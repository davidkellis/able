# Concurrency Executor Contract

Date: 2025-10-23  
Owners: Able Agents

## Purpose

Able v10 runtimes now share a minimal executor abstraction that drives `proc` and
`spawn` evaluation. This note documents the contract so TypeScript, Go, and any
future runtimes implement compatible semantics while retaining freedom over the
underlying scheduling strategy.

## High-level requirements

- **Executor surface.** Runtimes expose an executor with the methods below. The
  interface is intentionally small so alternative schedulers (goroutines,
  cooperative loops, thread pools) plug in without touching the interpreter’s
  evaluation code.

  ```ts
  type ExecutorTask = () => void;

  interface Executor {
    schedule(task: ExecutorTask): void;
    ensureTick(): void;
    flush(limit?: number): void;
    pendingTasks?(): number;
  }
  ```

  - `schedule` queues work for asynchronous execution.
  - `ensureTick` guarantees queued tasks will run “soon” (microtask/timer or
    immediate dispatch depending on the host).
  - `flush` progresses the queue synchronously up to an optional `limit`. This
    powers `proc_flush`, deterministic testing, and fixture harnesses.
  - `pendingTasks` is optional and only used by test helpers to assert the queue
    drained. Runtimes that cannot cheaply expose this count should omit it and
    adjust their tests accordingly.

- **Interpreter integration.** `InterpreterV10` takes an optional `executor`
  when constructed. If none is provided, it instantiates a cooperative executor
  (`CooperativeExecutor` in TypeScript, `SerialExecutor` in Go tests) with a
  shared default `maxSteps` of 1024 per flush. Production Go builds continue to
  use the goroutine executor; TypeScript relies on the cooperative implementation.

- **Helper semantics.**
  - `proc_yield` must yield control to the executor without completing the task,
    allowing other queued work to run.
  - `proc_cancelled` inspects the async payload set by the executor and raises
    when called outside a proc/future context.
  - `proc_flush` delegates to `executor.flush(limit)` to advance the queue
    deterministically.

- **Determinism expectations.** Fixtures and unit tests assume that calling
  `flush()` without a limit drains at most `schedulerMaxSteps` tasks (1024 by
  default) and that `ensureTick` guarantees eventual progress even if no flush
  occurs. Implementations should avoid unbounded recursion or starvation.

## Runtime mappings

- **TypeScript (`interpreter10/`).**
  - `CooperativeExecutor` maintains the existing FIFO queue, wrapping the former
    `schedulerQueue`/`processScheduler` logic. The interpreter delegates all
    scheduling to this executor.
  - Tests and harnesses drain work via `executor.flush()` rather than peeking
    at interpreter internals.
  - Fixtures automatically pick up the updated behaviour; no additional wiring
    is required.

- **Go (`interpreter10-go/`).**
  - `SerialExecutor` implements the same contract for deterministic tests.
  - `GoroutineExecutor` ignores `ensureTick` and `flush` (no-op), providing a
    “real” asynchronous executor for production runs while still satisfying the
    interface.
  - The parity harness now swaps in `GoroutineExecutor` for the cancellation and
    memoisation fixtures to validate behaviour under true concurrency.

## Testing guidance

- **Fixtures (`fixtures/ast/concurrency`).** These scenarios rely on the shared
  executor contract. Both runtimes must execute the fixtures with their native
  executor implementations, ensuring cancellation, memoisation, and helper
  semantics match.
- **Unit tests.** Test suites should interact with the executor through the
  public contract (`flush`, optional `pendingTasks`) rather than mutating
  interpreter internals. This keeps cross-runtime behaviour aligned.

## Follow-ups

- Document `proc_yield`/`proc_flush` guarantees in the v10 specification once
  wording is final.
- Investigate fairness fixtures that rely on specific scheduling orders; any
  additional coordination helpers must be added to both executors (or guarded by
  optional capabilities) before enabling those tests across runtimes.
