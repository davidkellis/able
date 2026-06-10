# Compiler non-capture/effect audit

Status: feasibility closed; no optimization retained (2026-07-17)

## Purpose

This note records the proof required before the compiler may reuse storage for
an identity-bearing nominal value across loop iterations. It supplements the
caller-owned nominal result ABI in `compiler-native-lowering.md`; it does not
authorize a lowering change.

The current ABI gives each ordinary eligible call distinct addressable storage.
Go may stack-allocate that storage when it does not escape. Reusing a prior
iteration's object is different: it changes Able-visible identity unless both
the caller and every callee prove the old object is unreachable.

## Required proof

A safe loop-carried reuse decision needs two independent analyses.

1. A per-parameter, interprocedural non-capture summary must prove that the
   input is not returned, stored in an aggregate or module binding, captured by
   a closure/iterator/task, passed through dynamic dispatch, or passed to a
   callee that lacks the same proof.
2. Caller-side alias and liveness dataflow must prove that no local, aggregate,
   closure, branch result, or previously adopted candidate can still refer to
   the object when its storage is overwritten.

A non-capturing callee alone is insufficient because the caller may retain an
alias. A dead caller binding alone is insufficient because the callee may have
stored the input. Fixed slot rotation is not a substitute because an alias can
remain live for an unbounded number of iterations.

## Existing compiler facts

- `resolveCallerOwnedResults` proves that a small nominal result is fresh or is
  produced by a proven-fresh tail-call chain. It does not summarize parameter
  effects.
- `resolveCompiledEnvironmentIndependence` computes a conservative function
  fixed point for package-environment access. Its fact is function-wide, not
  parameter-specific, and says nothing about aliases or object capture.
- IR closure/spawn capture lists identify lexical slots needed by generated
  closures. They are not a whole-program parameter escape analysis and do not
  cover aggregate storage, dynamic calls, or generator-level local liveness.
- Go escape analysis runs after Able lowering. It can choose stack versus heap
  for semantically distinct objects, but it cannot authorize merging their
  identities.

## Corpus census

A source census over active v12 examples, fixtures, and canonical stdlib found
28 `x = call(x, ...)` sites and 64 `x = x.method(...)` sites. After applying the
current small-scalar nominal result shape:

- primitive recurrences are outside this pointer-identity problem;
- Array, String, persistent-container, BigInt, and BigUint results are outside
  the eligible two-scalar-field carrier shape;
- the generic enumerable `acc = f(acc, value)` has an unknown callable and is
  necessarily capture-unknown;
- Rational's eligible temporary results are already eliminated by distinct
  caller-owned slots and are not loop-carried;
- only the user `RecurrenceState` fixture and the related signed/unsigned
  128-bit accumulation fixtures repeat the direct unconditional shape;
- Fixed Width conditionally adopts or rejects candidates and therefore fails
  caller-side lifetime proof.

This supplies one independent user recurrence plus one related numeric family,
not three unlike programs. Building the analysis or changing lowering would
therefore fail the project's generality admission rule.

## Semantic regression boundary

Compiler tests now cover all three distinct hazards:

- a caller retains an old result while advancing the current value;
- a callee stores the input in an Array before returning a fresh result; and
- a loop conditionally adopts one candidate, retains it, and later constructs
  rejected candidates from the retained best value.

Both functions in the new capture test already qualify for `_into` variants.
The test therefore guards the exact boundary between safe distinct
caller-owned storage and unsafe cross-iteration identity reuse.

## Re-entry criteria

Reopen this work only when at least three unlike verifier-backed applications
and at least two unrelated nominal definitions naturally exhibit the same
unconditional recurrence shape. At that point, design both proof halves,
reject unknown/dynamic paths conservatively, and measure exact allocations plus
alternating broad guards. Do not introduce named nominal, container, stdlib,
package, or benchmark rules.
