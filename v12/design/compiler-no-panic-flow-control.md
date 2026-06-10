# Compiler: Explicit Control Flow, No IIFEs

This note refines [compiler-native-lowering-guardrails.md](./compiler-native-lowering-guardrails.md)
for control flow and exception propagation.

## Status and scope

This is an active guardrail for normal static Able control flow, not a claim
that generated code contains no `panic`, `recover`, or closure anywhere. The
direct compiler emits `__ableControl` for raised/error control that crosses a
generated helper or dynamic boundary. It currently has a `Raise` kind plus
runtime value and error payloads; ordinary `return`, `break`, and `continue`
remain structural Go control when they stay within their lexical lowering
region.

Host faults, compiler invariants, unsupported internal bridge states, and
runtime callback containment can still legitimately use Go panic/recovery
mechanisms. They must not become the normal representation of a specified Able
return, loop exit, `raise`, `rescue`, `ensure`, propagation, or `or {}` path.
`PLAN.md` selects any performance work; this note does not select a control-flow
rewrite candidate.

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

Panics remain available for true compiler/runtime faults or host-level fatal
conditions where there is no meaningful language-level recovery path.

## Required Model

Normal Able control flow should lower to ordinary Go statements and returns.

When a lowered helper or dynamic boundary needs to communicate raised/error
control to its caller, it returns an explicit control value rather than
panicking. The current generated shape is:

Representative shape:

```go
type __ableControlKind uint8

const (
    __ableControlRaise __ableControlKind = iota + 1
)

type __ableControl struct {
    Kind  __ableControlKind
    Value runtime.Value
    Err   error
}
```

The exact generated shape may evolve, but a change must preserve explicit,
return-based propagation and the existing diagnostics/caller-frame behavior.
It must not introduce a general panic/recover stack for normal language control.

## IIFEs

IIFEs are not a normal-static-control mechanism.

Instead:

- expression lowering should emit setup lines plus a final value expression;
- statement contexts should execute those lines directly;
- helper boundaries should be named helpers or explicit local temporaries, not
  anonymous closures used merely as control/result wrappers.

This rule does not prohibit closures required by Able lambdas, callbacks, or
other semantically necessary generated initialization. Static source audits
should reject control-only IIFEs in the body under test, not perform a global
ban on every generated closure.

## Lowering Rules

### Returns

- Lower ordinary Able `return` to ordinary Go `return`.
- Do not add a return-control enum merely to model lexically local returns.
- If a future helper boundary needs non-local return propagation, specify and
  prove that representation before adding it; it is not part of the current
  `__ableControl` contract.

### Break / Continue

- Lower loop control with normal Go loop labels where a plain statement context
  is enough.
- Do not assume the current raised/error control carrier also represents
  break/continue. A future cross-helper form requires explicit semantics and
  focused proof.

### Breakpoints

- A breakpoint body should evaluate with normal statements and explicit result
  temporaries.
- A `break 'label value` should use ordinary branch/result logic in the current
  structural lowering. Any future helper-boundary representation needs a
  dedicated, source-backed contract.

### Raise / Rescue / Or Else

- `raise` should produce an explicit control signal representing an exception.
- `rescue` / `or_else` should branch on returned control information, consume
  exception signals when appropriate, and propagate anything else explicitly.
- The compiler must not rely on `panic` / `recover` to model the common
  exception path.

## Current generated-control record

- Generated compiled functions and relevant helper calls return
  `*__ableControl` alongside their value where raised/error control must
  propagate.
- `raise`, `rescue`, `or_else`, `ensure`, compiled helper calls, and generated
  dynamic `call_value` / `call_named` sites now branch on returned control
  where that control is applicable, instead of relying on panic/recovery for
  the normal raised-error path.
- Explicit dynamic callback boundaries now normalize callback failures back
  into ordinary Go `error` returns so boundary markers and diagnostics survive
  runtime callback failures.
- Residual dynamic member/helper paths now do the same: generated
  `__able_member_get`, `__able_member_set`, `__able_member_get_method`, and
  `__able_method_call*` helpers now return ordinary `error` /
  `*__ableControl` results instead of panicking internally, and the temporary
  recover-based bridge wrappers have been removed.

## Audit boundary

Focused generated-source audits cover representative static pattern/control
bodies and reject dynamic bridge helpers, `panic`, `recover`, and control-only
`func() ...` wrappers there. Separate tests verify that dynamic helpers no
longer emit the retired panic bridge wrappers and that the control bridge keeps
exit and native-error information intact.

## Change criteria

1. Classify a reported panic/recover or closure by its semantic boundary before
   changing it; a raw text search is not evidence of a language-control bug.
2. Keep normal static control on direct Go branches/returns and explicit
   raised/error returns where the source-backed helper boundary requires them.
3. Add focused behavioral and generated-source proof before changing the
   control representation. Do not select a performance change from this note
   without the shared-leaf evidence required by `PLAN.md`.
