from pathlib import Path

import pytest

from project_atlas.compiler import compile_target
from project_atlas.profile import ProjectProfile, load_profile
from project_atlas.scaffolder import initialize_project
from project_atlas.validator import validate_project


def test_init_validate_and_compile(tmp_path: Path):
    profile = load_profile(Path("examples/brasa/project-profile.json"))
    project = tmp_path / "brasa"
    result = initialize_project(project, profile)
    assert result.skills
    assert (project / "atlas.json").exists()
    assert (project / ".ai/skills/manifest.json").exists()
    assert (project / ".atlas/history/project-intelligence.json").exists()
    assert not (project / "PROJECT_MANIFEST.yaml").exists()
    assert not validate_project(project)

    created = compile_target(project, "codex")
    assert project / "AGENTS.md" in created
    assert (project / ".codex/skills/atlas-navigation/SKILL.md").exists()

    generic = compile_target(project, "generic")
    assert (project / ".atlas/runtime/compiled/generic/ENTRYPOINT.md") in generic

    traycer = compile_target(project, "traycer")
    assert (project / ".traycer/PROJECT_ATLAS.md") in traycer


def test_init_requires_explicit_model_roster(tmp_path: Path):
    profile = ProjectProfile({"version": 2, "project": {"name": "X", "type": ["cli"]}, "ai": {}})
    with pytest.raises(ValueError, match="Preferred LLMs"):
        initialize_project(tmp_path / "x", profile)
