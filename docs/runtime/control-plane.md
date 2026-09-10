# Agent Runtime Control Plane

## Purpose

The Control Plane is the explicit boundary between Prumo decisions and execution. It makes long-running, multi-agent, multi-provider work bounded, resumable, observable, and policy-governed.

## Planes

| Plane | Responsibility |
|-------|----------------|
| **Knowledge** | Goals, documentation, architecture knowledge, Experience, journal, traceability |
| **Control** | Run, Budget/Cost, Context Compiler, Model Router, Tool Gateway, Automation, Environment, Observability |
| **Quality** | Test Providers, evidence, gates, security, independent verification |
| **Execution** | Models, tools/MCPs, sandboxes/runtimes, external providers |
| **Integration** | OpenCode, Codex, Claude, Gemini, Kiro and other harness adapters |

Control Plane can consult Knowledge and Quality and mediate Execution. Providers and harnesses never own invariants.

## Run Engine

A `Run` executes a Plan/Task and has a persisted lifecycle:

```text
created → planning → ready → running
running → waiting_approval | waiting_resource | retrying | verifying | blocked
verifying → completed | failed
any active state → cancelled
```

Invalid transitions are rejected by Core.

### Checkpoint

A checkpoint must allow reconstruction after process death. It records run/task, phase, workspace hash/ref, completed/pending steps, evidence pointers, pending side effects, budget snapshot, and logical resume token. Checkpoint is distinct from Handoff.

### Retry and Resume

Classify rate-limit, timeout, provider 5xx, tool failure, invalid structured output, policy denied, product/test failure, environment failure, and unknown. Retryability is explicit; retry budget is separate. Resume uses canonical state + checkpoint + valid working state, never private model memory.

Side-effecting actions receive idempotency keys when supported. Ambiguous timeout becomes `unknown_side_effect`; it is not replayed automatically.

Cancellation propagates where possible, preserves partial evidence, and performs safe cleanup. Detect livelock through repeated tools/errors, no repository/evidence delta, cyclic delegation, and token burn without progress.

## Budget Manager

Budget scopes:

```text
user → workspace → project → Goal → run → agent → step → tool call
```

Dimensions: input/output/context tokens, tool-output tokens, LLM/tool calls, money, wall time, concurrency, network bytes, external-source count.

Soft limits reduce optional workforce/context or choose a cheaper route where quality policy permits. Hard limits checkpoint and stop safely. Budget overflow must not destroy continuity.

Pricing metadata is separate from model identity, versioned by date, and estimates state uncertainty. Rate limits classify 429/quota/concurrency responses and apply `Retry-After`, exponential backoff with jitter, queueing, and circuit breakers.

Every Run records initial budget, reservations, usage, avoided overage, cache hits when observable, and downgrade/stop reasons.

## Context Compiler

Input: intent, Goal, Task, risk, model, skills, repository state, canonical knowledge, working state, tool results.

```text
Candidates → Authority → Freshness → Relevance → Deduplicate → Token Cost → Priority Pack → Context Manifest
```

A Context Manifest reports included/excluded sources, estimated tokens, authority/trust, freshness, and selection reason. It never exposes chain-of-thought.

Distinguish physical context window, effective reliable window, and Task Context Budget. A larger model window does not justify sending more context.

Tool outputs use cap + summary + pointer + continuation. Pressure states are `healthy`, `pressure`, `compact`, `critical`; thresholds are model-profiled. Compaction checkpoints structured state for fresh reconstruction. Retrieval order: structural lookup → FTS/graph → embeddings only when benchmarked.

## Tool Gateway

Every mediated action follows:

```text
Agent → Tool Gateway → Policy → Permission → Budget → Sandbox → Tool/MCP
```

Tool descriptors declare id/version/source, read-only/idempotent/side-effecting/destructive status, reversibility, filesystem scope, network egress, credential scope, timeout, output size, trust, environment, and output schema.

Use lazy discovery: load summarized catalog/capabilities first, full schema only when needed. Before important side effects record intent + idempotency key + target; after record outcome/provider receipt.

MCP servers are untrusted by default. Identity, origin, version/provenance, permissions, roots, network policy, and secret scopes require governance.

Default policy examples: read inside root may be automatic; writes depend on permission; writes outside root denied; destructive actions require approval; credential-bearing network calls require egress + secret policy.

## Model Registry and Router

Registry stores provider/model metadata, context/modalities, structured output/tool support, reasoning controls, cache, privacy/data policy, pricing, rate limits, latency class, and Prumo eval history.

Router inputs: task class, complexity, risk, context/tools/modality needs, latency/cost budget, privacy/data class, and eval scores. Output includes route, alternatives, and policy constraints. Route decisions are explainable.

Record model ID/revision and evaluation date. Alias/provider changes require shadow eval and canary before critical promotion. Separate reasoning retry from action retry; never replay side effects during model fallback.

## Execution Environment

Providers declare isolation requirements. Fuzzers, active DAST, kernel tools, drivers, and other privileged tools cannot execute outside an authorized environment. Safe mode can disable destructive shell, external plugins, and writes outside root.

## Automation

```text
event → rule → conditions → actions → evidence
```

Automation contracts declare trigger, conditions, actions, approval, timeout, retry, concurrency key, budget, environment, failure policy, and ownership. Every execution has run/idempotency key. Exhausted/non-retryable failures go to DLQ with event, failure, attempts, and recovery hints.

Start event-driven; a daily/weekly scheduler is deferred until stable.

## Observability

Structured telemetry records events, model calls, token/cost usage, context manifests, selected skills/workforce, tools, policy decisions, environment, evidence, retries, failures, and checkpoints. It must explain what ran, why by rule/policy, context, budget, providers, evidence, and resume path — without persisting chain-of-thought.

Local-first storage uses SQLite/derived artifacts. OTLP is optional. Incident bundles are sanitized: versions, config hashes, connector/model IDs, environment fingerprint, events, and failures; secrets and classified content are redacted. Full prompts/outputs are not stored by default.

## Package and Runtime Manager

Prumo Packages initially include skill, connector, test-provider, experience-provider, and publication-adapter. Manifests declare protocol range, platform, dependencies, permissions, runtime requirements, checksum/signature/provenance, and install/cleanup strategy.

`prumo.lock` records resolved versions and integrity for reproducible capability sets. Auxiliary runtimes live under `$PRUMO_HOME/runtimes/...`, isolated from projects. Installation owner (Homebrew, WinGet, script, distro, dev build) is recorded; self-update respects package-manager ownership.

## Migration and Review

Migrations declare id, from/to version, preconditions, backup, transform, post-validation, reversibility, rollback, and affected canonical artifacts. They are deterministic, fixture-tested, and support `--dry-run` where possible.

Unified Review Queue covers Experience/Skill proposals, documentation contradictions/patches, security waivers, migration proposals, and publication approvals. Human approval is evidence/event, not invisible state mutation.

## OPEN QUESTIONS

- [ ] Exact persistent event/checkpoint schemas and versioning
- [ ] Initial local storage format before SQLite derived index is introduced
- [ ] Concrete sandbox implementations for each supported platform
- [ ] Budget threshold defaults and model-profile definitions
- [ ] First event sources for Automation MVP