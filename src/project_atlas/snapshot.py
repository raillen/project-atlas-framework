from __future__ import annotations

from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile

INCLUDE = [
    "ENTRYPOINT.md",
    "PROJECT_MANIFEST.yaml",
    "PROJECT_STATE.md",
    "docs",
    ".atlas",
    ".ai",
]


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
                    if child.is_file():
                        archive.write(child, child.relative_to(root))
    return output
