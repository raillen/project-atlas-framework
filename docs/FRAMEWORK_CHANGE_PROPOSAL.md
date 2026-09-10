# Framework Change Proposal Template

## Problem

What recurring engineering problem was observed?

## Evidence

Which projects, Goals, incidents or benchmarks demonstrate it?

## Existing behavior

What does Prumo currently do?

## Proposed generic rule

State the reusable rule without project-specific names or provider lock-in.

## Context/token impact

If this changes agent context or prompting:

- expected input effect;
- expected output effect;
- cache/retrieval effect;
- stopping condition;
- quality regression risk;
- benchmark plan.

## Persistence/format impact

Does it add a maintained file/format? If yes, why is Markdown/JSON/derived SQLite insufficient?

## Compatibility

Which schemas, catalogs, adapters, generated files or migrations are affected?

## Validation

How will the framework prove the change works and does not regress existing fixtures?
