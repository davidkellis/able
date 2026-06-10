# Compiled `structCacheKey` generality gate — 2026-07-13

## Decision

Keep no compiler, runtime, or canonical-stdlib change. `bridge.structCacheKey`
forms an environment-and-name cache key with `fmt.Sprintf`; it is a large
allocation source in the nominal conversion machinery used by TreeMap and
PersistentMap. It does not recur in the independent recursive-struct
application control, so changing the common bridge key representation would
be driven by a container translation shape rather than broad application use.

## Bounded profiles

All profiles use compiled generated mains, the external canonical
`able-stdlib`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.

| Workload | Shape | `structCacheKey` allocation objects |
| --- | --- | ---: |
| `tree_map_i32_small` | ordered mutable nominal map | 95,684 (33.55%) |
| `persistent_map_i32_small` | persistent HAMT nominal map | 109,182 (12.18%) |
| `binarytrees_small` | recursive `Node`/`DepthResult` structs, no map | 22 (0.0031%) |

`binarytrees_small` records 13,727,944 allocated main-phase bytes and 680,186
allocations, but only 22 are attributable to the candidate bridge key. The
TreeMap and PersistentMap stacks reach the key through generated struct-to-
runtime conversions; binary trees reach it only during a few setup/runtime
lookups. Thus the nominal map profiles do not establish a shared application
wall.

The new independent profile artifacts are retained with prefix
`20260713_struct_cache_key_binarytrees_small_compiled_`; the current TreeMap
and PersistentMap comparison artifacts are retained with prefix
`20260713_bridge_toint_native_`.

## Why no candidate is justified

Replacing `fmt.Sprintf` with a different bridge key format could improve the
two map implementations, but its lack of material cost in a non-map nominal
program fails the generality requirement. A change here would be too close to
an indirect container-specific optimization even though the helper itself has
a generic name. No timing candidate was run because the profile gate failed.

## Verification

- The bounded BinaryTrees compiled profile run completed successfully under
  the normal CPU, timeout, and memory guardrails.
- No source behavior changed, so no new semantic test is required.
- `git diff --check` passes.

The repository-wide `./run_all_tests.sh` remains blocked before Go tests by
the existing untracked `exec/12_09_nested_spawn_native_context` fixture missing
from the already-modified exec coverage index. This tranche leaves that
fixture and index untouched.

## Next recommendation

Refresh one bounded mixed generated-main profile set covering string,
primitive numeric Array, mutable map, persistent map, and recursive nominal
struct workloads. Choose a follow-up only if the same compiler/runtime symbol
is material in at least three distinct shapes; this resets the search away
from container residuals and toward improvements that can benefit ordinary
applications broadly.
