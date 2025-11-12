# Able Project Roadmap (v10 focus)

## Scope
- Maintain a canonical Able v10 language definition across interpreters and tooling, with the written specification remaining the source of truth for behaviour.
- Prioritise the Go interpreter until it matches the TypeScript implementation feature-for-feature (the only intentional divergence is that Go uses goroutines/channels while TypeScript simulates concurrency with a cooperative scheduler).
- Keep the TypeScript and Go AST representations structurally identical so tree-sitter output can feed either runtime (and future targets like Crystal); codify that AST contract inside the v10 specification once validated.
- Document process and responsibilities so contributors can iterate confidently.
- Modularize larger features into smaller, self-contained modules. Keep each file under one thousdand (i.e. 1000) lines of code.

## Existing Assets
- `spec/full_spec_v10.md`: authoritative semantics.
- `interpreter10/`: Bun-based TypeScript interpreter + AST definition (`src/ast.ts`) and extensive tests.
- `interpreter10-go/`: Go interpreter and canonical Able v10 runtime. Go-specific design docs live under `design/` (see `go-concurrency.md`, `typechecker.md`).
- Legacy work: `interpreter6/`, assorted design notes in `design/`, early stdlib sketches. Do not do any work in these directories.

## Ongoing Workstreams
- **Spec maintenance**: keep `spec/full_spec_v10.md` authoritative; log discrepancies in `spec/todo.md`.
- **Standard library**: coordinate with `stdlib/` efforts; ensure interpreters expose required builtin functions/types; track string/regex bring-up via `design/regex-plan.md` and the new spec TODOs covering byte-based strings with char/grapheme iterators.
- **Developer experience**: cohesive documentation, examples, CI improvements (Bun + Go test jobs).
- **Future interpreters**: keep AST schema + conformance harness generic to support planned Crystal implementation.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones now live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.

## Guardrails (must stay true)
- `interpreter10/scripts/run-parity.ts` is the authoritative entry point for fixtures/examples parity, and `run_all_tests.sh` must stay green (TS + Go unit tests, fixture suites, parity CLI).
- `interpreter10/testdata/examples/` remains curated; add programs only when parser/runtime support is complete (see `design/parity-examples-plan.md` for the active roster).
- Diagnostics parity is mandatory: both interpreters emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`. (✅ CLI now enforces diagnostics comparison; ✅ Go typechecker infers implicit `self`/iterator helpers so warn/strict runs match.)
- The JSON parity report (`tmp/parity-report.json`) must be archived in CI via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR` to keep regression triage fast.
- Track upcoming language work (channel select, advanced dynimport scenarios, interface dispatch additions) so fixtures/examples land immediately once syntax/runtime support exists.

## TODO

## DONE
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
- Generator laziness parity closed: iterator continuations now cover if/while/for/match across both runtimes, stdlib helpers (`stdlib/v10/src/concurrency/channel.able`, `stdlib/v10/src/collections/range.able`) use generator literals, and new fixtures (`fixtures/ast/control/iterator_*`, `fixtures/ast/stdlib/channel_iterator`, `fixtures/ast/stdlib/range_iterator`) keep the shared harness authoritative.
- Automatic time slicing verified for long-running procs: the new `concurrency/proc_time_slicing` fixture proves that handles without explicit `proc_yield()` still progress under repeated `proc_flush()` calls, capturing both the intermediate `Pending` status and the eventual resolved value across runtimes.
### AST → Parser → Typechecker Completion Plan _(reopen when new AST work appears)_
- Full sweep completed 2025-11-06 (strict fixture run, Go interpreter suite, Go parser harness, and `bun test` all green). Archive details in `LOG.md`; bring this plan back only if new AST/syntax changes introduce regressions.
