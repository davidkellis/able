# Able Language Specification: Functions

This section defines the syntax and semantics for function definition and invocation in **Able**, a hybrid functional and imperative programming language. Functions are designed to be minimal yet expressive, supporting multiple parameters and multi-expression bodies.

## Function Definition

### Syntax
Functions are defined with the following structure:
```
fn Identifier (ParameterList) [: ReturnType] { ExpressionList }
```
- **`fn`**: Keyword that introduces a function definition.
- **`Identifier`**: The function name, a valid identifier (e.g., `add`, `greet`).
- **`ParameterList`**: A comma-separated list of parameters, each defined as:
  - `Identifier [: Type]` (e.g., `a: i32`, `name: string`).
  - May be empty: `()`.
- **`[: ReturnType]`**: An optional return type annotation, written as `-> Type` (e.g., `-> i32`), where `Type` is any valid type expression.
- **`{ ExpressionList }`**: A block containing one or more expressions:
  - **`ExpressionList`**: A sequence of expressions separated by:
    - **Semicolon (`;`)**: Explicit delimiter between expressions on the same line.
    - **Newline**: Implicit delimiter when expressions appear on separate lines.
  - The last expression in the list determines the return value.

### Examples
- Single-expression function:
  ```
  fn add(a: i32, b: i32) -> i32 { a + b }
  ```
- Multi-expression function with semicolons:
  ```
  fn addAndDouble(a: i32, b: i32) -> i32 { c = a + b; c * 2 }
  ```
- Multi-expression function with newlines:
  ```
  fn addAndDouble(a: i32, b: i32) -> i32 {
    c = a + b
    c * 2
  }
  ```
- No parameters:
  ```
  fn zero() -> i32 { 0 }
  ```
- With comments:
  ```
  fn process(x: i32) -> i32 {
    y = x + 1  ## Increment x
    z = y * 2  ## Double the result
    z + 3      ## Add 3 and return
  }
  ```

### Semantics
- **Parameters**: Each parameter in `ParameterList` is bound to an argument value during invocation and scoped to the function body.
- **Execution**: Expressions in `ExpressionList` are evaluated sequentially from left to right (or top to bottom for newlines).
- **Return Value**: The value of the last expression in `ExpressionList` is returned, matching `ReturnType` if specified or inferred if omitted.
- **Type**: A function with parameters of types `T1, T2, ..., Tn` and return type `R` has the type `(T1, T2, ..., Tn) -> R`.
  - Example: `fn add(a: i32, b: i32) -> i32 { a + b }` has type `(i32, i32) -> i32`.
- **Scope**:
  - Parameters and variables defined within the block are local to the function.
  - The function name is introduced into the enclosing scope.
- **Type Checking**:
  - Parameter types must match argument types during invocation.
  - The final expression’s type must match the declared or inferred return type.

## Function Invocation

### Syntax
Functions are invoked with parentheses and a comma-separated argument list:
```
Identifier (ArgumentList)
```
- **`Identifier`**: The name of a defined function.
- **`ArgumentList`**: A comma-separated list of expressions providing argument values (e.g., `5, 3`, `"Alice"`).
  - May be empty: `()`.

### Examples
- `add(5, 3)`
  - Invokes `add`, returns `8`.
- `greet("Alice")`
  - Invokes `greet`, returns `"Hello, Alice"`.
- `zero()`
  - Invokes `zero`, returns `0`.
- `addAndDouble(2, 3)`
  - Invokes `addAndDouble`, returns `10` (i.e., `(2 + 3) * 2`).

### Semantics
- **Argument Passing**: Each argument in `ArgumentList` is evaluated and bound to the corresponding parameter by position.
- **Type Matching**: The type of each argument must match the type of the corresponding parameter (explicitly declared or inferred).
- **Evaluation**: The function’s `ExpressionList` is executed with parameters bound to argument values, and the result of the last expression is returned.

## Integration with Type Expressions
- **Function Types**: Functions are typed as `(TypeList) -> Type`.
  - **`TypeList`**: Comma-separated list of parameter types (e.g., `i32, i32`).
  - **`->`**: Arrow indicating the return type.
  - **`Type`**: The return type (e.g., `i32`, `string`).
- **Examples**:
  - `(i32, i32) -> i32` (type of `add`).
  - `(string) -> string` (type of `greet`).
  - `() -> i32` (type of `zero`).
- **Usage**: Function types can be used in type annotations, e.g., `f: (i32, i32) -> i32 = add`.

## Notes
- **Minimal Design**:
  - No explicit `return` keyword; the last expression implicitly determines the return value.
  - Semicolons and newlines provide flexible expression separation.
- **Expressiveness**: Supports both functional (pure expressions) and imperative (sequential assignments) styles within the body.
- **Comments**: Inline comments with `##` are supported within the block (e.g., `c = a + b  ## Sum`).
