---
name: multiplayer-networking
description: Server-authoritative replication, delta compression, packet size bounds, reconnection
---
# Multiplayer Networking

## 1. Server Authority
The server is the absolute source of truth. Clients only send input commands, never final state. The server simulates the game logic and replicates the authoritative state back to clients to prevent cheating.

## 2. Client Prediction
Implement client-side prediction to hide latency. The client immediately applies local input to its representation of the entity, predicting the server's response. If the server's replicated state diverges, the client must snap back or smoothly interpolate to the true state.

## 3. State Replication
Optimize state replication. Only replicate entities relevant to a specific client (Interest Management). Use spatial partitioning to determine which entities are in the client's view distance.

## 4. Delta Compression
Conserve bandwidth using delta compression. Replicate only the fields that have changed since the last acknowledged packet. Ensure robust handling of dropped packets when using delta compression over UDP.

## 5. Packet Size Bounds
Strictly bound packet sizes to stay below the MTU (Maximum Transmission Unit, typically ~1200 bytes for safe UDP over internet). Fragmentation causes massive latency spikes; avoid it at all costs.

## 6. Snapshot Interpolation
To render smooth movement of remote entities, buffer incoming server snapshots. Interpolate entity positions between the two most recent snapshots in the buffer. This adds slight delay but ensures smooth visuals despite network jitter.

## 7. Lag Compensation
Implement lag compensation (e.g., Rewind Server Time) for critical interactions like hit detection. The server rewinds the game state to the exact time the client fired the shot, verifying the hit based on the client's perspective.

## 8. Robust Reconnection
Design the architecture to handle seamless reconnections. Clients must be able to drop off and reconnect gracefully, retrieving the full current state without requiring a full game reload.

## 9. Bandwidth Monitoring
Continuously monitor and profile bandwidth usage per client. Implement aggressive culling and level-of-detail strategies for network updates when bandwidth thresholds are approached.

## 10. Network Protocol Security
Encrypt sensitive traffic and implement robust replay protection. Validate all incoming packets strictly. Disconnect clients sending malformed data or exceeding expected packet rates immediately to prevent DoS attacks.

