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
- Remaining interface/existential work is now a different category:
  non-object-safe/generic interface existentials, callable/function-type
  existentials, and the residual runtime-boundary surfaces that still
  legitimately require dynamic carriers.

## Relationship To Other Design Notes

- `compiler-monomorphization.md` is a subordinate design note for generic/native
  container lowering.
- `compiler-no-panic-flow-control.md` is a subordinate design note for explicit
  control-result propagation.
