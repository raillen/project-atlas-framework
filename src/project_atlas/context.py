from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

from .io import load_json


@dataclass(frozen=True)
class ContextPlan:
    profile: str
    strategy: str
    budget: dict[str, Any]
    reasons: list[str]


SMALL_HINTS = {"rename", "typo", "label", "copy", "text", "spelling"}
LARGE_HINTS = {"architecture", "migration", "security", "refactor", "rewrite", "redesign"}
PROGRESSIVE_HINTS = {"bug", "regression", "unknown", "investigate", "debug", "failure", "crash"}


def _pick_profile(task: str) -> tuple[str, str, list[str]]:
    tokens = set(task.lower().replace("/", " ").replace("-", " ").split())
    if tokens & SMALL_HINTS:
        return "small", "direct", ["localized-task-hint"]
    if tokens & LARGE_HINTS:
        return "large", "progressive-retrieval", ["high-impact-task-hint"]
    if tokens & PROGRESSIVE_HINTS:
        return "medium", "progressive-retrieval", ["uncertain-task-hint"]
    return "medium", "context-pack-or-structural-retrieval", ["default"]


def plan_context(root: Path, task: str) -> ContextPlan:
    atlas = load_json(root / "atlas.json")
    context = atlas.get("context", {})
    profile, strategy, reasons = _pick_profile(task)
    profiles = context.get("profiles", {})
    budget = dict(profiles.get(profile, profiles.get(context.get("budget_profile", "medium"), {})))
    if not budget:
        raise ValueError("atlas.json does not define usable context budget profiles")
    return ContextPlan(profile=profile, strategy=strategy, budget=budget, reasons=reasons)
