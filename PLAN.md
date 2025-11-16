# Able Project Roadmap (v11 focus)

## Scope
- Keep the frozen Able v10 toolchain available for historical reference while driving all new language, spec, and runtime work in v11.
- Keep the Go interpreter as the behavioural reference and ensure the TypeScript runtime + future ports match feature-for-feature (the concurrency implementation strategy may differ).
- Preserve a single AST contract for every runtime so tree-sitter output can target both the historical v10 branch and the actively developed v11 runtime; document any deltas immediately in the v11 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.

## Existing Assets
- `spec/full_spec_v10.md`: authoritative semantics for the archived toolchain. Keep it untouched unless a maintainer requests an errata fix.
- `spec/full_spec_v11.md`: active specification for all current work; every behavioural change must be described here.
- `v10/interpreters/{ts,go}/`: Frozen interpreters that match the v10 spec. Treat them as read-only unless a blocking support request lands.
- `v11/interpreters/{ts,go}/`, `v11/parser`, `v11/fixtures`, `v11/stdlib`: active development surface for Able v11.
- Legacy work: `interpreter6/`, assorted design notes in `design/`, early stdlib sketches. Do not do any work in these directories.

## Ongoing Workstreams
- **Spec maintenance**: stage and land all wording in `spec/full_spec_v11.md`; log discrepancies in `spec/TODO_v11.md`. Reference the v10 spec only when clarifying historical behaviour.
- **Standard library**: coordinate with `stdlib/` efforts; ensure interpreters expose required builtin functions/types; track string/regex bring-up via `design/regex-plan.md` and the new spec TODOs covering byte-based strings with char/grapheme iterators.
- **Developer experience**: cohesive documentation, examples, CI improvements (Bun + Go test jobs).
- **Future interpreters**: keep AST schema + conformance harness generic to support planned Crystal implementation.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones now live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.

## Guardrails (must stay true)
- `v11/interpreters/ts/scripts/run-parity.ts` remains the authoritative entry point for fixtures/examples parity; `./run_all_tests.sh --version=v11` must stay green (TS + Go unit tests, fixture suites, parity CLI). Run the v10 suite only when explicitly asked to investigate archival regressions.
- `v11/interpreters/ts/testdata/examples/` remains curated; add programs only when parser/runtime support is complete (see the corresponding `design/parity-examples-plan.md` for the active roster).
- Diagnostics parity is mandatory: both interpreters emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`. (✅ CLI now enforces diagnostics comparison; ✅ Go typechecker infers implicit `self`/iterator helpers so warn/strict runs match.)
- The JSON parity report (`tmp/parity-report.json`) must be archived in CI via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR` to keep regression triage fast.
- Track upcoming language work (awaitable orchestration, advanced dynimport scenarios, interface dispatch additions) so fixtures/examples land immediately once syntax/runtime support exists.

## TODO (as items are completed, move them to LOG.md)

### v11 Spec Delta Implementation Plan

5. **Optional generic parameter inference (§7.1.5)**
   - **Type checker:** detect free type names in function signatures, synthesize implicit `<T>` lists, hoist constraints, and block redeclaration conflicts; diagnostics should mention inferred names vs. explicit ones.
   - **AST:** extend function nodes to capture inferred generics (so later phases know whether a parameter list was explicit).
   - **Parser & mapping:** ensure signatures without `<...>` still emit the information required by the type checker and avoid regressing explicit generic support.
   - **Tests:** add fixtures spanning implicit generics, ambiguity errors, and import interactions; keep both interpreters’ inference behaviour aligned.

6. **Loop/range constructs & continue semantics (§§8.2–8.3)**
   - **AST:** add `loop` expression nodes, `continue` statements, range operator nodes for `..`/`...`, and metadata on loop break payload types.
   - **Runtime:** implement expression-valued `loop {}` plus `while`/`for` break payload propagation, enforce unlabeled `continue` rules (runtime error for labeled continue), and ensure range literals materialize iterables via the stdlib `Range` interface.
   - **Parser/mapping:** recognize the new keywords/tokens, wire precedence so `..`/`...` bind tighter than assignment but looser than arithmetic, and emit AST nodes consumed by both interpreters.
   - **Tests:** expand fixtures for loops returning values, retry loops, continue behaviour, and range-driven `for` loops; keep `bun run scripts/run-fixtures.ts` + Go parity green.

7. **`await` expression, Awaitable protocol, and concurrency errors (§§12.6–12.7 & Awaitable interface)**
   - **AST:** represent `await [arms...]` expressions (including default arms) and persist callback bodies for codegen.
   - **Scheduler/runtime:** implement the `Awaitable` interface (is_ready/register/commit), fairness when multiple arms are ready, cancellation of losers, propagation of proc cancellation, and the default-arm fall-through semantics in both interpreters.
   - **Stdlib errors:** add the required channel/mutex error structs (`ChannelClosed`, `ChannelNil`, `ChannelSendOnClosed`, `ChannelTimeout`, `MutexUnlocked`, `MutexPoisoned`) plus constructors and make all channel/mutex helpers surface them consistently.
   - **Parser/tests:** parse the `await [...]` form, update fixtures to cover channel send/recv arms, timer/default arms, fairness simulations, and run `bun run scripts/run-fixtures.ts`, `go test ./pkg/interpreter`, and `./run_all_tests.sh --version=v11` after wiring the scheduler.

8. **Module + stdlib resolution surface (§§13.6–13.7)**
   - **Loader:** implement the new search order (workspace roots, cwd/manual overrides, `ABLE_PATH`, `ABLE_MODULE_PATHS`, canonical stdlib via `ABLE_STD_LIB` or bundled path), deduplicate roots, sanitize hyphenated names, and reserve the `able.*` namespace.
   - **Tooling safeguards:** treat stdlib directories as read-only, raise collisions when the same package path appears in multiple roots, and make `dynimport` share the same lookup order.
   - **Tests/docs:** add harness coverage and documentation for the environment knobs, including overrides for fixtures and REPL usage.

9. **Stdlib API expansions (strings, arrays, regex, §§6.12 & 14.2)**
   - **String helpers:** implement the required `string` methods (`len_bytes/chars/graphemes`, `substring`, `split`, `replace`, `starts_with`, `ends_with`, iterator helpers) with byte/grapheme semantics shared across interpreters.
   - **Array helpers:** expose `size`, `push`, `pop`, `get`, `set`, `clear` with proper `IndexError` handling and ensure array types continue to satisfy `Iterable`.
   - **Regex module:** deliver `able.text.regex` with RE2-style determinism (core types, compile/match/split/replace helpers, regex sets, streaming scanner, grapheme-aware options) and bridge to host runtimes while guarding unsupported features via `RegexError`.
   - **Tests:** expand stdlib test suites + fixtures to cover each helper, regex compilation failures, streaming use cases, and confirm both interpreters return identical traces.

10. **Language-supported interface alignment (§14.1 + Awaitable interface)**
   - **Definitions:** codify the canonical interface declarations (Index/IndexMut, Iterable/Iterator, Apply, arithmetic/bitwise, comparison/hash, Display, Error, Clone, Default, Range, Proc/ProcError, Awaitable) inside the stdlib so syntax sugar lowers to these in both runtimes.
   - **Type checker/runtime integration:** ensure operator resolution consults these interfaces uniformly (e.g., `[]`, callable values, arithmetic ops), update parser + PT→AST lowering for indexing and callable value invocation, and guarantee diagnostics mention missing interfaces.
   - **Stdlib coverage:** audit built-in types (Array, Map, Range, Channel, Mutex, Regex types, async handles) for implementations of the required interfaces, updating fixtures/tests where gaps exist.
   - **Tests:** add parity fixtures verifying operator dispatch/method availability and run the full TS + Go suites plus `./run_all_tests.sh --version=v11` after changes.
