---
name: lang-go
description: Go idiomatic engineering, explicit error returns, context cancellation, goroutine lifecycle ownership, -race detector, and table-driven tests.
---

# Go Idiomatic Reliability Contract

## 1. Explicit Error Returns
- Every error return must be checked immediately. Never assign to blank identifier `_` for fallible operations.
- Wrap errors with contextual messages using `fmt.Errorf("...: %w", err)`.

## 2. Concurrency & Goroutine Ownership
- Every spawned goroutine must have an explicit owner and a guaranteed termination path (via `context.Context` or channel close).
- All tests must pass with the Go race detector (`go test -race ./...`).
