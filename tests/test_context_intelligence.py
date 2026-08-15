from pathlib import Path

from project_atlas.context import plan_context
from project_atlas.intelligence import record_task_report
from project_atlas.profile import load_profile
from project_atlas.scaffolder import initialize_project


def _project(tmp_path: Path) -> Path:
    root = tmp_path / "p"
    initialize_project(root, load_profile(Path("examples/brasa/project-profile.json")))
    return root


def test_context_plan_is_bounded(tmp_path: Path):
    root = _project(tmp_path)
    small = plan_context(root, "rename a label")
    assert small.profile == "small"
    assert small.strategy == "direct"
    assert small.budget.max_delegation_depth == 0

    large = plan_context(root, "architecture migration")
    assert large.profile == "large"
    assert large.budget.max_delegation_depth == 1


def test_intelligence_aggregates_input_and_output_separately(tmp_path: Path):
    root = _project(tmp_path)
    data = record_task_report(root, {
        "id": "T1",
        "status": "success",
        "tokens": {"input": 1200, "output": 300, "cached": 400, "intermediate_output": 100},
        "cost": {"direct": {"amount": 0.25, "provenance": "observed"}}
    })
    assert data["summary"]["input_tokens"] == 1200
    assert data["summary"]["output_tokens"] == 300
    assert data["summary"]["intermediate_output_tokens"] == 100
    assert data["summary"]["direct_cost"] == 0.25
