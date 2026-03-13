# Compiler: Explicit Control Flow, No IIFEs

This note refines [compiler-native-lowering.md](./compiler-native-lowering.md)
for control flow and exception propagation.

## Principle

Compiled Go output should not use `panic` / `recover` or IIFEs for normal Able
control flow.

That applies to:

- `return`
- `break`
- `continue`
- `breakpoint` / labeled non-local exits
- `raise`
- `rescue`
- `or_else`

Panics are reserved for true compiler/runtime faults or host-level fatal
conditions where there is no meaningful language-level recovery path.

## Required Model

Normal Able control flow should lower to ordinary Go statements and regular
returns.

When a lowered helper or expression context needs to communicate a non-local
effect to its caller, it should return an explicit control signal rather than
panic.

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

The exact generated shape may be specialized, but callers must be able to
distinguish:

- normal completion
- ordinary function return
- loop/breakpoint jump
- raised exception

using regular Go conditionals, not `recover`.

## IIFEs

IIFEs are not part of the target architecture.

Instead:

- expression lowering should emit setup lines plus a final value expression;
- statement contexts should execute those lines directly;
- helper boundaries should be named helpers or explicit local temporaries, not
  anonymous closures used as expression wrappers.

## Lowering Rules

### Returns

- Lower ordinary Able `return` to ordinary Go `return`.
- If a nested lowered helper must report a return to its caller, return a
  control signal with `Kind == __ableControlReturn`.

### Break / Continue

- Lower loop control with normal Go loop labels where a plain statement context
  is enough.
- If control must cross a lowered helper boundary, return
  `__ableControlBreak` / `__ableControlContinue` with the relevant label.

### Breakpoints

- A breakpoint body should evaluate with normal statements and explicit result
  temporaries.
- A `break 'label value` should become a returned control signal or ordinary
  branch logic, depending on whether a helper boundary exists.

### Raise / Rescue / Or Else

- `raise` should produce an explicit control signal representing an exception.
- `rescue` / `or_else` should branch on returned control information, consume
  exception signals when appropriate, and propagate anything else explicitly.
- The compiler must not rely on `panic` / `recover` to model the common
  exception path.

## Current Status

- Static compiled function bodies now use explicit `*__ableControl`
  propagation for the common non-local-flow path.
- `raise`, `rescue`, `or_else`, `ensure`, compiled helper calls, and generated
  dynamic `call_value` / `call_named` sites now branch on returned control
  instead of relying on `panic` / `recover`.
- Explicit dynamic callback boundaries now normalize callback failures back
  into ordinary Go `error` returns so boundary markers and diagnostics survive
  runtime callback failures.
- Residual dynamic member/helper paths now do the same: generated
  `__able_member_get`, `__able_member_set`, `__able_member_get_method`, and
  `__able_method_call*` helpers now return ordinary `error` /
  `*__ableControl` results instead of panicking internally, and the temporary
  recover-based bridge wrappers have been removed.

## Remaining Violations To Remove

- Any remaining IIFE-based expression wrappers (`func() T { ... }()`).
- Any lowering that still requires `defer` / `recover` stacks for loop or
  block control.
- Any remaining helper/runtime paths that still use `panic` for ordinary
  language-level error/control propagation instead of explicit returns.

## Execution Plan

1. Audit remaining helper/runtime paths for ordinary-language panic usage now
   that the residual dynamic member/call helpers no longer need recover-based
   containment.
2. Add/extend regression audits that fail when generated static code contains
   new IIFEs or panic-based flow-control scaffolding.
