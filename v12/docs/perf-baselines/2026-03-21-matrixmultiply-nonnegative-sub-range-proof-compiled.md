# 2026-03-21 Non-Negative Subtraction Range-Proof Snapshot for `matrixmultiply`

Scope:
- capture the next shared primitive-lowering step on the matrix benchmark
  family after inlining fixed-width checked integer `+` / `-`
- cover both:
  - the reduced benchmark fixture
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  - the full canonical benchmark
    `v12/examples/benchmarks/matrixmultiply.able`
- keep the work within the compiler contract:
  - this is a shared primitive range-proof change
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
| `bench/matrixmultiply_f64_small` | `0.1167s` | `0.1233s` | `0.0100s` | `7.00` | `build_matrix` now lowers `i - j` as direct signed subtraction when both loop operands are proven non-negative |
| `examples/benchmarks/matrixmultiply` | `1.1000s` | `1.1067s` | `0.0200s` | `13.00` | the full compiled benchmark keeps the subtraction proof too, but `i + j` still carries the widened checked-add branch |

Takeaway:
- subtraction-side inline overflow branching is now gone where shared static
  sign facts prove both operands non-negative
- on the matrix family, that cleanup remains effectively performance-neutral
  relative to the counted-loop baseline
- the remaining primitive residual is now stronger upper-bound range proofs
  for affine addition like `i + j`
