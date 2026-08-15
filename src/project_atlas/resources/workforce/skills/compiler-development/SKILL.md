---
name: compiler-development
description: Conformance test-first, parsing, AST/IR transformations, diagnostic quality, ABI compatibility
---
# Compiler Development

## 1. Conformance Test-First
Drive development using strict conformance tests based on the language specification. Tests must assert both valid code compilation and exact error messages for invalid code. Ensure 100% coverage of language grammar edge cases.

## 2. Robust Parsing
Implement resilient parsers capable of robust error recovery. Upon encountering a syntax error, the parser must attempt to resynchronize and continue parsing to report multiple errors, rather than crashing on the first failure.

## 3. AST Design
Design the Abstract Syntax Tree (AST) to faithfully represent the source code, including trivia (whitespace, comments) if required for tooling. Ensure the AST is immutable and easily traversable using visitor patterns.

## 4. Semantic Analysis
Perform rigorous semantic analysis, including type checking, scope resolution, and control flow analysis. Ensure all variables are initialized before use and all code paths return values. Validate lifetimes and ownership where applicable.

## 5. High-Quality Diagnostics
Emit precise, actionable diagnostic messages. Include specific file paths, line numbers, and column offsets. Provide visual context (snippets of the offending code) and suggest potential fixes. Obscure compiler errors are unacceptable.

## 6. Intermediate Representation (IR)
Design a robust Intermediate Representation. The IR should be simpler than the AST, enabling easier optimization passes. Consider using SSA (Static Single Assignment) form for data flow analysis.

## 7. Optimization Passes
Implement conservative, provably correct optimization passes. Start with simple constant folding, dead code elimination, and function inlining. Verify that optimizations do not alter program semantics.

## 8. ABI Compatibility
Adhere strictly to the target platform's Application Binary Interface (ABI). Ensure proper struct packing, calling conventions, register usage, and stack alignment to allow interoperability with C libraries and the OS.

## 9. Bootstrapping Strategy
If building a self-hosting compiler, maintain a clear bootstrapping strategy. Ensure older versions of the compiler can reliably build newer versions. Implement automated checks to verify bootstrapping integrity.

## 10. Memory Management
Manage compiler memory efficiently. Compilers allocate vast numbers of small objects (AST nodes). Consider using arena allocators for tree nodes to improve cache locality and eliminate overhead of individual deallocations.

