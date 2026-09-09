# C3 Safety Rules Checklist

- [ ] All heap allocations are paired immediately with a `defer mem::free(ptr)`.
- [ ] Functions handling sequences accept slices (`type[]`) rather than raw pointers (`type*`).
- [ ] Errors returned via `!Type` are unpacked with `if (catch err = result)` or handled with `try`.
- [ ] Public API functions include explicit doc contracts with `@param` constraints.
