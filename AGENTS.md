# AGENTS.md

This repository develops Project Atlas itself.

- Start with `ENTRYPOINT.md`, `FRAMEWORK.md`, and `docs/ATLAS.md`.
- Keep the generic protocol provider-neutral.
- Registries in `src/project_atlas/resources/catalog/` are canonical.
- Platform-specific behavior belongs in adapters, not core policy.
- Add/adjust JSON Schema whenever machine contracts change.
- Add tests for resolver, Goal state machine, compiler and validation behavior.
- Do not encode one project's preferred LLM roster as a framework default.
- Do not make Traycer, Codex, Claude Code, ChatGPT or another platform mandatory for the core framework.
