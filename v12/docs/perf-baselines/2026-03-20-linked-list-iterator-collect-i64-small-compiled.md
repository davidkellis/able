# 2026-03-20 Native Iterator Collect Mono-Array Snapshot

Scope:
- capture the compiled-only snapshot after closing the mono-array-enabled
  `Iterator.collect<Array i64>()` / specialized-array accumulator path
- target the checked-in benchmark fixture:
  `v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able`
- compare mono-array default-on versus explicit
  `--no-experimental-mono-arrays` on the same reduced iterator pipeline

Environment:
- date: `2026-03-20`
- commit: `c20aaa22`
- tree: `dirty`
- host: `Linux davidlinux 6.17.0-19-generic #19~24.04.2-Ubuntu SMP PREEMPT_DYNAMIC Fri Mar  6 23:08:46 UTC 2 x86_64 x86_64 x86_64 GNU/Linux`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct compiled parity check confirms the mono-array-enabled binary prints
  `382455000`
- the mono-array-enabled collect path now lowers through a generated compiled
  helper with a specialized `*__able_array_i64` accumulator instead of the
  residual `__able_method_call_node(...)` + `__able_array_i64_from(...)`
  bridge

Commands:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able

./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able
```

Results:

| benchmark | mode | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `bench/linked_list_iterator_collect_i64_small` | mono on | `0.1833s` | `0.2167s` | `0.0300s` | `14.00` | specialized `Iterator.collect<Array i64>()` helper |
| `bench/linked_list_iterator_collect_i64_small` | mono off | `0.1833s` | `0.2167s` | `0.0300s` | `13.33` | generic array accumulator path |

Takeaway:
- the mono-array-enabled `Iterator.collect<Array T>()` correctness bug is
  closed on the reduced iterator pipeline family
- on this reduced workload the fix is performance-neutral in wall-clock terms
  and essentially neutral in GC terms
- the next category is broader performance widening on the next hot generic
  container/runtime edges, not more cleanup on this specific collect bridge
