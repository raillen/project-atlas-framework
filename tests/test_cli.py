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
