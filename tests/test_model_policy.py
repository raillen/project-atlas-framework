from pathlib import Path

from project_atlas.model_policy import build_model_policy
from project_atlas.profile import load_profile


def test_model_policy_uses_only_project_roster():
    profile = load_profile(Path("examples/brasa/project-profile.yaml"))
    policy = build_model_policy(profile)
    roster = {m["id"] for m in policy["roster"]}
    assert "openai/gpt-5.6-sol" in roster
    for role in policy["roles"].values():
        assert set(role["preferred"]).issubset(roster)
        assert set(role["fallback"]).issubset(roster)
