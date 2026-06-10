# Compiled mixed generated-main profile refresh — 2026-07-13

## Decision

Keep no compiler, runtime, or canonical-stdlib change. The bounded profile set
does not identify a compiler or runtime operation that is material in three
independent application shapes. The large residuals are specific to String
conversion, persistent-HAMT conversion/hashing, or recursive-node allocation.
Changing one of those paths now would optimize a workload family rather than a
general Able program.

## Bounded profile set

All runs used compiled generated mains, the external canonical `able-stdlib`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, CPU 15, one run, and a 120-second
per-run timeout.

| Workload | Independent shape | Main allocated bytes | Main allocations | GCs |
| --- | --- | ---: | ---: | ---: |
| `run_length_encode_small` | String/UTF-8 traversal and building | 103,383,384 | 3,213,688 | 6 |
| `array_map_i32_small` | primitive numeric Array mapping | 2,576,104 | 112 | 0 |
| `hashmap_i32_small` | mutable nominal map | 3,004,312 | 11,634 | 0 |
| `persistent_map_i32_small` | persistent HAMT map | 56,833,344 | 898,914 | 4 |
| `binarytrees_small` | recursive nominal structs without a map | 13,738,448 | 680,350 | 1 |

The allocation attributions are disjoint:

- String work is led by `utf8_decode` (722,271 allocation objects), byte-array
  construction/lease bookkeeping, and character conversion.
- The numeric Array control has only 112 main allocations. Heap-profile
  writer frames dominate its allocation-difference file, so they are
  measurement work, not an application target.
- Mutable HashMap is small enough that profile-writer allocations and the
  one-time 4,225-object common-`i32` cache initialization obscure a useful
  residual. The already-kept scalar cache remains the appropriate general
  bridge improvement.
- PersistentMap is led by `bridge.ToUint` (172,853 objects), generated
  `HamtLeaf` conversion (112,182), hashing arithmetic, and the previously
  rejected map-only `structCacheKey` path (54,587 direct objects).
- BinaryTrees is 96.07% generated `make_tree` allocation (674,478 objects),
  rather than conversion or map machinery.

Thus no direct compiler/runtime symbol repeats materially across three shapes.

## CPU measurement guard

The first profile set requested CPU and exact allocation snapshots together.
Its CPU files were dominated by runtime traceback/profile-snapshot activity,
so they must not select an optimization. The independent CPU-only profile
mode (`ABLE_GO_PHASE_CPU_PROFILE_DIR`) then produced the following main-phase
samples without allocation snapshots:

| Workload | Measured duration | CPU samples | Dominant application path |
| --- | ---: | ---: | --- |
| `run_length_encode_small` | 446.59 ms | 450 ms | Go map operations and String `ArrayStoreEnsure`/character conversion |
| `persistent_map_i32_small` | 123.72 ms | 140 ms | generated `assoc_slot_spec`, allocation, hashing, and `ToUint` |
| `binarytrees_small` | 16.42 ms | 40 ms | generated `make_tree` allocation and GC assist |

`runtime.tryDeferToSpanScan` appears in all three CPU files, but its call trees
come from different String conversion, persistent-map hashing, and node
allocation/GC paths. It is allocation scanning, not a shared Able operation.
The BinaryTrees duration is also too short for stable CPU attribution. It
would be unsound to optimize that runtime frame directly or to infer a common
compiler lowering from it.

Artifacts are retained under
`v12/interpreters/go/.profiles/20260713_mixed_profile_refresh_*` and
`v12/interpreters/go/.profiles/20260713_mixed_cpu_only_*`.

## Why no candidate is justified

The profile gate requires one low-level compiler or runtime symbol to be
material in at least three distinct application shapes. This set instead
separates into String conversion, persistent-map conversion/hashing, and
recursive allocation, while the primitive numeric baseline is nearly
allocation-free. A generic rewrite of unions, cache keys, maps, or allocation
handling would be speculative and would violate the broad-program performance
rule.

No canonical `able-stdlib` change is needed.

## Verification

- All five bounded generated-main profile runs and all three clean CPU-only
  runs completed under the stated guards.
- No source behavior changed, so no new semantic test is required.
- `git diff --check` passes.

The repository-wide `./run_all_tests.sh` remains blocked before Go tests by
the existing untracked `exec/12_09_nested_spawn_native_context` fixture missing
from the already-modified exec coverage index. This tranche leaves that
fixture and index untouched.

## Next recommendation

Refresh the compiled-versus-Go external scorecard across the current
feature-rich benchmark suite, then profile only the lanes that miss the 95%
target. The mixed internal profiles do not expose a broad residual by
themselves; the scorecard will identify which real application categories are
still meaningfully behind their Go equivalents. Use the same pinned CPU and
memory guards, preserve stdout hashes, and use CPU-only phase profiles for
the selected misses. That keeps the next optimization driven by an observable
product gap rather than by a noisy internal frame.
