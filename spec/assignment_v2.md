Okay, let's consolidate the specification for assignment and destructuring based on our discussions. This reflects the current state where bindings are mutable by default and clarifies how patterns work with different struct types.

---

# Able Language Specification: Assignment and Destructuring

This section defines variable assignment and destructuring in Able. The core mechanism binds identifiers within a pattern to corresponding parts of a value produced by an expression.

## 1. Core Syntax

Variable assignment and destructuring use the following syntax:

```able
Pattern = Expression
```

-   **`Pattern`**: A pattern specifying the structure to match and the identifiers to bind. See details below.
-   **`=`**: The assignment operator.
-   **`Expression`**: Any valid Able expression that evaluates to a value.

## 2. Patterns

Patterns determine how the value from the `Expression` is deconstructed and which identifiers are bound.

### 2.1. Identifier Pattern

The simplest pattern binds the entire result of the `Expression` to a single identifier.

-   **Syntax**: `Identifier`
-   **Example**:
    ```able
    x = 42
    user_name = fetch_user_name()
    my_func = { a, b => a + b }
    ```
-   **Semantics**: Binds the identifier `Identifier` to the value produced by `Expression`.

### 2.2. Wildcard Pattern (`_`)

The wildcard `_` matches any value but does not bind it to any identifier. It's used to ignore parts of a value during destructuring.

-   **Syntax**: `_`
-   **Usage**: Primarily used *within* composite patterns (structs, arrays, tuples). Using `_` as the entire pattern (`_ = Expression`) is allowed and simply evaluates the expression, discarding its result.
-   **Example**:
    ```able
    { x: _, y } = get_point() ## Ignore the 'x' field, bind 'y'
    [_, second, _] = get_three_items() ## Ignore first and third items
    _ = function_with_side_effects() ## Evaluate function, ignore result
    ```

### 2.3. Struct Pattern (Named Fields)

Destructures instances of structs defined with named fields.

-   **Syntax**: `StructTypeName { Field1: PatternA, Field2 @ BindingB: PatternC, ShorthandField, ... }`
    *   `StructTypeName`: Optional, the name of the struct type. If omitted, the type is inferred or checked against the expected type.
    *   `Field: Pattern`: Matches the value of `Field` against the nested `PatternA`.
    *   `Field @ Binding: Pattern`: Matches the value of `Field` against nested `PatternC` and binds the original field value to the new identifier `BindingB`.
    *   `ShorthandField`: Equivalent to `ShorthandField: ShorthandField`. Binds the value of the field `ShorthandField` to an identifier with the same name.
    *   `...`: A literal `...` might be considered in the future to ignore remaining fields explicitly, but currently, extra fields in the value that are not mentioned in the pattern are simply ignored.
-   **Example**:
    ```able
    struct Point { x: f64, y: f64 }
    struct User { id: u64, name: String, address: String }

    p = Point { x: 1.0, y: 2.0 }
    u = User { id: 101, name: "Alice", address: "123 Main St" }

    ## Destructure Point
    Point { x: x_coord, y: y_coord } = p ## Binds x_coord=1.0, y_coord=2.0
    { x, y } = p                       ## Shorthand, binds x=1.0, y=2.0

    ## Destructure User
    { id, name @ user_name, address: addr } = u
    ## Binds id=101, user_name="Alice", addr="123 Main St"

    ## Ignore fields
    { id, name: _ } = u ## Binds id=101, ignores name, ignores address implicitly
    ```
-   **Semantics**: Matches fields by name. If `StructTypeName` is present, checks if the `Expression` value is of that type. Fails if a field mentioned in the pattern doesn't exist in the value.

### 2.4. Struct Pattern (Positional Fields / Named Tuples)

Destructures instances of structs defined with positional fields.

-   **Syntax**: `StructTypeName { Pattern0, Pattern1, ..., PatternN }`
    *   `StructTypeName`: Optional, the name of the struct type.
    *   `Pattern0, Pattern1, ...`: Patterns corresponding to the fields by their zero-based index. The number of patterns must match the number of fields defined for the struct type.
-   **Example**:
    ```able
    struct IntPair { i32, i32 }
    struct Coord3D { f64, f64, f64 }

    pair = IntPair { 10, 20 }
    coord = Coord3D { 1.0, -2.5, 0.0 }

    IntPair { first, second } = pair ## Binds first=10, second=20
    { x, y, z } = coord           ## Binds x=1.0, y=-2.5, z=0.0

    ## Ignore positional fields
    { _, y_val, _ } = coord       ## Binds y_val=-2.5
    ```
-   **Semantics**: Matches fields by position. If `StructTypeName` is present, checks the type. Fails if the number of patterns does not match the number of fields in the value's type.

### 2.5. Array Pattern

Destructures instances of the built-in `Array` type.

-   **Syntax**: `[Pattern1, Pattern2, ..., ...RestIdentifier]`
    *   `Pattern1, Pattern2, ...`: Patterns matching array elements by position from the start.
    *   `...RestIdentifier`: Optional. If present, matches all remaining elements *after* the preceding positional patterns and binds them as a *new* `Array` to `RestIdentifier`.
    *   `...`: If `...` is used without an identifier, it matches remaining elements but does not bind them.
-   **Example**:
    ```able
    data = [10, 20, 30, 40]

    [a, b, c, d] = data ## Binds a=10, b=20, c=30, d=40
    [first, second, ...rest] = data ## Binds first=10, second=20, rest=[30, 40]
    [x, _, y, ...] = data ## Binds x=10, y=30, ignores second element and rest
    [single] = [99] ## Binds single=99
    [] = [] ## Matches an empty array
    ```
-   **Semantics**: Matches elements by position. Fails if the array has fewer elements than required by the non-rest patterns.

### 2.6. Nested Patterns

Patterns can be nested arbitrarily within struct and array patterns.

-   **Example**:
    ```able
    struct Data { id: u32, point: Point } ## Point is { x: f64, y: f64 }
    struct Container { items: Array Data }

    val = Container { items: [ Data { id: 1, point: Point { x: 1.0, y: 2.0 } },
                               Data { id: 2, point: Point { x: 3.0, y: 4.0 } } ] }

    Container { items: [ Data { id: first_id, point: { x: first_x, y: _ }}, ...rest_data ] } = val
    ## Binds first_id = 1
    ## Binds first_x = 1.0
    ## Ignores y of the first point
    ## Binds rest_data = [ Data { id: 2, point: Point { x: 3.0, y: 4.0 } } ]
    ```

## 3. Semantics

1.  **Evaluation Order**: The `Expression` (right-hand side) is evaluated first to produce a value.
2.  **Matching & Binding**: The resulting value is then matched against the `Pattern` (left-hand side).
    *   If the value's structure and type match the pattern, identifiers within the pattern are bound to the corresponding parts of the value within the current scope.
    *   If the match fails (e.g., type mismatch, structural mismatch like wrong number of elements, missing field), a runtime error/panic occurs.
3.  **Mutability**: All bindings introduced via `Pattern = Expression` are **mutable** by default in the current specification (i.e., the bound identifier can be reassigned later using `=`).
4.  **Scope**: Bindings are introduced into the current lexical scope.
5.  **Type Checking**: The type system checks for compatibility between the type of the `Expression` and the structure expected by the `Pattern`. Type inference applies where possible.

---
