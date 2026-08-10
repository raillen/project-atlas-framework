# Architecture

Project Atlas separates durable policy from execution implementations.

```text
Project repository
├── Knowledge: ATLAS, docs, ADR/RFC
├── Goals: outcomes, acceptance, gates, evidence
├── AI workforce: selected agents/skills/recipes
├── Model policy: roster, routing, fallbacks
└── Adapters: generated platform-specific instructions
         ↓
Replaceable orchestrator (Traycer / native / other)
         ↓
Agent harnesses and LLM providers
         ↓
Git worktrees / implementation / CI / evidence
```

## Canonical vs generated

Canonical registries are provider-neutral YAML. The compiler turns selected entries into platform-native files such as `.codex/skills/*`, `.claude/agents/*` or a generic context pack. Generated files can be deleted and rebuilt.

## Capability resolution

The resolver uses project type, stack, features and risks. Core roles and skills are always present; bundles add domain defaults; selectors add stack/feature capabilities; risk rules force mandatory checks; skill dependencies are resolved transitively.

## Project state

Every initialized project receives a short `PROJECT_STATE.md` and machine-readable `PROJECT_MANIFEST.yaml`. These are recovery accelerators, not replacements for canonical documentation.

## Extensibility

Project-specific overrides should remain in the project. Proven reusable patterns can be proposed back to Project Atlas through an RFC/change proposal rather than copied blindly into core.
