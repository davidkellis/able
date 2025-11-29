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
- Status: typed-pattern match reachability regression is fixed (suites green); stay focused on the stdlib layering work in TODO items 7–10.

## Guardrails (must stay true)
- `v11/interpreters/ts/scripts/run-parity.ts` remains the authoritative entry point for fixtures/examples parity; `./run_all_tests.sh --version=v11` must stay green (TS + Go unit tests, fixture suites, parity CLI). Run the v10 suite only when explicitly asked to investigate archival regressions.
- `v11/interpreters/ts/testdata/examples/` remains curated; add programs only when parser/runtime support is complete (see the corresponding `design/parity-examples-plan.md` for the active roster).
- Diagnostics parity is mandatory: both interpreters emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`. (✅ CLI now enforces diagnostics comparison; ✅ Go typechecker infers implicit `self`/iterator helpers so warn/strict runs match.)
- The JSON parity report (`tmp/parity-report.json`) must be archived in CI via `ABLE_PARITY_REPORT_DEST`/`CI_ARTIFACTS_DIR` to keep regression triage fast.
- Track upcoming language work (awaitable orchestration, advanced dynimport scenarios, interface dispatch additions) so fixtures/examples land immediately once syntax/runtime support exists.

## TODO (working queue: tackle in order, move completed items to LOG.md)

### v11 Spec Delta Implementation Plan

8. **Stdlib API expansions (regex, §§6.12 & 14.2)** — **lowest priority; defer until higher items advance.**
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
      - Current working set restored in `v11/stdlib/src`: core/errors, core/interfaces, core/options, core/iteration, core/numeric, collections/array + enumerable + range + list + linked_list + vector + deque + queue + lazy_seq + hash_map + set + hash_set + heap + bit_set + tree_map + tree_set, text/string. Smoke tests live under `v11/stdlib/tests`.
      - For each module restored: add/restore minimal stdlib tests, run `bun test v11/interpreters/ts/test/stdlib/...` (or targeted), and `go test ./...` (fast) to keep suites green.
    - **Port to Able:** reimplement array helpers and string helpers in `v11/stdlib` atop kernel primitives; keep only bridges/low-level ops native. Add stdlib unit tests for the Able implementations and ensure interpreters call the stdlib versions in fixtures/examples.
    - **Runtime refactor:** demote native array/string methods to stdlib-backed calls (or remove where redundant), exposing only the minimal kernel hooks. Update typecheckers to source signatures from stdlib surfaces (or shared interface definitions) instead of hardcoded native members.
    - **Fixtures/parity:** update fixtures and exporter manifests to target stdlib-layer functions where behaviour moves; keep parity harness green after the shift.
    - **Docs/spec:** clarify kernel vs stdlib layering and mark any temporary native shims slated for removal; trim the PLAN item when the native surface matches the agreed minimal kernel.
      - Module loader now guards cycles and skips extern bodies at runtime; able.collections.array loads via run-module. Collections/enumerable helpers and text/string externs are back in src, with smoke tests validating them in the TS interpreter.
      - String helpers now exposed via `methods string`; TS interpreter + typechecker dispatch string members through stdlib-first resolution (with iterator adapters for struct-based iterators) and targeted stdlib override/integration tests added.
      - Began hiding the interim `String` wrapper: stdlib string helpers now return built-in strings with `u64` lengths, `String` is package-private, and `StringBuilder` operates on the built-in `string` surface.
      - Restored persistent `collections/list` into `v11/stdlib/src` with smoke coverage and TS ModuleLoader integration exercising concat/reverse/to_array + iteration.
      - Restored persistent `collections/vector` into `v11/stdlib/src` with smoke coverage plus a TS ModuleLoader integration ensuring push/set/pop + iteration work via the stdlib surface.
      - Restored persistent `collections/deque` and `collections/queue` into `v11/stdlib/src` with smoke coverage and TS ModuleLoader integration; both interpreters' typecheckers now treat them as for-loop iterables.
      - Restored persistent `collections/bit_set` into `v11/stdlib/src` with smoke + TS ModuleLoader coverage; both typecheckers now treat BitSet as a for-loop iterable of `i32`.
      - Restored `collections/tree_map` (sorted array-backed) and `collections/tree_set` into `v11/stdlib/src` with smoke coverage plus TS ModuleLoader integration validating ordering/updates.
      - Ordering equality + method dispatch for primitives fixed; TreeMap now routes through `Ord.cmp` again and covers custom key types in TS ModuleLoader tests.
      - HashSet stdlib use now typechecks without filtering diagnostics: imports seed HashSet bindings from package summaries and builtin method stubs cover new/with_capacity/add/remove/contains/size/clear/is_empty.
      - Typechecker builtin stubs now cover Array/List/Vector/HashMap with bool-aware signatures; iterable recognition includes these collections and nested generic applications (e.g., `HashMap string i32`) flatten correctly, so cross-module stdlib calls typecheck without coercing to `unknown`.
      - Go typechecker now recognises stdlib collections (`List`, `Vector`, `HashSet`, `Deque`, `Queue`, `BitSet`) as valid for-loop iterables; iterable helpers moved into their own file with regression coverage to keep `type_utils.go` under the 1k-line guardrail.
      - Go runtime + typechecker now prefer stdlib-defined Array methods before native shims, so method-set overrides and stdlib signatures apply with native fallbacks kept for compatibility.
      - Full TS + Go suite is green after the array stdlib-first resolution path (`./run_all_tests.sh --version=v11`).
      - TS interpreter typed-pattern matching now recognises interface implementations (e.g., `RangeError`, `IndexError`) for `Error` patterns, and stdlib string/array integration tests now cover RangeError/IndexError surfaces via the ModuleLoader.
      - Additional stdlib-first edge coverage landed for string split/replace and array pop/get/set out-of-bounds paths; native string replace fallback now returns the receiver when the needle is empty to mirror stdlib behaviour.
      - Fixed tree-sitter number literal mapping to keep hex literals as integers (even with `E` digits), so UTF-8 path bitwise checks no longer error and the multi-byte string split/replace tests are re-enabled.
      - Began trimming native array helpers in the TS interpreter: push/pop/get/set/size/clear now require the stdlib method set (Iterator fallback retained), and leetcode examples import `able.collections.array` to use the stdlib surface.
      - Removed remaining native string helpers in both interpreters (array iterator shim retained), mirroring Go/TS runtime and typechecker behaviour and adding guardrail tests for missing stdlib imports; added coverage to flag base-prefixed octal/binary literals with stray `e/E` markers so they surface diagnostics instead of being treated as exponents.
      - Aligned stdlib interface names with the spec (Apply.apply and Index/IndexMut.get/set return results), updating Array/HashMap implementations with inherent wrappers to avoid method name ambiguities in impl resolution.
      - TS typechecker now treats primitive strings as byte iterables (u8) to match the Go runtime/typechecker, ModuleLoader coverage exercises string for-loops via the stdlib iterator surface, and Go gets matching typechecker + runtime iterator coverage once the stdlib string module is loaded.
### Tutorials & Examples (cleanup backlog)
- Fix tutorials requiring missing stdlib imports/stubs (Channel/Mutex await sample) once the stdlib surfaces are wired back in.
- Update options/results/exceptions/interop tutorials to reference the `Error` interface correctly (or import the stdlib definition) so they typecheck.
- Adjust concurrency proc/spawn tutorial to avoid `proc_yield` misuse and return-type mismatches.
- Align package/import tutorials with correct package names (`tutorial_packages_*`) so the loader resolves dependencies.
