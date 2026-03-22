# Compiler Go Lowering Specification

## Status

This document is the canonical lowering specification for the v12 Go compiler.
It defines how Able language constructs map to Go code, which semantic helpers
are allowed, and where the explicit dynamic boundary begins.

This document is intentionally stricter than the current implementation. It
captures the target architecture the compiler must satisfy.

Related documents:
- `v12/design/compiler-go-lowering-plan.md`: ordered work plan from the current
  compiler to the target architecture described here.
- `v12/design/compiler-native-lowering.md`: short-form contract and guardrails.
- `v12/design/compiler-aot.md`: correctness-first AOT scope.
- `spec/full_spec_v12.md`: language semantics.

## Purpose

The compiler must compile Able semantics into equivalent low-level Go
implementations.

The compiler must not implement static Able semantics by delegating normal work
back to the interpreters. The interpreters are:
- the semantic oracle used for validation
- the engine behind explicit dynamic features
- not the primary implementation strategy for compiled static code

The compiler must therefore do three things well:
1. synthesize a native Go carrier for every statically representable Able value
2. synthesize direct Go control flow and dispatch for every statically
   representable Able operation
3. isolate the interpreter/runtime object model behind narrow, explicit dynamic
   boundaries

## Non-Negotiable Rules

1. Only primitive Able types may have primitive-specific lowering rules.
2. All non-primitive nominal types must lower through shared nominal lowering
   rules, not per-type compiler branches.
3. Language-level syntax and kernel ABI boundaries may have dedicated lowering
   rules, but those rules attach to the syntax or ABI, not to a named nominal
   type.
4. Static compiled code must not use interpreter evaluation or interpreter
   dispatch as a convenience mechanism.
5. Every lowering decision must come from a reusable synthesis point, not from
   ad hoc emitter-local special casing.

## What Is Allowed To Be Special

The compiler is allowed to have dedicated lowering rules only for these
categories:

1. Primitive scalar families.
2. Language syntax forms whose semantics are defined by the language itself:
   array literals, map literals, string interpolation, ranges, `spawn`,
   `await`, `dynimport`, `extern`, placeholder lambdas, and similar forms.
3. Dynamic-boundary ABI conversions.
4. Host-runtime services that the language explicitly requires: concurrency
   scheduler, process exit, explicit extern ABI, dynamic package registry.

Everything else must be handled by shared lowering machinery.

## Architectural Model

The compiler architecture must be organized around reusable lowering stages.
These stages are the only places where lowering knowledge is allowed to live.

### 1. Semantic Input

Input to codegen is not raw syntax.

It is a typed, resolved semantic graph containing at least:
- normalized type expressions
- resolved bindings
- resolved overloads and impls
- resolved interface/default-method targets where statically knowable
- explicit dynamic-feature markers
- explicit evaluation order

### 2. Canonical Type Normalization

Before Go codegen, the compiler must normalize type expressions so all later
stages see one canonical description of the Able type.

Normalization includes:
- expanding type aliases
- substituting bound generic parameters
- normalizing `Option T` / `Result T` aliases to their underlying union form
  when appropriate
- canonicalizing nullable and result forms
- canonicalizing named union aliases to the same structural representation used
  by the carrier synthesizer

All later codegen must consume normalized type expressions instead of
re-deriving type shape from syntax.

### 3. Carrier Synthesis

Carrier synthesis answers one question:

"What Go type represents this Able value on static compiled paths?"

Carrier synthesis must be centralized. Every codegen site that needs a Go type
must ask the carrier synthesizer rather than inventing a local fallback.

Carrier synthesis produces:
- the Go carrier type
- zero/nil form
- whether the carrier is nil-capable
- whether it is pointer-backed or value-backed
- how it crosses dynamic boundaries
- how it joins with sibling carriers

### 4. Dispatch Synthesis

Dispatch synthesis answers:
- how to read a field
- how to write a field
- how to call a function
- how to call a method
- how to call through an interface carrier
- how to call a callable value
- how to perform index operations

Dispatch synthesis must be shared and target-driven. The emitter should not ask
"is this a HashMap?" or "is this a LinkedList?". It should ask "what is the
receiver carrier and what operation is being performed?"

### 5. Control/Result Synthesis

Control synthesis answers how non-local Able semantics are represented in Go.

This includes:
- `return`
- `break`
- `continue`
- `breakpoint` exits
- `raise`
- `rethrow`
- `!`
- `or {}`
- `rescue`
- `ensure`

Where pure structured Go control flow is sufficient, the compiler must emit it
straightforwardly. Where helper boundaries need to propagate non-local control,
the compiler must use explicit control envelopes, not `panic` / `recover`.

### 6. Pattern/Join Synthesis

Pattern lowering and branch-join lowering must also be centralized.

This stage is responsible for:
- typed-pattern checks
- struct/array destructuring
- union/interface branch discrimination
- `if`/`match`/`rescue`/`or {}`/loop-result join carrier inference
- nil-aware joins
- recovered-type joins

### 7. Boundary Synthesis

Boundary synthesis is the only stage allowed to convert between native compiled
carriers and interpreter/runtime carriers.

A boundary must be explicit in the source semantics or explicit in the host ABI.
If there is no such boundary, native compiled code must stay native.

## Canonical Reusable Lowering Units

The compiler must expose one reusable synthesis path for each of the following.
No emitter-local alternatives are allowed.

1. `NormalizeTypeExpr`
   - input: spec-level Able type expression
   - output: canonical semantic type expression
2. `SynthesizeCarrier`
   - input: normalized type expression
   - output: Go carrier descriptor
3. `SynthesizeZeroValue`
   - input: carrier descriptor
   - output: zero/nil value form
4. `SynthesizeJoinCarrier`
   - input: N carrier descriptors plus normalized branch types
   - output: common carrier descriptor or explicit dynamic-boundary requirement
5. `SynthesizePattern`
   - input: pattern + subject carrier + subject type
   - output: ordered tests + extracted bindings + binding carriers
6. `SynthesizeDispatch`
   - input: operation + receiver carrier + resolved target metadata
   - output: direct Go expression/helper call
7. `SynthesizeControl`
   - input: control-producing construct
   - output: direct Go control or explicit control envelope
8. `SynthesizeBoundaryAdapter`
   - input: source carrier + target boundary carrier
   - output: conversion helpers

If a new lowering change cannot be expressed as a change to one of these shared
synthesis points, it is almost certainly the wrong fix.

## Canonical Go Shapes

The exact generated names may differ, but the architectural shapes are fixed.

### Nominal Struct Carrier

```go
type __able_struct_Point struct {
    X int32
    Y int32
}
```

Use pointers when mutation identity or aliasing must be preserved:

```go
type __able_struct_Buffer struct {
    Data *__able_array_u8
}
```

### Interface Carrier + Adapters

```go
type __able_iface_Drawable interface {
    __able_Draw() string
}

type __able_iface_Drawable_Point struct {
    value *__able_struct_Point
}

func (a __able_iface_Drawable_Point) __able_Draw() string {
    return __able_impl_Point_Drawable_draw(a.value)
}
```

### Union Carrier

```go
type __able_union_Error_or_i32 interface {
    __able_union_Error_or_i32_marker()
}

type __able_union_Error_or_i32_error struct {
    value runtime.ErrorValue
}

type __able_union_Error_or_i32_i32 struct {
    value int32
}
```

### Array Carrier

```go
type __able_array_i32 struct {
    elements []int32
}
```

Or, for generic-but-statically-representable element carriers:

```go
type __able_array_Drawable struct {
    elements []__able_iface_Drawable
}
```

### Callable Carrier

```go
type __able_fn_i32_to_bool func(int32) bool
```

Or, when capture representation needs a named type:

```go
type __able_closure_Filter struct {
    threshold int32
}

func (c *__able_closure_Filter) Invoke(v int32) bool {
    return v > c.threshold
}
```

### Control Envelope

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

The exact payload typing may be specialized, but ordinary control propagation
must use return-based signaling rather than `panic` / `recover`.

## Type-Form Lowering

This section is exhaustive at the language-construct level.

### Primitive Scalars

The primitive scalar family in this document matches `spec/full_spec_v12.md`
Section 4.2 exactly.

| Able type form | Go carrier | Notes |
| --- | --- | --- |
| `bool` | `bool` | direct host boolean |
| `i8/i16/i32/i64` | matching signed Go width | direct host integer |
| `i128` | compiler-owned 128-bit signed carrier | preserve full spec range semantics |
| `u8/u16/u32/u64` | matching unsigned Go width | direct host integer |
| `u128` | compiler-owned 128-bit unsigned carrier | preserve full spec range semantics |
| `f32/f64` | `float32` / `float64` | direct host float |
| `char` | `rune` | Unicode scalar value |
| `void` | `struct{}` | zero-sized completion value |

`nil` is not a freestanding carrier. It lowers as the typed zero/nil form of the
expected carrier. If no expected carrier exists and the expression is entering a
dynamic boundary, boundary synthesis must materialize the runtime nil form.

### Built-In Textual Scalar

| Able type form | Go carrier | Notes |
| --- | --- | --- |
| `String` | `string` | language-defined scalar built-in |

`String` lowering is permitted as a built-in scalar rule because string literals,
interpolation, and text indexing are language-level syntax. This is not a
pattern for granting special lowering to arbitrary nominal types.

### Type Alias

`type Alias = T`

Lowering rule:
- aliases do not introduce a new runtime carrier
- normalization expands the alias
- carrier synthesis uses the normalized target type

### Nominal Struct Types

`struct Name ...`

Lowering rule:
- generate one Go carrier for the nominal struct shape
- fields lower using the carrier synthesizer recursively
- mutability/identity are preserved by pointer usage where needed
- fully bound generic instantiations lower through the same struct template with
  substituted field carriers

This rule applies uniformly to:
- user-defined structs
- stdlib structs
- singleton structs
- positional structs
- generic structs

### Nominal Union Types

`union Name = A | B | ...`

Lowering rule:
- synthesize one Go interface carrier for the union
- synthesize one wrapper carrier per variant that cannot already serve as the
  interface representation directly
- synthesize coercion helpers from each member carrier to the union carrier
- `match` compiles to type switches / nil checks / discriminant checks on this
  generated family

### Nullable Types

`?T`

Lowering rule:
- if `T` already has a nil-capable carrier, use that nil-capable carrier form
- otherwise synthesize a native nullable wrapper around `T`
- success-path uses the unwrapped `T` carrier
- failure-path is typed nil, not `runtime.Value`

### Result Types

`!T` or `Error | T`

Lowering rule:
- synthesize a native result carrier when both error and success carriers are
  statically representable
- success path carries the native `T` carrier
- failure path carries the native `Error` carrier
- `!` is a control operation over this data carrier, not a request to box the
  whole result into `runtime.Value`

### Interfaces

`interface I ...`

Lowering rule:
- synthesize one Go interface carrier per fully bound Able interface type
- synthesize concrete adapters from each implementing nominal carrier
- synthesize runtime adapters only for values already crossing from the dynamic
  boundary
- default methods lower to compiled helpers that take the interface carrier and
  any explicit args

### Callable Types

`(P1, P2, ...) -> R`

Lowering rule:
- synthesize a Go function type or callable wrapper type whose params/return use
  synthesized carriers for `P1..Pn` and `R`
- closures lower to functions plus an environment carrier when captures exist
- bound methods lower to callable wrappers over the receiver carrier plus target
  impl/method

### Fully Bound Generic Nominal Types

`Box i32`, `Reader String`, `Pair i32 bool`

Lowering rule:
- substitute the concrete type arguments into the nominal template
- synthesize the resulting carrier through the same shared nominal rules
- do not fall back to `runtime.Value` merely because the type arrived via a
  generic path

### Unresolved / Truly Dynamic Types

Lowering rule:
- if the type remains genuinely dynamic after normalization and resolution, the
  value may use a dynamic carrier at an explicit boundary
- unresolved/dynamic is a boundary classification, not a default convenience
  fallback for static codegen

## Definition Lowering

### Package / Module

Able package/module lowering produces:
- one Go package per Able package
- Go top-level symbols for exported Able definitions
- a compiled package-init sequence for top-level evaluation ordering when the
  spec requires executable top-level work
- package metadata for the dynamic bridge when dynamic code must reference
  compiled definitions

### `import`

Static `import` lowers to compiled package references and initialization order.
It must not invoke interpreter package loading on static paths.

### `dynimport`

`dynimport` is an explicit dynamic-boundary construct.

Lowering rule:
- emit a call into the dynamic package registry / interpreter bridge
- adapt resulting dynamic values into native carriers only when a legal explicit
  adaptation exists

### `type`

Type alias lowering is compile-time only. No runtime code is emitted beyond what
is needed for reflection/diagnostics metadata.

### `struct`

Emit:
- Go carrier type
- constructor helpers as needed for literal lowering
- conversion helpers for explicit dynamic boundaries

### `union`

Emit:
- union carrier interface
- variant wrappers
- conversion helpers
- match helpers only if needed for generated-code reuse; pattern lowering must
  still remain native in shape

### `interface`

Emit:
- Go interface carrier
- adapter carriers
- default-method helpers
- explicit dynamic-boundary adapters

### `methods` and `impl`

Lowering rule:
- resolved methods/impls lower to compiled Go functions or receiver methods over
  already-synthesized nominal carriers
- generic impl specialization is driven by the same nominal substitution logic
  used everywhere else
- no per-type emitter branches are allowed

### `fn`

Lowering rule:
- top-level functions lower to Go functions
- local functions lower to nested helpers or closures
- parameter and return carriers use synthesized carriers
- body lowering uses the control, pattern, and dispatch synthesis paths

### Placeholder Lambdas and Partials

Placeholder syntax lowers to generated closures over the supplied expression and
placeholder bindings.

This is a syntax-level lowering rule, not a named-type rule.

### `extern`

`extern` is a boundary construct.

Lowering rule:
- emit ABI wrapper functions that convert native carriers to the host ABI form
- convert results back through the carrier synthesizer and boundary adapters
- do not route static semantics through interpreter calls unless the extern's own
  implementation boundary requires it

## Expression Lowering

### Identifier

Lowering rule:
- local binding -> Go local or captured field
- package binding -> compiled package symbol
- member implicit binding -> resolved through static dispatch synthesis
- recovered native carriers must be preferred over `runtime.Value` locals when
  concrete type information is still available

### Literal Expressions

#### Primitive literals

Lower directly to the corresponding Go literal or constant expression.

#### String interpolation

Lower to Go string-building code or shared compiled string helpers.
This is a language-syntax rule.

#### Array literal

Lower to compiler-owned array carrier construction.

The rule attaches to array-literal syntax, not to any named container type.

Representative shape:
- allocate compiler-owned array wrapper
- evaluate elements left-to-right
- convert each element to the synthesized element carrier
- append/store directly into native storage

#### Struct literal

Lower to direct construction of the synthesized nominal struct carrier.
Functional update lowers to ordered field-copy/override code over that same
carrier.

#### Map literal

Map literals are a language syntax form whose default semantic target is the
language-defined map type.

Lowering rule:
- evaluate entries left-to-right
- build the target map value through the language-defined map-construction ABI
- once constructed, the resulting nominal map value follows the same shared
  nominal lowering rules as any other non-primitive nominal value

The dedicated logic is attached to literal syntax, not to the `HashMap` name.

#### Iterator / generator literal

Lower to a generated iterator/controller state machine implemented in Go.
The state machine may use generated structs, labels, and explicit state enums.
It must not require interpreter evaluation on static paths.

### Member Access

`receiver.member`

Lowering rule:
- field read -> direct field access on the synthesized carrier
- method value -> direct compiled method/bound-callable synthesis
- interface member -> compiled interface/default-method dispatch synthesis
- dynamic member lookup is allowed only if the source semantics are already in a
  dynamic-boundary context

### Index Access

`object[index]`

Lowering rule:
- arrays -> direct wrapper/slice access plus explicit bounds semantics
- user nominal index protocols -> resolved compiled impl/method call
- dynamic helper routing is only allowed at explicit dynamic boundaries

### Function Calls

Lowering rule:
- fully resolved direct call -> direct Go call
- resolved static method call -> direct compiled impl/helper call
- resolved interface/default-method call -> generated native interface dispatch
- callable value call -> direct invocation of the synthesized callable carrier
- unresolved dynamic call -> explicit boundary call only if the source operation
  is dynamic by semantics

### Operators

Lowering rule:
- primitive arithmetic/comparison/logical operators -> primitive-lowering table
- non-primitive operators -> shared resolved impl/method call lowering
- overload selection must occur before emission; codegen emits the selected
  implementation, not a new runtime overload search

### Casts

Lowering rule:
- primitive casts -> primitive cast-lowering table
- nominal/interface/union casts -> shared coercion or boundary-adapter rules
- typed-pattern narrowing is handled by pattern synthesis, not by ad hoc cast
  helpers

### Range Expressions

Language syntax lowering.

Lower to a compiled range value or compiled range-construction helper using
native scalar carriers.

## Pattern Lowering

Patterns are lowered by shared pattern synthesis, not by emitter-local logic.

### Identifier Pattern

Lower to direct local binding with the synthesized carrier.
Typed identifier patterns perform a native compatibility/coercion check before
binding.

### Wildcard Pattern

Evaluate the source expression, discard the result, and emit no binding.

### Struct Pattern

Lower to:
1. nominal compatibility check against the expected struct carrier
2. ordered field extraction on the native carrier
3. direct local binding for each field binding

Renamed bindings such as `field::binding` are purely compile-time binding names.
They must lower to direct Go locals, never env lookups.

### Array Pattern

Lower to:
1. length/shape checks
2. element extraction on the native array carrier
3. rest-tail slicing through the same native array carrier family
4. direct local binding

### Typed Pattern

Lower through shared compatibility logic based on normalized Able types and the
synthesized carrier family. If the subject is statically representable, the
compiler must not fall back to runtime type-match helpers.

### Union / Interface Pattern

Lower to native discriminant checks:
- nil checks for nullable carriers
- type switches for generated union/interface carriers
- direct unwrap/bind steps for the matched branch

## Assignment and Binding Lowering

### `:=`

Lower to:
- Go local allocation
- RHS evaluation first
- pattern match/coercion second
- commit of new binding only on successful pattern match

### `=`

Lower to:
- RHS evaluation first
- resolved target storage evaluation once
- direct overwrite of the resolved storage location
- no repeated receiver/index evaluation for compound assignments

### Compound assignment

Lower to:
- evaluate receiver/index once
- read old value
- apply operator via primitive or resolved nominal operator lowering
- write back once

## Control-Flow Lowering

### Block / `do`

Lower to ordinary Go block structure, preserving evaluation order.

### `if` / `elsif` / `else`

Lower to:
- direct Go branching
- one synthesized join carrier for the overall result
- branch result coercion into that join carrier
- no `runtime.Value` join local when all branch types are statically
  representable

### `match`

Lower to:
- subject evaluation once
- ordered clause checks through shared pattern synthesis
- one synthesized join carrier for the expression result
- branch result coercion into that join carrier

### `while`

Lower to ordinary Go loop structure.
Where the shape is a counted loop, the compiler may lower it to a more direct Go
counted loop, but that is a control-form optimization, not a nominal-type rule.

### `for in`

Lower to one of:
- direct array iteration for statically known native array carriers
- direct iterator-carrier loop for statically known iterator carriers
- explicit dynamic-boundary iteration only when the iterable itself is dynamic

### `loop`

Lower to an ordinary infinite Go loop plus a synthesized loop-result carrier
when the loop is used as an expression.

### `breakpoint` / labeled `break`

Lower to structured block-result control. When a helper boundary must propagate
it, use the explicit control envelope.

### `break` / `continue`

Lower to direct Go `break` / `continue` when lexically local.
Use the explicit control envelope only across generated helper boundaries.

### `return`

Lower to direct Go `return` in ordinary function bodies.
Use the explicit control envelope only when a generated helper must propagate the
return across a lowering boundary.

## Error and Propagation Lowering

### `raise`

`raise` is control, not data.

Lower to:
- direct return of a raise control envelope when inside helper-mediated control
- or direct branch/return to the enclosing rescue-aware lowering region

### `rescue`

Lower to:
- evaluate the monitored expression
- intercept raise/error outcomes using native error/control carriers
- pattern-match rescue clauses using shared pattern synthesis
- join clause results using shared join synthesis

### `ensure`

Lower to:
- evaluate body
- record normal or exceptional completion
- run ensure block exactly once
- re-emit prior control outcome after ensure completes

### `rethrow`

Lower to re-emission of the currently handled raise/error control outcome.

### `!` propagation

Lowering depends on context, but the rule is general:
- unwrap the native success carrier
- if failure is present, emit the appropriate early-exit control from the
  current lowering region
- never box the whole result merely to implement propagation

### `or {}`

Lower to:
- evaluate primary expression once
- branch on native success/failure form
- bind handler variables with native carriers when statically representable
- join success and handler results through shared join synthesis

## Interface, Method, and Callable Lowering

### Static field/method dispatch

Lower to direct Go field reads/writes or direct compiled function/method calls.

### Default methods

Lower to generated compiled helpers that accept the interface carrier plus
explicit parameters. Default methods must not require interpreter dispatch on
static paths.

### Interface coercion

Lower to construction of the generated adapter carrier.
No runtime boxing is required unless the value is crossing an explicit dynamic
boundary.

### Generic interface/default-method dispatch

Lower to compiled dispatch helpers specialized by the fully bound interface type
and resolved method metadata. Runtime-based dispatch is allowed only when the
receiver already originated from the dynamic boundary.

### Bound method values

Lower to synthesized callable carriers that close over the receiver carrier and
selected target.

### Apply/callable protocols

Lower through the same callable carrier synthesis used for function types and
bound methods.

## Concurrency Lowering

Concurrency forms are language/runtime features, not nominal per-type fast
paths.

### `spawn`

Lower to:
- a compiled closure over synthesized carriers
- submission of that closure to the compiled scheduler/runtime service
- production of a `Future<T>` value using the language-defined future ABI

### `await`

Lower to the compiled scheduler/runtime await/select helper over already lowered
future/channel operations.

### `Future T`

Lowering rule:
- the future handle is a language/runtime service value
- explicit handle operations (`status`, `value`, `cancel`) lower through the
  future ABI
- implicit value-view lowering inserts the correct wait/value extraction path
  when the surrounding context requires `T`

### `Channel`, `Mutex`, scheduler helpers

These are language/runtime service values. Their semantics live in the compiled
runtime core, but user code still reaches them through the same shared nominal
and callable lowering rules once the carriers are synthesized.

## Dynamic Boundary Lowering

The dynamic boundary is explicit and finite.

Only the following categories may cross it:
- `dynimport`
- dynamic package mutation / definition
- dynamic evaluation / metaprogramming
- extern/host ABI crossings
- values that already originate from runtime dynamic payloads

### Boundary Rules

1. Convert native compiled carrier -> boundary carrier exactly at the edge.
2. Perform dynamic work.
3. Convert boundary carrier -> native compiled carrier immediately after the
   edge if the result type is statically representable.
4. Do not allow dynamic carriers to leak into surrounding static code beyond the
   explicit boundary result.

## What Must Never Appear On Static Paths

Unless the source program is itself invoking an explicit dynamic feature,
compiled static code must not rely on:
- interpreter `EvaluateProgram()` or expression evaluation fallback
- runtime overload/method search for already resolved static calls
- `runtime.Value` or `any` locals for statically representable values
- `panic` / `recover` for ordinary Able control flow
- IIFE scaffolding just to manufacture expression results
- env lookups for already resolved native bindings
- named-structure-specific compiler branches for non-primitive nominal types

## Acceptance Criteria

The lowering architecture described here is implemented when all of the
following are true:

1. Every statically representable Able type form synthesizes a native carrier.
2. Every static control-flow construct stays on native carriers.
3. Every static dispatch site lowers to direct compiled dispatch.
4. All remaining interpreter/runtime object-model crossings are explicit dynamic
   boundaries.
5. New nominal types compile correctly without adding new named-type lowering
   rules.
6. The compiled benchmark family is free of already-identified avoidable dynamic
   scaffolding on hot static paths.
