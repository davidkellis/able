# 2026-03-20 Native Iterator Default-Method Pipeline Snapshot

Scope:
- capture the compiled-only snapshot after closing the native iterator
  default-method path for the representative
  `LinkedList.lazy().map<i64>(...).filter(...).next()` pipeline
- target the checked-in benchmark fixture:
  `v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able`
- verify that ordinary default native-interface methods on iterator carriers
  now lower through direct compiled helpers instead of the runtime adapter
  method layer

Environment:
- date: `2026-03-20`
- commit: `c20aaa22`
- tree: `dirty`
- host: `Linux 6.17.0-19-generic #19~24.04.2-Ubuntu SMP PREEMPT_DYNAMIC Fri Mar  6 23:08:46 UTC 2 x86_64 GNU/Linux`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity check confirms the compiled binary prints
  `382455000`
- this tranche specifically isolates the `map/filter/next` iterator helper
  path; the first attempted `collect<Array i64>().reduce(...)` benchmark shape
  exposed a separate open issue on the mono-array-enabled CLI/default path and
  is not mixed into this snapshot

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/linked_list_iterator_pipeline_i64_small` | `0.1800s` | `0.2100s` | `0.0267s` | `13.33` | native iterator default-method `map/filter -> next()` path stays on compiled helpers |

Takeaway:
- the representative `LinkedList.lazy().map/filter -> next()` path is now
  closed on the shared native iterator carrier design
- this is a correctness/native-carrier tranche with a stable focused
  benchmark, not a broader container widening step
- the next follow-up on this family is the separate mono-array-enabled
  `Iterator.collect<Array T>()` / specialized-array accumulator interaction
