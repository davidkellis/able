# Able Language Specification: Error Handling (Revised with Exceptions and V-Style)

Able provides multiple mechanisms for handling errors and exceptional situations:

1.  **Explicit `return`:** Allows early exit from functions.
2.  **`Result T` and `Option T` (`?Type`) types:** Used with V-lang style propagation (`!`) and handling (`else {}`) for expected errors or absence.
3.  **Exceptions:** For exceptional conditions, using `raise` and `rescue`. Panics are implemented via exceptions.

## 1. Explicit `return` Statement

Functions can return a value before reaching the end of their body using the `return` keyword.

### Syntax
```able
return Expression
return // Equivalent to 'return void' if function returns void
```

-   **`return`**: Keyword initiating an early return.
-   **`Expression`**: Optional expression whose value is returned from the function. Its type must match the function's declared or inferred return type.
-   If `Expression` is omitted, the function must have a `void` return type, and `void` is implicitly returned.

### Semantics
-   Immediately terminates the execution of the current function.
-   The value of `Expression` (or `void`) is returned to the caller.
-   If used within nested blocks (like loops or `do` blocks) inside a function, it still returns from the *function*, not just the inner block.

### Example
```able
fn find_first_negative(items: Array i32) -> ?i32 {
  for item in items {
    if item < 0 {
      return Some i32 { value: item } ## Early return with value
    }
  }
  return nil ## Return nil if no negative found
}

fn process_or_skip(item: i32) -> void {
    if item == 0 {
        log("Skipping zero")
        return ## Early return void
    }
    process_item(item)
}
```

## 2. V-Lang Style Error Handling (`Option`/`Result`, `!`, `else`)

This mechanism is preferred for handling *expected* errors or optional values gracefully without exceptions.

### 2.1. Core Types

-   **`Option T` (`?Type`)**: Represents optional values. Defined implicitly as the union `nil | T`. Used when a value might be absent.
    ```able
    user: ?User = find_user(id) ## find_user returns nil or User
    ```
-   **`Result T`**: Represents the result of an operation that can succeed with a value of type `T` or fail with an error. Defined implicitly as the union `T | Error`.
    ```able
    ## The 'Error' interface (built-in or standard library)
    interface Error for T {
        fn message(self: Self) -> string;
        ## Potentially other methods like cause(), stacktrace()
    }

    ## Result T is implicitly: union Result T = T | Error

    ## Example function signature
    fn read_file(path: string) -> Result string { ... } ## Returns string or Error
    ```

### 2.2. Error/Option Propagation (`!`)

The postfix `!` operator simplifies propagating `nil` from `Option` types or `Error` from `Result` types up the call stack.

#### Syntax
```able
ExpressionReturningOptionOrResult!
```

#### Semantics
-   Applies to an expression whose type is `?T` (`nil | T`) or `Result T` (`T | Error`).
-   If the expression evaluates to the "successful" variant (`T`), the `!` operator unwraps it, and the overall expression evaluates to the unwrapped value (of type `T`).
-   If the expression evaluates to the "failure" variant (`nil` or an `Error`), the `!` operator causes the **current function** to immediately **`return`** that `nil` or `Error` value.
-   **Requirement:** The function containing the `!` operator must itself return a compatible `Option` or `Result` type (or a supertype union) that can accommodate the propagated `nil` or `Error`.

#### Example
```able
## Assuming read_file returns Result string (string | Error)
## Assuming parse_data returns Result Data (Data | Error)
fn load_and_parse(path: string) -> Result Data {
    content = read_file(path)! ## If read_file returns Err, load_and_parse returns it.
                               ## Otherwise, content is string.
    data = parse_data(content)! ## If parse_data returns Err, load_and_parse returns it.
                                ## Otherwise, data is Data.
    return data ## Return the successful Data value (implicitly Ok)
}

## Option example
fn get_nested_value(data: ?Container) -> ?Value {
    container = data! ## If data is nil, return nil from get_nested_value
    inner = container.get_inner()! ## Assuming get_inner returns ?Inner, propagate nil
    value = inner.get_value()! ## Assuming get_value returns ?Value, propagate nil
    return value
}
```

### 2.3. Error/Option Handling (`else {}`)

Provides a way to handle the `nil` or `Error` case of an `Option` or `Result` immediately, typically providing a default value or executing alternative logic.

#### Syntax
```able
ExpressionReturningOptionOrResult else { BlockExpression }
ExpressionReturningOptionOrResult else { |err| BlockExpression } // Capture error
```

#### Semantics
-   Applies to an expression whose type is `?T` (`nil | T`) or `Result T` (`T | Error`).
-   If the expression evaluates to the "successful" variant (`T`), the overall expression evaluates to that unwrapped value (`T`). The `else` block is *not* executed.
-   If the expression evaluates to the "failure" variant (`nil` or an `Error`):
    *   The `BlockExpression` inside the `else { ... }` is executed.
    *   If the form `else { |err| ... }` is used *and* the failure value was an `Error`, the error value is bound to the identifier `err` (or chosen name) within the scope of the `BlockExpression`. If the failure value was `nil`, `err` is not bound or has a `nil`-like value (TBD - let's assume it's only bound for `Error`).
    *   The entire `Expression else { ... }` expression evaluates to the result of the `BlockExpression`.
-   **Type Compatibility:** The type of the "successful" variant (`T`) and the type returned by the `BlockExpression` must be compatible. The overall expression has this common compatible type.

#### Example
```able
## Option handling
config_port: ?i32 = read_port_config()
port = config_port else { 8080 } ## Provide default value if config_port is nil
## port is i32

user_name = find_user(id) else { |err| ## Assuming find_user returns Result User
    log(`Failed to find user: ${err.message()}`)
    "Default User" ## Return default string
}
## user_name is string (compatible with User via Display? Or needs explicit handling?)
## Let's assume return type must be compatible:
user: ?User = find_user(id) else { |err|
    log(`Failed to find user: ${err.message()}`)
    nil ## Return nil if lookup failed
}

## Result Handling without capturing error detail
data = load_data() else {
    log("Loading failed, using empty.")
    [] ## Return empty array
}
```

## 3. Exceptions (`raise` / `rescue`)

For handling truly *exceptional* situations that disrupt normal control flow, often originating from deeper library levels or representing programming errors discovered at runtime. Panics are implemented as a specific kind of exception.

### 3.1. Raising Exceptions (`raise`)

The `raise` keyword throws an exception value, immediately interrupting normal execution and searching up the call stack for a matching `rescue` block.

#### Syntax
```able
raise ExceptionValue
```
-   **`raise`**: Keyword initiating the exception throw.
-   **`ExceptionValue`**: An expression evaluating to the value to be raised. This value should typically implement the standard `Error` interface, but technically any value *could* be raised (TBD - restricting to `Error` types is safer).

#### Example
```able
struct DivideByZeroError {} ## Implement Error interface
impl Error for DivideByZeroError { fn message(self: Self) -> string { "Division by zero" } }

fn divide(a: i32, b: i32) -> i32 {
    if b == 0 {
        raise DivideByZeroError {}
    }
    a / b
}
```

### 3.2. Rescuing Exceptions (`rescue`)

The `rescue` keyword provides a mechanism to catch exceptions raised during the evaluation of an expression. It functions similarly to `match` but operates on exceptions caught during the primary expression's execution.

#### Syntax
```able
MonitoredExpression rescue {
  case Pattern1 [if Guard1] => ResultExpressionList1,
  case Pattern2 [if Guard2] => ResultExpressionList2,
  ...
  [case _ => DefaultResultExpressionList] ## Catches any other exception
}
```

-   **`MonitoredExpression`**: The expression whose execution is monitored for exceptions.
-   **`rescue`**: Keyword initiating the exception handling block.
-   **`{ case PatternX [if GuardX] => ResultExpressionListX, ... }`**: Clauses similar to `match`.
    *   Execution starts by evaluating `MonitoredExpression`.
    *   **If No Exception:** The `rescue` block is skipped. The entire `rescue` expression evaluates to the normal result of `MonitoredExpression`.
    *   **If Exception Raised:** Execution of `MonitoredExpression` stops. The *raised exception value* becomes the subject matched against the `PatternX` in the `case` clauses sequentially.
    *   The first clause whose `PatternX` matches the exception value (and whose optional `GuardX` passes) is chosen.
    *   The corresponding `ResultExpressionListX` is executed. Its result becomes the value of the entire `rescue` expression.
    *   If no pattern matches the raised exception, the exception continues propagating up the call stack. A final `case _ => ...` can catch any otherwise unhandled exception within this `rescue`.
-   **Type Compatibility:** The normal result type of `MonitoredExpression` must be compatible with the result types of all `ResultExpressionListX` in the `rescue` block.

#### Example
```able
result = do {
            divide(10, 0) ## This will raise DivideByZeroError
         } rescue {
            case e: DivideByZeroError => {
                log("Caught division by zero!")
                0 ## Return 0 as the result of the rescue expression
            },
            case e: Error => { ## Catch any other Error
                log(`Caught other error: ${e.message()}`)
                -1
            },
            ## No final 'case _', other exceptions would propagate
         }
## result is 0

value = risky_operation() rescue { case _ => default_value } ## Provide default on any error
```

### 3.3. Panics

Panics are implemented as a specific, severe type of exception (e.g., `PanicError` implementing `Error`). Calling the built-in `panic(message)` function is equivalent to `raise PanicError { message: message }`. Typically, `rescue` blocks should *avoid* catching `PanicError` unless performing essential cleanup before potentially re-raising or terminating, as panics signal unrecoverable states or bugs. The default top-level program handler usually catches unhandled panics, prints details, and terminates.
