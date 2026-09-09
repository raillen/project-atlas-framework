# Installing Atlas v0.4

Atlas v0.4 distributes a single static binary. Python is not required to run the Go CLI.

## Development Build

```bash
go build -o /tmp/atlas ./cmd/atlas
/tmp/atlas --json version
```

## Release Build

```bash
VERSION=0.4.0 sh scripts/release.sh
ls dist/
cat dist/checksums.txt
```

## Setup and Portable Home

```bash
atlas --home ~/.atlas setup
atlas --home ./project-home setup
ATLAS_HOME=./project-home atlas setup
```

Setup records the installation manifest and detects available harnesses. Running setup twice converges to the same manifest.

## Install Connectors

```bash
atlas --home ~/.atlas install connector opencode
```

Connector state lives under `ATLAS_HOME/connectors/<id>/cleanup.json`.

## Uninstall Safely

```bash
atlas --home ~/.atlas uninstall
atlas --home ~/.atlas uninstall --connectors
atlas --home ~/.atlas uninstall --purge-cache
atlas --home ~/.atlas uninstall --purge-global-config
```

Uninstall never deletes repository data: `.ai/`, docs, goals, plans, evidence, and all project files remain untouched. Purge flags only remove cache, derived runtime state, or global configuration.
