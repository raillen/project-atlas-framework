.intel_syntax noprefix
.text
.globl safe_add_integers
.type safe_add_integers, @function

safe_add_integers:
    .cfi_startproc
    # Arguments: RDI = a, RSI = b
    # Return: RAX = a + b
    mov rax, rdi
    add rax, rsi
    ret
    .cfi_endproc

.size safe_add_integers, .-safe_add_integers

# Non-executable stack note
.section .note.GNU-stack,"",@progbits
