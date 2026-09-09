from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any


@dataclass
class SimulationResult:
    goal_id: str
    plan_id: str
    status: str  # "COMPLETED", "FAILED", "BLOCKED"
    runs: list[dict[str, Any]] = field(default_factory=list)
    events: list[dict[str, Any]] = field(default_factory=list)
    evidence: list[dict[str, Any]] = field(default_factory=list)
    gates: list[dict[str, Any]] = field(default_factory=list)


class FakeRuntime:
    """Deterministic, provider-neutral fake runtime for Project Atlas conformance testing."""

    def __init__(self, project_id: str = "conformance-test"):
        self.project_id = project_id
        self.events: list[dict[str, Any]] = []
        self.evidence: list[dict[str, Any]] = []
        self.runs: list[dict[str, Any]] = []
        self.gates: list[dict[str, Any]] = []
        self._event_counter = 0

    def emit_event(
        self,
        event_type: str,
        payload: dict[str, Any],
        goal_id: str | None = None,
        plan_id: str | None = None,
        task_id: str | None = None,
        run_id: str | None = None,
        attempt_id: str | None = None,
    ) -> dict[str, Any]:
        self._event_counter += 1
        event = {
            "id": f"evt-{self._event_counter:04d}",
            "type": event_type,
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "project_id": self.project_id,
            "goal_id": goal_id,
            "plan_id": plan_id,
            "task_id": task_id,
            "run_id": run_id,
            "attempt_id": attempt_id,
            "producer": "fake-runtime",
            "schema_version": 1,
            "payload": payload,
        }
        self.events.append(event)
        return event

    def create_evidence(
        self,
        evidence_type: str,
        goal_id: str,
        task_id: str | None = None,
        run_id: str | None = None,
        attempt_id: str | None = None,
        status: str = "passed",
        command: str | None = None,
        artifact: str | None = None,
    ) -> dict[str, Any]:
        ev_id = f"EV-{len(self.evidence) + 1:03d}"
        ev = {
            "id": ev_id,
            "type": evidence_type,
            "producer": "fake-runtime",
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "status": status,
            "goal_id": goal_id,
            "task_id": task_id,
            "run_id": run_id,
            "attempt_id": attempt_id,
            "command": command,
            "artifact": artifact,
            "confidence": "high",
        }
        self.evidence.append(ev)
        self.emit_event(
            "evidence.created",
            {"evidence_id": ev_id, "type": evidence_type, "status": status},
            goal_id=goal_id,
            task_id=task_id,
            run_id=run_id,
            attempt_id=attempt_id,
        )
        return ev

    def execute_plan(
        self,
        goal: dict[str, Any],
        plan: dict[str, Any],
        behavior: dict[str, Any] | None = None,
    ) -> SimulationResult:
        behavior = behavior or {}
        goal_id = goal["id"]
        plan_id = plan["id"]

        self.emit_event("goal.locked", {"goal_id": goal_id, "revision": goal.get("revision", 1)}, goal_id=goal_id)
        self.emit_event("plan.approved", {"plan_id": plan_id}, goal_id=goal_id, plan_id=plan_id)

        tasks = plan.get("tasks", [])
        task_map = {t["id"]: t for t in tasks if isinstance(t, dict)}

        for task_id, task in task_map.items():
            run_id = f"RUN-{task_id}"
            self.emit_event("task.started", {"role": task.get("role")}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)

            should_fail = behavior.get(f"{task_id}.fail", False)
            should_retry = behavior.get(f"{task_id}.retry", False)
            should_timeout = behavior.get(f"{task_id}.timeout", False)
            should_invalid = behavior.get(f"{task_id}.invalid_evidence", False)
            should_deny_tool = behavior.get(f"{task_id}.tool_denied", False)
            should_deny_perm = behavior.get(f"{task_id}.permission_denied", False)

            # Attempt 1
            att1_id = f"ATT-{task_id}-1"
            self.emit_event("attempt.started", {"number": 1, "backend": "cli"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)

            if should_timeout:
                self.emit_event("attempt.failed", {"error": "timeout", "status": "TIMED_OUT"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                self.emit_event("task.failed", {"reason": "Attempt failed"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
                return SimulationResult(goal_id=goal_id, plan_id=plan_id, status="FAILED", runs=self.runs, events=self.events, evidence=self.evidence, gates=self.gates)

            if should_deny_tool:
                self.emit_event("tool.failed", {"error": "Tool denied"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                self.emit_event("attempt.failed", {"error": "Tool denied"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                self.emit_event("task.failed", {"reason": "Attempt failed"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
                return SimulationResult(goal_id=goal_id, plan_id=plan_id, status="FAILED", runs=self.runs, events=self.events, evidence=self.evidence, gates=self.gates)

            if should_deny_perm:
                self.emit_event("permission.denied", {"error": "Permission denied"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                self.emit_event("attempt.failed", {"error": "Permission denied"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                self.emit_event("task.failed", {"reason": "Permission denied"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
                return SimulationResult(goal_id=goal_id, plan_id=plan_id, status="FAILED", runs=self.runs, events=self.events, evidence=self.evidence, gates=self.gates)

            if should_invalid:
                ev = self.create_evidence("test", goal_id=goal_id, task_id=task_id, run_id=run_id, attempt_id=att1_id, command="cargo test", status="invalid")
                self.emit_event("attempt.completed", {"status": "SUCCESS"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                gate_id = f"GATE-{task_id}"
                self.emit_event("gate.started", {"gate_id": gate_id}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
                self.emit_event("gate.failed", {"gate_id": gate_id, "evidence_id": ev["id"]}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
                self.gates.append({"id": gate_id, "status": "failed", "evidence": ev["id"]})
                self.emit_event("task.failed", {"reason": "Gate failed"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
                return SimulationResult(goal_id=goal_id, plan_id=plan_id, status="FAILED", runs=self.runs, events=self.events, evidence=self.evidence, gates=self.gates)

            if should_retry:
                self.emit_event("attempt.failed", {"error": "Simulated transient failure"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                # Attempt 2 (fallback)
                att2_id = f"ATT-{task_id}-2"
                self.emit_event("attempt.started", {"number": 2, "backend": "direct"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att2_id)
                ev = self.create_evidence("test", goal_id=goal_id, task_id=task_id, run_id=run_id, attempt_id=att2_id, command="cargo test")
                self.emit_event("attempt.completed", {"status": "SUCCESS"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att2_id)
                task_status = "DONE"
            elif should_fail:
                self.emit_event("attempt.failed", {"error": "Simulated permanent failure"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                self.emit_event("task.failed", {"reason": "Attempt failed"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
                return SimulationResult(
                    goal_id=goal_id,
                    plan_id=plan_id,
                    status="FAILED",
                    runs=self.runs,
                    events=self.events,
                    evidence=self.evidence,
                    gates=self.gates,
                )
            else:
                ev = self.create_evidence("test", goal_id=goal_id, task_id=task_id, run_id=run_id, attempt_id=att1_id, command="cargo test")
                self.emit_event("attempt.completed", {"status": "SUCCESS"}, goal_id=goal_id, plan_id=plan_id, task_id=task_id, run_id=run_id, attempt_id=att1_id)
                task_status = "DONE"

            # Check Task Gates
            gate_id = f"GATE-{task_id}"
            self.emit_event("gate.started", {"gate_id": gate_id}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
            self.emit_event("gate.passed", {"gate_id": gate_id, "evidence_id": ev["id"]}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
            self.gates.append({"id": gate_id, "status": "passed", "evidence": ev["id"]})

            self.emit_event("task.completed", {"status": task_status}, goal_id=goal_id, plan_id=plan_id, task_id=task_id)
            self.runs.append({
                "id": run_id,
                "task_id": task_id,
                "goal_id": goal_id,
                "plan_id": plan_id,
                "status": "COMPLETED",
                "started_at": datetime.now(timezone.utc).isoformat(),
                "evidence": [ev["id"]],
            })

        self.emit_event("goal.completed", {"goal_id": goal_id}, goal_id=goal_id)
        return SimulationResult(
            goal_id=goal_id,
            plan_id=plan_id,
            status="COMPLETED",
            runs=self.runs,
            events=self.events,
            evidence=self.evidence,
            gates=self.gates,
        )
