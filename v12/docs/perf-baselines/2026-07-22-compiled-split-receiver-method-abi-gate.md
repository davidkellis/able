# Compiled split-receiver method ABI gate — 2026-07-22

## Decision

Retain the general generated compiler change. Every generated compiled
instance method now has an optional direct core that receives its receiver
separately from explicit arguments. Statically resolved generic-union calls
use that core and no longer allocate a fresh receiver-prefixed slice.

The ordinary combined-argument wrapper remains registered and delegates to
the same core. Bound methods, partial calls, UFCS, interface dictionaries,
overloads, builtins, interpreter thunks, dynamic resolution, and the bridge
fallback therefore retain their existing ABI. The rule names no union,
record, container, stdlib type, or benchmark. No canonical-stdlib, bytecode,
language, or WASM change was needed.

## Audit and implementation

The feasibility audit rejected replacing `NativeFunctionValue.Impl`: that
would force unrelated dynamic callable paths to understand a second argument
shape. Instead, compiled-method entries gained optional `direct` metadata.
Single generated instance methods register a direct core; static generic-union
dispatch calls it when present and retains the old receiver-injection fallback
for builtins or other entries.

The direct core preserves:

- explicit and optional-last arity behavior;
- active package environment and `RuntimeData`;
- argument conversion and receiver position;
- Able control/native error conversion and nil normalization;
- mutable nominal-struct receiver writeback;
- the ordinary wrapper used by bound, partial, UFCS, and dynamic calls.

The large runtime-call generator was also split at a function boundary so all
modified generator files remain below 1,000 lines; generated code is
unchanged by that mechanical refactor.

## Measurement protocol

Current binaries from the retained direct-known-method implementation were
preserved before the source edit. Candidate binaries were built once and
reused. The gate contains 300 successful verifier-backed benchmark processes
with zero failures or timeouts:

- 160 owner timing processes: 20 per variant for four applications;
- 120 unrelated guard processes;
- 20 exact candidate allocation-stat processes.

All workstation samples are retained. Cohorts reverse current/candidate order.
Volatile N-Body, K-Nucleotide, and Matrix Multiply use 16 samples per variant;
Binary Trees and Mutex Ledger use six. Runs use one logical CPU,
`GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and a 55-second per-process cap.

## Owner wall gate

| Application | Samples/variant | Current mean | Candidate mean | Delta |
| --- | ---: | ---: | ---: | ---: |
| Binary Event Log | 20 | 0.6465 s | 0.6400 s | -1.01% |
| Option/Result Config | 20 | 0.1925 s | 0.1900 s | -1.30% |
| Manifest Normalization | 20 | 0.1975 s | 0.1905 s | -3.54% |
| Policy Record Dispatch | 20 | 0.2190 s | 0.2130 s | -2.74% |

All four unlike owners improve on the full retained average. Policy's median
moves from 0.210 to 0.215 seconds, a single ambiguous short-timer result, while
its 20-process mean and exact allocation evidence improve.

## Exact allocation gate

| Application | Current objects | Candidate objects | Object delta | Current bytes | Candidate bytes | Byte delta |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 4,465,436.8 | 4,240,162.6 | -5.04% | 273,425,644.8 | 267,265,742.4 | -2.25% |
| Option/Result Config | 1,360,322.0 | 1,212,866.8 | -10.84% | 46,044,529.6 | 41,719,256.0 | -9.39% |
| Manifest Normalization | 997,350.6 | 976,801.0 | -2.06% | 45,704,526.4 | 44,412,153.6 | -2.83% |
| Policy Record Dispatch | 961,314.8 | 955,685.0 | -0.59% | 48,748,643.2 | 48,604,616.0 | -0.30% |

The generated direct path contains no receiver-prefix append. The object
reductions are extremely stable across five independent processes per owner
and match removal of one hot per-call allocation.

## Unrelated guard gate

| Guard | Samples/variant | Current mean | Candidate mean | Delta |
| --- | ---: | ---: | ---: | ---: |
| Binary Trees | 6 | 31.120 s | 30.667 s | -1.46% |
| N-Body | 16 | 0.1888 s | 0.1738 s | -7.95% |
| K-Nucleotide | 16 | 3.1206 s | 2.9494 s | -5.49% |
| Matrix Multiply | 16 | 1.2475 s | 1.2425 s | -0.40% |
| Mutex Ledger | 6 | 0.5333 s | 0.5317 s | -0.31% |

No unrelated averaged workload regresses. The short-run improvements are not
claimed as independent speedups; they establish that the general ABI does not
trade owner allocation wins for a broad wall regression.

## Verification

The focused generated-source and executable tests pass:

```text
go test ./pkg/compiler -run 'TestCompilerStaticGenericUnionKnownMethod|TestCompilerTreatsSelfTypedFirstMethodParamAsInstanceReceiver|TestCompilerStandaloneGenericNamedUnionMethods' -count=1 -timeout 60s
```

A broader method/boundary selection covering optional arguments, bound method
values, imported and generic nominal methods, safe navigation, interface
coercion, fallback absence, and generic-union semantics also passes in 21.9
seconds. No test exceeds one minute.

The companion JSON retains all timing/allocation samples and semantic-contract
flags: `2026-07-22-compiled-split-receiver-method-abi-gate.json`.

## Next direction

Audit a synchronization-free, lazy native-call-context form for this optional
direct ABI. The earlier pool experiment and this tranche independently removed
one of the two helper allocations; the remaining successful-call allocation is
the fresh `NativeCallContext`. Passing the already-known environment and state
separately, and constructing a context only for uncommon error conversion,
could remove it without the rejected pool overhead.

First prove environment, `RuntimeData`, package swapping, and raised-control
semantics across the direct and fallback paths. Prototype only if the common
path remains allocation-free without reviving the closed program-wide or
spawn-scoped execution-context designs. Gate it against the same four owners
and unlike controls with repeated verifier-backed averages. Keep bytecode on
its independently measured frontier and continue to defer WASM.
