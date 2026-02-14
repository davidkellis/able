# Able Project Roadmap (v12 focus)

## Prompts

### New Session
Read AGENTS, PLAN, and the v12 spec, then start on the highest priority PLAN work. Proceed with next steps. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. You have permission to run tests. Tests should run quickly; no test should take more than one minute to complete.

### Next steps
Proceed with next steps as suggested; don't talk about doing it - do it. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. Tests should run quickly; no test should take more than one minute to complete.

## Scope
- Keep the frozen Able v10/v11 toolchains available for historical reference while driving all new language, spec, and runtime work in v12.
- Use the Go tree-walking interpreter as the behavioral reference and keep the Go bytecode interpreter in strict semantic parity.
- Preserve a single AST contract for every runtime so tree-sitter output can target both the historical branches and the actively developed v12 runtime; document any deltas immediately in the v12 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.

## Existing Assets
- `spec/full_spec_v10.md` + `spec/full_spec_v11.md`: authoritative semantics for archived toolchains. Keep them untouched unless a maintainer requests errata.
- `spec/full_spec_v12.md`: active specification for all current work; every behavioral change must be described here.
- `v10/` + `v11/`: frozen workspaces (read-only unless hotfix required).
- `v12/`: active development surface for Able v12 (`interpreters/go/`, `parser/`, `fixtures/`, `stdlib/`, `design/`, `docs/`).

## Ongoing Workstreams
- **Spec maintenance**: stage and land all wording in `spec/full_spec_v12.md`; log discrepancies in `spec/TODO_v12.md`.
- **Go runtimes**: maintain tree-walker + bytecode interpreter parity; keep diagnostics and fixtures aligned.
- **Tooling**: build a Go-based fixture exporter; update harnesses to remove TS dependencies.
- **Performance**: expand bytecode VM coverage; add perf harnesses for tree-walker vs bytecode.
- **WASM**: run the Go runtime in WASM with JS tree-sitter parsing and a defined host ABI.
- **Stdlib**: keep v12 stdlib aligned with runtime capabilities and spec requirements.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `v12/design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.

## Guardrails (must stay true)
- `./run_all_tests.sh` (v12 default) must stay green for the Go suites and fixtures.
- `./run_stdlib_tests.sh` must stay green (tree-walker + bytecode).
- Diagnostics parity is mandatory: tree-walker and bytecode interpreters must emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`.
- After regenerating tree-sitter assets under `v12/parser/tree-sitter-able`, force Go to relink the parser by deleting `v12/interpreters/go/.gocache` or running `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`.
- It is expected that some new fixtures will fail due to interpreter bugs/deficiencies. Implement fixtures strictly in accordance with the v12 spec semantics. Do not weaken or sidestep the behavior under test to "make tests pass".

## TODO (working queue: tackle in order, move completed items to LOG.md)
### Compiler AOT
- Compiler AOT: expand IR lowering coverage to full AST (control flow, patterns, error handling, concurrency, interop).
- Compiler AOT: emit Go packages per Able package with direct static calls.
- Compiler AOT: compile stdlib and kernel packages as Go code and wire into builds.
- Compiler AOT: compile `able.numbers.bigint` as part of stdlib compilation (relies on compiled arrays + numeric ops; no dedicated runtime primitive). In progress: `able build` precompile package discovery now includes core foundational/test packages (`able.spec`, `able.collections.enumerable`, `able.collections.list`, `able.collections.vector`, `able.collections.tree_map`, `able.collections.tree_set`, `able.collections.persistent_map`, `able.collections.persistent_sorted_set`, `able.collections.persistent_queue`, `able.collections.linked_list`, `able.collections.lazy_seq`, `able.collections.hash_map`, `able.collections.hash_set`, `able.collections.deque`, `able.collections.queue`, `able.collections.array`, `able.collections.range`, `able.collections.bit_set`, `able.collections.heap`, `able.concurrency`, `able.concurrency.concurrent_queue`, `able.math`, `able.core.numeric`, `able.fs`, `able.io.temp`, `able.os`, `able.process`, `able.term`, `able.test.protocol`, `able.test.harness`, `able.test.reporters`) and the numeric package set (`able.numbers.bigint`, `able.numbers.biguint`, `able.numbers.int128`, `able.numbers.uint128`, `able.numbers.rational`, `able.numbers.primitives`) by default, bigint/biguint/int128/uint128/rational/primitives stdlib exec fixture coverage plus list/vector/tree_map/tree_set/persistent_map/persistent_sorted_set/persistent_queue/linked_list/lazy_seq/hash_map/hash_set/deque/queue/array/range/bit_set/heap collection fixture coverage, channel/mutex/concurrent_queue stdlib concurrency fixture coverage, math/core-numeric fixture coverage, fs/path fixture coverage, io/temp fixture coverage, os fixture coverage, process fixture coverage, term fixture coverage, and test-harness/reporters fixture coverage is wired into compiled strict/no-fallback parity gates, `able test --compiled` now has strict no-fallback stdlib gates for `bigint.test.able` + `biguint.test.able`, for `int128.test.able` + `uint128.test.able` + `rational.test.able`, for `numbers_numeric.test.able`, for foundational suites (`simple.test.able`, `assertions.test.able`, `enumerable.test.able`), for collections suites (`list.test.able`, `vector.test.able`), for ordered-collections suites (`tree_map.test.able`, `tree_set.test.able`), for persistent-collections suites (`persistent_map.test.able`, `persistent_set.test.able`), for persistent sorted/FIFO suites (`persistent_sorted_set.test.able`, `persistent_queue.test.able`), for linked/lazy suites (`linked_list.test.able`, `lazy_seq.test.able`), for hash suites (`collections/hash_map_smoke.test.able`, `collections/hash_set.test.able`), for deque/queue smoke suites (`collections/deque_smoke.test.able`, `collections/queue_smoke.test.able`), for array/range smoke suites (`collections/array_smoke.test.able`, `collections/range_smoke.test.able`), for bit_set/heap suites (`bit_set.test.able`, `heap.test.able`), for concurrency suites (`concurrency/channel_mutex.test.able`, `concurrency/concurrent_queue.test.able`), for math/core numeric suites (`math.test.able`, `core/numeric_smoke.test.able`), for fs/path smoke suites (`fs_smoke.test.able`, `path_smoke.test.able`), for io smoke suite (`io_smoke.test.able`), for os smoke suite (`os_smoke.test.able`), for process smoke suite (`process_smoke.test.able`), for term smoke suite (`term_smoke.test.able`), for harness/reporters smoke suite (`harness_reporters_smoke.test.able`), and rescue-expression lowering now supports mixed monitored/clause result types in statement contexts without fallback so numeric conversion/underflow rescue assertions stay under strict compilation, including method-chain process spec coverage via selector imports for `with_cwd` and `with_env`; compiler fixture gate defaults are now tuned for sub-10m package runs while full sweeps remain available via `ABLE_COMPILER_EXEC_FIXTURES=all`, `ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all`, and `ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all`, plus dedicated wrappers (`v12/run_compiler_full_matrix.sh`, `v12/run_all_tests.sh --compiler-full-matrix`) and CI automation (`.github/workflows/compiler-full-matrix-nightly.yml`).
- Compiler AOT: fully bypass interpreter interface lookup once compiled dispatch covers all impl shapes (strict dispatch flag is in place when all unnamed impl methods are compileable).
- Compiler AOT: continue expanding compiled main no-bootstrap execution. Current scope now includes fallback-free, non-dynamic, fully-compileable multi-package programs with statically-seeded imports across package/selector/wildcard forms plus named impl namespace imports (public functions/structs/interfaces/unions/impl namespaces), including named impl namespaces with overloaded methods; package registration orchestration is emitted in a dedicated generated artifact (`compiled_package_aggregators.go`) while per-package registrars are emitted into per-package generated files (`compiled_pkg_registrar_*.go`); package definition seeding is emitted into per-package generated files (`compiled_pkg_defs_*.go`); callable registration is emitted into per-package generated files (`compiled_pkg_callables_*.go`); method/impl registration is emitted into per-package generated files (`compiled_pkg_methods_impls_*.go`); interface-dispatch + method-overload registration is emitted in a dedicated generated artifact (`compiled_interface_dispatch.go`); register/run entrypoint emission is emitted in a dedicated generated artifact (`compiled_register.go`); no-bootstrap static import seeding is emitted in a dedicated generated artifact (`compiled_import_seeding.go`); callable import seeding now binds directly to compiled call table entries; compiler `RequireNoFallbacks` option now provides an explicit fail-fast guard for no-silent-fallback builds, with CLI controls via `able build --no-fallbacks|--allow-fallbacks` and env-driven strict mode (`ABLE_COMPILER_REQUIRE_NO_FALLBACKS`) applied to both `able build` and `able test --compiled`; compiler fixture/parity gates now compile in strict no-fallback mode by default (`ABLE_COMPILER_FIXTURE_REQUIRE_NO_FALLBACKS=0|false|off|no` to temporarily disable for local debugging); `able build` now precompiles stdlib/kernel package graphs by default (`ABLE_BUILD_PRECOMPILE_STDLIB` + `--precompile-stdlib|--no-precompile-stdlib`) and bundles stdlib/kernel sources into external build outputs for bootstrap relocation.
- Compiler AOT: ensure everything in v12/design/compiler-aot.md is implemented.
### WASM
- WASM: prototype JS tree-sitter parsing that feeds AST into the Go/WASM runtime.
- WASM: build a minimal `ablewasm` runner (Node + browser harness) once the Go runtime builds to WASM.
- WASM: document the WASM deployment contract in `v12/docs/`.
### Regex syntax
- Regex syntax: add regex AST nodes and grammar in tree-sitter (quantifiers, groups, classes, alternation).
- Regex syntax: wire AST mapping for regex nodes in Go parser.
- Regex syntax: add fixtures/tests for regex AST output and exec behavior; keep stdlib engine parity.
- Regex syntax: add parser corpus cases that cover nested groups, alternation, and escaped quantifiers.
- Regex syntax: update `spec/TODO_v12.md` with remaining regex syntax/semantics gaps.
- Regex syntax: align stdlib regex implementation with parser outputs as grammar coverage expands.
