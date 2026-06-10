# Go Reference Concurrency Guide

Date: 2026-07-14
Status: Active contributor guide

> **Normative contract:** [spec §12](../../spec/full_spec_v12.md) defines the
> Able-language behavior of `spawn`, `Future`, `await`, channels, mutexes, and
> `future_*` helpers. This guide routes contributors through the active Go
> implementation. [go-concurrency-scheduler.md](go-concurrency-scheduler.md)
> records the detailed executor behavior and implementation rationale.

## Scope and defaults

The Go tree-walker and bytecode VM are the active reference implementations.
`New()` and `NewBytecode()` both use `SerialExecutor` by default, which runs a
FIFO task queue on one worker goroutine and gives fixtures reproducible
ordering. The default is also selected when `ABLE_EXECUTOR` is unset.

Use `NewWithExecutor(NewGoroutineExecutor(nil))`,
`NewBytecodeWithExecutor(NewGoroutineExecutor(nil))`, or
`ABLE_EXECUTOR=goroutine` only when a caller needs true parallel execution.
`future_yield()` is then an advisory Go scheduler yield, not a promise of a
particular task order. The environment value is restricted to `serial` and
`goroutine`.

The historical TypeScript/cooperative-continuation proposal is retained in
[concurrency-executor-contract.md](concurrency-executor-contract.md). It does
not define behavior for the active workspace.

## Observable behavior contributors must preserve

- `spawn` returns a `Future` with `Pending`, `Resolved`, `Cancelled`, or
  `Failed` status. A terminal outcome is memoized; cancellation is best-effort
  and idempotent.
- `future_yield()`, `future_cancelled()`, `future_flush()`, and
  `future_pending_tasks()` take zero arguments. The first two require an
  asynchronous task context; `future_pending_tasks()` is diagnostic only.
- A serial `future_flush()` drains runnable work but does not resume work that
  is genuinely await-blocked. A goroutine flush returns once all remaining
  tasks are blocked, preventing a blocked channel or I/O task from hanging the
  caller indefinitely.
- The runtime protects its own environments and task bookkeeping. Able code
  that mutates shared user data under the goroutine executor still needs the
  language-level channel or mutex synchronization that its semantics require.
- Tree-walker, bytecode, and compiled execution must remain observably
  aligned. Do not introduce an executor-specific language rule merely to make
  a scheduling test or benchmark faster.

## Changing concurrency behavior

Start with §12 and update it when the public rule changes. Then add focused
coverage for tree-walker and bytecode execution; keep compiler parity green
when lowering or runtime support changes. The primary coverage is in
`interpreter_concurrency*_test.go`, `interpreter_await_test.go`,
`interpreter_channels_mutex_test.go`, bytecode async tests, and the
`concurrency/` executable fixtures.

Scheduler tuning is not a standalone benchmark project. It needs the same
cross-application evidence as other VM work: a material shared hot leaf in
unlike workloads, a generally applicable candidate, and the full benchmark
screen before it is retained. See the detailed scheduler record for current
implementation boundaries and test expectations.
