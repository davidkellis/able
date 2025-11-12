# Able Language Specification (Draft)

**Version:** As of 2023-10-27 conversation (incorporating v2/v3 revisions)
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
    *   [5.1. Core Syntax](#51-core-syntax)
    *   [5.2. Patterns](#52-patterns)
        *   [5.2.1. Identifier Pattern](#521-identifier-pattern)
        *   [5.2.2. Wildcard Pattern (`_`)](#522-wildcard-pattern-_)
        *   [5.2.3. Struct Pattern (Named Fields)](#523-struct-pattern-named-fields)
        *   [5.2.4. Struct Pattern (Positional Fields / Named Tuples)](#524-struct-pattern-positional-fields--named-tuples)
        *   [5.2.5. Array Pattern](#525-array-pattern)
        *   [5.2.6. Nested Patterns](#526-nested-patterns)
    *   [5.3. Semantics](#53-semantics)
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
    *   [11.2. V-Lang Style Error Handling (`Option`/`Result`, `!`, `else`)](#112-v-lang-style-error-handling-optionresult--else)
        *   [11.2.1. Core Types](#1121-core-types)
        *   [11.2.2. Error/Option Propagation (`!`)](#1122-erroroption-propagation-)
        *   [11.2.3. Error/Option Handling (`else {}`)](#1123-erroroption-handling-else-)
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
*   **Operators:** Symbols with specific meanings (See Section [6.3](#63-operators)).
*   **Literals:** Source code representations of fixed values (See Section [4.2](#42-primitive-types) and Section [6.1](#61-literals)).
*   **Comments:** Line comments start with `##` and continue to the end of the line. Block comment syntax is TBD.
    ```able
    x = 1 ## Assign 1 to x (line comment)
    ```
*   **Whitespace:** Spaces, tabs, and form feeds are generally insignificant except for separating tokens.
*   **Newlines:** Significant as expression separators within blocks (See Section [3](#3-syntax-style--blocks)).
*   **Delimiters:** `()`, `{}`, `[]`, `<>`, `,`, `;`, `:`, `->`, `|`, `=`, `.`, `...`, `..`, `_`, `` ` ``, `#`, `?`, `!`, `~`, `=>`, `|>`.

## 3. Syntax Style & Blocks

*   **General Feel:** A blend of ML-family (OCaml, F#) and Rust syntax influences.
*   **Code Blocks `{}`:** Curly braces group sequences of expressions in specific syntactic contexts:
    *   Function bodies (`fn ... { ... }`)
    *   Struct literals (`Point { ... }`)
    *   Lambda literals (`{ ... => ... }`)
    *   Control flow branches (`if ... { ... }`, `match ... { case => ... }`, etc.)
    *   `methods`/`impl` bodies (`methods Type { ... }`)
    *   `do` blocks (`do { ... }`) (See Section [6.2](#62-block-expressions-do))
*   **Expression Separation:** Within blocks, expressions are evaluated sequentially. They are separated by **newlines** or optionally by **semicolons** (`;`). The last expression in a block determines its value unless otherwise specified (e.g., loops).
*   **Expression-Oriented:** Most constructs are expressions evaluating to a value (e.g., `if/or`, `match`, `breakpoint`, `rescue`, `do` blocks, assignments). Loops (`while`, `for`) evaluate to `nil`.

## 4. Types

Able is statically and strongly typed.

### 4.1. Type System Fundamentals

*   **Type:** A named classification representing a set of possible values (e.g., `bool` represents `{true, false}`). Every value has a single, specific type.
*   **Parametric Nature:** All types are conceptually parametric (zero or more type parameters).
*   **Type Expressions:** Syntax for denoting types (type names, space-delimited arguments, parentheses for grouping).
*   **Parameter Binding:** Parameters are **bound** when specific types/variables are provided, otherwise **unbound** (`_` or omitted).
*   **Concrete Type:** All type parameters are bound. Values have concrete types.
*   **Polymorphic Type / Type Constructor:** Has unbound parameters (e.g., `Array`, `Map String _`). Used in HKT contexts.
*   **Constraints:** Restrict generic parameters using interfaces (`T: Interface1 + Interface2`). Applied in definitions using `:` or `where` clauses.

### 4.2. Primitive Types

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

*(See Section [6.1](#61-literals) for detailed literal syntax.)*

### 4.3. Type Expression Syntax Details

*   **Simple:** `i32`, `string`, `MyStruct`
*   **Generic Application:** `Array i32`, `Map string User` (space-delimited arguments)
*   **Grouping:** `Map string (Array i32)` (parentheses control application order)
*   **Function:** `(i32, string) -> bool`
*   **Nullable Shorthand:** `?string` (Syntactic sugar for `nil | string`, see Section [4.6.2](#462-nullable-type-shorthand-))
*   **Wildcard:** `_` denotes an unbound parameter (e.g., `Map string _`).

### 4.4. Reserved Identifier (`_`) in Types

The underscore `_` can be used in type expressions to explicitly denote an unbound type parameter, contributing to forming a polymorphic type / type constructor. Example: `Map string _`.

### 4.5. Structs

Structs aggregate named or positional data fields into a single type. Able supports three kinds of struct definitions: singleton, named-field, and positional-field. A single struct definition must be exclusively one kind. All fields are mutable.

#### 4.5.1. Singleton Structs

Represent types with exactly one value, identical to the type name itself. Useful for simple enumeration variants or tags.

*   **Declaration:** `struct Identifier;` (or `struct Identifier {}`)
*   **Example:** `struct Success; struct Failure;`
*   **Instantiation & Usage:** Use the identifier directly: `status = Success`.
*   **Pattern Matching:** Matched using the identifier: `case Success => ...`.

#### 4.5.2. Structs with Named Fields

Group data under named fields.

*   **Declaration:** `struct Identifier [GenericParamList] { FieldName1: Type1, FieldName2: Type2 ... }`
    *   `[GenericParamList]` is optional, space-delimited (e.g., `<T>`).
*   **Example:** `struct Point { x: f64, y: f64 }`
*   **Instantiation:** Use `{ FieldName: Value, ... }`. Order irrelevant. All fields must be initialized. Field init shorthand `{ name }` is supported (`{ name }` is sugar for `name: name`).
    ```able
    p = Point { x: 1.0, y: 2.0 }
    username = "Alice"
    u = User { id: 101, username, is_active: true } ## Shorthand
    ```
*   **Field Access:** Dot notation: `instance.FieldName`. Example: `p.x`.
*   **Functional Update:** Create a new instance based on others using `...Source`. Later sources/fields override earlier ones.
    ```able
    addr = Address { ...base_addr, zip: "90210" }
    ```
*   **Field Mutation:** Modify fields in-place using assignment: `instance.FieldName = NewValue`. Example: `p.x = 3.0`.

#### 4.5.3. Structs with Positional Fields (Named Tuples)

Define fields by their position and type. Accessed by index.

*   **Declaration:** `struct Identifier [GenericParamList] { Type1, Type2 ... }`
*   **Example:** `struct IntPair { i32, i32 }`
*   **Instantiation:** Use `{ Value1, Value2, ... }`. Values must be provided in the defined order. All fields must be initialized.
    ```able
    pair = IntPair { 10, 20 }
    ```
*   **Field Access:** Dot notation with zero-based integer index: `instance.Index`.
    ```able
    first = pair.0 ## Accesses 10
    ```
    Compile-time error preferred for invalid literal indices. Runtime error otherwise.
*   **Functional Update:** Not supported via `...Source` syntax. Create new instances explicitly.
*   **Field Mutation:** Modify fields in-place using indexed assignment: `instance.Index = NewValue`.
    ```able
    pair.0 = 15
    ```

### 4.6. Union Types (Sum Types / ADTs)

Represent values that can be one of several different types (variants). Essential for modeling alternatives.

#### 4.6.1. Union Declaration

Define a new type as a composition of existing variant types using `|`.

*   **Syntax:** `union UnionTypeName [GenericParamList] = VariantType1 | VariantType2 | ... | VariantTypeN`
    *   `[GenericParamList]` optional space-delimited generics.
    *   Each `VariantType` must be a pre-defined, valid type name.
*   **Example:**
    ```able
    struct Red; struct Green; struct Blue;
    union Color = Red | Green | Blue

    union IntOrString = i32 | string

    union Option T = T | nil ## Common pattern
    ```

#### 4.6.2. Nullable Type Shorthand (`?`)

Provides concise syntax for types that can be either a specific type or `nil`.

*   **Syntax:** `?Type`
*   **Equivalence:** `?Type` is syntactic sugar for the union `nil | Type`.
*   **Example:** `maybe_user: ?User = find_user(id)`

#### 4.6.3. Constructing Union Values

Create a value of the union type by creating a value of one of its variant types.

```able
c: Color = Green
opt_val: ?i32 = 42
opt_nothing: ?i32 = nil
val: IntOrString = 100
```

#### 4.6.4. Using Union Values

The primary way to interact with union values is via `match` expressions (See Section [8.1.2](#812-pattern-matching-expression-match)), which allow safely deconstructing the value based on its current variant.

```able
display_name = maybe_name match {
  case s: string => s, ## Matches non-nil string
  case nil      => "Guest"
}
```

## 5. Bindings, Assignment, and Destructuring

This section defines variable assignment and destructuring in Able. The core mechanism binds identifiers within a pattern to corresponding parts of a value produced by an expression.

### 5.1. Core Syntax

```able
Pattern = Expression
```

-   **`Pattern`**: Specifies the structure to match and identifiers to bind.
-   **`=`**: The assignment operator.
-   **`Expression`**: Any valid Able expression that evaluates to a value.

### 5.2. Patterns

Patterns determine how the value from the `Expression` is deconstructed.

#### 5.2.1. Identifier Pattern

Binds the entire result of the `Expression` to a single identifier.

*   **Syntax**: `Identifier`
*   **Example**: `x = 42`, `user = get_current_user()`

#### 5.2.2. Wildcard Pattern (`_`)

Matches any value but does not bind it. Used to ignore parts of a value.

*   **Syntax**: `_`
*   **Example**: `{ x: _, y } = point`, `[_, second, _] = items`

#### 5.2.3. Struct Pattern (Named Fields)

Destructures instances of structs with named fields.

*   **Syntax**: `[StructTypeName] { Field1: PatternA, Field2 @ BindingB: PatternC, ShorthandField, ... }`
    *   `StructTypeName`: Optional type annotation/check.
    *   `Field: Pattern`: Match field value against nested pattern.
    *   `Field @ Binding: Pattern`: Match field value, bind original field value to `BindingB`.
    *   `ShorthandField`: Equivalent to `ShorthandField: ShorthandField`.
    *   Extra fields in value not mentioned in pattern are ignored.
*   **Example**:
    ```able
    struct Point { x: f64, y: f64 }
    p = Point { x: 1.0, y: 2.0 }
    { x, y: y_coord } = p ## Binds x=1.0, y_coord=2.0
    ```

#### 5.2.4. Struct Pattern (Positional Fields / Named Tuples)

Destructures instances of structs with positional fields.

*   **Syntax**: `[StructTypeName] { Pattern0, Pattern1, ... }`
    *   `StructTypeName`: Optional type annotation/check.
    *   Number of patterns must match the number of fields.
*   **Example**:
    ```able
    struct IntPair { i32, i32 }
    pair = IntPair { 10, 20 }
    { first, _ } = pair ## Binds first=10, ignores second element
    ```

#### 5.2.5. Array Pattern

Destructures instances of the built-in `Array` type.

*   **Syntax**: `[Pattern1, Pattern2, ..., ...RestIdentifier]` or `[Pattern1, ..., ...]`
    *   Patterns match elements from the start.
    *   `...RestIdentifier`: Optional, binds remaining elements as a new `Array`.
    *   `...` without identifier ignores remaining elements.
*   **Example**:
    ```able
    data = [10, 20, 30, 40]
    [first, second, ...rest] = data ## Binds first=10, second=20, rest=[30, 40]
    ```
*   Fails if the array has fewer elements than required by non-rest patterns.

#### 5.2.6. Nested Patterns

Patterns can be nested within struct and array patterns.

*   **Example**:
    ```able
    Container { items: [ Data { id: first_id, point: { x, y: _ }}, ...] } = val
    ```

### 5.3. Semantics

1.  **Evaluation Order**: `Expression` (RHS) evaluated first.
2.  **Matching & Binding**: Resulting value matched against `Pattern` (LHS).
    *   On success, identifiers in `Pattern` are bound in the current scope.
    *   On failure (type/structure mismatch), a runtime error/panic occurs.
3.  **Mutability**: Bindings introduced via `Pattern = Expression` are **mutable** by default. Identifiers can be reassigned using `=`. (Revisitable: `let`/`var` could be added).
4.  **Scope**: Bindings introduced into the current lexical scope.
5.  **Type Checking**: Compiler checks compatibility between `Expression` type and `Pattern` structure. Inference applies.
6.  **Result Value**: The assignment expression (`=`) evaluates to the **value of the RHS** after successful binding.

## 6. Expressions

Units of code that evaluate to a value.

### 6.1. Literals

Source code representations of fixed values.

*   **Integers:** `123`, `0`, `1_000_000`, `42i64`, `255u8`. Prefixes (e.g., `0x`) TBD. Default `i32` (TBC). Suffixes (e.g., `u8`, `i64`) specify type.
*   **Floats:** `3.14`, `0.0`, `-1.2e-5`, `1_000.0`, `2.718f`. Default `f64` (TBC). Suffix `f` for `f32`.
*   **Booleans:** `true`, `false`.
*   **Characters:** `'a'`, `' '`, `'\n'`, `'\u{1F604}'`. Single quotes. Escape sequences (`\n`, `\t`, `\\`, `\'`, `\u{...}`).
*   **Strings:** `"Hello"`, `""`, `` `Val: ${x}` ``. Double quotes or backticks (interpolation). Escape sequences same as char.
*   **Nil:** `nil`. Represents absence of data.
*   **Arrays:** `[1, 2, 3]`, `["a", "b"]`, `[]`. (Requires `Array` type definition in stdlib).
*   **Structs:** `{ field: val }`, `{ val1, val2 }`. (See Section [4.5](#45-structs)).

### 6.2. Block Expressions (`do`)

Execute a sequence of expressions within a new lexical scope.

*   **Syntax:** `do { ExpressionList }`
*   **Semantics:** Expressions evaluated sequentially. Introduces a scope for local bindings. The block evaluates to the value of the last expression.
*   **Example:** `result = do { x = f(); y = g(x); x + y }`

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
| 1          | `=`                   | Simple Assignment                       | Right-to-left | (Evaluates to RHS value)                                  |
| 1          | `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=` | Compound Assignment (TBD)               | Right-to-left | (Needs formal definition)                                 |
| 0          | `\|>`                 | Pipe Forward                            | Left-to-right | (Lowest precedence)                                       |

#### 6.3.2. Operator Semantics

*   Standard arithmetic, logical, bitwise, comparison operations.
*   `[]` for indexing (relies on `Index` interface).
*   `!` is logical NOT, `~` is bitwise NOT.
*   `.` for access (fields, methods, UFCS).
*   `()` for function/method calls.
*   `..`/`...` for range creation (relies on `Range` type/Iterable).
*   `=` assigns (evaluates to RHS value).
*   `&&`/`||` short-circuit on `bool` operands.
*   `|>` pipes (`x |> f` is equivalent to `f(x)`).

#### 6.3.3. Overloading (Via Interfaces)

Behavior for non-primitive types relies on implementing standard library interfaces (e.g., `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg` (for unary `-`), `Not` (for `~`), `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`, `PartialEq`, `Eq`, `PartialOrd`, `Ord`, `Index`, `IndexMut`). Logical `!` is typically not overloaded. See Section [14](#14-standard-library-interfaces-conceptual--tbd).

### 6.4. Function Calls

See Section [7.4](#74-function-invocation).

### 6.5. Control Flow Expressions

`if/or`, `match`, `breakpoint`, `rescue` evaluate to values. See Section [8](#8-control-flow) and Section [11](#11-error-handling). Loops (`while`, `for`) evaluate to `nil`.

### 6.6. String Interpolation

`` `Literal text ${Expression} more text` ``

*   Evaluates embedded `Expression`s (converting results to string, likely via a `Display` interface).
*   Concatenates literal parts and stringified expression results into a final `string`.
*   Escape literal backticks with `` \` ``, escape `${` with `\$`.

## 7. Functions

Functions are first-class values and support closures.

### 7.1. Named Function Definition

```able
fn Identifier[<GenericParamList>] ([ParameterList]) [-> ReturnType] {
  ExpressionList
}
```

*   `Identifier`: Function name.
*   `[<GenericParamList>]`: Optional generics using angle brackets (e.g., `<T: Bound, U>`).
*   `([ParameterList])`: Comma-separated parameters `ident: Type`. Types generally required. Parentheses required. Empty: `()`.
*   `[-> ReturnType]`: Optional return type annotation. Inferred if omitted.
*   `{ ExpressionList }`: Body block. Separated by newlines/semicolons.
*   **Return Value:** Implicitly returns the value of the *last expression*. Use `return` for early exit (See Section [7.3](#73-explicit-return-statement)).

### 7.2. Anonymous Functions and Closures

Capture lexical environment.

#### 7.2.1. Verbose Anonymous Function Syntax

Mirrors named function definition without the identifier.

```able
fn[<GenericParamList>] ([ParameterList]) [-> ReturnType] { ExpressionList }
```

#### 7.2.2. Lambda Expression Syntax

Concise syntax, primarily for single-expression bodies (or `do` blocks).

```able
{ [LambdaParameterList] [-> ReturnType] => Expression }
```

*   `[LambdaParameterList]`: Comma-separated identifiers, optional types (`id: Type`). No parentheses. Zero params: `{ => ... }`.
*   `[-> ReturnType]`: Optional return type.
*   `=>`: Separator.
*   `Expression`: The body expression.

#### 7.2.3. Closures

Both anonymous forms create closures, capturing variables from their definition scope.

```able
fn make_adder(amount: i32) -> (i32 -> i32) {
  { value => value + amount } ## Captures 'amount'
}
```

### 7.3. Explicit `return` Statement

Provides early exit from a function.

*   **Syntax:** `return Expression` or `return` (for `void` functions).
*   **Semantics:** Immediately terminates the current function, returning the value of `Expression` (must match function return type) or `void`. See Section [11.1](#111-explicit-return-statement) for more detail in context of error handling.

### 7.4. Function Invocation

#### 7.4.1. Standard Call

```able
FunctionName ( ArgumentList )
```
*   Arguments are comma-separated. `func()` for no arguments.

#### 7.4.2. Trailing Lambda Syntax

If the last argument is a lambda, it can follow the closing parenthesis. If it's the *only* argument, parentheses can be omitted.

```able
items.map(initial_value) { item => process(item) }
items.each { print(_) }
```

#### 7.4.3. Method Call Syntax

`ReceiverExpression . FunctionOrMethodName ( RemainingArgumentList )`

Used for instance methods, static methods, and provides Uniform Function Call Syntax (UFCS). See Section [9.3](#93-method-call-syntax-resolution-initial-rules) for resolution order.

#### 7.4.4. Callable Value Invocation (`Apply` Interface)

If `value` implements the `Apply` interface (TBD), `value(args...)` desugars to `value.apply(args...)`.

### 7.5. Partial Function Application

Create a new function by providing some arguments and using `_` as a placeholder for others.

*   **Syntax:** Use `_` in place of arguments.
    ```able
    add(5, _)          ## Creates fn(b: i32) -> i32 { add(5, b) }
    5.add              ## Equivalent to add(5, _) via method call syntax access
    process(_, data)   ## Creates fn(cb) { process(cb, data) }
    ```
*   **Semantics:** Creates a closure capturing the provided arguments and the original function. `receiver.method_name` creates a closure capturing `receiver` (equivalent to `method_name(receiver, _, ...)`).

### 7.6. Shorthand Notations

#### 7.6.1. Implicit First Argument Access (`#member`)

**Allowed in any function body (named, anonymous, lambda, method).**

*   **Syntax:** `#Identifier`
*   **Semantics:** Syntactic sugar for `param1.Identifier`, where `param1` is the **first parameter** of the function the `#member` expression appears within. Errors if the function has no parameters. Relies on convention that the first parameter is often the primary context (`self`).
*   **Example:**
    ```able
    fn process_data(d: Data, factor: i32) {
      ## #value is shorthand for d.value
      result = #value * factor
    }
    ```

#### 7.6.2. Implicit Self Parameter Definition (`fn #method`)

**Allowed only when defining functions within a `methods TypeName { ... }` block or an `impl Interface for Type { ... }` block.**

*   **Syntax:** `fn #method_name ([param2: Type2, ...]) [-> ReturnType] { ... }`
*   **Semantics:** Syntactic sugar for defining an **instance method**. Automatically adds `self: Self` as the first parameter. `fn #method(p2) { ... }` is equivalent to `fn method(self: Self, p2) { ... }`. `Self` refers to the type the `methods` or `impl` block is for.
*   **Example:**
    ```able
    impl Counter {
      fn #increment() { #value = #value + 1 } ## #value uses shorthand access too
    }
    ```

## 8. Control Flow

Constructs for conditional execution, looping, and non-local jumps.

### 8.1. Branching Constructs

Both `if/or` and `match` are expressions.

#### 8.1.1. Conditional Chain (`if`/`or`)

Evaluates conditions sequentially, executes block associated with the first true condition. Replaces `if/else if/else`.

*   **Syntax:**
    ```able
    if Cond1 { Blk1 }
    [or Cond2 { Blk2 }]
    ...
    [or { DefaultBlk }] // Final 'or' without condition acts as 'else'
    ```
*   **Semantics:** Evaluates conditions (`bool`) top-down. Executes first true block. Final `or { }` executes if no prior condition was true.
*   **Result Value:** Evaluates to the result of the executed block. Evaluates to `nil` if no block executes (no conditions true and no default `or {}`).
*   **Type Compatibility:** Results of all blocks must be compatible. If execution is not guaranteed (no default `or {}`), the result type is `?CompatibleType`.

#### 8.1.2. Pattern Matching Expression (`match`)

Selects a branch by matching a subject expression against patterns.

*   **Syntax:**
    ```able
    SubjectExpression match {
      case Pattern1 [if Guard1] => ResultExpressionList1,
      case Pattern2 [if Guard2] => ResultExpressionList2,
      ...
      [case _ => DefaultResultExpressionList]
    }
    ```
    *   Clauses separated by commas `,`.
    *   `Pattern`: See Section [5.2](#52-patterns). Bindings are local to the clause.
    *   `[if Guard]`: Optional `bool` guard expression using pattern variables.
*   **Semantics:** `SubjectExpression` evaluated once. Clauses checked top-down. First `Pattern` that matches *and* whose `Guard` (if present) is true selects the clause.
*   **Result Value:** Evaluates to the result of the chosen `ResultExpressionList`.
*   **Exhaustiveness:** Compiler SHOULD check exhaustiveness. Non-exhaustive matches MAY warn/error at compile time and SHOULD panic at runtime. `case _` ensures exhaustiveness.
*   **Type Compatibility:** All `ResultExpressionList` results must have compatible types.

### 8.2. Looping Constructs

Loops evaluate to `nil`. Traditional `break`/`continue` are replaced by `breakpoint`/`break`.

#### 8.2.1. While Loop (`while`)

Repeats body while condition is true.

*   **Syntax:** `while Condition { BodyExpressionList }`
*   **Semantics:** `Condition` (`bool`) checked before each iteration. Body executes if true. Evaluates to `nil`.

#### 8.2.2. For Loop (`for`)

Iterates over values produced by an `Iterable` expression.

*   **Syntax:** `for Pattern in IterableExpression { BodyExpressionList }`
*   **Semantics:** Gets an iterator from `IterableExpression`. For each yielded element, matches it against `Pattern` (bindings available in body), then executes body. Evaluates to `nil`. Relies on `Iterable`/`Iterator` interfaces (TBD).

#### 8.2.3. Range Expressions

Create iterable integer sequences.

*   **Syntax:** `StartExpr .. EndExpr` (inclusive), `StartExpr ... EndExpr` (exclusive).
*   **Semantics:** Creates range objects implementing `Iterable`.

### 8.3. Non-Local Jumps (`breakpoint` / `break`)

Mechanism for early exit from a labeled block, returning a value.

#### 8.3.1. Defining an Exit Point (`breakpoint`)

Marks a block that can be exited early. `breakpoint` is an expression.

*   **Syntax:** `breakpoint 'LabelName { ExpressionList }`
    *   `'LabelName`: Identifier prefixed with `'` uniquely naming this point.

#### 8.3.2. Performing the Jump (`break`)

Initiates early exit targeting a labeled `breakpoint`.

*   **Syntax:** `break 'LabelName ValueExpression`
    *   `'LabelName`: Must match a lexically enclosing `breakpoint`. Compile error if not found.
    *   `ValueExpression`: Result becomes the value of the exited `breakpoint`.

#### 8.3.3. Semantics

1.  `breakpoint` block executes normally. If it completes, its value is the last expression's value.
2.  If `break 'LabelName Value` occurs (possibly in nested calls):
    *   Execution stops immediately.
    *   `ValueExpression` is evaluated.
    *   Stack unwinds to the target `breakpoint 'LabelName`.
    *   The `breakpoint` expression itself evaluates to the result of `ValueExpression`.
3.  **Type Compatibility:** The `breakpoint` expression's type must be compatible with its normal completion value AND any `ValueExpression` types from `break`s targeting it.

## 9. Inherent Methods (`methods`)

Define methods (instance or static) directly associated with a specific struct or other type (TBD if applicable to others). Distinct from implementing interfaces.

### 9.1. Syntax

```able
methods [GenericParams] TypeName [GenericArgs] {
  [FunctionDefinitionList]
}
```

*   **`methods`**: Keyword.
*   **`[GenericParams]`**: Optional generics `<...>` for the block itself (rare).
*   **`TypeName`**: The type name (e.g., `Point`, `User`).
*   **`[GenericArgs]`**: Generic arguments if `TypeName` is generic (e.g., `methods Pair A B { ... }`).
*   **`{ [FunctionDefinitionList] }`**: Contains standard `fn` definitions for methods.

### 9.2. Method Definitions

Within a `methods TypeName { ... }` block:

1.  **Instance Methods:** Operate on an instance (`self`).
    *   Explicit `self`: `fn method_name(self: Self, ...) { ... }` (`Self` refers to `TypeName`).
    *   Shorthand `fn #`: `fn #method_name(...) { ... }` (implicitly adds `self: Self`). See Section [7.6.2](#762-implicit-self-parameter-definition-fn-method).
2.  **Static Methods:** Associated with the type, not an instance.
    *   Defined *without* `self` and *without* the `#` prefix: `fn static_name(...) { ... }`.

### 9.3. Method Call Syntax Resolution (Initial Rules)

When resolving `receiver.name(args...)`:

1.  Check if `name` is a **field** of `receiver`'s type. If yes, treat as field access (unless it's also callable?). *TBD: Precedence if field is callable.* Assume field access takes priority if not callable.
2.  Check for **inherent method** `name` defined in a `methods TypeName { ... }` block for `receiver`'s type.
3.  Check for **interface method** `name` applicable via an `impl Interface for TypeName { ... }` block. Use specificity rules ([Section 10.2.3](#1023-overlapping-implementations-and-specificity)) if multiple interfaces provide `name`.
4.  Check for **free function** `name` where the type of the *first parameter* matches `receiver`'s type (Uniform Function Call Syntax - UFCS). Call becomes `name(receiver, args...)`.
5.  If multiple valid candidates are found at the same step (e.g., conflicting interface methods) and cannot be disambiguated (e.g., by specificity or named impls), it's a compile-time error.
6.  If no candidate is found, it's a compile-time error.

## 10. Interfaces and Implementations

Define contracts (interfaces) and provide concrete implementations (`impl`) for types.

### 10.1. Interfaces

Define a contract of function signatures.

#### 10.1.1. Interface Usage Models

1.  **As Constraints:** Restrict generic parameters (`T: Interface`). Compile-time polymorphism.
2.  **As Types (Lens/Existential Type):** Use interface name as a type (`Array (Display)` or `dyn Display` - syntax TBD) representing any value implementing the interface. Enables dynamic dispatch.

#### 10.1.2. Interface Definition

*   **Syntax:**
    ```able
    interface InterfaceName [GenericParamList] for SelfTypePattern [where <ConstraintList>] {
      [FunctionSignatureList]
    }
    ```
    *   `[GenericParamList]`: Optional interface generics (e.g., `T`, `K V`).
    *   `for SelfTypePattern`: **Mandatory**. Defines structure of implementor.
        *   Concrete: `for Point`, `for Array i32`.
        *   Generic Var: `for T`.
        *   Type Constructor (HKT): `for M _`, `for Map K _`.
        *   Generic App: `for Array T`.
    *   `[FunctionSignatureList]`: `fn name(...) -> ReturnType;`. Can include instance (`self: Self`) or static methods.

#### 10.1.3. `Self` Keyword Interpretation

Within interface and corresponding `impl`: `Self` refers to the type matching `SelfTypePattern`.

*   `for Point`: `Self` is `Point`.
*   `for T`: `Self` is `T`.
*   `for M _`: `Self` is the constructor `M`. `Self A` means `M A`.
*   `for Array T`: `Self` is `Array T`.

#### 10.1.4. Composite Interfaces (Interface Aliases)

Define an interface as a combination of others.

*   **Syntax:** `interface NewName [Gens] = Iface1 [Args] + Iface2 [Args] ...`
*   **Semantics:** Implementing `NewName` requires implementing all constituents.
*   **Example:** `interface DisplayClone for T = Display + Clone` (assuming `Display`, `Clone` are defined `for T`).

### 10.2. Implementations (`impl`)

Provide concrete function bodies for an interface for a specific target.

#### 10.2.1. Implementation Declaration

*   **Syntax:**
    ```able
    [ImplName =] impl [<ImplGenericParams>] InterfaceName [InterfaceArgs] for Target [where <ConstraintList>] {
      [ConcreteFunctionDefinitions]
    }
    ```
    *   `[ImplName =]`: Optional name for disambiguation.
    *   `[<ImplGenericParams>]`: Optional impl generics (e.g., `<T: Numeric>`).
    *   `Target`: The type or type constructor implementing the interface. Must match interface's `SelfTypePattern`.
    *   `[ConcreteFunctionDefinitions]`: Full `fn` definitions matching interface signatures. `fn #method` shorthand can be used for instance methods.

#### 10.2.2. HKT Implementation Syntax

To implement `interface Iface for M _` for `Constructor`:

```able
impl [<ImplGens>] Iface TypeParamName for Constructor [where ...] {
  ## TypeParamName (e.g., A) is bound and usable in method signatures/bodies
  fn method<B>(self: Constructor TypeParamName, ...) -> Constructor B { ... }
}
## Example:
impl Mappable A for Array {
  fn map<B>(self: Array A, f: (A -> B)) -> Array B { ... }
}
```

#### 10.2.3. Overlapping Implementations and Specificity

Compiler uses specificity rules (most specific wins) to resolve which `impl` applies. Ambiguity is a compile error. Rules (simplified):

1.  Concrete (`impl for i32`) > Generic (`impl for T`).
2.  Concrete Generic (`impl for Array T`) > Interface Bound (`impl for T: Iterable`).
3.  Interface Bound (`impl for T: Iterable`) > Unconstrained (`impl for T`).
4.  Subset Union (`impl for i32 | f32`) > Superset Union (`impl for i32 | f32 | string`).
5.  More Constrained (`impl<T: A+B>`) > Less Constrained (`impl<T: A>`).

### 10.3. Usage

#### 10.3.1. Instance Method Calls

Use dot notation: `value.method(...)`. Resolved via Section [9.3](#93-method-call-syntax-resolution-initial-rules).

#### 10.3.2. Static Method Calls

Use `TypeName.static_method(...)`. `TypeName` must have an `impl` for the interface.

```able
zero_int = i32.zero() ## Assuming 'impl Zeroable for i32'
empty_array = Array.zero<f64>() ## Assuming 'impl Zeroable for Array', method is generic
```

#### 10.3.3. Disambiguation (Named Impls)

Qualify call with implementation name for static methods. Method call syntax TBD.

```able
sum_id = Sum.id() ## Assuming 'Sum = impl Monoid for i32'
prod_id = Product.id()
res = Sum.op(5, 6)
## Method call TBD: 5.Sum.op(6)? or requires static style?
```

#### 10.3.4. Interface Types (Dynamic Dispatch)

Using interface name as type (syntax TBD: `(Display)` or `dyn Display`). Method calls use dynamic dispatch.

```able
displayables: Array (Display) = [1, "hello", Point{x:0, y:0}]
for item in displayables {
  print(item.to_string()) ## Dynamic dispatch
}
```

## 11. Error Handling

Combines explicit `return`, V-lang style propagation (`!`/`else`), and exceptions (`raise`/`rescue`).

### 11.1. Explicit `return` Statement

Used for early exit from functions.

*   **Syntax:** `return Expression` or `return` (for `void` functions).
*   **Semantics:** Immediately terminates function, returns value. Value type must match function's return type. `return` alone returns `void`.

### 11.2. V-Lang Style Error Handling (`Option`/`Result`, `!`, `else`)

Preferred for expected errors or absence.

#### 11.2.1. Core Types

*   **`Option T` (`?Type`)**: Union `nil | T`. Represents optional values.
*   **`Result T`**: Union `T | Error`. Represents success (`T`) or failure (`Error`).
    *   Requires a standard `Error` interface (TBD). Conceptually:
        ```able
        interface Error for T { fn message(self: Self) -> string; }
        ```

#### 11.2.2. Error/Option Propagation (`!`)

Postfix `!` operator unwraps success or returns `nil`/`Error` from the current function.

*   **Syntax:** `ExpressionReturningOptionOrResult!`
*   **Semantics:**
    *   If expression yields `T`, result is `T`.
    *   If expression yields `nil` or `Error`, the current function immediately returns that `nil` or `Error`.
    *   Containing function must have a compatible return type (`?U`, `Result U`, or a supertype union).

#### 11.2.3. Error/Option Handling (`else {}`)

Handles `nil` or `Error` case immediately, often providing a default or alternative logic.

*   **Syntax:**
    ```able
    ExpressionReturningOptionOrResult else { BlockExpression }
    ExpressionReturningOptionOrResult else { |err| BlockExpression } // Capture error
    ```
*   **Semantics:**
    *   If expression yields `T`, result is `T`. `else` block skipped.
    *   If expression yields `nil` or `Error`:
        *   `BlockExpression` is executed.
        *   If `|err|` form used and value is `Error`, `err` is bound within the block.
        *   The entire `Expression else { ... }` evaluates to the result of `BlockExpression`.
    *   **Type Compatibility:** The success type `T` and the `BlockExpression` result type must be compatible.

### 11.3. Exceptions (`raise` / `rescue`)

For exceptional conditions disrupting normal flow.

#### 11.3.1. Raising Exceptions (`raise`)

Throws an exception value, unwinding the stack.

*   **Syntax:** `raise ExceptionValue`
*   **Semantics:** `ExceptionValue` (should implement `Error`, but TBD if enforced) is thrown. Execution searches up call stack for `rescue`.

#### 11.3.2. Rescuing Exceptions (`rescue`)

Catches exceptions during monitored execution. `rescue` is an expression.

*   **Syntax:**
    ```able
    MonitoredExpression rescue {
      case Pattern1 [if Guard1] => ResultExpressionList1,
      ...
      [case _ => DefaultResultExpressionList]
    }
    ```
*   **Semantics:**
    *   `MonitoredExpression` is evaluated.
    *   **No Exception:** Result is `MonitoredExpression`'s value. `rescue` block skipped.
    *   **Exception Raised:** Execution stops. Raised value matched against `case` patterns. First matching clause's `ResultExpressionList` is executed. Its result is the value of the `rescue` expression.
    *   Unmatched exceptions propagate up.
    *   **Type Compatibility:** Normal result type of `MonitoredExpression` must be compatible with all `ResultExpressionList` types.

#### 11.3.3. Panics

Built-in `panic(message: string)` raises a special `PanicError` (TBD). Equivalent to `raise PanicError{...}`. Avoid rescuing panics except for cleanup. Unhandled panics typically terminate the program.

## 12. Concurrency

Lightweight, Go-inspired primitives.

### 12.1. Concurrency Model Overview

*   Supports concurrent execution via implementation-defined mechanism (green threads, OS threads, etc.).
*   Guarantees potential for progress.
*   Synchronization primitives (channels, mutexes) TBD (likely in stdlib).

### 12.2. Asynchronous Execution (`proc`)

Starts async task, returns `Proc T` handle immediately.

#### 12.2.1. Syntax

```able
proc FunctionCall
proc BlockExpression // e.g., proc do { ... }
```

#### 12.2.2. Semantics

1.  Starts function/block asynchronously. Current thread does *not* block.
2.  Immediately returns a handle implementing `Proc T` interface. `T` is the return type of the function/block (`void` if none).

#### 12.2.3. Process Handle (`Proc T` Interface)

Conceptual interface for interacting with the async process.

```able
union ProcStatus = Pending | Resolved | Cancelled | Failed ProcError
struct ProcError { details: string } ## Example structure, TBD

interface Proc T for HandleType { ## HandleType is concrete type returned by 'proc'
  fn status(self: Self) -> ProcStatus;
  fn get_value(self: Self) -> Result T ProcError; ## Blocks caller until done
  fn cancel(self: Self) -> void; ## Non-blocking cancellation request
}
```

*   `get_value()`: Blocks caller, returns `Ok(T)` on success, `Err(ProcError)` on failure/cancellation. For `T=void`, returns `Ok(void)` conceptually.
*   `cancel()`: Cooperative cancellation request.

### 12.3. Thunk-Based Asynchronous Execution (`spawn`)

Starts async task, returns `Thunk T`. Evaluation blocks implicitly.

#### 12.3.1. Syntax

```able
spawn FunctionCall
spawn BlockExpression // e.g., spawn do { ... }
```

#### 12.3.2. Semantics

1.  Starts function/block asynchronously. Current thread does *not* block.
2.  Immediately returns `Thunk T` value. `T` is return type (`void` if none).
3.  **Implicit Blocking Evaluation:** When `Thunk T` value is used where `T` is needed, current thread **blocks** until completion.
    *   Success: Yields the result `T`.
    *   Failure (Panic): Propagates the panic to the evaluating thread.
    *   `Thunk void`: Blocks until complete, yields `void`.

### 12.4. Key Differences (`proc` vs `spawn`)

*   **Return:** `proc` -> `Proc T` (handle), `spawn` -> `Thunk T` (implicit future).
*   **Control:** `Proc T` offers explicit status checks, cancellation request, error-handled result retrieval.
*   **Result Access:** `Thunk T` blocks implicitly on evaluation, yielding value or propagating panic directly.
*   **Use Cases:** `proc` for management/fine control. `spawn` for simpler "fire and eventually get result" or side-effect tasks.

## 13. Packages and Modules

Organizes code into reusable libraries and namespaces.

### 13.1. Package Naming and Structure

*   Packages form a tree rooted at the library name (from `package.yml`).
*   Hierarchy follows directory structure.
*   Qualified names use `.` (e.g., `root_name.dir_name.sub_dir_name`).
*   Directory/Package path segments must be valid identifiers (hyphens `-` map to underscores `_`).

### 13.2. Package Configuration (`package.yml`)

Defines project metadata, dependencies, etc.

*   Located in the project root directory.
*   Specifies `name`, `version`, `dependencies`, etc.
    ```yaml
    name: my_cool_package
    version: 0.1.0
    dependencies:
      collections: { github: owner/repo, version: ~>1.0 }
    ```

### 13.3. Package Declaration in Source Files

Optional `package name;` declaration influences the qualified name.

*   If `path/to/file.able` contains `package my_mod;`, its fully qualified name is `root_name.path.to.my_mod`.
*   If `path/to/file.able` has no `package` declaration, its name is `root_name.path.to`.
*   Example Walkthrough: See `packages.md` input file for detailed illustration.

### 13.4. Importing Packages (`import`)

Brings definitions from other packages into scope.

*   **Syntax:**
    *   `import qualified.path.name`
    *   `import qualified.path.name.*` (Wildcard - Use with caution)
    *   `import qualified.path.{Identifier1, Identifier2}` (Selective)
    *   `import qualified.path as Alias` (Alias package)
    *   `import qualified.path.{Identifier as Alias}` (Alias specific identifier)
*   **Scope:** Imports can occur at package scope or any local scope.
*   **Semantics:** Creates new bindings in the importing scope referring to the imported definitions.

### 13.5. Visibility and Exports (`private`)

Controls access to definitions across package boundaries.

*   **Default:** All top-level definitions (functions, types, etc.) are **public** (exported) by default.
*   **`private` Keyword:** Prefixing a top-level definition with `private` restricts its visibility to the **current package** only.
    ```able
    ## Publicly visible
    fn helper(x: i32) -> i32 { x + 1 }

    ## Visible only within the package it's defined in
    private const INTERNAL_CONSTANT = 100;
    ```

## 14. Standard Library Interfaces (Conceptual / TBD)

Many language features rely on interfaces expected to be in the standard library. These require full definition.

*   **Operators:** `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg`, `Not` (Bitwise `~`), `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`.
*   **Comparison:** `PartialEq`, `Eq`, `PartialOrd`, `Ord`.
*   **Functions:** `Apply` (for callable values `value(args)`).
*   **Collections:** `Iterable`, `Iterator` protocol, `Index`, `IndexMut`, potentially `Collection`, `Sequence`, etc.
*   **Display:** `Display` (for `to_string`, used by interpolation).
*   **Error Handling:** `Error` (base interface for errors).
*   **Concurrency:** `Proc`, `ProcError` (details of handle and error).
*   **Cloning:** `Clone`.
*   **Hashing:** `Hash` (for map keys etc.).
*   **Default Values:** `Default`.
*   **Ranges:** `Range` (type returned by `..`/`...`).

## 15. To Be Defined / Refined

This specification is incomplete. Key areas requiring further work include:

*   **Standard Library Implementation:** Full definition and implementation of core types (`Array`, `Map`?, `Set`?, `Range`, `Option`/`Result` concrete structure, `Proc`, `Thunk`), IO, String methods, Math module, `Iterable`/`Iterator` protocol details, Operator interface definitions.
*   **Operator Details:** Formal precedence table verification, Compound assignment (`+=` etc.) semantics, Division/modulo by zero behavior, Integer overflow semantics (wrapping, panic?), Shift semantics/overflows.
*   **Type System Details:** Complete type inference rules, Variance rules (co/contra/invariance), Coercion rules (if any), "Interface type" / dynamic dispatch mechanism details (syntax like `Array (Display)` or `dyn Display`?, vtable layout?), Higher-Kinded Types (HKT) limitations and full capabilities (e.g., type alias support, constraints on HKTs).
*   **Error Handling:** `ProcError` structure, `PanicError` details, Panic unwinding/cleanup behavior (`defer` statement?), Interaction between different `Error` types and exceptions (e.g., can any value be raised?). Binding of `|err|` in `else` block for `nil` case.
*   **Concurrency:** Synchronization primitives (channels, mutexes, atomics?), Cooperative cancellation mechanism details, Scheduler details/guarantees (fairness, thread pool config?).
*   **Mutability:** Final confirmation of "mutable by default" bindings vs. `let`/`var`. `Array` element mutability definition. Need for immutable data structures in stdlib?
*   **FFI:** Mechanism for calling external (e.g., C) code.
*   **Metaprogramming:** Macros? Compile-time evaluation?
*   **Entry Point:** Convention for program entry (`main` function?).
*   **Tooling:** Compiler specifics, Package manager commands (`build`, `test`, `run`), Testing framework conventions.
*   **Lexical Details:** Block comment syntax? Full definitive keyword list? Numeric literal prefixes (`0x`, `0b`, `0o`).
*   **Named Impl Invocation:** Confirm syntax for instance method calls via named impls (e.g., `value.ImplName.method(...)`).
*   **Garbage Collection:** Specify any required GC behavior or user controls (if any).
*   **Field/Method Name Collisions:** Precedence if a callable field and method share a name.
