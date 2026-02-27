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
- Native-lowering requirement: static Able semantics should lower to host-native Go constructs (concrete structs/scalars/collections) rather than interpreter object-model execution paths; generic dynamic carriers are reserved for explicit boundary/ABI/polymorphic residual cases.
- Kickoff changes landed:
  - [x] Map statically-typed `Array ...` locals to compiled `*Array` instead of defaulting to `runtime.Value`.
  - [x] Fix local `=` declaration fallback so unbound local assignments do not compile to `__able_global_set/__able_global_get`.
  - [x] Lower typed `Array` index read/write (`arr[idx]`, `arr[idx] = v`) through direct `runtime.ArrayStore*` paths for compiled `*Array` receivers.
  - [x] Keep static no-fallback enforcement active while applying these optimizations.
- Immediate unit of work (execute in order):
  - [ ] Add a static native-lowering audit for user-defined nominal types (struct/union/interface views) and primitive locals to identify/remove avoidable `runtime.Value` carriers in compiled function bodies.
  - [ ] Add regression fixtures that assert struct-heavy static programs emit concrete Go typed locals/field access paths and avoid dynamic dispatch helpers (`__able_member_get_method`, `__able_call_value`, `__able_call_named`) outside explicit dynamic-boundary wrappers.
  - [ ] Encode/enforce allowed dynamic-carrier touchpoints in codegen (explicit dynamic boundary adapters, residual runtime-polymorphic dispatch, extern ABI conversion), with compile-time errors for new static misuse.
  - [x] Add call-site intrinsics for typed `Array` methods in hot paths (`push`, `len`, `get`, `set`) to bypass dynamic member lookup / `__able_call_value`.
  - [x] Add compiler regression fixtures proving typed-array locals in static code emit no `__able_global_get/__able_global_set` in compiled function bodies.
  - [x] Add compiler regression fixtures proving typed-array loops emit no `__able_member_get_method`/`__able_call_value` for `push/get/set/len`.
  - [x] Add fast-path loop lowering for `while` loops without explicit `break`/`continue`/`rescue` needs (avoid per-iteration closure + `defer` scaffolding).
  - [x] Extend array literal lowering so typed contexts keep native compiled `*Array` paths and avoid unnecessary struct<->runtime boxing.
  - [x] Audit `Array` stdlib compiled methods (`push`, `len`, `get`, `set`, `refresh_metadata`) for redundant runtime round-trips; remove avoidable metadata refresh churn.
  - [x] Add benchmark fixtures (`noop`, `sieve_count`, `sieve_full`) and track real/user/sys + GC count in CI-adjacent perf script (non-blocking, report-only).
  - [x] Document required/allowed `runtime.Value` usage categories in `spec/full_spec_v12.md` and note staged limits in `spec/TODO_v12.md`.
  - [x] Design and stage monomorphized container ABI proposal (`Array<T>` native element typing) with compatibility constraints before implementation.
  - [x] Add stage-0 mono-array flag scaffolding (`--experimental-mono-arrays`, `ABLE_EXPERIMENTAL_MONO_ARRAYS`) through compiler options and generated feature marker.
  - [ ] Implement staged monomorphized array lowering behind a compiler flag once design/spec update is approved.
    - Partial stage-1 landed: runtime typed stores (`i32`, `i64`, `bool`, `u8`) + compiler lowering for typed array literals, index read/write, and `Array.push/len/get/set` intrinsics when static element type is known.
    - Runtime array capacity growth now uses amortized expansion to remove per-push reallocation thrash in both dynamic and mono paths.
    - Added mono-array boundary regression coverage for explicit dynamic calls (compiled callback conversion + compiled->dynamic->compiled array roundtrip under `--experimental-mono-arrays`).
    - Added mono-array boundary coverage for nullable/union/interface callback conversion shapes (success + failure) under `--experimental-mono-arrays`.
    - Captured compiled-only perf snapshot (5-run avg, 2026-02-26, after native index-int + propagation/cast de-boxing):
      - `bench/noop`: mono `0.060s` real / `3.20` GC vs default `0.060s` / `3.00` GC.
      - `bench/sieve_count`: mono `0.072s` real / `5.00` GC vs default `0.094s` / `10.80` GC.
      - `bench/sieve_full`: mono `0.156s` real / `22.40` GC vs default `0.376s` / `54.40` GC.
    - Landed index-int lowering optimization for array read/write/get/set paths: native integer index types now avoid `bridge.ToInt` + `bridge.AsInt` boxing round-trips.
    - Landed mono propagation/cast de-boxing for typed index reads (`arr[idx]! as i64` style paths) so mono read values stay native where semantically safe.
    - Landed compatibility fixes for mixed array carriers and interface-typed assignment coercion:
      - `Array` struct converters now accept/synchronize raw `*runtime.ArrayValue` carriers at explicit runtime boundaries.
      - Interface-annotated local assignment now applies `bridge.MatchType` coercion so interface args are preserved in compiled dispatch.
    - Default-on rollout criteria (stage-1 gate):
      - strict compiler fixture audits pass in `ABLE_TYPECHECK_FIXTURES=strict` (`ExecFixtures`, `StrictDispatch`, `InterfaceLookupBypass`, `BoundaryFallbackMarker`, `ExecFixtureFallbacks`);
      - dynamic-boundary suites pass, including mono-array boundary tests;
      - compiled-only perf stability (5-run avg) stays within guardrails: mono real-time regression <= 10%, mono GC regression <= 15% vs default on `bench/noop`, `bench/sieve_count`, `bench/sieve_full`.
    - Gate status (2026-02-26):
      - strict fixture audits: PASS.
      - dynamic-boundary suites (`TestCompilerDynamicBoundary*`): PASS.
      - compiled-only perf (5-run avg):
        - `bench/noop`: default `0.062s` / `3.20` GC; mono `0.060s` / `3.20` GC.
        - `bench/sieve_count`: default `0.072s` / `5.40` GC; mono `0.074s` / `5.20` GC.
        - `bench/sieve_full`: default `0.164s` / `23.20` GC; mono `0.164s` / `23.00` GC.
      - Result: stage-1 rollout gate satisfied; default-on enabled in CLI flows with explicit opt-out (`--no-experimental-mono-arrays` / `ABLE_EXPERIMENTAL_MONO_ARRAYS=off`); remaining work is staged rollout mechanics and eventual flag retirement criteria.
- Definition of done for this workstream:
  - [x] Static typed-array hot paths (`push/get/set/len` + index ops) compile without dynamic member dispatch in generated function bodies.
  - [x] Static local-variable fallback semantics (`=` declares when unbound) stay local-scope and do not route through global environment helpers.
  - [x] Sieve-style benchmark shows measurable runtime and GC reduction versus pre-work baseline, with unchanged program output.
  - [x] No regressions in compiler strict static checks, fixture parity, and dynamic-boundary behavior.
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
