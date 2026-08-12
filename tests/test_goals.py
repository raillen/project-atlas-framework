from pathlib import Path

import pytest

from project_atlas.goals import new_goal, transition_goal
from project_atlas.io import dump_json, load_json


def test_goal_state_machine_and_evidence(tmp_path: Path):
    path = tmp_path / "P01-G01.goal.json"
    dump_json(new_goal("P01-G01", "Runtime", "P01"), path)
    transition_goal(path, "PLANNED")
    transition_goal(path, "LOCKED")
    transition_goal(path, "EXECUTING")
    transition_goal(path, "VERIFYING")
    transition_goal(path, "REVIEWING")
    with pytest.raises(ValueError):
        transition_goal(path, "DONE")
    goal = load_json(path)
    goal["evidence"] = ["ci:test-run-1"]
    dump_json(goal, path)
    transition_goal(path, "DONE")
    assert load_json(path)["state"] == "DONE"


def test_goal_rejects_illegal_transition(tmp_path: Path):
    path = tmp_path / "goal.json"
    dump_json(new_goal("G1", "Thing", "P0"), path)
    with pytest.raises(ValueError):
        transition_goal(path, "EXECUTING")


def test_legacy_goal_is_read_only_until_migrated(tmp_path: Path):
    path = tmp_path / "goal.goal.yaml"
    path.write_text("id: G1\ntitle: Thing\nphase: P0\nstate: DRAFT\nobjective: X\nacceptance: [X]\ngates: {tests: required}\ndependencies: []\nevidence: []\n")
    with pytest.raises(ValueError, match="read-only"):
        transition_goal(path, "PLANNED")
