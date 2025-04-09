# Assignment

## Part 1: Variable Assignment Semantics

### Syntax
Variable assignment binds identifiers to values, supporting scalars, sequences, and structural values with optional destructuring and renaming. The general form is:

```
Pattern = Expression
```

- **Pattern**: A binding pattern (scalar, sequence, or structural).
- **Expression**: A value-producing expression (literal, variable, or complex expression).
- **`=`**: The assignment operator.

#### Scalar Assignment
- **Syntax**: `Identifier [: Type] = Expression`
  - `Identifier`: A valid variable name (e.g., `x`, `foo`).
  - `[: Type]`: Optional type annotation (see Type Expressions below).
  - **Example**: `x: Int = 42`
- **Semantics**: Binds `Identifier` to the value of `Expression`. If `Type` is specified, `Expression` must evaluate to a value compatible with `Type`.

#### Sequence Destructuring
- **Syntax**: `[PatternList] [: Type] = Expression`
  - `PatternList`: Comma-separated list of sub-patterns:
    - Scalar pattern: `Identifier [: Type]` (e.g., `x: Int`).
    - Rest pattern: `...Identifier [: Type]` (e.g., `...rest: List Int`).
  - **Example**: `[a: Int, b: String, ...rest: Array Int] = [1, "hello", 2, 3]`
- **Semantics**:
  - `Expression` must evaluate to a sequence (e.g., list, array, tuple).
  - Each `Identifier` in `PatternList` binds to the corresponding element in the sequence by position.
  - `...Identifier` binds to the remaining elements as a sequence of the annotated `Type` (or inferred type if omitted).
  - **Length mismatch**: If the sequence is shorter than non-rest patterns, undefined behavior (e.g., error or default values if specified); if longer, excess elements are ignored unless captured by `...`.

#### Structural Destructuring
- **Syntax**: `{MemberList} [: Type] = Expression`
  - `MemberList`: Comma-separated list of member patterns:
    - `Field [@ Identifier] [: Type]`
      - `Field`: A named field in the structure (e.g., `name`, `age`).
      - `@ Identifier`: Optional renaming to a new variable name (e.g., `name @ n`).
      - `[: Type]`: Optional type annotation.
    - **Examples**: `{name: String}`, `{name @ n: String, age @ a: Int}`
  - **Example**: `{name @ n: String, age: Int} = {name: "Alice", age: 30}`
- **Semantics**:
  - `Expression` must evaluate to a structural value (e.g., record, object) with fields matching `Field` names in `MemberList`.
  - For each member:
    - If `@ Identifier` is present, bind `Identifier` to the value of `Field`.
    - If `@` is absent, bind `Field` as the identifier to its value.
    - If `Type` is specified, the field’s value must match `Type`.
  - Unmatched fields in `Expression` are ignored; missing fields in `Pattern` cause undefined behavior (e.g., error or null).

#### Nested Destructuring
- **Syntax**: Patterns can nest arbitrarily, e.g., `[Identifier, {MemberList}]`.
  - **Example**: `[x: Int, {y @ yVal: Float}] = [1, {y: 2.5}]`
- **Semantics**: Recursively apply the rules for sequence or structural destructuring at each level.

#### Default Values (Optional Extension)
- **Syntax**: `Identifier [: Type] = DefaultExpression`
  - **Example**: `[a: Int = 0, b: String = "none"] = [42]`
- **Semantics**: If `Expression` provides no value for `Identifier` (e.g., sequence too short), bind `Identifier` to `DefaultExpression`.

### Behavioral Notes
- **Scope**: Bindings are introduced into the current scope (functional: immutable; imperative: mutable, depending on language rules).
- **Type Checking**: If `Type` is specified, static or runtime checks ensure compatibility. If omitted, type inference applies (if supported).
- **Errors**: Mismatches (e.g., type, structure, or sequence length) result in undefined behavior (e.g., compile-time error, runtime exception).

---

## Part 2: Type Expressions Semantics

### Syntax
Type expressions denote the types of variables and values, supporting simple types, generic types with dual parameterization (space and period), and parenthesized sub-expressions.

#### Simple Types
- **Syntax**: `TypeName`
  - `TypeName`: A basic type identifier (e.g., `Int`, `String`, `Float`).
  - **Example**: `x: Int = 42`
- **Semantics**: Represents a non-parameterized type.

#### Generic Types
- **Syntax**: `TypeName ParameterList`
  - `TypeName`: A generic type constructor (e.g., `List`, `Map`, `Pair`).
  - `ParameterList`: One or more type expressions, delimited by:
    - **Space (` `)**: Loose binding between parameters.
    - **Period (`.`)**: Tight binding between parameters.
  - **Examples**:
    - `Map Int String` (space-delimited).
    - `Pair.Float` (period-delimited).

#### Parenthesized Sub-Expressions
- **Syntax**: `(TypeExpression)`
  - `TypeExpression`: Any valid type expression.
  - **Example**: `Map Int (Pair Float)`

#### Precedence and Binding
- **Rules**:
  - Period (`.`) binds more tightly than space (` `).
  - Parentheses `( )` override natural precedence, grouping sub-expressions explicitly.
- **Examples**:
  - `Map Int Pair.Float`:
    - `.` binds `Pair.Float` first → `Map Int (Pair.Float)`.
    - A map from `Int` to `Pair.Float`.
  - `Map Int Pair Float`:
    - Spaces only, left-to-right → `(Map Int Pair) Float`.
    - If `(Map Int Pair)` is a valid type, applies `Float` to it; otherwise, error.
  - `Map Int (Pair Float)`:
    - Parentheses group `Pair Float` → `Map Int (Pair Float)`.
    - A map from `Int` to `Pair Float`.
  - `List Map.Int String`:
    - `.` binds `Map.Int` → `List (Map.Int) String`.

### Semantics
- **Simple Types**: Denote atomic types with no parameters (e.g., `Int` is the type of integers).
- **Generic Types**:
  - `TypeName` is a type constructor expecting a specific number of parameters (e.g., `Map` takes 2, `List` takes 1).
  - Each parameter in `ParameterList` is itself a type expression, resolved recursively.
  - Space (` `) implies a looser, sequential application of parameters.
  - Period (`.`) implies a tighter, immediate application, grouping adjacent types.
- **Parentheses**: Force a specific grouping, treated as a single type expression.
- **Type Checking**:
  - The number and kinds of parameters must match `TypeName`’s arity and constraints.
  - Example: `Map Int String` is valid (2 parameters); `Map Int` is invalid (missing parameter).

### Behavioral Notes
- **Ambiguity**: If a type expression like `A B C.D` could be parsed multiple ways, precedence resolves it (e.g., `A B (C.D)`). Parentheses disambiguate further.
- **Equivalence**: `Pair.Float` and `Pair Float` may be semantically equivalent (depending on your type system), but `.` affects parsing in complex expressions.
- **Nesting**: Types can nest arbitrarily, e.g., `Array (Map Int (List Float))`.

---

## Combined Example
```plaintext
// Scalar
x: Int = 42

// Sequence with generics
[head: Int, ...tail: List (Pair Float)] = [1, {fst: 2.5, snd: 3.5}]

// Structural with renaming and mixed parameterization
{key @ k: String, value @ v: Map Int Pair.Float} = {key: "data", value: {1: {fst: 2.5, snd: 3.5}}}

// Nested
[x: Int, {y @ yVal: Array (Option Int)}] = [1, {y: [2, 3]}]
```

---

## Specification Notes
- **Flexibility**: The semantics allow both functional (immutable) and imperative (mutable) interpretations, depending on assignment rules.
- **Extensibility**: Default values, wildcards (e.g., `_`), or pattern guards could be added to variable assignment.
- **Type System**: Assumes a static type system; adjust for dynamic typing by relaxing checks.
- **Errors**: Undefined behaviors (e.g., mismatches) can be specified as compile-time errors, runtime exceptions, or custom handling.

---

This Markdown version should be easy to read and share. Let me know if you’d like further adjustments, such as adding more examples or formalizing it with EBNF!
