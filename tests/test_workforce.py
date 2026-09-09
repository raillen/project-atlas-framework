from __future__ import annotations

import json
from pathlib import Path

from jsonschema import Draft202012Validator
from referencing import Registry, Resource
from referencing.jsonschema import DRAFT202012

from project_atlas.workforce import (
    load_agent_package,
    load_recipe_package,
    load_skill_package,
    validate_workforce,
)

WORKFORCE_ROOT = Path(__file__).parent.parent / "src/project_atlas/resources/workforce"
SCHEMA_DIR = Path(__file__).parent.parent / "schemas"


def get_schema_registry() -> Registry:
    registry = Registry()
    for file_path in SCHEMA_DIR.glob("*.json"):
        try:
            content = json.loads(file_path.read_text(encoding="utf-8"))
            resource = Resource.from_contents(content, default_specification=DRAFT202012)
            registry = registry.with_resource(uri=file_path.name, resource=resource)
            if "$id" in content:
                registry = registry.with_resource(uri=content["$id"], resource=resource)
        except Exception:
            pass
    return registry


_REGISTRY = get_schema_registry()


def validate_json_schema(schema_name: str, instance: dict) -> list[str]:
    schema = json.loads((SCHEMA_DIR / schema_name).read_text(encoding="utf-8"))
    validator = Draft202012Validator(schema, registry=_REGISTRY)
    return [e.message for e in validator.iter_errors(instance)]


def test_all_skill_packages_conformance():
    skills_dir = WORKFORCE_ROOT / "skills"
    assert skills_dir.exists()
    count = 0
    for sdir in sorted(skills_dir.glob("*")):
        if sdir.is_dir():
            count += 1
            pkg = load_skill_package(sdir)
            assert pkg.manifest.id == sdir.name
            assert pkg.instructions, f"Skill {sdir.name} missing instructions"
            errs = validate_json_schema("skill.schema.json", pkg.manifest.raw)
            assert errs == [], f"Skill {sdir.name} schema validation errors: {errs}"
    assert count >= 90


def test_all_agent_packages_conformance():
    agents_dir = WORKFORCE_ROOT / "agents"
    assert agents_dir.exists()
    count = 0
    for adir in sorted(agents_dir.glob("*")):
        if adir.is_dir():
            count += 1
            pkg = load_agent_package(adir)
            assert pkg.manifest.id == adir.name
            assert pkg.instructions, f"Agent {adir.name} missing instructions"
            errs = validate_json_schema("agent.schema.json", pkg.manifest.raw)
            assert errs == [], f"Agent {adir.name} schema validation errors: {errs}"
    assert count >= 20


def test_all_recipe_packages_conformance():
    recipes_dir = WORKFORCE_ROOT / "recipes"
    assert recipes_dir.exists()
    count = 0
    for rdir in sorted(recipes_dir.glob("*")):
        if rdir.is_dir():
            count += 1
            pkg = load_recipe_package(rdir)
            assert pkg.manifest.id == rdir.name
            assert pkg.instructions, f"Recipe {rdir.name} missing instructions"
            errs = validate_json_schema("recipe.schema.json", pkg.manifest.raw)
            assert errs == [], f"Recipe {rdir.name} schema validation errors: {errs}"
    assert count >= 10


def test_validate_workforce_clean():
    errors = validate_workforce()
    assert errors == [], f"Workforce validation errors: {errors}"
