from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any

from .catalog import by_id, load_catalog
from .profile import ProjectProfile, any_match


@dataclass
class Resolution:
    agents: list[str]
    skills: list[str]
    recipes: list[str]
    reasons: dict[str, list[str]]
    traces: dict[str, list[dict[str, Any]]] = field(default_factory=dict)


def _selector_matches(item: dict[str, Any], profile: ProjectProfile) -> tuple[bool, list[str], list[dict[str, Any]]]:
    selector = item.get("select", {})
    reasons: list[str] = []
    traces: list[dict[str, Any]] = []
    if item.get("core"):
        reasons.append("core")
        traces.append({"type": "core", "value": "core"})
        return True, reasons, traces

    checks: list[bool] = []
    if selector.get("project_types"):
        ok = any_match(selector["project_types"], profile.project_types)
        checks.append(ok)
        if ok:
            reasons.append("project-type")
            traces.append({"type": "project-type", "value": list(selector["project_types"])})
    if selector.get("stack_any"):
        ok = any_match(selector["stack_any"], profile.stack_tokens)
        checks.append(ok)
        if ok:
            reasons.append("stack")
            traces.append({"type": "stack", "value": list(selector["stack_any"])})
    if selector.get("features_any"):
        ok = any_match(selector["features_any"], profile.features)
        checks.append(ok)
        if ok:
            reasons.append("feature")
            traces.append({"type": "feature", "value": list(selector["features_any"])})
    if selector.get("risk_any"):
        ok = any_match(selector["risk_any"], profile.risks)
        checks.append(ok)
        if ok:
            reasons.append("risk")
            traces.append({"type": "risk", "value": list(selector["risk_any"])})
    matched = bool(checks) and all(checks)
    return matched, reasons if matched else [], traces if matched else []


def resolve(profile: ProjectProfile) -> Resolution:
    skills_catalog = by_id("skills")
    agents_catalog = by_id("agents")
    recipes_catalog = by_id("recipes")
    bundles = load_catalog("bundles").get("bundles", [])

    selected_skills: set[str] = set()
    selected_agents: set[str] = set()
    selected_recipes: set[str] = set()
    reasons: dict[str, list[str]] = {}
    traces: dict[str, list[dict[str, Any]]] = {}

    for bundle in bundles:
        if str(bundle.get("project_type", "")).lower() in profile.project_types:
            bundle_trace = {"type": "bundle", "value": bundle["id"]}
            for skill in bundle.get("skills", []):
                selected_skills.add(skill)
                reasons.setdefault(skill, []).append(f"bundle:{bundle['id']}")
                traces.setdefault(skill, []).append(bundle_trace)
            for agent in bundle.get("agents", []):
                selected_agents.add(agent)
                reasons.setdefault(agent, []).append(f"bundle:{bundle['id']}")
                traces.setdefault(agent, []).append(bundle_trace)
            for recipe in bundle.get("recipes", []):
                selected_recipes.add(recipe)
                reasons.setdefault(recipe, []).append(f"bundle:{bundle['id']}")
                traces.setdefault(recipe, []).append(bundle_trace)

    for skill_id, item in skills_catalog.items():
        matched, why, trace_items = _selector_matches(item, profile)
        if matched:
            selected_skills.add(skill_id)
            reasons.setdefault(skill_id, []).extend(why)
            traces.setdefault(skill_id, []).extend(trace_items)

    risk_rules = load_catalog("bundles").get("risk_rules", [])
    for rule in risk_rules:
        trigger = set(map(str.lower, rule.get("when_any", [])))
        universe = profile.risks | profile.features | profile.project_types
        if trigger & universe:
            rule_trace = {"type": "risk-rule", "value": rule["id"]}
            for skill in rule.get("require_skills", []):
                selected_skills.add(skill)
                reasons.setdefault(skill, []).append(f"risk-rule:{rule['id']}")
                traces.setdefault(skill, []).append(rule_trace)

    for agent_id, item in agents_catalog.items():
        matched, why, trace_items = _selector_matches(item, profile)
        required_any = set(item.get("requires_skills_any", []))
        if matched or (required_any and required_any & selected_skills):
            selected_agents.add(agent_id)
            reasons.setdefault(agent_id, []).extend(why or ["skill-match"])
            traces.setdefault(agent_id, []).extend(trace_items or [{"type": "skill-match", "value": list(required_any & selected_skills)}])

    for recipe_id, item in recipes_catalog.items():
        matched, why, trace_items = _selector_matches(item, profile)
        required_any = set(item.get("requires_skills_any", []))
        if matched or (required_any and required_any & selected_skills):
            selected_recipes.add(recipe_id)
            reasons.setdefault(recipe_id, []).extend(why or ["skill-match"])
            traces.setdefault(recipe_id, []).extend(trace_items or [{"type": "skill-match", "value": list(required_any & selected_skills)}])

    changed = True
    while changed:
        changed = False
        for skill_id in list(selected_skills):
            for dep in skills_catalog.get(skill_id, {}).get("requires", []):
                if dep not in selected_skills:
                    selected_skills.add(dep)
                    reasons.setdefault(dep, []).append(f"dependency:{skill_id}")
                    traces.setdefault(dep, []).append({"type": "dependency", "value": skill_id})
                    changed = True

    return Resolution(
        agents=sorted(selected_agents),
        skills=sorted(selected_skills),
        recipes=sorted(selected_recipes),
        reasons={key: sorted(set(value)) for key, value in reasons.items()},
        traces=traces,
    )
