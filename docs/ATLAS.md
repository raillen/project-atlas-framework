# Project Atlas Framework — ATLAS

ATLAS is an intent router, not a requirement to read every document.

## I want to understand the framework

- [Framework contract](../FRAMEWORK.md)
- [Architecture](ARCHITECTURE.md)
- [Usage](USAGE.md)

## I want to document a project for users/developers/operators

- [Documentation System](DOCUMENTATION_SYSTEM.md)
- [Portability and platform adapters](PORTABILITY.md)
- [Validation and quality](QUALITY.md)

## I want to work with AI agents efficiently

- [Lean Progressive Context](LEAN_PROGRESSIVE_CONTEXT.md)
- [Agents, Skills and Recipes](AI_WORKFORCE.md)
- [Project Orchestration Protocol](PROJECT_ORCHESTRATION_PROTOCOL.md)

## I want to define measurable work

- [Goal System](GOAL_SYSTEM.md)
- [Project Intelligence](PROJECT_INTELLIGENCE.md)

## I want to evolve or migrate the framework

- [Contributing and evolution](CONTRIBUTING.md)
- [Framework change proposal](FRAMEWORK_CHANGE_PROPOSAL.md)
- [Migration v0.1 → v0.2](MIGRATION_V0_1_TO_V0_2.md)

## Machine contracts

- `schemas/project-profile.schema.json` — init/import profile.
- `schemas/atlas.schema.json` — canonical `atlas.json`.
- `schemas/goal.schema.json` — Goal records.
- `schemas/model-policy.schema.json` — model routing.
- `schemas/task-report.schema.json` — durable task intelligence.

## Canonical registries

- `src/project_atlas/resources/catalog/catalog.json`

## Persistent format rule

The framework-maintained surface is intentionally small:

- Markdown for human knowledge.
- JSON for machine contracts.
- SQLite only for derived/runtime state.

YAML is legacy read compatibility, not a v0.2 canonical format.
