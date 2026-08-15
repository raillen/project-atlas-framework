# GitHub CI Debugging

## Purpose
Diagnose and fix GitHub Actions workflow failures, matrix mismatches, and flaky pipeline steps.

## Use when
- Active Goal or Task explicitly requires github ci debugging operations.
- Operating in mode(s): `implementation`, `testing`.

## Do not use when
- Task is out of scope or unrelated to github ci debugging.
- Bounded token budget or capability policy denies required operations.

## Required context
- Repository context
- Issue / PR specifications
- Roadmap Goal

## Procedure
1. Resolve repository context and active Goal requirements for github ci debugging.
2. Ensure compliance with project GitHub Governance policies.
3. Format markdown according to standard template schemas.
4. Verify links, labels, milestones, and cross-references.
5. Commit or publish the validated artifact.

## Decision rules
- Always prefer smallest sufficient changeset over broad refactoring.
- Preserve existing public contracts and architecture invariants.
- Never weaken locked Goal acceptance criteria.

## Evidence required
- test
- review

## Output contract
- Structured GitHub artifact
- Validation confirmation

## Stop conditions
- Task acceptance criteria satisfied with evidence
- Token budget exhausted
- Blocked on external dependency

## Escalation rules
- Escalate to lead architect or human if locked Goal criteria cannot be met.
- Escalate immediately upon discovering unexpected security vulnerabilities or data loss risks.
