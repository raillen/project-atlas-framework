from pathlib import Path

import pytest

from project_atlas.goals import amend_goal, compute_goal_digest, new_goal, transition_goal, verify_goal_lock
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


def test_goal_v2_lock_and_mutation_detection(tmp_path: Path):
    path = tmp_path / "G02.goal.json"
    dump_json(new_goal("G02", "Security Engine", "P01"), path)
    transition_goal(path, "PLANNED")
    goal_locked = transition_goal(path, "LOCKED")
    assert "lock" in goal_locked
    assert goal_locked["lock"]["revision"] == 1
    assert goal_locked["lock"]["digest"] == compute_goal_digest(goal_locked)

    # Mutate objective without amendment
    goal_locked["objective"] = "Tampered objective"
    dump_json(goal_locked, path)

    # Next transition should fail due to lock violation
    with pytest.raises(ValueError, match="Illegal goal mutation detected"):
        transition_goal(path, "EXECUTING")


def test_goal_v2_amendment(tmp_path: Path):
    path = tmp_path / "G03.goal.json"
    dump_json(new_goal("G03", "Parser", "P01"), path)
    transition_goal(path, "PLANNED")
    transition_goal(path, "LOCKED")

    amendment = {
        "id": "AMD-002",
        "reason": "Add streaming parse criteria",
        "approved_by": "lead",
        "changes": {
            "acceptance": ["Supports AST streaming"]
        }
    }
    amended = amend_goal(path, amendment)
    assert amended["revision"] == 2
    assert len(amended["amendments"]) == 1
    assert amended["acceptance"] == ["Supports AST streaming"]
    valid, _ = verify_goal_lock(amended)
    assert valid is True

    # Now transition to EXECUTING should succeed
    transition_goal(path, "EXECUTING")
    assert load_json(path)["state"] == "EXECUTING"
