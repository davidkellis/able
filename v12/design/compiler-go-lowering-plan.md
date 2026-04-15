# Compiler Go Lowering Completion Plan

## Status

This is the ordered execution plan for reaching the lowering architecture defined
in `v12/design/compiler-go-lowering-spec.md`.

Compiler completion status:
- release validation completed on 2026-03-30.
- release validation passed through:
  - `GOFLAGS='-p=1' ./run_all_tests.sh --compiler`
  - `./run_stdlib_tests.sh`
- compiler-native encoding completion closed on 2026-04-14 under the stronger
  finish line that removes staged hybrid carriers from static compiled paths.
- representable static arrays now default to compiler-native specialized
  carriers, with remaining `runtime.ArrayValue` / `ArrayStore*` use limited to
  explicit dynamic/open or ABI edges.
- residual representable union/result/interface carrier-synthesis cleanup is
  complete on 2026-04-14; representable static carrier families now only
  admit `runtime.Value` / `any` at explicit dynamic/open or ABI edges.

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

Compiler release validation is complete, but stronger compiler-native encoding
completion is not:
- large parts of arrays, structs, interfaces, callables, joins, and control
  flow now stay native
- explicit dynamic-boundary enforcement exists
- benchmark work has already removed major scaffolding on hot array/matrix paths
- the array-native lowering tranche is complete on 2026-04-01; remaining
  `runtime.ArrayValue` / `ArrayStore*` use is limited to explicit dynamic or
  ABI edges plus the unspecialized wildcard-array ABI
- imported shadowed nominal typed-pattern bindings now preserve foreign package
  context through carrier reconstruction too, so direct field access stays
  native instead of round-tripping through nominal/runtime helpers
- mixed imported/local shadowed nominal joins now keep distinct native union
  members instead of collapsing on unqualified type-expression strings, and
  shadowed callable joins built from those nominal returns now stay on native
  callable-union carriers instead of widening to `fn(...) -> runtime.Value`
- lambda literals and placeholder lambdas now also narrow through expected
  callable members inside native union/result carriers, and semantic
  `Result` carrier synthesis preserves the callable member's resolved package
  context too, so imported semantic `Option` / `Result` aliases and direct
  union aliases built from shadowed callable returns stay on native callable
  carriers instead of failing `lambda expression type mismatch` or
  `placeholder lambda type mismatch`
- imported selector aliases for generic interfaces now preserve their source
  package through generic type normalization too, so nullable / union / result
  aliases built from foreign `Interface<T>` members keep their native carrier
  even when the caller shadows the same interface name locally
- raw imported selector-alias typed patterns now normalize in the lexical
  caller package first too, so generic `Result` / semantic-result matches on
  imported shadowed interface members resolve back onto that native interface
  carrier instead of widening through `runtime.Value`
- imported semantic `Result` aliases nested under outer `Result` carriers now
  stay on that same native interface carrier too, because alias expansion
  preserves the alias source package and builtin `Error` identities collapse
  across package contexts during nested result flattening
- imported semantic `Result` aliases over shadowed callable members now stay
  native under outer `Result` carriers too, because raw imported selector
  aliases nested inside function type expressions keep lexical caller-package
  normalization instead of being replayed under stale foreign package
  context; proof coverage now pins the same nested outer-result path for
  imported semantic `Option` aliases and imported generic union aliases over
  those shadowed callable members too
- proof coverage now also pins local parameterized union/result aliases with
  imported shadowed interface/callable actuals, so those generic alias locals
  stay on native carriers instead of widening through `runtime.Value` / `any`
- proof coverage now also pins generic specialization over those same
  imported shadowed alias actuals, so specialized helper signatures stay on
  native carriers instead of widening through `runtime.Value` / `any`
- proof coverage now also pins direct shared-mapper lookups and generic
  interface-method dispatch over those same imported shadowed alias actuals,
  so `lowerCarrierTypeInPackage(...)` and existential `pass<T>(...)`-style
  calls stay on native union/result/interface/callable carriers too
- imported generic interface-method calls now keep caller-side explicit type
  arguments on those same native carriers too: generic-method shape
  inference normalizes explicit type arguments in the lexical caller package
  and retries representable-carrier recovery while computing concrete
  param/return helper signatures, so imported `Echo.pass<T>(...)` and
  imported default generic method calls stop widening to `runtime.Value` /
  `any`
- imported generic interface default methods now also resolve nested
  selector-imported members inside those explicit type arguments before
  specializing the interface-package body, so
  `tagged.tagged<Outcome(() -> RemoteThing)>(...)` keeps the concrete native
  `Tagged<...>` carrier and avoids `__able_method_call_node(...)`
- native union synthesis now also retries representable-carrier recovery in
  the member's resolved package before keeping a residual `runtime.Value`
  member, and the imported shadowed interface/callable alias families above
  are now pinned against hidden `_runtime_Value` union variants too
- generic specialization now also retries representable-carrier recovery
  before rejecting a fully bound actual as broad, and the imported shadowed
  specialization matrix is now pinned across result / union alias families
  for both native interface and native callable actuals
- imported shadowed nullable interface/callable aliases now stay on those
  native carriers through generic specialization too, and imported shadowed
  callable union aliases like `Choice (() -> RemoteThing)` now specialize
  through native union helpers instead of falling back through
  `runtime.Value`
- proof coverage now also pins the adjacent broader imported-shadowed
  three-member alias surface, so `Choice3(RemoteReader i32)` and
  `Outcome3(() -> RemoteThing)` already stay on native
  union/interface/callable carriers too with no hidden `_runtime_Value`
  helper variants
- proof coverage now also pins those same broader imported-shadowed
  three-member alias families through imported generic-interface dispatch and
  imported default generic methods, so
  `echo.pass<Choice3(RemoteReader i32)>(...)` and
  `tagged.tagged<Outcome3(() -> RemoteThing)>(...)` stay on native
  union/result/interface/callable carriers too instead of widening helper
  signatures or falling back through `__able_method_call_node(...)`
- proof coverage now also pins those same broader imported-shadowed
  three-member alias families when the generic-interface receiver itself is a
  join-produced existential across concrete implementers, so joined
  `echo.pass<Choice3(RemoteReader i32)>(...)` and joined
  `tagger.tagged<Outcome3(() -> RemoteThing)>(...)` stay on native
  union/result/interface/callable carriers too instead of widening the join
  local or falling back through runtime helpers
- proof coverage now also pins broader outer-result shapes over those same
  imported-shadowed three-member alias families, so
  `!(Choice3(RemoteReader i32))` and `!(Outcome3(() -> RemoteThing))`
  already flatten/collapse to native union/result/interface/callable
  carriers too instead of regressing broader result/error families to
  `runtime.Value` / `any`
- proof coverage now also pins those same broader outer-result families
  through generic interface shape synthesis itself, so parameterized carriers
  like `Keeper(!(Choice3(RemoteReader i32)))` and
  `Keeper(!(Outcome3(() -> RemoteThing)))` keep native
  union/result/interface helper signatures too instead of widening interface
  params/returns to `runtime.Value` / `any`
- proof coverage now also pins the closed local interface existential family
  itself, so local aliases like `Either = (Reader i32) | Echo`,
  `Outcome = !Either`, and `Keeper<Either>` / `Keeper<Outcome>` helper
  synthesis stay on native union/result/interface carriers too instead of
  broadening those local helper params/returns to `runtime.Value` / `any`
- proof coverage now also pins the broader local multi-member
  interface/callable existential family, so local aliases like
  `Choice3 = (Reader i32) | Echo | String`,
  `Outcome3 = Error | (() -> Thing) | String`, generic interface dispatch
  over those same local families, and `Keeper<Choice3>` /
  `Keeper<Outcome3>` helper synthesis stay on native
  union/result/interface/callable carriers too instead of broadening local
  params, locals, or helper signatures to `runtime.Value` / `any`
- proof coverage now also pins the last local analogs of that broader
  family, so joined existential receivers calling
  `Echo.pass<Choice3>(...)` / `Tagger.tagged<Outcome3>(...)` stay on native
  carriers too, and local outer-result helper synthesis like
  `Keeper<!(Choice3)>` / `Keeper<!(Outcome3)>` keeps native
  union/result/interface signatures too instead of widening joined locals or
  helper params/returns to `runtime.Value` / `any`
- native union synthesis now refuses partially native helper families when
  any member still only maps to `runtime.Value` / `any`, so hidden
  `_runtime_Value` variants stop leaking into adjacent imported-shadowed
  specialization slices
- native interface method-shape collection, native interface impl-signature
  synthesis, and native callable signature materialization now also retry
  representable-carrier recovery after raw package-scoped mapping, so
  substituted imported shadowed alias families stay on native
  union/result/interface/callable carriers instead of broadening internally
  to `runtime.Value` / `any`
- imported generic struct members with shadowed nominal arguments now stay on
  specialized native carriers inside result / nested-result / nested
  union-result families too, because selector-imported nominal arguments now
  count as fully bound in the caller package during foreign specialization;
  proof coverage now also pins the same native-carrier behavior through
  generic specialization over imported shadowed generic-struct
  result/union aliases, and imported interface/callable arguments now keep
  those same specialized native `Box<...>` carriers too instead of falling
  back to base `*Box` plus dynamic member calls
- local generic nominal carriers over normalized nullable/union/result members
  are now pinned too, so `Box MaybeReader`, `Box Choice`, `Box Outcome`,
  `Box !(Choice)`, and `Box !(Outcome)` all stay on specialized native
  `Box<...>` carriers instead of falling back to base `*Box`,
  `runtime.Value`, or `any`; this closes the deeper `types.go` /
  `generator_native_unions.go` carrier-synthesis cleanup
- error-wrapped nominal struct typed matches now stay on those native struct
  carriers too, because generated `__able_struct_*_try_from(...)` /
  `__able_struct_*_from(...)` helpers now unwrap through the shared
  `__able_struct_instance(...)` path before enforcing the nominal definition
  check, fixing static `case _: IndexError` matches on array helper bounds
  results under the no-bootstrap boundary harness
- representable nested union/result members now flatten during carrier
  synthesis too, so outer unions like `(!T) | U` and `(A | B) | U`, plus
  direct nested result families like `!!T` and `!(A | B)`, lower to a single
  native union family instead of nesting native-union carriers
- generic specialization also keeps those native interface carriers now:
  `T -> T` on interface actuals stays on the native interface signature, and
  fully bound duplicate generic unions that normalize to the same imported
  shadowed interface collapse to that native interface carrier instead of
  widening through `runtime.Value`, `any`, or a synthetic union wrapper;
  no-op type substitution now preserves the imported selector's resolved
  package context too, so the specialized helper signature stays on that
  same foreign carrier
- interface join carrier selection is now nominal rather than structural, so
  unrelated same-shape interfaces stay distinct and join through a native
  union instead of collapsing onto one interface carrier
- propagated `rescue { case value => ... }` identifier joins now reuse native
  callable plus imported shadowed nominal/interface carriers when the
  monitored call already has a statically known native return type
- higher-order / unknown rescue call failures now stay on the dynamic error
  path instead of reusing the callback return type as a fake failure carrier,
  so handlers like `err.value match { ... }` compile without collapsing `err`
  to `String`
- rescue clause pattern compilation now clears unrelated outer
  `expectedTypeExpr` state before binding the rescued value too, so unresolved
  higher-order rescue handlers do not inherit the enclosing return type
- no-bootstrap bridge fallback stringification for raised non-`Error` values
  now stays aligned with interpreter-visible string output, so compiled rescue
  paths that inspect `err.value` no longer diverge under compiler-only runs
- nested struct-pattern field bindings now restore field-local expected type
  context before recursing into subpatterns, which keeps persistent map/set
  helper patterns on the correct native carriers
- dynamic typed-pattern casts now allocate temps on the caller context rather
  than a cloned probe context, so iterator-end / generator-yield matches stop
  colliding with surrounding temps during codegen
- raised imported shadowed nominal struct literals now prefer the compiler's
  syntax-aware struct-literal type reconstruction during failure inference, so
  propagated rescue joins keep the foreign native struct carrier instead of
  collapsing to a same-named local nominal union
- generated struct `*_from(...)` helpers now keep raw `fieldValue` temporaries
  compile-clean even when a residual unsupported conversion branch does not
  consume them, which cleared the compiled concurrency parity outlier slice
- implicit and explicit return expressions now preserve the declared Able
  return `TypeExpr` while compiling the returned expression too, so generic
  return paths like `unwrap(io_read(...))` keep nullable success carriers
  instead of degrading to the nil-capable Go carrier only
- static nullable typed matches on nil-capable native carriers now guard both
  the non-nil typed branch and the `case nil` branch directly too, so native
  interface and result-family whole-carrier matches no longer emit dead
  unconditional/false conditions ahead of the nil arm
- the nullable typed-match nil-guard relaxation is now limited to actual
  in-scope generic type variables, so concrete `?Interface<T>` /
  `?Result<T>` typed arms regain the required non-nil guard instead of
  compiling as unconditional typed branches
- representable outer unions built from native nullable/result members now
  keep direct inner-member literals and typed clause ordering on native
  carriers too, because nested wrapping is direct and union narrowing only
  removes a member when the pattern exhausts that whole v12 member type
- fully bound duplicate union/result members now collapse to their single
  native carrier during mapping too, so generic specializations like
  `T | String` at `T = String` and `!T` at `T = Error` keep native specialized
  signatures instead of falling back through synthetic union/runtime carriers
- the hard release gates are green:
  - `GOFLAGS='-p=1' ./run_all_tests.sh --compiler`
  - `./run_stdlib_tests.sh`
- in the current dirty tree, the latest timed compiler-gate rerun is green
  again end-to-end; the aggregated observed wall clock across the completed
  timed rerun segments is about `51m34s`
- staged hybrid carrier work remains in the static compiler architecture,
  especially:
  - residual union/result/interface carrier shapes that still retain
    `runtime.Value` / `any` members beyond the desired final host-native end-state
  - transitional mono-array/runtime-store machinery still present in-tree

## Ordered Work Program

This ordered work program is complete and retained as the record of how the
compiler was brought to release shape.

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

Status:
- complete on 2026-03-30.
- final release-blocking fixes in this milestone were:
  - range expression inferred-carrier correction back to shared iterable
    semantics instead of nominal range recoercion
  - concrete wrapped native-interface receiver dispatch preserving overrides
    instead of always selecting default helpers
  - union-aware numeric operator typing for all-numeric unions

## Immediate Ordered Queue

This is the concrete next queue derived from the stronger compiler-native finish
line.

Status:
- array-native lowering tranche complete on 2026-04-01; remaining
  `runtime.ArrayValue` / `ArrayStore*` use is limited to explicit dynamic or
  ABI edges plus the unspecialized wildcard-array ABI.
- mono-array transitional runtime-store scaffolding cleanup complete on
  2026-04-14; compiler-generated mono-array wrappers are now pure slice
  carriers, and only explicit dynamic / ABI boundaries still read or write
  runtime array-store state.
- alias/constraint revalidation closure complete on 2026-04-14; no-interpreter
  generic interface dispatch now performs alias expansion and interface
  constraint checks through compiler-emitted runtime metadata/helpers instead
  of interpreter registries.
- top-level compiler release gates reran green on 2026-04-14.

This stronger compiler-native finish-line queue is now closed. The next active
project queue is the bytecode performance program in `PLAN.md`.

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
