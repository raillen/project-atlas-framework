# Lean Progressive Context (LPC)

**Lean Progressive Context (LPC)** is the Prumo methodology for minimizing LLM context and output while preserving task correctness. Its implementation architecture is the **Progressive Context Architecture (PCA)**; the runtime component is the **Progressive Context Engine (PCE)**.

## One-sentence definition

Prumo LPC/PCA compiles canonical project knowledge into small, temporary task contexts, expands them only when evidence requires it, bounds both input and output, and discards operational context after durable knowledge has been promoted.

## Why

Large context windows do not eliminate context degradation. Loading a whole repository, all docs, old discussion and every rule wastes tokens and increases the chance of conflicting/stale information.

LPC optimizes for **minimum sufficient context**, not maximum available context.

## Core flow

```text
Task
 ↓
classify scope/risk
 ↓
small initial context
 ↓
WCC (Working Context Capsule)
 ↓
enough evidence? ── yes ──→ implement
       │
       no
       ↓
targeted expansion
       ↓
optional isolated child (depth 1)
       ↓
implement → test → docs delta → intelligence → context GC
```

## Context strategies

| Level | Strategy | Typical use |
|---|---|---|
| L0 | Direct | typo, rename, small localized change |
| L1 | Structural/symbol retrieval | known code symbol/module |
| L2 | Context Pack/task map | recurring known task |
| L3 | Progressive retrieval | ambiguous bug/feature impact |
| L4 | Isolated delegation | bounded specialized question |
| L5 | Recursive experiment | disabled by default |

## Working Context Capsule (WCC)

The WCC is rolling state, not an append-only transcript.

Recommended shape:

```text
TASK
goal:
scope:
risk:

CURRENT
...

RULES
...

POINTERS
DOC:...
SRC:...
TEST:...

EVIDENCE
...

OPEN
...

REJECTED
...
```

Rules:

- target a small token budget;
- replace stale information rather than appending forever;
- keep rejected hypotheses only when they prevent repeated investigation;
- reference canonical sources with pointers;
- do not create a permanent `CONTEXT.md` per task.

## Pointer over payload

Prefer:

```text
DOC:ARC-017#invariants
SRC:Timeline.cs#AddClip
TEST:TransitionTests
ADR:ADR-004
```

Resolve the payload only if required.

## Semantic fragmentation

Physical documentation should be organized for humans. The context engine should create **virtual chunks** from headings, code blocks, symbols, contracts and tests.

One `timeline.md` may expose virtual units such as:

```text
timeline#commands
timeline#persistence
timeline#invariants
```

No extra files are required.

## Input economy

Required techniques:

- retrieve progressively;
- filter by task/risk;
- use symbol/structural lookup before broad search;
- deduplicate;
- exclude build/generated/vendor directories;
- reuse cached derived indexes;
- use deltas within an active session;
- promote only relevant invariants/rules;
- stop retrieval when evidence is sufficient.

## Output economy

Output is a budget too.

Agents should:

- avoid narrating every exploration step;
- return compact findings/evidence from delegated work;
- prefer patches over rewriting documents;
- update docs after behavior stabilizes rather than at each microstep;
- store large intermediate artifacts externally and return a pointer/summary;
- avoid generating summaries that no consumer needs.

A child task should normally return a compact finding, evidence and confidence, not an essay.

## Budgets

Projects define target/hard limits in `prumo.json`.

Typical dimensions:

- context/input tokens;
- output tokens;
- retrieval tokens;
- expansion rounds;
- delegated calls;
- elapsed time;
- monetary cost.

Budgets are baselines, not excuses to omit necessary evidence. Exceeding a soft limit should trigger reevaluation; hard limits require escalation/fallback policy.

## Stopping

Stop expanding when:

- evidence is sufficient;
- critical contradictions are resolved;
- applicable tests/contracts are known;
- no critical open question remains;
- marginal expected value is lower than added cost.

## Context garbage collection

After a completed task:

- stable reusable knowledge → canonical docs;
- architectural decision → ADR/RFC;
- accepted debt → debt/project intelligence;
- durable evidence/cost → project history;
- remaining WCC/retrieval/temporary artifacts → delete.

Failed/interrupted tasks may keep a compact recovery capsule for a configurable TTL.

## Recursion policy

Prumo does not depend on Recursive Language Models.

- depth 0/external context ideas are compatible with LPC;
- single-level isolated delegation is allowed when bounded;
- deep recursion is experimental, disabled by default and requires a project ADR/benchmark before broad use.

## Model/rendering independence

Canonical files do not change format per model. A Context Renderer may present the same Context IR as compact KV, minimal Markdown or lightweight tagged text if benchmarks show a benefit.

## Metrics

At minimum record:

- available source estimate;
- retrieved tokens;
- injected/input tokens;
- cached tokens;
- intermediate output tokens;
- final output tokens;
- strategy;
- expansions/delegations;
- cost/latency;
- success.

The goal is not simply fewer tokens; it is fewer tokens **at equal or better task success**.
