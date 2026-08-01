# Top-three sustained compiled profile closure

Date: 2026-07-30

## Decision

**Retain no production code. Versioned Telemetry Pipeline, k-Nucleotide, and
Sudoku Masks have no exact material generated-code or generated-runtime owner
in common.**

The three applications are the largest current compiled rows by absolute time
above equivalent Go. All were regenerated with the current compiler under
`--no-fallbacks`, passed their public verifier, and produced 96-package final
Go dependency graphs that omit `able/interpreter-go/pkg/interpreter`.

Thirty independent main-only CPU processes, nine lightweight exact
main-allocation processes, and three one-object allocation-shape processes all
verified with zero failures and zero timeouts. Three additional unprofiled
smoke processes also verified.

After symbol normalization, the only flat-positive CPU symbol sampled in all
three merged profiles is `runtime.asyncPreempt`: 0.41% in Telemetry and 0.14%
in each of k-Nucleotide and Sudoku. It is Go runtime preemption noise, not an
Able lowering owner. No generated application, generated runtime, bridge, or
semantic helper symbol repeats in all three.

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, language, dependency, benchmark, fixture, frozen workspace,
or WASM source changed. With no candidate, no baseline/candidate/Go A/B cohort
was manufactured.

The exact 503,192 KiB task-created disk-backed workspace was removed after
validation. It contained the profiling compiler, Go cache, generated modules,
binaries, logs, summaries, and raw profiles. No `/tmp/able-*` artifact or new
Python cache remained.

## Protocol

The profiles used:

- CPU affinity: CPU 12;
- `GOMAXPROCS=1`;
- `GOGC=50`;
- `GOMEMLIMIT=1GiB`;
- ten CPU-only phase-profile processes per application;
- three lightweight exact `MemStats` phase deltas per application; and
- one `MemProfileRate=1` start/end allocation-shape process per application.

Each process had a 58-second execution limit. The longest exact allocation
profile was k-Nucleotide at 32.03 seconds; Telemetry took 28.91 seconds and
Sudoku 2.50 seconds. Every individual process stayed below the one-minute
rule.

The workstation carried sustained unrelated Marketlab CPU and disk load. The
ten CPU-profile process means of 2.223, 1.420, and 1.467 seconds are recorded
only as profile-run controls. They do not replace the current five-process
scoreboard or support a new timing claim. The profiler binaries were built
with the locally active Go 1.26.4 toolchain.

## Strict build controls

| Application | Source SHA-256 | Binary SHA-256 | Binary bytes | Dependencies | Interpreter |
| --- | --- | --- | ---: | ---: | --- |
| Versioned Telemetry Pipeline | `8701c450...6de20` | `f8fcc674...f7187` | 12,959,080 | 96 | absent |
| k-Nucleotide | `933749cb...a8a2a` | `b75d73d2...13277` | 15,488,040 | 96 | absent |
| Sudoku Masks | `222b321f...90e0` | `4c0a43e2...f71e` | 13,817,008 | 96 | absent |

All 42 profiled processes reproduced one stable verifier-approved output hash
per application. The current `ablec` binary has SHA-256
`8dcdbfdf8189a1d9e1459464625aad5e2e340ccae7b69b85d90efdb71f7fe66b`.

## CPU ownership

The ten main-only profiles per application were merged before attribution.

### Versioned Telemetry Pipeline

The merged profile contains 22.12 seconds of samples:

| Owner | Flat CPU | Cumulative CPU | Cumulative share |
| --- | ---: | ---: | ---: |
| checked signed multiply | 5.00 s | 5.07 s | 22.92% |
| signed divmod | 3.90 s | 4.14 s | 18.72% |
| checked signed add | 2.79 s | 2.81 s | 12.70% |
| `runtime.newobject` | 0.14 s | 4.26 s | 19.26% |
| Linear policy adapter | 0.43 s | 1.29 s | 5.83% |
| Adaptive policy adapter | 0.38 s | 1.71 s | 7.73% |

The arithmetic helpers retain Able overflow and division semantics. Their
general alternatives are already closed by unlike-program A/B evidence.
Policy dispatch is application-specific to this cohort.

### k-Nucleotide

The merged profile contains 14.09 seconds of samples:

| Owner | Flat CPU | Cumulative CPU | Cumulative share |
| --- | ---: | ---: | ---: |
| primitive HashMap key equality | 2.48 s | 3.06 s | 21.72% |
| primitive HashMap hash | 1.08 s | 1.66 s | 11.78% |
| generated HashMap entry search | 0.73 s | 3.96 s | 28.11% |
| generated HashMap raw set | 0.12 s | 7.49 s | 53.16% |
| generated HashMap raw get | 0.12 s | 5.82 s | 41.31% |
| `bridge.ToUint` | 0.11 s | 3.40 s | 24.13% |
| `bridge.ToInt` | 0.06 s | 2.17 s | 15.40% |
| `runtime.mallocgc` | 0.25 s | 4.20 s | 29.81% |

This is the canonical HashMap/runtime-value boundary already classified as a
named non-primitive nominal/kernel surface. It is material only in
k-Nucleotide here and cannot justify a HashMap-specific compiler rule.

### Sudoku Masks

The merged profile contains 14.58 seconds of samples:

| Owner | Flat CPU | Cumulative CPU | Cumulative share |
| --- | ---: | ---: | ---: |
| `find_best_empty` | 5.06 s | 13.16 s | 90.26% |
| `bit_count` | 2.24 s | 2.24 s | 15.36% |
| checked signed multiply | 2.12 s | 2.26 s | 15.50% |
| signed divmod | 2.04 s | 2.43 s | 16.67% |
| `square_index` | 1.42 s | 6.13 s | 42.04% |

The solver already operates on native `i32`, Boolean, and static Array
carriers. Its remaining cost is exact-cover search plus checked arithmetic.
The same arithmetic helpers occur in Telemetry, but not in k-Nucleotide, and
therefore fail the required three-unlike-application gate.

## Exact allocation ownership

Three independent lightweight main-phase counters were stable:

| Application | Allocated-byte samples | Mean bytes | Allocation samples | Mean allocations | Mean GC |
| --- | --- | ---: | --- | ---: | ---: |
| Versioned Telemetry Pipeline | 430,788,208 / 430,788,160 / 430,788,296 | 430,788,221.33 | 13,325,303 / 13,325,303 / 13,325,305 | 13,325,303.67 | 351.00 |
| k-Nucleotide | 598,183,536 / 598,183,552 / 598,183,552 | 598,183,546.67 | 12,232,489 / 12,232,489 / 12,232,489 | 12,232,489.00 | 274.33 |
| Sudoku Masks | 618,856 / 618,856 / 618,856 | 618,856.00 | 15,018 / 15,018 / 15,018 | 15,018.00 | 0.00 |

The one-object profiles identify different semantic sites:

- Telemetry allocates 13,208,878 escaping `Sample` objects in its hot update
  loop. They enter generic window storage and retain identity-bearing nominal
  semantics.
- k-Nucleotide allocates 7,999,998 `bridge.ToUint` and 3,961,373
  `bridge.ToInt` small-integer runtime values under HashMap raw set/get.
- Sudoku's measured solver is allocation-free after setup. Its small total is
  dominated by 7,600 output-string concatenations, 5,000 parse-path
  allocations, and 1,000 mask-array constructions.

The allocation profile writer itself is visible in the start/end profile
subtraction, especially for low-allocation Sudoku. Selection therefore uses
the independent exact phase counters for totals and the line-attributed
profile only for application-site shape.

Allocation and GC ancestry repeat in Telemetry and k-Nucleotide, but not
Sudoku. The concrete lifetimes also differ: escaping identity-bearing
`Sample` pointers versus runtime integer boxes. They are not one removable
general ABI.

## Admission result

The exact cross-application intersections are:

- checked arithmetic: Telemetry and Sudoku only;
- allocation and GC: Telemetry and k-Nucleotide only;
- HashMap hashing, equality, and runtime conversion: k-Nucleotide only;
- interface policy dispatch: Telemetry only; and
- exact-cover search, bit counting, and array scans: Sudoku only.

No exact generated-code or generated-runtime owner is material in all three.
Accordingly:

- no candidate was implemented;
- no closed checked-arithmetic, HashMap, nominal-identity, Array, GC, or
  launch-floor route was reopened;
- no benchmark, application, container, or non-primitive nominal special case
  was added; and
- no baseline/candidate/Go A/B was run.

The machine-readable companion is
`2026-07-30-top-three-sustained-compiled-profile-no-go.json`.

## Next

Build an exact semantic-parent boundary census from the current 66-row static
native-boundary evidence before choosing another compiled profile cohort.

Why: choosing applications by duration produced three different owners, while
coarse categories such as “allocation,” “interface,” or “runtime conversion”
would falsely group different semantics. Another duration-only cohort would
likely repeat this no-go.

What it entails: normalize each direct-main `runtime.Value`, bridge,
interface, union, callable, and nominal conversion by exact callee plus its
generated semantic parent; join those identities to current timing excess and
retained profile reach; and rank only exact triples spanning unlike
applications. Regenerate a module only when the retained census lacks enough
line-level identity. Admit profiling only for a boundary whose same semantic
parent appears in three applications. Preserve dynamic and escaping paths and
do not create named-container or non-primitive nominal rules.

Why it matters: the project goal is to remove real compiled/interpreted or
native/erased crossings. Exact parent identity can distinguish a broadly
lowerable boundary from three unrelated calls to the same generic helper,
reducing wasted profiling and keeping the next candidate general.
