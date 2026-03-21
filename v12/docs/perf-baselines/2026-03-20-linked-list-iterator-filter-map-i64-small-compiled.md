# 2026-03-20 Native Iterator `filter_map` Controller Snapshot

Scope:
- capture the compiled-only snapshot after closing the iterator-literal
  controller/runtime-value edge inside the native iterator default-method
  family
- target the checked-in benchmark fixture:
  `v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able`
- verify that `LinkedList.lazy().filter_map<i64>(...).collect<Array i64>()`
  stays on compiled iterator/controller helpers instead of routing `gen` back
  through dynamic member dispatch

Environment:
- date: `2026-03-20`
- commit: `c20aaa22`
- tree: `dirty`
- host: `Linux davidlinux 6.17.0-19-generic #19~24.04.2-Ubuntu SMP PREEMPT_DYNAMIC Fri Mar  6 23:08:46 UTC 2 x86_64 x86_64 x86_64 GNU/Linux`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct compiled parity check confirms the compiled binary prints
  `191952000`
- this tranche specifically isolates the remaining iterator-literal
  controller path: compiled iterator literals now bind `gen` as a
  compiler-owned `*__able_generator`, and `gen.yield(...)` / `gen.stop()` /
  bound `gen.yield` callable captures stay off `__able_method_call_node(...)`

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/linked_list_iterator_filter_map_i64_small` | `0.1267s` | `0.1433s` | `0.0133s` | `10.00` | native iterator controller + `filter_map` + `collect<Array i64>()` path |

Takeaway:
- the iterator default-method family is now closed through `filter_map`
- this tranche removes a real residual runtime edge inside iterator-literal
  bodies rather than widening to a new container family
- the next category is the next benchmark-worthy generic-container/runtime
  edge beyond iterator-literal controller cleanup
