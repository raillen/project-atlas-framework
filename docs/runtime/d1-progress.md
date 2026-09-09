# D1 Runtime Foundation Progress

Implemented baseline capabilities:

- `Run` lifecycle with explicit transition validation.
- `ExecutorSession` separated from Run identity.
- Checkpoint and ContinuationRecord v1 schemas/models.
- `atlas run` and `atlas continue` human/JSON/prompt paths.
- Derived continuation under `.atlas/runtime/`, ignored by Git.
- Hierarchical Budget Envelope package with reservations, soft/hard limits, and usage.
- Deterministic Context Compiler package with source ordering, deduplication, token budget, and pressure states.
- Structured Observability event append baseline.
- Failure taxonomy and side-effect journal baseline.
- Bounded retry classification helper.

Remaining D1 gate work:

- None. All D1 gate deliverables are complete and verified:
  - Persisted Run lifecycle (`internal/runtime`) with explicit transitions and validation.
  - CLI operations: `atlas run`, `atlas run show`, `atlas run cancel`, `atlas run resume`, `atlas run context`.
  - Continuation and prompt paths: `atlas continue`, `atlas continue --json`, `atlas continue --prompt`.
  - Hierarchical Budget Envelope (`internal/budget`) and budget CLI commands (`atlas budget`, `atlas budget explain`).
  - Deterministic Context Compiler (`internal/contextcompiler`) with source ordering, token limits, pressure states, and CLI manifest recompilation.
  - Observability events append (`internal/observability`) and sanitized debug bundle export (`atlas debug bundle`).
  - Failure taxonomy and side-effect journal baseline (`internal/runtime/failure`, `internal/runtime/journal`).
  - Portable continuation dogfood: fresh-executor fixture (`TestFreshExecutorContinuationDogfood`) and multi-process dogfood test (`TestPortableContinuationAcrossProcesses`).

Status: **COMPLETE (D1 Exit Gate Passed)**.

