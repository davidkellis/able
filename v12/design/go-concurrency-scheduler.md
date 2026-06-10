# Go Concurrency Design

Date: 2026-01-23
Status: Active Go implementation record; reconciled 2026-07-14

> **Normative contract:** spec §12 defines the public `spawn`, `Future`,
> `await`, and `future_*` behavior. This document records how the active Go
> reference implementation provides it. It does not select a scheduler
> performance project; any such work remains subject to the roadmap's
> cross-application selection rule.

## Scope

The active v12 workspace contains the Go tree-walker, bytecode VM, and AOT
compiler. The former TypeScript/cooperative continuation proposal is retained
only as historical context in [concurrency-executor-contract.md](concurrency-executor-contract.md).
It is not an implementation requirement for the active workspace.

## Executor surface

The interpreter uses a small internal executor abstraction for spawned work:

```go
type ProcTask func(context.Context) (runtime.Value, error)

type Executor interface {
    RunFuture(task ProcTask) *runtime.FutureValue
    Flush()
    PendingTasks() int
}
```

`ProcTask` is an internal Go name, unrelated to the removed Able-language
`proc` feature. `runtime.FutureValue` owns the context/cancel pair, terminal
status, result or failure, and memoized handle/value-view state.

### Serial executor

`New()` and `NewBytecode()` use `SerialExecutor` by default. It runs a FIFO
queue on one worker goroutine, which makes fixture execution deterministic
without requiring AST-level continuation snapshots. A task that calls
`future_yield()` is rescheduled behind runnable queued work. Re-entrant
`Future.value()` calls use `Drive` to progress a pending target without
deadlocking. Await-blocked work remains blocked until its registration wakes
it; it is not blindly requeued.

The worker publishes every dequeue-to-active handoff under the executor lock.
`Flush()` waits while that handoff is in flight, so it cannot mistake an
already-dequeued runnable task for an empty executor. The handoff marker is
cleared as the task becomes active, not when the task finishes: a task that
later pauses on a channel, mutex, timer, or other external wake remains
blocked work that `Flush()` may leave.

### Goroutine executor

`GoroutineExecutor` is available through the configurable interpreter
constructors and integration paths. It starts each spawned task in a Go
goroutine with a cancellable context. `future_yield()` calls `runtime.Gosched()`
as an advisory fairness hint only; callers must not depend on a particular
ordering.

## Future lifecycle and helpers

- `spawn` returns a `Future T` with `Pending`, `Resolved`, `Cancelled`, or
  `Failed` status. The first terminal outcome wins.
- `value()` and implicit value-view evaluation reuse a memoized result or
  `FutureError`; `cancel()` is best-effort and idempotent.
- `future_cancelled()` observes the current spawned task's cancellation flag
  and errors outside asynchronous work.
- `future_flush()` has **zero arguments**. It delegates to the configured
  executor. The serial executor drains runnable queued work. The goroutine
  executor waits for current work to settle but returns when every remaining
  task is blocked, so a blocked channel or I/O operation cannot make a flush
  hang forever.
- `future_pending_tasks()` is diagnostic only. It reports queued serial tasks
  or outstanding goroutine tasks and must not be used for functional program
  decisions.

## Error, cancellation, and blocking behavior

Executor task wrappers convert panics and evaluation failures into the common
Future failure representation. Cancellation signals the task context; code
that observes it resolves as cancelled, while an already-terminal task keeps
its terminal result. Nested waits retain progress guarantees: serial execution
drives the awaited future, and goroutine execution relies on Go's scheduler.

## Verification

Focused Go coverage includes:

- `interpreter_concurrency_test.go`: status/value/error/cancel, helper
  dispatch, pending-task reporting, and goroutine behavior.
- `interpreter_concurrency_executor_test.go`: serial fairness, re-entrant
  waits, memoization, the dequeue-to-active Flush handoff, blocked work, and
  goroutine parallelism.
- `interpreter_await_test.go` and bytecode async ordering tests: await,
  cancellation, and bytecode scheduling behavior.
- Executable fixtures `12_02_future_fairness_cancellation`,
  `12_03_spawn_future_status_error`, `12_04_future_handle_value_view`,
  `12_06_await_fairness_cancellation`, and `12_09_nested_spawn_native_context`.

The compiler emits an equivalent executor/runtime surface. Its generated
serial executor uses the same locked dequeue-to-active handoff invariant, and
`compiler_serial_flush_handoff_test.go` exercises that emitted code under the
race detector. Fixture and compiler coverage must remain green when changing
observable concurrency behavior.

## Boundaries

Timeout/select extensions, new scheduler policies, and a future non-Go runtime
need their own specified language behavior and coverage. They are not implied
by this completed design. Do not reopen scheduler tuning or add a
benchmark-shaped Future, Mutex, Channel, or executor optimization without a
qualifying shared performance leaf.
