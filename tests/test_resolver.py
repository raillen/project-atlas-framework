from pathlib import Path

from project_atlas.profile import load_profile
from project_atlas.resolver import resolve


def test_brasa_resolves_domain_workforce():
    profile = load_profile(Path("examples/brasa/project-profile.json"))
    result = resolve(profile)
    assert "architect" in result.agents
    assert "engine-engineer" in result.agents
    assert "renderer-engineer" in result.agents
    assert "networking-engineer" in result.agents
    assert "rendering-2d" in result.skills
    assert "multiplayer-networking" in result.skills
    assert "security-network" in result.skills
    assert "lean-progressive-context" in result.skills
    assert "project-intelligence" in result.skills
    assert "roslyn-context-indexing" in result.skills
    assert "engine-renderer" in result.recipes


def test_framework_catalog_references_are_valid():
    from project_atlas.validator import validate_framework
    assert validate_framework() == []
