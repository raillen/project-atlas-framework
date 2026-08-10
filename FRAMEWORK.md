# Framework Contract

## Purpose

Project Atlas is a versioned engineering protocol for durable, portable AI-assisted development. Prompts, skills, agents, adapters and orchestrator configurations are distributions of the protocol, not the protocol itself.

## Invariants

- **Git is durable memory.** Important project truth must survive loss of chat history.
- **Project decisions outrank framework defaults.** An explicit project ADR is authoritative for that project.
- **Canonical before generated.** Generated indexes/adapters never replace their canonical source.
- **Goals define completion.** Roadmaps describe direction; locked Goals define measurable done conditions.
- **Acceptance is protected.** Agents may propose amendments but cannot silently weaken a locked Goal.
- **Roles are abstract.** Agents are roles; models are routed to roles by project policy.
- **Skills are selected.** Do not install the entire skill registry into every project.
- **Review diversity is desirable.** High-risk work should use a reviewer from a different provider when practical.
- **Evidence over assertion.** Builds, tests, benchmarks, screenshots, logs and reviews are evidence.
- **Orchestrators are replaceable.** Traycer may be primary without becoming source of truth.
- **Model preferences are per-project.** Bootstrap explicitly obtains the preferred LLM/provider roster each time.
- **Portability first.** A generic Markdown adapter must remain sufficient for platforms unknown to the framework.

## Authority order

1. Explicit current user/project decision.
2. Accepted project ADR/RFC and locked Goal.
3. Canonical project documentation.
4. Project configuration/manifests.
5. Project Atlas Framework protocol.
6. Platform adapter.
7. Generic model/tool defaults.
