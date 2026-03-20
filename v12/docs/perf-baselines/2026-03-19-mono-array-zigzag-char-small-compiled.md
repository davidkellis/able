# 2026-03-19 Mono-Array `char` Zigzag Compiled Snapshot

Scope:
- compare compiled-mode output with staged mono-array text specialization
  enabled versus disabled
- target the reduced checked-in char-heavy benchmark fixture:
  `v12/fixtures/bench/zigzag_char_small/main.able`
- measure the first focused `Array char` / `Array (Array char)` workload after
  landing compiler-owned `[]rune` / nested text-row wrappers
- record the corrected mono-off baseline after fixing nested row identity for
  outer carrier arrays when staged scalar mono arrays are disabled

Environment:
- date: `2026-03-19`
- commit: `feebd884`
- tree: `dirty`
- host: `Linux 6.17.0-19-generic x86_64 GNU/Linux`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- mono-off comparison uses `--compiled-build-arg=--no-experimental-mono-arrays`

Commands:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/zigzag_char_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/zigzag_char_small/main.able
```

Results:

| benchmark | mono on real | mono on gc | mono off real | mono off gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/zigzag_char_small` | `0.9567s` | `88.00` | `1.0500s` | `384.00` | corrected mono-off baseline after preserving nested `Array (Array char)` row identity; both binaries print `8192000` |

Takeaway:
- the `Array char` / nested `Array (Array char)` widening is architecturally
  landed and `!Array char` propagation now compiles correctly on static paths
- the earlier mono-off comparison for this benchmark was invalid because the
  mono-off compiled path was boxing nested rows through `runtime.Value` and
  losing row mutation identity
- with that corrected, the staged text specialization is modestly faster on
  wall clock and dramatically lower on GC count for this workload
