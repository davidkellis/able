# Able Project Roadmap (v11 focus)

Standard onboarding prompt:
Read AGENTS, PLAN, and the v11 spec, then start on the highest priority PLAN work. Proceed with next steps. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. You have permission to run tests. Tests should run quickly; no test should take more than one minute to complete.

Standard next steps prompt:
Proceed with next steps as suggested; don't talk about doing it - do it. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. Tests should run quickly; no test should take more than one minute to complete.

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
- Status: interface alignment is in place (Apply/Index/Awaitable + stdlib helpers) and the v11 sweep is green; next focus is regex stdlib expansions and tutorial cleanup.

## Guardrails (must stay true)
- `v11/interpreters/ts/scripts/run-parity.ts` remains the authoritative entry point for fixtures/examples parity; `./run_all_tests.sh --version=v11` must stay green (TS + Go unit tests, fixture suites, parity CLI). Run the v10 suite only when explicitly asked to investigate archival regressions.
- `v11/interpreters/ts/testdata/examples/` remains curated; add programs only when parser/runtime support is complete (see the corresponding `design/parity-examples-plan.md` for the active roster).
- Diagnostics parity is mandatory: both interpreters emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`. (✅ CLI now enforces diagnostics comparison; ✅ Go typechecker infers implicit `self`/iterator helpers so warn/strict runs match.)
- The JSON parity report (`tmp/parity-report.json`) must be archived in CI via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR` to keep regression triage fast.
- Track upcoming language work (awaitable orchestration, advanced dynimport scenarios, interface dispatch additions) so fixtures/examples land immediately once syntax/runtime support exists.

## TODO (working queue: tackle in order, move completed items to LOG.md)

### Go stdlib/typechecker parity (no shortcuts; required for ablego/tutorials)
- [ ] **Ordering/Ord correctness:** Align `able.core.interfaces` and `able.core.errors` with the Go typechecker so `Ord.cmp` returns the declared `Ordering` variants and primitive/Error impls satisfy parameter/return types.
- [ ] **Extern acceptance:** Teach the Go typechecker to accept extern function bodies in stdlib modules (e.g., `able.collections.array` extern helpers) instead of rejecting them as unsupported statements.
- [ ] **Array Iterable/Iterator self typing:** Fix Go typechecker handling of `Iterable`/`Iterator` impls in `able.collections.array` so `Self` matches instantiated type args and member accesses (`length`, `storage_handle`, etc.) are recognized.
- [ ] **Clone/Default/Index parity:** Ensure the Go typechecker validates stdlib impls for Clone/Default/Index/IndexMut on primitives and Array without spurious generic-arity errors.
- [ ] **Rule: do not bypass stdlib checks:** Keep stdlib typechecking enabled in ablego; fix checker + stdlib so `v11/examples/tutorial/04_functions_and_lambdas.able` (and the rest of the tutorials) run cleanly with the full stdlib loaded.

### v11 Spec Delta Implementation Plan
- [ ] **Correct examples**: Fix broken examples in the v11/examples directory.
   - help me use the ablets and ablego scripts in the v11 directory to run the example programs in v11/examples/* and let's start identifying which ones don't work and then figure out what we'll need to do to correct them or fix the interpreters so that the example programs work. I want to identify problems first and capture those items in the PLAN and start working through why they don't work and getting things corrected where they're broken, expanding stdlib to make things work, etc.
   - identify broken examples that are due to missing stdlib features
   - identify broken examples that are due to language features from the v11 spec that are not yet implemented yet or are currently not working properly
      - these issues are the highest priority and should be addressed first
   - TS ablets sweep (current): everything passes except:
     - `greet_typecheck_fail.able` (intentional type error)
     - `leetcode4_median_of_two_sorted_arrays.able` — needs numeric conversion support (`as f64` / return type mismatch)
      - `leetcode8_string_to_integer.able` — literal `2147483648` overflows `i32` (example should widen or adjust literal)
      - `leetcode9_palindrome_number.able` — parser rejects `and` (tree-sitter parse error; grammar needs `and` or the sample should use `&&`)
   - Tutorials (TS): 01–14 typecheck after stdlib iterator/channel fixes; Go parity still pending.
- [ ] **Stdlib numeric conversions:** Add explicit conversion helpers on numeric types (e.g., to_f64, to_i64) in the stdlib and wire through TS/Go so examples/tests can rely on them instead of the nonexistent `as` cast.
- [ ] **Stdlib API expansions (regex, §§6.12 & 14.2)** — **lowest priority; defer until higher items advance.**
   - **Tests:** expand stdlib test suites + fixtures to cover each helper, regex compilation failures, streaming use cases, and confirm both interpreters return identical traces.

### Tutorials & Examples (cleanup backlog)
- v11/examples triage (ablets/ablego):
  - Root examples: `assign/greet/hello_world/loop` pass in TS (Go sweep pending); `greet_typecheck_fail` intentionally errors.
  - Tutorials: TS 01–14 typecheck; Go parity still needs a fresh run after stdlib plumbing.
  - Leetcode (TS): all but 4/8/9 typecheck (see above); Go not yet rerun.
- Fix tutorials requiring missing stdlib imports/stubs (Channel/Mutex await sample, array and string helpers) once the stdlib surfaces are wired back in; align Apply/Callable support across runtimes (Go still needs parity).
- Update options/results/exceptions/interop tutorials to reference/import the stdlib `Error` interface so they typecheck (done for TS; Go still pending).
- Adjust concurrency proc/spawn tutorial to avoid `proc_yield` misuse and return-type mismatches; revisit union-pattern check in Go (tutorial 03) and add Go support for prelude/extern.
- Go fixture/typechecker hygiene: keep overload fixtures (`overload_resolution_success`, UFCS overload cases) green by allowing method overloads without duplicate diagnostics and keeping diagnostic paths aligned with the TS harness.
