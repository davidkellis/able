# 2026-03-20 Static Array Frame-Elision Snapshot for `matrixmultiply`

Scope:
- capture the next shared built-in `Array` lowering step on the matrix
  benchmark family after removing synthetic call-frame scaffolding from static
  array factories and intrinsics
- cover both:
  - the reduced benchmark fixture
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  - the full canonical benchmark
    `v12/examples/benchmarks/matrixmultiply.able`
- keep the work within the compiler contract:
  - this is a shared built-in `Array` lowering change
  - no named non-primitive structure rule was added

Environment:
- date: `2026-03-20`
- tree: `dirty`
- runs:
  - reduced fixture: `3`
  - full benchmark: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity checks confirm:
  - reduced fixture output: `-28.500833332098754`
  - full benchmark output: `-95.58358333329998`

Commands:
```bash
./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able

./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output \
  v12/examples/benchmarks/matrixmultiply.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/examples/benchmarks/matrixmultiply.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/matrixmultiply_f64_small` | `0.1933s` | `0.1933s` | `0.0133s` | `7.00` | static `Array` factory/get/push frame churn removed from the reduced path |
| `examples/benchmarks/matrixmultiply` | `4.2267s` | `4.2300s` | `0.0367s` | `13.00` | full compiled benchmark is now comfortably under the `60s` harness budget |

Takeaway:
- synthetic `__able_push_call_frame(...)` / `__able_pop_call_frame()` scaffolding
  on static built-in `Array` operations was a major remaining cost on the
  matrix family
- once removed from the shared static `Array` path, the reduced matrix target
  dropped from `1.7567s` to `0.1933s`
- the full canonical `matrixmultiply` benchmark moved from timeout to a stable
  compiled average of `4.2267s`
- this closes the current macro-scale matrix built-in `Array` tranche
