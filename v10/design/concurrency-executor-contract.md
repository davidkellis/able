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
  - `proc_pending_tasks` surfaces `executor.pendingTasks` as an Able helper so
    fixtures/tests can assert that cooperative queues drain. Pre-emptive
    executors may return `0` (or a best-effort count of outstanding tasks) when
    their host runtime does not expose runnable-queue details; programs must not
    rely on the value for correctness.
  - Interpreters are expected to charge cooperative “time slices” at safe evaluation
    boundaries (statement iterations, loop bodies, pattern matches). Once a task
    reaches the configured `schedulerMaxSteps`, it should raise the shared yield
    signal, persist its evaluation state, and reschedule itself so long-running
    procs make forward progress even without explicit `proc_yield` calls. Manual
    `proc_yield` invocations must remain supported and should not re-run already
    completed statements when resumed.

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
  - The evaluator records continuation state for blocks, loops, and match expressions.
    When a `ProcYieldSignal` is raised (either by `proc_yield` or because
    `checkTimeSlice()` saw `schedulerMaxSteps` ticks), the interpreter snapshots
    the current node state, unwinds, and later resumes from the recorded point.
    This continuation layer is what lets a single-threaded JS host emulate Go-like
    pre-emption without user code changes.

- **Go (`interpreter10-go/`).**
  - **Production executor.** `GoroutineExecutor` launches each Able `proc`/`spawn`
    as a native goroutine. Go’s scheduler handles suspension, resumption, and
    fair progress automatically; we do not maintain an explicit continuation layer.
    The interpreter simply exposes the same Proc/Future API (`status`, `value`,
    `cancel`), delegating real concurrency to the host runtime.
  - **Deterministic test executor.** `SerialExecutor` keeps the same `Executor`
    surface for parity tests. It queues goroutine entry functions and drives them
    in FIFO order to make interleavings predictable. Importantly, each task still
    runs inside a goroutine, so stack/state preservation is provided by Go itself—
    no AST-level continuation bookkeeping is required.
  - The parity harness swaps between executors: `SerialExecutor` for deterministic
    unit tests, `GoroutineExecutor` for integration scenarios that should run under
    true concurrency. In both cases Able programs see the same observable semantics
    as the TypeScript interpreter.

## Testing guidance

- **Fixtures (`fixtures/ast/concurrency`).** These scenarios rely on the shared
  executor contract. Both runtimes must execute the fixtures with their native
  executor implementations, ensuring cancellation, memoisation, and helper
  semantics match.
- **Unit tests.** Test suites should interact with the executor through the
  public contract (`flush`, optional `pendingTasks`) rather than mutating
  interpreter internals. This keeps cross-runtime behaviour aligned.

## Scheduler tuning (`schedulerMaxSteps`)

The cooperative TypeScript executor (and the Go serial executor that mimics it
in parity tests) charges “time slices” as the interpreter evaluates loop bodies,
pattern matches, and statement boundaries. Once the interpreter accrues
`schedulerMaxSteps` slices it raises a `ProcYieldSignal`, unwinds, and reschedules
the task so that other queued work can run. The default budget is **1024**
steps per flush; this value balances throughput with fairness for the current
fixture corpus.

- **Configuring the budget.** `new InterpreterV10({ schedulerMaxSteps: N })`
  overrides the default for TypeScript runs (CLI tools expose the same option
  when they construct interpreters). Lower values make fairness fixtures
  (e.g. `concurrency/fairness_proc_round_robin`) more sensitive by forcing the
  cooperative executor to yield more often. Higher values are useful for
  benchmark runs that would otherwise spend too much time snapshotting continuations.
- **Go executors.** The production `GoroutineExecutor` ignores this budget
  entirely—the host scheduler handles pre-emption. The deterministic
  `SerialExecutor` uses the same cooperative contract when it replays fixtures,
  so lowering `schedulerMaxSteps` in TS tests often surfaces bugs that the Go
  parity suite would also catch.
- **Symptoms & guidance.**
  - If long-running Able code appears to “stick” under the cooperative executor,
    lower the budget (e.g. 256) to confirm time slicing is still happening, then
    inspect the new fairness fixtures for additional hints.
  - If benchmarks regress because the interpreter spends too much time yielding,
    raise the budget in controlled increments; once the fairness fixtures no
    longer report progress (or `proc_yield_flush` hangs) the budget is too high.

Keep the default at 1024 unless a workload justifies a different balance. Any
changes that ship to other contributors should include updated fixture coverage
or design notes so the cooperative and goroutine executors stay aligned.

## Follow-ups

- Document `proc_yield`/`proc_flush` guarantees in the v10 specification once
  wording is final.
- Investigate fairness fixtures that rely on specific scheduling orders; any
  additional coordination helpers must be added to both executors (or guarded by
  optional capabilities) before enabling those tests across runtimes.
