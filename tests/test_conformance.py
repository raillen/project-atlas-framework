from __future__ import annotations

import json
from pathlib import Path

from jsonschema import Draft202012Validator
from referencing import Registry, Resource
from referencing.jsonschema import DRAFT202012

from project_atlas.doctor import check_dag_cycles, diagnose_project
from project_atlas.fake_runtime import FakeRuntime
from project_atlas.goals import verify_goal_lock
from project_atlas.io import load_json

CONF_ROOT = Path(__file__).parent.parent / "conformance"
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


def validate_schema(schema_name: str, instance: dict) -> list[str]:
    schema = json.loads((SCHEMA_DIR / schema_name).read_text(encoding="utf-8"))
    validator = Draft202012Validator(schema, registry=_REGISTRY)
    return [e.message for e in validator.iter_errors(instance)]


# 1. Goal Conformance
def test_conformance_locked_goal():
    goal = load_json(CONF_ROOT / "goals/valid_locked_goal.json")
    assert validate_schema("goal.schema.json", goal) == []
    valid, msg = verify_goal_lock(goal)
    assert valid is True, msg


def test_conformance_goal_amendment():
    amendment = load_json(CONF_ROOT / "goals/goal_amendment.json")
    assert validate_schema("goal-amendment.schema.json", amendment) == []


# 2. Plan & DAG Conformance
def test_conformance_valid_plan_dag():
    plan = load_json(CONF_ROOT / "plans/valid_plan_dag.json")
    assert validate_schema("plan.schema.json", plan) == []
    tasks = plan["tasks"]
    assert check_dag_cycles(tasks) == []


def test_conformance_cyclic_dag_detection():
    data = load_json(CONF_ROOT / "plans/cyclic_dag_error.json")
    errors = check_dag_cycles(data["tasks"])
    assert len(errors) > 0
    assert any("cycle" in err.lower() for err in errors)


# 3. Context Conformance
def test_conformance_context_pack():
    cpack = load_json(CONF_ROOT / "context/valid_context_pack.json")
    assert validate_schema("context-pack.schema.json", cpack) == []


# 4. Evidence & Gates Conformance
def test_conformance_evidence():
    evidence = load_json(CONF_ROOT / "evidence/valid_test_evidence.json")
    assert validate_schema("evidence.schema.json", evidence) == []


def test_conformance_gate_waiver():
    waiver = load_json(CONF_ROOT / "gates/valid_gate_waiver.json")
    assert validate_schema("gate-waiver.schema.json", waiver) == []


# 5. Policy Conformance
def test_conformance_policies():
    perm = load_json(CONF_ROOT / "policies/permission_policy.json")
    assert validate_schema("permission-policy.schema.json", perm) == []

    trust = load_json(CONF_ROOT / "policies/trust_policy.json")
    assert validate_schema("trust-policy.schema.json", trust) == []


# 6. Event Protocol Conformance
def test_conformance_events():
    events = load_json(CONF_ROOT / "events/sample_event_stream.json")
    for evt in events:
        assert validate_schema("event.schema.json", evt) == []


# 7. Fake Runtime Simulation (End-to-End Execution Flow)
def test_conformance_fake_runtime_execution():
    runtime = FakeRuntime(project_id="conformance-project")
    goal = load_json(CONF_ROOT / "goals/valid_locked_goal.json")
    plan = load_json(CONF_ROOT / "plans/valid_plan_dag.json")

    # Happy path run
    result = runtime.execute_plan(goal, plan)
    assert result.status == "COMPLETED"
    assert len(result.runs) == 2
    assert len(result.events) >= 10
    assert len(result.evidence) >= 2
    assert len(result.gates) >= 2

    # Validate all generated events against event.schema.json
    for evt in result.events:
        errs = validate_schema("event.schema.json", evt)
        assert errs == [], f"Invalid event {evt}: {errs}"

    # Validate all generated evidence against evidence.schema.json
    for ev in result.evidence:
        errs = validate_schema("evidence.schema.json", ev)
        assert errs == [], f"Invalid evidence {ev}: {errs}"

    # Validate all generated runs against run.schema.json
    for r in result.runs:
        errs = validate_schema("run.schema.json", r)
        assert errs == [], f"Invalid run {r}: {errs}"


def test_conformance_fake_runtime_retry_and_fallback():
    runtime = FakeRuntime(project_id="conformance-retry-test")
    goal = load_json(CONF_ROOT / "goals/valid_locked_goal.json")
    plan = load_json(CONF_ROOT / "plans/valid_plan_dag.json")

    # Simulate transient retry on T-01
    result = runtime.execute_plan(goal, plan, behavior={"T-01.retry": True})
    assert result.status == "COMPLETED"
    retry_events = [e for e in result.events if e["type"] == "attempt.failed"]
    assert len(retry_events) == 1


# 8. End-to-End Conformance Project Doctor Check
def test_conformance_project_diagnostics():
    project_root = Path(__file__).parent.parent / "examples/conformance-project"
    findings = diagnose_project(project_root)
    errors = [f for f in findings if f.severity == "ERROR"]
    assert errors == [], f"Doctor reported errors on conformance project: {errors}"
