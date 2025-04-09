# Able Language Specification: Operators

This section defines the standard operators available in Able, their syntax, semantics, precedence, and associativity, closely following Rust's precedence model but using `~` for bitwise NOT.

## 1. Operator Precedence and Associativity

Operators are evaluated in a specific order determined by precedence (higher binds tighter) and associativity (order for operators of the same precedence).

| Precedence | Operator(s)           | Description                             | Associativity | Notes                                                     |
| :--------- | :-------------------- | :-------------------------------------- | :------------ | :-------------------------------------------------------- |
| 15         | `.`                   | Member Access (fields, methods, UFCS)   | Left-to-right |                                                           |
| 14         | `()`                  | Function/Method Call                    | Left-to-right |                                                           |
| 14         | `[]`                  | Indexing                                | Left-to-right |                                                           |
| 13         | `-` (unary)           | Arithmetic Negation                     | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 13         | `!` (unary)           | **Logical NOT**                         | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 13         | `~` (unary)           | **Bitwise NOT (Complement)**            | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 12         | `*`, `/`, `%`         | Multiplication, Division, Remainder     | Left-to-right |                                                           |
| 11         | `+`, `-` (binary)     | Addition, Subtraction                   | Left-to-right |                                                           |
| 10         | `<<`, `>>`            | Left Shift, Right Shift                 | Left-to-right |                                                           |
| 9          | `&` (binary)          | Bitwise AND                             | Left-to-right |                                                           |
| 8          | `^`                   | Bitwise XOR                             | Left-to-right |                                                           |
| 7          | `|` (binary)          | Bitwise OR                              | Left-to-right |                                                           |
| 6          | `>`, `<`, `>=`, `<=`    | Ordering Comparisons                    | Non-assoc     | Chaining requires explicit grouping (`(a<b) && (b<c)`) |
| 5          | `==`, `!=`            | Equality, Inequality Comparisons        | Non-assoc     | Chaining requires explicit grouping                     |
| 4          | `&&`                  | Logical AND (short-circuiting)          | Left-to-right |                                                           |
| 3          | `||`                  | Logical OR (short-circuiting)           | Left-to-right |                                                           |
| 2          | `..`, `...`           | Range Creation (inclusive, exclusive)   | Non-assoc     |                                                           |
| 1          | `=`                   | Simple Assignment                       | Right-to-left |                                                           |
| 1          | `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=` | Compound Assignment (TBD)               | Right-to-left | (Needs formal definition)                                 |
| 0          | `\|>`                 | Pipe Forward                            | Left-to-right | (Lowest precedence)                                       |

*(Note: Precedence levels are relative; specific numerical values may vary but the order shown is based on Rust.)*

## 2. Operator Semantics

*   **Arithmetic (`+`, `-`, `*`, `/`, `%`):** Standard math operations on numeric types.
*   **Comparison (`>`, `<`, `>=`, `<=`, `==`, `!=`):** Compare values, result `bool`. Equality/ordering behavior relies on `Equatable`/`Comparable` interface concepts (TBD).
*   **Logical (`&&`, `||`, `!`):**
    *   `&&` (Logical AND): Short-circuiting on `bool` values.
    *   `||` (Logical OR): Short-circuiting on `bool` values.
    *   `!` (Logical NOT): Unary operator, negates a `bool` value.
*   **Bitwise (`&`, `|`, `^`, `<<`, `>>`, `~`):**
    *   `&`, `|`, `^`: Standard bitwise AND, OR, XOR on integer types (`i*`, `u*`).
    *   `<<`, `>>`: Bitwise left shift, right shift on integer types.
    *   `~` (Bitwise NOT): Unary operator, performs bitwise complement on integer types.
*   **Unary (`-`):** Arithmetic negation for numeric types.
*   **Member Access (`.`):** Access fields/methods, UFCS, static methods.
*   **Function Call (`()`):** Invokes functions/methods.
*   **Indexing (`[]`):** Access elements within indexable collections (e.g., `Array`). Relies on `Indexable`/`IndexableMut` interface concepts (TBD).
*   **Range (`..`, `...`):** Create `Range` objects (inclusive `..`, exclusive `...`).
*   **Assignment (`=`):** Binds RHS value to LHS pattern. Evaluates to the RHS value.
*   **Compound Assignment (`+=`, etc. TBD):** Shorthand (e.g., `a += b` is like `a = a + b`). Need formal definition.
*   **Pipe Forward (`|>`):** `x |> f` evaluates to `f(x)`.

## 3. Overloading (Via Interfaces)

Behavior for non-primitive types relies on implementing standard library interfaces (e.g., `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Not` (for bitwise ~), `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`, `PartialEq`, `Eq`, `PartialOrd`, `Ord`, `Index`, `IndexMut`). These interfaces need definition. Note that logical `!` is typically not overloaded.
