from __future__ import annotations

from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from .io import dump_json, load_data

STATES = ["DRAFT", "PLANNED", "LOCKED", "EXECUTING", "VERIFYING", "REVIEWING", "DONE"]
TRANSITIONS = {
    "DRAFT": {"PLANNED", "BLOCKED"},
    "PLANNED": {"LOCKED", "DRAFT", "BLOCKED"},
    "LOCKED": {"EXECUTING", "BLOCKED"},
    "EXECUTING": {"VERIFYING", "BLOCKED"},
    "VERIFYING": {"REVIEWING", "EXECUTING", "BLOCKED"},
    "REVIEWING": {"DONE", "EXECUTING", "BLOCKED"},
    "BLOCKED": {"PLANNED", "LOCKED", "EXECUTING"},
    "DONE": set(),
}


def new_goal(goal_id: str, title: str, phase: str, objective: str = "") -> dict[str, Any]:
    return {
        "id": goal_id,
        "title": title,
        "phase": phase,
        "state": "DRAFT",
        "objective": objective or f"Define the measurable outcome for {title}.",
        "constraints": [],
        "non_goals": [],
        "acceptance": ["Replace this placeholder with objective acceptance criteria."],
        "gates": {
            "build": "required",
            "tests": "required",
            "review": "required",
            "documentation_impact": "required",
            "project_intelligence": "required",
        },
        "context": {
            "budget_profile": "medium",
            "max_delegation_depth": 1,
        },
        "dependencies": [],
        "evidence": [],
        "history": [
            {
                "at": datetime.now(timezone.utc).isoformat(),
                "event": "created",
                "state": "DRAFT",
            }
        ],
    }


def transition_goal(path: Path, target: str, reason: str = "") -> dict[str, Any]:
    goal = load_data(path)
    current = str(goal.get("state", "DRAFT")).upper()
    target = target.upper()
    allowed = TRANSITIONS.get(current, set())
    if target not in allowed:
        raise ValueError(f"Invalid goal transition: {current} -> {target}")
    if target == "DONE" and not goal.get("evidence"):
        raise ValueError("A goal cannot be marked DONE without evidence.")
    goal["state"] = target
    goal.setdefault("history", []).append(
        {
            "at": datetime.now(timezone.utc).isoformat(),
            "event": "state-transition",
            "from": current,
            "to": target,
            "reason": reason,
        }
    )
    if path.suffix.lower() == ".json":
        dump_json(goal, path)
    else:
        raise ValueError("Legacy YAML Goal is read-only in v0.2; migrate it to .goal.json before changing state.")
    return goal
