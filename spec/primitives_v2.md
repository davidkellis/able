# Able Language Specification: Primitive Types and Values

This section defines the built-in primitive types in Able, which form the basic building blocks for data representation.

## 1. Overview

Primitive types represent fundamental data categories like numbers, text characters, logical values, the concept of absence (`nil`), and the concept of *no value* (`void`). Values of primitive types (except `void`) are generally considered immutable (e.g., the number `5` cannot itself be changed into `6`), although bindings to them can be reassigned if the binding is mutable.

## 2. Primitive Type Definitions

| Type     | Description                                   | Literal Examples                    | Notes                                           |
| :------- | :-------------------------------------------- | :---------------------------------- | :---------------------------------------------- |
| `i8`     | 8-bit signed integer (-128 to 127)            | `-128`, `0`, `10`, `127`            |                                                 |
| `i16`    | 16-bit signed integer (-32,768 to 32,767)       | `-32768`, `1000`, `32767`           |                                                 |
| `i32`    | 32-bit signed integer (-2³¹ to 2³¹-1)           | `-2_147_483_648`, `0`, `42`         | Default type for integer literals (TBC).        |
| `i64`    | 64-bit signed integer (-2⁶³ to 2⁶³-1)           | `-9_223_...`, `1_000_000_000`       | Underscores `_` allowed for readability.        |
| `i128`   | 128-bit signed integer (-2¹²⁷ to 2¹²⁷-1)      | `-170_...`, `0`, `170_...`          |                                                 |
| `u8`     | 8-bit unsigned integer (0 to 255)             | `0`, `10`, `255`                  |                                                 |
| `u16`    | 16-bit unsigned integer (0 to 65,535)           | `0`, `1000`, `65535`              |                                                 |
| `u32`    | 32-bit unsigned integer (0 to 2³²-1)            | `0`, `4_294_967_295`              | Underscores `_` allowed for readability.        |
| `u64`    | 64-bit unsigned integer (0 to 2⁶⁴-1)            | `0`, `18_446_...`                 |                                                 |
| `u128`   | 128-bit unsigned integer (0 to 2¹²⁸-1)          | `0`, `340_...`                    |                                                 |
| `f32`    | 32-bit float (IEEE 754 single-precision)      | `3.14f`, `-0.5f`, `1e-10f`, `2.0f`  | Suffix `f` distinguishes from `f64`.              |
| `f64`    | 64-bit float (IEEE 754 double-precision)      | `3.14159`, `-0.001`, `1e-10`, `2.0`  | Default type for float literals (TBC).          |
| `string` | Immutable sequence of Unicode chars (UTF-8) | `"hello"`, `""`, `` `interp ${val}` `` | Double quotes or backticks (interpolation).      |
| `bool`   | Boolean logical values                        | `true`, `false`                     |                                                 |
| `char`   | Single Unicode scalar value (UTF-32)        | `'a'`, `'π'`, `'💡'`, `'\n'`, `'\u{1F604}'` | Single quotes. Supports escape sequences.       |
| `nil`    | Singleton type representing **absence of data**. | `nil`                               | **Type and value are both `nil` (lowercase)**. Often used with `?Type`. |
| `void`   | Type with **no values** (empty set).          | *(No literal value)*                | Represents computations completing without data. |

## 3. Literals

Literals are the source code representation of fixed values.

### 3.1. Integer Literals

-   **Syntax:** A sequence of digits `0-9`. Underscores `_` can be included anywhere except the start/end for readability and are ignored. Prefixes like `0x` (hex), `0o` (octal), `0b` (binary) may be supported (TBD).
-   **Type:** By default, integer literals are inferred as `i32` (this default is configurable/TBC). Type suffixes can explicitly specify the type (e.g., `100i64`, `255u8`, `0i128`).
-   **Examples:** `123`, `0`, `1_000_000`, `42i64`, `255u8`.

### 3.2. Floating-Point Literals

-   **Syntax:** Include a decimal point (`.`) or use scientific notation (`e` or `E`). Underscores `_` are allowed for readability.
-   **Type:** By default, float literals are inferred as `f64`. The suffix `f` explicitly denotes `f32`.
-   **Examples:** `3.14`, `0.0`, `-123.456`, `1e10`, `6.022e23`, `2.718f`, `_1.618_`, `1_000.0`.

### 3.3. Boolean Literals

-   **Syntax:** `true`, `false`.
-   **Type:** `bool`.

### 3.4. Character Literals

-   **Syntax:** A single Unicode character enclosed in single quotes `'`. Special characters can be represented using escape sequences:
    *   Common escapes: `\n` (newline), `\r` (carriage return), `\t` (tab), `\\` (backslash), `\'` (single quote), `\"` (double quote - though not strictly needed in char literal).
    *   Unicode escape: `\u{XXXXXX}` where `XXXXXX` are 1-6 hexadecimal digits representing the Unicode code point.
-   **Type:** `char`.
-   **Examples:** `'a'`, `' '`, `'%'`, `'\n'`, `'\u{1F604}'`.

### 3.5. String Literals

-   **Syntax:**
    1.  **Standard:** Sequence of characters enclosed in double quotes `"`. Supports the same escape sequences as character literals.
    2.  **Interpolated:** Sequence of characters enclosed in backticks `` ` ``. Can embed expressions using `${Expression}`. Escapes like `` \` `` and `\$` are used for literal backticks or dollar signs before braces.
-   **Type:** `string`. Strings are immutable.
-   **Examples:** `"Hello, world!\n"`, `""`, `` `User: ${user.name}, Age: ${user.age}` ``, `` `Literal: \` or \${` ``.

### 3.6. Nil Literal

-   **Syntax:** `nil`.
-   **Type:** `nil`. The type `nil` has only one value, also written `nil`.
-   **Usage:** Represents the absence of meaningful data. Often used with the `?Type` (equivalent to `nil | Type`) union shorthand. `nil` itself *only* has type `nil`, but can be assigned to variables of type `?SomeType`.

### 3.7. Void Type (No Literal)

-   **Type Name:** `void`.
-   **Values:** The `void` type represents the empty set; it has **no values**.
-   **Usage:** Primarily used as a return type for functions that perform actions (side effects) but do not produce any resulting data. It signifies successful completion without a value.
-   **Distinction from `nil`:** The type `nil` has one value (`nil`); the type `void` has zero values.

## 4. Semantics

-   **Immutability:** Primitive values themselves are immutable.
-   **Type Inference:** The compiler infers the type of literals or expressions. `void` is inferred for expressions/functions that produce no value.
-   **Usage:** Primitives are fundamental. `void` signals completion without data. `nil` signals potential absence of data.
