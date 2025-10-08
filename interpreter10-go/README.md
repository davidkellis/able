# Able v10 Go Interpreter

This package hosts the Go reference interpreter for the Able v10 language. The Go implementation is the canonical runtime that must match `spec/full_spec_v10.md` exactly and stay in lockstep with other interpreters on the shared AST and semantics.

Current focus:
- Define the canonical AST in Go (`pkg/ast`).
- Mirror the TypeScript AST helpers so shared fixtures remain compatible.
- Build the evaluator and runtime using Go-native concurrency primitives once the AST is complete.
- Keep the JSON fixtures under `fixtures/ast` green by running `go test ./pkg/interpreter` after any change to `interpreter10/scripts/export-fixtures.ts` (the tests hydrate every fixture to ensure parity with the TypeScript interpreter).

Development checklist and milestones live in the workspace root `PLAN.md` until this package grows its own detailed plan.

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

## Design decisions & contributor guidance

- **One responsibility per file:** each `interpreter_*.go` file owns a tight slice of behaviour (operations, patterns, members, type info). New helpers should be colocated with their subsystem; resist adding shared “misc” files.
- **Tests mirror production modules:** every major subsystem has a companion `*_test.go`. When adding behaviour, extend the matching test suite so failures point to the owning area.
- **Shared helpers live in `interpreter_test_helpers.go`:** add new AST builders or bootstrap functions there instead of duplicating scaffolding across suites.
- **Imports/impl resolution stay isolated:** functionality that mutates package state or performs impl lookup remains in `imports.go` / `impl_resolution.go` to avoid reintroducing cross-file coupling.
- **Keep `interpreter.go` tiny:** only interpreter lifecycle, environment wiring, and truly shared utilities should live in `interpreter.go`. Anything evaluator-specific belongs in a dedicated module.
