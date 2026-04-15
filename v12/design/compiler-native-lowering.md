# Compiler Native Lowering

## Status

Active design constraint for the v12 compiler. This document records the target
architecture and supersedes ad hoc decisions that keep static compiled code too
close to interpreter/runtime object-model carriers.

The exhaustive lowering map now lives in `v12/design/compiler-go-lowering-spec.md`.
The ordered execution plan now lives in `v12/design/compiler-go-lowering-plan.md`.
This document remains the short-form contract and guardrail summary.

Milestone status:
- compiler `PLAN.md` Milestone 1 is complete.
- the compiler now has explicit shared lowering entrypoints for:
  - type normalization and carrier synthesis
  - join/pattern synthesis
  - dispatch synthesis
  - control-envelope synthesis
  - boundary-adapter synthesis
- source-audit enforcement for those entrypoints lives in
  `v12/interpreters/go/pkg/compiler/compiler_lowering_facade_audit_test.go`.

Named stdlib/container examples in this document are proof cases for shared
lowering machinery, not architecture exceptions.

## Vision

The compiler should lower Able programs to native Go constructs whenever the
semantics are statically representable.

The end state is not "compiled code that still manipulates interpreter values
more efficiently." The end state is "compiled code that primarily manipulates
Go-native values and uses dynamic carriers only at explicit dynamic
boundaries."

## Completion Milestones

The compiler is not "done" when the current local fallback is smaller. It is
"done" when these high-level conditions are true:

1. Native carrier completeness:
   every statically representable Able type expression lowers to a native Go
   carrier instead of `runtime.Value` / `any`.
2. Native pattern/control-flow completeness:
   `if`, `match`, `rescue`, `or {}`, `loop`, and `breakpoint` keep native
   carriers end-to-end whenever branch and pattern types are statically
   representable.
3. Native dispatch completeness:
   statically resolved field/method/interface/index/call sites compile to
   direct Go dispatch instead of runtime helper dispatch.
4. Boundary containment completeness:
   residual runtime carriers remain only at explicit dynamic language or ABI
   boundaries, and audits fail if they leak back into static paths.
5. Performance completeness:
   the benchmarked compiled path no longer carries known avoidable runtime
   scaffolding on hot static code.

## Non-Negotiable Constraints

### Arrays

- Able arrays on static compiled paths must be represented in terms of native Go
  array-backed storage.
- Acceptable compiled representations are Go slices or compiler-owned Go wrapper
  structs around slices.
- Unacceptable compiled representations on static paths are `runtime.ArrayValue`,
  `runtime.ArrayStore*`, kernel `storage_handle`, or any other interpreter-owned
  structural carrier.
- If a compiled array must cross into explicit dynamic behavior, the compiler
  may generate a narrow adapter at that boundary, but the boundary form is not
  the static representation.

### Structs

- Able structs must lower to native Go structs or pointers to Go structs.
- Compiled locals, fields, params, and returns must remain native unless they
  are crossing an explicit dynamic boundary.
- The compiler must not auto-box struct values into `runtime.Value` just to
  preserve identity or make dispatch easier.

### Unions

- Able unions must lower to generated Go interfaces plus native variant
  carriers.
- `any` may be a temporary escape hatch during bring-up, but it is not the
  target representation for statically compiled union values.
- Pattern matching on unions should compile to native Go type checks / type
  switches over those generated interfaces and variants.

### Flow Control

- Compiled control flow must be expressed with ordinary Go conditionals,
  branches, loops, and returns.
- IIFEs are not part of the target architecture.
- `panic` / `recover` must not be used for ordinary Able control flow.
- Exceptions and non-local jumps should be represented as explicit return
  signals that the caller can inspect and propagate with normal Go logic.

### Boundary Discipline

- `runtime.Value`, `runtime.ArrayValue`, `runtime.StructInstanceValue`, bridge
  dispatch helpers, and interpreter-backed facilities are allowed only at
  explicit dynamic boundaries, extern ABI edges, and truly residual unsupported
  cases that are documented.
- Crossing a dynamic boundary should be narrow and explicit:
  - native compiled value -> boundary adapter
  - dynamic work
  - boundary adapter -> native compiled value
- The final explicit dynamic-boundary entry helper set for the current Go AOT
  compiler is:
  - `__able_call_value(...)`
  - `__able_call_named(...)`
  - generated `call_original` wrappers
- Generated carrier adapters such as `*_runtime_value`, `*_from`, and `*_to`
  are allowed only immediately adjacent to one of those explicit entries or an
  extern/host ABI edge; they are not permitted as a general static lowering
  substrate.

## Target Representation Map

| Able construct | Target compiled representation |
| --- | --- |
| primitives | native Go scalars |
| structs | native Go structs / pointers |
| unions | generated Go interfaces + native variant carriers |
| nullable struct-like values | nil-capable Go pointers or wrappers |
| arrays | native Go slices or compiler-owned slice wrappers |
| functions | native Go functions where representable |
| dynamic features | explicit boundary adapters using runtime/interpreter values |

## Control Signal Model

The compiler should move toward explicit control envelopes for non-local flow.

Representative shape:

```go
type __ableControlKind uint8

const (
    __ableControlNone __ableControlKind = iota
    __ableControlReturn
    __ableControlBreak
    __ableControlContinue
    __ableControlRaise
)

type __ableControl struct {
    Kind  __ableControlKind
    Label string
    Value any
}
```

Generated helpers and lowered subroutines should return enough information for
callers to distinguish:

- normal completion
- function return
- loop/breakpoint jump
- raised exception

The exact concrete shape may be specialized per generated helper or per static
type, but the architectural rule is fixed: propagate control with return values,
not `panic` / `recover`.

## Current Violations To Remove

- Overriding the kernel `Array` shape in compiler type mapping to
  `Elements []runtime.Value`.
- Converting compiled arrays through `runtime.ArrayValue` /
  `runtime.ArrayStore*` on normal static paths.
- Boxing struct locals into `runtime.Value` to preserve identity.
- Lowering unions to `any` or `runtime.Value` instead of generated Go
  interfaces.
- Using IIFEs to manufacture expression contexts.
- Using `panic` / `recover` for non-local jump or exception propagation in
  normal compiled flow.

## Execution Plan

1. Remove static-path array/runtime hybrids and define the compiler-native array
   ABI.
2. Keep struct locals native and repair static dispatch around native carriers.
3. Define the native union ABI and pattern-match lowering strategy.
4. Replace panic/recover and IIFE-based control lowering with explicit control
   return propagation.
5. Add regression audits that fail when static compiled code regresses back to
   interpreter structural carriers or panic-based flow control.

## Progress Snapshot

- Static compiled arrays now use a compiler-owned `Array` carrier with
  spec-visible metadata fields and hidden native slice storage on direct
  compiled paths.
- Static array literal/mutation lowering is already native for hot paths such
  as literal construction, `push`, `write_slot`, direct index assignment, and
  `clear`.
- Static array destructuring is now native for both `match` expressions and
  pattern assignments: rest tails lower to native compiled `*Array` values
  instead of `runtime.ArrayValue`.
- `match` expressions no longer blanket-convert struct subjects to
  `runtime.Value` before pattern dispatch; native struct/array subjects now
  stay native until an explicit dynamic boundary is entered.
- The generated `Array` boundary helpers now keep plain `runtime.ArrayValue`
  boundaries handle-free unless a handle is already present or a
  `StructInstanceValue` target explicitly requires storage-handle semantics.
- Remaining array work is boundary tightening, not reintroducing interpreter
  carriers onto direct compiled paths.
- The first native union slice is now landed for closed two-member carriers,
  and the first native `!T` slice now rides on that same generated carrier
  family for `runtime.ErrorValue | T` result shapes.
- The broader native-union carrier widening tranche is now landed too:
  multi-member nominal unions, generic alias unions that normalize to native
  nullable/result carriers, and interface/open unions with explicit residual
  `runtime.Value` members now stay on generated native union carriers instead
  of collapsing the whole union to `any`.
- Fully bound object-safe interfaces now also stay native on compiled static
  paths: the compiler emits generated Go interface carriers plus concrete and
  runtime adapters, and static params/returns, typed local assignment, struct
  fields, direct method dispatch, wrapper/lambda conversion, and dynamic
  callback boundaries now use those carriers instead of raw `runtime.Value`.
- The object-safe tranche is now closed under the strict no-fallback fixture
  audit too: the full interface audit is green again, the reporters fixture
  has a dedicated regression harness, runtime adapters round-trip `void` as
  `struct{}`, and pointer-backed native interface args are written back after
  runtime-backed dispatch.
- Native interface adapter population is now refreshed against the current impl
  set before reuse, which keeps no-fallback typed interface assignment and
  static interface-heavy fixtures on the native path instead of failing on a
  stale cached adapter view.
- No-bootstrap compiled result paths now derive concrete `Error` messages
  through compiled `Error.message()` impls before constructing
  `runtime.ErrorValue`, avoiding the old interpreter-dependent bridge fallback
  on those static result paths.
- Direct compiled `Error.message()` / `Error.cause()` calls now also stay on
  the native `runtime.ErrorValue` carrier, native concrete-error normalization
  preserves compiled cause payloads as well as messages, and struct field
  conversion now supports `Error` / `?Error` carriers without falling back to
  unsupported-field codegen.
- Explicit dynamic call boundaries now also participate in the control-result
  model: generated `__able_call_value(...)` / `__able_call_named(...)`
  helpers return `(runtime.Value, *__ableControl)`, compiled callsites branch
  on that control with ordinary Go conditionals, and callback-bearing runtime
  helpers convert control back into ordinary Go `error` returns.
- The residual dynamic-helper panic cleanup tranche is now complete too:
  generated `__able_member_get`, `__able_member_set`,
  `__able_member_get_method`, and `__able_method_call*` helpers now use
  explicit `error` / `*__ableControl` returns, and the temporary recover-based
  bridge wrappers are gone.
- Singleton struct boundary converters now also accept runtime
  `StructDefinitionValue` payloads, which keeps interpreted bare-singleton
  arguments compatible with compiled native struct/union params.
- The non-object-safe/generic interface existential tranche is now closed too:
  pure-generic interfaces keep generated native carriers instead of
  collapsing typed locals/params back to `runtime.Value`, and generic
  interface/default-interface methods now keep the receiver on that native
  carrier through generated compiled dispatch helpers.
- For statically-known generic call shapes, those helpers switch on the native
  interface carrier and call specialized compiled impls directly instead of
  converting the receiver to runtime for `__able_method_call_node(...)`.
- Cross-package generic-only interface adapters now survive shared adapter
  refresh and late helper generation too: the compiler tracks explicitly
  required concrete adapters separately from the refresh cache and emits
  concrete adapter types/helpers to a fixed point during render, which closes
  the old `Tokenizer <- Prefixer` missing-helper build hole without adding
  nominal-type-specific lowering rules.
- The only remaining runtime crossing on that surface is the explicit
  runtime-adapter case inside the generated helper, which is the expected
  dynamic boundary for interface values that already originate from runtime
  payloads.
- The callable/function-type existential tranche is now landed too:
  `FunctionTypeExpression` lowers to generated native Go callable carriers,
  lambdas/local functions/placeholder lambdas/bound method values stay on
  those carriers on static paths, wrapper/interface/struct-field conversion is
  explicit, and callable-heavy generic surfaces like
  `Iterator<T>.map/filter/filter_map/collect` now run on the narrowed
  receiver-plus-callable carrier design instead of broad dynamic helper
  fallback.
- The strict interface/global lookup audit no longer relies on one oversized
  default sweep: the default fixture set is now split across four
  deterministic top-level batch tests so each strict audit run stays under the
  repository's one-minute per-test target, while the unsuffixed selector test
  remains available for explicit fixture subsets via
  `ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES`.
- Allowed dynamic-carrier touchpoints are now also mechanically enforced:
  combined-source audits fail if representative static native paths regress to
  `__able_call_value(...)`, `__able_member_get*`, `__able_index*`,
  `__able_method_call_node(...)`, `bridge.MatchType(...)`,
  `__able_try_cast(...)`, `__able_any_to_value(...)`, or panic/IIFE-style
  control scaffolding, and representative static fixtures now execute under
  the boundary-marker harness with both fallback and explicit boundary counts
  at zero. Statically-known generic interface calls are now covered by that
  same fully-native expectation rather than a narrowed residual-runtime audit
  exception.
- The broader compiler re-audit tranche is now closed too: the surfaced
  native array mismatch under the zero-boundary harness on
  `06_12_02_stdlib_array_helpers` was a two-part error-wrapped nominal-struct
  gap. `__able_error_to_struct(...)` already preserves concrete wrapped struct
  payloads before falling back to the synthetic error view, and the remaining
  shared nominal-converter gap is now closed too: generated
  `__able_struct_*_try_from(...)` / `__able_struct_*_from(...)` helpers now
  unwrap through `__able_struct_instance(current)` after interface unwrapping
  before enforcing the nominal definition check. That keeps static
  `case _: IndexError` matches exhaustive on array bounds results under the
  zero-boundary fixture audit.
- The mono-array design tranche is now revised too: the older typed-runtime-
  store / handle-tag rollout has been superseded as the target architecture.
  Future `Array<T>` specialization work must converge on compiler-owned
  specialized wrappers over native Go slices, and a dedicated compiler audit
  now proves that enabling `ExperimentalMonoArrays` still keeps representative
  static array bodies on the compiler-owned array carrier today.
- The first specialized-wrapper slice is now landed too for explicit typed
  `Array i32` / `Array i64` / `Array bool` / `Array u8` positions: those types
  map to compiler-owned wrappers over native typed Go slices, direct typed
  literals and hot index/intrinsic paths operate on those wrappers, and the
  explicit mono-array dynamic-boundary suite stays green on the specialized
  carriers.
- The widened mono-array slice is now landed too: non-empty unannotated local
  array literals infer staged specialized carriers, `Array.new()` /
  `Array.with_capacity()` lower directly to compiler-owned static carriers on
  typed static paths, `reserve()` / `clone_shallow()` stay specialized, static
  array `for` loops iterate directly over typed slices, and array-pattern rest
  tails preserve specialized carriers instead of dropping back to generic
  `*Array`.
- The widened mono-array slice has now also been re-measured on the staged
  compiled fixtures (`bench/noop`, `bench/sieve_count`, `bench/sieve_full`):
  wall-clock stayed flat, but the heaviest staged array fixture reduced timed
  GC (`1.00` vs `3.00` with mono arrays disabled).
- The residual generic-array narrowing tranche is now closed too: static array
  bindings that still carry a recoverable element type now keep generic helper
  results (`get`, `pop`, `first`, `last`, `read_slot`) on native nullable
  carriers where representable; static and residual runtime-backed array `set`
  / index-assignment success is back in parity at `nil`; runtime-backed
  iterator interface carriers now accept raw `*runtime.IteratorValue` payloads
  directly; and `__able_control_from_error_with_node(...)` now preserves
  `__able_generator_stop` as iterator completion instead of surfacing it as a
  generic runtime error.
- The staged specialized wrapper set now includes `f64`, and nested
  `Array (Array f64)` hot loops now stay on the static compiler path too:
  native nullable propagation now supports concrete expected-type coercion for
  pointer-backed scalar carriers like `*float64`, which keeps
  `rows.get(j)!.get(i)!` from regressing the surrounding `push(...)` back to
  `__able_method_call_node(...)`.
- The full compiled `examples/benchmarks/matrixmultiply.able` path is back in
  parity too: under the existing `60s` `bench_perf` budget it now times out in
  both mono-on and mono-off modes instead of failing early with
  `runtime: runtime error` on the mono-array path.
- A reduced checked-in benchmark fixture now measures the staged `f64` slice
  directly:
  - `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  - compiled 3-run averages: mono on `5.4833s`, `280.00` GC; mono off
    `45.3133s`, `3568.67` GC.
- The remaining generic outer-row shell is now gone on the representative
  nested staged path too: `Array (Array f64)` lowers to
  `*__able_array_array_f64` over `[]*__able_array_f64`, rendered mono-array
  converters/native `Array` helpers handle pointer-backed specialized
  elements directly, and the post-outer-wrapper reduced benchmark snapshot is
  `5.7233s` / `252.00` GC with mono arrays on versus `44.5167s` /
  `3550.67` GC with mono arrays off.
- Compiler-owned array wrapper synthesis now covers broader native carrier
  element families too: generic inner arrays, native interfaces, native
  callables, and other representable pointer-backed carriers now lower to
  dedicated outer wrappers instead of falling back to generic `*Array` /
  `runtime.Value` element storage, and dynamic-boundary callback coverage now
  includes interface-carrier and callable-carrier arrays explicitly.
- The staged text scalar family is now specialized too:
  - `Array char` -> `*__able_array_char` over `[]rune`
  - `Array String` -> `*__able_array_String` over `[]string`
  - representative nested text rows (`Array (Array char)`) now lower to
    `*__able_array_array_char`
  - native result propagation now keeps `!Array char` on the static carrier
    path instead of routing its success branch back through
    `_from_value(__able_runtime, ...)`
- The first focused text benchmark now exists too:
  - `v12/fixtures/bench/zigzag_char_small/main.able`
  - corrected compiled 3-run averages: mono on `0.9567s`, `88.00` GC; mono
    off `1.0500s`, `384.00` GC
  - the earlier mono-off comparison was invalid because nested
    `Array (Array char)` rows were losing mutation identity when the outer
    array fell back to generic `*Array` / `runtime.Value` storage
- Carrier-array wrappers for already-native compiler carriers now stay
  available even with `ExperimentalMonoArrays` disabled, which preserves
  nested carrier identity on mono-off paths while still leaving staged scalar
  specialization itself behind the flag.
- The next shared built-in `Array` scalar-lowering step is now landed too:
  - static staged-array propagation returns concrete success element types on
    the compiled path instead of routing scalar success values back through
    `__able_nullable_*_to_value(...)`;
  - primitive numeric casts like `i32 -> f64` now lower directly to Go casts
    on static compiled paths instead of `__able_cast(...)`;
  - shared primitive `float -> int` casts now also lower natively through
    truncate/range/overflow checks instead of `__able_cast(...)` and
    `bridge.AsInt(...)`, which removes the last obvious runtime cast helper
    use from the full `matrixmultiply` entry path;
  - shared static built-in `Array` factories and intrinsics now lower without
    synthetic `__able_push_call_frame(...)` / `__able_pop_call_frame()`
    scaffolding on compiled static paths;
  - propagated static built-in `Array` accessors (`get`, `first`, `last`,
    `read_slot`, `pop`) now lower as direct bounds-check + element-load paths
    with nil control transfer instead of manufacturing pointer-backed nullable
    carriers on the success path;
  - the reduced matrix benchmark
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able` still measures in
    the same range at `0.1967s` / `7.33` GC over 3 compiled runs;
  - the full macro benchmark `v12/examples/benchmarks/matrixmultiply.able`
    now measures `3.4367s` / `13.67` GC over 3 compiled runs, which closes the
    current propagated static-array accessor pointer-carrier gap on the matrix
    family.
- The next shared primitive/control-flow step is now landed too:
  - canonical primitive counted loops of the form
    `loop { if i >= n { break } ... i = i + 1 }` now lower to direct Go
    `for i < n { ... i++ }` loops on compiled static paths;
  - the matcher is conservative by construction and now inspects nested
    function/lambda/iterator/ensure bodies so the fast path rejects loops that
    can still mutate the induction variable indirectly;
  - the compile-shape audit now proves `build_matrix` and `matmul` stay on
    direct counted loops, and `matmul` no longer carries
    `__able_checked_add_signed(...)` for loop induction;
  - the reduced matrix benchmark now measures `0.1133s` / `7.00` GC over
    3 compiled runs;
  - the full macro benchmark now measures `1.0833s` / `13.00` GC over
    3 compiled runs;
  - loop control scaffolding is no longer the limiting primitive residual on
    the matrix family.
- The next shared primitive affine-arithmetic step is now landed too:
  - fixed-width primitive checked `+` / `-` under 64 bits now lower inline on
    static compiled paths instead of calling the checked helper functions;
  - `int`, `uint`, `i64`, and `u64` intentionally remain on the existing
    helper path because they still depend on wider-width/runtime-width
    overflow machinery;
  - the compile-shape audit now proves `build_matrix` no longer carries
    `__able_checked_add_signed(...)` / `__able_checked_sub_signed(...)` for
    `i - j` / `i + j`; those now lower as inline `int64(...) +/- int64(...)`
    plus explicit range checks;
  - the reduced matrix benchmark remains `0.1133s` / `7.00` GC over 3 runs;
  - the full macro benchmark remains in the same band at `1.0867s` /
    `13.00` GC over 3 runs;
  - the remaining primitive residual is now the inline overflow branches
    themselves where static range proofs can prove they are unnecessary.
- The next shared primitive range-proof step is now landed too:
  - the compile context now tracks simple primitive integer sign facts per Go
    binding and carries them into child scopes while clearing them on
    rebinding/shadowing;
  - inline checked signed subtraction now lowers directly when both operands
    are proven non-negative;
  - the compile-shape audit now proves `build_matrix` lowers `i - j` as a
    direct signed subtraction, while `i + j` still carries the widened inline
    overflow branch;
  - the reduced matrix benchmark remains in the same band at `0.1167s` /
    `7.00` GC over 3 runs;
  - the full macro benchmark remains in the same band at `1.1000s` /
    `13.00` GC over 3 runs;
  - the remaining primitive residual is now stronger upper-bound proofs for
    affine addition like `i + j`, not subtraction.
- The next shared primitive upper-bound step is now landed too:
  - the compiler now carries simple primitive upper-bound facts across
    statically resolved function calls and seeds them back into callee param
    contexts before render;
  - counted-loop induction variables now inherit a conservative upper bound
    from their loop guard when the bound is statically known;
  - inline checked signed addition now lowers directly when both operands are
    proven non-negative and their combined upper bound fits the target width;
  - the compile-shape audit now proves `build_matrix` lowers both `i - j` and
    `i + j` as direct signed arithmetic with no widened `int64(...)` affine
    branch scaffolding left in the inner loop;
  - the reduced matrix benchmark remains in the same band at `0.1267s` /
    `7.00` GC over 3 runs;
  - the full macro benchmark remains in the same band at `1.1367s` /
    `13.00` GC over 3 runs;
  - the hot affine integer residual on the matrix path is now closed, so the
    next worthwhile category is no longer loop-affine primitive arithmetic on
    this benchmark family.
- The remaining primitive numeric scalar family is now staged too:
  - `Array i8`, `Array i16`, `Array u16`, `Array u32`, `Array u64`, and
    `Array f32` now lower to compiler-owned wrappers on the staged specialized
    path
  - reduced unsigned benchmark:
    `v12/fixtures/bench/sum_u32_small/main.able`
  - compiled 3-run averages: mono on `1.0933s`, `185.33` GC; mono off
    `1.6800s`, `21.33` GC
- The first broader non-array container carrier slice is landed too:
  - `HashMap K V` now lowers to native `*HashMap` carriers on static compiled
    paths instead of collapsing to `any`
  - typed/inferred map literals and `Array (HashMap K V)` shells stay native
    in the compiled body
  - shared nominal struct lowering now expands simple aliases before mapping,
    so kernel-style alias-backed fields like `HashMapHandle = i64` stay native
    on the same generic path instead of requiring another container-specific
    exception
  - the remaining map-literal handle edge now converts runtime handle values
    back into the native `HashMap.handle` carrier explicitly
  - `Map K V` interface params stay on the native interface carrier, and the
    old `HashSet.iterator()` / `Enumerable.iterator` fallback is closed via
    corrected generic interface return matching plus an explicit narrowed
    runtime adapter roundtrip at the residual ABI edge
  - the compiled control bridge now preserves exit signals before wrapping
    raised values, which closes the false `runtime error` failure on compiled
    stdlib flows that finish through `__able_os_exit(...)`
  - reduced container benchmark:
    `v12/fixtures/bench/hashmap_i32_small/main.able`
  - compiled 3-run average: `1.7633s`, `175.33` GC
- The nominal follow-through slice beyond `HashMap` is landed too:
  - `TreeMap` / `TreeSet` and `PersistentMap` / `PersistentSet` now compile
    through the same generic nominal struct/interface carrier path instead of
    requiring new per-type lowering rules
  - static compiled bodies no longer emit bare Go `len(...)` / `cap(...)` /
    string-byte-length calls on these paths; they use generated helpers so
    Able locals like `len` cannot break Go compilation
  - the compiled stdlib fixture gates `06_12_11_stdlib_collections_tree_map_set`
    and `06_12_12_stdlib_collections_persistent_map_set` are green again
- The next broader stdlib container families are now mechanically covered on
  that same shared native path too:
  - new no-fallback regressions pin representative static methods for
    `Deque` / `Queue`, `BitSet` / `Heap`, and
    `PersistentSortedSet` / `PersistentQueue`
  - those generated methods are now explicitly audited to avoid
    `__able_call_value(...)`, `__able_member_get_method(...)`,
    `__able_method_call_node(...)`, `bridge.MatchType(...)`, and
    `__able_try_cast(...)`
  - shared compiled fixture gates for
    `06_12_13_stdlib_collections_persistent_sorted_queue`,
    `06_12_16_stdlib_collections_deque_queue`, and
    `06_12_17_stdlib_collections_bit_set_heap` are green
  - reduced benchmark target:
    `v12/fixtures/bench/heap_i32_small/main.able`
  - compiled 3-run average: `7.7533s`, `1105.00` GC
- The next deeper generic-container correctness slice is closed too:
  - generic nullable/interface carriers such as `LazySeq.Source: ?(Iterator T)`
    now stay on generated native carriers instead of degrading to `any`
  - compiled `nil` lowering now emits typed Go nils for native nilable
    carriers (`(*ListNode)(nil)`, `__able_iface_Iterator_T(nil)`, etc.)
    instead of raw untyped `nil` temps
  - the compiled stdlib fixture gate
    `06_12_14_stdlib_collections_linked_list_lazy_seq` is green again on the
    same shared native path
- The first benchmark-worthy generic-container hot path is closed too:
  - static `for value in iterable` lowering now uses native concrete/interface
    receiver calls instead of `__able_resolve_iterator(...)`
  - native interface adapter synthesis now honors interface inheritance, so
    derived-interface impls such as `Enumerable A for LinkedList` synthesize
    matching native base-interface carriers like `Iterable A`
  - concrete native interface adapters now directly coerce compatible native
    interface return carriers, which removes recursive runtime-value
    round-trips on cyclic container graphs like `LinkedListIterator`
  - reduced benchmark fixture:
    `v12/fixtures/bench/linked_list_for_i32_small/main.able`
  - compiled 3-run average: `0.2000s`, `15.00` GC
- The next concrete generic/default container-method hot path is closed too:
  - higher-kinded interface self patterns like `Enumerable A for C _` now bind
    `C` to the concrete target type on compiled impl paths
  - bound type-constructor calls like `C.default()` now resolve to compiled
    impl methods instead of `__able_env_get("C")`
  - native `Iterator<T>` carriers now satisfy compiled iterable lowering
    directly inside generic default impl bodies, which removes the
    `to_runtime_value -> from_value -> iterator()` round-trip that previously
    overflowed on larger `LinkedList` graphs
  - reduced benchmark fixture:
    `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`
  - compiled 3-run average: `0.1667s`, `12.00` GC
- The callback/runtime-value carrier cleanup slice on that same hot path is
  closed too:
  - specialized impl functions now retain bound generic type bindings through
    compileability checks and render, so compiled bodies discovered while
    specializing other impls are emitted in the same pass
  - specialized sibling impls are cached early enough to break mutually
    recursive specialization loops during codegen
  - direct sibling selection inside default impl bodies now prefers those
    specialized sibling impls before the ordinary concrete-receiver path
  - that fixes the last `LinkedList` `Enumerable.lazy()` regression:
    specialized helpers now call `iterator_*_spec(...)` directly instead of
    bridging `Iterator_A -> runtime.Value -> Iterator_i32`
  - follow-up compiled 3-run average on
    `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`:
    `0.1667s`, `15.33` GC
- The remaining shared static nominal receiver/struct-literal refinement gap
  on that same reduced family is closed too:
  - recursive type substitution now resolves chained specialization bindings
    transitively instead of stopping on placeholder-to-placeholder maps
  - static nominal target refinement now upgrades bare targets and struct
    literals like `LazySeq { ... }` to the concrete specialized carrier when
    the specialization context already knows the expected target type
  - native interface concrete-impl matching now compares specialized receiver
    targets through the shared target-template path, which keeps
    `*LinkedList_i32` on the native `Iterable<i32>` adapter path
  - follow-up compiled 3-run average on
    `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`:
    `0.1633s`, `8.33` GC
- The next iterator default-method hot path is closed too:
  - ordinary default native-interface methods now lower through the same
    direct compiled-helper path already used for default generic methods when
    the receiver stays on a native interface carrier
  - on the representative `LinkedList.lazy().map<i64>(...).filter(...).next()`
    shape, `Iterator.filter` no longer routes through the runtime adapter
    method layer after `map(...)`; it resolves directly to the compiled
    default helper on the native iterator carrier
  - reduced benchmark fixture:
    `v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able`
  - compiled 3-run average: `0.1800s`, `13.33` GC
- The mono-array-enabled `Iterator.collect<Array T>()` follow-up is now closed
  too:
  - the compiler now emits a generated compiled helper with a specialized
    mono-array accumulator for `collect<Array i64>()` instead of routing
    through `__able_method_call_node(...)` plus `__able_array_i64_from(...)`
  - the shared generic default-method path now also has an explicit
    user-defined nominal proof case: `Iterator.collect<C>()` stays native for
    a `Default + Extend` accumulator struct (`SumCount`) without another
    named-container branch; the Array helper remains only as a fallback for
    the built-in `Array` exception
  - reduced benchmark fixture:
    `v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able`
  - compiled 3-run averages: mono on `0.1833s`, `14.00` GC; mono off
    `0.1833s`, `13.33` GC
- The iterator-literal controller/runtime-value edge is now closed too:
  - compiled iterator literals bind `gen` as a compiler-owned
    `*__able_generator` controller instead of an opaque runtime object
  - `gen.yield(...)`, `gen.stop()`, and bound `gen.yield` callable captures
    now lower directly through that controller path instead of
    `__able_method_call_node(...)`
  - native nilable/static-carrier conditions now lower as direct nil checks,
    which keeps `Iterator.filter_map` on the static path
  - reduced benchmark fixture:
    `v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able`
  - compiled 3-run average: `0.1267s`, `10.00` GC
- Shared generic nominal `methods` specialization is now closed too:
  - statically known concrete nominal targets now get specialized compiled
    method bodies on the shared nominal-method path instead of reusing the
    unspecialized `runtime.Value` signatures
  - this now has an explicit user-defined proof case (`Box T`) and a hot
    reduced benchmark proof case (`Heap i32`) without adding another
    named-structure lowering rule
  - reduced benchmark fixture:
    `v12/fixtures/bench/heap_i32_small/main.able`
  - compiled 3-run average after the tranche: `4.2000s`, `1811.67` GC
- Bound generic field/member carrier refinement inside those already-
  specialized nominal method bodies is now closed too:
  - fully bound generic fields like `self.items: Array T` now stay on their
    concrete native carrier once `T` is known, instead of re-entering the
    residual `runtime.Value` index path
  - user-defined proof case: `Bucket T { items: Array T }` now renders
    `Items *__able_array_i32` plus specialized `Bucket.push` /
    `Bucket.second` bodies under `ExperimentalMonoArrays`
  - sibling/default impl specialization now upgrades placeholder receiver
    bindings like `T -> T` to concrete bindings like `T -> i64`, which closes
    the remaining mono-array `Iterable.iterator` / `Iterable.each` execute gap
  - reduced benchmark fixture:
    `v12/fixtures/bench/heap_i32_small/main.able`
  - compiled 3-run average after the tranche: `0.7667s`, `91.33` GC
- Mixed-result control-flow join inference is now closed too:
  - `if`, `match`, and `rescue` expressions now use one shared join-type
    inference path, so when all branch result types are statically
    representable the compiler synthesizes or reuses a native carrier instead
    of defaulting the join local to `runtime.Value`
  - branch coercion now reuses the shared native static-coercion path, so
    native union/interface/callable/nullable carriers can serve directly as
    join results
  - static typed patterns now accept same-family nominal carriers through
    shared receiver compatibility instead of requiring exact Go carrier
    identity, which keeps specialized native carriers on the static match path
- Mixed-result `or {}` lowering now uses that same shared join path too:
  - representable mixed success/handler result shapes now stay on native
    carriers instead of defaulting the `or {}` result local to `runtime.Value`
  - nullable success paths now join on the unwrapped payload carrier rather
    than the pointer carrier
  - `err => ...` bindings now stay on the native failure carrier when the
    failure branch type is statically known, which keeps compiled
    option/result error handlers off the old runtime-value bridge
- `loop` and labeled `breakpoint` expressions now use that same shared
  control-flow join/coercion path too:
  - statically representable `break` payloads now bind directly onto native
    result carriers instead of forcing loop/breakpoint result temps through
    `runtime.Value`
  - labeled non-local `break` payloads now coerce directly onto the target
    breakpoint's native result carrier, including cases where the payload uses
    locals declared earlier in the breakpoint body
  - bare `break` continues to mean `nil`; when a loop result temp exists the
    compiler now writes that nil payload explicitly instead of accidentally
    reusing the loop's normal-completion `void` sentinel
- Shared existential join inference is now wider too:
  - when mixed concrete join branches already share a native existential
    carrier, the compiler now prefers that carrier before synthesizing a union
  - concrete zero-arg interface implementers now join directly on the native
    interface carrier instead of materializing a union local and then calling
    `__able_method_call_node(...)`
  - mixed native `Error` implementers now join directly on
    `runtime.ErrorValue` instead of generating an intermediate union carrier
  - pure-generic interface calls on those joined carriers still route through
    the compiled generic dispatch helper even when multiple concrete adapters
    exist for the interface
  - fully bound parameterized interface joins now use that same shared carrier
    preference too: candidate native interface carriers are materialized from
    the actual branch impl metadata, and bound base interfaces are materialized
    alongside child interfaces so inherited common carriers can be discovered
    without adding any named non-primitive rule
  - nil-capable joins now use that same shared inference too: `nil` branches
    are tracked separately from concrete branch carriers, the concrete carriers
    are joined first, and the result stays native when that joined carrier
    already has a nil zero value or a native nullable wrapper exists
  - type-expression-backed joins are now closed too: when a branch/local still
    reports `runtime.Value` or `any` but retains a concrete normalized Able
    `TypeExpr`, shared join inference now recovers the native carrier instead
    of widening the whole join back to `runtime.Value`; typed-pattern bindings
    preserve that `TypeExpr`, identifier lowering prefers the recovered native
    carrier on use, and `if` / `match` / `rescue` / `or {}` / loop /
    breakpoint joins all reuse that shared recovered-type path
- The remaining typed-pattern/control-flow closure on this family is now
  landed too:
  - recoverable typed pattern bindings now lower through one shared dynamic
    typed-pattern cast path, so the bound local stays on its native carrier
    instead of defaulting to `runtime.Value`
  - rescue typed bindings now stay native on recoverable carriers such as
    the native nullable scalar/error family, the native scalar family,
    `runtime.ErrorValue`, generated native nominal struct carriers,
    generated native union/result carriers, generated native interface
    carriers, and generated native callable carriers
  - dynamic rescue/match typed-pattern narrowing for those native nullable
    scalar/error, native scalar, native nominal struct, native union/result,
    native interface, native callable, and `runtime.ErrorValue` carriers now
    also skips `__able_try_cast(...)` and instead uses shared native matcher
    helpers / direct scalar runtime type checks / direct nullable helpers /
    direct error detection
  - native-union whole-value typed bindings now coerce directly onto the
    shared interface carrier instead of reboxing through `runtime.Value`
  - `or { err => ... }` bindings now also stay native on statically known
    failure carriers for propagated result-control paths and native
    error-union handler paths instead of defaulting those locals back to
    `runtime.Value`
  - shared handled-failure inference now follows propagated `!T` failures
    through block-wrapped and non-tail statement contexts too, so `rescue`
    bindings, `or { err => ... }` bindings, and rescue-clause local joins
    keep their native error carriers in those wrapped forms instead of
    widening back through `runtime.Value`
  - imported nullable / union / result aliases now retain their alias-source
    package during carrier synthesis too, so shadowed foreign nominal members
    stay on native pointer/union carriers instead of collapsing onto same-
    named local structs and falling back on typed-pattern mismatch
  - that same foreign-package context now survives later typed-pattern/local
    rebinding too, so imported nullable / union / result match bindings access
    shadowed foreign nominal members directly instead of round-tripping
    through nominal conversion helpers or runtime member dispatch
  - mixed imported/local shadowed nominal joins now key normalized type
    identity with resolved package context too, so same-named local and
    imported struct branches keep distinct native union members instead of
    collapsing onto one `Thing`-shaped join and falling back
  - shadowed callable joins built from those nominal returns now stay on
    native callable-union carriers too: callable metadata retains package
    context during carrier reconstruction, and join synthesis no longer treats
    unrelated native callable carriers as interchangeable just because both
    happen to be callable
  - lambda literals and placeholder lambdas now narrow through expected
    callable members inside native union/result carriers too, and semantic
    `Result` carrier synthesis now preserves the callable member's resolved
    package context, so imported semantic `Option` / `Result` aliases and
    direct union aliases built from shadowed callable returns no longer fall
    back with `lambda expression type mismatch` or
    `placeholder lambda type mismatch`
  - imported selector aliases for generic interfaces now preserve their source
    package during generic type normalization too, so `RemoteIface<T>`-style
    aliases keep foreign native interface carriers inside nullable / union /
    result aliases even when the caller shadows the same interface name
    locally
  - raw imported selector-alias typed patterns now normalize in the lexical
    caller package before any previously recorded foreign package is reused,
    so generic `Result` / semantic-result matches like
    `Outcome(RemoteIface<i32>) match { case value: RemoteIface<i32> => ... }`
    resolve back onto the same foreign native interface carrier instead of
    degrading to `runtime.Value`
  - imported semantic `Result` aliases nested under outer `Result` carriers
    now stay on that same foreign native interface carrier too: alias
    expansion preserves the alias source package, builtin `Error` identities
    collapse across package contexts during carrier synthesis, and final
    codegen invalidates stale normalization/function-carrier caches after
    compileability probing
  - imported semantic `Result` aliases over shadowed callable members now
    stay native under outer `Result` carriers too: raw imported selector
    aliases nested inside function type expressions no longer reuse stale
    recorded foreign package context during normalization, so typed patterns
    like `(() -> RemoteThing)` resolve back onto the foreign native callable
    carrier instead of widening to `__able_fn_*_to_runtime_Value`; proof
    coverage now pins the same outer-result path for imported semantic
    `Option` aliases and imported generic union aliases over those shadowed
    callable members too
  - proof coverage now also pins local parameterized union/result aliases
    with imported shadowed interface/callable actuals, so locals like
    `Choice (RemoteReader i32)` and `Outcome (() -> RemoteThing)` stay on
    native carriers instead of widening through `runtime.Value` / `any`
  - proof coverage now also pins generic specialization over those same
    imported shadowed alias actuals, so specialized helper signatures stay
    on native union/result/interface/callable carriers instead of widening to
    `runtime.Value` / `any`
  - proof coverage now also pins direct shared-mapper lookups and generic
    interface-method dispatch over those same imported shadowed alias
    actuals, so `lowerCarrierTypeInPackage(...)` plus existential
    `pass<T>(...)`-style calls stay on native
    union/result/interface/callable carriers too
  - imported generic interface-method calls now keep caller-side explicit
    type arguments on those same native carriers too: generic-method shape
    inference normalizes explicit type arguments in the lexical caller
    package and retries representable-carrier recovery while computing
    concrete param/return helper signatures, so imported `Echo.pass<T>(...)`
    and imported default generic method calls stop widening to
    `runtime.Value` / `any`
  - imported generic interface default methods now also resolve nested
    selector-imported members inside those explicit type arguments before
    specializing the interface-package body, so
    `tagged.tagged<Outcome(() -> RemoteThing)>(...)` keeps the concrete
    native `Tagged<...>` carrier and avoids `__able_method_call_node(...)`
  - native union synthesis now retries representable-carrier recovery in the
    member's resolved package before accepting a residual `runtime.Value`
    member, and the imported shadowed interface/callable alias families above
    are now pinned against hidden `_runtime_Value` union variants inside the
    generated helper family too
- generic specialization now retries the same representable-carrier
  recovery path before rejecting a fully bound actual as broad, and the
  imported shadowed specialization matrix is now pinned across result /
  union alias families for both native interface and native callable
  actuals
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
  synthesis, and native callable signature materialization now retry that
  same representable-carrier recovery pass after raw package-scoped
    mapping too, so substituted imported shadowed alias families stay on
    native union/result/interface/callable carriers instead of broadening
    internally to `runtime.Value` / `any`
  - imported generic struct members with shadowed nominal arguments now stay
    on specialized native carriers inside result / nested-result /
    nested-union-result families too: selector-imported nominal arguments now
    count as fully bound in the caller package, and foreign generic struct
    specialization keeps that caller-side package context through field
    substitution and field-type inference; proof coverage now also pins the
    same native-carrier behavior through generic specialization over
    imported shadowed generic-struct result/union aliases, and foreign
    generic structs no longer re-run an overly strict whole-expression
    concreteness gate after that fully bound argument check either, so
    imported interface/callable arguments such as `RemoteReader i32` and
    `() -> RemoteThing` keep specialized native `Box<...>` carriers instead
    of falling back to base `*Box` plus dynamic member calls
  - local generic nominal carriers over normalized nullable/union/result
    members now stay on those same specialized native carriers too:
    `Box MaybeReader`, `Box Choice`, `Box Outcome`, `Box !(Choice)`, and
    `Box !(Outcome)` all keep specialized native `Box<...>` carriers instead
    of falling back to base `*Box`, `runtime.Value`, or `any`; this closes
    the deeper `types.go` / `generator_native_unions.go`
    carrier-synthesis cleanup
  - generic specialization already keeps those native interface carriers too:
    direct `T -> T` specializations on interface actuals stay on the native
    interface signature, and fully bound duplicate generic unions like
    `T | RemoteIface<i32>` at `T = RemoteIface<i32>` collapse to that same
    foreign native interface carrier instead of widening through
    `runtime.Value`, `any`, or a synthetic union wrapper; the follow-on fix
    here was making no-op type substitution preserve the bound expression
    node so imported selector source-package context survives into the
    specialized helper signature too
  - representable nested union/result members now flatten during carrier
    synthesis too, so outer unions like `(!T) | U` and `(A | B) | U`, plus
    direct nested result families like `!!T` and `!(A | B)`, lower to one
    native union family instead of nesting a native union carrier inside
    another, including imported shadowed-interface `Result` members
  - join carrier selection is now nominal for interfaces, so unrelated
    same-shape interface joins like `Reader<i32> | Source<i32>` stay on a
    native union instead of silently collapsing to one interface carrier
  - propagated `rescue { case value => ... }` identifier joins now reuse those
    native carriers when the monitored call already has a statically known
    native return type, so callable plus imported shadowed nominal/interface
    failures no longer widen back through `runtime.Value`, native-error
    unions, or member access fallback
  - higher-order / unknown rescue call failures no longer reuse the callback
    return type as a fake failure carrier; statically known callees still
    recover native rescue carriers from their bodies, while unresolved
    callback failures stay on the dynamic error path so handlers like
    `err.value match { ... }` compile without collapsing `err` to `String`
  - rescue clause pattern compilation now also clears unrelated outer
    `expectedTypeExpr` context before binding the failure subject, so the
    unresolved higher-order rescue path stays on dynamic error carriers
    instead of inheriting the enclosing return type by accident
  - no-bootstrap bridge fallback stringification for raised non-`Error`
    values now stays aligned with interpreter-visible string output, so
    compiled rescue/string paths no longer diverge under compiler-only runs
  - nested struct-pattern field bindings now rebind recursive expected type
    context to the matched field type instead of leaking the outer nominal
    expectation inward, which keeps persistent-collection helper patterns on
    the correct native carriers
  - dynamic typed-pattern casts now allocate temps on the caller context
    instead of a cloned probe context, so iterator-end / generator-yield
    matches stop colliding with surrounding temps during codegen
  - raised imported shadowed nominal struct literals now prefer the
    compiler's syntax-aware struct-literal type reconstruction during failure
    inference, so propagated rescue joins keep the foreign native struct
    carrier instead of degrading into a same-named local nominal union
  - generated struct boundary `*_from(...)` helpers now keep raw `fieldValue`
    temporaries compile-clean even when a residual unsupported conversion
    branch does not consume them, which cleared the released compiled
    concurrency parity slice after the rescue fix
  - implicit and explicit return expressions now preserve the declared Able
    return `TypeExpr` while compiling the returned expression too, so generic
    return paths like `unwrap(io_read(...))` keep nullable success carriers
    instead of degrading to the bare nil-capable Go carrier
  - static nullable typed matches on nil-capable native carriers now guard
    both the non-nil typed branch and the `case nil` branch directly too, so
    whole-carrier native interface/result matches no longer emit dead
    unconditional/false conditions ahead of the actual nil arm
  - the nullable typed-match nil-guard relaxation is now limited to actual
    in-scope generic type variables, so concrete `?Interface<T>` /
    `?Result<T>` typed arms regain the required non-nil guard instead of
    compiling as unconditional typed branches
  - representable outer unions built from native nullable/result members now
    keep direct inner-member literals and clause-ordered typed matches on
    native carriers too: nested wrapping no longer rejects `"ok"` / `Box {}`
    against `?T | U` or `!T | U`, and native-union narrowing now only removes
    a member when the pattern exhausts that whole v12 member type instead of
    treating a non-nil subcase like `String` as if it also consumed `nil`
  - fully bound duplicate union/result members now collapse to their single
    native carrier during mapping too, so generic specializations no longer
    keep synthetic `runtime.Value | ...` / duplicate-result carrier shapes
    once all type parameters have been concretized
  - representative static pattern/control bodies are audited against
    `__able_try_cast(...)`, `bridge.MatchType(...)`, panic/recover, and
    IIFE-style scaffolding
- Milestone 4 is now closed:
  - shared dispatch recovery rehydrates recoverable `runtime.Value` / `any`
    call/member/index targets onto their native carriers before dispatch
  - local concrete/interface `Apply` bindings now lower through the shared
    static apply path
  - mixed-source pure-generic interface dispatch now prefers the more concrete
    compiled specialization instead of falling back to runtime method dispatch
  - representative dispatch coverage is now pinned by
    `compiler_dispatch_completeness_test.go`
- Milestone 5 is now closed:
  - static compiled helper families now use direct Go `_impl` runtime-core
    helpers on static paths instead of `__able_extern_*` or helper-to-helper
    `__able_extern_call(...)` chains
  - representative helper-body coverage now pins direct array/channel/mutex
    runtime-core helper usage
  - zero-arg callable syntax and `Await.default` zero-arg callback
    specialization now stay on native callable carriers across compiled static
    and spawned-task paths
  - representative static fixture coverage now includes:
    - `06_01_compiler_spawn_await`
    - `06_12_02_stdlib_array_helpers`
    - `06_12_19_stdlib_concurrency_channel_mutex_queue`
- Milestone 6 is complete:
  - the final explicit boundary helper set is now mechanically locked to
    `call_value`, `call_named`, and `call_original`
  - representative static no-bootstrap fixture execution is audited for zero:
    - fallback boundary calls
    - explicit dynamic boundary calls
    - interface/member lookup fallback calls
    - global lookup fallback calls
  - representative no-fallback static fixture batches remain green under the
    boundary-marker harness
- Milestone 7 is complete:
  - added the reduced recursion benchmark
    `v12/fixtures/bench/fib_i32_small/main.able`
  - shared compiled callable/runtime env scaffolding now swaps package envs
    only when the target env differs from the current env
  - representative current compiled numbers are recorded in
    `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`
- Milestone 8 is complete:
  - compiler release validation is green under:
    - `GOFLAGS='-p=1' ./run_all_tests.sh --compiler`
    - `./run_stdlib_tests.sh`
  - the milestone-closing shared semantic fixes were:
    - range expression inferred-carrier correction back to shared iterable
      lowering
    - native-interface default-method dispatch preserving concrete wrapped
      receiver overrides
    - all-numeric union operator typing through pairwise numeric promotion
  - release validation is closed, but the stronger compiler-native completion
    program remains active
  - the array-native lowering tranche is complete on 2026-04-01; remaining
    `runtime.ArrayValue` / `ArrayStore*` use is now limited to explicit
    dynamic or ABI edges plus the unspecialized wildcard-array ABI
  - the mono-array transition/runtime-store cleanup is complete on 2026-04-14;
    compiler-generated mono-array wrappers are now pure slice carriers and
    explicit boundary helpers are the only remaining mono-array touchpoints for
    `runtime.ArrayValue` / `ArrayStore*`
  - no-interpreter generic-interface dispatch now revalidates aliases and
    interface constraints from compiler-emitted runtime metadata/helpers
    instead of interpreter registries
  - the stronger compiler-native completion program is closed on 2026-04-14,
    and the next active project queue is bytecode performance

## Relationship To Other Design Notes

- `compiler-monomorphization.md` is a subordinate design note for generic/native
  container lowering.
- `compiler-no-panic-flow-control.md` is a subordinate design note for explicit
  control-result propagation.
