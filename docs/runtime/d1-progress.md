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

- Persisted Run lifecycle beyond bootstrap creation.
- `atlas run resume`, cancellation, retry taxonomy, and livelock detection CLI.
- Full Context Manifest rehydration from Goal/Git/evidence sources.
- Budget/cost/rate-limit commands and accounting integration remain minimal: budget inspection is available, reservation accounting is package-level.
- Observability explain/debug bundle CLI remains to be completed.
- Process-kill and fresh-executor dogfood without connector history: generic fresh-executor fixture passes; process-kill integration remains the final D1 gate.
- `atlas run context` now recompiles a deterministic Context Manifest for a Run.
