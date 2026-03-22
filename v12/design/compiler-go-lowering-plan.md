# Compiler Go Lowering Completion Plan

## Status

This is the ordered execution plan for reaching the lowering architecture defined
in `v12/design/compiler-go-lowering-spec.md`.

This plan is deliberately organized around completion goals, not around whatever
local emitter file was touched most recently.

## Finish Line

The compiler is done when it satisfies all of these conditions together:

1. Every statically representable Able value uses a native Go carrier.
2. Static control flow, patterns, and joins stay on native carriers.
3. Static dispatch is compiled directly to Go dispatch.
4. The interpreter is used only at explicit dynamic boundaries.
5. The compiled runtime implements language semantics directly in Go instead of
   modeling static execution in terms of interpreter behavior.
6. The benchmarked compiled path is fast and free of known avoidable dynamic
   scaffolding.

## Rules For The Remaining Work

1. Do not add new named-nominal fast paths for non-primitive types.
2. Do not fix static compiler bugs by routing through interpreter helpers.
3. Do not let emitter-local code invent new carrier or dispatch logic.
4. Put lowering knowledge into the shared synthesis points described in the
   lowering spec.
5. Treat interpreters as the behavioral oracle and dynamic-boundary engine, not
   as the implementation substrate for static compiled semantics.

## Current State Summary

Substantial native lowering is already landed:
- large parts of arrays, structs, interfaces, callables, joins, and control
  flow now stay native
- explicit dynamic-boundary enforcement exists
- benchmark work has already removed major scaffolding on hot array/matrix paths

But the compiler is not done because these broader goals are still incomplete:
- carrier synthesis still has representable fallback gaps
- some pattern/control surfaces still degrade when type recovery gets indirect
- some dispatch surfaces still retain residual runtime paths
- the compiled runtime still contains areas whose semantics are too closely tied
  to interpreter-oriented helpers
- validation is still distributed across many narrow slices rather than a small
  number of hard completion gates

## Ordered Work Program

Work through this list in order.

### 1. Centralize Lowering Knowledge

Status:
- complete on 2026-03-21.
- the canonical lowering facade now lives in:
  - `v12/interpreters/go/pkg/compiler/generator_lowering_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_control.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_boundaries.go`
- enforcement now lives in:
  - `v12/interpreters/go/pkg/compiler/compiler_lowering_facade_audit_test.go`

#### Goal

Make lowering decisions come from one reusable source per concern.

#### Why

Without this, the compiler keeps accumulating one-off emitter fixes and
structure-specific branches.

#### Required work

1. Establish one canonical type-normalization entrypoint.
2. Establish one canonical carrier-synthesis entrypoint.
3. Establish one canonical join/pattern-synthesis entrypoint.
4. Establish one canonical dispatch-synthesis entrypoint.
5. Establish one canonical control-envelope entrypoint.
6. Establish one canonical boundary-adapter entrypoint.
7. Audit codegen so emitters call these entrypoints instead of making local
   fallback decisions.

#### Main subsystems

- compiler type mapping
- native union/interface/callable synthesis
- pattern and join lowering
- dispatch lowering
- control-flow/result propagation
- dynamic-boundary adapters

#### Definition of done

- No emitter-local carrier fallback remains for representable static types.
- No new lowering change requires a named non-primitive special case.
- Code review can point to one shared synthesis point for each lowering class.

#### Proof

- source audit tests that fail on new ad hoc fallback patterns
- representative generated-source audits

### 2. Finish Native Carrier Completeness

Status:
- complete on 2026-03-22.
- `types.go` and `generator_native_unions.go` no longer silently widen fully
  bound representable carriers to `runtime.Value` / `any`.
- built-in `DivMod T` now lowers through the shared nominal struct path rather
  than a dedicated `any` fallback.
- shared native `Error | void` carrier support is now present, so native
  `!void` signatures no longer force runtime result carriers.
- compile-shape regressions now pin:
  - concrete `DivMod i32`
  - parameterized union/result alias locals containing `DivMod i32`
  - direct native `!void` returns

#### Goal

Every statically representable Able type expression maps to a native Go carrier.

#### Required work

1. Remove representable `runtime.Value` / `any` fallbacks from carrier synthesis.
2. Complete carrier synthesis for:
   - nullable values
   - results
   - closed unions
   - open unions with explicit residual dynamic members only when necessary
   - interfaces
   - callables
   - fully bound generic nominal types
3. Ensure alias-expanded and recovered type expressions take the same carrier
   path as directly written types.
4. Ensure join/carrier recovery can reuse normalized type expressions instead of
   giving up when an intermediate local is temporarily stored dynamically.

#### Main subsystems

- type normalization
- carrier synthesis
- native union synthesis
- nullable/result synthesis
- generic specialization metadata

#### Definition of done

- A representable static type never defaults to `runtime.Value` or `any`.
- Generated locals/params/returns/fields for representable types stay native.
- Only truly dynamic payloads retain dynamic carriers.

#### Proof

- carrier audit tests over representative type families
- generated-source scans for representable `runtime.Value` / `any` locals
- fixture slices covering unions, interfaces, callables, results, and generics

### 3. Finish Native Pattern And Control-Flow Completeness

Status:
- complete on 2026-03-22.
- shared join/native recovery now covers `if`, `match`, `rescue`, `or {}`,
  `loop`, and `breakpoint`, including nil-capable and common-existential join
  cases.
- typed pattern bindings now reuse recovered native carriers instead of
  defaulting the bound local to `runtime.Value`, including rescue bindings and
  native-union whole-value interface bindings.
- representative static pattern/control bodies are source-audited against
  runtime type-match helpers and panic/recover/IIFE control scaffolding.

#### Goal

All static control and pattern constructs stay native when their subject/result
shapes are statically representable.

#### Required work

1. Keep `if` / `match` / `rescue` / `or {}` joins on shared native carriers.
2. Keep loop and breakpoint result carriers native.
3. Remove runtime typed-pattern fallback when the subject type is statically
   recoverable.
4. Keep struct, array, union, interface, and typed patterns on native binding
   carriers.
5. Make `raise`, `rescue`, `ensure`, `rethrow`, and `!` use explicit native
   control/result carriers instead of interpreter-style helpers.
6. Remove remaining IIFE/panic/recover control scaffolding from static paths.

#### Main subsystems

- join inference
- pattern synthesis
- control envelope synthesis
- rescue/or/propagate lowering
- loop/breakpoint lowering

#### Definition of done

- Static pattern lowering does not require runtime type-match helpers.
- Static control constructs do not require `runtime.Value` join locals.
- Non-local control uses explicit control envelopes only where needed.

#### Proof

- generated-source audits for `__able_try_cast(...)`, `bridge.MatchType(...)`,
  `panic`, `recover`, and IIFE scaffolding on static fixtures
- fixture slices for `match`, `rescue`, `or {}`, loops, breakpoints, and typed
  patterns

### 4. Finish Native Dispatch Completeness

Status:
- complete on 2026-03-22.
- shared dispatch recovery now rehydrates recoverable `runtime.Value` / `any`
  call/member/index targets onto their native carriers before dispatch.
- local concrete/interface `Apply` bindings now use the shared static apply
  path instead of dynamic call helpers.
- mixed-source pure-generic interface dispatch now prefers the more concrete
  compiled specialization, which keeps representative generic interface calls
  off runtime method dispatch.
- proof now includes:
  - `v12/interpreters/go/pkg/compiler/compiler_dispatch_completeness_test.go`
  - the shared dispatch fixture slice in `PLAN.md`

#### Goal

All statically resolved operations compile to direct Go dispatch.

#### Required work

1. Static field access/assignment -> direct carrier field operations.
2. Static method calls -> direct compiled impl/helper calls.
3. Static interface/default-method calls -> compiled interface dispatch helpers.
4. Static callable invocations -> direct callable carrier invocation.
5. Static index operations -> direct wrapper/slice/protocol dispatch.
6. Bound method values and partials -> shared callable carrier synthesis.
7. Generic interface/default-method dispatch -> compiled specialization helpers,
   not runtime dispatch.
8. Remove residual runtime helper dispatch from static call/member/index paths.

#### Main subsystems

- method/call lowering
- interface dispatch synthesis
- callable synthesis
- index lowering
- assignment lowering

#### Definition of done

- Static compiled bodies do not emit runtime helper dispatch for resolved calls.
- New nominal types gain direct dispatch automatically through shared lowering.

#### Proof

- generated-source audits for `__able_call_value(...)`,
  `__able_method_call_node(...)`, `__able_member_get*`, and `__able_index*`
  outside explicit dynamic boundaries
- fixture slices for structs, interfaces, callables, indexable types, and
  generic/default-method surfaces

### 5. Finish Compiled Runtime Core Independence

Status:
- complete on 2026-03-22.
- static compiled helper families now lower directly to Go `_impl` runtime-core
  helpers on static paths instead of `__able_extern_*` or helper-to-helper
  `__able_extern_call(...)` chains.
- zero-arg callable syntax and `Await.default` zero-arg callback
  specialization now stay on native callable carriers across compiled static
  and spawned-task paths.

#### Goal

The compiled runtime must implement language semantics directly in Go instead of
borrowing interpreter-oriented execution as its normal implementation strategy.

#### Required work

1. Audit every compiled-runtime helper used on static paths.
2. Replace helpers whose semantics are still effectively modeled as interpreter
   operations with direct Go implementations.
3. Keep only true dynamic-boundary helpers dependent on interpreter/runtime
   object-model values.
4. Ensure array/map/range/string/concurrency runtime services are explicit Go
   runtime services, not thin wrappers around interpreter evaluation.
5. Keep control-flow semantics in compiled runtime helpers aligned with the
   shared control-envelope model.

#### Main subsystems

- compiled runtime helper emission
- array/map/string/range runtime core
- concurrency runtime core
- boundary helper families

#### Definition of done

- Static compiled helper families execute as normal Go logic.
- Interpreter execution is absent from static runtime core paths.

#### Proof

- source audit over emitted helper families
- fixture slices that exercise runtime-core helpers without dynamic features
- no-bootstrap and no-fallback compiler tests

### 6. Finish Boundary Containment

Status:
- complete on 2026-03-22.
- the final explicit dynamic-boundary helper set is now locked to:
  - `call_value` through `__able_call_value(...)`
  - `call_named` through `__able_call_named(...)`
  - `call_original` through generated original-wrapper calls
- representative static no-bootstrap fixture execution now proves:
  - zero fallback boundary calls
  - zero explicit dynamic boundary calls
  - zero interface/member lookup fallback calls
  - zero global lookup fallback calls
- representative proof now includes:
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_containment_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_touchpoint_audit_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_main_bootstrap_test.go`

#### Goal

All residual runtime/interpreter carriers live only at explicit dynamic or ABI
edges.

#### Required work

1. Enumerate the allowed boundary categories exactly.
2. Tighten boundary adapters so they convert at the edge and return to native
   carriers immediately after the edge.
3. Remove residual dynamic leakage from static paths.
4. Audit compiled fixtures so static fixtures cross zero dynamic boundaries.
5. Keep dynamic fixtures explicit and narrow.

#### Allowed boundary categories

- `dynimport`
- dynamic package mutation/definition
- dynamic evaluation/metaprogramming
- extern/host ABI conversion
- explicit compiled <-> dynamic callback boundaries
- values that already originate from dynamic runtime payloads

#### Allowed explicit boundary helper set

- `__able_call_value(...)`
  - explicit compiled value/callback -> dynamic callable entry
- `__able_call_named(...)`
  - explicit compiled -> dynamic named lookup/call entry
- generated `call_original` wrappers
  - explicit preserved original/dynamic implementation entry

All other generated helpers must either stay fully native or act only as
immediate carrier adapters adjacent to one of the three explicit entry helpers
or an extern/host ABI edge.

#### Definition of done

- A static source program does not cross boundary helpers unless it explicitly
  uses dynamic features.
- Dynamic fixtures prove only the allowed edges remain.

#### Proof

- boundary marker harnesses
- strict no-fallback compiler audit
- generated-source audits around boundary helper use

### 7. Finish Performance Completeness

Status:
- complete on 2026-03-22.
- added the reduced checked-in recursion benchmark:
  - `v12/fixtures/bench/fib_i32_small/main.able`
- shared compiled callable/runtime env scaffolding now swaps package envs only
  when necessary through:
  - `v12/interpreters/go/pkg/compiler/bridge/bridge_env_swap.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_env_helpers.go`
- representative performance closure is recorded in:
  - `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`
- current representative compiled results after the shared env-swap fast path:
  - `bench/fib_i32_small`: `2.7567s`, `0.00` GC
  - `bench/heap_i32_small`: `0.2900s`, `5.00` GC
  - `bench/linked_list_iterator_pipeline_i64_small`: `0.1433s`, `9.67` GC
  - `bench/matrixmultiply_f64_small`: `0.1167s`, `7.33` GC
  - `examples/benchmarks/matrixmultiply`: `1.0633s`, `13.33` GC

#### Goal

Make the compiled benchmark family fast without violating the lowering rules.

#### Required work

1. Use checked-in reduced benchmarks to isolate hot shared lowering gaps.
2. Remove only shared primitive/control/array/callable/dispatch scaffolding,
   never by adding named non-primitive fast paths.
3. Track benchmark deltas in checked-in perf snapshots.
4. Add regressions for the generated code shape responsible for each benchmark
   improvement.

#### Main benchmark families

- recursion/call overhead microbenchmarks
- matrix/array benchmarks
- iterator/container pipeline benchmarks
- map/set/container hot-path benchmarks

#### Definition of done

- Hot compiled benchmark paths no longer carry already-identified avoidable
  scaffolding.
- Performance work continues to respect the lowering spec.

#### Proof

- checked-in perf baselines
- generated-source audits for the hot path under study

### 8. Finish Validation And Merge Gates

#### Goal

Turn the above architecture into a hard release gate instead of a design
aspiration.

#### Required work

1. Keep one authoritative lowering spec and one authoritative completion plan.
2. Maintain generated-source audit tests for carrier, dispatch, control, and
   boundary rules.
3. Maintain full fixture parity across compiled mode and interpreters.
4. Maintain no-bootstrap and no-fallback audits.
5. Maintain reduced benchmark baselines for the main hot families.
6. Use the full repo suite as the final gate once targeted slices are green.

#### Definition of done

The compiler is releasable when:
- static fixtures are native, direct, and boundary-clean
- dynamic fixtures use only documented explicit boundaries
- compiled runtime helpers are direct Go implementations of the required
  semantics
- the benchmark family is materially improved and no longer regresses into known
  scaffolding classes

## Immediate Ordered Queue

This is the concrete next queue derived from the larger work program.

1. Run the compiled release matrix and close any remaining fixture failures
   under no-bootstrap/no-fallback enforcement.
2. Run `./run_all_tests.sh` and close compiler-side regressions.
3. Run `./run_stdlib_tests.sh` and close compiled-mode stdlib regressions.
4. Audit compiled diagnostics/failure behavior for release stability.
5. Confirm reproducible clean-checkout compiler builds and release-gate docs.

## How To Judge Proposed Compiler Changes

A proposed compiler change is correct when the answer to all of these is yes:

1. Does it improve a shared synthesis point instead of introducing a named
   non-primitive fast path?
2. Does it move static semantics farther away from interpreter execution rather
   than closer to it?
3. Does it preserve or improve explicit dynamic-boundary containment?
4. Does it come with generated-source proof and fixture proof?
5. Would the same change automatically help user-defined nominal types, not just
   one stdlib structure?

If the answer to any of those is no, it is probably the wrong fix.
