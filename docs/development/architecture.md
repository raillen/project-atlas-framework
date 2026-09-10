# Architecture

Prumo is built around a schema-first, compiler-driven architecture.

## High-Level Architecture Diagram

```text
User/Config -> [ CLI ] -> [ Scaffold / Init ]
                              |
                        [ Resolver ] <--- (Registry/Manifests)
                              |
                        [ Compiler ] ---> [ Target Adapters (Codex, Claude, etc.) ]
                              |
                        [ Workforce / Goals ]
```

## Module Responsibilities

- **cli:** Command-line interface entrypoints.
- **compiler:** Translates resolved states into target-specific artifacts (e.g., `.codex/`).
- **resolver:** Determines which skills, agents, and recipes apply based on project context.
- **validator:** Ensures manifests and configs match JSON schemas.
- **scaffolder:** Generates boilerplate for new projects.
- **workforce:** Manages agent definitions and rosters.
- **goals:** Defines high-level objectives.
- **context:** Manages environment and project state.
- **doctor:** Diagnostics and troubleshooting.
- **explain:** Provides traces for resolver decisions.
- **migration:** Upgrades configurations between schema versions.
- **fake_runtime:** A deterministic test environment for skills.

## Data Flow

`profile` -> `resolve` (filter applicable resources) -> `scaffold` (setup project) -> `compile` (generate platform-specific instructions).

## Schema-First Development

All configurations (skills, agents, recipes) are backed by strict JSON schemas. Changes to the domain model require schema updates first.

## Canonical vs Generated Content

Human-maintained (canonical) content lives in `src/prumo/resources/` or project config. Generated content (e.g., `.codex/`) is ephemeral and should not be edited manually.

## Protocol Versioning

Manifests use semantic versioning and explicit schema versions (`schema_version: "2.0"`).

## Key Design Decisions

- **Provider-Neutrality:** The core framework does not depend on any specific LLM or orchestrator.
- **Deterministic Resolution:** Given the same inputs, the resolver always produces the same output.
