# Able v12 Typechecker (Go)

Status: active Go architecture as of 2026-07-14
Language authority: [`spec/full_spec_v12.md`](../../spec/full_spec_v12.md)
Change-selection contract: [`typechecker-plan.md`](typechecker-plan.md)

The Go typechecker validates static facts over the shared v12 AST. It is used
by the Go CLI, tree-walker, bytecode interpreter, fixture harness, and compiler
analysis. It does not define alternate runtime semantics: where the spec keeps
a rule dynamic, interpreters must retain the corresponding runtime behaviour.

## Entry points and result

`Checker.CheckModule(*ast.Module)` checks a single parsed module. It is useful
for a self-contained module or tooling that has already supplied a correct
prelude.

`ProgramChecker.Check(*driver.Program)` is the normal package-aware entry
point. The loader supplies modules in dependency order. For every module, the
program checker builds an import prelude from already captured package surfaces,
runs a fresh `Checker`, adds module/file attribution to diagnostics, and then
captures that module's exports for downstream users.

`interpreter.TypecheckProgram` simply exposes the program-wide API at the
runtime boundary. A result contains:

| Field | Meaning |
| --- | --- |
| `Diagnostics` | `ModuleDiagnostic` values with package, files, a diagnostic, and best-effort `SourceHint`. |
| `Packages` | privacy-aware `PackageSummary` metadata for symbols, declarations, implementations, and method sets. |
| `Inferred` | one `InferenceMap` per checked package, keyed by AST node. |

The result is diagnostic data and reusable metadata, not a modified AST.
`PrepareProgramForEvaluation` is intentionally narrower: it performs the
declaration-side preparation needed by skip-check execution, but does not
validate bodies or produce the diagnostic result.

## Module-checking pipeline

`CheckModule` follows a recoverable, deterministic sequence:

1. Reset per-module inference, control-flow state, obligation state, and prior
   diagnostics.
2. Collect declarations and initial generic/where-clause obligations. This
   establishes structs, unions, interfaces, functions, aliases, implementations,
   and method sets before bodies are checked.
3. Combine active built-in implementations with the program prelude, then
   create the module body scope and check every statement and expression.
4. Resolve gathered constraints and validate implementation declarations.
5. Return all diagnostics that can be reported safely; only malformed checker
   inputs return a Go error.

The package is split by responsibility rather than by a monolithic visitor:

- `checker.go`, `statement_checker.go`, and `decls*.go` orchestrate collection
  and scopes.
- `types.go`, `env.go`, `inference.go`, and `type_*.go` provide checker-only
  type representation and substitution/assignability helpers.
- `constraint_solver*.go` and `implementation_validation.go` solve generic,
  interface, impl, and method-set obligations.
- Feature checkers such as `patterns.go`, `match.go`, `member_access*.go`,
  `functions.go`, `concurrency.go`, `array_literal.go`, and
  `index_expression.go` infer and validate their AST families.
- `program_checker*.go` coordinates imports, export capture, re-exports,
  visibility, package summaries, and source hints.

## Types, inference, and constraints

Checker type values represent primitive scalars, functions, nominal structs and
unions, interfaces, applications, aliases, futures, and unknown values.
`InferenceMap` records a `Type` against each relevant `ast.Node`; callers use a
clone, so later checks cannot mutate an already returned result.

Generic parameters and `where` clauses are represented as obligations. The
solver resolves the required interface and arity, then tests direct interface
compatibility, visible implementations, and visible method sets with the
necessary substitutions. It attaches contextual details to a diagnostic when a
method or nested obligation fails. This makes constraint failure reporting
consistent across declaration checking, calls, member access, and package
imports.

`UnknownType` is a deliberate recovery tool for dynamic or incompletely known
information. It prevents a cascade from concealing an earlier actionable
diagnostic; it is not evidence that a program is statically valid.

## Package and visibility semantics

Program checking is required for import-sensitive work. `ProgramChecker`:

- uses the loader's dependency order so an import sees an upstream package's
  captured surface;
- supplies imported symbols, implementations, and method sets in a prelude;
- distinguishes exported and private definitions in `PackageSummary`;
- reports unknown packages, private imports/selectors, private source-export
  attempts, and named-implementation binding collisions with a package/file
  source hint where available; and
- propagates named and wildcard source re-exports through the same package
  surface without creating a new nominal type.

A standalone `CheckModule` gives unresolved imports conservative placeholder
bindings so it can still check local structure. It must not be used to claim
that a multi-package program passed privacy or selector validation.

## Runtime, CLI, compiler, and fixtures

`able check`, default `able run`, and default `able build` use program-wide
checking. `able test` checks loaded production and test sources together unless
its explicit test typecheck mode is off. `EvaluateProgram` stops before
evaluation when diagnostics exist, unless the caller asks to allow them; it also
passes checked inference facts to the bytecode interpreter.

`Interpreter.EnableTypechecker` remains available for direct single-module
evaluation. `FailFast` controls whether diagnostics prevent that evaluation.
Fixture harnesses default to strict checking. The
`ABLE_TYPECHECK_FIXTURES=strict|warn|off` setting changes fixture diagnostic
handling only; it does not change the language's runtime checks.

Compiler analysis uses `NewProgramChecker().Check` as well. A compiler or VM
optimization may consume an inference fact only when it preserves the dynamic
fallback required by the language contract. The typechecker is not a licence to
introduce a separate native type system.

## Diagnostics and source attribution

`Diagnostic` carries a severity, message, origin node, and optional notes.
`ModuleDiagnostic` adds package/file context, and `SourceHint` gives the
loader's best available path/line/column attribution. General AST source spans
are not yet a guaranteed interface, so new diagnostics must work correctly when
only the package or file is available.

Diagnostics should be specific enough for users and stable enough for the CLI,
fixtures, and tooling. Prefer structured notes or source hints over encoding
extra data into a fragile message string. Continue collecting independent
diagnostics where recovery is safe.

## Extending the checker

Before editing this package, cite the relevant v12 rule (or record the missing
rule in `spec/TODO_v12.md`), reproduce the issue with a focused checker test,
and add a program-level case when package semantics are involved. When inferred
facts feed execution or lowering, cover the applicable tree-walker, bytecode,
and compiler boundary. Update the spec and this architecture note when a new
diagnostic becomes canonical.

The active plan deliberately does not contain a speculative TypeScript parity,
LSP, cache, or micro-optimization queue. Those require a source-backed consumer
and the evidence gate in `typechecker-plan.md`; the TypeScript-era chronology is
archived in `typechecker-plan-historical.md`.

The named-implementation/source-re-export audit records the exact static
import collision, privacy, and identity boundary:
`reexport-named-implementation-import-audit.md`.

## Verification

The package unit suite is the primary checker control:

```sh
cd v12/interpreters/go
go test ./pkg/typechecker
```

For an integration change, also run the focused CLI and program-evaluation
groups named in `typechecker-plan.md`, then the relevant fixture/interpreter and
compiler controls. Finish broad v12 work with `./run_all_tests.sh` from the
repository root.
