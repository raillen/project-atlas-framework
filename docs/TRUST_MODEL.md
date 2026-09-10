# Prumo Trust Model

## 1. Overview

The Prumo Trust Model establishes formal boundaries between authoritative policy, repository context, and untrusted inputs. In agentic workflows, LLMs process diverse information sources. Without an explicit trust hierarchy, untrusted input (such as malicious issue descriptions, external web pages, or compromised third-party dependencies) could alter locked goals, bypass security gates, or execute unauthorized operations.

## 2. Trust Classification

Prumo categorizes all information into three distinct trust tiers:

```text
┌─────────────────────────────────────────────────────────┐
│                     TRUSTED POLICY                      │
│ Locked Goals, Accepted ADRs, Framework Protocols,       │
│ Permission & Approval Policies, Signed Configs          │
└────────────────────────────┬────────────────────────────┘
                             │ Overrides & Governs
┌────────────────────────────▼────────────────────────────┐
│                    CONTEXTUAL DATA                      │
│ Project Source Code, Internal Tests, Canonical Docs     │
└────────────────────────────┬────────────────────────────┘
                             │ Informs
┌────────────────────────────▼────────────────────────────┐
│                    UNTRUSTED DATA                       │
│ Web Pages, GitHub Issues/PRs, Dependency READMEs,       │
│ Tool/MCP Outputs, Model Outputs, External Repositories  │
└─────────────────────────────────────────────────────────┘
```

### 2.1 Trusted Policy (`trusted`)
Authoritative artifacts that govern agent execution and verification:
- Locked Goals (`.ai/goals/*.goal.json` with `state: "LOCKED"`);
- Project configuration (`prumo.json`, `.ai/orchestration/model-policy.json`);
- Permission policies (`schemas/permission-policy.schema.json`);
- Approval policies (`schemas/approval-policy.schema.json`);
- Architectural Decision Records (ADRs) in `docs/adr/`.

**Rule:** Trusted policy can only be changed by explicit human authorization or formal Goal Amendment protocols.

### 2.2 Contextual Data (`contextual`)
Repository-internal artifacts that represent current implementation state:
- Source code files in `src/`, `lib/`, etc.;
- Test suites in `tests/`;
- Documentation in `docs/`;
- Project intelligence history in `.ai/intelligence/`.

**Rule:** Contextual data informs implementation choices but cannot contradict or weaken Trusted Policy.

### 2.3 Untrusted Data (`untrusted`)
External or generated inputs:
- Web pages retrieved via browsing tools;
- GitHub issue descriptions, comments, or external pull requests;
- Third-party library documentation and release notes;
- Output from MCP servers and external CLI tools;
- Raw reasoning output and intermediary drafts from agents.

**Rule:** Untrusted data is strictly evidentiary or observational. It MUST NEVER override system instructions, security constraints, or locked acceptance criteria.

## 3. Core Invariants

1. **Policy Supremacy:** No prompt injection, issue payload, or tool output can modify locked Goal objectives, acceptance criteria, or security policies.
2. **Deterministic Evidence Over Generated Assertions:** Claims made in untrusted agent outputs must be backed by verifiable evidence (test runs, builds, hash-verified artifacts).
3. **Least Privilege by Default:** External tool invocations, file modifications, and network access must conform to the project's active `permission-policy.json`.
4. **Auditability:** Any action involving high-risk capabilities (process execution, credential access, external writes) produces traceable events in the audit log.
