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
- Status: in progress. History/details for completed compiler shim slices are tracked in `LOG.md`.
- Completed milestone: shared runtime-call shim normalization and strict dispatch/no-fallback hardening are in place and green on focused + strict-total + strict no-fallback gates (latest timings recorded in `LOG.md`).
- Validation snapshot: full strict compiler matrix with `...FIXTURES=all` plus fallback audit is green (`v12/run_compiler_full_matrix.sh --typecheck-fixtures=strict`, latest gate timings in `LOG.md`).
- Spec snapshot: Compiler AOT contract gaps tracked in `spec/TODO_v12.md` are cleared; normative boundary/ABI/dispatch/compiled-dependency semantics are now documented in `spec/full_spec_v12.md`.
- Regression snapshot: fixed compiled stdlib CLI hang in `math + core/numeric` caused by UFCS fallback precedence in generated `__able_member_get_method`; interface dispatch now precedes UFCS fallback and compiled CLI no-fallback suite is green again (details/timings in `LOG.md`).
- Coverage snapshot: stdlib smoke strict lookup follow-up is closed by adding `06_12_20_stdlib_math_core_numeric`, `06_12_22_stdlib_io_temp`, `06_12_23_stdlib_os`, `06_12_24_stdlib_process`, `06_12_25_stdlib_term`, and `06_12_26_stdlib_test_harness_reporters` to default interface-lookup audits (`TestCompilerInterfaceLookupBypassForStaticFixtures`).
- Parity snapshot: block-local function-definition statements (`fn` inside function bodies) now lower directly in compiled mode (including recursive self-reference) and stay green under `RequireNoFallbacks`.
- Parity snapshot: block-local type-definition statements (`type`/`struct`/`union`/`interface`) now lower directly in compiled mode and stay green under `RequireNoFallbacks`, including local interface signatures that carry default impl bodies.
- Parity snapshot: block-local `methods` and `impl` definitions now lower directly in compiled mode (via explicit bridge statement evaluation) and stay green under `RequireNoFallbacks`.
- Parity snapshot: compiled definition metadata now preserves generic-parameter interface constraints, `where`-clause constraints, and interface signature default-impl bodies when rendering struct/union/interface definitions (package-level and block-local) under no-fallback gates.
- Parity snapshot: method receiver detection now matches interpreter semantics for `Self`-typed first parameters (even when the parameter is not named `self`), keeping instance-method registration/dispatch compiled and strict-gate green.
- Safety snapshot: bridge global-lookup hardening is in place (toggle + counters + entry-env seeding + bridge struct hydration), and strict-total interface/global lookup audits are now green on both default fixtures and `ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all`.
- Operations snapshot: compiler full-matrix runner is now bounded per suite (`ABLE_COMPILER_SUITE_TIMEOUT` default `25m`, hard wall `ABLE_COMPILER_SUITE_WALL_TIMEOUT` default `30m`) so stalled suites fail fast instead of hanging indefinitely.

Active remaining backlog (finish in this order):
- Compiler AOT: complete full-AST lowering parity (control flow, patterns, error handling, concurrency, interop) with no silent fallback in static programs.
- Compiler AOT: finish per-package codegen architecture for all compileable modules (direct static calls, typed dispatch, compiled registration).
- Compiler AOT: finish stdlib+kernel compiled-by-default pipeline end-to-end (including `able.numbers.bigint`) and keep strict parity gates green.
- Compiler AOT: complete explicit dynamic-boundary contract implementation (`dynimport`/`defpackage`/dynamic eval only) and prove no interpreter execution for non-dynamic programs.
- Compiler AOT: close all remaining items from `v12/design/compiler-aot.md` work breakdown and definition-of-done.

Definition of done for Compiler AOT (PLAN close criteria):
- Non-dynamic programs execute fully compiled with no interpreter execution.
- Dynamic features execute only through explicit boundary paths.
- Stdlib and kernel compile and execute directly in compiled mode.
- Compiler fixture + stdlib compiled gates are green in strict no-fallback mode.
- Spec semantics parity is preserved (`spec/full_spec_v12.md`).
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
