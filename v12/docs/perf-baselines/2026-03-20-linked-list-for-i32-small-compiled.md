# 2026-03-20 Native `LinkedList i32` Iterable Compiled Snapshot

Scope:
- capture a reduced compiled-only baseline for the first benchmark-worthy
  generic-container hot-path slice after the deeper generic-carrier
  correctness work
- target the checked-in `LinkedList i32` iterable benchmark fixture:
  `v12/fixtures/bench/linked_list_for_i32_small/main.able`
- exercise native `LinkedList -> Iterable -> Iterator` adapter lowering on the
  compiled path

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
  `1199940000`
- this tranche closed two shared native-interface bugs, not a benchmark-only
  exception:
  - inherited interface impls now synthesize native base-interface adapters
    (`Enumerable : Iterable`)
  - native interface concrete adapters now directly coerce compatible native
    interface return carriers instead of round-tripping through runtime values

Command:
```bash
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_for_i32_small/main.able
```

Results:

| benchmark | compiled real | compiled user | compiled sys | compiled gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/linked_list_for_i32_small` | `0.2000s` | `0.2867s` | `0.0233s` | `15.00` | reduced baseline for native `LinkedList -> Iterable -> Iterator` carrier lowering on the compiled path |

Takeaway:
- the first benchmark-worthy generic-container hot path in this program now
  stays on the shared native interface/container carrier path
- the remaining generic-container work is now broader performance widening and
  any other residual container/runtime edges, not this `LinkedList` iterable
  correctness hole
