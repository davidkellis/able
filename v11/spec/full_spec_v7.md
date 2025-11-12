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
    *   [9.3. Example: `methods` block for `Address`](#93-example-methods-block-for-address)
    *   [9.4. Method Call Syntax Resolution](#94-method-call-syntax-resolution)
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
*   **Identifiers:** Start with a letter (`a-z`, `A-Z`) or underscore (`_`), followed by letters, digits (`0-9`), or underscores. Typically `[a-zA-Z_][a-zA-Z0-9_]*`. Identifiers are case-sensitive. Package/directory names mapping to identifiers treat hyphens (`-`) as underscores. The identifier `_` is reserved as the wildcard pattern (see Section [5.2.2](#522-wildcard-pattern-_)) and for unbound type parameters (see Section [4.4](#44-reserved-identifier-_-in-types)). The tokens `@` and `@n` (e.g., `@1`, `@2`, ...) are reserved for expression placeholders and cannot be used as identifiers.
*   **Keywords:** Reserved words that cannot be used as identifiers: `fn`, `struct`, `union`, `interface`, `impl`, `methods`, `type`, `package`, `import`, `dynimport`, `extern`, `prelude`, `private`, `Self`, `do`, `return`, `if`, `or`, `else`, `while`, `for`, `in`, `match`, `case`, `breakpoint`, `break`, `raise`, `rescue`, `ensure`, `rethrow`, `proc`, `spawn`, `as`, `nil`, `true`, `false`, `void`.
*   **Reserved Tokens (non-identifiers):** `@` and numbered placeholders `@n` (e.g., `@1`, `@2`, ...), used for expression placeholder lambdas.
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
| `i32`    | 32-bit signed integer (-2³¹ to 2³¹-1)           | `-2_147_483_648`, `0`, `42_i32`      | Default type for integer literals.              |
| `i64`    | 64-bit signed integer (-2⁶³ to 2⁶³-1)           | `-9_223_..._i64`, `1_000_000_000_i64`| Underscores `_` allowed for readability.        |
| `i128`   | 128-bit signed integer (-2¹²⁷ to 2¹²⁷-1)      | `-170_..._i128`, `0_i128`, `170_..._i128`|                                                 |
| `u8`     | 8-bit unsigned integer (0 to 255)             | `0`, `10_u8`, `255_u8`              |                                                 |
| `u16`    | 16-bit unsigned integer (0 to 65,535)           | `0_u16`, `1000`, `65535_u16`        |                                                 |
| `u32`    | 32-bit unsigned integer (0 to 2³²-1)            | `0`, `4_294_967_295_u32`            | Underscores `_` allowed for readability.        |
| `u64`    | 64-bit unsigned integer (0 to 2⁶⁴-1)            | `0_u64`, `18_446_..._u64`           |                                                 |
| `u128`   | 128-bit unsigned integer (0 to 2¹²⁸-1)          | `0`, `340_..._u128`                 |                                                 |
| `f32`    | 32-bit float (IEEE 754 single-precision)      | `3.14_f32`, `-0.5_f32`, `1e-10_f32`, `2.0_f32` | Suffix `_f32`.                                  |
| `f64`    | 64-bit float (IEEE 754 double-precision)      | `3.14159`, `-0.001`, `1e-10`, `2.0`  | Default type for float literals. Suffix `_f64` optional if default. |
| `string` | Immutable sequence of Unicode chars (UTF-8) | `"hello"`, `""`, `` `interp ${val}` `` | Double quotes or backticks (interpolation).      |
| `bool`   | Boolean logical values                        | `true`, `false`                     |                                                 |
| `char`   | Single Unicode scalar value (UTF-32)        | `'a'`, `'π'`, `'💡'`, `'\n'`, `'\u{1F604}'` | Single quotes. Supports escape sequences.       |
| `nil`    | Singleton type representing **absence of data**. | `nil`                               | **Type and value are both `nil` (lowercase)**. Often used with `?Type`. |
| `void`   | Type with **no values** (empty set).          | *(No literal value)*                | Represents computations completing without data. In specific contexts like `Proc void` or `Thunk void`, it acts as a signal of successful completion. |

*(See Section [6.1](#61-literals) for detailed literal syntax.)*

### 4.3. Type Expression Syntax Details

*   **Simple:** `i32`, `string`, `MyStruct`
*   **Generic Application:** `Array i32`, `Map string User` (space-delimited arguments)
*   **Grouping:** `Map string (Array i32)` (parentheses control application order)
*   **Function:** `(i32, string) -> bool`
*   **Nullable Shorthand:** `?string` (Syntactic sugar for `nil | string`, see Section [4.6.2](#462-nullable-type-shorthand-))
*   **Result Shorthand:** `!string` (Syntactic sugar for `Error | string`, see Section [11.2.1](#1121-core-types-type-type))
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

Define a new type as a composition of existing variant types using `|`. The order of variants in the definition (`A | B` vs `B | A`) is generally not significant for type checking but might influence runtime representation or default pattern matching order (TBD). For consistency, this specification prefers the order `FailureVariant | SuccessVariant` where applicable (e.g., `nil | T`, `Error | T`).

##### Syntax
```able
union UnionTypeName [GenericParamList] = VariantType1 | VariantType2 | ... | VariantTypeN
```
-   **`union`**: Keyword.
-   **`UnionTypeName`**: The name of the new union type being defined.
-   **`[GenericParamList]`**: Optional space-delimited generic parameters applicable to the union itself.
-   **`=`**: Separator.
-   **`VariantType1 | VariantType2 | ...`**: List of one or more variant types separated by `|`.
    -   Each `VariantType` must be a pre-defined, valid type name (e.g., primitive, struct, another union, generic type application, or interface type).

##### Examples
```able
## Simple enumeration using singleton structs
struct Red; struct Green; struct Blue;
union Color = Red | Green | Blue

## Option type (direct)
union Option T = nil | T

## Result type (direct; uses interface type 'Error')
union Result T = Error | T

## Mixing types
union IntOrString = i32 | string

## Payload-bearing variants (named fields)
struct Circle { radius: f64 }
struct Rectangle { width: f64, height: f64 }
struct Triangle { a: f64, b: f64, c: f64 }
union Shape = Circle | Rectangle | Triangle
```

Note: Each `VariantType` in a union may be a concrete type (e.g., `i32`, `Point`), another union, a generic application (e.g., `Array i32`), or an interface type (e.g., `Error`, `Display`). Using an interface name as a variant denotes an existential value implementing that interface (dynamic dispatch). For example, `Result T = Error | T` is valid and corresponds to values that are either a concrete error type implementing `Error` or a success value of type `T`.

#### 4.6.2. Nullable Type Shorthand (`?`)

Provides concise syntax for types that can be either a specific type or `nil`.

##### Syntax
```able
?Type
```
-   **`?`**: Prefix operator indicating nullability.
-   **`Type`**: Any valid type expression.

##### Equivalence
`?Type` is syntactic sugar for the union `nil | Type`. This follows the `FailureVariant | SuccessVariant` convention.

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

## Option (direct)
opt_val: Option i32 = 42
opt_nothing: Option i32 = nil

## Result values are directly either Error or T
# res_ok: Result string = "Data loaded"
# res_err: Result string = SomeConcreteError { details: "File not found" }

val: IntOrString = 100
val2: IntOrString = "hello"

## Constructing payload-bearing variants
shape1: Shape = Circle { radius: 2.0 }
shape2: Shape = Rectangle { width: 3.0, height: 4.0 }
shape3: Shape = Triangle { a: 3.0, b: 4.0, c: 5.0 }
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

## Pattern matching on payload-bearing variants
shape_area = shape1 match {
  case Circle { radius } => 3.141592653589793 * radius * radius,
  case Rectangle { width, height } => width * height,
  case Triangle { a, b, c } => {
    s = (a + b + c) / 2.0
    (s * (s - a) * (s - b) * (s - c)).sqrt() ## Assume sqrt available
  }
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
    *   Handles both reassignment of existing bindings and initial binding of new ones, depending on the `LHS` and lexical context.
    *   **If `LHS` is an Identifier:**
        *   If `Identifier` exists as an accessible, mutable binding (found via lexical scoping, checking current scope first), it **reassigns** that binding.
        *   If `Identifier` does not exist lexically, it declares a **new** mutable binding in the **current** scope (initial binding).
    *   **If `LHS` is a Destructuring Pattern (e.g., `{x, y}`, `[a, b]`):**
        *   For each identifier within the `Pattern`:
            *   If the identifier matches an existing, accessible, mutable binding (found via lexical scoping), that existing binding is **reassigned**.
            *   If the identifier does *not* match any existing accessible binding, a **new** mutable binding is created in the **current** scope (initial binding).
    *   **If `LHS` is a Field/Index Access (`instance.field`, `array[index]`):**
        *   Performs **mutation** on the specified field or element, provided it's accessible and mutable.
    *   **Precedence:** Reassignment of existing bindings takes precedence over creating new ones if an identifier matches. To guarantee a new binding that shadows an outer one, use `:=`.
    *   It is a compile-time error if `LHS` attempts to reassign bindings/locations that are not accessible or not mutable, or access fields/indices that do not exist.
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

*   **Syntax**: `[StructTypeName] { Field1: PatternA, Field2 as BindingB: PatternC, ShorthandField, ... }`
    *   `StructTypeName`: Optional, the name of the struct type. If omitted, the type is inferred or checked against the expected type.
    *   `Field: Pattern`: Matches the value of `Field` against the nested `PatternA`.
    *   `Field as Binding: Pattern`: Matches the value of `Field` against nested `PatternC` and binds the original field value to the new identifier `BindingB` (only valid with `:=`).
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
    { id, name as user_name, address: addr } := u
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
#### 5.2.7. Typed Patterns (Type Guards)

Typed patterns refine a match by requiring the value to conform to a given type.

*   **Syntax**: `Identifier: Type` or `_: Type`
    *   `Identifier: Type`: Matches if the subject is of `Type`, and binds it to `Identifier`.
    *   `_: Type`: Matches if the subject is of `Type`, discards the bound value.
*   **Examples**:
    ```able
    value match {
      case s: string => print(s),
      case n: i32 => print(n + 1),
      case _: Error => log("got error"),
      case _ => print("unknown")
    }
    ```
*   **Semantics**: Acts as a runtime type guard within `match`/`rescue`. This does not introduce new static subtyping; it narrows within the matched branch only.

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
    *   **`:=`**: Resulting value matched against `Pattern` (LHS). **Always** creates new mutable bindings in the current scope for identifiers introduced in the pattern, potentially shadowing outer bindings. It is a compile-time error if any introduced identifier already exists as a binding *in the current scope*.
    *   **`=`**: Resulting value matched against `LHS` pattern/location specifier.
        *   If `LHS` is a destructuring pattern, identifiers within it are either **reassigned** (if they match existing accessible bindings) or cause **new initial bindings** to be created in the current scope (if they don't match existing bindings).
        *   If `LHS` is an identifier or field/index access, it performs **reassignment** or **mutation** on the existing target.
        *   It is a compile-time error if any target location for reassignment/mutation does not exist, is not accessible, or is not mutable.
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
| 13         | `^`                   | Exponentiation                          | Right-to-left | Binds tighter than unary `-`. `-x^2` == `-(x^2)`          |
| 12         | `-` (unary)           | Arithmetic Negation                     | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 12         | `!` (unary)           | **Logical NOT**                         | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 12         | `~` (unary)           | **Bitwise NOT (Complement)**            | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 11         | `*`, `/`, `%`         | Multiplication, Division, Remainder     | Left-to-right |                                                           |
| 10         | `+`, `-` (binary)     | Addition, Subtraction                   | Left-to-right |                                                           |
| 9          | `<<`, `>>`            | Left Shift, Right Shift                 | Left-to-right |                                                           |
| 8          | `&` (binary)          | Bitwise AND                             | Left-to-right |                                                           |
| 7          | `\xor`               | Bitwise XOR                             | Left-to-right |                                                           |
| 6          | `|` (binary)          | Bitwise OR                              | Left-to-right |                                                           |
| 6          | `>`, `<`, `>=`, `<=`    | Ordering Comparisons                    | Non-assoc     | Chaining requires explicit grouping (`(a<b) && (b<c)`) |
| 5          | `==`, `!=`            | Equality, Inequality Comparisons        | Non-assoc     | Chaining requires explicit grouping                     |
| 4          | `&&`                  | Logical AND (short-circuiting)          | Left-to-right |                                                           |
| 3          | `||`                  | Logical OR (short-circuiting)           | Left-to-right |                                                           |
| 2          | `..`, `...`           | Range Creation (inclusive, exclusive)   | Non-assoc     |                                                           |
| 1          | `:=`                  | **Declaration and Initialization**      | Right-to-left | See Section [5.1](#51-operators---)                     |
| 1          | `=`                   | **Reassignment / Mutation**             | Right-to-left | See Section [5.1](#51-operators---)                     |
| 1          | `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `\xor=`, `<<=`, `>>=` | Compound Assignment                      | Right-to-left | Desugars to `a = a OP b` (no `^=`).                       |
| 0          | `\|>`                 | Pipe Forward                            | Left-to-right | (Lowest precedence)                                       |

*(Note: Precedence levels are relative; specific numerical values may vary but the order shown is based on Rust.)*

#### 6.3.2. Operator Semantics

*   **Arithmetic (`+`, `-`, `*`, `/`, `%`):** Standard math operations on numeric types. Division (`/`) or remainder (`%`) by zero **raises a runtime exception** (e.g., `DivisionByZeroError`). See Section [11.3](#113-exceptions-raise--rescue).
    *   Numeric promotion (P1):
        -   Integer with integer widens to the larger integer kind (e.g., `i32` with `i64` → `i64`).
        -   Integer with float promotes to the float kind (e.g., `i32` with `f64` → `f64`).
        -   No implicit narrowing (including signed/unsigned). Use explicit casts for non-widening conversions.
    *   Integer overflow (O1):
        -   Checked by default. On overflow in `+`, `-`, `*`, raises a runtime exception `PanicError { message: "integer overflow" }`.
        -   Division and remainder by zero already raise `DivisionByZeroError`.
        -   Library provides explicit alternatives (names TBD): `wrapping_add/sub/mul`, `saturating_add/sub/mul`, `checked_add/sub/mul -> ?T` for performance-critical or specific semantics.
*   **Comparison (`>`, `<`, `>=`, `<=`, `==`, `!=`):** Compare values, result `bool`. Equality/ordering behavior relies on standard library interfaces (`PartialEq`, `Eq`, `PartialOrd`, `Ord`). See Section [14](#14-standard-library-interfaces-conceptual--tbd).
*   **Logical (`&&`, `||`, `!`):**
    *   `&&` (Logical AND): Short-circuiting on `bool` values.
    *   `||` (Logical OR): Short-circuiting on `bool` values.
    *   `!` (Logical NOT): Unary operator, negates a `bool` value.
*   **Bitwise (`&`, `|`, `\xor`, `<<`, `>>`, `~`):**
    *   `&`, `|`, `\xor`: Standard bitwise AND, OR, XOR on integer types (`i*`, `u*`).
    *   `<<`, `>>` (Shift semantics, S1):
        -   Shift count must be in range `0..bits` for the left operand's type. Out-of-range shift counts raise `PanicError { message: "shift out of range" }`.
        -   Right shift of signed integers is arithmetic (sign-extending), matching Go semantics; right shift of unsigned integers is logical (zero-filling).
    *   `~` (Bitwise NOT): Unary operator, performs bitwise complement on integer types.
*   **Unary (`-`):** Arithmetic negation for numeric types.
*   **Member Access (`.`):** Access fields/methods, UFCS, static methods. See Section [9.4](#94-method-call-syntax-resolution).
*   **Function Call (`()`):** Invokes functions/methods. See Section [7.4](#74-function-invocation).
*   **Indexing (`[]`):** Access elements within indexable collections (e.g., `Array`). Relies on standard library interfaces (`Index`, `IndexMut`). See Section [14](#14-standard-library-interfaces-conceptual--tbd).
*   **Range (`..`, `...`):** Create `Range` objects (inclusive `..`, exclusive `...`). See Section [8.2.3](#823-range-expressions).
*   **Declaration (`:=`):** Declares/initializes new variables. Evaluates to RHS. See Section [5.1](#51-operators---).
*   **Assignment (`=`):** Reassigns existing variables or mutates locations. Evaluates to RHS. See Section [5.1](#51-operators---).
*   **Compound Assignment (`+=`, etc.):** Shorthand (e.g., `a += b` is like `a = a + b`). Acts like `=`.
*   **Pipe Forward (`|>`):** `x |> f` evaluates to `f(x)`.

Compound assignment semantics:
*   Supported forms: `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `\xor=`, `<<=`, `>>=`. The exponent form `^=` is not supported.
*   Desugaring: `a OP= b` is defined as `a = a OP b`, where `OP` is the corresponding binary operator.
*   Single evaluation (C1): `a` is evaluated exactly once. The compiler lowers to an assignment to the same storage location without re-evaluating addressable subexpressions on the LHS.
    -   Example: `arr[i()] += f()` evaluates `i()` once, not twice.
*   Types: Assignment follows the same rules as plain `=` of the desugared form; the target of `a` must be assignable from the result type of `a OP b`.

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
*   Multiline strings are supported in both forms:
    -   Backticks: newlines are literal.
    -   Double quotes: newlines are literal characters within the string; standard escapes apply.

### 6.7. Generator Literal (`Iterator { ... }`)

Creates a value implementing `(Iterator T)` by writing imperative code that `yield`s elements on demand. Generators produce values lazily when the consumer calls `next()` on the iterator.

#### Syntax

```able
Iterator { gen => ExpressionList }
Iterator T { gen => ExpressionList }        ## Optional element type annotation
```

-   **`Iterator`**: Introduces a generator literal producing an iterator.
-   **`[T]`**: Optional element type annotation. If provided, all `yield`s must be compatible with `T`. If omitted, `T` is inferred from context or from the union of yielded value types.
-   **`gen`**: Identifier bound to the generator driver within `ExpressionList`.
-   **`ExpressionList`**: Imperative code that may call methods on `gen` (notably `gen.yield(value)` and `gen.stop()`). Last expression's value is ignored.

#### Generator Driver API (within the body)

```able
gen.yield(value: T) -> void   ## Yield a value and suspend until next() is called again
gen.stop() -> void            ## Terminate the generator early (subsequent next() => IteratorEnd)
```

#### Typing

-   The generator literal has type `(Iterator T)`.
-   If the annotation `Iterator T { ... }` is present, each `gen.yield(expr)` must type-check as `expr: T'` where `T'` is compatible with `T`.
-   Without an annotation, `T` is:
    -   the context-required element type if present (e.g., a function expects `(Iterator U)`), otherwise
    -   the least upper bound / union of all yielded expression types.
-   A generator that yields no values has element type `T` from context if available; otherwise it is a type error unless an explicit `T` annotation is provided.

#### Semantics

1.  Creation returns an object implementing `Iterator T` with internal suspended state and captured lexical environment.
2.  On `next()`:
    -   Resume execution of `ExpressionList` from the last suspension point until one of:
        -   `gen.yield(v)` is executed: suspend and return `v`.
        -   `gen.stop()` is executed or the body finishes: mark as complete; return `IteratorEnd` and subsequently always `IteratorEnd`.
        -   An exception is raised: propagate the exception to the caller of `next()` (rescuable via `rescue`).
3.  Re-entrancy: It is an error to call `next()` again while the generator is suspended at a `gen.yield` for the same iterator value (single-consumer, sequential usage assumed).
4.  `return` inside the generator body immediately terminates the generator (equivalent to `gen.stop()`).
5.  Local `ensure` blocks run when the generator terminates (normal or via `stop()`/exception).

#### Examples

Simple range generator (inclusive):

```able
fn range_inclusive(start: i32, end: i32) -> (Iterator i32) {
  Iterator i32 { gen =>
    i = start
    while i <= end {
      gen.yield(i)
      i = i + 1
    }
  }
}

for n in range_inclusive(3, 5) { print(n) } ## 3 4 5
```

Ad-hoc generator (filter-map):

```able
fn evens_to_strings(xs: Iterable i32) -> (Iterator string) {
  Iterator string { gen =>
    xs.each(fn (x: i32) {
      if x % 2 == 0 { gen.yield(`even:${x}`) }
    })
  }
}
```

Default `iterator` via generator (ties to Section 14 Core Iteration Protocol):

```able
## In the default of Iterable T.iterator(self)
fn iterator(self: Self) -> (Iterator T) {
  Iterator { gen => self.each(gen.yield) }
}
```

### 6.8. Arrays (`Array T`)

Resizable contiguous sequence of elements of type `T`.

#### Construction and Literals

```able
[]                    ## Array never-typed literal; requires context or annotation
[1, 2, 3]             ## Array i32
[1, 2.0]              ## Array (i32 | f64) unless context fixes to float
arr: Array i32 = []   ## Explicit type annotation
```

#### Typing and Mutability

-   The type `Array T` holds elements of type `T`.
-   Arrays are mutable containers; element mutation uses indexing and requires `IndexMut` in scope.
-   Assignment copies/moves the array binding; elements are moved unless `Clone` is used.

#### Core Operations

Indexing (panics on out-of-bounds):
```able
x = arr[i]                 ## Read element (may raise IndexError if i out of bounds)
arr[i] = v                 ## Write element (may raise IndexError)
```

Safe access:
```able
arr.get(i)    -> ?T        ## nil if out of bounds
arr.set(i, v) -> !void     ## Error if out of bounds
```

Length/capacity:
```able
arr.size()     -> u64
```

Push/pop:
```able
arr.push(v)   -> void
arr.pop()     -> ?T        ## nil if empty
```

Iteration:
```able
impl Iterable T for Array T { ... }   ## Provided by stdlib
for x in arr { ... }
```

Slicing (views; TBD if borrowed slices are first-class):
```able
arr.slice(start: usize, end: usize) -> Array T   ## TBD copy vs view semantics
```

#### Semantics

-   Bounds: `arr[i]` and `arr[i] = v` raise `IndexError` on out-of-bounds. `get`/`pop` return `?T`.
-   Growth: Amortized doubling
-   Equality/Hashing: Derive from elementwise comparisons if `T: Eq/Hash` (via interfaces in Section 14).

#### Examples

```able
arr: Array i32 = []
arr.push(1)
arr.push(2)
arr[0] = 3
first = arr.get(0) else { 0 }
sum = 0
for x in arr { sum = sum + x }
```

### 6.9. Lexical Details (Comments, Identifiers, Literals)

This section defines source comments and identifier rules.

#### Line Comments

-   Start with `##` and continue to end-of-line.
-   Inside string/interpolated literals, `##` is treated as ordinary text.
-   Block comments are not supported.
-   Doc comments are not supported; use ordinary line comments.

#### Identifiers

-   Source text is UTF-8 (Go-style). Byte Order Mark (BOM) is not permitted. End-of-line may be LF or CRLF.
-   Identifiers use Unicode rules like Go:
    -   identifier = letter { letter | digit }
    -   letter = Unicode letters (categories Lu, Ll, Lt, Lm, Lo, Nl) or `_`
    -   digit  = Unicode decimal digits (category Nd)
-   No normalization is applied; different Unicode sequences are distinct identifiers.
-   Keywords are reserved and cannot be used as identifiers.

#### Statement Termination and Line Joining

-   Newlines separate expressions/statements. Semicolons are optional and generally not used.
-   Implicit line joining applies within open delimiters `(` `[` `{` until their closing counterpart.

#### Trailing Commas

-   Trailing commas are permitted in array literals, struct literals, and import lists.
-   Trailing commas are disallowed in argument lists.

#### Numeric Literals

-   Integer bases: decimal, binary `0b...`, octal `0o...`, hexadecimal `0x...`.
-   Leading zero in a decimal literal is acceptable and does not imply octal.
-   Underscores `_` are permitted as digit separators between digits; not allowed at the start or end of the literal, nor adjacent to base prefixes or type suffixes.
-   Type suffixes are allowed and use the form `_<kind>` (e.g., `_i64`, `_u16`, `_i128`, `_u128`). The suffix determines the literal's type; otherwise, type is from context.
-   Floating-point literals support decimal forms (with optional exponent). Hexadecimal float forms are not recognized. No special `NaN` or `Inf` tokens.

### 6.10. Dynamic Metaprogramming (Interpreted)

Able supports a dynamic, interpreted metaprogramming mode that lets you define packages and package-level items (interfaces, impls, functions, structs, unions, methods) at runtime. Dynamic items are:

-   executed by the interpreter (not compiled ahead-of-time),
-   dynamically typed (no static typechecking on newly created items), and
-   usable from compiled code via explicit dynamic bridges and late-bound imports, keeping a clean separation from statically compiled code.

#### Static vs Dynamic Realms

-   Static realm: compile-time symbol table (types/functions/packages) resolved by `import`; used for static typing and dispatch; immutable at runtime.
-   Dynamic realm: runtime symbol table (packages and items created/extended at runtime); resolved by `dynimport`; does not affect static typing or static dispatch.
-   Cross-realm interactions are explicit via `dynimport`, `dyn.package/def_package`, and interface adapters like `dyn.as_interface`.

#### Package Objects and Definition

Dynamic code is organized into package objects (dynamic namespaces):

```able
## Get an existing package (dynamic or compiled-backed) as a dynamic package object
dyn.package(fully_qualified_name: string) -> !dyn.Package

## Define a new dynamic package; returns its package object
dyn.def_package(fully_qualified_name: string) -> !dyn.Package

methods dyn.Package {
  ## Define declarations inside this package's namespace using Able source text (interpreted).
  ## Valid constructs: interfaces, impls, package-level functions, structs, unions, methods.
  fn def(self: Self, code: string) -> !void
}
```

Relative vs absolute package paths inside `def`:
-   If the code inside `def` begins with `package Foo;` and `p` is bound to package `Qux`, the definitions land under `Qux.Foo` (relative subpackage).
-   Fully qualified paths (e.g., `package root.util.foo;` if your naming scheme supports it) are treated as absolute.

Special-form block sugar:
-   `p.def { ... }` is sugar for passing the enclosed Able source to `p.def("...")`. The block uses normal Able grammar (braces; no special `end`).

#### Dynamic Imports and Late Binding

-   Use `dynimport` (not `import`) to bind names from dynamic packages:
    ```able
    dynimport foo.{bar}
    ```
-   Late binding: names introduced by `dynimport` are resolved at use time. If `foo.bar` is redefined later via `p.def`, subsequent calls to `bar()` see the new definition.
-   Failure modes: If a `dynimport`ed name does not exist or is not callable at the time of use, invoking it raises an `Error`.

#### Interop Surfaces

-   Compiled → dynamic:
    -   Call dynamic functions/values via late-bound `dynimport` names.
    -   Or use reflective helpers (e.g., `dyn.call`, `dyn.construct`) if provided by the standard library.
-   Dynamic → compiled:
    -   Dynamic code can call exported compiled functions/types via a host binding table provided to the interpreter.
-   Interface adapters:
    -   To use dynamic objects with static code through interface lenses, adapt them explicitly: `dyn.as_interface(value, Display) -> ?Display`. Only object-safe methods are callable via interface-typed values.

#### Concurrency and Safety

-   Thread-safety: `dyn.package`, `dyn.def_package`, `dynimport`, and `p.def` are thread-safe. Redefinitions are atomic at symbol granularity.
-   Races: Late-bound lookups observe either the old or new definition; behavior is well-defined at the granularity of whole-symbol replacement.
-   Errors: Parse errors, missing names, arity/shape mismatches return `Error` from dynamic APIs or raise during invocation.

#### Performance and Deployment

-   Optional interpreter: Projects may exclude the interpreter in AOT builds. Dynamic facilities are only available when the interpreter is present.
-   Implementations may JIT or bytecode-compile dynamic functions internally; this is not visible at the language level.

#### Examples

Package-oriented usage with relative nesting:
```able
## In dynamic/interpreted execution
p = dyn.package("Qux")!
p.def {
  package Foo;
  import io.{puts}
  fn bar() { puts("hi") }
}

## From compiled code, bind a late-resolving reference to bar
package Qux.Client
dynimport Qux.Foo.{bar}
fn baz() {
  bar()  ## resolves at call time to current Qux.Foo.bar
}
```

Defining a new dynamic package and using a dynamic type:
```able
p2 = dyn.def_package("foo")!
p2.def(`struct Point { name: string, x: i32, y: i32 }`)!
dynimport foo.{Point}
pt := Point { name: "blah", x: 12, y: 82 }
```

This facility deliberately avoids static type guarantees; it offers a clear, explicit bridge between compiled and interpreted worlds via `dynimport`, dynamic package objects, and interface adapters.

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

Allows calling functions (both inherent/interface methods and qualifying free functions) using dot notation on the first argument. (See Section [9.4](#94-method-call-syntax-resolution) for resolution details).

##### Syntax
```able
ReceiverExpression . FunctionOrMethodName ( RemainingArgumentList )
```

##### Semantics (Simplified - see Section 9.4 for full rules)
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

Create a new function by providing some arguments and using `@` as a placeholder for others. Numbered placeholders `@1`, `@2`, ... refer to specific parameter positions. Mixing is allowed: unnumbered `@` fill remaining positions left-to-right; the arity is the maximum of the highest numbered placeholder and the count of unnumbered placeholders.

#### 7.5.1. Syntax
Use `@` (and `@n`) in place of arguments in a function or method call expression.
```able
function_name(Arg1, @, Arg3, ...)
instance.method_name(@, Arg2, ...)
```

#### 7.5.2. Syntax & Semantics
-   `function_name(Arg1, @, ...)` creates a closure.
-   `receiver.method_name(@, Arg2, ...)` creates a closure capturing `receiver`.
    *   To partially apply a static method, use `TypeExpr.static_method(@, ...)`. For a closure expecting the receiver as the first argument, use `InterfaceName.method(@, ...)` in function position and pass the receiver explicitly when calling.
-   `receiver.free_function_name` (using Method Call Syntax access without `()`) creates a closure equivalent to `free_function_name(receiver, @, ...)`.

#### 7.5.3. Examples
```able
add_10 = add(@, 10)      ## Function expects one arg: add(arg, 10)
result = add_10(5)       ## result is 15

## Assuming prepend exists: fn prepend(prefix: string, body: string) -> string
# prefix_hello = prepend("Hello, ", @) ## Function expects one arg
# msg = prefix_hello("World")          ## msg is "Hello, World"

## method call syntax access creates partially applied function
add_five = 5.add ## Creates function add(5, @) via Method Call Syntax access
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
#### 7.6.3. Placeholder Lambdas (`@`, `@n`)

Placeholders in expression positions create anonymous functions.

*   `@` without a number introduces fresh parameters left-to-right within the expression:
    *   `@ + 1` → `{ x => x + 1 }`
    *   `@ * @` → `{ x, y => x * y }`
*   Numbered placeholders refer to specific parameter positions; repeats reuse the same parameter:
    *   `@1 + @1` → `{ x => x + x }`
*   Mixing is allowed; arity is the maximum of the highest numbered placeholder and the count of unnumbered placeholders:
    *   `f(@1, @, @3)` → `{ x, y, z => f(x, y, z) }`
*   Scope: The smallest enclosing expression that expects a function determines the lambda boundary. If a placeholder spans a whole block, the entire block becomes the lambda body. Parentheses may be used for clarity without changing scope.
*   Errors:
    *   Using `@`/`@n` where a named identifier is required (outside expression placeholders) is a compile-time error.
    *   Arity mismatches between inferred placeholder lambdas and the expected function type at the call site are compile-time errors.


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
    *   Unification rules:
        -   Common supertype via union (C1). If branches yield unrelated types `A` and `B`, the chain type is `A | B`.
        -   Nil special-case (N1). If one branch is `nil` and the other is `T`, the chain type is `?T`.
        -   Result special-cases (R1/R2). If branches are `T` and `!U`, upcast to `!T`/`!U` and unify (overall `!(T | U)` if needed). If branches are `!A` and `!B`, result is `!(A | B)`.
        -   Numeric exactness (E1). No implicit numeric coercions here beyond global operator rules; if numeric types differ and are not otherwise coerced, the unified type is a union (e.g., `i32 | f64`).

    Examples:
    ```able
    if c1 { 1 } or { "x" }                ## i32 | string
    if c1 { nil } or { v: T }             ## ?T
    if c1 { v: T } or { w: U }            ## T | U
    if c1 { ok: T } or { read() }         ## if read() -> !U, then !(T | U)
    if c1 { 1 } or { 2.0 }                ## i32 | f64 (E1)
    ```

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
    *   **`PatternX`**: Pattern to match (Literal, Identifier, `_`, Typed Identifier `name: Type`, Typed Wildcard `_: Type`, Type/Variant, Struct `{}`, Array `[]`). Bound variables are local to this clause. See Section [5.2](#52-patterns).
    *   **`[if GuardX]`**: Optional `bool` guard expression using pattern variables.
    *   **`=>`**: Separator.
    *   **`ResultExpressionListX`**: Expressions executed if clause chosen; last expression's value is the result.

##### Semantics

1.  **Sequential Evaluation**: `SubjectExpression` evaluated once. `case` clauses checked top-to-bottom.
2.  **First Match Wins**: The first `PatternX` that matches *and* whose `GuardX` (if present) is true selects the clause.
3.  **Execution & Result**: The chosen `ResultExpressionListX` is executed. The `match` expression evaluates to the value of the last expression in that list.
4.  **Exhaustiveness**: Compiler SHOULD check for exhaustiveness (especially for unions). Non-exhaustive matches MAY warn/error at compile time and SHOULD panic at runtime. A `case _ => ...` usually ensures exhaustiveness.
5.  **Type Compatibility**: All `ResultExpressionListX` must yield compatible types. The `match` expression's type is this common type.
    *   Unification rules as for `if/or`:
        -   Union common supertype (C1); `nil` with `T` yields `?T` (N1).
        -   Result cases: `!A` with `!B` → `!(A | B)`; `T` with `!U` → `!(T | U)` (R1/R2).
        -   Numeric exactness (E1) — otherwise, union numeric types.

    Examples:
    ```able
    x match { case 1 => 1, case _ => "one" }     ## i32 | string
    x match { case v: T => v, case nil => nil } ## ?T
    r match { case a: A => a, case e: Error => default() } ## A | ReturnType(default)
    r match { case a: A => a, case e: Error => recover(e) } ## if recover: Error->A, overall A; else A | ReturnType(recover)
    ```

##### Example

```able
## Assuming Option T = nil | T
description = maybe_num match {
  case x: i32 if x > 0 => `Positive: ${x}`,
  case 0 => "Zero",
  case x: i32 => `Negative: ${x}`,
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
    *   Unification follows the same rules as `if/or` (C1, N1, R1/R2, E1). All normal block exits and `break` payloads are unified to produce the `breakpoint` expression's type (B1).

    Example:
    ```able
    result = breakpoint 'find {
      for x in xs { if p(x) { break 'find x } }
      nil
    }                    ## If xs: Array T, result: ?T (N1/B1)

    result2 = breakpoint 'mix {
      if c { break 'mix 1 } else { break 'mix "a" }
    }                    ## i32 | string (C1/B1)
    ```

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

### 9.4 Method Call Syntax Resolution

This section details the step-by-step process the Able compiler uses to determine *which* function or method to call when encountering the method call syntax: `ReceiverExpression.Identifier(ArgumentList)`.

**Resolution Steps:**

Let `ReceiverType` be the static type of the `ReceiverExpression`. The compiler attempts to resolve `Identifier` in the following order:

1.  **Check for Field Access:**
    *   Determine if `ReceiverType` has a field named `Identifier`.
    *   If a field exists:
        *   If the field's type implements the `Apply` interface (making it callable), the call resolves to invoking the `apply` method on the field's value (`ReceiverExpression.Identifier.apply(ArgumentList)`).
        *   If the field's type does *not* implement `Apply`, and parentheses `()` with arguments are present, this is a **compile-time error** (cannot call a non-callable field).
        *   If parentheses are absent (`ReceiverExpression.Identifier`), it resolves to accessing the field's value.

2.  **Check for Inherent Methods:**
    *   Determine if an inherent method named `Identifier` is defined for `ReceiverType` within a `methods ReceiverType { ... }` block.
    *   If found, the call resolves to this inherent method. `ReceiverExpression` is passed as the `self` argument (or equivalent first argument).

3.  **Check for Interface Methods (Trait Methods):**
    *   Identify *all* interfaces `I` that `ReceiverType` is known to implement (either directly via `impl I for ReceiverType` or through generic bounds like `T: I`).
    *   Filter this set to interfaces `I` that define a method named `Identifier`.
    *   **If exactly one such interface `I` is found:** The call resolves to the implementation of `Identifier` provided by the `impl I for ReceiverType` block.
    *   **If multiple such interfaces are found:** Apply the **Specificity Rules** (See Section [10.2.4](#1024-overlapping-implementations-and-specificity)) to find the *single most specific* implementation among the candidates.
        *   If a single most specific implementation exists, the call resolves to that implementation.
        *   If no single implementation is more specific than all others (ambiguity), this step fails, and resolution continues (or results in an error if no further steps match). **Note:** Explicit disambiguation might be required (See Section [10.3.3](#1033-disambiguation-named-impls)).
    *   **If no such interfaces are found:** This step fails.

4.  **Check for Universal Function Call Syntax (UFCS):**
    *   Search the current scope for a free (non-method) function named `Identifier` whose *first parameter* type is compatible with `ReceiverType`.
    *   If exactly one such function is found, the call resolves to this function, passing `ReceiverExpression` as the first argument (`Identifier(ReceiverExpression, ArgumentList)`).
    *   If multiple such functions exist (e.g., due to overloading based on later arguments, if supported) or none exist, this step fails.

**Precedence and Error Handling:**

*   The resolution stops at the **first step** that successfully finds a match.
*   **Precedence Order:** Field Access (Callable) > Inherent Method > Interface Method (after specificity) > UFCS.
*   If **ambiguity** arises within Step 3 (multiple equally specific interface implementations) and is not resolved by later steps (which is unlikely given the precedence), a **compile-time error** occurs, requiring explicit disambiguation.
*   If **no match** is found after all steps, a **compile-time error** occurs ("method not found").

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
    *(Note: Using an interface name in a type position denotes a dynamic/existential interface type; using it in a constraint position denotes a static constraint. See Section [10.3.4](#1034-interface-types-dynamic-dispatch).)*

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

Implementations provide the concrete function bodies for an interface for a specific type or type constructor. They establish the link showing that a type *satisfies* an interface contract.

#### 10.2.1. Implementation Declaration

Provides bodies for interface methods. Can use `fn #method` shorthand if desired. `impl` may be declared `private` to restrict visibility to the defining package.

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
    *   If the interface was defined using `interface ... { ... }` (without `for`), the `Target` **must not** be a bare type constructor (like `Array` or `Map K _`). It must be a concrete type (e.g., `i32`, `Point`), a fully bound generic type (e.g., `Array i32`), or a type variable `T` constrained elsewhere. Implementing such an interface for a type constructor requires the interface to use the `for M _` pattern. **Compile-time error:** Attempting to `impl InterfaceWithoutFor for TypeConstructor` is an error.
-   **`[where <ConstraintList>]`**: Optional clause for specifying constraints on `ImplGenericParams`.
-   **`{ [ConcreteFunctionDefinitions] }`**: Block containing the full definitions (`fn name(...) { body }`) for all functions required by the interface (unless they have defaults). Signatures must match. May override defaults.

-   **Method Availability:** An `impl` block associates the defined methods with the `Target` type for the specific `InterfaceName`. These methods become available through method call syntax (`value.method()`) on values of the `Target` type, subject to the resolution rules (Section [9.4](#94-method-call-syntax-resolution)). They are not directly imported into the scope like names via `import`; instead, the compiler finds them during method lookup based on the receiver's type and the implemented interfaces.

#### 10.2.2. Semantic Effect of `impl`

*   **Association, Not Scoping:** An `impl InterfaceName for TargetType { ... }` block does **not** directly add the defined methods (`ConcreteFunctionDefinitions`) into the namespace or scope of `TargetType` itself. Inherent methods defined in a `methods TargetType { ... }` block are directly associated with the type's scope.
*   **Contract Fulfillment:** The primary role of an `impl` block is to **register** with the compiler that `TargetType` conforms to `InterfaceName`. It provides the necessary evidence (the concrete method bodies) for this conformance.
*   **Method Lookup Table:** Conceptually, the compiler maintains a mapping: for each pair `(InterfaceName, TargetType)`, it knows the location of the function implementations provided by the corresponding `impl` block. This mapping is crucial for method resolution (Section [9.4](#94-method-call-syntax-resolution-revised-and-expanded)).
*   **Visibility:** The `impl` block itself follows the visibility rules of where it's defined. However, the *association* it creates between the type and interface becomes known wherever both the `TargetType` and `InterfaceName` are visible. Public types implementing public interfaces can be used polymorphically across package boundaries. Private types or implementations for private interfaces are restricted.
*   **Dispatch:** When an interface method is called (e.g., `receiver.interface_method()`), the compiler uses the type of `receiver` and the method name to look up the correct implementation via the `(InterfaceName, TargetType)` mapping established by the `impl` block (details in Section [9.4](#94-method-call-syntax-resolution-revised-and-expanded) and Section [10.3](#103-usage-revised)).

#### 10.2.3. HKT Implementation Syntax (Refined)

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

## Example for Option (direct union 'nil | T')
union Option T = nil | T
impl Mappable A for Option {
  fn map<B>(self: Self A, f: A -> B) -> Self B {
    self match { case a: A => f(a), case nil => nil }
  }
}
```
*(Note: This syntax is only applicable when the interface was defined using the `for M _` pattern.)*

#### 10.2.4. Examples of Implementations

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

#### 10.2.5. Overlapping Implementations and Specificity

When multiple `impl` blocks could apply to a given type and interface, Able uses specificity rules to choose the *most specific* implementation. If no single implementation is more specific, it results in a compile-time ambiguity error. Rules (derived from Rust RFC 1210, simplified):

1.  **Concrete vs. Generic:** Implementations for concrete types (`impl ... for i32`) are more specific than implementations for type variables (`impl ... for T`). (`Array i32` is more specific than `Array T`).
2.  **Concrete vs. Interface Bound:** Implementations for concrete types (`impl ... for Array T`) are more specific than implementations for types bound by an interface (`impl ... for T: Iterable`).
3.  **Interface Bound vs. Unconstrained:** Implementations for constrained type variables (`impl ... for T: Iterable`) are more specific than for unconstrained variables (`impl ... for T`).
4.  **Subset Unions:** Implementations for union types that are proper subsets are more specific (`impl ... for i32 | f32` is more specific than `impl ... for i32 | f32 | f64`).
5.  **Constraint Set Specificity:** An `impl` whose type parameters have a constraint set that is a proper superset of another `impl`'s constraints is more specific (`impl<T: A + B> ...` is more specific than `impl<T: A> ...`).

Ambiguities must be resolved manually, typically by qualifying the method call (see Section [10.3.3](#1033-disambiguation-named-impls)).

### 10.3. Usage

This section explains how interface methods are invoked and how different dispatch mechanisms work.

#### 10.3.1. Instance Method Calls and Dispatch

When calling a method using dot notation `receiver.method_name(args...)`:

*   **Resolution:** The compiler follows the steps outlined in Section [9.4](#94-method-call-syntax-resolution-revised-and-expanded). If resolution points to an interface method implementation (Step 3), the specific `impl` block is identified.
*   **Static Dispatch (Monomorphization):**
    *   If the `receiver` has a **concrete type** (`Point`, `Array i32`, etc.) at the call site, the compiler knows the exact `impl` block to use. It typically generates code that directly calls the specific function defined within that `impl` block.
    *   This also applies when the `receiver` has a **generic type `T` constrained by the interface** (e.g., `fn process<T: Display>(item: T) { item.to_string() }`). During compilation (often through monomorphization), the compiler generates specialized versions of `process` for each concrete type `T` used, and the `item.to_string()` call within each specialized version directly calls the correct implementation for that concrete type.
    *   This is the most common form of dispatch – efficient and resolved at compile time.
*   **Dynamic Dispatch:**
    *   This occurs when the `receiver` has an **interface type** (e.g., `Display` used in a type position — see Section [10.3.4](#1034-interface-types-dynamic-dispatch-revised)). The actual concrete type of the value is not known at compile time.
    *   The call `receiver.method_name(args...)` is dispatched at *runtime*. The `receiver` value (often represented as a "fat pointer" containing a pointer to the data and a pointer to a **vtable**) uses the vtable to find the address of the correct `method_name` implementation for the underlying concrete type and then calls it.
    *   This enables runtime polymorphism but incurs a small runtime overhead compared to static dispatch.

```able
p = Point { x: 1, y: 2 } ## Concrete type Point
s = p.to_string()       ## Static dispatch to 'impl Display for Point'

fn print_any<T: Display>(item: T) {
  print(item.to_string()) ## Static dispatch within monomorphized versions of print_any
}
print_any(p)      ## Instantiates print_any<Point>, calls Point's to_string
print_any("hi")   ## Instantiates print_any<string>, calls string's to_string

## Dynamic dispatch example (see 10.3.4)
displayables: Array Display = [p, "hi"]
for item in displayables {
  print(item.to_string()) ## Dynamic dispatch via vtable based on item's concrete type
}
```

#### 10.3.2. Static Method Calls

Static methods defined within an interface (those not taking `self`) are called on a fully bound type expression.

*   **Rule:** Always call statics using `TypeExpr.static_method(args...)`, where `TypeExpr` is a fully bound type (e.g., `i32`, `(Array f64)`, `Point`, or a type variable `T` in a context where `T: Interface`).
    *Example:*
    ```able
    zero_int = i32.zero()                 ## Calls impl Zeroable for i32
    empty_f64_arr = (Array f64).zero()    ## Calls impl Zeroable for Array f64

    fn make_zero<T: Zeroable>() -> T { T.zero() }  ## Generic static call via bound type variable
    ```
*   **Disambiguation using Named Impls:** If multiple `impl` blocks could provide the static method (e.g., `Sum` and `Product` implementing `Monoid`), use the implementation name:
    ```able
    sum_id = Sum.id()       ## Calls static 'id' from 'Sum = impl Monoid for i32'
    prod_id = Product.id()  ## Calls static 'id' from 'Product = impl Monoid for i32'
    ```

#### 10.3.3. Disambiguation (Named Impls and Explicit Paths)

When method call resolution (Section [9.4](#94-method-call-syntax-resolution-revised-and-expanded)) results in ambiguity (multiple equally specific interface methods) or when you need to explicitly choose a non-default implementation, use more qualified syntax:

1.  **Named Implementation Calls:** If an implementation was named (`ImplName = impl ...`), you can select its methods explicitly, but not via instance method syntax.
    *   **Static Methods:** `ImplName.static_method(args...)` (as shown above).
    *   **Instance Methods — disallowed as instance syntax:** You may not write `receiver.(ImplName.method)(...)` or otherwise select a named impl via `receiver.method(...)` syntax. Instead, call the method in function position by qualifying with the implementation name and passing the receiver explicitly, or use the pipeline operator:
        ```able
        ## Assuming Sum/Product impl Monoid for i32
        res_sum = Sum.op(5, 6)         ## 11
        res_prod = Product.op(5, 6)    ## 30

        ## Pipeline-friendly form
        res_sum2 = 5 |> Sum.op(6)
        ```
        This style avoids ambiguity while keeping instance method syntax unextended.

2.  **Fully Qualified Interface Calls:** To resolve ambiguity between interfaces even without named implementations, you can specify the interface explicitly using function style and passing the receiver as the first argument:
    ```able
    ## Assume Type implements InterfaceA and InterfaceB, both having 'conflicting_method'.
    val: Type = ...
    result_a = InterfaceA.conflicting_method(val, args...)
    result_b = InterfaceB.conflicting_method(val, args...)
    ```
    *   **Note:** This relies on the compiler being able to determine the correct `impl` for `InterfaceA for Type` and `InterfaceB for Type` respectively. If ambiguity still exists (e.g., multiple `impl InterfaceA for Type`), named implementations are likely necessary.

#### 10.3.4. Interface Types (Dynamic Dispatch)

Using an interface name as a type denotes a dynamic/existential interface value: a value of some concrete type that implements the interface, with method calls dispatched at runtime.

*   **Syntax:** Use the interface name directly in type positions (variable declarations, parameter/return types, composite types like arrays/maps/unions).
    ```able
    struct Circle { radius: f64 }
    struct Square { side: f64 }

    impl Display for Circle { fn to_string(self: Self) -> string { $"Circle({self.radius})" } }
    impl Display for Square { fn to_string(self: Self) -> string { $"Square({self.side})" } }

    ## Create an array holding values viewed through the 'Display' interface lens
    shapes: Array Display = [Circle { radius: 1.0 }, Square { side: 2.0 }]
    for s in shapes { print(s.to_string()) }
    ```
*   **Method Calls:** When a method is called on an interface-typed value (`item.to_string()` where `item` has type `Display`), the implementation corresponding to the underlying concrete type (captured at upcast time) is invoked at runtime.
*   **Static vs Dynamic Use:**
    -   In constraint positions (`T: Display` or `where T: Display`), `Display` is a static bound; calls are resolved at compile time.
    -   In type positions (`x: Display`, `Array Display`, `Error | T`), `Display` is a dynamic/existential; calls are resolved at runtime.
*   **Object Safety:** Only methods that are object-safe (no unconstrained generic method parameters, no returning `Self` except where boxed/erased is defined) are callable through interface-typed values. Object-safety rules are TBD and will be specified; non-object-safe methods are unavailable via interface values but usable via static bounds.
*   **Import-Scoped Model:** The concrete implementation used for a dynamic/interface-typed value is fixed at the upcast site (where a concrete value is converted to an interface type) based on impls in scope there. Consumers do not need that impl in scope to call methods on the received interface value.


## 11. Error Handling

Able provides multiple mechanisms for handling errors and exceptional situations:

1.  **Explicit `return`:** Allows early exit from functions.
2.  **`Option T` (`?Type`) and `Result T` (`!Type`) types:** Used with V-lang style propagation (`!`) and handling (`else {}`) for expected errors or absence.
3.  **Exceptions:** For exceptional conditions, using `raise` and `rescue`.

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
    if item < 0 { return item }
  }
  return nil
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
-   **`Result T` (`!Type`)**: Represents the result of an operation that can succeed with a value of type `T` or fail with an error. Defined implicitly as the union `Error | T`. This follows the `FailureVariant | SuccessVariant` convention.
    ```able
    ## The 'Error' interface (built-in or standard library, TBD)
    interface Error {
        fn message(self: Self) -> string
        fn cause(self: Self) -> ?Error
    }

    ## Result T is implicitly: union Result T = Error | T
    ## !Type is syntactic sugar for Result T

    ## Example function signature
    fn read_file(path: string) -> !string { ... } ## Returns Error or string
    ```

#### 11.2.2. Error/Option Propagation (`!`)

The postfix `!` operator simplifies propagating `nil` from `Option` types or `Error` from `Result` types up the call stack.

##### Syntax
```able
ExpressionReturningOptionOrResult!
```

##### Semantics
-   Applies to an expression whose type is `?T` (`nil | T`) or `!T` (`Error | T`).
-   If the expression evaluates to the "successful" variant (`T`), the `!` operator unwraps it, and the overall expression evaluates to the unwrapped value (of type `T`).
-   If the expression evaluates to the "failure" variant (`nil` or an `Error`), the `!` operator causes the **current function** to immediately **`return`** that `nil` or `Error` value.
-   **Requirement:** The function containing the `!` operator must itself return a compatible `Option` or `Result` type (or a supertype union) that can accommodate the propagated `nil` or `Error`.

##### Example
```able
## Assuming read_file returns !string (Error | string)
## Assuming parse_data returns !Data (Error | Data)
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
-   Applies to an expression whose type is `?T` (`nil | T`) or `!T` (`Error | T`).
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

For handling truly *exceptional* situations that disrupt normal control flow, often originating from deeper library levels or representing programming errors discovered at runtime. Division/Modulo by zero raises an exception.

#### 11.3.1. Raising Exceptions (`raise`)

The `raise` keyword throws an exception value, immediately interrupting normal execution and searching up the call stack for a matching `rescue` block.

##### Syntax
```able
raise ExceptionValue
```
-   **`raise`**: Keyword initiating the exception throw.
-   **`ExceptionValue`**: An expression evaluating to the value to be raised. The value must implement the `Error` interface; raising a non-`Error` value is a compile-time error.

##### Example
```able
struct DivideByZeroError {} ## Implement Error interface
impl Error for DivideByZeroError {
  fn message(self: Self) -> string { "Division by zero" }
  fn cause(self: Self) -> ?Error { nil } }
}

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

##### Error Interface

Errors must implement the `Error` interface. An optional `cause` enables error chaining.

```able
interface Error {
  fn message(self: Self) -> string
  fn cause(self: Self) -> ?Error
}

##### Standard Error Types

The standard library defines a small set of core error types implementing `Error` that the language/runtime may raise:

```able
## Arithmetic
struct DivisionByZeroError {}
impl Error for DivisionByZeroError { fn message(self: Self) -> string { "division by zero" } fn cause(self: Self) -> ?Error { nil } }

struct PanicError { message: string, cause: ?Error }
impl Error for PanicError { fn message(self: Self) -> string { self.message } fn cause(self: Self) -> ?Error { self.cause } }

## Indexing
struct IndexError { index: u64, length: u64 }
impl Error for IndexError { fn message(self: Self) -> string { `index ${self.index} out of bounds for length ${self.length}` } fn cause(self: Self) -> ?Error { nil } }
```

Language-defined raises map to these errors:

-   Division or remainder by zero raises `DivisionByZeroError`.
-   Integer overflow and shift-out-of-range raise `PanicError { message: ..., cause: nil }`.
-   Array out-of-bounds indexing raises `IndexError { index, length }`.

##### Raising Rules

-   Only values implementing `Error` may be raised with `raise`. Attempting to `raise` a non-`Error` value is a compile-time error.
-   `rescue` matches on the concrete error value (existential `Error`), including specific types like `DivisionByZeroError` and a catch-all `case _: Error`.
```

##### Rethrow

Within a `rescue` clause, `rethrow` re-raises the currently handled exception.

```able
data = risky() rescue {
  case e: ParseError => { log(e.message()); rethrow }
  case _: Error => default
}
```

##### Ensure

An `ensure` block always runs after normal completion or `rescue`. Its value is discarded; it cannot override the result unless it raises.

#### 11.3.3. Panics

Not supported. Use `raise`/`rescue` with `Error` values.

## 12. Concurrency

Able provides lightweight concurrency primitives inspired by Go, allowing asynchronous execution of functions and blocks using the `proc` and `spawn` keywords. The underlying scheduling and progress guarantees are inherited from the chosen compilation target (e.g., Go, Crystal) and are implementation-defined.

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
## Status of a Proc; variants are singleton structs. 'Failed' carries ProcError.
struct Pending;
struct Resolved;
struct Cancelled;
struct Failed { error: ProcError }
union ProcStatus = Pending | Resolved | Cancelled | Failed

## Represents an error occurring during process execution (details TBD)
## Could wrap panic information or specific error types.
struct ProcError { details: string } ## Example structure
impl Error for ProcError {
  fn message(self: Self) -> string { self.details }
  fn cause(self: Self) -> ?Error { nil } ## May wrap underlying error in future
}

## Interface for interacting with a process handle
interface Proc T for HandleType { ## HandleType is the concrete type returned by 'proc'
  ## Get the current status of the process
  fn status(self: Self) -> ProcStatus

  ## Attempt to retrieve the result value.
  ## Blocks the *calling* thread until the process status is Resolved, Failed, or Cancelled.
  ## Returns 'T' on success, or an Error on failure/cancelled. 'ProcError' implements Error.
  fn value(self: Self) -> !T

  ## Request cancellation of the asynchronous process.
  ## Best-effort and idempotent: if Pending, may transition to Cancelled; if Resolved, no effect.
  ## Races are allowed; whichever terminal state is reached first wins.
  fn cancel(self: Self) -> void
}
```

##### Semantics of Methods

-   **`status()`**: Returns the current state (`Pending`, `Resolved`, `Cancelled`, `Failed`) without blocking.
-   **`value()`**: Blocks the caller until the process finishes (resolves, fails, or is definitively cancelled).
    -   If `Resolved`, returns `value` where `value` has type `T`. For `Proc void`, this returns `void` (successful completion without data).
    -   If `Failed`, returns an error value of type `ProcError` (which implements `Error`) containing error details.
    -   If `Cancelled`, returns an error value of type `ProcError` indicating cancellation.
-   **`cancel()`**: Sends a cancellation signal to the asynchronous task. The task is not guaranteed to stop immediately or at all unless designed to check for cancellation signals.

##### Example Usage

```able
data_proc: Proc string = proc fetch_data("http://example.com")

## Check status without blocking
current_status = data_proc.status()
if match current_status { case Pending => true } { print("Still working...") }

## Block until done and get result (handle potential errors)
result = data_proc.value()
final_data = result match {
  case d: T => `Success: ${d}`,
  case e: Error => `Failed: ${e.message()}`
}
print(final_data)

## Request cancellation (fire and forget)
data_proc.cancel()
```

### 12.3. Thunk-Based Asynchronous Execution (`spawn`)

The `spawn` keyword also initiates asynchronous execution but returns a `Thunk T` value, which implicitly blocks and yields the result when evaluated. The result of a `Thunk T` is memoized: the first evaluation computes the result; subsequent evaluations return the memoized value (or error).

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
    *   If the computation fails (e.g., panics or raises an unhandled exception), evaluating the `Thunk T` yields an error value that implements `Error` (typically `ProcError`). This aligns error handling between `proc` and `spawn`.
    *   Evaluating a `Thunk void` blocks until completion. If successful, it yields `void`. If the underlying task fails, it yields a `ProcError` (which implements `Error`).

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
-   **Result Access:** `Thunk T` provides implicit result access; evaluating the thunk blocks and returns the value directly (or propagates panics). It lacks fine-grained status checks or cancellation via the handle itself.
-   **Use Cases:**
    *   `proc` is suitable when you need to manage the lifecycle of the async task, check its progress, handle failures explicitly, or potentially cancel it.
    *   `spawn` is simpler for "fire and forget" tasks where you only need the final result eventually and are okay with blocking for it implicitly (or propagating panics).

### 12.5. Synchronization Primitives (Crystal-style APIs, Go semantics)

Able provides standard library types `Channel T` and `Mutex` with APIs similar to Crystal, but the semantics are aligned with Go.

#### Channel T

Typed conduit for sending values between concurrent tasks.

Construction:
```able
## Unbuffered (rendezvous)
ch: Channel i32 = Channel.new(0)

## Buffered
ch_buf = Channel string |> new(64)
```

Core API:
```able
methods Channel T {
  ## Send a value. Blocks if the buffer is full (or if unbuffered until a receiver is ready).
  fn send(self: Self, value: T) -> void

  ## Receive a value. Blocks until a value is available or the channel is closed and drained.
  ## Returns nil when the channel has been closed and drained (Go's (x, ok) with ok=false).
  fn receive(self: Self) -> ?T

  ## Attempt to send without blocking. Returns true if the value was sent, false otherwise.
  fn try_send(self: Self, value: T) -> bool

  ## Attempt to receive without blocking. Returns a value if available, or nil if none/closed.
  fn try_receive(self: Self) -> ?T

  ## Close the channel. Further sends raise an error; receivers drain any buffered values, then receive() yields nil.
  fn close(self: Self) -> void

  ## Returns true if the channel has been closed.
  fn is_closed(self: Self) -> bool
}
```

Semantics (Go-compatible):
-   Unbuffered channels (capacity 0) are rendezvous; send/receive both block until paired.
-   Buffered channels block send when full and block receive when empty; element order is FIFO.
-   `close()` may be called by the last sender; multiple closes panic.
-   Sending on a closed channel panics.
-   `receive()` returns `?T` and yields `nil` when the channel is closed and drained.
-   Nil channels (uninitialized variables) block forever on send/receive; closing a nil channel panics.
-   Happens-before: a send happens-before the corresponding receive; closing happens-before a receive that returns the closed indication.

Iteration:
```able
## Channels are Iterable; iteration blocks, ending when closed and drained.
impl Iterable T for Channel T {
  fn iterator(self: Self) -> (Iterator T) {
    Iterator { gen =>
      loop {
        v = self.receive()
        if v == nil { break }
        gen.yield(v)
      }
    }
  }
}

for v in ch { print(v) } ## Ends when channel is closed and drained
```

Notes:
-   Multiplexing/select can be provided via library helpers or timer channels (`os.after(d)`); dedicated `select` syntax is TBD.
-   Timeouts and cancellation can be modeled using auxiliary channels or higher-level APIs.

#### Mutex

Mutual exclusion primitive to protect shared data.

Construction and API:
```able
m = Mutex.new()
m.lock()
## critical section
m.unlock()

## With helper to avoid forgetting to unlock
fn with_lock<T>(m: Mutex, f: () -> T) -> T {
  m.lock()
  result = f()
  m.unlock()
  result
}

val = with_lock(m, fn() { compute() })
```

Semantics:
-   Non-reentrant: locking a mutex already held by the current task blocks (deadlock).
-   No poisoning: if a panic occurs while the mutex is held, subsequent lockers proceed; ensuring state consistency is the user's responsibility.

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

#### Dynamic Imports (`dynimport`)

The `dynimport` statement binds identifiers from dynamically defined packages (created or extended at runtime via dynamic metaprogramming). It is resolved at runtime (interpreted context) and is invalid at compile time (AOT mode) unless an embedded interpreter is present and enabled.

*   **Syntax Forms**:
    *   Package import: `dynimport foo;`
    *   Selective import: `dynimport foo.{Point, do_something};`
    *   Aliased import: `dynimport foo as f;`
*   **Resolution**:
    *   Looks up a dynamic package object via `dyn.package("foo")` and binds requested names from its current dynamic namespace.
    *   Fails at runtime with `Error` if the package or names do not exist.
*   **Scope**: May appear at top level or in local scopes of interpreted execution. No effect in pure AOT contexts.
*   **Interoperability**: Dynamic imports can coexist with static imports; identical names follow normal shadowing rules (innermost wins).

### 13.5. Visibility and Exports (`private`)

*   **Default**: All identifiers defined at the top level (package scope) are public and exported by default.
*   **`private` Keyword**: Prefixing a top-level definition with `private` restricts its visibility to the current package only. Private identifiers cannot be imported or accessed directly from other packages.
*   **Implementations Visibility:**
    -   `impl` is a top-level definition and follows the same visibility rules as other top-level items.
        *   Default: public (exported).
        *   `private impl ...`: Visible only within the defining package.
    -   Import-scoped resolution: An implementation `(Interface, TargetType)` participates in implicit method resolution at a call site only if the `impl` is in scope at that site (defined locally or exported by a package that has been imported) and both `Interface` and `TargetType` are visible.
    -   Interface-typed values (dynamic dispatch) carry their implementation dictionary. If a package constructs a value of type `Interface` using a visible `impl` and returns it, consumers can call interface methods on that value even if the `impl` is not in scope in the consumer package.
    -   Unnamed coherence (per package scope): For any visible pair `(Interface, TargetType)`, at most one unnamed (default) `impl` may be in scope. If multiple unnamed implementations are in scope, it is a compile-time error in that package until imports are adjusted.
    -   Named implementations are never chosen implicitly. They require explicit selection (see Named Impl Invocation TBD) and follow the same visibility/import rules as other top-level items. Named impl identifiers must be unique within the importing scope; if collisions occur, use selective import with aliasing.
    -   No orphan restriction: Packages may define `impl Interface for TargetType` even if they do not own the interface or the type. Which implementation is used is determined solely by what impls are in scope in the using package (via its imports).

```able
## In package 'my_pkg'

foo = "bar" ## Public

private baz = "qux" ## Private to my_pkg

fn public_func() { ... }

private fn helper() { ... } ## Private to my_pkg

## Visibility examples

## Private concrete type and private impl; export only the interface view
private struct Hidden { value: i32 }
private impl Display for Hidden { fn to_string(self: Self) -> string { `${self.value}` } }

fn make_display() -> Display { Hidden { value: 42 } }

## In another package, resolution via imports
# import my_pkg
# d = my_pkg.make_display()
# print(d.to_string())        ## OK: dynamic dispatch works; impl dictionary came with 'd'
# h = my_pkg.Hidden { value: 7 } ## ERROR: 'Hidden' is private
#
# ## Competing third-party impls are isolated by imports
# ## If two packages provide different default impls for the same (Interface, Type),
# ## each consuming package chooses which one to import; only imported impls participate.
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
*   **Error Handling:** `Error` (base interface for errors; methods: `message(self) -> string`, `cause(self) -> ?Error`).
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
  ## Default: iterate by visiting each element. If 'each' isn't overridden,
  ## it is defined in terms of 'iterator'. Users may implement either 'each'
  ## or 'iterator' to make a type Iterable; the other is provided by default.
  fn each(self: Self, visit: T -> void) -> void {
    it = self.iterator()
    loop {
      nxt = it.next()
      match nxt {
        case v: T => visit(v),
        case IteratorEnd => break
      }
    }
  }

  ## Creates and returns a new iterator positioned at the start of the sequence.
  ## The return type '(Iterator T)' represents *any* type that implements Iterator T
  ## (an existential type / interface type).
  ##
  ## Default: if 'iterator' isn't overridden, derive it from 'each'. This
  ## constructs an iterator that drives 'each' to yield values.
  fn iterator(self: Self) -> (Iterator T) {
    Iterator { gen => self.each(gen.yield) }
  }
}
```

#### Range Interface

Provide construction of iterable ranges used by `..` (inclusive) and `...` (exclusive):

```able
interface Range S E Out {
  fn inclusive_range(start: S, end: E) -> Iterable Out
  fn exclusive_range(start: S, end: E) -> Iterable Out
}
```
The operators `StartExpr .. EndExpr` and `StartExpr ... EndExpr` are specified to
produce values that implement `Iterable Out` via `Range` implementations in scope
for the operand types.

## 15. Program Entry Point

Able programs may define one or more executables via `main` functions located in packages. This section specifies the entrypoint rules.

### 15.1. Location and Multiplicity

-   Multiple binaries are supported: any package that defines a non-private, top-level `fn main() -> void` produces an executable named after that package path (build tooling may provide renaming).

### 15.2. Signature and Arguments

-   Signature: `fn main() -> void`.
-   Command-line arguments are accessed at runtime via `os.args()`.

### 15.3. Exit Behavior

-   Returning from `main` exits with code 0.
-   An unhandled exception exits with code 1 and prints the error message.
-   To set a custom exit code, use a standard library function: `os.exit(code)`.

### 15.4. Background Work

-   The process terminates when `main` returns; background `proc`/`spawn` tasks are not awaited (fire-and-forget unless explicitly joined).

### 15.5. Constraints

-   `main` must be top-level, non-generic, non-private, and unique within any package producing a binary.

### 15.6. Example

```able
## Example Entry Point
fn main() {
    print("Hello, Able!")
    ## Access args via os.args()
    ## Exit explicitly if desired: os.exit(12)
}
```

## 16. Host Interop (Target-Language Inline Code)

Able allows embedding function bodies and package-scope preludes written in the target host language (e.g., Go, Crystal, TypeScript, Python, Ruby). This is distinct from FFI: host interop is for writing target-language code that is compiled/linked as part of the same binary the Able code compiles into.

### 16.1. Syntax

#### 16.1.1. Preludes (package scope)

```able
prelude go { import "time"; import "os" }
prelude crystal { require "time" }
prelude typescript { import { readFileSync } from "node:fs"; }
prelude python { import time }
prelude ruby { require "securerandom" }
```

Rules:
-   May appear only at package scope. Multiple preludes per target are allowed; they are concatenated in order.
-   Host code inside a prelude must follow the host language’s top-level syntax rules (e.g., imports for Go).

#### 16.1.2. Extern Host Function Bodies

```able
extern go fn now_nanos() -> i64 { return time.Now().UnixNano() }

extern crystal fn new_uuid() -> string { UUID.random.to_s }

extern typescript fn read_text(path: string) -> !string {
  try { return readFileSync(path, "utf8") } catch (e) { throw host_error(String(e)) }
}

extern python fn now_secs() -> f64 { return time.time() }

extern ruby fn new_uuid() -> string { SecureRandom.uuid }
```

Rules:
-   `extern <target> fn ...` provides a full function body in the given host language for that Able function signature.
-   Multiple extern bodies may be provided for the same function (one per target). If none match the current target:
    -   If a pure Able body exists, it is used as fallback.
    -   Otherwise, compilation errors.
-   Extern bodies are allowed only as full function bodies (no inline host expressions/macros).
-   Extern bodies are not permitted inside dynamic packages.

### 16.2. Type Mapping (Strict Core Set)

The following table summarizes mappings. Implementations MUST enforce copy-in/copy-out for arrays and MUST NOT expose pointers/references to Able-managed memory.

-   Integers: i8/i16/i32/i64/u8/u16/u32/u64/i128/u128 →
    -   Go: int8/int16/int32/int64/uint8/uint16/uint32/uint64/(128-bit if available)
    -   Crystal: Int8/Int16/Int32/Int64/Int128; UInt8/UInt16/UInt32/UInt64/UInt128
    -   TypeScript: number (use with care; IEEE-754; prefer i32/u32/f64)
    -   Python: int (arbitrary precision)
    -   Ruby: Integer (arbitrary precision)
-   Floats: f32/f64 → float32/float64 (Go); Float32/Float64 (Crystal); number (TS); float (Python); Float (Ruby)
-   Bool → bool (Go); Bool (Crystal); boolean (TS); bool (Python); TrueClass/FalseClass (Ruby)
-   String → string (Go/TS); String (Crystal/Ruby/Python)
-   Array T → []T (Go); Array(T) (Crystal); T[] (TS); list[T] (Python); Array(T) (Ruby) — copy-in/copy-out
-   ?T (Option) → nil/None/null for “no value” in the host; otherwise T mapping above
-   !T (Result) →
    -   Go: (T, error)
    -   Crystal/TS/Python/Ruby: return T or raise/throw; uncaught becomes Able Error

### 16.3. Error Mapping

-   Provide `host_error(message: string)` helper inside extern bodies to produce an Able `Error`.
-   Go: return (zero, err) or panic → Able `Error` at boundary.
-   Crystal/TypeScript/Python/Ruby: raise/throw → Able `Error` at boundary.

### 16.4. Concurrency and Execution

-   Extern bodies execute in the caller’s goroutine/fiber/thread and may block.
-   Target-specific constraints (e.g., Go package import placement, Crystal fibers) apply within preludes/bodies.

### 16.5. Placement and Hygiene

-   Preludes appear only at package scope; extern bodies only as full function bodies.
-   No extern bodies within dynamic packages.
-   Namespaces/imports in host code follow host language conventions.

### 16.6. Multi-Target and Fallback Rules

-   If multiple extern bodies for a function exist, the compiler selects the one whose `<target>` matches the configured compilation target.
-   If none match and a pure Able body exists, it is used. Otherwise, compilation errors.

### 16.7. Examples

Go:
```able
prelude go { import "time"; import "os" }

extern go fn now_nanos() -> i64 { return time.Now().UnixNano() }

extern go fn read_file(path: string) -> !string {
  data, err := os.ReadFile(path)
  if err != nil { return host_error(err.Error()) }
  return string(data)
}
```

Crystal:
```able
prelude crystal { require "uuid" }
extern crystal fn new_uuid() -> string { UUID.random.to_s }
```

TypeScript:
```able
prelude typescript { import { readFileSync } from "node:fs"; }
extern typescript fn read_text(path: string) -> !string {
  try { return readFileSync(path, "utf8") } catch (e) { throw host_error(String(e)) }
}
```

Python:
```able
prelude python { import time }
extern python fn now_secs() -> f64 { return time.time() }
```

Ruby:
```able
prelude ruby { require "securerandom" }
extern ruby fn new_uuid() -> string { SecureRandom.uuid }
```

# Todo

*   **Standard Library Implementation:** Core types (`Array`, `Map`?, `Set`?, `Range`, `Option`/`Result` details, `Proc`, `Thunk`), IO, String methods, Math, `Iterable`/`Iterator` protocol, Operator interfaces. Definition of standard `Error` interface.
*   **Type System Details:** Full inference rules, Variance, Coercion (if any), HKT limitations/capabilities.
*   **Concurrency:** Synchronization primitives (channels, mutexes?).
*   **Object Safety Rules:** Which interface methods are callable from interface-typed values; any boxing/erasure rules; formal vtable capture at upcast.
*   **Pattern Exhaustiveness:** Rules for open sets like `Error` and refutability constraints.
*   **Re-exports and Named Impl Aliasing:** Precise import/alias collision rules and diagnostics.
*   **Ranges:** Concrete type vs existential for `..` and `...` results.
*   **Tooling:** Compiler, Package manager commands, Testing framework.
