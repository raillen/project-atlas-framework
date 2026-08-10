# Project Atlas Framework — Agent Entrypoint

If you are an AI system using this framework:

1. Read `FRAMEWORK.md` for invariants.
2. Read `docs/ARCHITECTURE.md` for the protocol model.
3. For an actual project, read its `PROJECT_MANIFEST.yaml`, `PROJECT_STATE.md` and `docs/ATLAS.md`.
4. Read the active Goal and dependencies before making changes.
5. Use project manifests to load selected Agents, Skills and Recipes; do not activate the full registry indiscriminately.
6. Project-specific explicit decisions and ADRs outrank generic framework defaults.
7. Never silently weaken locked acceptance criteria.
8. Do not inherit the LLM roster from another project. Every project explicitly defines its preferred models/providers.
9. Treat orchestrators (Traycer, Codex, Claude Code, etc.) as replaceable execution adapters.
10. CI, tests and recorded evidence—not agent confidence—determine completion.

Portable conversation commands are documented in `docs/USAGE.md` and `prompts/`.
