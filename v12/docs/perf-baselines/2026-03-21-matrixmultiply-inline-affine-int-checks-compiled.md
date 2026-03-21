# 2026-03-21 Inline Affine Integer Checks Snapshot for `matrixmultiply`

Scope:
- capture the next shared primitive-lowering step on the matrix benchmark
  family after removing helper-call overhead from fixed-width checked integer
  `+` / `-`
- cover both:
  - the reduced benchmark fixture
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  - the full canonical benchmark
    `v12/examples/benchmarks/matrixmultiply.able`
- keep the work within the compiler contract:
  - this is a shared primitive-lowering change
  - no named non-primitive structure rule was added

Environment:
- date: `2026-03-21`
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
| `bench/matrixmultiply_f64_small` | `0.1133s` | `0.1133s` | `0.0100s` | `7.00` | `build_matrix` affine `i - j` / `i + j` now lower as inline `int64(...) +/- int64(...)` plus range checks instead of helper calls |
| `examples/benchmarks/matrixmultiply` | `1.0867s` | `1.0900s` | `0.0333s` | `13.00` | full compiled benchmark keeps the affine integer ops inline too, but remains in the same performance band as the counted-loop snapshot |

Takeaway:
- helper-call overhead for fixed-width checked integer `+` / `-` is now gone
  on static compiled paths
- on the matrix family, that cleanup is effectively performance-neutral after
  the earlier counted-loop win
- the remaining primitive residual is now the inline overflow branches
  themselves where shared static range proofs can show they are unnecessary
