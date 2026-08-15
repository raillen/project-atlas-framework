# Changelog

## 0.3.0 — Execution-Ready Protocol

- **Strict Machine Contracts:** 24 JSON schemas governing Goals v2, Plans, Tasks, Evidence, Gates, Events, Policies (Permissions, Approvals, Trust, Models, Execution), and Context Packs.
- **Goal System v2:** Cryptographic SHA256 locking, mutation prevention, formal Goal amendments with audit trails.
- **Standardized Workforce Packages:** Modular canonical directory structure for 98 Skills (`manifest.json` + `SKILL.md`), 26 Agents (`manifest.json` + `AGENT.md`), and 14 Recipes (`recipe.json` + `RECIPE.md`).
- **Explainable Workforce Resolution:** Multi-pass deterministic workforce resolver emitting explainability traces and reasons.
- **Platform Adapter Compiler v2:** Skill package compiler injecting complete skill bundles into Codex (`AGENTS.md`, `.codex/skills/`) and Claude Code (`CLAUDE.md`, `.claude/skills/`).
- **Event Protocol & Trust Model:** Structured append-only event stream specification and 3-tier trust hierarchy (trusted policy > contextual data > untrusted data).
- **Diagnostics & Explainability CLI:** `atlas doctor` with comprehensive DAG cycle and lock checks, `atlas explain` covering workforce/agents/skills/recipes/context/models, and `atlas migrate` with automated backup snapshots.
- **Fake Runtime & Conformance Suite:** Deterministic provider-neutral execution simulator and golden test vectors in `conformance/`.
- **Expanded Documentation & Real Examples:** Comprehensive guides across getting-started, user-guide, authoring, integration, reference, and development; real working example projects for `rust-cli`, `react-saas`, `rust-desktop`, `game-engine`, `atlas-flow`, and `conformance-project`.

## 0.2.0 — Lean Progressive Context evolution

- Adopt Lean Progressive Context (LPC) / Progressive Context Architecture (PCA).
- Add explicit input/output token economy, rolling Working Context Capsules, pointer-over-payload and context garbage collection rules.
- Make user, developer, operations and agent documentation first-class framework surfaces.
- Define documentation-site publishing as a derived view of canonical Markdown.
- Add Project Intelligence contract for task/project costs, token usage, effort, quality and debt.
- Reduce maintained project formats to Markdown + JSON; SQLite is runtime/derived state only.
- Migrate framework catalogs and generated project contracts from YAML to JSON.
- Keep YAML as legacy read compatibility for v0.1 migration only.
- Replace persistent generic context packs with runtime/generated context artifacts.
- Add semantic/structural fragmentation guidance instead of physical microfile proliferation.
- Keep deep recursive LLM execution experimental and disabled by default.
- Add migration guidance from v0.1.

## 0.1.0 — Initial implementation

- Git-native framework contract and recovery entrypoint.
- Project profile and workforce capability resolver.
- Agent, Skill, Recipe and project-type Bundle registries.
- Project Orchestration Protocol with per-project LLM selection and fallbacks.
- Goal state machine, acceptance/gate/evidence schema and CLI operations.
- Adapters for generic chat, ChatGPT, Claude, Kimi, Codex, Claude Code and Traycer.
- Platform adapter compiler.
- Project bootstrap, validation, doctor and portable snapshots.
- Initial test suite and GitHub Actions CI.
