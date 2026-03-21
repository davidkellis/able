# 2026-03-20 Static Array Propagation Pointer-Elision Snapshot for `matrixmultiply`

Scope:
- capture the next shared built-in `Array` lowering step on the matrix
  benchmark family after removing pointer-backed nullable-carrier construction
  from propagated static array accessors
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
| `bench/matrixmultiply_f64_small` | `0.1967s` | `0.2000s` | `0.0100s` | `7.33` | reduced benchmark is effectively flat; propagated `get(...)!` pointer carriers are gone from the hot path |
| `examples/benchmarks/matrixmultiply` | `3.4367s` | `3.4400s` | `0.0400s` | `13.67` | full compiled benchmark improves materially once propagated static `Array` accessors stop constructing pointer-backed nullable carriers |

Takeaway:
- propagated static built-in `Array` accessors (`get`, `first`, `last`,
  `read_slot`, `pop`) were still manufacturing pointer-backed nullable
  carriers on the success path
- removing that shared success-path pointer construction is effectively neutral
  on the reduced matrix target but reduces the full canonical benchmark from
  `4.2267s` to `3.4367s`
- this closes the propagated static-array accessor pointer-carrier gap on the
  current matrix family
