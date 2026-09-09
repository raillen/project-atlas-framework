from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

from jsonschema import Draft202012Validator
from referencing import Registry, Resource
from referencing.jsonschema import DRAFT202012

from .io import load_json
from .resources import resource_path


@dataclass
class SkillManifest:
    id: str
    name: str
    purpose: str
    raw: dict[str, Any] = field(default_factory=dict)


@dataclass
class SkillPackage:
    manifest: SkillManifest
    instructions: str
    checks: list[Path] = field(default_factory=list)
    templates: list[Path] = field(default_factory=list)
    references: list[Path] = field(default_factory=list)
    scripts: list[Path] = field(default_factory=list)
    examples: list[Path] = field(default_factory=list)


@dataclass
class AgentManifest:
    id: str
    name: str
    purpose: str
    raw: dict[str, Any] = field(default_factory=dict)


@dataclass
class AgentPackage:
    manifest: AgentManifest
    instructions: str


@dataclass
class RecipeManifest:
    id: str
    purpose: str
    raw: dict[str, Any] = field(default_factory=dict)


@dataclass
class RecipePackage:
    manifest: RecipeManifest
    instructions: str


def get_workforce_root() -> Path:
    return resource_path("workforce")


def get_schemas_root() -> Path:
    return resource_path("schemas")


def build_workforce_registry(schemas_dir: Path | None = None) -> Registry:
    s_dir = schemas_dir or get_schemas_root()
    registry = Registry()
    if s_dir.exists():
        for file_path in s_dir.glob("*.json"):
            try:
                content = json.loads(file_path.read_text(encoding="utf-8"))
                resource = Resource.from_contents(content, default_specification=DRAFT202012)
                registry = registry.with_resource(uri=file_path.name, resource=resource)
                if "$id" in content:
                    registry = registry.with_resource(uri=content["$id"], resource=resource)
            except Exception:
                pass
    return registry


def load_skill_package(skill_dir: Path) -> SkillPackage:
    manifest_path = skill_dir / "manifest.json"
    raw = load_json(manifest_path) if manifest_path.exists() else {}
    skill_md = skill_dir / "SKILL.md"
    instructions = skill_md.read_text(encoding="utf-8") if skill_md.exists() else ""
    return SkillPackage(
        manifest=SkillManifest(id=raw.get("id", skill_dir.name), name=raw.get("name", skill_dir.name), purpose=raw.get("purpose", ""), raw=raw),
        instructions=instructions,
        checks=list((skill_dir / "checks").glob("*")) if (skill_dir / "checks").exists() else [],
        templates=list((skill_dir / "templates").glob("*")) if (skill_dir / "templates").exists() else [],
        references=list((skill_dir / "references").glob("*")) if (skill_dir / "references").exists() else [],
        scripts=list((skill_dir / "scripts").glob("*")) if (skill_dir / "scripts").exists() else [],
        examples=list((skill_dir / "examples").glob("*")) if (skill_dir / "examples").exists() else [],
    )


def load_agent_package(agent_dir: Path) -> AgentPackage:
    manifest_path = agent_dir / "manifest.json"
    if not manifest_path.exists():
        manifest_path = agent_dir / "agent.json"
    raw = load_json(manifest_path) if manifest_path.exists() else {}
    agent_md = agent_dir / "AGENT.md"
    instructions = agent_md.read_text(encoding="utf-8") if agent_md.exists() else ""
    return AgentPackage(
        manifest=AgentManifest(id=raw.get("id", agent_dir.name), name=raw.get("name", agent_dir.name), purpose=raw.get("purpose", ""), raw=raw),
        instructions=instructions,
    )


def load_recipe_package(recipe_dir: Path) -> RecipePackage:
    manifest_path = recipe_dir / "recipe.json"
    if not manifest_path.exists():
        manifest_path = recipe_dir / "manifest.json"
    raw = load_json(manifest_path) if manifest_path.exists() else {}
    recipe_md = recipe_dir / "RECIPE.md"
    instructions = recipe_md.read_text(encoding="utf-8") if recipe_md.exists() else ""
    return RecipePackage(
        manifest=RecipeManifest(id=raw.get("id", recipe_dir.name), purpose=raw.get("purpose", ""), raw=raw),
        instructions=instructions,
    )


def load_all_skills(workforce_dir: Path | None = None) -> list[dict[str, Any]]:
    w_dir = (workforce_dir or get_workforce_root()) / "skills"
    skills = []
    if w_dir.exists():
        for skill_dir in sorted(w_dir.iterdir()):
            manifest = skill_dir / "manifest.json"
            if manifest.exists():
                data = load_json(manifest)
                skill_md = skill_dir / "SKILL.md"
                if skill_md.exists() and "instructions" not in data:
                    data["instructions"] = skill_md.read_text(encoding="utf-8")
                skills.append(data)
    return skills


def load_all_agents(workforce_dir: Path | None = None) -> list[dict[str, Any]]:
    w_dir = (workforce_dir or get_workforce_root()) / "agents"
    agents = []
    if w_dir.exists():
        for agent_dir in sorted(w_dir.iterdir()):
            manifest = agent_dir / "manifest.json"
            if not manifest.exists():
                manifest = agent_dir / "agent.json"
            if manifest.exists():
                data = load_json(manifest)
                agent_md = agent_dir / "AGENT.md"
                if agent_md.exists() and "instructions" not in data:
                    data["instructions"] = agent_md.read_text(encoding="utf-8")
                agents.append(data)
    return agents


def load_all_recipes(workforce_dir: Path | None = None) -> list[dict[str, Any]]:
    w_dir = (workforce_dir or get_workforce_root()) / "recipes"
    recipes = []
    if w_dir.exists():
        for recipe_dir in sorted(w_dir.iterdir()):
            manifest = recipe_dir / "recipe.json"
            if not manifest.exists():
                manifest = recipe_dir / "manifest.json"
            if manifest.exists():
                data = load_json(manifest)
                recipe_md = recipe_dir / "RECIPE.md"
                if recipe_md.exists() and "instructions" not in data:
                    data["instructions"] = recipe_md.read_text(encoding="utf-8")
                recipes.append(data)
    return recipes


def validate_workforce(
    workforce_dir: Path | None = None,
    schemas_dir: Path | None = None,
) -> list[str]:
    w_dir = workforce_dir or get_workforce_root()
    s_dir = schemas_dir or get_schemas_root()
    registry = build_workforce_registry(s_dir)
    errors: list[str] = []

    # Validate skills
    skill_schema_path = s_dir / "skill.schema.json"
    if skill_schema_path.exists():
        skill_schema = load_json(skill_schema_path)
        validator = Draft202012Validator(skill_schema, registry=registry)
        for skill_dir in sorted((w_dir / "skills").glob("*")):
            if skill_dir.is_dir():
                manifest = skill_dir / "manifest.json"
                if not manifest.exists():
                    errors.append(f"Skill package {skill_dir.name} is missing manifest.json")
                else:
                    data = load_json(manifest)
                    for err in validator.iter_errors(data):
                        errors.append(f"Skill {skill_dir.name}: {err.message}")
                if not (skill_dir / "SKILL.md").exists():
                    errors.append(f"Skill package {skill_dir.name} is missing SKILL.md")

    # Validate agents
    agent_schema_path = s_dir / "agent.schema.json"
    if agent_schema_path.exists():
        agent_schema = load_json(agent_schema_path)
        validator = Draft202012Validator(agent_schema, registry=registry)
        for agent_dir in sorted((w_dir / "agents").glob("*")):
            if agent_dir.is_dir():
                manifest = agent_dir / "manifest.json"
                if not manifest.exists():
                    manifest = agent_dir / "agent.json"
                if not manifest.exists():
                    errors.append(f"Agent package {agent_dir.name} is missing manifest.json")
                else:
                    data = load_json(manifest)
                    for err in validator.iter_errors(data):
                        errors.append(f"Agent {agent_dir.name}: {err.message}")

    # Validate recipes
    recipe_schema_path = s_dir / "recipe.schema.json"
    if recipe_schema_path.exists():
        recipe_schema = load_json(recipe_schema_path)
        validator = Draft202012Validator(recipe_schema, registry=registry)
        for recipe_dir in sorted((w_dir / "recipes").glob("*")):
            if recipe_dir.is_dir():
                manifest = recipe_dir / "recipe.json"
                if not manifest.exists():
                    manifest = recipe_dir / "manifest.json"
                if not manifest.exists():
                    errors.append(f"Recipe package {recipe_dir.name} is missing recipe.json")
                else:
                    data = load_json(manifest)
                    for err in validator.iter_errors(data):
                        errors.append(f"Recipe {recipe_dir.name}: {err.message}")

    return errors
