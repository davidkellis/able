# 2026-03-20 Shared Generic Nominal `methods` Specialization `Heap i32` Snapshot

Scope:
- capture the first reduced compiled-only benchmark after landing shared
  specialization for generic nominal `methods` blocks on statically known
  concrete targets
- target the checked-in `Heap i32` benchmark fixture:
  `v12/fixtures/bench/heap_i32_small/main.able`
- exercise the shared nominal-method specialization rule on a hot constrained
  generic method family without adding another named-structure lowering rule

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
- comparison baseline for this same reduced fixture remains
  `v12/docs/perf-baselines/2026-03-19-heap-i32-small-compiled.md`

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/heap_i32_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/heap_i32_small` | `4.2000s` | `10.3867s` | `0.6033s` | `1811.67` | shared generic nominal-method specialization landed; hot `Heap<T>` method calls now specialize on statically known targets |

Takeaway:
- this tranche materially improves wall-clock on the reduced `Heap i32`
  benchmark relative to the earlier `7.7533s` baseline
- GC count increased on this workload, which indicates the next shared AOT
  target is bound generic field/member carrier refinement inside specialized
  nominal method bodies rather than more call-signature specialization
