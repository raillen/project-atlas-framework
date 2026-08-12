# Changelog

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
