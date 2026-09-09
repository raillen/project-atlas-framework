# Monitoring Guide

Monitor the Go v0.4 CLI through deterministic health checks, exit codes, JSON envelopes, and repository state.

## Framework health

```bash
atlas framework-check
atlas --json framework-check
```

The command validates canonical adapters, schemas, catalog references, and workforce packages.

## Project health

```bash
atlas validate ./project
atlas doctor ./project
atlas --json doctor ./project
```

Run after Goal transitions, migrations, compiler changes, and before release. Treat `ERROR` findings as blockers. Warnings require review according to the project risk policy.

## CLI diagnostics

JSON output is stable for automation:

```bash
atlas --json version
atlas --json validate ./project
atlas --json doctor ./project > doctor.json
```

Use stderr for operational diagnostics and preserve stdout when consuming JSON.

## Compiler checks

```bash
for target in generic chatgpt claude kimi codex claude-code traycer; do
  atlas compile --target "$target" --path ./project
  echo "$target: $?"
done
```

Compare generated artifacts against the canonical project and source workforce. Do not hand-edit generated adapters.

## Repository state

```bash
git status --short
du -sh ./project/.ai ./project/.atlas ./project/.codex ./project/.claude 2>/dev/null
```

Investigate unexpected generated files, unbounded `.atlas/runtime` growth, or changes outside the task scope.

## CI health gates

```bash
gofmt -l cmd internal embedded_assets.go
go test ./... -race
go vet ./...
pytest
```

The Python suite remains the v0.3 oracle. It is not the runtime health check for Go installations.

## Suggested cadence

- Per change: `go test ./...`, `go vet ./...`, targeted `doctor`.
- Before merge: race tests, Python oracle, `framework-check`, all compiler targets.
- Before release: release build, checksums, install/setup/uninstall smoke tests, project-data preservation test.
