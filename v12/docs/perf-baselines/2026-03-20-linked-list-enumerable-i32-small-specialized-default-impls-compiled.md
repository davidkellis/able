# 2026-03-20 Concrete `Enumerable` `LinkedList i32` Specialized-Default-Impl Follow-Up Snapshot

Scope:
- capture the follow-up compiled-only snapshot after closing the remaining
  callback/runtime-value carrier regression inside the concrete
  `LinkedList` `Enumerable` default-impl hot path
- target the checked-in benchmark fixture:
  `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`
- verify that specialized `Enumerable.lazy()` now calls the specialized
  `iterator_*_spec(...)` path directly instead of round-tripping
  `Iterator_A -> runtime.Value -> Iterator_i32`

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
- this tranche is primarily a correctness/native-carrier cleanup, not a new
  target-specific speed hack:
  - specialized impls now retain bound generic type bindings through
    compileability and render
  - specialized siblings are cached early enough to break recursive
    specialization loops during codegen
  - direct sibling selection inside default impl bodies now prefers those
    specialized sibling impls before the ordinary concrete receiver path

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_enumerable_i32_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/linked_list_enumerable_i32_small` | `0.1667s` | `0.2233s` | `0.0200s` | `15.33` | specialized default-impl follow-up after removing the `Iterator_A` runtime round-trip regression |

Takeaway:
- the linked-list concrete `Enumerable` benchmark is correct and stable again
- wall-clock returned to the earlier linked-list baseline; this tranche closed
  a stack-overflowing native-carrier regression rather than delivering a new
  speed step
- the next benchmark-worthy category is a different callback/generic-container
  runtime edge beyond this `LinkedList` default-impl path
