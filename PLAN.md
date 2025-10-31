# Able Project Roadmap (v10 focus)

## Scope
- Maintain a canonical Able v10 language definition across interpreters and tooling, with the written specification remaining the source of truth for behaviour.
- Prioritise the Go interpreter until it matches the TypeScript implementation feature-for-feature (the only intentional divergence is that Go uses goroutines/channels while TypeScript simulates concurrency with a cooperative scheduler).
- Keep the TypeScript and Go AST representations structurally identical so tree-sitter output can feed either runtime (and future targets like Crystal); codify that AST contract inside the v10 specification once validated.
- Document process and responsibilities so contributors can iterate confidently.

## Existing Assets
- `spec/full_spec_v10.md`: authoritative semantics.
- `interpreter10/`: Bun-based TypeScript interpreter + AST definition (`src/ast.ts`) and extensive tests.
- `interpreter10-go/`: Go interpreter and canonical Able v10 runtime. Go-specific design docs live under `design/` (see `go-concurrency.md`, `typechecker.md`).
- Legacy work: `interpreter6/`, assorted design notes in `design/`, early stdlib sketches.

## Current Focus — Parser & AST Coverage
- Track parser + AST fixture completeness via `design/parser-ast-coverage.md`.
- Sequence of work:
  1. Fill the shared AST fixture suite (`fixtures/ast`) so every spec feature is represented.
  2. Ensure both interpreters execute those fixtures with full behavioral assertions.
  3. Add exhaustive parser tests that round-trip the surface syntax to the canonical AST.
- Near-term emphasis: expand the Go interpreter's fixture parity harness (`interpreter10-go/pkg/interpreter/fixtures_parity_test.go`) until every language feature is exercised and green, then circle back to the TypeScript runtime and AST evaluator once Go coverage is exhaustive.
- Blocker work until the checklist is entirely `Done`; prioritize filling `TODO`/`Partial` rows before broader feature work resumes.
- Documented host-backed design for `Channel<T>`/`Mutex` (see `design/channels-mutexes.md`); interpreters must add channel/mutex value kinds and native helpers so stdlib externs can wire in the real semantics and the remaining AST fixtures can land.

**Latest parser progress (2025-01-XX):**
- Shared AST corpus now includes fixtures for `if/or`, assignment variants, breakpoint expressions, and lambda/trailing-lambda calls; TypeScript interpreter and tree-sitter grammar have been regenerated and verified (`bun run scripts/export-fixtures.ts`, `npx tree-sitter test`).
- Go parser harness still needs to wire these fixtures in once grammar support lands (notably `value! else { … }` and trailing-lambda metadata) — see `interpreter10/PLAN.md` for follow-up.

## Phase α — Channel & Mutex Runtime Bring-up
1. **Runtime Foundations (TS first, Go in lock-step)**
   - Extend `V10Value`/Go equivalents with `channel` and `mutex` variants + scheduling metadata.
   - Register native helpers (`__able_channel_new/send/recv/close`, `__able_mutex_new/lock/unlock`) and surface them via `prelude`.
   - Integrate send/recv/lock/unlock with the cooperative scheduler (block/yield/resume semantics, FIFO queues).
   - TypeScript runtime now mirrors Go-style blocking semantics, including waiter queues for send/receive, cancellation cleanup, and dedicated Bun tests (`test/concurrency/channel_mutex.test.ts`).
   - Mirror the implementation in Go (wrap native `chan`/`sync.Mutex` while preserving semantics).
   - Add unit tests covering core operations (buffered/unbuffered channels, close behaviour, mutex ownership errors) at the interpreter level.
2. **Stdlib Wiring**
   - Implement `Channel<T>` / `Mutex` structs in `stdlib/v10/concurrency` using extern bodies that call the native helpers; add high-level helpers (`with_lock`, iteration).
   - Ensure extern implementations exist for both TS and Go targets (Crystal host remains TODO until the runtime exists).
3. **Fixtures & Coverage**
   - Add AST fixtures covering channel send/recv across procs, closing, iteration, mutex lock/unlock, and failure cases.
   - Update `design/parser-ast-coverage.md` rows 121–122 to `Partial`/`Done` once fixtures land.
   - Record semantic notes in the spec (already captured) and cross-link from stdlib docs.

## Phase β — End-to-End AST Evaluation Completion *(AST evaluator coverage complete ✅)*
1. **Fixture audit**
   - Sweep `design/parser-ast-coverage.md` to confirm every language feature is represented; open tickets for any remaining `TODO` fixtures.
   - §12.5 fixtures now cover buffered ops, nil-channel cancellation, closed-channel errors, and mutex contention; next up are parser assertions and mirroring the behaviour in the TypeScript runtime.
   - TypeScript interpreter now shares the executor contract with Go (`src/interpreter/executor.ts`), so the concurrency fixture harness runs the same cancellation/yield scenarios across both runtimes.
   - TypeScript interpreter now runs the shared `pipes/member_topic` fixture and covers UFCS/placeholder pipe edges via unit tests; parity notes updated accordingly.
   - Add missing fixtures (e.g., remaining concurrency edge cases, select/timeout stubs if spec settles).
   - Fairness-focused concurrency fixtures now exercise round-robin proc/proc+future scheduling via the serial executor coordination helper; see `design/concurrency-fixture-plan.md` for details.
2. **Interpreter verification**
   - Ensure both interpreters run the entire fixture suite with assertions (value, stdout, error cases). Add targeted unit tests where fixture coverage is insufficient (e.g., helper corner cases).
   - Automate fixture execution in CI (TS `bun run scripts/run-fixtures.ts`, Go parity harness).
3. **Regression safety nets**
   - Consider property tests / fuzzing around `Channel` and `Mutex` semantics (enqueue/dequeue ordering, contention).
   - Document residual risks in `design/` for future contributors.

## Phase γ — Parser Exhaustiveness & AST Round-trip
1. **Parser test suite expansion**
   - For each feature row in `design/parser-ast-coverage.md`, add parser unit tests that assert the CST → AST transformation matches expectations (`pkg/parser` in Go, TS parser harness).
   - TypeScript now exercises fixtures with `source.able` by shelling out to the Go parser CLI (`interpreter10/test/parser/fixtures_parser.test.ts`); replace this bridge with a native Bun/tree-sitter harness once the TS parser bindings exist.
   - Include negative tests (syntax errors, malformed constructs) where the spec defines them.
2. **Round-trip validation**
   - Ensure CST → AST → interpreter evaluation produces the same results as fixtures for representative programs.
   - Add snapshot tests where appropriate to lock down AST shape (Go golden files vs TS snapshots).
3. **Completion criteria**
   - Mark coverage rows `Done` only after parser + AST fixtures + runtime behaviour align.
   - Update documentation/spec where parser behaviour needs clarification.

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
1. **Parser groundwork**
   - Start the tree-sitter grammar effort now that the Go runtime + checker are canonical (design parser milestones, confirm AST emission format, bootstrap repository scaffolding).
   - Replace the interim TypeScript harness that shells out to `go run cmd/parse-module` with a Bun-native parser binding once the grammar is exported.
   - Audit remaining fixtures missing `source.able` files and backfill them so the parity suite exercises the whole corpus.
2. **CLI & package manager follow-up**
   - Land the remaining CLI surface (`able build`, `able test`, `able check`, `able env`, formatter stubs).
   - Expand coverage for `able deps install/update` (multi-target manifests, selective updates) and document failure modes.
   - Finish stdlib packaging: flesh out `stdlib/v10/src/**` implementations and ensure they ship via the cache/manifest flow.
3. **Concurrency hardening**
   - Add stress fixtures/property tests for `proc_flush`, nil-channel cancellation, and fairness across the goroutine executor to match the design notes.
   - Document the semantics in `design/channels-mutexes.md` once the additional coverage lands.
4. **Documentation refresh** – finish updating onboarding/README/PARITY notes to reflect the Go interpreter’s canonical status and document the parser handoff expectations.

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

### 2025-10-30
- Comments are now ignored during parser → AST mapping for both interpreters.
  - ✅ Go: `ModuleParser` / helper utilities skip `comment`, `line_comment`, `block_comment` nodes and `TestParseModuleIgnoresComments` asserts the behaviour.
  - ✅ TypeScript: `tree-sitter-mapper` filters the same node types; `fixtures_mapper.test.ts` covers the mapping path and `fixtures_parser.test.ts` ensures the raw grammar parses comment-heavy sources.
- TODO: audit remaining parser/mapping gaps per `design/parser-ast-coverage.md` (pipes/topic combos, functional updates, etc.) and backfill fixtures/tests.
- TODO: propagate comment ignoring to any remaining mapper helpers (struct literals, pattern visitors) still traversing raw children.
- TODO: Build end-to-end coverage across **all three facets** (parsing, tree → AST mapping, AST evaluation) for both interpreters. Use the coverage table to drive fixture additions, parser assertions, and runtime tests until every spec feature is green.
- TODO: Extend the **typechecker** suites (Go + TS) so they verify type rules and inference across modules. Assemble an exhaustive inference corpus exercising expression typing, generics, interfaces/impls, and cross-module reconciliation; ensure these scenarios are evaluated alongside runtime fixtures.

### 2025-10-31
- Regenerated the tree-sitter-able artifacts with the freshly rebuilt grammar (interface-composition fix now baked into `parser.c`/`.wasm`) using the local Emscripten toolchain; no diff surfaced, confirming the repo already carried the correct bits.
- Cleared local Go build caches (`.gocache`, `interpreter10-go/.gocache`) and re-ran `GOCACHE=$(pwd)/.gocache go test ./pkg/parser` to mimic CI picking up the refreshed grammar without stale entries.
- ACTION: propagate the cache-trim guidance to CI docs if flakes recur; otherwise move on to the remaining parser fixture gaps (`design/parser-ast-coverage.md`).
