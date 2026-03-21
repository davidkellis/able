# 2026-03-20 Concrete `Enumerable` `LinkedList i32` Compiled Snapshot

Scope:
- capture a reduced compiled-only baseline for the next generic-container
  hot-path tranche after the native `LinkedList -> Iterable -> Iterator`
  adapter work
- target the checked-in concrete `Enumerable` default-method fixture:
  `v12/fixtures/bench/linked_list_enumerable_i32_small/main.able`
- exercise direct compiled `Enumerable.map/filter/reduce` calls on a concrete
  `LinkedList i32` receiver

Environment:
- date: `2026-03-20`
- commit: `c20aaa22`
- tree: `dirty`
- host: `Linux 6.17.0-19-generic x86_64 GNU/Linux`
- runs: `3`
- timeout per run: `60s`

Harness details:
- compiled mode builds through `cmd/ablec` via `v12/bench_perf`
- direct `ablec` parity check confirms the compiled binary prints
  `382455000`
- this tranche closed shared generic-container lowering gaps, not a
  benchmark-only exception:
  - the compiler now binds higher-kinded interface self patterns like
    `Enumerable A for C _` to the concrete target type on compiled impl paths
  - static bound type-constructor calls like `C.default()` now resolve to
    compiled impl methods instead of `__able_env_get("C")`
  - native `Iterator<T>` carriers now satisfy compiled `for`-loop iterable
    lowering directly, so default `Enumerable` loops avoid iterator
    runtime-value round-trips that previously overflowed on large `LinkedList`
    graphs

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_enumerable_i32_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/linked_list_enumerable_i32_small` | `0.1667s` | `0.2167s` | `0.0167s` | `12.00` | reduced baseline for direct compiled concrete `Enumerable.map/filter/reduce` on `LinkedList i32` |

Takeaway:
- concrete generic/default `Enumerable` calls on `LinkedList` now stay on
  compiled impl functions instead of `__able_method_call_node(...)`
- the next remaining hot edge on this family is callback/carrier reduction
  inside the generic default impl bodies, not container-type resolution or
  iterator fallback correctness
