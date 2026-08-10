# Project Atlas Framework

Project Atlas is a Git-native, provider-agnostic framework for building and maintaining software with humans and AI agents. It makes the repository—not chat memory, a particular LLM, or an orchestrator—the durable source of truth.

It combines six operational layers:

1. **Knowledge** — ATLAS, canonical docs, ADRs/RFCs and relationships.
2. **Goals** — measurable outcomes, acceptance criteria, gates and evidence.
3. **Agents** — abstract roles with responsibilities and permissions.
4. **Skills** — reusable capabilities selected by project type, stack, features and risk.
5. **Recipes** — reusable workflows for common categories of work.
6. **Orchestration** — model routing, fallbacks, cross-provider review and replaceable orchestrators.

## Why

AI-assisted projects fail when decisions live only in chat history, every tool gets a different prompt, agents receive irrelevant context, or completion means "the model said it is done." Project Atlas addresses those problems with explicit project state, versioned protocol files and evidence-based gates.

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

Create a profile (see `examples/brasa/project-profile.yaml`) and run:

```bash
atlas init ../my-project --profile profile.yaml --non-interactive
atlas validate ../my-project --schemas ./schemas
```

Interactive `atlas init` explicitly asks for the LLM roster. **Project Atlas never silently inherits model preferences from another project.**

## Resolve the AI workforce

```bash
atlas resolve profile.yaml
```

The resolver composes:

```text
core
+ project-type bundle
+ stack skills
+ feature skills
+ risk-mandated skills
+ dependencies
+ project overrides
```

## Compile platform adapters

```bash
atlas compile --target codex --path ../my-project
atlas compile --target claude-code --path ../my-project
atlas compile --target traycer --path ../my-project
atlas compile --target chatgpt --path ../my-project
```

Canonical agents/skills remain provider-neutral; compilation produces platform-specific operational files.

## Goals

```bash
atlas goal new P01-G01 "Runtime window" --phase P01 --path ../my-project
atlas goal state P01-G01 PLANNED --path ../my-project
atlas goal state P01-G01 LOCKED --path ../my-project
atlas goal list --path ../my-project
```

A Goal cannot transition directly between arbitrary states, and `DONE` requires evidence.

## Recovery

```bash
atlas snapshot ../my-project
```

A new chat/tool can recover the project by reading `ENTRYPOINT.md`, `PROJECT_MANIFEST.yaml`, `PROJECT_STATE.md`, `docs/ATLAS.md`, the active Goal and linked canonical docs.

## Orchestration

Traycer is the default recommended adapter when it fits the project, but it is not canonical. See [`docs/PROJECT_ORCHESTRATION_PROTOCOL.md`](docs/PROJECT_ORCHESTRATION_PROTOCOL.md).

## Repository map

- `src/project_atlas/` — CLI and framework implementation.
- `src/project_atlas/resources/catalog/` — canonical Agent/Skill/Recipe/Bundle registries.
- `src/project_atlas/resources/adapters/` — platform adapters.
- `schemas/` — machine contracts.
- `docs/` — framework design and usage.
- `prompts/` — portable chat workflows.
- `examples/` — project profiles and fixtures.
- `tests/` — resolver, Goal, compiler and validation tests.

Start with [`ENTRYPOINT.md`](ENTRYPOINT.md) if you are an LLM or agent.
