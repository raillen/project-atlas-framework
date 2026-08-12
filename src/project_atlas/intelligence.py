from __future__ import annotations

from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from .io import dump_json, load_json

DEFAULT_PATH = Path(".atlas/history/project-intelligence.json")


def _store_path(root: Path) -> Path:
    atlas = load_json(root / "atlas.json")
    configured = atlas.get("intelligence", {}).get("path")
    return root / (configured or DEFAULT_PATH)


def load_intelligence(root: Path) -> dict[str, Any]:
    path = _store_path(root)
    if not path.exists():
        return {
            "version": 1,
            "updated_at": datetime.now(timezone.utc).isoformat(),
            "summary": {},
            "tasks": [],
            "debt": [],
        }
    return load_json(path)


def _num(report: dict[str, Any], *path: str) -> float:
    value: Any = report
    for key in path:
        if not isinstance(value, dict):
            return 0.0
        value = value.get(key)
    return float(value or 0)


def recompute_summary(data: dict[str, Any]) -> dict[str, Any]:
    tasks = list(data.get("tasks", []))
    summary = {
        "tasks": len(tasks),
        "input_tokens": int(sum(_num(t, "tokens", "input") for t in tasks)),
        "output_tokens": int(sum(_num(t, "tokens", "output") for t in tasks)),
        "cached_tokens": int(sum(_num(t, "tokens", "cached") for t in tasks)),
        "intermediate_output_tokens": int(sum(_num(t, "tokens", "intermediate_output") for t in tasks)),
        "direct_cost": round(sum(_num(t, "cost", "direct", "amount") for t in tasks), 6),
    }
    return summary


def record_task_report(root: Path, report: dict[str, Any]) -> dict[str, Any]:
    if not report.get("id"):
        raise ValueError("Task report requires id")
    data = load_intelligence(root)
    tasks = [t for t in data.get("tasks", []) if t.get("id") != report["id"]]
    tasks.append(report)
    data["tasks"] = tasks
    data["summary"] = recompute_summary(data)
    data["updated_at"] = datetime.now(timezone.utc).isoformat()
    path = _store_path(root)
    dump_json(data, path)
    return data
