from __future__ import annotations

import argparse
import dataclasses
import json
import shutil
import sys
from pathlib import Path

from .compiler import SUPPORTED_TARGETS, compile_target
from .context import plan_context
from .doctor import diagnose_project
from .explain import (
    explain_agent,
    explain_context,
    explain_execution,
    explain_model,
    explain_recipe,
    explain_skill,
    explain_workforce,
)
from .goals import amend_goal, new_goal, transition_goal
from .intelligence import load_intelligence, record_task_report
from .io import dump_json, load_data, load_json
from .migration import migrate_project
from .profile import ProjectProfile, load_profile
from .resolver import resolve
from .scaffolder import initialize_project
from .snapshot import create_snapshot
from .validator import validate_framework, validate_project


def _project_root(value: str | None) -> Path:
    return Path(value or ".").resolve()


def cmd_init(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    profile_path = Path(args.profile).resolve() if args.profile else None
    if profile_path:
        profile = load_profile(profile_path)
    else:
        if args.non_interactive:
            print("--profile is required with --non-interactive", file=sys.stderr)
            return 2
        name = input(f"Project name [{root.name}]: ").strip() or root.name
        project_type = input("Project type (e.g. web-saas, game-engine, cli): ").strip()
        models = input("Preferred LLMs (provider/model, comma separated): ").strip()
        if not models:
            print("At least one preferred LLM is required.", file=sys.stderr)
            return 2
        profile = ProjectProfile(
            {
                "version": 3,
                "protocol": {"version": 3, "compatible": ">=3 <4"},
                "project": {"name": name, "type": [project_type]},
                "stack": {},
                "features": [],
                "quality": {},
                "ai": {
                    "orchestrator": "native",
                    "autonomy": "agentic",
                    "preferred_models": [
                        {
                            "id": item.strip(),
                            "provider": item.split("/", 1)[0] if "/" in item else "unknown",
                        }
                        for item in models.split(",")
                        if item.strip()
                    ],
                },
            }
        )
    resolution = initialize_project(root, profile)
    print(f"Initialized Project Atlas v0.3 in {root}")
    print(
        f"Agents: {len(resolution.agents)} | "
        f"Skills: {len(resolution.skills)} | Recipes: {len(resolution.recipes)}"
    )
    return 0


def cmd_resolve(args: argparse.Namespace) -> int:
    profile = load_profile(Path(args.profile).resolve())
    result = resolve(profile)
    if getattr(args, "json", False):
        print(json.dumps({
            "agents": result.agents,
            "skills": result.skills,
            "recipes": result.recipes,
            "reasons": result.reasons,
            "traces": result.traces
        }, indent=2))
        return 0
    print("Agents:")
    for value in result.agents:
        print(f"  - {value}: {', '.join(result.reasons.get(value, []))}")
    print("Skills:")
    for value in result.skills:
        print(f"  - {value}: {', '.join(result.reasons.get(value, []))}")
    print("Recipes:")
    for value in result.recipes:
        print(f"  - {value}: {', '.join(result.reasons.get(value, []))}")
    return 0


def cmd_validate(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    errors = validate_project(root, Path(args.schemas).resolve() if args.schemas else None)
    if errors:
        print("Validation failed:")
        for error in errors:
            print(f"- {error}")
        return 1
    print("Project Atlas validation passed.")
    return 0


def cmd_goal_new(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    goal = new_goal(args.id, args.title, args.phase, args.objective or "")
    goal_dir = root / ".ai/goals"
    goal_dir.mkdir(parents=True, exist_ok=True)
    target = goal_dir / f"{args.id}.goal.json"
    if target.exists() and not args.force:
        print(f"Goal already exists: {target} (use --force to overwrite)", file=sys.stderr)
        return 2
    dump_json(goal, target)
    print(f"Created Goal {args.id} in {target}")
    return 0


def cmd_goal_state(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    matches = list((root / ".ai/goals").glob(f"**/{args.id}.goal.*"))
    if not matches:
        print(f"Goal {args.id} not found in {root / '.ai/goals'}", file=sys.stderr)
        return 2
    goal = transition_goal(matches[0], args.state, args.reason or "")
    print(f"Goal {args.id} transitioned to {goal['state']}")
    return 0


def cmd_goal_amend(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    matches = list((root / ".ai/goals").glob(f"**/{args.id}.goal.*"))
    if not matches:
        print(f"Goal {args.id} not found in {root / '.ai/goals'}", file=sys.stderr)
        return 2
    amendment_data = load_json(Path(args.file).resolve()) if args.file else {
        "reason": args.reason or "CLI amendment",
        "approved_by": args.approved_by or "human",
        "changes": {}
    }
    goal = amend_goal(matches[0], amendment_data)
    print(f"Goal {args.id} amended to revision {goal.get('revision')}")
    return 0


def cmd_goal_list(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    goals_dir = root / ".ai/goals"
    if not goals_dir.exists():
        print("No goals directory found.")
        return 0
    goals = []
    for path in sorted(goals_dir.glob("**/*.goal.*")):
        data = load_data(path)
        goals.append(
            {
                "id": data.get("id", path.stem),
                "title": data.get("title", ""),
                "phase": data.get("phase", ""),
                "state": data.get("state", "DRAFT"),
                "revision": data.get("revision", 1),
                "path": str(path.relative_to(root)),
            }
        )
    print(f"{'ID':<12} {'PHASE':<8} {'STATE':<12} {'REV':<5} {'TITLE'}")
    print("-" * 65)
    for g in goals:
        print(f"{g['id']:<12} {g['phase']:<8} {g['state']:<12} {g['revision']:<5} {g['title']}")
    return 0


def cmd_context_plan(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    plan = plan_context(root, args.task)
    if args.json:
        print(
            json.dumps(
                {
                    "task": plan.task,
                    "profile": plan.profile,
                    "strategy": plan.strategy,
                    "budget": dataclasses.asdict(plan.budget) if hasattr(plan.budget, "__dataclass_fields__") else (plan.budget.__dict__ if hasattr(plan.budget, "__dict__") else plan.budget),
                    "reasons": plan.reasons,
                },
                indent=2,
            )
        )
    else:
        print(f"Profile: {plan.profile}")
        print(f"Strategy: {plan.strategy}")
        for key, value in dataclasses.asdict(plan.budget).items():
            print(f"{key}: {value}")
    return 0


def cmd_report_add(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    report = load_json(Path(args.file).resolve())
    data = record_task_report(root, report)
    print(f"Recorded task {report['id']}; total tasks: {data['summary'].get('tasks', 0)}")
    return 0


def cmd_report_summary(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    data = load_intelligence(root)
    print(json.dumps(data.get("summary", {}), indent=2))
    return 0


def cmd_migrate(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    dry_run = getattr(args, "dry_run", False)
    report = migrate_project(root, dry_run=dry_run)
    if getattr(args, "json", False):
        print(json.dumps(report, indent=2))
    else:
        prefix = "[DRY-RUN] " if dry_run else ""
        print(f"{prefix}Project migration report (v{report.get('from_version')} -> v{report.get('to_version')}):")
        for change in report.get("changes", []):
            print(f"  - {change}")
        if report.get("snapshot"):
            print(f"  Snapshot backup created: {report['snapshot']}")
    return 0


def cmd_compile(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    created = compile_target(root, args.target)
    print(f"Compiled {args.target}: {len(created)} files")
    return 0


def cmd_snapshot(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    output = (
        Path(args.output).resolve()
        if args.output
        else root / ".atlas" / f"{root.name}-atlas-snapshot.zip"
    )
    create_snapshot(root, output)
    print(output)
    return 0


def cmd_doctor(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    findings = diagnose_project(root, Path(args.schemas).resolve() if getattr(args, "schemas", None) else None)
    has_errors = any(f.severity == "ERROR" for f in findings)
    if getattr(args, "json", False):
        print(json.dumps([
            {
                "category": f.category,
                "severity": f.severity,
                "message": f.message,
                "target": f.target
            }
            for f in findings
        ], indent=2))
    else:
        if not findings:
            print("Project Atlas Doctor: all checks passed cleanly.")
        else:
            print(f"Project Atlas Doctor found {len(findings)} issue(s):")
            for f in findings:
                print(f"[{f.severity}] ({f.category}) {f.message} {f'[{f.target}]' if f.target else ''}")
    return 1 if has_errors else 0


def cmd_explain(args: argparse.Namespace) -> int:
    topic = getattr(args, "topic", "")
    target = getattr(args, "target", None)
    root = _project_root(getattr(args, "path", "."))
    is_json = getattr(args, "json", False)

    result: Any = None
    if topic == "workforce":
        if target:
            profile = load_profile(Path(target).resolve())
        elif (root / "atlas.json").exists():
            profile = ProjectProfile(load_json(root / "atlas.json"))
        else:
            print("Please specify a profile path or run inside a Project Atlas workspace.", file=sys.stderr)
            return 2
        result = explain_workforce(profile)
    elif topic == "agent":
        if not target:
            print("Agent ID required.", file=sys.stderr)
            return 2
        result = explain_agent(target)
    elif topic == "skill":
        if not target:
            print("Skill ID required.", file=sys.stderr)
            return 2
        result = explain_skill(target)
    elif topic == "recipe":
        if not target:
            print("Recipe ID required.", file=sys.stderr)
            return 2
        result = explain_recipe(target)
    elif topic == "context":
        if not target:
            print("Task ID required.", file=sys.stderr)
            return 2
        result = explain_context(target, root)
    elif topic == "model":
        if not target:
            print("Role name required.", file=sys.stderr)
            return 2
        result = explain_model(target, root)
    elif topic == "execution":
        if not target:
            print("Profile ID required.", file=sys.stderr)
            return 2
        result = explain_execution(target, root)
    else:
        print(f"Unknown explain topic: {topic}", file=sys.stderr)
        return 2

    if result is None:
        print(f"Not found: {topic} {target}", file=sys.stderr)
        return 1

    if is_json:
        print(json.dumps(result, indent=2))
    else:
        if isinstance(result, dict):
            for k, v in result.items():
                if isinstance(v, (dict, list)):
                    print(f"{k}:")
                    print(json.dumps(v, indent=2))
                else:
                    print(f"{k}: {v}")
        else:
            print(result)
    return 0


def cmd_framework_check(args: argparse.Namespace) -> int:
    errors = validate_framework()
    if errors:
        print("Framework validation failed:")
        for error in errors:
            print(f"- {error}")
        return 1
    print("Project Atlas framework validation passed.")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="atlas", description="Project Atlas Framework CLI")
    parser.add_argument("--version", action="version", version="Project Atlas 0.3.0")
    sub = parser.add_subparsers(dest="command", required=True)

    init = sub.add_parser("init", help="Initialize Project Atlas in a project")
    init.add_argument("path", nargs="?", default=".")
    init.add_argument("--profile")
    init.add_argument("--non-interactive", action="store_true")
    init.set_defaults(func=cmd_init)

    resolve_p = sub.add_parser("resolve", help="Resolve agents, skills and recipes for a profile")
    resolve_p.add_argument("profile")
    resolve_p.add_argument("--json", action="store_true")
    resolve_p.set_defaults(func=cmd_resolve)

    validate = sub.add_parser("validate", help="Validate a Project Atlas project")
    validate.add_argument("path", nargs="?", default=".")
    validate.add_argument("--schemas")
    validate.set_defaults(func=cmd_validate)

    goal = sub.add_parser("goal", help="Manage Goals")
    goal_sub = goal.add_subparsers(dest="goal_command", required=True)
    goal_new = goal_sub.add_parser("new")
    goal_new.add_argument("id")
    goal_new.add_argument("title")
    goal_new.add_argument("--phase", required=True)
    goal_new.add_argument("--objective")
    goal_new.add_argument("--path", default=".")
    goal_new.add_argument("--force", action="store_true")
    goal_new.set_defaults(func=cmd_goal_new)

    goal_state = goal_sub.add_parser("state")
    goal_state.add_argument("id")
    goal_state.add_argument("state")
    goal_state.add_argument("--reason")
    goal_state.add_argument("--path", default=".")
    goal_state.set_defaults(func=cmd_goal_state)

    goal_amend_p = goal_sub.add_parser("amend")
    goal_amend_p.add_argument("id")
    goal_amend_p.add_argument("--file")
    goal_amend_p.add_argument("--reason")
    goal_amend_p.add_argument("--approved-by")
    goal_amend_p.add_argument("--path", default=".")
    goal_amend_p.set_defaults(func=cmd_goal_amend)

    goal_list = goal_sub.add_parser("list")
    goal_list.add_argument("--path", default=".")
    goal_list.set_defaults(func=cmd_goal_list)

    context = sub.add_parser("context", help="Lean Progressive Context utilities")
    context_sub = context.add_subparsers(dest="context_command", required=True)
    context_plan = context_sub.add_parser("plan", help="Create a deterministic initial context/budget plan")
    context_plan.add_argument("task")
    context_plan.add_argument("--path", default=".")
    context_plan.add_argument("--json", action="store_true")
    context_plan.set_defaults(func=cmd_context_plan)

    report = sub.add_parser("report", help="Project Intelligence")
    report_sub = report.add_subparsers(dest="report_command", required=True)
    report_add = report_sub.add_parser("add", help="Record/replace a task report from JSON")
    report_add.add_argument("file")
    report_add.add_argument("--path", default=".")
    report_add.set_defaults(func=cmd_report_add)
    report_summary = report_sub.add_parser("summary", help="Show aggregated project intelligence")
    report_summary.add_argument("--path", default=".")
    report_summary.set_defaults(func=cmd_report_summary)

    migrate = sub.add_parser("migrate", help="Migrate a legacy Project Atlas project")
    migrate.add_argument("path", nargs="?", default=".")
    migrate.add_argument("--dry-run", action="store_true")
    migrate.add_argument("--json", action="store_true")
    migrate.set_defaults(func=cmd_migrate)

    compile_p = sub.add_parser("compile", help="Compile the active AI workforce for a target")
    compile_p.add_argument("--target", required=True, choices=sorted(SUPPORTED_TARGETS))
    compile_p.add_argument("--path", default=".")
    compile_p.set_defaults(func=cmd_compile)

    snapshot = sub.add_parser("snapshot", help="Create a portable recovery snapshot")
    snapshot.add_argument("path", nargs="?", default=".")
    snapshot.add_argument("--output")
    snapshot.set_defaults(func=cmd_snapshot)

    doctor = sub.add_parser("doctor", help="Run comprehensive project health diagnostics")
    doctor.add_argument("path", nargs="?", default=".")
    doctor.add_argument("--schemas")
    doctor.add_argument("--json", action="store_true")
    doctor.set_defaults(func=cmd_doctor)

    explain = sub.add_parser("explain", help="Explain workforce, reasoning, context or policies")
    explain.add_argument("topic", choices=["workforce", "agent", "skill", "recipe", "context", "model", "execution"])
    explain.add_argument("target", nargs="?", default=None)
    explain.add_argument("--path", default=".")
    explain.add_argument("--json", action="store_true")
    explain.set_defaults(func=cmd_explain)

    framework_check = sub.add_parser("framework-check", help="Validate built-in catalogs and adapters")
    framework_check.set_defaults(func=cmd_framework_check)
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    try:
        return int(args.func(args))
    except (ValueError, FileNotFoundError, json.JSONDecodeError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2
