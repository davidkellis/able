# Able Language Specification: Type System Fundamentals

This section outlines the core concepts of Able's type system, focusing on types, type expressions, polymorphism, and constraints. Able features a static, strong type system with extensive type inference.

## Overview

A type is a name given to a set of values, and every value has an associated type. For example, `bool` is the name given to the set `{true, false}`, and since the value `true` is a member of the set `{true, false}`, it is of type `bool`. `TwoBitUnsignedInt` might be the type name we give to the set `{0, 1, 2, 3}`, such that `3` would be a value of type `TwoBitUnsignedInt`.

A type is denoted by a type expression. A type expression is a string. All types are parametric types, in that all types have zero or more type parameters.

Type parameters may be bound or unbound. A bound type parameter is a type parameter for which either a named type variable or a preexisting type name has been substituted. An unbound type parameter is a type parameter that is either unspecified or substituted by the placeholder type variable, denoted by `_`.

A type that has all its type parameters bound is called a concrete type. A type that has any of its type parameters unbound is called a polymorphic type, and a type expression that represents a polymorphic type is called a type constructor.


## 1. Types and Values

*   **Value:** Any piece of data that can be computed and manipulated within the Able language (e.g., the number `42`, the string `"hello"`, the boolean `true`, an instance of a `struct`).
*   **Type:** A type is a named classification representing a set of possible values. Every value in Able has a specific, well-defined type.
    *   Example: The type `bool` represents the set of values `{true, false}`. The value `true` has type `bool`.
    *   Example: If `struct Point { x: f64, y: f64 }` is defined, then `Point` is a type representing the set of all possible point structures with `f64` fields `x` and `y`. An instance `Point { x: 1.0, y: 0.0 }` has type `Point`.

## 2. Type Expressions

A type expression is the syntactic representation used in the Able source code to denote a type.

*   **Syntax:** Type expressions are composed of:
    *   **Type Names:** Identifiers that name a type (e.g., `i32`, `String`, `bool`, `Array`, `Point`, `Option`).
    *   **Type Arguments:** Other type expressions provided as parameters to a type name (e.g., `i32` in `Array i32`). Arguments are space-delimited.
    *   **Parentheses:** Used for grouping type sub-expressions to control application order (e.g., `Map String (List Int)`).
    *   **Nullable Shorthand:** `?TypeName` (desugars to a union `Nil | TypeName`).
    *   **Function Type Syntax:** `(ArgType1, ArgType2, ...) -> ReturnType`.
    *   **Wildcard Placeholder:** `_` used to explicitly denote an unbound type parameter.

## 3. Parametric Nature of Types

*   **Universally Parametric:** Conceptually, *all* types in Able are parametric, meaning they have zero or more type parameters associated with their definition.
    *   A primitive type like `i32` has zero type parameters.
    *   A generic type like `Array` (as defined) intrinsically has one type parameter (the element type).
    *   A generic type like `Map` intrinsically has two type parameters (key and value types).
*   **Type Parameters:** These act as placeholders in a type's definition that can be filled in (bound) with specific types (arguments) when the type is used.

## 4. Parameter Binding, Polymorphism, and Type Constructors

*   **Bound Type Parameter:** A type parameter is considered **bound** when a specific type (which could be a concrete type, a type variable from an enclosing scope, or another type constructor) is provided as an argument for it.
    *   In `Array i32`, the single type parameter of `Array` is bound to the concrete type `i32`.
    *   In `Map String User`, the two type parameters of `Map` are bound to `String` and `User`.
    *   In `struct Foo T { value: T }`, within the scope of `Foo`, `T` is a type variable acting as a bound parameter.
*   **Unbound Type Parameter:** A type parameter is considered **unbound** if:
    *   An argument for it is not specified.
    *   The wildcard placeholder `_` is explicitly used in its position.
*   **Concrete Type:** A type expression denotes a **concrete type** if *all* of its inherent type parameters (and those of any nested types) are bound to specific types or type variables. Values can only have concrete types.
    *   Examples: `i32`, `String`, `Array bool`, `Map String (Array i32)`, `Point`, `?String`.
*   **Polymorphic Type / Type Constructor:** A type expression denotes a **polymorphic type** (or acts as a **type constructor**) if it has one or more unbound type parameters. Type constructors cannot be the type of a runtime value directly but are used in contexts like interface implementations (`impl Mappable A for Array`) or potentially as type arguments themselves (if full HKTs are supported).
    *   Examples:
        *   `Array` (parameter is unspecified) - represents the "Array-ness" ready to accept an element type.
        *   `Array _` (parameter explicitly unbound) - same as above.
        *   `Map String` (second parameter unspecified) - represents a map constructor fixed to `String` keys, awaiting a value type. Equivalent to `Map String _`.
        *   `Map _ bool` (first parameter unbound) - represents a map constructor fixed to `bool` values, awaiting a key type.
        *   `Map` (both parameters unspecified) - represents the map constructor itself. Equivalent to `Map _ _`.
        *   `?` (desugared from `Nil | _` ?) - potentially the nullable type constructor.

## 5. Type Constraints

Type constraints restrict the types that can be used for a generic type parameter. They ensure that a given type implements one or more specified interfaces.

*   **Syntax:**
    *   `TypeParameter : Interface1` (Requires `TypeParameter` to implement `Interface1`)
    *   `TypeParameter : Interface1 + Interface2 + ...` (Requires implementation of all listed interfaces)
*   **Usage Locations:**
    1.  **Generic Parameter Lists:** Directly within angle brackets `< >` (if used) or space-delimited lists for function, struct, interface, or impl definitions.
        ```able
        fn process<T: Display>(item: T) { ... }
        struct Container T: Numeric + Clone { data: T }
        impl<T: Debug> Display for MyType T { ... }
        ```
    2.  **`where` Clauses:** For more complex constraints or better layout.
        ```able
        fn complex_op<K, V, R>(...)
          where K: Hash + Display,
                V: Numeric,
                R: Default {
          ...
        }
        ```

*   **Semantics:** The compiler enforces these constraints. If a type argument provided for a constrained parameter does not implement the required interface(s), a compile-time error occurs. Constraints allow the code within the generic scope to safely use the methods defined by the required interfaces on values of the constrained type parameter.
