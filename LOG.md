# Able Project Log

# 2025-12-12 — Inherent methods as functions (v11)
- Exported method functions now carry their method-set obligations: Go binds implicit `self` for method shorthand, substitutes the receiver into exported signatures, and preserves method-set context so free-call constraints fail when receivers/where-clauses are violated.
- TypeScript attaches method-set obligations/substitutions to exported method infos so direct calls enforce receiver typing and block missing `where` constraints; §9 in the spec now spells out the export + sugar model for inherent methods.
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/function_calls.test.ts`; `cd v11/interpreters/go && go test ./pkg/typechecker`.

# 2025-12-11 — UFCS overload priority (v11)
- Added overload priority metadata so inherent methods outrank interface/default impls without masking UFCS ambiguities; runtime dispatch now tags impl/default entries with lower priority and tolerates impl ambiguity when other candidates are present.
- Typechecker mirrors the priority model, keeping the highest-priority candidate per signature and sorting overloads by score then priority to stay aligned with runtime selection.
- Go interpreter/runtime carry the same priority metadata for parity with the TypeScript path.
- Module loaders now auto-discover bundled kernel/stdlib roots (no `ABLE_STD_LIB` knob); the `able` namespace is treated as a normal dependency resolved through the standard search paths/lockfile.
- Tests: `bun test test/stdlib/string_stdlib_integration.test.ts`; `bun test test/stdlib/concurrency_stdlib_integration.test.ts test/stdlib/await_stdlib_integration.test.ts test/stdlib/hash_set_stdlib_integration.test.ts test/stdlib/bit_set_stdlib_integration.test.ts test/stdlib/hash_map_stdlib_integration.test.ts`; `./run_all_tests.sh --version=v11`.

# 2025-12-10 — Conditional/unwrap syntax shift (v11)
- Landed the new `if/elsif/else` syntax and `{ expr or { err => ... } }` handlers across grammar/AST/TS+Go interpreters/typecheckers; regenerated tree-sitter artifacts and aligned AST fixtures/printers with the new `elseIfClauses`/`elseBody` layout.
- Updated fixtures, examples, tutorials, and tests (TS + Go) to drop the legacy `if/or` + `| err |` handler forms; handling blocks now use `binding =>` with optional binding, and parser mappers no longer misclassify handler statements as bindings.
- Exported the fixture corpus and reran the full parity + CLI + Go/TS suites to green.
- Tests: `./run_all_tests.sh --version=v11`.

# 2025-12-09 — Pipe/placeholder alignment (v11)
- Made bare `@` placeholders behave as `@1` across parsers/interpreters/typecheckers; placeholder lambdas now reuse the first argument when unnumbered tokens repeat, with new runtime tests in TS/Go to lock the behaviour.
- Pipe expressions now typecheck as callable invocations (including low-precedence `|>>`), emitting non-callable diagnostics instead of silently returning `unknown`, and placeholder-driven pipes follow the callable-only model.
- Exported the AST fixture corpus after the placeholder changes so shared fixtures/exports include explicit placeholder indices.
- Tests: `cd v11/interpreters/ts && bun test test/runtime/pipes.test.ts && bun test test/typechecker/function_calls.test.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run Placeholder && go test ./pkg/typechecker`.

## 2025-12-08 — Parity harness green (v11)
- Normalized Go fixture results (String kind) and fixed native Error methods to use receiver-bound arity so messages propagate correctly through fixture CLI output and parity reporting.
- Cleaned up TS fixture/parity scripts (type checks and stringify usage) and aligned stringification on both runtimes: TS honours `to_string`, Go bound-method stringification now includes targets, and stdout capture no longer misreports types.
- Parity sweep now fully green with the latest fixture export; report saved to `tmp/parity-report.json`.
- Updated the v11 spec + manuals/onboarding to canonicalize the `String` type (lowercase reserved for scalar primitives) and remove the remaining lowercase `string` references in type signatures/examples.
- Tests: `cd v11/interpreters/go && go test ./...`; `cd v11/interpreters/ts && bun run scripts/run-parity.ts`.

## 2025-12-07 — Operators tutorial + xor parity (v11)
- Added tutorial `02a_operators_and_builtin_types.able` to showcase built-in scalars, arithmetic (`/ // %% /%`), comparisons, bitwise/shift ops, and boolean logic; confirmed it runs in TS+Go.
- Enabled `\xor` in the Go interpreter/typechecker and added a shared fixture (`functions/bitwise_xor_operator`) so bitwise xor stays covered across runtimes.
- Go/TS suites remain green after exporting fixtures and rerunning the operators tutorial with both CLIs.

## 2025-12-06 — Option/Result `else` handling fixed (v11)
- Implemented spec-compliant `else {}` handling for `?T`/`!T` across TS/Go interpreters and typecheckers: failures now trigger handler blocks on `nil` or `Error` values, with error bindings and optional early-return narrowing.
- Updated the Go runtime type test so interface-backed errors (e.g., `MathError`) satisfy `Error`, letting `or-else` handlers bind user-defined errors instead of treating them as successes.
- The tutorial 09 example now runs to completion in both interpreters (`Handled error: need two numbers`), clearing the option-narrowing runtime/typechecker regression from the PLAN backlog.

## 2025-12-05 — Go stdlib typecheck parity (v11)
- Go ProgramChecker sweep (driver.NewLoader RootStdlib against `v11/stdlib/src`) now reports zero diagnostics; stdlib modules typecheck cleanly with the Go checker.
- `./run_all_tests.sh --version=v11` stays green (TS + Go units, fixtures, parity, CLI); Go unit suites no longer flag await manual waker.
- Removed the Go stdlib/typechecker parity worklist from PLAN since the stdlib checker is stable again.

## 2025-12-05 — UFCS method-style fallback finalized (v11)
- Clarified UFCS spec wording (§7.4/§9.4) covering pipe equivalence, receiver-compatible overload selection, and self-parameter trimming; marked the TODO as complete and removed the PLAN item.
- Added shared UFCS fixtures: `functions/ufcs_generic_overloads` (generic + overloaded free-function binding parity across pipe/member syntax) and `errors/ufcs_overload_ambiguity` (method-style ambiguity diagnostics). Exporter now emits these fixtures.
- Hardened the TS typechecker runtime: pipelines no longer emit spurious undefined-identifier diagnostics, and duplicate typecheck diagnostics are deduped in the fixture runner; fixed a missing runtime import in pattern handling.
- Updated the typecheck baseline to reflect new fixtures and existing diagnostics, keeping UFCS coverage aligned across interpreters.

## 2025-12-04 — Struct pattern shorthand + rename operator (v11)
- Codified `::` as the pattern rename operator in §5.2 (shorthand `field` = `field::field`, rename, rename+type, nested patterns) while keeping dot for namespace traversal; added examples for assignment/match and aligned import syntax guidance.
- Documented import renaming via `::` (`import pkg::alias`, `import pkg.{item::alias}`) in the spec and queued implementation work in PLAN.
- Updated the tree-sitter grammar and TS/Go ASTs to carry rename + type metadata, track struct kinds for positional patterns, and propagate type annotations into nested struct patterns when present.
- TS/Go runtimes and typecheckers now evaluate/check shorthand/rename/type struct patterns consistently across `=`/`:=`/`match`/`rescue`; fixtures/tutorials/exporter outputs refreshed (including `patterns/struct_pattern_rename` and updated baselines).
- Tests: `bun test` (v11/interpreters/ts); `cd v11/interpreters/go && go test ./...`.

## 2025-12-03 — Impl specificity + ambiguity parity (v11)
- Implemented the full impl-specificity lattice across Go/TS typecheckers and runtimes: concrete > generic, constraint superset, union subset, and more-instantiated generic tie-breaks with ambiguity diagnostics; named impls stay opt-in.
- Go/TS runtimes now register generic-target impls, bind `Self` during matching, and surface consistent ambiguity errors; fixtures/manifests (`interfaces/impl_specificity*`) and the typecheck baseline reflect the new coverage.
- Tests: `cd v11/interpreters/go && go test ./...`; `cd v11/interpreters/ts && bun test`.

## 2025-12-02 — Pipe precedence + low-precedence pipe (v11)
- Raised pipe (`|>`) precedence above assignment and added a low-precedence `|>>` operator in the grammar (tree-sitter regenerated) with matching TS/Go parser mappings.
- TS/Go interpreters and typecheckers now treat `|>>` identically to `|>` (topic/callable fallback, placeholder guards) so pipeline semantics stay in sync across runtimes.
- Added fixture `pipes/low_precedence_pipe` to cover assignment vs pipe grouping and `||` interactions; refreshed exported fixtures, and `bun test` (v11/interpreters/ts) plus `cd v11/interpreters/go && go test ./...` are green.

## 2025-12-02 — UFCS inherent instance methods (v11)
- UFCS resolution now considers inherent instance methods (excluding static/interface/named impl methods) when the first argument can serve as the receiver; TS/Go interpreters bind the receiver accordingly and fall back from identifier calls, with TS/Go typecheckers mirroring the UFCS candidate search and call handling.
- Added fixtures for UFCS calls into inherent methods plus a static-method negative case, refreshed the typecheck baseline, and documented the UFCS expansion in the spec (§7.4/§9.4).
- `bun test` (v11/interpreters/ts) and `cd v11/interpreters/go && go test ./...` remain green after the UFCS sweep.

## 2025-12-02 — Function/method overloading implemented (v11)
- Added function/method overload sets across the Go and TS interpreters (runtime dispatch with nullable-tail omission, bound/native/UFCS/dyn refs) and TS typechecker (arity filter, specificity scoring, ambiguity diagnostics); envs now merge duplicate names into overload sets.
- Updated the spec (§7.4.1, §7.7) to codify nullable trailing parameter omission and overload eligibility/specificity, removed the overloading TODO, and exported shared fixtures for overload success/ambiguity (functions and methods).
- `bun test` (v11/interpreters/ts) and `cd v11/interpreters/go && go test ./...` are green after the overload sweep.

## 2025-12-01 — Interface dispatch alignment + Awaitable stdlib
- Codified the language-backed interfaces in the stdlib (Apply, Index/IndexMut, Iterable defaults, Awaitable/Proc/Future handles, channel/mutex awaitables) and wired both interpreters/typecheckers to route callable invocation and `[]`/`[]=` through these impls, surfacing Apply/IndexMut diagnostics when missing.
- Added shared fixtures/tests for callable values and index assignment dispatch (`interfaces/apply_index_dispatch`, missing-impl diagnostics, Go/TS unit tests) plus new awaitable fixtures covering channel/mutex/timer arms and stdlib helpers (`concurrency/await_*`, ModuleLoader integration for Channel/Mutex/Await.default).
- Refreshed the stdlib concurrency surface (awaitable wrappers, channel/mutex helpers) and kept parity harnesses green; `./run_all_tests.sh --version=v11` now passes end-to-end after the interface alignment sweep (TS unit/CLI/tests+fixtures+parity, Go unit tests).

## 2025-11-27 — Channel for-loop iteration fixed (Go runtime)
- Channel iterators now terminate correctly after closure/empty reads (stdlib `ChannelIterator.next` returns `IteratorEnd` before the typed arm), so Go for-loops over channels no longer hang.
- The Go ModuleLoader concurrency smoke test now iterates a channel with a for-loop (summing values) via the stdlib surface, and the stdlib channel/mutex smoke test mirrors the for-loop path.
- `go test ./...` (v11/interpreters/go) and the TS stdlib integration test for concurrency remain green.

## 2025-11-26 — Kernel vs stdlib layering complete (v11)
- Primitive strings now live entirely in the stdlib: helpers/iterators/wrap-unwrap operate on built-in strings, with only the three kernel bridges left native. Go/TS runtimes and typecheckers resolve string members via the stdlib and surface import hints when missing.
- Arrays now prefer stdlib methods end-to-end (size inline, minimal native shims retained for compatibility), with runtimes/typecheckers pointing diagnostics at `able.collections.array`.
- Go runtime member access now consults stdlib method sets before native fallbacks, and primitive `string` is treated as iterable (`u8`) in the Go typechecker, matching the TS path. ModuleLoader + Go test suites stay green.

## 2025-11-20 — Typed-pattern match reachability fix
- Added a reachability guard for typed `match` patterns so clauses whose annotations cannot match the subject still typecheck but are treated as unreachable, preventing spurious return/branch diagnostics in the Go checker (mirrors TS intent).
- Full v11 suites run green after the fix (`go test ./pkg/typechecker`, `go test ./...`, `./run_all_tests.sh --version=v11`).
- Extended the v11 stdlib with a `collections.set` interface and `HashSet` implementation (plus smoke + TS ModuleLoader integration tests) and refreshed stdlib docs/PLAN so the restored surface tracks the new module.
- Go typechecker now recognises stdlib collections (`List`, `Vector`, `HashSet`) as valid for-loop iterables; iterable helpers moved out of `type_utils.go` to stay under the 1k-line guardrail, and a new regression test covers generic element inference for these types.

## 2025-11-19 — File modularization cleanup
- Split Go typechecker declaration/type utility stacks into dedicated files (`decls_*`, `type_substitution.go`) and shrank `type_utils.go` beneath the 1k-line guardrail; Go AST definitions now live across `ast.go`, `type_expressions.go`, and `patterns.go` so each file stays lean.
- Broke the long TS fixture exporter (`proc_scheduling.ts`) into `proc_scheduling_part{1,2}.ts` with a tiny aggregate shim to keep per-file size under 1000 lines.
- `go test ./...` (v11/interpreters/go) remains green after the split.

## 2025-11-18 — String helper surface complete
- Added the full string helper set to both interpreters: `len_chars`/`len_graphemes`, `substring` (code-point offsets with `RangeError` on invalid bounds), `split` (empty delimiter splits graphemes), `replace`, `starts_with`/`ends_with`, and `chars`/`graphemes` iterators (Segmenter-aware in TS, rune fallback in Go). Shared helpers keep `len_*` in sync with iterator `size()`.
- Typecheckers now understand the string surface (Go member access signatures, TS call resolution for optional substring length and string primitives), and Go array literal inference merges element types instead of rejecting unions.
- Added the `strings/string_methods` AST fixture covering length/slicing/split/replace/prefix/suffix cases; `./run_all_tests.sh --version=v11` runs green after the fixture/typechecker/runtime updates.

## 2025-11-17 — Loop/range constructs & continue semantics complete
- Locked down the full §8.2–§8.3 feature set across both runtimes: `loop {}` expressions, `while`, and `for` now return the last `break` payload (or `nil` on exhaustion) while rejecting labeled `break`/`continue` targets. The Go and TypeScript interpreters share identical behavior, including iterator-driven loops and generator continuations.
- Added the Range interface runtime registries (`src/interpreter/range.ts`, `pkg/interpreter/range_runtime.go`) so range literals first delegate to stdlib implementations before falling back to synthesized arrays, matching the spec requirement that ranges materialize via the Range interface.
- Tree-sitter mappers + AST builders already exposed `LoopExpression`, `ContinueStatement`, and `RangeExpression`; we confirmed parser precedence and metadata are round-tripping, so the AST contract is now exercised end-to-end by fixtures such as `control/loop_expression` and `control/for_range_break`.
- Typechecker metadata is in place: both checkers push loop contexts, record break payload types, and emit diagnostics for labeled `continue`, so assignments using loop expressions inherit the correct type.
- Fresh runtime + parity coverage landed in TS (`test/control_flow/{while,for}.test.ts`, `test/runtime/iterators.test.ts`) and Go (`pkg/interpreter/interpreter_control_flow_test.go`). `./run_all_tests.sh --version=v11` runs green (TS unit + CLI suites, fixtures/parity harness, Go parser/interpreter/typechecker tests), so PLAN item 6 is officially complete and moved here from the TODO list.

# 2025-11-17 — Typed declaration + literal adoption work finalized
- Locked down the binding semantics for `:=`/`=` by keeping typed patterns intact across AST parsing/mapping (declaration + fallback assignment), enforcing “`:=` introduces at least one new binding” in both interpreters/typecheckers, and ensuring runtime evaluation order stays deterministic (RHS once, receivers/indexers evaluated exactly once) even for compound assignments and safe-navigation forms.
- Verified the runtime/typechecker literal-adoption flow now covers every context listed in the v11 spec (arrays, maps, ranges, iterator yields, async bodies, function bodies/returns/arguments, struct literals, typed patterns), and refreshed the AST fixtures (`patterns/typed_destructuring`, `patterns/typed_equals_assignment`) so typed destructuring + typed `=` assignments keep parity between TypeScript and Go.
- Fixed a regression uncovered by the fixture sweep where the TS AST builder always serialized `isSafe: false` on member-access nodes; it now omits the flag unless `?.` is present, matching the Go AST (`json:"safe,omitempty"`) and letting fixtures/tests pass again. Full `./run_all_tests.sh --version=v11` run is green.

## 2025-11-16 — Safe navigation operator implemented
- Tree-sitter grammar/parser now treat `?.` as part of member-access, the TypeScript/Go AST mappers expose a `safe` flag on `MemberAccessExpression` nodes, and the generated parser artifacts (`grammar.json`, `parser.c`, node types, WASM) have been regenerated so fixtures and tooling pick up the new operator.
- TypeScript + Go interpreters short-circuit safe member access/calls (returning `nil` when receivers are `nil`, skipping argument evaluation, and mirroring dot semantics otherwise) while rejecting assignments that attempt to use `?.`.
- The Go typechecker wraps safe-navigation results in `NullableType` only when the receiver may be `nil`, so redundant usage on non-optional receivers still typechecks as plain dot access. New unit tests in both runtimes (`test/runtime/safe_navigation.test.ts`, `interpreter_safe_navigation_test.go`) cover the runtime semantics, and `bun test …`, `go test ./pkg/typechecker ./pkg/interpreter`, plus `./v11/ablego ./v11/examples/rosettacode/factorial.able` all remain green.

## 2025-11-15 — Map literal support landed
- Extended the shared tree-sitter grammar plus both AST layers (TS + Go) with `MapLiteral`/entry/spread nodes, regenerated parser artifacts, and wired the fixtures/exporter so `#{ ... }` forms round-trip through `module.json` + source generation.
- Implemented map literal evaluation in the TypeScript interpreter (new hash_map value kind, literal/spread semantics, fixture assertions) and added runtime+typechecker unit tests covering spreads, duplicates, and diagnostics.
- Mirrored the same behavior in the Go interpreter/typechecker (hash map insertion helper, literal evaluation, new `MapType`, diagnostics) and added Go unit tests—`./v11/run_all_tests.sh` now exercises the new fixtures end-to-end.

## 2025-11-14 — Type alias declarations wired end-to-end
- The shared tree-sitter grammar gained a `type_alias_definition` rule (space-delimited generics + optional `where` clauses) plus corpus coverage, and both the TypeScript + Go AST mappers now surface `TypeAliasDefinition` statements.
- TypeScript + Go parsers/interpreters ignore alias declarations at runtime while the TypeScript typechecker/summary plumbing keeps tracking them; the TypeScript fixture exporter/pretty-printer now emits `type Foo T where … = Expr` syntax.
- Added the `types/type_alias_definition` AST fixture so both runtimes/typecheckers exercise alias declarations, updated the fixture baseline (`bun run scripts/export-fixtures.ts && ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts -- --write-typecheck-baseline`), and kept Go green with `go test ./pkg/parser ./pkg/interpreter`.

## 2025-11-13 — v11 Spec Expansion Complete
- Closed the v11 spec TODO slate: every deferred item now lives in `spec/full_spec_v11.md`, covering mutable `=` semantics (§5.3.1), map literals (§6.1.9), struct functional updates (§4.5.2), type aliases (§4.7), safe navigation (§6.3.4), typed `=` declarations (§5.1–§5.1.1), contextual literal typing (§6.1.1/§6.3.2), optional generic parameters (§7.1.5/§4.5/§10.1), await/async coordination plus channel error surfaces (§12.6–§12.7), the `loop` expression (§8.2.3), Array/String APIs (§6.8/§6.1.5/§6.12.1), stdlib packaging + module search paths (§13.6–§13.7), and the regex/text modules (§14.2). `spec/TODO_v11.md` now reflects that completion and will track only newly scheduled language work.
- The root PLAN no longer lists “Expand the v11 specification” as an open milestone; remaining TODOs focus on implementing the documented features across the interpreters, parser, and stdlib.

## 2025-11-12 — Versioned Workspace Split
- Introduced a dedicated `v10/` workspace that now owns the frozen Able v10 assets (`design/`, `docs/`, `examples/`, `fixtures/`, `parser/`, `stdlib/`, `interpreter10` → `v10/interpreters/ts`, `interpreter10-go` → `v10/interpreters/go`, plus helper scripts). This removes ambiguity about where new work should land and keeps the archived toolchain intact for maintenance.
- Added a version-dispatching `run_all_tests.sh` at the repo root (`./run_all_tests.sh --version=v10|v11 --typecheck-fixtures=...`) so CI and contributors can target either workspace without remembering individual paths.
- Updated repo-wide onboarding docs (`README.md`, `AGENTS.md`, `PLAN.md`) to describe the multi-version layout, explain how to run tests per version, and drop completed roadmap items covering the initial workspace bootstrap/freeze.
- Copied the legacy v10 docs (`v10/README.md`, `v10/AGENTS.md`, `v10/PLAN.md`, `v10/LOG.md`) so historical context remains close to the frozen code while the root docs focus on cross-version coordination.

## 2025-11-11 — Stdlib Module Search Paths
- **Pipe semantics parity**: Added the `pipes/multi_stage_chain` AST fixture so multi-stage pipelines that mix `%` topic steps, placeholder-built callables, and bound methods stay covered; `bun run scripts/run-fixtures.ts` (TypeScript) and `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter` (Go) stay green with no parity divergences observed.
- **Typechecker strict fixtures**: TypeScript’s checker now hoists struct identifiers for static calls, predeclares `:=` bindings (so proc handles can reference themselves), binds iterator driver aliases plus struct/array pattern destructures, and hides private package members behind the standard “has no symbol” diagnostic. The full fixture suite now passes under `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts`, and manifests/baselines were updated where diagnostics are expected.
- **Dynamic interface collections & iterables**: added the shared fixture `interfaces/dynamic_interface_collections` (plus exporter wiring) so both interpreters prove that range-driven loops and map-like containers storing interface values still choose the most specific impl even when union targets overlap. `bun run scripts/run-fixtures.ts` is green; Go parity remains blocked by existing fixture failures (`functions/hkt_interface_impl_ok`, `imports/static_alias_private_error`).
- **Privacy & import spec gaps**: enforced package privacy across both interpreters/typecheckers (selectors, aliases, wildcard bindings), tightened dynimport placeholder handling (TypeScript + Go parity plus fixture coverage), and updated `spec/full_spec_v10.md` with the canonical `Proc`/`Future` handle definitions so runtimes and tooling share the same ABI (`ProcStatus`, `ProcError`, `status/value/cancel` contracts).
- **Interface & impl completeness**: higher-kinded/visibility edge cases now have end-to-end coverage. Overlapping impl ambiguity diagnostics land in both interpreters/typecheckers, TypeScript enforces method-set backed constraints with parity coverage, both interpreters honour interface self-type patterns (including higher-kinded targets) and reject bare constructors unless `for …` is declared, and parser/exporter propagate the `for …` patterns so fixtures surface the diagnostics automatically. Shared AST fixtures (`errors/interface_self_pattern_mismatch`, `errors/interface_hkt_constructor_mismatch`, plus the positive `functions/hkt_interface_impl_ok`) keep both runtimes + typecheckers aligned.

- Proc/future cancellation coverage is complete: both interpreters expose `Future.cancel()`, cancellation transitions produce the `Cancelled` status, and the new `concurrency/future_cancel_nested` fixture exercises a proc awaiting a spawned future that is cancelled mid-flight. The exporter and strict harness generate the fixture automatically, keeping the TypeScript + Go runtimes/typecheckers in lockstep for nested cancellation chains across both executors.
- Channel error helpers now emit the stdlib error structs in both runtimes. TypeScript + Go runtimes route every channel error through `ChannelClosed`, `ChannelNil`, or `ChannelSendOnClosed`, and new Bun/Go tests cover the behaviour (`interpreter10/test/concurrency/channel_mutex.test.ts`, `interpreter10-go/pkg/interpreter/interpreter_channels_mutex_test.go`).
- Lightweight executor diagnostics are now wired into both runtimes: the new `proc_pending_tasks()` helper surfaces cooperative queue length via `CooperativeExecutor.pendingTasks`, Go’s serial and goroutine executors expose the same data, and coverage comes from a dedicated AST fixture (`concurrency/proc_executor_diagnostics`) plus Bun/Go unit tests. The spec, manuals, and `design/concurrency-executor-contract.md` now describe the helper so fixtures can assert drain behaviour deterministically.
- String host bridge externs (`__able_string_from_builtin`, `__able_string_to_builtin`, `__able_char_from_codepoint`) are wired into both interpreters and typecheckers with dedicated tests (`interpreter10/test/string_host.test.ts`, `interpreter10-go/pkg/interpreter/interpreter_string_host_test.go`).
- Hasher externs (`__able_hasher_create/write/finish`) now back the stdlib hash maps across TypeScript and Go, complete with parity tests and stub support for tooling.
- Added the `concurrency/channel_error_rescue` AST fixture and exposed `Error` member access (message/value) in both interpreters so Able code can assert the struct payloads produced by the channel helpers.
- Go parity now runs the new `concurrency/channel_error_rescue` fixture under the goroutine executor, and a dedicated Go test (`TestChannelErrorRescueExposesStructValue`) verifies that rescuing channel errors exposes the struct payload via `err.value`.
- Added the `errors/result_error_accessors` AST fixture so both interpreters exercise `err.message()/cause()/value` inside `!T else { |err| ... }` flows; fixture exporter + TS harness updated accordingly.
- Go typechecker now recognises `Error.message()`, `.cause()`, and `.value`, and the spec documents the runtime-provided `Error.value` payload hook; the typechecker baseline entry for `channel_error_rescue` was removed once diagnostics cleared.
- Proc/future runtime errors now record their cause payloads in both interpreters, the new `concurrency/proc_error_cause` fixture exercises `err.cause()` end-to-end, and matching Bun/Go tests keep the regression harness green.
- Generator laziness parity closed: iterator continuations now cover if/while/for/match across both runtimes, stdlib helpers (`stdlib/src/concurrency/channel.able`, `stdlib/src/collections/range.able`) use generator literals, and new fixtures (`fixtures/ast/control/iterator_*`, `fixtures/ast/stdlib/channel_iterator`, `fixtures/ast/stdlib/range_iterator`) keep the shared harness authoritative.
- Automatic time slicing verified for long-running procs: the new `concurrency/proc_time_slicing` fixture proves that handles without explicit `proc_yield()` still progress under repeated `proc_flush()` calls, capturing both the intermediate `Pending` status and the eventual resolved value across runtimes.
### AST → Parser → Typechecker Completion Plan _(reopen when new AST work appears)_
- Full sweep completed 2025-11-06 (strict fixture run, Go interpreter suite, Go parser harness, and `bun test` all green). Archive details in `LOG.md`; bring this plan back only if new AST/syntax changes introduce regressions.

## 2025-11-09 — Executor Diagnostics Helper
- Added the `proc_pending_tasks()` runtime helper so Able programs/tests can observe cooperative executor queue depth. TypeScript wires it through `CooperativeExecutor.pendingTasks()` while the Go runtime surfaces counts from both the serial and goroutine executors (best-effort on the latter via atomic counters). The helper is registered with both typecheckers so Warn/Strict fixture runs understand the signature.
- New coverage keeps the helper honest: Bun unit tests exercise the helper directly (`interpreter10/test/concurrency/proc_spawn_scheduling.test.ts`), Go gains matching tests (`TestProcPendingTasksSerialExecutor`, `TestProcPendingTasksGoroutineExecutor`), and a serial-only AST fixture (`fixtures/ast/concurrency/proc_executor_diagnostics/`) ensures both interpreters prove that `proc_flush` drains the cooperative queue.
- Spec + docs now describe the helper alongside `proc_yield`/`proc_flush` (see `spec/full_spec_v10.md`, `docs/manual/*.md/html`, `design/concurrency-executor-contract.md`, and `AGENTS.md`), and the PLAN TODO for “Concurrency ergonomics” is officially closed out.

## 2025-11-08 — Fixture Diagnostics Parity Enforcement
- `interpreter10/scripts/run-parity.ts` now diffs typechecker diagnostics for every AST fixture so warn/strict parity runs catch unexpected checker output even when manifests do not declare expectations. The parity JSON report captures the mismatched diagnostics to speed up triage.
- Added `test/scripts/parity/fixtures_compare.test.ts` to cover the helper logic that determines when diagnostics mismatches should fail parity.
- Go’s typechecker now treats unannotated `self` parameters inside `methods {}` / `impl` blocks as `Self`/concrete receiver types and seeds iterator literals with the implicit `gen` binding, eliminating the extra diagnostics that used to appear only in Go’s warn/strict runs.
- Added Go regression tests (`pkg/typechecker/checker_impls_test.go`, `pkg/typechecker/checker_iterators_test.go`) so implicit `self` bindings and iterator generator helpers stay covered.
- Bun’s typechecker mirrors the same behaviour: iterator literals now predefine the implicit `gen` binding, implicit `self` parameters default to `Self`, and new tests (`test/typechecker/method_sets.test.ts`, `test/typechecker/iterators.test.ts`) keep the coverage locked in.
- The Go fixture runner and CLI now continue evaluating programs in warn/strict modes even when diagnostics are reported (matching the Bun harness), so parity checks capture both the expected runtime results and the shared diagnostics payloads.

## 2025-11-07 — AST → Parser → Typechecker Cycle Revalidated
- Added proc handle memoization fixtures (success + cancellation) and ensured both interpreters plus the Go parser harness run them under strict typechecking (`bun run scripts/run-fixtures.ts`, `cd interpreter10-go && GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser`).
- Verified the full suite remains green (`./run_all_tests.sh --typecheck-fixtures=strict`, `bun test`, `cd interpreter10-go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`), keeping the Priority 0 gate satisfied.
- Updated PLAN.md to mark the current AST → Parser → Typechecker cycle complete and advance the focus to Phase α (Channel & Mutex runtime bring-up).
- Added stdlib specs for channel/mutex behaviour (`stdlib/tests/concurrency/channel_mutex.test.able`) so the Phase α bullet “add unit tests covering core operations” is now satisfied.

## 2025-11-07 — Serial Executor Future Reentrancy
- Go’s SerialExecutor now exposes a `Drive` helper that runs pending proc/future tasks inline, so nested `future.value()` / `proc_handle.value()` calls no longer deadlock and match the TypeScript scheduler semantics. The helper steals the targeted handle from the deterministic queue, executes it re-entrantly (including repeated slices when `proc_yield` fires), and restores the outer task context once the awaited handle resolves.
- Fixtures `concurrency/future_value_reentrancy` and `concurrency/proc_flush_fairness` now pass under the Go interpreter’s serial executor, keeping the newly added fairness/re-entrancy corpus green for both runtimes (`ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts`, `cd interpreter10-go && GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`).
- PLAN immediate actions trimmed: the “Concurrency stress coverage & docs” placeholder has been cleared now that the blocking fixtures run cleanly; follow-up concurrency work can graduate to design docs instead of the top-level plan.
- Added a Go-side regression test (`TestSerialExecutorFutureValueReentrancy`) that mirrors the new fixture to ensure nested `future.value()` calls stay green even if future contributors touch the executor; the design note `design/go-concurrency-scheduler.md` now documents the inline-driving behaviour and follow-up doc work.
- Added a companion fixture (`concurrency/proc_value_reentrancy`) plus TypeScript exporter wiring + Go regression test (`TestSerialExecutorProcValueReentrancy`) so both interpreters exercise nested `proc.value()` waits under the serial executor; parser coverage tables were updated accordingly.
- Documented the goroutine-executor fairness contract (what `proc_yield`/`proc_flush` mean under `GoroutineExecutor`, how we rely on Go’s scheduler, and when tests must fall back to the serial executor) to close out the remaining PLAN follow-up (`design/go-concurrency-scheduler.md`).
- Updated the v10 spec to codify re-entrant `proc.value()` / `future.value()` semantics so both interpreters (and future targets) must guarantee deadlock-free nested waits (§12.2.5 “Re-entrant waits”).
- Added direct Go unit tests (`TestProcHandleValueMemoizesResult`, `TestProcHandleValueCancellationMemoized`) to ensure repeated `value()` calls return memoized results/errors even after cancellation, satisfying the remaining “exercise repeated value() paths” item from the concurrency plan.
- Introduced the `concurrency/proc_value_memoization` fixture (plus exporter wiring) so both interpreters prove proc handles memoize values, and updated the Go parity harness to run it with the goroutine executor.
- Added `concurrency/proc_value_cancel_memoization` to assert that cancelled proc handles return the same error for repeated `value()` calls without re-running their bodies; the fixture exporter, AST corpus, and fixture run all cover this scenario now.

## 2025-11-06 — Tree-Sitter Mapper Modularization Complete
- Go declarations/patterns/imports all run through the shared `parseContext`, removing the last raw-source helpers so both runtimes share one parser contract (`interpreter10-go/pkg/parser/{declarations,patterns,statements,expressions}_parser*.go`).
- TypeScript parser README now calls out the shared context, and `PLAN.md` logged the Step 6 regression sweep so future contributors know the refactor is locked in (`interpreter10/src/parser/README.md`, `PLAN.md`).
- Wrapper exports like `parseExpression(node, source)` / `parseBlock` / `parsePattern` were removed; all Go parser consumers now flow through the context pipeline.
- Tests: `./run_all_tests.sh --typecheck-fixtures=warn` and `cd interpreter10-go && GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser`.
- Follow-up: confirm any remaining helpers that still accept raw source (e.g., host-target parsing) genuinely require it before migrating them to `parseContext`.

## 2025-11-03 — AST → Parser → Typechecker Completion Sweep
Status: ✅ Completed. Parser coverage table reads fully `Done`, TypeScript + Go checkers share the bool/comparison semantics, and `ABLE_TYPECHECK_FIXTURES=strict bun run scripts/run-fixtures.ts` executes cleanly after wiring the remaining builtin signatures, iterator annotations, and error fixture diagnostics into the manifests.

Open items (2025-11-02 audit):
- [x] Iterator literals dropped both the optional binding identifier and optional element type annotation. Update both ASTs + parsers/interpreters so the metadata survives round-trips and execution.
- [x] Teach both typecheckers to honor the iterator element annotation so every `yield` matches the declared element type (TS `src/typechecker/checker.ts`, Go `pkg/typechecker/literals.go` + iterator tests).
- [x] Carry iterator element metadata through `for` loops so typed loop patterns validate across array/range/iterator inputs (TS `src/typechecker/checker.ts`, Go `pkg/typechecker/{control_flow,literals,type_utils}.go` + new cross-loop tests).
- [x] Give the TS checker parity with Go for block/proc/spawn typing and array/range literal inference so all three stages (AST → parser → typechecker) agree on the element/result metadata (§6.8, §6.10, §12.2).
- [x] Enforce async-only builtins (`proc_yield`) and add concurrency smoke tests so TS emits the same diagnostics as Go when authors call scheduler helpers outside `proc`/`spawn`.
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
- Cleared local Go build caches (`.gocache`, `interpreter10-go/.gocache`) and re-ran `GOCACHE=$(pwd)/.gocache GO_PARSER_FIXTURES=1 go test ./pkg/parser` to mimic CI picking up the refreshed grammar without stale entries.
- ACTION: propagate the cache-trim guidance to CI docs if flakes recur; otherwise move on to the remaining parser fixture gaps (`design/parser-ast-coverage.md`).
- Mirrored the TypeScript placeholder auto-lift guardrails inside the Go interpreter so pipe placeholders evaluate eagerly, keeping the shared `pipes/topic_placeholder` fixture green.
- Parser sweep: both the TypeScript mapper and Go parser now skip inline comments when traversing struct literals, call/type argument lists, and struct definitions, with fresh TS/Go tests guarding the behaviour.
- TypeScript checker scaffold landed: basic environment/type utilities exist under `interpreter10/src/typechecker`, exported via the public index, and fixtures respect `ABLE_TYPECHECK_FIXTURES` ahead of the full checker port.
- TypeScript checker now emits initial diagnostics for logical operands, range bounds, struct pattern validation, and generic interface constraints so the existing error fixtures pass under `ABLE_TYPECHECK_FIXTURES=warn`.

### 2025-11-01
- TypeScript checker grew a declaration-collection sweep that registers interfaces, structs, methods, and impl blocks before expression analysis, mirroring the Go checker’s phase ordering.
- Implementation validation now checks that every `impl` supplies the interface’s required methods, that method generics/parameters/returns mirror the interface signature, and flags stray definitions; only successful impls count toward constraint satisfaction.
- Canonicalised type formatting to the v10 spec (`Array i32`, `Result string`) and keyed implementation lookups by the fully-instantiated type so generic targets participate in constraint checks.
- Extended the TypeScript checker’s type info to capture `nullable`, `result`, `union`, and function signatures, and taught constraint resolution to recognise specialised impls like `Show` for `Result string`, with fresh tests covering the new cases.
- Added focused Bun tests under `test/typechecker/` plus the `ABLE_TYPECHECK_FIXTURES` harness run to lock in the new behaviour and guard future regressions.
- Mirrored the Go checker and test suite to use the same spec-facing type formatter and wrapper-aware constraint logic so diagnostics now reference `Array i32`, `string?`, etc., and added parity tests (`interpreter10-go/pkg/typechecker/*`).

### 2025-11-02
- The Bun CLI suite (`interpreter10/test/cli/run_module_cli.test.ts`) now covers multi-file packages, custom loader search paths via the new `ABLE_MODULE_PATHS` env, and strict vs warn typecheck enforcement so the ModuleLoader refactor stays covered without pulling `stdlib/` into every run.
- Introduced the `able test` skeleton inside `scripts/run-module.ts`: it parses the planned flags/filters, materialises run options + reporter selection, and prints a deterministic plan summary before exiting with code `2` while the stdlib testing packages remain unparsable. (See `design/testing-cli-design.md` / `design/testing-cli-protocol.md`.)
- Extracted the shared package-scanning helpers (`discoverRoot`, `indexSourceFiles`, etc.) into `scripts/module-utils.ts` so other tooling (fixtures runner, future harnesses) can reuse the multi-module discovery logic without duplicating it.
- **Deferral noted:** full stdlib/testing integration is still on pause until the parser accepts `stdlib/src/testing/*`; once that unblocks, wire the CLI skeleton into the able.testing harness per the design notes.

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
- Added Bun smoke tests for nil-channel cancellation and mutex re-entry errors (`interpreter10/test/concurrency/channel_mutex.test.ts`) to mirror the Go parity suite.
- Documented the audit and captured the remaining TODO (map native errors to `ChannelClosed`/`ChannelNil`/`ChannelSendOnClosed`) in `design/channels-mutexes.md` and `spec/TODO.md`.
- Cleared `Phase α` from the active roadmap; next milestone is Phase 4 (cross-interpreter parity/tooling).

### 2025-11-07 — Fixture Parity Harness (Phase 4 Kick-off)
- Added a Go CLI entry point (`cmd/fixture`) that evaluates a single AST fixture and emits normalized JSON (result kind/value, stdout, diagnostics) with respect to `ABLE_TYPECHECK_FIXTURES`. The helper reuses the interpreter infrastructure and supports serial/goroutine executors while sandboxing the Go build cache.
- Refactored the TypeScript fixture loader into `scripts/fixture-utils.ts` so both the CLI harness and Bun tests can hydrate modules, install runtime stubs, and intercept `print` output consistently.
- Rebuilt `run-fixtures.ts` on top of the shared utilities (no behavior change) to keep fixture execution logic single-sourced.
- Introduced a Bun parity test (`test/parity/fixtures_parity.test.ts`) that exercises a representative slice of the shared fixture corpus (currently 20 fixtures across basics + concurrency) against both interpreters and asserts matching results/stdout via the new Go CLI.

### 2025-11-08 — Parity CLI Reporting
- Added `interpreter10/scripts/run-parity.ts`, reusable parity helpers, and Bun parity suites that now share the same execution/diffing logic across AST fixtures and curated examples.
- `run_all_tests.sh` now invokes the parity CLI so local + CI runs execute the same cross-interpreter verification and drop a JSON report at `tmp/parity-report.json` for machine-readable diff tracking; `tmp/` landed in `.gitignore` to keep artifacts out of commits.
- The helper script also honors `ABLE_PARITY_REPORT_DEST` or `CI_ARTIFACTS_DIR` so pipelines can copy the parity JSON into their artifact buckets without bespoke wrapper scripts.
- Updated `interpreter10/README.md` and `interpreter10-go/README.md` with parity CLI instructions, env knobs (`ABLE_PARITY_MAX_FIXTURES`, `ABLE_PARITY_REPORT_DEST`), and guidance on keeping the cross-interpreter harness green.
- Added Go package docs (`pkg/interpreter/doc.go`, `pkg/typechecker/doc.go`) plus README guidance on regenerating `go doc`/pkg.go.dev pages so the documentation workstream is unblocked.
- Landed `dynimport_parity` and `dynimport_multiroot` in `interpreter10/testdata/examples/` to cover dynamic package aliasing, selector imports, and multi-root dynimport scenarios end-to-end; the parity README + plan now list them alongside the other curated programs, and the Go CLI + Bun harness honor `ABLE_MODULE_PATHS` when resolving shared deps.
- Authored `docs/parity-reporting.md` and linked it from the workspace README so CI pipelines know how to persist `tmp/parity-report.json` via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR`.
- The Go CLI (`cmd/able`) now honors `ABLE_MODULE_PATHS` in addition to `ABLE_PATH`, with new tests ensuring the search-path env works; stdlib docs reference the alias so multi-root dynimport scenarios can rely on a single env knob across interpreters.
- Fixed the `..`/`...` range mapping bug in both parsers so inclusive ranges now follow the spec (TS + Go parser updates, new parser unit tests, interpreter for-loop regression tests, and fizzbuzz-style parity coverage).

### Phase 5 Foundations — Parser Alignment
- Canonical AST mapping now mirrors the fixture corpus across both runtimes. The TypeScript mapper’s fixture parity suite (`bun test test/parser/fixtures_mapper.test.ts`) and the Go parser harness (`go test ./pkg/parser`) stay green, so every tree-sitter node shape maps to the shared AST contract with span/origin metadata.
- `tree-sitter-able` grammar coverage is complete for the v10 surface (see `design/parser-ast-coverage.md`); new syntax is added directly with fixture+cXPath tests so the grammar remains authoritative.
- Translators and loaders are live in both interpreters: TypeScript’s `ModuleLoader` and Go’s `driver.Loader` now ingest `.able` source via tree-sitter, hydrate canonical AST modules, and feed them to their respective typechecker/interpreter pipelines.
- End-to-end parse → typecheck → interpret tests exercise both runtimes: `ModuleLoader pipeline with typechecker` (Bun) covers the TS path, and `pkg/interpreter/program_pipeline_test.go` drives the Go loader/interpreter via `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`.
- Diagnostic coverage now rides on the same pipelines: the new Bun test asserts missing import selectors surface typechecker errors before evaluation, and the Go suite verifies that `EvaluateProgram` halts (or proceeds when `AllowDiagnostics` is set) when return-type violations are reported.

### 2025-11-22
- TS typechecker imports now seed struct/interface bindings from package summaries (with generic placeholders) so stdlib types preserve their shape when referenced from dependent modules.
- Added builtin HashSet method stubs (new/with_capacity/add/remove/contains/size/clear/is_empty) to the TS typechecker, letting the hash_set ModuleLoader integration typecheck without ignoring bool-condition diagnostics.
- HashSet stdlib integration test now expects zero diagnostics; the stdlib Bun suite remains green.

### 2025-11-23
- Quarantine stdlib iterators now return the explicit `IteratorEnd {}` sentinel and match on the sentinel type to avoid pattern collisions; iterator imports across array/list/linked_list/vector/lazy_seq/string/automata DSL modules now pull from `core.iteration`.
- TS stdlib integration suite rerun to confirm no regressions in the active modules; Go unit tests remain green.
- TS typechecker now treats `Iterator` interface annotations as structural for literal overflow checks, so iterator literals annotated as `Iterator u8` surface integer-bound diagnostics on yielded values; `run_all_tests.sh --version=v11` is green.

### 2025-11-24
- Primitive `string` is now treated as an iterable of `u8` across both runtimes: TS/Go typecheckers recognise string for-loops, diagnostics reference `array, range, string, or iterator`, and new tests cover typed-pattern matches plus runtime iteration backed by the stdlib string module.
- Added ModuleLoader + Go runtime tests to ensure string iteration requires importing `able.text.string` and yields byte values; stdlib README documents the import requirement.

### 2025-11-25
- Added a guardrail fixture for missing Apply/IndexMut implementations (`interfaces/apply_index_missing_impls`) and regenerated exports/baseline so warn/strict runs expect the shared diagnostics.
- Aligned Go typechecker diagnostics with TS for non-callable Apply targets and Index-only [] assignments (now report “non-callable … missing Apply implementation” and “cannot assign via [] without IndexMut …”), keeping the parity suite green.
- Full v11 sweep rerun after the additions (`./run_all_tests.sh --version=v11`) is green.

### 2025-12-03
- Swapped import alias syntax to the `::` rename operator across the tree-sitter grammar, TS/Go parsers, and module loaders, rejecting legacy `as` aliases while keeping dot traversal unchanged.
- Updated fixtures/docs/tests to the new syntax (package aliases, selective aliases for static/dynimport, struct-pattern rename coverage) and re-exported the shared fixture corpus.
- Ran `./run_all_tests.sh --version=v11` (TS+Go units, fixtures, parity); all suites passed and parity report saved to `v11/tmp/parity-report.json`.

### 2025-12-09
- Finalized the v11 operator surface: `%`/`//`/`/%` follow Euclidean semantics, `%=` compounds and the dot-prefixed bitwise set (`.& .| .^ .<< .>>`) are supported, `^` acts as exponent (bitwise xor remains dotted), and the operator interfaces in `core.interfaces` mirror the runtime behavior. Parser/AST/typechecker/stdlib fixtures all align and legacy `%%` syntax is gone.
- Verified the full sweep (`./run_all_tests.sh --version=v11`) stays green after the operator updates (TS + Go units, fixtures, parity harness).

### 2025-12-10
- Kernel/stdlib discovery now covers the v11 layout: search-path collectors scan `v11/kernel/src` and `v11/stdlib/src`, TS ModuleLoader/Go CLI auto-load kernel packages when bundled, and new tests in both runtimes assert the v11 scan paths.
- TS/Go CLI/module loader tests updated to exercise the bundled scan; stdlib README notes the expanded auto-detection coverage.
- Manifested runs now pin the bundled boot packages: Go `able deps install` injects both stdlib and kernel into `package.lock`, kernel search-path discovery honors module-path env entries, and the TS CLI reads `package.yml`/`package.lock` to add dependency roots before falling back to bundled detection. New CLI tests cover lock-required runs plus pinned stdlib/kernel boot without env overrides.

### 2025-12-10 — Ratio & Numeric Conversions
- Implemented exact `Ratio` struct and normalization helpers in the stdlib kernel, added `to_r` conversions for integers/floats, and expanded numeric smoke tests (`v11/stdlib/src/core/numeric.able`, `v11/stdlib/tests/core/numeric_smoke.test.able`).
- TypeScript runtime/typechecker now treat `Ratio` as numeric: builtin `__able_ratio_from_float`, exact ratio arithmetic/comparisons, NumericConversions support, and new ratio tests (`v11/interpreters/ts/src/interpreter/{numeric.ts,operations.ts,numeric_host.ts}`, `v11/interpreters/ts/src/typechecker/**`, `v11/interpreters/ts/test/{typechecker/numeric.test.ts,basics/ratio.test.ts}`).
- Go runtime/typechecker mirror the Ratio struct/builtin and conversion helpers with adjusted constraint diagnostics plus coverage in interpreter/typechecker suites (`v11/interpreters/go/pkg/interpreter/*`, `v11/interpreters/go/pkg/typechecker/*`).
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/numeric.test.ts`; `cd v11/interpreters/ts && bun test test/basics/ratio.test.ts`; `cd v11/interpreters/go && go test ./pkg/typechecker`; `cd v11/interpreters/go && go test ./pkg/interpreter`.
