from __future__ import annotations

from importlib.resources import files
from pathlib import Path


def resource_path(*parts: str) -> Path:
    return Path(str(files("project_atlas").joinpath("resources", *parts)))
