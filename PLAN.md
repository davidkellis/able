# Able Project Roadmap (v10 focus)

## Scope
- Maintain a canonical Able v10 language definition across interpreters and tooling, with the written specification remaining the source of truth for behaviour.
- Build and ship a Go reference interpreter that evaluates any AST produced by the v10 spec; this becomes the authoritative runtime implementation.
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
1. Port the remaining struct scenarios from the TypeScript suite to Go (mutating fields, static methods, functional update edge cases) and mirror any missing TS coverage.
2. Extend shared fixtures when new struct behaviour lands (e.g., functional update, method access) and keep both harnesses passing (`bun run scripts/run-fixtures.ts`, `go test ./pkg/interpreter`).
3. Continue folding design-note behaviour into `spec/full_spec_v10.md` as milestones complete; add new todos to `spec/todo.md` when gaps appear.

## Tracking & Reporting
- Update this plan as milestones progress; log decisions in `design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Short weekly status notes can live in `PLAN.md` under a future "Status Log" section when work begins.
