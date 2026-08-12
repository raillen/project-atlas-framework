from __future__ import annotations

from pathlib import Path

from .io import dump_json, load_data
from .profile import ProjectProfile
from .scaffolder import initialize_project

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
