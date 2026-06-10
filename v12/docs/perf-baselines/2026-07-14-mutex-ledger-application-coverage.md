# Mutex Ledger application coverage — 2026-07-14

## Decision

Add `mutex_ledger` as the 29th verifier-backed external application, a
dedicated `mutex-ledger` suite, and a member of the `concurrency` and broad
`coverage` suites. Keep the stable 16-program `generality` suite unchanged.

Mutex Ledger has four workers process 2,048 deterministic numeric entries
each. Each entry updates a shared three-field nominal `Ledger` through the
public `able.concurrency.with_lock` helper. The workers return independent
subtotals after `future_flush()` has allowed their work to complete. Every
implementation prints the same checked result:

```text
4:2048:8192:4098178594:292511:4098178594
```

The matching Go, Python, and Ruby programs use their ordinary mutex and thread
facilities. The workload does not assume a lock-acquisition order; it checks
only the observable mutual-exclusion and completion result. It therefore adds
public Mutex/`with_lock` coverage without assigning value to a particular host
scheduler policy.

## Correctness repairs exposed by the application

This coverage work found two generic runtime defects; neither is a performance
optimization or a benchmark-specific path.

1. Bytecode raw integer stack cells are reusable scratch cells. A value that
   escapes into a struct field, array literal, or array write must be
   materialized first. Otherwise later arithmetic can overwrite an earlier
   aggregate element. The VM now materializes such values at those generic
   aggregate-escape boundaries. Named-struct member plans also validate their
   field metadata against the instruction member name before using a positional
   slot.
2. Both the bytecode interpreter and generated compiled runtime made an
   unlocking task wait synchronously for a particular locker handoff. With
   three or more waiters, another locker could consume the signal and leave all
   tasks asleep. Unlock now performs normal condition-variable notification;
   lock acquisition no longer signals an unlocker that is not waiting.

Focused regressions pin the aggregate lifetime, bytecode multi-waiter, and
compiled multi-waiter cases. The canonical `able-stdlib` source needs no
change: its public `with_lock` implementation already expresses the specified
lock/ensure/unlock contract.

## Verifier-backed comparison

This local three-process screen used the normal bounded 45-second process cap.
It is a product-status row, not a new scorecard baseline: the run was not
CPU-pinned and one new application cannot identify a general optimization
candidate. Every Able output passed the sibling Ruby verifier.

| Mode/reference | Mean real seconds | Relative result |
| --- | ---: | ---: |
| Go 1.26 | 0.0045 | baseline |
| Python | 0.0379 | baseline |
| Ruby | 0.0626 | baseline |
| Able compiled | 0.9300 | 206.67x Go |
| Able bytecode | 0.5467 | 14.42x Python; 8.73x Ruby |
| Able tree-walker | 0.7767 | semantic comparison only |

The exact collection commands were:

```text
./v12/bench_refresh_go_refs --benchmarks mutex_ledger --runs 3 --timeout 30 \
  --output-json /tmp/mutex-ledger-go.json

./v12/bench_refresh_interpreter_refs --benchmarks mutex_ledger --runs 3 --timeout 30 \
  --output-json /tmp/mutex-ledger-interpreters.json

./v12/bench_compare_external --benchmarks mutex_ledger \
  --modes compiled,bytecode,treewalker --runs 3 --timeout 45 \
  --reference-json /tmp/mutex-ledger-interpreters.json \
  --go-reference-json /tmp/mutex-ledger-go.json
```

The dedicated bytecode lowering audit passes, and the complete `coverage`
audit now passes with 29 applications, 127 functions, and 7,041 instructions.
The compiled dynamic-boundary audit also completes with a verified binary.

## Conclusion and next gate

The new row is intentionally retained despite its material current gap. It is
one application with a distinct public synchronization boundary, so it cannot
justify a Mutex-, closure-, struct-, or benchmark-specific optimization. Before
profiling or changing a shared scheduler/runtime path, add a distinct portable
application for the remaining public `Mutex.await_lock`/`Awaitable` boundary
and require the same concrete descendant to repeat across it and Mutex Ledger.
That gives a genuine two-application synchronization gate while preserving the
rule that real-program breadth, not a fast synthetic benchmark, selects work.
