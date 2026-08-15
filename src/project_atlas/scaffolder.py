from __future__ import annotations

from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from .io import dump_json, write_text
from .model_policy import build_model_policy
from .profile import ProjectProfile
from .resolver import Resolution, resolve

FRAMEWORK_VERSION = "0.2.0"


def _manifest(kind: str, values: list[str], resolution: Resolution) -> dict[str, Any]:
    return {
        "generated_by": {"project_atlas": FRAMEWORK_VERSION},
        kind: values,
        "reasons": {value: resolution.reasons.get(value, []) for value in values},
    }


def _context_policy() -> dict[str, Any]:
    return {
        "methodology": "lean-progressive-context",
        "architecture": "progressive-context-architecture",
        "mode": "progressive",
        "budget_profile": "medium",
        "profiles": {
            "small": {
                "context_target_tokens": 3000,
                "context_hard_tokens": 6000,
                "output_target_tokens": 500,
                "output_hard_tokens": 1000,
                "max_expansion_rounds": 1,
                "max_delegation_depth": 0,
            },
            "medium": {
                "context_target_tokens": 8000,
                "context_hard_tokens": 16000,
                "output_target_tokens": 1500,
                "output_hard_tokens": 3000,
                "max_expansion_rounds": 2,
                "max_delegation_depth": 1,
            },
            "large": {
                "context_target_tokens": 16000,
                "context_hard_tokens": 32000,
                "output_target_tokens": 3000,
                "output_hard_tokens": 6000,
                "max_expansion_rounds": 3,
                "max_delegation_depth": 1,
            },
        },
        "deep_recursion": {"enabled": False, "experimental": True},
        "runtime": {
            "database": ".atlas/runtime/atlas.db",
            "completed_context_ttl": "0d",
            "failed_context_ttl": "7d",
        },
    }


def _ensure_gitignore(root: Path) -> None:
    path = root / ".gitignore"
    current = path.read_text(encoding="utf-8") if path.exists() else ""
    required = [".atlas/runtime/", ".atlas/cache/"]
    missing = [line for line in required if line not in current.splitlines()]
    if not missing:
        return
    addition = "\n# Project Atlas derived/runtime state\n" + "\n".join(missing) + "\n"
    path.write_text(current.rstrip() + addition, encoding="utf-8")


def initialize_project(root: Path, profile: ProjectProfile) -> Resolution:
    if not profile.preferred_models:
        raise ValueError(
            "Preferred LLMs/providers are required per project. Add ai.preferred_models to the profile."
        )

    root.mkdir(parents=True, exist_ok=True)
    resolution = resolve(profile)
    project = profile.raw.get("project", {})
    project_name = project.get("name", root.name)
    now = datetime.now(timezone.utc).isoformat()

    atlas_config = {
        "version": 2,
        "protocol": {"version": 3, "compatible": ">=3 <4"},
        "framework": {"name": "project-atlas-framework", "version": FRAMEWORK_VERSION},
        "project": project,
        "stack": profile.raw.get("stack", {}),
        "features": profile.raw.get("features", []),
        "risk": profile.raw.get("risk", []),
        "quality": profile.raw.get("quality", {}),
        "documentation": {
            "entrypoint": "docs/ATLAS.md",
            "canonical_format": "markdown",
            "site": {
                "enabled": True,
                "source": "docs",
                "generated": True,
                "public_internal_views": True,
            },
            "audiences": ["user", "developer", "operations", "agent"],
            "virtual_chunking": True,
        },
        "context": _context_policy(),
        "intelligence": {
            "enabled": True,
            "path": ".atlas/history/project-intelligence.json",
            "task_reports": True,
            "track_input_tokens": True,
            "track_output_tokens": True,
            "track_cost": True,
            "distinguish_observed_estimated": True,
        },
        "orchestration": {
            "protocol": "POP",
            "orchestrator": profile.orchestrator,
            "autonomy": profile.autonomy,
        },
        "goals": {"active_phase": "P00", "active_goal": None},
        "ai": profile.raw.get("ai", {}),
    }
    dump_json(atlas_config, root / "atlas.json")

    dump_json(_manifest("agents", resolution.agents, resolution), root / ".ai/agents/manifest.json")
    dump_json(_manifest("skills", resolution.skills, resolution), root / ".ai/skills/manifest.json")
    dump_json(_manifest("recipes", resolution.recipes, resolution), root / ".ai/recipes/manifest.json")

    model_policy = build_model_policy(profile)
    dump_json(model_policy, root / ".ai/orchestration/model-policy.json")
    dump_json(
        {
            "version": 2,
            "primary": profile.orchestrator,
            "protocol": "project-orchestration-protocol-v2",
            "autonomy": profile.autonomy,
            "source_of_truth": "repository",
            "context_methodology": "LPC",
        },
        root / ".ai/orchestration/orchestrator.json",
    )
    dump_json(
        {
            "version": 2,
            "rules": [
                "availability fallback",
                "quality fallback after bounded failed attempts",
                "targeted context expansion before broad re-read",
                "cross-provider review for high-risk changes when possible",
                "human escalation after fallback/budget exhaustion",
            ],
            "unbounded_retry": False,
            "unbounded_recursion": False,
        },
        root / ".ai/orchestration/fallbacks.json",
    )
    dump_json(
        {"version": 2, "records": [], "note": "Append measured project-local model performance only."},
        root / ".ai/orchestration/model-scorecard.json",
    )

    dump_json(
        {
            "version": 1,
            "updated_at": now,
            "summary": {
                "tasks": 0,
                "input_tokens": 0,
                "output_tokens": 0,
                "cached_tokens": 0,
                "direct_cost": 0.0,
            },
            "tasks": [],
            "debt": [],
        },
        root / ".atlas/history/project-intelligence.json",
    )

    write_text(
        root / "PROJECT_STATE.md",
        f"""# Current Project State

- Project: **{project_name}**
- Framework: **Project Atlas {FRAMEWORK_VERSION}**
- Current phase: **P00 — Foundation**
- Current goal: **not selected**
- Context methodology: **Lean Progressive Context (LPC)**
- Last updated: `{now}`

## Next action

Define and lock the first measurable Goal before implementation begins.

## Recovery order

1. `ENTRYPOINT.md` or the platform adapter.
2. `atlas.json`.
3. `PROJECT_STATE.md`.
4. `docs/ATLAS.md`.
5. Active Goal under `.ai/goals/`.
6. Only relevant canonical docs/symbols/tests selected by the context strategy.

Do not load the entire repository by default.
""",
    )

    write_text(
        root / "docs/ATLAS.md",
        f"""# Project Atlas — {project_name}

This is the intent router for humans and agents. Add links as stable documentation is created; do not create empty documentation solely to populate this map.

## Current state

- [Project state](../PROJECT_STATE.md)
- `atlas.json` — canonical project configuration

## I want to use the product

Add user tutorials, how-to guides, reference and explanations under `docs/user/` as needed.

## I want to develop/contribute

Add onboarding, codebase tour, build/test/debug and task-oriented development guides under `docs/developer/`.

## I want to operate/support it

Add deployment, configuration, observability, runbooks, backup/recovery, troubleshooting and release guidance under `docs/operations/` / `docs/support/` as needed.

## I am an AI agent

1. Read the active Goal.
2. Use the smallest sufficient context.
3. Prefer structural/symbol/document-section pointers.
4. Expand only when evidence is insufficient.
5. Keep output bounded.
6. Update only impacted canonical docs.
7. Record evidence/intelligence and garbage-collect temporary context.

## Architecture / decisions / specs

Add stable architecture, ADR/RFC and specifications as the project grows.

## Goals

Goals live under `.ai/goals/<phase>/` and define measurable completion.

## Durable intelligence

Compact project/task intelligence lives in `.atlas/history/project-intelligence.json`.
""",
    )

    write_text(
        root / "ENTRYPOINT.md",
        """# Project Atlas entrypoint

1. Read `atlas.json`, `PROJECT_STATE.md` and `docs/ATLAS.md`.
2. Read the active Goal and its dependencies.
3. Start with the minimum sufficient context; do not read the entire repository.
4. Prefer Context Packs/task maps, document sections, symbols and related tests.
5. Expand context only when evidence is insufficient; delegation depth is bounded by `atlas.json`.
6. Never weaken acceptance criteria silently.
7. Keep code, tests and canonical docs synchronized through a Documentation Delta.
8. Keep intermediate output compact and do not persist task-specific context files.
9. Before completion, record evidence/project intelligence and remove temporary context.
""",
    )

    (root / ".ai/goals/P00").mkdir(parents=True, exist_ok=True)
    (root / ".atlas/runtime").mkdir(parents=True, exist_ok=True)
    _ensure_gitignore(root)
    return resolution
