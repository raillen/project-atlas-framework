from __future__ import annotations

from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from .io import dump_yaml, write_text
from .model_policy import build_model_policy
from .profile import ProjectProfile
from .resolver import Resolution, resolve


def _manifest(kind: str, values: list[str], resolution: Resolution) -> dict[str, Any]:
    return {
        "generated_by": {"project_atlas": "0.1.0"},
        kind: values,
        "reasons": {value: resolution.reasons.get(value, []) for value in values},
    }


def initialize_project(root: Path, profile: ProjectProfile) -> Resolution:
    if not profile.preferred_models:
        raise ValueError(
            "Preferred LLMs/providers are required per project. Add ai.preferred_models to the profile."
        )
    root.mkdir(parents=True, exist_ok=True)
    resolution = resolve(profile)
    dump_yaml(profile.raw, root / ".atlas/project-profile.yaml")

    dump_yaml(_manifest("agents", resolution.agents, resolution), root / ".ai/agents/manifest.yaml")
    dump_yaml(_manifest("skills", resolution.skills, resolution), root / ".ai/skills/manifest.yaml")
    dump_yaml(_manifest("recipes", resolution.recipes, resolution), root / ".ai/recipes/manifest.yaml")

    model_policy = build_model_policy(profile)
    dump_yaml(model_policy, root / ".ai/orchestration/model-policy.yaml")
    dump_yaml(
        {
            "version": 1,
            "primary": profile.orchestrator,
            "protocol": "project-orchestration-protocol-v1",
            "autonomy": profile.autonomy,
            "source_of_truth": "repository",
        },
        root / ".ai/orchestration/orchestrator.yaml",
    )
    dump_yaml(
        {
            "version": 1,
            "rules": [
                "availability fallback",
                "quality fallback after two failed attempts",
                "cross-provider review for high-risk changes when possible",
                "human escalation after fallback exhaustion",
            ],
        },
        root / ".ai/orchestration/fallbacks.yaml",
    )
    dump_yaml(
        {"version": 1, "records": [], "note": "Append measured project-local model performance only."},
        root / ".ai/orchestration/model-scorecard.yaml",
    )

    project = profile.raw.get("project", {})
    project_name = project.get("name", root.name)
    now = datetime.now(timezone.utc).isoformat()
    dump_yaml(
        {
            "framework": {"name": "project-atlas-framework", "version": "0.1.0"},
            "project": {"name": project_name, "type": project.get("type", [])},
            "documentation": {"entrypoint": "docs/ATLAS.md"},
            "orchestration": {"protocol": "POP", "orchestrator": profile.orchestrator},
            "goals": {"active_phase": "P00", "active_goal": None},
            "ai": {"policy": ".ai/orchestration/model-policy.yaml"},
        },
        root / "PROJECT_MANIFEST.yaml",
    )
    write_text(
        root / "PROJECT_STATE.md",
        f"""# Current Project State

- Project: **{project_name}**
- Framework: **Project Atlas 0.1.0**
- Current phase: **P00 — Foundation**
- Current goal: **not selected**
- Last updated: `{now}`

## Next action

Define and lock the first measurable Goal before implementation begins.

## Recovery order

1. `ENTRYPOINT.md` or `AGENTS.md` for the active platform.
2. `PROJECT_MANIFEST.yaml`.
3. `PROJECT_STATE.md`.
4. `docs/ATLAS.md`.
5. Active Goal under `.ai/goals/`.
6. Relevant canonical docs and ADRs.
""",
    )
    write_text(
        root / "docs/ATLAS.md",
        f"""# Project Atlas — {project_name}

This is the navigation root for humans and agents.

## Current state

- [Project state](../PROJECT_STATE.md)
- [Project manifest](../PROJECT_MANIFEST.yaml)
- [Project profile](../.atlas/project-profile.yaml)

## AI workforce

- [Agents manifest](../.ai/agents/manifest.yaml)
- [Skills manifest](../.ai/skills/manifest.yaml)
- [Recipes manifest](../.ai/recipes/manifest.yaml)
- [Model policy](../.ai/orchestration/model-policy.yaml)
- [Orchestrator](../.ai/orchestration/orchestrator.yaml)

## Goals

Goals live under `.ai/goals/<phase>/` and are the authoritative definition of completion.

## Documentation contract

Add product, architecture, implementation, quality, security, governance and operations documents here as the project grows. Keep this index current.
""",
    )
    write_text(
        root / "ENTRYPOINT.md",
        """# Project Atlas entrypoint

1. Read `PROJECT_MANIFEST.yaml` and `PROJECT_STATE.md`.
2. Read `docs/ATLAS.md`.
3. Read the active Goal and its dependencies.
4. Load only the agents/skills/recipes selected in `.ai/` manifests.
5. Follow the Project Orchestration Protocol and quality gates.
6. Never weaken acceptance criteria silently.
7. Keep code, tests and documentation synchronized.
""",
    )
    (root / ".ai/goals/P00").mkdir(parents=True, exist_ok=True)
    return resolution
