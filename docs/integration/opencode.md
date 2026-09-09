# OpenCode Native Integration

This guide details the Project Atlas native harness integration for [OpenCode](https://opencode.ai).

## Overview

Project Atlas integrates natively with OpenCode via the OpenCode Native Harness (`opencode`), providing:
- In-process TypeScript plugin (`.opencode/plugins/atlas.ts`)
- Primary Atlas orchestrator agent (`.opencode/agents/atlas.md`)
- Domain-specific subagents (`architect`, `executor`, `verifier`)
- Pre-tool validation (tool guards) with synchronous veto (`pre_tool_block`)
- Lifecycle session hooks (`session.start`, `session.end`, `tool.before_execute`, `tool.after_execute`)
- OpenCode slash commands (`/goal`, `/plan`, `/trace`, `/experience`, `/adopt`, `/status`)
- Cleanup manifests for zero-residue uninstallation

## Installation

Install the OpenCode native harness into your repository:

```bash
atlas connector install opencode
# or via backward-compatible command
atlas install connector opencode
```

This compiles the `.opencode/` workspace directory:
```text
.opencode/
├── opencode.json             # OpenCode workspace configuration
├── plugins/
│   └── atlas.ts              # TypeScript plugin (tool guards & lifecycle hooks)
├── agents/
│   ├── atlas.md              # Primary Atlas orchestrator
│   ├── architect.md          # Architecture & schema subagent
│   ├── executor.md           # Implementation subagent
│   └── verifier.md           # Quality gate & test subagent
├── skills/
│   ├── code-review/SKILL.md
│   ├── goal-management/SKILL.md
│   └── evidence-collection/SKILL.md
├── guards/
│   └── tool-policy.json      # Tool guard security policies
├── commands/
│   └── atlas.json            # Slash commands
└── .atlas-generated.json     # Ownership marker
```

## Tool Guards

Tool guards enforce the Project Atlas trust model prior to tool invocation (`pre_tool_block`):
- Dangerous shell commands (e.g. `rm -rf /`, `mkfs`) are blocked immediately.
- Sensitive files (`.git/`, `.atlas/credentials`, `.env`) are protected from destructive write operations.
- Operations requiring human confirmation (e.g. `git push --force`) trigger user review.

## Lifecycle Hooks

- **`session.start`**: Injects Lean Progressive Context (LPC) and active Goal details without bloating initial context.
- **`session.end`**: Synthesizes session summaries and emits structured session events into `.atlas/experience/`.
- **`tool.before_execute` / `tool.after_execute`**: Records tool calls and results deterministically for auditability.

## Validation and Cleanup

Validate the native harness installation:
```bash
atlas connector validate opencode
```

Safely uninstall without deleting repository files:
```bash
atlas connector uninstall opencode
```
