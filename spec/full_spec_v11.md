# Able Language Specification (Draft)

**Version:** 2025-11-11
**Status:** Draft

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
    *   [4.7. Type Aliases](#47-type-aliases)
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
    *   [6.7. Generator Literal (`Iterator { ... }`)](#67-generator-literal-iterator)
    *   [6.8. Arrays (`Array T`)](#68-arrays-array-t)
    *   [6.9. Lexical Details (Comments, Identifiers, Literals)](#69-lexical-details-comments-identifiers-literals)
    *   [6.10. Dynamic Metaprogramming (Interpreted)](#610-dynamic-metaprogramming-interpreted)
    *   [6.11. Truthiness and Boolean Contexts](#611-truthiness-and-boolean-contexts)
    *   [6.12. Standard Library API (Required)](#612-standard-library-api-required)
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
        *   [7.6.3. Placeholder Lambdas (`@`, `@n`)](#763-placeholder-lambdas--n)
    *   [7.7. Function Overloading](#77-function-overloading)
8.  [Control Flow](#8-control-flow)
    *   [8.1. Branching Constructs](#81-branching-constructs)
        *   [8.1.1. Conditional Chain (`if`/`elsif`/`else`)](#811-conditional-chain-ifelsifelse)
        *   [8.1.2. Pattern Matching Expression (`match`)](#812-pattern-matching-expression-match)
    *   [8.2. Looping Constructs](#82-looping-constructs)
        *   [8.2.1. While Loop (`while`)](#821-while-loop-while)
        *   [8.2.2. For Loop (`for`)](#822-for-loop-for)
        *   [8.2.3. Loop Expression (`loop`)](#823-loop-expression-loop)
        *   [8.2.4. Continue Statement (`continue`)](#824-continue-statement-continue)
        *   [8.2.5. Range Expressions](#825-range-expressions)
    *   [8.3. Non-Local Jumps (`breakpoint` / `break`)](#83-non-local-jumps-breakpoint--break)
        *   [8.3.1. Defining an Exit Point (`breakpoint`)](#831-defining-an-exit-point-breakpoint)
        *   [8.3.2. Performing the Jump (`break`)](#832-performing-the-jump-break)
        *   [8.3.3. Semantics](#833-semantics)
        *   [8.3.4. Example](#834-example)
        *   [8.3.5. Loop Break Result](#835-loop-break-result)
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
        *   [11.2.3. Handling (`or {}`)](#1123-handling-or-)
    *   [11.3. Exceptions (`raise` / `rescue`)](#113-exceptions-raise--rescue)
        *   [11.3.1. Raising Exceptions (`raise`)](#1131-raising-exceptions-raise)
        *   [11.3.2. Rescuing Exceptions (`rescue`)](#1132-rescuing-exceptions-rescue)
        *   [11.3.3. Runtime Exceptions (no panic abstraction)](#1133-runtime-exceptions-no-panic-abstraction)
12. [Concurrency](#12-concurrency)
    *   [12.1. Concurrency Model Overview](#121-concurrency-model-overview)
    *   [12.2. Asynchronous Execution (`proc`)](#122-asynchronous-execution-proc)
        *   [12.2.1. Syntax](#1221-syntax)
        *   [12.2.2. Semantics](#1222-semantics)
        *   [12.2.3. Process Handle (`Proc T` Interface)](#1223-process-handle-proc-t-interface)
    *   [12.3. Future-Based Asynchronous Execution (`spawn`)](#123-future-based-asynchronous-execution-spawn)
        *   [12.3.1. Syntax](#1231-syntax)
        *   [12.3.2. Semantics](#1232-semantics)
    *   [12.4. Key Differences (`proc` vs `spawn`)](#124-key-differences-proc-vs-spawn)
    *   [12.5. Synchronization Primitives (Crystal-style APIs, Go semantics)](#125-synchronization-primitives-crystal-style-apis-go-semantics)
    *   [12.6. `await` Expression and the `Awaitable` Protocol](#126-await-expression-and-the-awaitable-protocol)
    *   [12.7. Channel and Mutex Error Types](#127-channel-and-mutex-error-types)
13. [Packages and Modules](#13-packages-and-modules)
    *   [13.1. Package Naming and Structure](#131-package-naming-and-structure)
    *   [13.2. Package Configuration (`package.yml`)](#132-package-configuration-packageyml)
    *   [13.3. Package Declaration in Source Files](#133-package-declaration-in-source-files)
    *   [13.4. Importing Packages (`import`)](#134-importing-packages-import)
    *   [13.5. Visibility and Exports (`private`)](#135-visibility-and-exports-private)
14. [Standard Library Interfaces (Conceptual / TBD)](#14-standard-library-interfaces-conceptual--tbd)
15. [Program Entry Point](#15-program-entry-point)
16. [Host Interop (Target-Language Inline Code)](#16-host-interop-target-language-inline-code)
17. [Tooling: Testing Framework](#17-tooling-testing-framework)

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
*   **Keywords:** Reserved words that cannot be used as identifiers: `fn`, `struct`, `union`, `interface`, `impl`, `methods`, `type`, `package`, `import`, `dynimport`, `extern`, `prelude`, `private`, `Self`, `do`, `return`, `if`, `elsif`, `else`, `or`, `while`, `for`, `in`, `match`, `case`, `breakpoint`, `break`, `raise`, `rescue`, `ensure`, `rethrow`, `proc`, `spawn`, `nil`, `void`, `true`, `false`.
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
*   **Expression Separation:** Within blocks, expressions are evaluated sequentially. They are separated by **newlines** or optionally by **semicolons** (`;`). The last expression in a block determines its value unless otherwise specified (e.g., loops).
*   **Expression-Oriented:** Most constructs are expressions evaluating to a value (e.g., `if/elsif/else`, `match`, `breakpoint`, `rescue`, `do` blocks, assignment/declaration (`=`, `:=`)). Loops (`while`, `for`) evaluate to `void`.

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
*   **Type Names:** Identifiers that name a type (e.g., `i32`, `String`, `bool`, `Array`, `Point`, `Option`).
    *   **Type Arguments:** Other type expressions provided as parameters to a type name (e.g., `i32` in `Array i32`). Arguments are space-delimited.
*   **Parentheses:** Used for grouping type sub-expressions to control application order (e.g., `Map String (Array i32)`).
    *   **Nullable Shorthand:** `?TypeName` (desugars to a union `nil | TypeName`). See Section [4.6.2](#462-nullable-type-shorthand-).
    *   **Result Shorthand:** `!TypeName` (desugars to a union `Error | TypeName`). See Section [11.2.1](#1121-core-types-type-type).
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
    *   In `Map String User`, the two type parameters of `Map` are bound to `String` and `User`.
    *   In `struct Foo T { value: T }`, within the scope of `Foo`, `T` is a type variable acting as a bound parameter.
*   **Unbound Type Parameter:** A type parameter is considered **unbound** if:
    *   An argument for it is not specified.
    *   The wildcard placeholder `_` is explicitly used in its position. See Section [4.4](#44-reserved-identifier-_-in-types).
*   **Concrete Type:** A type expression denotes a **concrete type** if *all* of its inherent type parameters (and those of any nested types) are bound to specific types or type variables. Values can only have concrete types.
    *   Examples: `i32`, `String`, `Array bool`, `Map String (Array i32)`, `Point`, `?String`.
*   **Polymorphic Type / Type Constructor:** A type expression denotes a **polymorphic type** (or acts as a **type constructor**) if it has one or more unbound type parameters. Type constructors cannot be the type of a runtime value directly but are used in contexts like interface implementations (`impl Mappable A for Array`) or potentially as type arguments themselves (if full HKTs are supported). Interface-typed existentials such as `Display` are concrete runtime types; `Display _` (an interface with an unbound parameter) is not a concrete type unless all its parameters are bound. When using interface names in type positions (existential/dynamic types), all interface type parameters must be fully bound. For example, `Display` is valid, `Display String` (if parameterized) is valid, but `Mappable _` is not a valid type and cannot appear in type positions.
    *   Examples:
        *   `Array` (parameter is unspecified) - represents the "Array-ness" ready to accept an element type.
        *   `Array _` (parameter explicitly unbound) - same as above.
        *   `Map String` (second parameter unspecified) - represents a map constructor fixed to `String` keys, awaiting a value type. Equivalent to `Map String _`.
        *   `Map _ bool` (first parameter unbound) - represents a map constructor fixed to `bool` values, awaiting a key type.
        *   `Map` (both parameters unspecified) - represents the map constructor itself. Equivalent to `Map _ _`.
        *   `?` (type-level operator) denotes the nullable constructor mapping `T` to `nil | T`; it is not a standalone type.


    Value positions require concrete types (no unbound parameters):

    - A parameter, variable, field, or return type annotation must be a concrete type. Using a type constructor with unbound parameters (e.g., `Array`, `Map String _`, `Mappable _`) in a value position is invalid.
    - Instead, make the function/definition generic and bind the parameters via type variables.

    ```able
    ## VALID (generic binds the element type)
    fn len<T>(xs: Array T) -> u64 { ... }

    ## INVALID (unbound element type in value position)
    # fn bad(x: Array) -> u64 { ... }

    ## VALID (map with generic value type)
    fn keys<V>(m: Map String V) -> Array String { ... }
    ```

    - Call-site inference binds generic parameters from arguments/results; annotations that leave parameters unbound are rejected. This reconciles the rule “values can only have concrete types” with convenient polymorphism at call sites.

    Interface/existential use must be fully bound in type positions:

    - Using an interface name as a type denotes a dynamic/existential type and must be fully bound with all its own parameters (if any). `Display` (no params) and `Display String` (if parameterized) are valid types; `Mappable _` is not.

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
        >   **Example:** `Array i32`, `Map String User`, `struct Pair T U`, `interface Mappable K V`
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

#### 4.1.6. Subtyping and Inference (clarifications)

- There is no general structural subtyping in Able. Interface implementations do not create a static subtyping lattice. Instead, use:
  - Algebraic unions for disjunction of alternatives.
  - Existential/interface types for dynamic dispatch views.
- Type argument inference occurs at function call sites from argument and expected return types. It does not permit leaving unbound parameters in value annotations. Where the compiler cannot infer generics, specify them explicitly, e.g., `identity<i64>(0)`.

#### 4.1.7. Variance, Coercion, and Higher-Kinded Types (minimal rules)

- Variance:
  - Type parameters are invariant unless a type explicitly declares variance in its definition (TBD syntax). Stdlib containers are invariant by default (`Array T` is invariant in `T`).
- Coercion:
  - There is no implicit subtyping-based coercion. Coercions occur only via explicit constructors/conversions or unions (upcast into a union). Numeric widening follows operator rules in §6.3.2; no silent narrowing.
- HKTs (higher-kinded interfaces):
  - Supported at the interface level via `for M _` patterns (e.g., `interface Mappable A for M _`). Implementations target type constructors as in §10.2.3. HKTs for concrete types beyond interface usage (e.g., first-class type constructors in arbitrary positions) are intentionally limited and may be expanded later.

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
| `String` | Immutable sequence of Unicode chars (UTF-8) | `"hello"`, `""`, `` `interp ${val}` `` | Double quotes or backticks (interpolation). String literals evaluate to `String` directly.      |
| `bool`   | Boolean logical values                        | `true`, `false`                     |                                                 |
| `char`   | Single Unicode code point        | `'a'`, `'π'`, `'💡'`, `'\n'`, `'\u{1F604}'` | Single quotes. Supports escape sequences.       |
| `nil`    | Singleton type representing **absence of data**. | `nil`                               | **Type and value are both `nil` (lowercase)**. Often used with `?Type`. Falsy. |
| `void`   | Singleton unit type with exactly one value.   | `void`                              | Represents successful completion without data. Distinct from `nil`. Truthy. Any expression may be implicitly coerced to `void`; the produced value is discarded. |

*(See Section [6.1](#61-literals) for detailed literal syntax.)*

> **Single `String` type + naming rules**
>
> Able exposes exactly one textual type: `String`. String literals evaluate to `String` directly, and the standard library methods described in §6.12.1 live on that same type. Naming follows a simple rule: lowercase is reserved for scalar primitives (`bool`, numeric widths, `char`, `nil`, `void`); every other built-in/nominal type uses PascalCase (`String`, `Array`, `Result`, `Error`, `Range`, etc.). Older material may show lowercase `string`; treat it only as a historical alias for `String`.

### 4.3. Type Expression Syntax Details

*   **Simple:** `i32`, `String`, `MyStruct`
*   **Generic Application:** `Array i32`, `Map String User` (space-delimited arguments)
*   **Grouping:** `Map String (Array i32)` (parentheses control application order)
*   **Function:** `(i32, String) -> bool`
*   **Nullable Shorthand:** `?String` (Syntactic sugar for `nil | String`, see Section [4.6.2](#462-nullable-type-shorthand-))
*   **Result Shorthand:** `!String` (Syntactic sugar for `Error | String`, see Section [11.2.1](#1121-core-types-type-type))
*   **Wildcard:** `_` denotes an unbound parameter (e.g., `Map String _`).

### 4.4. Reserved Identifier (`_`) in Types

The underscore `_` can be used in type expressions to explicitly denote an unbound type parameter, contributing to forming a polymorphic type / type constructor. Example: `Map String _`.

### 4.5. Structs

Structs aggregate named or positional data fields into a single type. Able supports three kinds of struct definitions: singleton, named-field, and positional-field. A single struct definition must be exclusively one kind. All fields are mutable.

Note on immutability: The language does not provide `const`/`immutable` qualifiers for bindings or fields. Immutability is achieved by design (e.g., exposing no mutators, returning new values) or by using library-provided persistent data structures. Projects may adopt conventions enforcing single-assignment or immutable APIs; the core language does not enforce it.

#### 4.5.4. Immutability Patterns (Guidance)

- Prefer API designs that avoid in-place mutation. Provide “with-” style constructors that copy and override selected fields:
  ```able
  Address { ...base, zip: 90210 }
  ```
- Expose read-only accessors and avoid exporting mutators; keep mutating helpers internal.
- Offer persistent data structures in stdlib for common types (e.g., vectors/maps) and document their semantics.
- Use builders for complex construction, then hand out values with no mutators.
- Concurrency: prefer sharing immutable data across tasks to avoid coordination costs; if mutation is required, use `Mutex` or channels (§12.5).

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
-   **`[GenericParamList]`**: Optional space-delimited generics (e.g., `T`, `K V: Display`). Constraints can be inline or in the `where` clause. When omitted, Able infers generic parameters from free type names appearing in the field types (see §7.1.5). For example, `struct Box { value: T }` implicitly declares `struct Box T { value: T }`.
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
Create a new instance based on other instances using spread syntax. The literal always produces a *fresh* struct; no source value is mutated.

```able
StructType { ...Source1, ...Source2, FieldOverride: NewValue, ... }
addr = Address { ...base_addr, zip: "90210" }
```

Semantics:

1.  **Named structs only:** Functional update applies exclusively to named-field structs. Positional structs raise `"Functional update only supported for named structs"` if `...` appears inside the literal.
2.  **Type compatibility:** Every spread expression must evaluate to the *same struct type* (after substituting concrete type arguments) as the literal being constructed. For generics, both the literal and the spread sources must share identical type arguments after inference; otherwise evaluation raises `"Functional update source must be same struct type"`.
    ```able
    Point T := struct { x: T, y: T }
    p_i32 := Point i32 { x: 1, y: 2 }
    origin_f64 := Point f64 { x: 0.0, y: 0.0 }

    Point { ...p_i32, x: 10 }      ## OK (T inferred as i32)
    Point { ...origin_f64, x: 5 }   ## Error: mismatched type arguments
    ```
3.  **Evaluation order:** Entries evaluate left-to-right. Each spread expression runs once and copies its fields into the accumulating result before later spreads/overrides run. Field expressions (`field: expr`) evaluate after all earlier spreads in the literal body.
4.  **Overwrite rules:** Later sources overwrite earlier sources field-by-field. Explicit field clauses (including shorthand `field`) override whatever value the same field had from prior spreads.
5.  **Partial spreads:** There is no syntax to spread a subset of fields; each `...Source` copies every field defined on the struct. To drop a field, simply override it with a new value (including `nil` where allowed).
6.  **Diagnostics:**
    - Spreading a non-struct or a struct of a different nominal type raises `"Functional update source must be same struct type"`.
    - Using `...` in a positional struct literal raises `"Functional update only supported for named structs"`.
    - Missing required fields after all spreads/overrides are processed raise the same diagnostics as ordinary struct literals.
7.  **Immutability guarantee:** Since updates allocate a fresh struct, later mutation of the new value does not affect the sources. This mirrors record-update semantics in other languages and keeps aliasing predictable.
8.  **Compile-time enforcement:** Static checkers are encouraged to verify type compatibility and redundant spreads, but the runtime rules above are the authoritative behaviour.

Example combining multiple spreads and overrides:

```able
Address := struct {
  line1: String,
  line2: ?String,
  city: String,
  state: String,
  zip: String
}

home := Address {
  line1: "123 Able St",
  line2: nil,
  city: "San Mateo",
  state: "CA",
  zip: "94401"
}

shipment := Address {
  ...home,
  line2: "Attn: Receiving",
  zip: "94070",
  ...overrides   ## later overrides win
}
```
`shipment` copies every field from `home`, adjusts two of them, then applies whatever entries `overrides` supplies. None of the source structs are mutated.

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
-   **`[GenericParamList]`**: Optional space-delimited generics. Constraints can be inline or in the `where` clause. Free type names inside the field list implicitly become generic parameters when no list is provided.
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
-   The number of values supplied in a positional literal must match the field count. Otherwise evaluation raises `"Struct 'Identifier' expects N fields, got M"`.

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

Define a new type as a composition of existing variant types using `|`. The order of variants in the definition (`A | B` vs `B | A`) is not semantically significant for type checking. Implementations may choose any internal representation; authors should not rely on variant position. For readability the spec adopts the conventional order `FailureVariant | SuccessVariant` where applicable (e.g., `nil | T`, `Error | T`). Operators such as propagation (`!`) are defined by the presence of specific failure variants (`nil` for `Option`, a value implementing `Error` for `Result`), not by their position in the union.

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
union IntOrString = i32 | String

## Payload-bearing variants (named fields)
struct Circle { radius: f64 }
struct Rectangle { width: f64, height: f64 }
struct Triangle { a: f64, b: f64, c: f64 }
union Shape = Circle | Rectangle | Triangle
```

Note: Each `VariantType` in a union may be a concrete type (e.g., `i32`, `Point`), another union, a generic application (e.g., `Array i32`), or an interface type (e.g., `Error`, `Display`). Using an interface name as a variant denotes an existential value implementing that interface (dynamic dispatch). For example, `Result T = Error | T` is valid and corresponds to values that are either a concrete error type implementing `Error` or a success value of type `T`.

Interface variants: construction and matching

- Construction/upcast: When a union lists an interface like `Error`, any concrete value whose type implements that interface can be used directly; the upcast to the existential interface variant is implicit.

``` able
res1: !String = "ok"              ## success variant
res2: !String = IndexError { index: 5, length: 2 } ## implicitly upcasts to Error | String
```

- Matching/narrowing: Pattern matching can use typed patterns to narrow existential interface variants either to the interface itself or to specific concrete error types.

``` able
r: !i32 = some_op()
msg = r match {
  case n: i32       => `ok:${n}`,      ## success
  case e: Error     => e.message(),    ## interface-wide handler (open set)
}

msg2 = r match {
  case n: i32             => `ok:${n}`,
  case e: IndexError      => `bad index ${e.index}`,
  case _: Error           => "failed",  ## ensure coverage of other errors
}
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
`?Type` is syntactic sugar for the union `nil | Type`. This follows the `FailureVariant | SuccessVariant` convention. The `?` operator applies only to type positions (it does not prefix expressions or constructors).

##### Examples
```able
name: ?String = "Alice"
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
# res_ok: Result String = "Data loaded"
# res_err: Result String = SomeConcreteError { details: "File not found" }

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

maybe_name: ?String = get_name_option()
display_name = maybe_name match {
  case s: String => s, ## Matches non-nil string
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

### 4.7. Type Aliases

Type aliases introduce a new identifier that stands for an existing type expression. They are purely compile-time substitutions: an alias does **not** define a fresh nominal type, allocate storage, or change runtime representation. Everywhere the alias appears, the compiler behaves exactly as if the aliased type expression had been written directly.

#### 4.7.1. Syntax

```able
type Identifier [GenericParamList] [where <ConstraintList>] = TypeExpression
```

-   **`type`**: Keyword introducing the alias. Aliases are declared at package scope alongside structs/unions/interfaces; local aliases inside functions/blocks are prohibited to keep the type namespace static and tooling-friendly.
-   **`Identifier`**: Name of the alias. Alias names share the type namespace with structs/unions/interfaces and must be unique within a module.
-   **`[GenericParamList]`**: Optional **space-delimited** generic parameters with inline constraints (same delimiter rules as §4.1.5). Angle-bracket syntax (`<T>`) is not permitted for aliases. A trailing `where` clause may express additional bounds.
-   **`TypeExpression`**: Any valid type expression (primitives, structs, unions, interface existentials, generic applications, shorthand such as `?T`/`!T`, other aliases, etc.). The right-hand side may reference earlier aliases so long as expansion terminates.

Examples:

```able
type UserID = u64
type Result T = Error | T
type MapOf K V = Map K V
type Either L R = L | R

import time
type Timestamp = time.Instant
```

#### 4.7.2. Generic Aliases & Substitution

-   Generic aliases behave like other generic type definitions: every use must ultimately bind all parameters (either explicitly or via inference) so the alias denotes a concrete type. Wildcards (`_`) follow the same rules described in §4.1.4.
-   Expansion is capture-avoiding and eager. When type checking encounters an alias name, it substitutes the right-hand side—plugging in any provided type arguments—before performing equality checks, constraint solving, or interface lookups. Consequently, `type UserID = u64` means `UserID` and `u64` are indistinguishable to the typechecker.
-   Because aliases are transparent, they cannot be used to work around coherence rules. Declaring both `impl Display for u64` and `impl Display for UserID` is a redeclaration error; both target the same underlying type.

#### 4.7.3. Scope, Visibility, and Imports

-   Aliases obey the module’s visibility/export mechanics (e.g., `pub type` in host scaffolding, explicit export lists, etc.). Importing a module brings its exported aliases into scope just like other types: `import net.http; request: http.Request = ...`.
-   Within a module, aliases participate in lexical scoping: once declared, they are visible throughout the remainder of the file. Re-declaring the same alias name in the same scope is an error; shadowing via nested modules is allowed but discouraged unless intentionally creating versioned namespaces.
-   Because aliases live in the type namespace, they do not interfere with value-level identifiers. `type Path = String` can coexist with `fn Path() -> Path` (though readability may suffer).
-   Importing aliases is **type-only**: `import pkg.{Alias}` (or `import pkg.{Alias::Local}`) makes the alias name available in type positions within the importing module, but does **not** create a runtime value binding. Renames apply only to the type namespace.

#### 4.7.4. Interaction with `methods`, `impl`, and Type-Based Features

-   `methods`/`impl` headers accept any type expression. If an alias appears there, it is expanded before registration. Attaching methods to an alias therefore attaches them to the underlying type:
    ```able
    type UserID = u64
    methods UserID {
      fn to_hex(self: UserID) -> String { self.format("0x%X") }
    }
    ## Equivalent to `methods u64 { ... }` and subject to the same restrictions on primitives.
    ```
-   Interface constraints, `where` clauses, and type inference treat aliases identically to their expansions. Diagnostics should mention both names when it improves clarity (e.g., “`UserID` (alias of `u64`) does not implement `Hash`”).
-   Aliases have no runtime identity, so they do not affect reflection metadata, host-language FFI bindings, or serialization schemas; exporters should emit the canonical underlying type.

#### 4.7.5. Recursion and Termination Rules

-   Alias resolution must terminate. Direct self-reference (`type Node = Node`) or mutually-recursive aliases without an intervening nominal type are rejected at compile time (`"Type alias expansion does not terminate"`).
-   Recursive data structures should be modeled with structs/unions and may then be wrapped by an alias for convenience:
    ```able
    struct ListNode T { value: T, next: ?ListNode T }
    type List T = ?ListNode T
    ```
    Here `List` is legal because the recursion flows through the nominal `ListNode` definition rather than the alias itself.

#### 4.7.6. Usage Examples

```able
type Milliseconds = u64
type Handler T = fn(T) -> void
type RequestResult = Result Response

fn set_timeout(duration: Milliseconds, cb: Handler void) {
  scheduler.schedule(duration, cb)
}

fn fetch(url: String) -> RequestResult {
  http.request(url)
}
```

`Milliseconds` is interchangeable with `u64`, but the alias documents intent and enables API-specific validation without inventing a new nominal type. The same applies to `RequestResult`, which is simply `Result Response` with a descriptive name.
```

## 5. Bindings, Assignment, and Destructuring

This section defines variable binding, assignment, and destructuring in Able. Able uses `=` and `:=` for binding identifiers within patterns to values. `=` primarily handles reassignment/mutation but implicitly declares bindings when no matching identifier exists; `:=` is the tool for explicitly declaring new bindings (and shadowing) while optionally reassigning existing names in the same pattern when at least one new name is introduced. Bindings are mutable by default.

### 5.0. Mutability Model

-   **Binding vs Value mutability:** There is a strong distinction between a binding (the name-to-location association) and the value bound to it. These are independent:
    -   Binding mutability (rebinding): By default, bindings may be reassigned; `=` can rebind an existing name to a new value; `:=` declares new bindings (and may reassign existing names in the same pattern when at least one new name is introduced).
    -   Value mutability (in-place mutation): Values are not assumed to be immutable. Unless a type or API is explicitly documented as immutable, values are generally mutable (e.g., struct fields, array elements, map entries) and can be changed in place.
-   **Reference semantics:** All Able values are references to heap-allocated objects (including structs, arrays, maps, user-defined aggregates). Mutating a value through one binding is observable through every other binding that aliases the same value. There is no implicit copy-on-write or value-type behaviour; APIs that need isolation must perform explicit `clone` operations or create fresh values.
-   **Important distinction:** Rebinding a name (e.g., `x = ...`) replaces which value the name refers to. Mutating a value (e.g., `x.field = ...`, `arr[i] = ...`) changes the underlying value itself. Even if you avoid rebinding `x`, mutating through `x` will update the value that any other aliasing references observe.
-   **Design note:** Favor immutable designs where appropriate by using types that expose no mutators or are explicitly documented as immutable. Projects may also adopt single-assignment discipline by policy; the language does not add per-binding mutability annotations.

Additional note: There is no `const` keyword and no per-field immutability modifier. Mutation control is expressed by API design and by choosing value vs. rebinding operations.

### 5.1. Operators (`:=`, `=`)

*   **Declaration (`:=`)**: `Pattern := Expression`
    *   Declares **new** mutable bindings in the **current** lexical scope for identifiers introduced by `Pattern` that do not already exist in the current scope, and
    *   Reassigns identifiers in `Pattern` that already exist in the current scope.
    *   At least one identifier in the `Pattern` must be new to the current scope; otherwise, it is a compile-time error ("no new bindings on left side of :=").
    *   This is the **required** operator for **shadowing**: if an identifier introduced by `Pattern` has the same name as a binding in an *outer* scope but not in the current scope, `:=` creates a new, distinct binding in the current scope that shadows the outer one.
    *   Example (Shadowing and update):
        ```able
        package_var := 10 ## Assume declared at package level

        fn my_func() {
          ## package_var is accessible here (value 10)
          package_var := 20  ## Declares NEW local binding 'package_var', shadows package-level one
          print(package_var)  ## prints 20 (local)

          ## '=' reassigns the innermost binding in scope. With a local 'package_var',
          ## this modifies the local, not the package-level binding.
          package_var = 30  ## Reassigns the local 'package_var'
        }
        my_func()
        print(package_var)  ## prints 10 (package-level was unaffected by local ':=')
        ```

*   **Assignment (`=`)**: `LHS = Expression`
    *   Performs **reassignment** of existing bindings or **mutation** of fields/elements. If no matching binding exists, it implicitly declares one in the current scope (see §5.3.1). Use `:=` when you need the compiler to enforce “this introduces a new binding” (e.g., for deliberate shadowing).
    *   **If `LHS` is an Identifier:**
        *   Able looks up the identifier in the current lexical scope chain. If found, the binding is reassigned.
        *   If no binding exists in any accessible scope, `=` allocates a new mutable binding in the current scope and assigns the RHS value.
        *   Attempting to assign to an immutable binding or a value imported as read-only is still an error.
    *   **If `LHS` is a Destructuring Pattern (e.g., `Point {x, y}`, `[a, b]`):**
        *   The pattern is matched atomically. For each identifier inside the pattern:
            - If a mutable binding exists (current or enclosing scope), it is reassigned.
            - If no binding exists, the identifier is declared in the current scope and then assigned.
        *   Match failure leaves all bindings untouched and returns an `Error` (see §5.3).
    *   **If `LHS` is a Field/Index Access (`instance.field`, `array[index]`):**
        *   Performs **mutation** on the specified field or element, provided it's accessible and mutable. These forms never declare new bindings.
    *   It is a compile-time error if `LHS` attempts to reassign an immutable binding, mutate a non-existent field/index, or otherwise violates access checks. Implementations without a static checker must surface a runtime error instead of silently ignoring the assignment.
    *   Example (Implicit declaration vs. reassignment):
        ```able
        count = 1        ## No prior binding => declares `count` in the current scope
        count = count+1  ## Reassignment

        do {
          count = 5      ## Reassigns outer binding (no new local binding introduced)
          local_total = 0  ## Declares new binding because none existed
        }

        { acc, delta } = { acc: 10, delta: 5 }
        ## Declares acc/delta if missing; otherwise reassigns them atomically
        ```

### 5.2. Patterns

Patterns are used on the left-hand side of `:=` (declaration) and `=` (assignment) to determine how the value from the `Expression` is deconstructed and which identifiers are bound or assigned to.

Pattern binding forms (named-field contexts) follow these rules (the same grammar applies in `:=` / `=` assignments and in `match` / `rescue` patterns):
- `::` is the **rename operator** for patterns. It never performs namespace traversal; use dot (`.`) for packages/static methods.
- `field` is shorthand for `field::field` (bind the field to a local of the same name).
- `field::binding` binds the field to the specified binding name.
- `field: Type` binds the field to a same-named local with a type annotation. `:` never renames.
- `field::binding: Type` combines a rename with a type annotation.
- Any of the above may be followed by a nested pattern to destructure the field value, e.g., `field::b: Address { street, city }` destructures the field value after binding/annotating `b`.

#### 5.2.1. Identifier Pattern

The simplest pattern binds the entire result of the `Expression` to a single identifier.

*   **Syntax**: `Identifier`
*   **Usage (`:=`)**: Declares a new binding `Identifier`.
    ```able
    x := 42
    user_name := fetch_user_name()
    my_func := { a, b => a + b }
    ```
*   **Usage (`=`)**: Reassigns the nearest mutable binding if one exists; otherwise declares a new binding in the current scope before assigning. Use `:=` when you intend to shadow an outer binding and want the compiler to enforce “new binding required.”
    ```able
    x = 50      ## Reassigns existing x if present, otherwise declares it
    counter = 0 ## Declares `counter` if not already defined
    ```

#### 5.1.1. Typed Declarations with `:`

You may annotate identifiers (or nested pattern bindings) with a type using `Identifier: Type` syntactic sugar alongside either operator:

-   `Identifier: Type := Expression`
-   `Identifier: Type = Expression`

In both cases the type annotation narrows what values may be bound, while the operator determines whether new bindings must be introduced (`:=`) or whether the runtime may fall back to reassigning/implicitly declaring (`=`).

Key rules:

1.  **Evaluation order:** The RHS always evaluates first. Errors raised before binding mean no mutation/declaration occurs.
2.  **Type checking:** The evaluated value must be assignable to `Type`. Static checkers should enforce this, but runtimes must raise `"Typed pattern mismatch"` (or a specific diagnostic) if the value fails the annotation at execution time.
3.  **Operator behavior:**
    -   `:=` enforces the “at least one new binding in the current scope” rule (§5.1). Typed names behave exactly like untyped names with the additional type assertion.
    -   `=` follows the rules described earlier: reuse an existing binding if present; otherwise implicitly declare one in the current scope. The annotation applies in either case, effectively declaring the static type when a new binding is created.
4.  **Shadowing & compatibility:** When reassigning through either operator, the annotated type must be compatible with the binding’s declared type. Attempting to narrow to an incompatible type is a compile-time error.
5.  **Patterns:** Typed fields inside destructuring patterns inherit these semantics. For example, `{ count: i64 := arr.size() }` declares `count` with type `i64`, while `{ count: i64 = arr.size() }` reassigns/declares depending on existing bindings.

Examples:

```able
total: i64 := 0          ## Declares `total` with type i64
total: i64 = total + 5   ## Reassigns after asserting the type stays i64

ratio: f64 = compute_ratio()  ## Declares if `ratio` was absent, otherwise reassigns

fn demo() {
  threshold: u32 := 10   ## Forces a new local binding, shadowing any outer one
  limit: u32 = threshold ## Reuses the just-declared binding
}
```

Typed destructuring remains governed by §5.2.7; the addition here simply clarifies that annotations work uniformly with both `:=` and `=` while respecting their respective binding rules.

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

Note on `_` vs `%` vs `@`:

- `_` in patterns: wildcard (ignore value) — only valid in pattern positions (§5.2). Not an identifier.
- `_` in types: unbound type parameter placeholder (§4.4). Forms a polymorphic type constructor.
- `%` in expressions: modulo operator (§6.3.2). No dedicated pipe-topic token exists.
- `@`, `@n` in expressions: placeholder lambdas (§7.6.3).

#### 5.2.3. Struct Pattern (Named Fields)

Destructures instances of structs defined with named fields.

*   **Syntax**: `[StructTypeName] { FieldEntry, ... }`
    *   `StructTypeName` (optional): If present, the value must be an instance of this struct; otherwise type is inferred/checked.
    *   `FieldEntry` forms:
        - `field` — shorthand binding (`field::field`), binds the field to a local of the same name.
        - `field::binding` — binds the field to `binding` (rename). `::` is reserved for pattern renames, never for namespace traversal.
        - `field: Type` — binds the field to a same-named local with a type annotation (`:` never renames).
        - `field::binding: Type` — binds to `binding` with a type annotation.
        - Any of the above may be followed by a nested pattern to destructure the field value, e.g., `address::addr: Address { street, city }`.
    *   `...` is not supported; unmentioned fields are ignored, extra pattern fields still error.

*   **Example**:
    ```able
    struct Point { x: f64, y: f64 }
    struct Address { street: String, city: String, zip: String }
    struct Person { name: String, age: i32, address: Address }

    p := Point { x: 1.0, y: 2.0 }
    home := Address { street: "123 Main", city: "SF", zip: "94107" }
    david := Person { name: "David", age: 40, address: home }

    ## Shorthand bindings
    { x, y } := p                 ## binds x = 1.0, y = 2.0
    { x::x_coord } := p           ## binds x_coord = 1.0

    ## Type annotations on bindings
    { x: f64, y } := p            ## x annotated f64, y inferred
    { x::x64: f64 } := p          ## rename + type annotation

    ## Nested destructuring with rename + type
    Person { name::who, address::addr: Address { street, city, zip } } := david
    ## who="David", addr=david.address, street/city/zip bound from addr

    ## Assignment (reuse existing bindings)
    existing_x = 0.0; existing_y = 0.0
    Point { x::existing_x, y::existing_y } = p ## reassigns existing_x, existing_y

    ## Match/rescue patterns use the same grammar
    p match { case Point { x, y } => `Point: ${x}, ${y}` }
    ```
*   **Semantics**: Matches fields by name. If `StructTypeName` is present, the value must be an instance of that struct type; otherwise evaluation fails (`"struct type mismatch in destructuring"`). Each referenced field must exist (`"Missing field 'name' during destructuring"`); extra pattern fields are an error and unmentioned fields are ignored. Type annotations assert compatibility on the bound identifier (widening allowed, narrowing diagnosed). The same rules apply in assignment (`=`), declaration (`:=`), and pattern positions in `match`/`rescue`.

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
    existing_a = 0 ## existing binding
    existing_b = 0 ## existing binding
    { existing_a, existing_b } = IntPair { 100, 200 } ## Assigns 100 to existing_a, 200 to existing_b
    { new_x, new_y, new_z } := coord ## Declare new_x, new_y, new_z in current scope
    ```
*   **Semantics**: Matches fields by position. If `StructTypeName` is present, the value must be an instance of that struct type; otherwise the match fails. The number of patterns must equal the field arity (`"struct field count mismatch in destructuring"`). Patterns expect positional instances (`"expected positional struct value"`).

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
    existing_head = 0 ## existing binding
    ## Note: Assigning to a rest pattern with '=' is likely invalid or needs careful definition.
    ##       Typically, '=' would assign to existing elements by index/pattern.
    [existing_head, element_1] = [1, 2] ## Assigns 1 to existing_head; element_1 must already exist
    ```
*   **Semantics**: Matches elements by position. Fails if the array has fewer elements than required by the non-rest patterns. Rest patterns must be identifiers or `_`; other forms raise `"unsupported rest pattern type"`.
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
      case s: String => print(s),
      case n: i32 => print(n + 1),
      case _: Error => log("got error"),
      case _ => print("unknown")
    }
    ```
*   **Semantics**: Acts as a runtime type guard within `match`/`rescue`. This does not introduce new static subtyping; it narrows within the matched branch only.

Typed patterns in `:=`/`=`:

- Typed patterns are also permitted on the left-hand side of `:=` and `=` within struct, array, or standalone identifier patterns. The assignment/declaration succeeds only if the runtime value conforms to the annotated type; otherwise evaluation raises `"Typed pattern mismatch in assignment"`.
- Compile-time checkers **may** surface diagnostics when they can statically prove the mismatch (e.g., in warn-mode runs), but these diagnostics are advisory. A program that proceeds will still evaluate according to the runtime rule above, producing an `Error` value if the value fails the annotation at execution time.

``` able
## Union destructuring with typed pattern in assignment
val: ?i32 = get_opt()
x: i32 = 0
_ = { x: i32 } = val   ## succeeds only if val is a non-nil i32; else yields Error

## Direct typed identifier in declaration (:=) from a dynamic value
{ n: i32 } := next_value()  ## declares n if next_value() is an i32; else Error
```

Union values are destructured with `match` (§8.1.2). Assignment/declaration does not perform variant selection; use `match` to branch on variants, then use typed patterns inside the selected branch as needed.

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

1.  **Evaluation Order**: The `Expression` (right-hand side) is evaluated first to produce a value. Any side effects of the RHS occur before any binding or reassignment effects on the LHS.
2.  **Matching & Binding/Assignment**: The resulting value is then matched against the `Pattern` or `LHS` (left-hand side).
    *   **`:=`**: Resulting value matched against `Pattern` (LHS). Declares new bindings in the current scope for names not already bound in the current scope and reassigns names that already exist in the current scope. At least one new name must be introduced; otherwise, it is a compile-time error.
    *   **`=`**: Resulting value matched against `LHS` pattern/location specifier.
        *   If `LHS` is a destructuring pattern, all identifiers within it must already exist as accessible, mutable bindings. Otherwise, it is a compile-time error. Use `:=` to declare new bindings.
        *   If `LHS` is an identifier or field/index access, it performs **reassignment** or **mutation** on the existing target.
        *   It is a compile-time error if any target location for reassignment/mutation does not exist, is not accessible, or is not mutable.
    *   **Match Failure**: If the value's structure or type does not match the pattern/LHS, the assignment/declaration expression evaluates to an `Error` value (a value whose type implements `Error`).
3.  **Mutability**: Bindings introduced via `:=` are mutable by default. The `=` operator requires the target location(s) (variable, field, index) to be mutable.
4.  **Scope**: `:=` introduces bindings into the current lexical scope. `=` operates on existing bindings/locations found according to lexical scoping rules.
5.  **Type Checking**: The compiler checks for compatibility between the type of the `Expression` and the structure expected by the `Pattern`/`LHS`. Type inference applies where possible.
6.  **Result Value and Truthiness**:
    *   Success: Both assignment (`=`) and declaration (`:=`) expressions evaluate to the **value of the RHS** after successful binding/assignment. In boolean contexts, this is truthy unless the RHS itself is falsy by §6.11 (e.g., `nil`, `false`, or an `Error`).
    *   Failure: If the match fails, the expression evaluates to an `Error` value (which is falsy in boolean contexts). Implementations should provide a specific error describing the mismatch; programs may branch on success/failure using `if/elsif/else`.

#### 5.3.1. Mutable Reassignment (`=`)

Plain `=` is the mechanism for mutating existing bindings, fields, and indexed locations. Able uses mutable lexical storage for every binding introduced via `:=`; reassigning that binding writes a new reference into the same storage slot rather than creating a shadow copy.

*   **Storage semantics:** `:=` allocates storage for each identifier it introduces. `=` reuses that storage and overwrites it with the newly evaluated value. Reads that occur after the assignment within the same scope (including inside loops) observe the updated value.
*   **Declaration fallback:** If a name referenced inside the `=` pattern is not bound in the current lexical scope (nor any enclosing scope) the runtime treats the assignment as a declaration and allocates a fresh mutable binding in the current scope before writing the value. In other words, `=` prefers reassignment but degrades gracefully into declaration when no binding exists. Use `:=` when you want the compiler to enforce “this must declare something new”; use `=` when you want “assign here, declaring if necessary.”
    ```able
    ## `result` already exists, so `=` reassigns
    result := 0
    result = compute()

    ## `total` does not exist yet; `=` implicitly declares it in the current scope
    total = 42
    print(total)  ## OK, binding created moments earlier
    ```
    Shadowing rules follow the normal lexical model: if an outer scope defines `name` but the current scope does not, `name = value` mutates the outer binding. To deliberately shadow, use `name := value`.
*   **Snapshot matching:** When the left-hand side is a destructuring pattern, the runtime evaluates the RHS once, attempts to match the resulting value against the pattern using the *current* bindings, and only commits the reassignment if the entire pattern succeeds. Match failure leaves every target binding untouched and surfaces the usual `Error` value for failed assignment.
*   **Evaluation order:** RHS side effects happen first. Once the RHS produces a value, destructuring proceeds left-to-right. Individual identifiers update in textual order. Field/index targets (`point.x`, `arr[i]`) evaluate their receiver/index expressions exactly once and then mutate the resolved storage location. Implementations must not re-evaluate the receiver or index when lowering compound assignments (e.g., `arr[i] += 1`).
*   **Mutability checks:** Attempting `=` against an immutable binding, a missing identifier, or a non-assignable location is a compile-time error. Interpreters running without the static checker should surface a runtime error instead of silently ignoring the assignment.
*   **Destructuring updates:**
    ```able
    Point := struct { x: i64, y: i64 }
    current := Point { x: 1, y: 2 }
    Point { x, y } = current
    x = x + 10           ## writes back into x's storage cell
    y = y + 20
    ```
    Because `{ x, y } = current` succeeds atomically, either both bindings are updated to refer to `current`'s fields or neither changes if the RHS fails to match the pattern shape.
*   **Loops and imperative code:**
    ```able
    count := 3
    while count > 0 {
      print(count)
      count = count - 1    ## `=` must mutate the existing binding
    }
    ```
    Each iteration observes the newly written value of `count` because `=` mutates the storage cell allocated by the initial `:=`.
*   **Fields and indexed locations:** Assignments such as `point.x = point.x + 1` or `arr[i] = arr[i] * scale` mutate the designated storage in place. Indexing relies on the `IndexMut` interface; map and array entries update without replacing the containing aggregate value.

`=` introduces new bindings when no such binding had existed in the lexical scope or the enclosing lexical scope. When a binding had alrady existed, then the `=` functions as a reassignment, in which the original binding is mutated. Use `:=` when you need to declare a new binding or shadow names; use `=` when existing bindings need to point at a different value or when introducing a new binding without having to think about whether a binding already existed.

## 6. Expressions

Units of code that evaluate to a value.

### 6.1. Literals

Literals are the source code representation of fixed values.

#### 6.1.1. Integer Literals

-   **Syntax:** A sequence of digits `0-9`. Underscores `_` can be included anywhere except the start/end for readability and are ignored. Prefixes `0x` (hex), `0o` (octal), `0b` (binary) are supported.
-   **Type:** By default, integer literals are inferred as `i32` (this default is configurable/TBC). Type suffixes can explicitly specify the type: `_i8`, `_u8`, `_i16`, `_u16`, `_i32`, `_u32`, `_i64`, `_u64`, `_i128`, `_u128`.
-   **Examples:** `123`, `0`, `1_000_000`, `42_i64`, `255_u8`, `0xff`, `0b1010_1111`, `0o777_i16`.

##### Literal Typing & Context

1.  **Default types:** Unsuffixed integer literals start life as `i32`. Float literals default to `f64` (see §6.1.2). These defaults apply only when no other information constrains the literal.
2.  **Explicit suffix wins:** Supplying a suffix (`42_u8`, `0xff_i64`) locks the literal to that type. Subsequent contexts must accept that exact type; otherwise compilation fails.
3.  **Contextual adoption:** When a literal appears in a context that requires a specific integer type (variable declaration with annotation, parameter type, struct field, return expression, typed pattern, etc.), the compiler attempts to adopt that type. Adoption succeeds if—and only if—the literal’s value fits within the target type’s representable range. Otherwise it is a compile-time error (`"literal 300 does not fit in u8"`).
    *   **Examples:**
        ```able
        small: u8 = 200        ## OK (fits)
        # too_big: u8 = 300    ## Error: literal does not fit in u8

        duration: i64 = 1_000_000_000_000  ## Literal adopts i64 from the annotation
        offset := 0         ## No annotation ⇒ defaults to i32 until constrained later
        ```
4.  **Flow into generics:** Inside generic functions/structs, unsuffixed literals remain “untyped integers” until constraints resolve them. At the point where a type parameter is bound (e.g., `T: Numeric` inferred as `i64`), pending literals adopt the resolved type and must fit.
5.  **Implicit widening contexts:** Assigning or passing an integer literal to a wider integer type (e.g., literal `5` to `i64`, literal `200` to `u32`) is always allowed as long as the literal fits. Able does **not** automatically narrow; when a narrower type is required, ensure the literal fits exactly or add an explicit conversion helper.
6.  **No hidden rounding:** Integer literals never auto-convert to floats unless the context explicitly requires a floating-point type (e.g., `f64`, `f32`). In such contexts, the literal adopts the requested float type with an exact conversion if representable; otherwise the value is rounded according to IEEE-754 rules.
7.  **Diagnostics:**
    -   Literal-too-wide for context ⇒ compile-time error pointing to the literal and the expected type.
    -   Literal used with no resolving context and exceeding the default type (e.g., `3_000_000_000` with default `i32`) ⇒ compile-time error unless a suffix or explicit context is supplied.
    -   Tooling SHOULD surface notes suggesting adding a suffix or annotation when inference stalls.

#### 6.1.2. Floating-Point Literals

-   **Syntax:** Include a decimal point (`.`) or use scientific notation (`e` or `E`). Underscores `_` are allowed for readability.
-   **Type:** By default, float literals are inferred as `f64`. The suffixes `_f32` and `_f64` explicitly denote the type.
-   **Examples:** `3.14`, `0.0`, `-123.456`, `1e10`, `6.022e23`, `2.718_f32`, `_1.618_`, `1_000.0`, `1.0_f64`.

#### 6.1.3. Boolean Literals

-   **Syntax:** `true`, `false`.
-   **Type:** `bool`.

#### 6.1.4. Character Literals

-   **Syntax:** A single Unicode scalar value (code point) enclosed in single quotes `'`. Special characters can be represented using escape sequences:
    *   Common escapes: `\n` (newline), `\r` (carriage return), `\t` (tab), `\\` (backslash), `\'` (single quote), `\"` (double quote - though not strictly needed in char literal).
    *   Unicode escape: `\u{XXXXXX}` where `XXXXXX` are 1-6 hexadecimal digits representing the Unicode scalar value.
-   **Type:** `char`.
-   **Validation:** After escape processing the literal must contain exactly one Unicode scalar value. Sequences that expand to multiple scalars are rejected at compile time.
-   **Examples:** `'a'`, `' '`, `'%'`, `'\n'`, `'\u{1F604}'`.

#### 6.1.5. String Literals

-   **Syntax:**
    1.  **Standard:** Sequence of characters enclosed in double quotes `"`. Supports the same escape sequences as character literals.
    2.  **Interpolated:** Sequence of characters enclosed in backticks `` ` ``. Can embed expressions using `${Expression}`. Escapes like `` \` `` and `\$` are used for literal backticks or dollar signs before braces. See Section [6.6](#66-string-interpolation).
-   **Type:** `String`. Strings are immutable.
-   **Examples:** `"Hello, world!\n"`, `""`, `` `User: ${user.name}, Age: ${user.age}` ``, `` `Literal: \` or \${` ``.

##### String Representation

-   `String` values represent immutable sequences of UTF-8 bytes.
-   Operations that inspect textual structure (code points, grapheme clusters, normalisation) are performed through library routines (`String.chars()`, `String.graphemes()`, `String.to_nfc()`, etc.) rather than implicit runtime behaviour.
-   The `char` type corresponds to a single Unicode code point value (`u32` range). A distinct `Grapheme` type in the standard library models user-perceived characters; it is derived from strings via segmentation helpers.
-   Unless specified otherwise, indices and spans refer to byte offsets within the UTF-8 sequence.
-   Required `String` helper methods are listed in §6.12.1.

#### 6.1.6. Nil Literal

-   **Syntax:** `nil`.
-   **Type:** `nil`. The type `nil` has only one value, also written `nil`.
-   **Usage:** Represents the absence of meaningful data. Often used with the `?Type` (equivalent to `nil | Type`) union shorthand. `nil` itself *only* has type `nil`, but can be assigned to variables of type `?SomeType`.

#### 6.1.7. Array Literals

-   **Syntax:** `[Expression, ...]` with optional trailing commas/newlines. Empty arrays use `[]`.
-   **Evaluation order:** Elements evaluate left-to-right and each value is stored into a fresh `Array` allocation. The literal expression's value is the fully populated array.
-   **Typing:** Every element must be assignable to a single element type `T`. Mixed literals form the least upper bound (e.g., `[1, 2.0]` infers `T = i32 | f64`). An empty literal requires contextual typing or an explicit annotation (e.g., `values: Array String = []`).
-   **Examples:**
    ```able
    evens := [0, 2, 4, 6]
    matrix := [
      [1, 0],
      [0, 1]
    ]
    ```

#### 6.1.8. Struct Literals

-   **Syntax:** `TypeName { field: value, ... }` for named structs; `TypeName { value1, value2, ... }` for positional structs (named tuples). Spread/functional update forms (`TypeName { ...existing, field: override }`) follow the struct semantics in §4.5.
-   **Evaluation order:** The struct expression evaluated after its arguments/fields evaluate left-to-right. Each literal produces a fresh heap allocation.
-   **Typing:** The literal's shape must match the struct definition. For named structs all required fields must be assigned exactly once. For positional structs the arity must match. Type inference flows from the struct definition.
-   **Examples:**
    ```able
    struct Point { x: f64, y: f64 }
    origin := Point { x: 0.0, y: 0.0 }

    struct Pair { String, i32 }
    result := Pair { "count", 42 }
    ```

#### 6.1.9. Map Literals

Able v11 introduces a dedicated literal form for hash-map values. Map literals construct the standard library `HashMap K V` (from `able.collections.hash_map`), which implements the `Map` interface.

-   **Syntax:** `#{ KeyExpr: ValueExpr, ... }`. Entries are comma-delimited; trailing commas and multiline formatting are permitted. Whitespace is insignificant outside expressions. The empty literal is written `#{}`.
-   **Type inference:** Literal entries must agree on a single key type `K` and value type `V`. The compiler infers `HashMap K V` from the entries when possible. If either dimension cannot be inferred (e.g., `#{}`), provide context or an explicit annotation (either `HashMap K V` or the `Map K V` interface):
    ```able
    counts: HashMap String u64 = #{}
    view: Map String u64 = counts
    data := #<String, String>{ "name": user.name, "id": user.id }
    ```
-   **Key constraints:** The key type must satisfy the same requirements imposed by the standard library `Map` implementation (typically `Hash + Eq`). Using a key expression whose type cannot act as a map key is a compile-time error.
-   **Evaluation order:** Entries evaluate left-to-right. Each `KeyExpr` evaluates exactly once; `ValueExpr` may observe side effects from prior insertions. The literal produces a fresh mutable map and inserts entries sequentially.
-   **Duplicate keys:** Later entries with the same key overwrite earlier ones, including entries introduced by spreads.
-   **Spread/updates:** Literal bodies may include `...Expression` clauses that merge the entries from an existing map value before processing subsequent key/value pairs:
    ```able
    merged := #{
      ...defaults,
      "timeout_ms": 5000,
      ...user_overrides
    }
    ```
    Each spread expression must evaluate to a `HashMap K V` (or another map implementation that can enumerate entries) compatible with the literal's inferred key/value types. Spreads execute in place: entries insert in the position of the spread, and later clauses can overwrite them.
-   **Runtime checks:** Attempting to spread a non-map value or inserting a key/value that does not conform to the inferred `Map K V` type raises an `Error` at runtime (or a compile-time diagnostic when statically provable).
-   **Examples:**
    ```able
    headers := #{
      "content-type": "application/json",
      "x-trace": request.trace_id
    }

    overrides := #{
      ...headers,
      "x-trace": new_trace_id,
      user.id: user.token
    }
    ```
    The second literal copies the entries from `headers`, replaces the `x-trace` key, and adds a user-specific entry. Because literals allocate new map instances, later mutation of `overrides` does not affect `headers`.

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
| 14         | `!` (postfix)         | Propagate `?T`/`!T`                     | Left-to-right | Applies to expressions like `arr[i]!`, `foo()!`, `a.b()!` |
| 13         | `^`                   | Exponentiation                          | Right-to-left | Binds tighter than unary `-`. `-x^2` == `-(x^2)`          |
| 12         | `-` (unary)           | Arithmetic Negation                     | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 12         | `!` (unary)           | **Logical NOT**                         | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 12         | `.~` (unary)          | **Bitwise NOT (Complement)**            | Non-assoc     | (Effectively Right-to-left in practice)                 |
| 11         | `*`, `/`, `//`, `%`, `/%` | Multiplication; float div; Euclidean int div; modulo; div-with-remainder | Left-to-right | `//`, `%`, and `/%` use Euclidean integer division. |
| 10         | `+`, `-` (binary)     | Addition, Subtraction                   | Left-to-right |                                                           |
| 9          | `.<<`, `.>>`          | Left Shift, Right Shift                 | Left-to-right |                                                           |
| 8          | `.&` (binary)         | Bitwise AND                             | Left-to-right |                                                           |
| 7          | `.^`                  | Bitwise XOR                             | Left-to-right |                                                           |
| 6          | `.|` (binary)         | Bitwise OR                              | Left-to-right |                                                           |
| 6          | `>`, `<`, `>=`, `<=`    | Ordering Comparisons                    | Non-assoc     | Chaining requires explicit grouping (`(a<b) && (b<c)`) |
| 5          | `==`, `!=`            | Equality, Inequality Comparisons        | Non-assoc     | Chaining requires explicit grouping                     |
| 4          | `&&`                  | Logical AND (short-circuiting)          | Left-to-right |                                                           |
| 3          | `||`                  | Logical OR (short-circuiting)           | Left-to-right |                                                           |
| 2          | `..`, `...`           | Range Creation (inclusive, exclusive)   | Non-assoc     |                                                           |
| 1          | `\|>`                 | Pipe Forward                            | Left-to-right | Binds tighter than assignment; looser than `||`/`&&`      |
| 0          | `:=`                  | **Declaration and Initialization**      | Right-to-left | See Section [5.1](#51-operators---)                     |
| 0          | `=`                   | **Reassignment / Mutation**             | Right-to-left | See Section [5.1](#51-operators---)                     |
| 0          | `+=`, `-=`, `*=`, `/=`, `%=`, `.&=`, `.|=`, `.^=`, `.<<=`, `.>>=` | Compound Assignment                      | Right-to-left | Desugars to `a = a OP b` (no `^=`).                       |
| -1         | `\|>>`                | Low-Precedence Pipe Forward             | Left-to-right | Binds looser than assignment (lowest)                    |

*(Note: Precedence levels are relative; specific numerical values may vary but the order shown is based on Rust.)*

#### 6.3.2. Operator Semantics

*   **Arithmetic (`+`, `-`, `*`, `/`, `//`, `%`, `/%`, `^`):**
    *   Numeric promotion (P1) for `+`, `-`, `*`, `/`:
        1.  **Classify operands.** Each operand is either an integer (signed/unsigned, width 8–128) or a float (`f32`, `f64`). Literals adopt operand types per §6.1.1 before this step.
        2.  **Floating precedence (`/` only):** If either operand is floating-point, both convert to the wider float (if any side is `f64`, promote both to `f64`; otherwise promote to `f32`). The expression type is that float.
        3.  **Integer-only rules (`+`, `-`, `*`, `/` when both operands are ints and `Ratio` is not involved):**
            -   Identical signedness ⇒ promote both to the wider width while keeping signedness (e.g., `i16 + i64` → `i64`).
            -   Mixed signed/unsigned ⇒ promote both to a signed type whose width can represent both ranges. Compute `needed_bits = max(width(lhs)+1, width(rhs)+1)` and choose the smallest signed type ≥ `needed_bits` (preferring `i16`, `i32`, `i64`, `i128`). If no built-in width suffices, emit a compile-time error requesting an explicit big-int implementation or narrowing cast.
            -   When the unsigned operand already subsumes the signed range (e.g., `u128` with `i32`), the compiler may keep the unsigned width to avoid unnecessary sign changes. Otherwise the signed rule above applies.
        4.  **Result type:** Arithmetic expressions evaluate to the promoted type. Assigning the result to an even wider type is allowed; narrowing to a smaller type requires an explicit conversion helper.
        5.  **Examples:**
            ```able
            total: i64 = 0
            total = total + 1      ## stays i64

            mix = 5_i8 + 10_u16    ## promotes to i32 (needs 17 bits)
            wide = 1_u64 + (-1_i32) ## promotes to i65 ⇒ implemented as i128

            precise = 1_000_000 + 0.5_f64  ## float wins ⇒ f64
            ```
    *   **Division family:**
        -   `/` (float division): follows promotion rules above; integer / integer yields `f64`. Division by zero **raises** `DivisionByZeroError`.
        -   `//` (Euclidean integer division): integers only. Defines `q` and `r` such that `a = b * q + r` and `0 <= r < |b|` when `b != 0`. If `b > 0`, this matches `q = floor(a / b)`; if `b < 0`, it matches `q = ceil(a / b)`. Zero divisor **raises** `DivisionByZeroError`. Examples: `5 // 3 = 1`; `-5 // 3 = -2`; `5 // -3 = -1`; `-5 // -3 = 2`.
        -   `%` (Euclidean remainder): integers only. `r` is the non-negative remainder from the Euclidean pair above. Zero divisor **raises** `DivisionByZeroError`. Examples: `5 % 3 = 2`; `-5 % 3 = 1`; `5 % -3 = 2`; `-5 % -3 = 1`.
        -   `/%` (Euclidean div-with-remainder): integers only. Returns `DivMod<T> { quotient: T, remainder: T }` where `T` matches the operand integer type. Uses the same `q`/`r` as `//`/`%`; zero divisor **raises** `DivisionByZeroError`.
    *   **Exponentiation (`^`):**
        -   Floats: follows IEEE-754; negative exponents allowed; `NaN`/`Inf` propagate per host rules.
        -   Integers: defined for non-negative exponents; negative exponents on integers are a runtime error. Overflow raises `OverflowError`.
    *   **Ratio-aware operations:** The standard library provides a `Ratio` type (`num: i64`, `den: i64`, `den > 0`, gcd-reduced) plus `to_r()` conversions on integers/floats. Arithmetic on `Ratio` mirrors `+`, `-`, `*`, `/` with exact rational results; mixed Ratio/primitive operations follow the Ratio implementation rules (see §6.12.3).
    *   Integer overflow (O1):
        -   Checked by default. On overflow in `+`, `-`, `*`, raises a runtime exception `OverflowError { message: "integer overflow" }`.
        -   Division and remainder by zero already raise `DivisionByZeroError`.
        -   Library provides explicit alternatives (names TBD): `wrapping_add/sub/mul`, `saturating_add/sub/mul`, `checked_add/sub/mul -> ?T` for performance-critical or specific semantics.
*   **Comparison (`>`, `<`, `>=`, `<=`, `==`, `!=`):** Compare values, result `bool`. Equality/ordering behavior relies on standard library interfaces (`PartialEq`, `Eq`, `PartialOrd`, `Ord`). See Section [14](#14-standard-library-interfaces-conceptual--tbd).
*   **Logical (`&&`, `||`, `!`):**
    *   Truthiness-based semantics (see §6.11). All operands are accepted; values are interpreted for truthiness.
    *   `a && b` (AND):
        -   Evaluate `a`. If `a` is falsy, result is `a` (no evaluation of `b`).
        -   Otherwise, evaluate `b` and return `b`.
        -   Type: union of operand types after flow; commonly `TypeOf(a) | TypeOf(b)`.
    *   `a || b` (OR):
        -   Evaluate `a`. If `a` is truthy, result is `a` (no evaluation of `b`).
        -   Otherwise, evaluate `b` and return `b`.
        -   Type: union of operand types after flow; commonly `TypeOf(a) | TypeOf(b)`.
    *   `!x` (NOT): Returns a `bool` equal to `false` iff `x` is truthy, `true` iff `x` is falsy.
    *   Object identity is preserved for returned operands (no coercion); only `!` yields a `bool`.
*   **Bitwise (`.&`, `.|`, `.^`, `.<<`, `.>>`, `.~`):**
    *   `.&`, `.|`, `.^`: Standard bitwise AND, OR, XOR on integer types (`i*`, `u*`).
    *   `.<<`, `.>>` (Shift semantics, S1):
        -   Shift count must be in range `0..bits` for the left operand's type. Out-of-range shift counts raise `ShiftOutOfRangeError { message: "shift out of range" }`.
        -   Right shift of signed integers is arithmetic (sign-extending), matching Go semantics; right shift of unsigned integers is logical (zero-filling).
    *   `.~` (Bitwise NOT): Unary operator, performs bitwise complement on integer types.
*   **Unary (`-`):** Arithmetic negation for numeric types.
*   **Member Access (`.`):** Access fields/methods, UFCS, static methods. See Section [9.4](#94-method-call-syntax-resolution).
*   **Function Call (`()`):** Invokes functions/methods. See Section [7.4](#74-function-invocation).
*   **Indexing (`[]`):** Access elements within indexable collections (e.g., `Array`). Relies on standard library interfaces (`Index`, `IndexMut`). See Section [14](#14-standard-library-interfaces-conceptual--tbd).
*   **Range (`..`, `...`):** Create `Range` objects (inclusive `..`, exclusive `...`). See Section [8.2.3](#823-range-expressions).
*   **Declaration (`:=`):** Declares/initializes new variables. Evaluates to RHS. See Section [5.1](#51-operators---).
*   **Assignment (`=`):** Reassigns existing variables or mutates locations. Evaluates to RHS. See Section [5.1](#51-operators---).
*   **Compound Assignment (`+=`, etc.):** Shorthand (e.g., `a += b` is like `a = a + b`). Acts like `=`.
*   **Pipe Forward (`|>`):** Binds tighter than assignment, looser than `||`/`&&`.
    - Evaluate the LHS (`subject`), then evaluate the RHS once to a callable value.
    - Invoke that callable with `subject` prepended as the first argument, followed by any explicit RHS arguments (if the RHS is itself a partially applied callable).
    - If the RHS is not callable, this is an error. If the callable expects more args than supplied, the result is a partial application; if too many are supplied, it is an error.
    - Placeholders (`@`, `@n`) remain available to build RHS callables concisely; bare `@` is equivalent to `@1`. The RHS can be any expression yielding a callable—parentheses are not required.
    - Examples: `x |> (@ + 1)` applies the unary lambda to `x`; `5 |> add` yields a partial expecting the second argument; `5 |> @ * @ |> @ + @` applies successive placeholder lambdas and yields `50`.
*   **Low-precedence Pipe (`|>>`)**: Identical semantics to `|>` but with lower precedence than assignment. Useful for post-processing an assignment’s value without extra parentheses:
    - `a = 5 |>> print` parses as `(a = 5) |>> print`, performs the assignment, then pipes the resulting value to `print`.

Compound assignment semantics:
*   Supported forms: `+=`, `-=`, `*=`, `/=`, `%=`, `.&=`, `.|=`, `.^=`, `.<<=`, `.>>=`. The exponent form `^=` is not supported.
*   Desugaring: `a OP= b` is defined as `a = a OP b`, where `OP` is the corresponding binary operator.
*   Single evaluation (C1): `a` is evaluated exactly once. The compiler lowers to an assignment to the same storage location without re-evaluating addressable subexpressions on the LHS.
    -   Example: `arr[i()] += f()` evaluates `i()` once, not twice.
*   Types: Assignment follows the same rules as plain `=` of the desugared form; the target of `a` must be assignable from the result type of `a OP b`.

Precedence examples:
```able
a = x |> (@ + 1)      ## parses as a = (x |> (@ + 1))
a := x |> (@ + 1)     ## same, with declaration
flag = cond || x |> (@ + 1)   ## parses as flag = (cond || (x |> (@ + 1))) because |> binds looser than ||
a = 5 |>> print    ## parses as (a = 5) first, then pipes the resulting value to print
```

#### 6.3.3. Overloading (Via Interfaces)

Behavior for non-primitive types relies on implementing standard library interfaces (e.g., `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg` (for unary `-`), `Not` (for bitwise `.~`), `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`, `PartialEq`, `Eq`, `PartialOrd`, `Ord`, `Index`, `IndexMut`). These interfaces need definition (See Section [14](#14-standard-library-interfaces-conceptual--tbd)). Note that logical `!` is not overloaded.

> **Operator participation rule:** User code cannot define new operators or overload existing ones directly. The only way for a custom type (e.g., a struct) to participate in an operator is to implement the corresponding standard interface (`Add`, `Eq`, `Index`, etc.). If the interface is not implemented (or not in scope), the operator is unavailable for that type. This keeps the grammar fixed and makes operator behavior explicitly tied to the shared interface contracts.

Operator-to-interface mapping (when operands are not primitives):

- `+` → `Add`
- `-` (binary) → `Sub`; `-` (unary) → `Neg`
- `*` → `Mul`; `/` → `Div`; `%` → `Rem`
- `.&` → `BitAnd`; `.|` → `BitOr`; `.^` → `BitXor`
- `.<<` → `Shl`; `.>>` → `Shr`; `.~` → `Not` (bitwise complement)
- `==`, `!=` → `PartialEq`/`Eq`
- `<`, `>`, `<=`, `>=` → `PartialOrd`/`Ord`
- `[]` (indexing) → `Index`; `[]=` (mutation) → `IndexMut`

These operations are available only when a single applicable implementation is visible in scope per the import-scoped model.

#### 6.3.4. Safe Navigation Operator (`?.`)

`?.` short-circuits member access and method calls when the receiver is `nil`. It is primarily used to traverse optional values without repeatedly writing `match`/`if` guards.

*   **Syntax:**
    ```able
    receiver?.field
    receiver?.method(args...)
    receiver?.field?.nested_field   ## chaining is allowed
    ```
    `receiver` may be any expression. `field`/`method` resolution follows the same rules as normal `.` access (§6.3, §7.4). Safe navigation does not currently apply to indexing; use helper methods (e.g., `arr?.get(i)`) for containers.
*   **Evaluation order:**
    1. Evaluate `receiver` exactly once.
    2. If it evaluates to `nil`, the entire `?.` expression evaluates to `nil` immediately; subsequent field/method lookups and argument expressions are skipped.
    3. If it is non-`nil`, the operator behaves like standard `.` access: resolve the member/method, then (for calls) evaluate arguments left-to-right and invoke the method.
*   **Typing:**
    -   Let the underlying access (if performed via plain `.`) have type `T`. Then `receiver?.member` has type `?T` (i.e., `nil | T`).
    -   The usual union simplification rules apply (`nil | ?U` collapses to `?U`).
    -   When chaining (`a?.b?.c`), each step adopts the `?`-wrapped result of the previous step, so the final type reflects the deepest access that actually executes.
*   **Method calls:** `receiver?.method(args...)` returns `?R`, where `R` is the method’s return type. Arguments are evaluated only when the receiver is non-`nil`.
*   **Side effects:** Because the receiver evaluates even when it is `nil`, any side effects in the receiver expression still occur. Side effects in arguments execute only when the receiver is non-`nil`.
*   **Diagnostics:** Using `?.` on a value whose type statically excludes `nil` is permitted (it simply behaves like `.`), but linters may warn about redundant usage. Applying `?.` to something that cannot be treated as a member access (e.g., literals without fields) is a compile-time error.

Examples:

```able
user: ?User = find_user(id)
name = user?.profile?.display_name ?? "Guest"   ## ?? is a hypothetical coalesce helper

timeout_ms = config?.network?.timeouts?.connect ?? 1_000

fn maybe_length(s: ?String) -> ?u64 {
  s?.len()
}

log(user?.session()?.token ?? "no session")
```

Each expression above returns `nil` if any receiver in the chain is `nil`; otherwise it produces the same value as the corresponding `.` access. The helpers `??`/`len()` are shown for illustration; they follow ordinary lookup rules once the safe-navigation step succeeds.

### 6.4. Function Calls

See Section [7.4](#74-function-invocation).

### 6.5. Control Flow Expressions

`if/elsif/else`, `match`, `breakpoint`, `rescue`, `do`, `:=`, `=` evaluate to values. See Section [8](#8-control-flow) and Section [11](#11-error-handling). Loops (`while`, `for`) evaluate to `void`.

Assignment/Declaration results: Both `=` and `:=` evaluate to the RHS value on successful matching/binding. If the pattern fails to match, the expression evaluates to an `Error` value (see §5.3).

### 6.6. String Interpolation

`` `Literal text ${Expression} more text` ``

*   Evaluates embedded `Expression`s (converting via `Display` interface).
*   Concatenates parts into a final `String`.
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

Note: `IteratorEnd` is a singleton struct used to signal end of iteration; see Section 14 (Core Iteration Protocol) for its definition.

```able
gen.yield(value: T) -> void  ## Yield a value and suspend until next() is called again
gen.stop() -> void           ## Terminate the generator early (subsequent next() => IteratorEnd)
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
fn evens_to_strings(xs: Iterable i32) -> (Iterator String) {
  Iterator String { gen =>
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

Indexing (raises on out-of-bounds):
```able
x = arr[i]                 ## Read element (may raise IndexError if i out of bounds)
arr[i] = v                 ## Write element (may raise IndexError)
```

Safe access:
```able
arr.get(i)    -> ?T        ## nil if out of bounds
arr.set(i, v) -> !nil      ## Error if out of bounds
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

-   Bounds: `arr[i]` and `arr[i] = v` surface `IndexError` via the `Index`/`IndexMut` result (`!`) on out-of-bounds. `get`/`pop` return `?T`.
-   Growth: Amortized doubling
-   Equality/Hashing: Derive from elementwise comparisons if `T: Eq/Hash` (via interfaces in Section 14).
-   See §6.12.2 for the required standard library helper methods (size/push/pop/etc.).

#### Examples

```able
arr: Array i32 = []
arr.push(1)
arr.push(2)
arr[0] = 3
first = arr.get(0) or { 0 }
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
dyn.package(fully_qualified_name: String) -> !dyn.Package

## Define a new dynamic package; returns its package object
dyn.def_package(fully_qualified_name: String) -> !dyn.Package

methods dyn.Package {
  ## Define declarations inside this package's namespace using Able source text (interpreted).
  ## Valid constructs: interfaces, impls, package-level functions, structs, unions, methods.
  fn def(self: Self, code: String) -> !nil
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

-   The interpreter is part of the standard Able runtime when dynamic features are used; dynamic facilities are available whenever the interpreter is present.
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
p2.def(`struct Point { name: String, x: i32, y: i32 }`)!
dynimport foo.{Point}
pt := Point { name: "blah", x: 12, y: 82 }
```

This facility deliberately avoids static type guarantees; it offers a clear, explicit bridge between compiled and interpreted worlds via `dynimport`, dynamic package objects, and interface adapters.


### 6.11. Truthiness and Boolean Contexts

Able adopts Ruby-like truthiness with explicit integration into the type system and error model.

1. Boolean contexts are expressions expecting a `bool` (e.g., `if`/`or` conditions, `while` conditions, guards in `match`/`rescue`). Any value used in a boolean context is first interpreted for truthiness.
2. Truthiness rules:
   - Falsy values: `false`, `nil`, any value implementing the `Error` interface.
   - Truthy values: all other values, including the singleton `void` value.
3. Interface participation: Explicit interface conversion ("Booleanish") is not required; the above rules apply uniformly. Libraries may provide an opt-in `ToBool` interface and adapter functions, but the core language defines truthiness by value kind, not by an interface.
4. Unions in boolean contexts:
   - `?T` (`nil | T`): considered falsy iff the value is `nil`; otherwise truthy.
   - `!T` (`Error | T`): considered falsy iff the value is an `Error`; otherwise truthy.
   - General unions: falsy iff the runtime value is `nil`, `false`, or an `Error`.
5. Short-circuiting operators:
   - `&&` and `||` operate on truthiness and return one of their operands (see §6.3.2).
   - `!` returns a `bool` based on truthiness.
6. Examples:
```able
## Option
if maybe_user { login(maybe_user!) } or { show_login() }

## Result
if data = load() { render(data) } or { |e| log(e.message()) }

## Assignment success/failure (see §5.3)
if { x, y } = try_pair() { use(x, y) } or { handle_fail() }

## void is truthy
if do { log("side-effect"); void } { ok() } or { unreachable() }
```

Note: Empty collections (e.g., `Array T` with size 0) are truthy. Only `false`, `nil`, and any value implementing `Error` are falsy.

### 6.12. Standard Library API (Required)

#### 6.12.0. Kernel Library (Host Built-ins)

The interpreters ship a **minimal** “kernel” library implemented in the host runtime. Higher-level helpers described in later subsections are written in Able and layer on top of this kernel surface. The kernel exists so the stdlib can bootstrap itself without depending on host-only functionality.

The kernel exposes only the following:
- **Global functions:** `print`, `proc_yield`, `proc_cancelled`, `proc_flush`, `proc_pending_tasks`.
- **Concurrency bridges:** channel primitives `__able_channel_new/send/receive/try_send/try_receive/await_try_send/await_try_recv/close/is_closed`; mutex primitives `__able_mutex_new/lock/unlock`; await waker helpers.
- **String/char bridges:** `__able_string_from_builtin`, `__able_string_to_builtin`, `__able_char_from_codepoint` (and UTF-8 validation/byte iterators as needed).
- **Hasher bridges:** `__able_hasher_create`, `__able_hasher_write`, `__able_hasher_finish`.
- **Array buffer hooks:** host-level allocation and slot access functions (e.g., `__able_array_new/with_capacity`, `__able_array_read`, `__able_array_write`, `__able_array_grow`). These are not user-facing and exist solely so the stdlib `Array` implementation can manage storage.
- **Error methods:** `message() -> String`, `cause() -> ?error`; `value` field is accessible for payloads.
- **Iterator methods:** `next() -> T | IteratorEnd`, `close()`.
- **Proc/Future methods:** `status()`, `value()`, `cancel()` (future cancel may be a no-op depending on target runtime).

All user-facing array and string helpers in §§6.12.1–6.12.2 live in the Able stdlib. Some runtimes currently ship temporary native shims for these helpers; they are **not** part of the kernel contract and will be removed once the stdlib owns the behaviour.

#### 6.12.1. String & Grapheme Helpers

`String` is the canonical immutable UTF-8 container provided by the language. String literals evaluate to `String` directly. Runtimes may still expose helper functions (e.g., `string_from_builtin`, `string_to_builtin`) to convert to/from host-native encodings, but Able programs always operate on the built-in `String` type. Each `String` owns an `Array u8` buffer; mutation happens only through builders (e.g., `StringBuilder`) that emit a new canonical value.

**Required operations**

| Method | Signature | Description |
| --- | --- | --- |
| `len_bytes()` | `fn len_bytes(self: Self) -> u64` | Delegates to `bytes().size()`; returns the number of UTF-8 bytes in the string. |
| `len_chars()` | `fn len_chars(self: Self) -> u64` | Delegates to `chars().size()`; counts Unicode code points. |
| `len_graphemes()` | `fn len_graphemes(self: Self) -> u64` | Delegates to `graphemes().size()`; counts user-perceived characters. |
| `substring(start, length?)` | `fn substring(self: Self, start: u64, length: ?u64 = nil) -> !String` | `start`/`length` are expressed in code points. The runtime converts them to byte offsets and raises `RangeError` when indices are out of bounds or would split a code point. |
| `split(delimiter)` | `fn split(self: Self, delimiter: String) -> Array String` | Splits by the delimiter (matched by code points). An empty delimiter splits into grapheme clusters. |
| `replace(old, new)` | `fn replace(self: Self, old: String, new: String) -> String` | Returns a new `String` with all non-overlapping occurrences of `old` replaced by `new`. |
| `starts_with(prefix)` | `fn starts_with(self: Self, prefix: String) -> bool` | Tests byte-prefix equality (`prefix.len_bytes()` must fit wholly inside the receiver). |
| `ends_with(suffix)` | `fn ends_with(self: Self, suffix: String) -> bool` | Tests byte-suffix equality. |

Implementations may surface richer helpers (`to_lower`, `to_upper`, normalization utilities, builders, etc.), but the methods above are the portable baseline expected by the v11 spec.

**Iteration & views**

`String` must expose explicit iterators so callers can choose the granularity they need:

| Helper | Signature | Description |
| --- | --- | --- |
| `bytes()` | `fn bytes(self: Self) -> (Iterator u8)` | Iterates raw UTF-8 bytes from left to right without allocating (direct view over the internal `Array u8`). |
| `chars()` | `fn chars(self: Self) -> (Iterator char)` | Yields Unicode code points. Invalid sequences raise `StringEncodingError` (implements `Error`). |
| `graphemes()` | `fn graphemes(self: Self) -> (Iterator Grapheme)` | Yields extended grapheme clusters computed from scalar values. |
| `Grapheme.len_bytes()` | `fn len_bytes(self: Self) -> u64` | Returns the byte width of the cluster. |
| `Grapheme.bytes()` | `fn bytes(self: Self) -> Array u8` | Clones the underlying byte span for low-level processing. |
| `Grapheme.as_string()` | `fn as_string(self: Self) -> String` | Materialises the grapheme as its own `String`. |

The iterator types (`StringBytesIter`, `StringCharsIter`, `StringGraphemesIter`) implement `Iterator` so they compose with the rest of the collection API (`map`, `take`, etc.). Each iterator also implements `fn size(self: Self) -> u64` (returning the number of elements remaining); the `len_*` helpers are defined in terms of these `size()` calls, ensuring a single source of truth for length calculations. Grapheme segmentation adopts the Unicode rules; runtimes MAY ship simplified tables as long as they agree across interpreters.

**Byte-indexing rules**

- Parameters ending in `_bytes` (e.g., `len_bytes`, any future `slice_bytes` helpers) interpret offsets as UTF-8 byte indices. Callers are responsible for providing well-formed boundaries.
- Parameters that omit the suffix (e.g., `substring`, `split`, `replace`) interpret offsets/counts as Unicode scalar values unless documented otherwise. Implementations must convert these counts to byte offsets and raise `RangeError` when a boundary would bisect a code point.
- Grapheme-level methods (`len_graphemes`, `graphemes()`) derive their counts from the iterator and therefore reflect user-perceived characters. Regex, formatting, and diagnostics may opt into grapheme-aware semantics by explicitly using these APIs (§14.2).
- Because the iterator `size()` implementations are authoritative, any optimization (e.g., cached byte length) must keep iterator state in sync so that `bytes().size()`, `chars().size()`, and `graphemes().size()` always report the same numbers as `len_bytes`, `len_chars`, and `len_graphemes` respectively.

**Errors and builders**

Invalid UTF-8 detected during decoding or iteration produces `StringEncodingError`, an `Error` with the offending byte offset. Indexing mistakes (negative offsets, out-of-bounds slices, or attempts to split a code point) raise `RangeError` (or a more specific subtype such as `StringIndexError`). Because `String` values are immutable, construction and concatenation flow through `StringBuilder`, which offers `push_char`, `push_bytes`, `push_string`, and `finish() -> Result String`.

Implementation note: these helpers live in the Able stdlib built atop the kernel string/char bridges. Any temporary native implementations in a runtime are compatibility shims and not part of the kernel contract.

#### 6.12.2. Array Helpers

`Array T` values expose the following minimum API (all methods mutate the receiver unless noted):

| Method | Signature | Description |
| --- | --- | --- |
| `size()` | `fn size(self: Self) -> u64` | Returns the current element count. |
| `push(value)` | `fn push(self: Self, value: T) -> void` | Appends `value`, growing the buffer as needed. |
| `pop()` | `fn pop(self: Self) -> ?T` | Removes and returns the last element, or `nil` if empty. |
| `get(index)` | `fn get(self: Self, index: u64) -> ?T` | Returns the element at `index` or `nil` if out of bounds. |
| `set(index, value)` | `fn set(self: Self, index: u64, value: T) -> !nil` | Writes to `index`; returns `IndexError` (implements `Error`) when out of bounds. |
| `clear()` | `fn clear(self: Self) -> void` | Removes all elements (capacity may be retained). |

Implementations must raise `IndexError` when out-of-range `set`/`push`/`pop` operations cannot be satisfied. Arrays must continue to implement `Iterable T` so `for` loops work uniformly.

Implementation note: these helpers live in the Able stdlib and are backed by kernel array buffer hooks for allocation and slot access. Any existing native implementations in a runtime are transitional and not part of the kernel contract.

#### 6.12.3. Numeric Helpers (Ratio, DivMod)

**`Ratio`**

-   Struct: `struct Ratio { num: i64, den: i64 }` with invariants:
    -   `den > 0`.
    -   `(num, den)` always reduced by `gcd(|num|, den)`.
    -   Sign carried by `num` (negative ratios have `num < 0`).
    -   Construction with `den == 0` is invalid and must raise `DivisionByZeroError`.
-   Conversions:
    -   `i*.to_r() -> Ratio` (exact; overflow to `i64` is an error).
    -   `f*.to_r() -> !Ratio` (exact binary expansion; rejects `NaN`/`Inf`/overflow).
-   Arithmetic: `+`, `-`, `*`, `/` defined on `Ratio` (and mixed Ratio/int/float where implemented) return reduced `Ratio` or raise `DivisionByZeroError` when dividing by zero.
-   Equality/ordering: defined via cross-multiplication on reduced forms.

**`DivMod`**

-   Built-in generic struct surfaced by the prelude: `struct DivMod T { quotient: T, remainder: T }`.
-   Produced by the `/%` operator on integer operands. `quotient`/`remainder` use the Euclidean pair `a = b*q + r` with `0 <= r < |b|` (so `q` is floor-based when `b > 0` and ceil-based when `b < 0`).
-   Division by zero raises `DivisionByZeroError`.

**Flooring helpers**

-   Stdlib functions `div_floor`, `mod_floor`, and `div_mod_floor` are provided for integer types and forward to the matching integer methods.
-   Semantics follow floor division: `q = floor(a / b)` and `r = a - b*q`. For `b > 0`, this matches Euclidean `//`/`%`; for `b < 0`, the remainder is negative or zero.
-   Each integer type also exposes methods `div_floor`, `mod_floor`, and `div_mod_floor` with the same semantics and error behavior (division by zero raises `DivisionByZeroError`).

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
-   **`[-> ReturnType]`**: Optional return type annotation. If omitted, the return type is inferred from the body's final expression or explicit `return` statements. If the body's last expression evaluates to `void` (e.g., a loop or a function that performs only side effects) and there's no explicit `return`, the return type is `void`. If the function is intended to return no data, use `-> void`. When the declared return type is `void`, the value of the last expression (or any explicit `return` expression) is ignored and the function is considered to return `void`.
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
  print(message) ## Assuming print returns void
  return ## Explicit return void
}

## Function with side effects and inferred void return type
fn log_and_nil(name: String) { ## Implicitly returns void
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

#### 7.1.4. Generic Argument Inference and Annotation Rules

- Generic parameters in `fn` definitions may be omitted at call sites; the compiler infers them from argument types and, when needed, the expected return type at the call site.
- Annotations in value positions must be concrete. To accept values of a polymorphic family (e.g., any `Array T`), introduce a generic parameter and use it in the annotation: `fn f<T>(xs: Array T) { ... }`.
- When inference is insufficient or ambiguous, provide explicit generics: `identity<i64>(0)`.
- It is a compile-time error to annotate parameters, locals, or fields with unbound type constructors (e.g., `Array`, `Map String _`).

#### 7.1.5. Optional Generic Parameter Declarations

Able infers function-level type parameters when the signature uses free type names that are not already defined in the surrounding scope. This allows concise definitions without `<T, U>` headers when the type variables are obvious from context.

**Detection rules:**

1.  Collect every type identifier used in the parameter list, return type, and `where` clause.
2.  Remove identifiers that refer to known types in scope (structs, aliases, interfaces, type parameters from enclosing scopes, etc.).
3.  Any remaining identifiers are treated as inferred generic parameters. The compiler synthesizes a `<...>` list in lexical order of first appearance and applies the same constraint rules as if the parameters had been declared explicitly.

**Constraints:**

- Inline constraints (`T: Display`) or `where` clauses referencing inferred parameters remain valid. The compiler hoists those constraints onto the synthesized parameter list.
- If the body introduces a conflicting declaration for an inferred name, it is a compile-time error (“cannot redeclare inferred type parameter `T`”).
- Ambiguity arises when a name is both a known type and appears as a would-be parameter; developers must disambiguate by explicitly declaring `<T>` (for generics) or fully qualifying the known type (e.g., `pkg.Type`).

**Examples:**

```able
## Explicit generics
fn choose_first<T, U>(first: T, second: U) -> T { first }

## Equivalent implicit form
fn choose_first(first: T, second: U) -> T { first }

## Constraints still work
fn render(item: T) -> String where T: Display { item.to_string() }

## Mixed explicit + implicit: OK but redundant; prefer one style
fn map<T>(value: T, map_fn: (T) -> U) -> Array T { [ map_fn(value) ] }
```

**Diagnostics:**

- If the compiler cannot infer all type arguments at call sites, it reports “cannot infer type for parameter `<T>`” just as if `<T>` had been written explicitly.
- If an inferred name collides with a type imported into the same scope, the spec requires the author to disambiguate via explicit `<>` syntax. Implementations should point to both declarations in the error message.

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
return ## Equivalent to 'return void' if function returns void
```

#### Semantics
-   Immediately terminates the execution of the current function.
-   If the surrounding function's return type is `void`, `return` with no expression is equivalent to `return void`; any expression used with `return` is coerced to `void` before the function completes.
-   The value of `Expression` (or `nil`) is returned to the caller.
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

- If a call supplies **fewer arguments than the callable’s arity**, it produces a *partially applied* callable that captures the provided arguments and expects the remaining ones. Supplying more arguments than the callable accepts is an error.
- If the **final parameter is declared nullable** (`?T`), the call may omit that argument; `nil` is supplied implicitly. This applies to free functions and methods (including shorthand forms) but only to the last parameter.
  ```able
  fn log(message: String, tag: ?String) -> void { ... }
  log("hi")         ## equivalent to log("hi", nil)
  log("hi", "info") ## explicit tag
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

Allows calling functions using dot notation on the first argument. In Able, "methods" are ordinary functions; the `methods` block syntax is a convenience for defining functions whose first parameter is the receiver type. (See Section [9.4](#94-method-call-syntax-resolution) for the full resolution rules and import scoping.)

##### Syntax
```able
ReceiverExpression . FunctionOrMethodName ( RemainingArgumentList )
```

##### Semantics (Simplified)
When `receiver.name(args...)` is encountered:
1.  If `receiver` has a callable field `name`, invoke it.
2.  Otherwise, resolve `name` against the unified callable pool consisting only of callable *names that are actually in scope* whose first parameter is compatible with the receiver (inherent functions from `methods` blocks, interface methods, and free/UFCS functions). Pick the single most specific candidate.
3.  Invoke the chosen callable with `receiver` prepended as the first argument.
4.  Ambiguity or no match results in an error.

Notes:
- Functions defined inside `methods Type { ... }` whose first parameter is `Self` are exported under their unqualified name (e.g., `foo`) and are eligible for method-call sytactic sugar and UFCS.
- Functions in `methods` blocks whose first parameter is **not** `Self` are *type-qualified functions*; they export as `Type.foo` (Type = the base nominal name) and are called only with type qualification (`Type.foo<T>(...)`), never via receiver syntax.
- Method-call syntax is sugar for calling a function with the receiver as the first argument; there are no special “method objects” beyond the callable values themselves.

##### Example (Method Call Syntax on Free Function)
```able
fn add(a: i32, b: i32) -> i32 { a + b }
res = 4.add(5) ## Resolved via Method Call Syntax to add(4, 5) -> 9

## Inherent instance methods are also callable via UFCS:
methods Point {
  fn #norm() -> f64 { (#x * #x + #y * #y).sqrt() }
}
point = Point { x: 3.0, y: 4.0 }
dist = point.norm()  ## Calls Point.norm via inherent method step (preferred)
dist2 = norm(point) ## UFCS-style invocation of the same method
```

#### 7.4.4. Callable Value Invocation (`Apply` Interface)

If `value` implements the `Apply` interface, `value(args...)` desugars to `value.apply(args...)`. (See Section [14](#14-standard-library-interfaces-conceptual--tbd)). All ordinary function values and closures produced by placeholder lambdas implement `Apply` implicitly with their natural arity; user-defined types may implement `Apply` to become callable.
```able
## Conceptual Example
# impl Apply for Integer { fn apply(self: Integer, a: Integer) -> Integer { self * a } }
# thirty = 5(6) ## Calls 5.apply(6)
```

### 7.5. Partial Function Application

Create a new function by providing some arguments and using `@` as a placeholder for others. Numbered placeholders `@1`, `@2`, ... refer to specific parameter positions; bare `@` is equivalent to `@1`, and repeated unnumbered placeholders all reference the first argument. The arity is the maximum placeholder index present (at least 1 whenever a placeholder appears).

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

## Assuming prepend exists: fn prepend(prefix: String, body: String) -> String
# prefix_hello = prepend("Hello, ", @) ## Function expects one arg
# msg = prefix_hello("World")          ## msg is "Hello, World"

## method call syntax access creates partially applied function
add_five = 5.add ## Creates function add(5, @) via Method Call Syntax access
result_pa = add_five(20)  ## result_pa is 25
```

Note: Expression placeholders use `@`/`@n` only. The underscore `_` is not an expression placeholder; it is reserved for wildcard patterns (§5.2.2) and unbound type parameters in type expressions (§4.4).

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
struct Data { value: i32, name: String }
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

#### 7.6.3. Placeholder Lambdas (`@`, `@n`)

Placeholders in expression positions create anonymous functions.

*   Bare `@` is equivalent to `@1` (the first argument).
    *   `@ + 1` → `{ x => x + 1 }`
    *   `@ * @` → `{ x => x * x }` (duplicates the first argument)
*   Numbered placeholders refer to specific parameter positions; repeats reuse the same parameter:
    *   `@1 + @1` → `{ x => x + x }`
*   Mixing is allowed; arity is the maximum placeholder index used:
    *   `f(@1, @2, @3)` → `{ x, y, z => f(x, y, z) }`
*   Scope: The smallest enclosing expression that expects a function determines the lambda boundary. If a placeholder spans a whole block, the entire block becomes the lambda body. Parentheses may be used for clarity without changing scope.
*   Nesting: Placeholders inside an explicit lambda (`{ ... }`), iterator body, `proc`, or `spawn` are scoped to that construct; they do not implicitly convert the outer expression into a placeholder lambda.
*   Errors:
    *   Using `@`/`@n` where a named identifier is required (outside expression placeholders) is a compile-time error.
    *   Arity mismatches between inferred placeholder lambdas and the expected function type at the call site are compile-time errors.

Example (nested placeholders remain local):
```able
builder = { ## explicit lambda; returns a callable that doubles its input
  fn_from_placeholder = (@ * 2)   ## placeholder inside lambda produces a separate function
  fn_from_placeholder
}

double = builder()
double(5) ## => 10; the outer lambda stays a normal function, only the placeholder expression becomes callable
```

Interaction with pipes:

- In a pipe step, the RHS must evaluate to a callable and is invoked with the subject as the first argument. Placeholders can be used to construct such callables: e.g., `x |> add(@, 1)`.
- `@`/`@n` construct anonymous functions within ordinary expressions; there is no dedicated pipe-topic token, and `%` is reserved for modulo.


### 7.7. Function Overloading

Able allows multiple functions (or inherent methods) with the same name in the **same scope** when their parameter shapes differ. Overloading never consults return types and does not change interface method contracts (interfaces still declare a single signature per method).

**Eligibility**
- Arity must match the call-site arity (positional arguments only; no defaults/varargs). If the final parameter is nullable (`?T`), the call may omit it (treated as `nil`).
- Parameter types must be compatible with the call-site arguments after applying ordinary type rules; a candidate that requires explicit generics is eligible only when the call supplies them.

**Resolution order**
1. Collect visible definitions with the target name in the applicable scope (top-level or inherent method set).
2. Filter by arity match.
3. Type-check each remaining candidate against the call-site arguments, performing generic inference per the existing rules. Discard candidates whose inference fails or whose parameter types are incompatible.
4. If exactly one candidate remains, select it.
5. If multiple remain, choose the **most specific**:
   - A non-generic signature beats a generic one when both match.
   - Between generics, a candidate whose parameter types are strictly tighter (concrete types or more instantiated shapes vs type variables) beats a looser one.
   - If no single candidate is strictly most specific, the call is ambiguous.

**Diagnostics**
- If no overloads match: “no overloads match call” (or the existing missing/argument-mismatch diagnostic).
- If multiple equally specific matches remain: “ambiguous overload; candidates: ...”.

**Non-participating surfaces**
- Interface methods are not overloaded; each interface method name has one signature, and multiple interface impls with the same method name remain subject to the impl-specific specificity/ambiguity rules in §10.2.5.
- Named implementations are never chosen implicitly via instance syntax; use the impl name in function position (`ImplName.method(value, ...)`) to disambiguate when needed.
**UFCS Callable Pool**: Inherent methods (when imported), interface methods (with impls in scope), and free functions in scope can all be candidates via method syntax per §9.4; named impls still require explicit selection.

## 8. Control Flow

This section details the constructs Able uses to control the flow of execution, including conditional branching, pattern matching, looping, range expressions, and non-local jumps.

### 8.1. Branching Constructs

Branching constructs allow choosing different paths of execution based on conditions or patterns. Both `if/elsif/else` and `match` are expressions.

#### 8.1.1. Conditional Chain (`if`/`elsif`/`else`)

This construct evaluates conditions sequentially and executes the block associated with the first true condition, with an optional default.

##### Syntax

```able
if Condition1 { ExpressionList1 }
[elsif Condition2 { ExpressionList2 }]
...
[elsif ConditionN { ExpressionListN }]
[else { DefaultExpressionList }]
```

-   **`if Condition1 { ExpressionList1 }`**: Required start. Executes `ExpressionList1` if `Condition1` (`bool`) is true.
-   **`elsif ConditionX { ExpressionListX }`**: Optional clauses. Executes `ExpressionListX` if its `ConditionX` (`bool`) is the first true condition in the chain.
-   **`else { DefaultExpressionList }`**: Optional final default block, executed if no preceding conditions were true.
-   **`ExpressionList`**: Sequence of expressions; the last expression's value is the result of the block.

##### Semantics

1.  **Sequential Evaluation**: Conditions are evaluated strictly in order.
2.  **First True Wins**: Execution stops at the first true `ConditionX`. The corresponding `ExpressionListX` is executed, and its result becomes the value of the chain.
3.  **Default Clause**: Executes if no conditions are true and the clause exists.
4.  **Result Value**: The chain evaluates to the result of the executed block. If no block executes (no conditions true and no default `else {}`), it evaluates to `nil`.
5.  **Type Compatibility**:
    *   If a default `else {}` guarantees execution, all result expressions must have compatible types. The chain's type is this common type.
    *   If no default `else {}` exists, non-`nil` results must be compatible. The chain's type is `?CompatibleType`.
    *   Unification rules:
        -   Common supertype via union (C1). If branches yield unrelated types `A` and `B`, the chain type is `A | B`.
        -   Nil special-case (N1). If one branch is `nil` and the other is `T`, the chain type is `?T`.
        -   Result special-cases (R1/R2). If branches are `T` and `!U`, upcast to `!T`/`!U` and unify (overall `!(T | U)` if needed). If branches are `!A` and `!B`, result is `!(A | B)`.
        -   Numeric exactness (E1). No implicit numeric coercions here beyond global operator rules; if numeric types differ and are not otherwise coerced, the unified type is a union (e.g., `i32 | f64`).

    Examples:
    ```able
    if c1 { 1 } else { "x" }                ## i32 | String
    if c1 { nil } else { v: T }             ## ?T
    if c1 { v: T } elsif c2 { w: U } else { read() } ## T | U
    if c1 { ok: T } else { read() }         ## if read() -> !U, then !(T | U)
    if c1 { 1 } else { 2.0 }                ## i32 | f64 (E1)
    ```

##### Example

```able
grade = if score >= 90 { "A" }
        elsif score >= 80 { "B" }
        else { "C or lower" } ## Guarantees String result
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
4.  **Exhaustiveness**: Compiler SHOULD check for exhaustiveness (especially for unions). Non-exhaustive matches MAY warn/error at compile time and SHOULD raise an exception at runtime. A `case _ => ...` usually ensures exhaustiveness.
    *   Open sets: When matching on an existential/interface type (e.g., `Error`), the set of possible concrete variants is open. Exhaustiveness for that component requires either a wildcard `case _ => ...` or at least `case _: Error => ...` to cover the open set.
5.  **Type Compatibility**: All `ResultExpressionListX` must yield compatible types. The `match` expression's type is this common type.
    *   Unification rules as for `if/elsif/else`:
        -   Union common supertype (C1); `nil` with `T` yields `?T` (N1).
        -   Result cases: `!A` with `!B` → `!(A | B)`; `T` with `!U` → `!(T | U)` (R1/R2).
        -   Numeric exactness (E1) — otherwise, union numeric types.

    Examples:
    ```able
    x match { case 1 => 1, case _ => "one" }     ## i32 | String
    x match { case v: T => v, case nil => nil } ## ?T
    r match { case a: A => a, case e: Error => default() } ## A | ReturnType(default)
    r match { case a: A => a, case e: Error => recover(e) } ## A | ReturnType(recover)
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

Loops execute blocks of code repeatedly. Loop expressions (`while`, `for`) evaluate to `void`.

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
-   If an element yielded by the iterator does not match the `Pattern`, the assignment expression evaluates to an `Error` value (some value implementating `Error`).

##### Example

```able
items = ["a", "b"] ## Array implements Iterable
for item in items { print(item) } ## Prints "a", "b"

total = 0
for i in 1..3 { ## Range 1..3 implements Iterable
  total = total + i
} ## total becomes 6 (1+2+3)
```

#### 8.2.3. Loop Expression (`loop`)

`loop { ... }` executes its body indefinitely until a `break` leaves the loop. Unlike `while`/`for`, `loop` is an expression that evaluates to the value supplied by the `break` that exits it.

##### Syntax

```able
loop {
  BodyExpressionList
}
```

##### Semantics

1.  The loop body runs, then control immediately begins the next iteration. There is no implicit condition.
2.  The only structured way to exit is `break`. A bare `break` yields `nil`; `break value` yields `value`. Labeled `break 'label` may target an enclosing `breakpoint` per §8.3.
3.  `continue` behaves identically to other loops: it skips the rest of the body and starts the next iteration.
4.  Because `loop` is an expression, it can appear wherever a value is expected.
5.  Authors should ensure the body can eventually execute a `break`; otherwise the loop is intentionally infinite.

#### 8.2.4. Continue Statement (`continue`)

Skips the remainder of the current loop iteration and proceeds with the next iteration of the innermost enclosing `for` or `while` loop.

##### Syntax

```able
continue
```

##### Semantics

-   `continue` statements may only appear inside a loop. Encountering `continue` transfers control to the loop's next iteration (or terminates the loop if its condition is now false or the iterator is exhausted).
-   `continue` never carries a value and always evaluates to `nil`.
-   **Labeled continues are not part of Able v11.** Attempting to write `continue 'label` (or any variant with a label) is a static error when detectable by tooling; interpreters must raise a runtime error with message `"Labeled continue not supported"` if such syntax is executed.

##### Example

```able
sum = 0
for n in numbers {
  if n < 0 { continue } ## Skip negative values
  sum = sum + n
}
```

#### 8.2.5. Range Expressions

Provide a concise way to create iterable sequences of integers.

##### Syntax

```able
StartExpr .. EndExpr   ## Inclusive range [StartExpr, EndExpr]
StartExpr ... EndExpr  ## Exclusive range [StartExpr, EndExpr)
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
break ['LabelName] [ValueExpression]
```

-   **`break`**: Keyword.
-   **`'LabelName`** (optional): A label identifying a target `breakpoint`. When present, must match a lexically enclosing `breakpoint`. When omitted inside a loop, the loop itself is the target.
-   **`ValueExpression`** (optional): Expression whose result becomes the value of the exited construct. Defaults to `nil` when omitted.

#### 8.3.3. Semantics

1.  **`breakpoint` Block Execution**:
    *   Evaluates `ExpressionList`.
    *   If execution finishes normally, the `breakpoint` expression evaluates to the result of the *last expression* in `ExpressionList`.
    *   If a `break 'LabelName ...` targeting this block occurs during execution (possibly in nested calls), execution stops immediately.
2.  **`break` Execution**:
    *   If a label is provided, finds the innermost lexically enclosing `breakpoint` with the matching `'LabelName`; otherwise, targets the innermost loop.
    *   Evaluates `ValueExpression` (or `nil` if omitted).
    *   Unwinds the call stack up to the target construct.
    *   Causes the target expression (loop or `breakpoint`) to evaluate to the result of `ValueExpression` (or `nil`).
    *   Labeled breaks targeting loops are not permitted in this revision and must raise an implementation error.
3.  **Type Compatibility**: The type of the `breakpoint` expression must be compatible with both the type of its block's final expression *and* the type(s) of the `ValueExpression`(s) from any `break` statements targeting it.
    *   Unification follows the same rules as `if/elsif/else` (C1, N1, R1/R2, E1). All normal block exits and `break` payloads are unified to produce the `breakpoint` expression's type (B1).

    Example:
    ```able
    result = breakpoint 'find {
      for x in xs { if p(x) { break 'find x } }
      nil
    }                    ## If xs: Array T, result: ?T (N1/B1)

    result2 = breakpoint 'mix {
      if c { break 'mix 1 } else { break 'mix "a" }
    }                    ## i32 | String (C1/B1)
    ```

4.  **Asynchrony boundary**: `break` only unwinds the current synchronous call stack. It cannot cross asynchronous boundaries introduced by `proc` or `spawn`. Attempting to target a `breakpoint` that is not in the current synchronous stack is a compile-time error when detectable, otherwise a runtime error.

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

#### 8.3.5. Loop Break Result

Loops evaluate to the value provided by the `break` that terminates them, which enables expression-oriented looping patterns.

##### Examples

```able
first_even = loop {
  value = stream.next()
  if value == nil { break nil }
  if value % 2 == 0 { break value }
}

## Retry until an operation succeeds, then return its payload
result = loop {
  outcome = try_fetch()
  outcome match {
    case { ok: payload } => break payload,
    case { retry } => continue,
    case { fatal: err } => break err
  }
}
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

1.  **Method-style functions (first param `Self`):** Operate on an instance of `TypeName`. Defined using:
    *   Explicit `self`: `fn method_name(self: Self, ...) { ... }`
    *   Shorthand `fn #`: `fn #method_name(...) { ... }` (implicitly adds `self: Self` as the first parameter). See Section [7.6.2](#762-implicit-self-parameter-definition-fn-method).
    *   `Self` refers to `TypeName` (with its generic arguments, if any).
    *   **Export/name:** Unless marked `private`, these export under the unqualified name (`method_name`) and are eligible for method-call sugar/UFCS when the name is in scope.
2.  **Type-qualified functions (first param not `Self`):** Associated with the nominal type but **not** usable via receiver syntax.
    *   Defined without `self` as the first parameter and without `#`.
    *   Exported symbol is `TypeName.function_name` (TypeName = the base nominal name; generic arguments are ignored for the symbol).
    *   Called only with type qualification: `TypeName.function_name<T>(...)` or `pkg.TypeName.function_name<T>(...)`. The qualifier never carries type arguments; they follow the function name with angle brackets.
    *   Never participates in method-call sugar/UFCS because the first parameter is not `Self`.

### 9.3. Example: `methods` block for `Address`

```able
struct Address { house_number: u16, street: String, city: String, state: String, zip: u16 }

methods Address {
  ## Instance method using shorthand definition and access
  fn #to_s() -> String {
    ## #house_number is shorthand for self.house_number, etc. See Section 7.6.1
    `${#house_number} ${#street}\n${#city}, ${#state} ${#zip}`
  }

  ## Instance method using explicit self
  fn update_zip(self: Self, zip_code: u16) -> void {
    self.zip = zip_code ## Could also use #zip here
  }

  ## Static method (constructor pattern)
  fn from_parts(hn: u16, st: String, ct: String, sta: String, zp: u16) -> Self {
    Address { house_number: hn, street: st, city: ct, state: sta, zip: zp }
  }
}

## Usage
addr = Address.from_parts(123, "Main St", "Anytown", "CA", 90210) ## Call static method
addr_string = addr.to_s()     ## Call instance method
addr.update_zip(90211)        ## Call instance method (mutates addr)
```

**Exports and function form:**

*   Functions declared inside a `methods` block behave like ordinary module-level functions. Unless marked `private`, they are exported and can be imported alongside other functions.
*   If the first parameter is `Self`, `receiver.method(args...)` is sugar for calling the exported function with `receiver` as the first argument.
*   If the first parameter is **not** `Self`, the export is `Type.method` and it is called only as `Type.method(...)` (or `pkg.Type.method(...)`); it never participates in `receiver.method(...)` sugar.
*   Generic/`where` obligations on the `methods` block continue to apply whether the function is invoked via member syntax (where applicable) or directly by name.

**Imports and visibility:**

*   `import pkg.*` brings both `foo` and `Type.foo` symbols into scope. Selective/rename imports work on both (`import pkg.{foo, Type.bar::alias}`).
*   Package-alias-only imports (`import pkg`) do **not** surface any function names; use qualification (`pkg.foo`, `pkg.Type.foo`) in that case.
*   Re-exports/aliases do not create new nominal types; method sets and type-qualified function symbols attach to the underlying type name.

### 9.4 Method Call Syntax Resolution

This section details how Able resolves `ReceiverExpression.Identifier(ArgumentList)` and how import scoping interacts with method-call sugar/UFCS.

**Resolution Steps (Unified Callable Pool):**

Let `ReceiverType` be the static type of `ReceiverExpression`. The compiler resolves as follows:

1.  **Callable Field:** If `ReceiverType` has a field named `Identifier` whose value is callable (a function/closure or implements `Apply`), call it directly with `ReceiverExpression` removed and `ArgumentList` as arguments. If the field exists but is not callable, invoking it is a compile-time error.
2.  **Callable Pool (names in scope only):** Gather callables named `Identifier` that are **actually in scope by name** (locally defined or imported via wildcard/selective/rename; package-alias-only imports do not contribute names) and whose first parameter is compatible with `ReceiverType`. Compatibility covers direct type equality (after alias canonicalization), interface satisfaction, and generic constraints (e.g., `fn foo<T: Display>(x: T, ...)` when `ReceiverType` satisfies `Display`). Sources:
    *   Inherent functions from `methods` blocks for the receiver type, but only if their **function names** are in scope.
    *   Interface methods for interfaces that `ReceiverType` implements and whose impls are in scope.
    *   Free/UFCS functions that are in scope and whose first parameter matches.
    *   **Excluded:** type-qualified functions (`Type.Identifier`) because their first parameter is not `Self`; call them via type qualification instead.
3.  **Select Most Specific:** Apply the existing overload/specificity rules (arity, type specificity, constraints) to choose a single best candidate from the pool. There is no category priority within the pool beyond specificity. The receiver is treated as the first argument for scoring.
4.  **Ambiguity:** If no single most specific candidate exists, emit a compile-time ambiguity error. The caller must disambiguate with explicit qualification (`pkg.name(receiver, ...)`, `Type.name(receiver, ...)`, `Interface.name(receiver, ...)`, or by narrowing imports).
5.  **Dispatch:** Invoke the chosen callable directly with `ReceiverExpression` prepended as the first argument, followed by `ArgumentList`. No pipe transformation is involved.

**Precedence and Error Handling:**

*   If a callable field is usable, it wins. Otherwise, resolution uses the unified pool. No match yields "method not found".
*   Named implementations still require explicit selection where applicable; visibility rules apply (import scope for free functions/inherent methods; interface impls must be in scope). Type-qualified functions require type qualification; they are never candidates for receiver syntax.

### 9.5. Method-Set Generics and Where-Clause Obligations

`methods` blocks may declare generic parameters and `where` clauses that impose additional constraints on the receiver or on helper type parameters. These obligations are enforced every time a method from the block is invoked.

- The `Self` alias always refers to the fully instantiated receiver type. Constraints such as `where Self: Display` therefore require the receiver itself to satisfy the referenced interface before the call can succeed.
- Generic parameters declared on the `methods` block (e.g., `methods Pair T { ... }`) behave like function-level generics. Any constraints placed on those parameters—either inline (`T: Display`) or through a `where` clause—must be satisfied once the compiler substitutes the concrete types inferred from the call.
- If a method inside the block introduces additional generics with their own constraints, those obligations are checked alongside the block-level requirements.

When method resolution (Section [9.4](#94-method-call-syntax-resolution)) selects a method from a `methods` block, the compiler instantiates all generic parameters, substitutes `Self` with the receiver type, and then validates the collected obligations. Failure to satisfy any constraint is a compile-time error, and the call does not proceed.

**Example**

```able
interface Display { fn show(Self) -> String }

struct Wrapper { value: String }

methods Wrapper where Self: Display {
  fn describe(self: Self) -> String {
    self.show()
  }
}
```

Calling `Wrapper.describe` requires a matching `impl Display for Wrapper`; otherwise the compiler reports that the `methods Wrapper.describe` constraint is not satisfied. The shared fixtures `errors/method_set_where_constraint` and `functions/method_set_where_constraint_ok` illustrate the failing and successful cases.

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
-   **`[GenericParamList]`**: Optional space-delimited generic parameters for the interface itself (e.g., `T`, `K V`). Constraints use `Param: Bound` inline or in the `where` clause. Free type names inside the signature list implicitly become parameters when no list is provided.
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
-   **`[GenericParamList]`**: Optional space-delimited generic parameters for the interface itself. Constraints use `Param: Bound` inline or in the `where` clause. When omitted, Able infers generics from free type names inside the signature list (mirroring §7.1.5).
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

Note on recursive `Self` constraints:

- Interfaces may reference `Self` in their own signatures (e.g., `fn next(self: Self) -> ?Self`). Recursive constraints over `Self` (e.g., requiring `Self: Interface` within the same interface) are allowed only when well-founded (no infinite regress) and remain an advanced feature; implementations must satisfy such constraints explicitly. Full formal rules are out of scope for v11 and may be tightened in a future revision.

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
    fn greet(self: Self) -> String { "Hello!" } ## Default implementation
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
-   **`[InterfaceArgs]`**: Space-delimited type expressions for the interface's generic parameters (if any), one argument per parameter. Space-delimited generic applications are only formed when parenthesized; otherwise each term is parsed as its own argument. Parenthesized or prefixed types (`?`/`!`) count as a single argument only when the generic application is parenthesized. Example: for a single-parameter interface, `impl Foo (Map K V) for ...` is valid while `impl Foo Map K V for ...` supplies three arguments and will fail arity checks. For a two-parameter interface, `impl Foo (Map K V) (Array T) for ...` supplies two arguments.
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
*   **Method Lookup Table:** Conceptually, the compiler maintains a mapping: for each pair `(InterfaceName, TargetType)`, it knows the location of the function implementations provided by the corresponding `impl` block. This mapping is crucial for method resolution (Section [9.4](#94-method-call-syntax-resolution)).
*   **Visibility:** The `impl` block itself follows the visibility rules of where it's defined. However, the *association* it creates between the type and interface becomes known wherever both the `TargetType` and `InterfaceName` are visible. Public types implementing public interfaces can be used polymorphically across package boundaries. Private types or implementations for private interfaces are restricted.
*   **Dispatch:** When an interface method is called (e.g., `receiver.interface_method()`), the compiler uses the type of `receiver` and the method name to look up the correct implementation via the `(InterfaceName, TargetType)` mapping established by the `impl` block (details in Section [9.4](#94-method-call-syntax-resolution) and Section [10.3](#103-usage)).

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
    fn greet(self: Self) -> String { "Hi from MyGreeter!" } ## Overrides default
}
```

#### 10.2.5. Overlapping Implementations and Specificity

When multiple `impl` blocks could apply to a given type and interface, Able uses specificity rules to choose the *most specific* implementation. If no single implementation is more specific, it results in a compile-time ambiguity error. Rules (derived from Rust RFC 1210, simplified):

**Applicability set**
- Consider only visible, unnamed impls for `(Interface, TargetType)` that unify with the receiver type (including generic substitutions).
- Named impls are never chosen implicitly; they require explicit qualification.

**Specificity ordering (strongest → weakest)**
1. **Concrete target beats generic target**
   `impl Show for Array i32` > `impl<T> Show for Array T`.
2. **Constraint superset beats subset**
   `impl<T: A+B> Show for T` > `impl<T: A> Show for T`.
3. **Union subset beats union superset**
   `impl Show for i32 | f32` > `impl Show for i32 | f32 | f64`.
4. **More-instantiated generics beat less-instantiated**
   `impl<T> Show for Pair T i32` > `impl<U V> Show for Pair U V` when the call site has `Pair i32 i32`.
5. If still tied, the call is ambiguous; no impl is chosen.

**Ambiguity handling**
- If multiple impls remain with no single most-specific winner, emit an ambiguity diagnostic listing candidates and suggest explicit disambiguation (e.g., hiding an import or calling a named impl).
- Interface methods are not overloaded; overloading rules do not resolve interface-impl ambiguity. Named impls remain opt-in via explicit qualification and are not considered in implicit selection.

**Dynamic/interface types**
- For interface-typed (existential) values, the impl dictionary is fixed at upcast time using the same rules. Ambiguity at upcast is an error.

**Examples**
```able
interface Show for T { fn show(self: Self) -> String }

## Concrete vs generic
impl Show for Array i32 { fn show(self: Self) -> String { `arr<i32>(${self.size()})` } }
impl<T> Show for Array T { fn show(self: Self) -> String { `arr(${self.size()})` } }
nums_i32: Array i32 = [1,2,3]
nums_str: Array String = ["a"]
nums_i32.show()  ## uses Array i32 impl (more specific)
nums_str.show()  ## uses generic Array T impl

## Constraint superset vs subset
impl<T: Eq+Hash> Show for T { fn show(self: Self) -> String { "eq+hash" } }
impl<T: Eq>      Show for T { fn show(self: Self) -> String { "eq only" } }
x: MyEqHash = ...
x.show() ## picks Eq+Hash impl as more specific

## Union subset vs superset
impl Show for i32 | f32 { fn show(self: Self) -> String { "narrow" } }
impl Show for i32 | f32 | f64 { fn show(self: Self) -> String { "wide" } }
v: i32 | f32 = 1
v.show() ## uses narrow impl

## Ambiguous (no winner)
impl<T: A> Show for T { ... }
impl<T: B> Show for T { ... }
y: T where T: A + B = ...
y.show() ## ambiguity diagnostic (no single most specific)

## Named impl disambiguation
Fmt = impl Show for Array i32 { fn show(self: Self) -> String { "fmt" } }
## Instance syntax does not pick named impls; call explicitly:
Fmt.show(nums_i32)
```

### 10.3. Usage

This section explains how interface methods are invoked and how different dispatch mechanisms work.

#### 10.3.1. Instance Method Calls and Dispatch

When calling a method using dot notation `receiver.method_name(args...)`:

*   **Resolution:** The compiler follows the steps outlined in Section [9.4](#94-method-call-syntax-resolution). If resolution points to an interface method implementation (Step 3), the specific `impl` block is identified.
*   **Static Dispatch (Monomorphization):**
    *   If the `receiver` has a **concrete type** (`Point`, `Array i32`, etc.) at the call site, the compiler knows the exact `impl` block to use. It typically generates code that directly calls the specific function defined within that `impl` block.
    *   This also applies when the `receiver` has a **generic type `T` constrained by the interface** (e.g., `fn process<T: Display>(item: T) { item.to_string() }`). During compilation (often through monomorphization), the compiler generates specialized versions of `process` for each concrete type `T` used, and the `item.to_string()` call within each specialized version directly calls the correct implementation for that concrete type.
    *   This is the most common form of dispatch – efficient and resolved at compile time.
*   **Dynamic Dispatch:**
    *   This occurs when the `receiver` has an **interface type** (e.g., `Display` used in a type position — see Section [10.3.4](#1034-interface-types-dynamic-dispatch)). The actual concrete type of the value is not known at compile time.
    *   The call `receiver.method_name(args...)` is dispatched at *runtime*. The `receiver` value (often represented as a "fat pointer" containing a pointer to the data and a pointer to a **vtable**) uses the vtable to find the address of the correct `method_name` implementation for the underlying concrete type and then calls it.
    *   This enables runtime polymorphism but incurs a small runtime overhead compared to static dispatch.

```able
p = Point { x: 1, y: 2 } ## Concrete type Point
s = p.to_string()       ## Static dispatch to 'impl Display for Point'

fn print_any<T: Display>(item: T) {
  print(item.to_string()) ## Static dispatch within monomorphized versions of print_any
}
print_any(p)      ## Instantiates print_any<Point>, calls Point's to_string
print_any("hi")   ## Instantiates print_any<String>, calls String.to_string

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

When method call resolution (Section [9.4](#94-method-call-syntax-resolution)) results in ambiguity (multiple equally specific interface methods) or when you need to explicitly choose a non-default implementation, use more qualified syntax:

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

    impl Display for Circle { fn to_string(self: Self) -> String { $"Circle({self.radius})" } }
    impl Display for Square { fn to_string(self: Self) -> String { $"Square({self.side})" } }

    ## Create an array holding values viewed through the 'Display' interface lens
    shapes: Array Display = [Circle { radius: 1.0 }, Square { side: 2.0 }]
    for s in shapes { print(s.to_string()) }
    ```
*   **Method Calls:** When a method is called on an interface-typed value (`item.to_string()` where `item` has type `Display`), the implementation corresponding to the underlying concrete type (captured at upcast time) is invoked at runtime.
*   **Static vs Dynamic Use:**
    -   In constraint positions (`T: Display` or `where T: Display`), `Display` is a static bound; calls are resolved at compile time.
    -   In type positions (`x: Display`, `Array Display`, `Error | T`), `Display` is a dynamic/existential; calls are resolved at runtime.
*   **Object Safety (minimum rule-set):** Methods callable via interface-typed values must be object-safe:
    -   No generic method parameters that are not fully constrained by the interface’s type parameters.
    -   No `Self` in return position unless wrapped in an interface-typed existential (or otherwise erased) defined by the interface.
    -   `self: Self` receiver only (no by-value moves across dynamic boundary unless the interface specifies the ownership model).
    Non-object-safe methods remain callable under static bounds (e.g., `T: Interface`) but are unavailable through interface-typed values.
*   **Import-Scoped Model:** The concrete implementation used for a dynamic/interface-typed value is fixed at the upcast site (where a concrete value is converted to an interface type) based on impls in scope there. Consumers do not need that impl in scope to call methods on the received interface value.

*   **Exhaustiveness reminder:** Because interface types represent open sets of implementors, pattern matching on an interface-typed value is only exhaustively covered with a wildcard or an explicit `case _: Interface` clause.


## 11. Error Handling

Able provides multiple mechanisms for handling errors and exceptional situations:

1.  **Explicit `return`:** Allows early exit from functions.
2.  **`Option T` (`?Type`) and `Result T` (`!Type`) types:** Used with V-lang style propagation (`!`) and handling (`or {}`) for expected errors or absence.
3.  **Exceptions:** For exceptional conditions, using `raise` and `rescue`.

### 11.1. Explicit `return` Statement

Functions can return a value before reaching the end of their body using the `return` keyword. (See also Section [7.3](#73-explicit-return-statement)).

#### Syntax
```able
return Expression
return ## Equivalent to 'return void' if function returns void
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

### 11.2. V-Lang Style Error Handling (`Option`/`Result`, `!`, `or`)

This mechanism is the default for handling *expected* errors or optional values gracefully without exceptions.

Policy:
-   Public and internal APIs that can fail in expected ways SHOULD return `!T` (or `?T` when absence is not an error) and use `!`/`or {}` at call sites.
-   Use exceptions only for truly exceptional conditions (see Section 11.3), not for routine control flow or recoverable failures.

#### 11.2.1. Core Types (`?Type`, `!Type`)

-   **`Option T` (`?Type`)**: Represents optional values. Defined implicitly as the union `nil | T`. Used when a value might be absent. (See Section [4.6.2](#462-nullable-type-shorthand-)).
    ```able
    user: ?User = find_user(id) ## find_user returns nil or User
    ```
-   **`Result T` (`!Type`)**: Represents the result of an operation that can succeed with a value of type `T` or fail with an error. Defined implicitly as the union `Error | T`. This follows the `FailureVariant | SuccessVariant` convention. The `!` shorthand is purely syntactic and does not depend on the variant order in user-declared unions.
    ```able
    ## The 'Error' interface (built-in or standard library, TBD)
    interface Error {
        fn message(self: Self) -> String
        fn cause(self: Self) -> ?Error
    }

    ## Result T is implicitly: union Result T = Error | T
    ## !Type is syntactic sugar for Result T

    ## Example function signature
    fn read_file(path: String) -> !String { ... } ## Returns Error or String
    ```

    Notes:
    - Shorthands compose positionally in types and apply to the immediate type to their right. For example, `?(!T)` denotes `nil | (Error | T)`. Parentheses are recommended when combining shorthands for readability.
    - The shorthands commute: `?(!T)` and `!(?T)` both denote `nil | Error | T`.

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
-   **Requirement:** The function containing the `!` operator must itself return a compatible `Option` or `Result` type, or a supertype union that contains `nil` and/or `Error` respectively. For example, a function returning `nil | Error | T` may use `!` on both `?U` and `!V` values.

Nested and composite cases:

- If the operand’s type is a union that may include both `nil` and `Error` (e.g., `nil | Error | T`, possibly arising from `?(!T)`), `expr!` unwraps the `T` on success and returns early on either `nil` or `Error`.

- For `?(!T)` and `!(?T)`, a single postfix `!` is sufficient and unambiguous: if the value is `nil`, return `nil`; if it is an `Error`, return that error; otherwise yield the `T` value. There is no need to chain `!`.
    - Symmetry: the same behavior applies to `!(?T)` since both forms normalize to `nil | Error | T`. A single postfix `!` on a `nil | Error | T` union is unambiguous: `nil` and `Error` are mutually exclusive failure variants.

``` able
## Single-step propagation/unwrap across both nil and Error
## !(?i32) ≡ nil | Error | i32
fn flatten(x: ?(!i32)) -> !(?i32) { x! }
```

##### Example
```able
## Assuming read_file returns !String (Error | String)
## Assuming parse_data returns !Data (Error | Data)
fn load_and_parse(path: String) -> !Data {
    content = read_file(path)! ## If read_file returns Err, load_and_parse returns it.
                               ## Otherwise, content is String.
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

#### 11.2.3. Error/Option Handling (`or {}`)

Provides a way to handle the `nil` or `Error` case of an `Option` or `Result` immediately, typically providing a default value or executing alternative logic. Implementations also treat any value that implements the `Error` interface as a failure for the purposes of `or` handling.

##### Syntax
```able
ExpressionReturningOptionOrResult or { BlockExpression }
ExpressionReturningOptionOrResult or { err => BlockExpression } ## Capture error
```

##### Semantics
-   Applies to an expression whose type is `?T` (`nil | T`) or `!T` (`Error | T`). Implementations must also treat *any* value whose runtime type implements the `Error` interface as a failure value for `or` (even if the static type is a plain union).
-   If the expression evaluates to the "successful" variant (`T`), the overall expression evaluates to that unwrapped value (`T`). The `or` block is *not* executed.
-   If the expression evaluates to the "failure" variant (`nil` or an `Error`/`Error`-implementer):
    *   The `BlockExpression` inside the `or { ... }` is executed.
    *   If the form `or { err => ... }` is used *and* the failure value was an `Error` (or an `Error`-implementer), the error value is bound to the identifier `err` (or chosen name) within the scope of the `BlockExpression`. If the failure value was `nil`, `err` is not bound.
    *   The entire `Expression or { ... }` expression evaluates to the result of the `BlockExpression`.
-   **Type Compatibility:** The type of the "successful" variant (`T`) and the type returned by the `BlockExpression` must be compatible. The overall expression has this common compatible type. If the two types are distinct and no expected type is provided by the surrounding context, the overall type is inferred as their union.

##### Example
```able
## Option handling
config_port: ?i32 = read_port_config()
port = config_port or { 8080 } ## Provide default value if config_port is nil
## port is i32

## Result Handling with error capture
user: ?User = find_user(id) or { err => ## Assuming find_user returns !User
    log(`Failed to find user: ${err.message()}`)
    nil ## Return nil if lookup failed
}

## Result Handling without capturing error detail
data = load_data() or { ## Assuming load_data returns !Array T
    log("Loading failed, using empty.")
    [] ## Return empty array
}
```

### 11.3. Exceptions (`raise` / `rescue`)

For handling truly *exceptional* situations that disrupt normal control flow, often originating from deeper library levels or representing programming errors discovered at runtime. Division/Modulo by zero raises an exception. Exceptions are orthogonal to `Option`/`Result`: the `!` propagation operator does not interact with exceptions; use `rescue` to handle them.

Policy:
-   Exceptions (via `raise`) are reserved for unrecoverable errors: programmer bugs (e.g., out-of-bounds, integer overflow), invariant/contract violations, resource corruption, or OS-level fatal errors.
-   Do not use exceptions for expected error cases in library or application APIs. Prefer returning `!T` and handling with `!`/`or {}`.
-   Interop: Exceptions from host languages should be converted to `!T` at the boundary where feasible (see Section 16). Use `rescue` sparingly for top-level fault containment.
-   Tooling note: Projects may enable lints/warnings to discourage `raise`/`rescue` in API implementations, except for approved exceptional cases.

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
  fn message(self: Self) -> String { "Division by zero" }
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
-   **Type Compatibility:** The normal result type of `MonitoredExpression` must be compatible with the result types of all `ResultExpressionListX` in the `rescue` block. If multiple handler branches produce distinct types and no common supertype is otherwise constrained by context, the overall type is the least upper bound (typically a union) of the normal and handler result types.

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
  fn message(self: Self) -> String
  fn cause(self: Self) -> ?Error
}

Runtime-generated errors (raised by the interpreter or stdlib helpers) automatically implement these members. They also expose a `value` field that returns the underlying payload used to construct the error (or `nil` when no payload exists). This enables Able programs to write handlers such as `err.value match { case ChannelClosed {} => ... }` without needing to reify the channel helper structs manually.

##### Standard Error Types

The standard library defines a small set of core error types implementing `Error` that the language/runtime may raise:

```able
## Arithmetic
struct DivisionByZeroError {}
impl Error for DivisionByZeroError { fn message(self: Self) -> String { "division by zero" } fn cause(self: Self) -> ?Error { nil } }

## Indexing
struct IndexError { index: u64, length: u64 }
impl Error for IndexError { fn message(self: Self) -> String { `index ${self.index} out of bounds for length ${self.length}` } fn cause(self: Self) -> ?Error { nil } }
```

Language-defined raises map to these errors:

-   Division or remainder by zero raises `DivisionByZeroError`.
-   Integer overflow raises `OverflowError { message: "integer overflow" }`; shift-out-of-range raises `ShiftOutOfRangeError { message: "shift out of range" }`.
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

#### 11.3.3. Runtime Exceptions (no panic abstraction)

Able does not have a distinct panic mechanism. All exceptional conditions are modeled as exceptions (values implementing `Error`) and handled via `raise`/`rescue`.

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

1.  **Asynchronous Start**: The target `FunctionCall` or `BlockExpression` is submitted to the runtime *executor*—a scheduling abstraction that may be backed by goroutines/threads (Go) or a cooperative queue (TypeScript). Execution starts independently of the caller, and the current task does *not* block.
2.  **Return Value**: The `proc` expression immediately returns a value whose type implements the `Proc T` interface.
    -   `T` is the return type of the `FunctionCall` or the type of the value the `BlockExpression` evaluates to.
    -   If the function/block returns `void`, the return type is `Proc void`.
3.  **Independent Execution**: The asynchronous task runs independently until it completes, fails, or is cancelled.

#### 12.2.3. Example

```able
fn fetch_data(url: String) -> String {
  ## ... perform network request ...
  "Data from {url}"
}

proc_handle = proc fetch_data("http://example.com") ## Starts fetching data
## proc_handle has type `Proc String`

computation_handle = proc do {
  x = compute_part1()
  y = compute_part2()
  x + y ## Block evaluates to the sum
} ## computation_handle has type `Proc i32` (assuming sum is i32)

side_effect_proc = proc { log_message("Starting background task...") } ## Returns Proc void
```

#### 12.2.4. Process Handle (`Proc T` Interface)

The `Proc T` interface provides methods to interact with an ongoing asynchronous process started by `proc`. The structs, union, and interface below form the canonical runtime surface; every interpreter/runtime MUST expose them with the exact spellings shown because shared fixtures and the Go/TypeScript typecheckers target these names.

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
## Could wrap exception information or specific error types.
struct ProcError { details: String } ## Example structure
impl Error for ProcError {
  fn message(self: Self) -> String { self.details }
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

Implementations may enrich `ProcError` with platform-specific fields, but they MUST preserve the `message()` contract and keep `status()`, `value()`, and `cancel()` available with the semantics spelled out below.

##### Semantics of Methods

-   **`status()`**: Returns the current state (`Pending`, `Resolved`, `Cancelled`, `Failed`) without blocking.
-   **`value()`**: Blocks the caller until the process finishes (resolves, fails, or is definitively cancelled).
    -   If `Resolved`, returns `value` where `value` has type `T`.
    -   If `Failed`, returns an error value of type `ProcError` (which implements `Error`) containing error details.
    -   If `Cancelled`, returns an error value of type `ProcError` indicating cancellation.
-   **`cancel()`**: Sends a cancellation signal to the asynchronous task. The task is not guaranteed to stop immediately or at all unless designed to check for cancellation signals.

##### Example Usage

```able
data_proc: Proc String = proc fetch_data("http://example.com")

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

Propagation inside `proc` tasks:

- Within a `proc` task body, the postfix `!` operator behaves as usual inside the task. Early returns triggered by `!` return from the task’s function. Observers then see the resulting `Proc` state:
  - If the task function returns `!T`, callers of `value()` receive that `!T` as-is (success or error) wrapped in the `Proc` protocol.
  - If the task function raises an exception (unhandled), `value()` returns `ProcError` describing the failure (or re-raises under a target that maps exceptions directly).

Cancellation example:

``` able
handle = proc do {
  ch = Channel.new(0)
  ## ... periodically check a user-provided cancellation flag or channel ...
  ## on cancel, exit early (return void)
}
handle.cancel()
st = handle.status()
```

#### 12.2.5. Runtime Helpers and Scheduling

Able exposes a small set of helper functions for coordinating asynchronous work. These functions are implemented natively by each runtime but share the same observable behaviour:

-   **`proc_yield()`** &mdash; May only be called from inside a `proc`/`spawn` task. Signals the executor that the current task is willing to yield so other queued work can run. Cooperative executors MUST requeue the yielding task behind any currently runnable work (providing a round-robin effect when multiple tasks yield). On Go’s goroutine executor the call is a no-op advisory hint: Go’s scheduler decides when the goroutine runs next. Programs MUST NOT rely on fairness guarantees beyond “other tasks have an opportunity to make progress”.
-   **`proc_cancelled()`** &mdash; May be called from inside a `proc` task to observe whether cancellation has been requested. Returns `true` when the task’s handle has been cancelled; `false` otherwise. Calling it outside asynchronous context raises a runtime error. Implementations SHOULD integrate this check with their cancellation primitives (e.g., Go `context.Context`).
-   **`proc_flush(limit?: i32)`** &mdash; Drains work from the executor’s queue up to an optional step limit (defaulting to 1024). Provided primarily for deterministic testing and fixture harnesses. Production runtimes with preemptive schedulers (e.g., Go) may treat it as a no-op because work progresses independently; cooperative schedulers MUST honour the limit to avoid starvation.
-   **`proc_pending_tasks()`** &mdash; Returns an `i32` count of runnable tasks currently queued on the executor. Cooperative runtimes MUST surface their queue length so fixtures/tests can assert that drains complete. Pre-emptive runtimes MAY return `0` (or a best-effort approximation of outstanding work) when their host scheduler does not expose per-queue metrics. The helper is diagnostic-only; programs MUST NOT rely on its value for functional correctness.

The helpers rely on the executor contract:

```able
## Conceptual interface implemented by each runtime.
interface Executor {
  fn schedule(task: () -> void) -> void
  fn ensure_tick() -> void
  fn flush(limit: i32 = 1024) -> void
  fn pending_tasks() -> i32 ## optional / best-effort for diagnostic helpers
}
```

Runtimes can provide different executor implementations (goroutine-backed, cooperative queue, deterministic serial) as long as the observable semantics described above are preserved.

##### Scheduler fairness

Runtime schedulers are expected to provide eventual forward progress for runnable procs, matching the behaviour of the Go runtime. Hosts that already provide pre-emptive scheduling—such as Go—inherit this property automatically and require no additional instrumentation. Cooperative implementations SHOULD emulate the same effect (for example, by yielding after bounded interpreter work or after blocking operations) so user-visible semantics remain aligned across runtimes. Implementations may offer additional test-only executors that enforce deterministic interleavings (e.g., always running the next ready task after a `proc_yield()`), but those guarantees are outside the core language contract.

##### Re-entrant waits

`proc_handle.value()` and `future.value()` may themselves be called from within other `proc`/`spawn` tasks (and even nested inside additional `value()` calls). Implementations MUST treat these operations as re-entrant:

- If the awaited task has already completed, `value()` returns immediately with the memoized result/error just as it would in top-level code.
- If the awaited task is still pending, the caller simply blocks until the awaited task resolves. Cooperative executors MUST continue to make progress by either resuming the awaited task inline (e.g., by draining the deterministic queue) or by ensuring it stays scheduled; they may not deadlock merely because the wait originated from another task.
- Pre-emptive runtimes (e.g., the Go goroutine executor) can rely on the host scheduler, but they must still guarantee that nested waits eventually unblock once the awaited task finishes.

This requirement ensures programs like nested futures/procs (see the shared `concurrency/future_value_reentrancy` and `concurrency/proc_value_reentrancy` fixtures) behave the same way across interpreters and targets.

### 12.3. Future-Based Asynchronous Execution (`spawn`)

The `spawn` keyword initiates asynchronous execution and returns a `Future T` value, which implicitly blocks and yields the result when evaluated in a `T` context. The result of a `Future T` is memoized: the first evaluation computes the result; subsequent evaluations return the memoized value (or error).

#### 12.3.1. Syntax

```able
spawn FunctionCall
spawn BlockExpression
```

-   **`spawn`**: Keyword initiating thunk-based asynchronous execution.
-   **`FunctionCall` / `BlockExpression`**: Same as for `proc`.

#### 12.3.2. Semantics

1.  **Asynchronous Start**: Starts the function or block execution asynchronously, similar to `proc`. The current thread does not block.
2.  **Return Value**: Immediately returns a value of the special built-in type `Future T`.
    -   `T` is the return type of the function or the evaluation type of the block.
    *   If the function/block returns `void`, the return type is `Future void`.
3.  **Implicit Blocking Evaluation**: The core feature of `Future T` is its evaluation behavior. When a value of type `Future T` is used in a context requiring a value of type `T` (e.g., assignment, passing as argument, part of an expression), the current thread **blocks** until the associated asynchronous computation completes.
    *   If the computation completes successfully with value `v` (type `T`), the evaluation of the `Future T` yields `v`.
    *   If the computation fails (raises an unhandled exception), evaluating the `Future T` re-raises that exception in the evaluating context. Use `rescue` to handle such failures.
    *   If the computation itself returns a `!T` (i.e., the underlying function returns `Error | T`), evaluating the `Future !T` yields that union value unchanged; no implicit wrapping occurs beyond memoization.
    *   Evaluating a `Future void` blocks until completion. If successful, it yields `void`. If the underlying task fails, it raises the exception to the evaluating context.

    Interaction with `!`:

    - Since `Future !T` evaluates to `!T`, the postfix `!` operator can be applied directly to a `Future !T` value. `future_result!` will block until completion, then propagate the `Error` early or return the unwrapped `T`.

#### 12.3.3. Example

```able
fn expensive_calc(n: i32) -> i32 {
  ## ... time-consuming work ...
  n * n
}

future_result: Future i32 = spawn expensive_calc(10)
future_void: Future void = spawn { log_message("Background log started...") }

print("Spawned tasks...") ## Executes immediately

## Evaluation blocks here until expensive_calc(10) finishes:
final_value = future_result
print(`Calculation result: ${final_value}`) ## Prints "Calculation result: 100"

## Evaluation blocks here until the logging block finishes:
_ = future_void ## Assigning to _ forces evaluation/synchronization
print("Background log finished.")
```

#### 12.3.4. Future Handle (Canonical Interface)

Even though values of type `Future T` behave like `T` when evaluated, the runtime also exposes an explicit handle so tooling and user code can inspect or await futures without forcing implicit evaluation in expression position. The handle reuses the `ProcStatus`/`ProcError` structs from Section 12.2.4 and MUST provide the following methods:

```able
interface Future T for FutureHandle {
  ## Non-blocking inspection of the memoized computation.
  fn status(self: Self) -> ProcStatus

  ## Blocks until the underlying asynchronous work settles and returns the memoized !T payload.
  ## Successful futures yield the resolved value; cancelled/failed futures yield ProcError.
  fn value(self: Self) -> !T
}
```

-   Futures are read-only views. There is intentionally no `cancel()` method; cancellation is requested via the originating `Proc` handle (if the program retained one) or by wiring cancellation-aware channels/flags into the asynchronous work.
-   `Future.status()` mirrors `Proc.status()` and MUST return the singleton structs `Pending`, `Resolved`, `Cancelled`, or `Failed`.
-   `Future.value()` is idempotent: once the future resolves, every subsequent call returns the cached result or error without re-running the computation. Typecheckers therefore model the return type as `!T`.

Implementations may expose additional helper functions (e.g., scheduler instrumentation) but MUST keep the methods above available so that diagnostics and fixtures can reason about futures uniformly across runtimes.

### 12.4. Using `proc` vs `spawn`

-   **Return Type:** `proc` returns `Proc T` (an interface handle); `spawn` returns `Future T` (a transparent, memoized result).
-   **Control:** `Proc T` offers explicit control (check status, attempt cancellation, get result via method call potentially handling errors).
-   **Result Access:** `Future T` provides implicit result access; evaluating the future blocks and returns the value directly. If the underlying computation raises, evaluation re-raises that exception in the evaluating context; use `rescue` to handle it.
    -   Accessing a `Future T` value has the same semantics as accessing a variable of type `T`; the access blocks until the value is available and yields the value directly (or re-raises on failure).
-   **Use Cases:**
    *   Use `proc` when you need to manage the lifecycle of the async task, check its progress, handle failures explicitly, or potentially cancel it.
    *   Use `spawn` for minimal syntax and transparent, memoized result delivery.

### 12.5. Synchronization Primitives (Crystal-style APIs, Go semantics)

Able provides standard library types `Channel T` and `Mutex` with APIs similar to Crystal, but the semantics are aligned with Go. These types are declared in Able source and use `extern <target>` bodies plus `prelude` initialisation to call into host-provided helpers; interpreters MUST expose the required native helpers so the APIs function uniformly across targets.

#### Channel T

Typed conduit for sending values between concurrent tasks.

Construction:
```able
## Unbuffered (rendezvous)
ch: Channel i32 = Channel.new(0)

## Buffered
ch_buf = Channel String |> new(64)
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

  ## Async helpers used by await (§12.6). They return Awaitable arms that invoke callbacks when ready.
  fn try_recv<R>(self: Self, cb: fn(T) -> R) -> (Awaitable R)
  fn try_send<R>(self: Self, value: T, cb: fn(void) -> R) -> (Awaitable R)
}
```

Semantics (Go-compatible):
-   Unbuffered channels (capacity 0) are rendezvous; send/receive both block until paired.
-   Buffered channels block send when full and block receive when empty; element order is FIFO.
-   `close()` may be called by the last sender; multiple closes raise a `ClosedChannelError`.
-   Sending on a closed channel raises a `SendOnClosedChannelError`.
-   `receive()` returns `?T` and yields `nil` when the channel is closed and drained.
-   Nil channels (uninitialized variables) block forever on send/receive; closing a nil channel raises a `NilChannelError`.
-   Cancellation is cooperative: if a suspended send/receive observes that its owning `proc` has been cancelled before progress is made, it unwinds with a cancellation error (mirroring Go’s `context` cancellation). Cancellation does **not** implicitly close the channel.
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

- Notes:
-   Use the `await` expression (§12.6) when coordinating channel operations with timers, sockets, or other async primitives.
-   Timeouts and cancellation can be modeled using auxiliary channels or higher-level APIs. Long-running tasks should periodically check for cancellation via user-defined channels or flags; there is no implicit ambient cancellation context.
-   Implementation note: no dedicated AST nodes exist for channels or mutexes. All operations are ordinary method calls; runtimes must extend their value representations (e.g., `V10Value`) with host-backed channel/mutex variants and integrate blocking operations with the cooperative scheduler.

Shared-data guidance and patterns:

- Prefer message passing (channels) over shared mutable memory. Design APIs to transfer ownership or pass immutable snapshots between tasks.
- When sharing mutable values across tasks, guard all access with `Mutex` and keep critical sections minimal. Avoid holding a lock across blocking operations (e.g., channel send/receive, I/O).
- Use `with_lock` helpers to guarantee unlock on early returns/exceptions.
- Consider copy-on-write or persistent structures when contention is high.

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
-   Non-reentrant: locking a mutex already held by the current task blocks permanently. This matches Go’s `sync.Mutex`: the waiter never acquires the lock, so programs must ensure a task releases the mutex before calling `lock()` again (for example, by structuring critical sections carefully or using `with_lock` helpers).
-   No poisoning: if an exception occurs while the mutex is held, subsequent lockers proceed; ensuring state consistency is the user's responsibility.

### 12.6. `await` Expression and the `Awaitable` Protocol

The `await` expression multiplexes asynchronous operations. Any value that implements the `Awaitable` interface can participate, allowing channels, timers, sockets, futures, and user-defined async helpers to share the same orchestration surface.

#### 12.6.1. Syntax

```able
await [ Awaitable1, Awaitable2, ... ]
```

-   The iterable may be an array literal or any collection of `Awaitable`s.
-   Each element may capture a callback. When `await` commits to an element, it invokes the element’s callback and yields that callback’s return value.
-   `Await.default(fn() -> R)` returns an always-ready arm that runs only when every other arm is pending. Only one default arm is permitted.

#### 12.6.2. Evaluation Rules

1.  Evaluate the iterable once, collecting the arms in source order.
2.  For every non-default arm call `is_ready()` (see §12.6.3). If one or more arms report `Ready(payload)`, choose a winner using fairness rules and execute its callback immediately.
3.  If none report ready:
    - Execute the default arm immediately if present; otherwise
    - Call `register(waker)` on every arm, park the current task, and block until any waker fires. When woken, the runtime re-checks readiness, commits to the ready arm, and cancels the others via their registration handles.
4.  The value of the `await` expression is the result returned by the chosen arm’s callback.

#### 12.6.3. `Awaitable` Interface

```able
interface Awaitable Output for SelfType {
  fn is_ready(self: Self) -> bool;
  fn register(self: Self, waker: AwaitWaker) -> AwaitRegistration;
  fn commit(self: Self) -> Output;
  fn is_default(self: Self) -> bool { false }
}
```

-   `is_ready()` is a non-blocking probe. It must not consume the underlying resource irrevocably; if the arm needs to capture data (e.g., received channel value) it stores it internally so `commit()` can read it later.
-   `register` attaches the arm to the runtime scheduler. Arms MUST call `waker.wake()` when they transition from pending to ready.
-   `AwaitRegistration` is an opaque handle the runtime uses to cancel the arm if another arm wins first.
-   `commit()` executes exactly once for the winning arm; it should invoke the user callback and return its result. Any buffered value captured during `is_ready()` is consumed here.
-   `is_default()` returns `true` only for the default arm returned by `Await.default`; the default implementation returns `false` so other awaitables do not opt in accidentally.

Channels, timers, sockets, futures, and other async APIs expose concrete `Awaitable` implementations that wrap their existing blocking operations. None of the underlying semantics change; only the orchestration surface does.

#### 12.6.4. Examples

```able
result = await [
  channel1.try_recv { msg =>
    print(`ch1: ${msg}`)
    msg
  },
  channel2.try_send(packet) {
    print("sent packet")
    true
  },
  Await.sleep(2.seconds) {
    print("timeout")
  },
  Await.default {
    print("idle")
  }
]

arms := sockets.map { sock => sock.try_recv { bytes => handle_sock(sock, bytes) } }
arms.push(cancel.try_recv(fn(_) => "cancelled"))
result = await arms
```

Because `await` accepts any iterable, callers can build arm lists dynamically.

#### 12.6.5. Fairness & Cancellation

-   **Fair selection:** When multiple arms become ready, runtimes must choose fairly so no arm is starved by source order (shuffle, rotating cursor, etc.).
-   **Blocking:** Only the winning arm’s operation commits; others remain pending for future `await` invocations.
-   **Cancellation:** If the enclosing `proc` is cancelled while blocked, runtimes must wake the awaiting task and propagate the same cancellation error used by other blocking primitives.

### 12.7. Channel and Mutex Error Types

Standardized error structs:

- `ChannelClosed`: Raised by APIs that treat closure as exceptional (e.g., strict receive helpers).
- `ChannelNil`: Raised when attempting to operate on a `nil` channel reference.
- `ChannelSendOnClosed`: Raised immediately when sending on a closed channel (including inside `await`).
- `ChannelTimeout`: Raised by helper APIs that implement deadlines/timeouts.
- `MutexUnlocked`: Raised when `unlock()` is called on an unlocked mutex.
- `MutexPoisoned`: Reserved for future poisoned-mutex semantics when a task panics while holding the lock.

Implementations must convert host-language panics/exceptions into these error values before surfacing them to Able code.

## 13. Packages and Modules

Packages form a tree of namespaces rooted at the name of the library, and the hierarchy follows the directory structure of the source files within the library.

### 13.1. Package Naming and Structure

*   **Root Package Name**: Every package has a root package name, defined by the `name` field in the `package.yml` file located in the package's root directory.
*   **Unqualified Names**: All individual package name segments (directory names, names declared with `package`) must be valid identifiers.
*   **Qualified Names**: Package paths are composed of segments delimited by periods (`.`), e.g., `root_package.sub_dir.module_name`.
*   **Directory Mapping**: Directory names containing source files are part of the package path. Hyphens (`-`) in directory names are treated as underscores (`_`) when used in the package path. Example: A directory `my-utils` becomes `my_utils` in the package path.
    -   Imports use the mapped identifier form. For a directory `my-utils`, write `import my_utils;` (not `import my-pkg;`).

### 13.2. Package Declaration in Source Files

*   **Single-segment declaration**: `package` accepts exactly one identifier segment (e.g., `package math;`). Qualified names such as `package util.math;` or any use of `::` are invalid; nested packages must be expressed via directories.
*   **Directory prefix always included**: The directory path of the file (relative to the package root, after hyphens are normalized to underscores) always prefixes the package path. A package declaration appends one additional segment; it never skips or replaces the directory-derived prefix.
*   **Optional Declaration**: A source file can optionally declare which sub-package its contents belong to using `package <unqualified-name>;`.
*   **Implicit vs. explicit package**:
    *   If a file `src/foo/bar.able` contains `package my_bar;`, and the root package name is `my_pkg`, its fully qualified package is `my_pkg.foo.my_bar`.
    *   If a file `src/foo/baz.able` has *no* `package` declaration, its fully qualified package is determined by its directory path relative to the root: `my_pkg.foo`.
    *   If a file `src/foo/qux.able` declares `package root;`, its fully qualified package is `my_pkg.foo.root` (the `foo` directory still contributes to the package path).
*   **Multiple Files**: Multiple files can contribute to the same fully qualified package name, either by residing in the same directory (without `package` declarations) or by declaring the same package name within different directories.

#### Example

Assume a package root `/home/david/projects/greet` with `package.yml` specifying `name: hello_world`.

```
/home/david/projects/greet/
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
    *   Renamed package import: `import internationalization::i18n;` (imports package under alias)
    *   Renamed selective import: `import io.{puts::print_line, gets};` (imports specific items with optional renames)
    *   `::` in imports is the rename operator only; package traversal continues to use dot (`.`). Outside imports and struct patterns, `::` has no meaning.
*   **Scope**: Imports can occur at the top level of a file (package scope) or within any local scope (e.g., inside a function).
*   **Binding Semantics**: Importing an identifier creates a new binding in the current scope. This binding refers to the same underlying definition (function, type, etc.) as the original identifier in the imported package.
*   **Re-exports preserve identity**:
    -   Selective or wildcard imports that are immediately re-exported (e.g., `import able.kernel.{Array}; export Array;` or `export * from able.kernel;`) do **not** create a new type. They bind the same nominal definition to an additional package path.
    -   Method or `impl` blocks declared for a re-exported struct/union/interface extend that original nominal type. Once the exporting package is imported, those methods become available on all values of that type, regardless of where the value was constructed.
    -   The only way to get a distinct nominal type is to declare a new `struct/union/interface` with its own name. Aliases (§4.7) and re-exports never introduce a fresh nominal type.
*   **Aliases extend the underlying type (even when the alias is private)**:
    -   `type Alias = Target` never introduces a new nominal type. Any `methods Alias { ... }` or `impl Interface for Alias` attaches to `Target`.
    -   The alias binding itself follows normal visibility rules; a private alias cannot be imported. However, public methods/impls defined for that alias are exported and become available to all packages that import the defining package (directly or through a re-exporting package that depends on it). Consumers do **not** need the alias binding to call those methods; importing the package is enough.
    -   Method/impl visibility is determined by the method/impl declaration's own `private` flag, not by the alias binding's visibility.
    -   Alias chains must be fully expanded to the canonical underlying type before method-set/impl lookup and dispatch.
    -   This applies to all packages, not just `able.kernel`: a package that wraps `foo.Bar` with `type Baz = foo.Bar` can add methods on `Baz`, and those methods extend `foo.Bar` for any client that imports the wrapper package.
    -   Implementations and method sets must be keyed by the canonical underlying type so that alias visibility does not gate method availability. The only way to obtain a distinct nominal type is to declare a fresh `struct/union/interface`.
    -   If multiple visible method sets attach the same method signature to the same canonical type, the call is ambiguous. Distinct signatures participate in overload resolution per §7.7; `impl` selection still follows §10.2.5.

#### Dynamic Imports (`dynimport`)

The `dynimport` statement binds identifiers from dynamically defined packages (created or extended at runtime via dynamic metaprogramming). It is resolved at runtime in the interpreter and is always available in Able builds that include the interpreter.

*   **Syntax Forms**:
    *   Package import: `dynimport foo;`
    *   Selective import: `dynimport foo.{Point, do_something};`
    *   Renamed import: `dynimport foo::f;`
*   **Resolution**:
    *   Looks up a dynamic package object via `dyn.package("foo")` and binds requested names from its current dynamic namespace.
    *   Fails at runtime with `Error` if the package or names do not exist.
*   **Scope**: May appear at top level or in local scopes of interpreted execution.
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
    -   Named implementations are never chosen implicitly. They require explicit selection (see Named Impl Invocation TBD) and follow the same visibility/import rules as other top-level items. Named impl identifiers must be unique within the importing scope; if collisions occur, use selective import with renaming (`import pkg.{ImplName::Alias}`).
    -   No orphan restriction: Packages may define `impl Interface for TargetType` even if they do not own the interface or the type. Which implementation is used is determined solely by what impls are in scope in the using package (via its imports).

    -   Specificity with multiple visible impls: If more than one unnamed `impl` is visible for the same `(Interface, TargetType)`, and one is strictly more specific (§10.2.5), the more specific one is chosen; otherwise, ambiguity is a compile-time error. Use imports to hide the undesired impl or call explicitly via a named implementation.

```able
## In package 'my_pkg'

foo = "bar" ## Public

private baz = "qux" ## Private to my_pkg

fn public_func() { ... }

private fn helper() { ... } ## Private to my_pkg

## Visibility examples

## Private concrete type and private impl; export only the interface view
private struct Hidden { value: i32 }
private impl Display for Hidden { fn to_string(self: Self) -> String { `${self.value}` } }

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

Typing and dynamic imports:

- Names brought in via `dynimport` are late-bound and not statically typed. In statically compiled code, they can be:
  - Called dynamically (raising an Error at runtime if shape/arity is incompatible), or
  - Adapted to interface types explicitly via `dyn.as_interface(value, Interface)` to cross into statically typed APIs.
- Using `dynimport`ed values directly in static type positions (e.g., as a parameter with a specific static type) is not permitted unless adapted as above.

### 13.6. Standard Library Packaging (`able.*`)

-   Able distributes Able-authored kernel and standard library bundles versioned with the toolchain. The kernel lives under `<tool-root>/kernel/src` and is automatically injected ahead of workspace code to surface host bridges (scheduler/channel/mutex/string/hasher shims).
-   The standard library is a normal package named `able` resolved through the usual search path + lockfile rules. When no pinned dependency is present, tooling falls back to the bundled copy at `<tool-root>/stdlib/src` (or `<tool-root>/stdlib/v11/src` when installed in a multi-version layout).
-   `import able.*` always resolves against whichever `able` root the loader selects. User code MUST NOT publish an `able` namespace; any root whose manifest declares `name: able` is treated as the stdlib, and the loader reports collisions rather than shadowing.
-   Tooling treats the bundled kernel/stdlib as read-only. Local edits rely on the dependency resolver (lockfile pin) or general search-path overrides (`ABLE_MODULE_PATHS`/workspace deps); there is no stdlib-specific environment knob.

### 13.7. Module Search Paths & Environment Overrides

`able run`/`able check` build a search order for package discovery before loading any modules. The combined list (duplicates removed) is constructed as follows:

1.  The directory containing the entry manifest plus every ancestor discovered by walking up from the entry file until a `package.yml` is found (see §13.1–§13.4). This is the *workspace root* and is always first.
2.  The current working directory and any explicit paths passed on the CLI (reserved for future flags) so ad-hoc scripts can import sibling files without manifests.
3.  Entries from `ABLE_PATH` (OS path list). This predates `ABLE_MODULE_PATHS` and is kept for backward compatibility.
4.  Entries from `ABLE_MODULE_PATHS` (also an OS path list). This is the preferred knob for injecting additional package roots—fixtures, local dependency mirrors, dynamic package sandboxes, etc. Both interpreters honor it for static imports *and* `dynimport`.
5.  Bundled kernel/stdlib roots discovered via the auto-detection logic described in §13.6 (including installed bundles) plus any resolver-provided copies (e.g., a lockfile pin for `able`).

Each search root is normalised to an absolute directory, verified to exist, and assigned a root package name. If a `package.yml` is present, its `name` field (after sanitising hyphen → underscore) becomes the root segment; otherwise the directory name is used. All `.able` files under the root participate in package assembly, and `package` statements inside those files may append further segments. If the same package path appears in multiple roots, the loader reports a collision rather than silently shadowing one with another.

`dynimport` uses the same ordered search list. At runtime it scans each root for the requested package path, falling back to later roots when a module is absent. This makes it possible to host production packages under the workspace root while letting tests or REPL sessions inject overrides via `ABLE_MODULE_PATHS=/tmp/able-overrides`.

## 14. Standard Library Interfaces (Conceptual / TBD)

Many language features rely on interfaces expected to be in the standard library. These require full definition.

Editorial note on built-ins vs. stdlib:

- Aside from primitives (`i*`, `u*`, `f*`, `bool`, `char`, `nil`, `void`), core collection/concurrency types used in this spec (e.g., `String`, `Array T`, `HashMap K V`, `Channel T`, `Mutex`, `Range`, plus interfaces like `Map K V`) are defined in the standard library. Syntactic constructs that reference them (array literals/patterns, indexing, ranges `..`/`...`) rely on those stdlib interfaces being in scope (e.g., `Index`, `Iterable`, `Range`). Implementations MUST provide a canonical stdlib that satisfies these expectations for the syntax to be usable. The kernel library bundled with the interpreter (v11 loads `v11/kernel`) supplies the foundational interfaces and minimal implementations; the stdlib is a normal dependency resolved via the package manager (defaulting to the bundled `able` stdlib when unspecified).

### 14.1. Language-Supported Interface Catalogue

The following interfaces enable core syntax/semantics. A type participates in a feature by implementing the corresponding interface(s) and ensuring the implementations are in scope at the call site.

#### 14.1.1. Indexing (`Index` / `IndexMut`)

```able
interface Index Key Value for Self {
  ## Return the element at `key`, raising IndexError (or subtype) if out of range.
  ## Calls return an Error on failure so callers can propagate via `!` or inspect directly.
  fn get(self: Self, key: Key) -> !Value
}

interface IndexMut Key Value for Self {
  ## Write `value` at `key`, raising IndexError if out of range.
  ## Calls return an Error on failure so callers can propagate via `!` or inspect directly.
  fn set(self: Self, key: Key, value: Value) -> !void
}
```

-   `receiver[key]` desugars to `receiver.Index.get(key)` and returns `!Value`.
-   `receiver[key] = value` desugars to `receiver.IndexMut.set(key, value)` (returning `!void`) and therefore requires both `Index` + `IndexMut`.
-   Implementations should surface `IndexError` (or a subtype implementing `Error`) when `key` is invalid so callers can propagate or inspect it.

Example:

```able
struct Foo { items: Array i64 }

impl Index u64 i64 for Foo {
  fn get(self: Self, idx: u64) -> !i64 { self.items[idx] }
}

impl IndexMut u64 i64 for Foo {
  fn set(self: Self, idx: u64, value: i64) -> !void {
    self.items[idx] = value
  }
}

foo := Foo { items: [10, 20, 30] }
value = foo[1]        ## -> 20
foo[2] = 40           ## Mutates foo.items[2]
```

#### 14.1.2. Iteration (`Iterator` / `Iterable`)

```able
struct IteratorEnd;

interface Iterator T for Self {
  fn next(self: Self) -> T | IteratorEnd;
}

interface Iterable T for Self {
  fn each(self: Self, visit: T -> void) -> void { ... }
  fn iterator(self: Self) -> (Iterator T) { ... }
}
```

-   `for element in collection { ... }`, `collection.each(...)`, and generator-based helpers rely on these interfaces (see “Core Iteration Protocol” below for default bodies).
-   Implementers may override `each` or `iterator` (or both) for efficiency. At least one must be supplied; the companion default derives the other.

Example—exposing a custom ring buffer as iterable:

```able
struct RingBuffer { storage: Array i64, len: u32 }

impl Iterable i64 for RingBuffer {
  fn iterator(self: Self) -> (Iterator i64) {
    Iterator i64 { gen =>
      i = 0
      while i < self.len {
        gen.yield(self.storage[i])
        i = i + 1
      }
    }
  }
}
```

#### 14.1.3. Callable Values (`Apply`)

```able
interface Apply Args Result for Self {
  fn apply(self: Self, args: Args) -> Result
}
```

-   `value(args...)` lowers to `value.apply(args...)` whenever the callee is a non-function value that implements `Apply`.
-   Implementations typically choose a tuple or struct to model `Args`. Closure literals/functions implement `Apply` implicitly; user-defined types can opt in for DSL-style callables.

Example:

```able
struct Multiplier { factor: i64 }

impl Apply i64 i64 for Multiplier {
  fn apply(self: Self, input: i64) -> i64 { self.factor * input }
}

doubler := Multiplier { factor: 2 }
result = doubler(21)   ## -> 42 via Apply.apply
```

#### 14.1.4. Arithmetic & Bitwise Operators

Operators are never customized directly; user-defined types participate by implementing the standard interfaces below. Each mapping corresponds to §6.3.2/§6.3.3 and the compiler requires that exactly one applicable implementation is in scope.

```able
interface Add Rhs Result for Self { fn add(self: Self, rhs: Rhs) -> Result }
interface Sub Rhs Result for Self { fn sub(self: Self, rhs: Rhs) -> Result }
interface Mul Rhs Result for Self { fn mul(self: Self, rhs: Rhs) -> Result }
interface Div Rhs Result for Self { fn div(self: Self, rhs: Rhs) -> Result }
interface Rem Rhs Result for Self { fn rem(self: Self, rhs: Rhs) -> Result }

interface Neg Result for Self { fn neg(self: Self) -> Result }
interface Not Result for Self { fn not(self: Self) -> Result }            ## bitwise complement

interface BitAnd Rhs Result for Self { fn bit_and(self: Self, rhs: Rhs) -> Result }
interface BitOr  Rhs Result for Self { fn bit_or(self: Self, rhs: Rhs) -> Result }
interface BitXor Rhs Result for Self { fn bit_xor(self: Self, rhs: Rhs) -> Result }

interface Shl Shift Result for Self { fn shl(self: Self, amount: Shift) -> Result }
interface Shr Shift Result for Self { fn shr(self: Self, amount: Shift) -> Result }
```

- `+` → `Add`
- `-` (binary) → `Sub`; `-` (unary) → `Neg`
- `*` → `Mul`
- `/` → `Div`
- `%` → `Rem`
- `.&` → `BitAnd`
- `.|` → `BitOr`
- `.^` → `BitXor`
- `.<<` → `Shl`
- `.>>` → `Shr`
- `.~` → `Not`

Implementers pick whatever `Result`/`Shift` types make sense (often `Self`/`u32`), but all operands must be concrete types.

#### 14.1.5. Comparison & Hashing

- `==`/`!=` require `PartialEq` (or the stronger `Eq`).
- `<`, `<=`, `>`, `>=` require `PartialOrd` (or `Ord`).
- Hash-based containers use `Hash` together with `Eq`.

```able
interface PartialEq Rhs for Self { fn eq(self: Self, other: Rhs) -> bool }
interface Eq for Self : PartialEq Self { fn eq(self: Self, other: Self) -> bool } ## Total equality

enum Ordering = Less | Equal | Greater
interface PartialOrd Rhs for Self { fn partial_cmp(self: Self, other: Rhs) -> Ordering }
interface Ord for Self : PartialOrd Self { fn cmp(self: Self, other: Self) -> Ordering }

interface Hash for Self {
  fn hash(self: Self, state: Hasher) -> void ## 'Hasher' is the stdlib sink used by maps/sets
}
```

Implementations must satisfy the usual algebraic laws (reflexivity, antisymmetry, etc.); violations result in undefined behavior for language-provided containers.

#### 14.1.6. Display & Errors

-   `Display` supplies `fn to_string(self: Self) -> String` (or similar) so values can appear in string interpolation and diagnostic output.
-   `Error` defines the minimum surface for throwable/reportable errors (`fn message(self: Self) -> String`, `fn cause(self: Self) -> ?Error`). Runtime-raised exceptions must produce values implementing `Error`.

#### 14.1.7. Clone, Default, Range, Concurrency

-   `Clone` produces deep-ish copies where needed (e.g., copying values out of borrowed containers).
-   `Default` provides fallback initialization (`fn default() -> Self`).
-   `Range` constructs iterables from `..`/`...` syntax (definitions below).
-   `Proc` / `ProcError` describe asynchronous handles (§12.2).

Built-in implementations:
- Primitives (`bool`, all integer/float widths, `String`) ship with implicit implementations of `Display`, `Clone`, `Default`, `Eq`/`Ord`, and `Hash` provided by the runtimes. These are always in scope and cannot be redefined by user code.
- Because the runtimes supply these unnamed impls, the stdlib must avoid re-declaring duplicate unnamed implementations for the same `(Interface, Type)` pairs to keep coherence/ambiguity rules intact. Where additional behavior is needed (e.g., custom hashing for composite structs), define implementations on those composite types instead of reintroducing primitive impls.

These interfaces, along with their implementations, form the contract between the language and the standard library. Authors creating new collection or numeric types should conform to these signatures so their types slot seamlessly into existing syntax.

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

### Awaitable Interface (Async Coordination)

`await` (§12.6) relies on the following runtime protocol:

```able
struct AwaitWaker { wake: fn() -> void }
struct AwaitRegistration { cancel: fn() -> void }

interface Awaitable Output for SelfType {
  fn is_ready(self: Self) -> bool;
  fn register(self: Self, waker: AwaitWaker) -> AwaitRegistration;
  fn commit(self: Self) -> Output;
  fn is_default(self: Self) -> bool { false }
}
```

-   `Await.default(fn() -> R)` is defined as an always-ready implementation whose `is_ready` returns `true`, whose `register` is a no-op, and whose `commit` simply invokes the supplied callback.
-   `is_default()` is reserved for that helper; other awaitables leave it at the default `false` implementation so the runtime can spot the single fallback arm.
-   Channel helpers, timers, sockets, futures, and other async APIs MUST implement this interface so the scheduler can coordinate them uniformly.
-   Cooperative runtimes drive the protocol directly; host runtimes (Go, TS, etc.) adapt `register`/`waker` to their native event APIs (wait queues, promises, epoll, …).

### 14.2. Text Modules (`able.text.regex`)

`able.text.regex` provides deterministic, RE2-style regular-expression facilities that run in time linear to the size of the pattern plus the haystack. The module sits on top of the `String`/`Grapheme` helpers from §6.12.1 and exposes a uniform API across every runtime.

#### 14.2.1. Core Types

-   `struct Regex { pattern: String, options: RegexOptions, program: RegexHandle }` — immutable compiled expression. `RegexHandle` is an opaque runtime-specific value; Able code never inspects it directly.
-   `struct RegexOptions { case_insensitive: bool, multiline: bool, dot_matches_newline: bool, unicode: bool, anchored: bool, unicode_case: bool, grapheme_mode: bool }`
    -   Defaults: `case_insensitive=false`, `multiline=false`, `dot_matches_newline=false`, `unicode=true`, `anchored=false`, `unicode_case=false`, `grapheme_mode=false`.
    -   `grapheme_mode` toggles cluster-aware iteration; when true, quantifiers and the `.` atom advance using `String.graphemes()` rather than raw code points.
-   `struct RegexError = InvalidPattern { message: String, span: Span } | UnsupportedFeature { message: String, hint: ?String } | CompileFailure { message: String }`
-   `struct Span { start: u64, end: u64 }` — byte offsets into the haystack (`start` inclusive, `end` exclusive). Offsets are measured in UTF-8 bytes and never exceed `String.len_bytes()`.
-   `struct Match { matched: String, span: Span, groups: Array Group, named_groups: Map String Group }`
-   `struct Group { name: ?String, value: ?String, span: ?Span }` — `span=nil` when the group did not participate in the match.
-   `struct Replacement = Literal(String) | Function(fn (match: Match) -> String)`
-   `struct RegexIter` implements `(Iterator Match)` and yields successive matches against a haystack.
-   `struct RegexSet` stores multiple compiled patterns and exposes aggregate matching APIs.
-   `struct RegexScanner` represents a stateful, streaming matcher that supports chunked input (e.g., sockets/files) without re-scanning from the beginning.

All structs are shareable across threads/procs; compiled programs are immutable and may be cached.

#### 14.2.2. Constructors & Helpers

| Helper | Signature | Description |
| --- | --- | --- |
| `Regex.compile` | `fn compile(pattern: String, options: RegexOptions = RegexOptions.default()) -> Result Regex RegexError` | Parses, validates, and compiles the pattern. No error is raised at runtime; invalid syntax is reported via `RegexError`. |
| `Regex.is_match` | `fn is_match(self: Regex, haystack: String) -> bool` | Returns true if at least one match exists. |
| `Regex.match` | `fn match(self: Regex, haystack: String) -> ?Match` | Returns the first match (if any). |
| `Regex.find_all` | `fn find_all(self: Regex, haystack: String) -> RegexIter` | Returns an iterator over non-overlapping matches. The iterator captures a reference to the haystack to keep spans valid. |
| `Regex.replace` | `fn replace(self: Regex, haystack: String, replacement: Replacement) -> Result String RegexError` | Applies either a literal replacement (`$1`/`\k<name>` substitutions follow the RE2 rules) or a callback that receives the current `Match`. |
| `Regex.split` | `fn split(self: Regex, haystack: String, limit: ?u64 = nil) -> Array String` | Splits on matches, mirroring `String.split` semantics with an optional match limit. |
| `Regex.scan` | `fn scan(self: Regex, haystack: String) -> RegexScanner` | Produces a stateful scanner when chunked processing is required. |
| `regex_is_match` | `fn regex_is_match(pattern: String, haystack: String, options: RegexOptions = RegexOptions.default()) -> Result bool RegexError` | Convenience helper used throughout the stdlib/testing packages. |

`Regex.to_program()` (implementation-defined) may expose the compiled automaton for tooling and debugging. Engines MAY additionally surface `Regex.captured_names()` and other introspection helpers provided they remain deterministic.

#### 14.2.3. Execution Semantics

-   Matching walks the haystack in code-point units by default. All pattern constructs—ranges, classes, `.`—interpret text as Unicode scalar values. When `grapheme_mode` is enabled, the engine segments the haystack using `String.graphemes()` before evaluating atoms and quantifiers so user-perceived characters stay intact.
-   Every API guarantees linear time with respect to `pattern.len_chars() + haystack.len_chars()`. Backreferences and other constructs that require unbounded backtracking are rejected with `RegexError.UnsupportedFeature`.
-   Matches borrow from the haystack. `Match.matched` and each group’s `value` are `String` instances created by slicing the original bytes; their `span` offsets always refer to the source haystack and remain valid even if the haystack is larger than the matched segment.
-   Capture groups are numbered in declaration order; named groups populate both `groups` (by ordinal) and `named_groups` (by identifier). Groups that do not participate in the match return `value=nil`, `span=nil`.
-   Replacement callbacks run synchronously and may `raise`; the error propagates to the caller and aborts the replace operation at the current match.

#### 14.2.4. Regex Sets & Streaming

-   `RegexSet.compile(patterns: Array String, options: RegexOptions = RegexOptions.default()) -> Result RegexSet RegexError` compiles every pattern into a single automaton. Matching APIs include:
    -   `fn is_match(self: RegexSet, haystack: String) -> bool`
    -   `fn matches(self: RegexSet, haystack: String) -> Array u64` (indices of patterns that matched)
    -   `fn iter(self: RegexSet, haystack: String) -> (Iterator RegexSetMatch)` where `RegexSetMatch { pattern_index: u64, span: Span }`
-   `RegexScanner` exposes `fn feed(self: RegexScanner, chunk: String) -> void` and `fn next(self: RegexScanner) -> Match | IteratorEnd`. Scanners must honour Able’s cooperative scheduling: long-running scans call `proc_yield()` between chunks when running under the interpreter.
-   Streaming scanners and regex sets share the same deterministic guarantees as single-pattern regexes. Partial matches that span chunk boundaries buffer the necessary bytes internally until a decision can be made.

#### 14.2.5. Integration Points

-   The testing helpers (`able.spec.match_regex`) delegate to `regex_is_match`, so diagnostic output and failure modes flow through the regex module rather than bespoke matchers.
-   Channel/mutex diagnostics (§12.7) and the text processing utilities consume regex spans directly; consumers should treat `Span.start`/`Span.end` as byte offsets and rely on `String.substring` or `String.graphemes()` when human-facing presentation is needed.
-   Because regex captures rely on byte offsets, any API that displays indices alongside grapheme-oriented UIs must convert them explicitly (e.g., by counting graphemes up to `span.start`).

## 15. Program Entry Point

Able programs may define one or more executables via `main` functions located in packages. This section specifies the entrypoint rules.

### 15.1. Location and Multiplicity

-   Multiple binaries are supported: any package that defines a non-private, top-level `fn main() -> void` produces an executable named after that package path (build tooling may provide renaming). If dependencies also define `main`, they produce their own binaries when built as roots; they do not affect the current package's binary unless explicitly selected by tooling.

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
fn main() -> void {
    print("Hello, Able!")
    ## Access args via os.args()
    ## Exit explicitly if desired: os.exit(12)
}
```

## 16. Host Interop (Target-Language Inline Code)

Able allows embedding function bodies and package-scope preludes written in the target host language (e.g., Go, Crystal, TypeScript, Python, Ruby). This is distinct from FFI: host interop is for writing target-language code that is compiled/linked as part of the same binary the Able code compiles into. Structs/unions are not implicitly mapped across the boundary; only the core primitive and container mappings listed below are supported. Passing complex data structures requires explicit serialization or manually mirrored struct definitions on the host side with adapter code.

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
-   Host code inside a prelude must follow the host language's top-level syntax rules (e.g., imports for Go).

#### 16.1.2. Extern Host Function Bodies

```able
extern go fn now_nanos() -> i64 { return time.Now().UnixNano() }

extern crystal fn new_uuid() -> String { UUID.random.to_s }

extern typescript fn read_text(path: String) -> !String {
  try { return readFileSync(path, "utf8") } catch (e) { throw host_error(string(e)) }
}

extern python fn now_secs() -> f64 { return time.time() }

extern ruby fn new_uuid() -> String { SecureRandom.uuid }
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
-   String → string (Go/TS); string (Crystal/Ruby/Python)
-   Array T → []T (Go); Array(T) (Crystal); T[] (TS); list[T] (Python); Array(T) (Ruby) — copy-in/copy-out
-   ?T (Option) → nil/None/null for "no value" in the host; otherwise T mapping above
-   !T (Result) →
    -   Go: (T, error)
    -   Crystal/TS/Python/Ruby: return T or raise/throw; uncaught becomes Able Error

### 16.3. Error Mapping

-   Provide `host_error(message: String)` helper inside extern bodies to produce an Able `Error`. The helper's name and signature are standardized across targets; implementations MUST expose it wherever extern bodies are permitted.
-   Go: return (zero, err) → Able `Error` at boundary when `err != nil`.
-   Crystal/TypeScript/Python/Ruby: raise/throw → Able `Error` at boundary.

### 16.4. Concurrency and Execution

-   Extern bodies execute in the caller's goroutine/fiber/thread and may block.
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

extern go fn read_file(path: String) -> !String {
  data, err := os.ReadFile(path)
  if err != nil { return host_error(err.Error()) }
  return string(data)
}
```

Crystal:
```able
prelude crystal { require "uuid" }
extern crystal fn new_uuid() -> String { UUID.random.to_s }
```

TypeScript:
```able
prelude typescript { import { readFileSync } from "node:fs"; }
extern typescript fn read_text(path: String) -> !String {
  try { return readFileSync(path, "utf8") } catch (e) { throw host_error(string(e)) }
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
extern ruby fn new_uuid() -> String { SecureRandom.uuid }
```

## 17. Tooling: Testing Framework

Able ships a stdlib-backed testing framework (`able.test.*` protocol + `able.spec`
DSL). Tooling must treat test modules as a distinct build profile so production
builds remain slim but tests can still share package scope.

### 17.1. Test Modules

- Any source file ending in `.test.able` or `.spec.able` is a **test module**.
- Test modules belong to the same package namespace as production modules in the
  same directory tree, so they may access private members.
- Standard commands (`able build`, `able run`, `able check`, etc.) ignore test
  modules unless `--with-tests` is explicitly enabled.
- `able test` always enables the test profile and typechecks production + test
  sources together to preserve privacy semantics.

### 17.2. `able test` Command Contract

- `able test [OPTIONS] [TARGETS...]` discovers test modules in scope, loads them
  once, and evaluates the registered frameworks via `able.test.harness`.
- Discovery uses `able.test.protocol.DiscoveryRequest` with CLI-supplied path,
  name, and tag filters; `--list` performs discovery only.
- Execution uses `able.test.harness.run_plan` with `RunOptions` populated from
  flags like `--shuffle`, `--repeat`, `--parallel`, and `--fail-fast`.
- `--format` selects the reporter output format (doc/progress/tap/json).
- Exit codes: `0` success, `1` test failures, `2` discovery/runtime errors.

### 17.3. Standard Library Surface

- `able.test.protocol` defines the shared structs and interfaces.
- `able.test.registry` stores framework registrations triggered by imports.
- `able.test.harness` provides `discover_all` + `run_plan` orchestration.
- `able.test.reporters` supplies default reporters used by the CLI.
- `able.spec` is the default spec-style DSL; importing it registers a framework.

# Todo

*   **Standard Library Implementation:** Core types (`Array`, `Map`?, `Set`?, `Range`, `Option`/`Result` details, `Proc`, `Future`), IO, string methods, Math, `Iterable`/`Iterator` protocol, Operator interfaces. Definition of standard `Error` interface.
*   **Type System Details:** Full inference rules, Variance, Coercion (if any), HKT limitations/capabilities.
*   **Object Safety Rules:** Which interface methods are callable from interface-typed values; any boxing/erasure rules; formal vtable capture at upcast.
*   **Pattern Exhaustiveness:** Rules for open sets like `Error` and refutability constraints.
*   **Re-exports and Named Impl Aliasing:** Precise import/alias collision rules and diagnostics.
*   **Ranges:** Concrete type vs existential for `..` and `...` results.
*   **Tooling:** Compiler, Package manager commands.

# Unresolved questions

* Shared Data in Concurrency (12.5): Unresolved—awaiting "races and ownership patterns" note with examples.
* HKTs/Variance/Coercion: Unresolved—awaiting minimal rules.
* Self Interpretation (10.1.3): Unresolved—no recursive details yet.
