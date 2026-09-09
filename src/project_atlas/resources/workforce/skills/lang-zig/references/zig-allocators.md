# Zig Allocators & Resource Management Reference

1. **std.mem.Allocator**: Standard interface for all allocations. Pass it down call stacks rather than keeping global allocators.
2. **GeneralPurposeAllocator (GPA)**: Safety-oriented allocator detecting double-free, use-after-free, and memory leaks.
3. **ArenaAllocator**: Bundles multiple allocations into a single lifetime, released together with `arena.deinit()`.
