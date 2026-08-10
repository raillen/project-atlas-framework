# Goal System

A roadmap says what should happen over time. A Goal defines an outcome that can be independently verified.

## Hierarchy

`Vision → Project Goals → Phase → Goals → Subgoals → Task DAG → Evidence`

## Required fields

A Goal has an ID, phase, objective, constraints, non-goals, acceptance criteria, gates, dependencies, evidence and state history.

## State machine

```text
DRAFT → PLANNED → LOCKED → EXECUTING → VERIFYING → REVIEWING → DONE
          ↑          ↓          ↓             ↓
          └──────── BLOCKED ─────┴─────────────┘
```

The CLI enforces legal transitions. `DONE` requires evidence.

## Locking

Once LOCKED, acceptance criteria are protected. Discoveries that genuinely change the outcome require an explicit Goal Amendment with rationale and review; test failure is not a reason to relax acceptance.

## Evidence

Evidence can include test reports, CI runs, benchmarks, screenshots, logs, artifacts, review reports and manual gate records. Store stable evidence paths/URLs when practical.
