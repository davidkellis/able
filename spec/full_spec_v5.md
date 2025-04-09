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

Able uses `:=` for declaration and initialization, and `=` for reassignment or mutation. Bindings are mutable by default.

### 5.1. Operators (`:=`, `=`)

*   **Declaration (`:=`)**: `Pattern := Expression`
    *   Declares **new** mutable bindings for all identifiers introduced in `Pattern` within the **current** scope.
    *   Initializes these bindings using the corresponding values from `Expression` via matching.
    *   It is a compile-time error if any identifier introduced by `Pattern` already exists as a binding in the *current* scope.
    *   Shadowing: If an identifier introduced by `Pattern` has the same name as a binding in an *outer* scope, the new binding shadows the outer one within the current scope.
*   **Assignment/Mutation (`=`)**: `LHS = Expression`
    *   Assigns the value of `Expression` to existing, accessible, and mutable locations specified by `LHS`.
    *   `LHS` can be:
        *   An identifier (`x = value`) referring to an existing mutable binding.
        *   A mutable field access (`instance.field = value`).
        *   A mutable index access (`array[index] = value`).
        *   A pattern (`{x, y} = point`) where all identifiers/locations within the pattern refer to existing, accessible, and mutable bindings/locations.
    *   It is a compile-time error if `LHS` refers to bindings/locations that do not exist, are not accessible, or are not mutable.

### 5.2. Patterns

Patterns are used on the left-hand side of `:=` and `=` to declare or assign to structured data.

#### 5.2.1. Identifier Pattern

*   **Syntax**: `Identifier`
*   **Usage:**
    *   `x := 10` (Declare and initialize `x`)
    *   `x = 20` (Reassign existing `x`)

#### 5.2.2. Wildcard Pattern (`_`)

Matches any value but binds nothing. Used to ignore parts.

*   **Syntax**: `_`
*   **Usage:** `{ x: _, y } := point`, `_ = side_effect_func()`

#### 5.2.3. Struct Pattern (Named Fields)

*   **Syntax**: `[StructTypeName] { Field1: PatternA, Field2 @ BindingB: PatternC, ShorthandField, ... }`
*   **Usage (`:=`)**: Declares new bindings specified in nested patterns.
    ```able
    Point { x: x_coord, y } := get_point() ## Declares x_coord and y
    ```
*   **Usage (`=`)**: Assigns to existing bindings/locations specified in nested patterns.
    ```able
    { x: existing_x, y: point.y } = new_values ## Assigns to existing_x and point.y
    ```

#### 5.2.4. Struct Pattern (Positional Fields / Named Tuples)

*   **Syntax**: `[StructTypeName] { Pattern0, Pattern1, ... }`
*   **Usage (`:=`)**: Declares new bindings.
    ```able
    IntPair { first, second } := make_pair() ## Declares first and second
    ```
*   **Usage (`=`)**: Assigns to existing bindings/locations.
    ```able
    { var1, arr[0] } = calculation() ## Assigns to existing var1 and arr[0]
    ```

#### 5.2.5. Array Pattern

*   **Syntax**: `[Pattern1, ..., ...RestIdentifier]` or `[Pattern1, ..., ...]`
*   **Usage (`:=`)**: Declares new bindings.
    ```able
    [head, ...tail] := my_list ## Declares head and tail
    ```
*   **Usage (`=`)**: Assigns to existing bindings/locations.
    ```able
    [x, y, _] = three_items ## Assigns to existing x and y
    ```
*   **Mutability:** Array elements are mutable (via index assignment `arr[idx] = val`). Requires `Array` type to support `IndexMut` interface (TBD).

#### 5.2.6. Nested Patterns

Patterns can be nested arbitrarily for both `:=` and `=`.

### 5.3. Semantics of Assignment/Declaration

1.  **Evaluation Order**: `Expression` (RHS) evaluated first.
2.  **Matching & Binding/Assignment**:
    *   **`:=`**: Resulting value matched against `Pattern` (LHS). New mutable bindings created in current scope. Error if identifiers already exist *in current scope*.
    *   **`=`**: Resulting value matched against `LHS` pattern/location specifier. Values assigned to existing, accessible, mutable bindings/locations. Error if target doesn't exist, isn't accessible, or isn't mutable.
    *   Match Failure: If the value's structure doesn't match the pattern, a runtime error/panic occurs for both `:=` and `=`.
3.  **Mutability**: Bindings introduced via `:=` are mutable. `=` requires the target to be mutable.
4.  **Scope**: `:=` introduces bindings into the current lexical scope. `=` operates on existing bindings found according to lexical scoping rules.
5.  **Type Checking**: Compiler checks compatibility between `Expression` type and `Pattern`/`LHS` structure. Inference applies.
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

*   **Syntax:** `for Pattern in IterableExpression { Body }` (Relies on `Iterable`).

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

Define contracts and provide implementations.

### 10.1. Interfaces

Define a contract of function signatures, potentially with default implementations.

#### 10.1.1. Interface Usage Models

1.  **As Constraints:** (`T: Interface`). Compile-time polymorphism.
2.  **As Types (Lens/Existential Type):** (`Array (Display)` or `dyn Display` - TBD). Enables **dynamic dispatch**.

#### 10.1.2. Interface Definition

*   **Syntax:**
    ```able
    interface Name [Gens] for SelfTypePattern [where...] {
      fn signature(...);                 ## Signature only
      fn name_with_default(...) { ... } ## Signature with default body
    }
    ```
    *   `for SelfTypePattern` is mandatory.

#### 10.1.3. `Self` Keyword Interpretation

`Self` refers to the type matching `SelfTypePattern`.

#### 10.1.4. Composite Interfaces (Interface Aliases)

*   **Syntax:** `interface NewName = Iface1 + Iface2 ...`
*   Requires implementing all constituents.

### 10.2. Implementations (`impl`)

Provide concrete function bodies for an interface for a specific target.

#### 10.2.1. Implementation Declaration

*   **Syntax:**
    ```able
    [ImplName =] impl [<ImplGens>] IfaceName [Args] for Target [where...] {
      ## Must provide bodies for signatures without defaults in interface.
      fn required_method(...) { ... }

      ## May override default implementations from interface.
      [fn method_with_default(...) { ... }]
    }
    ```

#### 10.2.2. HKT Implementation Syntax

See previous specification version for details. Example: `impl Mappable A for Array { ... }`.

#### 10.2.3. Overlapping Implementations and Specificity

Compiler uses specificity rules (most specific wins) to resolve `impl`. Ambiguity is error. (See previous spec version for rule details).

### 10.3. Usage

#### 10.3.1. Instance Method Calls

`value.method(...)`. Resolved via Section [9.3](#93-method-call-syntax-resolution-initial-rules).

#### 10.3.2. Static Method Calls

`TypeName.static_method(...)`.

#### 10.3.3. Disambiguation (Named Impls)

Qualify static call: `ImplName.static_method(...)`. Instance call TBD.

#### 10.3.4. Interface Types (Dynamic Dispatch)

Using interface name as type (e.g., `dyn Display` - TBD) allows heterogeneous collections and uses **dynamic dispatch** for method calls.

```able
displayables: Array (dyn Display) := [1, "hello", Point{x:0, y:0}]
for item in displayables {
  print(item.to_string()) ## Dynamic dispatch
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

Required for many features.

*   **Operators:** `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg`, `Not`, `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`.
*   **Comparison:** `PartialEq`, `Eq`, `PartialOrd`, `Ord`.
*   **Functions:** `Apply`.
*   **Collections:** `Iterable`, `Iterator`, `Index`, `IndexMut`.
*   **Display:** `Display`.
*   **Error Handling:** `Error`.
*   **Concurrency:** `Proc`, `ProcError`.
*   **Cloning:** `Clone`.
*   **Hashing:** `Hash`.
*   **Default Values:** `Default`.
*   **Ranges:** `Range`.

### Iterable and Iterator Interfaces

```able
## Marker type indicating end of iteration
struct IteratorEnd;

## Iterator interface: produces a sequence of values of type T
interface Iterator T for Self {
    fn next(self: Self) -> T | IteratorEnd;
}

## Iterable interface: can produce an Iterator over its elements
interface Iterable T for Self {
    fn iterator(self: Self) -> Iterator T;
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
