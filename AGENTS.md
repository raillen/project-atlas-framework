# AGENTS.md

This repository develops Project Atlas itself.

- Start with `ENTRYPOINT.md`, `FRAMEWORK.md`, and `docs/ATLAS.md`.
- Preserve provider-neutral core policy.
- Follow Lean Progressive Context: smallest sufficient context, progressive expansion, bounded output.
- Treat `src/project_atlas/resources/catalog/catalog.json` as the canonical workforce registry.
- Platform-specific behavior belongs in adapters, not core policy.
- New human-maintained project artifacts use Markdown or JSON only.
- YAML support is legacy read compatibility; do not generate new YAML artifacts.
- Derived indexes, summaries and Working Context Capsules belong in runtime/cache and must not become canonical files.
- Add/adjust JSON Schema whenever machine contracts change.
- Add tests for resolver, Goals, compiler, validation, scaffolding and migration behavior.
- Keep `ENTRYPOINT.md` and generated adapters short.
- Do not encode one project's preferred model roster as a framework default.
- Do not make any orchestrator/model provider mandatory for the core framework.
- New context mechanisms must report token/cost impact and have a stopping condition.
- Deep recursive execution remains experimental and disabled by default.
