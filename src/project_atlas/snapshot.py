from __future__ import annotations

from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile

INCLUDE = [
    "ENTRYPOINT.md",
    "atlas.json",
    "PROJECT_STATE.md",
    "docs",
    ".ai",
    ".atlas/history",
]

EXCLUDED_PARTS = {"runtime", "cache", "__pycache__"}


def create_snapshot(root: Path, output: Path) -> Path:
    output.parent.mkdir(parents=True, exist_ok=True)
    with ZipFile(output, "w", compression=ZIP_DEFLATED) as archive:
        for rel in INCLUDE:
            path = root / rel
            if not path.exists():
                continue
            if path.is_file():
                archive.write(path, path.relative_to(root))
            else:
                for child in path.rglob("*"):
                    if child.is_file() and not (set(child.relative_to(root).parts) & EXCLUDED_PARTS):
                        archive.write(child, child.relative_to(root))
    return output
