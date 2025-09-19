# Able Language Specification: Concurrency

Able provides lightweight concurrency primitives inspired by Go, allowing asynchronous execution of functions and blocks using the `proc` and `spawn` keywords.

## 1. Concurrency Model Overview

-   Able supports concurrent execution, allowing multiple tasks to run seemingly in parallel.
-   The underlying mechanism (e.g., OS threads, green threads, thread pool, event loop) is implementation-defined but guarantees the potential for concurrent progress.
-   Communication and synchronization between concurrent tasks (e.g., channels, mutexes) are not defined in this section but would typically be provided by the standard library.

## 2. Asynchronous Execution (`proc`)

The `proc` keyword initiates asynchronous execution of a function call or a block, returning a handle (`Proc T`) to manage the asynchronous process.

### Syntax

```able
proc FunctionCall
proc BlockExpression
```

-   **`proc`**: Keyword initiating asynchronous execution.
-   **`FunctionCall`**: A standard function or method call expression (e.g., `my_function(arg1)`, `instance.method()`).
-   **`BlockExpression`**: A `do { ... }` block expression.

### Semantics

1.  **Asynchronous Start**: The target `FunctionCall` or `BlockExpression` begins execution asynchronously, potentially on a different thread or logical task. The current thread does *not* block.
2.  **Return Value**: The `proc` expression immediately returns a value whose type implements the `Proc T` interface.
    -   `T` is the return type of the `FunctionCall` or the type of the value the `BlockExpression` evaluates to.
    -   If the function/block returns `void`, the return type is `Proc void`.
3.  **Independent Execution**: The asynchronous task runs independently until it completes, fails, or is cancelled.

### Example

```able
fn fetch_data(url: string) -> string {
  ## ... perform network request ...
  "Data from {url}"
}

proc_handle = proc fetch_data("http://example.com") ## Starts fetching data
## proc_handle has type `Proc string`

computation_handle = proc do {
  x = compute_part1()
  y = compute_part2()
  x + y ## Block evaluates to the sum
} ## computation_handle has type `Proc i32` (assuming sum is i32)

side_effect_proc = proc { log_message("Starting background task...") } ## Returns Proc void
```

## 3. Process Handle (`Proc T` Interface)

The `Proc T` interface provides methods to interact with an ongoing asynchronous process started by `proc`.

### Definition (Conceptual)

```able
## Represents the status of an asynchronous process
union ProcStatus = Pending | Resolved | Cancelled | Failed ProcError

## Represents an error occurring during process execution (details TBD)
## Could wrap panic information or specific error types.
struct ProcError { details: string } ## Example structure

## Interface for interacting with a process handle
interface Proc T for HandleType { ## HandleType is the concrete type returned by 'proc'
  ## Get the current status of the process
  fn status(self: Self) -> ProcStatus;

  ## Attempt to retrieve the result value.
  ## Blocks the *calling* thread until the process status is Resolved, Failed, or Cancelled.
  ## Returns Ok(T) on success, Err(ProcError) on failure or if cancelled before resolution.
  fn get_value(self: Self) -> Result T ProcError;

  ## Request cancellation of the asynchronous process.
  ## This is a non-blocking request. Cancellation is cooperative; the async task
  ## must potentially check for cancellation requests to terminate early.
  ## Guarantees no specific timing or success, only signals intent.
  fn cancel(self: Self) -> void;
}
```

### Semantics of Methods

-   **`status()`**: Returns the current state (`Pending`, `Resolved`, `Cancelled`, `Failed`) without blocking.
-   **`get_value()`**: Blocks the caller until the process finishes (resolves, fails, or is definitively cancelled).
    -   If `Resolved`, returns `Ok(value)` where `value` has type `T`. For `Proc void`, returns `Ok(void)` conceptually (or `Ok(nil)` if `void` cannot be wrapped? TBD). Let's assume `Ok(value)` works even for `T=void`.
    -   If `Failed`, returns `Err(ProcError)` containing error details.
    -   If `Cancelled`, returns `Err(ProcError)` indicating cancellation.
-   **`cancel()`**: Sends a cancellation signal to the asynchronous task. The task is not guaranteed to stop immediately or at all unless designed to check for cancellation signals.

### Example Usage

```able
data_proc: Proc string = proc fetch_data("http://example.com")

## Check status without blocking
current_status = data_proc.status()
if match current_status { case Pending => true } { print("Still working...") }

## Block until done and get result (handle potential errors)
result = data_proc.get_value()
final_data = result match {
  case Ok { value: d } => `Success: ${d}`,
  case Err { error: e } => `Failed: ${e.details}`
}
print(final_data)

## Request cancellation (fire and forget)
data_proc.cancel()
```

## 4. Thunk-Based Asynchronous Execution (`spawn`)

The `spawn` keyword also initiates asynchronous execution but returns a `Thunk T` value, which implicitly blocks and yields the result when evaluated.

### Syntax

```able
spawn FunctionCall
spawn BlockExpression
```

-   **`spawn`**: Keyword initiating thunk-based asynchronous execution.
-   **`FunctionCall` / `BlockExpression`**: Same as for `proc`.

### Semantics

1.  **Asynchronous Start**: Starts the function or block execution asynchronously, similar to `proc`. The current thread does not block.
2.  **Return Value**: Immediately returns a value of the special built-in type `Thunk T`.
    -   `T` is the return type of the function or the evaluation type of the block.
    *   If the function/block returns `void`, the return type is `Thunk void`.
3.  **Implicit Blocking Evaluation**: The core feature of `Thunk T` is its evaluation behavior. When a value of type `Thunk T` is used in a context requiring a value of type `T` (e.g., assignment, passing as argument, part of an expression), the current thread **blocks** until the associated asynchronous computation completes.
    *   If the computation completes successfully with value `v` (type `T`), the evaluation of the `Thunk T` yields `v`.
    *   If the computation fails (e.g., panics), that panic is **propagated** to the thread evaluating the `Thunk T`.
    *   Evaluating a `Thunk void` blocks until completion and then yields `void` (effectively synchronizing).

### Example

```able
fn expensive_calc(n: i32) -> i32 {
  ## ... time-consuming work ...
  n * n
}

thunk_result: Thunk i32 = spawn expensive_calc(10)
thunk_void: Thunk void = spawn { log_message("Background log started...") }

print("Spawned tasks...") ## Executes immediately

## Evaluation blocks here until expensive_calc(10) finishes:
final_value = thunk_result
print(`Calculation result: ${final_value}`) ## Prints "Calculation result: 100"

## Evaluation blocks here until the logging block finishes:
_ = thunk_void ## Assigning to _ forces evaluation/synchronization
print("Background log finished.")
```

## 5. Key Differences (`proc` vs `spawn`)

-   **Return Type:** `proc` returns `Proc T` (an interface handle); `spawn` returns `Thunk T` (a special type).
-   **Control:** `Proc T` offers explicit control (check status, attempt cancellation, get result via method call potentially handling errors).
-   **Result Access:** `Thunk T` provides implicit result access; evaluating the thunk blocks and returns the value directly (or propagates panic). It lacks fine-grained status checks or cancellation via the handle itself.
-   **Use Cases:**
    *   `proc` is suitable when you need to manage the lifecycle of the async task, check its progress, handle failures explicitly, or potentially cancel it.
    *   `spawn` is simpler for "fire and forget" tasks where you only need the final result eventually and are okay with blocking for it implicitly (or propagating panics).
