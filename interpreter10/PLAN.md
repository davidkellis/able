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
  - Include at least one evaluation test that constructs an AST using the new feature in isolation and asserts the evaluated result (e.g., `2 + 3` evaluates to `5`).
  - Keep tests deterministic; use snapshots only where helpful.
  - Prefer direct `evaluate` calls for unit tests; add module tests where relevant.

### Done Definition Per Milestone
- Implementation merged and typechecks clean.
- All new tests passing with high coverage.
- TODO item checked off.


### Status
- Milestones implemented in this pass: 1–12 (runtime core, blocks/assignments, ops/ranges, functions/lambdas, control flow, structs/member access, string interpolation, pattern matching, error handling, module/imports, interfaces/impls/methods, proc/spawn placeholders).
- Extras implemented:
  - Static methods on struct definitions
  - Destructuring in function parameters and assignments (array/struct/typed); for-loop destructuring
  - Array member access via `.index`
  - Privacy enforcement for imports (functions/types/interfaces/unions)
  - Modules/imports enhancements: wildcard imports, package alias objects, dynimport selectors/alias/wildcard
  - Named impl exposure as `impl_namespace`; unnamed impl coherence (reject multiple unnamed impls per (Interface,Type))
  - UFCS fallback for free functions
  - String interpolation prefers struct `to_string`
  - Compound assignments across id/member/index; i32 shift range checks (0..31)
  - TypedPattern and minimal param type checks (primitives, arrays, structs, Error)
  - Minimal generics support: accept type args on calls; enforce interface-based where-constraints at call sites; type-arg count checks
  - Tests now cover constraints on both free functions and instance methods
  - Bind generic type arguments into function env as `${T}_type` strings for introspection; tests added
  - Struct instantiation enforces generic constraints and propagates type arguments for downstream impl lookup
  - Interface method resolution honors inherent-method precedence, selects the most specific impl, and raises on ambiguous overlaps
  - Named impl namespaces expose metadata (`interface`, `target`, `interface_args`) for inspection
  - Concurrency handles: `proc` now returns a status/value/cancel handle and `spawn` returns a memoized future handle
  - Proc/Future `status` expose `ProcStatus` structs and `value()` yields `!T` (success or `ProcError`), enabling propagation with `!`
  - Interfaces support default methods, dynamic dispatch via interface-typed values, and stricter constraint specificity resolution

### Next steps (prioritized)
1) Full v10 concurrency semantics (partially complete)
   - ✅ Cooperative scheduler queues `proc`/`spawn` runners so they progress without explicit joins; tests cover background completion
   - ✅ Cancellation now transitions pending tasks to `Cancelled` with `ProcError` payloads and leaves resolved tasks untouched; coverage exercises both paths
   - ✅ Cooperative helpers (`proc_yield`, `proc_cancelled`) let tasks interleave and proactively observe cancellation, with stress tests covering multi-handle fairness (`ABC` trace ordering) and cooperative cancellation (`wx` trace) via direct runner control
   - ✅ Added nested proc/future stress tests covering nested yields and updated README with recommended usage patterns for library authors

2) Interface & impl completeness (partially complete)
   - ✅ Union-target impls now compare subset precedence (smaller unions win) with improved ambiguity diagnostics when ties remain
   - ✅ Constraint superset precedence now considers inherited interfaces, with coverage for multi-parameter where clauses and dynamic interface dispatch
   - Remaining: higher-kinded/parameterised base-interface chains, mixed visibility/import scenarios, and explicit tooling guidance for disambiguation
   - Tests: overlapping impl resolution edge cases (higher-order generics), interface inheritance in collections, explicit/named impl disambiguation

3) Remaining spec gaps
   - Enforce full privacy model across packages (functions/types/interfaces/impls) and import visibility rules
   - Expand wildcard/import semantics (wildcards with aliasing, dyn package privacy) and expose standard `Proc`/`Future` interface structures
   - Tests: privacy enforcement, import variations, interface struct coverage

4) Generator laziness & iterator parity
   - Draft design for suspend/resume continuations that mirror the Go runtime ✅ (see `design/ts-generator-continuations.md`)
   - ✅ Implement iterator value kinds + sentinel + generator context (straight-line yields first)
   - Extend frame saving across control flow (if/while/for/match) and port Go iterator tests (control-flow coverage achieved; Go parity still pending)
   - Update stdlib helpers (`Channel.iterator`, range iterators) once the runtime hooks land; refresh fixtures

5) Concurrency ergonomics
   - ✅ Surface cooperative yielding APIs, cancellation observers, and ensure futures/procs participate cleanly in collections per spec
   - Remaining: long-running proc stress tests, mixed sync/async loops, `value()` re-entrancy safeguards, and documenting best practices for polling `proc_cancelled`

6) Dynamic interface collections & iterables
   - Extend coverage to ranges/maps of interface values, ensure iteration and higher-order combinators honour most-specific dispatch
   - Tests: `for`/`while` loops over mixed interface unions, comprehension-like patterns, nested collections

7) Performance and maintainability
   - Env lookups and method cache (map hot-paths); micro-benchmarks in tests
   - Split interpreter into modules (values, env, eval nodes)

5) Developer experience
   - Expand examples; README/PLAN alignment; doc comments for evaluator helpers
   - Add coverage target and CI script (bun test --coverage)

### Acceptance criteria for the above
- New behavior covered by focused tests (positive and failure paths)
- No regressions in existing suite; lints remain clean
- PLAN.md updated per milestone completion
