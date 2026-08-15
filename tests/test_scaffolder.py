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

    # Codex compilation (Gate B)
    created = compile_target(project, "codex")
    assert project / "AGENTS.md" in created
    assert (project / ".codex/skills/clean-code/SKILL.md").exists()
    assert (project / ".codex/skills/clean-code/manifest.json").exists()
    assert (project / ".codex/skills/clean-code/checks/cleanliness-checklist.md").exists()
    assert (project / ".codex/skills/clean-code/templates/clean-code-review.md").exists()

    # Claude Code compilation (Gate B)
    created_claude = compile_target(project, "claude-code")
    assert project / "CLAUDE.md" in created_claude
    assert (project / ".claude/skills/threat-modeling/SKILL.md").exists()
    assert (project / ".claude/skills/threat-modeling/checks/threat-modeling-checklist.md").exists()

    generic = compile_target(project, "generic")
    assert (project / ".atlas/runtime/compiled/generic/ENTRYPOINT.md") in generic

    traycer = compile_target(project, "traycer")
    assert (project / ".traycer/PROJECT_ATLAS.md") in traycer


def test_init_requires_explicit_model_roster(tmp_path: Path):
    profile = ProjectProfile({"version": 2, "project": {"name": "X", "type": ["cli"]}, "ai": {}})
    with pytest.raises(ValueError, match="Preferred LLMs"):
        initialize_project(tmp_path / "x", profile)
