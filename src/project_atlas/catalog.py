from __future__ import annotations

from functools import lru_cache
from typing import Any

from .io import load_json
from .resources import resource_path


@lru_cache(maxsize=None)
def load_registry() -> dict[str, Any]:
    manifest = load_json(resource_path("catalog", "catalog.json"))
    registry: dict[str, Any] = {"version": manifest.get("version", 2)}
    for name, filename in manifest.get("sections", {}).items():
        section = load_json(resource_path("catalog", str(filename)))
        registry[name] = section.get(name, [])
    return registry


def load_catalog(name: str) -> dict[str, Any]:
    data = load_registry()
    if name in data:
        return {name: data[name], "risk_rules": data.get("risk_rules", [])}
    return data


def by_id(name: str) -> dict[str, dict[str, Any]]:
    entries = load_registry().get(name, [])
    return {item["id"]: item for item in entries}
