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
comparison and `v12/fixtures/bench/hashmap_i32_small/main.able` for the first
broader native-container (`HashMap i32 i32` + `Map i32 i32`) comparison and
`v12/fixtures/bench/heap_i32_small/main.able` for the broader array-backed
container family (`Heap i32`) comparison and the shared generic nominal-method
specialization follow-up and bound generic field/member carrier refinement
follow-up on that same benchmark and
`v12/fixtures/bench/linked_list_for_i32_small/main.able` for the first
benchmark-worthy generic-container hot-path (`LinkedList -> Iterable ->
Iterator`) comparison and
`v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` for the next
concrete generic/default-method container hot path (`LinkedList.map/filter/reduce`)
comparison and the shared static nominal receiver/struct-literal closure
follow-up on that same benchmark and
`v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able` for
the next iterator default-method hot path
(`LinkedList.lazy().map<i64>(...).filter(...).next()`) comparison and
`v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able` for the
mono-array-enabled iterator collect/reduce follow-up
(`LinkedList.lazy().map<i64>(...).filter(...).collect<Array i64>().reduce(...)`)
comparison and
`v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able` for
the iterator-literal controller / `filter_map` follow-up
(`LinkedList.lazy().filter_map<i64>(...).collect<Array i64>().reduce(...)`)
comparison, while the full
`matrixmultiply` workload in `v12/examples/benchmarks/matrixmultiply.able`
remains the canonical suite entry used by `v12/bench_suite`. Current focused
snapshots for these reduced fixtures are checked in at:

- `v12/docs/perf-baselines/2026-03-19-mono-array-f64-matrixmultiply-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-scalar-propagation-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-float-int-casts-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-frame-elision-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-propagation-pointer-elision-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-u32-sum-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-hashmap-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-heap-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-heap-i32-generic-nominal-method-specialization-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-heap-i32-bound-generic-field-carrier-refinement-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-for-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-specialized-default-impls-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-pipeline-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-collect-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-filter-map-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`

The reduced recursion/call-overhead benchmark is now:
- `v12/fixtures/bench/fib_i32_small/main.able`

The current representative compiled Milestone 7 snapshot is:
- `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`
  - `bench/fib_i32_small`: `2.7567s`, `0.00` GC
  - `bench/heap_i32_small`: `0.2900s`, `5.00` GC
  - `bench/linked_list_iterator_pipeline_i64_small`: `0.1433s`, `9.67` GC
  - `bench/matrixmultiply_f64_small`: `0.1167s`, `7.33` GC
  - `examples/benchmarks/matrixmultiply`: `1.0633s`, `13.33` GC

The `zigzag_char_small` snapshot was corrected after fixing mono-off nested
carrier identity for `Array (Array char)`, so use the checked-in snapshot
rather than any earlier ad hoc mono-off timings.

The current best matrix snapshots are now:
- reduced target:
  `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`,
  which records `0.1133s` / `7.00` GC on the compiled
  `matrixmultiply_f64_small` target after removing the synthetic static-array
  loop-induction checked-arithmetic scaffolding through shared primitive
  counted-loop lowering
- full canonical benchmark:
  `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`,
  which records `1.0833s` / `13.00` GC on the compiled
  `v12/examples/benchmarks/matrixmultiply.able` path

The follow-up affine integer snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
proves the remaining `build_matrix` `i - j` / `i + j` helper calls are gone,
but it is effectively performance-neutral relative to the counted-loop
snapshot on this benchmark family.

The follow-up subtraction range-proof snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
proves the widened inline overflow branch is now gone for `build_matrix`
`i - j`, but `i + j` still carries the widened checked-add path and the
benchmark remains in the same band as the counted-loop baseline.

The follow-up upper-bound range-proof snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
proves the remaining widened inline overflow branch is now gone for
`build_matrix` `i + j` too. The inner-loop affine add/sub gap is closed, but
the benchmark still remains in the same band as the counted-loop baseline.

The iterator-pipeline family is now split intentionally:
- `linked_list_iterator_pipeline_i64_small` isolates the already-closed native
  `map/filter/next` path
- `linked_list_iterator_collect_i64_small` isolates the now-closed
  mono-array-enabled `collect<Array i64>().reduce(...)` follow-up
- `linked_list_iterator_filter_map_i64_small` isolates the now-closed
  iterator-literal controller / `filter_map(...).collect<Array i64>()`
  follow-up

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
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/hashmap_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/heap_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_for_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_enumerable_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/fib_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able

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
