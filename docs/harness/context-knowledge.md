# Context v2, Knowledge, Documentation Compiler

## Context Compiler v2 (`internal/harness/contextv2`)

Pipeline: authority/trust/privacy gates → freshness → exact/structured
retrieval → FTS/BM25 (+optional embeddings + typed graph) → RRF → MMR/dedup
→ dependency/coverage constraints → marginal-utility-per-token packing →
progressive disclosure L0–L4 → replayable Manifest v2 (relevance, authority,
freshness/rev, provenance, token estimate, reason, method, compression,
full-content pointer). No-LLM path is first-class; embeddings are
derived/optional. Extends (not replaces) `internal/contextcompiler` baseline.

## Knowledge Runtime (`internal/harness/knowledge`)

Typed records with stable IDs: Source, Section, Claim, Finding, Decision,
Requirement, Constraint, Risk, Assumption, OpenQuestion, ResearchRecord,
Evidence, MemoryRecord. Each carries authority/trust/freshness/status,
provenance, lineage/supersession, typed relationships. Canonical records stay
simple; indexes/summaries/embeddings are derived.

Knowledge API: Locate/Get/Search/Related/Explain/Context/Affected +
ValidateDelta/CommitDelta/Supersede/Promote/Refresh/Snapshot. Mutation is
`KnowledgeDelta → Validate (policy/review) → Commit → KnowledgeChanged →
derived rebuild` with optimistic-concurrency-friendly supersession.

## Contradiction / Coverage / Readiness

Separate engines. Typed `contradicts` edges (never substring heuristics);
relationship-driven coverage (requirement ← evidences/covers); revision-aware
evidence; readiness policy = no contradictions + no uncovered reqs + no
blocker open-questions. LLM proposes candidates only; it never decides
authority.

## Documentation Compiler (`internal/harness/doccompile`)

Knowledge IR → Document IR → dependency DAG → incremental evaluator →
renderer → canonical serializer → atomic writer. Stable IDs, SHA-256
fingerprints, reverse invalidation via DAG order, red-green digests,
CAS/no-op writes, whole-file generation, atomic replace, artifact manifests,
provenance. AST-aware Markdown regions + JSON Pointer/Patch are the next
increment (interfaces reserved).

## Human Documentation Runtime (HD0–HD4 baseline)

`DocumentationSpec`, profiles, `DocumentationUnit`/Page Graph, deterministic
planner, README/reference generation, basic docs tree, coverage/readiness —
this `docs/harness/` tree + `promotion-report` are the first instance.
Website/Starlight (HD5+) follows the dependency roadmap and never blocks the
Harness.

## Local Intelligence

Decoupled by design: generation/rerank → llama.cpp/GGUF worker, embeddings →
ONNX worker, OpenVINO → optional adapter. Workers are supervised processes,
never mandatory Core bindings. Concrete models stay benchmark-driven;
No-LLM remains first-class.
