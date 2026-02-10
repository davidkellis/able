# Able v12 Compiler Vision (Correctness-First AOT)

This document defines the correctness-first vision for the Able v12 compiler, the intended architecture, and a complete work breakdown. It is authoritative for compiler implementation scope and must stay aligned with `spec/full_spec_v12.md`.

## Vision

The Able v12 compiler is a correctness-first ahead-of-time compiler that emits Go code implementing the v12 spec semantics exactly. It must compile all non-dynamic code to Go and execute it without interpreter involvement. The interpreter is used only for explicitly dynamic features and must not be used as a silent fallback.

## Non-Negotiable Principles

- Spec fidelity is mandatory. Any deviation from `spec/full_spec_v12.md` is a bug.
- No silent fallback. If code is not compileable, the compiler must fail or explicitly cross the dynamic boundary.
- The AST contract is canonical. Any compiler IR must preserve AST semantics.
- Determinism matters. Evaluation order, diagnostics, and concurrency semantics must match the interpreters.
- The compiled output must be reproducible and relocatable.

## Scope

- Target language is Go for v12.
- The compiler produces Go packages per Able package, then builds a single native binary.
- The compiler must compile stdlib and kernel to Go packages.
- Dynamic facilities are supported via an explicit dynamic boundary and interpreter bridge.

## Dynamic Boundary Contract

Dynamic features are the only reasons the interpreter may execute code:

- `dynimport`
- `defpackage`
- dynamic metaprogramming and evaluation

The boundary is explicit and bidirectional.

- Compiled code may call the interpreter when dynamic features are executed.
- The interpreter may call compiled functions when dynamic code references compiled definitions.

Value conversion at the boundary must be lossless or raise a runtime error.

## Compiler Architecture

### Frontend

- Parse and typecheck the full module graph.
- Preserve typechecker outputs: resolved types, overloads, impls, interface dispatch tables.
- Mark dynamic feature usage per module and per function.

### Typed IR

Lower AST to a typed IR that makes evaluation order, effects, and error flow explicit.

The IR must encode:

- Exact evaluation order.
- Pattern assignment semantics.
- Control flow semantics (if, match, loops, breakpoint).
- Error propagation (`!`, `or`, `raise`, `rescue`, `ensure`).
- Interface dispatch selection and overload resolution.

### Codegen (Go)

Emit Go packages for each Able package. Generated code uses fully typed values and direct calls for known types. Only dynamic boundary calls use `runtime.Value`.

### Compiled Runtime Core

Implement core types and operations in Go:

- `Array`
- `Ratio`
- `String`
- `Channel`, `Mutex`, `Future`

`BigInt` is a stdlib type (`able.numbers.bigint`) implemented in Able. It is compiled as part of stdlib compilation and does not require a dedicated runtime primitive beyond compiled arrays and numeric ops.

### Transition Notes (Current Work)

- Array literals now lower to kernel `Array` handles in compiled output. Compiled `__able_array_*` externs now use the shared runtime array store (no interpreter bridge).
- Compiled `__able_hash_map_*` externs now use the runtime hash map store, with hashing/equality dispatched through the compiler bridge.
- Compiled runtime helpers use interpreter fast-path operators for numeric/string ops before falling back to interpreter dispatch.
- Compiled string/char externs (`__able_String_*`, `__able_char_*`) now execute in the compiled runtime without interpreter calls.
- Compiled numeric externs (`__able_ratio_from_float`, `__able_f32_bits`, `__able_f64_bits`, `__able_u64_mul`) now execute in the compiled runtime without interpreter calls.
- Compiled runtime now handles Ratio arithmetic and comparisons directly (no interpreter fallback).

### Dynamic Bridge

Provide explicit conversion and invocation between compiled values and interpreter values. The interpreter is only invoked through these bridge points.

## Spec Alignment Map

The compiler must implement the following semantics exactly.

Bindings and assignment.
Spec reference: Section 5, especially 5.3.

Expressions and operators.
Spec reference: Section 6, especially 6.3.2 and 6.5.

Control flow expressions and loops.
Spec reference: Section 8.

Methods and UFCS resolution.
Spec reference: Section 9.4 and 7.4.3.

Interfaces, impl resolution, and dispatch.
Spec reference: Section 10.

Error handling and `or` propagation.
Spec reference: Section 11.

Concurrency, futures, and scheduling.
Spec reference: Section 12.

Packages and imports.
Spec reference: Section 13.

Host interop.
Spec reference: Section 16.

Testing and diagnostics.
Spec reference: Section 17 and diagnostics guidelines.

## Correctness Invariants

- All non-dynamic code runs without interpreter execution.
- Evaluation order matches interpreter and spec.
- Every compiled value has a deterministic runtime representation.
- Every dynamic boundary call is explicit and traceable.
- Interface and overload resolution results match the typechecker.
- Stdlib and kernel are compiled and callable without dynamic dispatch.

## Testing Strategy

All compiler correctness must be validated by tests and fixtures.

### Compiler Fixture Parity

- Run every exec fixture through compiled output.
- Compare stdout, stderr, exit codes.
- Ensure compiled output does not invoke interpreter for non-dynamic fixtures.

### Dynamic Boundary Fixtures

- Dedicated fixtures that exercise dynimport, defpackage, eval.
- Ensure compiled code calls interpreter only at boundary.
- Ensure interpreter can call compiled functions correctly.

### Stdlib and Kernel Parity

- Run stdlib tests under compiled mode.
- Compare results to bytecode and tree-walker outputs.

### Typechecker and Diagnostics Parity

- Ensure compiled diagnostics match interpreter diagnostics.
- Ensure typecheck warnings and errors are consistent.

### Concurrency

- Run concurrency fixtures in compiled mode.
- Ensure scheduler semantics match interpreters.

### ABI and Value Conversion

- Test conversion between compiled values and runtime.Value.
- Verify lossless conversions or correct errors.

## Work Breakdown

The work below is exhaustive. Each item must be completed to reach the vision.

### Phase 0: Documentation and Spec Alignment

1. Maintain this vision doc and keep it aligned with spec and implementation.
2. Define explicit dynamic boundary semantics in `spec/full_spec_v12.md`.
3. Record compiler-specific gaps in `spec/TODO_v12.md`.
4. Document compiled runtime ABI for core types.

### Phase 1: Frontend Graph and Metadata

1. Build a full module dependency graph from the entry.
2. Preserve typechecker outputs in a compiler-visible form.
3. Annotate dynamic feature usage per module and per function.
4. Add tests for dynamic feature detection.

### Phase 2: Typed IR

1. Define typed IR structures.
2. Lower AST to IR with explicit evaluation order.
3. Encode error flow semantics in IR.
4. Encode pattern assignment semantics in IR.
5. Encode match and loop semantics in IR.
6. Encode method resolution in IR.
7. Encode interface dispatch in IR.
8. Add IR validation tests.

### Phase 3: Compiled Runtime Core

1. Implement compiled Array with correct semantics.
2. Implement compiled Ratio with correct semantics.
3. Implement compiled String helpers with correct semantics.
4. Implement compiled Channel, Mutex, Future with correct semantics.
5. Add parity tests against interpreter and stdlib fixtures.

### Phase 4: Codegen Core

1. Emit Go packages for each Able package.
2. Emit compiled functions and methods with direct calls.
3. Emit control flow constructs correctly.
4. Emit error handling constructs correctly.
5. Emit pattern assignment logic correctly.
6. Emit match logic correctly.
7. Emit static operator lowering without dynamic dispatch.
8. Add codegen tests per feature.

### Phase 5: Interfaces and Overloads

1. Emit dictionary tables for interface dispatch.
2. Emit compiled overload resolution for all call sites.
3. Validate interface dispatch parity with fixtures.

### Phase 6: Stdlib and Kernel Compilation

1. Compile stdlib packages as Go packages.
2. Compile kernel packages as Go packages.
3. Wire stdlib and kernel into the compiled program by default.
4. Validate stdlib and kernel parity fixtures.

### Phase 7: Dynamic Boundary Bridge

1. Define conversion between compiled values and runtime.Value.
2. Implement compiled to interpreter boundary calls.
3. Implement interpreter to compiled callbacks.
4. Add fixtures for dynamic boundary correctness.
5. Add tests that ensure no interpreter execution for static programs.

### Phase 8: Execution and Tooling

1. Emit a compiled `main` that executes compiled code directly.
2. Only initialize the interpreter when dynamic features are invoked.
3. Emit reproducible build trees with Go module wiring.
4. Ensure compiled artifacts are relocatable.
5. Add tooling diagnostics for dynamic-only features.

### Phase 9: Fixture and Regression Coverage

1. Full exec fixture parity in compiled mode.
2. Full stdlib test suite in compiled mode.
3. Bytecode and tree-walker parity checks for compiled output.
4. Regression tests for evaluation order and error semantics.

### Phase 10: Validation and Performance Baseline

1. Verify no silent fallback paths remain.
2. Produce a performance baseline for compiled outputs.
3. Ensure semantics parity dominates performance concerns.

## Definition of Done

The compiler is complete when:

- Every non-dynamic program runs fully compiled with no interpreter execution.
- Dynamic features are handled only at explicit boundaries.
- Stdlib and kernel are compiled and called directly.
- All fixtures and tests are green in compiled mode.
- `spec/full_spec_v12.md` is fully implemented.
