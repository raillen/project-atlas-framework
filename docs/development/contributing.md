# Contributing

Thank you for contributing to Project Atlas!

## Development Setup

1. Clone the repository.
2. Run `poetry install` (or equivalent) to install dependencies.
3. Run `pre-commit install` to set up formatting hooks.

## Code Style

Follow PEP 8 for Python code. Use `black` for formatting and `isort` for imports.

## Adding a New Schema

1. Update the schema definition in `src/project_atlas/resources/schemas/`.
2. Generate models if applicable.
3. Add tests in `test_schemas.py`.

## Adding a New Skill, Agent, or Recipe

1. Create the package directory in `src/project_atlas/resources/catalog/`.
2. Write the manifest and markdown file.
3. Ensure conformance tests pass.

## Adding a CLI Command

Add the command to `src/project_atlas/cli/` using the established CLI framework (e.g., Typer or argparse).

## PR Process

- Open a PR with a descriptive title.
- Ensure all CI checks (pytest, linting) pass.
- Request review from core maintainers.

## AGENTS.md Rules

Adhere strictly to the rules defined in the repository's root `AGENTS.md` (e.g., provider-neutrality, Lean Progressive Context).
