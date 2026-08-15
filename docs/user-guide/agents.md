# Agent Contract v2

Agents (`schemas/agent.schema.json`) represent operational roles.

## Core Agents
- `Explorer`: Read-only mapping of repository, assumptions, and impact.
- `Architect`: Interface boundaries, ADRs, and high-risk design decisions.
- `Implementer`: Scoped implementation with tests.
- `Debugger`: Root-cause reproduction, isolation, and repair.
- `Test Engineer`: Verification suites and regression tests.
- `Reviewer`: Independent cross-provider code and contract review.
- `Documentation Maintainer`: Canonical markdown synchronization.
- `Release Verifier`: Gate verification, checksums, and rollback readiness.
