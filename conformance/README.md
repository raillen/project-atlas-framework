# Conformance Testing: Project Atlas v0.3 Oracle vs. v0.4 Go

## Purpose
This directory hosts differential test fixtures, golden references, and validation datasets used to ensure behavioral parity between:
- Project Atlas v0.3 (Python reference oracle)
- Project Atlas v0.4 (Go implementation)

## Structure
- `V03_BASELINE.json`: Metadata defining the verified v0.3 baseline, test suites, and commands.
- `fixtures/`: Test inputs, sample workspaces, and invalid inputs for edge case testing.
- `golden/`: Canonical expected stdout, stderr, and JSON envelope outputs captured from Python v0.3.

## Differential Testing Workflow
When porting a command to Go:
1. Run fixture through Python v0.3:
   ```bash
   atlas <cmd> <args> --json > conformance/golden/<feature>.golden.json
   ```
2. Run identical fixture through Go v0.4:
   ```bash
   go run ./cmd/atlas <cmd> <args> --json
   ```
3. Assert exact match in:
   - Exit code
   - Envelope structure (`ok`, `protocol_version`, `data`, `diagnostics`)
   - Canonical artifacts written to the workspace
