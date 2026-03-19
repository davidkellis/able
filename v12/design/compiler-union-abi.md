# Compiler Union ABI

## Status

Active compiler-native-lowering workstream with multiple landed slices.

This note defines the target ABI for compiled Able unions and records the
current staged limitations so implementation can proceed in ordered steps.

## Current State

The compiler is not yet lowering general unions natively.

Current lowering facts:

- `types.go` now maps the initial native union carrier family to generated Go
  carrier interfaces:
  - the original closed two-member nominal slice;
  - broader multi-member nominal unions;
  - generic alias unions that normalize to native nullable/result carrier
    families;
  - broader interface/open unions that can keep non-native payload branches as
    explicit residual `runtime.Value` union members.
- `types.go` now maps the first native `ResultTypeExpression` slice to those
  same generated carriers when the result shape normalizes to the current
  `runtime.ErrorValue | T` form, but broader result shapes that require deeper
  existential/interface lowering still lower to `any`.
- plain `Error` type positions now also map to `runtime.ErrorValue` on
  compiled static paths, but this is still a narrow special-case rather than a
  general native interface-carrier ABI.
- `?Error` now also maps to `*runtime.ErrorValue` on compiled static paths,
  extending that same narrow special-case into the nullable/error-recovery
  surface without solving arbitrary interface existentials yet.
- the narrow native `Error` special-case cleanup is now complete: direct
  compiled `Error.message()` lowers to `runtime.ErrorValue.Message`, direct
  compiled `Error.cause()` lowers to native payload extraction plus narrow
  nullable-error coercion only when needed, and compiled concrete-error
  normalization now preserves both message and cause payloads when
  constructing `runtime.ErrorValue`.
- `mapNullableType(...)` now keeps pointer/slice carriers native and the full
  compiler-native scalar nullable family native too:
  - `?bool -> *bool`
  - `?String -> *string`
  - `?char -> *rune`
  - `?f32 -> *float32`
  - `?f64 -> *float64`
  - `?isize -> *int`
  - `?usize -> *uint`
  - `?i8 -> *int8`
  - `?i16 -> *int16`
  - `?i32 -> *int32`
  - `?i64 -> *int64`
  - `?u8 -> *uint8`
  - `?u16 -> *uint16`
  - `?u32 -> *uint32`
  - `?u64 -> *uint64`
- Typed patterns and wrapper coercions no longer fall back to dynamic
  `bridge.MatchType(...)` / `__able_try_cast(...)` for the landed native union
  carrier family above, but they still do for the broader remaining result /
  existential union surface.
- Interface/runtime dispatch still reasons about residual interface/open union
  members through runtime type expressions because those payloads remain
  explicit `runtime.Value` branches inside the generated native carrier.
- Fully bound object-safe interface existentials now have a native carrier ABI
  too: generated Go interface carriers plus concrete/runtime adapters now cover
  static params/returns, typed local assignment, struct fields, direct method
  dispatch, wrapper/lambda conversion, and dynamic callback boundaries without
  collapsing those static paths back to `runtime.Value`.

This is a staged implementation fallback, not the target ABI.

### First Landed Slice

The first code-bearing pass is now in place for the compiler-native scalar
nullable family:

- compiled static nullable scalar locals/params/returns use native Go pointer
  carriers instead of `any`;
- compiled wrappers and compiled lambdas convert those shapes explicitly at ABI
  boundaries through generated helpers instead of `any`;
- static typed `match` on those shapes lowers through nil checks plus direct
  payload dereference instead of `__able_try_cast(...)`;
- static `or {}` on those shapes lowers through nil checks plus direct payload
  unwrap on the success path;
- runtime helper emission for the generated nullable boundary helpers is split
  into `generator_render_runtime_nullable.go` so the compiler runtime-helper
  renderer stays under the repo file-size limit while the carrier family grows.

### First Closed-Union Slice

The next code-bearing pass is now also in place for a narrow closed-union
surface:

- direct two-member `UnionTypeExpression` shapes whose branches already map to
  native Go carriers now lower to generated Go interfaces plus wrapper carrier
  structs instead of `any`;
- named union definitions that normalize to the same two-branch native shape
  now map to those same generated carriers;
- compiled params/returns and explicit wrapper/lambda boundary adapters use
  explicit generated union helpers on those shapes;
- static typed `match` on those carriers now unwraps branches through generated
  helper functions instead of `__able_try_cast(...)`;
- static typed `match` can now also recognize the closed-union case where
  exactly one concrete member implements the requested interface name,
  starting with `case err: Error => ...`; those branches still keep the union
  discrimination native and only convert the matched whole value to
  `runtime.Value` at the binding edge when the handler needs the dynamic
  `Error` surface;
- static `or {}` on native unions whose failure branch is a native `Error`
  implementer now treats that branch as the failure path via generated union
  unwrap helpers instead of forcing the whole expression back through
  `runtime.Value` error probing.

This is still a deliberately narrow slice, not the final union ABI.

This is progress in the intended direction, not the final union ABI.

### First Native Result Slice

The next code-bearing pass is now also in place for the narrow native result
surface that fits the first closed-union ABI:

- `!T` now lowers to the same generated native carrier family when its error
  branch normalizes to `runtime.ErrorValue` and the success branch `T` already
  maps to a native Go carrier;
- compiled returns, branch coercion, wrapper conversion, and propagation on
  those shapes now stay on the native carrier instead of lowering the whole
  result path to `any`;
- for no-bootstrap compiled execution, concrete `Error` implementers now build
  `runtime.ErrorValue` payloads by calling the compiled `Error.message()`
  implementation first, instead of depending on interpreter-backed
  `bridge.ErrorValue(...)` recovery to discover the message text at runtime.
- direct compiled `Error.message()` / `Error.cause()` calls now stay on that
  same native carrier instead of routing back through dynamic member helpers.

This is still a deliberately narrow slice, not the final result ABI.

### Native `Error` Carrier Widening

The next incremental widening is now also in place for plain `Error`:

- direct `Error` params/returns on compiled static paths now use
  `runtime.ErrorValue` instead of `runtime.Value`;
- nullable `?Error` now uses `*runtime.ErrorValue` instead of `any` /
  `runtime.Value`;
- explicit unions such as `String | Error` now reuse the same native carrier
  family as the current `runtime.ErrorValue | T` result slice;
- direct compiled `Error` method calls and error-bearing struct fields now also
  stay on that native carrier family;
- this special-case is now internally consistent, but it is still not a
  general native lowering for arbitrary interface existential types.

### Broader Carrier Widening

The next widening tranche is now also landed for the broader native-union
surface needed before moving on to control-result work:

- closed nominal unions are no longer limited to the first two-member slice;
  broader multi-member nominal unions now stay on generated native union
  carriers instead of `any`;
- generic alias unions such as `Option T = nil | T` and
  `Result T = Error | T` now normalize onto the already-landed native nullable
  and native result carrier families when their bound members fit those
  carrier rules;
- interface/open unions such as `String | Tag` now also stay on generated
  native union carriers; non-native existential branches remain explicit as
  residual `runtime.Value` members inside that carrier instead of forcing the
  entire union back to `any`;
- singleton struct boundary converters now accept both
  `StructInstanceValue` and `StructDefinitionValue`, which keeps interpreted
  callers passing bare singleton values compatible with compiled native
  struct/union params.

This completes the current carrier-widening category. The later control-result,
object-safe interface, generic-interface existential, and callable/function-type
existential tranches are now also landed. Remaining work is now different in
kind: broader audit/enforcement of explicit dynamic-carrier edges plus further
specialization/performance work on top of the native carrier ABI.

## Target ABI

### Closed Unions

For a closed union such as:

```able
union MaybeInt = nil | i32
union Shape = Circle | Rect
```

the compiler should generate host-native carriers:

```go
type __able_union_MaybeInt interface {
	__able_union_MaybeInt()
}

type __able_union_MaybeInt_i32 int32

type __able_union_Shape interface {
	__able_union_Shape()
}

type __able_union_Shape_Circle struct{ Value *Circle }
type __able_union_Shape_Rect struct{ Value *Rect }
```

The exact generated names are an implementation detail. The architectural rules
are not:

- unions compile to generated Go interfaces;
- concrete variants compile to native carrier types;
- nominal struct variants stay native pointers/structs inside those carriers;
- `any` is not the steady-state union representation.

### Nullable

Nullable is a special union case:

- `?Point` -> `*Point`
- `?Array i32` -> `*Array` today, later monomorphized array carrier pointer
- `?i32` -> dedicated nullable wrapper or generated union carrier, not `any`

Pointer/slice-backed nullable forms should stay nil-capable native Go values.
Value-type nullable forms need generated carrier wrappers.

### Result

`!T` / `Result T E` should follow the same rule:

- native success carrier for `T`
- native error carrier for the error branch
- explicit generated interface/wrapper family rather than `any`

This aligns with the broader no-panic control-result direction.

## Pattern Matching

Union pattern matching should lower to native Go type discrimination:

- type switches for generated union interfaces
- nil checks for nullable pointer forms
- direct variant extraction for wrapped payloads

Dynamic `bridge.MatchType(...)` should remain only at explicit ABI/dynamic
boundaries, not the primary compiled pattern engine.

## Boundary Discipline

When compiled union values cross a dynamic boundary:

- native union carrier -> boundary adapter -> `runtime.Value`
- `runtime.Value` -> boundary adapter -> native union carrier

The adapters should be generated per union shape or per canonical type
signature. They should not force static compiled locals/params/returns back to
`any`.

## Bring-Up Order

1. Add union-shape analysis/canonicalization so equivalent unions share one
   generated ABI.
2. Keep nullable pointer/slice forms native and introduce a carrier strategy
   for nullable value types.
3. Generate Go interfaces + wrappers for closed union shapes referenced by
   compiled params/returns/locals.
4. Lower typed pattern checks on native union carriers before replacing
   `bridge.MatchType(...)` at wrappers.
5. Narrow `any` use to explicit dynamic boundaries and residual unsupported
   generic cases only.

## First Implementation Targets

The first code-bearing union ABI pass should focus on the smallest useful set:

1. Nullable value types (now expanded to the compiler-native scalar family)
   without `any`.
2. Closed two-branch unions over native nominal/native carrier types (`A | B`
   where `A/B` already map to native Go carriers).
3. Native pattern matching over those carriers in compiled `match`.
4. Wrapper conversion helpers for compiled params/returns on those same shapes.

Do not start with:

- fully generic open unions;
- interface existential unions;
- every dynamic boundary shape at once.

## Non-Goals For The First Pass

- solving all generic higher-kinded interface dispatch;
- eliminating every runtime type-expression helper;
- redesigning the runtime/interface dispatch subsystem before native union
  carriers exist.
