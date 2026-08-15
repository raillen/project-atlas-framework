# Project Atlas Event Protocol

## 1. Overview

The Project Atlas Event Protocol defines a standardized, append-only, provider-neutral envelope for recording lifecycle events across Goals, Plans, Tasks, Runs, Attempts, Context, Tools, Evidence, and Gates.

Events are emitted by execution runtimes (such as Atlas Flow or CLI runners) and can be consumed for:
- Observability and tracing;
- Project Intelligence metrics aggregation;
- Audit trails and conformance verification;
- State machine synchronization.

## 2. Event Envelope Specification

All events conform to `schemas/event.schema.json`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "id": "evt-20260814-001",
  "type": "task.started",
  "timestamp": "2026-08-14T17:00:00Z",
  "project_id": "project-atlas-framework",
  "goal_id": "G-001",
  "plan_id": "P-001",
  "task_id": "T-001",
  "run_id": "R-001",
  "attempt_id": "A-001",
  "producer": "atlas-flow",
  "schema_version": 1,
  "payload": {
    "role": "implementer",
    "skills": ["clean-code", "secure-coding"]
  }
}
```

### 2.1 Envelope Fields

| Field | Type | Description |
|---|---|---|
| `id` | String | Unique event identifier (UUID or deterministic ID). |
| `type` | String | Event type from canonical vocabulary. |
| `timestamp` | String | ISO 8601 UTC timestamp. |
| `project_id` | String? | Project identifier. |
| `goal_id` | String? | Active Goal identifier. |
| `plan_id` | String? | Active Plan identifier. |
| `task_id` | String? | Active Task identifier. |
| `run_id` | String? | Active Run execution identifier. |
| `attempt_id` | String? | Active Attempt execution identifier. |
| `producer` | String | Producer identity (runtime, adapter, agent, tool). |
| `payload` | Object | Event-specific payload. |
| `schema_version` | Integer? | Envelope/payload schema version (default: 1). |

## 3. Canonical Event Vocabulary

### 3.1 Project Lifecycle
- `project.opened`: Project workspace loaded and verified.

### 3.2 Goal Lifecycle
- `goal.created`: Goal drafted.
- `goal.locked`: Goal acceptance criteria and constraints locked.
- `goal.amended`: Locked Goal amended through explicit approval.
- `goal.completed`: Goal satisfied with all required evidence.

### 3.3 Plan Lifecycle
- `plan.created`: Plan snapshot generated from Goal.
- `plan.reviewed`: Plan reviewed for DAG correctness and risk.
- `plan.approved`: Plan approved for execution.
- `plan.superseded`: Plan replaced by an amended version.

### 3.4 Task Lifecycle
- `task.queued`: Task added to execution queue.
- `task.started`: Task execution initiated.
- `task.blocked`: Task execution blocked by missing dependencies or human review.
- `task.completed`: Task finished successfully with required evidence.
- `task.failed`: Task failed.

### 3.5 Agent & Tool Lifecycle
- `agent.spawned`: Agent instance created with assigned role and skills.
- `agent.ready`: Agent initialized with context pack.
- `agent.failed`: Agent encountered unrecoverable runtime error.
- `agent.stopped`: Agent execution terminated under stop conditions.
- `tool.started`: Tool invocation began.
- `tool.completed`: Tool invocation finished.
- `tool.failed`: Tool invocation failed.

### 3.6 Context Lifecycle
- `context.created`: Initial minimal ContextPack constructed.
- `context.expanded`: Context expanded progressively under budget.
- `context.exhausted`: Context budget exhausted or stop reason reached.

### 3.7 Evidence & Gate Lifecycle
- `evidence.created`: Durable or operational evidence recorded.
- `gate.started`: Verification gate execution initiated.
- `gate.passed`: Verification gate satisfied by evidence.
- `gate.failed`: Verification gate failed.
- `gate.waived`: Verification gate bypassed through approved waiver.

### 3.8 Review & Execution
- `review.started`: Cross-review initiated.
- `review.completed`: Cross-review findings emitted.
- `run.created`: Task execution run initialized.
- `run.started`: Run started.
- `run.completed`: Run completed.
- `run.failed`: Run failed.
- `attempt.started`: Individual execution attempt initiated.
- `attempt.completed`: Attempt completed.
- `attempt.failed`: Attempt failed with retry/fallback reasoning.

## 4. Invariants and Append Semantics

1. **Immutability:** Once emitted, an event is never mutated or overwritten.
2. **Deterministic Replay:** Project state transitions can be audited or reconstructed by reading the chronological event stream.
3. **No Unbounded Payload Blobs:** Large outputs (logs, ASTs, diffs) must be referenced via hash or pointer path in `evidence`, not embedded wholly in event payloads.
