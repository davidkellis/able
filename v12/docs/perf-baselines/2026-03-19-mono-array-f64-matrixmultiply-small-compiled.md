# 2026-03-19 Mono-Array `f64` Matrix Multiply Compiled Snapshot

Scope:
- compare compiled-mode output with staged mono-array `f64` specialization
  enabled versus disabled
- target the reduced checked-in nested `Array (Array f64)` benchmark fixture:
  `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
- keep the full `v12/examples/benchmarks/matrixmultiply.able` example as a
  separate parity check only; at the current `60s` harness budget it still
  times out in both modes

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
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
```

Results:

| benchmark | mono on real | mono on gc | mono off real | mono off gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/matrixmultiply_f64_small` | `5.4833s` | `280.00` | `45.3133s` | `3568.67` | staged `f64` specialization + nested get/push fix materially improve both wall-clock and GC pressure |

Parity note:
- `./v12/bench_perf --runs 1 --timeout 60 --modes compiled v12/examples/benchmarks/matrixmultiply.able`
  now times out with mono arrays enabled instead of aborting early with
  `runtime: runtime error`
- `./v12/bench_perf --runs 1 --timeout 60 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/examples/benchmarks/matrixmultiply.able`
  also times out, matching the historical compiled baseline for the full
  `1000x1000` example

Takeaway:
- the staged `f64` mono-array slice is now producing a meaningful performance
  win once the benchmark directly exercises the specialized nested array path
- the next performance category should focus on broader nested/container
  carrier reduction beyond the current scalar-specialized inner-row strategy
