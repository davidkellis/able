# Able v12 Spec TODOs

This list tracks the remaining v12 items after audit; completed work should be removed.

## Parser gaps
- None currently tracked (cast `as` line breaks fixed 2026-02-06).

## Compiler AOT gaps
- Evaluate whether no-interpreter static runtimes should add runtime-independent alias/constraint revalidation for generic interface dispatch, or keep this permanently compile-time-only and document it as final.

## Stdlib externalization gaps
- Confirm and document canonical stdlib resolution contract end-to-end (`able setup`, cache layout, lockfile pins, and `able override` precedence).
- Clarify collision/error semantics when multiple `name: able` roots are visible through `ABLE_MODULE_PATHS`, lockfile sources, or overrides.

## Compiler AOT performance / dynamic-carrier staged limits
- `runtime.Value` usage categories are now documented in `spec/full_spec_v12.md` under the AOT boundary section.
- Native-lowering target is now captured in `v12/design/compiler-native-lowering.md`: static compiled code should primarily manipulate native Go carriers, not interpreter object-model values.
- Control-flow target is now captured in `v12/design/compiler-no-panic-flow-control.md`: compiled control flow should use ordinary Go branches/returns plus explicit control-result signaling, not IIFEs or `panic`/`recover`.
- Desired end-state: compiled polymorphism lowers primarily to host-native mechanisms (Go interfaces/concrete dispatch/generic specialization), with `runtime.Value` used only for explicit dynamic boundaries and residual non-representable cases.
- Native-lowering mandate: static compiled code should represent nominal/user-defined program values with host-native concrete structures (not interpreter object-model carriers) and should never invoke interpreter execution paths unless entering explicit dynamic features.
- Desired container end-state: compiled arrays use native Go array-backed storage on static paths; `runtime.ArrayValue`, `ArrayStore*`, and kernel `storage_handle` are boundary mechanisms only, not the compiler's internal static representation.
- Desired nominal-type end-state: compiled structs remain Go structs/pointers and compiled unions lower to generated Go interfaces plus native variant carriers; `any` is a staged fallback only, not the target union ABI.
- Current staged compiler limit: the in-flight `Array -> Elements []runtime.Value` hybrid plus conversions through `runtime.ArrayValue` / `ArrayStore*` is not an approved final architecture and must be replaced by a true compiler-native array ABI.
- Union-ABI target and bring-up order are now captured in `v12/design/compiler-union-abi.md`; the first code-bearing slice should replace `any` for nullable value carriers before widening to closed nominal unions.
- Current progress note: the native nullable-value slice now covers the
  compiler-native scalar family: `?bool`, `?String`, `?char`, `?f32`, `?f64`,
  `?isize`, `?usize`, `?i8`, `?i16`, `?i32`, `?i64`, `?u8`, `?u16`, `?u32`,
  `?u64`. These now use native Go pointer carriers on compiled static paths,
  with explicit generated boundary helpers for compiled-wrapper/lambda
  conversion, native typed `match` nil/payload checks, and native `or {}` nil
  branching instead of routing those cases through `any`.
- Current progress note: the first closed two-member native union slice is now
  landed for direct `UnionTypeExpression` shapes and named union definitions
  that normalize to the same two-branch native carrier form. Those shapes now
  use generated Go interfaces plus wrapper carrier structs on compiled static
  paths, explicit boundary helpers for wrapper/lambda conversion, and native
  typed `match` branch extraction helpers instead of `__able_try_cast(...)`.
  Static `or {}` now also recognizes native `Error`-implementer failure
  branches on those carriers and keeps the success path native, converting the
  failure payload to `runtime.Value` only at the handler edge when an `err`
  binding is requested. Static `case err: Error => ...` match branches on
  those same carriers now also discriminate natively, converting the matched
  whole error value to `runtime.Value` only at the branch binding edge.
- Current progress note: the first native `!T` slice is now landed for
  `ResultTypeExpression` shapes that normalize to the same
  `runtime.ErrorValue | T` two-member native carrier form. Compiled
  returns/propagation on those shapes now stay on the native carrier, and
  no-bootstrap concrete `Error` implementers derive `runtime.ErrorValue`
  messages through compiled `Error.message()` impls instead of the old
  interpreter-dependent bridge fallback.
- Current progress note: the broader native-union carrier widening tranche is
  now landed for this phase too. Multi-member nominal unions, generic alias
  unions that normalize to native nullable/result carrier families, and
  interface/open unions like `String | Tag` now stay on generated native union
  carriers instead of collapsing the whole union to `any`. When a branch is
  not yet representable as a host-native carrier, it stays explicit as a
  residual `runtime.Value` member inside that generated union carrier rather
  than forcing the entire union path back through `any`.
- Current progress note: plain `Error` type positions now also use the native
  `runtime.ErrorValue` carrier on compiled static paths, which keeps explicit
  `Error` returns and explicit `String | Error` unions off `runtime.Value`.
- Current progress note: `?Error` now also stays native on compiled static
  paths via `*runtime.ErrorValue`, with explicit generated nullable helper
  adapters at wrapper/lambda boundaries and native nullable `match` lowering
  instead of `any`.
- Current progress note: the narrow native `Error` carrier cleanup is now
  complete for direct compiled method use too: `Error.message()` lowers to
  `runtime.ErrorValue.Message`, `Error.cause()` lowers to direct payload
  extraction plus narrow nullable-error coercion, native concrete-error
  normalization preserves both compiled message and cause payloads, and struct
  field conversion now supports `Error` / `?Error` carriers without falling
  back to unsupported-field codegen.
- Current progress note: the compiler now synthesizes a built-in `Array` carrier for
  static lowering, preserves spec-visible metadata fields on the compiled Go
  struct, and lowers several hot static array paths (`literal`, `push`,
  `write_slot`, direct index assignment, `clear`, and static array
  destructuring/rest bindings) to native slice mutations/bindings with metadata
  sync. `match` expressions also no longer blanket-box struct subjects before
  pattern dispatch, so direct compiled `Array` patterns stay native on static
  paths. The generated `Array` boundary helpers now also keep plain
  `runtime.ArrayValue` boundaries handle-free unless a handle already exists or
  a `StructInstanceValue` target explicitly requires storage-handle semantics.
  The residual dynamic array helpers now consistently use the shared runtime
  array unwrapping shim, and current static compiler slices continue to bypass
  them for native `*Array` paths. Reachability tests now also prove the helper
  layer remains available from explicit dynamic package/member/index usage, and
  the error-wrapper bridge for those array bounds paths now preserves concrete
  wrapped struct payloads (for example `IndexError`) before synthesizing an
  anonymous error struct view, so static `case _: IndexError` matches stay
  exhaustive under the zero-boundary harness too.
  Remaining work is to keep shrinking the explicit `runtime.ArrayValue` /
  `ArrayStore*` boundary surface in the residual dynamic helpers and then
  extend the same native strategy to structs/unions.
- Current progress note: unannotated local struct declarations no longer
  default back to `runtime.Value`; static struct field/method tests now assert
  native `*Struct` locals and direct compiled access without extract/writeback
  shims, and targeted compiler coverage now also asserts native direct-call
  param passing, native returns, and mutation-through-call behavior for static
  struct paths. Wrapper returns for native struct/array values now also use
  explicit `__able_struct_*_to` conversion instead of broad `any` conversion.
  Singleton struct boundary converters now also accept runtime
  `StructDefinitionValue` payloads, so interpreted callers can pass bare
  singleton values into compiled native struct/union params without falling
  back to a struct-instance-only boundary.
  The remaining struct work is to extend this native lowering across residual
  dynamic-boundary adapters and any remaining ABI surfaces that still box
  unnecessarily.
- Current staged compiler limit: the whole-union fallback to `any` is no
  longer true for the first native nullable/error/result family, multi-member
  nominal unions, generic alias unions that normalize to those carrier
  families, or interface/open unions that can keep non-native payloads in
  explicit residual `runtime.Value` union members. It still applies to result
  and union shapes that require deeper interface/existential lowering than the
  current residual-carrier strategy provides.
- Narrowing note: the nullable fallback statement above is no longer true for
  the compiler-native scalar family listed above; it still applies to the
  broader remaining union/result surface and to nullable/value-union shapes not
  yet moved onto native carriers.
- Narrowing note: the union fallback statement above is no longer true for the
  first landed closed two-member native union slice; it now also excludes the
  broader carrier-widening tranche for multi-member nominal unions, generic
  alias unions that normalize to native carrier families, and interface/open
  unions with explicit residual `runtime.Value` members. It still applies to
  broader interface/existential lowering beyond that residual-carrier strategy,
  broader result/error shapes beyond the current `runtime.ErrorValue | T`
  slice, and other union surfaces not yet moved onto native carriers.
- Current progress note: compiled static control flow plus explicit dynamic
  call boundaries now propagate explicit control-result signals instead of raw
  panic on the common path. Generated `call_value` / `call_named` helpers now
  return `(runtime.Value, *__ableControl)`, compiled callsites branch on that
  control with ordinary Go conditionals, and callback-boundary failure markers
  stay intact under dynamic callback failures.
- Current progress note: the residual dynamic-helper panic cleanup tranche is
  now complete too. Generated `__able_member_get`, `__able_member_set`,
  `__able_member_get_method`, and `__able_method_call*` helpers now use
  explicit `error` / `*__ableControl` returns instead of raw panic, and the
  temporary recover-based bridge wrappers are gone.
- Current progress note: fully bound object-safe interfaces now lower to
  generated native Go interface carriers plus concrete/runtime adapters across
  static params, returns, typed local assignment, struct fields, direct method
  dispatch, concrete receiver `Index` / `IndexMut`, default-interface method
  calls, concrete `Apply`, wrapper/lambda conversion, and dynamic callback
  boundaries. The strict no-fallback interface fixture audit is green again
  end-to-end, `06_12_26_stdlib_test_harness_reporters` now has a dedicated
  regression harness, runtime adapters now round-trip `void` as `struct{}`
  and write back mutated pointer-backed interface args after runtime dispatch,
  and native interface `*_from_value(...)` helpers now recover concrete
  compiled adapters directly before falling back to the generic runtime
  adapter path.
- Current progress note: the non-object-safe/generic interface existential
  tranche is now landed too. Pure-generic interfaces keep generated native
  carriers instead of collapsing typed locals/params back to `runtime.Value`,
  generic interface/default-interface methods now keep the receiver on that
  native carrier and cross into runtime only at the explicit generic dispatch
  edge, runtime dispatch results convert back into the best-known native Go
  carrier before re-entering compiled code, and the strict interface lookup
  audit is green with total interface/global lookup counts forced to zero.
- Current progress note: the callable/function-type existential tranche is now
  landed too. `FunctionTypeExpression` lowers to generated native callable
  carriers, and direct lambdas, local functions, placeholder lambdas, bound
  method values, function-typed params/fields, wrapper boundaries, and
  interface conversions now stay on those carriers on static compiled paths.
- Current progress note: the strict interface/global lookup audit now defaults
  to four deterministic batch tests so each strict run stays below the repo's
  one-minute per-test target; the unsuffixed
  `TestCompilerInterfaceLookupBypassForStaticFixtures` selector remains
  available for explicit fixture subsets via
  `ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES`.
- Current progress note: allowed dynamic-carrier touchpoints are now
  mechanically enforced by combined-source native-lowering audits plus a
  zero-explicit-boundary fixture audit. Representative static native paths now
  fail tests if they regress to `__able_call_value(...)`,
  `__able_member_get*`, `__able_index*`, `__able_method_call_node(...)`,
  `bridge.MatchType(...)`, `__able_try_cast(...)`, `__able_any_to_value(...)`,
  or panic/IIFE-style control scaffolding, while the residual generic
  interface edge is explicitly audited to stay narrowed to
  `__able_iface_*_to_runtime_value(...)` plus `__able_method_call_node(...)`.
  The zero-explicit-boundary fixture audit now also includes
  `06_08_array_ops_mutability`, so native array mutation/bounds/error handling
  is covered there too.
- Current staged compiler limit: remaining compiler-native work is now a
  different category: broader performance-oriented
  specialization/monomorphization rather than missing callable existentials
  or missing touchpoint enforcement.
- Stage-0 flag scaffolding landed: `--experimental-mono-arrays` and `ABLE_EXPERIMENTAL_MONO_ARRAYS` flow through compiler options; current CLI default is ON with explicit opt-out.
- The mono-array design has now been revised: the earlier typed-runtime-store /
  handle-tag rollout is superseded as the final architecture. Future
  mono-array work must target compiler-generated specialized wrappers over
  native Go slices; `runtime.ArrayValue`, `ArrayStore*`, and runtime typed
  stores are boundary/residual machinery only.
- Historical stage-1 partial remains in-tree behind the flag: Go runtime typed
  stores (`i32`, `i64`, `bool`, `u8`) and some compiler lowering for typed
  literals/index plus `push/len/get/set` intrinsics when static element type
  is known.
- Stage-1 boundary coverage now includes explicit dynamic-call mono-array roundtrip fixtures plus nullable/union/interface callback conversion success/failure fixtures under `--experimental-mono-arrays`.
- Stage-1 index optimization landed: array read/write/get/set lowering now keeps native integer indices as native `int` where safe instead of boxing through `bridge.ToInt` + `bridge.AsInt`.
- Stage-1 propagation/cast optimization landed: mono typed index propagation paths now avoid boxing `i32` reads into `runtime.Value` when a native widening cast is semantically safe (e.g., `i32 -> i64`).
- Stage-1 compatibility fixes landed:
  - `Array` struct converters now accept/synchronize raw `*runtime.ArrayValue` carriers at explicit runtime boundaries.
  - Interface-annotated local assignment now enforces interface coercion via `bridge.MatchType`, preserving interface args for compiled dispatch.
- Stage-1 strict sweep status (2026-02-26): compiler strict fixture audits and `TestCompilerDynamicBoundary*` are green.
- Stage-1 perf snapshot (compiled-only, 5-run avg, 2026-02-26, post compatibility fixes): `bench/noop` default `0.062s` / `3.20` GC vs mono `0.060s` / `3.20`; `bench/sieve_count` default `0.072s` / `5.40` vs mono `0.074s` / `5.20`; `bench/sieve_full` default `0.164s` / `23.20` vs mono `0.164s` / `23.00`.
- Guardrail landed: `TestCompilerExperimentalMonoArraysStaticBodyStaysOnCompilerOwnedArrayCarrier` now proves that enabling `ExperimentalMonoArrays` still keeps representative static array bodies on the compiler-owned array carrier instead of regressing to `runtime.ArrayValue`, `ArrayStore*`, or dynamic helper dispatch.
- Stage-1 specialized-wrapper slice is now landed for the first staged element
  set on explicit typed positions: `Array i32`, `Array i64`, `Array bool`, and
  `Array u8` now lower to compiler-owned wrappers over native Go slices
  (`*__able_array_i32`, `*__able_array_i64`, `*__able_array_bool`,
  `*__able_array_u8`) when `--experimental-mono-arrays` is enabled.
- The direct typed mono-array hot path now covers typed literals, `push`,
  `get`, `set`, `read_slot`, `write_slot`, direct `arr[idx]`, direct
  `arr[idx] = value`, and explicit wrapper/lambda/interface/union/struct
  boundary conversion on those staged wrappers.
- Stage-1 widening slice is now landed too: non-empty unannotated local array
  literals infer staged specialized carriers, `Array.new()` /
  `Array.with_capacity()` lower directly to compiler-owned static carriers on
  typed static paths, `reserve()` / `clone_shallow()` stay specialized, static
  array `for` loops iterate directly over typed slices, and array-pattern rest
  tails preserve specialized carriers instead of dropping back to generic
  `*Array`.
- Stage-1 compiled remeasurement snapshot (2026-03-19, 5-run averages via
  `v12/bench_perf`, compiled mode built through `cmd/ablec`): `bench/noop`
  mono on `0.0100s` / `0.00` GC vs mono off `0.0100s` / `0.00`;
  `bench/sieve_count` mono on `0.0100s` / `0.00` GC vs mono off `0.0100s` /
  `0.00`; `bench/sieve_full` mono on `0.0200s` / `1.00` GC vs mono off
  `0.0200s` / `3.00`.
- Residual generic static-array narrowing landed too: when a static compiled
  array binding still carries a recoverable element type, generic helper
  results such as `get`, `pop`, `first`, `last`, and `read_slot` now prefer
  native nullable carriers instead of dropping back to `runtime.Value`.
- The staged specialized wrapper set now includes `f64` too:
  `Array f64 -> *__able_array_f64`, with explicit wrapper/runtime boundary
  helpers and dynamic-boundary callback coverage.
- Native nullable propagation now handles concrete expected-type coercion for
  pointer-backed carriers such as `*float64`, which keeps nested static
  `rows.get(j)!.get(i)!` expressions on the compiler-native path instead of
  falling back to dynamic method dispatch.
- Static built-in `Array` scalar propagation now returns concrete success
  element types directly on the compiled path, and primitive numeric casts
  such as `i32 -> f64` now lower to direct Go casts. The reduced matrix
  benchmark `v12/fixtures/bench/matrixmultiply_f64_small/main.able` now
  measures `1.9733s` / `7.00` GC over 3 compiled runs, down from the earlier
  `5.7233s` / `252.00` GC post-outer-wrapper snapshot.
- Shared primitive `float -> int` casts now also lower natively on static
  compiled paths through truncate/range/overflow checks instead of
  `__able_cast(...)` / `bridge.AsInt(...)`; the latest reduced matrix
  snapshot for `v12/fixtures/bench/matrixmultiply_f64_small/main.able` is now
  `1.7567s` / `7.00` GC.
- Shared static built-in `Array` factories and intrinsics now lower without
  synthetic `__able_push_call_frame(...)` / `__able_pop_call_frame()`
  scaffolding on compiled static paths. The latest reduced matrix snapshot for
  `v12/fixtures/bench/matrixmultiply_f64_small/main.able` is now `0.1933s` /
  `7.00` GC, and the full compiled
  `v12/examples/benchmarks/matrixmultiply.able` benchmark now completes in
  `4.2267s` / `13.00` GC over 3 runs instead of timing out.
- A reduced checked-in compiler benchmark target now exists for the staged
  `f64` slice: `v12/fixtures/bench/matrixmultiply_f64_small/main.able`.
  Current compiled 3-run averages on that target are:
  mono on `5.4833s` / `280.00` GC vs mono off `45.3133s` / `3568.67` GC.
- Nested typed outer rows are now specialized too: `Array (Array f64)` lowers
  to a compiler-owned outer wrapper (`*__able_array_array_f64`) over
  `[]*__able_array_f64` instead of the generic `*Array` / `[]runtime.Value`
  shell.
- Compiler-owned array wrapper synthesis now also covers broader native
  carrier element families beyond nested mono arrays, including generic inner
  arrays, native interface carriers, and native callable carriers.
- The staged specialized text scalar family is now landed too:
  `Array char -> *__able_array_char` over `[]rune`,
  `Array String -> *__able_array_String` over `[]string`, and
  `Array (Array char) -> *__able_array_array_char` on the representative
  nested text-row path.
- Native result propagation for specialized text arrays is closed too:
  `!Array char` now re-wraps native success branches through the static
  coercion path instead of incorrectly routing specialized carriers back
  through `_from_value(__able_runtime, ...)`.
- Carrier-array wrappers for already-native compiler carriers now remain
  available even when staged scalar mono arrays are disabled, which closes the
  mono-off nested-text-row identity bug on `Array (Array char)`.
- Reduced checked-in text benchmark target:
  `v12/fixtures/bench/zigzag_char_small/main.able`.
  Corrected compiled 3-run averages are mono on `0.9567s` / `88.00` GC vs
  mono off `1.0500s` / `384.00` GC.
- The staged specialized primitive numeric family is now broader too:
  `Array i8`, `Array i16`, `Array u16`, `Array u32`, `Array u64`,
  `Array isize`, `Array usize`, and `Array f32` now lower to compiler-owned
  wrappers when `ExperimentalMonoArrays` is enabled.
- Reduced checked-in unsigned benchmark target:
  `v12/fixtures/bench/sum_u32_small/main.able`.
  Current compiled 3-run averages are mono on `1.0933s` / `185.33` GC vs
  mono off `1.6800s` / `21.33` GC.
- Post-outer-wrapper snapshot (`v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`):
  `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  mono on `5.7233s` / `252.00` GC vs mono off `44.5167s` / `3550.67` GC.
- Array `set` / index-assignment parity is restored across static and residual
  runtime-backed paths: success returns `nil`, failure remains `IndexError`.
- Runtime-backed iterator interface carriers now accept raw
  `*runtime.IteratorValue` payloads directly, and generator stop is preserved
  as iterator completion through `__able_control_from_error_with_node(...)`
  instead of being surfaced as a generic runtime error. That closes the
  earlier `06_12_18_stdlib_collections_array_range` compiler/runtime mismatch.
- Staged Go runtime/compiler limit: the old runtime-backed mono-array experiment
  still exists in-tree, and current CLI flag plumbing remains default-on, but
  that hybrid path is not the accepted end state. The widened specialized slice
  reduced timed GC on `bench/sieve_full`, but did not yet move wall-clock time
  on the staged compiled trio; the residual generic array mismatches that were
  blocking `06_08_array_ops_mutability`, `06_12_02_stdlib_array_helpers`, and
  `06_12_18_stdlib_collections_array_range` are now closed, and nested outer
  row carriers are specialized too, and broader native carrier-array families
  now stay compiler-owned as well, and the text scalar family now stays on
  specialized wrappers too. The remaining mono-array work is now primarily
  performance work plus broader container lowering: `HashMap`, `TreeMap`, and
  `PersistentMap` families now stay on native compiler carriers on static
  compiled paths, but broader container families and deeper generic container
  paths still need the same treatment and measurement.
- Staged Go compiler note: callable/function-type existential surfaces no
  longer default to dynamic carrier values on static compiled paths; residual
  dynamic carrier use should now be limited to explicit dynamic boundaries,
  open runtime-polymorphic dispatch, and other semantically necessary ABI
  edges.
- Staged Go compiler note: `HashMap K V` now lowers to native `*HashMap`
  carriers on static compiled paths, typed/untyped map literals and
  `Array (HashMap K V)` shells stay native in generated code, and the old
  `HashSet.iterator()` / `Iterator T` generic interface return fallback is now
  replaced by an explicit narrowed runtime adapter roundtrip instead of
  compiler fallback.
- Staged Go compiler note: generic nominal struct lowering now expands simple
  type aliases before host-type mapping, so alias-backed fields like
  `HashMapHandle = i64` lower through the same native carrier path instead of
  regressing those fields to `runtime.Value`; the explicit map-literal handle
  boundary and compiled exit-signal bridge have been tightened to match that
  contract.
- Staged Go compiler note: shared generic default-method lowering now has an
  explicit user-defined nominal proof case too: `Iterator.collect<C>()`
  stays on the compiled `Default + Extend` path for a user accumulator struct
  (`SumCount`) without introducing another named-container rule. The
  residual specialized collect helper remains only for the built-in `Array`
  exception and only as a fallback behind that shared generic path.
- Staged Go compiler note: compiled static bodies should not emit bare Go
  builtin calls where Able user bindings can shadow them. Current slice/string
  static lowering now routes through generated helpers
  (`__able_slice_len`, `__able_slice_cap`, `__able_string_len_bytes`) so
  container-heavy nominal code like `TreeMap` / `PersistentMap` no longer
  fails due to Go builtin shadowing.
- Staged Go compiler note: the next stdlib container family slice is now
  mechanically audited on the same shared native-carrier path too.
  Representative compiled methods for `Deque` / `Queue`, `BitSet` / `Heap`,
  and `PersistentSortedSet` / `PersistentQueue` now have no-fallback
  regressions proving they stay on native locals and avoid dynamic helper
  regressions; reduced benchmark target
  `v12/fixtures/bench/heap_i32_small/main.able` now has a follow-up shared
  generic nominal-method specialization snapshot at `4.2000s` / `1811.67` GC
  over 3 compiled runs.
- Staged Go compiler note: generic nominal `methods` blocks now specialize on
  statically known concrete targets too, so ordinary user-defined generic
  nominal types (for example `Box T`) and stdlib generic nominal types
  (for example `Heap T`) can render concrete compiled method signatures
  without introducing another named-structure lowering rule. The next shared
  gap is no longer nominal-method signature specialization; bound generic
  field/member carrier refinement inside those already-specialized method
  bodies is now closed too.
- Staged Go compiler note: fully bound generic struct fields/members now stay
  on their concrete native carriers inside specialized nominal method bodies.
  A user-defined `Bucket T { items: Array T }` proof case now pins
  `Items *__able_array_i32` plus specialized `Bucket.push` / `Bucket.second`
  bodies under `ExperimentalMonoArrays`, the mono-array `Iterable.iterator` /
  `Iterable.each` execute gap is closed, and the reduced
  `v12/fixtures/bench/heap_i32_small/main.able` benchmark now has a follow-up
  snapshot at `0.7667s` / `91.33` GC over 3 compiled runs.
- Staged Go compiler note: the remaining shared generic nominal
  default/static receiver and struct-literal refinement gap on the reduced
  `LinkedList -> Enumerable -> LazySeq` family is now closed too. Recursive
  type substitution now resolves chained bindings transitively, static nominal
  target refinement upgrades `LazySeq { ... }` to concrete specialized
  carriers like `LazySeq<i32>` when the expected type is known, and native
  interface adapter synthesis now matches specialized concrete receivers like
  `*LinkedList_i32` through the shared target-template path instead of
  falling back.
- Staged Go compiler note: the next deeper generic-container correctness
  slice is now closed too. Generic nullable/interface carriers like
  `LazySeq.Source: ?(Iterator T)` now stay on generated native carriers
  instead of collapsing to `any`, and compiled nil lowering now emits typed Go
  nils for native nilable carriers (`(*ListNode)(nil)`,
  `__able_iface_Iterator_T(nil)`, etc.) instead of invalid raw `nil` temps.
  The compiled fixture gate
  `06_12_14_stdlib_collections_linked_list_lazy_seq` is green again on that
  shared native path.
- Staged Go compiler note: the first benchmark-worthy generic-container hot
  path is now closed too. The compiled `LinkedList -> Iterable -> Iterator`
  path stays native because inherited interface impls now synthesize native
  base-interface adapters and concrete native interface adapters directly
  coerce compatible native interface return carriers instead of round-tripping
  through runtime values.
- Staged Go compiler note: the next concrete generic/default container-method
  hot path is now closed too. Higher-kinded self patterns like
  `Enumerable A for C _` now bind `C` to the concrete target on compiled impl
  paths, bound type-constructor calls like `C.default()` resolve statically,
  and native `Iterator<T>` carriers now satisfy compiled iterable lowering
  directly inside those default impl bodies.
- Staged Go compiler note: the callback/runtime-value carrier cleanup slice on
  that same concrete default-impl hot path is now closed too. Specialized impl
  functions now retain bound generic type bindings through compileability and
  render, default-impl sibling calls prefer specialized sibling impls before
  the ordinary concrete receiver path, and the compiled
  `LinkedList.map/filter/reduce` benchmark no longer overflows by bridging
  `Iterator_A -> runtime.Value -> Iterator_i32` through cyclic
  `LinkedListIterator` / `ListNode` conversion.
- Staged Go compiler note: the same reduced
  `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` benchmark
  now has a follow-up shared static nominal closure snapshot at `0.1633s` /
  `8.33` GC over 3 compiled runs, with direct compiled output parity at
  `382455000`.
- Staged Go compiler note: the next iterator default-method hot path is now
  closed too. Ordinary default native-interface methods now lower to direct
  compiled helpers on native iterator carriers, so the representative compiled
  `LinkedList.lazy().map<i64>(...).filter(...).next()` path stays native
  instead of re-entering the runtime adapter method layer.
- Staged Go compiler note: the mono-array-enabled `Iterator.collect<Array T>()`
  follow-up is now closed too. That path now lowers through a generated
  compiled helper with a specialized mono-array accumulator instead of the old
  residual dynamic bridge.
- Staged Go compiler note: the iterator-literal controller/runtime-value edge
  is now closed too. Compiled iterator literals bind `gen` as a
  compiler-owned `*__able_generator`, `gen.yield(...)` / `gen.stop()` and
  bound `gen.yield` callables lower directly, and native nilable/static
  carrier conditions now use direct nil checks, which keeps
  `Iterator.filter_map` on the static path.
- Current staged Go compiler limit on this family: the next work is no longer
  iterator-literal/controller cleanup; it is the next hot
  generic-container/runtime edge beyond the now-closed
  `map/filter/filter_map/collect` iterator family.
- Pending workstream: implement the revised monomorphized container ABI
  (`Array<T>` specialized compiler-owned wrappers over native Go slices) under
  the existing gated rollout plumbing; accepted proposal captured in
  `v12/design/monomorphized-container-abi.md`.
- Pending workstream: broaden native lowering beyond arrays (struct/union/interface-call-site specialization where statically representable) and add regression guards that fail when new static paths regress to dynamic carrier helpers.
