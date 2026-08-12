from pathlib import Path

from project_atlas.migration import migrate_v01_to_v02


def test_migrate_legacy_yaml_project(tmp_path: Path):
    root = tmp_path / "legacy"
    (root / ".atlas").mkdir(parents=True)
    (root / ".ai/goals/P00").mkdir(parents=True)
    (root / ".atlas/project-profile.yaml").write_text("""
version: 1
project:
  name: Legacy
  type: [cli]
ai:
  orchestrator: native
  autonomy: agentic
  preferred_models:
    - id: test/model
      provider: test
""")
    (root / "PROJECT_MANIFEST.yaml").write_text("framework: {name: project-atlas-framework, version: 0.1.0}\n")
    (root / ".ai/goals/P00/P00-G01.goal.yaml").write_text("""
id: P00-G01
title: Foundation
phase: P00
state: DRAFT
objective: Foundation
acceptance: [Works]
gates: {tests: required}
dependencies: []
evidence: []
""")

    migrate_v01_to_v02(root)

    assert (root / "atlas.json").exists()
    assert (root / ".ai/goals/P00/P00-G01.goal.json").exists()
    assert not (root / ".atlas/project-profile.yaml").exists()
    assert not (root / "PROJECT_MANIFEST.yaml").exists()
