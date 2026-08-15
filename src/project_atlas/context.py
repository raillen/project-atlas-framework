from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Optional

from .io import load_json


@dataclass(frozen=True)
class ContextItem:
    id: str
    kind: str
    source: str
    locator: str
    range: Optional[dict[str, Any]] = None
    symbol: Optional[str] = None
    reason: Optional[str] = None
    priority: Optional[str] = None
    estimated_tokens: Optional[int] = None
    provenance: Optional[Any] = None
    trust: Optional[str] = None
    freshness: Optional[str] = None
    digest: Optional[str] = None


@dataclass(frozen=True)
class ContextBudget:
    max_input_tokens: Optional[int] = None
    context_target_tokens: Optional[int] = None
    context_hard_tokens: Optional[int] = None
    output_target_tokens: Optional[int] = None
    output_hard_tokens: Optional[int] = None
    max_expansion_rounds: Optional[int] = None
    max_retrieved_tokens: Optional[int] = None
    max_injected_tokens: Optional[int] = None
    max_intermediate_output_tokens: Optional[int] = None
    max_expansions: Optional[int] = None
    max_delegations: Optional[int] = None
    max_delegation_depth: Optional[int] = None
    deadline: Optional[str] = None
    cost_budget: Optional[Any] = None
    stop_reasons: Optional[list[str]] = None


@dataclass(frozen=True)
class ContextRequest:
    task_id: str
    intent: str
    required_topics: Optional[list[str]] = None
    required_symbols: Optional[list[str]] = None
    budget: Optional[ContextBudget] = None


@dataclass(frozen=True)
class ContextPack:
    id: str
    task_id: str
    created_at: str
    items: list[ContextItem] = field(default_factory=list)
    estimated_tokens: Optional[int] = None
    actual_tokens: Optional[int] = None
    reason: Optional[str] = None


@dataclass(frozen=True)
class ContextPlan:
    profile: str
    strategy: str
    budget: ContextBudget
    reasons: list[str]
    task: str = ""


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
    budget_dict = dict(profiles.get(profile, profiles.get(context.get("budget_profile", "medium"), {})))
    if not budget_dict:
        raise ValueError("atlas.json does not define usable context budget profiles")
    budget = ContextBudget(**budget_dict)
    return ContextPlan(profile=profile, strategy=strategy, budget=budget, reasons=reasons, task=task)
