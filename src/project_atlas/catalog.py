from __future__ import annotations

from functools import lru_cache
from typing import Any

from .io import load_yaml
from .resources import resource_path


@lru_cache(maxsize=None)
def load_catalog(name: str) -> dict[str, Any]:
    return load_yaml(resource_path("catalog", f"{name}.yaml"))


def by_id(name: str) -> dict[str, dict[str, Any]]:
    raw = load_catalog(name)
    key = name if name in raw else next(iter(raw), name)
    entries = raw.get(key, [])
    return {item["id"]: item for item in entries}
