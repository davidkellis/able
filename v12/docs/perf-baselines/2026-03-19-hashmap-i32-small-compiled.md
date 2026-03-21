# 2026-03-19 Native `HashMap` Compiled Snapshot

Scope:
- capture a reduced compiled-only baseline for the first broader native
  container carrier slice beyond arrays
- target the checked-in `HashMap` / `Map` benchmark fixture:
  `v12/fixtures/bench/hashmap_i32_small/main.able`
- exercise both direct native `HashMap i32 i32` lowering and native
  `Map i32 i32` interface dispatch in the same compiled workload

Environment:
- date: `2026-03-19`
- commit: `c20aaa22`
- tree: `dirty`
- host: `Linux 6.17.0-19-generic x86_64 GNU/Linux`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity check confirms the compiled binary prints `4498503`

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/hashmap_i32_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/hashmap_i32_small` | `1.7633s` | `2.6933s` | `0.1333s` | `175.33` | reduced baseline for native `HashMap` carrier lowering plus native `Map` interface dispatch |

Takeaway:
- the first broader native container carrier slice is now represented by a
  checked-in reduced benchmark target
- this snapshot is a baseline for future `HashMap` / `HashSet` / map-interface
  tuning work, not a flag-on/flag-off comparison
