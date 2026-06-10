# Bytecode equality argument-ownership gate — 2026-07-17

## Decision

Keep the generic equality-dispatch ownership change. Cached equality dispatch
constructs a fresh two-element argument slice from the left and right values;
it now invokes the resolved callable through the existing caller-owned mutable
argument path. Parameter coercion may rewrite those two private slice slots
without first cloning the slice. No Able value, collection, caller-owned
argument list, benchmark identity, primitive name, or nominal type is inspected.

Also keep `boolean_reconciliation_small`, a bounded non-text fixture that
compares independently built boolean state vectors and reduces a reconciliation
checksum. It runs with identical output in compiled, tree-walker, and bytecode
modes and adds primitive equality performance coverage outside character/text
workloads.

## Admission profiles

Fresh one-process profiles used the post-scalar bytecode test binary,
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, CPU 0, canonical external
`able-stdlib`, a warmed `main`, and one bounded measured call.

| Workload | Baseline runtime | `ensureMutableCallArgs` flat objects |
| --- | ---: | ---: |
| Run-length encode | 2.701 s | 901,147 (33.03%) |
| Automata DFA | 3.468 s | 147,460 (15.33%) |
| Unicode Scalar Pipeline | 6.630 s | 1,589,296 (27.61%) |
| Boolean Reconciliation | 929.8 ms | 327,690 (19.20%) |
| Persistent Map `i32` control | 826.0 ms | no material runtime leaf; only cache initialization |

The exact clone therefore repeats in four unlike application shapes and two
primitive domains. An exploratory user-defined nominal record reconciliation
also reproduced the leaf at 20.22%, but its compiled CLI path required an
unavailable interpreter bridge. It was replaced rather than retained as a
shared benchmark. Existing custom-`Eq` execution fixtures remain the semantic
guard for user-defined implementations.

## Candidate semantics

`applyCachedEqualityDispatch(...)` previously called exported `CallFunction`,
which conservatively marked every supplied argument slice as borrowed. The
resolved function's parameter-coercion stage consequently copied the new local
two-element slice through `ensureMutableCallArgs(...)` before writing a
coerced slot.

The call site now uses `callCallableValueMutable(...)`. That declaration is
safe because the slice literal is created at the call site and is never shared.
It does not grant mutation access to the underlying Able values; it only lets
the call machinery replace elements in this private Go slice. Partial calls,
caller-provided slices, and other exported `CallFunction` users retain their
existing stable-copy rules.

A focused tree-walker/bytecode test invokes cached equality through a function
whose `i32` inputs require `f64` parameter coercion. Existing tests continue to
verify that ordinary `CallFunction` and partial-function calls do not mutate
their callers' argument slices.

## Repeated performance gate

Each sample is an independent process. Three-sample rows alternate pair order;
volatile Automata and Array Map rows add five candidate-first pairs.

| Workload | Samples/side | Baseline mean | Candidate mean | Timing | Allocation result |
| --- | ---: | ---: | ---: | ---: | ---: |
| Run-length encode | 3 | 2.5929 s | 2.4717 s | 4.67% faster | 2,230,700 -> 1,270,730 objects (-43.0%) |
| Automata DFA | 8 | 3.3877 s | 3.4034 s | 0.46% slower; neutral | 687,700 -> 555,472 objects (-19.2%) |
| Unicode Scalar Pipeline | 3 | 6.5681 s | 6.3786 s | 2.89% faster | 8,730,899 -> 6,961,412 objects (-20.3%) |
| Boolean Reconciliation | 3 | 926.268 ms | 841.345 ms | 9.17% faster | 1,136,476 -> 743,257 objects (-34.6%) |
| Iterator Collect guard | 3 | 423.170 ms | 418.197 ms | 1.18% faster | identical |
| Base64 guard | 3 | 67.796 ms | 68.203 ms | 0.60% slower; neutral | identical |
| Numeric Array Map guard | 8 | 69.301 ms | 72.356 ms | one 96 ms outlier | identical |
| Reverse Complement guard | 3 | 1.090 ms | 1.029 ms | 5.67% faster | effectively identical |
| String Split/Join guard | 3 | 1.0075 s | 1.0016 s | 0.59% faster | effectively identical |

Automata's five candidate-first pairs differ by only 0.22%. Array Map's five
candidate-first pairs are 68.753 ms candidate versus 68.774 ms baseline, a
0.03% improvement; its combined mean is distorted by one isolated 96 ms
candidate sample. Both are classified as neutral on this active workstation.

The post-change Boolean profile removes `ensureMutableCallArgs` entirely.
Measured memory falls from 28.08 MB to 15.49 MB and allocations from 1,136,538
to 743,319. The next allocation owner is now the remaining cached equality
dispatch itself, including construction of the one private two-value call
slice.

## Verification

- Focused cached-equality, parameter-coercion, caller-slice stability, partial
  call, and custom `Eq` fixture tests pass in 0.585 s.
- The runtime package passes in 0.052 s.
- Boolean Reconciliation prints `2139002832` in compiled, tree-walker, and
  bytecode modes.
- `just bench-catalog-check` passes with 35 portable applications, 79 local
  fixtures, and 114 combined programs; all feature-coverage tests pass.
- `just bench-selection-check` passes all protocol tests with the unchanged
  63-row reviewed selection.

## Next recommendation

Profile the remaining `applyCachedEqualityDispatch` allocation and CPU wall
across Boolean Reconciliation, Run-length, Unicode Scalar Pipeline, and a
user-defined custom-`Eq` execution workload. Evaluate a generic fixed-arity
two-argument call path only if it can remove the remaining private slice
allocation while remaining reentrant for nested equality and preserving
partial/coercing call behavior.

Why: this tranche removes the redundant defensive clone, but the post-change
Boolean profile still assigns 34.44% of sampled objects to cached equality
dispatch. The work entails escape-analysis and allocation profiles, a
reentrancy test where one equality method performs another equality call, a
coercing custom-`Eq` guard, and the same repeated scalar/iterator/byte/numeric
performance gate. A fixed-arity path is potentially applicable to every
two-operand interface dispatch without naming a benchmark or nominal type.
WASM remains deferred.
