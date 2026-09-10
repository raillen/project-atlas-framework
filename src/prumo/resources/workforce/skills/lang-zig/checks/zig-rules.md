# Zig Safety Checklist

- [ ] All functions performing dynamic allocation accept `allocator: std.mem.Allocator`.
- [ ] Every allocated slice/object is released using `defer allocator.free(...)` or `allocator.destroy(...)`.
- [ ] Error paths clean up intermediate allocations using `errdefer`.
- [ ] Test suites execute under `std.testing.allocator` with 0 leak reports.
- [ ] Casts using `@ptrCast` are registered in `.prumo/escape-hatches.json`.
