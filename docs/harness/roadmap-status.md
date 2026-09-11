# Harness roadmap status (HA0–HA11) + split gate

Snapshot: `2026-09-11-d41f18bb65f5`. ACCEPTED != implemented. Truth table:

| ID | Scope | Design | Implementation |
|----|-------|--------|----------------|
| HA0 | AgentRuntime contracts/events | ACCEPTED | implemented (`internal/harness/agent` + schemas) |
| HA1 | FakeProvider + conformance + first ModelProvider | ACCEPTED | implemented (Fake + OpenAI-compat + suite) |
| HA2 | Native Agent state machine | ACCEPTED | implemented (reentrant Runner, safe points) |
| HA3 | ToolGateway + Permission integration | ACCEPTED | implemented (ACI catalog + perm engine wired in Runner) |
| HA4 | checkpoint/restart/resume + budgets + observability | ACCEPTED | implemented (atomic store, idempotent journal, budget hook, AgentEvent) |
| HA5 | Coding ACI + Sandbox baseline | ACCEPTED | partial (ACI baseline + path containment + local/worktree env; container/gVisor future) |
| HA6 | 2nd ModelProvider + Gateway/fallback | ACCEPTED | code-complete (openai-compat + anthropic adapters, gateway routing/fallback/CB; live keys environmental) |
| HA7 | 1st external AgentProvider | ACCEPTED | partial (OpenCode server + Codex CLI + FakeAgent; live matrix pending) |
| HA8 | Handoff cross-provider dogfood | ACCEPTED | baseline (typed bundle + CLI; cross-provider e2e pending live creds) |
| HA9 | Context v2 / Knowledge integration | ACCEPTED | baseline (contextv2, knowledge, delta, contradiction/coverage/readiness) |
| HA10 | multi-agent/worktrees | ACCEPTED | baseline (team roles, ownership guard, worktree alloc) |
| HA11 | compatibility/eval matrix | ACCEPTED | partial (unit+conformance+CLI evals; provider-matrix + fuzz pending) |
| HD0–HD4 | Human Docs Runtime baseline | ACCEPTED | partial (this tree + doccompile DAG/CAS; planner/site/i18n future) |
| Local inference | workers (llama.cpp/ONNX) | ACCEPTED | deferred (interfaces reserved; No-LLM first-class) |

## Split gate (`raillen/prumo-code` NOT created)

- [x] HA0 contracts
- [x] HA1 FakeProvider + first ModelProvider
- [x] HA2 Native Agent
- [x] HA3 Tool/Permission integration
- [x] HA4 checkpoint/restart/resume
- [~] HA5 Coding ACI + Sandbox baseline (baseline usable, hardening continues)
- [x] headless coding Run end-to-end (`prumo agent run` → tool → checkpoint)
- [~] versioned public protocol (envelope + 3 schemas + negotiation kernel v0.1.0 + `prumo agent protocol`; full IDL/SDK pending)
- [~] reconnect/replay (resume+handoff CLI done; JSONL event replay via `prumo agent events` done; daemon reconnect pending)

Verdict: NOT READY for split — see `HARNESS_IMPLEMENTATION_REPORT.md`.
The first Prumo Code must build with zero `internal/` imports; that
conformance test is not yet fully wired (protocol IDL + SDK pending).
