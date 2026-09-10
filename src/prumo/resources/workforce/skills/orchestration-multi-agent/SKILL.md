# Multi-Agent Orchestration

## Purpose
Plan execution as dependency DAGs, isolate work, and cap retries.

## Use when
Task or Goal requires multi-agent orchestration capabilities.

## Do not use when
Not relevant to active Goal or out of project scope.

## Required context
- Active Goal and Task definition
- Relevant source modules and architecture docs

## Procedure
1. Load minimal relevant context.
2. Plan as a dependency DAG, isolate work, cap retries/delegation depth, diversify review and let CI/evidence decide completion.
3. Execute required verification gates.

## Decision rules
- Favor smallest sufficient changeset.
- Adhere strictly to project architecture.

## Evidence required
- Automated test runs, build outputs, or verification logs.

## Output contract
- Clean, tested implementation or findings report.

## Stop conditions
- Sufficient evidence collected or budget exhausted.

## Escalation rules
- Escalate to lead/human if locked Goal or security boundary is impacted.
