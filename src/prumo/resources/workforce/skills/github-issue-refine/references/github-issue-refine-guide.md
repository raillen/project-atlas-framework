# GitHub Issue Refinement — Technical Reference Guide

## Overview & Purpose
Refine existing issues by clarifying acceptance criteria, technical dependencies, and milestones.

## Core Architecture Principles
1. **Explicit Domain Boundaries**: Align all operations strictly with modular architectural boundaries.
2. **Deterministic Behavior**: Ensure repeatable, verifiable results with zero hidden side-effects.
3. **Defense in Depth**: Validate inputs against canonical schemas before execution.
4. **Lean Context**: Operate only on the minimum required context without speculative expansions.

## Operational Standards
- **Inputs**: Repository context, Issue / PR specifications, Roadmap Goal
- **Outputs**: Structured GitHub artifact, Validation confirmation
- **Required Capabilities**: filesystem.read, filesystem.write, process.spawn
- **Evidence Contract**: test, review

## Common Pitfalls & Anti-Patterns
- Modifying shared state without cryptographic or process locks.
- Suppressing runtime errors or ignoring validation failures.
- Producing unbounded output that violates LPC token limits.

## Recommended References
- Prumo Architecture Blueprint (`docs/architecture/overview.md`)
- Clean Code Engineering Contract (`docs/architecture/clean-code-contract.md`)
- Testing Quality Strategy (`docs/development/testing-strategy.md`)
