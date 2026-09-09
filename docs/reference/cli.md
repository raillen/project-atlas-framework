# CLI Reference

The active CLI is the Go binary built from `cmd/atlas`.

```bash
atlas <command> [options]
```

Use the [usage manual](../manual/usage.md) for workflows and examples. This page records the current command surface.

## Global options

| Option | Meaning |
|--------|---------|
| `--json` | Emit the machine-readable envelope on stdout |
| `--home <path>` | Use an isolated global Atlas home |

`atlas --help` is not implemented yet. Unknown commands return exit code `2`.

## Version and project discovery

```bash
atlas version
atlas --json version
atlas status --path <project>
atlas --json status --path <project>
```

## Installation lifecycle

```bash
atlas setup
atlas --home <path> setup
atlas install connector <id>
atlas uninstall
atlas uninstall --connectors
atlas uninstall --purge-cache
atlas uninstall --purge-global-config
```

See [installation](../manual/installation.md) and [uninstallation](../manual/uninstallation.md).

## Project lifecycle

```bash
atlas init <path> --profile <profile.json> --non-interactive
atlas validate [path]
atlas doctor [path]
atlas framework-check
```

`init` requires `--profile` in non-interactive mode.

## Goals

```bash
atlas goal new <id> <title> --phase <phase> [--objective <text>] [--path <path>]
atlas goal state <id> <state> [--reason <text>] [--path <path>]
atlas goal amend <id> [--file <path>] [--reason <text>] [--approved-by <actor>] [--path <path>]
atlas goal list [--path <path>]
```

States: `DRAFT`, `PLANNED`, `LOCKED`, `EXECUTING`, `VERIFYING`, `REVIEWING`, `BLOCKED`, `DONE`.

## Context and intelligence

```bash
atlas context plan <task> [--path <path>] [--json]
atlas report add <report.json> [--path <path>]
atlas report summary [--path <path>] [--json]
```

## Migration and snapshots

```bash
atlas migrate [path] [--dry-run] [--json]
atlas snapshot [path] [--output <archive.zip>]
```

## Documentation deltas

```bash
atlas docs delta propose --goal <goal> [--path <project>] [changed ...] [--json]
atlas docs delta list [--path <project>] [--json]
atlas docs delta show --id <delta> [--path <project>] [--json]
atlas docs delta transition --id <delta> --state <state> [--evidence <id>]... [--path <project>] [--json]
```

Delta states are `proposed`, `reviewed`, `accepted`, `rejected`, and `applied`. Applying a delta requires evidence and does not itself edit canonical documentation.

Migration should be previewed with `--dry-run`. Project data is not removed by uninstall.

## Resolution and explanation

```bash
atlas resolve <profile.json> [--json]
atlas explain workforce <profile.json> [--json]
atlas explain agent <id> [--json]
atlas explain skill <id> [--json]
atlas explain recipe <id> [--json]
atlas explain context <task-id> [--path <project>] [--json]
atlas explain model <role> [--path <project>] [--json]
atlas explain execution <profile> [--path <project>] [--json]
```

## Compiler targets

```bash
atlas compile --target generic [--path <project>] [--json]
atlas compile --target chatgpt [--path <project>] [--json]
atlas compile --target claude [--path <project>] [--json]
atlas compile --target kimi [--path <project>] [--json]
atlas compile --target codex [--path <project>] [--json]
atlas compile --target claude-code [--path <project>] [--json]
atlas compile --target traycer [--path <project>] [--json]
```

## JSON envelope

Successful commands return:

```json
{
  "protocol_version": "1",
  "ok": true,
  "data": {},
  "diagnostics": [],
  "warnings": []
}
```

Exit codes:

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Validation or gate failure |
| `2` | Invalid command or arguments |
| `3` | Configuration error |
| `4` | Capability unavailable |
| `5` | Internal error |
| `6` | Project not found |
