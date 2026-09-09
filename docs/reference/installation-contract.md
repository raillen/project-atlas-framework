# Atlas v0.4 Installation Contract

This contract defines how the Go binary, global state, and project-local state coexist without destructive side effects.

## Home Directory Layout

`ATLAS_HOME` resolution order:

1. `--home <path>` CLI flag.
2. `ATLAS_HOME` environment variable.
3. `~/.atlas` fallback (`%USERPROFILE%\.atlas` on Windows).

Layout under `ATLAS_HOME`:

```text
ATLAS_HOME/
├── bin/atlas
├── config/installation.json
├── cache/
├── logs/
├── connectors/<id>/cleanup.json
└── runtimes/
```

The process working directory is the only project state touched. Global state never leaks into repositories unless the user explicitly runs a project-scoped command.

## Installation Manifest

`ATLAS_HOME/config/installation.json`:

```json
{
  "atlas_version": "0.4.x",
  "binary_path": "...",
  "connectors": {},
  "created_paths": [],
  "managed_config_fragments": []
}
```

`created_paths` records only files and directories created by installation. User-owned files are never recorded as managed.

## Cleanup Manifest

Each connector records:

```json
{
  "connector": "opencode",
  "scope": "global",
  "created_paths": [],
  "managed_fragments": [],
  "backups": []
}
```

Uninstall removes only recorded paths and fragments. Any path containing user modifications not produced by Atlas is reported as a leftover instead of deleted.

## Safety Rules

- Default uninstall never deletes repository data: `.ai/`, docs, goals, plans, evidence, or any project file.
- `--purge-cache` removes only cache, derived databases, and logs.
- `--purge-global-config` removes only configuration under `ATLAS_HOME/config`.
- `--connectors` removes only connector state tracked in cleanup manifests.
- `setup` never writes outside `ATLAS_HOME`, except when compiling project-local adapters into the current repository.
- Install and setup are idempotent: repeated execution converges to the same manifest.

## Release Distribution

Releases publish checksum files and a manifest with platform binaries:

```text
dist/
├── atlas-linux-amd64
├── atlas-linux-arm64
├── atlas-darwin-amd64
├── atlas-darwin-arm64
├── atlas-windows-amd64.exe
├── atlas-windows-arm64.exe
├── checksums.txt
└── release.json
```

Homebrew and WinGet/Scoop channels consume the same signed artifacts. The install script verifies checksum before replacing the binary atomically.
