# Statically Monomorphic Captured-Callable Carrier Retained

Date: 2026-07-30

Decision: retain the general compiled callable-inference change.

## Immediate result

Fresh local lambdas now keep one native Go callable carrier when every
statically known use agrees on a fully bound parameter and result signature.
The inference reaches direct calls and direct callable arguments, including a
captured invocation inside a statically typed inline callback. It also follows
generic named-union instance-method callbacks and sequentially known local
receivers inside nested blocks.

The analysis remains fail-closed. Dynamic, conflicting, stored, escaping,
host, runtime-service, insufficiently typed function, and iterator uses keep
the erased `runtime.Value` carrier. The implementation contains no benchmark,
application, `Result`, container, or non-primitive nominal-name rule.

One adjacent general inference defect was corrected: ordinary resolved
instance methods now account for their `self` parameter when selecting an
argument type, while generic named-union instance methods use their existing
concrete method resolution path.

## Generated carrier proof

| Application | Before | After |
| --- | --- | --- |
| Versioned Telemetry Pipeline | `runtime.Value, runtime.Value -> runtime.Value` | `Sample, Sample -> i64` |
| Manifest Normalization | `runtime.Value, runtime.Value -> NormalizedManifest` | `ManifestRecord, i64 -> NormalizedManifest` |
| Binary Event Log | `runtime.Value, runtime.Value -> i64` | `EventRecord, i64 -> i64` |

All three strict dependency graphs omit `pkg/interpreter`. No stdlib, runtime,
interpreter, bytecode VM, language, dependency, or WASM source changed.

## Repeated A/B results

Every process passed its public verifier. The longer telemetry cohort used
five order-balanced baseline/candidate/Go processes per lane. The two short
applications used ten order-rotated processes and a monotonic high-resolution
clock because centisecond process timing cannot resolve their change.

| Application | Baseline mean | Candidate mean | Change | Go mean | Candidate/Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| Versioned Telemetry Pipeline | 54.3540 s | 2.1260 s | -96.09% | 0.2060 s | 10.32x |
| Manifest Normalization | 0.012190 s | 0.007667 s | -37.11% | 0.003480 s | 2.20x |
| Binary Event Log | 0.114900 s | 0.049313 s | -57.08% | 0.006457 s | 7.64x |

Three exact main-phase allocation-counter processes per lane confirm that the
wall improvements come from removing the erased boundary:

| Application | Allocated bytes | Change | Allocations | Change | GC cycles | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Versioned Telemetry | 33.823 GB -> 0.431 GB | -98.73% | 1,056,825,647 -> 13,325,303 | -98.74% | 28,244 -> 351 | -98.76% |
| Manifest Normalization | 3.560 MB -> 1.397 MB | -60.76% | 76,230 -> 27,075 | -64.48% | 3 -> 1 | -66.67% |
| Binary Event Log | 62.024 MB -> 9.322 MB | -84.97% | 1,020,540 -> 171,066 | -83.24% | 57 -> 8 | -85.96% |

The standard five-process scorecard harness was also refreshed with Go
1.26.5 on CPU 12. It records Binary Event Log at 0.0900 seconds versus Go at
0.0097 seconds, Manifest Normalization at 0.0440 seconds versus Go at 0.0048
seconds, and Versioned Telemetry at 2.0680 seconds versus Go at 0.2078 seconds.
The Manifest row is dominated by process-launch noise at the harness's
centisecond resolution; the ten-sample high-resolution A/B and the independent
64.48% allocation-count reduction are the retention evidence.

## Correctness and safety guards

Focused tests cover direct monomorphic invocation, a native nested callback,
a generic named-union callback, a future block binding, stored nested escape,
conflicting nested signatures, and execution of the native callback result.
The existing direct-argument, imported callable, dynamic, and conflict guards
remain green. Generic named-union and stdlib `Option`/`Result` specialization
guards pass. The parser fixture lane, every non-compiler Go package, all 34
bounded compiler batches, and the complete bytecode fixture corpus pass.

Machine-readable samples and counters are in
`2026-07-30-statically-monomorphic-captured-callable-retained.json`. The
official verifier-backed source report is
`2026-07-30-statically-monomorphic-captured-callable-compiled.json`.

## What comes next

Refresh interpreter-free compiled CPU and exact allocation profiles across at
least three unlike current applications.

Why: the erased captured-callable owner has been removed, so the old profiles
cannot identify the largest remaining shared cost.

What it entails: generate current strict binaries, confirm that their
dependency graphs still omit the interpreter, collect repeated CPU and exact
allocation evidence, and select only an exact general compiler/runtime owner
that is material in all three.

Why it matters: the retained rule produces large gains but the applications
remain 2.20x-10.32x behind equivalent Go. Reprofiling is the shortest honest
route toward the 95%-of-Go goal without reopening closed routes or introducing
application, container, or nominal-type special cases.
