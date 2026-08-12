# Migration v0.1 → v0.2

v0.2 adopts Lean Progressive Context and reduces the canonical maintained formats to Markdown + JSON.

## Major project file changes

| v0.1 | v0.2 |
|---|---|
| `.atlas/project-profile.yaml` + `PROJECT_MANIFEST.yaml` | `atlas.json` |
| `.ai/**/manifest.yaml` | `.ai/**/manifest.json` |
| `.ai/orchestration/*.yaml` | `.ai/orchestration/*.json` |
| `.ai/goals/**/*.goal.yaml` | `.ai/goals/**/*.goal.json` |
| persistent generic compiled `CONTEXT.md` | runtime `.atlas/runtime/compiled/...` |
| project YAML catalogs/config | JSON |

## Compatibility policy

v0.2 may read legacy YAML for migration. It must not generate new YAML canonical files.

PyYAML may remain a temporary compatibility dependency during the v0.2 transition.

## Migration procedure

1. Create a clean branch.
2. Preserve canonical Markdown docs/ADRs/specs unchanged unless their content is obsolete.
3. Merge project profile + manifest truth into `atlas.json`.
4. Convert AI manifests/policies/Goals to JSON.
5. Remove generated context packs that duplicate canonical knowledge.
6. Ensure `.atlas/runtime/` and `.atlas/cache/` are gitignored.
7. Add/initialize compact Project Intelligence data if desired.
8. Update `ENTRYPOINT.md`/`docs/ATLAS.md` to LPC/PCA routing.
9. Validate.
10. Compile required platform adapters again.

## Do not migrate by

- converting Markdown documentation to JSON;
- creating a JSON duplicate for every Markdown file;
- persisting generated summaries;
- splitting every document into microfiles;
- enabling recursive agent execution by default.

## After migration

Use Progressive Context:

```text
ATLAS/Goal
→ minimum context
→ structural/known pack
→ targeted expansion
→ bounded delegation if needed
→ documentation delta
→ intelligence
→ GC
```
