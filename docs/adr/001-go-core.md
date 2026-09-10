# ADR 001: Go Core for Prumo v0.4

## Status

Accepted

## Context

Prumo v0.3 is implemented in Python (`src/prumo/`). The v0.4 architecture requires a single-binary, zero-CGO distribution that can run on Linux, macOS, and Windows without external runtime dependencies. Python's interpreter requirement, virtual environment management, and startup latency are blockers for the target user experience.

## Decision

Go will be the official language of the Prumo Core and CLI starting with v0.4. The migration will proceed via conformance-driven parity:

1. Freeze v0.3 observable contracts (exit codes, JSON envelopes, filesystem effects, schema validation).
2. Build Go foundation (module, CLI skeleton, envelope, project discovery, error model).
3. Port protocol domain (Goals, Plans, Tasks, Events, Evidence, Gates, Resolver, Validator).
4. Achieve 100% conformance on critical contracts against Python v0.3 oracle.
5. Port compiler targets (Codex, Claude Code, Generic + ancillary).
6. Deprecate Python runtime only after verified parity.

Python v0.3 remains the executable specification and conformance oracle during migration. No new v0.4 features are implemented in Python.

## Consequences

Positive:
- Single binary distribution via `go build` + `go:embed` for schemas/resources.
- Zero CGO by default preserves cross-compilation and distribution simplicity.
- Standard library covers most needs; external deps only when justified.
- `go test -race`, `go vet`, `staticcheck` provide strong quality gates.

Negative:
- Migration effort: ~17 Python modules → Go packages.
- JSON Schema Draft 2020-12 validation requires external library or Python subprocess delegation during bootstrap.
- Goal digest parity requires exact replication of Python's `json.dumps(sort_keys=True, separators=(',', ':'))` behavior.

## References

- `docs/architecture/overview.md`
- `docs/migration/protocol-inventory.md`
- `docs/migration/conformance-strategy.md`
- `docs/development/phases.md` (M0–M1 gates)