# Able Project Log

# 2026-01-24 — Iterator interface returns + constraint-arity fixture cleanup (v11)
- Go interpreter: treat Iterator interface return values as iterators during for/each by accepting `IteratorValue` in `adaptIteratorValue`.
- Fixtures: removed duplicate constraint-interface-arity diagnostics from exported manifests via the TS fixture definitions; re-exported fixtures.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/go && ABLE_TYPECHECK_FIXTURES=strict go test ./pkg/interpreter -run 'TestFixtureParityStringLiteral/errors/constraint_interface_arity' -count=1`; `./run_all_tests.sh --version=v11`.

# 2026-01-24 — Interface dictionary arg dispatch + fixture expansion (v11)
- Go interpreter: coerce interface-typed generic values into interface dictionaries so interface arguments are preserved for bindings, params, and return coercions.
- Fixtures: added interface dictionary exec coverage for default chains, overrides, named impl + inherent method calls, interface inheritance, interface-arg dispatch (bindings/params/returns), and union-target dispatch; added AST error fixtures for ambiguous impl constraints + missing interface methods; updated exec coverage index and typecheck baseline.
- Tests: `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_11_interface_generic_args_dispatch -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_12_interface_union_target_dispatch -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_13_interface_param_generic_args -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_14_interface_return_generic_args -count=1`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_13_interface_param_generic_args bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_14_interface_return_generic_args bun run scripts/run-fixtures.ts`.

# 2026-01-24 — Named impl method resolution fix (v11)
- Interpreters (TS + Go): attach named-impl context to impl methods so default methods (and peers) can resolve sibling methods via `self.method()` in the same impl.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_05_interface_named_impl_defaults bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_06_interface_generic_param_dispatch bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_05_dynimport_interface_dispatch bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_05_interface_named_impl_defaults`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/10_06_interface_generic_param_dispatch`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/13_05_dynimport_interface_dispatch`.

# 2026-01-24 — Unified Future regression passes + fs rename fix (v11)
- Stdlib fs (TS): switched `fs_remove`/`fs_rename` externs to sync Node calls to avoid flaky exists checks during stdlib runs.
- Plan: removed the unified Future regression-pass TODO after completing the full sweeps.
- Tests: `./run_all_tests.sh --version=v11`; `./run_stdlib_tests.sh`.

# 2026-01-24 — Go test wording cleanup (v11)
- Go tests: updated concurrency/await diagnostics to say “future handle” instead of “proc handle”.
- Tests not run (text-only change).

# 2026-01-24 — Interface dictionary exec coverage expansion (v11)
- Exec fixtures: added coverage for named impl disambiguation across packages with defaults, generic-constraint dispatch across packages, and dynimport-returned interface values.
- Coverage index: registered `exec/10_05_interface_named_impl_defaults`, `exec/10_06_interface_generic_param_dispatch`, and `exec/13_05_dynimport_interface_dispatch`.
- Tests not run (fixture-only changes).

# 2026-01-24 — Unified Future model naming cleanup (v11)
- TypeScript tests: renamed `proc_spawn_*` concurrency tests/helpers to `future_spawn_*` and updated imports (including await tests).
- Docs: updated `v11/interpreters/ts/README.md` and `v11/stdlib/src/README.md` to use Future terminology.
- Plan: collapsed the unified Future model checklist to the remaining regression-pass item.
- Tests not run (rename + docs + plan updates only).

# 2026-01-24 — Bytecode VM prototype (v11)
- TypeScript interpreter: added a minimal stack-based bytecode VM plus a small AST->bytecode lowering path (literals, identifiers, `:=`/`=`, `+`, blocks).
- Tests: added VM-vs-tree-walker conformance checks for literals, assignment + arithmetic, and module bodies.
- Tests: `cd v11/interpreters/ts && bun test test/vm/bytecode_vm.test.ts`.

# 2026-01-24 — Core IR + runtime ABI design (v11)
- Design: expanded `v11/design/compiler-interpreter-vision.md` with a typed core IR outline and runtime ABI surface (interface dictionaries, concurrency, errors, dynamic hooks).
- Tests not run (design-only update).

# 2026-01-24 — Stdlib copy helpers speedup (v11)
- Stdlib fs: routed `copy_file`/`copy_dir` through host externs (Go + TS) and removed the Able-level directory traversal to keep `copy_dir` under the test-time budget.
- Tests: `./v11/ablets test v11/stdlib/tests/fs.test.able --format tap --name "able.fs::copies directory trees with overwrite control"` (≈59.6s); `./run_stdlib_tests.sh`; `./run_all_tests.sh --version=v11`.

# 2026-01-24 — Interface dispatch fixture coverage (v11)
- Exec fixtures: added `exec_10_04_interface_dispatch_defaults_generics` to cover cross-package default interface methods + generic interface method calls on interface-typed values.
- Coverage index: registered the new exec fixture entry.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_04_interface_dispatch_defaults_generics bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter`.

# 2026-01-24 — Future handle test/fixture cleanup (v11)
- TypeScript tests: align await stdlib integration helper with `Future.value() -> !T` by returning `!Array bool`.
- Exec fixtures: update `12_04_future_handle_value_view` stdout expectations to match the renamed output text.
- Go tests: disambiguate duplicate future-handle/serial-executor test names introduced during the future renames.
- Tests: `./v11/export_fixtures.sh`; `./run_all_tests.sh --version=v11`; `./run_stdlib_tests.sh`.

# 2026-01-24 — Regex quantifier parsing + scan (v11)
- Stdlib regex: implemented literal-token parsing with quantifiers (`*`, `+`, `?`, `{m}`, `{m,}`, `{m,n}`), updated match/find_all/scan to use token spans, and fixed a match-case return syntax issue.
- Design: updated `v11/design/regex-plan.md` to reflect the partial Phase 1 status and active regex fixture.
- Tests: `./v11/ablets check v11/stdlib/src/text/regex.able`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=14_02_regex_core_match_streaming bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter`.

# 2026-01-23 — Constraint arity regression coverage (v11)
- Typechecker (TS): added constraint interface arity diagnostics for missing/mismatched interface type arguments.
- Fixtures/tests: added `errors/constraint_interface_arity` fixture (calls wrapped in a non-invoked helper to avoid runtime errors), plus new TS/Go typechecker regression tests for constraint arity.
- Baseline: regenerated `v11/fixtures/ast/typecheck-baseline.json` after the fixes.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts --write-typecheck-baseline`.

# 2026-01-23 — Typechecker default enforcement (v11)
- Harnesses: defaulted fixture/test/parity typechecking to strict, with explicit warn/off overrides; run_all_tests now always passes `ABLE_TYPECHECK_FIXTURES` (strict by default) into fixture/parity/Go test runs.
- Docs: refreshed v11 + interpreter readmes and parity reporting notes to reflect strict defaults and explicit overrides.
- Tests not run (docs + harness configuration only).

# 2026-01-23 — Stdlib typecheck verification (v11)
- Verified stdlib packages typecheck cleanly in TS + Go by importing all stdlib packages via a temporary stdlib typecheck harness.
- Verified `.examples/foo.able` runs with strict typechecking in both interpreters.
- Tests: `./v11/ablets check tmp/stdlib_typecheck.able`; `./v11/ablego check tmp/stdlib_typecheck.able`; `./v11/ablets .examples/foo.able`; `./v11/ablego .examples/foo.able`.

# 2026-01-22 — Stdlib Eq/Ord/Hash audit (v11)
- Stdlib: audited `v11/stdlib/src` for generic `Eq`/`Ord`/`Hash` constraints and kernel alias usage; no type-arg uses remain, so the PLAN item was cleared.
- Tests not run (audit + plan/log updates only).

# 2026-01-22 — Go parser nested generics fix (v11)
- Go parser: avoid flattening parenthesized generic applications in type argument lists so nested types like `Iterable (Option String)` stay nested; fixes `TestParseInterfaceCompositeNestedGenerics` and exec fixture `04_03_type_expression_syntax`.
- Tests: `cd v11/interpreters/go && go test ./pkg/parser -run TestParseInterfaceCompositeNestedGenerics -count=1`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/04_03_type_expression_syntax -count=1`; `./run_all_tests.sh --version=v11`; `./run_stdlib_tests.sh`.

# 2026-01-22 — TS typechecker import symbol scoping (v11)
- Typechecker (TS): record symbol origins for imports/locals and filter unqualified/UFCS free-function resolution to the explicitly imported symbol source; builtins tagged for scope filtering.
- Tests: `bun test test/typechecker/function_calls.test.ts` and `bun test test/typechecker/duplicates.test.ts` in `v11/interpreters/ts`.
- Tests: `./run_stdlib_tests.sh --version=v11`.

# 2026-01-21 — TS typechecker import scoping for functions (v11)
- Typechecker (TS): track current package for function infos, filter call resolution to imported packages, and avoid cross-package duplicate declaration errors in stdlib/test runs.
- Tests: `./run_stdlib_tests.sh --version=v11`.

# 2026-01-21 — Full test runs (v11)
- Tests: `./run_all_tests.sh --version=v11`.

# 2026-01-21 — Full test runs (v11)
- Tests: `./run_all_tests.sh --version=v11` (passed).
- Tests: `./run_stdlib_tests.sh --version=v11` failed in TS stdlib due to duplicate declaration diagnostics (e.g., `able.text.regex` vs `able.text.string`, `able.spec` vs `able.spec.assertions`, and duplicated helpers across stdlib collection smoke tests); Go stdlib tests passed.

# 2026-01-21 — TS interpreter types refactor (v11)
- Refactored `v11/interpreters/ts/src/interpreter/types.ts` into `types/format.ts`, `types/primitives.ts`, `types/unions.ts`, `types/structs.ts`, and `types/helpers.ts` while keeping the public augmentations unchanged.
- Tests not run (refactor only).

# 2026-01-21 — TS impl resolution refactor (v11)
- Split `v11/interpreters/ts/src/interpreter/impl_resolution.ts` into stage modules for constraints, candidates, specificity ranking, defaults, and diagnostics under `v11/interpreters/ts/src/interpreter/impl_resolution/`.
- Tests not run (refactor only).

# 2026-01-21 — TS implementation collection refactor (v11)
- Split `v11/interpreters/ts/src/typechecker/checker/implementation-collection.ts` into collection vs validation/self-pattern helpers in `implementation-validation.ts`.
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/implementation_validation.test.ts`.

# 2026-01-21 — TS function call refactor (v11)
- Split `v11/interpreters/ts/src/typechecker/checker/function-calls.ts` into call-shape parsing, overload resolution, and diagnostics helpers (`function-call-parse.ts`, `function-call-resolve.ts`, `function-call-errors.ts`).
- Tests: `cd v11/interpreters/ts && bun test test/typechecker/function_calls.test.ts`; `cd v11/interpreters/ts && bun test test/typechecker/method_sets.test.ts`; `cd v11/interpreters/ts && bun test test/typechecker/apply_interface.test.ts`.

# 2026-01-21 — TypeCheckerBase size trim (v11)
- Extracted `checkFunctionDefinition`/`checkReturnStatement` into `v11/interpreters/ts/src/typechecker/checker/checker_base_functions.ts` and trimmed `v11/interpreters/ts/src/typechecker/checker_base.ts` under 900 lines.
- Tests not run (refactor only).

# 2026-01-21 — Interfaces fixture refactor (v11)
- Split `v11/interpreters/ts/scripts/export-fixtures/fixtures/interfaces.ts` into modular fixture files under `v11/interpreters/ts/scripts/export-fixtures/fixtures/interfaces/`.
- Tests not run (fixture refactor only).

# 2026-01-20 — Go typechecker builtin arity fallback (v11)
- Go typechecker: prefer builtin arity when generic base names are not found in the env, fixing Array T without explicit imports in parity examples.
- State: `./run_all_tests.sh --version=v11` green; parity examples passing.
- Next: resume PLAN TODOs (regex parser + quantifiers).

# 2026-01-19 — Eager/lazy collections verification (v11)
- Design: noted that `String`/`StringBuilder` keep bespoke eager `map`/`filter` (char-only) until an HKT wrapper exists (`v11/design/collections-eager-lazy-split.md`).
- Plan/spec: removed the eager/lazy TODO item from `PLAN.md` and cleared the resolved spec TODO entry (`spec/TODO_v11.md`).
- Tests: `./v11/ablets test v11/stdlib/tests/enumerable.test.able --format tap`; `./v11/ablego test v11/stdlib/tests/enumerable.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/text/string_methods.test.able --format tap`; `./v11/ablego test v11/stdlib/tests/text/string_methods.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/text/string_builder.test.able --format tap`; `./v11/ablego test v11/stdlib/tests/text/string_builder.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/collections/hash_map_smoke.test.able --format tap` (no tests found); `./v11/ablego test v11/stdlib/tests/collections/hash_map_smoke.test.able --format tap` (no tests found).

# 2026-01-19 — Iterator.map generator yield fix (v11)
- Stdlib: wrapped `Iterator.map` match-arm yields in a block so generator resumes advance instead of repeating the same value.
- Tests: `./v11/ablets test v11/stdlib/tests/enumerable.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/collections/hash_map_smoke.test.able --format tap` (no tests found).

# 2026-01-19 — Eager/lazy collections wiring (v11)
- Stdlib: moved lazy adapters to `Iterator`, introduced HKT `Enumerable` + `lazy` bridge, and refactored collection `Enumerable` impls/`Iterable` overrides.
- Collections: added shared `MapEntry` in `able.collections.map`, switched `PersistentMap` to `Iterable` entries, removed `Enumerable` from `BitSet`, and added HashMap value iteration plus value-based `map` (keys preserved).
- Plan: removed completed eager/lazy sub-steps from `PLAN.md`.
- Tests not run (stdlib + plan updates only).

# 2026-01-17 — Iterable collect type refs (v11)
- Interpreters: bind generic type parameters as runtime type refs so static interface methods (like `C.default()`) resolve in TS + Go.
- Stdlib/tests: disambiguate `collect` by terminating the `C.default()` statement and import `able.collections.array` in iteration tests so `Array` impls load for the Go typechecker.
- Tests: `./v11/ablets test v11/stdlib/tests/core/iteration.test.able`; `./v11/ablego test v11/stdlib/tests/core/iteration.test.able`.

# 2026-01-15 — Hash/Eq fixture and test coverage (v11)
- Fixtures: added AST fixtures for primitive hashing, kernel hasher availability, custom Hash/Eq, and collision handling; added exec fixtures for primitive hashing plus custom Hash/Eq + collisions; updated exec coverage index.
- Tests: added TS + Go unit coverage for hash helper builtins and kernel HashMap dispatch (custom + collision keys).
- Tests not run (edited code + fixtures only).

# 2026-01-15 — Remove host hasher bridges (v11)
- Kernel: dropped the `__able_hasher_*` extern declarations and the unused `HasherHandle` alias so hashing flows through `KernelHasher` only.
- Interpreters: removed host hasher state/builtins from Go + TypeScript, along with runtime stub and typechecker builtin entries.
- Docs/spec: scrubbed hasher bridge references from the kernel contract and extern execution/design notes.
- Tests not run (edited code + docs only).

# 2026-01-15 — Kernel Hash/Eq runtime alignment (v11)
- Kernel: added primitive `Eq`/`Ord`/`Hash` impls (ints/bool/char/String) plus float-only `PartialEq`/`PartialOrd`, and wired the Able-level FNV-1a hasher with raw byte helpers.
- Stdlib: trimmed duplicate interface impls and routed map hashing through the sink-style `Hash.hash` API.
- Interpreters: hash map kernels now dispatch `Hash.hash`/`Eq.eq`; numeric equality follows IEEE semantics; Go/TS typecheckers exclude floats from `Eq`/`Hash`.
- Fixtures: added float equality + hash-map key rejection exec coverage.
- Tests not run (edited code + docs only).

# 2026-01-15 — Kernel interfaces + Hash/Eq plan (v11)
- Documented the kernel-resident interface plan for primitive `Hash`/`Eq`, including runtime/stdlib/typechecker alignment, spec deltas, and fixture coverage (`v11/design/kernel-interfaces-hash-eq.md`, `v11/design/stdlib-v11.md`).
- Captured the spec update checklist in `spec/TODO_v11.md` and expanded the roadmap work breakdown in `PLAN.md`.
- Tests not run (planning/doc updates only).

# 2026-01-15 — Manual syntax alignment (v11)
- Manual docs now reflect v11 lexical rules and pipe semantics: line comments use `##`, string literals can be double-quoted or backticks with interpolation, and pipe docs no longer reference a `%` topic token (`v11/docs/manual/manual.md`, `v11/docs/manual/variables.html`).
- Tests not run (docs-only changes).

# 2026-01-15 — Primitive Hash/Eq constraints (v11)
- TypeScript typechecker now treats primitive numeric types as satisfying `Hash`/`Eq` constraints (matching Go) and the example iterates directly over `String` so `for` sees an `Iterable` (`v11/interpreters/ts/src/typechecker/checker/implementation-constraints.ts`, `.examples/foo.able`).
- Tests: `./v11/ablets .examples/foo.able`; `./v11/ablego .examples/foo.able`.

# 2026-01-13 — Runtime diagnostics formatting (v11)
- Runtime errors now emit `runtime:` diagnostics with locations + call-site notes in both interpreters; CLI/runtime harnesses share the same formatter.
- Added Go runtime diagnostic formatting test and updated exec fixture stderr expectations to include locations.
- Tests: `cd v11/interpreters/go && go test ./pkg/interpreter -run RuntimeDiagnostics`; `cd v11/interpreters/ts && bun test test/runtime/runtime_diagnostics.test.ts`.

# 2026-01-13 — Parser diagnostics formatting (v11)
- Parser diagnostics: route syntax/mapping failures through shared formatting, add span/expectation extraction from tree-sitter/mapper nodes, and normalize parser error messages for CLI output.
- CLI/tests: TS + Go loaders now surface parser diagnostics in the same format as typechecker output; added parser diagnostic tests for localized expectation messages.
- Tests: `cd v11/interpreters/ts && bun test test/parser/diagnostics.test.ts`; `cd v11/interpreters/go && go test ./pkg/driver -run ParserDiagnostics`.

# 2026-01-13 — Diagnostics groundwork + union normalization (v11)
- Design: added diagnostics overhaul note with warning policy, span/notes shape, and union normalization rules (`v11/design/diagnostics-overhaul.md`).
- Typecheckers: normalized unions with nullable/alias expansion, redundant-member warnings, and single-member collapse (TS + Go); added warning severity plumbing and location end spans.
- CLI/fixtures: warning-prefixed formatting for typechecker diagnostics in TS + Go; Go CLI diagnostics now use location-first formatting.
- Tests: `bun test test/typechecker/union_normalization.test.ts`; `go test ./pkg/typechecker -run 'UnionNormalization|NormalizeUnionTypes'`.

# 2026-01-13 — Lowercase path package cleanup (v11)
- Stdlib: ensured the path module works under the lowercase package name by importing `Path` into stdlib tests and avoiding module shadowing in `fs.write_bytes`.
- Tests: `./v11/ablets test v11/stdlib/tests/path.test.able`; `./run_stdlib_tests.sh --version=v11`; `./run_all_tests.sh --version=v11`.

# 2026-01-13 — Stdlib fs convenience helpers (v11)
- Stdlib fs: added `read_lines`, `write_lines`, `copy_file`, `copy_dir`, `touch`, `remove_file`, and `remove_dir` helpers; `touch` now uses host `utimes`/`Chtimes`, `copy_dir` uses an explicit task stack to avoid iterator re-entrancy, and `fs_path` prioritizes string inputs to keep Go/TS behavior aligned.
- Tests: expanded `v11/stdlib/tests/fs.test.able` to cover line IO, copy helpers + overwrite behavior, touch, and explicit remove wrappers.
- Tests: `./v11/ablets test v11/stdlib/tests/fs.test.able`; `./v11/ablego test v11/stdlib/tests/fs.test.able`.

# 2026-01-13 — Path API completion (v11)
- Stdlib Path: added `current`/`home`/`absolute`/`expand_home`/`normalize` helpers, `/` join sugar, and filesystem wrappers (`exists`, `is_file`, `is_dir`, `stat`, `read_text`, `write_text`).
- Go typechecker: allow `/` to resolve via `Div` interface implementations when operands are non-numeric.
- Tests: expanded `v11/stdlib/tests/path.test.able` with cwd/home/absolute/expand_home, join sugar, and fs helper coverage.
- Tests: `./v11/ablets test v11/stdlib/tests/path.test.able`; `./v11/ablego test v11/stdlib/tests/path.test.able`.

# 2026-01-13 — Proc cancellation test alignment (v11)
- TS tests: move cooperative cancellation check to run after `proc_yield`, aligning with proc resume semantics.
- Tests: `./run_all_tests.sh --version=v11`.

# 2026-01-13 — TS call continuation yields (v11)
- TS interpreter: preserve function-call environments across proc yields so async extern calls resume without replaying earlier statements; stop resetting block/module indices on manual yields.
- Tests: `./run_stdlib_tests.sh`.

# 2026-01-11 — TS concurrency continuation fixes (v11)
- TS interpreter: added module-level continuation state so entrypoint yields resume without replaying declarations; await commit now resumes across cooperative yields; future/proc awaitBlocked handling unified; proc/future value waits handle immediate waker completion.
- Tests: `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts`; `./run_all_tests.sh --version=v11`.

# 2026-01-11 — Stdlib OS/process/term scaffolding (v11)
- Stdlib: added `able.os`, `able.process`, and `able.term` modules with host externs and Able-level helpers (`ProcessSpec`, `ProcessStatus`, `ProcessSignal`, `TermSize`, line editor, env/cwd helpers).
- IO: extended TS `IoHandle` support to allow stream-backed handles (child process pipes) in `io_read`/`io_write`/`io_flush`/`io_close`.
- Go/TS hosts: implemented process spawning + wait/kill plumbing with cached status, plus term TTY hooks (Go raw mode/size limited to linux/amd64).
- Tests: not run.

# 2026-01-11 — Go extern host plugin exports (v11)
- Go interpreter: capture extern package name at definition time so host calls resolve after module evaluation; generate exported plugin wrappers for extern functions and bump extern cache version.
- Tests: `cd v11/interpreters/go && go test ./pkg/interpreter -run ExecFixtures`.
- Full sweep: `./run_all_tests.sh --version=v11`.

# 2026-01-10 — Extern host execution (v11)
- TS: added per-package host module caching with `host_error`/handle aliases and routed externs through dynamic imports; empty non-kernel externs now raise runtime errors.
- Go: added per-package plugin generation + caching, with extern calls marshaled through host functions; introduced `HostHandle` runtime values for IoHandle/ProcHandle.
- Fixtures/tests: updated extern exec fixture expectation and extern unit tests.
- Tests: `cd v11/interpreters/ts && bun test test/basics/externs.test.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run Extern`.

# 2026-01-10 — String interpolation continuation (v11)
- TS interpreter: added continuation state for string interpolation so time-slice yields resume mid-interpolation without re-running completed parts (fixes exec string helpers under scheduler yields).
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_12_01_stdlib_string_helpers bun run scripts/run-fixtures.ts`.
- Follow-up: ran `./run_all_tests.sh --version=v11`; all suites green and parity report refreshed at `v11/tmp/parity-report.json`.
- TS interpreter: switched the `read_text` host extern stub to async `fs.promises.readFile` so blocking IO can suspend only the calling proc via the native Promise bridge.

# 2026-01-02 — Placeholder member access in typechecker (v11)
- Go typechecker: treat member access on placeholder expressions as unknown instead of rejecting the access, fixing `@.method()` in pipe shorthand.
- Tests: `bun test test/parity/examples_parity.test.ts -t pipes_topics/main.able`; `./run_all_tests.sh --version=v11`.

# 2026-01-02 — String host accepts struct-backed String (v11)
- Go interpreter: `__able_String_from_builtin` now accepts struct-backed `String` values by extracting the `bytes` field, fixing stdlib string helpers when `String` is a struct instance.
- Go interpreter tests: added coverage for struct-backed `String` conversions in the string host builtins.
- Tests: `go test ./pkg/interpreter -run StringFromBuiltin`; `./v11/ablego test v11/stdlib/tests/text/string_methods.test.able`.

# 2026-01-02 — Parse empty function types + numeric method resolution (v11)
- Go parser: treat empty parenthesized types as zero-arg function parameters when parsing `() -> T`, preventing `()->T` from collapsing into a simple type.
- Go typechecker: allow method-set lookup for integer/float values so numeric helpers like `to_r` resolve after importing `able.core.numeric`.
- Tests: `./v11/ablego test v11/stdlib/tests/core/numeric_smoke.test.able`; `go test ./pkg/parser ./pkg/typechecker`.

# 2026-01-02 — Type matching fixes + Clone primitives (v11)
- Go interpreter: expanded runtime type matching to compare alias-expanded value types, let unknown type names match struct instances (generic union members like `T`), and accept struct-backed `String` values; added Result/Option generic matching fallback.
- Go interpreter: primitives now satisfy `Clone` via built-in method lookup, fixing stdlib TreeSet String constraints without extra imports.
- Tests: `ABLE_TYPECHECK_FIXTURES=off ./v11/ablego test v11/stdlib/tests --list`.

# 2026-01-02 — Method resolution disambiguation (v11)
- Go interpreter: tightened UFCS/member lookup to filter candidates by in-scope overloads and disambiguate same-name types via method target matching, fixing local `Channel.send` collisions while preserving alias reexports and Ratio methods.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`.

# 2026-01-02 — Serial executor fairness ordering (v11)
- Go interpreter: queue freshly created serial tasks ahead of resumed ones so `proc_yield` round-robins even when procs are created in separate eval calls.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter`; `./run_all_tests.sh --version=v11`.

# 2026-01-01 — Trailing lambdas + struct definition bindings (v11)
- Reattached trailing lambda bodies inside expression lists in the Go parser so call arguments match block parsing and `suite.it(...) { ... }` forms register correctly.
- Go interpreter now treats native/bound/partial callables as valid function values for `fn(...)` type checks.
- Go runtime records struct definitions separately from value bindings (including imports) so struct literals resolve even when constructors shadow names.
- Tests: `./run_all_tests.sh --version=v11`; `./v11/ablego test /home/david/sync/projects/able/v11/stdlib/tests/simple.test.able`; `./v11/ablets test /home/david/sync/projects/able/v11/stdlib/tests/simple.test.able`.

# 2026-01-01 — Regex literal matching (v11)
- Implemented literal-only regex compile/match/find_all using byte-span search and stored literal bytes in the regex handle.
- Updated the stdlib matcher test to exercise substring matches and refreshed docs to reflect partial regex support.
- Tests: `cd v11/interpreters/ts && ABLE_TYPECHECK_FIXTURES=off bun run scripts/run-module.ts test ../../stdlib/tests`.

# 2025-12-31 — Stdlib/runtime test fixes (v11)
- Resolved Iterable/Enumerable method ambiguity by preferring explicit impl methods over default interface methods (fixes Vector.each).
- Aligned hasher host builtins to return i64 handles/hashes (TS runtime + stubs + typechecker) and updated stdlib string integration expectations for subString errors.
- Tests: `./run_all_tests.sh --version=v11`.

# 2025-12-31 — Add idiomatic Able style guide (v11)
- Added a new documentation guide covering idiomatic Able conventions with examples at `v11/docs/idiomatic-able.md`.
- Cleared completed PLAN TODO entries now that the queue is empty.
- Tests: not run (docs-only changes).

# 2025-12-31 — Fix stdlib spec test discovery filters (v11)
- Corrected `able.spec` discovery filtering so empty exclude lists no longer filter every example; regenerated `v11/stdlib/src/spec.able`.
- Tests: `cd v11/interpreters/ts && ABLE_TYPECHECK_FIXTURES=off bun run scripts/run-module.ts test ../../stdlib/tests --list`.

# 2025-12-30 — Stdlib test syntax cleanup (v11)
- Rewrote stdlib tests to avoid non-spec syntax: replaced `let`/`let mut` with `:=`, switched prefix `match` to postfix form, wrapped `raise` in match/lambda bodies, removed `TestEvent.` qualifiers, and used struct-literal constructors for NFAChar.
- Updated static calls in tests from `Type::method` to `Type.method` and replaced `_` lambda parameters with `_ctx`.
- Tests: `bun run scripts/run-module.ts test ../../stdlib/tests` (fails: `able.text.automata` package missing from stdlib src; still in quarantine).

# 2025-12-30 — Verbose anonymous function syntax (v11)
- Added tree-sitter support for `fn(...) { ... }` anonymous functions and mapped them to `LambdaExpression` with generics/where clauses in TS/Go parsers.
- Added exec fixture `exec/07_02_01_verbose_anonymous_fn` plus conformance/coverage index updates.
- Tests: `bun test test/parser/fixtures_parser.test.ts`; `ABLE_FIXTURE_FILTER=07_02_01_verbose_anonymous_fn bun run scripts/run-fixtures.ts`; `go test -a ./pkg/interpreter -run TestExecFixtures/07_02_01_verbose_anonymous_fn$`.

# 2025-12-30 — Go parser interface-arg test fix (v11)
- Fixed Go parser test coverage so `TestParseImplInterfaceArgsParentheses` is a real test (no longer embedded in a raw string for the breakpoint fixture).
- Regenerated the tree-sitter parser/wasm and forced a Go rebuild so the cgo parser picks up the updated grammar.
- Tests: `bun test test/parser/fixtures_parser.test.ts`; `go test -a ./pkg/parser -run TestParseImplInterfaceArgsParentheses`; `./run_all_tests.sh --version=v11`.

# 2025-12-30 — Interface arg parentheses (Option 2b)
- Updated `spec/full_spec_v11.md` to clarify that interface args are space-delimited type expressions and generic applications only form when parenthesized.
- Updated tree-sitter grammar + TS/Go type parsers for `interface_type_*` nodes, and removed interface-arg arity grouping in TS/Go interpreter + typechecker.
- Added TS/Go parser tests asserting interface-arg splitting vs parenthesized generic args; rebuilt tree-sitter parser/wasm.
- Docs: noted parenthesized generic applications in `v11/parser/README.md`.
- Tests: not run (parser build via `npm run build`).

# 2025-12-30 — Go `able test` CLI wiring + integration tests (v11)
- Implemented Go `able test` CLI (args/targets, discovery, harness/run, reporters, exit codes) plus loader support for include packages and interpreter helpers for array/method access.
- Added lightweight Go CLI tests for `able test` list/dry-run using a stub stdlib harness to keep the suite fast and isolated.
- Tests: `cd v11/interpreters/go && go test ./cmd/able -run TestTestCommand`.

# 2025-12-30 — TypeScript `able test` CLI wiring (v11)
- Completed TS CLI `able test` wiring end-to-end (discovery/filter/run/report/exit), including kernel `Array` decoding for reporters/list output and array-length handling for discovery results.
- Bound trailing lambdas inside expression lists so `suite.it("...") { ... }` attaches correctly within `describe` bodies.
- Updated stdlib tests to `import able.spec.*` so suite method names are in scope; fixed a malformed `able.spec.assertions` import.
- Shifted testing protocol `Framework` methods to return `?Failure` and regenerated `v11/stdlib/src/spec.able`.
- Tests: `cd v11/interpreters/ts && bun test test/cli/run_module_cli.test.ts`.

# 2025-12-30 — Drop snapshots from testing framework (v11)
- Removed snapshot matchers, stores, and CLI flag references from the stdlib testing DSL; deleted the snapshot store design doc and scrubbed snapshot mentions from testing docs/design notes/spec.
- Updated stdlib tests and the TS CLI skeleton tests to drop `match_snapshot`/`--update-snapshots`.
- Tests: not run.

# 2025-12-30 — Testing module suffix policy (v11)
- Codified test modules as `.test.able` or `.spec.able` in `spec/full_spec_v11.md` (new Tooling: Testing Framework section) and resolved the open decision in `v11/design/testing-plan.md`.
- Removed the suffix-policy TODO from `PLAN.md` now that the rule is set.
- Tests: not run.

# 2025-12-30 — Split stdlib testing into able.test + able.spec (v11)
- Moved testing protocol/harness/reporters/snapshots into `v11/stdlib/src/test` and the user DSL into `v11/stdlib/src/spec`/`v11/stdlib/src/spec.able`, renaming packages and framework id to `able.spec`.
- Migrated quarantine stdlib tests into `v11/stdlib/tests` and updated imports to `able.spec` + `able.test.*`.
- Updated testing docs/spec references to the new namespace split and renamed internal `RspecFramework` to `SpecFramework`.
- Tests: not run (stdlib/CLI integration pending).

# 2025-12-29 — Alias re-export follow-up validation (v11)
- Progress: ran `./run_all_tests.sh --version=v11`; all green.
- State: parity report refreshed at `v11/tmp/parity-report.json`.
- Next: resume PLAN backlog (regex stdlib expansion + tutorial cleanup).

# 2025-12-29 — Typechecker method-set dedupe + alias impl diag alignment (v11)
- Prevented duplicate method-set candidates by tagging method-set function infos, skipping them during member lookup, and marking static method-set entries as type-qualified.
- Kept method-set targets in generic form to preserve where-clause constraint enforcement, and deduped method sets/impl records across session preludes.
- Aligned alias re-export impl ambiguity diagnostics to the canonical target label and updated the baseline accordingly.
- Tests: `./run_all_tests.sh --version=v11`.

# 2025-12-29 — Alias re-export impl ambiguity fixture (v11)
- Added AST error fixture `errors/alias_reexport_impl_ambiguity` covering duplicate impl registration when alias-attached impls target the same canonical type.
- Filled `module.json` for alias re-export method/impl ambiguity fixtures and updated the typecheck baseline for the new diagnostic.
- Tests: `cd v11/interpreters/ts && bun test test/parser/fixtures_mapper.test.ts`; `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts`.

# 2025-12-29 — Document Euclidean division (v11)
- Clarified in `spec/full_spec_v11.md` that `//`, `%`, and `/%` use Euclidean integer division (non-negative remainder), with examples and negative divisor behavior.

# 2025-12-29 — Add floor division helpers (v11)
- Added stdlib `div_floor`/`mod_floor`/`div_mod_floor` functions and methods for `i32`, `i64`, `u32`, `u64`, plus updated numeric helper fixture coverage and spec text.
- Removed the integer floor-division TODO from `spec/TODO_v11.md` now that helpers are implemented.

# 2025-12-29 — Spec TODO audit (v11)
- Trimmed `spec/TODO_v11.md` to the remaining items with detailed scope and open questions (alias/re-export method propagation).

# 2025-12-29 — Quarantine host regex hooks (v11)
- Removed TS/Go regex host hooks and wiring; stdlib `able.text.regex` now raises `RegexUnsupportedFeature` instead of calling host engines.
- Quarantined the exec regex fixture (`exec/14_02_regex_core_match_streaming`) and marked coverage as planned; TS regex integration test now skipped pending the stdlib engine.
- Updated conformance/testing/manual docs and regex design notes to reflect stdlib-only regex plans.
- Tests: `./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-29 — v11 fixture sweep
- Progress: ran `./run_all_tests.sh --version=v11 --fixture`.
- Current state: TypeScript fixtures, parity harness, and Go tests all green; parity report at `v11/tmp/parity-report.json`.
- Next: continue PLAN backlog (regex stdlib expansion and tutorial cleanup).

# 2025-12-29 — Regex core match/streaming fixture (v11)
- Implemented stdlib regex core helpers (`Regex.compile`, `is_match`, `match`, `find_all`, `scan` + streaming `RegexScanner.feed/next`) with host-backed compile/find externs.
- Added TS/Go regex host builtins and runtime state for compiled handles, plus span/match struct construction (empty groups/named groups for now).
- Added exec fixture `exec/14_02_regex_core_match_streaming`, updated conformance plan + coverage index, and removed the PLAN backlog item.
- Updated regex stdlib integration test and testing matcher docs now that regex helpers are live.
- Go runtime now treats `IteratorEnd {}` as matching the IteratorEnd sentinel during return type checks, aligning iterator method returns with pattern matching behavior.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=14_02_regex_core_match_streaming bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/14_02_regex_core_match_streaming$`.

# 2025-12-29 — Interface dispatch fixture + IteratorEnd runtime alignment (v11)
- Added exec fixture `exec/14_01_language_interfaces_index_apply_iterable` covering Index/IndexMut, Iterable/Iterator, and Apply dispatch; updated the conformance plan + coverage index and cleared the PLAN backlog item.
- Stdlib `iteration.each` now matches `IteratorEnd {}` before visiting values, keeping IteratorEnd from flowing into visitor callbacks.
- Go runtime now treats `IteratorEnd` as a first-class type in `matchesType`/type inference and equality comparisons, aligning match behavior with TS.
- Added exec fixture `exec/14_01_operator_interfaces_arithmetic_comparison` covering arithmetic/comparison operator interface dispatch with Display/Clone/Default helpers; updated the conformance plan + coverage index and cleared the PLAN backlog item.
- TS/Go runtimes now route unary `-` through `Neg` interface impls when operands are non-numeric, and both runtimes dispatch comparison operators via `Eq`/`PartialEq` and `Ord`/`PartialOrd` when available.
- Go static method lookup now includes impl methods so interface-provided statics like `Default.default()` resolve on types.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_06_stdlib_package_resolution bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_07_search_path_env_override bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=14_01_language_interfaces_index_apply_iterable bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/13_06_stdlib_package_resolution$`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/13_07_search_path_env_override$`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/14_01_language_interfaces_index_apply_iterable$`.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=14_01_operator_interfaces_arithmetic_comparison bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/14_01_operator_interfaces_arithmetic_comparison$`.
- Added stdlib `os` module with `args()` and runtime `__able_os_args` builtins for TS/Go; added exec fixture `exec/15_02_entry_args_signature` plus coverage/conformance updates and removed the PLAN item.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=15_02_entry_args_signature bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/15_02_entry_args_signature$`.
- Added `os.exit` runtime support (TS/Go) with CLI/fixture harness handling; added exec fixture `exec/15_03_exit_status_return_value` plus coverage/conformance updates and removed the PLAN item.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=15_03_exit_status_return_value bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/15_03_exit_status_return_value$`.
- Added exec fixture `exec/15_04_background_work_flush` to assert background tasks are not awaited on exit; updated coverage/conformance and removed the PLAN item.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=15_04_background_work_flush bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/15_04_background_work_flush$`.
- Added exec fixture `exec/16_01_host_interop_inline_extern` covering extern host bindings; updated coverage/conformance and removed the PLAN item.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=16_01_host_interop_inline_extern bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/16_01_host_interop_inline_extern$`.

# 2025-12-28 — Exec fixtures for errors + concurrency (v11)
- Added `exec/11_03_rescue_rethrow_standard_errors` covering arithmetic/indexing runtime errors, rescue/ensure, and rethrow semantics; updated the exec coverage index and removed the PLAN backlog item.
- Added `exec/12_02_future_fairness_cancellation` covering `future_yield` fairness, cancellation via `future_cancelled`, and `future_flush` queue drains; updated the exec coverage index + conformance plan and removed the PLAN backlog item.
- Added `exec/12_03_spawn_future_status_error` and `exec/12_04_future_handle_value_view` for future status/value/error propagation and handle/value behaviour; updated the exec coverage index + conformance plan and cleared the PLAN items.
- Added `exec/12_05_mutex_lock_unlock` and `exec/12_06_await_fairness_cancellation` for mutex/await semantics; updated the exec coverage index + conformance plan and cleared the PLAN items.
- TS/Go runtimes now raise standard errors (`DivisionByZeroError`, `OverflowError`, `ShiftOutOfRangeError`, `IndexError`) with `Error.value` payloads for rescue matching; `!` propagation now raises any `Error` value and index fallback returns `IndexError` payloads.
- Tests: `cd v11/interpreters/ts && bun run scripts/export-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=11_03_rescue_rethrow_standard_errors bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_02_future_fairness_cancellation bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_03_spawn_future_status_error bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_04_future_handle_value_view bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_05_mutex_lock_unlock bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_06_await_fairness_cancellation bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter`.
- Added `exec/12_07_channel_mutex_error_types` and `exec/13_03_package_config_prelude` for channel/mutex error payloads and package.yml root-name/prelude parsing; updated the exec coverage index + conformance plan and cleared the PLAN items.
- Tests: `cd v11/interpreters/ts && bun run scripts/export-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=12_07_channel_mutex_error_types bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_03_package_config_prelude bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtures/(12_07_channel_mutex_error_types|13_03_package_config_prelude)$'`.
- Added `exec/13_04_import_alias_selective_dynimport` covering import aliases, selective renames, and dynimport bindings; updated the exec coverage index + conformance plan and removed the PLAN backlog item.
- TS runtime now treats primitive types as satisfying `Hash`/`Eq` constraints for interface enforcement and returns `IndexError` values on out-of-bounds array assignments to align IndexMut semantics.
- Updated TS division/ratio tests to assert `RaiseSignal` error payloads rather than raw error messages.
- Tests: `./run_all_tests.sh --version=v11`.

# 2025-12-26 — Impl specificity exec fixture + array type-arg dispatch (v11)
- Added `exec/10_02_impl_specificity_named_overrides` covering impl specificity ordering, named impl disambiguation, and HKT targets; updated the exec coverage index, conformance plan, and removed the PLAN backlog item.
- TS runtime now derives array element type arguments during method resolution so concrete vs generic impl selection matches spec intent.
- Tests: `cd v11/interpreters/ts && bun run scripts/export-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_02_impl_specificity_named_overrides bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && GOCACHE=/home/david/sync/projects/able/.tmp/go-build GOMODCACHE=/home/david/sync/projects/able/.tmp/gomod go test ./pkg/interpreter -run TestExecFixtures/10_02_impl_specificity_named_overrides$`.

# 2025-12-24 — Truthiness exec fixture + boolean context alignment (v11)
- Added `exec/06_11_truthiness_boolean_context` to cover truthiness rules, unary `!`, and `&&`/`||` operand returns; updated coverage index + conformance plan and cleared the PLAN item.
- TS/Go runtimes now evaluate `!`, `&&`, and `||` via truthiness (returning operands) and dynimport supports late-bound packages without eager loader failures; dyn refs now re-check privacy at call time.
- Typecheckers no longer require bool for conditions/guards or logical operands; Go typechecker/interpreter tests updated to match truthiness semantics.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_11_truthiness_boolean_context bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/06_11_truthiness_boolean_context$`; `cd v11/interpreters/go && go test ./pkg/typechecker -run Truthiness`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestLogicalOperandsTruthiness$`; `cd v11/interpreters/go && go test ./pkg/driver -run TestLoaderDynImportDependencies$`.

# 2025-12-24 — Dynamic metaprogramming exec fixture + runtime support (v11)
- Added `exec/06_10_dynamic_metaprogramming_package_object` to cover dyn package creation/lookup, dynamic definitions, and late-bound dynimport redefinitions; updated coverage index + conformance plan and cleared the PLAN item.
- Implemented dyn runtime helpers in TS/Go: `dyn.package`, `dyn.def_package`, and `dyn.Package.def` parse/evaluate dynamic code and replace prior definitions without overload merging.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_10_dynamic_metaprogramming_package_object bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/06_10_dynamic_metaprogramming_package_object$`.

# 2025-12-24 — Lexical line-join + trailing commas exec fixture (v11)
- Added `exec/06_09_lexical_trailing_commas_line_join` to cover delimiter line-joining and trailing commas in arrays/structs/imports; updated conformance plan and exec coverage index.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_09_lexical_trailing_commas_line_join bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/06_09_lexical_trailing_commas_line_join$`.

# 2025-12-24 — Array ops exec fixture + IndexMut error surfacing (v11)
- Added `exec/06_08_array_ops_mutability` to cover array mutation, bounds handling, and iteration, plus updated the conformance plan and coverage index.
- Index assignment now returns IndexError values from IndexMut implementations instead of silently discarding them (TS + Go interpreters).
- Tests: `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run TestExecFixtures/06_08_array_ops_mutability$`.

# 2025-12-24 — Exec fixtures for structs, unions, methods, interfaces, and packages (v11)
- Added exec fixtures for numeric literal contextual typing (plus overflow diag), positional structs, nullable truthiness, Option/Result construction, union guarded match coverage (plus a non-exhaustive diag), union payload patterns, method imports/UFCS instance-vs-static, interface dynamic dispatch, and directory-based package structure.
- Interpreters now coerce numeric values to float parameter contexts at runtime; TS/Go typecheckers accept integer literals in float contexts per spec.
- Updated `v11/fixtures/exec/coverage-index.json`, `v11/docs/conformance-plan.md`, and pruned completed PLAN items.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_05_03_struct_positional_named_tuple bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_06_01_union_payload_patterns bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_06_02_nullable_truthiness bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_06_03_union_construction_result_option bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_06_04_union_guarded_match_exhaustive bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_01_literals_numeric_contextual bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=09_02_methods_instance_vs_static bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=10_03_interface_type_dynamic_dispatch bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=13_01_package_structure_modules bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtures/(04_05_03_struct_positional_named_tuple|04_06_01_union_payload_patterns|04_06_02_nullable_truthiness|04_06_03_union_construction_result_option|04_06_04_union_guarded_match_exhaustive(_diag)?|06_01_literals_numeric_contextual(_diag)?|09_02_methods_instance_vs_static|10_03_interface_type_dynamic_dispatch|13_01_package_structure_modules)$'`.

# 2025-12-23 — Core exec fixtures + literal escape parsing (v11)
- Added exec fixtures for struct named updates (plus diagnostic), string/char literal escapes, control-flow expression values, and lambda closures with explicit return; updated exec coverage index + conformance plan and removed completed PLAN items.
- TS/Go parsers now unescape string/char literals with spec escapes (including `\'` and `\u{...}`) to align literal parsing across runtimes.
- Tests: `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=04_05_02_struct_named_update_mutation bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_01_literals_string_char_escape bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=06_05_control_flow_expr_value bun run scripts/run-fixtures.ts`; `cd v11/interpreters/ts && ABLE_FIXTURE_FILTER=07_02_lambdas_closures_capture bun run scripts/run-fixtures.ts`; `cd v11/interpreters/go && go test ./pkg/interpreter -run 'TestExecFixtures/(04_05_02_struct_named_update_mutation|04_05_02_struct_named_update_mutation_diag|06_01_literals_string_char_escape|06_05_control_flow_expr_value|07_02_lambdas_closures_capture)$'`.

# 2025-12-22 — Division-by-zero exec fixture (v11)
- Added `exec/04_02_primitives_truthiness_numeric_diag` to assert division-by-zero errors, plus inline semantics comment in the package-visibility fixture module.
- Normalized TS numeric division errors to use lowercase `division by zero` for parity with the stdlib error message and Go runtime.
- Updated exec coverage index + conformance plan to include the new diagnostic fixture.
- Tests: `./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-22 — Kernel alias normalization for typed patterns (v11)
- Normalized runtime type matching to map KernelChannel/KernelMutex/KernelRange/KernelRangeFactory/KernelRatio/KernelAwaitable/AwaitWaker/AwaitRegistration to their stdlib names so typed patterns match kernel aliases.
- Tests: `GOCACHE=/home/david/sync/projects/able/.tmp/go-build GOMODCACHE=/home/david/sync/projects/able/.tmp/gomod go test ./pkg/interpreter -run TestStdlibChannelMutexModuleLoader`; `GOCACHE=/home/david/sync/projects/able/.tmp/go-build GOMODCACHE=/home/david/sync/projects/able/.tmp/gomod ./run_all_tests.sh --version=v11`.

# 2025-12-22 — Singleton struct exec fixture (v11)
- Added `exec/04_05_01_struct_singleton_usage` covering singleton struct tags and pattern matching; updated exec coverage index and PLAN backlog.
- Treated singleton struct definitions as runtime values in TS/Go (typed pattern checks + type-name reporting) and matched singleton identifiers as constant patterns; Go `valuesEqual` now handles struct definition pointers.
- Tests: `GOCACHE=/home/david/sync/projects/able/.tmp/go-build GOMODCACHE=/home/david/sync/projects/able/.tmp/gomod ./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-20 — Reserved underscore type aliases (v11)
- Added `exec/04_04_reserved_underscore_types` for `_` placeholder type expressions and updated the exec coverage index + PLAN backlog.
- Added AST fixture `errors/type_alias_underscore_reserved` and enforced runtime/typechecker rejection of type alias name `_` in TS + Go.
- Refreshed the typecheck baseline to include the new alias diagnostic.
- Tests: `./run_all_tests.sh --version=v11 --fixture`; `ABLE_TYPECHECK_FIXTURES=warn bun run scripts/run-fixtures.ts -- --write-typecheck-baseline`.

# 2025-12-20 — Go type matching + constraint parity fixes (v11)
- Normalized runtime type matching for kernel alias names (`KernelArray`, `KernelHashMap`) and treated generic type args as wildcards during concrete matches to align alias recursion fixtures.
- Treated primitives as satisfying `Hash`/`Eq` method presence during impl resolution and expanded intrinsic Hash/Eq coverage to include all integer/float types.
- Stopped Go typechecker from typechecking impl/method bodies to mirror TS diagnostics for interface conformance tests.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestFixtureParityStringLiteral/strings/String_methods -count=1`; `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtures/04_07_05_alias_recursion_termination -count=1`; `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestExecFixtures/06_01_literals_array_map_inference -count=1`; `GOCACHE=$(pwd)/.gocache go test ./pkg/interpreter -run TestFixtureParityStringLiteral/interfaces/apply_index_missing_impls -count=1`; `./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-19 — Type expression syntax exec fixture (v11)
- Added `exec/04_03_type_expression_syntax` to cover nested type expressions with generic application and unions, including inline semantics comments.
- Updated `v11/fixtures/exec/coverage-index.json` and removed the completed PLAN backlog item.
- Tests not run (not requested).

# 2025-12-19 — Primitives truthiness exec fixture (v11)
- Added `exec/04_02_primitives_truthiness_numeric` to cover literal forms, Euclidean `//`/`%`, and truthiness for `nil`/`false`/`Error`/`void`.
- Updated `v11/fixtures/exec/coverage-index.json` and removed the completed PLAN backlog item.
- Tests not run (not requested).

# 2025-12-19 — Exec fixture reimplementation pass (v11)
- Reimplemented exec fixtures for §2, §3, §4.1, §4.7.2–§4.7.5, §5.0–§5.3, §6.1.7–§6.1.9, §6.2, and §6.3.1 with fresh programs aligned to spec semantics; refreshed manifests/expectations accordingly.
- Updated `PLAN.md` note to emphasize spec-first fixtures even when interpreters fail.
- Tests not run (exec fixtures pending re-sweep).

# 2025-12-19 — Exec fixture expansion (v11)
- Seeded exec fixtures for §2 (lexical comments/identifiers/trailing commas), §3 (block expression separation/value), and §4.1.4–§4.1.6 (generic inference with constrained interface methods), each with inline semantics notes.
- Coverage index and PLAN backlog updated to reflect the new fixtures; exec coverage guard stays green. Go runtime updated to treat generic-parameter types as wildcards during overload selection so the new §4.1 fixture runs.
- Tests: `./run_all_tests.sh --version=v11 --fixture`.

# 2025-12-18 — Exec fixtures renamed to spec sections (v11)
- Renamed exec fixtures/packages to use spec section prefixes, refreshed manifests/comments, and regenerated the coverage index + docs with the spec-based naming scheme.
- Added/updated exec fixtures for async proc/spawn scheduling, methods/UFCS, option/result handling via raise/rescue, package visibility with explicit aliases, and alias/union generic combos; each case documents the exercised semantics inline.
- Conformance docs now carry the seeded coverage matrix and exec fixtures guide keyed to the new IDs, with the JSON coverage index tracked alongside.
- `run_all_tests.sh` runs `scripts/check-exec-coverage.mjs` before the suites; `./run_all_tests.sh --version=v11` passes after the exec fixture sweep.
- Seeded new exec fixtures for §4.7.2–§4.7.5 (generic alias substitution, import visibility, alias-backed methods/impls, and recursive alias termination via nominal indirection) with manifests and package wiring.
- Coverage index and PLAN backlog updated to reflect the new alias fixtures.
- Added exec fixtures for §5.0–§5.2 (mutability declaration vs assignment, identifier/wildcard typed patterns, struct pattern rename with typed nesting) and brought coverage index + PLAN in sync.

# 2025-12-13 — Host interop tutorial unblocked + prefix match guard (v11)
- Added the Go-side `read_text` extern for tutorial 14 (imports `os`, returns `host_error` on failure) so the host interop example now runs on the Go runtime alongside the existing TS path.
- Tree-sitter grammar usage is now enforced to reject prefix-style `match <expr> { ... }`: added Go parser and TS tree-sitter/mapper tests that expect syntax errors for the legacy ordering, matching the spec requirement of `<expr> match { ... }`.
- Tests: `cd v11/interpreters/go && go test ./pkg/parser -run PrefixMatch`; `cd v11/interpreters/ts && bun test test/parser/fixtures_parser.test.ts`.

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

### 2025-12-24
- Added exec fixture `exec/06_03_operator_overloading_interfaces` to cover Add/Index/IndexMut operator dispatch and updated the conformance plan + coverage index.
- Go interpreter now dispatches arithmetic/bitwise operators to interface implementations when operands are non-numeric.
- Range expressions now enforce integer bounds in both runtimes/typecheckers; updated range diagnostics/tests to match the v11 spec.
- Cleared the operator-overloading exec fixture item from the v11 PLAN backlog.
- Added composition exec fixtures for combined behavior: `exec/09_00_methods_generics_imports_combo` (imports + generics + methods) and `exec/11_00_errors_match_loop_combo` (match + loop + rescue), with coverage index + conformance plan updates.
- Added exec fixture `exec/06_03_safe_navigation_nil_short_circuit` to cover `?.` short-circuiting, receiver evaluation, and argument skipping, and updated coverage + conformance docs.
- Added exec fixture `exec/06_04_function_call_eval_order_trailing_lambda` to cover left-to-right call argument evaluation and trailing lambda equivalence, with coverage + conformance updates.
- Added exec fixture `exec/06_06_string_interpolation` to cover interpolation escapes and multiline string literals, with coverage + conformance updates.
- Added exec fixture `exec/06_07_generator_yield_iterator_end` to cover yield/stop semantics and IteratorEnd exhaustion behavior, with coverage + conformance updates.
- Tree-sitter grammar now allows multiline double-quoted strings; TS/Go parsers unescape interpolation text for `\\$`/`\\`` (and `\\\\`) so backtick escapes follow the v11 spec.
- Pattern matching now treats `IteratorEnd {}` as a match for the iterator end sentinel in both interpreters.
- Added exec fixture `exec/06_12_01_stdlib_string_helpers` covering required string helper semantics (lengths, substring bounds, split/replace, prefix/suffix) and updated coverage/conformance tracking; cleared the PLAN backlog item.
- Added exec fixture `exec/06_12_02_stdlib_array_helpers` for the required array helper API (size, push/pop, get/set, clear) with coverage/conformance updates; cleared the PLAN backlog item.
- Added exec fixture `exec/06_12_03_stdlib_numeric_ratio_divmod` covering Ratio normalization/to_r and Euclidean /% results with coverage/conformance updates; cleared the PLAN backlog item.
- Added `as` cast expressions to the grammar + AST contract and implemented explicit numeric/interface casts in both interpreters and typecheckers.
- Stdlib numeric cleanup: replaced unsupported `const`/`mut`/`else if`, normalized i128 constants, removed duplicate Ratio numerator/denominator methods in favor of kernel definitions, and added statement terminators where the parser requires them.
- Inherent methods now resolve without requiring the method name in the caller scope (TS + Go) so stdlib extensions work through package imports; refreshed the numeric ratio/divmod exec fixture import to use stdlib `Ratio`.

### 2025-12-26
- Added exec fixtures `exec/07_01_function_definition_generics_inference` (implicit/explicit generics + return inference), `exec/07_03_explicit_return_flow` (explicit return flow), `exec/07_04_trailing_lambda_method_syntax` (method call syntax + trailing lambda parity), `exec/07_04_apply_callable_interface` (Apply callables), `exec/07_05_partial_application` (placeholder partial application), and `exec/07_06_shorthand_member_placeholder_lambdas` (implicit member/method shorthand + placeholder lambdas); updated the conformance plan + coverage index and cleared the related PLAN backlog items.
- TS parser/typechecker now accept implicit member assignments as valid assignment targets, matching the runtime semantics.
- Tests: `ABLE_FIXTURE_FILTER=07_01_function_definition_generics_inference bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_01_function_definition_generics_inference$`; `ABLE_FIXTURE_FILTER=07_03_explicit_return_flow bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_03_explicit_return_flow$`; `ABLE_FIXTURE_FILTER=07_04_trailing_lambda_method_syntax bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_04_trailing_lambda_method_syntax$`; `ABLE_FIXTURE_FILTER=07_04_apply_callable_interface bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_04_apply_callable_interface$`; `ABLE_FIXTURE_FILTER=07_05_partial_application bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_05_partial_application$`; `ABLE_FIXTURE_FILTER=07_06_shorthand_member_placeholder_lambdas bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestExecFixtures/07_06_shorthand_member_placeholder_lambdas$`.
- Method resolution now honors name/type-in-scope gating without breaking kernel/primitive method access; fixed UFCS name lookup and aligned TS/Go member resolution with type-name visibility.
- Updated TS typechecker tests to match truthiness semantics (if/while/match/rescue guards) and swapped the diagnostic location test to an undefined identifier.
- Refreshed fixtures for truthiness (`errors/logic_operand_type`) and string/numeric exec imports; updated the AST typecheck baseline for the logic operand fixture.
- Tests: `./run_all_tests.sh --version=v11` (TS + Go units, fixtures, parity) with parity report in `v11/tmp/parity-report.json`.

### 2025-12-27
- Added exec fixtures `exec/07_07_overload_resolution_runtime`, `exec/08_01_if_truthiness_value`, and `exec/08_01_match_guards_exhaustiveness`; updated the conformance plan + coverage index and cleared the related PLAN backlog items.
- `if` expressions now return `nil` (not `void`) when no branch matches and there is no else, aligning TS/Go runtimes with the v11 spec; updated `errors/rescue_guard` fixture expectation.
- Tests: `bun run scripts/export-fixtures.ts`; `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`); `./run_all_tests.sh --version=v11`.
- Next: continue exec fixture backlog starting at `exec/08_02_while_continue_break` and `exec/08_02_loop_expression_break_value`.

### 2025-12-28
- Added exec fixtures `exec/08_02_while_continue_break`, `exec/08_02_loop_expression_break_value`, and `exec/08_02_range_inclusive_exclusive`; updated the conformance plan + coverage index and cleared the related PLAN backlog items.
- Updated the generated AST fixture expectation for `errors/rescue_guard` in the TS export fixture source to keep it aligned with nil-returning `if` expressions.
- Tests: `bun run scripts/export-fixtures.ts`; `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`).
- Next: continue exec fixture backlog starting at `exec/08_03_breakpoint_nonlocal_jump`.
- Added exec fixtures `exec/13_06_stdlib_package_resolution` and `exec/13_07_search_path_env_override`; updated the conformance plan + coverage index and cleared the related PLAN backlog items.
- Exec fixture runners now honor manifest-provided env overrides for module search paths; TS uses CLI-style search path resolution and Go exec fixtures mirror the env-driven roots.

### 2025-12-29
- Added exec fixture `exec/08_03_breakpoint_nonlocal_jump` and updated the conformance plan + coverage index; cleared the related PLAN backlog item.
- Tests: `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`).
- Next: continue exec fixture backlog starting at `exec/09_05_method_set_generics_where`.
- Codified alias/re-export method propagation and conflict semantics in the v11 spec and cleared the remaining spec TODO.
- Added AST fixture `errors/alias_reexport_method_ambiguity` (with setup packages) plus baseline entry; TS/Go typecheckers now surface ambiguous overloads when multiple method sets attach the same signature.
- Tests: `ABLE_FIXTURE_FILTER=alias_reexport_method_ambiguity bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter -run TestFixtureParityStringLiteral/errors/alias_reexport_method_ambiguity`.

### 2025-12-30
- Added exec fixture `exec/09_05_method_set_generics_where` covering method-set generics/where constraints for instance + UFCS calls; updated the conformance plan + coverage index and cleared the PLAN backlog item.
- Go interpreter now enforces method-set generic/where constraints during method calls; call helpers were split into `v11/interpreters/go/pkg/interpreter/call_helpers.go` to keep call logic files under 1000 lines.
- Aligned TS + Go parser handling of impl interface args so space-delimited arg lists are not collapsed into a single generic application; fixture mapper parity is green again.
- Spec: documented interface-arg grouping by interface arity (greedy left-to-right grouping with parentheses to force a single argument).
- TS/Go typechecker + runtime now group impl interface args by interface arity, so unparenthesized generic applications like `Map K V` remain a single argument when the interface expects one.
- Tests: `bun run scripts/export-fixtures.ts`; `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`).

### 2025-12-31
- Added exec fixture `exec/10_01_interface_defaults_composites` covering interface defaults, implicit vs explicit `Self`, and composite aliases; updated the conformance plan + coverage index and cleared the PLAN backlog item.
- TS/Go runtimes now treat composite interfaces as base-interface bundles for interface coercion and method dispatch (including default methods), with interface checks honoring base interfaces.
- Tests: `bun run scripts/export-fixtures.ts`; `bun run scripts/run-fixtures.ts`; `go test ./pkg/interpreter` (with a temp `GOCACHE`).
- Added exec fixture `exec/07_02_01_verbose_anonymous_fn` covering verbose anonymous functions (generics + where clauses).
- Stdlib fixes: added numeric interfaces to `core/numeric`, corrected rational i128 min/max constants, and added `Queue.is_empty` inherent method to avoid interface ambiguity.
- String stdlib: added `__able_char_to_codepoint` host builtin (TS/Go/kernel) and rewrote `char_to_utf8` to use it; string smoke tests now import `able.text.string`.
- Test cleanup: corrected array filter expectations, added semicolons to avoid `Array.new()`/`for` parse ambiguity, removed unsupported heap comparator test, and renamed lazy seq iterator to avoid duplicate impls.

### 2025-12-26
- Added exec fixture `exec/11_01_return_statement_type_enforcement` and updated the conformance plan + coverage index; cleared the related PLAN backlog item.
- TS/Go runtimes now enforce return type checks (including bare return for non-void), and the TS/Go typecheckers ignore unreachable tail expressions after return while rejecting bare returns in non-void functions.
- Updated `exec/11_00_errors_match_loop_combo` to use `//` so integer division stays within `i32` per the v11 spec.
- Added exec fixture `exec/11_02_option_result_or_handlers` (Option/Result `or {}` handlers) and updated the conformance plan + coverage index; cleared the related PLAN backlog item.
- Return type enforcement now treats iterator values as `Iterator` interface matches in TS/Go, and Go generic interface checks accept interface implementers; updated the pipeline diagnostics test to expect runtime mismatch under `AllowDiagnostics`.
- Tests: `./run_all_tests.sh --version=v11 --fixture`; `./run_all_tests.sh --version=v11`.

### 2026-01-10
- TS interpreter entrypoint tasks now preserve raw runtime errors, support re-entrant `proc_flush`, and prioritize generator continuations so iterator yields work inside proc contexts.
- CLI/fixture/parity runners now bind entrypoint `main` calls via a dedicated environment to avoid missing symbol errors; entrypoint-only async helpers still gate user-facing `proc_yield`/`proc_cancelled`.
- Tests: `bun test test/concurrency/native_suspend.test.ts`; `bun test test/cli/run_module_cli.test.ts`; `bun test test/parity/examples_parity.test.ts`; `bun test test/parity/fixtures_parity.test.ts`.

### 2026-01-11
- Added `dyn.Package.eval`/`dyn.eval` plumbing for REPL-style evaluation (parse errors return `ParseError` with `is_incomplete`), and mirrored dynamic-definition rebinding rules in Go imports/definitions.
- Added stdlib `able.repl` module (line editor, `:help`/`:quit`, prints non-`void` results) plus `able repl` CLI support for TS and Go.
- Spec: documented `ParseError`/`Span` plus dynamic eval APIs and REPL-oriented parse error semantics.
- Tests: `bun run scripts/run-fixtures.ts`; `./run_all_tests.sh --version=v11`.

### 2026-01-13
- Ran the full v11 sweep; parity report refreshed at `v11/tmp/parity-report.json`.
- Tests: `./run_all_tests.sh --version=v11`.
- Added `able.io.temp` for temp file/dir creation + cleanup helpers, and added `io.puts`/`io.gets` wrappers in `able.io`.
- Extended stdlib IO tests with temp helper coverage.
- Tests: `./v11/ablets test v11/stdlib/tests/io.test.able`; `./v11/ablego test v11/stdlib/tests/io.test.able`.
- Expanded Path tests for mixed separators and UNC roots, and fs tests for missing directory reads.
- Tests: `./v11/ablets test v11/stdlib/tests/fs.test.able`; `./v11/ablego test v11/stdlib/tests/fs.test.able`; `./v11/ablets test v11/stdlib/tests/path.test.able`; `./v11/ablego test v11/stdlib/tests/path.test.able`.
- Expanded stdlib IO/Path/fs edge coverage (non-positive reads, empty paths, remove missing, empty read_lines, differing roots).
- Tests: `./v11/ablets test v11/stdlib/tests/io.test.able`; `./v11/ablego test v11/stdlib/tests/io.test.able`; `./v11/ablets test v11/stdlib/tests/fs.test.able`; `./v11/ablego test v11/stdlib/tests/fs.test.able`; `./v11/ablets test v11/stdlib/tests/path.test.able`; `./v11/ablego test v11/stdlib/tests/path.test.able`.
- Added a PermissionDenied stdlib fs test and fixed Go extern singleton struct decoding so union kinds map correctly from host strings; tightened permission error detection in Go stdlib error mapping.
- Tests: `./v11/ablego test v11/stdlib/tests/fs.test.able`.
- Ran the full v11 sweep; parity report refreshed at `v11/tmp/parity-report.json`.
- Tests: `./run_all_tests.sh --version=v11`.
- Lowercased the `able.io.path` package name and updated stdlib/tests/docs imports and call sites to use `path.*`.
- Tests: `./run_all_tests.sh --version=v11`.
- Next: finish stdlib IO coverage (errors, path normalization, IO handle edge cases) and keep `./run_all_tests.sh --version=v11` green.

### 2026-01-14
- Preserved entrypoint runtime diagnostic context in the TS scheduler so raw runtime errors keep locations/stack notes.
- Standardized TS diagnostics path normalization to the repo root, added shared path helpers, and normalized fixture origins accordingly.
- Corrected infix/postfix expression spans so runtime diagnostics point at full expressions instead of suffix/right operands.
- Propagated return-statement context into return type mismatch errors and refreshed fixture expectations/baselines for new paths/notes.
- Tests: `bun run scripts/run-fixtures.ts`.
- Split Go interpreter member resolution into smaller modules (`interpreter_members.go`, `interpreter_method_resolution.go`) to keep files under 900 lines.
- Tests: `./run_all_tests.sh --version=v11` (passed); `./run_stdlib_tests.sh` (TS stdlib failed at `v11/stdlib/tests/fs.test.able:202`, Go stdlib passed).
- Fixed `fs.copy_dir` overwrite behavior by clearing destination contents after a removal attempt when needed.
- Tests: `./run_stdlib_tests.sh`; `./run_all_tests.sh --version=v11`.
- Dropped redundant where-clauses from `HashMap` impls so stdlib typechecking stays clean in strict mode.
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

### 2026-01-16
- Added common `HashSet` set operations (union/intersect/difference/symmetric difference, subset/superset/disjoint) plus new spec coverage.
- Tests: `./v11/ablets test v11/stdlib/tests/collections/hash_set.test.able`; `./v11/ablego test v11/stdlib/tests/collections/hash_set.test.able`.
- Spec: documented the always-loaded `able.kernel` contract (core interfaces, HashMap, KernelHasher, hash bridges) and clarified map literal key constraints plus hash container semantics.
- Spec: defined the `Hasher` interface and tied primitive Hash/Eq/Ord impls to the kernel library; aligned kernel string/char bridge names.
- Spec: enumerated kernel-resident types/interfaces/methods and listed the full `Hasher` helper surface with default semantics.
- TS interpreter: track struct definitions and treat concrete type names as taking precedence over interface names during runtime coercion/matching to fix `Range` vs `Range` interface collisions.
- Go typechecker: unwrap interface aliases (e.g., `Clone`, `Eq`, `Hash`) when collecting impls, validating impls, and solving constraints.
- Spec TODOs: cleared the kernel hashing contract items now captured in `spec/full_spec_v11.md`.
- Tests: `./run_all_tests.sh --version=v11`; `./run_stdlib_tests.sh --version=v11`.
- Go typechecker: allow impl targets to be interface types (supporting `impl Iterable T for Iterator T` matches).
- Go interpreter: treat missing generic args as wildcards when matching impl targets; record iterator values as `Iterator _` for runtime type info.
- Stdlib: fixed iterable helper signatures in `able.core.iteration` after adding iterator-as-iterable support.
- Tests: `./v11/ablets test v11/stdlib/tests/core/iteration.test.able`; `./v11/ablego test v11/stdlib/tests/core/iteration.test.able`.
- Go typechecker: allow interface default methods to satisfy member access when an impl omits the method body.
- Tests: `./v11/ablego test v11/stdlib/tests/core/iteration.test.able`.
- Stdlib: added `default<T: Default>()` helper in `able.core.interfaces`.
- Stdlib/spec: added `Extend` interface + `Iterable.collect` default, with Array/HashSet impls and iteration tests.
- Next: resume the PLAN work queue (regex parser + quantifiers).

### 2026-01-18
- Spec: clarified interface dynamic dispatch as dictionary-based (default methods + interface-impl method availability).
- TS typechecker: added type-parameter tracking in expressions, inference for interface-method generics, and base-interface method candidates.
- TS interpreter: interface values now carry method dictionaries (incl. iterator natives), interface-member binding handles native methods, and for-loops accept interface-wrapped iterators.
- Go typechecker: collect transitive impls/method sets for imports, preserve interface metadata on impls for default methods, and write inferred call type arguments into the AST.
- Stdlib: fixed `Iterable.map`/`filter_map` generic parameter syntax.
- Tests: `v11/ablets .examples/foo.able`; `v11/ablego .examples/foo.able`; `v11/ablets test v11/stdlib/tests/core/iteration.test.able`; `v11/ablego test v11/stdlib/tests/core/iteration.test.able`; `v11/ablets test v11/stdlib/tests/collections/hash_set.test.able`; `v11/ablego test v11/stdlib/tests/collections/hash_set.test.able`; `v11/ablets test v11/stdlib/tests/collections/hash_set_smoke.test.able`; `v11/ablego test v11/stdlib/tests/collections/hash_set_smoke.test.able`.

### 2026-01-19
- Spec: documented package-qualified member access as yielding first-class values (type aliases remain type-only).
- TS typechecker: package member access now resolves symbol types from summaries (function values included), enabling `pkg.fn` usage in expressions.
- Tests: `v11/ablets .examples/foo.able`; `v11/ablego .examples/foo.able`.
- Stdlib tests: added iteration coverage for `collect` via Default/Extend and package-qualified function values.
- Tests: `v11/ablets test v11/stdlib/tests/core/iteration.test.able`; `v11/ablego test v11/stdlib/tests/core/iteration.test.able`.
- Spec: documented the `Default` interface signature and its stdlib helper.
- Design: captured the eager vs lazy collections split (`Iterable` minimal, lazy adapters on `Iterator`, eager `Enumerable`).
- Spec: documented the `Enumerable` interface and lazy/eager split in the iteration section; updated core iteration protocol.
- Stdlib: made `Enumerable` parseable under current grammar (removed base-interface clause and HKT `where` constraints); documented `Iterable`’s "implement either each/iterator" intent; moved `Queue` operations to inherent methods.
- TS interpreter: added Error/Awaitable/Iterator interface-value handling for native values (default methods + await helpers), and allow generic interface values to satisfy `matchesType`.
- TS tests: aligned `Display` dispatch test with `to_string`.
- Tests: `bun test` in `v11/interpreters/ts`; `go test ./...` in `v11/interpreters/go`.
- Stdlib: added explicit `Enumerable.lazy` impl for `Array` to keep lazy iterators reachable under Go.
- Tests: `./v11/ablego test v11/stdlib/tests/enumerable.test.able --format tap`.

### 2026-01-20
- Go typechecker: instantiate generic unions when resolving type annotations and normalize applied union types for assignability.
- Parser/typechecker: where-clause subjects now accept type expressions; interface bases parse via `for ... : ...`; fixture printer updated and AST fixtures regenerated with new `where` shape.
- Typechecker: base interface signatures now participate in impl validation; self-type pattern names map to concrete `Self` substitutions.
- Kernel/stdlib: added `PartialEq`/`PartialOrd` impls for non-float primitives and big-number types; `Ord` impls now define `partial_cmp` to satisfy base interface contracts.
- Runtime: impl method resolution now prefers direct interface matches over base interface matches to avoid ambiguity (TS + Go).
- Fixtures: updated `implicit_generic_where_ambiguity` diagnostics + typecheck baseline; adjusted TreeMap stdlib test to include `partial_cmp` on custom `Ord` keys.
- Tests: `bun run scripts/export-fixtures.ts`; `cd v11/interpreters/ts && bun test`; `cd v11/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./...`.
- Kernel: compare String bytes for `PartialEq`/`Eq` to avoid recursive `==` on struct-backed strings.
- TS/Go interpreters: lambdas now treat `return` as local by catching return signals.
- Tests: `./v11/ablets test v11/stdlib/tests/text/string_methods.test.able --format tap`; `./v11/ablets test v11/stdlib/tests/text/string_builder.test.able --format tap`.

### 2026-01-21
- Typechecker: higher-kinded self patterns now reject concrete targets unless the impl is still a type constructor (arity-aware in TS + Go).
- Parser: type application parsing prefers left-associative space-delimited arguments; tree-sitter assets + corpus refreshed and Go parser relinked.
- Stdlib: removed Array overrides from `Enumerable` impl to rely on interface defaults; added exec fixtures for type-arg arity + associativity diagnostics.
- Tests: `cd v11/parser/tree-sitter-able && tree-sitter test -u`; `cd v11/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`; `./run_all_tests.sh --version=v11`.
- TS interpreter: separate interface vs subject type arguments in impl resolution (`findMethod`, `resolveInterfaceImplementation`, `matchImplEntry`, `typeImplementsInterface`) and widen runtime generic skipping for nested type expressions.
- TS interpreter: bind receiver type arguments into function env as `type_ref` and allow typed patterns to resolve generic type refs (fallback to wildcard for unknown generic names) to avoid non-exhaustive matches in generic matchers.
- Tests: `/home/david/sync/projects/able/v11/ablets test /home/david/sync/projects/able/v11/stdlib/tests/assertions.test.able`; `/home/david/sync/projects/able/v11/ablets test /home/david/sync/projects/able/v11/stdlib/tests` (timed out after 60s).
- Typechecker: overload resolution now prefers non-generic matches over generic, with generic specificity ranking aligned across TS + Go; unknown argument types no longer satisfy overload sets in Go to match TS (UFCS ambiguity case).
- CLI: test runner skips typechecking in `--list`/`--dry-run` modes to avoid spurious stdlib diagnostics.
- Fixtures: updated UFCS overload expectations and typecheck baseline entries; refreshed export-fixtures manifests for overload cases.
- Tests: `GOCACHE=$(pwd)/.gocache go test ./pkg/typechecker` in `v11/interpreters/go`; `bun test test/parity/fixtures_parity.test.ts -t "functions/ufcs_generic_overloads"`; `ABLE_FIXTURE_FILTER=errors/ufcs_overload_ambiguity bun run scripts/run-fixtures.ts`; `./run_all_tests.sh --version=v11`.
- Typechecker: bind higher-kinded self pattern placeholders to partially applied targets and apply interface args to `Self` when the impl target is still a type constructor (TS + Go); flatten applied types during substitution for nested constructor applications.
- Tests: `go test ./pkg/typechecker -run TestImplementationAllowsPartiallyAppliedConstructorWithSelfPattern` in `v11/interpreters/go`; `bun test test/typechecker/implementation_validation.test.ts -t "partially applied constructor"` in `v11/interpreters/ts`.
- Typechecker (Go): only apply explicit interface args to `Self` for constructor targets so inferred self-pattern args don't double-apply in method validation.
- Tests: `GOCACHE=/tmp/able-gocache go test ./pkg/typechecker` in `v11/interpreters/go`; `bun test test/typechecker` in `v11/interpreters/ts`.
- Typechecker (TS): avoid overriding impl generic substitutions with self-pattern bindings during impl validation to prevent nested applied types in stdlib interface checks.
- Tests: `./run_all_tests.sh --version=v11`.
- Typechecker: add regression coverage to ensure self-pattern placeholders remain in-scope for interface method signatures (TS + Go).
- Tests: `GOCACHE=/tmp/able-gocache go test ./pkg/typechecker -run TestImplementationAllowsSelfPatternPlaceholderInMethodSignature` in `v11/interpreters/go`; `bun test test/typechecker/implementation_validation.test.ts -t "self placeholders"` in `v11/interpreters/ts`.
- Typechecker: add package-scoped duplicate declaration coverage so same symbol names across packages do not conflict (TS session + Go program checker).
- Tests: `GOCACHE=/tmp/able-gocache go test ./pkg/typechecker -run TestProgramCheckerAllowsDuplicateNamesAcrossPackages` in `v11/interpreters/go`; `bun test test/typechecker/duplicates.test.ts -t "same symbol name"` in `v11/interpreters/ts`.
- Fixtures: added exec coverage for typechecker return mismatch diagnostics to mirror runtime behavior checks (`exec/11_01_return_statement_typecheck_diag`).
- Tests: `ABLE_FIXTURE_FILTER=11_01_return_statement_typecheck_diag bun run scripts/run-fixtures.ts` in `v11/interpreters/ts`; `GOCACHE=/tmp/able-gocache go test ./pkg/interpreter -run TestExecFixtures/11_01_return_statement_typecheck_diag` in `v11/interpreters/go`.

### 2026-01-22
- Stdlib: moved `collect` to the `Iterator` interface and kept `Iterable` focused on `each`/`iterator`.
- Spec: documented `Iterator.collect` in both iteration sections and removed `collect` from `Iterable`.
- Design: updated the eager/lazy collections split doc to reflect `Iterator.collect`.
- Tests not run (docs + stdlib interface change only).

### 2026-01-23
- Typechecker (TS): scope duplicate declaration tracking by package (prelude-safe) and allow local bindings to shadow package aliases during member access.
- Kernel: added `__able_os_args`/`__able_os_exit` externs to `v11/kernel/src/kernel.able` for kernel/stdlib alignment.
- Tests: `./run_stdlib_tests.sh --version=v11`; `./run_all_tests.sh --version=v11`.
- Typechecker: enforce missing type-argument diagnostics for concrete type annotations (TS + Go), while allowing constructor targets for impls/method sets; avoid duplicate arity diagnostics for constraints.
- Fixtures: added builtin type-arity + partial-application regression fixtures; refreshed typecheck baseline.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts --write-typecheck-baseline`; `cd v11/interpreters/ts && bun test test/typechecker/constraint_arity.test.ts`; `cd v11/interpreters/go && GOCACHE=$(pwd)/.gocache go test ./pkg/typechecker`.
- Fixtures: updated nested struct destructuring + Apply/Index dispatch fixtures to unwrap index results (`!`) so happy-path tests typecheck cleanly; refreshed typecheck baseline.
- Tests: `./v11/export_fixtures.sh`; `cd v11/interpreters/ts && bun run scripts/run-fixtures.ts --write-typecheck-baseline`.
- Spec: unified async model around `spawn`/`Future`, removed `proc`, renamed helpers to `future_*`, and rewrote Section 12 accordingly.
- Design: added `v11/design/future-unification.md`; updated concurrency/AST/typechecker/stdlib design notes to align with unified Future semantics.
- Plan: added a comprehensive implementation breakdown for the unified Future change.
- Tests not run (docs/spec/plan changes only).
- Parser/AST: removed `proc` keyword/`proc_expression` from tree-sitter, regenerated grammar artifacts, updated parser corpus; removed `ProcExpression` from TS/Go AST schemas and parser mappers; fixture JSON now uses `SpawnExpression`.
- Runtime: await/channel helpers now accept future contexts (TS + Go), and `proc_cancelled` works inside spawned futures in Go.
- Typechecker: future `cancel()` is allowed (TS + Go), and concurrency/typechecker tests updated to use spawn/future semantics.
- Fixtures: `.able` sources updated from `proc` → `spawn`, and expected error strings updated from `Proc failed/cancelled` → `Future failed/cancelled`.
- Tests not run (parser + runtime + fixture changes only).
