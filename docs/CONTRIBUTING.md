# Contributing and Framework Evolution

Project Atlas is intended to learn from real projects without becoming a dumping ground for project-specific instructions.

A reusable change should identify the observed problem, evidence from one or more projects, proposed generic rule, affected schemas/catalogs/adapters and compatibility impact. Breaking protocol changes require a major version and migration guidance.

## Checklist

- Keep core provider-neutral.
- Keep `ENTRYPOINT.md` and root `AGENTS.md` short.
- Update schemas when contracts change.
- Add tests for behavior changes.
- Do not bake a current model ranking into canonical agent definitions.
- Prefer capabilities/risk selectors over technology-specific duplication.
