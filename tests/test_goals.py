from pathlib import Path

import pytest

from project_atlas.goals import new_goal, transition_goal
from project_atlas.io import dump_yaml, load_yaml


def test_goal_state_machine_and_evidence(tmp_path: Path):
    path = tmp_path / "P01-G01.goal.yaml"
    dump_yaml(new_goal("P01-G01", "Runtime", "P01"), path)
    transition_goal(path, "PLANNED")
    transition_goal(path, "LOCKED")
    transition_goal(path, "EXECUTING")
    transition_goal(path, "VERIFYING")
    transition_goal(path, "REVIEWING")
    with pytest.raises(ValueError):
        transition_goal(path, "DONE")
    goal = load_yaml(path)
    goal["evidence"] = ["ci:test-run-1"]
    dump_yaml(goal, path)
    transition_goal(path, "DONE")
    assert load_yaml(path)["state"] == "DONE"


def test_goal_rejects_illegal_transition(tmp_path: Path):
    path = tmp_path / "goal.yaml"
    dump_yaml(new_goal("G1", "Thing", "P0"), path)
    with pytest.raises(ValueError):
        transition_goal(path, "EXECUTING")
