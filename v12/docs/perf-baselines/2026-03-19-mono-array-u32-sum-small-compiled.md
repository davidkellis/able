# 2026-03-19 Mono-Array `u32` Sum Compiled Snapshot

Scope:
- compare compiled-mode output with staged mono-array unsigned specialization
  enabled versus disabled
- target the reduced checked-in `Array u32` benchmark fixture:
  `v12/fixtures/bench/sum_u32_small/main.able`
- measure the widened primitive numeric family on a workload dominated by typed
  push, index read, and index mutation

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
- direct `ablec` parity check confirms both binaries print `17999997000003`

Commands:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/sum_u32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/sum_u32_small/main.able
```

Results:

| benchmark | mono on real | mono on gc | mono off real | mono off gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/sum_u32_small` | `1.0933s` | `185.33` | `1.6800s` | `21.33` | staged unsigned specialization materially improves wall-clock and user time on typed `u32` array build/read/mutate paths |

Takeaway:
- the remaining primitive numeric scalar slice is now covered by staged
  compiler-owned array wrappers
- on this `u32` workload the specialized path is materially faster even though
  raw GC count is not lower, which indicates the generic fallback cost here is
  dominated more by boxing/runtime dispatch than by GC count alone
