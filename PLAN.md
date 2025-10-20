# Able Project Roadmap (v10 focus)

## Scope
- Maintain a canonical Able v10 language definition across interpreters and tooling, with the written specification remaining the source of truth for behaviour.
- Prioritise the Go interpreter until it matches the TypeScript implementation feature-for-feature (the only intentional divergence is that Go uses goroutines/channels while TypeScript simulates concurrency with a cooperative scheduler).
- Keep the TypeScript and Go AST representations structurally identical so tree-sitter output can feed either runtime (and future targets like Crystal); codify that AST contract inside the v10 specification once validated.
- Document process and responsibilities so contributors can iterate confidently.

## Existing Assets (2025-09-05)
- `spec/full_spec_v10.md`: authoritative semantics.
- `interpreter10/`: Bun-based TypeScript interpreter + AST definition (`src/ast.ts`) and extensive tests.
- Legacy work: `interpreter6/`, assorted design notes in `design/`, early stdlib sketches.

## Phase 0 — Canonical AST Alignment
1. Inventory every node defined in `interpreter10/src/ast.ts` (types, helpers, DSL aliases).
2. Document a language-agnostic AST schema (names, fields, invariants) anchored by the Go data structures; flag any TypeScript-specific quirks (e.g., optional fields, default builders).
3. Decide whether revisions are needed in the TS AST to satisfy the canonical schema; capture required changes as tickets.
4. Deliverables: schema doc (likely `design/ast-schema-v10.md`), parity checklist, backlog items for TS adjustments if required.

## Phase 1 — Go Interpreter Foundations
1. Repository skeleton
   - Create `interpreter10-go/` with `go.mod`, module layout (`cmd/`, `pkg/ast`, `pkg/interpreter`, `pkg/runtime`, `internal/tests`).
   - Adopt Go 1.22+, enable tooling (lint, gofmt, go test).
2. AST implementation
   - Port canonical AST to Go structs/interfaces.
   - Provide builders mirroring the TS DSL helpers for tests.
   - Implement JSON/YAML (de)serialization helpers if helpful for cross-language fixtures.
3. Test harness
   - Establish `pkg/ast/ast_test.go` validating round-trips and parity with TS snapshots (shared fixtures).
   - Set up golden file strategy for complex nodes (only if needed).

## Phase 2 — Runtime Infrastructure in Go
1. Runtime values & environment
   - Design tagged union equivalent (e.g., `type Value interface{}` with dedicated structs + discriminant) covering all v10 kinds.
   - Implement lexical `Environment` stack with `Define/Assign/Get`, mirroring TS semantics (error cases, shadowing, module scope).
2. Interpreter skeleton
   - Create evaluator entry points (`Evaluate(node ast.Node) (Value, error)`, `InterpretModule(ast.Module)`).
   - Port control-flow signals (`return`, `raise`, `break`, `break label`) using Go errors or sum types.
3. Concurrency model mapping
   - Model Able `proc` as goroutines with a scheduler abstraction wrapping channels/mutexes to preserve deterministic test execution.
   - Represent `Proc` handle, `Future`, `ProcStatus`, `ProcError` using Go structs; ensure blocking semantics (`value()`) align with spec using channels/condition vars.
   - Implement cooperative helpers analogues (`proc_yield`, `proc_cancelled`) leveraging Go contexts or explicit channels.

## Phase 3 — Feature Parity Implementation
Follow the milestone order already validated in TypeScript, adapting to Go idioms:
1. Literals, arrays, identifiers, assignments.
2. Blocks, scopes, unary/binary ops, ranges.
3. Functions/lambdas, closures, call semantics.
4. Control flow (`if/or`, `while`, `for`, labeled breaks).
5. Structs, member access, mutation rules.
6. String interpolation.
7. Pattern matching (identifier → typed patterns).
8. Error handling (`raise`, `rescue`, `!`, `ensure`, `or else`, `rethrow`).
9. Modules/imports/package handling (module env, privacy).
10. Interfaces/impls/method dispatch, UFCS fallback.
11. Generics/where clauses for functions and structs as spec requires (mirror runtime enforcement).
12. Concurrency (`proc`, `spawn`, helpers) with deterministic scheduler, cancellation, memoization.
13. Host interop stubs (`prelude`, `extern`) as no-ops or placeholders until tree-sitter integration.
For each milestone: port representative TS tests into Go, add new cases where Go-specific behavior needs coverage, keep PLAN updated.

### Typechecker (Go)
- ✅ Skeleton package with literal typing, identifier resolution, declaration collection for structs/unions/interfaces/functions (parameter + return types), and inference map storage.
- ✅ Declaration pass captures generics, where clauses, and rejects redefinitions with diagnostics.
- ✅ Expression coverage spans function calls (incl. generics), member access, control flow, pattern matching, async constructs (`proc`/`spawn`), aggregates, and literals.
- ✅ Pattern typing: destructuring assignments, match clauses, function parameter patterns, and guards.
- ✅ Constraint solving: interface/impl obligations, trait bounds, result/option helpers, and method-set where clauses with contextual diagnostics.
- ✅ Diagnostics: standardised messaging (span plumbing deferred until parser emits locations).
- ✅ Integration: optional checker invocation before interpretation, exposes inference metadata for future compiler/LSP consumers.

## Phase 4 — Cross-Interpreter Parity & Tooling
1. Shared AST fixtures
   - Generate machine-readable AST samples from TS DSL to validate Go parser compatibility.
   - Introduce parity tests that evaluate the same AST on both interpreters and diff results.
2. Conformance harness
   - Build CLI to run Go + TS interpreters against shared suites (`examples/`, synthetic tests).
   - Track divergences and file bugs.
3. Documentation
   - Update interpreter READMEs, regenerate API docs (Go: `pkg.go.dev` ready comments).

## Phase 5 — Tree-sitter Parser Integration
1. Finalize canonical AST mapping for parser output.
2. Complete or rebuild `tree-sitter-able` grammar for v10 syntax.
3. Implement translators: parser AST → canonical AST → language-specific structures.
4. Add end-to-end tests: parse → interpret (TS + Go) → compare results.

## Ongoing Workstreams
- **Spec maintenance**: keep `spec/full_spec_v10.md` authoritative; log discrepancies in `spec/todo.md`.
- **Standard library**: coordinate with `stdlib/` efforts; ensure interpreters expose required builtin functions/types.
- **Developer experience**: cohesive documentation, examples, CI improvements (Bun + Go test jobs).
- **Future interpreters**: keep AST schema + conformance harness generic to support planned Crystal implementation.

## Immediate Next Actions
1. **Parser groundwork** – start the tree-sitter grammar effort now that the Go runtime + checker are canonical (design parser milestones, confirm AST emission format, bootstrap repository scaffolding).
2. **TypeScript parity follow-up** – mirror the finalised Go semantics (checker integration hooks, strict fixture mode) in the Bun interpreter so both runtimes expose the same behaviour ahead of parser integration.
3. **Documentation refresh** – finish updating onboarding/README/PARITY notes to reflect the Go interpreter’s canonical status and document the parser handoff expectations.

### Next steps (prioritized)
1. **Tree-sitter parser** – finalise grammar design, emit canonical AST JSON, and wire parser → interpreter integration tests.
2. **TypeScript checker enablement** – prepare the Bun interpreter to honour `ABLE_TYPECHECK_FIXTURES` once its checker lands, keeping parity traffic flowing during parser rollout.
3. **Spec & tooling updates** – keep `spec/full_spec_v10.md` and `spec/todo.md` in sync with the now-canonical Go behaviour, and ensure documentation/onboarding emphasises the new workflow.
4. **Performance & maintainability** – profile the Go runtime (method caches, environment lookups) and add CI coverage (Go + Bun + fixtures) once parser work starts emitting shared ASTs.

## Tracking & Reporting
- Update this plan as milestones progress; log decisions in `design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Short weekly status notes can live in `PLAN.md` under a future "Status Log" section when work begins.

## Status Log

### 2025-09-29
- Static import parity expanded: shared fixtures now cover wildcard/alias success + privacy errors along with a two-hop re-export chain; Go tests mirror the new scenarios (`interpreter10-go/pkg/interpreter/interpreter_test.go`).
- Dyn import alias metadata verified in Go (alias binds `DynPackageValue` with expected name/path) and environment error casing aligned with TypeScript so fixture diagnostics stay in sync.
- **Next focus:** extend re-export coverage to deeper chains + nested manifests, record dyn-import metadata for parity harnesses, and draft the Go goroutine scheduler design note before starting proc/spawn work.

### 2025-10-04
- Shared fixtures now cover generic function application, generic constraint errors, and struct functional updates; Go fixtures/tests load the new AST shapes.
- Go normalises destructuring/logical error strings and validates unknown generic interfaces during definition, matching TS diagnostics.
- **Next focus:** enforce generics at call time, port interface/impl diagnostics with fixtures, then resume the goroutine scheduler implementation.

### 2025-10-06
- Split the Go interpreter into feature-specific files (`eval_statements.go`, `eval_expressions.go`, `definitions.go`, `imports.go`, `impl_resolution.go`) so each subsystem can evolve independently; mirrored the split on the test side and updated the README with the new layout.
- Refreshed immediate actions/next steps to include finishing the modularisation work before resuming concurrency and remaining parity backlog.
- **Next focus:** migrate pattern-matching helpers/runtime utilities into dedicated files, then continue the parity checklist (interface diagnostics, concurrency scheduler).

### 2025-10-07
- Captured the cross-interpreter modularisation plan: TypeScript interpreter/tests will be split into feature-focused files mirroring the Go layout, with sub-1k LOC targets and parity test runs guarding behaviour.
- Immediate roadmap updated to tackle TS modularisation first, then stabilise the Go layout, run a parity audit, and sequence the Go reference handoff followed by the tree-sitter parser effort.
- **Next focus:** execute the TS interpreter split, keep fixtures/tests green, and document any deviations for the Go parity follow-up.

### 2025-10-14
- Hardened the Go concurrency suite with focused unit tests covering cancellation-before-start, cooperative cancellation observation (`proc_cancelled`), and future memoisation/error propagation (`interpreter10-go/pkg/interpreter/interpreter_concurrency_test.go`).
- Added a native `proc_flush()` helper in both interpreters so tests/fixtures can deterministically drain the cooperative scheduler, and exported shared fixtures for proc cancellation, `proc_cancelled` misuse, and future memoisation (`fixtures/ast/concurrency/**`).
- **Tests:** `GOCACHE=$(pwd)/.gocache go test ./...`
- **Next focus:** design a portable yield-fairness fixture now that `proc_flush()` exists, then mirror it across TS/Go once scheduling semantics are aligned.

### 2025-10-15

› I forgot that I want to add a safe navigation operator so that users can do foo?.bar?.baz() in the v10 language spec. does that conflict with any language features
  in the v10 language spec?
• No direct clash. In v10 the only places ? is spoken for are:

  - the nullable shorthand at the type level (?T → nil | T, §4.6.2 of spec/full_spec_v10.md);
  - the wider Option/Result discussion that builds on that shorthand (§11.2).

  Expression‑level syntax never uses a bare ?, identifiers cannot contain ?, and the propagation operator is the postfix !, not ?. Member access today is just ..
  Because ?. isn’t currently tokenised or given semantics, introducing it as a “safe member access” operator would sit alongside . without colliding with existing
  grammar or operators—you’d just need to teach the lexer/parser to prefer a ?. token where appropriate and spell out its precedence/behaviour in the spec.

  So you’re clear to add foo?.bar?.baz(); the main follow-up work is defining how it short‑circuits (likely returning nil to mesh with the existing ?T unions) and
  updating the spec/AST/interpreters accordingly.

- We will plan to add the safe member access operator to the language spec and the canonical AST as well as to the interpreters and the typechecker.

### 2025-10-21
- Split the Go constraint-solver expression tests so the heavy cases now live in `interpreter10-go/pkg/typechecker/checker_expressions_constraints_test.go`, keeping every test file comfortably under the 1k LOC ceiling.
- Shifted the struct/interface/module definitions into `interpreter10-go/pkg/ast/definitions.go`, leaving `pkg/ast/ast.go` below the 1k target without touching the AST contract.
- Normalised method-set instantiation by stripping substituted type parameters/where clauses in `type_utils.go`, which unblocked the `Formatter<T>` satisfaction tests and keeps `go test ./pkg/typechecker` green.
- Tightened the method-set regression by asserting the `"via method 'format'"` and `"via method set"` contexts appear on unsatisfied obligations, covering the new substitution behaviour.
- **Next focus:** continue the Go typechecker hardening (start with additional method-set + where-clause edge cases) and then move back into parser prep once the constraint solver dust settles.

### 2025-10-23
- Extended the Go typechecker’s constraint coverage with two `Self`-scoped method-set cases so `Formatter<string>`-style where clauses now propagate context into diagnostics; updated `typesEquivalentForSignature` to treat applied struct targets and struct instances as equivalent so those obligations resolve cleanly (`interpreter10-go/pkg/typechecker/implementation_validation.go`, `checker_expressions_constraints_test.go`).
- Verified the new scenarios with `GOCACHE=$(pwd)/.gocache go test ./pkg/typechecker`.
- Logged parser prep as the next major milestone: derive the concrete operator table, enumerate tokens, and scaffold the tree-sitter grammar in `parser10/tree-sitter-able`.
- **Next focus:** finish documenting the parser plan (token/precedence tables, AST mapping) and bootstrap the tree-sitter grammar without the deferred safe-navigation operator (planned for v11).
