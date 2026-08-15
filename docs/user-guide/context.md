# Lean Progressive Context (LPC/PCA)

LPC guarantees minimal token consumption and maximum signal:
- Load smallest sufficient context first (`ENTRYPOINT.md` + active Goal).
- Expand progressively by symbol/file only when needed.
- Enforce token budgets (`context-budget.schema.json`).
- Garbage-collect intermediate reasoning traces.
