# Writing Recipes

Recipes define complex workflows that coordinate multiple agents and skills across a Directed Acyclic Graph (DAG) of steps.

## Recipe Package Directory Structure

```text
recipe-name/
├── recipe.json         # Formal workflow definition
└── RECIPE.md           # Human-readable explanation
```

## Manifest Reference (`recipe.json`)

```json
{
  "id": "fullstack-feature",
  "version": "1.0.0",
  "purpose": "End-to-end implementation of a feature.",
  "preconditions": ["db_running", "repo_clean"],
  "inputs": ["feature_spec"],
  "steps": [
    {
      "id": "design-db",
      "agent": "backend-dev",
      "action": "schema-design"
    },
    {
      "id": "implement-api",
      "depends_on": ["design-db"],
      "agent": "backend-dev",
      "skills": ["api-gen"]
    }
  ],
  "agents": ["backend-dev", "frontend-dev"],
  "skills": ["schema-design", "api-gen"],
  "gates": [
    {"step": "design-db", "type": "human_approval"}
  ],
  "artifacts": ["schema.sql", "api.ts"],
  "fallbacks": [
    {"step": "implement-api", "action": "retry_with_simplified_spec"}
  ],
  "stop_conditions": ["Feature deployed", "Validation failed 3 times"]
}
```

## Defining Step DAGs
Use the `depends_on` array in each step to define execution order. The runner will execute independent steps in parallel.

## Gate Placement Strategy
Place `human_approval` or `automated_test` gates after critical steps (e.g., database schema changes) before proceeding to dependent steps.

## Fallback Configuration
Define fallback actions for flaky steps to ensure resilience (e.g., switching to a smarter model or simplifying the task).
