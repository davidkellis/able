# Interpreter-free compiled runtime profile refresh

Date: 2026-07-24

## Decision

Retain no compiler, generated-runtime, bridge, canonical-stdlib, application,
language, dependency, or WASM change. Fresh CPU and allocation evidence across
three unlike strict interpreter-free applications contains no exact generated
code or Able runtime symbol shared by all three.

The only full intersection is Go allocator and collector machinery. Its callers
split into text/map boundary conversion, recursive Array/search allocation, and
already-Go-equivalent nominal tree allocation. That aggregate is not one
general Able optimization, and the closed GC-ballast route is not reopened.

## Applications and artifact contract

The governing applications were selected before profiling:

- K-Nucleotide: serial text decoding, primitive map hashing/equality, sorting,
  and integer boundary conversion;
- Sudoku Masks: serial recursive search, static Arrays, bit arithmetic, and
  signed quotient; and
- Binary Trees: four-CPU recursive nominal allocation and traversal.

Each application was compiled once with `ablec --no-fallbacks`. `go list
-deps` over each generated tree and `go tool nm` over each executable found no
`able/interpreter-go/pkg/interpreter` dependency or symbol. All profile,
allocation, and timing processes passed the public benchmark verifier with
stable output hashes.

The canonical stdlib source state remained the current 70-file tree
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
The companion JSON records source, generated Go, executable, reference, and
verifier hashes.

## Repeated timing against equivalent Go

K-Nucleotide and Sudoku used CPU 0, `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, and
`GOGC=50`. Binary Trees used CPUs 0-3, the required goroutine executor,
`GOMAXPROCS=4`, `GOMEMLIMIT=1GiB`, and normal Go GC as required by the
benchmark. Processes were order-balanced between Able and Go.

| Application | Able samples | Able mean | Go samples | Go mean | Able / Go |
| --- | --- | ---: | --- | ---: | ---: |
| K-Nucleotide | 3.02, 2.86, 2.57, 2.53, 3.20 s | 2.8360 s | 0.05, 0.08, 0.06, 0.05, 0.06 s | 0.0600 s | 47.2667x |
| Sudoku Masks | 2.21, 2.09, 2.32, 2.33, 2.12 s | 2.2140 s | 0.69, 0.71, 0.76, 0.68, 0.61 s | 0.6900 s | 3.2087x |
| Binary Trees | 8.23, 8.19, 8.09 s | 8.1700 s | 8.22, 8.80, 7.84 s | 8.2867 s | 0.9859x |

Binary Trees is 1.41% faster than equivalent Go in this fresh cohort. The two
other rows remain far outside the 1.0526x target ratio, but their exact hot
owners do not overlap each other or Binary Trees.

## Main-only CPU profiles

Three independent CPU-profile processes per application ran with
`GOGC=50`, `GOMEMLIMIT=1GiB`, the same CPU/executor policy, and a 55-second
cap. Only the registered Able `main` phase was merged.

| Application | Merged CPU samples | Leading exact owners |
| --- | ---: | --- |
| K-Nucleotide | 9.06 s | primitive map key equality 8.28% flat; primitive map hash 4.42%; generated map find 12.91% cumulative |
| Sudoku Masks | 6.75 s | generated `find_best_empty` 31.11% flat / 92.15% cumulative; checked multiply 12.89%; signed divmod 12.44% |
| Binary Trees | 130.62 s | Go span scanning 22.98% flat; `runtime.mallocgc` 71.21% cumulative; generated `make_tree` 77.78% cumulative |

The complete exact-symbol intersection contains 26 symbols. All 26 belong to
the Go allocator or collector; zero begin with generated `main.` or
`able/interpreter-go`. The largest common flat symbol is
`runtime.tryDeferToSpanScan` at 7.84%, 1.19%, and 22.98%.
`runtime.mallocgc` is cumulative beneath 38.74%, 11.70%, and 71.21%.

Those shared parents do not establish one Able owner:

- K-Nucleotide allocates String/byte conversion results and dynamic primitive
  boundary values below map operations;
- Sudoku allocates the source-required best-position Array inside its search;
  and
- Binary Trees allocates the same two-pointer node shape and recursive tree
  graph as its Go reference.

## Allocation profiles and exact counters

Each application received one sampled cumulative allocation profile for caller
attribution and three separate lightweight exact main-phase counter processes.
The counter means are authoritative.

| Application | Mean main bytes | Mean allocations | Mean GC under profile contract | Dominant sampled owner |
| --- | ---: | ---: | ---: | --- |
| K-Nucleotide | 1,216,971,963 | 27,733,835 | 196.33 | builtin String conversion 34.50%, `bridge.ToUint` 29.82%, `bridge.ToInt` 15.26% |
| Sudoku Masks | 156,390,397 | 7,802,771 | 132.67 | generated `find_best_empty`, 98.64% |
| Binary Trees | 9,820,590,973 | 613,771,245 | 205.67 | generated `make_tree`, 100% |

The exact counts were stable within each application. They reinforce rather
than weaken the CPU classification: no allocation definition, carrier, or
boundary repeats across all three.

## Why no candidate advanced

- K-Nucleotide's map/string/primitive boundary is material only in that row.
  Extending it from this evidence would be a named-container or one-family
  optimization.
- Sudoku's generated checked arithmetic and search result allocation are not
  present in the other two. The quotient route and non-improving-result
  construction rules already have dedicated correctness and breadth records.
- Binary Trees' node shape already matches Go. Pooling, ballast, or GC tuning
  would violate both the benchmark rules and the closed package-cut decision.
- The shared allocator/collector symbols are consequences of three different
  allocation graphs, not a generated-runtime mechanism.

The required three-unlike-program bar therefore fails before implementation.
No candidate A/B was warranted, and no production diff was created.

## Verification

- 3/3 strict fallback-free builds completed.
- 3/3 generated dependency graphs and executables omitted the interpreter
  package.
- 9/9 accepted CPU-profile processes verified.
- 9/9 allocation-counter processes verified.
- 3/3 sampled allocation-profile processes verified.
- 26/26 timed Able/Go processes verified.

One initial Binary Trees diagnostic cohort used the default serial executor.
It was rejected immediately and is absent from every table and decision above.
All accepted Binary Trees evidence uses `ABLE_EXECUTOR=goroutine`.

## Next recommendation

Return selection to the bytecode engine and refresh a three-unlike-application
semantic-owner gate from the current scorecard.

Why: this interpreter-free compiled refresh has no legal shared generated
owner, while bytecode still owns the large majority of the cross-engine target
excess. More local compiled work would choose among unrelated application
bodies or reopen a closed allocator/GC route.

What it entails: freeze one current bytecode artifact, collect repeated warmed
main-only CPU and allocation profiles plus matched Python/Ruby timings for
three unlike misses, and intersect exact VM-owned semantic leaves below
dispatcher/stack parents. Select only a new operation that is material in all
three and absent from the closure ledger; otherwise retain no code again.

Why it matters: it moves effort to the largest remaining target budget while
preserving the evidence standard that prevented a benchmark, named-container,
or non-primitive nominal shortcut here. WASM remains deferred.

## Evidence

- `2026-07-24-interpreter-free-compiled-runtime-profile-refresh.json`
