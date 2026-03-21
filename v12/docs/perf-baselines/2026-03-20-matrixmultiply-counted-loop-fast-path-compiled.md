# 2026-03-20 Counted-Loop Fast-Path Snapshot for `matrixmultiply`

Scope:
- capture the next shared primitive/control-flow lowering step on the matrix
  benchmark family after removing canonical loop-induction checked-arithmetic
  scaffolding
- cover both:
  - the reduced benchmark fixture
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  - the full canonical benchmark
    `v12/examples/benchmarks/matrixmultiply.able`
- keep the work within the compiler contract:
  - this is a shared primitive/control-flow lowering change
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
| `bench/matrixmultiply_f64_small` | `0.1133s` | `0.1133s` | `0.0100s` | `7.00` | canonical `loop { if i >= n { break } ... i = i + 1 }` induction now lowers to direct `for i < n { ... i++ }` |
| `examples/benchmarks/matrixmultiply` | `1.0833s` | `1.0767s` | `0.0433s` | `13.00` | full compiled benchmark now keeps matrix loop control off `__able_checked_add_signed(...)` in `matmul` and on direct counted loops in `build_matrix` / `matmul` |

Takeaway:
- loop-induction checked arithmetic was the next shared primitive/control-flow
  residual on the matrix family after the earlier built-in `Array` cleanup
- lowering the canonical counted-loop shape through one shared primitive rule
  reduces the reduced matrix target from `0.1967s` to `0.1133s`
- the same shared rule reduces the full canonical `matrixmultiply` benchmark
  from `3.4367s` to `1.0833s`
- the remaining matrix-family primitive residual is no longer loop control;
  it is affine checked integer arithmetic inside `build_matrix`, notably
  `i - j` and `i + j`
