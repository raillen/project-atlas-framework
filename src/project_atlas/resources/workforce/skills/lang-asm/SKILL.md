---
name: lang-asm
description: Assembly x86/x86-64 calling conventions (System V AMD64 / Microsoft x64), 16-byte stack alignment, CFI directives, non-executable stack, and register clobber safety.
---

# Assembly x86/x86-64 Engineering & ABI Safety Contract

## 1. Calling Conventions & Register Discipline
- **System V AMD64 (Linux, macOS, BSD)**:
  - Integer Arguments: `RDI`, `RSI`, `RDX`, `RCX`, `R8`, `R9`.
  - Return Value: `RAX` (and `RDX` for 128-bit).
  - Callee-Saved Registers: `RBX`, `RBP`, `R12`, `R13`, `R14`, `R15`. MUST be preserved across function calls.
  - Caller-Saved (Scratch): `RAX`, `RCX`, `RDX`, `RSI`, `RDI`, `R8`, `R9`, `R10`, `R11`.
- **Microsoft x64 (Windows)**:
  - Integer Arguments: `RCX`, `RDX`, `R8`, `R9`.
  - Mandatory 32-byte Shadow Space allocated by caller.
  - Callee-Saved: `RBX`, `RBP`, `RDI`, `RSI`, `R12`, `R13`, `R14`, `R15`.

## 2. Mandatory Stack Alignment & Security
- **16-Byte Alignment**: The stack pointer (`RSP`) MUST be 16-byte aligned immediately before any `call` instruction. (Upon function entry after `call`, `RSP` is 8 mod 16 due to the return address pushed).
- **CFI Directives**: Include `.cfi_startproc`, `.cfi_endproc`, and `.cfi_def_cfa_offset` for stack unwinding and debugging.
- **Non-Executable Stack**: Every assembly file must declare `.section .note.GNU-stack,"",@progbits` to prevent security mitigations from marking stack memory executable.
- **Spectre / Branch Safety**: Use `lfence` or retpoline for speculative execution mitigation on indirect jumps when processing untrusted inputs.
