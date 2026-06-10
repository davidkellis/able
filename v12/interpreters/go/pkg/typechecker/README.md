# Able v12 Typechecker (Go)

The Go typechecker owns v12 static diagnostics over the shared AST. The
program-wide entry point is normally the right one:

```go
result, err := typechecker.NewProgramChecker().Check(program)
```

It checks the dependency-ordered `driver.Program`, returns package-qualified
diagnostics, privacy-aware package summaries, and per-package inference maps.
`Checker.CheckModule` is available for a single parsed module; callers needing
imports, visibility, re-exports, or source attribution should use
`ProgramChecker` instead. Runtime/CLI users can call
`interpreter.TypecheckProgram`, which exposes the same `CheckResult`.

Key implementation groups:

- `checker.go`, `decls*.go`, and `statement_checker.go`: module orchestration,
  declaration collection, and scoped body checking.
- `types.go`, `env.go`, and `inference.go`: checker-only type metadata,
  environments, and AST-node side tables.
- `constraint_solver*.go`, `implementation_validation.go`, and
  `member_access*.go`: generic/interface/impl/method-set resolution.
- `program_checker*.go`: program imports, export capture, visibility, package
  summaries, re-exports, and source-hinted diagnostics.

`Interpreter.EnableTypechecker` is a module-evaluation integration point;
`Interpreter.EvaluateProgram` normally uses `TypecheckProgram` instead so
program-wide semantics are preserved. Fixture runs are strict by default;
`ABLE_TYPECHECK_FIXTURES=strict|warn|off` selects only their diagnostic policy.

The language specification remains authoritative. Static diagnostics must not
silently replace runtime checks for dynamic language rules. The active contract
and change-selection gate live in `v12/design/typechecker-plan.md`; the
architecture reference is `v12/design/typechecker.md`.

Tests: `go test ./pkg/typechecker`.
