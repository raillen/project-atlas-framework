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

## Model Gateway (internal, separate from Workforce)

Package `internal/harness/gateway`. `ModelRoute/RouteTarget/ProviderHealth/
QuotaState`, healthy-first deterministic routing, retry, circuit breaker
(3 failures → open + 30s cooldown), cost/latency/capability/privacy inputs.

Rule: transparent fallback ONLY before observable side effects; after side
effects the gateway refuses and the caller must do an explicit Handoff.
