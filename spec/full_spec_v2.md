
# Able Language Specification (Draft)

**Version:** As of 2023-10-27 conversation (incorporating latest revisions)
**Status:** Incomplete Draft

## 1. Core Philosophy & Goals

*   **Hybrid:** Blends functional and imperative programming paradigms. Aims for functional safety and expressiveness where practical, while allowing familiar imperative style and side effects.
*   **Minimal & Expressive:** Strives for a simple, concise syntax while providing powerful features.
*   **Type System:** Statically and strongly typed with extensive type inference to reduce boilerplate.
*   **Memory Management:** Garbage collected. No manual memory management or Rust-style borrow checker.
*   **Target Audience:** Developers seeking a modern language combining functional features with imperative flexibility.

## 2. Lexical Structure

*   **Character Set:** UTF-8 recommended.
*   **Identifiers:** Rules typically `[a-zA-Z_][a-zA-Z0-9_]*`. Case-sensitive. Package/directory names mapping to identifiers treat hyphens as underscores. The identifier `_` is reserved as a wildcard/placeholder and cannot be used as a regular variable name.
*   **Keywords:** Reserved words including `fn`, `struct`, `union`, `interface`, `impl`, `methods`, `if`, `or`, `while`, `for`, `in`, `match`, `case`, `breakpoint`, `break`, `type`, `package`, `import`, `private`, `nil`, `true`, `false`, `void`, `Self`, `proc`, `spawn`, `raise`, `rescue`, `do`. (More may exist).
*   **Operators:** See Section 11: Operators. Includes arithmetic, comparison, logical, bitwise, assignment, range, member access, indexing, function call, pipe forward.
*   **Literals:** For integers, floats, strings (`"`), characters (`'`), booleans, `nil`, arrays (`[]`), structs (`{}`), interpolated strings (`` ` ``). Underscores allowed in numbers (`1_000`).
*   **Comments:** Line comments start with `##`. Block comments TBD.
    ```able
    x = 1 ## This is a line comment
    ```
*   **Whitespace:** Generally insignificant, used to separate tokens. Newlines act as expression separators in blocks.
*   **Delimiters:** Parentheses `()`, curly braces `{}`, square brackets `[]`, angle brackets `<>`, comma `,`, semicolon `;`, colon `:`, arrow `->`, pipe `|`, equals `=`, dot `.`, ellipsis `...`, underscore `_`, backtick `` ` ``, hash `#`, question mark `?`, exclamation mark `!`, tilde `~`.

## 3. Syntax Style & Blocks

*   **General Feel:** A blend of ML-family (OCaml, F#) and Rust syntax.
*   **Code Blocks:** Curly braces `{}` are used to group sequences of expressions in specific contexts: function bodies, struct/lambda literals, control flow branches (`if`/`or`/`match`/`breakpoint`/`rescue`), `do` blocks.
*   **Expression Separation:** Expressions within blocks are separated by newlines or optionally by semicolons `;`.
*   **Expression-Oriented:** Constructs like `if/or`, `match`, `breakpoint`, `rescue`, and `do` blocks are expressions that evaluate to a value. Loops (`while`, `for`) evaluate to `nil`. Assignment (`=`) evaluates to the RHS value.

## 4. Types

### 4.1. Type System Fundamentals

*   **Value:** Any piece of data that can be computed and manipulated.
*   **Type:** A named classification representing a set of possible values. Every value has a specific type.
*   **Parametric Nature:** All types are conceptually parametric (zero or more type parameters).
*   **Type Expressions:** Syntax used to denote types (Type names, space-delimited arguments, parentheses for grouping).
*   **Parameter Binding:** Parameters are bound when specific types/variables are provided. Unbound parameters use `_` or are omitted.
*   **Concrete Type:** All type parameters are bound. Values have concrete types.
*   **Polymorphic Type / Type Constructor:** Has one or more unbound type parameters (e.g., `Array`, `Map String _`).
*   **Constraints:** Restrict generic parameters using interfaces (`T: Interface1 + Interface2`). Specified after the parameter or in `where` clauses.

### 4.2. Primitive Types

| Type     | Description                                   | Literal Examples                    | Notes                                           |
| :------- | :-------------------------------------------- | :---------------------------------- | :---------------------------------------------- |
| `i8`     | 8-bit signed integer                          | `-128`, `0`, `127`                  |                                                 |
| `i16`    | 16-bit signed integer                         | `-32768`, `1000`, `32767`           |                                                 |
| `i32`    | 32-bit signed integer                         | `-2_147_483_648`, `0`, `42`         | Default type for integer literals (TBC).        |
| `i64`    | 64-bit signed integer                         | `-9_223_...`, `1_000_000_000`       | Underscores `_` allowed for readability.        |
| `i128`   | 128-bit signed integer                        | `-170_...`, `0`, `170_...`          |                                                 |
| `u8`     | 8-bit unsigned integer                        | `0`, `10`, `255`                  |                                                 |
| `u16`    | 16-bit unsigned integer                       | `0`, `1000`, `65535`              |                                                 |
| `u32`    | 32-bit unsigned integer                       | `0`, `4_294_967_295`              | Underscores `_` allowed for readability.        |
| `u64`    | 64-bit unsigned integer                       | `0`, `18_446_...`                 |                                                 |
| `u128`   | 128-bit unsigned integer                      | `0`, `340_...`                    |                                                 |
| `f32`    | 32-bit float (IEEE 754)                       | `3.14f`, `-0.5f`, `1e-10f`, `2.0f`  | Suffix `f` distinguishes from `f64`.              |
| `f64`    | 64-bit float (IEEE 754)                       | `3.14159`, `-0.001`, `1e-10`, `2.0`  | Default type for float literals (TBC).          |
| `string` | Immutable Unicode sequence (UTF-8)            | `"hello"`, `""`, `` `interp ${val}` `` | Double quotes or backticks (interpolation).      |
| `bool`   | Boolean logical values                        | `true`, `false`                     |                                                 |
| `char`   | Single Unicode scalar value (UTF-32)          | `'a'`, `'π'`, `'💡'`, `'\n'`, `'\u{1F604}'` | Single quotes. Escape sequences.                |
| `nil`    | Singleton type for **absence of data**.       | `nil`                               | Type and value are `nil`. Often used with `?Type`. |
| `void`   | Type with **no values** (empty set).          | *(No literal value)*                | For functions/computations producing no data.     |

### 4.3. Type Expressions Details
*   **Simple:** `i32`, `string`
*   **Generic:** `Array i32`, `Map string User` (space-delimited)
*   **Grouping:** `Map string (Array i32)` (parentheses for precedence)
*   **Function:** `(i32, string) -> bool`
*   **Nullable:** `?string` (desugars to `nil | string`)

## 5. Structs

Composite types grouping data. Fields are mutable.

### 5.1. Singleton Structs
Declare `struct Identifier`. Type name is the only value. Used for tags/enums.
```able
struct Red; struct Green; struct Blue;
color = Red
```

### 5.2. Structs with Named Fields
Declare `struct Name [Gens] { field1: Type1, field2: Type2 ... }`.
Instantiate `Name [Args] { field1: val1, field2: val2 }`. Order irrelevant. Shorthand `{ field }` allowed.
Access `instance.field`.
Functional update `Name { ...src1, ...src2, field: new_val }`. Later overrides earlier.
Mutate `instance.field = new_val`.
```able
struct Point { x: f64, y: f64 }
p = Point { x: 1.0, y: 0.0 }
p.x = 2.0
p2 = Point { y: 5.0, ...p } ## p2 is { x: 2.0, y: 5.0 }
```

### 5.3. Structs with Positional Fields (Named Tuples)
Declare `struct Name [Gens] { Type1, Type2 ... }`.
Instantiate `Name [Args] { val1, val2 ... }`. Order matters.
Access `instance.index` (zero-based).
Mutate `instance.index = new_val`. Functional update `...` not supported.
```able
struct IntPair { i32, i32 }
pair = IntPair { 10, 20 }
first = pair.0 ## 10
pair.1 = 30
```

## 6. Union Types (Sum Types / ADTs)

Represent values that can be one of several types (variants).

### 6.1. Declaration
```able
union UnionTypeName [GenericParamList] = VariantType1 | VariantType2 | ...
## Variant types must be pre-defined.
union Color = Red | Green | Blue
union Option T = nil | T ## Or nil | Some T if Some is defined
union Result T E = Ok T | Err E ## Assuming Ok/Err structs
```

### 6.2. Construction & Usage
Create instances of variants. Use `match` to deconstruct.

## 7. Bindings, Assignment, and Destructuring

*   **Assignment:** `Pattern = Expression`. Binds identifiers in `Pattern` to parts of the value from `Expression`. Evaluates to the RHS value.
*   **Patterns:** Identifier (`x`), Wildcard (`_`), Literal, Struct (`{field: P1, ...}` or `{P1, P2}`), Array (`[P1, P2, ...Rest]`), Nested patterns.
*   **Mutability:** Bindings introduced via `=` are currently **mutable by default**. (Revisitable).

```able
{ name, age @ a } = fetch_user()
[head, ...tail] = my_list
```

## 8. Expressions

*   **Operators:** See Section 11. Precedence/associativity defined.
*   **Function Calls:** See Section 9.
*   **Control Flow:** `if/or`, `match`, `breakpoint`, `rescue` are expressions.
*   **Block Expressions:** `do { ExpressionList }` creates a scope, evaluates sequentially, yields the last expression's value.
*   **String Interpolation:** `` `Hello ${name}!` ``.

## 9. Functions

First-class values.

### 9.1. Named Function Definition
```able
fn Identifier[<GenericParams>] ([ParamList]) [-> ReturnType] { ExpressionList }
## ParamList: identifier: Type, ...
## Last expression is implicit return value.
```

### 9.2. Anonymous Functions and Closures
Capture lexical environment.

*   **Verbose Syntax:** `fn[<Gens>] ([Params]) [-> RetType] { Body }`
*   **Lambda Syntax:** `{ [ParamList] [-> RetType] => Expression }`
    *   `ParamList`: `ident: Type, ...` (no parens). Zero params: `{ => ... }`.

### 9.3. `return` Statement
Allows early exit from a function.
```able
return Expression ## Returns value
return           ## Returns void
```

### 9.4. Function Invocation
*   **Standard:** `func(args)`
*   **Trailing Lambda:** `func(other_args) { lambda }` or `func { lambda }`
*   **Method Call Syntax:** `receiver.func(args)` (includes UFCS for free functions).
*   **Callable Values:** `value(args)` (if `value` implements `Apply`).

### 9.5. Partial Function Application
Use `_` as placeholder for arguments. Creates a new function.
```able
add_5 = add(5, _)
add_five = 5.add ## Via Method Call Syntax access
```

### 9.6. Shorthands within Function Definitions
*   **`#member` Access:** In *any* function body, shorthand for `first_param.member`. Error if no params.
*   **`fn #method` Definition:** In `methods` or `impl` blocks *only*, shorthand for instance method `fn method(self: Self, ...)`.

## 10. Inherent Methods (`methods`)

Define instance and static methods directly associated with a type.

### Syntax
```able
methods [Gens] TypeName [Args] {
  fn instance_method(self: Self, ...) { ... }
  fn #shorthand_instance_method(...) { ... } ## Implicit self: Self
  fn static_method(...) { ... } ## No self
}
```
*   Instance methods called via `instance.method(...)`.
*   Static methods called via `TypeName.static_method(...)`.

## 11. Interfaces and Implementations

Define contracts (interfaces) and provide implementations for types.

### 11.1. Interface Definition
Contract defining only function signatures and a required self type structure.
```able
interface Name [Gens] for SelfTypePattern [where...] {
  fn instance_sig(self: Self, ...);
  fn static_sig(...);
}
## SelfTypePattern: ConcreteType, TypeVar T, HKT M _, GenericApp Array T, etc.
## Self keyword refers to SelfTypePattern.
```
*   **Usage:** As constraints (`T: Iface`) or types (`Array (Iface)` - dynamic dispatch TBD).
*   **Composite:** `interface Comp = I1 + I2 ...`

### 11.2. Implementation Definition
Provide concrete method bodies for an interface for a target type/constructor.
```able
[ImplName =] impl [<ImplGens>] InterfaceName [IfaceArgs] for Target [where...] {
  fn instance_method(self: Self, ...) { /* body */ }
  fn #shorthand_instance(...) { /* body */ }
  fn static_method(...) { /* body */ }
}
## HKT Impl: impl IfaceName TypeParam for Constructor ...
```
*   `Target` must match the interface's `SelfTypePattern`.
*   Named impls `ImplName = impl...` allow disambiguation.
*   Specificity rules resolve overlapping implementations.

### 11.3. Invocation
*   Instance/Interface methods: `value.method(...)`.
*   Static interface methods require type context or qualification: `TypeName.static_method(...)` or `ImplName.static_method(...)`. Disambiguation needed.

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

*   Semantics defined as expected (arithmetic, logical, bitwise, comparison).
*   `=` evaluates to RHS value. `&&`, `||` short-circuit.
*   Overloading via standard library interfaces (TBD).

## 13. Control Flow

### 13.1. Branching
*   **`if/or` Chain:**
    ```able
    if Cond1 { Block1 } or Cond2 { Block2 } or { DefaultBlock }
    ```
    Evaluates conditions sequentially, executes first true block. Evaluates to block result or `nil`.
*   **`match` Expression:**
    ```able
    Subject match { case Pat1 [if Guard1] => Block1, case Pat2 => Block2, ... }
    ```
    Evaluates `Subject`, finds first matching `case`, executes block. Evaluates to block result. Should be exhaustive.

### 13.2. Looping
Loops evaluate to `nil`. No `break`/`continue`. Use `breakpoint`.
*   **`while` Loop:** `while Condition { Body }`
*   **`for` Loop:** `for Pattern in IterableExpression { Body }` (uses `Iterable` interface)
*   **Range Expressions:** `start..end` (inclusive), `start...end` (exclusive). Create `Iterable` range objects.

### 13.3. Non-Local Jumps
*   **`breakpoint 'Label { Body }`**: Defines an expression block that can be exited early. Evaluates to last expression or `break` value.
*   **`break 'Label Value`**: Jumps to enclosing `breakpoint 'Label`, making it return `Value`. Unwinds stack.

## 14. Concurrency

Lightweight concurrency via `proc` and `spawn`.

### 14.1. `proc`
`proc FunctionCall` or `proc do { ... }`. Starts async task, returns `Proc T` handle immediately.
*   **`Proc T` Interface:** (Conceptual) `status() -> ProcStatus`, `get_value() -> Result T ProcError` (blocks), `cancel() -> void`. Panics in proc return `Err` from `get_value`.

### 14.2. `spawn`
`spawn FunctionCall` or `spawn do { ... }`. Starts async task, returns `Thunk T` immediately.
*   **`Thunk T` Type:** Evaluating a `Thunk T` value blocks until computation finishes, returning `T` or propagating panics.

## 15. Error Handling

Combines V-lang style, exceptions, and `return`.

### 15.1. `return` Statement
`return Expression` or `return` (for `void`). Early exit from function.

### 15.2. V-Lang Style (`Option`/`Result`, `!`, `else`)
*   **Types:** `?Type` (implicit `nil | T`), `Result T` (implicit `T | Error`). `Error` is an interface.
*   **Propagation (`!`):** `expr!` unwraps success (`T`) or returns `nil`/`Error` from current function. Requires compatible return type.
*   **Handling (`else {}`):** `expr else { DefaultBlock }` or `expr else { |err| ErrorBlock }`. Executes block on `nil`/`Error`, otherwise yields unwrapped `T`. Result type must be compatible.

### 15.3. Exceptions (`raise`/`rescue`)
For exceptional/unrecoverable conditions.
*   **`raise ExceptionValue`**: Throws an exception (should implement `Error`).
*   **`MonitoredExpr rescue { case Pat1 => Block1, ... }`**: Executes `MonitoredExpr`. If exception raised, matches exception value against `case` patterns, executes corresponding block. `rescue` expression evaluates to normal result or rescue block result. Unmatched exceptions propagate.
*   **Panics:** Built-in `panic(msg)` raises a special `PanicError`. `rescue` should generally avoid catching panics.

## 16. Packages and Modules

*   **Config:** `package.yml` (name, version, deps).
*   **Naming:** Based on config name + directory structure (hyphens -> underscores).
*   **Declaration:** Optional `package name;` in files.
*   **Imports:** `import path`, `import path.*`, `import path.{id1, id2}`, `import path as alias`, `import path.{id as alias}`.
*   **Visibility:** Top-level definitions public by default. `private` keyword makes definition package-private.

## 17. String Interpolation

`` `Literal text ${Expression} more text` ``

## 18. To Be Defined / Refined

*   **Standard Library:** Core types (`Array`, `Map`?, `Set`?, `Range`, `Option`/`Result` details, `Error`, `Proc`, `Thunk`), IO, String methods, Math, `Iterable`/`Iterator` protocol, Operator interfaces (`Addable`, `Equatable`, etc.).
*   **Operator Details:** Formal precedence table, compound assignment (`+=`), division/modulo by zero, shift semantics.
*   **Type System Details:** Full inference rules, variance, coercion (if any), "interface type" / dynamic dispatch mechanism details (`Array (Display)`?).
*   **Error Handling:** `ProcError` structure, panic cleanup behavior, interaction between `Error` types and exceptions.
*   **Concurrency:** Synchronization primitives (channels, mutexes?), cooperative cancellation mechanism, scheduler details.
*   **Mutability:** Final confirmation of "mutable by default" or introduction of `let`/`var`. `Array` mutability.
*   **FFI:** Mechanism for calling external code.
*   **Metaprogramming:** Macros?
*   **Entry Point:** `main` function convention?
*   **Tooling:** Compiler specifics, package manager commands, testing.
*   **Lexical Details:** Block comments? Full keyword list? Numeric literal prefixes/suffixes.
*   **Named Impl Invocation:** Confirm syntax like `value.ImplName.method(...)`.
