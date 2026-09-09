# Project Atlas Framework

Project Atlas is a Git-native protocol and CLI for software projects built with humans and AI agents.

The repository is the durable source of truth. Atlas stores canonical project state in Markdown, JSON, JSON Schema, and Git; generated adapters, caches, indexes, and runtime state remain derived.

## Current release line

Project Atlas v0.4 is a pure Go distribution (ADR 002).

- **Go v0.4** is the official single-binary CLI and Core implementation.
- Python v0.3 has been completely retired (ADR 002).
- Zero external runtime dependencies (no Python, pip, or virtualenv required).
- All canonical assets (schemas, catalog, workforce, adapters) are embedded directly into the Go binary.

## What Atlas provides

- Project initialization from a profile.
- Deterministic agent, skill, recipe, risk, and model-policy resolution.
- Goal lifecycle with SHA-256 lock integrity and formal amendments.
- Plan DAG validation, Events, Evidence, Gates, and Doctor diagnostics.
- JSON Schema Draft 2020-12 validation with local `$ref` resolution.
- Lean Progressive Context planning and project intelligence reports.
- Compiler adapters for Generic, ChatGPT, Claude, Kimi, Codex, Claude Code, and Traycer.
- Machine-readable JSON envelopes for automation and harness integrations.
- Conformance tests comparing Go behavior with golden specification baselines.
- Portable installation state, connector ownership, setup, and safe uninstall.

## Quick start from source

Requirements:

- Go 1.22+;
- Git.

```bash
git clone git@github.com:raillen/project-atlas-framework.git
cd project-atlas-framework

go run ./cmd/atlas version
go run ./cmd/atlas --json version
```

Expected version output:

```text
0.4.0-dev
```

The CLI currently uses the subcommand surface directly. `atlas --help` is not yet implemented; use the [CLI reference](docs/manual/usage.md#command-reference).

## Initialize a project

Create a profile with at least one preferred model, then initialize a project:

```bash
go run ./cmd/atlas init ./my-project \
  --profile examples/brasa/project-profile.json \
  --non-interactive

go run ./cmd/atlas validate ./my-project
go run ./cmd/atlas doctor ./my-project
```

`atlas init` creates canonical project files such as `atlas.json`, `.ai/`, `docs/ATLAS.md`, `PROJECT_STATE.md`, and `.atlas/history/`. It does not install a harness globally.

## Work with Goals

```bash
go run ./cmd/atlas goal new P00-G01 "Foundation" \
  --phase P00 \
  --objective "Establish the project foundation." \
  --path ./my-project

go run ./cmd/atlas goal state P00-G01 PLANNED --path ./my-project
go run ./cmd/atlas goal state P00-G01 LOCKED --path ./my-project
go run ./cmd/atlas goal list --path ./my-project
```

Locked Goals must be changed through `goal amend`; direct edits are detected by the lock digest.

## Compile a harness adapter

```bash
go run ./cmd/atlas compile --target generic --path ./my-project
go run ./cmd/atlas compile --target codex --path ./my-project
go run ./cmd/atlas compile --target claude-code --path ./my-project
```

Generated artifacts are derived. The canonical project files and workforce packages remain the source of truth.

## Install, setup, and uninstall

Install the latest published development release with checksum verification:

```bash
curl --fail --silent --show-error --location https://raw.githubusercontent.com/raillen/project-atlas-framework/main/scripts/install.sh | sh
```

For a reversible uninstall, review the script before running it:

```bash
curl --fail --location https://raw.githubusercontent.com/raillen/project-atlas-framework/main/scripts/uninstall.sh -o uninstall.sh
sh uninstall.sh --mode pure --dry-run
sh uninstall.sh --mode pure
```

Use a portable installation home when testing or working in CI:

```bash
go run ./cmd/atlas --home ./atlas-home setup
go run ./cmd/atlas --home ./atlas-home install connector opencode
go run ./cmd/atlas --home ./atlas-home uninstall --connectors --purge-cache --purge-global-config
```

Uninstall never removes project files, `.ai/`, docs, Goals, Plans, Evidence, or other repository data. Read the [installation manual](docs/manual/installation.md) and [uninstall manual](docs/manual/uninstallation.md).

## Machine output

Commands that support automation accept `--json`:

```bash
go run ./cmd/atlas --json version
go run ./cmd/atlas --json doctor ./my-project
go run ./cmd/atlas --json framework-check
```

The envelope is:

```json
{
  "protocol_version": "1",
  "ok": true,
  "data": {},
  "diagnostics": [],
  "warnings": []
}
```

stdout is reserved for JSON when `--json` is used. Diagnostics belong on stderr. JSON output contains no ANSI formatting.

## Development checks

```bash
gofmt -l cmd internal embedded_assets.go
go test ./... -race
go vet ./...
```

The Go suite includes comprehensive tests for initialization, resolver behavior, Goals, Plans, Doctor, Explain, migration, snapshots, all seven compiler targets, connectors, and adoption engine.

## Documentation map

- [Install manual](docs/manual/installation.md)
- [Uninstall manual](docs/manual/uninstallation.md)
- [Usage manual](docs/manual/usage.md)
- [CLI reference](docs/manual/usage.md#command-reference)
- [First project](docs/getting-started/first-project.md)
- [Core concepts](docs/getting-started/concepts.md)
- [v0.4 product scope](docs/product/scope-v0.4.md)
- [Architecture](docs/architecture/overview.md)
- [ADR 001: Go Core](docs/adr/001-go-core.md)
- [ADR 002: Retire Python Runtime](docs/adr/002-retire-python-runtime.md)
- [Migration status](docs/migration/v0.3-to-v0.4-go.md)
- [Conformance strategy](docs/migration/conformance-strategy.md)
- [Development blueprint](docs/development/implementation-blueprint.md)
- [Testing strategy](docs/development/testing-strategy.md)
- [Security and trust model](docs/security/trust-model.md)
- [Runtime Control Plane](docs/runtime/control-plane.md)
- [Documentation System v2](docs/governance/documentation-system.md)
- [Documentation source map](docs/SOURCE_MAP.json)

Use `docs/ATLAS.md` as the repository documentation router. Do not load the entire documentation tree for a single task.

## License and contribution

Before contributing, read `AGENTS.md`, the [coding standards](docs/development/coding-standards.md), the [dependency rules](docs/architecture/dependency-rules.md), and the [testing strategy](docs/development/testing-strategy.md).
