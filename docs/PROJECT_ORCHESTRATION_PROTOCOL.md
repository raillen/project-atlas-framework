# Project Orchestration Protocol (POP)

POP defines how Project Atlas turns Goals into agentic work while remaining independent of a specific orchestrator.

## Bootstrap rule AI-001

**Never infer a new project's preferred LLM/model roster from another project.** Explicitly obtain or define the roster for every project before generating routing configuration.

## Flow

```text
Vision → Phase → Goal → Task DAG → isolated execution → verification → review → evidence → gate
```

## Roles and models

Roles such as Architect, Implementer, Tester and Security Reviewer are stable abstractions. `model-policy.yaml` maps the project's explicitly selected roster onto roles. Model/provider names must not be embedded into canonical Agent definitions.

## Fallbacks

1. **Availability fallback** — provider/model unavailable.
2. **Quality fallback** — two failed repair attempts cause cross-model escalation.
3. **Confidence fallback** — high-risk design may require independent plans/review.
4. **Human escalation** — fallback exhaustion or policy conflict stops autonomous execution.

Retries are bounded; the orchestrator must not loop indefinitely.

## Review diversity

For high-risk changes, prefer a reviewer from a different provider than the implementer when the configured roster makes this possible. This reduces correlated failure modes; it does not replace tests.

## Work isolation

Independent Goals/tasks should use branches/worktrees or equivalent isolated checkouts. Parallelization is allowed only where dependency analysis indicates safe independence.

## Source of truth

Orchestrator artifacts may mirror Goals but repository Goal files remain canonical. Traycer Smart/YOLO behavior, for example, may propose an amendment but must not silently modify locked acceptance criteria.

## Completion

An agent's declaration of success has no authority. The required gates and evidence recorded in the Goal determine completion.
