# Compiled runtime interface boundary

Date: 2026-07-17

## Decision

Retain the first generated-runtime boundary tranche. The compiler bridge no
longer imports or stores the concrete interpreter type, generated registration
uses the bridge capability contract, compiled thunks live in the shared runtime
value layer, and static launchers resolve executor policy without importing an
interpreter helper.

This is an architectural prerequisite, not a claimed application-speed win.
Static generated binaries still import the interpreter for two generic operator
helpers, so the expensive interpreter package initializer remains. The next
tranche must remove those last two roots before startup/RSS and the Binary Trees
GC guard can make a meaningful accept/reject decision.

## Retained changes

- `bridge.Interpreter` names the exact dynamic operations available to generated
  code and the compiler bridge. `*interpreter.Interpreter` satisfies it in a
  compile-time test.
- Production bridge files have no concrete interpreter import. A source audit
  prevents that dependency from returning.
- Private bytecode raw integer/float carriers expose boundary materialization,
  allowing the bridge to stabilize VM values without importing VM code.
- Interpreter raise signals expose only their carried value through a small
  error capability; their concrete control-signal representation remains
  private.
- `runtime.CompiledThunk` is the shared thunk identity. The old interpreter name
  remains an alias for API compatibility.
- Generated registration, package registration, and print helpers accept the
  bridge contract rather than `*interpreter.Interpreter`.
- Static generated launchers use `bridge.ExecutorKindFromEnvironment`; dynamic
  launchers still construct and use the full interpreter.
- A redundant generated exit-signal inspection was removed: both sides of the
  old branch returned the same error unchanged.

## Dependency result

`go list -deps ./pkg/compiler/bridge` contains no
`able/interpreter-go/pkg/interpreter` package. A freshly generated no-bootstrap
Noop application now has exactly two top-level concrete interpreter references:

1. `interpreter.ApplyBinaryOperatorFast(...)`
2. `interpreter.ApplyUnaryOperatorFast(...)`

The prior concrete registration signatures, thunk assertions, static executor
lookup, and static exit inspection are absent. Package-interface AST decoding is
retained only for output modes that retain interpreter bootstrap metadata.

## Retained-state measurement

Five fresh compiled Noop processes all completed successfully:

| Metric | Retained result |
| --- | ---: |
| Mean wall time | 0.0660 s |
| Mean user time | 0.0560 s |
| Mean system time | 0.0320 s |
| Mean GC cycles | 3.00 |

The generated binary is 13,985,560 bytes and still contains 329 symbols whose
names are rooted in the interpreter package. `GODEBUG=inittrace=1` continues to
attribute 55 ms, 38,001,992 bytes, and 707,332 allocations to interpreter
initialization. This is expected: an interface boundary changes dependency
direction, but the two operator imports still make the package and all of its
initializers reachable.

No broad performance comparison is claimed for this tranche. The static Noop
result remains in the existing roughly 60-70 ms process band, which is the
correct negative control until the final import roots are removed.

## Verification

- Full compiler bridge tests pass, including concrete-interface conformance and
  the no-concrete-import audit.
- Focused bytecode boundary materialization and completed-run tests pass.
- Static no-bootstrap boundary-clean fixtures pass.
- Dynamic callback, named-call, and original-call fallback tests pass.
- Static and dynamic generated-main selection tests pass.
- Focused generated thunk, control-error, future registration, and package
  metadata tests pass.
- The full `pkg/interpreter` aggregate command was not used as a pass/fail gate:
  its fixture-parity aggregate exceeded the repository's one-minute test rule.
  The relevant interpreter tests were run as a bounded partition and passed.

## Next recommendation

Move stable primitive operator evaluation behind an interpreter-independent
runtime operation package, then make generated interpreter imports conditional
on actual bootstrap/fallback requirements.

Why: binary and unary generic operator evaluation are the only remaining
concrete interpreter roots in a static generated application. Removing any
other small helper cannot change package reachability, startup allocations, or
RSS.

What it entails: extract the stable-value portion of binary/unary evaluation
with shared conformance tests against the current interpreter behavior; retain
VM-raw handling and dynamic dispatch in the full interpreter adapter; prove a
no-bootstrap generated source and `go list` dependency graph omit the
interpreter; then run repeated alternating short applications, a true dynamic
fallback application, allocation-light TapeLang, and allocation-heavy Binary
Trees under the existing OOM/GC guardrails. Reject the split if semantic parity
fails or the averaged Binary Trees/guard cohort regresses. Do not compensate by
changing `GOGC`, retaining heap ballast, or adding benchmark/type/container
special cases.
