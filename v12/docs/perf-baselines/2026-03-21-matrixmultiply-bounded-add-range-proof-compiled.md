# 2026-03-21 Bounded Addition Range-Proof Snapshot for `matrixmultiply`

Scope:
- capture the next shared primitive-lowering step on the matrix benchmark
  family after removing the subtraction-side widened overflow branch
- cover both:
  - the reduced benchmark fixture
    `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
  - the full canonical benchmark
    `v12/examples/benchmarks/matrixmultiply.able`
- keep the work within the compiler contract:
  - this is a shared primitive/function range-proof change
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
| `bench/matrixmultiply_f64_small` | `0.1267s` | `0.1267s` | `0.0100s` | `7.00` | `build_matrix` now lowers both `i - j` and `i + j` as direct signed arithmetic in the inner loop |
| `examples/benchmarks/matrixmultiply` | `1.1367s` | `1.1367s` | `0.0300s` | `13.00` | the full compiled benchmark keeps the same affine proof, but remains in the same performance band as the earlier counted-loop snapshot |

Takeaway:
- the widened inline overflow branch is now gone for both affine integer ops
  in `build_matrix`
- the matrix inner-loop affine add/sub gap is now closed through shared
  primitive/function range proofs
- on this benchmark family, the result remains effectively performance-neutral
  relative to the counted-loop baseline, so the next worthwhile work is a
  different category than loop-affine primitive arithmetic
