# Able v11 Typechecker (Go)

Status: in-progress (2025-10-19)

- `checker.go` exposes `Checker.CheckModule`, currently supporting literal typing, let bindings, and a global declaration pass.
- `env.go` implements lexical environments for type information.
- `types.go` defines core type representations (primitives, structs, functions, proc/future, unknown).
- `inference.go` tracks inferred types per AST node.
- `literals.go` handles literal expressions, identifier lookup, and basic statement dispatch.
- `array_literal.go` infers `Array` element types and validates literal element compatibility.
- `index_expression.go` validates indexing operations and propagates element types.
- `range_expression.go` validates numeric ranges and records their element type.
- `decls.go` collects top-level struct/union/interface/function signatures before body checking.
- `constraint_solver.go` enforces trait/where-clause obligations with contextual diagnostics.
- `typechecker_integration.go` integrates with the interpreter so callers can enable a pre-flight typecheck (optionally fail-fast on diagnostics) via `Interpreter.EnableTypechecker`. Fixture harnesses may toggle this with the `ABLE_TYPECHECK_FIXTURES` environment flag (`warn`/`strict`).

- Next steps:
- Flesh out documentation/design notes for the completed surface.
- Expand diagnostics with span data once the AST carries source locations.
- Continue parity validation against the TypeScript interpreter (async helper diagnostics, concurrency fixtures, etc.).

Design background: see `design/typechecker.md` and `design/typechecker-plan.md`.

Tests: `go test ./pkg/typechecker`.
