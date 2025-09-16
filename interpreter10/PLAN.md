## Able v10 Interpreter Plan

This document tracks the implementation plan for the v10 interpreter inside `interpreter10/`.

### Goals
- Build an interpreter that executes the v10 AST (`src/ast.ts`) following `spec/full_spec_v10.md`.
- Implement features incrementally from foundational to advanced, with thorough tests for each.
- Keep code clear and modular to enable future optimizations and targets.

### Architecture
- Interpreter class with:
  - Global and nested `Environment` for name bindings.
  - Runtime `Value` union (tagged) for primitives, arrays, structs, functions, etc.
  - Signal types for control flow (return, raise, break), similar to v6.
  - Public `evaluate(node)` for unit tests and `interpretModule(module)` for running modules.

### Implementation Milestones
1) Runtime core & primitives
   - Add runtime tags for: string, ints (typed via `integerType` defaulting to `i32`), floats (`f64` default), bool, char, nil, array.
   - Implement `evaluate` dispatch for literal nodes and array literal.
   - Tests: Each literal produces expected runtime value; arrays evaluate element-wise.

2) Blocks, identifiers, assignment (:=, =) with identifiers
   - `BlockExpression` scope; `Identifier` lookup; `AssignmentExpression` for identifiers.
   - Tests: declaration, reassignment, shadowing rules in blocks; expression-as-statement semantics.

3) Unary/binary operators and ranges
   - Implement `UnaryExpression`, `BinaryExpression`, and `RangeExpression` (return an iterable value/tag).
   - Tests: numeric ops across i32/f64, string concatenation, boolean logic, range iteration basics.

4) Functions and lambdas
   - `FunctionDefinition`, `LambdaExpression`, closures, call evaluation; pattern params later.
   - Tests: arity checks, closures, returns, nested functions.

5) Control flow: if/or, while, for
   - `IfExpression` with `or` clauses, `WhileLoop`, `ForLoop` over arrays/ranges, `BreakStatement`.
   - Tests: branches, fallthrough, loop control, labeled breaks later if needed.

6) Structs and member access
   - `StructDefinition`, `StructLiteral`, `MemberAccessExpression`, positional/named, functional update.
   - Tests: construction, access, mutation, validation.

7) String interpolation
   - `StringInterpolation` evaluating embedded expressions.
   - Tests: multiple parts, nested expressions.

8) Pattern matching
   - `MatchExpression` and patterns (identifier, wildcard, literal, struct, array, typed later).
   - Tests: simple to complex cases, guards, exhaustiveness errors.

9) Error handling
   - `RaiseStatement`, `RescueExpression`, `PropagationExpression` (!), `OrElseExpression`, `EnsureExpression`, `RethrowStatement`.
   - Tests: raising, rescuing by pattern, ensure semantics, propagation.

10) Packages & imports
   - `PackageStatement`, `ImportStatement`, `DynImportStatement`; link to builtins; stubs for module objects.
   - Tests: selective imports, aliasing, warnings for missing.

11) Interfaces/impls/methods
   - `InterfaceDefinition`, `ImplementationDefinition`, `MethodsDefinition`; method lookup/binding.
   - Tests: inherent vs interface methods, self binding, conflicts.

12) Concurrency (proc/spawn)
   - `ProcExpression`, `SpawnExpression` – stubs to start; later extend to async semantics.
   - Tests: construction and basic behavior placeholders.

### Testing Strategy
- Use Bun’s built-in test runner (`bun test`).
- For each milestone:
  - Add focused test file(s) under `test/`, covering success and error cases.
  - Keep tests deterministic; use snapshots only where helpful.
  - Prefer direct `evaluate` calls for unit tests; add module tests where relevant.

### Done Definition Per Milestone
- Implementation merged and typechecks clean.
- All new tests passing with high coverage.
- TODO item checked off.


