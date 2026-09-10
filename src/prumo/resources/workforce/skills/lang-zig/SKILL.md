---
name: lang-zig
description: Zig systems programming, explicit allocator passing (std.mem.Allocator), defer/errdefer cleanup, error union handling, and zero hidden control flow.
---

# Zig Explicit Allocation & Comptime Verification Contract

## 1. Explicit Memory Allocation
- **Zero Hidden Allocations**: No library or helper function shall allocate memory without receiving an explicit `std.mem.Allocator` parameter.
- **Deterministic Cleanup with `defer` & `errdefer`**: Every allocation must be paired with `defer allocator.free(slice)` or `errdefer` for failure rollback.
- **Leak Detection**: Tests must use `std.testing.allocator`, which fails tests if any memory is leaked.

## 2. Control Flow & Error Discipline
- **No Hidden Control Flow**: Zig has no hidden operator overloading or implicit function calls.
- **Error Unions (`!T`)**: Handle errors explicitly using `try`, `catch`, or pattern matching. Do not discard errors with `_ = fallible() catch {}` without documented justification.
- **Comptime Validation**: Leverage `@compileError` and `comptime` checks to assert type invariants before runtime.
