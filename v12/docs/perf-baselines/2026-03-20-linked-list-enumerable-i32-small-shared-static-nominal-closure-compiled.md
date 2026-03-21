# 2026-03-20 Shared Static-Nominal Closure Snapshot for `LinkedList` `Enumerable` (`i32`)

Scope:
- capture the follow-up compiled-only snapshot after closing the remaining
  shared static nominal receiver/struct-literal refinement gap on the reduced
  `LinkedList -> Enumerable -> LazySeq` family
- target the checked-in benchmark fixture:
  `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`
- verify that the compiled path stays native through specialized
  `Iterable<i32>` adapter synthesis and concrete `LazySeq<i32>`
  construction, without fallback dispatch

Environment:
- date: `2026-03-20`
- tree: `dirty`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity check confirms the compiled binary prints
  `382455000`
- this tranche is a shared-lowering closure, not a target-specific
  named-container optimization:
  - recursive type substitution now resolves chained specialization bindings
    transitively
  - expected-type refinement now upgrades static nominal targets and struct
    literals like `LazySeq { ... }` to their concrete specialized carrier
  - native interface concrete-impl matching now recognizes specialized
    receivers through the shared target-template path

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_enumerable_i32_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/linked_list_enumerable_i32_small` | `0.1633s` | `0.1700s` | `0.0167s` | `8.33` | shared static nominal receiver/struct-literal closure on the reduced `LinkedList -> Enumerable -> LazySeq` family |

Takeaway:
- the reduced `LinkedList` `Enumerable` benchmark is now fully back on the
  shared static nominal path with no residual fallback gap
- this follow-up improves GC materially versus the earlier linked-list
  follow-ups while keeping wall-clock slightly better as well
- the next benchmark-worthy category is a different generic
  container/runtime edge beyond this now-closed family
