from __future__ import annotations

import argparse
import json
import shutil
import sys
from pathlib import Path

from .compiler import SUPPORTED_TARGETS, compile_target
from .context import plan_context
from .goals import new_goal, transition_goal
from .intelligence import load_intelligence, record_task_report
from .migration import migrate_v01_to_v02
from .io import dump_json, load_data, load_json
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
                "version": 2,
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
    print(f"Initialized Project Atlas v0.2 in {root}")
    print(
        f"Agents: {len(resolution.agents)} | "
        f"Skills: {len(resolution.skills)} | Recipes: {len(resolution.recipes)}"
    )
    return 0


def cmd_resolve(args: argparse.Namespace) -> int:
    profile = load_profile(Path(args.profile).resolve())
    result = resolve(profile)
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
    path = root / ".ai/goals" / args.phase / f"{args.id}.goal.json"
    if path.exists() and not args.force:
        print(f"Goal already exists: {path}", file=sys.stderr)
        return 2
    dump_json(goal, path)
    print(path)
    return 0


def _find_goal(root: Path, goal_id: str) -> Path:
    json_matches = list((root / ".ai/goals").glob(f"**/{goal_id}.goal.json"))
    if len(json_matches) == 1:
        return json_matches[0]
    legacy = list((root / ".ai/goals").glob(f"**/{goal_id}.goal.yaml"))
    if len(legacy) == 1:
        return legacy[0]
    raise ValueError(
        f"Expected exactly one goal with id {goal_id}, "
        f"found {len(json_matches) + len(legacy)}"
    )


def cmd_goal_state(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    path = _find_goal(root, args.id)
    goal = transition_goal(path, args.state, args.reason or "")
    print(f"{args.id}: {goal['state']}")
    return 0


def cmd_goal_list(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    paths = sorted((root / ".ai/goals").glob("**/*.goal.json"))
    paths += sorted((root / ".ai/goals").glob("**/*.goal.yaml"))
    for path in paths:
        goal = load_data(path)
        legacy = " [legacy]" if path.suffix == ".yaml" else ""
        print(f"{goal.get('id')}\t{goal.get('state')}\t{goal.get('title')}{legacy}")
    return 0


def cmd_context_plan(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    plan = plan_context(root, args.task)
    if args.json:
        print(
            json.dumps(
                {
                    "profile": plan.profile,
                    "strategy": plan.strategy,
                    "budget": plan.budget,
                    "reasons": plan.reasons,
                },
                indent=2,
            )
        )
    else:
        print(f"Profile: {plan.profile}")
        print(f"Strategy: {plan.strategy}")
        for key, value in plan.budget.items():
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
    created = migrate_v01_to_v02(root)
    print(f"Migrated Project Atlas v0.1 -> v0.2: {len(created)} canonical files created/converted")
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
    project_file = root / "atlas.json"
    legacy = root / "PROJECT_MANIFEST.yaml"
    checks = {
        "python": shutil.which("python") or shutil.which("python3"),
        "git": shutil.which("git"),
        "atlas-project": project_file.exists() or legacy.exists(),
        "canonical-v0.2": project_file.exists(),
        "runtime-gitignored": _gitignore_contains(root, ".atlas/runtime/"),
    }
    failed = False
    for name, value in checks.items():
        ok = bool(value)
        failed |= not ok and name in {"python", "git", "atlas-project"}
        print(f"{'OK' if ok else 'WARN' if name not in {'python','git','atlas-project'} else 'FAIL'}  {name}: {value or 'not found'}")
    return 1 if failed else 0


def _gitignore_contains(root: Path, value: str) -> bool:
    path = root / ".gitignore"
    return path.exists() and value in path.read_text(encoding="utf-8").splitlines()


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
    parser.add_argument("--version", action="version", version="Project Atlas 0.2.0")
    sub = parser.add_subparsers(dest="command", required=True)

    init = sub.add_parser("init", help="Initialize Project Atlas in a project")
    init.add_argument("path", nargs="?", default=".")
    init.add_argument("--profile")
    init.add_argument("--non-interactive", action="store_true")
    init.set_defaults(func=cmd_init)

    resolve_p = sub.add_parser("resolve", help="Resolve agents, skills and recipes for a profile")
    resolve_p.add_argument("profile")
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
    migrate.set_defaults(func=cmd_migrate)

    compile_p = sub.add_parser("compile", help="Compile the active AI workforce for a target")
    compile_p.add_argument("--target", required=True, choices=sorted(SUPPORTED_TARGETS))
    compile_p.add_argument("--path", default=".")
    compile_p.set_defaults(func=cmd_compile)

    snapshot = sub.add_parser("snapshot", help="Create a portable recovery snapshot")
    snapshot.add_argument("path", nargs="?", default=".")
    snapshot.add_argument("--output")
    snapshot.set_defaults(func=cmd_snapshot)

    framework_check = sub.add_parser("framework-check", help="Validate built-in catalogs and adapters")
    framework_check.set_defaults(func=cmd_framework_check)

    doctor = sub.add_parser("doctor", help="Check local prerequisites/project format")
    doctor.add_argument("path", nargs="?", default=".")
    doctor.set_defaults(func=cmd_doctor)
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    try:
        return int(args.func(args))
    except (ValueError, FileNotFoundError, json.JSONDecodeError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2
