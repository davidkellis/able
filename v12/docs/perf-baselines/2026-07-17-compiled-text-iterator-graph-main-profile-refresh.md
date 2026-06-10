# Compiled text, iterator, and graph main-profile refresh

Date: 2026-07-17

## Decision

Complete the compiled Word Frequency, Document Audit, Lexical Rollup, and
Dependency Plan profile gate and retain no compiler, generated-runtime,
bytecode VM, canonical-stdlib, benchmark, fixture, or language change.

Fresh verifier-backed generated-main CPU profiles and exact allocation
counters divide below generic Go runtime parents into four different semantic
owners. No exact compiler-owned operation is material in three unlike
applications. This gate therefore does not add a String, HashMap, iterator,
Queue/Deque, graph, key-type, or named-container lowering rule. WASM remains
deferred.

## Verified comparison contract

The current compiler built all four executables before any selection timing.
Each executable then received five independent one-CPU launches from the
catalog working directory with the catalog arguments and public Ruby verifier.
Five fresh Go-reference processes used the same one-CPU contract. Every
outlier remains in the arithmetic mean.

| Application | Able samples (s) | Able mean | Fresh Go mean | Able / Go |
| --- | --- | ---: | ---: | ---: |
| Word Frequency | 0.25, 0.19, 0.20, 0.19, 0.20 | 0.2060 s | 0.0056 s | 36.79x |
| Document Audit | 0.07, 0.07, 0.07, 0.07, 0.07 | 0.0700 s | 0.0040 s | 17.50x |
| Lexical Rollup | 0.08, 0.08, 0.08, 0.09, 0.08 | 0.0820 s | 0.0041 s | 20.00x |
| Dependency Plan | 0.07, 0.07, 0.07, 0.07, 0.07 | 0.0700 s | 0.0037 s | 18.92x |

All twenty Able and twenty Go processes verified. The Able and Go stdout
hashes agree within each application. The current Able sources and all input
and verifier fingerprints match the checked-in scorecard contracts.

## Generated-main CPU profiles

CPU-only phase profiling used the exact timed Able executables with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and CPU 0. Every launch was a
separate process and passed its public verifier. Word Frequency supplied ample
samples after 30 launches; the three shorter mains were expanded to 100
launches each. Only `main.cpu.pprof` was merged, excluding launcher/bootstrap
work.

| Application | Profiles | Main samples | Material concrete owners |
| --- | ---: | ---: | --- |
| Word Frequency | 30 | 4.03 s | generated HashMap find 45.66% flat/46.65% cumulative; `String.split` 38.71% cumulative; allocation 26.80% cumulative; byte/String conversion |
| Document Audit | 100 | 90 ms | `String.contains` 44.44% cumulative; generator/filter 44.44%; fast line reading 22.22% |
| Lexical Rollup | 100 | 800 ms | generator 56.25% cumulative; lazy map 36.25%; channel send 22.50%; iterator next 18.75%; `String.contains` 8.75% |
| Dependency Plan | 100 | 60 ms | deployment resolver 83.33% cumulative; Queue/Deque dequeue 33.33%; allocation 50.00%; common `i32` boxing 16.67% |

Document Audit and Dependency Plan remain close to CPU-profiler sampling
resolution even after 100 independent launches, so their percentages are
coarse. Their concrete stacks and exact allocation counters nevertheless agree
with the earlier preserved-binary profiles. They are sufficient to reject a
three-application shared descendant, not to rank small changes within either
short main.

## Exact main allocation counters

Separate allocation-only phase processes avoided CPU-profiler distortion.
The phase counters below are authoritative; start/end allocation-profile
serialization itself dominates the `.pprof` difference in the shorter
applications and is excluded from the interpretation.

| Application | Main bytes | Main allocations | Application-owned shape |
| --- | ---: | ---: | --- |
| Word Frequency | 31,184,888 | 720,431 | String/byte conversion, formatting, split, and map values |
| Document Audit | 374,200 | 1,968 | file-line materialization and iterator union values |
| Lexical Rollup | 2,392,960 | 30,980 | iterator union values, map/filter transport, String and integer conversion |
| Dependency Plan | 475,192 | 18,631 | common integer boxes, Queue/Deque elements, and graph lists |

The allocation owners do not repeat across three programs. Generic
`mallocgc` and collector span scanning are consequences of these different
owners, not a legal compiler or stdlib candidate.

## Generality reconciliation

- Hash probing and String splitting dominate only Word Frequency.
- Document Audit and Lexical Rollup share a lazy iterator pipeline, but the
  portable coverage record intentionally has only those two natural default-
  iterator consumers; the three-program admission rule is not permission to
  manufacture a third benchmark.
- `String.contains` is material in Document Audit and present in Lexical
  Rollup, but this four-program cohort does not contain a third material
  consumer.
- Dependency Plan's resolver, Queue/Deque operations, and integer boxes are
  absent or immaterial in the other three applications.
- Go collector helpers appear under different Able callers and cannot justify
  workload-specific GC policy, heap ballast, or container lowering.

No candidate advanced to an A/B gate. Canonical `../able-stdlib` required no
change.

## Verification and cleanup

- 8/8 initial build/verifier smoke processes passed.
- 20/20 repeated Able timing processes passed.
- 20/20 fresh Go-reference timing processes passed.
- 330/330 CPU-only phase-profile processes passed.
- 4/4 allocation-only phase-profile processes passed.
- `go test ./pkg/profilehook -count=1 -timeout 60s` passed.
- Raw generated trees, executables, stdout captures, and profiles were removed
  after this aggregate record was written.

## Next recommendation

Run a bounded validated-String provenance feasibility gate across I-Before-E,
Document Audit, and Lexical Rollup before implementing a compiler change.

Why: these are three existing verifier-backed applications with unlike outer
algorithms—direct file/text filtering, a document lazy pipeline, and a bounded
lexical lazy pipeline—but all naturally call the same compiled
`String.contains` implementation. Current profiles put that operation at
44.44% cumulative in Document Audit and 8.75% in Lexical Rollup; the latest
I-Before-E profile attributed 23.5% to it and another material share to UTF-8
validation. The generated fast path currently calls `utf8.ValidString` on
both operands before `strings.Contains`. A central proof that a String came
from a validated literal, filesystem decoder, checked constructor, or
StringBuilder could remove repeated validation across many String operations
without recognizing any application or named container.

What it entails: first refresh I-Before-E with the same preserved-binary
CPU/allocation contract and separate actual substring-search cost from UTF-8
validation and entry-wrapper cost in all three programs. Inventory every
String producer and invalid-string escape, especially
`String.from_bytes_unchecked`, and admit no candidate unless validity can be
represented by one conservative compiler fact that survives calls and merges.
If admitted, add invalid-UTF-8 fallback tests and repeated order-balanced A/B
cohorts for all three applications, then guard Word Frequency split,
Reverse Complement byte handling, regex, and bytecode/tree-walker String
parity. Revert if the shared concrete leaf is only the host substring search,
the proof becomes source-shape-specific, or any broad control regresses.
