# 2026-03-20 Shared Native Scalar Array-Propagation Snapshot for `matrixmultiply_f64_small`

Scope:
- capture the reduced compiled-only benchmark after removing the remaining
  built-in `Array` scalar propagation/runtime-cast crossings on the reduced
  matrix path
- target the checked-in benchmark fixture:
  `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
- prove this is still a shared-lowering step, not a named-structure rule:
  - static built-in `Array` propagation now returns concrete success element
    types on the compiled path
  - primitive numeric casts like `i32 -> f64` now lower directly to Go casts

Environment:
- date: `2026-03-20`
- tree: `dirty`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity check confirms the compiled binary prints
  `-28.500833332098754`
- comparison baselines for this same reduced fixture remain:
  - `v12/docs/perf-baselines/2026-03-19-mono-array-f64-matrixmultiply-small-compiled.md`
  - `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`

Commands:
```bash
./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able

./v12/bench_perf --runs 1 --timeout 60 --modes compiled \
  v12/examples/benchmarks/matrixmultiply.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/matrixmultiply_f64_small` | `1.9733s` | `1.9767s` | `0.0167s` | `7.00` | shared built-in `Array` scalar propagation and primitive cast lowering keep the inner numeric loop native |
| `examples/benchmarks/matrixmultiply` | `timeout` | `n/a` | `n/a` | `n/a` | full macro benchmark still exceeds the current `60s` harness budget |

Takeaway:
- this tranche materially improves both wall-clock and GC count on the reduced
  matrix benchmark versus the earlier `5.7233s` / `252.00` GC snapshot
- the reduced matrix path is no longer paying runtime-value conversion costs
  in the scalar multiply/add loop
- the next category is the remaining macro-scale built-in `Array` lowering on
  the full matrix path, using the same shared `Array` and primitive rules
