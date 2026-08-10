# Portability and Platform Adapters

The generic Markdown protocol is the compatibility floor. Platform adapters improve ergonomics but cannot own essential policy.

Supported initial targets:

- `generic`
- `chatgpt`
- `claude`
- `kimi`
- `codex`
- `claude-code`
- `traycer`

Run `atlas compile --target <target>` in an initialized project. Codex and Claude Code receive generated role/skill files; chat platforms receive a compact context pack; Traycer receives an orchestration bridge document.

## Conversation commands

These are portable conventions, not required platform syntax:

- `ProjectAtlas: bootstrap` — structure a new project and gather explicit AI preferences.
- `ProjectAtlas: finalize` — turn approved discussion into canonical project docs/config.
- `ProjectAtlas: update` — incorporate new approved decisions.
- `ProjectAtlas: audit` — check implementation/documentation/protocol drift.
- `ProjectAtlas: recover` — reconstruct project context from Git state.
- `ProjectAtlas: migrate` — apply documented framework version migrations.
