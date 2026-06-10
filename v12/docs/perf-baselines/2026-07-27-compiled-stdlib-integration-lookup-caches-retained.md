# Compiled stdlib integration lookup-cache tranche

Date: 2026-07-27

## Decision

Retain:

- feature-package sharding for the six compiled stdlib integration rows that
  exceeded one minute;
- a read-only package export-name-set cache invalidated when static imports or
  source re-exports grow;
- a unique struct-name lookup cache invalidated when the collected struct
  table changes size;
- focused zero-allocation and invalidation/ambiguity regression tests for both
  caches.

The test split preserves every prior expected-output assertion. The compiler
caches are general lookup bookkeeping: they do not name a collection,
benchmark, or non-primitive nominal type, and they do not change lowering,
generated program behavior, runtime carriers, or dynamic boundaries.

No canonical stdlib implementation, runtime, interpreter, bytecode VM,
language, benchmark, dependency, or WASM change was required.

## Complete compiled-lane census

The initial warm, diagnostics-off census ran all 26 then-current compiled
stdlib cases in one Go test process:

```text
ABLE_RUN_COMPILED_CLI_INTEGRATION=1
ABLE_COMPILER_DIAGNOSTICS=0
TMPDIR=/var/tmp
GOCACHE=/var/tmp/able-go-cache
go test ./cmd/able -run '^TestTestCommandCompiledRunsStdlib' \
  -count=1 -timeout=45m -v
```

The lane passed in 1,182.728 seconds. Six rows exceeded one minute:

| Original row | Time |
|---|---:|
| HashMap smoke / HashSet | 99.35 s |
| List / Vector | 91.59 s |
| LinkedList / LazySeq | 73.26 s |
| PersistentSortedSet / PersistentQueue | 70.67 s |
| BigInt / BigUint | 67.79 s |
| PersistentMap / PersistentSet | 61.28 s |

The other 20 rows ranged from 12.93 to 54.51 seconds. TreeMap/TreeSet,
BitSet/Heap, concurrency, math, host-service, reporters, text, and the already
split foundational cases therefore stayed unchanged.

Each over-limit row was split at its existing feature-package boundary:

- BigInt and BigUint;
- List and Vector;
- PersistentMap and PersistentSet;
- PersistentSortedSet and PersistentQueue;
- LinkedList and LazySeq;
- HashMap smoke and HashSet.

The first isolated pass showed that sharding alone fixed the aggregate-only
violations, but four single-package graphs remained at or above the boundary:

| Isolated case | Time |
|---|---:|
| Vector | 79.50 s |
| HashSet | 95.58 s |
| LazySeq | 65.83 s |
| PersistentQueue | 60.00 s |

## Phase attribution

A temporary opt-in observer measured discovery, loading, Able compilation,
generated-output writing, Go build, and execution. It was removed before final
verification.

| Case | Total | Able compile | Go build | Execution |
|---|---:|---:|---:|---:|
| Vector | 79.35 s | 61.127 s | 17.562 s | 0.057 s |
| LazySeq | 64.51 s | 47.610 s | 16.242 s | 0.058 s |
| PersistentQueue | 60.45 s | 43.562 s | 16.175 s | 0.062 s |
| HashSet | 93.65 s | 76.555 s | 16.410 s | 0.064 s |

Execution was immaterial. A generated-Go build cache alone could not solve
Vector or HashSet because Able compilation already exceeded one minute.

Separate CPU/allocation profiles for Vector, LazySeq, and HashSet identified
two repeated lookup costs:

1. wildcard import/export checks rebuilt and sorted the same immutable package
   export-name sets on every query;
2. `structInfoByNameUnique` linearly scanned the entire collected struct table
   on repeated type and native-interface checks.

Before the first cache, package export-name discovery reached 61.99% cumulative
sampled allocation in HashSet and 43.67% in LazySeq. After that cache removed
the owner, unique struct-name scans still consumed 20–27% cumulative CPU in
Vector, PersistentQueue, and HashSet.

## Retained compiler rules

### Package export-name sets

`importableNameSet` now returns a cached read-only set for each package.
The cache shares the existing import-resolution lifecycle and is cleared
whenever a static import or source re-export binding grows.

### Unique struct names

`structInfoByNameUnique` now caches successful, missing, and ambiguous results.
The cache records the current collected struct-table size and clears itself
when that size changes, covering built-in and late test additions without
requiring a nominal-name special case.

Focused tests prove:

- cached package export sets allocate zero times on repeated reads;
- a new re-export invalidates a cached empty export set;
- cached unique struct lookup allocates zero times on repeated reads;
- adding one struct invalidates a miss;
- adding a second same-named struct invalidates a unique hit and preserves the
  ambiguity result.

## Repeated retained measurements

Three independent final processes per governing case ran with the temporary
observer removed:

| Case | Runs | Mean test time | Worst | Mean process wall | Mean peak RSS |
|---|---|---:|---:|---:|---:|
| Vector | 54.71, 55.10, 55.33 s | 55.047 s | 55.33 s | 56.333 s | 3,186,384 KB |
| LazySeq | 39.00, 39.40, 40.06 s | 39.487 s | 40.06 s | 40.563 s | 2,973,980 KB |
| PersistentQueue | 41.79, 41.45, 42.47 s | 41.903 s | 42.47 s | 43.013 s | 2,933,904 KB |
| HashSet | 45.40, 43.56, 44.73 s | 44.563 s | 45.40 s | 45.647 s | 3,093,763 KB |

Relative to the same-case phased baselines, retained mean test times changed:

| Case | Baseline | Retained mean | Change |
|---|---:|---:|---:|
| Vector | 79.35 s | 55.047 s | -30.63% |
| LazySeq | 64.51 s | 39.487 s | -38.79% |
| PersistentQueue | 60.45 s | 41.903 s | -30.68% |
| HashSet | 93.65 s | 44.563 s | -52.42% |

The final 32-case compiled stdlib lane passed in 1,038.379 seconds with a
55.70-second worst row. This is 12.20% below the original 1,182.728-second
26-case census despite six additional independent compile/build processes.

## Verification

Passed:

- cache zero-allocation, invalidation, and ambiguity tests;
- generic Enumerable/Iterator/filter-map guards;
- generic interface boundary and Result/Option specialization guards;
- broad imported/shadowed alias guards;
- `go test ./cmd/ablec`;
- all 32 strict compiled stdlib integration cases with every expected output;
- `./run_all_tests.sh` in 633.96 seconds;
- the canonical external stdlib in tree-walker mode in 17 seconds and bytecode
  mode in 15 seconds.

The full handoff retained fallback-free compiled behavior and changed no
generated execution rule.

## Next recommendation

Keep production application-performance mutation paused and evaluate a
content-addressed generated-output/build cache for the now-bounded compiled
stdlib lane.

Why: every individual case is below one minute, but the complete cold lane
still takes 1,038 seconds and remains opt-in rather than routine handoff
coverage.

What it entails: define a cache key from all Able source/dependency
fingerprints, compiler options, runner configuration, and Go/toolchain
identity; keep artifacts on disk; prove cold-miss correctness, repeated warm
hits, and invalidation after source/configuration changes across at least
three unlike cases. Do not move a slow first build into hidden setup, weaken
assertions, or change production lowering.

Why it is important: practical full compiled-CLI verification makes future
native-carrier and boundary work safer without reopening closed runtime
optimization routes. Generated execution ownership did not change in this
tranche, so the frozen application runtime profiles remain authoritative.
Do not begin WASM work.
