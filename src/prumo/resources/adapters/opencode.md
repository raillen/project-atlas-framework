# OpenCode Native Adapter

This project integrates with OpenCode via the Prumo Native Harness.

- Read `ENTRYPOINT.md`, `prumo.json`, the active Goal and `docs/PRUMO.md`.
- Use Lean Progressive Context: smallest sufficient context, progressive expansion, pointer over payload.
- Do not scan/read the entire repository by default.
- Respect context/output budgets and stop when evidence is sufficient.
- Keep delegation bounded; deep recursion is disabled unless explicitly configured.
- OpenCode native plugin `.opencode/plugins/prumo.ts` intercepts tool calls with tool guards.
- Session start and end lifecycle hooks synchronize context and experience events.
- Specialized subagents (`architect`, `executor`, `verifier`) are defined under `.opencode/agents/`.
- Custom commands are registered in `.opencode/commands/prumo.json`.
