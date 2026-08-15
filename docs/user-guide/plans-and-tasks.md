# Plans, Tasks & DAGs

A **Plan** (`schemas/plan.schema.json`) defines a directed acyclic graph (DAG) of **Tasks** (`schemas/task.schema.json`).

## Invariants
- Tasks must have unique IDs within the plan.
- Dependency graphs must be strictly acyclic (no cycles).
- Each task declares required skills, isolation scope, and budget.
