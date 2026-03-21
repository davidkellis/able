# 2026-03-20 Shared Native Float-to-Int Cast Snapshot for `matrixmultiply_f64_small`

Scope:
- capture the next shared primitive-lowering step on the reduced compiled-only
  matrix benchmark after removing the remaining `float -> int` runtime cast
  crossings from the full benchmark entry path
- target the checked-in benchmark fixture:
  `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
- keep the work within the compiler contract:
  - only built-in `Array` semantics and primitive numeric casts receive
    primitive-specific lowering
  - no named non-primitive structure rule was added

Environment:
- date: `2026-03-20`
- tree: `dirty`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity check confirms the compiled binary prints
  `-28.500833332098754`
- full macro smoke uses the same compiled-mode harness on
  `v12/examples/benchmarks/matrixmultiply.able`

Commands:
```bash
./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able

./v12/bench_perf --runs 1 --timeout 60 --modes compiled \
  v12/examples/benchmarks/matrixmultiply.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/matrixmultiply_f64_small` | `1.7567s` | `1.7600s` | `0.0100s` | `7.00` | shared primitive `float -> int` casts now stay native on the matrix entry path |
| `examples/benchmarks/matrixmultiply` | `timeout` | `n/a` | `n/a` | `n/a` | full macro benchmark still exceeds the current `60s` harness budget |

Takeaway:
- the reduced matrix benchmark improved again after replacing the remaining
  `__able_cast(...)` / `bridge.AsInt(...)` entry-path casts with native
  truncate/range-check/overflow lowering
- the generated full benchmark `main` body no longer depends on runtime cast
  helpers for `(n / 2) as i32` style primitive conversions
- the next category remains the broader macro-scale built-in `Array` work on
  the full matrix path beyond these primitive entry casts
