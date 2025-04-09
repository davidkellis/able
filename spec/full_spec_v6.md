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

*   **Type:** A named classification representing a set of possible values (e.g., `bool` represents `{true, false}`). Every value has a single, specific type.
*   **Parametric Nature:** All types are conceptually parametric (zero or more type parameters).
*   **Type Expressions:** Syntax for denoting types (type names, space-delimited arguments, parentheses for grouping).
*   **Parameter Binding:** Parameters are **bound** when specific types/variables are provided, otherwise **unbound** (`_` or omitted).
*   **Concrete Type:** All type parameters are bound. Values have concrete types.
*   **Polymorphic Type / Type Constructor:** Has unbound parameters (e.g., `Array`, `Map string _`). Used in HKT contexts.
*   **Constraints:** Restrict generic parameters using interfaces (`T: Interface1 + Interface2`). Applied in definitions using `:` or `where` clauses.

### 4.2. Primitive Types

| Type     | Description                                   | Literal Examples                    | Notes                                           |
| :------- | :-------------------------------------------- | :---------------------------------- | :---------------------------------------------- |
| `i8`     | 8-bit signed integer (-128 to 127)            | `-128`, `0`, `10`, `127_i8`         |                                                 |
| `i16`    | 16-bit signed integer (-32,768 to 32,767)       | `-32768`, `1000`, `32767_i16`        |                                                 |
| `i32`    | 32-bit signed integer (-2³¹ to 2³¹-1)           | `-2_147_483_648`, `0`, `42_i32`      | Default type for integer literals (TBC).        |
| `i64`    | 64-bit signed integer (-2⁶³ to 2⁶³-1)           | `-9_223_..._i64`, `1_000_000_000_i64`| Underscores `_` allowed for readability.        |
| `i128`   | 128-bit signed integer (-2¹²⁷ to 2¹²⁷-1)      | `-170_..._i128`, `0_i128`, `170_...`|                                                 |
| `u8`     | 8-bit unsigned integer (0 to 255)             | `0`, `10_u8`, `255_u8`              |                                                 |
| `u16`    | 16-bit unsigned integer (0 to 65,535)           | `0_u16`, `1000`, `65535_u16`        |                                                 |
| `u32`    | 32-bit unsigned integer (0 to 2³²-1)            | `0`, `4_294_967_295_u32`            | Underscores `_` allowed for readability.        |
| `u64`    | 64-bit unsigned integer (0 to 2⁶⁴-1)            | `0_u64`, `18_446_...`               |                                                 |
| `u128`   | 128-bit unsigned integer (0 to 2¹²⁸-1)          | `0`, `340_..._u128`                 |                                                 |
| `f32`    | 32-bit float (IEEE 754 single-precision)      | `3.14_f32`, `-0.5_f32`, `1e-10_f32`  | Suffix `_f32`.                                  |
| `f64`    | 64-bit float (IEEE 754 double-precision)      | `3.14159`, `-0.001`, `1e-10`, `2.0`  | Default type for float literals (TBC). Suffix `_f64`. |
| `string` | Immutable sequence of Unicode chars (UTF-8) | `"hello"`, `""`, `` `interp ${val}` `` | Double quotes or backticks (interpolation).      |
| `bool`   | Boolean logical values                        | `true`, `false`                     |                                                 |
| `char`   | Single Unicode scalar value (UTF-32)        | `'a'`, `'π'`, `'💡'`, `'\n'`, `'\u{1F604}'` | Single quotes. Supports escape sequences.       |
| `nil`    | Singleton type representing **absence of data**. | `nil`                               | **Type and value are both `nil` (lowercase)**. Used with `?Type`. |
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

Structs aggregate named or positional data fields into a single type. Fields are mutable. A single struct definition must be exclusively one kind (singleton, named-field, or positional-field).

#### 4.5.1. Singleton Structs

Represent types with exactly one value, identical to the type name itself.

*   **Declaration:** `struct Identifier;` (or `struct Identifier {}`)
*   **Usage:** Use the identifier directly as the value: `status = Success`. Matched as `case Success => ...`.

#### 4.5.2. Structs with Named Fields

Group data under named fields.

*   **Declaration:** `struct Identifier [GenericParamList] { FieldName1: Type1, FieldName2: Type2 ... }`
*   **Instantiation:** Use `{ FieldName: Value, ... }`. Order irrelevant. All fields must be initialized. Field init shorthand `{ name }` supported.
    ```able
    p := Point { x: 1.0, y: 2.0 }
    username := "Alice"
    u := User { id: 101, username, is_active: true } ## Shorthand
    ```
*   **Field Access:** Dot notation: `instance.FieldName`. Example: `p.x`.
*   **Functional Update:** Create a new instance based on others using `...Source`. Later sources/fields override earlier ones.
    ```able
    addr2 := Address { ...addr1, zip: "90210" }
    ```
*   **Field Mutation:** Modify fields in-place using assignment (`=`): `instance.FieldName = NewValue`. Example: `p.x = 3.0`. Requires `p` to be an existing mutable binding.

#### 4.5.3. Structs with Positional Fields (Named Tuples)

Define fields by their position and type. Accessed by index.

*   **Declaration:** `struct Identifier [GenericParamList] { Type1, Type2 ... }`
*   **Instantiation:** Use `{ Value1, Value2, ... }`. Values must be provided in the defined order. All fields must be initialized.
    ```able
    pair := IntPair { 10, 20 }
    ```
*   **Field Access:** Dot notation with zero-based integer index: `instance.Index`. Example: `first := pair.0`.
*   **Functional Update:** Not supported via `...Source` syntax.
*   **Field Mutation:** Modify fields in-place using indexed assignment (`=`): `instance.Index = NewValue`. Example: `pair.0 = 15`. Requires `pair` to be an existing mutable binding.

### 4.6. Union Types (Sum Types / ADTs)

Represent values that can be one of several different types (variants).

#### 4.6.1. Union Declaration

Define a new type as a composition of existing variant types using `|`.

*   **Syntax:** `union UnionTypeName [GenericParamList] = VariantType1 | VariantType2 | ... | VariantTypeN`
*   **Example:**
    ```able
    struct Red; struct Green; struct Blue;
    union Color = Red | Green | Blue
    union IntOrString = i32 | string
    union Option T = T | nil
    union Result T ErrorType = T | ErrorType
    ```

#### 4.6.2. Nullable Type Shorthand (`?`)

Concise syntax for types that can be `nil` or a specific type.

*   **Syntax:** `?Type`
*   **Equivalence:** `?Type` is syntactic sugar for the union `nil | Type`.
*   **Example:** `maybe_user: ?User = find_user(id)`

*(See Section [11.2.1](#1121-core-types-type-type) for `!Type` shorthand)*

#### 4.6.3. Constructing Union Values

Create a value of the union type by creating a value of one of its variant types.

```able
c: Color = Green
opt_val: ?i32 = 42
maybe_error: !string = "Success value"
```

#### 4.6.4. Using Union Values

Primarily used with `match` expressions (See Section [8.1.2](#812-pattern-matching-expression-match)).

```able
display_name = maybe_name match {
  case s: string => s,
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

Source code representations of fixed values.

*   **Integers:**
    *   Decimal: `123`, `0`, `1_000_000`
    *   Hexadecimal: `0xff`, `0xDEAD_BEEF`
    *   Binary: `0b1010`, `0b1111_0000`
    *   Octal: `0o777`, `0o123_456`
    *   Type Suffixes: `_i8`, `_u8`, `_i16`, `_u16`, `_i32`, `_u32`, `_i64`, `_u64`, `_i128`, `_u128`. Example: `42_i64`, `255_u8`, `0xff_u32`.
    *   Default Type: `i32` (TBC) if no suffix.
*   **Floats:** `3.14`, `0.0`, `-1.2e-5`, `1_000.0`.
    *   Type Suffixes: `_f32`, `_f64`. Example: `2.718_f32`, `1.0_f64`.
    *   Default Type: `f64` (TBC) if no suffix.
*   **Booleans:** `true`, `false`.
*   **Characters:** `'a'`, `' '`, `'\n'`, `'\u{1F604}'`. Single quotes. Escape sequences (`\n`, `\t`, `\\`, `\'`, `\u{...}`).
*   **Strings:** `"Hello"`, `""`, `` `Val: ${x}` ``. Double quotes or backticks (interpolation). Escape sequences same as char.
*   **Nil:** `nil`. Represents absence of data.
*   **Arrays:** `[1, 2, 3]`, `["a", "b"]`, `[]`. (Requires `Array` type definition in stdlib).
*   **Structs:** `{ field: val }`, `{ val1, val2 }`. (See Section [4.5](#45-structs)).

### 6.2. Block Expressions (`do`)

Execute a sequence of expressions within a new lexical scope.

*   **Syntax:** `do { ExpressionList }`
*   **Semantics:** Expressions evaluated sequentially. Introduces scope. Evaluates to the value of the last expression.
*   **Example:** `result := do { x := f(); y := g(x); x + y }`

### 6.3. Operators

Symbols performing operations.

#### 6.3.1. Operator Precedence and Associativity

(Follows Rust model, `~` for bitwise NOT)

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
| 1          | `:=`                  | **Declaration and Initialization**      | Right-to-left |                                                           |
| 1          | `=`                   | **Reassignment / Mutation**             | Right-to-left |                                                           |
| 1          | `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=` | Compound Assignment (TBD)               | Right-to-left | (Needs formal definition, acts like `=`)                  |
| 0          | `\|>`                 | Pipe Forward                            | Left-to-right | (Lowest precedence)                                       |

#### 6.3.2. Operator Semantics

*   Arithmetic (`+`, `-`, `*`, `/`, `%`): Standard math operations. Division (`/`) or remainder (`%`) by zero **raises a runtime exception** (e.g., `DivisionByZeroError`).
*   Comparison (`>`, `<`, `>=`, `<=`, `==`, `!=`): Result `bool`. Relies on standard library interfaces (`Eq`, `Ord`).
*   Logical (`&&`, `||`, `!`): Short-circuiting `&&`, `||` on `bool`. `!` negates `bool`.
*   Bitwise (`&`, `|`, `^`, `<<`, `>>`, `~`): Standard operations on integer types. `~` is complement.
*   Unary (`-`): Arithmetic negation.
*   Member Access (`.`): Access fields/methods, UFCS.
*   Function Call (`()`): Invokes functions/methods.
*   Indexing (`[]`): Access elements within indexable collections (relies on `Index`/`IndexMut`).
*   Range (`..`, `...`): Create `Range` objects.
*   Declaration (`:=`): Declares/initializes new variables. Evaluates to RHS.
*   Assignment (`=`): Reassigns existing variables or mutates locations. Evaluates to RHS.
*   Compound Assignment (`+=`, etc. TBD): Shorthand (e.g., `a += b` like `a = a + b`). Acts like `=`.
*   Pipe Forward (`|>`): `x |> f` evaluates to `f(x)`.

#### 6.3.3. Overloading (Via Interfaces)

Behavior for non-primitive types relies on implementing standard library interfaces (e.g., `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg`, `Not` (for `~`), `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`, `PartialEq`, `Eq`, `PartialOrd`, `Ord`, `Index`, `IndexMut`). Logical `!` typically not overloaded. See Section [14](#14-standard-library-interfaces-conceptual--tbd).

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

First-class values, support closures.

### 7.1. Named Function Definition

```able
fn Identifier[<GenericParamList>] ([ParameterList]) [-> ReturnType] {
  ExpressionList
}
```

*   Uses `fn`. Returns last expression implicitly, or uses explicit `return`.

### 7.2. Anonymous Functions and Closures

Capture lexical environment.

#### 7.2.1. Verbose Anonymous Function Syntax

```able
fn[<GenericParamList>] ([ParameterList]) [-> ReturnType] { ExpressionList }
```

#### 7.2.2. Lambda Expression Syntax

```able
{ [LambdaParameterList] [-> ReturnType] => Expression }
```

*   Zero params: `{ => ... }`.

#### 7.2.3. Closures

Both forms create closures.

### 7.3. Explicit `return` Statement

Provides early exit from a function.

*   **Syntax:** `return Expression` or `return` (for `void` functions).
*   **Semantics:** Immediately terminates function, returning value. See Section [11.1](#111-explicit-return-statement).

### 7.4. Function Invocation

#### 7.4.1. Standard Call

`FunctionName ( ArgumentList )`

#### 7.4.2. Trailing Lambda Syntax

`Function(Args) { Lambda }` or `Function { Lambda }`

#### 7.4.3. Method Call Syntax

`ReceiverExpression . FunctionOrMethodName ( RemainingArgumentList )`
(See Section [9.3](#93-method-call-syntax-resolution-initial-rules) for resolution).

#### 7.4.4. Callable Value Invocation (`Apply` Interface)

If `value` implements `Apply`, `value(args...)` desugars to `value.apply(args...)`.

### 7.5. Partial Function Application

Use `_` as placeholder for arguments. Creates a closure.

```able
add_10 := add(_, 10)
add_five := 5.add
```

### 7.6. Shorthand Notations

#### 7.6.1. Implicit First Argument Access (`#member`)

**Allowed in any function body.**

*   **Syntax:** `#Identifier`
*   **Semantics:** Sugar for `param1.Identifier` where `param1` is the first parameter.

#### 7.6.2. Implicit Self Parameter Definition (`fn #method`)

**Allowed only within `methods` or `impl` blocks.**

*   **Syntax:** `fn #method_name(...) { ... }`
*   **Semantics:** Defines instance method, implicitly adds `self: Self` as first parameter.

## 8. Control Flow

### 8.1. Branching Constructs

Expressions evaluating to a value.

#### 8.1.1. Conditional Chain (`if`/`or`)

*   **Syntax:** `if Cond1 { Blk1 } [or Cond2 { Blk2 }] ... [or { DefBlk }]`
*   **Semantics:** Executes first true block or default `or {}`. Evaluates to block result or `nil`. Types must be compatible (`?Type` if `nil` possible).

#### 8.1.2. Pattern Matching Expression (`match`)

*   **Syntax:** `Subj match { case Pat [if Grd] => Blk, ... [case _ => DefBlk] }`
*   **Semantics:** Evaluates `Subj`, executes block of first matching `case`. Evaluates to block result. Should be exhaustive. Types must be compatible.

### 8.2. Looping Constructs

Loops evaluate to `nil`. Use `breakpoint`/`break` for early exit.

#### 8.2.1. While Loop (`while`)

*   **Syntax:** `while Condition { Body }`

#### 8.2.2. For Loop (`for`)

Iterates over values produced by an expression whose type implements the `Iterable T` interface.

*   **Syntax:** `for Pattern in IterableExpression { BodyExpressionList }`
*   **Semantics:**
    1.  The `IterableExpression` is evaluated. Its type must implement the `Iterable T` interface for some element type `T`.
    2.  The `iterator()` method is called on the result, producing a value that implements `Iterator T`. Let's call this the *iterator instance*.
    3.  The loop begins:
        *   The `next()` method is called on the *iterator instance*.
        *   The result of `next()` is matched:
            *   If it matches the `Pattern` (meaning it's a value of type `T`), the pattern variables are bound, and the `BodyExpressionList` is executed. The loop continues to the next iteration.
            *   If it is the value `IteratorEnd`, the loop terminates.
            *   If it's a value of type `T` but doesn't match the `Pattern`, a runtime error/panic occurs.
    4.  The `for` loop expression always evaluates to `nil`.
    5.  Loop exit occurs when the iterator yields `IteratorEnd` or via a non-local jump (`break`).
*   **Example:**
    ```able
    items := ["a", "b", "c"] ## Assuming Array string implements Iterable string
    for item in items {
        print(item) ## Prints "a", "b", "c"
    }

    total := 0
    for i in 1..3 { ## Assuming Range implements Iterable i32
        total = total + i
    } ## total becomes 6
    ```

#### 8.2.3. Range Expressions

*   **Syntax:** `StartExpr .. EndExpr` (inclusive), `StartExpr ... EndExpr` (exclusive). Creates `Iterable` range objects.

### 8.3. Non-Local Jumps (`breakpoint` / `break`)

Mechanism for early exit from a labeled block, returning a value.

#### 8.3.1. Defining an Exit Point (`breakpoint`)

*   **Syntax:** `breakpoint 'LabelName { ExpressionList }` (Is an expression).

#### 8.3.2. Performing the Jump (`break`)

*   **Syntax:** `break 'LabelName ValueExpression`

#### 8.3.3. Semantics

*   `break` jumps to enclosing `breakpoint 'LabelName`, making it return `ValueExpression`. Unwinds stack. Normal completion returns last expression. Type must be compatible.

## 9. Inherent Methods (`methods`)

Define methods directly associated with a type.

### 9.1. Syntax

```able
methods [Gens] TypeName [Args] { [FunctionDefinitionList] }
```

### 9.2. Method Definitions

Within `methods TypeName`:
1.  **Instance Methods:** `fn name(self: Self, ...)` or `fn #name(...)`.
2.  **Static Methods:** `fn name(...)` (no `self` or `#`).

### 9.3. Method Call Syntax Resolution (Initial Rules)

When resolving `receiver.name(args...)`:
1.  Check for **field** `name`.
2.  Check for **inherent method** `name`.
3.  Check for **interface method** `name` (use specificity).
4.  Check for **free function** `name` (UFCS).
5.  Ambiguity or not found -> error.

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
-   **`[GenericParamList]`**: Optional space-delimited generic parameters for the interface itself (e.g., `T`, `K V`). Constraints use `Param: Bound`.
-   **`for`**: Keyword introducing the self type pattern (mandatory in this form).
-   **`SelfTypePattern`**: Specifies the structure of the type(s) that can implement this interface. This defines the meaning of `Self` (see Section [10.1.3](#1013-self-keyword-interpretation)).
    *   **Concrete Type:** `TypeName [TypeArguments]` (e.g., `for Point`, `for Array i32`).
    *   **Generic Type Variable:** `TypeVariable` (e.g., `for T`).
    *   **Type Constructor (HKT):** `TypeConstructor _ ...` (e.g., `for M _`, `for Map K _`). `_` denotes unbound parameters.
    *   **Generic Type Constructor:** `TypeConstructor TypeVariable ...` (e.g., `for Array T`).
-   **`[where <ConstraintList>]`**: Optional constraints on generic parameters used in `GenericParamList` or `SelfTypePattern`.
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
-   **`[<ImplGenericParams>]`**: Optional comma-separated generics for the implementation itself (e.g., `<T: Numeric>`). Use `<>` delimiters.
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
-   **`[where <ConstraintList>]`**: Optional constraints on `ImplGenericParams`.
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

Combines explicit `return`, Option/Result types with operators, and exceptions.

### 11.1. Explicit `return` Statement

Used for early exit. See Section [7.3](#73-explicit-return-statement).

### 11.2. Option/Result Types and Operators (`?Type`, `!Type`, `!`, `else`)

Preferred for expected errors or absence.

#### 11.2.1. Core Types (`?Type`, `!Type`)

*   **`Option T` (`?Type`)**: Represents optional values. Union `nil | T`.
*   **`Result T` (`!Type`)**: Represents success (`T`) or failure (`Error`).
    *   Requires a standard `Error` interface (TBD). Conceptually: `interface Error for T { fn message(self: Self) -> string; }`
    *   `!Type` is **syntactic sugar** for the union `Type | Error`. (The underlying type might still be referred to as `Result Type Error` in compiler messages/internals).
    *   Example function signatures:
        ```able
        fn find_user(id: u64) -> ?User
        fn read_file(path: string) -> !string
        ```

#### 11.2.2. Propagation (`!`)

Postfix `!` operator unwraps success (`T`) or returns `nil`/`Error` from the current function.

*   **Syntax:** `ExpressionReturningOptionOrResult!`
*   **Semantics:**
    *   If expr is `T` -> result is `T`.
    *   If expr is `nil` or `Error` -> current function returns that `nil` or `Error`.
    *   Containing function must return compatible `?U` or `!U`.
*   **Example:**
    ```able
    fn process_file(path: string) -> !Data {
      content := read_file(path)!  ## Returns Error if read_file fails
      parse_data(content)!       ## Returns Error if parse_data fails
                                 ## Returns Data on success
    }
    ```

#### 11.2.3. Handling (`else {}`)

Handles `nil` or `Error` case immediately.

*   **Syntax:**
    ```able
    ExpressionReturningOptionOrResult else { BlockExpression }
    ExpressionReturningOptionOrResult else { err => BlockExpression } // Capture error
    ```
*   **Semantics:**
    *   If expr is `T` -> result is `T`. `else` block skipped.
    *   If expr is `nil` or `Error`:
        *   `BlockExpression` is executed.
        *   If `err => ...` form used and value is `Error`, `err` is bound within the block.
        *   The entire `... else { ... }` expression evaluates to the result of `BlockExpression`.
    *   **Type Compatibility:** `T` and `BlockExpression` result must be compatible.
*   **Example:**
    ```able
    port := config_lookup("port") else { 8080 } ## Default if nil/Error
    user := find_user(id) else { err =>
        log(`Failed: ${err.message()}`)
        create_guest_user()
    }
    ```

### 11.3. Exceptions (`raise` / `rescue`)

For exceptional conditions. Division/Modulo by zero raises an exception.

#### 11.3.1. Raising Exceptions (`raise`)

*   **Syntax:** `raise ExceptionValue` (Should implement `Error`).

#### 11.3.2. Rescuing Exceptions (`rescue`)

Catches exceptions. `rescue` is an expression.

*   **Syntax:** `MonitoredExpr rescue { case Pat [if Grd] => Blk, ... }`
*   **Semantics:** If `MonitoredExpr` raises, matches exception against `case`. Executes first matching block. Result is normal result or rescue block result. Unmatched exceptions propagate.

#### 11.3.3. Panics

Built-in `panic(message: string)` raises a special `PanicError` (TBD). Avoid rescuing unless for cleanup.

## 12. Concurrency

Lightweight, Go-inspired primitives.

### 12.1. Concurrency Model Overview

*   Supports concurrent execution. Mechanism implementation-defined.
*   Synchronization primitives TBD (stdlib).

### 12.2. Asynchronous Execution (`proc`)

Starts async task, returns `Proc T` handle.

#### 12.2.1. Syntax

`proc FuncCall`, `proc do { ... }`

#### 12.2.2. Semantics

Starts async task, returns `Proc T` immediately. `T` is return type.

#### 12.2.3. Process Handle (`Proc T` Interface)

Conceptual interface: `status()`, `get_value() -> !T`, `cancel()`.

### 12.3. Thunk-Based Asynchronous Execution (`spawn`)

Starts async task, returns `Thunk T`. Evaluation blocks implicitly.

#### 12.3.1. Syntax

`spawn FuncCall`, `spawn do { ... }`

#### 12.3.2. Semantics

Starts async task, returns `Thunk T`. Evaluating `Thunk T` blocks until completion, yields `T` or propagates panic.

### 12.4. Key Differences (`proc` vs `spawn`)

*   `proc` -> Explicit handle (`Proc T`), `spawn` -> Implicit future (`Thunk T`).
*   `Proc T` gives control (status, cancel, error handling).
*   `Thunk T` blocks implicitly on use.

## 13. Packages and Modules

Organizes code.

### 13.1. Package Naming and Structure

*   Tree structure based on `package.yml` name and directories.
*   Qualified names: `root.dir.subdir`. Hyphens in dir -> underscore.

### 13.2. Package Configuration (`package.yml`)

*   Root file defining `name`, `version`, `dependencies`, etc.

### 13.3. Package Declaration in Source Files

*   Optional `package name;` influences qualified name.

### 13.4. Importing Packages (`import`)

*   `import path`, `import path.*`, `import path.{id}`, `import path as alias`, etc.

### 13.5. Visibility and Exports (`private`)

*   Top-level definitions public by default.
*   `private` keyword restricts visibility to current package.

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
