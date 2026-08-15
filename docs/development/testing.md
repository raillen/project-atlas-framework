# Testing

Project Atlas emphasizes rigorous testing to ensure deterministic behavior and backward compatibility.

## Test Suite Organization

- `test_schemas.py`: Validates schema definitions.
- `test_goals.py`: Goal evaluation logic.
- `test_resolver.py`: Resolution algorithm correctness.
- `test_scaffolder.py`: Project initialization.
- `test_conformance.py`: Validates skills against contracts.
- `test_workforce.py`: Agent/Workforce definition logic.
- `test_cli.py`: CLI command integration tests.
- `test_migration.py`: Schema migration tools.

## Running Tests

Run `pytest` from the repository root. Use `pytest -v` for verbose output.

## Conformance Suite

Ensures that provided skills adhere strictly to the schema and structural rules (e.g., verifying `SKILL.md` sections).

## FakeRuntime

A deterministic execution environment used to test agent interactions and tool calls without invoking real LLMs or touching the live filesystem unnecessarily.

## Golden Fixtures

Tests relying on complex compilation outputs use golden fixtures stored in `tests/fixtures/`. Update these carefully when compiler output intentionally changes.
