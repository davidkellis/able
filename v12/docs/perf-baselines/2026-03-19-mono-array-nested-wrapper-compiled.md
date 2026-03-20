# 2026-03-19 — Nested Mono-Array Outer Wrapper Compiled Snapshot

Target:

- `v12/fixtures/bench/matrixmultiply_f64_small/main.able`

Commands:

```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
```

Results:

- mono on: `5.7233s` real, `7.2600s` user, `0.1233s` sys, `252.00` GC
- mono off: `44.5167s` real, `137.8733s` user, `2.2033s` sys, `3550.67` GC

Interpretation:

- The outer `Array (Array f64)` shell is now compiler-owned too
  (`*__able_array_array_f64` over `[]*__able_array_f64`), so static compiled
  paths no longer box specialized rows through `runtime.Value` in the outer
  container.
- This structural cleanup preserves the large mono-on win over mono-off on the
  reduced matrix benchmark, but it does not add a second visible wall-clock
  improvement over the earlier `f64` slice by itself.
