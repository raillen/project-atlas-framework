# C3 Safety & Memory Model Reference

1. **Defer Semantics**: `defer` runs at the close of the surrounding scope in reverse order of declaration.
2. **Slices**: Bounded views with pointer and length. Out-of-bounds indexing in debug builds triggers a panic.
3. **Optional Error Handling**: C3 avoids C-style ambiguous return values by introducing native optional error types (`!int`).
