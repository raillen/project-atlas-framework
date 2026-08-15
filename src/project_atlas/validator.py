from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from jsonschema import Draft202012Validator
from referencing import Registry, Resource
from referencing.jsonschema import DRAFT202012

from .catalog import by_id, load_registry
from .io import load_data, load_json
from .resources import resource_path

REQUIRED_V2_FILES = [
    "docs/ATLAS.md",
    "PROJECT_STATE.md",
    "atlas.json",
    ".ai/orchestration/model-policy.json",
    ".ai/agents/manifest.json",
    ".ai/skills/manifest.json",
    ".ai/recipes/manifest.json",
    ".atlas/history/project-intelligence.json",
]

REQUIRED_V1_FILES = [
    "docs/ATLAS.md",
    "PROJECT_STATE.md",
    "PROJECT_MANIFEST.yaml",
    ".atlas/project-profile.yaml",
    ".ai/orchestration/model-policy.yaml",
    ".ai/agents/manifest.yaml",
    ".ai/skills/manifest.yaml",
    ".ai/recipes/manifest.yaml",
]


def build_schema_registry(schema_dir: Path) -> Registry:
    registry = Registry()
    if schema_dir.exists():
        for file_path in schema_dir.glob("*.json"):
            try:
                content = json.loads(file_path.read_text(encoding="utf-8"))
                resource = Resource.from_contents(content, default_specification=DRAFT202012)
                registry = registry.with_resource(uri=file_path.name, resource=resource)
                if "$id" in content:
                    registry = registry.with_resource(uri=content["$id"], resource=resource)
            except Exception:
                pass
    return registry


def validate_against_schema(
    data: dict[str, Any],
    schema: dict[str, Any],
    registry: Registry | None = None,
) -> list[str]:
    errors = Draft202012Validator(schema, registry=registry).iter_errors(data)
    return [f"{'.'.join(map(str, e.absolute_path)) or '<root>'}: {e.message}" for e in errors]


def _schema_dir(schema_dir: Path | None) -> Path:
    return schema_dir or resource_path("schemas")


def _validate_v2(root: Path, schema_dir: Path) -> list[str]:
    errors: list[str] = []
    registry = build_schema_registry(schema_dir)
    for rel in REQUIRED_V2_FILES:
        if not (root / rel).exists():
            errors.append(f"missing required file: {rel}")

    checks = [
        (root / "atlas.json", schema_dir / "atlas.schema.json"),
        (root / ".ai/orchestration/model-policy.json", schema_dir / "model-policy.schema.json"),
    ]
    for data_path, schema_path in checks:
        if data_path.exists() and schema_path.exists():
            data = load_json(data_path)
            schema = json.loads(schema_path.read_text(encoding="utf-8"))
            for error in validate_against_schema(data, schema, registry=registry):
                errors.append(f"{data_path.relative_to(root)}: {error}")

    goal_schema = schema_dir / "goal.schema.json"
    if goal_schema.exists() and (root / ".ai/goals").exists():
        schema = json.loads(goal_schema.read_text(encoding="utf-8"))
        for path in sorted((root / ".ai/goals").glob("**/*.goal.json")):
            for error in validate_against_schema(load_json(path), schema, registry=registry):
                errors.append(f"{path.relative_to(root)}: {error}")


    atlas_yaml_candidates = [
        root / "PROJECT_MANIFEST.yaml",
        root / ".atlas/project-profile.yaml",
        root / ".ai/agents/manifest.yaml",
        root / ".ai/skills/manifest.yaml",
        root / ".ai/recipes/manifest.yaml",
        root / ".ai/orchestration/model-policy.yaml",
        root / ".ai/orchestration/orchestrator.yaml",
        root / ".ai/orchestration/fallbacks.yaml",
        root / ".ai/orchestration/model-scorecard.yaml",
    ]
    atlas_yaml_candidates.extend((root / ".ai/goals").glob("**/*.goal.yaml") if (root / ".ai/goals").exists() else [])
    generated_yaml = [p for p in atlas_yaml_candidates if p.exists()]
    if generated_yaml:
        errors.append(
            "v0.2 Atlas canonical state contains legacy YAML; migrate/remove: "
            + ", ".join(str(p.relative_to(root)) for p in generated_yaml[:10])
        )
    return errors


def _validate_v1(root: Path, schema_dir: Path) -> list[str]:
    errors: list[str] = []
    for rel in REQUIRED_V1_FILES:
        if not (root / rel).exists():
            errors.append(f"missing required legacy file: {rel}")
    checks = [
        (root / ".atlas/project-profile.yaml", schema_dir / "project-profile.schema.json"),
        (root / "PROJECT_MANIFEST.yaml", schema_dir / "project-manifest.schema.json"),
        (root / ".ai/orchestration/model-policy.yaml", schema_dir / "model-policy.schema.json"),
    ]
    for data_path, schema_path in checks:
        if data_path.exists() and schema_path.exists():
            schema = json.loads(schema_path.read_text(encoding="utf-8"))
            for error in validate_against_schema(load_data(data_path), schema):
                errors.append(f"{data_path.relative_to(root)}: {error}")
    return errors


def validate_project(root: Path, schema_dir: Path | None = None) -> list[str]:
    schemas = _schema_dir(schema_dir)
    if (root / "atlas.json").exists():
        return _validate_v2(root, schemas)
    if (root / "PROJECT_MANIFEST.yaml").exists() or (root / ".atlas/project-profile.yaml").exists():
        return _validate_v1(root, schemas)
    return ["not a recognized Project Atlas project: missing atlas.json"]


def validate_framework() -> list[str]:
    errors: list[str] = []
    skills = by_id("skills")
    agents = by_id("agents")
    recipes = by_id("recipes")
    registry = load_registry()

    for key in ("agents", "skills", "recipes", "bundles"):
        ids = [item.get("id") for item in registry.get(key, [])]
        if len(ids) != len(set(ids)):
            errors.append(f"duplicate IDs in catalog section: {key}")

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

    for bundle in registry.get("bundles", []):
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

    for schema in ["atlas.schema.json", "project-profile.schema.json", "goal.schema.json", "model-policy.schema.json", "task-report.schema.json"]:
        if not resource_path("schemas", schema).exists():
            errors.append(f"missing schema: {schema}")
    return errors
