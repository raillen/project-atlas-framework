from pathlib import Path

import pytest

from project_atlas.compiler import compile_target
from project_atlas.profile import ProjectProfile, load_profile
from project_atlas.scaffolder import initialize_project
from project_atlas.validator import validate_project


def test_init_validate_and_compile(tmp_path: Path):
    profile = load_profile(Path("examples/brasa/project-profile.yaml"))
    project = tmp_path / "brasa"
    result = initialize_project(project, profile)
    assert result.skills
    assert not validate_project(project)
    created = compile_target(project, "codex")
    assert project / "AGENTS.md" in created
    assert (project / ".codex/skills/atlas-navigation/SKILL.md").exists()
    traycer = compile_target(project, "traycer")
    assert (project / ".traycer/PROJECT_ATLAS.md") in traycer


def test_init_requires_explicit_model_roster(tmp_path: Path):
    profile = ProjectProfile({"version": 1, "project": {"name": "X", "type": ["cli"]}, "ai": {}})
    with pytest.raises(ValueError, match="Preferred LLMs"):
        initialize_project(tmp_path / "x", profile)
