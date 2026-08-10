# Agents, Skills and Recipes

## Taxonomy

- **Agent** — who performs work: role, responsibilities, permissions.
- **Skill** — what capability is available.
- **Recipe** — how a category of work should proceed.
- **Goal** — which measurable outcome must be achieved.
- **Model Policy** — which configured LLMs may perform roles and fallback order.
- **Orchestrator** — what coordinates execution.

## Selection rule

Do not load every skill/agent into every project. The resolver composes the smallest useful workforce from core + project-type bundles + stack + features + risk requirements + project overrides.

## Why roles are model-independent

`renderer-engineer` remains `renderer-engineer` whether a project routes it to GPT, DeepSeek, Kimi, GLM, Claude or another model. This makes model replacement and scorecard-driven routing possible without rewriting project methodology.

## Risk-mandated capabilities

Some domains force skills. Public web projects require security checks; performance-critical systems require benchmarks; multiplayer requires network testing/security; explicit accessibility requirements activate accessibility/visual checks.

## Registry evolution

Add generic capabilities to the framework only when they are reusable. One-off project instructions belong in the project profile/overrides.
