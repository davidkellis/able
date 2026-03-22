# Monomorphized Container ABI (Compiler-Native Arrays, Go AOT)

Date: 2026-03-19
Owners: Able compiler/runtime contributors

## Purpose

Define the accepted direction for monomorphized `Array<T>` lowering in the Go
AOT compiler.

This note supersedes the earlier runtime-typed-store / handle-tag rollout plan
as the target architecture. Existing flag plumbing and historical experiments
may remain in-tree temporarily, but new implementation work must converge on
compiler-owned native Go storage on static compiled paths.

Named stdlib/container examples in this document are proof cases for shared
lowering machinery, not architecture exceptions.

This design supports:

- the v12 AOT no-fallback contract,
- the compiler native-lowering contract,
- reduction of `runtime.Value` boxing on hot compiled array paths,
- performance work that does not reintroduce interpreter/runtime structural
  carriers as the default compiled representation.

## Status

Current repo state:

- `--experimental-mono-arrays` / `ABLE_EXPERIMENTAL_MONO_ARRAYS` plumbing
  exists.
- historical experimental work added typed runtime stores plus boundary tests.
- static compiled array lowering today already stays on the compiler-owned
  `Array` carrier for direct paths, but that carrier still stores elements as
  `[]runtime.Value`.

Accepted direction after the 2026-03-19 audit:

- the old typed-runtime-store / `storage_handle`-tag plan is not the final ABI;
- future mono-array work must specialize compiler-owned array carriers over
  native Go element storage;
- `runtime.ArrayValue`, `ArrayStore*`, and runtime typed stores are boundary /
  residual machinery only, not the representation that static compiled code
  should manipulate directly.

## Non-Negotiable Constraints

1. Static compiled array values must stay on compiler-owned native Go storage.
2. Explicit dynamic boundaries may adapt to/from runtime carriers, but those
   adapters must remain narrow and boundary-only.
3. Source-visible Able semantics cannot change:
   - aliasing,
   - mutation visibility,
   - bounds/error behavior,
   - cloning semantics,
   - `length` / `capacity` observability,
   - iteration order.
4. The compiler must not silently fall back to interpreter execution or broad
   dynamic helper dispatch when a static mono-array path is representable.
5. Mono-array work must preserve the broader native-lowering contract:
   no new IIFEs, no panic/recover for ordinary flow, and no regression to broad
   `runtime.Value` dispatch helpers on static paths.

## Rejected Direction

The following are explicitly rejected as the final mono-array architecture:

- tagged `storage_handle` routing between legacy and typed runtime stores as
  the primary compiled representation;
- lowering static compiled `Array<T>` to `runtime.ArrayValue` or `ArrayStore*`
  APIs by default;
- keeping mono-array values in runtime-managed typed stores and treating the
  compiler as a thin handle emitter;
- widening the existing experimental runtime-store path instead of migrating
  toward compiler-owned native Go storage.

Historical runtime-store work can still be used as compatibility scaffolding or
measurement reference, but it is not the end state and should not receive new
feature surface area unless required strictly for boundary compatibility.

## Target Representation

For statically monomorphic arrays, the compiler should emit specialized,
compiler-owned wrapper types selected by the resolved element type.

Examples:

| Able type | Target compiled representation |
| --- | --- |
| `Array i32` | `*__able_array_i32` wrapping `[]int32` |
| `Array i64` | `*__able_array_i64` wrapping `[]int64` |
| `Array bool` | `*__able_array_bool` wrapping `[]bool` |
| `Array u8` | `*__able_array_u8` wrapping `[]byte` |
| `Array char` | `*__able_array_char` wrapping `[]rune` |
| `Array String` | `*__able_array_String` wrapping `[]string` |
| `Array Point` | later `*__able_array_Point` or `[]*Point` wrapper |

Each specialized wrapper should preserve the spec-visible metadata fields the
current compiler already exposes on the generic carrier:

- `Length int32`
- `Capacity int32`
- `Storage_handle int64`

The element backing storage, however, must be native Go storage for the element
kind, not `[]runtime.Value`.

## ABI Shape

### Specialized wrapper form

Illustrative shape:

```go
type __able_array_i32 struct {
    Length         int32
    Capacity       int32
    Storage_handle int64
    Elements       []int32
}
```

Notes:

- wrapper naming is compiler-internal; exact identifier spelling is not part of
  the language contract;
- the wrapper is the static compiled representation;
- `Storage_handle` remains metadata for compatibility and explicit boundary
  synchronization, not the authority for static storage.

### Alias semantics

Aliasing must be pointer-based:

- `b = a` shares the same specialized wrapper pointer;
- mutation through any alias updates the same backing slice;
- metadata synchronization updates the shared wrapper in place.

### Generic / unspecialized fallback

If element type is not yet in the staged specialization set or cannot be proven
statically, the compiler may remain on the generic compiler-owned `*Array`
carrier with `[]runtime.Value` elements.

That fallback is acceptable only as a residual static compiler path while
specialization coverage expands. It does not justify routing specialized arrays
through runtime carriers.

## Lowering Rules

When the mono-array flag is enabled and `Array<T>` resolves to a staged element
kind:

1. Array literals should emit the specialized wrapper directly.
2. `push`, `pop`, `get`, `set`, indexing, `len`, `capacity`, `clear`, cloning,
   and iteration should operate on the native typed slice.
3. Pattern matching and destructuring over specialized arrays should remain on
   the specialized wrapper until an explicit dynamic boundary is entered.
4. Numeric paths should avoid boxing into `runtime.Value` unless the type of the
   surrounding expression requires it.
5. Conversion back to `runtime.Value` should happen only:
   - at explicit dynamic boundaries,
   - at ABI wrappers,
   - at residual runtime-polymorphic surfaces that are still intentionally
     unspecialized.

## Boundary Rules

Boundary adapters are allowed only at explicit dynamic/ABI edges.

Required adapter directions:

- specialized wrapper -> `runtime.ArrayValue`
- `runtime.ArrayValue` -> specialized wrapper
- specialized wrapper -> generic compiler `*Array` carrier when crossing into a
  still-unspecialized compiled path
- generic compiler `*Array` carrier -> specialized wrapper when re-entering a
  proven-safe mono-array path

Boundary adapters must:

- preserve aliasing semantics where required by the receiving side,
- preserve bounds/error behavior,
- preserve mutation writeback when the runtime side mutates the array,
- avoid leaking runtime handles back into the default static compiled path.

## Implementation Order

1. Freeze the revised architecture in docs and guardrail tests.
2. Add compiler-generated specialized wrapper emission for the first staged
   element set (`i32`, `i64`, `bool`, `u8`).
3. Move hot intrinsics and direct index read/write onto specialized wrappers.
4. Add explicit boundary adapters between specialized wrappers and runtime
   carriers.
5. Extend iteration/pattern/destructuring/clone paths.
6. Re-benchmark compiled mode and decide whether the flag behavior/default
   should change.

## Guardrails

The following must remain true while the implementation is in progress:

- static compiled mono-array function bodies must not regress to
  `runtime.ArrayValue`, `runtime.ArrayStore*`, `__able_call_value(...)`,
  `__able_method_call_node(...)`, or dynamic index helpers for representable
  specialized paths;
- the old runtime-store experiment should not gain new surface area unless a
  change is strictly boundary-related;
- every new mono-array lowering slice must come with targeted compiler tests and
  dynamic-boundary tests.

## Validation Gates

Correctness:

- compiler static no-fallback audits remain green;
- zero-explicit-boundary native fixture audits remain green;
- dynamic-boundary mono-array tests continue to prove explicit boundary
  conversion behavior.

Performance:

- compiled benchmark deltas have now been re-measured after the widened
  specialized-wrapper slice:
  - snapshot: `v12/docs/perf-baselines/2026-03-19-mono-array-widened-compiled.md`
  - `bench/noop`: flat at `0.0100s`, `0.00` GC
  - `bench/sieve_count`: flat at `0.0100s`, `0.00` GC
  - `bench/sieve_full`: flat at `0.0200s`, but timed GC dropped from `3.00`
    to `1.00`
  - implication: the widened slice is reducing allocation pressure, but the
    next visible runtime win requires shrinking residual generic carriers
    rather than just widening the existing explicit-typed slice
- the first residual generic-array cleanup slice is now closed too:
  - inferred local `Array T` bindings retain recoverable element-type metadata
    on static paths, so generic helper results such as `get`, `pop`, `first`,
    `last`, and `read_slot` can stay on native nullable carriers where
    representable;
  - static and residual runtime-backed array `set` / index-assignment success
    now return `nil` again, restoring parity with the `nil | IndexError`
    contract;
  - runtime-backed iterator interface carriers now accept raw
    `*runtime.IteratorValue` payloads directly, and generator stop stays
    iterator completion instead of surfacing as a generic runtime error;
  - result: `06_08_array_ops_mutability`,
    `06_12_02_stdlib_array_helpers`, and
    `06_12_18_stdlib_collections_array_range` are green again on the current
    compiler-native path
- the staged specialized wrapper set now includes `f64`, and the nested
  `Array (Array f64)` path is closed too:
  - `Array f64` now lowers to `*__able_array_f64`;
  - pointer-backed nullable propagation now supports concrete expected-type
    coercion for carriers like `*float64`, which keeps nested
    `rows.get(j)!.get(i)!` expressions on the static path;
  - the full compiled `examples/benchmarks/matrixmultiply.able` benchmark no
    longer fails early under mono arrays; it now times out in parity with
    mono-off at the current `60s` harness limit.
- a reduced checked-in benchmark fixture now provides an apples-to-apples
  measurement target for the staged `f64` slice:
  - `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  - compiled 3-run averages:
    - mono on: `5.4833s`, `280.00` GC
    - mono off: `45.3133s`, `3568.67` GC
- the remaining generic outer row shell is now gone on the representative
  nested staged path too:
  - `Array (Array f64)` now lowers to `*__able_array_array_f64` over
    `[]*__able_array_f64` instead of `*Array` with `[]runtime.Value` rows;
  - rendered mono-array converters and native `Array` core helpers now treat
    pointer-backed specialized elements as first-class carriers, including nil
    propagation for `read_slot` / `pop` and explicit nested runtime boundary
    conversion;
  - post-outer-wrapper compiled 3-run averages on
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`:
    - mono on: `5.7233s`, `252.00` GC
    - mono off: `44.5167s`, `3550.67` GC
  - implication: the architectural cleanup is complete for this nested row
    case, but further wall-clock movement now depends on broader carrier
    reduction beyond nested mono-array wrappers.
- the next shared built-in `Array` scalar-lowering step is now landed too:
  - static staged-array propagation returns concrete success element types on
    the compiled path instead of routing scalar success values through
    `__able_nullable_*_to_value(...)`;
  - primitive numeric casts such as `i32 -> f64` now lower directly to Go
    casts on static compiled paths instead of `__able_cast(...)`;
  - shared primitive `float -> int` casts now also lower natively through
    truncate/range/overflow checks instead of `__able_cast(...)` and
    `bridge.AsInt(...)`, which removes the remaining obvious runtime cast
    helper use from the full matrix entry path;
  - shared static built-in `Array` factories and intrinsics now lower without
    synthetic `__able_push_call_frame(...)` / `__able_pop_call_frame()`
    scaffolding on compiled static paths;
  - propagated static built-in `Array` accessors (`get`, `first`, `last`,
    `read_slot`, `pop`) now lower as direct bounds-check + element-load paths
    with nil control transfer instead of manufacturing pointer-backed nullable
    carriers on the success path;
  - the reduced matrix benchmark
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
    now measures `0.1967s` / `7.33` GC over 3 compiled runs;
  - the full macro benchmark `v12/examples/benchmarks/matrixmultiply.able`
    now measures `3.4367s` / `13.67` GC over 3 compiled runs;
  - implication: the reduced nested numeric hot loop is now back on the
    native path, the primitive entry-cast gap is closed, and the propagated
    static-array accessor pointer-carrier gap is closed too.
- the next shared primitive/control-flow step is now landed too:
  - canonical primitive counted loops of the form
    `loop { if i >= n { break } ... i = i + 1 }` now lower to direct Go
    `for i < n { ... i++ }` loops on compiled static paths;
  - the matcher is conservative by construction and traverses nested
    function/lambda/iterator/ensure bodies before enabling the fast path;
  - the reduced matrix benchmark
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
    now measures `0.1133s` / `7.00` GC over 3 compiled runs;
  - the full macro benchmark `v12/examples/benchmarks/matrixmultiply.able`
    now measures `1.0833s` / `13.00` GC over 3 compiled runs;
  - implication: loop-control checked arithmetic is no longer the limiting
    primitive residual on this matrix family; the remaining shared primitive
    target is affine checked integer arithmetic inside `build_matrix`.
- the next shared primitive affine-arithmetic step is now landed too:
  - fixed-width primitive checked `+` / `-` under 64 bits now lower inline on
    static compiled paths instead of calling the checked helper functions;
  - `int`, `uint`, `i64`, and `u64` intentionally remain on the helper path
    because they still depend on the wider-width/runtime-width overflow
    machinery;
  - `build_matrix` no longer carries
    `__able_checked_add_signed(...)` / `__able_checked_sub_signed(...)`; its
    affine integer ops now lower as inline `int64(...) +/- int64(...)` plus
    explicit range checks;
  - the reduced matrix benchmark remains `0.1133s` / `7.00` GC over 3 runs;
  - the full macro benchmark remains `1.0867s` / `13.00` GC over 3 runs;
  - implication: helper-call overhead is no longer the limiting primitive
    residual on this matrix family; the next target is eliminating provably
    unnecessary inline overflow branches via shared range-proof lowering.
- shared primitive sign/range facts now remove the subtraction-side branch
  where both fixed-width signed operands are proven non-negative:
  - the compile context now tracks simple non-negative facts per Go binding
    and carries them into child scopes while clearing them on
    rebinding/shadowing;
  - `build_matrix` now lowers `i - j` as a direct signed subtraction instead
    of widened `int64(...)` subtraction plus overflow-branch scaffolding;
  - the reduced matrix benchmark remains `0.1167s` / `7.00` GC over 3 runs;
  - the full macro benchmark remains `1.1000s` / `13.00` GC over 3 runs;
  - implication: subtraction-side overflow branching is now closed, and the
    remaining primitive matrix residual is stronger upper-bound proofs for
    affine addition like `i + j`.
- shared primitive upper-bound propagation now removes the addition-side
  branch too when fixed-width signed operands are proven non-negative and
  bounded:
  - the compiler now carries simple primitive upper-bound facts across
    statically resolved function calls and seeds them back into callee param
    contexts before render;
  - counted-loop induction variables now inherit a conservative upper bound
    from their loop guard when the bound is statically known;
  - `build_matrix` now lowers both `i - j` and `i + j` as direct signed
    arithmetic with no widened affine branch scaffolding left in the inner
    loop;
  - the reduced matrix benchmark remains `0.1267s` / `7.00` GC over 3 runs;
  - the full macro benchmark remains `1.1367s` / `13.00` GC over 3 runs;
  - implication: the hot affine integer residual on the matrix path is now
    closed, so further work there is a different category than loop-affine
    primitive arithmetic.
- compiler-owned array wrapper synthesis now covers broader native carrier
  element families beyond nested mono arrays too:
  - generic inner arrays that are still genuinely generic on the inner level
    (for example `Array (Array (HashMap String i32))` when the inner array is
    not otherwise representable as a staged specialized wrapper)
  - native interface carriers (`Array Greeter` ->
    `*__able_array_iface_Greeter`)
  - native callable carriers (`Array (i32 -> i32)` ->
    `*__able_array_fn_int32_to_int32`)
  - other representable pointer-backed carrier families sharing the same
    explicit runtime adapter rules
  - dynamic-boundary callback coverage now includes interface-carrier arrays
    and callable-carrier arrays explicitly
- the staged text scalar family is now specialized too:
  - `Array char` -> `*__able_array_char` over `[]rune`
  - `Array String` -> `*__able_array_String` over `[]string`
  - representative nested text rows (`Array (Array char)`) now lower to
    `*__able_array_array_char`
  - native result propagation now keeps `!Array char` on the static carrier
    path instead of re-entering the union through `_from_value(...)`
- a reduced checked-in text benchmark now measures that slice directly:
  - `v12/fixtures/bench/zigzag_char_small/main.able`
  - compiled 3-run averages:
    - mono on: `0.9567s`, `88.00` GC
    - mono off: `1.0500s`, `384.00` GC
  - note: the earlier mono-off comparison for this fixture was invalid because
    the mono-off compiled path was breaking nested row identity on
    `Array (Array char)` by boxing rows through `runtime.Value`
  - implication: with carrier-array identity preserved on mono-off, the text
    specialization slice is directionally correct and does not need rollback;
    the next work is broader carrier reduction again
- the remaining primitive numeric scalar family is now staged too:
  - `Array i8` -> `*__able_array_i8` over `[]int8`
  - `Array i16` -> `*__able_array_i16` over `[]int16`
  - `Array u16` -> `*__able_array_u16` over `[]uint16`
  - `Array u32` -> `*__able_array_u32` over `[]uint32`
  - `Array u64` -> `*__able_array_u64` over `[]uint64`
  - `Array f32` -> `*__able_array_f32` over `[]float32`
- a reduced checked-in unsigned benchmark now measures that slice directly:
  - `v12/fixtures/bench/sum_u32_small/main.able`
  - compiled 3-run averages:
    - mono on: `1.0933s`, `185.33` GC
    - mono off: `1.6800s`, `21.33` GC
  - implication: typed unsigned specialization is now clearly faster on wall
    clock, even though GC count is not the metric improving on this workload
- broader non-array native container lowering has started too:
  - static `HashMap K V` positions now use native `*HashMap` carriers instead
    of `any`
  - typed/untyped map literals, `Array (HashMap K V)` shells, and `Map K V`
    interface views now stay on native compiler carriers
  - the shared nominal struct path now expands simple aliases before host-type
    mapping, so alias-backed container fields keep native host carriers
    instead of reintroducing `runtime.Value` through field lowering
  - explicit map-literal handle conversion is now narrowed to the boundary that
    rehydrates the native `HashMap.handle` carrier
  - stdlib `HashSet.iterator()` / `Enumerable.iterator` generic interface
    returns now use corrected generic matching plus an explicit narrowed
    runtime adapter roundtrip instead of compiler fallback
  - compiled exit-signal propagation now survives the control bridge on this
    path, so successful stdlib container flows that end in `__able_os_exit(0)`
    do not degrade into generic runtime failures
  - reduced benchmark fixture:
    `v12/fixtures/bench/hashmap_i32_small/main.able`
  - compiled 3-run average: `1.7633s`, `175.33` GC
  - `TreeMap` / `TreeSet` and `PersistentMap` / `PersistentSet` now also
    compile through the same shared nominal carrier path instead of requiring
    more container-specific compiler exceptions
  - static compiled bodies on these paths now use generated builtin helpers
    (`__able_slice_len`, `__able_slice_cap`, `__able_string_len_bytes`) so
    Able locals cannot shadow raw Go builtins and break container-heavy code
  - the next stdlib container families are now mechanically audited on that
    same shared path too: representative compiled methods for `Deque` /
    `Queue`, `BitSet` / `Heap`, and `PersistentSortedSet` /
    `PersistentQueue` are pinned by no-fallback regressions that reject
    `__able_call_value(...)`, `__able_member_get_method(...)`,
    `__able_method_call_node(...)`, `bridge.MatchType(...)`, and
    `__able_try_cast(...)`
  - reduced benchmark fixture:
    `v12/fixtures/bench/heap_i32_small/main.able`
  - compiled 3-run average: `7.7533s`, `1105.00` GC
  - the next deeper generic-container correctness slice is now closed too:
    `LazySeq.Source: ?(Iterator T)` stays on a generated native interface
    carrier, and compiled nil lowering now emits typed Go nils for native
    nilable carriers instead of raw `nil` temps
  - implication: the container-lowering program has moved beyond arrays,
    beyond the first nominal map/set families, beyond the next stdlib
    container audit slice, and beyond the generic-container correctness gap;
    the first benchmark-worthy generic-container hot path is now closed too:
    `LinkedList -> Iterable -> Iterator` stays native because inherited
    interface impls now synthesize native base-interface adapters and concrete
    interface adapters directly coerce compatible native interface returns.
  - reduced benchmark fixture:
    `v12/fixtures/bench/linked_list_for_i32_small/main.able`
  - compiled 3-run average: `0.2000s`, `15.00` GC
  - implication: the next widening slices should target broader
    performance-oriented generic container/runtime edges and only benchmark
    surfaces that are actually hot enough to measure
  - concrete generic/default container-method slice now closed too:
    `LinkedList.map/filter/reduce` resolves through compiled `Enumerable`
    impls, binds `C` to the concrete target on compiled impl paths, and treats
    native `Iterator<T>` carriers as directly iterable inside those default
    impl bodies
  - reduced benchmark fixture:
    `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`
  - compiled 3-run average: `0.1667s`, `12.00` GC
  - follow-up callback/runtime carrier cleanup on that same hot path is now
    closed too:
    - specialized impls now retain bound type bindings through compileability
      and render
    - specialized sibling selection inside default impl bodies now wins before
      the ordinary concrete receiver path
    - specialized `Enumerable.lazy()` helpers now call
      `iterator_*_spec(...)` directly instead of bridging
      `Iterator_A -> runtime.Value -> Iterator_T`
  - the next iterator default-method slice is now closed too:
    `LinkedList.lazy().map<i64>(...).filter(...).next()` stays on compiled
    iterator helpers because ordinary default native-interface methods now
    lower through the direct compiled-helper path on native iterator carriers
  - reduced benchmark fixture:
    `v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able`
  - compiled 3-run average: `0.1800s`, `13.33` GC
  - the mono-array-enabled `Iterator.collect<Array T>()` follow-up is now
    closed too:
    - the compiler emits a generated compiled helper with a specialized
      mono-array accumulator instead of the old residual specialized-array
      bridge
    - shared generic lowering is still primary here: `Iterator.collect<C>()`
      now has a user-defined nominal proof case (`SumCount`) on the ordinary
      `Default + Extend` path, and the Array helper remains only as the
      built-in fallback
    - reduced benchmark fixture:
      `v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able`
    - compiled 3-run averages: mono on `0.1833s`, `14.00` GC; mono off
      `0.1833s`, `13.33` GC
  - the iterator-literal controller/runtime-value edge is now closed too:
    - compiled iterator literals bind `gen` as a compiler-owned
      `*__able_generator`
    - `gen.yield(...)`, `gen.stop()`, and bound `gen.yield` callable captures
      now lower directly instead of routing through dynamic member dispatch
    - native nilable/static-carrier conditions now lower as direct nil checks,
      which keeps `Iterator.filter_map` on the static path
    - reduced benchmark fixture:
      `v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able`
    - compiled 3-run average: `0.1267s`, `10.00` GC
  - shared generic nominal `methods` specialization is now closed too:
    - statically known concrete nominal targets now get specialized compiled
      method bodies on the shared nominal-method path instead of reusing
      unspecialized `runtime.Value` signatures
    - this has both a user-defined proof case (`Box T`) and a hot reduced
      benchmark proof case (`Heap i32`) without adding another
      nominal-type-specific lowering rule
    - reduced benchmark fixture:
      `v12/fixtures/bench/heap_i32_small/main.able`
    - compiled 3-run average after the tranche: `4.2000s`, `1811.67` GC
  - bound generic field/member carrier refinement is now closed too:
    - fully bound generic fields like `self.items: Array T` now stay on their
      concrete native carriers once `T` is known
    - user-defined proof case: `Bucket T { items: Array T }` now renders
      `Items *__able_array_i32` plus specialized method bodies under
      `ExperimentalMonoArrays`
    - receiver-derived sibling/default impl bindings now upgrade placeholder
      self-bindings like `T -> T` to concrete bindings like `T -> i64`,
      closing the remaining mono-array `Iterable.iterator` / `Iterable.each`
      execute gap
    - reduced benchmark fixture:
      `v12/fixtures/bench/heap_i32_small/main.able`
    - compiled 3-run average after the tranche: `0.7667s`, `91.33` GC
  - the remaining shared static nominal receiver/struct-literal gap on the
    reduced `LinkedList -> Enumerable -> LazySeq` family is closed too:
    - recursive type substitution now resolves chained specialization
      bindings transitively
    - static nominal target refinement now upgrades bare targets and
      `LazySeq { ... }` literals to the concrete specialized carrier when the
      expected type is already known
    - native interface concrete-impl matching now recognizes specialized
      receivers through the shared target-template path rather than falling
      back
    - reduced benchmark fixture:
      `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`
    - compiled 3-run average after the tranche: `0.1633s`, `8.33` GC
  - implication: the next widening slice on this family is now beyond
    iterator collect, iterator-controller correctness, nominal-method
    signature specialization, and bound generic field/member carrier
    refinement, and the shared static nominal receiver/struct-literal
    closure; the next target is the next benchmark-worthy generic
    container/runtime edge beyond the families already closed
- historical runtime-store perf numbers are reference data only, not proof that
  the accepted architecture is complete.

## Open Questions

1. Should specialized wrapper names be fully concrete per element type, or share
   a generated internal generic helper pattern where Go typing allows it?
2. What is the minimal staged element set after `i32`, `i64`, `bool`, `u8`,
   `f64`
   that yields meaningful wins without overcomplicating codegen?
3. How much of pattern/destructuring lowering should be specialized before
   widening to maps/sets/other containers?

## Commit-Ready Checklist

- [x] Revise the mono-array design so the final target is compiler-owned native
      Go storage rather than runtime typed stores / handle tagging.
- [x] Add a compiler guardrail test proving experimental mono-array static paths
      still stay on the compiler-owned array carrier today.
- [x] Emit specialized wrapper types for the first staged element set.
- [x] Lower the first direct typed hot-path slice onto those specialized
      wrappers (`literal`, `push`, `get`, `set`, `read_slot`, `write_slot`,
      direct index read/write).
- [x] Add the first explicit specialized-wrapper <-> runtime carrier adapters
      needed by wrappers, lambdas, callables, interfaces, unions, and struct
      field conversion.
- [x] Widen specialization beyond explicit typed positions to constructors /
      stdlib factories, unannotated locals, clone / iteration / pattern paths,
      and other residual generic `*Array` surfaces.
- Widened slice details:
  - non-empty unannotated local array literals now infer staged specialized
    carriers when the element family is staged;
  - `Array.new()` / `Array.with_capacity()` now lower directly to
    compiler-owned static carriers on typed static paths;
  - `reserve()` / `clone_shallow()`, static array `for` iteration, and
    array-pattern rest tails now preserve specialized carriers instead of
    dropping back to generic `*Array`.
- [x] Re-run compiled perf snapshots after specialized-wrapper lowering lands.
- [x] Close the first residual generic-array cleanup slice so the known
      compiler/runtime fixture mismatches are gone before the next benchmark
      pass.
- [x] Eliminate the representative nested outer-row generic shell so arrays of
      staged mono-array carriers stay compiler-owned end-to-end on static
      paths.
- [x] Widen compiler-owned array wrapper synthesis beyond nested mono arrays to
      broader native carrier element families already owned by the compiler.
