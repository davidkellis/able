# Bytecode fixed-arity equality call gate — 2026-07-17

## Decision

Keep the generic two-argument call-buffer path and use it for cached equality
dispatch. The previous tranche removed a redundant defensive copy, but its
fresh two-element argument slice still escaped to the heap once per comparison.
The new path checks out a reusable two-value buffer for interpreted functions
and native functions that explicitly promise to borrow arguments. Callables
that could retain the argument slice continue to receive an ordinary owned
slice.

The buffer remains checked out until the call returns. Nested equality calls
therefore acquire independent storage, and partial application copies bound
arguments before the pooled buffer is cleared. No Able value, primitive type,
nominal type, collection, benchmark, or implementation name selects the path.

No compiler, stdlib, language, or benchmark-source change is retained. A
temporary nominal `Key`/`Eq` workload was used only as diagnostic evidence and
removed after the gate.

## Admission profiles

Fresh post-ownership profiles used one warmed measured call, the canonical
external `able-stdlib`, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, CPU 0,
and the existing bytecode runtime harness. Allocation profiling used rate 1.

| Workload | Equality calls | Baseline flat allocation objects |
| --- | ---: | ---: |
| Boolean Reconciliation | 393,216 | 393,216 in `applyCachedEqualityDispatch` |
| Run-length encode | 959,952 | 959,952 in `applyCachedEqualityDispatch` |
| Unicode Scalar Pipeline | 1,769,472 | 1,769,472 in `applyCachedEqualityDispatch` |
| Temporary custom nominal `Eq` | 262,144 | 262,144 in `applyCachedEqualityDispatch` |

Escape analysis independently reports the two-element slice literal at the
cached dispatch call as escaping through `callCallableValueMutable(...)`.
The exact one-object-per-comparison relationship across primitive boolean,
character/text, Unicode scalar, and user-defined nominal equality satisfies
the cross-workload admission rule.

Post-change profiles remove `applyCachedEqualityDispatch` from the flat
allocation table in all four workloads. The next owners diverge:

- Boolean Reconciliation: raw integer result materialization.
- Run-length encode: string interpolation.
- Unicode Scalar Pipeline: raw integer results and character iteration.
- Custom nominal `Eq`: initial struct construction.

The already-closed raw-integer result family is not reopened merely because it
remains visible in two consumers.

## Safety contract

`callCallableValue2Mutable(...)` has a conservative reuse gate:

- interpreted functions and overload sets may use ephemeral argument storage;
- native functions may use it only when `BorrowArgs` is true;
- bound functions inherit the underlying callable's decision;
- dynamic, partial, unknown, and non-borrowing native targets fall back to a
  fresh owned slice.

Focused tests cover parameter coercion, nested/reentrant borrowed calls,
non-borrowing native retention, and partial application followed by scratch
reuse. Existing ordinary-call and partial-call mutation guards remain green.

## Repeated performance gate

Every timing sample is a separate process. Positive rows use alternating pair
order. Short/noisy Base64 and Array Map guards add five candidate-first pairs,
as required for this active workstation.

| Workload | Samples/side | Baseline mean | Candidate mean | Result | Allocation result |
| --- | ---: | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 5 | 948.744 ms | 913.487 ms | 3.72% faster | 743,273 -> 350,056 (-52.9%) |
| Custom nominal `Eq` | 5 | 551.360 ms | 544.165 ms | 1.31% faster | 270,468 -> 8,327 (-96.9%) |
| Run-length encode | 5 | 3.1242 s | 3.0146 s | 3.51% faster | 1,270,760 -> 310,803 (-75.5%) |
| Unicode Scalar Pipeline | 3 | 7.3576 s | 6.9859 s | 5.05% faster | 6,961,435 -> 5,191,968 (-25.4%) |
| Iterator Collect guard | 3 | 460.542 ms | 419.571 ms | 8.90% faster | identical |
| Base64 guard | 8 | 74.612 ms | 74.041 ms | 0.76% faster | effectively identical |
| Numeric Array Map guard | 8 | 79.957 ms | 77.000 ms | 3.70% faster | identical |
| Reverse Complement guard | 3 | 1.146 ms | 1.118 ms | 2.41% faster | effectively identical |
| String Split/Join guard | 3 | 1.0503 s | 1.0186 s | 3.01% faster | effectively identical |

The allocation reduction is exact and material in all equality-heavy rows.
The unrelated guards do not reveal synchronization or initialization cost from
the additional pool.

## Verification

- Fixed-arity coercion, reentrancy, retained-native, partial-copy, equality
  cache, ordinary-call mutation, partial-chain, and custom-`Eq` parity tests
  pass in 0.616 seconds.
- `go test ./pkg/runtime -count=1 -timeout 55s` passes in 0.062 seconds.
- All changed Go files remain below 1,000 lines and `git diff --check` passes.

## Next recommendation

Collect fresh CPU-only profiles without allocation sampling across Boolean
Reconciliation, Run-length, Unicode Scalar Pipeline, custom nominal `Eq`, and
one unrelated iterator/map control. Reconcile the call trees beneath cached
equality, type lookup, function invocation, and bytecode frame setup, and admit
a candidate only if the same exact interpreter-owned CPU leaf is material in
at least three unlike programs.

Why: the fixed-arity allocation wall is now closed, and the post-change
allocation owners diverge. Allocation-rate profiling heavily perturbs CPU
stacks, while retrying the previously closed raw-integer carrier family would
optimize a familiar metric without new broad evidence. This next work entails
bounded CPU-only profiles, call-tree reconciliation, a semantics test matched
to the selected shared leaf, and the same repeated equality/text/iterator/byte/
numeric guard gate. WASM remains deferred.
