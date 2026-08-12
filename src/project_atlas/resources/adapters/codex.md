# AGENTS.md

This project uses Project Atlas v0.2.

- Read `ENTRYPOINT.md`, `atlas.json`, the active Goal and `docs/ATLAS.md`.
- Use Lean Progressive Context: minimum sufficient context, progressive expansion, pointer over payload.
- Do not scan/read the entire repository by default.
- Respect context/output budgets and stop when evidence is sufficient.
- Keep delegation bounded; deep recursion is disabled unless the project explicitly says otherwise.
- Implement only locked Goal scope; tests/evidence determine completion.
- Compute Documentation Delta; patch only impacted canonical docs.
- Do not create task-specific `CONTEXT.md`/`SUMMARY.md` files.
- Record compact Project Intelligence where the project workflow supports it.
- Generated/runtime context is not canonical and should be garbage-collected.
