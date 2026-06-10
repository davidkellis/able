# Full bytecode collection-fixture health sweep — 2026-07-15

## Scope

This closes the bounded bytecode health milestone for the local
`fixture-collections` suite.  It validates fixture completion and expected
output, not comparative performance and not cross-mode parity.

## Method

The 27 fixtures were split into three independent nine-fixture batches so each
process could remain within the project’s short-test limit.  Every batch used:

```text
GOMEMLIMIT=1GiB
GOGC=50
GOMAXPROCS=1
outer timeout=55 seconds
per-fixture timeout=25 seconds
```

Each invocation used `bench_fixture_validate --modes bytecode`.  The harness
therefore marks parity as `partial`: compiled and tree-walker runs were
intentionally skipped.  In every case the bytecode result was `ok`, its output
matched the fixture’s baseline output, and no fixture timed out or errored.

| Batch | Fixtures | Result |
| --- | --- | --- |
| 1 | Array helpers, bit set, concurrent queue, deque, hash set/map, heap | 9/9 bytecode `ok` |
| 2 | Iterator matcher, lazy sequences, linked-list variants, list | 9/9 bytecode `ok` |
| 3 | Persistent collections, queue, tree map/set, vector | 9/9 bytecode `ok` |

The formerly bounded `deque_i32_small` and `heap_i32_small` cases completed in
batch 1.  `queue_i32_small` also completed in batch 3.  The per-batch JSON and
Markdown reports remain under `v12/tmp/fixture-health-2026-07-15/` while this
worktree is active.

## Decision

The collection health milestone is complete, so the stale active-plan wording
about deque and heap timeout cases is removed.  This is neither a timing
claim nor permission to optimize Array, queue, heap, or any named collection.
The external-scorecard selection gate remains unchanged: select a generic
performance change only after a material cross-cutting source change or a
shared concrete descendant hotspot in three unlike portable applications.
