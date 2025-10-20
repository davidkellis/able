# Able v10 Go Interpreter

This package hosts the Go reference interpreter for the Able v10 language. The Go implementation is the canonical runtime that must match `spec/full_spec_v10.md` exactly and stay in lockstep with other interpreters on the shared AST and semantics.

This repository now hosts the stable, canonical runtime. Shared AST fixtures
(`fixtures/ast`) and the strict test harness keep the Go and Bun interpreters in
lockstep with the specification (`./run_all_tests.sh --typecheck-fixtures=strict`).

## Package layout

Interpreter code now lives in focused files so feature work remains approachable:

Core interpreter:
- `pkg/interpreter/interpreter.go` – interpreter struct, module evaluation entry point, package bookkeeping shared across evaluators.
- `pkg/interpreter/interpreter_operations.go` – numeric/boolean operations, comparison helpers, and arithmetic utilities.
- `pkg/interpreter/interpreter_signals.go` – `break`/`continue`/`raise`/`return` signal types plus conversion helpers.
- `pkg/interpreter/interpreter_stringify.go` – stringification helpers (`to_string`, struct formatting, value display).
- `pkg/interpreter/interpreter_members.go` – member/index access logic, UFCS binding, dyn package resolution.
- `pkg/interpreter/interpreter_patterns.go` – pattern assignment helpers, destructuring, typed pattern coercion.
- `pkg/interpreter/interpreter_types.go` – interface/type info, impl lookup, method resolution, type coercion.

Evaluator subsystems:
- `pkg/interpreter/eval_statements.go` – evaluation of statements (blocks, loops, pattern for-loops, control-flow).
- `pkg/interpreter/eval_expressions.go` – expressions, call semantics, lambda invocation, unary/binary operators.
- `pkg/interpreter/definitions.go` – struct/interface/impl/function definition handlers, literal construction helpers.
- `pkg/interpreter/imports.go` – import/dynimport evaluation, privacy enforcement, package registry helpers.
- `pkg/interpreter/impl_resolution.go` – trait/impl discovery, constraint scoring, and method selection logic.

Test suites:
- `pkg/interpreter/interpreter_test_helpers.go` – shared AST builders/bootstrap helpers.
- `pkg/interpreter/interpreter_test.go` – base evaluation smoke tests (literals, scopes, arithmetic).
- `pkg/interpreter/interpreter_imports_test.go` – static/dynamic import behaviour and privacy errors.
- `pkg/interpreter/interpreter_errors_test.go` – rescue/or-else/ensure/raise semantics.
- `pkg/interpreter/interpreter_numeric_test.go` – numeric operations, comparisons, logical diagnostics.
- `pkg/interpreter/interpreter_generics_test.go` – generics/impl diagnostics and UFCS behaviour.
- `pkg/interpreter/interpreter_patterns_test.go` – pattern matching, destructuring, loop patterns.
- `pkg/interpreter/interpreter_structs_test.go` – struct literals, mutation, array index helpers.
- `pkg/interpreter/interpreter_interfaces_test.go` – interface dispatch, defaults, union specificity checks.
- `pkg/interpreter/impl_resolution_test.go` – focused trait resolution parity cases carried over from TypeScript.

This separation mirrors the TypeScript interpreter’s feature boundaries, keeps diffs small, and allows future contributors to extend a subsystem without wading through a monolithic file. When adding new language features, prefer extending the relevant feature file or creating a new one instead of growing `interpreter.go` or the umbrella tests.

## Concurrency model

- **Executors:** `New()` boots with `SerialExecutor` to keep tests deterministic; production callers should swap to `NewGoroutineExecutor` for real concurrency. Both implementations satisfy the same `Executor` contract.
- **Thread-safe environments:** `runtime.Environment` guards scope maps with an `RWMutex`, allowing goroutines spawned by `proc`/`spawn` to share globals safely.
- **Async task state:** Breakpoint and raise stacks live inside an `evalState` that is stored on the async payload, so each goroutine carries its own interpreter state without serialising the entire runtime.
- **Cooperative helpers:** `proc_yield`, `proc_cancelled`, and `proc_flush` are exposed as native functions. `proc_cancelled` now errors when called outside an async task; tests cover the happy path and misuse.
- **User-managed synchronisation:** The runtime ensures its internal structures are safe; user code must still guard shared data (e.g. via native mutex helpers) when true parallelism is enabled.

See `design/go-concurrency.md` for a deeper dive and open follow-ups.

## Typechecker status

- The Go-native typechecker in `pkg/typechecker` now covers the full v10 surface
  (declarations, expressions, patterns, constraint solving).
- Design notes and future enhancement ideas live in `design/typechecker.md` and
  `design/typechecker-plan.md` (spans, incremental checking, TS parity).
- Use the checker to gate fixtures with `ABLE_TYPECHECK_FIXTURES=strict` before
  running the interpreter; diagnostics surface contextual method-set failures
  (e.g., `"via method 'format'"`, `"via method set"`).

### Typechecker integration

- `Interpreter.EnableTypechecker` wires the checker into module evaluation. Use
  `TypecheckConfig{FailFast: true}` to abort execution on diagnostics.
- Fixture runs can enable the checker via `ABLE_TYPECHECK_FIXTURES=warn` (log
  diagnostics) or `ABLE_TYPECHECK_FIXTURES=strict` (fail before evaluation).
  The default remains disabled so existing fixtures continue to run unchanged.

## Design decisions & contributor guidance

- **One responsibility per file:** each `interpreter_*.go` file owns a tight slice of behaviour (operations, patterns, members, type info). New helpers should be colocated with their subsystem; resist adding shared “misc” files.
- **Tests mirror production modules:** every major subsystem has a companion `*_test.go`. When adding behaviour, extend the matching test suite so failures point to the owning area.
- **Shared helpers live in `interpreter_test_helpers.go`:** add new AST builders or bootstrap functions there instead of duplicating scaffolding across suites.
- **Imports/impl resolution stay isolated:** functionality that mutates package state or performs impl lookup remains in `imports.go` / `impl_resolution.go` to avoid reintroducing cross-file coupling.
- **Keep `interpreter.go` tiny:** only interpreter lifecycle, environment wiring, and truly shared utilities should live in `interpreter.go`. Anything evaluator-specific belongs in a dedicated module.
