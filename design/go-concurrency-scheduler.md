# Go Concurrency Scheduler Design

This document outlines the design of the concurrency scheduler for the Go interpreter, which will implement the `proc` and `spawn` semantics defined in the Able v10 specification.

## 1. Goals

- Implement a scheduler that uses Go's native concurrency primitives (goroutines and channels) to execute Able's `proc` and `spawn` expressions.
- Ensure the observable behavior of the Go interpreter's concurrency model is identical to the TypeScript interpreter's cooperative scheduler, as per the project's principles.
- Provide a deterministic execution order for tests.

## 2. Core Components

### 2.1. Scheduler

The `Scheduler` will be a new component in the Go interpreter responsible for managing and executing asynchronous tasks. It will have the following responsibilities:

-   **Task Queue:** Maintain a queue of tasks to be executed.
-   **Worker Goroutines:** A pool of worker goroutines that will execute tasks from the queue.
-   **Task Management:** Add new tasks to the queue, and manage their lifecycle.

### 2.2. Task

A `Task` will represent a unit of work to be executed by the scheduler. It will contain:

-   **Expression:** The `ProcExpression` or `SpawnExpression` to be evaluated.
-   **Environment:** The environment in which the expression should be evaluated.
-   **Handle:** A `ProcHandleValue` or `FutureHandleValue` to store the result of the computation.

### 2.3. ProcHandleValue and FutureHandleValue

The existing `ProcHandleValue` will be used for `proc` expressions. A new `FutureHandleValue` will be created for `spawn` expressions, which will include memoization logic.

## 3. Execution Flow

1.  When the interpreter encounters a `ProcExpression` or `SpawnExpression`, it will create a new `Task` and add it to the `Scheduler`'s task queue.
2.  The `Scheduler`'s worker goroutines will pick up tasks from the queue and execute them.
3.  The worker will evaluate the expression in the task's environment.
4.  Upon completion, the worker will update the corresponding `ProcHandleValue` or `FutureHandleValue` with the result.

## 4. Cooperative Yielding and Cancellation

To ensure parity with the TypeScript interpreter, the Go scheduler will support cooperative yielding and cancellation:

-   **`proc_yield()`:** When a task calls `proc_yield()`, the worker will requeue the task and pick up a new one from the queue. This can be implemented using a channel that the worker will block on, and the `proc_yield` function will send a value to the channel.
-   **`proc_cancelled()`:** The `ProcHandleValue` will have a `cancelRequested` flag. The `proc_cancelled()` function will check this flag.

## 5. Deterministic Testing

To ensure deterministic testing, the scheduler will be configurable to use a single worker goroutine and a predictable task queueing mechanism during tests.

## 6. Next Steps

1.  Implement the `Scheduler` component.
2.  Implement the `Task` struct.
3.  Create the `FutureHandleValue`.
4.  Integrate the `Scheduler` with the interpreter's evaluation of `ProcExpression` and `SpawnExpression`.
5.  Implement `proc_yield()` and `proc_cancelled()` native functions.
6.  Add comprehensive tests to verify the scheduler's functionality and parity with the TypeScript interpreter.
