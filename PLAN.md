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

7. **Kernel vs stdlib layering (immediate priority)**
   - **Audit & contract:** enumerate current kernel natives (arrays, string helpers, concurrency primitives, hasher, string bridges) and define the minimal kernel in `design/stdlib-v11.md`/`spec/full_spec_v11.md` (keep only bridges/scheduler/channels/mutex/hasher).
   - **Port to Able:** reimplement array and string helpers in `v11/stdlib` atop kernel primitives; keep only minimal hooks native. Add stdlib tests proving behaviour and update interpreters/fixtures to rely on the stdlib surface.
   - **Runtime/typechecker refactor:** remove or shim native array/string methods to delegate to stdlib; source signatures from stdlib interfaces instead of hardcoded natives. Ensure iterators/IndexError surfaces remain stable.
   - **Parity/docs:** adjust fixtures/exporter to target stdlib implementations, keep `./run_all_tests.sh --version=v11` green, and document any temporary shims slated for removal.

8. **Stdlib API expansions (arrays, regex, §§6.12 & 14.2)**
   - **Array helpers:** expose `size`, `push`, `pop`, `get`, `set`, `clear` with proper `IndexError` handling and ensure array types continue to satisfy `Iterable`. (Interpreter builtins + tests added; confirm iterator/IndexError parity.)
   - **Regex module:** deliver `able.text.regex` with RE2-style determinism (core types, compile/match/split/replace helpers, regex sets, streaming scanner, grapheme-aware options) and bridge to host runtimes while guarding unsupported features via `RegexError`.
   - **Tests:** expand stdlib test suites + fixtures to cover each helper, regex compilation failures, streaming use cases, and confirm both interpreters return identical traces.

9. **Language-supported interface alignment (§14.1 + Awaitable interface)**
   - **Definitions:** codify the canonical interface declarations (Index/IndexMut, Iterable/Iterator, Apply, arithmetic/bitwise, comparison/hash, Display, Error, Clone, Default, Range, Proc/ProcError, Awaitable) inside the stdlib so syntax sugar lowers to these in both runtimes.
   - **Type checker/runtime integration:** ensure operator resolution consults these interfaces uniformly (e.g., `[]`, callable values, arithmetic ops), update parser + PT→AST lowering for indexing and callable value invocation, and guarantee diagnostics mention missing interfaces.
   - **Stdlib coverage:** audit built-in types (Array, Map, Range, Channel, Mutex, Regex types, async handles) for implementations of the required interfaces, updating fixtures/tests where gaps exist.
   - **Tests:** add parity fixtures verifying operator dispatch/method availability and run the full TS + Go suites plus `./run_all_tests.sh --version=v11` after changes.

10. **Kernel vs stdlib layering (minimize native surface)**
    - **Audit & contract:** enumerate current kernel natives (arrays, string helpers, concurrency, hasher, string bridges) and document the intended minimal kernel in `design/stdlib-v11.md` + `spec/full_spec_v11.md` (kernel stays byte/string bridges, scheduler, channels/mutex, hasher).
    - **Rebuild stdlib incrementally (current focus):**
      - Keep an empty working stdlib in `v11/stdlib/src` and a quarantined copy in `v11/stdlib/quarantine/`.
      - Restore one module at a time from quarantine → `v11/stdlib/src/`, fixing parse/typecheck/test issues before adding the next. Recommended order: `core/errors`, `core/interfaces`, `core/options`, `core/iteration` (already restored) → `collections/array` + `collections/enumerable` → `text/string` → `core/numeric` → `collections/range` → `collections/list` → `collections/vector` → `collections/hash_map`/`set` → remaining collections.
      - Current working set restored in `v11/stdlib/src`: core/errors, core/interfaces, core/options, core/iteration, core/numeric, collections/array + enumerable + range, text/string. Smoke tests live under `v11/stdlib/tests`.
      - For each module restored: add/restore minimal stdlib tests, run `bun test v11/interpreters/ts/test/stdlib/...` (or targeted), and `go test ./...` (fast) to keep suites green.
    - **Port to Able:** reimplement array helpers and string helpers in `v11/stdlib` atop kernel primitives; keep only bridges/low-level ops native. Add stdlib unit tests for the Able implementations and ensure interpreters call the stdlib versions in fixtures/examples.
    - **Runtime refactor:** demote native array/string methods to stdlib-backed calls (or remove where redundant), exposing only the minimal kernel hooks. Update typecheckers to source signatures from stdlib surfaces (or shared interface definitions) instead of hardcoded native members.
    - **Fixtures/parity:** update fixtures and exporter manifests to target stdlib-layer functions where behaviour moves; keep parity harness green after the shift.
    - **Docs/spec:** clarify kernel vs stdlib layering and mark any temporary native shims slated for removal; trim the PLAN item when the native surface matches the agreed minimal kernel.
      - Module loader now guards cycles and skips extern bodies at runtime; able.collections.array loads via run-module. Collections/enumerable helpers and text/string externs are back in src, with smoke tests validating them in the TS interpreter.
