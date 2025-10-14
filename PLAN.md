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
1. **Concurrency coverage hardening** – Go unit tests now cover cancellation paths, `proc_cancelled`, and future memoization, and shared fixtures exercise cancellation, `proc_cancelled`, and future memoisation; next step is adding yield-fairness coverage so TypeScript mirrors the strengthened semantics.
2. **Documentation & parity tracker refresh** – keep the concurrency design note up to date as implementation details settle, and record progress in `interpreter10-go/PARITY.md` so remaining gaps are obvious to future contributors.
3. **TypeScript scheduler alignment** – port the executor/handle semantics into the Bun interpreter so both runtimes expose identical `proc`/`spawn` behaviour ahead of fixture expansion.

### Next steps (prioritized)
1. **Go concurrency semantics** – Follow the executor-based plan to deliver `proc`/`spawn`, futures, cancellation, and helper natives backed by goroutines/channels.
2. **TypeScript parity follow-up** – Mirror the finalised Go semantics in the TS cooperative scheduler so fixtures continue to pass cross-interpreter.
3. **Reference interpreter transition** – Once concurrency lands and parity is revalidated, designate the Go runtime as canonical, documenting the handover criteria and updating onboarding/docs.
4. **Interface & generics completeness** – Finish higher-order interface inheritance, unnamed/named impl diagnostics, and visibility rules; mirror TS tests.
5. **Spec & privacy alignment** – Enforce the full privacy model in Go, broaden wildcard/dyn import semantics, and keep `spec/todo.md` in sync as gaps close.
6. **Performance & maintainability** – Optimise hot paths (environment lookups, method caches), add micro-benchmarks, and refine modular boundaries as the interpreter grows.
7. **Tree-sitter pipeline** – After both interpreters stabilise, begin mapping the canonical AST to parser production rules and build the tree-sitter grammar that emits interoperable AST JSON.
8. **Developer experience** – Update docs/examples, ensure PLAN/README stay accurate, and add coverage/CI targets (Go + Bun) once parity stabilises.

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

### 2025-10-18
- Hardened the Go concurrency suite with focused unit tests covering cancellation-before-start, cooperative cancellation observation (`proc_cancelled`), and future memoisation/error propagation (`interpreter10-go/pkg/interpreter/interpreter_concurrency_test.go`).
- Added a native `proc_flush()` helper in both interpreters so tests/fixtures can deterministically drain the cooperative scheduler, and exported shared fixtures for proc cancellation, `proc_cancelled` misuse, and future memoisation (`fixtures/ast/concurrency/**`).
- **Tests:** `GOCACHE=$(pwd)/.gocache go test ./...`
- **Next focus:** design a portable yield-fairness fixture now that `proc_flush()` exists, then mirror it across TS/Go once scheduling semantics are aligned.
