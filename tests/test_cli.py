import json
from pathlib import Path

from project_atlas.cli import main


def test_cli_smoke(tmp_path: Path):
    project = tmp_path / "project"
    assert main(["init", str(project), "--profile", "examples/brasa/project-profile.json", "--non-interactive"]) == 0
    assert main(["goal", "new", "P00-G01", "Foundation", "--phase", "P00", "--path", str(project)]) == 0
    assert main(["goal", "list", "--path", str(project)]) == 0
    assert main(["context", "plan", "add a transition", "--path", str(project), "--json"]) == 0
    assert main(["compile", "--target", "generic", "--path", str(project)]) == 0
    assert main(["compile", "--target", "codex", "--path", str(project)]) == 0
    assert main(["compile", "--target", "claude-code", "--path", str(project)]) == 0

    report = tmp_path / "report.json"
    report.write_text(json.dumps({
        "id": "TASK-1",
        "status": "success",
        "tokens": {"input": 1000, "output": 200, "cached": 100},
        "cost": {"direct": {"amount": 0.12, "currency": "USD", "provenance": "observed"}}
    }))
    assert main(["report", "add", str(report), "--path", str(project)]) == 0
    assert main(["report", "summary", "--path", str(project)]) == 0
    assert main(["snapshot", str(project)]) == 0

    # Doctor check on valid project
    assert main(["doctor", str(project)]) == 0
    assert main(["doctor", str(project), "--json"]) == 0

    # Explain commands
    assert main(["explain", "workforce", "--path", str(project)]) == 0
    assert main(["explain", "agent", "implementer"]) == 0
    assert main(["explain", "skill", "clean-code"]) == 0
    assert main(["explain", "recipe", "feature-standard"]) == 0
    assert main(["explain", "context", "T-001", "--path", str(project)]) == 0
    assert main(["explain", "model", "implementer", "--path", str(project)]) == 0


def test_doctor_fails_on_inconsistent_project(tmp_path: Path):
    # Gate E: doctor must fail on an inconsistent project
    bad_project = tmp_path / "bad-project"
    bad_project.mkdir()

    # Broken atlas.json with missing fields and bad version
    (bad_project / "atlas.json").write_text(json.dumps({"version": 1}))

    # Inconsistent goal with broken dependency
    goals_dir = bad_project / ".ai/goals"
    goals_dir.mkdir(parents=True)
    (goals_dir / "G1.goal.json").write_text(json.dumps({
        "id": "G1",
        "title": "Goal 1",
        "phase": "P01",
        "state": "PLANNED",
        "objective": "Broken",
        "acceptance": ["A"],
        "gates": {"tests": "required"},
        "dependencies": ["NON_EXISTENT_GOAL"],
        "evidence": []
    }))

    # Doctor should report errors and return exit code 1
    assert main(["doctor", str(bad_project)]) == 1
