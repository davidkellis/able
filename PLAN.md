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
- Status: **COMPLETE**. All definition-of-done criteria met. History in `LOG.md`.
- No-bootstrap execution: non-dynamic programs run fully compiled (`interpreter.New()` instantiated for runtime services, `EvaluateProgram()` never called). Validated via `TestCompilerNoBootstrapExecFixtures`: 222 pass, 13 fail (12 inherently dynamic/IO + 1 pre-existing), 5 skip out of 240 total.
- Bootstrap skip detection: `TestCompilerMainSkips` (7 tests) verifies generated `main.go` omits `EvaluateProgram()` for static programs.
- Fallback audit: clean (`TestCompilerExecFixtureFallbacks` runs by default).
- Full matrix: `v12/run_compiler_full_matrix.sh --typecheck-fixtures=strict` green.
- Spec: compiler AOT contract fully documented in `spec/full_spec_v12.md`.
### Compiler AOT Boundary Hardening (active priority)
- Goal: enforce v12 AOT contract that compiled static code does not use interpreter execution paths; interpreter usage is allowed only for explicit dynamic features (`dynimport`, `dyn.def_package`, `dyn.eval`, etc.).
- Immediate unit of work (execute in order):
  - [x] Make static fallback rejection the default for `able build` (require-no-fallbacks on by default for non-dynamic builds; keep explicit override for migration/debug).
  - [x] Wire compile-time policy: when dynamic features are not present, any collected fallback is a hard compile error (not warning/runtime boundary).
  - [x] Add/strengthen tests so static fixtures assert zero boundary fallback calls by default (remove env-gated audit behavior for core static checks).
  - [x] Keep dynamic fixtures explicit: dynamic-boundary tests must prove boundary calls only occur for explicit dynamic operations.
  - [x] Remove static fallback sites that currently route through interpreter evaluation (starting with local `methods` / `impl` statement evaluation paths).
  - [x] Remove static named/value call fallback to bridge interpreter dispatch; unresolved static calls must fail compile.
  - [x] Eliminate unconditional interpreter bootstrap in static generated `main.go`; static path must not require interpreter initialization.
  - [x] Update `spec/full_spec_v12.md` and `spec/TODO_v12.md` to reflect enforcement status and any temporary implementation limits.
- Definition of done for this workstream:
  - [x] Non-dynamic compiled programs execute without interpreter evaluation fallback calls (`__ABLE_BOUNDARY_FALLBACK_CALLS=0` in static audit runs).
  - [x] Non-dynamic compiled `main.go` omits interpreter bootstrap/eval paths and does not require interpreter-backed bridge operations for static semantics.
  - [x] Dynamic programs still function with explicit boundary transitions and retain parity with tree-walker/bytecode behavior.
  - [x] `./run_all_tests.sh` and compiler fixture audits stay green with the new strict policy.
### Compiler AOT Performance and `runtime.Value` Reduction (active priority)
- Goal: minimize `runtime.Value` usage in static compiled code; keep it only where semantically required (explicit dynamic boundary crossing, interface/runtime-polymorphic dispatch, and ABI conversion points).
- Kickoff changes landed:
  - [x] Map statically-typed `Array ...` locals to compiled `*Array` instead of defaulting to `runtime.Value`.
  - [x] Fix local `=` declaration fallback so unbound local assignments do not compile to `__able_global_set/__able_global_get`.
  - [x] Lower typed `Array` index read/write (`arr[idx]`, `arr[idx] = v`) through direct `runtime.ArrayStore*` paths for compiled `*Array` receivers.
  - [x] Keep static no-fallback enforcement active while applying these optimizations.
- Immediate unit of work (execute in order):
  - [ ] Add call-site intrinsics for typed `Array` methods in hot paths (`push`, `len`, `get`, `set`) to bypass dynamic member lookup / `__able_call_value`.
  - [ ] Add compiler regression fixtures proving typed-array locals in static code emit no `__able_global_get/__able_global_set` in compiled function bodies.
  - [ ] Add compiler regression fixtures proving typed-array loops emit no `__able_member_get_method`/`__able_call_value` for `push/get/set/len`.
  - [ ] Add fast-path loop lowering for `while` loops without explicit `break`/`continue`/`rescue` needs (avoid per-iteration closure + `defer` scaffolding).
  - [ ] Extend array literal lowering so typed contexts keep native compiled `*Array` paths and avoid unnecessary struct<->runtime boxing.
  - [ ] Audit `Array` stdlib compiled methods (`push`, `len`, `get`, `set`, `refresh_metadata`) for redundant runtime round-trips; remove avoidable metadata refresh churn.
  - [ ] Add benchmark fixtures (`noop`, `sieve_count`, `sieve_full`) and track real/user/sys + GC count in CI-adjacent perf script (non-blocking, report-only).
  - [ ] Document required/allowed `runtime.Value` usage categories in `spec/full_spec_v12.md` and note staged limits in `spec/TODO_v12.md`.
  - [ ] Design and stage monomorphized container ABI proposal (`Array<T>` native element typing) with compatibility constraints before implementation.
  - [ ] Implement staged monomorphized array lowering behind a compiler flag once design/spec update is approved.
- Definition of done for this workstream:
  - [ ] Static typed-array hot paths (`push/get/set/len` + index ops) compile without dynamic member dispatch in generated function bodies.
  - [ ] Static local-variable fallback semantics (`=` declares when unbound) stay local-scope and do not route through global environment helpers.
  - [ ] Sieve-style benchmark shows measurable runtime and GC reduction versus pre-work baseline, with unchanged program output.
  - [ ] No regressions in compiler strict static checks, fixture parity, and dynamic-boundary behavior.
### WASM
- WASM: prototype JS tree-sitter parsing that feeds AST into the Go/WASM runtime (**in progress**).
  - Landed staging scaffold: `cmd/ablewasm` (`GOOS=js GOARCH=wasm`) + `pkg/wasmhost` JSON bridge and `v12/wasm/` Node prototype (`web-tree-sitter` subset adapter + runner).
  - Next: broaden AST adapter coverage beyond the initial expression/import subset and wire it to the host ABI path in `v12/docs/wasm-host-abi.md`.
- WASM: build a minimal `ablewasm` runner (Node + browser harness) once the Go runtime builds to WASM.
- WASM: document the WASM deployment contract in `v12/docs/`.
### Regex syntax
- Regex syntax: add regex AST nodes and grammar in tree-sitter (quantifiers, groups, classes, alternation).
- Regex syntax: wire AST mapping for regex nodes in Go parser.
- Regex syntax: add fixtures/tests for regex AST output and exec behavior; keep stdlib engine parity.
- Regex syntax: add parser corpus cases that cover nested groups, alternation, and escaped quantifiers.
- Regex syntax: update `spec/TODO_v12.md` with remaining regex syntax/semantics gaps.
- Regex syntax: align stdlib regex implementation with parser outputs as grammar coverage expands.
