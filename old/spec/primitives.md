# Able Language Specification: Primitive Types

This section defines the primitive types in **Able**, a hybrid functional and imperative language designed for simplicity and expressiveness. Primitive types form the foundation of Able’s type system, supporting both scalar values and a generic sequence type.

## Primitive Types

Able defines the following primitive types:

| Type    | Description                          | Literal Examples          |
|---------|--------------------------------------|---------------------------|
| `i8`    | 8-bit signed integer (-128 to 127)  | `-128`, `0`, `127`       |
| `i16`   | 16-bit signed integer (-32,768 to 32,767) | `-32768`, `42`, `32767`  |
| `i32`   | 32-bit signed integer (-2³¹ to 2³¹-1) | `-2_147_483_648`, `100`  |
| `i64`   | 64-bit signed integer (-2⁶³ to 2⁶³-1) | `-9_223_372_036_854_775_808`, `500` |
| `i128`  | 128-bit signed integer (-2¹²⁷ to 2¹²⁷-1) | `-2^127`, `0`, `2^127-1` |
| `u8`    | 8-bit unsigned integer (0 to 255)   | `0`, `255`               |
| `u16`   | 16-bit unsigned integer (0 to 65,535) | `0`, `65535`             |
| `u32`   | 32-bit unsigned integer (0 to 2³²-1) | `0`, `4_294_967_295`     |
| `u64`   | 64-bit unsigned integer (0 to 2⁶⁴-1) | `0`, `18_446_744_073_709_551_615` |
| `u128`  | 128-bit unsigned integer (0 to 2¹²⁸-1) | `0`, `2^128-1`           |
| `f32`   | 32-bit floating-point number (IEEE 754 single-precision) | `3.14`, `-0.5`, `2.0`    |
| `f64`   | 64-bit floating-point number (IEEE 754 double-precision) | `3.14159`, `-0.001`, `1e-10` |
| `string`| Unicode strings                     | `"hello"`, `""`          |
| `bool`  | Boolean values                      | `true`, `false`          |
| `char`  | Unicode character (UTF-32 scalar)   | `'a'`, `'π'`, `'💡'`     |
| `nil`   | Singleton type with one value       | `nil`                    |
| `Array` | Homogeneous ordered sequence (generic) | `[1, 2, 3]`, `["a", "b"]` |

### Detailed Definitions

#### Integer Types
- **Signed Integers**: `i8`, `i16`, `i32`, `i64`, `i128`
  - Represent signed integers of varying bit widths.
  - Literal syntax supports underscores for readability (e.g., `1_000`).
- **Unsigned Integers**: `u8`, `u16`, `u32`, `u64`, `u128`
  - Represent non-negative integers of varying bit widths.
  - Same literal syntax as signed integers (e.g., `4_294_967_295`).

#### Floating-Point Types
- **`f32`**: 32-bit IEEE 754 single-precision floating-point.
  - Range: ~±1.18e-38 to ±3.4e38; ~7 decimal digits precision.
- **`f64`**: 64-bit IEEE 754 double-precision floating-point.
  - Range: ~±2.23e-308 to ±1.8e308; ~15 decimal digits precision.
- Literal syntax includes decimals and scientific notation (e.g., `3.14`, `1e-10`).

#### String and Character Types
- **`string`**: A sequence of Unicode characters encoded as UTF-8.
  - Literal syntax: Double-quoted strings (e.g., `"hello"`).
- **`char`**: A single Unicode scalar value (UTF-32, 32-bit).
  - Literal syntax: Single-quoted characters (e.g., `'a'`, `'💡'`).

#### Boolean Type
- **`bool`**: Represents logical truth values.
  - Literal values: `true`, `false`.

#### Nil Type
- **`nil`**: A singleton type with exactly one value, `nil`.
  - Represents a 'nothing' value that is outside all domains.
  - Literal: `nil`.

#### Void type
- **`void`**: A type with no values. This type is an empty set.
  - Represents the absence of value. This is useful as a function return type in which all we care about is knowing that the function returned normally.

#### Array Type
- **`Array`**: A generic type representing a homogeneous ordered sequence.
  - Syntax: `Array T` where `T` is any type (e.g., `Array i32`, `Array string`).
  - Literal syntax: Square brackets with comma-separated elements (e.g., `[1, 2, 3]`).
  - Elements must share a common type.

### Semantics
- **Type Usage**: Primitive types can be used in variable assignments (e.g., `x: i32 = 42`), type expressions (e.g., `Array char`), and expressions.
- **Immutability**: Values of primitive types are immutable by default; mutability depends on assignment context (functional vs. imperative).
- **Inference**: Type annotations are optional; if omitted, the type is inferred from the literal or expression (e.g., `x = 42` infers `i32` unless specified).
