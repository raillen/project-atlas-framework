# Harness daemon (local)

Package `internal/harness/daemon`. A local Unix-socket server hosting
headless runs: `start/status/list/events/cancel/protocol` as JSON lines.

- Run records (`daemon-run-<id>.json`) and the JSONL timeline
  (`events-<id>.jsonl`) persist under the store dir, so clients can
  disconnect, the daemon can restart, and runs stay observable
  (reconnect baseline; remote transport is future work).
- Cancellation is cooperative at state-machine safe points; cancelled runs
  record `cancelled`, permission waits record `yielded`.
- CLI: `prumo agent serve --path . [--socket ...]` blocks until SIGINT/
  SIGTERM; `prumo agent ps` lists runs; `prumo agent logs --run <id>`
  replays the timeline. `serve` uses the Coding ACI workspace; providers
  resolve via `model.ForName` (fake default, real adapters need keys/URLs).
- Tests: lifecycle (start→complete→events→list→protocol), cancel of a
  blocking run, and reconnect (new server, same store).
