# Serial executor Flush handoff correction retained

Date: 2026-07-27

## Decision

Retain a general scheduler correctness correction in the v12 Go
`SerialExecutor`.

`Flush()` now accounts for the interval after the background worker removes a
runnable task from the queue and before that task is published as active. The
accounting ends atomically when the task becomes active, so blocked channel,
mutex, timer, and await work keeps the existing rule that Flush may return
when no runnable work remains.

This is a scheduler invariant. It is not a benchmark optimization, an
application rule, a named-container special case, or a change to the Able
language contract.

## Trigger

The runtime release-boundary review added race-enabled focused verification:

```text
go test -race ./pkg/interpreter \
  -run '^Test(InterpreterConcurrency|SerialExecutor|GoroutineExecutor|Future|Await|InterpreterChannels|Bytecode.*Concurrent|Bytecode.*Async)' \
  -count=1 -timeout 55s
```

The lane exposed an intermittent semantic failure in
`TestAwaitExpressionManualWaker`: the expected resumed value was `42`, but
the observer saw the pre-resume value `0`. The race detector did not report an
unsynchronized memory access; race instrumentation widened a scheduler
ordering window.

## Root cause

The serial worker previously performed these steps separately:

1. lock the executor and remove the next task from `queue`;
2. unlock;
3. enter `runSerialTask`;
4. lock again and publish the task as `active`.

An external `Flush()` could acquire the lock between steps 2 and 4. At that
instant the queue was empty and no task was active, so Flush returned even
though runnable work had already been claimed by the worker.

## Retained implementation

The executor now increments `workerInFlight` while dequeuing under the lock.
`Flush()` treats that marker as runnable work. When the worker publishes the
task as active, the same locked transition clears the marker and leaves the
existing `active` state responsible for the task.

The marker deliberately does not remain set for the task's whole execution.
An intermediate version did that and the broad interpreter corpus stopped at
`concurrency/mutex_contention/serial`: a task paused on a mutex was incorrectly
treated as runnable worker work. That version was not retained. The focused
blocked-task fixture and complete package pass prove the final handoff is
narrow rather than a change to blocked-Flush semantics.

`TestSerialExecutorFlushWaitsForDequeuedTask` deterministically removes a task
from the queue, starts Flush during the otherwise invisible interval, proves
that Flush waits, and releases the handoff. It does not depend on timing the
real worker goroutine.

## Verification

The retained final state passed:

- the deterministic dequeue-handoff guard 50 consecutive times;
- `concurrency/mutex_contention/serial` 10 consecutive times;
- the focused await/fairness group 10 consecutive times;
- the two-test race reproducer 20 consecutive times;
- the complete scheduler, Future, await, channel, and async bytecode race
  lane;
- the full short interpreter package in 75.624 seconds; and
- `go vet`, `gofmt`, and `git diff --check` for the reviewed runtime boundary.

The runtime boundary review and required complete handoff provide the broader
verification result.

## Scope

No runtime carrier, bytecode primitive lane, Array representation, compiler
lowering, standard-library source, dependency, benchmark, reference
implementation, fixture expectation, or WASM path changed. The correction
does not invalidate the closed performance-owner censuses because it changes
only when Flush observes a claimed task, not any hot execution carrier or
compiled/interpreted boundary.
