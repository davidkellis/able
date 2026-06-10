# Able v12 Typechecker: Active Go Contract

Status: active as of 2026-07-14
Authority: `spec/full_spec_v12.md` defines language behaviour; this note records
the current Go checker boundary. The tree-walker and bytecode interpreters remain
the runtime references when a static diagnostic and runtime behaviour differ.

The TypeScript interpreter is not part of the active v12 toolchain. Its former
parity work is historical and does not select implementation work.

## Purpose and ownership

`v12/interpreters/go/pkg/typechecker` checks the shared Able AST and exposes
diagnostics plus reusable type facts. It has three public levels:

- `typechecker.New().CheckModule` checks one already-parsed module.
- `typechecker.NewProgramChecker().Check` checks the dependency-ordered modules
  supplied by `driver.Program` and builds package summaries.
- `interpreter.TypecheckProgram` is the Go runtime/CLI-facing alias for the
  program-wide path.

Program-wide checking is the normal path for packages. It resolves imported
public surfaces before checking a downstream module, preserves package and
file attribution in `ModuleDiagnostic`, and returns a `CheckResult` containing:

- diagnostics;
- privacy-aware `PackageSummary` data (symbols, structs, interfaces,
  functions, implementations, and method sets); and
- an `InferenceMap` for each checked package.

Inference facts are side tables keyed by AST nodes. They must not become a
second AST schema. `PrepareProgramForEvaluation` deliberately performs only
the declaration-side preparation that runtime evaluation requires; it is not a
replacement for diagnostics or a licence to give unchecked programs different
language semantics.

## Current checked surface

The Go checker has declaration collection, scoped body checking, inference,
patterns, overload and member resolution, generic substitution, interface/impl
and method-set constraint solving, concurrency helper diagnostics, package
imports, re-exports, visibility checks, and source-hinted diagnostics. The
focused `checker_*_test.go` files and `program_checker_test.go` are the
executable inventory; this list is a navigation aid, not an independent
language authority.

For one module, `CheckModule` gathers declarations, applies a prelude, checks
bodies, resolves obligations, validates implementations, and accumulates
diagnostics where recovery is safe. For a `driver.Program`, `ProgramChecker`
repeats that process in loader dependency order, supplies imported declarations
through the next module's prelude, and captures its export surface after it is
checked. Use the program-wide entry point whenever imports or visibility matter.

## Integration contract

- `able check`, normal `able run`, and `able build` use program-wide checking
  and report diagnostics before continuing. The build command has its explicit
  skip-check option for exceptional workflows.
- `able test` checks its loaded production and test modules together unless its
  explicit typecheck mode is disabled. This preserves same-package privacy
  semantics.
- `Interpreter.EvaluateProgram` checks by default, stops evaluation on
  diagnostics unless `AllowDiagnostics` is requested, and provides checked
  inference facts to bytecode execution.
- The fixture harness defaults to strict checking. `ABLE_TYPECHECK_FIXTURES`
  accepts `strict`, `warn`, or `off` only as an explicit diagnostic policy for
  fixture/debug workflows; it does not alter Able runtime rules.
- Compiler analysis uses the same `ProgramChecker` result rather than creating
  a compiler-specific type model.

Diagnostic text and severity are part of tooling compatibility. Add structured
context through `ModuleDiagnostic`, `SourceHint`, and diagnostic notes rather
than teaching callers to parse an ad-hoc new message shape. AST source spans
are not yet a general contract; current file/line attribution is best effort
from loader node origins.

## Selecting new checker work

There is no active checker implementation candidate solely from this roadmap.
The spec's open type-system items—full inference/variance/coercion rules and
interface-dictionary details—need language decisions before a checker
implementation is selected. Static named-implementation import collisions,
selector renaming, and explicit source re-exports are now covered through the
shared AST and program-wide package surface. See
`reexport-named-implementation-import-audit.md`.

A new checker change must begin with all of the following:

1. A cited v12 specification rule, or a documented spec gap in `spec/TODO_v12.md`.
2. A minimal AST/loader or source-level reproduction and a checker regression
   test that distinguishes the intended diagnostic from runtime behaviour.
3. A program-wide case when imports, visibility, re-exports, impl propagation,
   or source attribution are involved.
4. Interpreter and compiler coverage when the inferred fact can affect either
   execution/lowering path; the change must not make static checking the sole
   definition of a runtime rule.
5. Documentation/spec updates in the same tranche when the diagnostic becomes
   canonical.

Do not schedule speculative incremental caches, LSP APIs, TypeScript parity,
or typechecker micro-optimisations. A performance change needs the same
verifier-backed, cross-application evidence required by the active performance
policy; no named-container, source-shape, or benchmark-only shortcut is
permitted.

## Verification

Run the focused package and integration controls under the repository's
one-process memory guard before handing off a checker change:

```sh
cd v12/interpreters/go
env GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 GOCACHE=$(pwd)/.gocache \
  go test ./pkg/typechecker -run '^(TestProgramChecker|TestCheckerPackage)' -count=1 -timeout 55s
env GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 GOCACHE=$(pwd)/.gocache \
  go test ./cmd/able -run '^TestRunProgramTypecheck' -count=1 -timeout 55s
env GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 GOCACHE=$(pwd)/.gocache \
  go test ./pkg/interpreter -run '^(TestInterpreterPipelineTypecheckDiagnostics|TestInterpreterTypechecker)' -count=1 -timeout 55s
```

Expand the group to the changed semantic surface, then use `./run_all_tests.sh`
before a broad handoff. Tests must remain bounded below one minute each.

## Historical record

`typechecker-plan-historical.md` preserves the 2025 bootstrap, package-export,
and retired TypeScript/Bun parity chronology. It is not an implementation queue.
