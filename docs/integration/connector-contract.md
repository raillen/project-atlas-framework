# Project Atlas Connector Contract

This document defines the Connector Contract specification for Project Atlas v0.4 harnesses and integrations.

## Overview

Project Atlas remains strictly provider-neutral and harness-agnostic at its core. Harness integrations (such as Google Antigravity, OpenCode, Codex, Claude Code, and Gemini CLI) interact with the framework via the **Connector Contract**.

## Connector Contract Specification

A connector defines its capabilities, lifecycle hooks, and enforcement primitives machine-readably in conformance with `schemas/connector-contract.schema.json`:

```json
{
  "id": "opencode",
  "version": "0.4.0",
  "protocol_range": ">=0.4.0 <0.5.0",
  "capabilities": [
    "advise",
    "restrict_tools",
    "pre_tool_block",
    "post_tool_verify",
    "isolate_subagents",
    "session_hooks",
    "native_plugin",
    "commands",
    "subagents"
  ],
  "enforcement": "strict",
  "hooks": [
    "session.start",
    "session.end",
    "tool.before_execute",
    "tool.after_execute"
  ]
}
```

## Supported Enforcement Primitives

| Primitive | Description |
|-----------|-------------|
| `advise` | Post-facto recommendation and guidance without execution block. |
| `restrict_tools` | Static denial of prohibited or unapproved tools based on policy. |
| `pre_tool_block` | Synchronous veto evaluating parameters and command before execution. |
| `post_tool_verify` | Post-execution validation of tool outputs and evidence artifacts. |
| `isolate_subagents` | Dedicated context boundary and restricted toolsets for subagents. |
| `session_hooks` | Lifecycle events for session initialization and completion. |
| `native_plugin` | In-process execution plugin (e.g., TypeScript plugin). |

## Lifecycle Operations

1. **Compile (`atlas compile <target>`)**:
   Generates native harness files (e.g., `.opencode/`), ownership markers (`.atlas-generated.json`), and primary/subagent definitions.

2. **Install (`atlas connector install <id>`)**:
   Compiles native artifacts into the active workspace, creates a `CleanupManifest` in `ATLAS_HOME/connectors/<id>/cleanup.json`, and records the installation state in `ATLAS_HOME/config/installation.json`.

3. **Validate (`atlas connector validate <id>`)**:
   Verifies that required native artifacts exist, ownership markers are intact, and configurations match contract expectations.

4. **Uninstall (`atlas connector uninstall <id>`)**:
   Uses the recorded `CleanupManifest` to safely remove managed files without deleting user-created files or project repository data.
