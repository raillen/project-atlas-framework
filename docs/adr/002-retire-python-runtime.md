# ADR 002: Retire Python v0.3 Runtime and Packaging

## Status

Accepted

## Context

ADR 001 established the migration of Project Atlas from Python v0.3 to Go v0.4. Across Milestones M0 through M12, the core framework, CLI operations, protocol domains, compiler targets, connectors, schema validation, doctor diagnostics, and adoption engine were completely implemented in pure Go with zero CGO dependencies. Differential conformance suites verified contract parity across CLI behavior, compiler targets, goal resolution, and schema validation.

Section 6 of `docs/migration/v0.3-to-v0.4-go.md` defines the 7 criteria for Python runtime deprecation and removal:
1. Full Draft 2020-12 schema validation parity verified (satisfied via `internal/validation`).
2. Critical CLI, protocol, filesystem, migration, and compiler contracts pass differential conformance (satisfied via `conformance/golden` and `internal/cliops`).
3. All explain/runtime surfaces implemented in Go (satisfied via `internal/cliops` and `internal/runtime`).
4. Go binaries cross-compile across platforms without external dependencies (satisfied via `scripts/build-release.sh` and pure Go standard library).
5. Installation documentation updated to use Go binary distributions (satisfied).
6. Conformance verified through complete milestone delivery M0–M12 (satisfied).
7. Legacy Python modules, test suite, and packaging classified and retired by explicit decision.

## Decision

Completely remove the legacy Python v0.3 runtime (`src/project_atlas/*.py`), Pytest test suite (`tests/`), and Python packaging configuration (`pyproject.toml`). Project Atlas is now a 100% pure Go distribution.

The canonical resource registries, schemas, and adapters (`src/project_atlas/resources/`, `schemas/`, `adapters/`) are preserved as embedded assets consumed by Go via `embedded_assets.go`.

CI/CD automation (`.github/workflows/ci.yml`) is streamlined to test and validate Go exclusively (`go fmt`, `go vet`, `go test -race`, and CLI smoke tests), eliminating Python environment dependencies.

## Consequences

Positive:
- Zero Python dependencies: no Python interpreter, pip, poetry, or virtualenv required.
- Single self-contained binary distribution with fast startup and instant execution.
- Simplified CI/CD pipeline: single `go-test` job running unit, race, and CLI smoke tests.
- Clean repository structure: no ambiguous duality between Go and Python implementations.
- Full compliance with Project Atlas v0.4 architecture and repository policies.

Negative:
- Python v0.3 oracle tests cannot be run against live Python modules; conformance is preserved via frozen golden fixtures (`conformance/golden/`).

## References

- `docs/adr/001-go-core.md`
- `docs/migration/v0.3-to-v0.4-go.md`
- `docs/architecture/overview.md`
