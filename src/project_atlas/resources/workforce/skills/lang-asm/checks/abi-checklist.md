# x86-64 Assembly Verification Checklist

- [ ] Function prologue saves all modified callee-saved registers (`RBX`, `RBP`, `R12`-`R15`).
- [ ] Function epilogue restores callee-saved registers in reverse order.
- [ ] Stack pointer `RSP` is 16-byte aligned before every nested `call`.
- [ ] `.cfi_startproc` and `.cfi_endproc` surround each function.
- [ ] `.section .note.GNU-stack,"",@progbits` present at end of file.
- [ ] Caller-saved scratch registers are not assumed to persist across `call` instructions.
- [ ] Direction flag `DF` is cleared (`cld`) if string operations (`movsb`, `stosb`) are performed.
