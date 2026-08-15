# Server Reconciliation

Server reconciliation is a technique used in authoritative server multiplayer games to handle situations where the client's predicted state diverges from the server's authoritative state.

## Concept

1. The client continuously sends inputs to the server along with a sequence number.
2. The client immediately predicts the outcome of these inputs and updates its local state (Client-side Prediction).
3. The server processes inputs as they arrive and periodically broadcasts its authoritative state, including the sequence number of the last input it processed.
4. When the client receives a server state, it checks if its predicted state matches. If they diverge, a reconciliation occurs.

## The Reconciliation Process

1. **State Snap:** The client overwrites its local state with the server's authoritative state.
2. **Replay:** The client identifies all inputs that have not yet been acknowledged by the server (inputs with sequence numbers greater than the one acknowledged in the server update).
3. **Re-simulation:** The client re-applies these unacknowledged inputs to the new state.

## Edge Cases

- **Floating Point Determinism:** If the client and server simulate physics slightly differently, minor discrepancies can trigger constant reconciliation. Deadzones or thresholds are needed.
- **Visual Snapping:** If reconciliation causes a large positional change, the visual representation should interpolate to the new position over a few frames to hide the snap from the player.
