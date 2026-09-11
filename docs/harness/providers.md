# Providers: Model vs Agent + Model Gateway

## ModelProvider (Prumo owns the loop)

Package `internal/harness/model`. Contract: Capabilities, Models,
Stream→EventStream, Cancel, Usage, Health. Normalized events: TextDelta,
ReasoningSummaryDelta, ToolCallDelta/Ready, UsageUpdated, Warning, Error,
Completed, Cancelled.

- `FakeModelProvider`: deterministic scripts (`response|tool call|permission|
  tool result|continuation|complete`), shared conformance suite.
- `OpenAICompat`: first real adapter (`/chat/completions` SSE, tool_calls
  normalization, usage, retryable 429/5xx, cancel via ctx). No vendor types
  escape; swap the library without changing domain contracts.
- `Anthropic`: second real adapter (`/v1/messages` SSE: text deltas,
  `tool_use` blocks with `partial_json` merge, usage, overloaded/rate-limit
  retryability, cancel via ctx). httptest covers text, tool-use merge,
  retryable errors and cancel. Live keys remain environmental
  (`--api-key` / `PRUMO_MODEL_API_KEY`); wire mapping is code-complete.

## AgentProvider (external runtime owns the loop)

Package `internal/harness/extagent`. Contract: Discover/Status/Capabilities/
Models, CreateSession/ResumeSession, Send/Cancel, Approve/Deny, Events, Close.

- `OpenCodeServer`: `opencode serve` HTTP adapter (session/message/permissions/cancel).
- `CodexCLI`: structured `codex exec --json` adapter.
- `FakeAgent`: conformance double (session/events/permissions/usage/cancel/resume).

External state is normalized, never canonical.

## Availability matrix (`prumo agent providers`)

`internal/harness/extagent/probe.go` reports honest availability: binaries +
versions + server reachability, never assumed interop. Measured 2026-09-11
(dev host):

```text
fake             model  available   builtin      deterministic double
openai-compat    model  unconfigured             needs PRUMO_MODEL_BASE_URL
anthropic        model  unconfigured             needs PRUMO_MODEL_API_KEY
opencode-cli     agent  available   1.18.30      binary present
opencode-server  agent  unavailable              nothing on 127.0.0.1:4096
codex-cli        agent  available   0.153.4      binary present
acp-generic      agent  unconfigured             needs PRUMO_ACP_URL
```

Live session interop against those CLIs/servers is the next matrix step;
the adapters + probe are the conformance baseline it will run against.

## Model Gateway (internal, separate from Workforce)

Package `internal/harness/gateway`. `ModelRoute/RouteTarget/ProviderHealth/
QuotaState`, healthy-first deterministic routing, retry, circuit breaker
(3 failures → open + 30s cooldown), cost/latency/capability/privacy inputs.

Rule: transparent fallback ONLY before observable side effects; after side
effects the gateway refuses and the caller must do an explicit Handoff.
