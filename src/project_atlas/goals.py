from __future__ import annotations

import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from .io import dump_json, load_data

STATES = ["DRAFT", "PLANNED", "LOCKED", "EXECUTING", "VERIFYING", "REVIEWING", "BLOCKED", "DONE"]
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


def compute_goal_digest(goal: dict[str, Any]) -> str:
    """Compute deterministic SHA256 digest over locked Goal fields."""
    payload = {
        "objective": str(goal.get("objective", "")).strip(),
        "acceptance": sorted(str(x).strip() for x in goal.get("acceptance", [])),
        "constraints": sorted(str(x).strip() for x in goal.get("constraints", [])),
        "non_goals": sorted(str(x).strip() for x in goal.get("non_goals", [])),
        "gates": goal.get("gates", {}),
    }
    encoded = json.dumps(payload, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(encoded).hexdigest()


def verify_goal_lock(goal: dict[str, Any]) -> tuple[bool, str]:
    """Verify that a locked goal has not been modified without amendment."""
    state = str(goal.get("state", "DRAFT")).upper()
    lock = goal.get("lock")
    if not lock:
        if state in {"LOCKED", "EXECUTING", "VERIFYING", "REVIEWING", "DONE"}:
            return False, "Goal is in locked/executing state but lacks a lock record."
        return True, "Goal is not locked."
    expected_digest = lock.get("digest")
    current_digest = compute_goal_digest(goal)
    if expected_digest != current_digest:
        return False, f"Lock digest mismatch: expected {expected_digest}, got {current_digest}"
    return True, "Lock integrity valid."


def new_goal(goal_id: str, title: str, phase: str, objective: str = "") -> dict[str, Any]:
    return {
        "id": goal_id,
        "title": title,
        "phase": phase,
        "revision": 1,
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

    # When transitioning to LOCKED, record the cryptographic digest and lock metadata
    if target == "LOCKED":
        goal["revision"] = int(goal.get("revision", 1))
        goal["lock"] = {
            "revision": goal["revision"],
            "digest": compute_goal_digest(goal),
            "locked_at": datetime.now(timezone.utc).isoformat(),
        }
    elif current in {"LOCKED", "EXECUTING", "VERIFYING", "REVIEWING"} and goal.get("lock"):
        valid, msg = verify_goal_lock(goal)
        if not valid:
            raise ValueError(f"Illegal goal mutation detected: {msg}")

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


def amend_goal(path: Path, amendment: dict[str, Any]) -> dict[str, Any]:
    """Apply a formal amendment to a locked goal, updating its revision and lock."""
    goal = load_data(path)
    current = str(goal.get("state", "DRAFT")).upper()
    if current not in {"LOCKED", "EXECUTING", "VERIFYING", "REVIEWING", "BLOCKED"}:
        raise ValueError(f"Cannot amend goal in {current} state (must be locked/active).")

    new_rev = int(goal.get("revision", 1)) + 1
    goal["revision"] = new_rev

    changes = amendment.get("changes", {})
    if "objective" in changes:
        goal["objective"] = changes["objective"]
    if "acceptance" in changes:
        goal["acceptance"] = changes["acceptance"]
    if "constraints" in changes:
        goal["constraints"] = changes["constraints"]
    if "non_goals" in changes:
        goal["non_goals"] = changes["non_goals"]
    if "gates" in changes:
        goal["gates"] = changes["gates"]

    digest = compute_goal_digest(goal)
    amendment_record = {
        "id": amendment.get("id", f"AMD-{new_rev:03d}"),
        "goal_id": goal["id"],
        "revision": new_rev,
        "reason": amendment.get("reason", "Formal amendment"),
        "approved_by": amendment.get("approved_by", "human"),
        "approved_at": amendment.get("approved_at", datetime.now(timezone.utc).isoformat()),
        "changes": changes,
        "digest": digest,
    }
    goal.setdefault("amendments", []).append(amendment_record)
    goal["lock"] = {
        "revision": new_rev,
        "digest": digest,
        "locked_at": amendment_record["approved_at"],
    }
    goal.setdefault("history", []).append(
        {
            "at": amendment_record["approved_at"],
            "event": "amended",
            "revision": new_rev,
            "reason": amendment_record["reason"],
            "approved_by": amendment_record["approved_by"],
        }
    )
    if path.suffix.lower() == ".json":
        dump_json(goal, path)
    else:
        raise ValueError("Legacy YAML Goal is read-only; migrate to JSON first.")
    return goal
