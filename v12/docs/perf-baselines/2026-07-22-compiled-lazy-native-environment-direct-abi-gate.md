# Compiled lazy native-environment direct ABI gate — 2026-07-22

## Decision

Retain the synchronization-free direct-method ABI refinement. A statically
resolved compiled instance-method call now passes the active runtime
environment directly and does not construct a `NativeCallContext` on its
successful path. The context is constructed only after the direct path is
bypassed or when Able control must be converted to a native error.

The ordinary `NativeFunctionValue` wrapper remains externally unchanged for
bound methods, partial application, UFCS, overloads, interface dictionaries,
builtins, interpreter thunks, and dynamic fallback. The rule applies to every
generated compiled instance method and names no nominal type, container,
stdlib API, application, or benchmark. No canonical-stdlib, bytecode,
language, or WASM change was needed.

## Semantic audit

The direct wrapper's normal path consumes `NativeCallContext.Env` only:

- package environment swapping uses `Env`;
- generated compiled bodies receive their ordinary typed arguments and do not
  consume the native context;
- `RuntimeData` remains reachable through the environment;
- raised-control conversion selects its interpreter environment from `Env`;
- neither the compiled body adapter nor raised-control conversion reads
  `NativeCallContext.State`.

The retained entry therefore carries `*runtime.Environment`, not a context or
an environment/state pair. Its error branch constructs a context containing
that environment immediately before `__able_control_to_error`. Dynamic and
fallback paths retain the original full context including `State`.

Two correct exploratory forms were discarded. Separate environment/state
parameters expanded every compatibility wrapper. A by-value pair preserved
allocations but produced a larger binary and an unnecessary transport helper.
The environment-only form keeps each compatibility adapter as one inlinable
return and grows the N-Body binary by only 6,216 bytes (0.035%).

## Measurement protocol

The retained split-receiver binaries are the current side; final environment-
only binaries were built once and preserved. All workstation samples remain
in the companion JSON. Cohorts reverse variant order and use one logical CPU,
`GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, public verifiers, and a 55-second
per-process cap.

The final gate contains 488 successful verifier-backed benchmark processes
with zero failures or timeouts:

- 320 owner timings: 40 processes per variant in four applications;
- 148 unrelated guards: 20 N-Body processes per variant, six Binary Trees
  processes per variant, and 16 per variant for K-Nucleotide, Matrix Multiply,
  and Mutex Ledger;
- 20 exact candidate allocation-stat processes.

## Owner wall gate

| Application | Samples/variant | Current mean | Candidate mean | Delta | Medians |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 40 | 0.62175 s | 0.62475 s | +0.48% | 0.615 / 0.590 s |
| Option/Result Config | 40 | 0.17725 s | 0.17150 s | -3.24% | 0.170 / 0.170 s |
| Manifest Normalization | 40 | 0.18825 s | 0.19050 s | +1.20% | 0.180 / 0.180 s |
| Policy Record Dispatch | 40 | 0.20650 s | 0.20875 s | +1.09% | 0.210 / 0.210 s |

Option's improvement is statistically separated in the approximate difference
interval. The other three intervals cross zero. Binary retains a 1.72-second
candidate workstation pause instead of trimming it; its candidate median is
better. Manifest and Policy medians are unchanged. This is a neutral broad
wall result, not a claim that every owner became faster.

## Exact allocation gate

| Application | Current objects | Candidate objects | Object delta | Current bytes | Candidate bytes | Byte delta |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 4,240,162.6 | 4,014,876.4 | -5.31% | 267,265,742.4 | 261,858,484.8 | -2.02% |
| Option/Result Config | 1,212,866.8 | 1,065,408.4 | -12.16% | 41,719,256.0 | 38,180,118.4 | -8.48% |
| Manifest Normalization | 976,801.0 | 956,391.4 | -2.09% | 44,412,153.6 | 44,630,545.6 | +0.49% |
| Policy Record Dispatch | 955,685.0 | 950,050.6 | -0.59% | 48,604,616.0 | 48,470,915.2 | -0.28% |

Objects fall stably in every owner across five independent processes and match
removal of one remaining hot call-context allocation. Manifest's byte count is
slightly higher despite 20,410 fewer objects; byte counters varied between its
processes, so the exact object count and wall gate govern this decision.

## Unrelated guard gate

| Guard | Samples/variant | Current mean | Candidate mean | Delta |
| --- | ---: | ---: | ---: | ---: |
| N-Body | 20 | 0.1780 s | 0.1650 s | -7.30% |
| Binary Trees | 6 | 30.6567 s | 31.0733 s | +1.36% |
| K-Nucleotide | 16 | 3.0763 s | 3.1256 s | +1.61% |
| Matrix Multiply | 16 | 1.1925 s | 1.2300 s | +3.14% |
| Mutex Ledger | 16 | 0.5425 s | 0.5313 s | -2.07% |

Every approximate difference interval crosses zero. An earlier 40-sample
N-Body audit of the larger environment/state prototype also found identical
main-phase allocation counts—41 objects and 1,096 bytes—proving N-Body does
not exercise this optimized path. These rows are broad non-regression guards,
not promoted speedup claims.

## Verification

Focused generated-source, raised-control, standalone generic-union, overload,
internal-name collision, and opt-in execution-context compatibility tests pass.
A broader executable selection covering optional parameters, bound methods,
imported and generic nominal methods, safe navigation, interface coercion,
UFCS, overloads, and dynamic fallback passes in 26.5 seconds. No test exceeds
one minute. All touched generator files remain below 1,000 lines and
`git diff --check` is clean.

The companion JSON retains every selected timing/allocation sample, hashes,
binary sizes, and semantic-contract flag:
`2026-07-22-compiled-lazy-native-environment-direct-abi-gate.json`.

## Next direction

Refresh bounded main-only CPU and exact allocation profiles for Binary Event
Log, Option/Result Config, Manifest Normalization, and Policy Record Dispatch
from this retained ABI. The two known per-call allocations—the bound-method
box/redispatch and the receiver slice/context pair—are now gone, so old
profiles cannot select the next candidate reliably.

Admit another compiler change only when one concrete generated/runtime leaf
repeats across at least three unlike applications and survives the same broad
repeated-average gate. Likely categories to inspect are general nominal result
conversion, union value conversion, or another generated allocation descendant,
but the profiles must choose among them. Keep bytecode on its separately
measured frontier and continue to defer WASM.
