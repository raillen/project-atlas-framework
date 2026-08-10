from pathlib import Path

from project_atlas.cli import main


def test_cli_smoke(tmp_path: Path):
    project = tmp_path / "project"
    assert main(["init", str(project), "--profile", "examples/brasa/project-profile.yaml", "--non-interactive"]) == 0
    assert main(["goal", "new", "P00-G01", "Foundation", "--phase", "P00", "--path", str(project)]) == 0
    assert main(["goal", "list", "--path", str(project)]) == 0
    assert main(["compile", "--target", "generic", "--path", str(project)]) == 0
    assert main(["snapshot", str(project)]) == 0
