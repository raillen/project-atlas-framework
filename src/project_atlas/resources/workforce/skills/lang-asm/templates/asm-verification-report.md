# x86-64 Assembly Audit Report

- **Target Architecture**: `x86_64`
- **Target ABI**: `System V AMD64` / `Microsoft x64`
- **Stack Alignment**: Verified (16-byte boundary honored before calls)
- **CFI Unwind Directives**: Present and valid
- **Non-Executable Stack**: Present (`.note.GNU-stack`)
- **Register Clobber Audit**: All callee-saved registers preserved
