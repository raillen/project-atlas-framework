from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from jsonschema import Draft202012Validator

from .io import load_yaml
from .resources import resource_path
from .catalog import by_id, load_catalog


REQUIRED_PROJECT_FILES = [
    "docs/ATLAS.md",
    "PROJECT_STATE.md",
    "PROJECT_MANIFEST.yaml",
    ".atlas/project-profile.yaml",
    ".ai/orchestration/model-policy.yaml",
    ".ai/agents/manifest.yaml",
    ".ai/skills/manifest.yaml",
    ".ai/recipes/manifest.yaml",
]


def load_schema(repo_root: Path, name: str) -> dict[str, Any]:
    bundled = Path(__file__).resolve().parents[2] / "schemas" / name
    if bundled.exists():
        return json.loads(bundled.read_text(encoding="utf-8"))
    # Installed wheel fallback: schema copies are embedded next to generated projects only.
    raise FileNotFoundError(f"Schema not found: {name}")


def validate_against_schema(data: dict[str, Any], schema: dict[str, Any]) -> list[str]:
    errors = Draft202012Validator(schema).iter_errors(data)
    return [f"{'.'.join(map(str, e.absolute_path)) or '<root>'}: {e.message}" for e in errors]


def validate_project(root: Path, schema_dir: Path | None = None) -> list[str]:
    errors: list[str] = []
    for rel in REQUIRED_PROJECT_FILES:
        if not (root / rel).exists():
            errors.append(f"missing required file: {rel}")

    schema_dir = schema_dir or resource_path("schemas")
    checks = [
        (root / ".atlas/project-profile.yaml", schema_dir / "project-profile.schema.json"),
        (root / "PROJECT_MANIFEST.yaml", schema_dir / "project-manifest.schema.json"),
        (root / ".ai/orchestration/model-policy.yaml", schema_dir / "model-policy.schema.json"),
    ]
    for data_path, schema_path in checks:
        if data_path.exists() and schema_path.exists():
            data = load_yaml(data_path)
            schema = json.loads(schema_path.read_text(encoding="utf-8"))
            for error in validate_against_schema(data, schema):
                errors.append(f"{data_path.relative_to(root)}: {error}")

    goal_schema = schema_dir / "goal.schema.json"
    if goal_schema.exists():
        schema = json.loads(goal_schema.read_text(encoding="utf-8"))
        for path in sorted((root / ".ai/goals").glob("**/*.goal.yaml")) if (root / ".ai/goals").exists() else []:
            for error in validate_against_schema(load_yaml(path), schema):
                errors.append(f"{path.relative_to(root)}: {error}")
    return errors


def validate_framework() -> list[str]:
    errors: list[str] = []
    skills = by_id("skills")
    agents = by_id("agents")
    recipes = by_id("recipes")

    for agent_id, agent in agents.items():
        for skill in agent.get("requires_skills_any", []):
            if skill not in skills:
                errors.append(f"agent {agent_id} references unknown skill {skill}")
    for recipe_id, recipe in recipes.items():
        for skill in recipe.get("requires_skills_any", []):
            if skill not in skills:
                errors.append(f"recipe {recipe_id} references unknown skill {skill}")
    for skill_id, skill in skills.items():
        for dep in skill.get("requires", []):
            if dep not in skills:
                errors.append(f"skill {skill_id} references unknown dependency {dep}")

    for bundle in load_catalog("bundles").get("bundles", []):
        for agent in bundle.get("agents", []):
            if agent not in agents:
                errors.append(f"bundle {bundle['id']} references unknown agent {agent}")
        for skill in bundle.get("skills", []):
            if skill not in skills:
                errors.append(f"bundle {bundle['id']} references unknown skill {skill}")
        for recipe in bundle.get("recipes", []):
            if recipe not in recipes:
                errors.append(f"bundle {bundle['id']} references unknown recipe {recipe}")

    for adapter in ["generic", "chatgpt", "claude", "kimi", "codex", "claude-code", "traycer"]:
        if not resource_path("adapters", f"{adapter}.md").exists():
            errors.append(f"missing adapter: {adapter}")
    return errors
