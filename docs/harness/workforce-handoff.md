# Handoff v2 + Workforce

## Handoff v2 (`internal/harness/handoff`)

Typed, pointer-first continuation. A bundle carries refs (never a full
transcript) to Run, Goal, Plan, Tasks, Checkpoint, workspace rev, Context
Manifest, Decisions, Requirements, Research, Changes, Evidence, Gates,
Budget, Permissions, Pending Effects + continuation summary. `Validate`
requires run/checkpoint/context_manifest/workspace_rev. Safe-point handoff is
preferred; Planning→Build preserves provenance without turning a
PlanningSession into a Run. After side effects: explicit Handoff, never
silent provider switch.

## Workforce / multi-agent (`internal/harness/team`)

`AgentTeam/AgentRole/AgentBinding` with delegation modes solo, manual,
suggested, bounded-auto. Rules: separate worktrees per concurrent writer,
per-role budgets, least context, explicit ownership. Reviewers get
diff+requirements+evidence — never the full implementer trajectory.
`Validate` rejects two writers on one worktree without policy.
`AllocateWorktree` provisions isolated git worktrees; merge/review/conflict
is an explicit operation. Cross-provider handoff is covered by the eval suite.

## External connectivity

Priorities implemented: OpenCode server, Codex structured CLI, ACP-shaped
`FakeAgent` conformance. Compatibility matrix grows behind the
`AgentProvider` port; canonical state never absorbs external session shape.
