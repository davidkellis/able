# Go Concurrency Design

This document captures the implementation strategy for Able v12 concurrency inside the Go interpreter.  
The guiding requirement is to **express Able’s unified `spawn`/`Future` semantics directly in terms of Go’s native concurrency primitives** (goroutines, channels, contexts, sync utilities) while keeping the runtime abstraction as thin as possible. The TypeScript interpreter now implements the same executor contract (see `design/concurrency-executor-contract.md`) on top of its cooperative scheduler, so both runtimes share identical helper semantics.

## 1. Goals and Constraints

* **Spec compliance.** Implement Section 12 of the v12 spec (`spawn`, `Future` interface, cancellation, status reporting, `future_yield`, `future_cancelled`) and surface the expected status/value/cancellation behaviours.
* **Go-native execution.** Launch asynchronous work with goroutines, use `context.Context` (and/or simple channels) for cancellation, and rely on Go’s blocking primitives for `value()`/future evaluation.
* **Lightweight abstraction.** Avoid building a bespoke scheduler. Instead manage bookkeeping around goroutines so the interpreter can translate between Able constructs and Go handles.
* **Deterministic tests.** Provide a deterministic execution mode for unit tests (single-threaded executor with controlled scheduling points) so parity suites remain repeatable.
* **Interop with existing runtime values.** Consolidate async handles into a unified `FutureValue` (or a renamed `FutureHandleValue`) so external code (TS harness, AST fixtures) can stay unchanged after renaming.
* **Error propagation.** Ensure panic/exception paths are converted into `FutureError` (matching the spec) and that `Future` evaluation re-raises errors in the caller context.

## 2. Architecture Overview

### 2.1 Execution Interfaces

Introduce a small `Executor` abstraction to decouple task submission from the underlying scheduling strategy:

```go
type Executor interface {
    RunFuture(fn FutureTask) *runtime.FutureValue
}

type FutureTask func(ctx context.Context) (runtime.Value, error)
```

* **Default executor (`GoroutineExecutor`).** For production runs, submit each task as a goroutine. Use `context.WithCancel` to propagate cancellation requests from the handle to the task body. House result/error propagation inside the handle structs.
* **Test executor (`SerialExecutor`).** For deterministic suites, run tasks synchronously (or in a controlled queue) on the calling goroutine. Still honour `future_yield`/`cancel` APIs by simulating cooperative checkpoints.
* The interpreter owns a singleton executor (defaulting to `GoroutineExecutor`) but exposes a test hook to swap it during unit tests.

### 2.2 Handle Structures

`runtime.FutureValue` is the unified async handle. It encapsulates both the handle view (status/value/cancel) and the memoized result used for implicit evaluation. We will:

* Add a `ctx`/`cancel` pair plus an internal goroutine-safe state machine (`Pending` → `Resolved`|`Cancelled`|`Failed`).
* Track the underlying result (`runtime.Value`) *or* an error value (`runtime.Value` implementing `Error`). Convert Go `error`/`panic` results into `FutureError` values per spec.
* Implement methods:
  * `status()` → snapshot without blocking.
  * `value()` → block on `sync.Cond` until terminal state, then:
    * Resolved: return success runtime value (wrapped in Able’s `!T` union semantics).
    * Failed/Cancelled: return `FutureError` as the error branch.
  * `cancel()` → invoke the stored `cancel` function, set state to `Cancelled` if still pending.
* Memoize the result once resolved so repeated `value()` calls and implicit evaluations reuse the same cached value/error.

### 2.3 Interpreter Integration

* **`spawn` evaluation**
  1. Capture the callee/block as a `FutureTask` closure (with its lexical environment and interpreter reference).
  2. Submit to the executor, obtaining a `FutureValue`.
  3. Return the handle as the result of the Able expression.
* **Native helpers**
  * `future_yield()` → call `runtime.Gosched()` (or, under the `SerialExecutor`, rotate the task queue) to provide cooperative hints.
  * `future_cancelled()` → check the handle’s cancellation flag (`context.Done()` / internal bool).

### 2.4 Cancellation & Error Semantics

* The task closure receives a `context.Context`; long-running library code can watch `ctx.Done()` to exit early.
* When cancellation is requested:
  * Set handle status to `Cancelled`.
  * The goroutine may still finish later; the first terminal state wins (standard Go cancellation semantics).
* Any panic or returned `error` from the task gets converted into `runtime.ErrorValue` representing `FutureError` (message + optional cause) before updating the handle. This keeps `value()` results aligned with the spec’s `!T`.

### 2.5 Deterministic Mode

* The `SerialExecutor` runs tasks immediately but still simulates asynchronous boundaries by:
  * Recording tasks in a queue.
  * Running each task to completion unless it calls `future_yield`, in which case it requeues itself.
* Useful for unit tests that assert specific interleavings (mirroring the TS cooperative scheduler).
* Production builds stick with `GoroutineExecutor`.

## 3. TypeScript Interpreter Implications

* No change to the TS runtime design: it already maintains a cooperative scheduler.  
* Documentation in this note plus the spec should clarify how TS must reflect:
  * `Future` handle API (`status`, `value`, `cancel`).
  * Future memoization.
  * Cancellation observation helpers.
* Tests should remain cross-language: fixtures for concurrency scenarios should pass on both runtimes.

## 4. Testing Strategy

1. **Unit tests (Go).**
   * Validate the state machine for `FutureValue`: pending → resolved/failed/cancelled.
   * Ensure `value()` blocks/unblocks correctly and returns the right union branch.
   * Confirm cancellation requests mark the handle and that tasks observing the context exit early.
   * Future memoization: multiple evaluations return cached values without rerunning the task.
2. **Integration tests.**
   * Parity fixtures matching the TS suite (multiple futures, cancellation, yield fairness).
   * Stress tests for racing `cancel()`/completion, repeated `value()` calls.
3. **Deterministic executor tests.**
   * Re-run key scenarios with the `SerialExecutor` to guarantee stable ordering for parity assertions.
4. **Inline driving for nested waits.**
   * When a future synchronously awaits another handle (e.g., nested `future.value()` calls or `future_flush` draining the queue), the serial executor now exposes `Drive(handle)` to steal and execute that task inline until it leaves the `Pending` state. This mirrors the TypeScript cooperative executor and prevents deadlocks in fixtures like `concurrency/future_value_reentrancy`.

## 5. Goroutine Executor Fairness

The production executor (`GoroutineExecutor`) defers scheduling decisions to Go’s runtime. Each Able `spawn` task runs inside its own goroutine with an attached `context.Context` for cancellation. While Go does not provide strict FIFO fairness guarantees, it does ensure that runnable goroutines eventually make progress. Able interprets that guarantee using the following rules:

* **`future_yield()` implementation.** In goroutine mode the helper simply invokes `runtime.Gosched()`. This cooperatively hints to Go’s scheduler that other goroutines should run, but it does **not** provide deterministic ordering. Authors should treat it as a best-effort fairness nudge rather than a strict context switch.
* **`future_flush()` semantics.** Because goroutines execute independently, `future_flush` becomes a no-op in goroutine mode. The helper still returns `nil` so programs can stay portable, but only the serial executor drains queued work synchronously.
* **Fairness expectations.** Programs must not assume alternation or round-robin behaviour when running under the goroutine executor; the only guarantee is forward progress. Tests that need deterministic ordering should run under the serial executor.
* **Fixture guidance.** Shared fixtures that assert trace ordering (e.g., `future_flush_fairness`) rely on the serial executor and should be mirrored in Go via either the fixture parity harness or dedicated unit tests. When adding new concurrency scenarios, include both a serial-executor regression test and an explanation of how the goroutine executor maintains spec compliance (usually via `Gosched` + eventual progress).

These notes satisfy the follow-up from the concurrency PLAN item (“Document remaining scheduler guarantees”), and future contributors should reference this section before modifying the executor helpers.

## 6. Open Questions / Follow-ups

* **`future_yield` semantics:** Calling `runtime.Gosched()` is lightweight but not strictly deterministic. Document and test expected behaviour; use the deterministic executor for precise interleavings in tests.
* **Timeout/Select helpers:** Out of scope for the initial pass, but the design should allow library-level helpers (e.g., `Channel.select`, timers) to integrate seamlessly.
* **TS parity docs:** Update TypeScript design notes once the Go implementation lands to keep the cooperative scheduler aligned.

## 7. Current Status (2025-10-18)

Implementation highlights now in `master`:

* **Executor layer shipped.** `GoroutineExecutor` and `SerialExecutor` live in `pkg/interpreter/executor.go`, both funnelling `FutureTask` closures through shared bookkeeping so handles observe consistent lifecycles.
* **Handle lifecycle extended.** `runtime.FutureValue` gained context/cancel wiring, status snapshots, and memoised results; it owns both handle and value-view behaviour.
* **Interpreter integration.** `spawn` expressions enqueue cloned-environment tasks, and native helpers (`future_yield`, `future_cancelled`, `future_flush`, `status`, `value`, `cancel`) mirror the spec via bound methods in `interpreter_concurrency.go`.
* **Testing.** Runtime unit tests cover resolve/fail/cancel basics, and new interpreter tests assert resolved + failed handle/future scenarios under the serial executor.

Key decisions captured during implementation:

* **Environment cloning.** Async tasks capture a snapshot of the lexical environment chain to isolate concurrent mutation, matching TypeScript behaviour.
* **Async context guard.** A private payload tracks whether the current task is a spawned task so helpers like `future_cancelled` can error when misused.
* **Error propagation shape.** Failures are uniform `runtime.ErrorValue` instances carrying a `future_error` payload struct so downstream code can inspect details consistently.

Outstanding work before we call concurrency “done”:

1. Exercise cancellation/yield paths (including repeated `value()` calls) in dedicated Go tests and parity fixtures.
