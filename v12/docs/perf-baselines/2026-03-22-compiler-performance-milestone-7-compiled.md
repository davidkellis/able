# 2026-03-22 Compiler Performance Milestone 7 Snapshot

Scope:
- close the compiler performance milestone with one shared callable/runtime
  scaffolding change instead of another nominal-type-specific fast path
- add the reduced checked-in recursion benchmark:
  `v12/fixtures/bench/fib_i32_small/main.able`
- remeasure representative compiled benchmark families after switching
  generated package-env entry scaffolding from unconditional `SwapEnv(...)`
  to conditional `SwapEnvIfNeeded(...)`

Environment:
- date: `2026-03-22`
- tree: `dirty`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct parity check for the new recursion fixture confirms compiled output:
  `102334155`
- the recursion baseline before this tranche used the same reduced `fib(40)`
  source as an ad hoc target and measured `10.1300s`

Commands:
```bash
./v12/bench_perf --runs 1 --timeout 60 --modes compiled --keep --show-output \
  v12/fixtures/bench/fib_i32_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/fib_i32_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/heap_i32_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/examples/benchmarks/matrixmultiply.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/fib_i32_small` | `2.7567s` | `2.7567s` | `0.0100s` | `0.00` | recursive compiled functions no longer pay redundant package-env swaps on every entry |
| `bench/heap_i32_small` | `0.2900s` | `0.2867s` | `0.0133s` | `5.00` | shared compiled method/function env entry is cheaper on the nominal generic container path too |
| `bench/linked_list_iterator_pipeline_i64_small` | `0.1433s` | `0.1467s` | `0.0200s` | `9.67` | iterator/default-method pipeline picks up the same shared callable env-entry win |
| `bench/matrixmultiply_f64_small` | `0.1167s` | `0.1200s` | `0.0067s` | `7.33` | array/matrix family stays in the current best performance band after the callable/runtime change |
| `examples/benchmarks/matrixmultiply` | `1.0633s` | `1.0633s` | `0.0300s` | `13.33` | full canonical benchmark remains fast and slightly improves on the prior `1.0833s` snapshot |

Takeaway:
- Milestone 7 is now closed with all four planned benchmark families covered by
  checked-in reduced targets or the canonical full matrix benchmark.
- The remaining major recursion/call overhead gap was redundant package-env
  swapping on compiled entry. Closing that shared scaffolding reduced the new
  `fib_i32_small` benchmark from `10.1300s` to `2.7567s`.
- The same shared change also improved representative container/iterator paths
  without adding any new nominal-type-specific lowering rule.
- The next compiler milestone is release validation, not another open-ended
  performance tranche.
