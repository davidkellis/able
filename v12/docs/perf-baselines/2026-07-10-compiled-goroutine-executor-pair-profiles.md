# Compiled Goroutine-Executor Pair Profiles

## Decision

Keep no implementation change in this profiling tranche. The paired profiles
do, however, qualify one generic next candidate: replace the compiled
goroutine executor's busy-yielding `Flush` wait with an event-driven progress
wait. The candidate is shared generated runtime code, repeats beneath the same
concrete function in two independent spawned programs, and is not a channel,
BinaryTrees, task-count, or nominal-container specialization.

## Method

- Built the maintained `binarytrees_small` fixture as a current compiled
  binary. It is the same independently authored BinaryTrees `spawn` /
  `future_flush` algorithm, bounded at depth 12 for repeatable fixture runs.
- Confirmed the generated-binary default is the serial executor, then selected
  the same explicit `ABLE_EXECUTOR=goroutine` mode used by the existing
  Channel-Rollup profile. This is required for a like-for-like test of the Go
  pre-emptive executor; serial output is only a selection control.
- Ran 100 normal, collector-free BinaryTrees processes with `GOMEMLIMIT=1GiB`
  and `GOGC=50`. Every output matched the serial control exactly, including
  `stretch tree of depth 13\\t check: 16383` and `long lived tree of depth
  12\\t check: 8191`.
- Merged the goroutine profiles into
  `.profiles/20260710_binarytrees_small_goroutine_compiled_main_collector_free.cpu.pprof`.
  The serial control profile is retained as
  `.profiles/20260710_binarytrees_small_compiled_main_collector_free.cpu.pprof`.
- Compared it with the five-launch, output-checked Channel-Rollup merged
  profile from the preceding current compiled-main tranche, also executed with
  `ABLE_EXECUTOR=goroutine`.

## Paired result

| Concrete generated runtime descendant | Channel-Rollup (19.67 s samples) | BinaryTrees (2.96 s samples) | Selection result |
| --- | ---: | ---: | --- |
| `(*__able_goroutine_executor).Flush` | 39.3% cumulative, 16.5% flat | 16.6% cumulative, 8.8% flat | Repeats materially. |
| `Flush` loop's `time.Sleep(0)` line | 4.56 s cumulative | 0.24 s cumulative | Repeats in the same busy-yield wait loop. |
| `bridge.currentGID` | 56.7% cumulative | 1.35% cumulative | Does not repeat materially; not a candidate. |
| Program body | Channel send/receive and environment work | tree construction 61.2% cumulative and allocation/GC | Workload-local; do not optimize from it. |

BinaryTrees under the default serial executor is tree-allocation dominated and
does not select the goroutine path. Under the explicitly selected goroutine
executor it preserves the identical result, launches its five spawned workers
through `RunFuture`, and reaches the same `Flush` loop. That distinguishes the
shared executor wait from BinaryTrees' tree workload and from Channel-Rollup's
channel operations.

## Candidate boundary and guards

The generated goroutine executor currently loops over atomic `pending` and
`blocked` counts and calls `time.Sleep(0)` until all work completes or all
remaining work is blocked. An event-driven progress notification can let
`Flush` sleep until a future completes or changes blocked state rather than
repeatedly yielding and rereading those counts.

Any implementation must preserve these semantics:

- return promptly when no work remains and when every remaining task is
  blocked;
- wake on completion, blocking, unblocking, and future creation without
  missing a transition or stranding concurrent `Flush` callers;
- retain goroutine executor re-entrant waits, cancellation, and error
  propagation; and
- leave the serial executor and sequential generated programs unchanged.

The existing `TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks` and
compiler/interpreter concurrency parity coverage are mandatory guards, in
addition to output checks and bounded A/B process measurements for both
Channel-Rollup and BinaryTrees with Lexical-Rollup as a serial guard.

## Next recommendation

Implement and evaluate a generated goroutine-executor progress notifier for
`Flush`, likely a broadcast-capable condition/sequence mechanism around task
completion and blocked-state transitions. Why: the paired profiles now show
the same busy-yield loop materially in two independent concurrent programs,
while ruling out `currentGID` as a broad target. The work entails changing only
the shared compiler runtime template, adding missed-wakeup and multi-flush
tests, running the existing blocked-task/parity guards, and comparing both
concurrency programs plus Lexical-Rollup before retention. Do not alter an
Able benchmark, specialize channels or BinaryTrees, or change stdlib APIs.
