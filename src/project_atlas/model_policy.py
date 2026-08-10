from __future__ import annotations

from typing import Any

from .profile import ProjectProfile

ROLE_GROUPS = {
    "architecture": ["architect", "goal-planner", "final-arbiter"],
    "implementation": ["implementer", "debugger"],
    "quality": ["tester", "reviewer", "security-reviewer", "release-verifier"],
    "documentation": ["documentation-maintainer"],
    "visual": ["ux-reviewer", "visual-reviewer"],
}


def build_model_policy(profile: ProjectProfile) -> dict[str, Any]:
    roster = profile.preferred_models
    if not roster:
        raise ValueError(
            "No preferred LLM roster configured. Project Atlas requires model preferences "
            "to be explicitly selected for every new project."
        )

    ids = [item["id"] if isinstance(item, dict) else str(item) for item in roster]
    providers = {
        (item["id"] if isinstance(item, dict) else str(item)): (
            item.get("provider", "unknown") if isinstance(item, dict) else "unknown"
        )
        for item in roster
    }
    explicit = profile.raw.get("ai", {}).get("role_preferences", {})
    roles: dict[str, Any] = {}
    all_roles = sorted({role for values in ROLE_GROUPS.values() for role in values})
    for role in all_roles:
        preferred = explicit.get(role, ids)
        if isinstance(preferred, str):
            preferred = [preferred]
        preferred = [model for model in preferred if model in ids]
        if not preferred:
            preferred = ids.copy()
        roles[role] = {
            "preferred": preferred,
            "fallback": [model for model in ids if model not in preferred],
        }

    return {
        "version": 1,
        "selection_rule": "cheapest-reliable-model-that-passes-gates",
        "cross_provider_review": True,
        "roster": [
            {"id": model_id, "provider": providers[model_id], "enabled": True}
            for model_id in ids
        ],
        "roles": roles,
    }
