from __future__ import annotations

import json
from pathlib import Path

from jsonschema import Draft202012Validator
from referencing import Registry, Resource
from referencing.jsonschema import DRAFT202012

SCHEMA_DIR = Path(__file__).parent.parent / "schemas"


def build_registry() -> Registry:
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


_REGISTRY = build_registry()


def load_schema(name: str) -> dict:
    path = SCHEMA_DIR / name
    assert path.exists(), f"Schema file {name} does not exist"
    return json.loads(path.read_text(encoding="utf-8"))


def validate(schema_name: str, instance: dict) -> list[str]:
    schema = load_schema(schema_name)
    validator = Draft202012Validator(schema, registry=_REGISTRY)
    return [e.message for e in validator.iter_errors(instance)]



# 1. Atlas Schema
def test_atlas_schema_positive():
    valid = {
        "version": 3,
        "protocol": {"version": 3, "compatible": ">=3 <4"},
        "framework": {"name": "project-atlas-framework", "version": "0.3.0"},
        "project": {"name": "test-project", "type": "web-app"},
        "documentation": {
            "entrypoint": "ENTRYPOINT.md",
            "canonical_format": "markdown",
            "site": {},
            "audiences": ["developer", "agent"]
        },
        "context": {
            "methodology": "lean-progressive-context",
            "mode": "progressive",
            "profiles": {"default": {}},
            "deep_recursion": {"enabled": False},
            "runtime": {}
        },
        "intelligence": {"enabled": True, "path": ".ai/intelligence"},
        "orchestration": {"protocol": "atlas", "orchestrator": "atlas-flow"},
        "goals": {"active_phase": "P01"},
        "ai": {"preferred_models": ["gpt-4o", "claude-3-7-sonnet"]}
    }
    assert validate("atlas.schema.json", valid) == []


def test_atlas_schema_negative():
    invalid = {"version": 1}  # Missing required fields and invalid version
    assert len(validate("atlas.schema.json", invalid)) > 0


# 2. Goal Schema
def test_goal_schema_positive():
    valid = {
        "id": "P01-G01",
        "title": "Core Runtime",
        "phase": "P01",
        "revision": 2,
        "state": "LOCKED",
        "lock": {
            "revision": 2,
            "digest": "abcdef123456",
            "locked_at": "2026-08-14T17:00:00Z"
        },
        "objective": "Build the core runtime.",
        "acceptance": ["Runtime starts deterministically."],
        "gates": {"tests": "required"},
        "dependencies": [],
        "evidence": [],
        "amendments": [
            {
                "id": "AMD-002",
                "revision": 2,
                "reason": "Scope clarification",
                "approved_by": "human",
                "approved_at": "2026-08-14T17:00:00Z"
            }
        ]
    }
    assert validate("goal.schema.json", valid) == []


def test_goal_schema_negative():
    invalid = {
        "id": "Invalid ID with spaces!",
        "title": "",
        "state": "UNKNOWN_STATE"
    }
    assert len(validate("goal.schema.json", invalid)) > 0


# 3. Goal Amendment Schema
def test_goal_amendment_schema_positive():
    valid = {
        "id": "AMD-002",
        "goal_id": "P01-G01",
        "revision": 2,
        "reason": "Add integration test requirement",
        "approved_by": "lead-architect",
        "approved_at": "2026-08-14T17:00:00Z",
        "changes": {
            "acceptance": ["Runtime starts deterministically.", "Passes integration suite."]
        }
    }
    assert validate("goal-amendment.schema.json", valid) == []


def test_goal_amendment_schema_negative():
    invalid = {"id": "AMD-001"}  # Missing goal_id, revision, changes, etc.
    assert len(validate("goal-amendment.schema.json", invalid)) > 0


# 4. Plan & Task Schemas
def test_plan_schema_positive():
    valid = {
        "id": "PLAN-001",
        "version": 1,
        "goal_id": "P01-G01",
        "goal_revision": 1,
        "created_at": "2026-08-14T17:00:00Z",
        "created_by": "architect",
        "status": "APPROVED",
        "tasks": ["T-001", "T-002"],
        "edges": [{"from": "T-001", "to": "T-002"}],
        "assumptions": ["Rust toolchain available."],
        "risks": ["Async runtime deadlock risk."]
    }
    assert validate("plan.schema.json", valid) == []


def test_plan_schema_negative():
    invalid = {"id": "PLAN-001", "tasks": []}  # minItems: 1
    assert len(validate("plan.schema.json", invalid)) > 0


def test_task_schema_positive():
    valid = {
        "id": "T-001",
        "goal_id": "P01-G01",
        "plan_id": "PLAN-001",
        "title": "Implement CLI argument parser",
        "objective": "Support --json and --dry-run flags.",
        "dependencies": [],
        "role": "implementer",
        "required_skills": ["clean-code"],
        "status": "PLANNED"
    }
    assert validate("task.schema.json", valid) == []


def test_task_schema_negative():
    invalid = {"id": "T-001", "title": "Missing required fields"}
    assert len(validate("task.schema.json", invalid)) > 0


# 5. Evidence, Gate & Waiver Schemas
def test_evidence_schema_positive():
    valid = {
        "id": "EV-001",
        "type": "test",
        "producer": "pytest",
        "timestamp": "2026-08-14T17:00:00Z",
        "status": "passed",
        "goal_id": "P01-G01",
        "task_id": "T-001",
        "command": "pytest tests/test_cli.py",
        "confidence": "high"
    }
    assert validate("evidence.schema.json", valid) == []


def test_evidence_schema_negative():
    invalid = {"id": "EV-001", "type": "invalid-type"}
    assert len(validate("evidence.schema.json", invalid)) > 0


def test_gate_schema_positive():
    valid = {
        "id": "GATE-TESTS",
        "name": "Unit and Integration Tests",
        "type": "test",
        "required": True,
        "required_evidence": ["EV-001"],
        "status": "passed"
    }
    assert validate("gate.schema.json", valid) == []


def test_gate_schema_negative():
    invalid = {"id": "GATE-01", "type": "unsupported"}
    assert len(validate("gate.schema.json", invalid)) > 0


def test_gate_waiver_schema_positive():
    valid = {
        "gate_id": "GATE-PERF",
        "reason": "Benchmark environment temporarily unavailable in CI",
        "approved_by": "security-officer",
        "approved_at": "2026-08-14T17:00:00Z"
    }
    assert validate("gate-waiver.schema.json", valid) == []


def test_gate_waiver_schema_negative():
    invalid = {"gate_id": "GATE-01"}  # Missing reason, approved_by, approved_at
    assert len(validate("gate-waiver.schema.json", invalid)) > 0


# 6. Run & Attempt Schemas
def test_run_schema_positive():
    valid = {
        "id": "RUN-001",
        "task_id": "T-001",
        "goal_id": "P01-G01",
        "plan_id": "PLAN-001",
        "status": "RUNNING",
        "started_at": "2026-08-14T17:00:00Z",
        "attempts": ["ATT-001"]
    }
    assert validate("run.schema.json", valid) == []


def test_attempt_schema_positive():
    valid = {
        "id": "ATT-001",
        "run_id": "RUN-001",
        "number": 1,
        "agent_role": "implementer",
        "skill_set": ["clean-code"],
        "model_profile": "primary-coder",
        "execution_backend": "cli",
        "started_at": "2026-08-14T17:00:00Z",
        "status": "SUCCESS"
    }
    assert validate("attempt.schema.json", valid) == []


# 7. Event Protocol Schema
def test_event_schema_positive():
    valid = {
        "id": "EVT-001",
        "type": "task.completed",
        "timestamp": "2026-08-14T17:00:00Z",
        "producer": "atlas-flow",
        "payload": {"exit_code": 0}
    }
    assert validate("event.schema.json", valid) == []


def test_event_schema_negative():
    invalid = {"id": "EVT-001", "type": "task.completed"}  # Missing timestamp, producer, payload
    assert len(validate("event.schema.json", invalid)) > 0


# 8. Permission, Approval & Trust Policies
def test_permission_policy_schema_positive():
    valid = {
        "version": 1,
        "mode": "controlled",
        "capabilities": [
            "filesystem.read",
            {"capability": "filesystem.write", "scope": "project-worktree"},
            {"capability": "network.http", "hosts": ["github.com"]}
        ],
        "risk_classes": {
            "READ": "automatic",
            "SAFE_WRITE": "confirm",
            "EXECUTE": "confirm"
        }
    }
    assert validate("permission-policy.schema.json", valid) == []


def test_approval_policy_schema_positive():
    valid = {
        "version": 1,
        "require_approval": ["process.spawn", "git.push"],
        "approvers": ["human", "lead-reviewer"],
        "timeout_seconds": 300,
        "fallback": "deny"
    }
    assert validate("approval-policy.schema.json", valid) == []


def test_trust_policy_schema_positive():
    valid = {
        "version": 1,
        "trusted_sources": ["locked-goal", "adr"],
        "contextual_sources": ["src/", "tests/"],
        "untrusted_sources": ["web", "mcp-output", "github-issue"]
    }
    assert validate("trust-policy.schema.json", valid) == []


# 9. Context Protocol Schemas
def test_context_item_schema_positive():
    valid = {
        "id": "CTX-001",
        "kind": "file",
        "source": "src/main.rs",
        "locator": "src/main.rs",
        "trust": "contextual",
        "estimated_tokens": 150
    }
    assert validate("context-item.schema.json", valid) == []


def test_context_request_schema_positive():
    valid = {
        "task_id": "T-001",
        "intent": "Implement command line options",
        "required_topics": ["cli", "argparse"]
    }
    assert validate("context-request.schema.json", valid) == []


def test_context_plan_schema_positive():
    valid = {
        "task_id": "T-001",
        "strategy": "progressive-retrieval",
        "planned_items": ["CTX-001"]
    }
    assert validate("context-plan.schema.json", valid) == []


def test_context_pack_schema_positive():
    valid = {
        "id": "CPACK-001",
        "task_id": "T-001",
        "items": [
            {
                "id": "CTX-001",
                "kind": "file",
                "source": "src/main.rs",
                "locator": "src/main.rs"
            }
        ],
        "created_at": "2026-08-14T17:00:00Z"
    }
    assert validate("context-pack.schema.json", valid) == []


def test_context_budget_schema_positive():
    valid = {
        "max_input_tokens": 8000,
        "max_retrieved_tokens": 4000,
        "max_injected_tokens": 2000,
        "max_expansions": 3,
        "stop_reasons": ["sufficient_evidence", "budget_exhausted"]
    }
    assert validate("context-budget.schema.json", valid) == []


# 10. Model & Execution Policies
def test_model_policy_v2_schema_positive():
    valid = {
        "version": 2,
        "selection_rule": "least_workforce",
        "cross_provider_review": True,
        "roster": ["claude-3-7-sonnet", "gpt-4o"],
        "profiles": {
            "primary-coder": {
                "provider": "anthropic",
                "model": "claude-3-7-sonnet",
                "capabilities": ["code", "tool-use"],
                "cost_class": "high"
            }
        },
        "roles": {
            "implementer": "primary-coder",
            "reviewer": "claude-3-7-sonnet"
        }
    }
    assert validate("model-policy.schema.json", valid) == []


def test_execution_policy_schema_positive():
    valid = {
        "version": 1,
        "default_backend": "cli",
        "backends": {
            "cli": {
                "type": "cli",
                "harness": "acp",
                "endpoint": "/usr/local/bin/codex"
            }
        },
        "routing": {
            "implementer": "cli"
        }
    }
    assert validate("execution-policy.schema.json", valid) == []


# 11. Agent, Skill & Recipe Schemas
def test_agent_schema_positive():
    valid = {
        "id": "implementer",
        "name": "Implementer",
        "version": 2,
        "purpose": "Implement scoped Goal work with minimal unrelated change.",
        "core": True,
        "inputs": ["Goal", "Task", "Plan", "ContextPack"],
        "outputs": ["Code", "Tests", "Evidence"],
        "required_skills": ["clean-code"],
        "risk_level": "medium",
        "stop_conditions": ["evidence_complete", "budget_exhausted"]
    }
    assert validate("agent.schema.json", valid) == []


def test_skill_schema_positive():
    valid = {
        "id": "clean-code",
        "name": "Clean Code",
        "version": 1,
        "schema_version": 2,
        "purpose": "Maintain high cohesion, low coupling, and clear module boundaries.",
        "risk_level": "low",
        "modes": ["implementation", "review"],
        "provenance": {
            "origin": "framework",
            "license": "MIT"
        }
    }
    assert validate("skill.schema.json", valid) == []


def test_recipe_schema_positive():
    valid = {
        "id": "feature-standard",
        "name": "Standard Feature Development",
        "version": 1,
        "purpose": "End-to-end feature workflow from exploration to review.",
        "steps": [
            {
                "id": "step-explore",
                "name": "Explore codebase",
                "agent_role": "explorer",
                "skills": ["atlas-navigation"]
            },
            {
                "id": "step-implement",
                "name": "Implement feature",
                "depends_on": ["step-explore"],
                "agent_role": "implementer",
                "skills": ["clean-code"]
            }
        ]
    }
    assert validate("recipe.schema.json", valid) == []


# Open Question (E-G01 Living Plan)
def test_open_question_schema_positive():
    valid = {
        "id": "OQ-001",
        "version": 1,
        "scope": "goal:G042",
        "contract": "architecture.system",
        "topic": "canonical state ownership",
        "priority": "blocker",
        "blocking": True,
        "reason": "implementation cannot start until the owner is known",
        "suggested_answers": ["control-plane", "project core"],
        "status": "open",
        "eligible_owners": ["human-maintained"],
        "question": "What owns canonical state?",
        "evidence": ["docs/architecture/overview.md"]
    }
    assert validate("open-question.schema.json", valid) == []


def test_open_question_schema_negative():
    unknown_priority = {
        "id": "OQ-002", "scope": "goal:G042", "priority": "curiosity", "status": "open"
    }
    errors = validate("open-question.schema.json", unknown_priority)
    assert errors, "expected unknown priority to be rejected"

    unknown_status = {
        "id": "OQ-003", "scope": "goal:G042", "priority": "blocker", "status": "draft"
    }
    errors = validate("open-question.schema.json", unknown_status)
    assert errors, "expected unknown status to be rejected"

    missing_scope = {"id": "OQ-004", "priority": "blocker", "status": "open"}
    errors = validate("open-question.schema.json", missing_scope)
    assert errors, "expected missing scope to be rejected"


# Decision Proposal (E-G02 Living Plan)
def test_decision_proposal_schema_positive():
    valid = {
        "id": "DP-001",
        "statement": "Control Plane owns canonical state",
        "classification": "explicit-decision",
        "scope": "goal:G042",
        "actor": "user",
        "source": "project-owner",
        "authority": "user-decision",
        "confidence": "unknown",
        "status": "accepted",
        "affected": ["docs/architecture/overview.md"],
        "rationale": "single writer avoids conflicting mutations",
        "alternatives": ["project core", "external service"],
        "evidence": ["docs/architecture/overview.md"],
        "created_at": "2026-09-09T12:00:00Z",
        "resolved_at": "2026-09-09T12:05:00Z"
    }
    assert validate("decision-proposal.schema.json", valid) == []


def test_decision_proposal_schema_unresolved_classification():
    valid = {
        "id": "DP-101", "statement": "storage engine not yet chosen",
        "classification": "unresolved",
        "authority": "agent-suggestion", "confidence": "low", "status": "proposed"
    }
    assert validate("decision-proposal.schema.json", valid) == []


def test_decision_proposal_schema_negative():
    unknown_classification = {
        "id": "DP-002", "statement": "x", "classification": "opinion",
        "authority": "agent-suggestion", "confidence": "unknown", "status": "proposed"
    }
    errors = validate("decision-proposal.schema.json", unknown_classification)
    assert errors, "expected unknown classification to be rejected"

    unknown_status = {
        "id": "DP-003", "statement": "x", "classification": "explicit-decision",
        "authority": "user-decision", "confidence": "unknown", "status": "merged"
    }
    errors = validate("decision-proposal.schema.json", unknown_status)
    assert errors, "expected unknown status to be rejected"

    missing_authority = {
        "id": "DP-004", "statement": "x", "classification": "explicit-decision",
        "confidence": "unknown", "status": "accepted"
    }
    errors = validate("decision-proposal.schema.json", missing_authority)
    assert errors, "expected missing authority to be rejected"

    unexpected_field = {
        "id": "DP-005", "statement": "x", "classification": "explicit-decision",
        "authority": "user-decision", "confidence": "unknown", "status": "accepted",
        "llm_temperature": 0.7
    }
    errors = validate("decision-proposal.schema.json", unexpected_field)
    assert errors, "expected unknown attribute to be rejected"


# Authority model helper tests (E-G02 Living Plan)
def test_authority_order():
    order = ["invariant", "user-decision", "project-decision", "documented-evidence",
             "inferred-state", "agent-suggestion", "external"]
    for i, authority in enumerate(order):
        errors = validate("decision-proposal.schema.json", {
            "id": "DP-A%d" % i, "statement": "x", "classification": "hypothesis",
            "authority": authority, "confidence": "low", "status": "proposed"
        })
        assert errors == [], "authority %s should be accepted" % authority


# Planning Session (E-G07 Living Plan resume/checkpoint + context compilation)
def _planning_session_valid() -> dict:
    return {
        "version": 1,
        "id": "SES-001",
        "run_id": "RUN-9",
        "scope": "goal:G042",
        "goal": "M6 living plan ready",
        "decisions": [
            {
                "id": "DP-001",
                "statement": "Control Plane owns canonical state",
                "classification": "explicit-decision",
                "authority": "user-decision",
                "confidence": "unknown",
                "status": "accepted"
            }
        ],
        "open_questions": [
            {
                "id": "OQ-001",
                "scope": "goal:G042",
                "priority": "blocker",
                "status": "open",
                "question": "What owns audit state?"
            }
        ],
        "last_preview": {
            "extracted_decisions": [],
            "blockers_resolved": 1,
            "open_remaining": 1,
            "preview_mandatory": True
        },
        "updated_at": "2026-09-09T12:00:00Z"
    }


def test_planning_session_schema_positive():
    assert validate("planning-session.schema.json", _planning_session_valid()) == []


def test_planning_session_schema_reuses_linked_models():
    payload = _planning_session_valid()
    payload["decisions"][0]["status"] = "merged"
    assert validate("planning-session.schema.json", payload), \
        "invalid decision must fail through the referenced model"


def test_planning_session_schema_negative():
    missing_scope = _planning_session_valid()
    del missing_scope["scope"]
    errors = validate("planning-session.schema.json", missing_scope)
    assert errors, "expected missing scope to be rejected"

    unexpected_field = _planning_session_valid()
    unexpected_field["transcript"] = "full raw conversation"
    errors = validate("planning-session.schema.json", unexpected_field)
    assert errors, "expected raw transcript field to be rejected"

    bad_preview_type = _planning_session_valid()
    bad_preview_type["last_preview"]["blockers_resolved"] = -1
    errors = validate("planning-session.schema.json", bad_preview_type)
    assert errors, "expected negative blockers_resolved to be rejected"
