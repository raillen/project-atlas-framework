# Incident Response Playbook

Troubleshooting guide for common Project Atlas v0.4 Go CLI failures and recovery procedures.

Use the [installation manual](../manual/installation.md), [uninstallation manual](../manual/uninstallation.md), and [usage manual](../manual/usage.md) for normal operations.

## Quick response matrix

| Symptom | First action |
|---------|--------------|
| `atlas: command not found` | Verify the binary path and `PATH`; see installation manual |
| `project root not found` | Run `atlas status --path <project>` and verify `atlas.json` |
| Validation failure | Run `atlas doctor <project> --json` |
| Goal transition invalid | Inspect Goal state and follow the documented state machine |
| Goal lock mismatch | Stop manual edits; inspect Git and use `goal amend` |
| Compiler failure | Run `atlas framework-check`, then regenerate from canonical resources |
| Global state issue | Use a fresh `--home` for isolation |
| Conformance mismatch | Compare exit code, JSON, stdout/stderr, files, and canonical artifacts |

## 1. Binary and PATH

```bash
command -v atlas
atlas version
atlas --json version
```

For a source build:

```bash
go build -trimpath -o ./atlas ./cmd/atlas
./atlas version
```

For an isolated environment:

```bash
atlas --home ./recovery-home setup
```

## 2. Project diagnostics

```bash
atlas status --path ./project
atlas validate ./project
atlas doctor ./project
atlas --json doctor ./project > doctor.json
```

Do not delete `.ai/`, docs, Goals, Plans, Evidence, or project files to clear a diagnostic. Restore from Git or use the documented migration/snapshot flow.

## 3. Goal failures

Valid lifecycle:

```text
DRAFT → PLANNED → LOCKED → EXECUTING → VERIFYING → REVIEWING → DONE
```

`DONE` requires evidence. A locked Goal cannot be changed directly:

```bash
atlas goal list --path ./project
atlas goal state P00-G01 PLANNED --path ./project
atlas goal amend P00-G01 --file amendment.json --path ./project
```

A lock mismatch indicates that canonical Goal content changed without an approved amendment. Check `git diff` before taking any action.

## 4. Compiler failures

```bash
atlas framework-check
atlas compile --target generic --path ./project
atlas compile --target codex --path ./project
atlas compile --target claude-code --path ./project
```

Canonical catalog and workforce packages live under `src/project_atlas/resources/`. Generated adapters are disposable and should not be hand-edited.

## 5. Global installation recovery

```bash
atlas --home ./clean-home setup
atlas --home ./clean-home install connector opencode
atlas --home ./clean-home uninstall --connectors --purge-cache
```

The `--home` flag prevents recovery tests from touching the normal global state.

## 6. Conformance failure

Record:

- exact Go and Python commands;
- exit codes;
- stdout/stderr;
- parsed JSON difference;
- filesystem difference;
- runtime/tool versions;
- fixture hash.

Python v0.3 remains the oracle until the migration exit criteria are formally met. Do not update golden outputs merely to make CI green.

## 7. Incident evidence

Keep sanitized logs and command output. Do not include secrets, credentials, complete sensitive prompts, or classified repository contents. Use Git history, snapshots, and `atlas doctor --json` as primary evidence.
