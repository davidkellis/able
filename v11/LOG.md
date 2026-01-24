# Able Project Log

## 2026-01-24 — Iterator interface returns + constraint-arity fixture cleanup
- Go interpreter: treat Iterator interface return values as iterators during for/each by accepting `IteratorValue` in `adaptIteratorValue`.
- Fixtures: removed duplicate constraint-interface-arity diagnostics from exported manifests via the TS fixture definitions; re-exported fixtures.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict go test ./pkg/interpreter -run 'TestFixtureParityStringLiteral/errors/constraint_interface_arity' -count=1`; `./run_all_tests.sh --version=v11`.

## 2026-01-24 — Interface dictionary arg dispatch + fixture expansion
- Go interpreter: coerce interface-typed generic values into interface dictionaries so interface arguments are preserved for bindings, params, and return coercions.
- Fixtures: added interface dictionary exec coverage for default chains, overrides, named impl + inherent method calls, interface inheritance, interface-arg dispatch (bindings/params/returns), and union-target dispatch; added AST error fixtures for ambiguous impl constraints + missing interface methods; updated exec coverage index and typecheck baseline.
- Tests: `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_11_interface_generic_args_dispatch -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_12_interface_union_target_dispatch -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_13_interface_param_generic_args -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_14_interface_return_generic_args -count=1`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_13_interface_param_generic_args bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_14_interface_return_generic_args bun run scripts/run-fixtures.ts`.

## 2026-01-24 — Named impl method resolution fix
- Interpreters (TS + Go): attach named-impl context to impl methods so default methods (and peers) can resolve sibling methods via `self.method()` in the same impl.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_05_interface_named_impl_defaults bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_06_interface_generic_param_dispatch bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_05_dynimport_interface_dispatch bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_05_interface_named_impl_defaults`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_06_interface_generic_param_dispatch`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/13_05_dynimport_interface_dispatch`.

## 2026-01-15 — Hash/Eq fixture and test coverage
- Fixtures: added AST fixtures for primitive hashing, kernel hasher availability, custom Hash/Eq, and collision handling; added exec fixtures for primitive hashing plus custom Hash/Eq + collisions; updated exec coverage index.
- Tests: added TS + Go unit coverage for hash helper builtins and kernel HashMap dispatch (custom + collision keys).
- Tests not run (edited code + fixtures only).

## 2026-01-15 — Remove host hasher bridges
- Kernel: dropped the `__able_hasher_*` extern declarations and the unused `HasherHandle` alias so hashing flows through `KernelHasher` only.
- Interpreters: removed host hasher state/builtins from Go + TypeScript, along with runtime stub and typechecker builtin entries.
- Docs/spec: scrubbed hasher bridge references from the kernel contract and extern execution/design notes.
- Tests not run (edited code + docs only).

## 2026-01-15 — Kernel Hash/Eq runtime alignment
- Kernel: added primitive `Eq`/`Ord`/`Hash` impls (ints/bool/char/String) plus float-only `PartialEq`/`PartialOrd`, and wired the Able-level FNV-1a hasher with raw byte helpers.
- Stdlib: trimmed duplicate interface impls and routed map hashing through the sink-style `Hash.hash` API.
- Interpreters: hash map kernels now dispatch `Hash.hash`/`Eq.eq`; numeric equality follows IEEE semantics; Go/TS typecheckers exclude floats from `Eq`/`Hash`.
- Fixtures: added float equality + hash-map key rejection exec coverage.
- Tests not run (edited code + docs only).

## 2026-01-15 — Kernel interfaces + Hash/Eq plan
- Added the kernel interface/primitive Hash/Eq design plan plus stdlib linkage notes (`v11/design/kernel-interfaces-hash-eq.md`, `v11/design/stdlib-v11.md`).
- Updated `spec/TODO_v11.md` and expanded the detailed work breakdown in `PLAN.md`.
- Tests not run (planning/doc updates only).

## 2026-01-15 — Manual syntax alignment
- Updated the v11 manual docs to match spec lexing and pipe semantics (line comments use `##`, string literals can use double quotes or backticks with interpolation, and pipe docs no longer mention a `%` topic token) in `v11/docs/manual/manual.md` and `v11/docs/manual/variables.html`.
- Tests not run (docs-only changes).

## 2026-01-15 — Primitive Hash/Eq constraints
- Updated the TS typechecker to treat primitive numeric types as satisfying `Hash`/`Eq` constraints (matching Go) and adjusted the example to iterate directly over `String` so `for` sees an `Iterable` (`v11/interpreters/ts/src/typechecker/checker/implementation-constraints.ts`, `.examples/foo.able`).
- Tests: `./v11/ablets .examples/foo.able`; `./v11/ablego .examples/foo.able`.

## 2026-01-14 — Go module splits
- Split the Go constraint solver, literals, and member-access helpers into focused files to keep modules under 900 lines (`v11/interpreters/go/pkg/typechecker/constraint_solver_impls.go`, `v11/interpreters/go/pkg/typechecker/constraint_solver_methods.go`, `v11/interpreters/go/pkg/typechecker/statement_checker.go`, `v11/interpreters/go/pkg/typechecker/member_access_methods.go`, `v11/interpreters/go/pkg/typechecker/member_access_matching.go`) and trimmed `v11/interpreters/go/pkg/typechecker/constraint_solver.go`, `v11/interpreters/go/pkg/typechecker/literals.go`, and `v11/interpreters/go/pkg/typechecker/member_access_helpers.go`.
- Split the Go extern host into focused files (`v11/interpreters/go/pkg/interpreter/extern_host_cache.go`, `v11/interpreters/go/pkg/interpreter/extern_host_builder.go`, `v11/interpreters/go/pkg/interpreter/extern_host_coercion.go`, `v11/interpreters/go/pkg/interpreter/extern_host_module.go`) and trimmed `v11/interpreters/go/pkg/interpreter/extern_host.go`.
- Split the Go interpreter type helpers into focused files (`v11/interpreters/go/pkg/interpreter/interpreter_type_info.go`, `v11/interpreters/go/pkg/interpreter/interpreter_interface_lookup.go`, `v11/interpreters/go/pkg/interpreter/interpreter_type_matching.go`, `v11/interpreters/go/pkg/interpreter/interpreter_type_coercion.go`) and removed the oversized `v11/interpreters/go/pkg/interpreter/interpreter_types.go`.
- Split the Go interpreter operations helpers into focused files (`v11/interpreters/go/pkg/interpreter/interpreter_operations_dispatch.go`, `v11/interpreters/go/pkg/interpreter/interpreter_operations_ratio.go`, `v11/interpreters/go/pkg/interpreter/interpreter_operations_arithmetic.go`, `v11/interpreters/go/pkg/interpreter/interpreter_operations_compare.go`) and removed the oversized `v11/interpreters/go/pkg/interpreter/interpreter_operations.go`.
- Tests not run.

## 2026-01-12 — Stdlib fs/io/os coverage
- Added stdlib spec tests for `able.fs`, `able.io`, and `able.os` covering open/read/write, directory ops, metadata, buffered IO, and environment/cwd helpers (`v11/stdlib/tests/fs.test.able`, `v11/stdlib/tests/io.test.able`, `v11/stdlib/tests/os.test.able`).
- Tests not run.

## 2025-12-30 — Drop snapshots from testing framework
- Removed snapshot matchers/stores and the `--update-snapshots` flag across stdlib tests, docs, and CLI design notes; deleted the snapshot store design doc and scrubbed spec references.
- Regenerated `v11/stdlib/src/spec.able` after removing snapshot exports from the spec DSL.
- Tests not run.

## 2025-12-30 — Testing module suffix policy
- Added the testing framework section to `spec/full_spec_v11.md`, codifying `.test.able` and `.spec.able` as test modules plus the `able test` contract.
- Cleared the suffix-policy open decision in `v11/design/testing-plan.md` and removed the corresponding TODO from `PLAN.md`.
- Tests not run.

## 2025-11-11 — Stdlib Module Search Paths
- **Pipe semantics parity**: Added the `pipes/multi_stage_chain` AST fixture so multi-stage pipelines that mix `%` topic steps, placeholder-built callables, and bound methods stay covered; `bun run scripts/run-fixtures.ts` (TypeScript) and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` (Go) stay green with no parity divergences observed.
- **Typechecker strict fixtures**: TypeScript’s checker now hoists struct identifiers for static calls, predeclares `:=` bindings (so future handles can reference themselves), binds iterator driver aliases plus struct/array pattern destructures, and hides private package members behind the standard “has no symbol” diagnostic. The full fixture suite now passes under `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts`, and manifests/baselines were updated where diagnostics are expected.
- **Dynamic interface collections & iterables**: added the shared fixture `interfaces/dynamic_interface_collections` (plus exporter wiring) so both interpreters prove that range-driven loops and map-like containers storing interface values still choose the most specific impl even when union targets overlap. `bun run scripts/run-fixtures.ts` is green; Go parity remains blocked by existing fixture failures (`functions/hkt_interface_impl_ok`, `imports/static_alias_private_error`).
- **Privacy & import spec gaps**: enforced package privacy across both interpreters/typecheckers (selectors, aliases, wildcard bindings), tightened dynimport placeholder handling (TypeScript + Go parity plus fixture coverage), and updated `spec/full_spec_v11.md` with the canonical `Future` handle definitions so runtimes and tooling share the same ABI (`FutureStatus`, `FutureError`, `status/value/cancel` contracts).
- **Interface & impl completeness**: higher-kinded/visibility edge cases now have end-to-end coverage. Overlapping impl ambiguity diagnostics land in both interpreters/typecheckers, TypeScript enforces method-set backed constraints with parity coverage, both interpreters honour interface self-type patterns (including higher-kinded targets) and reject bare constructors unless `for …` is declared, and parser/exporter propagate the `for …` patterns so fixtures surface the diagnostics automatically. Shared AST fixtures (`errors/interface_self_pattern_mismatch`, `errors/interface_hkt_constructor_mismatch`, plus the positive `functions/hkt_interface_impl_ok`) keep both runtimes + typecheckers aligned.

- Future cancellation coverage is complete: both interpreters expose `Future.cancel()`, cancellation transitions produce the `Cancelled` status, and the new `concurrency/future_cancel_nested` fixture exercises a task awaiting a spawned future that is cancelled mid-flight. The exporter and strict harness generate the fixture automatically, keeping the TypeScript + Go runtimes/typecheckers in lockstep for nested cancellation chains across both executors.
- Channel error helpers now emit the stdlib error structs in both runtimes. TypeScript + Go runtimes route every channel error through `ChannelClosed`, `ChannelNil`, or `ChannelSendOnClosed`, and new Bun/Go tests cover the behaviour (`v11/interpreters/ts/test/concurrency/channel_mutex.test.ts`, `interpreter-go/pkg/interpreter/interpreter_channels_mutex_test.go`).
- Lightweight executor diagnostics are now wired into both runtimes: the new `future_pending_tasks()` helper surfaces cooperative queue length via `CooperativeExecutor.pendingTasks`, Go’s serial and goroutine executors expose the same data, and coverage comes from a dedicated AST fixture (`concurrency/future_executor_diagnostics`) plus Bun/Go unit tests. The spec, manuals, and `design/concurrency-executor-contract.md` now describe the helper so fixtures can assert drain behaviour deterministically.
- String host bridge externs (`__able_string_from_builtin`, `__able_string_to_builtin`, `__able_char_from_codepoint`) are wired into both interpreters and typecheckers with dedicated tests (`v11/interpreters/ts/test/string_host.test.ts`, `interpreter-go/pkg/interpreter/interpreter_string_host_test.go`).
- Hasher externs (`__able_hasher_create/write/finish`) now back the stdlib hash maps across TypeScript and Go, complete with parity tests and stub support for tooling.
- Added the `concurrency/channel_error_rescue` AST fixture and exposed `Error` member access (message/value) in both interpreters so Able code can assert the struct payloads produced by the channel helpers.
- Go parity now runs the new `concurrency/channel_error_rescue` fixture under the goroutine executor, and a dedicated Go test (`TestChannelErrorRescueExposesStructValue`) verifies that rescuing channel errors exposes the struct payload via `err.value`.
- Added the `errors/result_error_accessors` AST fixture so both interpreters exercise `err.message()/cause()/value` inside `!T else { |err| ... }` flows; fixture exporter + TS harness updated accordingly.
- Go typechecker now recognises `Error.message()`, `.cause()`, and `.value`, and the spec documents the runtime-provided `Error.value` payload hook; the typechecker baseline entry for `channel_error_rescue` was removed once diagnostics cleared.
- Future runtime errors now record their cause payloads in both interpreters, the new `concurrency/future_error_cause` fixture exercises `err.cause()` end-to-end, and matching Bun/Go tests keep the regression harness green.
- Generator laziness parity closed: iterator continuations now cover if/while/for/match across both runtimes, stdlib helpers (`stdlib/src/concurrency/channel.able`, `stdlib/src/collections/range.able`) use generator literals, and new fixtures (`fixtures/ast/control/iterator_*`, `fixtures/ast/stdlib/channel_iterator`, `fixtures/ast/stdlib/range_iterator`) keep the shared harness authoritative.
- Automatic time slicing verified for long-running futures: the new `concurrency/future_time_slicing` fixture proves that handles without explicit `future_yield()` still progress under repeated `future_flush()` calls, capturing both the intermediate `Pending` status and the eventual resolved value across runtimes.
### AST → Parser → Typechecker Completion Plan _(reopen when new AST work appears)_
- Full sweep completed 2025-11-06 (strict fixture run, Go interpreter suite, Go parser harness, and `bun test` all green). Archive details in `LOG.md`; bring this plan back only if new AST/syntax changes introduce regressions.

## 2025-11-09 — Executor Diagnostics Helper
- Added the `future_pending_tasks()` runtime helper so Able programs/tests can observe cooperative executor queue depth. TypeScript wires it through `CooperativeExecutor.pendingTasks()` while the Go runtime surfaces counts from both the serial and goroutine executors (best-effort on the latter via atomic counters). The helper is registered with both typecheckers so Warn/Strict fixture runs understand the signature.
- New coverage keeps the helper honest: Bun unit tests exercise the helper directly (`v11/interpreters/ts/test/concurrency/proc_spawn_scheduling.test.ts`), Go gains matching tests (`TestProcPendingTasksSerialExecutor`, `TestProcPendingTasksGoroutineExecutor`), and a serial-only AST fixture (`fixtures/ast/concurrency/future_executor_diagnostics/`) ensures both interpreters prove that `future_flush` drains the cooperative queue.
- Spec + docs now describe the helper alongside `future_yield`/`future_flush` (see `spec/full_spec_v11.md`, `docs/manual/*.md/html`, `design/concurrency-executor-contract.md`, and `AGENTS.md`), and the PLAN TODO for “Concurrency ergonomics” is officially closed out.

## 2025-11-08 — Fixture Diagnostics Parity Enforcement
- `v11/interpreters/ts/scripts/run-parity.ts` now diffs typechecker diagnostics for every AST fixture so warn/strict parity runs catch unexpected checker output even when manifests do not declare expectations. The parity JSON report captures the mismatched diagnostics to speed up triage.
- Added `test/scripts/parity/fixtures_compare.test.ts` to cover the helper logic that determines when diagnostics mismatches should fail parity.
- Go’s typechecker now treats unannotated `self` parameters inside `methods {}` / `impl` blocks as `Self`/concrete receiver types and seeds iterator literals with the implicit `gen` binding, eliminating the extra diagnostics that used to appear only in Go’s warn/strict runs.
- Added Go regression tests (`pkg/typechecker/checker_impls_test.go`, `pkg/typechecker/checker_iterators_test.go`) so implicit `self` bindings and iterator generator helpers stay covered.
- Bun’s typechecker mirrors the same behaviour: iterator literals now predefine the implicit `gen` binding, implicit `self` parameters default to `Self`, and new tests (`test/typechecker/method_sets.test.ts`, `test/typechecker/iterators.test.ts`) keep the coverage locked in.
- The Go fixture runner and CLI now continue evaluating programs in warn/strict modes even when diagnostics are reported (matching the Bun harness), so parity checks capture both the expected runtime results and the shared diagnostics payloads.

## 2025-11-07 — AST → Parser → Typechecker Cycle Revalidated
- Added proc handle memoization fixtures (success + cancellation) and ensured both interpreters plus the Go parser harness run them under strict typechecking (`bun run scripts/run-fixtures.ts`, `cd interpreter-go && GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser`).
- Verified the full suite remains green (`./run_all_tests.sh --typecheck-fixtures=strict`, `bun test`, `cd interpreter-go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`), keeping the Priority 0 gate satisfied.
- Updated PLAN.md to mark the current AST → Parser → Typechecker cycle complete and advance the focus to Phase α (Channel & Mutex runtime bring-up).
- Added stdlib specs for channel/mutex behaviour (`stdlib/tests/concurrency/channel_mutex.test.able`) so the Phase α bullet “add unit tests covering core operations” is now satisfied.

## 2025-11-07 — Serial Executor Future Reentrancy
- Go’s SerialExecutor now exposes a `Drive` helper that runs pending tasks inline, so nested `future.value()` calls no longer deadlock and match the TypeScript scheduler semantics. The helper steals the targeted handle from the deterministic queue, executes it re-entrantly (including repeated slices when `future_yield` fires), and restores the outer task context once the awaited handle resolves.
- Fixtures `concurrency/future_value_reentrancy` and `concurrency/future_flush_fairness` now pass under the Go interpreter’s serial executor, keeping the newly added fairness/re-entrancy corpus green for both runtimes (`ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts`, `cd interpreter-go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`).
- PLAN immediate actions trimmed: the “Concurrency stress coverage & docs” placeholder has been cleared now that the blocking fixtures run cleanly; follow-up concurrency work can graduate to design docs instead of the top-level plan.
- Added a Go-side regression test (`TestSerialExecutorFutureValueReentrancy`) that mirrors the new fixture to ensure nested `future.value()` calls stay green even if future contributors touch the executor; the design note `design/go-concurrency-scheduler.md` now documents the inline-driving behaviour and follow-up doc work.
- Added a companion fixture (`concurrency/future_value_reentrancy`) plus TypeScript exporter wiring + Go regression test (`TestSerialExecutorProcValueReentrancy`) so both interpreters exercise nested `future.value()` waits under the serial executor; parser coverage tables were updated accordingly.
- Documented the goroutine-executor fairness contract (what `future_yield`/`future_flush` mean under `GoroutineExecutor`, how we rely on Go’s scheduler, and when tests must fall back to the serial executor) to close out the remaining PLAN follow-up (`design/go-concurrency-scheduler.md`).
- Updated the v11 spec to codify re-entrant `future.value()` semantics so both interpreters (and future targets) must guarantee deadlock-free nested waits (§12.2.5 “Re-entrant waits”).
- Added direct Go unit tests (`TestProcHandleValueMemoizesResult`, `TestProcHandleValueCancellationMemoized`) to ensure repeated `value()` calls return memoized results/errors even after cancellation, satisfying the remaining “exercise repeated value() paths” item from the concurrency plan.
- Introduced the `concurrency/future_value_memoization` fixture (plus exporter wiring) so both interpreters prove future handles memoize values, and updated the Go parity harness to run it with the goroutine executor.
- Added `concurrency/future_value_cancel_memoization` to assert that cancelled future handles return the same error for repeated `value()` calls without re-running their bodies; the fixture exporter, AST corpus, and fixture run all cover this scenario now.

## 2025-11-06 — Tree-Sitter Mapper Modularization Complete
- Go declarations/patterns/imports all run through the shared `parseContext`, removing the last raw-source helpers so both runtimes share one parser contract (`interpreter-go/pkg/parser/{declarations,patterns,statements,expressions}_parser*.go`).
- TypeScript parser README now calls out the shared context, and `PLAN.md` logged the Step 6 regression sweep so future contributors know the refactor is locked in (`v11/interpreters/ts/src/parser/README.md`, `PLAN.md`).
- Wrapper exports like `parseExpression(node, source)` / `parseBlock` / `parsePattern` were removed; all Go parser consumers now flow through the context pipeline.
- Tests: `./run_all_tests.sh --typecheck-fixtures=warn` and `cd interpreter-go && GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser`.
- Follow-up: confirm any remaining helpers that still accept raw source (e.g., host-target parsing) genuinely require it before migrating them to `parseContext`.

## 2025-11-03 — AST → Parser → Typechecker Completion Sweep
Status: ✅ Completed. Parser coverage table reads fully `Done`, TypeScript + Go checkers share the bool/comparison semantics, and `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts` executes cleanly after wiring the remaining builtin signatures, iterator annotations, and error fixture diagnostics into the manifests.

Open items (2025-11-02 audit):
- [x] Iterator literals dropped both the optional binding identifier and optional element type annotation. Update both ASTs + parsers/interpreters so the metadata survives round-trips and execution.
- [x] Teach both typecheckers to honor the iterator element annotation so every `yield` matches the declared element type (TS `src/typechecker/checker.ts`, Go `pkg/typechecker/literals.go` + iterator tests).
- [x] Carry iterator element metadata through `for` loops so typed loop patterns validate across array/range/iterator inputs (TS `src/typechecker/checker.ts`, Go `pkg/typechecker/{control_flow,literals,type_utils}.go` + new cross-loop tests).
- [x] Give the TS checker parity with Go for block/proc/spawn typing and array/range literal inference so all three stages (AST → parser → typechecker) agree on the element/result metadata (§6.8, §6.10, §12.2).
- [x] Enforce async-only builtins (`future_yield`) and add concurrency smoke tests so TS emits the same diagnostics as Go when authors call scheduler helpers outside `proc`/`spawn`.
- [x] Implement `if`/`while` diagnostics + inference in the TS checker so control-flow expressions match the Go implementation (§8.1/§8.2).
- [x] Mirror Go's match/rescue guard enforcement in the TS checker (§8.1.2 / §11.3).
- [x] Enforce package privacy + import diagnostics in the TS checker so private packages/definitions behave identically to Go (updated `imports.test.ts` + package summaries carry `visibility` metadata).

## Historical Status Notes

### 2025-10-30
- Comments are now ignored during parser → AST mapping for both interpreters.
  - ✅ Go: `ModuleParser` / helper utilities skip `comment`, `line_comment`, `block_comment` nodes and `TestParseModuleIgnoresComments` asserts the behaviour.
  - ✅ TypeScript: `tree-sitter-mapper` filters the same node types; `fixtures_mapper.test.ts` covers the mapping path and `fixtures_parser.test.ts` ensures the raw grammar parses comment-heavy sources.
- TODO: audit remaining parser/mapping gaps per `design/parser-ast-coverage.md` (pipes/topic combos, functional updates, etc.) and backfill fixtures/tests.
- DONE: comment skipping now wired through struct literals, struct patterns, and related mapper helpers across both runtimes.
- TODO: Build end-to-end coverage across **all three facets** (parsing, tree → AST mapping, AST evaluation) for both interpreters. Use the coverage table to drive fixture additions, parser assertions, and runtime tests until every spec feature is green.
- TODO: Extend the **typechecker** suites (Go + TS) so they verify type rules and inference across modules. Assemble an exhaustive inference corpus exercising expression typing, generics, interfaces/impls, and cross-module reconciliation; ensure these scenarios are evaluated alongside runtime fixtures.

### 2025-10-31
- Regenerated the tree-sitter-able artifacts with the freshly rebuilt grammar (interface-composition fix now baked into `parser.c`/`.wasm`) using the local Emscripten toolchain; no diff surfaced, confirming the repo already carried the correct bits.
- Cleared local Go build caches (`.gocache`, `interpreter-go/.gocache`) and re-ran `GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser` to mimic CI picking up the refreshed grammar without stale entries.
- ACTION: propagate the cache-trim guidance to CI docs if flakes recur; otherwise move on to the remaining parser fixture gaps (`design/parser-ast-coverage.md`).
- Mirrored the TypeScript placeholder auto-lift guardrails inside the Go interpreter so pipe placeholders evaluate eagerly, keeping the shared `pipes/topic_placeholder` fixture green.
- Parser sweep: both the TypeScript mapper and Go parser now skip inline comments when traversing struct literals, call/type argument lists, and struct definitions, with fresh TS/Go tests guarding the behaviour.
- TypeScript checker scaffold landed: basic environment/type utilities exist under `v11/interpreters/ts/src/typechecker`, exported via the public index, and fixtures respect `ABLE_TYPECHECK_FIXTURES` ahead of the full checker port.
- TypeScript checker now emits initial diagnostics for logical operands, range bounds, struct pattern validation, and generic interface constraints so the existing error fixtures pass under `ABLE_TYPECHECK_FIXTURES=warn`.

### 2025-11-01
- TypeScript checker grew a declaration-collection sweep that registers interfaces, structs, methods, and impl blocks before expression analysis, mirroring the Go checker’s phase ordering.
- Implementation validation now checks that every `impl` supplies the interface’s required methods, that method generics/parameters/returns mirror the interface signature, and flags stray definitions; only successful impls count toward constraint satisfaction.
- Canonicalised type formatting to the v11 spec (`Array i32`, `Result string`) and keyed implementation lookups by the fully-instantiated type so generic targets participate in constraint checks.
- Extended the TypeScript checker’s type info to capture `nullable`, `result`, `union`, and function signatures, and taught constraint resolution to recognise specialised impls like `Show` for `Result string`, with fresh tests covering the new cases.
- Added focused Bun tests under `test/typechecker/` plus the `ABLE_TYPECHECK_FIXTURES` harness run to lock in the new behaviour and guard future regressions.
- Mirrored the Go checker and test suite to use the same spec-facing type formatter and wrapper-aware constraint logic so diagnostics now reference `Array i32`, `string?`, etc., and added parity tests (`interpreter-go/pkg/typechecker/*`).

### 2025-11-02
- The Bun CLI suite (`v11/interpreters/ts/test/cli/run_module_cli.test.ts`) now covers multi-file packages, custom loader search paths via the new `ABLE_MODULE_PATHS` env, and strict vs warn typecheck enforcement so the ModuleLoader refactor stays covered without pulling `stdlib/` into every run.
- Introduced the `able test` skeleton inside `scripts/run-module.ts`: it parses the planned flags/filters, materialises run options + reporter selection, and prints a deterministic plan summary before exiting with code `2` while the stdlib testing packages remain unparsable. (See `design/testing-cli-design.md` / `design/testing-cli-protocol.md`.)
- Extracted the shared package-scanning helpers (`discoverRoot`, `indexSourceFiles`, etc.) into `scripts/module-utils.ts` so other tooling (fixtures runner, future harnesses) can reuse the multi-module discovery logic without duplicating it.
- **Deferral noted:** full stdlib/testing integration is still on pause until the parser accepts `stdlib/src/test/*`; once that unblocks, wire the CLI skeleton into the `able.test` harness per the design notes.

### 2025-11-05
- Step 6 regression sweep ran end-to-end: `./run_all_tests.sh --typecheck-fixtures=warn` stayed green post-refactor, and `GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser` uncovered/validated the Go-side AST gaps.
- Go AST parity improvements: `FunctionCall.arguments` now serialises as an empty array when no args exist, `NilLiteral` carries `value: null`, and break/continue statements omit label/value when not supplied (matching the TS mapper contract).
- Parser helper normalisation no longer nukes empty slices before fixture comparison, and parameter lists default to `[]`, which brought the channel/mutex fixtures back in sync.
- Fixture `structs/functional_update` gained explicit `\"isShorthand\": false` flags so both interpreters agree on struct literal metadata going forward.

### 2025-11-06
- Go CLI now exposes `able check` alongside `able run`, sharing the manifest/target resolution pipeline. Both commands surface ProgramChecker diagnostics + package summaries and fail fast when typechecking reports issues, keeping the TypeScript + Go CLIs aligned.
- Added dedicated CLI tests to cover `able check` success/failure cases so future refactors keep the new mode wired through manifest resolution, typechecker reporting, and exit codes.

### 2025-11-07 — Phase α (Channel/Mutex) Completion
- Audited channel/mutex stdlib wiring across both runtimes: helper registration, typechecker signatures, fixtures, and prelude exports now match; no AST or scheduler drift detected.
- Added Bun smoke tests for nil-channel cancellation and mutex re-entry errors (`v11/interpreters/ts/test/concurrency/channel_mutex.test.ts`) to mirror the Go parity suite.
- Documented the audit and captured the remaining TODO (map native errors to `ChannelClosed`/`ChannelNil`/`ChannelSendOnClosed`) in `design/channels-mutexes.md` and `spec/TODO.md`.
- Cleared `Phase α` from the active roadmap; next milestone is Phase 4 (cross-interpreter parity/tooling).

### 2025-11-07 — Fixture Parity Harness (Phase 4 Kick-off)
- Added a Go CLI entry point (`cmd/fixture`) that evaluates a single AST fixture and emits normalized JSON (result kind/value, stdout, diagnostics) with respect to `ABLE_TYPECHECK_FIXTURES`. The helper reuses the interpreter infrastructure and supports serial/goroutine executors while sandboxing the Go build cache.
- Refactored the TypeScript fixture loader into `scripts/fixture-utils.ts` so both the CLI harness and Bun tests can hydrate modules, install runtime stubs, and intercept `print` output consistently.
- Rebuilt `run-fixtures.ts` on top of the shared utilities (no behavior change) to keep fixture execution logic single-sourced.
- Introduced a Bun parity test (`test/parity/fixtures_parity.test.ts`) that exercises a representative slice of the shared fixture corpus (currently 20 fixtures across basics + concurrency) against both interpreters and asserts matching results/stdout via the new Go CLI.

### 2025-11-08 — Parity CLI Reporting
- Added `v11/interpreters/ts/scripts/run-parity.ts`, reusable parity helpers, and Bun parity suites that now share the same execution/diffing logic across AST fixtures and curated examples.
- `run_all_tests.sh` now invokes the parity CLI so local + CI runs execute the same cross-interpreter verification and drop a JSON report at `tmp/parity-report.json` for machine-readable diff tracking; `tmp/` landed in `.gitignore` to keep artifacts out of commits.
- The helper script also honors `ABLE_PARITY_REPORT_DEST` or `CI_ARTIFACTS_DIR` so pipelines can copy the parity JSON into their artifact buckets without bespoke wrapper scripts.
- Updated `v11/interpreters/ts/README.md` and `interpreter-go/README.md` with parity CLI instructions, env knobs (`ABLE_PARITY_MAX_FIXTURES`, `ABLE_PARITY_REPORT_DEST`), and guidance on keeping the cross-interpreter harness green.
- Added Go package docs (`pkg/interpreter/doc.go`, `pkg/typechecker/doc.go`) plus README guidance on regenerating `go doc`/pkg.go.dev pages so the documentation workstream is unblocked.
- Landed `dynimport_parity` and `dynimport_multiroot` in `v11/interpreters/ts/testdata/examples/` to cover dynamic package aliasing, selector imports, and multi-root dynimport scenarios end-to-end; the parity README + plan now list them alongside the other curated programs, and the Go CLI + Bun harness honor `ABLE_MODULE_PATHS` when resolving shared deps.
- Authored `docs/parity-reporting.md` and linked it from the workspace README so CI pipelines know how to persist `tmp/parity-report.json` via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR`.
- The Go CLI (`cmd/able`) now honors `ABLE_MODULE_PATHS` in addition to `ABLE_PATH`, with new tests ensuring the search-path env works; stdlib docs reference the alias so multi-root dynimport scenarios can rely on a single env knob across interpreters.
- Fixed the `..`/`...` range mapping bug in both parsers so inclusive ranges now follow the spec (TS + Go parser updates, new parser unit tests, interpreter for-loop regression tests, and fizzbuzz-style parity coverage).

### Phase 5 Foundations — Parser Alignment
- Canonical AST mapping now mirrors the fixture corpus across both runtimes. The TypeScript mapper’s fixture parity suite (`bun test test/parser/fixtures_mapper.test.ts`) and the Go parser harness (`go test ./pkg/parser`) stay green, so every tree-sitter node shape maps to the shared AST contract with span/origin metadata.
- `tree-sitter-able` grammar coverage is complete for the v11 surface (see `design/parser-ast-coverage.md`); new syntax is added directly with fixture+cXPath tests so the grammar remains authoritative.
- Translators and loaders are live in both interpreters: TypeScript’s `ModuleLoader` and Go’s `driver.Loader` now ingest `.able` source via tree-sitter, hydrate canonical AST modules, and feed them to their respective typechecker/interpreter pipelines.
- End-to-end parse → typecheck → interpret tests exercise both runtimes: `ModuleLoader pipeline with typechecker` (Bun) covers the TS path, and `pkg/interpreter/program_pipeline_test.go` drives the Go loader/interpreter via `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`.
- Diagnostic coverage now rides on the same pipelines: the new Bun test asserts missing import selectors surface typechecker errors before evaluation, and the Go suite verifies that `EvaluateProgram` halts (or proceeds when `AllowDiagnostics` is set) when return-type violations are reported.

### 2026-01-14
- Go typechecker implementation validation now skips prelude impls, aligns impl target labels with TS by alias-expanding and wildcard-normalizing target type expressions, and tracks local type declarations to avoid rewriting in-module types.
- Where-clause mismatch diagnostics for impl methods now point at the impl definition node to match fixture baselines.
- Tests: `./run_all_tests.sh --version=v11`.
- Ran `./run_stdlib_tests.sh` (TS + Go stdlib suites) to confirm current stdlib coverage remains green.
- Added a focused Go typechecker test to lock in alias-expanded impl label canonicalization.
- Split the Go test CLI harness into smaller files (`test_cli*.go`) to keep modules under 900 lines.
- Tests: `cd v11/interpreters/go && go test ./pkg/typechecker -run ImplementationLabelCanonicalizes && go test ./cmd/able`.
- Split TS implementation collection helpers into `implementation-collection-helpers.ts` so the main module stays under 900 lines.
- Split Go program checker helpers into `program_checker_types.go` and `program_checker_summaries.go` for smaller modules.
- Extracted Go loader helper utilities into `loader_helpers.go`.
- Moved Able CLI dependency fetchers (registry/git) into `deps_fetchers.go` to shrink `deps_resolver.go`.
- Tests: `cd v11/interpreters/go && go test ./pkg/typechecker -run ImplementationLabelCanonicalizes && go test ./cmd/able`.
- Fixed `fs.copy_dir` overwrite behavior by clearing destination contents after a removal attempt when needed.
- Tests: `./run_stdlib_tests.sh`; `./run_all_tests.sh --version=v11`.
- Dropped redundant where-clauses from `HashMap` impls so stdlib typechecking is clean in strict mode.
- Updated strict fixture manifests for expected typechecker diagnostics and refreshed the baseline for `expressions/map_literal_spread`.
- Removed the `PathType` alias from fs helpers; paths now use `Path | String` directly and fs helpers extend `Path`.
- Tests: `./run_all_tests.sh --version=v11 --typecheck-fixtures-strict --fixture`; `./run_stdlib_tests.sh`.
- Fixed impl validation to compare interface method where clauses against method-level constraints instead of impl-level where clauses (TS + Go typecheckers).
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/implementation_validation.test.ts`; `cd v11/interpreters/go && go test ./pkg/typechecker`.
- Added exec fixture `exec/10_02_impl_where_clause` to cover impl-level where clauses without method-level duplication and updated the coverage index.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_02_impl_where_clause bun run scripts/run-fixtures.ts`.
- Added exec fixture `exec/04_05_04_struct_literal_generic_inference`, updated the exec coverage index, and enforced exec-fixture typechecking when manifests specify diagnostics (TS + Go).
- Fixed struct literal generic type-argument handling in the TS and Go typecheckers (placeholder args in TS; inferred args in Go).
- Tests: `./run_all_tests.sh --version=v11 --typecheck-fixtures-strict`.
- Clarified spec call-site inference to include return-context expected types and documented return-context inference design notes; updated PLAN work queue.

### 2026-01-15
- TS typechecker now uses expected return types to drive generic call inference (explicit + implicit return paths), with new return-context unit tests.
- Go typechecker now propagates expected return types through implicit return blocks, plus a focused unit test for implicit-return inference.
- Added exec fixture `exec/07_08_return_context_generic_call_inference` and updated the exec coverage index.
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/return_context_inference.test.ts`; `cd v11/interpreters/go && go test ./pkg/typechecker -run TestGenericCallInfersFromImplicitReturnExpectedType`.
- TS typechecker now treats method-shorthand exports as taking implicit self for overload resolution, and uses receiver substitutions when enforcing method-set where clauses on exported function calls.
- TS runtime now treats unresolved generic type arguments on struct instances as wildcard matches when comparing against concrete generic types.
- Tests: `./run_all_tests.sh --version=v11 --typecheck-fixtures-strict`.
- Documented kernel Hash/Eq decisions (sink-style hashing, IEEE float equality, floats not Eq/Hash), updated spec wording, and expanded the PLAN work breakdown for interpreter alignment.
- Extended the kernel Hash/Eq plan to move the default `Hasher` implementation into Able with host bitcast helpers; updated spec TODOs and PLAN tasks accordingly.
- Added a kernel-level FNV-1a Hasher (Able code) with big-endian byte emission, introduced `__able_f32_bits`/`__able_f64_bits`/`__able_u64_mul` helpers in TS/Go, and updated stdlib hashing call sites + tests to use the new sink-style Hash API.

### 2026-01-23
- Renamed async scheduler helpers to `future_*` across v11 runtimes, docs, fixtures, and tests (`future_yield`, `future_cancelled`, `future_flush`, `future_pending_tasks`) while keeping fixture IDs intact.
- Updated typechecker diagnostics for `future_yield` to reference async tasks rather than `proc` bodies.
- Tests: `./run_all_tests.sh --version=v11 --fixture`.
