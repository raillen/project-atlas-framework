---
name: lang-c3
description: C3 systems programming, explicit defer resource release, slices, contract annotations (@param/@return), error returns (!Type), and safe conversions.
---

# C3 Semantic Safety & Contract Assertions Contract

## 1. Core Principles
- **Explicit Cleanup with `defer`**: All allocated resources, file descriptors, and locks must have an immediate `defer` cleanup block defined upon acquisition.
- **Slices over Raw Pointers**: Always prefer bounded slices (`int[]`) rather than raw pointers (`int*`). Slices contain length and prevent out-of-bounds access.
- **Error Results (`!Type`)**: Handle error unions explicitly with `try` or `catch`. Never ignore fallible operation results.
- **Contract Annotations**: Use `@param` bounds checks and `@pure` annotations to establish clear mathematical constraints on functions.
