# 2026-03-19 Native `Heap i32` Compiled Snapshot

Scope:
- capture a reduced compiled-only baseline for the broader stdlib container
  family audit tranche beyond nominal map/set carriers
- target the checked-in `Heap i32` benchmark fixture:
  `v12/fixtures/bench/heap_i32_small/main.able`
- exercise native `Heap i32` carrier lowering plus hot `push` / `pop` /
  `len` behavior on the compiled path

Environment:
- date: `2026-03-19`
- commit: `c20aaa22`
- tree: `dirty`
- host: `Linux 6.17.0-19-generic x86_64 GNU/Linux`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity check confirms the compiled binary prints
  `-211812354`
- the first draft of this fixture was larger and averaged `53.0167s` per
  compiled run, so it was trimmed before being recorded as the checked-in
  reduced target

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/heap_i32_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/heap_i32_small` | `7.7533s` | `16.5467s` | `0.5967s` | `1105.00` | reduced baseline for native `Heap i32` carrier lowering and broader array-backed container audit coverage |

Takeaway:
- the broader stdlib container families in this slice now have explicit
  no-fallback audit coverage on the shared native carrier path
- this snapshot is a checked-in baseline for future `Heap` / `Deque` /
  `Queue` / `BitSet` / `PersistentSortedSet` / `PersistentQueue` tuning work
