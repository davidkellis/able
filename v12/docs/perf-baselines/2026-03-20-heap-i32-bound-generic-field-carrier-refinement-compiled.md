# 2026-03-20 Bound Generic Field/Member Carrier Refinement `Heap i32` Snapshot

Scope:
- capture the reduced compiled-only benchmark after refining bound generic
  field/member carriers inside already-specialized nominal method bodies
- target the checked-in `Heap i32` benchmark fixture:
  `v12/fixtures/bench/heap_i32_small/main.able`
- prove the shared generic nominal lowering rule now keeps fully bound generic
  fields like `self.data: Array T` on their concrete native carriers without
  adding another named-structure lowering rule

Environment:
- date: `2026-03-20`
- tree: `dirty`
- host: `Linux 6.17.0-19-generic x86_64 GNU/Linux`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity check confirms the compiled binary prints
  `-211812354`
- comparison baselines for this same reduced fixture remain:
  - `v12/docs/perf-baselines/2026-03-19-heap-i32-small-compiled.md`
  - `v12/docs/perf-baselines/2026-03-20-heap-i32-generic-nominal-method-specialization-compiled.md`

Commands:
```bash
./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output \
  v12/fixtures/bench/heap_i32_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/heap_i32_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/heap_i32_small` | `0.7667s` | `0.9300s` | `0.0400s` | `91.33` | bound generic field/member carriers now stay native inside specialized nominal method bodies |

Takeaway:
- this tranche materially improves both wall-clock and GC count on the reduced
  `Heap i32` benchmark relative to the earlier `4.2000s` / `1811.67` snapshot
- the next category is no longer bound generic field/member carrier refinement;
  it is the next benchmark-worthy generic container/runtime edge that still
  crosses residual runtime carriers
