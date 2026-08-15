# Project Atlas Framework (v0.3)

Project Atlas is an execution-ready, Git-native, provider-agnostic framework and protocol for building and maintaining software with humans and AI agents. The repository—not chat memory, a specific LLM, an orchestrator or a generated documentation site—is the durable source of truth.

Version 0.3 establishes the **Execution-Ready Protocol**: formal JSON schema contracts for all state transitions, Goal v2 cryptographic locking and amendments, modular workforce packages (98 skills, 26 agents, 14 recipes), explainable deterministic resolution, platform compilation for Codex & Claude Code, a 3-tier trust hierarchy, and a deterministic conformance suite with fake runtime simulation.

## Key Capabilities

1. **Strict Machine Contracts** — 24 standard JSON schemas governing Goals, Plans, Tasks, Evidence, Gates, Events, and Policies.
2. **Goal System v2** — Outcome milestones with SHA256 integrity locks, mutation detection, and auditable amendment workflows.
3. **Standardized Workforce Packages** — Self-contained directories containing manifests, instructions (`SKILL.md`, `AGENT.md`, `RECIPE.md`), checks, templates, and scripts.
4. **Lean Progressive Context (LPC/PCA)** — Smallest sufficient context, pointer-over-payload, and bounded token budgets.
5. **Explainability & Diagnostics** — `atlas doctor` for full-project health checks, DAG cycle detection, and `atlas explain` for complete workforce/reasoning transparency.
6. **Platform Compilers** — Automatic adapter compilation for Codex (`AGENTS.md`, `.codex/skills/`), Claude Code (`CLAUDE.md`, `.claude/skills/`), Traycer, ChatGPT, and Generic LLMs.
7. **Conformance Engine** — Deterministic fake runtime and golden test suite in `conformance/`.

## Install

```bash
python -m pip install -e .
atlas --version
```
Expected output: `Project Atlas 0.3.0`

## Start a Project

```bash
# Initialize a project from profile
atlas init ./my-project --profile examples/conformance-project/atlas.json --non-interactive

# Check health diagnostics
atlas doctor ./my-project

# Explain workforce selection
atlas explain workforce --path ./my-project
```

## Manage Goals

```bash
# Create a new Goal
atlas goal new P01-G01 "Core Engine" --phase P01 --objective "Implement core engine features."

# Transition to PLANNED and LOCK acceptance criteria
atlas goal state P01-G01 PLANNED
atlas goal state P01-G01 LOCKED

# Formally amend a locked Goal
atlas goal amend P01-G01 --file amendment.json
```

## Compile Platform Adapters

```bash
atlas compile --target codex
atlas compile --target claude-code
```

## Run Diagnostics & Conformance Suite

```bash
# Test the entire framework and conformance suite
pytest
```

## Documentation
See [`docs/ATLAS.md`](docs/ATLAS.md) for the complete intent router and guides across Getting Started, User Guide, Authoring, Integrations, Protocols, and Reference.
