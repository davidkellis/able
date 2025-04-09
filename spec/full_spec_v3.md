# Able Language Specification (Draft)

**Version:** As of 2023-10-27 conversation
**Status:** Incomplete Draft - Requires Standard Library and further refinement.

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
*   **Identifiers:** Start with a letter (`a-z`, `A-Z`) or underscore (`_`), followed by letters, digits (`0-9`), or underscores. Typically `[a-zA-Z_][a-zA-Z0-9_]*`. Identifiers are case-sensitive. Package/directory names mapping to identifiers treat hyphens (`-`) as underscores. The identifier `_` is reserved (see Section 4.4).
*   **Keywords:** Reserved words that cannot be used as identifiers. Includes: `fn`, `struct`, `union`, `interface`, `impl`, `methods`, `if`, `or`, `while`, `for`, `in`, `match`, `case`, `breakpoint`, `break`, `type`, `package`, `import`, `private`, `nil`, `true`, `false`, `void`, `Self`, `proc`, `spawn`, `raise`, `rescue`, `do`, `else`. (List may be incomplete).
*   **Operators:** Symbols with specific meanings (See Section 12).
*   **Literals:** Source code representations of fixed values (See Sections 4.2, 8.4).
*   **Comments:** Line comments start with `##` and continue to the end of the line. Block comment syntax is TBD.
    ```able
    x = 1 ## Assign 1 to x (line comment)
    ```
*   **Whitespace:** Spaces, tabs, and form feeds are generally insignificant except for separating tokens. Newlines are significant as expression separators within blocks.
*   **Delimiters:** `()`, `{}`, `[]`, `<>`, `,`, `;`, `:`, `->`, `|`, `=`, `.`, `...`, `..`, `_`, `` ` ``, `#`, `?`, `!`, `~`, `=>`, `|>`

## 3. Syntax Style & Blocks

*   **General Feel:** A blend of ML-family (OCaml, F#) and Rust syntax influences.
*   **Code Blocks `{}`:** Curly braces group sequences of expressions in specific syntactic contexts:
    *   Function bodies (`fn ... { ... }`)
    *   Struct literals (`Point { ... }`)
    *   Lambda literals (`{ ... => ... }`)
    *   Control flow branches (`if ... { ... }`, `match ... { case => ... }`, etc.)
    *   `methods`/`impl` bodies (`methods Type { ... }`)
    *   `do` blocks (`do { ... }`) (See Section 8.2)
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
*   **Constraints:** Restrict generic parameters using interfaces (`T: Interface1 + Interface2`).

### 4.2. Primitive Types

| Type     | Description                     | Literal Value(s) | Notes                                  |
| :------- | :------------------------------ | :--------------- | :------------------------------------- |
| `i8`-`i128` | Signed Integers (8-128 bit)   | `123`, `-10`     | Underscores `_` allowed. Suffixes TBD. |
| `u8`-`u128` | Unsigned Integers (8-128 bit) | `123`, `0`       | Underscores `_` allowed. Suffixes TBD. |
| `f32`    | 32-bit float (IEEE 754)         | `3.14f`, `1.0f`  | Suffix `f`.                            |
| `f64`    | 64-bit float (IEEE 754)         | `3.14`, `1.0`    | Default for float literals (TBC).      |
| `string` | Immutable Unicode (UTF-8)       | `"abc"`, `` `a=${v}` `` | Standard or Interpolated.              |
| `bool`   | Boolean                         | `true`, `false`  |                                        |
| `char`   | Unicode Scalar Value (UTF-32)   | `'a'`, `'\n'`    | Single quotes. Escape sequences.       |
| `nil`    | Absence of Data type            | `nil`            | Singleton type. Type and value are `nil`. |
| `void`   | No Value type                   | *(None)*         | Empty set. For side-effecting functions. |

### 4.3. Type Expression Syntax Details
*   **Simple:** `i32`, `string`
*   **Generic Application:** `Array i32`, `Map string User` (space-delimited arguments)
*   **Grouping:** `Map string (Array i32)` (parentheses control application order)
*   **Function:** `(i32, string) -> bool`
*   **Nullable Shorthand:** `?string` (Syntactic sugar for `nil | string`)

### 4.4. Reserved Identifier (`_`) in Types
The underscore `_` can be used in type expressions to explicitly denote an unbound type parameter, contributing to forming a polymorphic type / type constructor. Example: `Map string _`.

## 5. Structs

Composite types grouping data. Fields are mutable. Cannot mix named and positional fields in one struct.

### 5.1. Singleton Structs
Declare `struct Identifier;`. Type name is the only value.
**Usage:** `status = Success`. Matched as `case Success => ...`.

### 5.2. Structs with Named Fields
Declare `struct Name [Gens] { field1: Type1, field2: Type2 ... }`.
Instantiate `Name [Args] { field1: val1, field2: val2 }`. Order irrelevant. Shorthand `{ field }` allowed.
Access `instance.field`.
Functional update `Name { ...src1, ...src2, field: new_val }`. Later overrides earlier.
Mutate `instance.field = new_val`.

### 5.3. Structs with Positional Fields (Named Tuples)
Declare `struct Name [Gens] { Type1, Type2 ... }`.
Instantiate `Name [Args] { val1, val2 ... }`. Order matters.
Access `instance.index` (zero-based). Compile/runtime error on invalid index.
Mutate `instance.index = new_val`. Functional update `...` not supported.

## 6. Union Types (Sum Types / ADTs)

Represent values that can be one of several variant types.

### 6.1. Declaration
```able
union UnionTypeName [GenericParamList] = VariantType1 | VariantType2 | ...
```
*   Variants must be pre-defined types.
*   `GenericParamList` is space-delimited.

### 6.2. Common Unions (Implicit Definitions)
*   **Option:** `?Type` is sugar for `nil | Type`. Represents optional values.
*   **Result:** `Result T` is sugar for `T | Error`. Requires a standard `Error` interface. Represents success (`T`) or failure (`Error`).

### 6.3. Construction & Usage
Create instances of variants. Primarily used with `match` for deconstruction.

## 7. Bindings, Assignment, and Destructuring

*   **Assignment:** `Pattern = Expression`. Binds identifiers in `Pattern` to corresponding parts of the value from `Expression`.
    *   **Semantics:** The `Expression` (RHS) is evaluated first. The resulting value is matched against the `Pattern` (LHS). If successful, bindings occur in the current scope. If the match fails, a runtime error/panic occurs.
    *   **Result Value:** The assignment expression (`=`) evaluates to the **value of the RHS** after successful binding.
*   **Patterns:**
    *   Identifier: `x` (binds the whole value)
    *   Wildcard: `_` (matches anything, binds nothing)
    *   Literal: `10`, `"hello"` (matches specific value)
    *   Struct (Named): `{ field1: P1, field2 @ b: P2, shorthand }`
    *   Struct (Positional): `{ P1, P2, _ }`
    *   Array: `[P1, P2, ...rest, _]`
    *   Nested patterns are allowed.
*   **Mutability:** Bindings introduced via assignment are currently **mutable by default**. Identifiers can be reassigned using `=`. (This decision is revisitable - `let`/`var` could be introduced).

## 8. Expressions

Units of code that evaluate to a value.

### 8.1. Literals
Source code representations of fixed values (numbers, strings, bools, `nil`, arrays, structs). See Section 4.2 and specific type sections.

### 8.2. Block Expressions (`do`)
Execute a sequence of expressions within a new lexical scope.
*   **Syntax:** `do { ExpressionList }`
*   **Semantics:** Expressions evaluated sequentially. The block evaluates to the value of the last expression. Introduces a scope for local bindings.

### 8.3. Operators
Symbols performing operations (See Section 12).

### 8.4. Function Calls
See Section 9.4.

### 8.5. Control Flow Expressions
`if/or`, `match`, `breakpoint`, `rescue` evaluate to values (See Section 13).

### 8.6. String Interpolation
`` `Literal text ${Expression} more text` ``
*   Evaluates embedded `Expression`s (converting results to string via `Display` interface), concatenates parts into a final `string`. Escape `` \` `` and `\$`.

## 9. Functions

First-class values, support closures.

### 9.1. Named Function Definition
```able
fn Identifier[<GenericParams>] ([ParamList]) [-> ReturnType] {
  ExpressionList ## Last expression is implicit return
}
```
*   `GenericParams`: `<T: Bound, U>`
*   `ParamList`: `ident1: Type1, ident2: Type2, ...` (Types generally required)

### 9.2. Anonymous Functions (Closures)
Capture lexical environment.

*   **Verbose:** `fn[<Gens>] ([Params]) [-> RetType] { Body }`
*   **Lambda:** `{ [ParamList] [-> RetType] => Expression }`
    *   `ParamList`: `id1: T1, id2, ...` (No parens). Zero params: `{ => ... }`.

### 9.3. `return` Statement
Provides early exit from a function.
*   `return Expression`: Returns `Expression`'s value. Must match function return type.
*   `return`: Returns `void`. Function must return `void`.

### 9.4. Function Invocation
*   **Standard:** `func(args)`
*   **Trailing Lambda:** `func(args) { lambda }` or `func { lambda }` (if lambda is only arg).
*   **Method Call Syntax:** `receiver.func(args)` (See Section 10.3). Includes UFCS role.
*   **Callable Values:** `value(args)` (if `value` implements `Apply`).

### 9.5. Partial Function Application
Create new functions using `_` as argument placeholders.
*   `add(5, _)` creates `(i32) -> i32`.
*   `5.add` creates `(i32) -> i32` (equivalent to `add(5, _)`).

### 9.6. Shorthands within Function Definitions
*   **`#member` Access:** In *any* function body, `param1.member` shorthand. Error if no params.
*   **`fn #method` Definition:** In `methods` or `impl` blocks *only*, defines instance method with implicit `self: Self` first parameter.

## 10. Inherent Methods (`methods`)

Define methods directly associated with a type.

### 10.1. Syntax
```able
methods [Gens] TypeName [Args] {
  ## Instance Methods:
  fn method1(self: Self, ...) { ... }
  fn #method2(...) { ... } ## Implicit self: Self

  ## Static Methods:
  fn static_method(...) { ... } ## No self or #
}
```

### 10.2. Invocation
*   **Instance:** `instance.method(...)`
*   **Static:** `TypeName.static_method(...)`

### 10.3. Method Call Syntax Resolution
When resolving `receiver.name(args...)`:
1.  Check for field `name`.
2.  Check for inherent method `name` (in `methods TypeName`).
3.  Check for interface method `name` (in `impl Interface for TypeName`). Use specificity rules.
4.  Check for free function `name` where `param1` type matches `receiver` (UFCS role).
If ambiguous or not found, error.

## 11. Interfaces and Implementations

Define and implement contracts (interfaces) containing only function signatures.

### 11.1. Interface Definition
```able
interface Name [Gens] for SelfTypePattern [where...] {
  fn instance_sig(self: Self, ...);
  fn static_sig(...);
}
## SelfTypePattern: Concrete, GenericVar T, HKT M _, GenericApp Array T, etc.
## Self keyword refers to SelfTypePattern.
```
*   **Usage:** As constraints (`T: Iface`) or types (`Array (Iface)` - dynamic dispatch TBD).
*   **Composite:** `interface Comp = I1 + I2 ...`

### 11.2. Implementation Definition
Provide concrete method bodies for an interface for a target.
```able
[ImplName =] impl [<ImplGens>] InterfaceName [IfaceArgs] for Target [where...] {
  fn instance_method(self: Self, ...) { /* body */ }
  fn #shorthand_instance(...) { /* body */ }
  fn static_method(...) { /* body */ }
}
## HKT Impl: impl IfaceName TypeParam for Constructor ...
```
*   `Target` must match the interface's `SelfTypePattern`.
*   Named impls `ImplName = impl...` allow disambiguation via `ImplName.method(...)` or qualified calls.
*   Specificity rules resolve overlapping implementations.

### 11.3. Static Method Invocation
Use `TypeName.static_method(...)` or potentially `ImplName.static_method(...)` for disambiguation.

## 12. Operators

(Precedence follows Rust model)

| Precedence | Operator(s) | Description               | Associativity |
| :--------- | :---------- | :------------------------ | :------------ |
| 15         | `.`         | Member Access             | Left-to-right |
| 14         | `()`        | Function Call             | Left-to-right |
| 14         | `[]`        | Indexing                  | Left-to-right |
| 13         | `-` (unary) | Arithmetic Negation       | Non-assoc     |
| 13         | `!` (unary) | Logical NOT               | Non-assoc     |
| 13         | `~` (unary) | Bitwise NOT               | Non-assoc     |
| 12         | `*`, `/`, `%` | Multiply, Divide, Remainder | Left-to-right |
| 11         | `+`, `-`    | Add, Subtract             | Left-to-right |
| 10         | `<<`, `>>`  | Bitwise Shifts            | Left-to-right |
| 9          | `&` (binary)| Bitwise AND               | Left-to-right |
| 8          | `^`         | Bitwise XOR               | Left-to-right |
| 7          | `|` (binary)| Bitwise OR                | Left-to-right |
| 6          | `>`, `<`, `>=`, `<=` | Ordering Comparison       | Non-assoc     |
| 5          | `==`, `!=`  | Equality Comparison       | Non-assoc     |
| 4          | `&&`        | Logical AND               | Left-to-right |
| 3          | `||`        | Logical OR                | Left-to-right |
| 2          | `..`, `...` | Range Creation            | Non-assoc     |
| 1          | `=`         | Assignment                | Right-to-left |
| 1          | `+=`, etc.  | Compound Assign (TBD)     | Right-to-left |
| 0          | `\|>`       | Pipe Forward              | Left-to-right |

*   **Semantics:** Standard arithmetic, logical, bitwise, comparison. `[]` for indexing. `!` is logical NOT, `~` is bitwise NOT. `.` for access. `()` for calls. `..`/`...` for ranges. `=` assigns (evaluates to RHS). `&&`/`||` short-circuit. `|>` pipes (`x |> f == f(x)`).
*   **Overloading:** Via standard library interfaces (e.g., `Add`, `Eq`, `Ord`, `Index`, `Not` etc. - TBD).

## 13. Control Flow

### 13.1. Branching
*   **`if/or` Chain:** `if Cond1 { Blk1 } or Cond2 { Blk2 } or { DefBlk }`. Evaluates conditions sequentially, executes first true block. Evaluates to block result or `nil`. Types must be compatible (or `?Type` if `nil` possible).
*   **`match` Expression:** `Subj match { case Pat1 [if Grd1] => Blk1, ... }`. Evaluates `Subj`, finds first matching `case`, executes block. Evaluates to block result. Should be exhaustive (compiler check + `_` case). Types must be compatible.

### 13.2. Looping
Loops evaluate to `nil`. No `break`/`continue`. Use `breakpoint`.
*   **`while` Loop:** `while Condition { Body }`. Repeats body while condition is true.
*   **`for` Loop:** `for Pattern in IterableExpression { Body }`. Iterates over `Iterable` values.
*   **Range Expressions:** `start..end` (inclusive), `start...end` (exclusive). Create `Iterable` range objects.

### 13.3. Non-Local Jumps
*   **`breakpoint 'Label { Body }`**: Defines labeled block, evaluates to last expr or `break` value.
*   **`break 'Label Value`**: Jumps to enclosing `breakpoint 'Label`, making it return `Value`. Unwinds stack.

## 14. Concurrency

Lightweight, Go-inspired primitives.

### 14.1. `proc`
`proc FuncCall` or `proc do { ... }`. Starts async task, immediately returns `Proc T` handle.
*   **`Proc T` Interface:** (Conceptual) Provides `status()`, `get_value()` (blocks, returns `Result T ProcError`), `cancel()`. Handles panics within the proc by returning `Err`.

### 14.2. `spawn`
`spawn FuncCall` or `spawn do { ... }`. Starts async task, immediately returns `Thunk T`.
*   **`Thunk T` Type:** Evaluating a `Thunk T` blocks until completion, returns `T` or propagates panics.

## 15. Error Handling

Combines V-lang style, exceptions, and `return`.

### 15.1. `return` Statement
`return Expression` or `return` (for `void`). Early exit from function.

### 15.2. V-Lang Style (`Option`/`Result`, `!`, `else`)
*   **Types:** `?Type` (sugar for `nil | T`), `Result T` (sugar for `T | Error`). `Error` is an interface.
*   **Propagation (`!`):** `expr!` unwraps success (`T`) or returns `nil`/`Error` from current function. Calling function must have compatible return type.
*   **Handling (`else {}`):** `expr else { DefaultBlock }` or `expr else { |err| ErrorBlock }`. Executes block on `nil`/`Error`, otherwise yields unwrapped `T`. Result type must be compatible with `T`.

### 15.3. Exceptions (`raise`/`rescue`)
For exceptional conditions.
*   **`raise ExceptionValue`**: Throws exception (should implement `Error`).
*   **`MonitoredExpr rescue { case Pat1 => Block1, ... }`**: Catches exceptions from `MonitoredExpr`. Matches exception value against `case` patterns. Evaluates to normal result or rescue block result. Unmatched exceptions propagate.
*   **Panics:** Built-in `panic(msg)` raises a special `PanicError`. Avoid rescuing panics except for cleanup.

## 16. Packages and Modules

*   **Config:** `package.yml` (name, version, deps). Directory containing it is root.
*   **Naming:** Based on config name + directory structure (hyphens -> underscores).
*   **Declaration:** Optional `package name;` in files influences qualified name.
*   **Imports:** `import path`, `import path.*`, `import path.{id1, id2}`, `import path as alias`, `import path.{id as alias}`.
*   **Visibility:** Top-level definitions public by default. `private` keyword restricts visibility to the current package.

## 17. Standard Library Interfaces (Conceptual / TBD)

Many language features rely on interfaces expected to be in the standard library:
*   **Operators:** `Add`, `Sub`, `Mul`, `Div`, `Rem`, `Neg`, `Not` (Bitwise `~`), `BitAnd`, `BitOr`, `BitXor`, `Shl`, `Shr`.
*   **Comparison:** `PartialEq`, `Eq`, `PartialOrd`, `Ord`.
*   **Functions:** `Apply`.
*   **Collections:** `Iterable`, `Iterator` protocol, `Index`, `IndexMut`.
*   **Display:** `Display` (for `to_string`, used by interpolation).
*   **Error Handling:** `Error`.
*   **Concurrency:** `Proc`.

## 18. To Be Defined / Refined

*   **Standard Library Implementation:** Core types (`Array`, `Map`?, `Set`?, `Range`, `Option`/`Result` details, `Proc`, `Thunk`), IO, String methods, Math, `Iterable`/`Iterator` protocol, Operator interfaces implementation.
*   **Operator Details:** Formal precedence table verification, compound assignment (`+=`), division/modulo by zero, shift semantics/overflows.
*   **Type System Details:** Full inference rules, variance, coercion (if any), "interface type" / dynamic dispatch mechanism details (`Array (Display)`?), HKT limitations/capabilities.
*   **Error Handling:** `ProcError` structure, `PanicError` details, panic cleanup behavior (`defer`?), interaction between `Error` types and exceptions.
*   **Concurrency:** Synchronization primitives (channels, mutexes?), cooperative cancellation mechanism, scheduler details/guarantees.
*   **Mutability:** Final confirmation of "mutable by default" or introduction of `let`/`var`. `Array` mutability definition.
*   **FFI:** Mechanism for calling external code.
*   **Metaprogramming:** Macros?
*   **Entry Point:** `main` function convention?
*   **Tooling:** Compiler specifics, package manager commands, testing framework.
*   **Lexical Details:** Block comments? Full keyword list? Numeric literal prefixes/suffixes.
*   **Named Impl Invocation:** Confirm syntax like `value.ImplName.method(...)`.
*   **Garbage Collection:** Specify any implications or controls (if any).
