# v12 Performance Benchmark Harness

`v12/bench_suite` runs the shared benchmark suite in three execution modes:

- `compiled`
- `treewalker`
- `bytecode`

It emits machine-readable JSON with:

- git commit + dirty state
- machine profile (OS/kernel/arch/CPU/memory)
- harness config (timeouts, runs, benchmark list)
- per-run status/timing/GC metrics
- per-benchmark summary rows

`v12/bench_perf` is the lighter per-target helper for focused perf checks. Its
compiled mode now builds through `cmd/ablec` directly, so compiled fixture
benchmarking measures the current compiler path without pulling in unrelated
`able build` package/bootstrap behavior. It also accepts repeated
`--compiled-build-arg` flags for controlled comparisons such as
`--no-experimental-mono-arrays`.

For targeted compiler-lowering checks, prefer checked-in fixture targets under
`v12/fixtures/bench/` so the benchmark package metadata is reproducible from
the repo. Recent mono-array work uses
`v12/fixtures/bench/matrixmultiply_f64_small/main.able` for the staged nested
`Array (Array f64)` comparison and
`v12/fixtures/bench/zigzag_char_small/main.able` for the staged text-scalar
(`Array char` / `Array (Array char)`) comparison and
`v12/fixtures/bench/sum_u32_small/main.able` for the staged unsigned numeric
comparison, while the full
`matrixmultiply` workload in `v12/examples/benchmarks/matrixmultiply.able`
remains the canonical suite entry used by `v12/bench_suite`. Current focused
snapshots for these reduced fixtures are checked in at:

- `v12/docs/perf-baselines/2026-03-19-mono-array-f64-matrixmultiply-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-u32-sum-small-compiled.md`

The `zigzag_char_small` snapshot was corrected after fixing mono-off nested
carrier identity for `Array (Array char)`, so use the checked-in snapshot
rather than any earlier ad hoc mono-off timings.

## Benchmarks Covered

- `fib`
- `binarytrees`
- `matrixmultiply`
- `quicksort`
- `sudoku`
- `i_before_e`

## Usage

```bash
# default suite (all benchmarks, all modes)
./v12/bench_suite

# targeted compiled mono-array comparison
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/zigzag_char_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/zigzag_char_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/sum_u32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/sum_u32_small/main.able

# reproducible baseline example
./v12/bench_suite \
  --runs 1 \
  --timeout 30 \
  --build-timeout 240 \
  --output-json v12/docs/perf-baselines/2026-03-03-benchmark-baseline.json
```

## JSON Output

The output file includes:

- `results`: one row per `(benchmark, mode, run_index)`
- `summary`: aggregated `ok/timeout/error` counts and average metrics for successful runs

Statuses:

- `ok`: command exited 0 within timeout
- `timeout`: command exceeded timeout
- `error`: non-timeout non-zero exit, including compiled-build failure
