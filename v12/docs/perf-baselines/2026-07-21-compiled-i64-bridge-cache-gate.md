# Compiled `i64` Bridge Cache Gate

Date: 2026-07-21

## Decision

Retain no production change. A suffix-preserving common-value cache for
compiled `i64` bridge values removed a large shared allocation wall and made
K-Nucleotide and Inventory Reconciliation materially faster, but it made the
allocation-light TapeLang guard 4.17% slower over ten alternating verified
pairs. The candidate was fully removed.

This was a primitive compiler/interpreter boundary trial, not a HashMap,
Result, application, or other non-primitive nominal special case. No Able
source, canonical stdlib, verifier, reference implementation, benchmark
contract, VM, or language semantics changed.

## Selection evidence

Six unlike current compiled misses were built once and profiled under
`GOMEMLIMIT=1GiB`, `GOGC=50`, their catalog CPU/executor policy, and a
60-second process limit:

- K-Nucleotide: text, unsigned hashing, and dynamic map calls;
- Sudoku Masks: arrays and search;
- Inventory Reconciliation: records and dynamic map calls;
- Unicode Scalar Pipeline: text and iterators;
- Option/Result Config: unions and configuration flow; and
- Validated Job Pipeline: concurrency, unions, and generic nominal flow.

Exact measured-main allocation snapshots completed for five applications.
K-Nucleotide's exact snapshot exceeded the bound, so its ordinary cumulative
allocation profile was used instead; the normal verified program completed in
under three seconds.

The same primitive bridge leaf, `bridge.ToInt`, was material in three unlike
applications:

- K-Nucleotide boxed `i64` map handles alongside `u64` keys;
- Inventory boxed `i64` handles, keys, and values; and
- Validated Job boxed `i64` values through generic Result/operator paths.

At the inspected map boundary sites, generated `[]runtime.Value` argument
slices did not escape in Go's `-m=2` output. The concrete allocations were
interface boxes for `runtime.IntegerValue`; Validated Job independently
repeated the same leaf through a different generic boundary. This cleared the
three-application admission rule for extending the already-retained
common-`i32` primitive bridge cache to `i64`.

## Candidate

`bridge.ToInt` retained its existing common `i32` path and temporarily added a
separate lazy `i64` cache for `-128..4095`. Separate storage preserved the
observable integer suffix. Out-of-range values and all other integer kinds
used the existing `runtime.NewSmallInt` path. Unit tests covered value/suffix
preservation, range fallback, and zero steady-state allocations for cached
values.

For every A/B application, baseline and candidate `compiled.go` hashes were
identical. Only the copied bridge dependency differed.

## Allocation result

All profiled outputs verified.

| Application | Baseline main bytes / objects | Candidate main bytes / objects | Change |
| --- | ---: | ---: | ---: |
| Inventory Reconciliation | 68,744,928 / 1,630,307 | 17,116,608 / 553,185 | -75.1% bytes; -66.1% objects |
| Validated Job Pipeline | 75,552,272 / 1,983,279 | 60,202,856 / 1,662,029 | -20.3% bytes; -16.2% objects |
| K-Nucleotide, sampled cumulative | 30,782,043 objects | 21,445,235 objects | -30.3% objects |

K-Nucleotide's sampled `bridge.ToInt` objects fell from 11,786,095 to
3,331,565 (-71.7%). `bridge.ToUint` remained the largest candidate allocation
leaf at 7,340,367 objects.

## Repeated wall-time gate

Every process ran the external Ruby verifier. Pair order alternated. Short or
volatile rows received ten pairs; the two clear primary wins and Sudoku used
five. Reported values are arithmetic means, as required for the shared
workstation.

| Application | Pairs | Baseline | Candidate | Change | Mean GC baseline -> candidate |
| --- | ---: | ---: | ---: | ---: | ---: |
| K-Nucleotide | 5 | 3.558 s | 3.014 s | -15.3% | 59.0 -> 41.4 |
| Inventory Reconciliation | 5 | 0.280 s | 0.206 s | -26.4% | 7.8 -> 5.4 |
| Validated Job Pipeline | 10 | 2.949 s | 2.988 s | +1.3% | 11.0 -> 10.1 |
| Option/Result Config | 10 | 0.200 s | 0.195 s | -2.5% | 6.8 -> 7.0 |
| Unicode Scalar Pipeline | 10 | 0.231 s | 0.233 s | +0.9% | 6.2 -> 5.9 |
| Sudoku Masks | 5 | 1.820 s | 1.794 s | -1.4% | 14.6 -> 14.6 |
| TapeLang Alphabet | 10 | 3.690 s | 3.844 s | **+4.2%** | 4.1 -> 3.9 |

Validated Job split 5/10 pair wins and had nearly identical mean user/system
time, so it supplies no demonstrated wall-time win despite lower allocation.
TapeLang split 4/10 wins. Removing each side's minimum and maximum still left
the candidate about 3.1% slower.

An exact TapeLang allocation-only check found that both binaries allocated
exactly 282,552 bytes in 4,274 measured-main objects and performed zero
measured-main GCs. A subsequent clean CPU profile established that TapeLang's
hot main loop does not call `bridge.ToInt`; the movement therefore cannot be
attributed to an executed range/cache branch. The broad timing rejection
remains valid, but that earlier causal interpretation is corrected by
`2026-07-21-compiled-dynamic-i64-boundary-cache.md`. Source-identical Binary
Trees baseline/candidate binaries were prepared, but the selection gate
stopped at the earlier established-guard failure rather than spending another
long cohort.

## Restored state and next direction

The `i64` cache, its branch, and its tests were removed. The retained common
`i32` cache is unchanged.

Restored verification passed:

- `go test -race ./pkg/compiler/bridge -count=1 -timeout 60s`;
- focused compiler dynamic-boundary, integer, HashMap, and union controls in
  19.394 seconds; and
- `go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s` in
  25.828 seconds.

The next useful question is narrower than another global cache-size or range
trial: classify generated `i64` boxing sites by whether the resulting
`runtime.Value` actually escapes across a dynamic compiler/interpreter
boundary. If the same escaping boundary shape is material in at least three
unlike applications, a dedicated boundary helper could apply reuse only where
Go otherwise allocates, leaving TapeLang's non-escaping conversions on the
branch-free path. That work requires generated-site and escape-analysis
reconciliation first; it does not authorize HashMap, Result, benchmark, or
named-nominal special cases.
