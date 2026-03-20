# Compiler Native Lowering

## Status

Active design constraint for the v12 compiler. This document records the target
architecture and supersedes ad hoc decisions that keep static compiled code too
close to interpreter/runtime object-model carriers.

## Vision

The compiler should lower Able programs to native Go constructs whenever the
semantics are statically representable.

The end state is not "compiled code that still manipulates interpreter values
more efficiently." The end state is "compiled code that primarily manipulates
Go-native values and uses dynamic carriers only at explicit dynamic
boundaries."

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
  carrier while narrowing the runtime crossing to the explicit generic
  dispatch edge.
- That generic dispatch edge is intentionally narrow: compiled code converts
  the native receiver to runtime only for `__able_method_call_node(...)`,
  writes back pointer-backed native carriers after the runtime call, and then
  converts the returned `runtime.Value` back into the best-known native Go
  carrier before continuing on the compiled static path.
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
  at zero. The intentionally residual generic-interface edge is audited
  separately to stay narrowed to `__able_iface_*_to_runtime_value(...)` plus
  `__able_method_call_node(...)`.
- The broader compiler re-audit tranche is now closed too: the last surfaced
  native array mismatch under the zero-boundary harness was
  `runtime.ErrorValue` -> anonymous synthetic struct conversion dropping the
  wrapped `IndexError` definition, which broke static `case _: IndexError`
  matches on array bounds results. `__able_error_to_struct(...)` now preserves
  concrete wrapped struct payloads before falling back to the synthetic error
  view, and the zero-boundary fixture audit now includes
  `06_08_array_ops_mutability`.
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
- The remaining primitive numeric scalar family is now staged too:
  - `Array i8`, `Array i16`, `Array u16`, `Array u32`, `Array u64`,
    `Array isize`, `Array usize`, and `Array f32` now lower to compiler-owned
    wrappers on the staged specialized path
  - reduced unsigned benchmark:
    `v12/fixtures/bench/sum_u32_small/main.able`
  - compiled 3-run averages: mono on `1.0933s`, `185.33` GC; mono off
    `1.6800s`, `21.33` GC
- Remaining work is now broader non-scalar/container carrier reduction again,
  plus wider specialization/monomorphization on top of the now mostly-native
  carrier ABI.

## Relationship To Other Design Notes

- `compiler-monomorphization.md` is a subordinate design note for generic/native
  container lowering.
- `compiler-no-panic-flow-control.md` is a subordinate design note for explicit
  control-result propagation.
