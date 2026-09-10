# ADR 004: Rename Project Atlas Framework to Prumo

# Status

Accepted

# Context

- The project was originally called Project Atlas Framework, with CLI `atlas`, manifest `atlas.json`, runtime directory `.atlas/`, documentation router `docs/ATLAS.md`, home variable `ATLAS_HOME`, and Go module `github.com/raillen/project-atlas-framework`.
- The scope evolved beyond the meaning of a traditional framework. The system is now better described as a Git-native project protocol and CLI for software built by humans and AI agents.
- The project has no meaningful external user base yet. Pre-v1 is the safest point for a complete identity change.
- Versions v0.3 and earlier were released under the Project Atlas name (Python runtime, then Go Core per ADR 001). Historical baselines (`conformance/V03_BASELINE.json`, `conformance/M0_CONFORMANCE_MATRIX.json`) keep the original names as a factual record of that line.

# Decision

Rename everything to:

```text
Prumo
```

Including:

```text
atlas CLI → prumo (cmd/prumo, binary prumo / prumo.exe)
atlas.json → prumo.json
.atlas/ → .prumo/
ATLAS_HOME → PRUMO_HOME
docs/ATLAS.md → docs/PRUMO.md
atlas-navigation skill → prumo-navigation
schemas/atlas.schema.json → schemas/prumo.schema.json
schemas/atlas-package.schema.json → schemas/prumo-package.schema.json
schemas/atlas-migration.schema.json → schemas/prumo-migration.schema.json
ATLAS_* envelope codes → PRUMO_*
atlas_version / atlas_generated fields → prumo_version / prumo_generated
.atlas-generated.json ownership markers → .prumo-generated.json
src/project_atlas/ → src/prumo/
github.com/raillen/project-atlas-framework → github.com/raillen/prumo
```

No compatibility layer is maintained: no `atlas` alias, no dual `atlas.json`/`prumo.json` loading, no parallel `.atlas/` support.

The rename ships as v0.5.0 (pre-v1 minor bump for a breaking identity change). The `framework` object key in `prumo.json` and the `framework-check` command name are kept to avoid unnecessary contract churn; the canonical `framework.name` value is `prumo`.

# Schema `$id` strategy

There is no canonical product domain yet, so schema `$id` values use the repository as the stable namespace:

```text
https://raw.githubusercontent.com/raillen/prumo/main/schemas/<name>.schema.json
```

The `$id` is an identifier, not a remote fetch target: validation resolves `$ref` locally (relative file references plus the documented base). If a dedicated domain is adopted later, the `$id` base can move in a versioned schema migration.

# Rationale

"Prumo" (plumb line) represents alignment, guidance, and keeping a project structurally correct relative to a reference. This fits the project's role in coordinating project state, agents, goals, plans, evidence, documentation, policies, automation, and harness integrations. It is shorter to type, easier to brand, and removes the generic "framework" terminology while the architecture and protocol semantics stay unchanged.

# Consequences

Positive:

- stronger identity;
- shorter CLI;
- easier branding;
- removes generic "framework" terminology;
- matches the actual architectural role.

Negative:

- existing local projects need regeneration or manual rename (`prumo.json`, `.prumo/`, `docs/PRUMO.md`);
- old commands, paths, imports, and schema `$id` values no longer work;
- connector output regenerated from canonical resources.

# Compatibility

None intentionally provided because the software is pre-v1 and currently has no relevant external user base. Version-to-version project migrations (v1/v2/v3 manifest upgrades) are unaffected; only the identity carrying them changed.
