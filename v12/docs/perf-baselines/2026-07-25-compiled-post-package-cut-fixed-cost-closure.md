# Compiled post-package-cut fixed-cost closure

Date: 2026-07-25

## Decision

Retain no compiler or runtime change from this tranche.

The post-interpreter-package-cut launch floor is real but smaller than the
centisecond scorecard suggested. Across six unlike strict applications,
high-resolution whole-process Able means range from 3.743 ms to 15.915 ms.
Generated bootstrap and registration account for 0.393-0.930 ms. No exact
open registration, initialization, or allocation owner is both material and
shared by at least three unlike applications.

The remaining large gaps are predominantly inside application work. The next
tranche should therefore intersect main-phase CPU and allocation owners across
the short applications that allocate at least about 1 MB after registration,
rather than micro-optimizing semantically required registration.

Machine-readable results are in
`2026-07-25-compiled-post-package-cut-fixed-cost-closure.json`.

## Scope and controls

The cohort covers six unlike current strict target misses:

| application | principal coverage |
| --- | --- |
| Dependency Plan | graph traversal, Arrays, topological processing |
| Concurrent Scene Tiles | goroutines, futures, worker aggregation |
| Document Audit | file input, text processing, iteration |
| Array Slice Window | native Array slicing and copying |
| Option/Result Config | unions, errors, generic callables |
| Sensor Calibration | nominal values, numeric work, file input |

Every Able application was freshly generated with the current compiler and
`--no-fallbacks`. Every Able and Go smoke process passed the existing Ruby
verifier and produced the expected stdout SHA-256. Each final Able dependency
graph contains 96 packages and omits `able/interpreter-go/pkg/interpreter`.

Serial applications ran pinned to CPU 5 with `GOMAXPROCS=1` and
`GOMEMLIMIT=1GiB`. Concurrent Scene Tiles used CPUs 5,10,15,11,
`GOMAXPROCS=4`, and `ABLE_EXECUTOR=goroutine`. File-driven applications used
their suite-defined working directories and inputs.

All large build and profile artifacts were placed under disk-backed
`/var/tmp`. Temporary generated-source instrumentation was applied only to
copied generated modules after the normal binaries and source hashes were
recorded; it was never applied to repository compiler sources.

## Whole-process result

The runner used monotonic clocks around direct child execution and collected
process resource usage. Each row contains 100 accepted Able and 100 accepted
Go processes after ten order-balanced warmup pairs. All 1,200 accepted
processes and all 120 warmups verified their output hashes.

| application | Able mean ms | Able p50 | Able p95 | Go mean ms | Go p50 | Go p95 | Able / Go | excess ms |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Dependency Plan | 4.174 | 4.157 | 4.410 | 1.432 | 1.426 | 1.574 | 2.915x | 2.742 |
| Concurrent Scene Tiles | 4.477 | 4.365 | 5.337 | 1.111 | 1.074 | 1.487 | 4.030x | 3.366 |
| Document Audit | 3.743 | 3.320 | 5.391 | 1.418 | 1.284 | 2.012 | 2.640x | 2.325 |
| Array Slice Window | 5.392 | 5.212 | 6.375 | 1.349 | 1.321 | 1.586 | 3.997x | 4.043 |
| Option/Result Config | 15.915 | 15.580 | 18.682 | 1.025 | 1.008 | 1.243 | 15.527x | 14.890 |
| Sensor Calibration | 9.959 | 8.770 | 15.179 | 2.319 | 2.089 | 3.765 | 4.294x | 7.640 |

These measurements supersede the coarse 24-52 ms Able and 3.8-5.3 ms Go
scorecard floor for fixed-cost attribution. They do not replace the governing
scorecard rows, whose harness and precision differ.

## Internal phase attribution

Temporary generated-source timers measured `runMain`, `RegisterIn`,
`RunRegisteredMain`, and registration descendants. Each row is the mean of
100 independently launched and verifier-backed instrumented processes after
ten warmups. Instrumentation changes layout and adds reporting cost, so the
phase values identify ownership and should not be substituted for the normal
whole-process means.

| application | outside runMain ms | bootstrap ms | RegisterIn ms | method packages ms | interface dispatch ms | packages ms | package inits ms | app main ms |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Dependency Plan | 1.508 | 0.519 | 0.512 | 0.052 | 0.153 | 0.243 | 0.008 | 0.645 |
| Concurrent Scene Tiles | 1.907 | 0.851 | 0.841 | 0.107 | 0.138 | 0.401 | 0.011 | 2.145 |
| Document Audit | 1.869 | 0.930 | 0.920 | 0.113 | 0.145 | 0.510 | 0.011 | 1.091 |
| Array Slice Window | 1.990 | 0.566 | 0.558 | 0.049 | 0.224 | 0.214 | 0.010 | 4.921 |
| Option/Result Config | 2.374 | 0.404 | 0.393 | 0.057 | 0.070 | 0.195 | 0.011 | 13.466 |
| Sensor Calibration | 2.287 | 0.938 | 0.928 | 0.163 | 0.129 | 0.482 | 0.010 | 6.940 |

The remaining registration time also includes small seed and builtin phases.
No exact registration descendant reaches even 0.511 ms, and no descendant is
material in three unlike normal whole-process rows. Aggregate registration is
measurable, but it combines required package definitions, method
registrations, interface dictionaries, import visibility, and package
initializers rather than exposing one removable general owner.

## Allocation and initialization evidence

Twenty separate phase-allocation runs per application all verified:

| application | bootstrap bytes | bootstrap allocs | app-main bytes | app-main allocs | app-main GCs |
| --- | ---: | ---: | ---: | ---: | ---: |
| Dependency Plan | 358,726 | 4,884.60 | 198,712 | 14,406 | 0 |
| Concurrent Scene Tiles | 457,130 | 6,225.95 | 1,192,276 | 37,056 | 0 |
| Document Audit | 585,675 | 7,653.80 | 373,474 | 1,952 | 0 |
| Array Slice Window | 258,668 | 3,566.25 | 1,441,352 | 24,012 | 0 |
| Option/Result Config | 215,716 | 2,958.25 | 9,937,764 | 182,900 | 3 |
| Sensor Calibration | 557,231 | 7,170.65 | 3,610,043 | 57,534 | 1 |

Three verifier-backed `GODEBUG=inittrace=1` processes per language and
application show Able initialization completing in 0.897-1.020 ms versus
0.290-0.440 ms for Go. Generated `main.init` contributes only
0.053-0.177 ms. The remainder includes parser/driver imports retained by
fallback-capable launcher helpers in the generated source; these helpers and
their imports are declaration-only linker roots already closed by the
generated static-closure analysis.

One exact bootstrap allocation profile per application repeats these
semantically rooted leaders:

- AST identifiers and simple type expressions used by runtime-visible
  definitions;
- compiled method registration;
- compiled interface-dispatch registration and union-target expansion;
- public import seeding.

The parent owners repeat, but the exact timed phases are only 0.049-0.224 ms
for method/interface work and 0.195-0.510 ms for package work. Public import
seeding is only a subset of the package phase. Eliminating duplicate transient
callable construction during import seeding was inspected, but its upper
bound is too small to satisfy the three-program materiality gate.

## Admission result

No candidate advanced to implementation or twenty-cohort A/B/Go measurement:

1. Process and Go initialization are outside an Able lowering rule.
2. Declaration-only parser/driver imports are an already-closed linker route.
3. Generated definition and interface registration is semantically required
   by imports, callbacks, dynamic calls, interfaces, and fallback-capable
   launchers. Workload-based pruning remains disallowed.
4. No exact registration descendant is material in three unlike programs.
5. Array Slice Window, Option/Result Config, and Sensor Calibration instead
   spend 4.921-13.466 ms in application main; their main allocations exceed
   bootstrap allocations by 5.6x-46.1x.

This is a verifier-backed negative result, not evidence that fixed cost is
zero. It establishes that the next material compiled opportunity is below the
application-work boundary rather than in shared launch registration.

## Verification

- 12/12 fresh Able/Go smoke processes passed.
- 1,200/1,200 accepted whole-process measurements passed, plus 120 warmups.
- 120/120 phase-allocation processes passed.
- 600/600 instrumented phase-timing processes passed, plus 60 warmups.
- 36/36 initialization-trace processes passed.
- 6/6 exact bootstrap-allocation profile processes passed.
- `go test ./cmd/ablec` passed in 5.873 seconds.
- All six strict Able dependency graphs omit the interpreter package.

No production compiler, runtime, interpreter, VM, language, stdlib,
dependency, or WASM change was made.

## Recommendation

Next run a main-phase CPU/allocation owner intersection across Concurrent
Scene Tiles, Array Slice Window, Option/Result Config, and Sensor Calibration.
They are unlike applications and each allocates at least about 1 MB after
registration. Collect repeated main-only CPU profiles and exact allocation
profiles, classify every shared leaf by native carrier versus
`runtime.Value`/semantic boundary, and advance at most one exact general owner
only if it is material in at least three programs. Use a fourth program as a
negative or regression control and require semantic guards plus at least
twenty order-balanced baseline/candidate/Go cohorts before retaining code.

This is next because the current decomposition places the remaining material
cost inside lowered application execution. It is important because matching
Go now depends on finding a broadly repeated boxing, conversion, dispatch, or
allocation boundary in real application work; registration micro-optimization
cannot close the measured gaps.
