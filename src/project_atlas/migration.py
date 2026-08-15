from __future__ import annotations

from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from .goals import compute_goal_digest
from .io import dump_json, load_data, load_json
from .profile import ProjectProfile
from .scaffolder import initialize_project
from .snapshot import create_snapshot

LEGACY_FILES = [
    "PROJECT_MANIFEST.yaml",
    ".atlas/project-profile.yaml",
    ".ai/agents/manifest.yaml",
    ".ai/skills/manifest.yaml",
    ".ai/recipes/manifest.yaml",
    ".ai/orchestration/model-policy.yaml",
    ".ai/orchestration/orchestrator.yaml",
    ".ai/orchestration/fallbacks.yaml",
    ".ai/orchestration/model-scorecard.yaml",
]


def migrate_v01_to_v02(root: Path) -> list[Path]:
    profile_path = root / ".atlas/project-profile.yaml"
    if not profile_path.exists():
        raise FileNotFoundError("Legacy .atlas/project-profile.yaml not found")
    profile = ProjectProfile(load_data(profile_path))
    legacy_goals: list[tuple[Path, dict]] = []
    goals_root = root / ".ai/goals"
    if goals_root.exists():
        for path in sorted(goals_root.glob("**/*.goal.yaml")):
            legacy_goals.append((path, load_data(path)))

    initialize_project(root, profile)
    created: list[Path] = [root / "atlas.json"]

    for old_path, goal in legacy_goals:
        new_path = old_path.parent / old_path.name.replace(".goal.yaml", ".goal.json")
        dump_json(goal, new_path)
        created.append(new_path)

    for rel in LEGACY_FILES:
        path = root / rel
        if path.exists():
            path.unlink()
    for old_path, _ in legacy_goals:
        if old_path.exists():
            old_path.unlink()

    return created


def migrate_v02_to_v03(root: Path, dry_run: bool = False) -> dict[str, Any]:
    atlas_json_path = root / "atlas.json"
    if not atlas_json_path.exists():
        raise FileNotFoundError("atlas.json not found; cannot migrate to v0.3")

    report: dict[str, Any] = {
        "from_version": 2,
        "to_version": 3,
        "dry_run": dry_run,
        "changes": [],
        "snapshot": None
    }

    if not dry_run:
        snapshot_out = root / ".atlas" / "snapshots" / f"pre_migration_v03_{datetime.now(timezone.utc).strftime('%Y%m%d%H%M%S')}.zip"
        snapshot_path = create_snapshot(root, snapshot_out)
        report["snapshot"] = str(snapshot_path)

    # 1. Update atlas.json
    atlas_data = load_json(atlas_json_path)
    if atlas_data.get("version", 2) < 3 or "protocol" not in atlas_data:
        atlas_data["version"] = 3
        atlas_data["protocol"] = {"version": 3, "compatible": ">=3 <4"}
        report["changes"].append("Updated atlas.json to version 3 with protocol metadata")
        if not dry_run:
            dump_json(atlas_data, atlas_json_path)

    # 2. Update model-policy.json
    mp_path = root / ".ai/orchestration/model-policy.json"
    if mp_path.exists():
        mp_data = load_json(mp_path)
        if mp_data.get("version", 1) < 2 or "profiles" not in mp_data:
            mp_data["version"] = 2
            if "profiles" not in mp_data:
                roster = mp_data.get("roster", [])
                mp_data["profiles"] = {
                    item: {
                        "provider": item.split("/")[0] if "/" in item else "unknown",
                        "model": item.split("/")[-1],
                        "capabilities": ["code", "tool-use"],
                        "cost_class": "medium"
                    }
                    for item in roster if isinstance(item, str)
                }
            report["changes"].append("Upgraded model-policy.json to v2 with profiles")
            if not dry_run:
                dump_json(mp_data, mp_path)

    # 3. Update Goals to Goal v2 format (ensure revision and lock if locked)
    goals_root = root / ".ai/goals"
    if goals_root.exists():
        for gpath in sorted(goals_root.glob("**/*.goal.json")):
            goal_data = load_json(gpath)
            modified = False
            if "revision" not in goal_data:
                goal_data["revision"] = 1
                modified = True
            state = str(goal_data.get("state", "DRAFT")).upper()
            if state in {"LOCKED", "EXECUTING", "VERIFYING", "REVIEWING", "DONE"} and "lock" not in goal_data:
                goal_data["lock"] = {
                    "revision": goal_data["revision"],
                    "digest": compute_goal_digest(goal_data),
                    "locked_at": datetime.now(timezone.utc).isoformat(),
                }
                modified = True
            if modified:
                report["changes"].append(f"Updated Goal {gpath.name} to Goal v2 contract")
                if not dry_run:
                    dump_json(goal_data, gpath)

    return report


def migrate_project(root: Path, dry_run: bool = False) -> dict[str, Any]:
    profile_yaml = root / ".atlas/project-profile.yaml"
    if profile_yaml.exists():
        if dry_run:
            return {"from_version": 1, "to_version": 3, "dry_run": True, "changes": ["Migrate v1 YAML to v2 JSON, then v2 to v3"]}
        migrate_v01_to_v02(root)
        return migrate_v02_to_v03(root, dry_run=False)
    return migrate_v02_to_v03(root, dry_run=dry_run)
