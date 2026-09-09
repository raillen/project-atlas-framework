# Documentation System v2

M5 makes documentation an engineering subsystem without requiring an LLM to decide policy.

## Commands

```bash
atlas docs contracts
atlas docs contracts show architecture.system
atlas docs profiles
atlas docs audit --json
atlas docs readiness --goal G042 --json
```

These commands are read-only. Documentation Delta application remains governed by Repository Change Governance.

## Coverage states

| State | Meaning |
|---|---|
| `missing` | No canonical binding exists. |
| `partial` | Binding exists but required knowledge is missing. |
| `implementation-ready` | Knowledge is sufficient for the scoped implementation. |
| `verified` | Required knowledge and declared evidence were verified. |
| `stale` | A relevant change trigger invalidated assumptions. |
| `not-applicable` | Applicability rules excluded the contract deterministically. |

Coverage is not reduced to a percentage. Blocking contracts and missing knowledge remain primary.

M5 status: the deterministic foundation is implemented through contract/profile resolution, applicability, bindings, coverage, readiness, impact, Delta proposal, and deterministic contradiction/staleness findings. Semantic prose contradiction assistance, full evidence verification, and Delta application lifecycle integration remain explicit follow-up work before the M5 exit gate.

## Contract and profile sources

- Built-in contracts: `docs/contracts/builtin.json`.
- Built-in profiles: `docs/profiles/builtin.json`.
- Atlas bindings: `docs/contracts/bindings.json`.
- Machine schemas: `schemas/documentation-*.schema.json`.

## Authority

Canonical Git documents and locked protocol decisions outrank generated references, external sources, and agent proposals. Contradictions are findings; they are never silently resolved by recency or model output.

## Scope boundary

M5 implements contract/profile resolution, applicability, bindings, coverage, readiness, impact, Delta foundation, and structured staleness/contradiction findings. Interview planning belongs to M6; adoption belongs to M7; full traceability belongs to M8; experience belongs to M9.

## Current command surface

```bash
atlas docs contracts
atlas docs profiles
atlas docs audit --json
atlas docs readiness --goal G042 --json
atlas docs impact --path . changed/path.go --json
atlas docs delta --goal G042 --json
atlas docs contradictions --json
```

Coverage is deterministic and evidence-oriented. Current contradiction detection is intentionally conservative: it emits structured findings only when deterministic rules exist; it never silently resolves prose conflicts. Causal staleness uses contract update triggers and changed paths.
