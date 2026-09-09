# M5 Documentation System v2 — Dogfood Report

Run against the Project Atlas repository:

```bash
atlas --json docs audit --path .
atlas --json docs readiness --goal M5 --path .
```

Observed result:

- Applicable profiles: `core-software`, `cli`.
- Applicable contracts: architecture, CLI, installation, product vision, scope, security, testing.
- Readiness: blocked.
- Blocking contracts include architecture, installation, product vision, and scope.

This is intentional. The engine reports missing semantic knowledge rather than fabricating completeness. The result is the baseline for the next Documentation Delta iteration.
