# Compiler Monomorphization Design

This note refines [compiler-native-lowering.md](./compiler-native-lowering.md)
for generic/container-heavy code generation.

Named stdlib/container examples in this document are proof cases for shared
lowering machinery, not architecture exceptions.

## Goal

Use monomorphization and static specialization to keep compiled code on native
Go carriers instead of dynamic/interpreter carriers.

Monomorphization is a means to the native-lowering end state. It is not a
license to keep `Array<T>` on top of `runtime.ArrayValue`, `ArrayStore*`,
`runtime.Value`, or `any` on otherwise static paths.

## Representation Targets

| Able type | Target compiled representation |
| --- | --- |
| `Array i32` | compiler-owned wrapper around `[]int32` |
| `Array String` | compiler-owned wrapper around `[]string` |
| `Array Point` | compiler-owned wrapper around `[]*Point` |
| `Array (A | B)` | compiler-owned wrapper over native union-carrier elements |
| `Point` | `*Point` |
| `A | B` | generated Go interface plus native variants |
| `?Point` | nil-capable `*Point` |

`any` is acceptable only as a temporary implementation escape hatch or at an
explicit dynamic boundary. It is not the target representation for compiled
union values or generic containers.

## Arrays

### Required direction

- Static compiled arrays must use compiler-owned native Go storage.
- Array literals should become native wrapper construction over typed Go slices.
- `len`, `push`, `pop`, `get`, `set`, indexing, iteration, and cloning should
  operate directly on that native storage.
- The compiler must not treat kernel `Array { length, capacity, storage_handle }`
  plus runtime handles as the authority for static array storage.

### Boundary rule

If a compiled array crosses into explicit dynamic behavior, the compiler may
generate adapters:

- native specialized array wrapper -> runtime/kernel boundary value
- boundary value -> native specialized array wrapper

That conversion is boundary logic only. It must not leak back into the default
static lowering path.

### Revised mono-array design reference

The detailed staged plan now lives in
[monomorphized-container-abi.md](./monomorphized-container-abi.md).

Important update from the 2026-03-19 audit:

- the older typed-runtime-store / handle-tag rollout plan is superseded as the
  final architecture;
- future mono-array work must target compiler-owned specialized wrappers over
  native Go slices;
- existing runtime-store experiments are historical scaffolding only.

## Unions And Generics

For generic code that cannot be fully specialized yet:

- prefer generated Go interfaces and native wrapper types;
- use `any` only as a temporary residual fallback;
- document every residual `any` use as a staged limit, not as the target ABI.

Pattern matching should eventually compile to native Go type checks over those
generated interfaces/variants rather than runtime-value inspection.

## Current Work To Avoid

The following are not acceptable as final monomorphization outcomes:

- `Array<T>` lowering that still depends on `runtime.ArrayValue` /
  `ArrayStore*` on static paths;
- typed runtime-store / handle-tag routing as the primary compiled
  representation of mono arrays;
- compiler-only rewrites of the kernel `Array` struct shape that hide
  `storage_handle` behind `Elements []runtime.Value` and stop there;
- union lowering that stops at `any` instead of generating native carrier
  interfaces;
- struct-local boxing into `runtime.Value` to paper over missing static ABI
  work.

## Execution Order

1. [x] Freeze the compiler-native mono-array ABI in docs and guardrail tests.
2. [x] Emit specialized array wrapper types for staged element kinds.
3. [x] Specialize the first direct typed hot-path slice over those wrappers:
   explicit typed literals, `push/get/set/read_slot/write_slot`, direct index
   read/write, and wrapper/lambda/runtime boundary conversion.
4. [x] Widen specialization coverage to constructors / stdlib factories,
   unannotated local inference, clone / iteration / pattern paths, and other
   residual generic `*Array` surfaces.
5. [x] Re-measure compiled benchmark deltas once the widened slice lands.
   - Recorded in
     `v12/docs/perf-baselines/2026-03-19-mono-array-widened-compiled.md`.
   - Result: `bench/noop` and `bench/sieve_count` stayed flat on wall-clock,
     while `bench/sieve_full` stayed flat on wall-clock but dropped timed GC
     from `3.00` to `1.00`.
   - Interpretation: the widened slice is reducing allocation pressure on the
     heaviest staged array benchmark, but the next material runtime win now
     depends on shrinking residual generic `*Array` / `runtime.Value` paths.
6. [x] Narrow the first residual generic/`runtime.Value` array slice until the
   remaining static mismatches are either closed or explicitly logged.
   - Inferred local `Array T` bindings now keep enough recoverable metadata for
     generic static helper results (`get`, `pop`, `first`, `last`,
     `read_slot`) to stay on native nullable carriers where representable.
   - Static and residual runtime-backed array `set` / index-assignment success
     is back in parity at `nil`, restoring the compiled `nil | IndexError`
     contract for the array helper fixtures.
   - Runtime-backed iterator interface boundaries now accept raw
     `*runtime.IteratorValue` carriers directly, and generator stop now stays
     iterator completion instead of becoming a generic runtime error.
   - Result: the previously-blocked `06_08_array_ops_mutability`,
     `06_12_02_stdlib_array_helpers`, and
     `06_12_18_stdlib_collections_array_range` compiler fixtures are green
     again on the current native-lowering arc.
7. [x] Extend the staged specialized array set to `f64` and re-measure a
   nested `Array (Array f64)` benchmark target.
   - `Array f64` now lowers to `*__able_array_f64` on the staged specialized
     path.
   - Native nullable propagation now handles pointer-backed scalar carriers
     such as `*float64` when a concrete `float64` is required, which keeps
     nested expressions like `rows.get(j)!.get(i)!` on the static path instead
     of forcing the surrounding `push(...)` back through
     `__able_method_call_node(...)`.
   - The full compiled `v12/examples/benchmarks/matrixmultiply.able` path no
     longer aborts early under mono arrays; at `60s` it now matches mono-off
     by timing out rather than failing.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-mono-array-f64-matrixmultiply-small-compiled.md`
   - Result: mono on `5.4833s` / `280.00` GC vs mono off `45.3133s` /
     `3568.67` GC over 3 compiled runs.
8. [x] Remove the remaining generic outer row shell for nested staged
   mono-array carriers.
   - Nested typed arrays such as `Array (Array f64)` now lower to dedicated
     outer wrappers like `*__able_array_array_f64` over `[]*__able_array_f64`
     instead of generic `*Array` with `[]runtime.Value` row storage.
   - Rendered mono-array converters and native `Array` core helpers now
     handle pointer-backed specialized elements directly, including nilable
     `read_slot` / `pop` results and explicit nested runtime boundary
     conversion.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`
   - Result: mono on `5.7233s` / `252.00` GC vs mono off `44.5167s` /
     `3550.67` GC over 3 compiled runs.
   - Interpretation: this closes the last generic outer-row carrier on the
     representative nested static path, but does not add a second visible
     wall-clock step beyond the earlier `f64` slice by itself.
9. [x] Remove the remaining scalar runtime-value crossings on the reduced
   matrix path through shared built-in `Array` and primitive lowering rules.
   - Static staged-array propagation now returns concrete success element
     types directly on the compiled path, so nested `get(...)!` / index
     propagation no longer routes scalar success values through
     `__able_nullable_*_to_value(...)`.
   - Primitive numeric casts such as `i32 -> f64` now lower directly to Go
     casts on static compiled paths instead of `__able_cast(...)`.
   - Shared primitive `float -> int` casts now also lower natively through
     truncate/range/overflow checks instead of `__able_cast(...)` and
     `bridge.AsInt(...)`, which removes the remaining obvious runtime cast
     helper use from the full matrix entry path.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-float-int-casts-compiled.md`
   - Result:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
     now measures `1.7567s` / `7.00` GC over 3 compiled runs.
   - Interpretation: the reduced matrix scalar loop is now back on the native
     path and the entry-path primitive cast gap is closed too.
9. [x] Remove synthetic call-frame scaffolding from shared static built-in
   `Array` factories/intrinsics on the matrix path.
   - Shared static built-in `Array` factories and intrinsics no longer emit
     synthetic `__able_push_call_frame(...)` / `__able_pop_call_frame()`
     pairs on compiled static paths.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-frame-elision-compiled.md`
   - Result:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
     now measures `0.1933s` / `7.00` GC over 3 compiled runs, and the full
     `v12/examples/benchmarks/matrixmultiply.able` benchmark now measures
     `4.2267s` / `13.00` GC over 3 compiled runs.
   - Interpretation: synthetic static-array frame churn was the dominant
     remaining macro-scale cost on the matrix family.
10. [x] Remove pointer-backed nullable-carrier construction from propagated
   static built-in `Array` accessors on the matrix path.
   - Propagated static built-in `Array` accessors (`get`, `first`, `last`,
     `read_slot`, `pop`) now lower as direct bounds-check + element-load paths
     with nil control transfer instead of manufacturing pointer-backed
     nullable carriers on the success path.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-propagation-pointer-elision-compiled.md`
   - Result:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
     now measures `0.1967s` / `7.33` GC over 3 compiled runs, and the full
     `v12/examples/benchmarks/matrixmultiply.able` benchmark now measures
     `3.4367s` / `13.67` GC over 3 compiled runs.
   - Interpretation: propagated static-array accessor pointer construction was
     the next shared macro-scale residual on the matrix family; that gap is
     now closed.
11. [x] Lower canonical counted primitive loops directly on the matrix path.
   - Shared counted-loop recognition now lowers canonical primitive loop
     shapes like `loop { if i >= n { break } ... i = i + 1 }` to direct
     Go `for i < n { ... i++ }` loops when the induction variable and bound
     stay on primitive integer carriers and the body does not mutate the
     counter outside the trailing increment.
   - The matcher now traverses nested function/lambda/iterator/ensure bodies
     too, so the fast path stays conservative instead of assuming nested
     control flow cannot touch the induction variable.
   - The compile-shape audit now proves `build_matrix` and `matmul` use direct
     counted loops, and `matmul` no longer carries
     `__able_checked_add_signed(...)` for loop induction.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`
   - Result:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
     now measures `0.1133s` / `7.00` GC over 3 compiled runs, and the full
     `v12/examples/benchmarks/matrixmultiply.able` benchmark now measures
     `1.0833s` / `13.00` GC over 3 compiled runs.
   - Interpretation: loop-induction checked arithmetic was the dominant
     remaining shared primitive/control-flow cost on the matrix family;
     the next residual is affine checked integer arithmetic inside
     `build_matrix`, not loop-control scaffolding.
12. [x] Inline fixed-width primitive checked `+` / `-` on static compiled
    paths.
   - Shared primitive checked addition/subtraction for fixed-width integers
     under 64 bits now lowers inline on static compiled paths instead of
     calling the checked helper functions.
   - This is intentionally narrow: `int`, `uint`, `i64`, and `u64` stay on
     the existing helper path because those widths still depend on the wider
     overflow machinery.
   - The compile-shape audit now proves `build_matrix` no longer carries
     `__able_checked_add_signed(...)` / `__able_checked_sub_signed(...)`; its
     affine integer ops now lower as inline `int64(...) +/- int64(...)` plus
     explicit range checks.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
   - Result:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
     remains `0.1133s` / `7.00` GC over 3 compiled runs, and the full
     `v12/examples/benchmarks/matrixmultiply.able` benchmark remains
     `1.0867s` / `13.00` GC over 3 compiled runs.
   - Interpretation: the affine helper-call gap is now closed, but the next
     primitive residual is the inline overflow branches themselves where a
     static range proof can show they are unnecessary.
13. [x] Use shared non-negative range proofs to remove provably-safe signed
    subtraction overflow branches on static compiled paths.
   - The compile context now tracks simple primitive integer sign facts per
     Go binding, carries them into child scopes, and clears them on
     rebinding/shadowing.
   - Inline checked signed subtraction now lowers directly when both operands
     are proven non-negative, so `build_matrix` now emits `i - j` as a direct
     signed subtraction while `i + j` remains widened.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
   - Result:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
     remains `0.1167s` / `7.00` GC over 3 compiled runs, and the full
     `v12/examples/benchmarks/matrixmultiply.able` benchmark remains
     `1.1000s` / `13.00` GC over 3 compiled runs.
   - Interpretation: subtraction-side overflow branching is now closed
     through shared primitive range proofs; the next primitive residual is
     stronger upper-bound proofs for affine addition like `i + j`.
14. [x] Use shared upper-bound range proofs to remove provably-safe signed
    addition overflow branches on static compiled paths.
   - The compiler now carries simple primitive upper-bound facts across
     statically resolved function calls and seeds them back into callee param
     contexts before render.
   - Counted-loop induction variables now inherit a conservative upper bound
     from their loop guard when the bound is statically known.
   - Inline checked signed addition now lowers directly when both operands are
     proven non-negative and their combined upper bound fits the target width,
     so `build_matrix` now emits both `i - j` and `i + j` as direct signed
     arithmetic in the inner loop.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
   - Result:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
     remains `0.1267s` / `7.00` GC over 3 compiled runs, and the full
     `v12/examples/benchmarks/matrixmultiply.able` benchmark remains
     `1.1367s` / `13.00` GC over 3 compiled runs.
   - Interpretation: the matrix inner-loop affine add/sub branch gap is now
     closed through shared primitive/function range proofs; the next
     worthwhile category is no longer loop-affine primitive arithmetic on the
     matrix path.
15. [x] Widen compiler-owned array wrapper synthesis to broader native carrier
   element families.
   - Arrays of generic inner arrays, native interface carriers, native
     callable carriers, and other representable pointer-backed carrier types
     now synthesize dedicated outer wrappers instead of falling back to
     generic `*Array` / `runtime.Value` storage.
   - Dynamic-boundary callback coverage now explicitly includes
     interface-carrier arrays and callable-carrier arrays.
   - This tranche is architectural rather than benchmark-driven; no new
     shared benchmark snapshot was added because the current benchmark set does
     not materially exercise these carrier families.
10. [x] Extend the staged specialized scalar set to the text family and
    re-measure it on a char-heavy compiled benchmark.
   - `Array char` now lowers to `*__able_array_char` over `[]rune`.
   - `Array String` now lowers to `*__able_array_String` over `[]string`.
   - Nested `Array (Array char)` now lowers to `*__able_array_array_char`.
   - Native result propagation for `!Array char` now re-wraps specialized
     success branches through static coercion instead of routing them back
     through `_from_value(__able_runtime, ...)`.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/zigzag_char_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
   - The initial mono-off comparison from this slice was invalid because the
     mono-off compiled path for nested text rows was producing the wrong
     result.
11. [x] Preserve compiler-owned carrier-array identity even with staged scalar
    mono arrays disabled, and correct the text benchmark baseline.
   - Carrier-array wrappers for already-native compiler carriers now synthesize
     regardless of `ExperimentalMonoArrays`, while staged scalar
     specializations remain flag-gated.
   - This fixes the mono-off `Array (Array char)` bug where the outer generic
     `*Array` boxed rows through `runtime.Value` and then lost row mutation
     identity on `rows[idx]!.push(...)`.
   - Corrected benchmark result: mono on `0.9567s` / `88.00` GC vs mono off
     `1.0500s` / `384.00` GC over 3 compiled runs.
   - Interpretation: with a valid mono-off baseline, the text-scalar slice is
     directionally correct and does not require a text-path rollback.
12. [x] Extend the staged specialized scalar set to the remaining primitive
    numeric family and benchmark a representative unsigned path.
   - Staged wrappers now also cover `Array i8`, `Array i16`, `Array u16`,
     `Array u32`, `Array u64`, and `Array f32`.
   - Focused regressions pin the generated wrappers for those types and cover
     dynamic-boundary callback conversion for `Array u32` / `Array f32`.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/sum_u32_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-mono-array-u32-sum-small-compiled.md`
   - Result: mono on `1.0933s` / `185.33` GC vs mono off `1.6800s` /
     `21.33` GC over 3 compiled runs.
   - Interpretation: the remaining primitive scalar family now shows a real
     compiled wall-clock win on a typed unsigned workload, even though GC
     count is not the improving metric on that benchmark.
13. [x] Land the first broader non-array native container slice for `HashMap`.
   - Static `HashMap K V` positions now use native `*HashMap` carriers instead
     of `any`.
   - Typed/untyped map literals, `Array (HashMap K V)` shells, and `Map K V`
     interface params now stay on the native lowering path.
   - Shared nominal struct lowering now expands simple aliases before host-type
     mapping, so alias-backed container fields like `HashMapHandle = i64` stay
     on the generic native struct path.
   - The explicit map-literal handle boundary now normalizes runtime handle
     values back into the native `HashMap.handle` carrier.
   - Stdlib `HashSet.iterator()` / `Enumerable.iterator` generic interface
     returns now use corrected generic matching plus an explicit narrowed
     runtime adapter roundtrip instead of compiler fallback.
   - The compiled control bridge now preserves exit signals before wrapping
     raised values, which closes the surfaced `HashSet.union` stdlib
     integration failure.
   - Reduced benchmark fixture:
     `v12/fixtures/bench/hashmap_i32_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-hashmap-i32-small-compiled.md`
   - Result: compiled 3-run average `1.7633s` / `175.33` GC.
   - Interpretation: this slice closes correctness/native-lowering gaps around
     broader containers and gives the performance program a checked-in
     `HashMap` target for future map/set tuning.
14. [x] Close the nominal tree/persistent container follow-through on the
    shared carrier pipeline.
   - `TreeMap` / `TreeSet` and `PersistentMap` / `PersistentSet` now compile
     through the same generic nominal struct/interface path used by `HashMap`
     instead of introducing new type-specific compiler rules.
   - Static compiled bodies now call generated builtin helpers
     (`__able_slice_len`, `__able_slice_cap`, `__able_string_len_bytes`) on
     these paths so Able locals like `len` no longer break Go compilation.
   - The previously failing compiled stdlib fixture gates
     `06_12_11_stdlib_collections_tree_map_set` and
     `06_12_12_stdlib_collections_persistent_map_set` are green again.
   - Interpretation: the next container category is no longer “make tree or
     persistent maps compile”; it is broader carrier reduction beyond the
     nominal map/set families already covered.
15. [x] Mechanically audit the next stdlib container families on that same
    shared native carrier path and add a reduced benchmark for the first hot
    representative.
   - Representative no-fallback regressions now pin static compiled method
     bodies for `Deque` / `Queue`, `BitSet` / `Heap`, and
     `PersistentSortedSet` / `PersistentQueue`.
   - Those tests assert native locals plus the absence of
     `__able_call_value(...)`, `__able_member_get_method(...)`,
     `__able_method_call_node(...)`, `bridge.MatchType(...)`, and
     `__able_try_cast(...)` in representative compiled methods.
   - Shared compiled fixture gates for
     `06_12_13_stdlib_collections_persistent_sorted_queue`,
     `06_12_16_stdlib_collections_deque_queue`, and
     `06_12_17_stdlib_collections_bit_set_heap` are green on the same shared
     lowering path.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/heap_i32_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-heap-i32-small-compiled.md`
   - Result: compiled 3-run average `7.7533s` / `1105.00` GC, direct compiled
     output `-211812354`.
   - Interpretation: the next category is no longer “make these families stay
     native”; it is deeper generic container paths and any remaining generic
     carrier fallbacks.
16. [x] Close the next deeper generic-container correctness slice on that same
    shared native path.
   - Generic nullable/interface carriers now remain native on deeper container
     shapes like `LazySeq.Source: ?(Iterator T)` instead of collapsing to
     `any`.
   - Compiled `nil` lowering now emits typed Go nils for native nilable
     carriers (`(*ListNode)(nil)`, `__able_iface_Iterator_T(nil)`, etc.)
     rather than invalid raw `nil` temps in generated code.
   - The compiled stdlib fixture gate
     `06_12_14_stdlib_collections_linked_list_lazy_seq` is green again on the
     shared container/interface carrier path.
   - Interpretation: the next category is no longer generic-container
     correctness; it is benchmark-worthy generic container hot paths and any
     remaining residual carrier fallbacks.
17. [x] Close the next iterator default-method hot path on that same shared
    native path.
   - Ordinary default native-interface methods now lower through the direct
     compiled-helper path on native iterator carriers instead of the runtime
     adapter method layer.
   - The representative
     `LinkedList.lazy().map<i64>(...).filter(...).next()` path now stays on
     compiled iterator helpers end-to-end.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-pipeline-i64-small-compiled.md`
   - Result: compiled 3-run average `0.1800s` / `13.33` GC, direct compiled
     output `382455000`.
18. [x] Close the mono-array-enabled `Iterator.collect<Array T>()` /
    specialized-array accumulator interaction on this iterator family.
  - `Iterator.collect<Array i64>()` now lowers through a generated compiled
    helper with a specialized `*__able_array_i64` accumulator instead of the
    residual `__able_method_call_node(...)` + `__able_array_i64_from(...)`
    bridge.
  - The shared generic default-method path now has an explicit user-defined
    nominal proof case too: `Iterator.collect<C>()` stays native for a
    `Default + Extend` accumulator struct (`SumCount`) without adding a new
    named-container rule. The Array helper is now documented as the built-in
    fallback behind that shared generic path, not the primary model.
  - The compiler-side fix is intentionally array-specific because `Array` is
    a language/kernel special form; this does not change the broader
    no-per-nominal-special-casing rule for ordinary user-defined structs.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-collect-i64-small-compiled.md`
   - Result: compiled 3-run averages are flat on this reduced workload:
     mono on `0.1833s` / `14.00` GC versus mono off `0.1833s` / `13.33` GC,
     with direct compiled output `382455000`.
   - Interpretation: this tranche closes a correctness/native-carrier gap
     rather than introducing a new speed step.
19. [x] Close the iterator-literal controller/runtime-value edge that remained
    inside the same iterator family.
   - Compiled iterator literals now bind `gen` as a compiler-owned
     `*__able_generator` controller instead of a generic runtime object.
   - `gen.yield(...)`, `gen.stop()`, and bound `gen.yield` callable captures
     now lower directly through that controller path instead of
     `__able_method_call_node(...)`.
   - Native nilable/static-carrier conditions now lower as direct nil checks,
     which keeps `Iterator.filter_map` on the static path.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-filter-map-i64-small-compiled.md`
   - Result: compiled 3-run average `0.1267s` / `10.00` GC, direct compiled
     output `191952000`.
   - Interpretation: the iterator default-method family is now closed through
     `filter_map`; the next category is the next hot generic-container/runtime
     edge beyond iterator-literal controller cleanup.
20. [x] Specialize shared generic nominal `methods` when the concrete target
    type is statically known.
  - Added shared nominal-method specialization parallel to the existing impl
    specialization path, so statically known targets now get specialized
    compiled method bodies instead of reusing unspecialized
    `runtime.Value` signatures.
  - Added a user-defined proof case (`Box T`) to pin that this is a general
    nominal-lowering rule, not a container-specific exception.
  - Re-measured the first hot constrained-generic nominal-method benchmark:
    `v12/fixtures/bench/heap_i32_small/main.able`
  - Snapshot:
    `v12/docs/perf-baselines/2026-03-20-heap-i32-generic-nominal-method-specialization-compiled.md`
  - Result: compiled 3-run average `4.2000s` / `1811.67` GC, direct compiled
    output `-211812354`.
21. [x] Refine bound generic field/member carriers inside already-specialized
    nominal method bodies.
  - Added a user-defined proof case (`Bucket T { items: Array T }`) and pinned
    it under `ExperimentalMonoArrays`, which proves fully bound generic fields
    now stay on their concrete native carriers (`Items *__able_array_i32`)
    inside specialized nominal method bodies.
  - Static array index lowering now returns concrete element types directly
    when an expected type is known, with explicit control transfer on bounds
    failures, instead of materializing a temporary `runtime.Value`.
  - Receiver-derived sibling/default impl bindings now upgrade placeholder
    self-bindings like `T -> T` to concrete bindings like `T -> i64`, which
    closes the remaining mono-array `Iterable.iterator` / `Iterable.each`
    execute gap and the surfaced `PersistentSortedQueue` specialization miss.
  - Re-measured:
    `v12/fixtures/bench/heap_i32_small/main.able`
  - Snapshot:
    `v12/docs/perf-baselines/2026-03-20-heap-i32-bound-generic-field-carrier-refinement-compiled.md`
  - Result: compiled 3-run average `0.7667s` / `91.33` GC, direct compiled
    output `-211812354`.
  - Interpretation: this shared field/member carrier tranche is now closed.
22. [x] Close the remaining shared static nominal receiver/struct-literal gap
    on the reduced `LinkedList -> Enumerable -> LazySeq` family.
  - Recursive type substitution now resolves chained specialization bindings
    transitively, expected-type refinement upgrades static nominal targets and
    struct literals to their concrete specialized carrier, and native
    interface concrete-impl matching now recognizes specialized receivers
    through the shared target-template path.
  - This closes the surfaced residual fallback gap for
    `ConcreteEnumerableGenericMethods...`,
    `LinkedListIterableAdapter...`, and
    `LazySeqIteratorCarrier...` without adding a `LinkedList`-specific or
    `LazySeq`-specific lowering rule.
  - Re-measured:
    `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`
  - Snapshot:
    `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-shared-static-nominal-closure-compiled.md`
  - Result: compiled 3-run average `0.1633s` / `8.33` GC, direct compiled
    output `382455000`.

## Next Category

The next monomorphization/container-lowering category is now beyond the
audited stdlib families already covered, beyond the generic-container
correctness fixes, and beyond the first benchmark-worthy generic-container hot
path (`LinkedList -> Iterable -> Iterator`) and the next concrete
generic/default-method slice (`LinkedList.map/filter/reduce`) and its
follow-up callback/runtime carrier cleanup and ordinary default iterator-method
slice and the mono-array-enabled iterator-collect closure and the shared
generic nominal-method specialization tranche and the bound generic
field/member carrier refinement tranche and the shared static
nominal receiver/struct-literal closure tranche:

1. Identify the next benchmark-worthy generic container/runtime edge that
   still crosses residual runtime carriers.
2. Then broaden performance-oriented carrier reduction across the next
   benchmark-worthy generic container/runtime edges that are still hot enough
   to matter.
