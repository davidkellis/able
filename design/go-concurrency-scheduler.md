# Go Concurrency Design

This document captures the implementation strategy for Able v10 concurrency inside the Go interpreter.  
The guiding requirement is to **express Able’s `proc`/`spawn` semantics directly in terms of Go’s native concurrency primitives** (goroutines, channels, contexts, sync utilities) while keeping the runtime abstraction as thin as possible. The TypeScript interpreter will continue to emulate the same semantics on top of its cooperative scheduler; the Go implementation should therefore present a clear contract that the TS runtime can mirror.

## 1. Goals and Constraints

* **Spec compliance.** Implement Section 12 of the v10 spec (`proc`, `spawn`, `Proc` interface, futures, cancellation, status reporting, `proc_yield`, `proc_cancelled`) and surface the expected status/value/cancellation behaviours.
* **Go-native execution.** Launch asynchronous work with goroutines, use `context.Context` (and/or simple channels) for cancellation, and rely on Go’s blocking primitives for `value()`/future evaluation.
* **Lightweight abstraction.** Avoid building a bespoke scheduler. Instead manage bookkeeping around goroutines so the interpreter can translate between Able constructs and Go handles.
* **Deterministic tests.** Provide a deterministic execution mode for unit tests (single-threaded executor with controlled scheduling points) so parity suites remain repeatable.
* **Interop with existing runtime values.** Reuse/extend `ProcHandleValue` and `FutureValue` from `pkg/runtime/values.go` so external code (TS harness, AST fixtures) can stay unchanged.
* **Error propagation.** Ensure panic/exception paths are converted into `ProcError` (matching the spec) and that `Future` evaluation re-raises errors in the caller context.

## 2. Architecture Overview

### 2.1 Execution Interfaces

Introduce a small `Executor` abstraction to decouple task submission from the underlying scheduling strategy:

```go
type Executor interface {
    RunProc(fn ProcTask) *runtime.ProcHandleValue
    RunFuture(fn ProcTask) *runtime.FutureValue
}

type ProcTask func(ctx context.Context) (runtime.Value, error)
```

* **Default executor (`GoroutineExecutor`).** For production runs, submit each task as a goroutine. Use `context.WithCancel` to propagate cancellation requests from the handle to the task body. House result/error propagation inside the handle structs.
* **Test executor (`SerialExecutor`).** For deterministic suites, run tasks synchronously (or in a controlled queue) on the calling goroutine. Still honour `proc_yield`/`cancel` APIs by simulating cooperative checkpoints.
* The interpreter owns a singleton executor (defaulting to `GoroutineExecutor`) but exposes a test hook to swap it during unit tests.

### 2.2 Handle Structures

`runtime.ProcHandleValue` already exists with status and condition variables. We will:

* Add a `ctx`/`cancel` pair plus an internal goroutine-safe state machine (`Pending` → `Resolved`|`Cancelled`|`Failed`).
* Track the underlying result (`runtime.Value`) *or* an error value (`runtime.Value` implementing `Error`). Convert Go `error`/`panic` results into `ProcError` values per spec.
* Implement methods:
  * `status()` → snapshot without blocking.
  * `value()` → block on `sync.Cond` until terminal state, then:
    * Resolved: return success runtime value (wrapped in Able’s `!T` union semantics).
    * Failed/Cancelled: return `ProcError` as the error branch.
  * `cancel()` → invoke the stored `cancel` function, set state to `Cancelled` if still pending.

`runtime.FutureValue` will wrap a `ProcHandleValue`:

* Use `sync.Once` + `chan struct{}` to memoize the result.
* Evaluation (e.g., when the future is used in an expression) blocks until the underlying handle resolves and returns the memoized value/error. Failures re-raise in the evaluating context (as per spec).

### 2.3 Interpreter Integration

* **`proc` evaluation**
  1. Capture the callee/block as a `ProcTask` closure (with its lexical environment and interpreter reference).
  2. Submit to the executor, obtaining a `ProcHandleValue`.
  3. Return the handle as the result of the Able expression.
* **`spawn` evaluation**
  1. Wrap the same `ProcTask` and submit via `RunFuture`.
  2. Return the resulting `FutureValue`.
* **Native helpers**
  * `proc_yield()` → call `runtime.Gosched()` (or, under the `SerialExecutor`, rotate the task queue) to provide cooperative hints.
  * `proc_cancelled()` → check the handle’s cancellation flag (`context.Done()` / internal bool).

### 2.4 Cancellation & Error Semantics

* The task closure receives a `context.Context`; long-running library code can watch `ctx.Done()` to exit early.
* When cancellation is requested:
  * Set handle status to `Cancelled`.
  * The goroutine may still finish later; the first terminal state wins (standard Go cancellation semantics).
* Any panic or returned `error` from the task gets converted into `runtime.ErrorValue` representing `ProcError` (message + optional cause) before updating the handle. This keeps `value()` results aligned with the spec’s `!T`.

### 2.5 Deterministic Mode

* The `SerialExecutor` runs tasks immediately but still simulates asynchronous boundaries by:
  * Recording tasks in a queue.
  * Running each `proc` task to completion unless it calls `proc_yield`, in which case it requeues itself.
* Useful for unit tests that assert specific interleavings (mirroring the TS cooperative scheduler).
* Production builds stick with `GoroutineExecutor`.

## 3. TypeScript Interpreter Implications

* No change to the TS runtime design: it already maintains a cooperative scheduler.  
* Documentation in this note plus the spec should clarify how TS must reflect:
  * `Proc` handle API (`status`, `value`, `cancel`).
  * Future memoization.
  * Cancellation observation helpers.
* Tests should remain cross-language: fixtures for concurrency scenarios should pass on both runtimes.

## 4. Testing Strategy

1. **Unit tests (Go).**
   * Validate the state machine for `ProcHandleValue`: pending → resolved/failed/cancelled.
   * Ensure `value()` blocks/unblocks correctly and returns the right union branch.
   * Confirm cancellation requests mark the handle and that tasks observing the context exit early.
   * Future memoization: multiple evaluations return cached values without rerunning the task.
2. **Integration tests.**
   * Parity fixtures matching the TS suite (multiple procs/futures, cancellation, yield fairness).
   * Stress tests for racing `cancel()`/completion, repeated `value()` calls.
3. **Deterministic executor tests.**
   * Re-run key scenarios with the `SerialExecutor` to guarantee stable ordering for parity assertions.

## 5. Open Questions / Follow-ups

* **`proc_yield` semantics:** Calling `runtime.Gosched()` is lightweight but not strictly deterministic. Document and test expected behaviour; use the deterministic executor for precise interleavings in tests.
* **Timeout/Select helpers:** Out of scope for the initial pass, but the design should allow library-level helpers (e.g., `Channel.select`, timers) to integrate seamlessly.
* **TS parity docs:** Update TypeScript design notes once the Go implementation lands to keep the cooperative scheduler aligned.

## 6. Current Status (2025-10-18)

Implementation highlights now in `master`:

* **Executor layer shipped.** `GoroutineExecutor` and `SerialExecutor` live in `pkg/interpreter/executor.go`, both funnelling `ProcTask` closures through shared bookkeeping so handles/futures observe consistent lifecycles.
* **Handle lifecycle extended.** `runtime.ProcHandleValue` gained context/cancel wiring, status snapshots, and memoised results; `FutureValue` now wraps a handle directly and defers to the same state machine.
* **Interpreter integration.** `proc` / `spawn` expressions enqueue cloned-environment tasks, and native helpers (`proc_yield`, `proc_cancelled`, `proc_flush`, `status`, `value`, `cancel`) mirror the spec via bound methods in `interpreter_concurrency.go`.
* **Testing.** Runtime unit tests cover resolve/fail/cancel basics, and new interpreter tests assert resolved + failed handle/future scenarios under the serial executor.

Key decisions captured during implementation:

* **Environment cloning.** Async tasks capture a snapshot of the lexical environment chain to isolate concurrent mutation, matching TypeScript behaviour.
* **Async context guard.** A private payload tracks whether the current task is a proc or future so helpers like `proc_cancelled` can error when misused.
* **Error propagation shape.** Failures are uniform `runtime.ErrorValue` instances carrying a `proc_error` payload struct so downstream code can inspect details consistently.

Outstanding work before we call concurrency “done”:

1. Exercise cancellation/yield paths (including repeated `value()` calls) in dedicated Go tests and parity fixtures.
2. Port the executor/handle contract to the TypeScript interpreter so both runtimes share identical helper APIs.
3. Harden serial executor fairness (e.g., explicit `proc_yield` interleavings) and document the scheduling guarantees we expect fixtures to rely on.
4. Update spec prose/fixtures once the TypeScript side lands, especially around error payload structure (`ProcError`) and helper semantics.
