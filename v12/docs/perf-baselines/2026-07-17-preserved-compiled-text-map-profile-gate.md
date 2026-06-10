# Preserved compiled text/map profile gate

Date: 2026-07-17

## Decision

Complete the preserved compiled Word Frequency, Document Audit, and Dependency
Plan gate and retain no compiler, generated-runtime, bytecode VM,
canonical-stdlib, benchmark, fixture, or language change. The fingerprinted
generated-main profiles divide into three different semantic owners. No
String, map, iterator, Queue, Deque, graph, or named-container candidate meets
the three-unlike-application admission rule.

The simultaneously captured Go initialization trace does expose one exact
three-way process wall outside generated `main`: the imported interpreter
package eagerly constructs bytecode integer caches in 57-58 ms, allocating
about 38.0 MB in 707,341 objects before every static compiled application.
That is a general runtime/package-boundary target, but this tranche does not
retry the rejected lazy-cache candidate or retain dummy heap ballast.

## Preserved timing evidence

All binaries were built before timing and then reused without rebuilding.
Every process used canonical external `able-stdlib`, catalog CPU 0 from pool
0-15, `GOMAXPROCS=1`, a 55-second limit, catalog working directories and
arguments, and the public Ruby verifier. The generated tree and executable
fingerprints in the report match the later profile artifacts byte-for-byte.

| Application | Able pooled mean | Cohort spread | Fresh Go | Able / Go | Timing status |
| --- | ---: | ---: | ---: | ---: | :---: |
| Word Frequency | 0.2000 s | 4.08% | 0.0063 s | 31.75x | accepted |
| Document Audit | 0.0810 s | 18.92% | 0.0051 s | 15.88x | rejected |
| Dependency Plan | 0.0720 s | 5.71% | 0.0046 s | 15.65x | accepted |

The primary lane completed 15/15 fresh Go and 30/30 Able executions. Document
Audit exceeded the 15% independent-cohort gate, so two additional five-run
cohorts reused the same executable. All ten additional executions verified.
The twenty-run pooled mean is 0.0775 s, but the four cohort means span
0.072-0.088 s, or 22.22%. The larger sample therefore confirms rejection
rather than averaging away the disagreement.

Machine-readable timing records:

- `2026-07-17-preserved-text-map-go-reference.json`
- `2026-07-17-preserved-text-map-cohorts.json`
- `2026-07-17-preserved-text-map-document-audit-extended.json`

## Generated-main CPU profiles

CPU-only phase profiling used the exact timed executables with
`GOMEMLIMIT=1GiB` and `GOGC=50`. Every launch passed its verifier and each
application produced one stable stdout hash. Merging 20 Word Frequency, 100
Document Audit, and 100 Dependency Plan `main.cpu.pprof` files supplied the
following attribution:

| Application | Main samples | Material concrete owners |
| --- | ---: | --- |
| Word Frequency | 2.76 s | `__able_hash_map_find_entry` 47.10% flat/48.55% cumulative; `String.split` 34.06% cumulative; byte-array/String conversion 11.59% cumulative |
| Document Audit | 100 ms | iterator generator/filter 60% cumulative; `String.contains` 50% cumulative; Go span scanning 40% flat |
| Dependency Plan | 150 ms | deployment-plan resolver 93.33% cumulative; checked add 13.33% flat; `ToInt` and Queue dequeue each 20% cumulative |

The apparent common leaf, `runtime.tryDeferToSpanScan`, is collector work with
different Able callers: large String/map conversion allocation, iterator and
file-line allocation, and Queue/graph objects. It is not one compiler or
stdlib operation. Hash-map entry lookup occurs materially only in Word
Frequency; String containment only in Document Audit; Queue/Deque and checked
integer work only in Dependency Plan. The profile gate therefore admits no
application-main candidate and does not run an A/B control for a nonexistent
candidate.

## Exact allocation counters

Separate allocation-only phase launches avoided CPU-profiler distortion. The
phase counters, not allocation-profile serialization stacks, are authoritative:

| Application | Main bytes | Main allocations | Main allocation shape |
| --- | ---: | ---: | --- |
| Word Frequency | 31,184,904 | 720,431 | String/byte conversion, formatting, split, and map values |
| Document Audit | 374,152 | 1,967 | iterator union values and file-line materialization |
| Dependency Plan | 475,192 | 18,631 | Queue/Deque, graph lists, and common integer boxes |

Allocation snapshot writing dominates the start/end `.pprof` difference for
the two short applications, but it is outside these exact main counters. The
application-owned residuals again do not repeat across all three.

## Shared process-start wall

One verifier-backed `GODEBUG=inittrace=1` launch per fingerprinted binary
identifies why the short complete-process rows remain far from Go even though
their sampled generated mains are tiny:

| Application | Interpreter package init | Init bytes | Init allocations |
| --- | ---: | ---: | ---: |
| Word Frequency | 58 ms | 38,003,880 | 707,341 |
| Document Audit | 57 ms | 38,003,880 | 707,341 |
| Dependency Plan | 58 ms | 38,003,896 | 707,341 |

The owner is the eager bytecode small-integer and raw-i32 cache construction
in the monolithic interpreter package imported by the compiled bridge. It is
an exact cross-application wall, unlike the divided generated-main costs.

The prior lazy-initialization experiment is still disqualified: although it
removed this startup work from short compiled applications, it reduced their
initial live heap and made allocation-heavy Binary Trees collect about 35%
more often, regressing its wall time 7.74% across twenty alternating pairs.
Changing `GOGC`, keeping dummy ballast, or selecting behavior by workload
would violate user runtime policy and the broad-program rule.

## Verification and cleanup

- 15/15 fresh Go reference executions verified.
- 40/40 preserved Able timing executions verified, including the extended
  volatile Document Audit cohorts.
- 220/220 CPU-only profile executions verified with one stdout hash per app.
- Three allocation-only and three initialization-trace executions verified.
- Binary SHA-256 fingerprints match the preserved timing report.
- No source candidate was written, so numeric/array A/B controls were neither
  needed nor represented as evidence.
- Raw generated trees, executables, stdout captures, and profiles were removed
  after this aggregate record was written.

## Next recommendation

Perform a bounded feasibility gate for packed eager bytecode integer-cache
storage, then prototype it only if cached value identity and dynamic type
semantics can remain centralized.

Why: the eager cache is now an exact, reproducible 57-58 ms and 38 MB startup
wall in three unlike compiled applications and also affects every interpreter
process. The rejected lazy design changed GC pacing because it removed the
live cache. A packed eager representation could reduce hundreds of thousands
of per-element interface-box allocations while retaining real cache storage,
without laziness, a runtime-policy override, or fake ballast.

What it entails: inventory all raw-i32 and `IntegerValue` dynamic-type and
identity assumptions first; the raw carrier currently has about 63 references
and integer values have hundreds, so a representation change must use one
central extraction/materialization boundary rather than scattered type cases.
Measure inittrace bytes, objects, and wall time for a minimal raw-cache trial
before touching the larger boxed caches. Require focused identity/type/stack
tests, tree-walker and bytecode numeric fixtures, and repeated startup/hot-path
benchmarks. If feasible, gate preserved short applications alongside
allocation-heavy Binary Trees, allocation-light TapeLang, numeric Array/map,
and normal bytecode workloads. Revert on any broad wall-time regression; do
not use `GOGC` changes, dummy heap ballast, named-container rules, or WASM work.
