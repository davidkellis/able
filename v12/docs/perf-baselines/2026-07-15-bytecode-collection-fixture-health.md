# Bytecode collection-fixture health recheck — 2026-07-15

## Purpose

This is a bounded health recheck of the two local collection fixtures called
out by older plan notes.  It is not an external performance comparison and it
does not authorize a VM, compiler, generated-runtime, canonical-stdlib, or
benchmark change.

## Environment

Each command used one process and the following limits:

```text
GOMEMLIMIT=1GiB
GOGC=50
GOMAXPROCS=1
```

The single-mode harness runs used a 55-second outer cap.  The validating
rerun used the same limits with a 25-second per-fixture cap.

## Results

`bench_suite` completed both bytecode fixtures successfully:

| Fixture | Wall time | Garbage collections |
| --- | ---: | ---: |
| `deque_i32_small` | 0.73 s | 7 |
| `heap_i32_small` | 4.14 s | 6 |

`bench_fixture_validate --benchmarks deque_i32_small,heap_i32_small --modes
bytecode --timeout 25` also reported bytecode `ok` for both manifest-required
exit-zero fixtures.  Its reported values were `26846973` for the deque and
`-211812354` for the heap.  The report labels parity `partial` because the
tree-walker and compiled modes were intentionally not run; this is a
bytecode-health check, not a cross-mode parity claim.

`bench_bytecode_audit --suite corpus-full` passed in 0.448 s.  Its static
lowering audit covers 109 programs, 410 functions, and 20,205 instructions.
That confirms the corpus remains available to exercise bytecode lowering, but
it is not a timing result.

The machine-readable and Markdown command reports are retained under
`v12/tmp/fixture-health-2026-07-15/` while the worktree is active.

## Decision

Historical timeout wording for these two bytecode fixture invocations is no
longer current.  There is no failing collection health signal and no repeated
cross-application leaf.  Keep the source-selection gate from the post-extern
scorecard: profile and consider a general optimization only after a material
cross-cutting source change, or after a newly needed portable application
reveals the same descendant hotspot in three unlike applications.  Do not
select a named-container or local-fixture-specific fast path.
