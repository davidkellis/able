# Able Language Specification (Draft)

**Version:** As of 2023-10-27 conversation (incorporating v2/v3 revisions and user updates)
**Status:** Incomplete Draft - Requires Standard Library definition and further refinement on TBD items.

## Table of Contents

1.  [Introduction](#1-introduction)
    *   [1.1. Core Philosophy & Goals](#11-core-philosophy--goals)
    *   [1.2. Document Scope](#12-document-scope)
2.  [Lexical Structure](#2-lexical-structure)
3.  [Syntax Style & Blocks](#3-syntax-style--blocks)
4.  [Types](#4-types)
    *   [4.1. Type System Fundamentals](#41-type-system-fundamentals)
    *   [4.2. Primitive Types](#42-primitive-types)
    *   [4.3. Type Expression Syntax Details](#43-type-expression-syntax-details)
    *   [4.4. Reserved Identifier (`_`) in Types](#44-reserved-identifier-_-in-types)
    *   [4.5. Structs](#45-structs)
        *   [4.5.1. Singleton Structs](#451-singleton-structs)
        *   [4.5.2. Structs with Named Fields](#452-structs-with-named-fields)
        *   [4.5.3. Structs with Positional Fields (Named Tuples)](#453-structs-with-positional-fields-named-tuples)
    *   [4.6. Union Types (Sum Types / ADTs)](#46-union-types-sum-types--adts)
        *   [4.6.1. Union Declaration](#461-union-declaration)
        *   [4.6.2. Nullable Type Shorthand (`?`)](#462-nullable-type-shorthand-)
        *   [4.6.3. Constructing Union Values](#463-constructing-union-values)
        *   [4.6.4. Using Union Values](#464-using-union-values)
5.  [Bindings, Assignment, and Destructuring](#5-bindings-assignment-and-destructuring)
    *   [5.1. Operators (`:=`, `=`)](#51-operators---)
    *   [5.2. Patterns](#52-patterns)
        *   [5.2.1. Identifier Pattern](#521-identifier-pattern)
        *   [5.2.2. Wildcard Pattern (`_`)](#522-wildcard-pattern-_)
        *   [5.2.3. Struct Pattern (Named Fields)](#523-struct-pattern-named-fields)
        *   [5.2.4. Struct Pattern (Positional Fields / Named Tuples)](#524-struct-pattern-positional-fields--named-tuples)
        *   [5.2.5. Array Pattern](#525-array-pattern)
        *   [5.2.6. Nested Patterns](#526-nested-patterns)
    *   [5.3. Semantics of Assignment/Declaration](#53-semantics-of-assignmentdeclaration)
6.  [Expressions](#6-expressions)
    *   [6.1. Literals](#61-literals)
    *   [6.2. Block Expressions (`do`)](#62-block-expressions-do)
    *   [6.3. Operators](#63-operators)
        *   [6.3.1. Operator Precedence and Associativity](#631-operator-precedence-and-associativity)
        *   [6.3.2. Operator Semantics](#632-operator-semantics)
        *   [6.3.3. Overloading (Via Interfaces)](#633-overloading-via-interfaces)
    *   [6.4. Function Calls](#64-function-calls)
    *   [6.5. Control Flow Expressions](#65-control-flow-expressions)
    *   [6.6. String Interpolation](#66-string-interpolation)
7.  [Functions](#7-functions)
    *   [7.1. Named Function Definition](#71-named-function-definition)
    *   [7.2. Anonymous Functions and Closures](#72-anonymous-functions-and-closures)
        *   [7.2.1. Verbose Anonymous Function Syntax](#721-verbose-anonymous-function-syntax)
        *   [7.2.2. Lambda Expression Syntax](#722-lambda-expression-syntax)
        *   [7.2.3. Closures](#723-closures)
    *   [7.3. Explicit `return` Statement](#73-explicit-return-statement)
    *   [7.4. Function Invocation](#74-function-invocation)
        *   [7.4.1. Standard Call](#741-standard-call)
        *   [7.4.2. Trailing Lambda Syntax](#742-trailing-lambda-syntax)
        *   [7.4.3. Method Call Syntax](#743-method-call-syntax)
        *   [7.4.4. Callable Value Invocation (`Apply` Interface)](#744-callable-value-invocation-apply-interface)
    *   [7.5. Partial Function Application](#75-partial-function-application)
    *   [7.6. Shorthand Notations](#76-shorthand-notations)
        *   [7.6.1. Implicit First Argument Access (`#member`)](#761-implicit-first-argument-access-member)
        *   [7.6.2. Implicit Self Parameter Definition (`fn #method`)](#762-implicit-self-parameter-definition-fn-method)
8.  [Control Flow](#8-control-flow)
    *   [8.1. Branching Constructs](#81-branching-constructs)
        *   [8.1.1. Conditional Chain (`if`/`or`)](#811-conditional-chain-ifor)
        *   [8.1.2. Pattern Matching Expression (`match`)](#812-pattern-matching-expression-match)
    *   [8.2. Looping Constructs](#82-looping-constructs)
        *   [8.2.1. While Loop (`while`)](#821-while-loop-while)
        *   [8.2.2. For Loop (`for`)](#822-for-loop-for)
        *   [8.2.3. Range Expressions](#823-range-expressions)
    *   [8.3. Non-Local Jumps (`breakpoint` / `break`)](#83-non-local-jumps-breakpoint--break)
        *   [8.3.1. Defining an Exit Point (`breakpoint`)](#831-defining-an-exit-point-breakpoint)
        *   [8.3.2. Performing the Jump (`break`)](#832-performing-the-jump-break)
        *   [8.3.3. Semantics](#833-semantics)
9.  [Inherent Methods (`methods`)](#9-inherent-methods-methods)
    *   [9.1. Syntax](#91-syntax)
    *   [9.2. Method Definitions](#92-method-definitions)
    *   [9.3. Method Call Syntax Resolution (Initial Rules)](#93-method-call-syntax-resolution-initial-rules)
10. [Interfaces and Implementations](#10-interfaces-and-implementations)
    *   [10.1. Interfaces](#101-interfaces)
        *   [10.1.1. Interface Usage Models](#1011-interface-usage-models)
        *   [10.1.2. Interface Definition](#1012-interface-definition)
        *   [10.1.3. `Self` Keyword Interpretation](#1013-self-keyword-interpretation)
        *   [10.1.4. Composite Interfaces (Interface Aliases)](#1014-composite-interfaces-interface-aliases)
    *   [10.2. Implementations (`impl`)](#102-implementations-impl)
        *   [10.2.1. Implementation Declaration](#1021-implementation-declaration)
        *   [10.2.2. HKT Implementation Syntax](#1022-hkt-implementation-syntax)
        *   [10.2.3. Overlapping Implementations and Specificity](#1023-overlapping-implementations-and-specificity)
    *   [10.3. Usage](#103-usage)
        *   [10.3.1. Instance Method Calls](#1031-instance-method-calls)
        *   [10.3.2. Static Method Calls](#1032-static-method-calls)
        *   [10.3.3. Disambiguation (Named Impls)](#1033-disambiguation-named-impls)
        *   [10.3.4. Interface Types (Dynamic Dispatch)](#1034-interface-types-dynamic-dispatch)
11. [Error Handling](#11-error-handling)
    *   [11.1. Explicit `return` Statement](#111-explicit-return-statement)
    *   [11.2. Option/Result Types and Operators (`?Type`, `!Type`, `!`, `else`)](#112-optionresult-types-and-operators-type-type--else)
        *   [11.2.1. Core Types (`?Type`, `!Type`)](#1121-core-types-type-type)
        *   [11.2.2. Propagation (`!`)](#1122-propagation-)
        *   [11.2.3. Handling (`else {}`)](#1123-handling-else-)
    *   [11.3. Exceptions (`raise` / `rescue`)](#113-exceptions-raise--rescue)
        *   [11.3.1. Raising Exceptions (`raise`)](#1131-raising-exceptions-raise)
        *   [11.3.2. Rescuing Exceptions (`rescue`)](#1132-rescuing-exceptions-rescue)
        *   [11.3.3. Panics](#1133-panics)
12. [Concurrency](#12-concurrency)
    *   [12.1. Concurrency Model Overview](#121-concurrency-model-overview)
    *   [12.2. Asynchronous Execution (`proc`)](#122-asynchronous-execution-proc)
        *   [12.2.1. Syntax](#1221-syntax)
        *   [12.2.2. Semantics](#1222-semantics)
        *   [12.2.3. Process Handle (`Proc T` Interface)](#1223-process-handle-proc-t-interface)
    *   [12.3. Thunk-Based Asynchronous Execution (`spawn`)](#123-thunk-based-asynchronous-execution-spawn)
        *   [12.3.1. Syntax](#1231-syntax)
        *   [12.3.2. Semantics](#1232-semantics)
    *   [12.4. Key Differences (`proc` vs `spawn`)](#124-key-differences-proc-vs-spawn)
13. [Packages and Modules](#13-packages-and-modules)
    *   [13.1. Package Naming and Structure](#131-package-naming-and-structure)
    *   [13.2. Package Configuration (`package.yml`)](#132-package-configuration-packageyml)
    *   [13.3. Package Declaration in Source Files](#133-package-declaration-in-source-files)
    *   [13.4. Importing Packages (`import`)](#134-importing-packages-import)
    *   [13.5. Visibility and Exports (`private`)](#135-visibility-and-exports-private)
14. [Standard Library Interfaces (Conceptual / TBD)](#14-standard-library-interfaces-conceptual--tbd)
15. [To Be Defined / Refined](#15-to-be-defined--refined)

## 1. Introduction

### 1.1. Core Philosophy & Goals
Able is a **hybrid functional and imperative** programming language designed with the following goals:
*   **Minimal & Expressive Syntax:** Strive for clarity and low syntactic noise while providing powerful features.
*   **Static & Strong Typing:** Ensure type safety at compile time with extensive **type inference**.
*   **Functional Features:** Support first-class functions, closures, algebraic data types (unions), pattern matching, and encourage immutability where practical.
*   **Imperative Flexibility:** Allow mutable state, side effects, looping constructs, and exception handling when needed.
*   **Memory Management:** Utilize **garbage collection**.
*   **Concurrency:** Provide lightweight, Go-inspired concurrency primitives.
*   **Pragmatism:** Blend theoretical concepts with practical usability.

### 1.2. Document Scope
This document specifies the syntax and semantics of the Able language core features as defined to date. It does not yet include a full standard library specification, detailed tooling requirements, or finalized definitions for all TBD items.

## 2. Lexical Structure

Defines how raw text is converted into tokens.

*   **Character Set:** UTF-8 source files are recommended.
*   **Identifiers:** Start with a letter (`a-z`, `A-Z`) or underscore (`_`), followed by letters, digits (`0-9`), or underscores. Typically `[a-zA-Z_][a-zA-Z0-9_]*`. Identifiers are case-sensitive. Package/directory names mapping to identifiers treat hyphens (`-`) as underscores. The identifier `_` is reserved as the wildcard pattern (see Section [5.2.2](#522-wildcard-pattern-_)) and for unbound type parameters (see Section [4.4](#44-reserved-identifier-_-in-types)).
*   **Keywords:** Reserved words that cannot be used as identifiers. Includes: `fn`, `struct`, `union`, `interface`, `impl`, `methods`, `if`, `or`, `while`, `for`, `in`, `match`, `case`, `breakpoint`, `break`, `type`, `package`, `import`, `private`, `nil`, `true`, `false`, `void`, `Self`, `proc`, `spawn`, `raise`, `rescue`, `do`, `else`. (List may be incomplete).
*   **Operators:** Symbols with specific meanings (See Section [6.3](#63-operators)). Includes assignment/declaration operators `:=` and `=`.
*   **Literals:** Source code representations of fixed values (See Section [4.2](#42-primitive-types) and Section [6.1](#61-literals)).
*   **Comments:** Line comments start with `##` and continue to the end of the line. Block comment syntax is TBD.
    ```able
    x := 1 ## Assign 1 to x (line comment)
    ```
*   **Whitespace:** Spaces, tabs, and form feeds are generally insignificant except for separating tokens.
*   **Newlines:** Significant as expression separators within blocks (See Section [3](#3-syntax-style--blocks)).
*   **Delimiters:** `()`, `{}`, `[]`, `<>`, `,`, `;`, `:`, `->`, `|`, `=`, `:=`, `.`, `...`, `..`, `_`, `` ` ``, `#`, `?`, `!`, `~`, `=>`, `|>`.

## 3. Syntax Style & Blocks

*   **General Feel:** A blend of ML-family (OCaml, F#) and Rust syntax influences.
*   **Code Blocks `{}`:** Curly braces group sequences of expressions in specific syntactic contexts:
    *   Function bodies (`fn ... { ... }`)
    *   Struct literals (`Point { ... }`)
    *   Lambda literals (`{ ... => ... }`)
    *   Control flow branches (`if ... { ... }`, `match ... { case => ... }`, `else { ... }` etc.)
    *   `methods`/`impl` bodies (`methods Type { ... }`)
    *   `do` blocks (`do { ... }`) (See Section [6.2](#62-block-expressions-do))
*   **Expression Separation:** Within blocks, expressions are evaluated sequentially. They are separated by **newlines** or optionally by **semicolons** (`;`). The last expression in a block determines its value unless otherwise specified (e.g., loops, assignments).
*   **Expression-Oriented:** Most constructs are expressions evaluating to a value (e.g., `if/or`, `match`, `breakpoint`, `rescue`, `do` blocks, assignment/declaration (`=`, `:=`)). Loops (`while`, `for`) evaluate to `nil`.

## 4. Types

Able is statically and strongly typed.

### 4.1. Type System Fundamentals

A type is a name given to a set of values, and every value has an associated type. For example, `bool` is the name given to the set `{true, false}`, and since the value `true` is a member of the set `{true, false}`, it is of type `bool`. `TwoBitUnsignedInt` might be the type name we give to the set `{0, 1, 2, 3}`, such that `3` would be a value of type `TwoBitUnsignedInt`.

A type is denoted by a type expression. A type expression is a string. All types are parametric types, in that all types have zero or more type parameters.

Type parameters may be bound or unbound. A bound type parameter is a type parameter for which either a named type variable or a preexisting type name has been substituted. An unbound type parameter is a type parameter that is either unspecified or substituted by the placeholder type variable, denoted by `_`.

A type that has all its type parameters bound is called a concrete type. A type that has any of its type parameters unbound is called a polymorphic type, and a type expression that represents a polymorphic type is called a type constructor.

#### 4.1.1. Types and Values

*   **Value:** Any piece of data that can be computed and manipulated within the Able language (e.g., the number `42`, the string `"hello"`, the boolean `true`, an instance of a `struct`).
*   **Type:** A type is a named classification representing a set of possible values. Every value in Able has a specific, well-defined type.
    *   Example: The type `bool` represents the set of values `{true, false}`. The value `true` has type `bool`.
    *   Example: If `struct Point { x: f64, y: f64 }` is defined, then `Point` is a type representing the set of all possible point structures with `f64` fields `x` and `y`. An instance `Point { x: 1.0, y: 0.0 }` has type `Point`.

#### 4.1.2. Type Expressions

A type expression is the syntactic representation used in the Able source code to denote a type.

*   **Syntax:** Type expressions are composed of:
    *   **Type Names:** Identifiers that name a type (e.g., `i32`, `string`, `bool`, `Array`, `Point`, `Option`).
    *   **Type Arguments:** Other type expressions provided as parameters to a type name (e.g., `i32` in `Array i32`). Arguments are space-delimited.
    *   **Parentheses:** Used for grouping type sub-expressions to control application order (e.g., `Map string (Array i32)`).
    *   **Nullable Shorthand:** `?TypeName` (desugars to a union `nil | TypeName`). See Section [4.6.2](#462-nullable-type-shorthand-).
    *   **Result Shorthand:** `!TypeName` (desugars to a union `TypeName | Error`). See Section [11.2.1](#1121-core-types-type-type).
    *   **Function Type Syntax:** `(ArgType1, ArgType2, ...) -> ReturnType`. See Section [7](#7-functions).
    *   **Wildcard Placeholder:** `_` used to explicitly denote an unbound type parameter. See Section [4.4](#44-reserved-identifier-_-in-types).

#### 4.1.3. Parametric Nature of Types

*   **Universally Parametric:** Conceptually, *all* types in Able are parametric, meaning they have zero or more type parameters associated with their definition.
    *   A primitive type like `i32` has zero type parameters.
    *   A generic type like `Array` (as defined) intrinsically has one type parameter (the element type).
    *   A generic type like `Map` intrinsically has two type parameters (key and value types).
*   **Type Parameters:** These act as placeholders in a type's definition that can be filled in (bound) with specific types (arguments) when the type is used.

#### 4.1.4. Parameter Binding, Polymorphism, and Type Constructors

*   **Bound Type Parameter:** A type parameter is considered **bound** when a specific type (which could be a concrete type, a type variable from an enclosing scope, or another type constructor) is provided as an argument for it.
    *   In `Array i32`, the single type parameter of `Array` is bound to the concrete type `i32`.
    *   In `Map string User`, the two type parameters of `Map` are bound to `string` and `User`.
    *   In `struct Foo T { value: T }`, within the scope of `Foo`, `T` is a type variable acting as a bound parameter.
*   **Unbound Type Parameter:** A type parameter is considered **unbound** if:
    *   An argument for it is not specified.
    *   The wildcard placeholder `_` is explicitly used in its position. See Section [4.4](#44-reserved-identifier-_-in-types).
*   **Concrete Type:** A type expression denotes a **concrete type** if *all* of its inherent type parameters (and those of any nested types) are bound to specific types or type variables. Values can only have concrete types.
    *   Examples: `i32`, `string`, `Array bool`, `Map string (Array i32)`, `Point`, `?string`.
*   **Polymorphic Type / Type Constructor:** A type expression denotes a **polymorphic type** (or acts as a **type constructor**) if it has one or more unbound type parameters. Type constructors cannot be the type of a runtime value directly but are used in contexts like interface implementations (`impl Mappable A for Array`) or potentially as type arguments themselves (if full HKTs are supported).
    *   Examples:
        *   `Array` (parameter is unspecified) - represents the "Array-ness" ready to accept an element type.
        *   `Array _` (parameter explicitly unbound) - same as above.
        *   `Map string` (second parameter unspecified) - represents a map constructor fixed to `string` keys, awaiting a value type. Equivalent to `Map string _`.
        *   `Map _ bool` (first parameter unbound) - represents a map constructor fixed to `bool` values, awaiting a key type.
        *   `Map` (both parameters unspecified) - represents the map constructor itself. Equivalent to `Map _ _`.
        *   `?` (desugared from `nil | _` ?) - potentially the nullable type constructor.

#### 4.1.5. Type Constraints

Type constraints restrict the types that can be used for a generic type parameter. They ensure that a given type implements one or more specified interfaces.

*   **Syntax:**
    *   `TypeParameter : Interface1` (Requires `TypeParameter` to implement `Interface1`)
    *   `TypeParameter : Interface1 + Interface2 + ...` (Requires implementation of all listed interfaces)
*   **Usage Locations:**
    1.  **Generic Parameter Lists:** Directly within angle brackets `< >` (if used) or space-delimited lists for function, struct, interface, or impl definitions.

        > **Delimiter Rules for Generic Parameters**
        >
        > - When generic parameters are enclosed in angle brackets `<...>`, **parameters must be comma-delimited**.
        >   **Example:** `<A, B, C>`, `<T: Display, U: Clone>`
        >
        > - When generic parameters are specified **without** angle brackets (such as in type applications, struct or union declarations, or interface headers), **parameters are space-delimited**.
        >   **Example:** `Array i32`, `Map string User`, `struct Pair T U`, `interface Mappable K V`
        >
        > - Constraints on parameters can be specified inline (e.g., `T: Display`) or in a `where` clause, regardless of delimiter style.
        ```able
        fn process<T: Display>(item: T) { ... }
        struct Container T: Numeric + Clone { data: T }
        impl<T: Debug> Display for MyType T { ... }
        ```
    2.  **`where` Clauses:** As an alternative or addition to inline constraints, a `where` clause can be placed after the parameter list (for functions) or the type/interface declaration (for structs, interfaces, impls) to specify constraints. This is often preferred for readability with multiple or complex constraints.
        ```able
        ## Function with where clause
        fn complex_fn<A, B, C>(a: A, b: B) -> C
          where A: Hash, B: Display, C: Default + Cmp B {
          ## ... function body ...
        }

        ## Struct with where clause
        struct ConstrainedContainer K V
          where K: Hash + Eq, V: Clone {
          key: K,
          value: V
        }

        ## Interface with where clause
        interface AdvancedMappable A for M _
          where M: Iterable A {
          ## ... signatures ...
        }

        ## Implementation with where clause
        impl<T> Display for MyType T
          where T: Numeric + Default {
          ## ... implementation ...
        }
        ```

*   **Semantics:** The compiler enforces these constraints regardless of whether they are defined inline or in a `where` clause. If a type argument provided for a constrained parameter does not implement the required interface(s), a compile-time error occurs. Constraints allow the code within the generic scope to safely use the methods defined by the required interfaces on values of the constrained type parameter.

### 4.2. Primitive Types

| Type     | Description                                   | Literal Examples                    | Notes                                           |
| :------- | :-------------------------------------------- | :---------------------------------- | :---------------------------------------------- |
| `i8`     | 8-bit signed integer (-128 to 127)            | `-128`, `0`, `10`, `127_i8`         |                                                 |
| `i16`    | 16-bit signed integer (-32,768 to 32,767)       | `-32768`, `1000`, `32767_i16`        |                                                 |
| `i32`    | 32-bit signed integer (-2³¹ to 2³¹-1)           | `-2_147_483_648`, `0`, `42_i32`      | Default type for integer literals (TBC).        |
| `i64`    | 64-bit signed integer (-2⁶³ to 2⁶³-1)           | `-9_223_..._i64`, `1_000_000_000_i64`| Underscores `_` allowed for readability.        |
| `i128`   | 128-bit signed integer (-2¹²⁷ to 2¹²⁷-1)      | `-170_..._i128`, `0_i128`, `170_..._i128`|                                                 |
| `u8`     | 8-bit unsigned integer (0 to 255)             | `0`, `10_u8`, `255_u8`              |                                                 |
| `u16`    | 16-bit unsigned integer (0 to 65,535)           | `0_u16`, `1000`, `65535_u16`        |                                                 |
| `u32`    | 32-bit unsigned integer (0 to 2³²-1)            | `0`, `4_294_967_295_u32`            | Underscores `_` allowed for readability.        |
| `u64`    | 64-bit unsigned integer (0 to 2⁶⁴-1)            | `0_u64`, `18_446_..._u64`           |                                                 |
| `u128`   | 128-bit unsigned integer (0 to 2¹²⁸-1)          | `0`, `340_..._u128`                 |                                                 |
| `f32`    | 32-bit float (IEEE 754 single-precision)      | `3.14_f32`, `-0.5_f32`, `1e-10_f32`, `2.0_f32` | Suffix `_f32`.                                  |
| `f64`    | 64-bit float (IEEE 754 double-precision)      | `3.14159`, `-0.001`, `1e-10`, `2.0`  | Default type for float literals (TBC). Suffix `_f64` optional if default. |
| `string` | Immutable sequence of Unicode chars (UTF-8) | `"hello"`, `""`, `` `interp ${val}` `` | Double quotes or backticks (interpolation).      |
| `bool`   | Boolean logical values                        | `true`, `false`                     |                                                 |
| `char`   | Single Unicode scalar value (UTF-32)        | `'a'`, `'π'`, `'💡'`, `'\n'`, `'\u{1F604}'` | Single quotes. Supports escape sequences.       |
| `nil`    | Singleton type representing **absence of data**. | `nil`                               | **Type and value are both `nil` (lowercase)**. Often used with `?Type`. |
| `void`   | Type with **no values** (empty set).          | *(No literal value)*                | Represents computations completing without data. |

*(See Section [6.1](#61-literals) for detailed literal syntax.)*

### 4.3. Type Expression Syntax Details

*   **Simple:** `i32`, `string`, `MyStruct`
*   **Generic Application:** `Array i32`, `Map string User` (space-delimited arguments)
*   **Grouping:** `Map string (Array i32)` (parentheses control application order)
*   **Function:** `(i32, string) -> bool`
*   **Nullable Shorthand:** `?string` (Syntactic sugar for `nil | string`, see Section [4.6.2](#462-nullable-type-shorthand-))
*   **Result Shorthand:** `!string` (Syntactic sugar for `string | Error`, see Section [11.2.1](#1121-core-types-type-type))
*   **Wildcard:** `_` denotes an unbound parameter (e.g., `Map string _`).

### 4.4. Reserved Identifier (`_`) in Types

The underscore `_` can be used in type expressions to explicitly denote an unbound type parameter, contributing to forming a polymorphic type / type constructor. Example: `Map string _`.

### 4.5. Structs

Structs aggregate named or positional data fields into a single type. Able supports three kinds of struct definitions: singleton, named-field, and positional-field. A single struct definition must be exclusively one kind. All fields are mutable.

#### 4.5.1. Singleton Structs

Represent types with exactly one value, identical to the type name itself. Useful for simple enumeration variants or tags.

##### Declaration
```able
struct Identifier
```
*(Optionally `struct Identifier {}`)*

-   **`Identifier`**: Names the type and its unique value (e.g., `Red`, `EOF`, `Success`).

##### Instantiation & Usage
Use the identifier directly as the value.
```able
status = Success
color_val: Color = Red ## Assuming 'union Color = Red | ...'
```
Matched using the identifier in patterns: `case Red => ...`.

#### 4.5.2. Structs with Named Fields

Group data under named fields.

##### Declaration
```able
struct Identifier [GenericParamList] [where <ConstraintList>] {
  FieldName1: Type1,
  FieldName2: Type2
  FieldName3: Type3 ## Comma or newline separated, trailing comma ok
}
```
-   **`Identifier`**: Struct type name.
-   **`[GenericParamList]`**: Optional space-delimited generics (e.g., `T`, `K V: Display`). Constraints can be inline or in the `where` clause.
-   **`[where <ConstraintList>]`**: Optional clause for specifying constraints on `GenericParamList`.
-   **`FieldName: Type`**: Defines a field with a unique name within the struct.

##### Instantiation
Use `{ FieldName: Value, ... }`. Order doesn't matter. All fields must be initialized. Field init shorthand `{ name }` is supported.
```able
Identifier [GenericArgs] { Field1: Value1, Field2: Value2, ... }
## GenericArgs space-delimited if explicit, often inferred.
p = Point { x: 1.0, y: 2.0 }
username = "Alice" ## Assume username exists
u = User { id: 101, username, is_active: true } ## Shorthand
```

##### Field Access
Dot notation: `instance.FieldName`.
```able
x_coord = p.x
```

##### Functional Update
Create a new instance based on others using `...Source`. Later sources/fields override earlier ones.
```able
StructType { ...Source1, ...Source2, FieldOverride: NewValue, ... }
addr = Address { ...base_addr, zip: "90210" }
```

##### Field Mutation
Modify fields in-place using assignment (`=`). Requires the binding (`instance`) to be mutable.
```able
instance.FieldName = NewValue
p.x = p.x + 10.0
```

#### 4.5.3. Structs with Positional Fields (Named Tuples)

Define fields by their position and type. Accessed by index.

##### Declaration
```able
struct Identifier [GenericParamList] [where <ConstraintList>] {
  Type1,
  Type2
  Type3 ## Comma or newline separated, trailing comma ok
}
```
-   **`Identifier`**: Struct type name (e.g., `IntPair`, `Coord3D`).
-   **`[GenericParamList]`**: Optional space-delimited generics. Constraints can be inline or in the `where` clause.
-   **`[where <ConstraintList>]`**: Optional clause for specifying constraints on `GenericParamList`.
-   **`Type`**: Defines a field by its type at a specific zero-based position.

##### Instantiation
Use `{ Value1, Value2, ... }`. Values must be provided in the defined order. All fields must be initialized.
```able
Identifier [GenericArgs] { Value1, Value2, ... }
pair = IntPair { 10, 20 }
```

##### Field Access
Dot notation with zero-based integer index: `instance.Index`.
```able
first = pair.0 ## Accesses 10
second = pair.1 ## Accesses 20
```
Compile-time error preferred for invalid literal indices. Runtime error otherwise.

##### Functional Update
Not supported via `...Source` syntax for positional structs. Create new instances explicitly.

##### Field Mutation
Modify fields in-place using indexed assignment (`=`). Requires the binding (`instance`) to be mutable.
```able
instance.Index = NewValue
pair.0 = pair.0 + 5
```

### 4.6. Union Types (Sum Types / ADTs)

Represent values that can be one of several different types (variants). Essential for modeling alternatives (e.g., success/error, presence/absence, different kinds of related data).

#### 4.6.1. Union Declaration

Define a new type as a composition of existing variant types using `|`.

##### Syntax
```able
union UnionTypeName [GenericParamList] = VariantType1 | VariantType2 | ... | VariantTypeN
```
-   **`union`**: Keyword.
-   **`UnionTypeName`**: The name of the new union type being defined.
-   **`[GenericParamList]`**: Optional space-delimited generic parameters applicable to the union itself.
-   **`=`**: Separator.
-   **`VariantType1 | VariantType2 | ...`**: List of one or more variant types separated by `|`.
    -   Each `VariantType` must be a pre-defined, valid type name (e.g., primitive, struct, another union, generic type application).

##### Examples
```able
## Simple enumeration using singleton structs
struct Red; struct Green; struct Blue;
union Color = Red | Green | Blue

## Option type (conceptual - assumes Some struct exists)
## struct Some T { value: T }
union Option T = Some T | nil ## More direct using 'nil' type

## Result type (conceptual - assumes Ok/Err structs exist)
## struct Ok T { value: T }
## struct Err E { error: E }
## union Result T E = Ok T | Err E

## Mixing types
union IntOrString = i32 | string
```

#### 4.6.2. Nullable Type Shorthand (`?`)

Provides concise syntax for types that can be either a specific type or `nil`.

##### Syntax
```able
?Type
```
-   **`?`**: Prefix operator indicating nullability.
-   **`Type`**: Any valid type expression.

##### Equivalence
`?Type` is syntactic sugar for the union `nil | Type`.
*(Note: Defined as `nil | Type` rather than `Type | nil`)*

##### Examples
```able
name: ?string = "Alice"
age: ?i32 = nil
maybe_user: ?User = find_user(id)
```

*(See Section [11.2.1](#1121-core-types-type-type) for `!Type` shorthand)*

#### 4.6.3. Constructing Union Values

Create a value of the union type by creating a value of one of its variant types.

```able
c: Color = Green
opt_val: Option i32 = Some i32 { value: 42 } ## Assuming Some struct definition
opt_nothing: Option i32 = nil

## Assuming Ok/Err struct definitions
# res_ok: Result string string = Ok string { value: "Data loaded" }
# res_err: Result string string = Err string { error: "File not found" }

val: IntOrString = 100
val2: IntOrString = "hello"
```

#### 4.6.4. Using Union Values

The primary way to interact with union values is via `match` expressions (See Section [8.1.2](#812-pattern-matching-expression-match)), which allow safely deconstructing the value based on its current variant.

```able
## Assuming struct F { deg: f64 }, struct C { deg: f64 }, struct K { deg: f64 }
## and union Temp = F | C | K
# temp: Temp = F { deg: 32.0 }
# desc = temp match {
#   case F { deg } => `Fahrenheit: ${deg}`,
#   case C { deg } => `Celsius: ${deg}`,
#   case K { deg } => `Kelvin: ${deg}`
# }

maybe_name: ?string = get_name_option()
display_name = maybe_name match {
  case s: string => s, ## Matches non-nil string
  case nil      => "Guest"
}
```

## 5. Bindings, Assignment, and Destructuring

This section defines variable binding, assignment, and destructuring in Able. Able uses `=` and `:=` for binding identifiers within patterns to values. `=` primarily handles reassignment but can also introduce initial bindings, while `:=` is used explicitly for declaring new bindings, especially for shadowing. Bindings are mutable by default.

### 5.1. Operators (`:=`, `=`)

*   **Declaration (`:=`)**: `Pattern := Expression`
    *   **Always** declares **new** mutable bindings for all identifiers introduced in `Pattern` within the **current** lexical scope.
    *   Initializes these new bindings using the corresponding values from `Expression` via matching.
    *   This is the **required** operator for **shadowing**: if an identifier introduced by `Pattern` has the same name as a binding in an *outer* scope, `:=` creates a new, distinct binding in the current scope that shadows the outer one.
    *   It is a compile-time error if any identifier introduced by `Pattern` already exists as a binding *within the current scope*.
    *   Example (Shadowing):
        ```able
        package_var := 10 ## Assume declared at package level

        fn my_func() {
          ## package_var is accessible here (value 10)
          package_var := 20  ## Declares NEW local binding 'package_var', shadows package-level one
          print(package_var)  ## prints 20 (local)

          ## To modify the package-level variable, use '=':
          # package_var = 30 ## This would be an error if the local 'package_var := 20' exists.
                           ## If the local didn't exist, this would modify the package var.
        }
        my_func()
        print(package_var)  ## prints 10 (package-level was unaffected by local :=)
        ```

*   **Assignment / Initial Binding (`=`)**: `LHS = Expression`
    *   Performs one of two actions based on the `LHS` and current scope:
        1.  **Reassignment/Mutation:** If `LHS` refers to existing, accessible, and mutable locations (bindings, fields, indices) found through lexical scoping rules (checking current scope first, then enclosing scopes), it assigns the value of `Expression` to those locations.
        2.  **Initial Binding:** If `LHS` is a `Pattern` and *none* of the identifiers introduced by the pattern exist as accessible bindings via lexical scoping, it declares **new** mutable bindings for those identifiers in the **current** scope and initializes them. This is effectively a convenience for initial declaration when shadowing is not intended or possible.
    *   **Precedence:** Reassignment (Action 1) takes precedence. `=` will modify an existing accessible binding before creating a new one with the same name. To guarantee a new binding and shadow an outer variable, use `:=`.
    *   `LHS` can be:
        *   An identifier (`x = value`).
        *   A mutable field access (`instance.field = value`).
        *   A mutable index access (`array[index] = value`).
        *   A pattern (`{x, y} = point`) where identifiers either refer to existing mutable locations or introduce new bindings (if no existing binding is found).
    *   It is a compile-time error if `LHS` refers to bindings/locations that do not exist (and cannot be created via initial binding), are not accessible, or are not mutable.
    *   Example (Initial Binding vs. Reassignment):
        ```able
        ## Initial binding (assuming 'a' doesn't exist yet)
        a = 10

        ## Reassignment
        a = 20

        b := 100 ## Declare 'b' in current scope
        b = 200  ## Reassign 'b' in current scope

        do {
          c = 5 ## Initial binding of 'c' in inner scope using '='
          a = 30 ## Reassigns 'a' from outer scope using '='
          b := 50 ## Declares NEW 'b' in inner scope using ':=' (shadows outer 'b')
          b = 60  ## Reassigns inner 'b' using '='
        }
        print(a) ## prints 30
        print(b) ## prints 200 (outer 'b' was shadowed, not reassigned)
        ## 'c' is out of scope here
        ```

### 5.2. Patterns

Patterns are used on the left-hand side of `:=` (declaration) and `=` (assignment/initial binding) to determine how the value from the `Expression` is deconstructed and which identifiers are bound or assigned to.

#### 5.2.1. Identifier Pattern

The simplest pattern binds the entire result of the `Expression` to a single identifier.

*   **Syntax**: `Identifier`
*   **Usage (`:=`)**: Declares a new binding `Identifier`.
    ```able
    x := 42
    user_name := fetch_user_name()
    my_func := { a, b => a + b }
    ```
*   **Usage (`=`)**: Reassigns an existing mutable binding `Identifier`, or creates an initial binding if `Identifier` doesn't exist lexically.
    ```able
    x = 50 ## Reassigns existing x, or initial binding if x doesn't exist
    ```

#### 5.2.2. Wildcard Pattern (`_`)

The wildcard `_` matches any value but does not bind it to any identifier. It's used to ignore parts of a value during destructuring.

*   **Syntax**: `_`
*   **Usage**: Primarily used *within* composite patterns (structs, arrays). Using `_` as the entire pattern (`_ = Expression`) is allowed and simply evaluates the expression, discarding its result.
*   **Example**:
    ```able
    { x: _, y } := get_point() ## Declare y, ignore x
    [_, second, _] := get_three_items() ## Declare second, ignore first and third
    _ = function_with_side_effects() ## Evaluate function, ignore result
    ```

#### 5.2.3. Struct Pattern (Named Fields)

Destructures instances of structs defined with named fields.

*   **Syntax**: `[StructTypeName] { Field1: PatternA, Field2 @ BindingB: PatternC, ShorthandField, ... }`
    *   `StructTypeName`: Optional, the name of the struct type. If omitted, the type is inferred or checked against the expected type.
    *   `Field: Pattern`: Matches the value of `Field` against the nested `PatternA`.
    *   `Field @ Binding: Pattern`: Matches the value of `Field` against nested `PatternC` and binds the original field value to the new identifier `BindingB` (only valid with `:=`).
    *   `ShorthandField`: Equivalent to `ShorthandField: ShorthandField`. Binds the value of the field `ShorthandField` to an identifier with the same name.
    *   `...`: (Not currently supported) Might be considered to ignore remaining fields explicitly. Currently, extra fields in the value not mentioned in the pattern are ignored.
*   **Example**:
    ```able
    struct Point { x: f64, y: f64 }
    struct User { id: u64, name: String, address: String }

    p := Point { x: 1.0, y: 2.0 }
    u := User { id: 101, name: "Alice", address: "123 Main St" }

    ## Destructure Point (Declaration :=)
    Point { x: x_coord, y: y_coord } := p ## Declares x_coord=1.0, y_coord=2.0
    { x, y } := p                       ## Shorthand declaration, declares x=1.0, y=2.0

    ## Destructure User (Declaration :=)
    { id, name @ user_name, address: addr } := u
    ## Declares id=101, user_name="Alice", addr="123 Main St"

    ## Ignore fields (Declaration :=)
    { id, name: _ } := u ## Declares id=101, ignores name, ignores address implicitly

    ## Reassignment / Initial Binding Example (=)
    existing_x = 0.0 ## Assume initial binding or reassignment
    existing_y = 0.0 ## Assume initial binding or reassignment
    { x: existing_x, y: existing_y } = Point { x: 5.0, y: 6.0 } ## Assigns 5.0 to existing_x, 6.0 to existing_y
    { id: new_id, name: new_name } = u ## Initial binding for new_id, new_name (if they don't exist)
    ```
*   **Semantics**: Matches fields by name. If `StructTypeName` is present, checks if the `Expression` value is of that type. Fails if a field mentioned in the pattern doesn't exist in the value.

#### 5.2.4. Struct Pattern (Positional Fields / Named Tuples)

Destructures instances of structs defined with positional fields.

*   **Syntax**: `[StructTypeName] { Pattern0, Pattern1, ..., PatternN }`
    *   `StructTypeName`: Optional, the name of the struct type.
    *   `Pattern0, Pattern1, ...`: Patterns corresponding to the fields by their zero-based index. The number of patterns must match the number of fields defined for the struct type.
*   **Example**:
    ```able
    struct IntPair { i32, i32 }
    struct Coord3D { f64, f64, f64 }

    pair := IntPair { 10, 20 }
    coord := Coord3D { 1.0, -2.5, 0.0 }

    ## Declaration (:=)
    IntPair { first, second } := pair ## Declares first=10, second=20
    { x, y, z } := coord           ## Declares x=1.0, y=-2.5, z=0.0

    ## Ignore positional fields (Declaration :=)
    { _, y_val, _ } := coord       ## Declares y_val=-2.5

    ## Reassignment / Initial Binding Example (=)
    existing_a = 0 ## Assume initial binding or reassignment
    existing_b = 0 ## Assume initial binding or reassignment
    { existing_a, existing_b } = IntPair { 100, 200 } ## Assigns 100 to existing_a, 200 to existing_b
    { new_x, new_y, new_z } = coord ## Initial binding for new_x, new_y, new_z (if they don't exist)
    ```
*   **Semantics**: Matches fields by position. If `StructTypeName` is present, checks the type. Fails if the number of patterns does not match the number of fields in the value's type.

#### 5.2.5. Array Pattern

Destructures instances of the built-in `Array` type.

*   **Syntax**: `[Pattern1, Pattern2, ..., ...RestIdentifier]` or `[Pattern1, ..., ...]`
    *   `Pattern1, Pattern2, ...`: Patterns matching array elements by position from the start.
    *   `...RestIdentifier`: Optional. If present, matches all remaining elements *after* the preceding positional patterns and binds them as a *new* `Array` to `RestIdentifier` (only valid with `:=`).
    *   `...`: If `...` is used without an identifier, it matches remaining elements but does not bind them.
*   **Example**:
    ```able
    data := [10, 20, 30, 40]

    ## Declaration (:=)
    [a, b, c, d] := data ## Declares a=10, b=20, c=30, d=40
    [first, second, ...rest] := data ## Declares first=10, second=20, rest=[30, 40]
    [x, _, y, ...] := data ## Declares x=10, y=30, ignores second element and rest
    [single] := [99] ## Declares single=99
    [] := [] ## Matches an empty array (declares nothing)

    ## Reassignment / Initial Binding Example (=)
    existing_head = 0 ## Assume initial binding or reassignment
    ## Note: Assigning to a rest pattern with '=' is likely invalid or needs careful definition.
    ##       Typically, '=' would assign to existing elements by index/pattern.
    [existing_head, element_1] = [1, 2] ## Assigns 1 to existing_head, assigns 2 to element_1 (initial binding if needed)
    ```
*   **Semantics**: Matches elements by position. Fails if the array has fewer elements than required by the non-rest patterns.
*   **Mutability:** Array elements themselves are mutable (via index assignment `arr[idx] = val`). Requires `Array` type to support `IndexMut` interface (TBD).

#### 5.2.6. Nested Patterns

Patterns can be nested arbitrarily within struct and array patterns for both `:=` and `=`.

*   **Example (`:=`)**:
    ```able
    struct Point { x: f64, y: f64 } ## Assuming Point is defined elsewhere
    struct Data { id: u32, point: Point }
    struct Container { items: Array Data }

    val := Container { items: [ Data { id: 1, point: Point { x: 1.0, y: 2.0 } },
                               Data { id: 2, point: Point { x: 3.0, y: 4.0 } } ] }

    Container { items: [ Data { id: first_id, point: { x: first_x, y: _ }}, ...rest_data ] } := val
    ## Declares first_id = 1
    ## Declares first_x = 1.0
    ## Ignores y of the first point
    ## Declares rest_data = [ Data { id: 2, point: Point { x: 3.0, y: 4.0 } } ]
    ```

### 5.3. Semantics of Assignment/Declaration

1.  **Evaluation Order**: The `Expression` (right-hand side) is evaluated first to produce a value.
2.  **Matching & Binding/Assignment**: The resulting value is then matched against the `Pattern` or `LHS` (left-hand side).
    *   **`:=`**: Resulting value matched against `Pattern` (LHS). New mutable bindings created in current scope for identifiers introduced in the pattern. It is a compile-time error if any introduced identifier already exists as a binding *in the current scope*.
    *   **`=`**: Resulting value matched against `LHS` pattern/location specifier. Values assigned to existing, accessible, mutable bindings/locations identified within the LHS. It is a compile-time error if any target location does not exist, is not accessible, or is not mutable.
    *   **Match Failure**: If the value's structure or type does not match the pattern/LHS, a runtime error/panic occurs for both `:=` and `=`.
3.  **Mutability**: Bindings introduced via `:=` are mutable by default. The `=` operator requires the target location(s) (variable, field, index) to be mutable.
4.  **Scope**: `:=` introduces bindings into the current lexical scope. `=` operates on existing bindings/locations found according to lexical scoping rules.
5.  **Type Checking**: The compiler checks for compatibility between the type of the `Expression` and the structure expected by the `Pattern`/`LHS`. Type inference applies where possible.
6.  **Result Value**: Both assignment (`=`) and declaration (`:=`) expressions evaluate to the **value of the RHS** after successful binding/assignment.

## 6. Expressions

Units of code that evaluate to a value.

### 6.1. Literals

Literals are the source code representation of fixed values.

#### 6.1.1. Integer Literals

-   **Syntax:** A sequence of digits `0-9`. Underscores `_` can be included anywhere except the start/end for readability and are ignored. Prefixes `0x` (hex), `0o` (octal), `0b` (binary) are supported.
-   **Type:** By default, integer literals are inferred as `i32` (this default is configurable/TBC). Type suffixes can explicitly specify the type: `_i8`, `_u8`, `_i16`, `_u16`, `_i32`, `_u32`, `_i64`, `_u64`, `_i128`, `_u128`.
-   **Examples:** `123`, `0`, `1_000_000`, `42_i64`, `255_u8`, `0xff`, `0b1010_1111`, `0o777_i16`.

#### 6.1.2. Floating-Point Literals

-   **Syntax:** Include a decimal point (`.`) or use scientific notation (`e` or `E`). Underscores `_` are allowed for readability.
-   **Type:** By default, float literals are inferred as `f64`. The suffixes `_f32` and `_f64` explicitly denote the type.
-   **Examples:** `3.14`, `0.0`, `-123.456`, `1e10`, `6.022e23`, `2.718_f32`, `_1.618_`, `1_000.0`, `1.0_f64`.

#### 6.1.3. Boolean Literals

-   **Syntax:** `true`, `false`.
-   **Type:** `bool`.

#### 6.1.4. Character Literals

-   **Syntax:** A single Unicode character enclosed in single quotes `'`. Special characters can be represented using escape sequences:
    *   Common escapes: `\n` (newline), `\r` (carriage return), `\t` (tab), `\\` (backslash), `\'` (single quote), `\"` (double quote - though not strictly needed in char literal).
    *   Unicode escape: `\u{XXXXXX}` where `XXXXXX` are 1-6 hexadecimal digits representing the Unicode code point.
-   **Type:** `char`.
-   **Examples:** `'a'`, `' '`, `'%'`, `'\n'`, `'\u{1F604}'`.

#### 6.1.5. String Literals

-   **Syntax:**
    1.  **Standard:** Sequence of characters enclosed in double quotes `"`. Supports the same escape sequences as character literals.
    2.  **Interpolated:** Sequence of characters enclosed in backticks `` ` ``. Can embed expressions using `${Expression}`. Escapes like `` \` `` and `\$` are used for literal backticks or dollar signs before braces. See Section [6.6](#66-string-interpolation).
-   **Type:** `string`. Strings are immutable.
-   **Examples:** `"Hello, world!\n"`, `""`, `` `User: ${user.name}, Age: ${user.age}` ``, `` `Literal: \` or \${` ``.

#### 6.1.6. Nil Literal

-   **Syntax:** `nil`.
-   **Type:** `nil`. The type `nil` has only one value, also written `nil`.
-   **Usage:** Represents the absence of meaningful data. Often used with the `?Type` (equivalent to `nil | Type`) union shorthand. `nil` itself *only* has type `nil`, but can be assigned to variables of type `?SomeType`.

#### 6.1.7. Void Type (No Literal)

-   **Type Name:** `void`.
-   **Values:** The `void` type represents the empty set; it has **no values**.
-   **Usage:** Primarily used as a return type for functions that perform actions (side effects) but do not produce any resulting data. It signifies successful completion without a value.
-   **Distinction from `nil`:** The type `nil` has one value (`nil`); the type `void` has zero values.

#### 6.1.8. Other Literals (Conceptual)

-   **Arrays:** `[1, 2, 3]`, `["a", "b"]`, `[]`. (Requires `Array` type definition in stdlib).
-   **Structs:** `{ field: val }`, `{ val1, val2 }`. (See Section [4.5](#45-structs)).

### 6.2. Block Expressions (`do`)

Execute a sequence of expressions within a new lexical scope.

*   **Syntax:** `do { ExpressionList }`
*   **Semantics:** Expressions evaluated sequentially. Introduces scope. Evaluates to the value of the last expression.
*   **Example:** `result := do { x := f(); y := g(x); x + y }`

### 6.3. Operators

This section defines the standard operators available in Able, their syntax, semantics, precedence, and associativity, closely following Rust's precedence model but using `~` for bitwise NOT.

#### 6.3.1. Operator Precedence and Associativity

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
| 1          | `:=`                  | **Declaration and Initialization**      | Right-to-left | See Section [5.1](#51-operators---)                     |
| 1          | `=`                   | **Reassignment / Mutation**             | Right-to-left | See Section [5.1](#51-operators---)                     |
| 1          | `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=` | Compound Assignment (TBD)               | Right-to-left | (Needs formal definition, acts like `=`)                  |
| 0          | `\|>`                 | Pipe Forward                            | Left-to-right | (Lowest precedence)                                       |

*(Note: Precedence levels are relative; specific numerical values may vary but the order shown is based on Rust.)*

#### 6.3.2. Operator Semantics

*   **Arithmetic (`+`, `-`, `*`, `/`, `%`):** Standard math operations on numeric types. Division (`/`) or remainder (`%`) by zero **raises a runtime exception** (e.g., `DivisionByZeroError`). See Section [11.3](#113-exceptions-raise--rescue).
*   **Comparison (`>`, `<`, `>=`, `<=`, `==`, `!=`):** Compare values, result `bool`. Equality/ordering behavior relies on standard library interfaces (`PartialEq`, `Eq`, `PartialOrd`, `Ord`). See Section [14](#14-standard-library-interfaces-conceptual--tbd).
*   **Logical (`&&`, `||`, `!`):**
    *   `&&` (Logical AND): Short-circuiting on `bool` values.
    *   `||` (Logical OR): Short-circuiting on `bool` values.
    *   `!` (Logical NOT): Unary operator, negates a `bool` value.
*   **Bitwise (`&`, `|`, `^`, `<<`, `>>`, `~`):**
    *   `&`, `|`, `^`: Standard bitwise AND, OR, XOR on integer types (`i*`, `u*`).
    *   `<<`, `>>`: Bitwise left shift, right shift on integer types.
    *   `~` (Bitwise NOT): Unary operator, performs bitwise complement on integer types.
*   **Unary (`-`):** Arithmetic negation for numeric types.
*   **Member Access (`.`):** Access fields/methods, UFCS, static methods. See Section [9.3](#93-method-call-syntax-resolution-initial-rules).
*   **Function Call (`()`):** Invokes functions/methods. See Section [7.4](#74-function-invocation).
*   **Indexing (`[]`):** Access elements within indexable collections (e.g., `Array`). Relies on standard library interfaces (`Index`, `IndexMut`). See Section [14](#14-standard-library-interfaces-conceptual--tbd).
*   **Range (`..`, `...`):** Create `Range` objects (inclusive `..`, exclusive `...`). See Section [8.2.3](#823-range-expressions).
*   **Declaration (`:=`):** Declares/initializes new variables. Evaluates to RHS. See Section [5.1](#51-operators---).
*   **Assignment (`=`):** Reassigns existing variables or mutates locations. Evaluates to RHS. See Section [5.1](#51-operators---).
*   **Compound Assignment (`+=`, etc. TBD):** Shorthand (e.g., `a += b` is like `a = a + b`). Need formal definition. Acts like `=`.
*   **Pipe Forward (`|>`):** `x |> f` evaluates to `f(x)`.

#### 6.3.3. Overloading (Via Interfaces)

Behavior for non-primitive types relies on implementing standard library interfaces (e.g., `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg` (for `-`), `Not` (for bitwise `~`), `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`, `PartialEq`, `Eq`, `PartialOrd`, `Ord`, `Index`, `IndexMut`). These interfaces need definition (See Section [14](#14-standard-library-interfaces-conceptual--tbd)). Note that logical `!` is typically not overloaded.

### 6.4. Function Calls

See Section [7.4](#74-function-invocation).

### 6.5. Control Flow Expressions

`if/or`, `match`, `breakpoint`, `rescue`, `do`, `:=`, `=` evaluate to values. See Section [8](#8-control-flow) and Section [11](#11-error-handling). Loops (`while`, `for`) evaluate to `nil`.

### 6.6. String Interpolation

`` `Literal text ${Expression} more text` ``

*   Evaluates embedded `Expression`s (converting via `Display` interface).
*   Concatenates parts into a final `string`.
*   Escape `` \` `` and `\$`.

## 7. Functions

This section defines the syntax and semantics for function definition, invocation, partial application, and related concepts like closures and anonymous functions in Able. Functions are first-class values.

### 7.1. Named Function Definition

Defines a function with a specific identifier in the current scope.

#### 7.1.1. Syntax
```able
fn Identifier[<GenericParamList>] ([ParameterList]) [-> ReturnType] [where <ConstraintList>] {
  ExpressionList
}
```

-   **`fn`**: Keyword introducing a function definition.
-   **`Identifier`**: The function name (e.g., `add`, `process_data`).
-   **`[<GenericParamList>]`**: Optional **comma-delimited** generic parameters and constraints enclosed in angle brackets (e.g., `<T>`, `<T: Display>`, `<A, B, C>`). Constraints can be specified inline here or in the `where` clause.
-   **`([ParameterList])`**: Required parentheses enclosing the parameter list.
    -   **`ParameterList`**: Comma-separated list of parameters, each defined as `Identifier: Type` (e.g., `a: i32`, `user: User`). Type annotations are generally required unless future inference rules allow omission.
    -   May be empty: `()`.
-   **`[-> ReturnType]`**: Optional return type annotation. If omitted, the return type is inferred from the body's final expression or explicit `return` statements. If the body's last expression evaluates to `nil` (e.g., assignment, loop) and there's no explicit `return`, the return type is `nil`. If the function is intended to return nothing, use `-> void`.
-   **`[where <ConstraintList>]`**: Optional clause placed after the return type (or parameter list if no return type) to specify constraints on `GenericParamList`.
-   **`{ ExpressionList }`**: The function body block. Contains one or more expressions separated by newlines or semicolons.
    -   **Return Value**: The value of the *last expression* in `ExpressionList` is implicitly returned, *unless* an explicit `return` statement is encountered.

#### 7.1.2. Examples
```able
## Simple function (implicit return)
fn add(a: i32, b: i32) -> i32 { a + b }

## Generic function (implicit return)
fn identity<T>(val: T) -> T { val }

## Function with side effects and explicit void return
fn greet(name: String) -> void {
  message = `Hello, ${name}!`
  print(message) ## Assuming print returns nil or void
  return ## Explicit return void
}

## Function with side effects and inferred nil return type
fn log_and_nil(name: String) { ## Implicitly returns nil
  message = `Logging: ${name}`
  print(message)
}


## Multi-expression body (implicit return)
fn process(x: i32) -> String {
  y = x * 2
  z = y + 1
  `Result: ${z}` ## Last expression is the return value
}
```

#### 7.1.3. Semantics
-   Introduces the `Identifier` into the current scope, bound to the defined function value.
-   Parameters are bound to argument values during invocation and are local to the function body scope.
-   The function body executes sequentially.
-   The type of a function is `(ParamType1, ParamType2, ...) -> ReturnType`.

### 7.2. Anonymous Functions and Closures

Functions can be created without being bound to a name at definition time. They capture their lexical environment (forming closures).

#### 7.2.1. Verbose Anonymous Function Syntax

Mirrors named function definition but omits the identifier. Useful for complex lambdas or when explicit generics are needed.

##### Syntax
```able
fn[<GenericParamList>] ([ParameterList]) [-> ReturnType] { ExpressionList }
```

##### Example
```able
mapper = fn(x: i32) -> String { `Value: ${x}` }
generic_fn = fn<T: Display>(item: T) -> void { print(item.to_string()) }
```

#### 7.2.2. Lambda Expression Syntax

Concise syntax, primarily for single-expression bodies.

##### Syntax
```able
{ [LambdaParameterList] [-> ReturnType] => Expression }
```
-   **`{ ... }`**: Lambda delimiters.
-   **`[LambdaParameterList]`**: Comma-separated identifiers, optional types (`ident: Type`). No parentheses used. Zero parameters represented by empty list before `=>`.
-   **`[-> ReturnType]`**: Optional return type.
-   **`=>`**: Separator.
-   **`Expression`**: Single expression defining the return value.

##### Examples
```able
increment = { x => x + 1 }
adder = { x: i32, y: i32 => x + y }
get_zero = { => 0 }
complex_lambda = { x, y => do { temp = x + y; temp * temp } } ## Using a block expression
```

#### 7.2.3. Closures

Both anonymous function forms create closures. They capture variables from the scope where they are defined. Captured variables are accessed according to the mutability rules of the original binding (currently mutable by default).

```able
fn make_adder(amount: i32) -> (i32 -> i32) {
  adder_lambda = { value => value + amount } ## Captures 'amount'
  ## Explicit return needed here as last expression is an assignment
  return adder_lambda
}
add_5 = make_adder(5)
result = add_5(10) ## result is 15
```

### 7.3. Explicit `return` Statement

Provides early exit from a function. (See also Section [11.1](#111-explicit-return-statement)).

#### Syntax
```able
return Expression
return // Equivalent to 'return void' if function returns void
```

#### Semantics
-   Immediately terminates the execution of the current function.
-   The value of `Expression` (or `void`) is returned to the caller.
-   If used within nested blocks (like loops or `do` blocks) inside a function, it still returns from the *function*, not just the inner block.

### 7.4. Function Invocation

#### 7.4.1. Standard Call

Parentheses enclose comma-separated arguments.
```able
Identifier ( ArgumentList )
```
```able
add(5, 3)
identity<String>("hello") ## Explicit generic argument
```

#### 7.4.2. Trailing Lambda Syntax

```able
Function ( [OtherArgs] ) LambdaExpr
Function LambdaExpr ## If lambda is only argument
```

If the last argument is a lambda, it can follow the closing parenthesis. If it's the *only* argument, parentheses can be omitted.
```able
items.reduce(0) { acc, x => acc + x }
items.map { item => item.process() }
```

#### 7.4.3. Method Call Syntax

Allows calling functions (both inherent/interface methods and qualifying free functions) using dot notation on the first argument. (See Section [9.3](#93-method-call-syntax-resolution-initial-rules) for resolution details).

##### Syntax
```able
ReceiverExpression . FunctionOrMethodName ( RemainingArgumentList )
```

##### Semantics (Simplified - see Section 9.3 for full rules)
When `receiver.name(args...)` is encountered:
1.  Check for field `name`.
2.  Check for inherent method `name`.
3.  Check for interface method `name`.
4.  Check for free function `name` applicable via UFCS (Universal Function Call Syntax).
5.  Invoke the first match found, passing `receiver` appropriately.
6.  Ambiguity or no match results in an error.

##### Example (Method Call Syntax on Free Function)
```able
fn add(a: i32, b: i32) -> i32 { a + b }
res = 4.add(5) ## Resolved via Method Call Syntax to add(4, 5) -> 9
```

#### 7.4.4. Callable Value Invocation (`Apply` Interface)

If `value` implements the `Apply` interface, `value(args...)` desugars to `value.apply(args...)`. (See Section [14](#14-standard-library-interfaces-conceptual--tbd)).
```able
## Conceptual Example
# impl Apply for Integer { fn apply(self: Integer, a: Integer) -> Integer { self * a } }
# thirty = 5(6) ## Calls 5.apply(6)
```

### 7.5. Partial Function Application

Create a new function by providing some arguments and using `_` as a placeholder for others.

#### 7.5.1. Syntax
Use `_` in place of arguments in a function or method call expression.
```able
function_name(Arg1, _, Arg3, ...)
instance.method_name(_, Arg2, ...)
```

#### 7.5.2. Syntax & Semantics
-   `function_name(Arg1, _, ...)` creates a closure.
-   `receiver.method_name(_, Arg2, ...)` creates a closure capturing `receiver`.
-   `TypeName::method_name(_, Arg2, ...)` (if static access is needed/allowed) creates a closure expecting `self` as the first argument.
-   `receiver.free_function_name` (using Method Call Syntax access without `()`) creates a closure equivalent to `free_function_name(receiver, _, ...)`.

#### 7.5.3. Examples
```able
add_10 = add(_, 10)      ## Function expects one arg: add(arg, 10)
result = add_10(5)       ## result is 15

## Assuming prepend exists: fn prepend(prefix: string, body: string) -> string
# prefix_hello = prepend("Hello, ", _) ## Function expects one arg
# msg = prefix_hello("World")          ## msg is "Hello, World"

## method call syntax access creates partially applied function
add_five = 5.add ## Creates function add(5, _) via Method Call Syntax access
result_pa = add_five(20)  ## result_pa is 25
```

### 7.6. Shorthand Notations

#### 7.6.1. Implicit First Argument Access (`#member`)

Within the body of any function (named, anonymous, lambda, or method), the syntax `#Identifier` provides shorthand access to a field or method of the function's *first parameter*.

##### Syntax
```able
#Identifier
```

##### Semantics
-   Syntactic sugar for `param1.Identifier`, where `param1` is the **first parameter** of the function the `#member` expression appears within.
-   If the function has *no* parameters, using `#member` is a compile-time error.
-   Inside a function `fn func_name(param1: Type1, param2: Type2, ...) { ... }`, an expression `#member` within the function body is syntactic sugar for `param1.member`.
-   This relies on the *convention* that the first parameter often represents the primary object or context (`self`).
-   The `param1` value must have a field or method named `member` accessible via the dot (`.`) operator.
-   This applies regardless of whether the first parameter is explicitly named `self`.

##### Example
```able
struct Data { value: i32, name: string }
methods Data {
    ## Inside an instance method, #value means self.value
    fn display(self: Self) -> void {
        print(`Data '${#name}' has value ${#value}`)
    }
}

## Inside a free function
fn process_data(d: Data, factor: i32) -> i32 {
  ## #value is shorthand for d.value
  incremented = #value + 1
  incremented * factor
}

d = Data { value: 10, name: "Test" }
d.display() ## Prints "Data 'Test' has value 10"
result = process_data(d, 5) ## result is (10 + 1) * 5 = 55
```

#### 7.6.2. Implicit Self Parameter Definition (`fn #method`)

**Allowed only when defining functions within a `methods TypeName { ... }` block or an `impl Interface for Type { ... }` block.**

##### Syntax
```able
fn #method_name ([param2: Type2, ...]) [-> ReturnType] { ... }
```

##### Semantics
-   Syntactic sugar for defining an **instance method**. Automatically adds `self: Self` as the first parameter.
-   `fn #method(p2) { ... }` is equivalent to `fn method(self: Self, p2) { ... }`.
-   `Self` refers to the type the `methods` or `impl` block is for.

##### Example
```able
struct Counter { value: i32 }
methods Counter {
  ## Define increment using shorthand
  fn #increment() -> void {
    #value = #value + 1 ## #value means self.value
  }

  ## Equivalent explicit definition:
  # fn increment(self: Self) -> void {
  #  self.value = self.value + 1
  # }

  ## Define add using shorthand
  fn #add(amount: i32) -> void {
    #value = #value + amount
  }
}

c = Counter { value: 5 }
c.increment() ## c.value becomes 6
c.add(10)     ## c.value becomes 16
```

## 8. Control Flow

This section details the constructs Able uses to control the flow of execution, including conditional branching, pattern matching, looping, range expressions, and non-local jumps.

### 8.1. Branching Constructs

Branching constructs allow choosing different paths of execution based on conditions or patterns. Both `if/or` and `match` are expressions.

#### 8.1.1. Conditional Chain (`if`/`or`)

This construct evaluates conditions sequentially and executes the block associated with the first true condition. It replaces traditional `if/else if/else`.

##### Syntax

```able
if Condition1 { ExpressionList1 }
[or Condition2 { ExpressionList2 }]
...
[or ConditionN { ExpressionListN }]
[or { DefaultExpressionList }] // Final 'or' without condition acts as 'else'
```

-   **`if Condition1 { ExpressionList1 }`**: Required start. Executes `ExpressionList1` if `Condition1` (`bool`) is true.
-   **`or ConditionX { ExpressionListX }`**: Optional clauses. Executes `ExpressionListX` if its `ConditionX` (`bool`) is the first true condition in the chain.
-   **`or { DefaultExpressionList }`**: Optional final default block, executed if no preceding conditions were true.
-   **`ExpressionList`**: Sequence of expressions; the last expression's value is the result of the block.

##### Semantics

1.  **Sequential Evaluation**: Conditions are evaluated strictly in order.
2.  **First True Wins**: Execution stops at the first true `ConditionX`. The corresponding `ExpressionListX` is executed, and its result becomes the value of the `if/or` chain.
3.  **Default Clause**: Executes if no conditions are true and the clause exists.
4.  **Result Value**: The `if/or` chain evaluates to the result of the executed block. If no block executes (no conditions true and no default `or {}`), it evaluates to `nil`.
5.  **Type Compatibility**:
    *   If a default `or {}` guarantees execution, all result expressions must have compatible types. The chain's type is this common type.
    *   If no default `or {}` exists, non-`nil` results must be compatible. The chain's type is `?CompatibleType`.

##### Example

```able
grade = if score >= 90 { "A" }
        or score >= 80 { "B" }
        or { "C or lower" } ## Guarantees String result
```

#### 8.1.2. Pattern Matching Expression (`match`)

Selects a branch by matching a subject expression against a series of patterns, executing the code associated with the first successful match. `match` is an expression.

##### Syntax

```able
SubjectExpression match {
  case Pattern1 [if Guard1] => ResultExpressionList1
  [ , case Pattern2 [if Guard2] => ResultExpressionList2 ]
  ...
  [ , case PatternN [if GuardN] => ResultExpressionListN ]
  [ , case _ => DefaultResultExpressionList ] ## Optional wildcard clause
}
```

-   **`SubjectExpression`**: The value to be matched.
-   **`match`**: Keyword initiating matching.
-   **`{ ... }`**: Block containing match clauses separated by commas `,`.
-   **`case PatternX [if GuardX] => ResultExpressionListX`**: A match clause.
    *   **`case`**: Keyword.
    *   **`PatternX`**: Pattern to match (Literal, Identifier, `_`, Type/Variant, Struct `{}`, Array `[]`). Bound variables are local to this clause. See Section [5.2](#52-patterns).
    *   **`[if GuardX]`**: Optional `bool` guard expression using pattern variables.
    *   **`=>`**: Separator.
    *   **`ResultExpressionListX`**: Expressions executed if clause chosen; last expression's value is the result.

##### Semantics

1.  **Sequential Evaluation**: `SubjectExpression` evaluated once. `case` clauses checked top-to-bottom.
2.  **First Match Wins**: The first `PatternX` that matches *and* whose `GuardX` (if present) is true selects the clause.
3.  **Execution & Result**: The chosen `ResultExpressionListX` is executed. The `match` expression evaluates to the value of the last expression in that list.
4.  **Exhaustiveness**: Compiler SHOULD check for exhaustiveness (especially for unions). Non-exhaustive matches MAY warn/error at compile time and SHOULD panic at runtime. A `case _ => ...` usually ensures exhaustiveness.
5.  **Type Compatibility**: All `ResultExpressionListX` must yield compatible types. The `match` expression's type is this common type.

##### Example

```able
## Assuming Option T = T | nil and Some is a struct { value: T }
description = maybe_num match {
  case Some { value: x } if x > 0 => `Positive: ${x}`,
  case Some { value: 0 } => "Zero",
  case Some { value: x } => `Negative: ${x}`,
  case nil => "Nothing"
}
```

### 8.2. Looping Constructs

Loops execute blocks of code repeatedly. Loop expressions (`while`, `for`) evaluate to `nil`.

#### 8.2.1. While Loop (`while`)

Repeats execution as long as a condition is true.

##### Syntax

```able
while Condition {
  BodyExpressionList
}
```

-   **`while`**: Keyword.
-   **`Condition`**: `bool` expression evaluated before each iteration.
-   **`{ BodyExpressionList }`**: Loop body executed if `Condition` is true.

##### Semantics

-   `Condition` checked. If `true`, body executes. Loop repeats. If `false`, loop terminates.
-   Always evaluates to `nil`.
-   Loop exit occurs when `Condition` is false or via a non-local jump (`break`).

##### Example

```able
counter = 0
while counter < 3 {
  print(counter)
  counter = counter + 1
} ## Prints 0, 1, 2. Result is nil.
```

#### 8.2.2. For Loop (`for`)

Iterates over a sequence produced by an expression whose type implements the `Iterable` interface.

##### Syntax

```able
for Pattern in IterableExpression {
  BodyExpressionList
}
```

-   **`for`**: Keyword.
-   **`Pattern`**: Pattern to bind/deconstruct the current element yielded by the iterator. See Section [5.2](#52-patterns).
-   **`in`**: Keyword.
-   **`IterableExpression`**: Expression evaluating to a value implementing `Iterable T` for some `T`.
-   **`{ BodyExpressionList }`**: Loop body executed for each element. Pattern bindings are available.

##### Semantics

-   The `IterableExpression` produces an iterator (details governed by the `Iterable` interface implementation, see Section [14](#14-standard-library-interfaces-conceptual--tbd)).
-   The body executes once per element yielded by the iterator, matching the element against `Pattern`.
-   Always evaluates to `nil`.
-   Loop terminates when the iterator is exhausted or via a non-local jump (`break`).
-   If an element yielded by the iterator does not match the `Pattern`, a runtime error/panic occurs.

##### Example

```able
items = ["a", "b"] ## Array implements Iterable
for item in items { print(item) } ## Prints "a", "b"

total = 0
for i in 1..3 { ## Range 1..3 implements Iterable
  total = total + i
} ## total becomes 6 (1+2+3)
```

#### 8.2.3. Range Expressions

Provide a concise way to create iterable sequences of integers.

##### Syntax

```able
StartExpr .. EndExpr   // Inclusive range [StartExpr, EndExpr]
StartExpr ... EndExpr  // Exclusive range [StartExpr, EndExpr)
```

-   **`StartExpr`, `EndExpr`**: Integer expressions.
-   **`..` / `...`**: Operators creating range values.

##### Semantics

-   Syntactic sugar for creating values (e.g., via `Range.inclusive`/`Range.exclusive`) that implement the `Iterable` interface (and likely a `Range` interface). See Section [14](#14-standard-library-interfaces-conceptual--tbd).

### 8.3. Non-Local Jumps (`breakpoint` / `break`)

Provides a mechanism for early exit from a designated block (`breakpoint`), returning a value and unwinding the call stack if necessary. Replaces traditional `break`/`continue`.

#### 8.3.1. Defining an Exit Point (`breakpoint`)

Marks a block that can be exited early. `breakpoint` is an expression.

##### Syntax

```able
breakpoint 'LabelName {
  ExpressionList
}
```

-   **`breakpoint`**: Keyword.
-   **`'LabelName`**: A label identifier (single quote prefix) uniquely naming this point within its lexical scope.
-   **`{ ExpressionList }`**: The block of code associated with the breakpoint.

#### 8.3.2. Performing the Jump (`break`)

Initiates an early exit targeting a labeled `breakpoint` block.

##### Syntax

```able
break 'LabelName ValueExpression
```

-   **`break`**: Keyword.
-   **`'LabelName`**: The label identifying the target `breakpoint` block. Must match a lexically enclosing `breakpoint`. Compile error if not found.
-   **`ValueExpression`**: Expression whose result becomes the value of the exited `breakpoint` block.

#### 8.3.3. Semantics

1.  **`breakpoint` Block Execution**:
    *   Evaluates `ExpressionList`.
    *   If execution finishes normally, the `breakpoint` expression evaluates to the result of the *last expression* in `ExpressionList`.
    *   If a `break 'LabelName ...` targeting this block occurs during execution (possibly in nested calls), execution stops immediately.
2.  **`break` Execution**:
    *   Finds the innermost lexically enclosing `breakpoint` with the matching `'LabelName`.
    *   Evaluates `ValueExpression`.
    *   Unwinds the call stack up to the target `breakpoint` block.
    *   Causes the target `breakpoint` expression itself to evaluate to the result of `ValueExpression`.
3.  **Type Compatibility**: The type of the `breakpoint` expression must be compatible with both the type of its block's final expression *and* the type(s) of the `ValueExpression`(s) from any `break` statements targeting it.

#### 8.3.4. Example

```able
search_result = breakpoint 'finder {
  data = [1, 5, -2, 8]
  for item in data {
    if item < 0 {
      break 'finder item ## Exit early, return the negative item
    }
    ## ... process positive items ...
  }
  nil ## Default result if loop completes without breaking
}
## search_result is -2
```

## 9. Inherent Methods (`methods`)

Define methods (instance or static) directly associated with a specific struct type using a `methods` block. This is distinct from implementing interfaces (Section [10](#10-interfaces-and-implementations)).

### 9.1. Syntax

```able
methods [GenericParams] TypeName [GenericArgs] {
  [FunctionDefinitionList]
}
```
-   **`methods`**: Keyword initiating the block for defining inherent methods.
-   **`[GenericParams]`**: Optional generics `<...>` for the block itself (rare).
-   **`TypeName`**: The struct type name (e.g., `Point`, `User`).
-   **`[GenericArgs]`**: Generic arguments if `TypeName` is generic (e.g., `methods Pair A B { ... }`).
-   **`{ [FunctionDefinitionList] }`**: Contains standard `fn` definitions.

### 9.2. Method Definitions

Within a `methods TypeName { ... }` block:

1.  **Instance Methods:** Operate on an instance of `TypeName`. Defined using:
    *   Explicit `self`: `fn method_name(self: Self, ...) { ... }`
    *   Shorthand `fn #`: `fn #method_name(...) { ... }` (implicitly adds `self: Self` as the first parameter). See Section [7.6.2](#762-implicit-self-parameter-definition-fn-method).
    *   `Self` refers to `TypeName` (with its generic arguments, if any).
2.  **Static Methods:** Associated with the type itself, not a specific instance. Defined *without* `self` as the first parameter and *without* using the `#` prefix shorthand.
    *   `fn static_name(...) { ... }`
    *   Often used for constructors or type-level operations.

### 9.3. Example: `methods` block for `Address`

```able
struct Address { house_number: u16, street: string, city: string, state: string, zip: u16 }

methods Address {
  ## Instance method using shorthand definition and access
  fn #to_s() -> string {
    ## #house_number is shorthand for self.house_number, etc. See Section 7.6.1
    `${#house_number} ${#street}\n${#city}, ${#state} ${#zip}`
  }

  ## Instance method using explicit self
  fn update_zip(self: Self, zip_code: u16) -> void {
    self.zip = zip_code ## Could also use #zip here
  }

  ## Static method (constructor pattern)
  fn from_parts(hn: u16, st: string, ct: string, sta: string, zp: u16) -> Self {
    Address { house_number: hn, street: st, city: ct, state: sta, zip: zp }
  }
}

## Usage
addr = Address.from_parts(123, "Main St", "Anytown", "CA", 90210) ## Call static method
addr_string = addr.to_s()     ## Call instance method
addr.update_zip(90211)        ## Call instance method (mutates addr)
```

### 9.4. Method Call Syntax Resolution (Initial Rules)

When resolving `receiver.name(args...)`:
1.  Check for **field** `name` on the `receiver`. If found and callable (implements `Apply`), call it. If found and not callable, error (unless accessing the field value was the intent).
2.  Check for **inherent method** `name` defined in a `methods TypeName { ... }` block for the type of `receiver`.
3.  Check for **interface method** `name` from interfaces implemented by the type of `receiver`. Use specificity rules (See Section [10.2.4](#1024-overlapping-implementations-and-specificity)) if multiple interfaces provide `name`.
4.  Check for **free function** `name` in scope that takes `receiver` as its first argument (Universal Function Call Syntax - UFCS).
5.  If multiple steps match (e.g., inherent and interface method), inherent methods typically take precedence (TBD - confirm precedence rules).
6.  If ambiguity remains or no match is found, result in a compile-time error.

## 10. Interfaces and Implementations

This section defines **interfaces**, which specify shared functionality (contracts), and **implementations** (`impl`), which provide the concrete behavior for specific types or type constructors conforming to an interface. This system enables polymorphism, code reuse, and abstraction.

### 10.1. Interfaces

An interface defines a contract consisting of a set of function signatures. A type or type constructor fulfills this contract by providing implementations for these functions.

#### 10.1.1. Interface Usage Models

Interfaces serve two primary purposes:

1.  **As Constraints:** They restrict the types allowable for generic parameters, ensuring those types provide the necessary functionality defined by the interface. This is compile-time polymorphism.
    ```able
    fn print_item<T: Display>(item: T) { ... } ## T must implement Display
    ```
2.  **As Types (Lens/Existential Type):** An interface name can be used as a type itself, representing *any* value whose concrete type implements the interface. This allows treating different concrete types uniformly through the common interface. This typically involves dynamic dispatch.
    ```able
    ## Assuming Circle and Square implement Display
    shapes: Array (Display) = [Circle{...}, Square{...}] ## Array holds values seen through the Display "lens"
    for shape in shapes {
        print(shape.to_string()) ## Calls the appropriate to_string via dynamic dispatch
    }
    ```
    *(Note: The exact syntax and mechanism for interface types like `Array (Display)` or `dyn Display` need further specification, but the concept is adopted. See Section [10.3.4](#1034-interface-types-dynamic-dispatch).)*

#### 10.1.2. Interface Definition

Interfaces are defined using the `interface` keyword. There are two primary forms:

**Syntax Form 1 (Explicit Self Type Pattern):**

```able
interface InterfaceName [GenericParamList] for SelfTypePattern [where <ConstraintList>] {
  [FunctionSignatureList]
}
```

-   **`interface`**: Keyword.
-   **`InterfaceName`**: The identifier naming the interface (e.g., `Display`, `Mappable`).
-   **`[GenericParamList]`**: Optional space-delimited generic parameters for the interface itself (e.g., `T`, `K V`). Constraints use `Param: Bound` inline or in the `where` clause.
-   **`for`**: Keyword introducing the self type pattern (mandatory in this form).
-   **`SelfTypePattern`**: Specifies the structure of the type(s) that can implement this interface. This defines the meaning of `Self` (see Section [10.1.3](#1013-self-keyword-interpretation)).
    *   **Concrete Type:** `TypeName [TypeArguments]` (e.g., `for Point`, `for Array i32`).
    *   **Generic Type Variable:** `TypeVariable` (e.g., `for T`).
    *   **Type Constructor (HKT):** `TypeConstructor _ ...` (e.g., `for M _`, `for Map K _`). `_` denotes unbound parameters.
    *   **Generic Type Constructor:** `TypeConstructor TypeVariable ...` (e.g., `for Array T`).
-   **`[where <ConstraintList>]`**: Optional clause for specifying constraints on generic parameters used in `GenericParamList` or `SelfTypePattern`.
-   **`{ [FunctionSignatureList] }`**: Block containing function signatures (methods).
    *   Each signature follows `fn name[<MethodGenerics>]([Param1: Type1, ...]) -> ReturnType;`.
    *   Methods can be instance methods (typically taking `self: Self` as the first parameter) or static methods (no `self: Self` parameter).
    *   Interface definitions may include default method implementations: `fn name(...) { ... }`. Implementations of the interface are allowed to override these default definitions.

**Syntax Form 2 (Implicit Self Type):**

```able
interface InterfaceName [GenericParamList] [where <ConstraintList>] {
  [FunctionSignatureList]
}
```

-   This form omits the `for SelfTypePattern` clause.
-   **`[GenericParamList]`**: Optional space-delimited generic parameters for the interface itself. Constraints use `Param: Bound` inline or in the `where` clause.
-   **`[where <ConstraintList>]`**: Optional clause for specifying constraints on `GenericParamList`.
-   **Semantics:**
    *   The `Self` type within the `FunctionSignatureList` refers directly to the concrete type that implements the interface.
    *   Implementations (`impl`) for this form of interface **cannot** target a type constructor (like `Array` or `Map K _`). They must target specific types (like `i32`, `Point`, `Array i32`, or a generic type variable `T` constrained elsewhere). This form is unsuitable for defining HKT interfaces like `Mappable`.

#### 10.1.3. `Self` Keyword Interpretation

Within the interface definition (and corresponding `impl` blocks):

*   **If `interface ... for SelfTypePattern ...` is used:**
    *   If `SelfTypePattern` is `MyType Arg1`, `Self` means `MyType Arg1`.
    *   If `SelfTypePattern` is `T`, `Self` means `T`.
    *   If `SelfTypePattern` is `MyConstructor _`, `Self` refers to the constructor `MyConstructor`. `Self Arg` means `MyConstructor Arg`. This is used for HKTs.
    *   If `SelfTypePattern` is `MyConstructor T`, `Self` means `MyConstructor T`.
*   **If `interface ... { ... }` (without `for`) is used:**
    *   `Self` refers to the concrete type provided in the `impl ... for Target` block (e.g., if `impl MyIface for i32`, then `Self` is `i32` within that impl).

#### 10.1.4. Examples of Interface Definitions

```able
## Form 1: Explicit 'for T' (Generic over implementing type)
interface Display for T {
  fn to_string(self: Self) -> String;
}

interface Clone for T {
  fn clone(self: Self) -> Self;
}

## Form 1: Explicit 'for Array i32' (Specific generic type application)
interface IntArrayOps for Array i32 {
  fn sum(self: Self) -> i32;
}

## Form 1: Explicit 'for M _' (HKT - Type Constructor)
interface Mappable A for M _ {
  fn map<B>(self: Self A, f: (A -> B)) -> Self B; ## Self=M, Self A = M A
}

## Form 1: Static method example
interface Zeroable for T {
  fn zero() -> Self; ## Static method, returns an instance of the implementing type T
}

## Form 2: Implicit Self Type (Cannot be used for HKTs)
interface Hashable {
  fn hash(self: Self) -> u64; ## Self will be the implementing type (e.g., i32, Point)
}

## Form 2: Interface with default implementation
interface Greeter {
    fn greet(self: Self) -> string { "Hello!" } ## Default implementation
}
```

#### 10.1.5. Composite Interfaces (Interface Aliases)

Define an interface as a combination of other interfaces.

**Syntax:**
```able
interface NewInterfaceName [GenericParams] [for SelfTypePattern] = Interface1 [Args] + Interface2 [Args] + ...
```
-   Implementing `NewInterfaceName` requires implementing all constituent interfaces (`Interface1`, `Interface2`, etc.).
-   *(TBD: The exact rules for how `for` clauses (or their absence) interact across constituents and the composite definition need clarification. Assume for now that if a `for` clause is present, it must be consistent across all constituents that require one. If constituents use the implicit self type form, the composite likely inherits that semantic.)*

**Example:**
```able
## Assuming Reader, Writer interfaces exist (likely using 'for T' or implicit self)
interface ReadWrite = Reader + Writer

## Assuming Display/Clone are defined 'for T'
interface DisplayClone for T = Display + Clone

## Assuming Hashable/Eq use implicit self type
interface HashableEq = Hashable + Eq
```

### 10.2. Implementations (`impl`)

Implementations provide the concrete function bodies for an interface for a specific type or type constructor.

#### 10.2.1. Implementation Declaration

Provides bodies for interface methods. Can use `fn #method` shorthand if desired.

**Syntax:**
```able
[ImplName =] impl [<ImplGenericParams>] InterfaceName [InterfaceArgs] for Target [where <ConstraintList>] {
  [ConcreteFunctionDefinitions]
}
```

-   **`[ImplName =]`**: Optional name for the implementation, used for disambiguation. Followed by `=`.
-   **`impl`**: Keyword.
-   **`[<ImplGenericParams>]`**: Optional comma-separated generics for the implementation itself (e.g., `<T: Numeric>`). Use `<>` delimiters. Constraints can be specified inline here or in the `where` clause.
-   **`InterfaceName`**: The name of the interface being implemented.
-   **`[InterfaceArgs]`**: Space-delimited type arguments for the interface's generic parameters (if any).
-   **`for`**: Keyword (mandatory).
-   **`Target`**: The specific type or type constructor implementing the interface.
    *   If the interface was defined using `interface ... for SelfTypePattern ...`, the `Target` must structurally match the `SelfTypePattern`.
        *   If interface is `for Point`, `Target` must be `Point`.
        *   If interface is `for T`, `Target` can be any type `SomeType` (or a type variable `T` in generic impls).
        *   If interface is `for M _`, `Target` must be `TypeConstructor` (e.g., `Array`). See HKT syntax below.
        *   If interface is `for Array T`, `Target` must be `Array U` (where `U` might be constrained).
    *   If the interface was defined using `interface ... { ... }` (without `for`), the `Target` **must not** be a bare type constructor (like `Array`). It must be a specific type (e.g., `i32`, `Point`, `Array i32`, or a type variable `T`).
-   **`[where <ConstraintList>]`**: Optional clause for specifying constraints on `ImplGenericParams`.
-   **`{ [ConcreteFunctionDefinitions] }`**: Block containing the full definitions (`fn name(...) { body }`) for all functions required by the interface (unless they have defaults). Signatures must match. May override defaults.

-   **Distinction from `methods`:** `methods Type { ... }` defines *inherent* methods (part of the type's own API). `impl Interface for Type { ... }` fulfills an *external contract* defined by the interface. An inherent method defined in `methods Type` may be used to explicitly satisfy an interface requirement in an `impl` block, but the `impl Interface for Type` block is still needed to declare the conformance.

#### 10.2.2. HKT Implementation Syntax (Refined)

To implement an interface defined `for M _` (like `Mappable`) for a concrete constructor like `Array`:

```able
[ImplName =] impl [<ImplGenerics>] InterfaceName TypeParamName for TypeConstructor [where ...] {
  ## 'TypeParamName' (e.g., A) is bound here and usable below.
  fn method<B>(self: TypeConstructor TypeParamName, ...) -> TypeConstructor B { ... }
}
## Example:
impl Mappable A for Array {
  fn map<B>(self: Array A, f: (A -> B)) -> Array B { ... }
}

## Example for Option (assuming 'union Option T = T | nil')
union Option T = T | nil
impl Mappable A for Option {
  fn map<B>(self: Self A, f: A -> B) -> Self B { self match { case a: A => f(a), case nil => nil } }
}
```
*(Note: This syntax is only applicable when the interface was defined using the `for M _` pattern.)*

#### 10.2.3. Examples of Implementations

```able
## Implementing Display (defined 'for T') for Point
impl Display for Point {
  fn to_string(self: Self) -> String { `Point({self.x}, {self.y})` } ## Self is Point
}

## Implementing Hashable (defined without 'for') for i32
impl Hashable for i32 {
  fn hash(self: Self) -> u64 { compute_i32_hash(self) } ## Self is i32
}

## Implementing Hashable for Point
impl Hashable for Point {
    fn hash(self: Self) -> u64 { compute_point_hash(self.x, self.y) } ## Self is Point
}

## ERROR: Cannot implement Hashable (no 'for') for a type constructor
# impl Hashable A for Array { ... }

## Implementing Zeroable (defined 'for T', static method) for i32
impl Zeroable for i32 {
  fn zero() -> Self { 0 } ## Self is i32 here
}

## Implementing Zeroable for Array T (Requires generic impl)
impl<T> Zeroable for Array T {
  fn zero() -> Self { [] } ## Self is Array T
}

## Named Monoid Implementations for i32 (Assuming 'interface Monoid for T')
Sum = impl Monoid for i32 {
  fn id() -> Self { 0 }
  fn op(self: Self, other: Self) -> Self { self + other }
}
Product = impl Monoid for i32 {
  fn id() -> Self { 1 }
  fn op(self: Self, other: Self) -> Self { self * other }
}

## Overriding a default method
struct MyGreeter;
impl Greeter for MyGreeter {
    fn greet(self: Self) -> string { "Hi from MyGreeter!" } ## Overrides default
}
```

#### 10.2.4. Overlapping Implementations and Specificity

When multiple `impl` blocks could apply to a given type and interface, Able uses specificity rules to choose the *most specific* implementation. If no single implementation is more specific, it results in a compile-time ambiguity error. Rules (derived from Rust RFC 1210, simplified):

1.  **Concrete vs. Generic:** Implementations for concrete types (`impl ... for i32`) are more specific than implementations for type variables (`impl ... for T`). (`Array i32` is more specific than `Array T`).
2.  **Concrete vs. Interface Bound:** Implementations for concrete types (`impl ... for Array T`) are more specific than implementations for types bound by an interface (`impl ... for T: Iterable`).
3.  **Interface Bound vs. Unconstrained:** Implementations for constrained type variables (`impl ... for T: Iterable`) are more specific than for unconstrained variables (`impl ... for T`).
4.  **Subset Unions:** Implementations for union types that are proper subsets are more specific (`impl ... for i32 | f32` is more specific than `impl ... for i32 | f32 | f64`).
5.  **Constraint Set Specificity:** An `impl` whose type parameters have a constraint set that is a proper superset of another `impl`'s constraints is more specific (`impl<T: A + B> ...` is more specific than `impl<T: A> ...`).

Ambiguities must be resolved manually, typically by qualifying the method call (see Section [10.3.3](#1033-disambiguation-named-impls)).

### 10.3. Usage

#### 10.3.1. Instance Method Calls

Use dot notation on a value whose type implements the interface. Resolution follows rules in Section [9.3](#93-method-call-syntax-resolution-initial-rules), considering specificity.
```able
p = Point { x: 1, y: 2 }
s = p.to_string() ## Calls Point's impl of Display.to_string

arr = [1, 2, 3]
arr_mapped = arr.map({ x => x * 2 }) ## Calls Array's impl of Mappable.map
```

#### 10.3.2. Static Method Calls

Use `TypeName.static_method(...)` notation. The `TypeName` must have an `impl` for the interface containing the static method.
```able
zero_int = i32.zero()          ## Calls i32's impl of Zeroable.zero
empty_f64_array = (Array f64).zero() ## Calls Array T's impl of Zeroable.zero
## TBD: Syntax for calling static methods on generic types like Array needs confirmation.
## Maybe Array.zero<f64>() or (Array f64).zero() ? Let's assume the latter for now.
```

#### 10.3.3. Disambiguation (Named Impls)

If multiple implementations exist (e.g., `Sum` and `Product` for `Monoid for i32`), qualify the call with the implementation name:
```able
sum_id = Sum.id()             ## 0
prod_id = Product.id()         ## 1
res = Sum.op(5, 6)          ## 11 (Calls Sum's op)
res2 = Product.op(5, 6)       ## 30

## TBD: Interaction between named impls and instance method call syntax
## (`value.ImplName.method(...)`) needs confirmation/specification.
## For now, assume named impls primarily disambiguate static calls or free functions.
```

#### 10.3.4. Interface Types (Dynamic Dispatch)

Using an interface as a type allows storing heterogeneous values implementing the same interface. Method calls typically use dynamic dispatch. The exact syntax is TBD (`dyn Iface` or `(Iface)`).

```able
## Assuming syntax 'dyn Iface'
displayables: Array (dyn Display) = [1, "hello", Point{...}]
for item in displayables {
  print(item.to_string()) ## Dynamic dispatch selects correct to_string impl
}
```

## 11. Error Handling

Able provides multiple mechanisms for handling errors and exceptional situations:

1.  **Explicit `return`:** Allows early exit from functions.
2.  **`Option T` (`?Type`) and `Result T` (`!Type`) types:** Used with V-lang style propagation (`!`) and handling (`else {}`) for expected errors or absence.
3.  **Exceptions:** For exceptional conditions, using `raise` and `rescue`. Panics are implemented via exceptions.

### 11.1. Explicit `return` Statement

Functions can return a value before reaching the end of their body using the `return` keyword. (See also Section [7.3](#73-explicit-return-statement)).

#### Syntax
```able
return Expression
return // Equivalent to 'return void' if function returns void
```

-   **`return`**: Keyword initiating an early return.
-   **`Expression`**: Optional expression whose value is returned from the function. Its type must match the function's declared or inferred return type.
-   If `Expression` is omitted, the function must have a `void` return type, and `void` is implicitly returned.

#### Semantics
-   Immediately terminates the execution of the current function.
-   The value of `Expression` (or `void`) is returned to the caller.
-   If used within nested blocks (like loops or `do` blocks) inside a function, it still returns from the *function*, not just the inner block.

#### Example
```able
fn find_first_negative(items: Array i32) -> ?i32 {
  for item in items {
    if item < 0 {
      ## Assuming Option T = T | nil and Some is struct { value: T }
      return Some { value: item } ## Early return with value
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

### 11.2. V-Lang Style Error Handling (`Option`/`Result`, `!`, `else`)

This mechanism is preferred for handling *expected* errors or optional values gracefully without exceptions.

#### 11.2.1. Core Types (`?Type`, `!Type`)

-   **`Option T` (`?Type`)**: Represents optional values. Defined implicitly as the union `nil | T`. Used when a value might be absent. (See Section [4.6.2](#462-nullable-type-shorthand-)).
    ```able
    user: ?User = find_user(id) ## find_user returns nil or User
    ```
-   **`Result T` (`!Type`)**: Represents the result of an operation that can succeed with a value of type `T` or fail with an error. Defined implicitly as the union `T | Error`.
    ```able
    ## The 'Error' interface (built-in or standard library, TBD)
    ## Conceptually:
    interface Error for T {
        fn message(self: Self) -> string;
        ## Potentially other methods like cause(), stacktrace()
    }

    ## Result T is implicitly: union Result T = T | Error
    ## !Type is syntactic sugar for Result T

    ## Example function signature
    fn read_file(path: string) -> !string { ... } ## Returns string or Error
    ```

#### 11.2.2. Error/Option Propagation (`!`)

The postfix `!` operator simplifies propagating `nil` from `Option` types or `Error` from `Result` types up the call stack.

##### Syntax
```able
ExpressionReturningOptionOrResult!
```

##### Semantics
-   Applies to an expression whose type is `?T` (`nil | T`) or `!T` (`T | Error`).
-   If the expression evaluates to the "successful" variant (`T`), the `!` operator unwraps it, and the overall expression evaluates to the unwrapped value (of type `T`).
-   If the expression evaluates to the "failure" variant (`nil` or an `Error`), the `!` operator causes the **current function** to immediately **`return`** that `nil` or `Error` value.
-   **Requirement:** The function containing the `!` operator must itself return a compatible `Option` or `Result` type (or a supertype union) that can accommodate the propagated `nil` or `Error`.

##### Example
```able
## Assuming read_file returns !string (string | Error)
## Assuming parse_data returns !Data (Data | Error)
fn load_and_parse(path: string) -> !Data {
    content = read_file(path)! ## If read_file returns Err, load_and_parse returns it.
                               ## Otherwise, content is string.
    data = parse_data(content)! ## If parse_data returns Err, load_and_parse returns it.
                                ## Otherwise, data is Data.
    return data ## Return the successful Data value (implicitly wrapped in Result)
}

## Option example
fn get_nested_value(data: ?Container) -> ?Value {
    container = data! ## If data is nil, return nil from get_nested_value
    inner = container.get_inner()! ## Assuming get_inner returns ?Inner, propagate nil
    value = inner.get_value()! ## Assuming get_value returns ?Value, propagate nil
    return value
}
```

#### 11.2.3. Error/Option Handling (`else {}`)

Provides a way to handle the `nil` or `Error` case of an `Option` or `Result` immediately, typically providing a default value or executing alternative logic.

##### Syntax
```able
ExpressionReturningOptionOrResult else { BlockExpression }
ExpressionReturningOptionOrResult else { |err| BlockExpression } // Capture error
```

##### Semantics
-   Applies to an expression whose type is `?T` (`nil | T`) or `!T` (`T | Error`).
-   If the expression evaluates to the "successful" variant (`T`), the overall expression evaluates to that unwrapped value (`T`). The `else` block is *not* executed.
-   If the expression evaluates to the "failure" variant (`nil` or an `Error`):
    *   The `BlockExpression` inside the `else { ... }` is executed.
    *   If the form `else { |err| ... }` is used *and* the failure value was an `Error`, the error value is bound to the identifier `err` (or chosen name) within the scope of the `BlockExpression`. If the failure value was `nil`, `err` is not bound or has a `nil`-like value (TBD - let's assume it's only bound for `Error`).
    *   The entire `Expression else { ... }` expression evaluates to the result of the `BlockExpression`.
-   **Type Compatibility:** The type of the "successful" variant (`T`) and the type returned by the `BlockExpression` must be compatible. The overall expression has this common compatible type.

##### Example
```able
## Option handling
config_port: ?i32 = read_port_config()
port = config_port else { 8080 } ## Provide default value if config_port is nil
## port is i32

## Result Handling with error capture
user: ?User = find_user(id) else { |err| ## Assuming find_user returns !User
    log(`Failed to find user: ${err.message()}`)
    nil ## Return nil if lookup failed
}

## Result Handling without capturing error detail
data = load_data() else { ## Assuming load_data returns !Array T
    log("Loading failed, using empty.")
    [] ## Return empty array
}
```

### 11.3. Exceptions (`raise` / `rescue`)

For handling truly *exceptional* situations that disrupt normal control flow, often originating from deeper library levels or representing programming errors discovered at runtime. Panics are implemented as a specific kind of exception. Division/Modulo by zero raises an exception.

#### 11.3.1. Raising Exceptions (`raise`)

The `raise` keyword throws an exception value, immediately interrupting normal execution and searching up the call stack for a matching `rescue` block.

##### Syntax
```able
raise ExceptionValue
```
-   **`raise`**: Keyword initiating the exception throw.
-   **`ExceptionValue`**: An expression evaluating to the value to be raised. This value should typically implement the standard `Error` interface, but technically any value *could* be raised (TBD - restricting to `Error` types is safer).

##### Example
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

#### 11.3.2. Rescuing Exceptions (`rescue`)

The `rescue` keyword provides a mechanism to catch exceptions raised during the evaluation of an expression. It functions similarly to `match` but operates on exceptions caught during the primary expression's execution. `rescue` forms an expression.

##### Syntax
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

##### Example
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

#### 11.3.3. Panics

Panics are implemented as a specific, severe type of exception (e.g., `PanicError` implementing `Error`). Calling the built-in `panic(message: string)` function is equivalent to `raise PanicError { message: message }`. Typically, `rescue` blocks should *avoid* catching `PanicError` unless performing essential cleanup before potentially re-raising or terminating, as panics signal unrecoverable states or bugs. The default top-level program handler usually catches unhandled panics, prints details, and terminates.

## 12. Concurrency

Able provides lightweight concurrency primitives inspired by Go, allowing asynchronous execution of functions and blocks using the `proc` and `spawn` keywords.

### 12.1. Concurrency Model Overview

-   Able supports concurrent execution, allowing multiple tasks to run seemingly in parallel.
-   The underlying mechanism (e.g., OS threads, green threads, thread pool, event loop) is implementation-defined but guarantees the potential for concurrent progress.
-   Communication and synchronization between concurrent tasks (e.g., channels, mutexes) are not defined in this section but would typically be provided by the standard library.

### 12.2. Asynchronous Execution (`proc`)

The `proc` keyword initiates asynchronous execution of a function call or a block, returning a handle (`Proc T`) to manage the asynchronous process.

#### 12.2.1. Syntax

```able
proc FunctionCall
proc BlockExpression
```

-   **`proc`**: Keyword initiating asynchronous execution.
-   **`FunctionCall`**: A standard function or method call expression (e.g., `my_function(arg1)`, `instance.method()`).
-   **`BlockExpression`**: A `do { ... }` block expression.

#### 12.2.2. Semantics

1.  **Asynchronous Start**: The target `FunctionCall` or `BlockExpression` begins execution asynchronously, potentially on a different thread or logical task. The current thread does *not* block.
2.  **Return Value**: The `proc` expression immediately returns a value whose type implements the `Proc T` interface.
    -   `T` is the return type of the `FunctionCall` or the type of the value the `BlockExpression` evaluates to.
    -   If the function/block returns `void`, the return type is `Proc void`.
3.  **Independent Execution**: The asynchronous task runs independently until it completes, fails, or is cancelled.

#### 12.2.3. Example

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

#### 12.2.4. Process Handle (`Proc T` Interface)

The `Proc T` interface provides methods to interact with an ongoing asynchronous process started by `proc`. (See Section [14](#14-standard-library-interfaces-conceptual--tbd) for conceptual definition).

##### Definition (Conceptual)

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
  fn get_value(self: Self) -> Result T ProcError; ## Note: Result T is T | Error

  ## Request cancellation of the asynchronous process.
  ## This is a non-blocking request. Cancellation is cooperative; the async task
  ## must potentially check for cancellation requests to terminate early.
  ## Guarantees no specific timing or success, only signals intent.
  fn cancel(self: Self) -> void;
}
```

##### Semantics of Methods

-   **`status()`**: Returns the current state (`Pending`, `Resolved`, `Cancelled`, `Failed`) without blocking.
-   **`get_value()`**: Blocks the caller until the process finishes (resolves, fails, or is definitively cancelled).
    -   If `Resolved`, returns `Ok(value)` where `value` has type `T`. For `Proc void`, returns `Ok(void)` conceptually (or `Ok(nil)` if `void` cannot be wrapped? TBD). Let's assume `Ok(value)` works even for `T=void`.
    -   If `Failed`, returns `Err(ProcError)` containing error details.
    -   If `Cancelled`, returns `Err(ProcError)` indicating cancellation.
-   **`cancel()`**: Sends a cancellation signal to the asynchronous task. The task is not guaranteed to stop immediately or at all unless designed to check for cancellation signals.

##### Example Usage

```able
data_proc: Proc string = proc fetch_data("http://example.com")

## Check status without blocking
current_status = data_proc.status()
if match current_status { case Pending => true } { print("Still working...") }

## Block until done and get result (handle potential errors)
result = data_proc.get_value()
final_data = result match {
  case Ok { value: d } => `Success: ${d}`,
  case Err { error: e } => `Failed: ${e.details}` ## Assuming Result T = T | Error
}
print(final_data)

## Request cancellation (fire and forget)
data_proc.cancel()
```

### 12.3. Thunk-Based Asynchronous Execution (`spawn`)

The `spawn` keyword also initiates asynchronous execution but returns a `Thunk T` value, which implicitly blocks and yields the result when evaluated.

#### 12.3.1. Syntax

```able
spawn FunctionCall
spawn BlockExpression
```

-   **`spawn`**: Keyword initiating thunk-based asynchronous execution.
-   **`FunctionCall` / `BlockExpression`**: Same as for `proc`.

#### 12.3.2. Semantics

1.  **Asynchronous Start**: Starts the function or block execution asynchronously, similar to `proc`. The current thread does not block.
2.  **Return Value**: Immediately returns a value of the special built-in type `Thunk T`.
    -   `T` is the return type of the function or the evaluation type of the block.
    *   If the function/block returns `void`, the return type is `Thunk void`.
3.  **Implicit Blocking Evaluation**: The core feature of `Thunk T` is its evaluation behavior. When a value of type `Thunk T` is used in a context requiring a value of type `T` (e.g., assignment, passing as argument, part of an expression), the current thread **blocks** until the associated asynchronous computation completes.
    *   If the computation completes successfully with value `v` (type `T`), the evaluation of the `Thunk T` yields `v`.
    *   If the computation fails (e.g., panics), that panic is **propagated** to the thread evaluating the `Thunk T`.
    *   Evaluating a `Thunk void` blocks until completion and then yields `void` (effectively synchronizing).

#### 12.3.3. Example

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

### 12.4. Key Differences (`proc` vs `spawn`)

-   **Return Type:** `proc` returns `Proc T` (an interface handle); `spawn` returns `Thunk T` (a special type).
-   **Control:** `Proc T` offers explicit control (check status, attempt cancellation, get result via method call potentially handling errors).
-   **Result Access:** `Thunk T` provides implicit result access; evaluating the thunk blocks and returns the value directly (or propagates panic). It lacks fine-grained status checks or cancellation via the handle itself.
-   **Use Cases:**
    *   `proc` is suitable when you need to manage the lifecycle of the async task, check its progress, handle failures explicitly, or potentially cancel it.
    *   `spawn` is simpler for "fire and forget" tasks where you only need the final result eventually and are okay with blocking for it implicitly (or propagating panics).

## 13. Packages and Modules

Packages form a tree of namespaces rooted at the name of the library, and the hierarchy follows the directory structure of the source files within the library.

### 13.1. Package Naming and Structure

*   **Root Package Name**: Every package has a root package name, defined by the `name` field in the `package.yml` file located in the package's root directory.
*   **Unqualified Names**: All individual package name segments (directory names, names declared with `package`) must be valid identifiers.
*   **Qualified Names**: Package paths are composed of segments delimited by periods (`.`), e.g., `root_package.sub_dir.module_name`.
*   **Directory Mapping**: Directory names containing source files are part of the package path. Hyphens (`-`) in directory names are treated as underscores (`_`) when used in the package path. Example: A directory `my-utils` becomes `my_utils` in the package path.

### 13.2. Package Declaration in Source Files

*   **Optional Declaration**: A source file can optionally declare which sub-package its contents belong to using `package <unqualified-name>;`.
*   **Implicit Package**:
    *   If a file `src/foo/bar.able` contains `package my_bar;`, and the root package name is `my_pkg`, its fully qualified package is `my_pkg.foo.my_bar`.
    *   If a file `src/foo/baz.able` has *no* `package` declaration, its fully qualified package is determined by its directory path relative to the root: `my_pkg.foo`.
*   **Multiple Files**: Multiple files can contribute to the same fully qualified package name, either by residing in the same directory (without `package` declarations) or by declaring the same package name within different directories.

#### Example

Assume a package root `/home/david/projects/hello-world` with `package.yml` specifying `name: hello_world`.

```
/home/david/projects/hello-world/
├── package.yml         (name: hello_world)
├── foo.able            (no package declaration)
├── bar.able            (contains: package bar;)
├── utils/
│   ├── helpers.able    (no package declaration)
│   └── formatters.able (contains: package fmt;)
└── my-data/
    └── models.able     (contains: package data_models;)

```

This structure defines the following packages:

*   `hello_world`: Contains definitions from `foo.able`.
*   `hello_world.bar`: Contains definitions from `bar.able`.
*   `hello_world.utils`: Contains definitions from `utils/helpers.able`.
*   `hello_world.utils.fmt`: Contains definitions from `utils/formatters.able`.
*   `hello_world.my_data.data_models`: Contains definitions from `my-data/models.able`. (Note `my-data` -> `my_data`).

### 13.3. Package Configuration (`package.yml`)

A single file, `package.yml`, defines a project's build configuration, dependencies, package metadata, etc. The directory containing `package.yml` is the root of the library.

```yaml
name: hello_world
version: 1.0.0
license: MIT

authors:
- David <david@conquerthelawn.com>

dependencies:
  collections:
    github: davidkellis/able-collections
    version: ~>0.16.0 ## Example version constraint
```

### 13.4. Importing Packages (`import`)

The `import` statement makes identifiers from other packages available in the current scope.

*   **Syntax Forms**:
    *   Package import: `import io;` (makes `io.puts` etc. available)
    *   Wildcard import: `import io.*;` (brings all public identifiers from `io` into scope - use with caution)
    *   Selective import: `import io.{puts, gets, SomeType};` (brings specific identifiers into scope)
    *   Aliased import: `import internationalization as i18n;` (imports package under alias)
    *   Aliased selective import: `import io.{puts as print_line, gets};` (imports specific items with aliases)
*   **Scope**: Imports can occur at the top level of a file (package scope) or within any local scope (e.g., inside a function).
*   **Binding Semantics**: Importing an identifier creates a new binding in the current scope. This binding refers to the same underlying definition (function, type, etc.) as the original identifier in the imported package.

### 13.5. Visibility and Exports (`private`)

*   **Default**: All identifiers defined at the top level (package scope) are public and exported by default.
*   **`private` Keyword**: Prefixing a top-level definition with `private` restricts its visibility to the current package only. Private identifiers cannot be imported or accessed directly from other packages.

```able
## In package 'my_pkg'

foo = "bar" ## Public

private baz = "qux" ## Private to my_pkg

fn public_func() { ... }

private fn helper() { ... } ## Private to my_pkg
```

## 14. Standard Library Interfaces (Conceptual / TBD)

Many language features rely on interfaces expected to be in the standard library. These require full definition.

*   **Iteration:**
    *   `struct IteratorEnd;` (Singleton type signalling end of iteration).
    *   `interface Iterator T for SelfType { fn next(self: Self) -> T | IteratorEnd; }`
    *   `interface Iterable T for SelfType { fn iterator(self: Self) -> (Iterator T); }`
*   **Operators:** `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg`, `Not` (Bitwise `~`), `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`.
*   **Comparison:** `PartialEq`, `Eq`, `PartialOrd`, `Ord`.
*   **Functions:** `Apply` (for callable values `value(args)`).
*   **Collections Indexing:** `Index`, `IndexMut`.
*   **Display:** `Display` (for `to_string`, used by interpolation).
*   **Error Handling:** `Error` (base interface for errors).
*   **Concurrency:** `Proc`, `ProcError` (details of handle and error).
*   **Cloning:** `Clone`.
*   **Hashing:** `Hash` (for map keys etc.).
*   **Default Values:** `Default`.
*   **Ranges:** `Range` (type returned by `..`/`...`, implements `Iterable`).

### Core Iteration Protocol

Iteration in Able is based on the `Iterable` and `Iterator` interfaces, along with a special singleton type `IteratorEnd` to signal completion.

```able
## Singleton struct used to signal the end of iteration.
struct IteratorEnd;

## Interface for stateful iterators producing values of type T.
## SelfType represents the concrete iterator type holding the iteration state.
interface Iterator T for SelfType {
  ## Retrieves the next element from the iterator.
  ## Returns the element (of type T) or IteratorEnd if iteration is complete.
  ## This method typically mutates the iterator's internal state.
  fn next(self: Self) -> T | IteratorEnd;
}

## Interface for types that can produce an iterator over elements of type T.
## SelfType represents the collection type (e.g., Array T, Range).
interface Iterable T for SelfType {
  ## Creates and returns a new iterator positioned at the start of the sequence.
  ## The return type '(Iterator T)' represents *any* type that implements Iterator T
  ## (an existential type / interface type).
  fn iterator(self: Self) -> (Iterator T);
}
```

## 15. To Be Defined / Refined

*   **Standard Library Implementation:** Core types (`Array`, `Map`?, `Set`?, `Range`, `Option`/`Result` details, `Proc`, `Thunk`), IO, String methods, Math, `Iterable`/`Iterator` protocol, Operator interfaces. Definition of standard `Error` interface.
*   **Operator Details:** Compound assignment (`+=`) semantics. Integer overflow behavior (wrapping? panic?). Shift semantics.
*   **Type System Details:** Full inference rules, Variance, Coercion (if any), "Interface type" syntax (`dyn Display`?) and mechanism details, HKT limitations/capabilities.
*   **Error Handling:** `ProcError`/`PanicError` structure. Panic cleanup (`defer`?). Raising non-`Error` values? `|err|` binding for `nil`. `DivisionByZeroError` details.
*   **Concurrency:** Synchronization primitives (channels, mutexes?), Cancellation details. Scheduler guarantees.
*   **Mutability:** Final decision on "mutable by default" vs `let`/`var`. Need for immutable data structures.
*   **Arrays:** Formal definition (`Array T`), including mutability of elements via `IndexMut`.
*   **FFI:** Mechanism for calling external code.
*   **Metaprogramming:** Macros?
*   **Entry Point:** The program entry point is a function named `main` with the signature `fn main() -> void` (or inferred `void`) defined in the root package (TBC precise package requirement).
    ```able
    ## Example Entry Point
    fn main() {
        print("Hello, Able!")
    }
    ```
*   **Tooling:** Compiler, Package manager commands, Testing framework.
*   **Lexical Details:** Block comment syntax? Full keyword list confirmation.
*   **Named Impl Invocation:** Syntax for instance method calls via named impls (`value.ImplName.method(...)`).
*   **Garbage Collection:** Specifics, controls?
*   **Field/Method Name Collisions:** Precedence rules.
