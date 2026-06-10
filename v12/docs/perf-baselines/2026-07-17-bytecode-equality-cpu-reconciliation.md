# Bytecode equality CPU reconciliation — 2026-07-17

## Decision

Keep no production change. CPU-only profiles confirm that equality dispatch is
still a large cumulative path after the allocation tranches, but its cost is
distributed across function invocation, parameter coercion, bytecode frame
execution, raw-value materialization, and workload-specific method bodies.
Equality-cache lookup itself is too small to justify a cache-key redesign.

A temporary exact-`*runtime.FunctionValue` branch bypassed generic callable
classification after a cached equality hit. It preserved semantics and
allocations, but full repeated workstation averages were neutral-to-negative.
The branch was reverted completely.

No VM, runtime, compiler, stdlib, language, fixture, or benchmark-source change
is retained. A temporary custom nominal `Eq` program and all profile binaries
are removed after this record.

## CPU-only profiles

Profiles used the retained fixed-arity tree, canonical external `able-stdlib`,
one warmed measured call, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and
CPU 0. Allocation sampling was disabled.

| Workload | Runtime | `applyEqualityInterface` cumulative | Cached call cumulative |
| --- | ---: | ---: | ---: |
| Boolean Reconciliation | 834 ms | 36.59% | 34.15% |
| Run-length encode | 2.668 s | 40.98% | 38.72% |
| Unicode Scalar Pipeline | 6.714 s | 32.34% | 29.64% |
| Temporary custom nominal `Eq` | 462 ms | 44.44% | 42.22% |
| Iterator Collect control | 407 ms | not material | not material |
| Numeric Array Map control | 69 ms | not reached | not reached |

The cache lookup is only 1.22% flat in Boolean, 0.75% flat / 1.50%
cumulative in Run-length, 0.15% flat / 0.90% cumulative in Unicode, and below
the 10 ms sampling quantum in custom `Eq`. Aggregate `mapaccess2_faststr`
samples arise from several unrelated caches and environments; they are not
evidence for one shared map optimization.

Within cached equality calls, the recurring work is instead spread among:

- `invokeFunction(...)` and detached VM execution;
- `invokeFunctionBindArgsForSlotLayout(...)` and simple-type coercion checks;
- raw integer materialization and stack snapshots;
- cached call-name/native calls inside the invoked Able method;
- workload-specific character, interpolation, array, and nominal-field work.

No exact leaf is independently material in three unlike programs. The
parameter-coercion subtree is the strongest remaining shared candidate for a
separate admission audit, but this tranche does not assume that its cumulative
cost can be removed.

## Rejected direct-function candidate

The existing two-argument helper temporarily called a direct
`*runtime.FunctionValue` through `callResolvedFunctionValue(...)` when no call
node was present. All other callable shapes retained generic dispatch.
Coercion, reentrancy, retained-native arguments, and partial-copy tests passed,
and allocation counts were unchanged exactly.

Each timing is a separate process. Boolean and custom `Eq` received ten samples
per side after isolated slow candidate processes made the first cohort
volatile. Those samples remain in the averages, following the workstation
measurement rule.

| Workload | Samples/side | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 10 | 906.637 ms | 912.351 ms | 0.63% slower |
| Custom nominal `Eq` | 10 | 509.649 ms | 561.923 ms | 10.26% slower |
| Run-length encode | 5 | 2.7317 s | 2.7311 s | 0.02% faster; neutral |
| Unicode Scalar Pipeline | 3 | 7.2127 s | 7.3269 s | 1.58% slower |

Bypassing the small classification shell does not remove the coercion/frame
work below it and introduces no reliable broad win. Unrelated performance
guards were unnecessary after the candidate failed its four admission rows.

## Verification

- Candidate-time fixed-arity, cached-equality, coercion, reentrancy,
  retained-native, partial-copy, and partial-chain tests passed.
- The rejected branch was reverted before final correctness verification.
- All touched Go files remain below 1,000 lines and `git diff --check` passes.

## Next recommendation

Audit parameter-coercion planning beneath
`invokeFunctionBindArgsForSlotLayout(...)` across the same four equality
consumers plus Iterator Collect and another unrelated call-heavy workload.
Measure how often each parameter actually requires conversion versus merely
repeating an already-exact type check. Admit a cached coercion-plan or exact-
shape guard only if the same avoidable decision work is material in at least
three unlike programs and the cache remains valid across generic bindings,
method receivers, and dynamic calls.

Why: cache lookup and callable classification are now measured as small, while
parameter binding/coercion recurs beneath every large equality subtree and also
belongs to general function calls. The work entails temporary counters,
CPU-only profiles, generic/union/interface coercion correctness tests, and the
same repeated equality/text/iterator/byte/numeric performance gate. This is a
general call-boundary investigation, not an equality or benchmark-specific
fast path. WASM remains deferred.
