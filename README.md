# Project Atlas Framework

Project Atlas is a Git-native, provider-agnostic framework for building and maintaining software with humans and AI agents. The repository—not chat memory, a specific LLM, an orchestrator or a generated documentation site—is the durable source of truth.

Version 0.2 evolves Atlas around **Lean Progressive Context (LPC)** and the **Progressive Context Architecture (PCA)**: use the smallest sufficient workforce and context, expand only when evidence requires it, budget both input and output tokens, persist only durable knowledge, and garbage-collect temporary context.

## Operational layers

1. **Knowledge** — ATLAS, canonical Markdown docs, ADRs/RFCs, specs and relationships.
2. **Goals** — measurable outcomes, acceptance criteria, gates and evidence.
3. **Documentation surfaces** — user, developer, operations and agent guidance.
4. **AI workforce** — selected provider-neutral Agents, Skills and Recipes.
5. **Progressive Context** — task maps/context packs, structural retrieval, WCC, budgets and stopping rules.
6. **Orchestration** — model routing, bounded fallbacks, isolated delegation and cross-provider review.
7. **Project Intelligence** — token/cost/effort/quality/debt history with observed-vs-estimated provenance.
8. **Publishing** — documentation site and dashboards derived from canonical sources.

## Persistent format policy

New projects deliberately keep the maintained surface small:

- **Markdown** — human knowledge and documentation.
- **JSON** — project configuration, manifests, Goals and compact durable intelligence.
- **SQLite** — derived/runtime indexes and temporary context only; normally gitignored.

YAML is accepted only as a v0.1 migration input.

## Install

```bash
python -m pip install -e .
atlas --version
```

For isolated installation:

```bash
pipx install .
```

## Start a project

Create a JSON profile (see `examples/brasa/project-profile.json`) and run:

```bash
atlas init ../my-project --profile examples/brasa/project-profile.json --non-interactive
atlas validate ../my-project --schemas ./schemas
```

Interactive `atlas init` explicitly asks for the model roster. Atlas never silently inherits model preferences from another project.

A new project is scaffolded around `atlas.json`, `PROJECT_STATE.md`, `docs/ATLAS.md`, compact JSON workforce manifests and JSON Goals.

## Resolve the AI workforce

```bash
atlas resolve examples/brasa/project-profile.json
```

The resolver composes the smallest useful workforce from core capabilities, project-type bundles, stack/features, risks, dependencies and project overrides.

## Compile platform adapters

```bash
atlas compile --target codex --path ../my-project
atlas compile --target claude-code --path ../my-project
atlas compile --target traycer --path ../my-project
```

Generated adapters are replaceable views. Generic/chat context artifacts are written under `.atlas/runtime/compiled/` and should not be committed.

## Goals

```bash
atlas goal new P01-G01 "Runtime window" --phase P01 --path ../my-project
atlas goal state P01-G01 PLANNED --path ../my-project
atlas goal state P01-G01 LOCKED --path ../my-project
atlas goal list --path ../my-project
```

New Goals are JSON. `DONE` requires evidence.

## Lean Progressive Context

LPC/PCA does not require deep recursion. The preferred order is:

```text
Direct
→ structural/symbol retrieval
→ known Context Pack
→ progressive retrieval
→ single-level isolated delegation if necessary
→ stop
```

Working context is rolling and ephemeral, not a growing collection of `SUMMARY.md`/`CONTEXT.md` files. See [`docs/LEAN_PROGRESSIVE_CONTEXT.md`](docs/LEAN_PROGRESSIVE_CONTEXT.md).

## Documentation system

Atlas treats documentation for users, contributors, operators and agents as first-class. Canonical Markdown may be structurally chunked virtually; physical microfile proliferation is discouraged. A documentation site is a generated view of the same sources. See [`docs/DOCUMENTATION_SYSTEM.md`](docs/DOCUMENTATION_SYSTEM.md).

## Project Intelligence

Each completed task should record a compact report and update project-level intelligence: input/output tokens, cost, effort, tests, docs impact, risk, debt and confidence. Dashboards consume the durable JSON data; they are not a separate database. See [`docs/PROJECT_INTELLIGENCE.md`](docs/PROJECT_INTELLIGENCE.md).

## Recovery

A new tool/model should start at `ENTRYPOINT.md`, then `atlas.json`, `PROJECT_STATE.md`, `docs/ATLAS.md`, the active Goal and only relevant linked knowledge.

```bash
atlas snapshot ../my-project
```

Snapshots exclude runtime/cache state.

## Repository map

- `src/project_atlas/` — CLI and framework implementation.
- `src/project_atlas/resources/catalog/catalog.json` — canonical workforce registry.
- `src/project_atlas/resources/adapters/` — provider/platform adapters.
- `schemas/` — machine contracts.
- `docs/` — framework protocol and guidance.
- `prompts/` — portable chat workflows.
- `examples/` — project profiles/fixtures.
- `tests/` — resolver, Goals, compiler, validation and migration tests.

Start with [`ENTRYPOINT.md`](ENTRYPOINT.md) if you are an LLM or agent.
