# Contributing

The short contribution guide is maintained at [`CONTRIBUTING.md`](../../CONTRIBUTING.md). This page contains repository-internal engineering rules.

## Repository entrypoints

Read `AGENTS.md`, `ENTRYPOINT.md`, `FRAMEWORK.md`, and `docs/ATLAS.md` before changing behavior. New work must preserve the provider-neutral core policy and Lean Progressive Context discipline.

## Development setup

```bash
git clone git@github.com:raillen/project-atlas-framework.git
cd project-atlas-framework
go test ./...
go vet ./...
gofmt -l cmd internal embedded_assets.go
```

The Go test suite includes differential conformance against Python v0.3. Python remains required only for the oracle and Python test context:

```bash
python -m pip install -e '.[dev]'
pytest
```

## Code and package conventions

- Follow [coding standards](coding-standards.md).
- Respect [dependency rules](../architecture/dependency-rules.md).
- Do not introduce `utils`, `common`, `helpers` or other semantic-free packages.
- Keep the CLI thin: business behavior belongs in application services, not `main.go`.
- Return explicit, wrapped errors; never use `panic` for user/config failures.
- Treat JSON key order as an output-serialization detail; compare generated JSON semantically in conformance tests.
- Keep `embedded_assets.go` at the module root because Go `//go:embed` cannot traverse outside `internal/`.

## Schemas, workforce, and CLI

- Update the canonical schema in `schemas/` before changing a machine contract.
- Keep agents, skills, and recipes under `src/project_atlas/resources/` until the migration explicitly moves canonical content.
- Add differential tests for resolver, Goals, compiler, validation, scaffolding, migration, doctor, and explain behavior.
- Validate all compiler targets after changing shared compiler logic.
- Run `atlas framework-check` behavior coverage after catalog/workforce changes.

## Migration discipline

- Do not delete, move, or rewrite the Python v0.3 runtime while it remains the conformance oracle.
- Do not change the Atlas protocol only to make Go implementation easier.
- New v0.4 capabilities require closed Goals, documentation, and acceptance criteria.
- Lock content, ADRs, and schemas must change explicitly, not silently.

## Pull request process

- Keep changes review-sized and scoped to one Goal or milestone.
- Report `go test -race ./...`, `go vet ./...`, `gofmt -l`, and `pytest` results.
- Record protocol-affecting decisions in the canonical document and ADR first.
- Request review from maintainers before merging protocol, security, or release changes.
