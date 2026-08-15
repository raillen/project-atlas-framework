from __future__ import annotations

from dataclasses import dataclass
import json
from pathlib import Path
from typing import Any

from .catalog import by_id
from .goals import verify_goal_lock
from .io import load_json
from .validator import validate_project


@dataclass
class DoctorFinding:
    category: str
    severity: str  # "ERROR", "WARNING", "INFO"
    message: str
    target: str = ""


def check_dag_cycles(tasks: list[dict[str, Any]]) -> list[str]:
    """Validate that task dependencies form an acyclic graph."""
    errors = []
    task_map = {t["id"]: t for t in tasks if isinstance(t, dict) and "id" in t}
    graph = {tid: list(task_map[tid].get("dependencies", [])) for tid in task_map}

    # Check for self-dependencies and missing dependencies
    for tid, deps in graph.items():
        if tid in deps:
            errors.append(f"Task {tid} has a self-dependency.")
        for dep in deps:
            if dep not in task_map:
                errors.append(f"Task {tid} depends on non-existent task '{dep}'.")

    # Cycle detection via DFS
    visited: dict[str, int] = {}  # 0: unvisited, 1: visiting, 2: visited

    def dfs(node: str, path: list[str]) -> bool:
        visited[node] = 1
        for neighbor in graph.get(node, []):
            if neighbor not in graph:
                continue
            if visited.get(neighbor, 0) == 1:
                cycle_str = " -> ".join(path + [neighbor])
                errors.append(f"DAG cycle detected: {cycle_str}")
                return True
            if visited.get(neighbor, 0) == 0:
                if dfs(neighbor, path + [neighbor]):
                    return True
        visited[node] = 2
        return False

    for node in graph:
        if visited.get(node, 0) == 0:
            dfs(node, [node])

    return errors


def diagnose_project(root: Path, schema_dir: Path | None = None) -> list[DoctorFinding]:
    findings: list[DoctorFinding] = []

    # 1. Base Project & Schema Validation
    validation_errors = validate_project(root, schema_dir)
    for err in validation_errors:
        findings.append(DoctorFinding(
            category="schema/structure",
            severity="ERROR",
            message=err,
            target=str(root)
        ))

    # 2. Protocol & Version Compatibility
    atlas_json_path = root / "atlas.json"
    if atlas_json_path.exists():
        try:
            atlas_data = load_json(atlas_json_path)
            version = atlas_data.get("version", 1)
            if version < 2:
                findings.append(DoctorFinding(
                    category="version",
                    severity="ERROR",
                    message=f"Project version {version} is deprecated; run 'atlas migrate'",
                    target="atlas.json"
                ))
            protocol = atlas_data.get("protocol")
            if protocol and protocol.get("version", 0) > 3:
                findings.append(DoctorFinding(
                    category="protocol",
                    severity="WARNING",
                    message=f"Project protocol version {protocol.get('version')} is newer than framework runtime v0.3",
                    target="atlas.json"
                ))
        except Exception as e:
            findings.append(DoctorFinding(
                category="parse",
                severity="ERROR",
                message=f"Failed to parse atlas.json: {e}",
                target="atlas.json"
            ))

    # 3. Goal Consistency & Lock Integrity
    goals_dir = root / ".ai/goals"
    goal_ids: set[str] = set()
    if goals_dir.exists():
        for gpath in sorted(goals_dir.glob("**/*.goal.json")):
            try:
                goal_data = load_json(gpath)
                gid = goal_data.get("id", gpath.stem)
                goal_ids.add(gid)

                # Check lock integrity on locked goals
                state = str(goal_data.get("state", "DRAFT")).upper()
                if state in {"LOCKED", "EXECUTING", "VERIFYING", "REVIEWING", "DONE"}:
                    valid, msg = verify_goal_lock(goal_data)
                    if not valid:
                        findings.append(DoctorFinding(
                            category="goal-lock",
                            severity="ERROR",
                            message=f"Goal {gid} lock integrity violation: {msg}",
                            target=str(gpath.relative_to(root))
                        ))

                # Check goal dependencies
                for dep in goal_data.get("dependencies", []):
                    # Will verify after scanning all goals
                    pass
            except Exception as e:
                findings.append(DoctorFinding(
                    category="goal",
                    severity="ERROR",
                    message=f"Failed to read goal file {gpath.name}: {e}",
                    target=str(gpath.relative_to(root))
                ))

        # Re-verify goal dependencies exist
        for gpath in sorted(goals_dir.glob("**/*.goal.json")):
            try:
                goal_data = load_json(gpath)
                gid = goal_data.get("id")
                for dep in goal_data.get("dependencies", []):
                    if dep not in goal_ids:
                        findings.append(DoctorFinding(
                            category="goal-dependencies",
                            severity="ERROR",
                            message=f"Goal {gid} references non-existent dependency goal '{dep}'",
                            target=str(gpath.relative_to(root))
                        ))
            except Exception:
                pass

    # 4. Plans, Tasks, and DAG Cycle Checks
    plans_dir = root / ".ai/plans"
    if plans_dir.exists():
        for ppath in sorted(plans_dir.glob("**/*.plan.json")):
            try:
                plan_data = load_json(ppath)
                pid = plan_data.get("id", ppath.stem)
                pgid = plan_data.get("goal_id")
                if pgid and pgid not in goal_ids:
                    findings.append(DoctorFinding(
                        category="plan",
                        severity="ERROR",
                        message=f"Plan {pid} references unknown goal '{pgid}'",
                        target=str(ppath.relative_to(root))
                    ))

                tasks = plan_data.get("tasks", [])
                task_objects = [t for t in tasks if isinstance(t, dict)]
                if task_objects:
                    dag_errors = check_dag_cycles(task_objects)
                    for derr in dag_errors:
                        findings.append(DoctorFinding(
                            category="dag",
                            severity="ERROR",
                            message=f"Plan {pid} {derr}",
                            target=str(ppath.relative_to(root))
                        ))
            except Exception as e:
                findings.append(DoctorFinding(
                    category="plan",
                    severity="ERROR",
                    message=f"Failed to validate plan {ppath.name}: {e}",
                    target=str(ppath.relative_to(root))
                ))

    # 5. Workforce Manifest Reference Check
    known_skills = by_id("skills")
    known_agents = by_id("agents")
    known_recipes = by_id("recipes")

    skills_manifest = root / ".ai/skills/manifest.json"
    if skills_manifest.exists():
        try:
            s_list = load_json(skills_manifest).get("skills", [])
            for sk in s_list:
                if sk not in known_skills and not (root / f".ai/skills/{sk}").exists():
                    findings.append(DoctorFinding(
                        category="workforce",
                        severity="WARNING",
                        message=f"Selected skill '{sk}' is not registered in framework catalog or local skills",
                        target=".ai/skills/manifest.json"
                    ))
        except Exception:
            pass

    agents_manifest = root / ".ai/agents/manifest.json"
    if agents_manifest.exists():
        try:
            a_list = load_json(agents_manifest).get("agents", [])
            for ag in a_list:
                if ag not in known_agents and not (root / f".ai/agents/{ag}").exists():
                    findings.append(DoctorFinding(
                        category="workforce",
                        severity="WARNING",
                        message=f"Selected agent '{ag}' is not registered in framework catalog or local agents",
                        target=".ai/agents/manifest.json"
                    ))
        except Exception:
            pass

    # Recipe Manifest Check
    recipes_manifest = root / ".ai/recipes/manifest.json"
    if recipes_manifest.exists():
        try:
            r_list = load_json(recipes_manifest).get("recipes", [])
            for r in r_list:
                if r not in known_recipes and not (root / f".ai/recipes/{r}").exists():
                    findings.append(DoctorFinding("workforce", "WARNING", f"Selected recipe '{r}' is not registered", ".ai/recipes/manifest.json"))
        except Exception:
            pass
    else:
        findings.append(DoctorFinding("workforce", "WARNING", "Missing recipes manifest", ".ai/recipes/manifest.json"))

    # Skill Provenance Check
    for sk_file in (root / ".ai/skills").glob("**/*.skill.json"):
        try:
            sk_data = load_json(sk_file)
            if "provenance" not in sk_data or not isinstance(sk_data["provenance"], dict) or "origin" not in sk_data["provenance"]:
                findings.append(DoctorFinding("skill-provenance", "ERROR", "Skill missing provenance.origin", str(sk_file.relative_to(root))))
        except Exception:
            pass

    # Model Policy Diagnostics
    model_policy = root / ".ai/orchestration/model-policy.json"
    if model_policy.exists():
        try:
            mp_data = load_json(model_policy)
            roles = mp_data.get("roles", {})
            profiles = mp_data.get("profiles", {})
            for role, profile_name in roles.items():
                if profile_name not in profiles and profile_name != "default":
                    findings.append(DoctorFinding("model-policy", "ERROR", f"Role '{role}' references missing profile '{profile_name}'", ".ai/orchestration/model-policy.json"))
            for pname, prof in profiles.items():
                if "provider" not in prof or "model" not in prof:
                    findings.append(DoctorFinding("model-policy", "ERROR", f"Profile '{pname}' is incomplete (missing provider or model)", ".ai/orchestration/model-policy.json"))
        except Exception:
            pass

    # Execution Policy Checks
    exec_policy = root / ".ai/orchestration/execution-policy.json"
    if exec_policy.exists():
        try:
            ep_data = load_json(exec_policy)
            for bname, backend in ep_data.get("backends", {}).items():
                if "type" not in backend:
                    findings.append(DoctorFinding("execution-policy", "ERROR", f"Backend '{bname}' missing type", ".ai/orchestration/execution-policy.json"))
        except Exception:
            pass

    # Policy Conflicts
    if model_policy.exists() and exec_policy.exists():
        try:
            mp_data = load_json(model_policy)
            ep_data = load_json(exec_policy)
            if mp_data.get("enforce_execution_isolation", False) and ep_data.get("default_isolation", "none") == "none":
                findings.append(DoctorFinding("policy-conflict", "WARNING", "Model policy requires isolation but execution policy defaults to none", "policies"))
        except Exception:
            pass

    # Permission Scopes
    perm_policy = root / ".ai/orchestration/permission-policy.json"
    if perm_policy.exists() and agents_manifest.exists():
        try:
            pp_data = load_json(perm_policy)
            allowed_scopes = set(pp_data.get("allowed_scopes", []))
            for ag in load_json(agents_manifest).get("agents", []):
                ag_path = root / f".ai/agents/{ag}"
                if ag_path.exists():
                    ag_data = load_json(ag_path)
                    for scope in ag_data.get("permissions", {}).get("scopes", []):
                        if scope not in allowed_scopes:
                            findings.append(DoctorFinding("permission-scopes", "ERROR", f"Agent '{ag}' requires scope '{scope}' not allowed by permission policy", str(ag_path.relative_to(root))))
        except Exception:
            pass

    # Invalid Gates
    for ppath in (root / ".ai/plans").glob("**/*.plan.json"):
        try:
            plan_data = load_json(ppath)
            for task in plan_data.get("tasks", []):
                for gate in task.get("gates", []):
                    if "evidence_type" not in gate:
                        findings.append(DoctorFinding("invalid-gates", "ERROR", f"Gate in task '{task.get('id')}' missing evidence_type", str(ppath.relative_to(root))))
        except Exception:
            pass

    # Missing Evidence References
    evidence_dir = root / ".ai/evidence"
    evidence_ids = set()
    if evidence_dir.exists():
        for epath in evidence_dir.glob("**/*.evidence.json"):
            try:
                ev_data = load_json(epath)
                if "id" in ev_data:
                    evidence_ids.add(ev_data["id"])
            except Exception:
                pass
    for ppath in (root / ".ai/plans").glob("**/*.plan.json"):
        try:
            plan_data = load_json(ppath)
            for task in plan_data.get("tasks", []):
                for ev_id in task.get("evidence", []):
                    if ev_id not in evidence_ids:
                        findings.append(DoctorFinding("missing-evidence", "WARNING", f"Task '{task.get('id')}' references missing evidence '{ev_id}'", str(ppath.relative_to(root))))
        except Exception:
            pass

    # Compiled Adapter Freshness
    adapters_dir = root / ".atlas/runtime/adapters"
    if adapters_dir.exists():
        for adapter in adapters_dir.glob("*.py"):
            source = root / "ENTRYPOINT.md"
            if source.exists() and source.stat().st_mtime > adapter.stat().st_mtime:
                findings.append(DoctorFinding("adapter-freshness", "WARNING", f"Compiled adapter '{adapter.name}' is older than ENTRYPOINT.md", str(adapter.relative_to(root))))

    return findings
