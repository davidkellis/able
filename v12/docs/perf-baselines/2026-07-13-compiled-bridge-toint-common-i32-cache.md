# Compiled common `i32` bridge cache — 2026-07-13

## Decision

Keep a compiler-bridge cache for immutable common `i32` runtime values. The
shared `bridge.ToInt` conversion now returns a cached `runtime.IntegerValue`
for canonical `i32` values from -128 through 4095. Other integer suffixes and
out-of-range values retain the previous `runtime.NewSmallInt` path exactly.

The cache is a primitive scalar conversion rule at the common compiler bridge;
it does not identify `HashMap`, `TreeMap`, `PersistentMap`, or any other
nominal type. Cached `IntegerValue`s are stored by value in a `runtime.Value`
interface, so a caller receives a value copy rather than mutable shared state.
`sync.Once` makes initialization safe for concurrent compiled applications.

## Shared evidence

Fresh compiled generated-main profiles used the external canonical
`able-stdlib`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The controls
cover three independent map representations and a primitive numeric Array
control.

Before the cache, `bridge.ToInt` attributed 9,017 allocation objects in
HashMap, 107,539 in TreeMap, 74,789 in PersistentMap, and one in Array-map.
The candidate reduces the direct bridge attribution to 4,008, one, 5,947, and
one respectively. Each map process pays the one-time 4,225-object cache
initialization; the phase totals below include that cost.

| Workload | Baseline bytes / allocations / GCs | Kept bytes / allocations / GCs |
| --- | ---: | ---: |
| `hashmap_i32_small` | 2,930,952 / 12,253 / 0 | 3,002,504 / 11,615 / 0 |
| `tree_map_i32_small` | 36,164,280 / 353,762 / 3 | 31,250,000 / 250,243 / 2 |
| `persistent_map_i32_small` | 59,837,320 / 963,579 / 5 | 57,112,024 / 899,869 / 4 |
| `array_map_i32_small` | 2,576,104 / 112 / 0 | 2,576,104 / 112 / 0 |

Compact baseline and candidate artifacts are retained in
`v12/interpreters/go/.profiles/` with the prefixes
`20260713_bridge_toint_` and `20260713_bridge_toint_native_`.

## Timing gate

Each lane was compiled and run nine times with
`bench_perf --cpu-affinity 15`, the same external stdlib, and the same
memory/GC guardrails. Every baseline/candidate pair has the same stdout
SHA-256 hash.

| Workload | Baseline | Kept cache | Change |
| --- | ---: | ---: | ---: |
| `hashmap_i32_small` | 0.1078 s | 0.0889 s | 17.5% faster |
| `tree_map_i32_small` | 0.1456 s | 0.1244 s | 14.6% faster |
| `persistent_map_i32_small` | 0.1967 s | 0.1967 s | neutral |
| `array_map_i32_small` | 0.0844 s | 0.0822 s | 2.6% faster |

The neutral persistent and non-map control results are important: the cache
does not turn a single collection benchmark into a global fast path. The
larger TreeMap and PersistentMap allocation reductions demonstrate that the
same primitive bridge conversion benefits more than one nominal translation
pipeline.

## Verification

- New bridge tests cover cached boundary values, fallback behavior for
  out-of-range/other-suffix inputs, and zero allocation for a cached value.
- `go test -v ./pkg/compiler/bridge -count=1 -timeout 60s` passes.
- Focused compiler map carrier/execution tests pass for HashMap, TreeMap, and
  PersistentMap. The same invocation still has two existing generated-source
  shape assertion failures: `TestCompilerHashMapStaticCarrierStaysNative`
  expects the pre-context method name, and
  `TestCompilerTreeMapBenchmarkShapePreservesNativeCarriers` rejects a
  runtime-value match branch emitted by the already-dirty generator work.
  Neither assertion is affected by this bridge-only cache.
- `git diff --check` passes.

The repository-wide `./run_all_tests.sh` remains blocked before Go tests by
the existing untracked `exec/12_09_nested_spawn_native_context` fixture missing
from the already-modified exec coverage index. This tranche leaves that
fixture and index untouched. No canonical-stdlib source change is needed.

## Next recommendation

Profile `bridge.structCacheKey` across TreeMap, PersistentMap, and a
non-map nominal-struct workload. It is the next repeated residual conversion
cost in both map representations, but it uses `fmt.Sprintf` to form a cache
key and must be demonstrated outside maps before considering a general
environment-and-name key representation change.
