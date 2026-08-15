# Project Bootstrap

## Purpose
Initialize Project Atlas configuration, schema validation, and minimal workforce for a new repository.

## Required Inputs
- Project profile JSON
- Target directory

## Step DAG & Dependencies
1. **Capture project preferences and stack** (`profile`)
   - **Role:** `explorer`
   - **Skills:** `atlas-navigation`
   - **Input:** Project profile
   - **Output:** Validated ProjectProfile object
   - **Evidence Required:** `review`
   - **Dependencies:** None (entry step)
2. **Resolve smallest useful workforce** (`resolve`)
   - **Role:** `architect`
   - **Skills:** `clean-code`
   - **Input:** ProjectProfile
   - **Output:** Resolution plan
   - **Evidence Required:** `review`
   - **Dependencies:** `profile`
3. **Scaffold canonical directory layout and atlas.json** (`scaffold`)
   - **Role:** `implementer`
   - **Skills:** `documentation`, `goal-management`
   - **Input:** Resolution plan
   - **Output:** Scaffolded workspace
   - **Evidence Required:** `test`
   - **Dependencies:** `resolve`
4. **Validate generated configuration against schemas** (`validate`)
   - **Role:** `tester`
   - **Skills:** `testing-quality`
   - **Input:** Scaffolded workspace
   - **Output:** Validation pass report
   - **Evidence Required:** `test`
   - **Dependencies:** `scaffold`
   - **Gates:** `tests`

## Fallback Strategy
- On `step_failed`: **retry_with_alternate_model** (Max Retries: 2)
- On `gate_failed`: **escalate_to_human**

## Required Gates & Evidence
- **Gates:** `tests`
- **Artifacts:** `atlas.json`, `docs/ATLAS.md`, `.ai/ manifests`, `Initial Goal`

## Stop Conditions
- Project initialized and validated with atlas doctor
