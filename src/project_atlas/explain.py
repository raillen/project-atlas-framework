from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from .catalog import by_id, load_catalog
from .io import load_json
from .profile import ProjectProfile, load_profile
from .resolver import resolve


def explain_workforce(profile: ProjectProfile) -> dict[str, Any]:
    res = resolve(profile)
    return {
        "summary": {
            "agents_count": len(res.agents),
            "skills_count": len(res.skills),
            "recipes_count": len(res.recipes),
        },
        "agents": {
            agent: {
                "reasons": res.reasons.get(agent, []),
                "traces": res.traces.get(agent, [])
            }
            for agent in res.agents
        },
        "skills": {
            skill: {
                "reasons": res.reasons.get(skill, []),
                "traces": res.traces.get(skill, [])
            }
            for skill in res.skills
        },
        "recipes": {
            recipe: {
                "reasons": res.reasons.get(recipe, []),
                "traces": res.traces.get(recipe, [])
            }
            for recipe in res.recipes
        }
    }


def explain_skill(skill_id: str) -> dict[str, Any] | None:
    skills = by_id("skills")
    if skill_id not in skills:
        return None
    skill = skills[skill_id]
    return {
        "id": skill.get("id"),
        "name": skill.get("name"),
        "version": skill.get("version", 1),
        "purpose": skill.get("purpose"),
        "risk_level": skill.get("risk_level", "low"),
        "modes": skill.get("modes", []),
        "requires": skill.get("requires", []),
        "select": skill.get("select", {}),
        "instructions": skill.get("instructions", "").strip(),
        "provenance": skill.get("provenance", {"origin": "framework", "license": "MIT"})
    }


def explain_agent(agent_id: str) -> dict[str, Any] | None:
    agents = by_id("agents")
    if agent_id not in agents:
        return None
    agent = agents[agent_id]
    return {
        "id": agent.get("id"),
        "name": agent.get("name"),
        "version": agent.get("version", 2),
        "purpose": agent.get("purpose"),
        "risk_level": agent.get("risk_level", "low"),
        "required_skills": agent.get("required_skills", []),
        "requires_skills_any": agent.get("requires_skills_any", []),
        "inputs": agent.get("inputs", []),
        "outputs": agent.get("outputs", []),
        "permissions": agent.get("permissions", {}),
        "instructions": agent.get("instructions", "").strip()
    }


def explain_recipe(recipe_id: str) -> dict[str, Any] | None:
    recipes = by_id("recipes")
    if recipe_id not in recipes:
        return None
    recipe = recipes[recipe_id]
    return {
        "id": recipe.get("id"),
        "name": recipe.get("name"),
        "purpose": recipe.get("purpose"),
        "steps": recipe.get("steps", []),
        "instructions": recipe.get("instructions", "").strip()
    }


def explain_context(task_id: str, root: Path | None = None) -> dict[str, Any]:
    # Look for task / context pack in root
    items = []
    budget = {"max_input_tokens": 8000, "max_retrieved_tokens": 4000}
    if root:
        cpack_path = root / f".atlas/runtime/context/{task_id}.cpack.json"
        if cpack_path.exists():
            data = load_json(cpack_path)
            enhanced_items = []
            for item in data.get("items", []):
                if isinstance(item, dict):
                    enhanced_items.append({
                        "id": item.get("id"),
                        "why_loaded": item.get("reason"),
                        "source": item.get("source"),
                        "estimated_token_cost": item.get("estimated_tokens"),
                        "trust_level": item.get("trust")
                    })
                else:
                    enhanced_items.append(item)
            return {
                "task_id": task_id,
                "strategy": "progressive-retrieval",
                "estimated_tokens": data.get("estimated_tokens", 0),
                "items": enhanced_items,
                "budget": budget
            }
    return {
        "task_id": task_id,
        "strategy": "progressive-retrieval",
        "budget": budget,
        "notes": "Context is lazily expanded following Lean Progressive Context (LPC/PCA)."
    }


def explain_model(role: str, root: Path | None = None) -> dict[str, Any]:
    if root:
        policy_path = root / ".ai/orchestration/model-policy.json"
        if policy_path.exists():
            policy = load_json(policy_path)
            model_target = policy.get("roles", {}).get(role, "default")
            profiles = policy.get("profiles", {})
            profile_info = profiles.get(model_target) if isinstance(model_target, str) else None
            return {
                "role": role,
                "target": model_target,
                "profile": profile_info,
                "selection_rule": policy.get("selection_rule", "least_workforce"),
                "cross_provider_review": policy.get("cross_provider_review", True)
            }
    return {
        "role": role,
        "target": "default",
        "selection_rule": "least_workforce",
        "cross_provider_review": True
    }


def explain_execution(profile_id: str, root: Path) -> dict[str, Any]:
    policy_path = root / ".ai/orchestration/execution-policy.json"
    if policy_path.exists():
        policy = load_json(policy_path)
        backends = policy.get("backends", {})
        backend = backends.get(profile_id)
        if backend:
            return {
                "profile_id": profile_id,
                "backend": backend,
                "default_fallback": policy.get("default_fallback")
            }
    return {
        "profile_id": profile_id,
        "backend": None,
        "error": "Execution policy not found or profile missing."
    }
