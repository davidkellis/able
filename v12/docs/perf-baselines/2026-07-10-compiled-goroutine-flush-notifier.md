# Compiled Goroutine Executor Flush Notifier

## Decision

Keep the shared generated-runtime `Flush` notifier. The goroutine executor no
longer repeatedly calls `time.Sleep(0)` while rereading its `pending` and
`blocked` counters. Instead, it waits for a broadcast progress transition.
This is one compiler-template change for every generated goroutine-executor
program; it does not recognize a Channel, BinaryTrees, task-count, source, or
nominal-container shape. No `able-stdlib` change is required.

## Implementation

`pkg/compiler/generator_render_runtime_future.go` now gives
`__able_goroutine_executor` a mutex-protected progress generation and
condition variable. Future creation, completion, blocking, unblocking, and
blocked-handle cleanup publish a progress transition. `Flush` checks the same
existing terminal conditions—no pending work, or all remaining work
blocked—then waits until the generation advances before rechecking.

The generation is observed while holding the condition lock, so a transition
between the state check and `Wait` cannot be missed. `Broadcast` wakes every
concurrent `Flush` waiter. The serial executor, Able API, cancellation, task
outcome handling, and future value semantics are unchanged.

## Safety coverage

- Added `TestCompilerGeneratedGoroutineFlushNotifiesAllWaiters`. It renders a
  current generated runtime, verifies that `Flush` waits on progress rather
  than `time.Sleep(0)`, then runs generated Go tests under `-race`.
- The generated tests hold a future open while two concurrent `Flush` calls
  wait, release it, and require both callers to wake. They also mark the only
  remaining future blocked and require `Flush` to return promptly.
- Existing `TestCompilerFutureFlushReturnsWithBlockedGoroutineTasks` continues
  to validate compiled/tree-walker blocked-task behavior.
- `go test ./... -count=1 -timeout 60s` passes.

## Guarded process measurements

Current candidate binaries and retained pre-change binaries ran with
`GOMEMLIMIT=1GiB` and `GOGC=50`; goroutine workloads also used
`ABLE_EXECUTOR=goroutine`. Every run reproduced its checked output.

| Workload | Runs | Baseline mean (s) | Candidate mean (s) | Result |
| --- | ---: | ---: | ---: | --- |
| Channel-Rollup | 5 | 1.700 | 1.594 | 6.2% lower wall time; retained target win. |
| BinaryTrees small | 20 | 0.0410 | 0.0405 | Within 10 ms clock-resolution noise; no regression. |
| Lexical-Rollup serial guard | 15 | 0.0947 | 0.0933 | Within small-run noise; unchanged executor path. |

Candidate CPU profiles retain correct output across five Channel-Rollup and
100 BinaryTrees launches. The previous shared `Flush`/`time.Sleep(0)` CPU
wall no longer appears among material candidate samples:

- Channel-Rollup's pre-change merged profile attributed 39.3% cumulative and
  16.5% flat samples to generated `Flush`; the candidate profile has no
  material `Flush` sample.
- BinaryTrees' pre-change merged profile attributed 16.6% cumulative and
  8.8% flat samples to generated `Flush`; its candidate profile likewise has
  no material `Flush` sample.

The retained candidate profiles are
`.profiles/20260710_{channel_rollup,binarytrees_small}_compiled_flush_notifier.cpu.pprof`.

## Next recommendation

Profile a third independent compiled goroutine workload, starting with the
bounded `channel_roundtrip_i32_small` fixture, against Channel-Rollup and
BinaryTrees. Why: removing the shared flush poll reveals `bridge.currentGID`
and environment-swap work in Channel-Rollup, but BinaryTrees does not make
that path material. A third program that uses actual concurrent blocking is
needed before treating any remaining bridge leaf as general. The work entails
building/output-checking the current fixture under `ABLE_EXECUTOR=goroutine`,
collecting merged collector-free profiles, and selecting a new candidate only
if the same concrete bridge descendant repeats without regressing the serial
guard. Do not add a Channel, fixture, task-count, or named-container fast
path.
