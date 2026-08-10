from __future__ import annotations

import argparse
import shutil
import sys
from pathlib import Path

from .compiler import SUPPORTED_TARGETS, compile_target
from .goals import new_goal, transition_goal
from .io import dump_yaml, load_yaml
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
                "version": 1,
                "project": {"name": name, "type": [project_type]},
                "stack": {},
                "features": [],
                "quality": {},
                "ai": {
                    "orchestrator": "traycer",
                    "autonomy": "agentic",
                    "preferred_models": [
                        {"id": item.strip(), "provider": item.split("/", 1)[0] if "/" in item else "unknown"}
                        for item in models.split(",")
                        if item.strip()
                    ],
                },
            }
        )
    resolution = initialize_project(root, profile)
    print(f"Initialized Project Atlas in {root}")
    print(f"Agents: {len(resolution.agents)} | Skills: {len(resolution.skills)} | Recipes: {len(resolution.recipes)}")
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
    path = root / ".ai/goals" / args.phase / f"{args.id}.goal.yaml"
    if path.exists() and not args.force:
        print(f"Goal already exists: {path}", file=sys.stderr)
        return 2
    dump_yaml(goal, path)
    print(path)
    return 0


def cmd_goal_state(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    matches = list((root / ".ai/goals").glob(f"**/{args.id}.goal.yaml"))
    if len(matches) != 1:
        print(f"Expected exactly one goal with id {args.id}, found {len(matches)}", file=sys.stderr)
        return 2
    goal = transition_goal(matches[0], args.state, args.reason or "")
    print(f"{args.id}: {goal['state']}")
    return 0


def cmd_goal_list(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    for path in sorted((root / ".ai/goals").glob("**/*.goal.yaml")):
        goal = load_yaml(path)
        print(f"{goal.get('id')}\t{goal.get('state')}\t{goal.get('title')}")
    return 0


def cmd_compile(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    created = compile_target(root, args.target)
    print(f"Compiled {args.target}: {len(created)} files")
    return 0


def cmd_snapshot(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    output = Path(args.output).resolve() if args.output else root / ".atlas" / f"{root.name}-atlas-snapshot.zip"
    create_snapshot(root, output)
    print(output)
    return 0


def cmd_doctor(args: argparse.Namespace) -> int:
    root = _project_root(args.path)
    checks = {
        "python": shutil.which("python") or shutil.which("python3"),
        "git": shutil.which("git"),
        "atlas-project": (root / "PROJECT_MANIFEST.yaml").exists(),
        "profile": (root / ".atlas/project-profile.yaml").exists(),
    }
    failed = False
    for name, value in checks.items():
        ok = bool(value)
        failed |= not ok
        print(f"{'OK' if ok else 'FAIL'}  {name}: {value or 'not found'}")
    return 1 if failed else 0


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
    parser.add_argument("--version", action="version", version="Project Atlas 0.1.0")
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

    doctor = sub.add_parser("doctor", help="Check local prerequisites")
    doctor.add_argument("path", nargs="?", default=".")
    doctor.set_defaults(func=cmd_doctor)
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    try:
        return int(args.func(args))
    except (ValueError, FileNotFoundError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2
