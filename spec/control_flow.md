# Able Language Specification: Control Flow - Branching

This section specifies the branching control flow structures in Able: the `if/or` chain and the `match` pattern matching expression. Both constructs are expressions, meaning they evaluate to a value.

## 1. Conditional Chain (`if`/`or`)

This construct evaluates conditions sequentially and executes the block associated with the first true condition. It replaces traditional `if/else if/else` structures.

### Syntax

```able
if Condition1 { ExpressionList1 }
[or Condition2 { ExpressionList2 }]
...
[or ConditionN { ExpressionListN }]
[or { DefaultExpressionList }] // Final 'or' without condition acts as 'else'
```

-   **`if Condition1 { ExpressionList1 }`**: The required starting clause. `Condition1` (a `bool` expression) is evaluated. If `true`, `ExpressionList1` is executed.
-   **`or ConditionX { ExpressionListX }`**: Optional subsequent clauses. If all preceding `if` or `or` conditions were false, `ConditionX` (a `bool` expression) is evaluated. If `true`, `ExpressionListX` is executed.
-   **`or { DefaultExpressionList }`**: Optional final clause *without* a condition. If all preceding conditions were false, `DefaultExpressionList` is executed.
-   **`ExpressionList`**: A sequence of one or more expressions (separated by newlines or semicolons). The value of the last expression in the list determines the result of the executed block.

### Semantics

1.  **Sequential Evaluation**: Conditions (`Condition1`, `Condition2`, ...) are evaluated strictly in order.
2.  **First True Wins**: Evaluation of conditions and execution stops as soon as a `ConditionX` evaluates to `true`. The corresponding `ExpressionListX` is executed, and its result (the value of its last expression) becomes the value of the entire `if/or` chain.
3.  **Default Clause**: If a final `or { ... }` clause exists and no preceding conditions were true, `DefaultExpressionList` is executed, and its result becomes the value of the chain.
4.  **No Match / Nil Result**: If no conditions are true and there is no final `or { ... }` clause, no block is executed, and the entire `if/or` expression evaluates to `nil`.
5.  **Expression Result**: The `if/or` chain is an expression. Its value is determined by the executed block, or `nil` if no block executes.
6.  **Type Compatibility:**
    *   If a final `or { ... }` clause guarantees a block always executes, all possible result expressions (the last expression in each block) must have compatible types. The type of the `if/or` chain is this common compatible type.
    *   If there is no final `or { ... }` clause, the chain might evaluate to `nil`. The non-`nil` result expressions must still be compatible with each other. The type of the `if/or` chain will be a nullable version of their common type (e.g., `?CompatibleType`).

### Examples

```able
score = 75
grade = if score >= 90 { "A" }
        or score >= 80 { "B" }
        or score >= 70 { "C" }
        or score >= 60 { "D" }
        or { "F" } ## Final 'or' acts as else, guarantees a String result
## grade is "C"

x = 0
y = -5
## Example without a final 'else', result type is ?i32
maybe_code: ?i32 = if x > 0 { log("Positive X"); 1 }
                   or y > 0 { log("Positive Y"); 2 }
## If x=0, y=-5, no condition is true, maybe_code becomes nil.

## Type incompatibility -> error
# error_result = if condition1 { 10 } or { "Default" } ## Error: branches return i32 and String
```

## 2. Pattern Matching Expression (`match`)

Selects a branch by matching a subject expression against a series of patterns, executing the code associated with the first successful match. `match` is an expression.

### Syntax

```able
SubjectExpression match {
  case Pattern1 [if Guard1] => ResultExpressionList1
  [ , case Pattern2 [if Guard2] => ResultExpressionList2 ]
  ...
  [ , case PatternN [if GuardN] => ResultExpressionListN ]
  [ , case _ => DefaultResultExpressionList ] ## Optional wildcard clause
}
```

-   **`SubjectExpression`**: The value to be matched. Evaluated once.
-   **`match`**: Keyword initiating the pattern matching block.
-   **`{ ... }`**: Block containing match clauses.
-   **`case PatternX [if GuardX] => ResultExpressionListX`**: A match clause:
    *   **`case`**: Keyword introducing a pattern clause.
    *   **`PatternX`**: The pattern to match against `SubjectExpression` (Literal, Identifier, Wildcard `_`, Type/Variant, Struct `{}`, Array `[]`). Variables bound in the pattern are local to this clause's guard and result scope.
    *   **`[if GuardX]`**: Optional boolean guard expression, evaluated only if `PatternX` structurally matches. Uses variables bound by `PatternX`.
    *   **`=>`**: Separator.
    *   **`ResultExpressionListX`**: A sequence of one or more expressions (separated by newlines or semicolons). Executed if this clause is chosen. The value of the *last expression* in this list becomes the result of the entire `match` expression.
-   **`,` (Comma Separator)**: Clauses *within* the `match` block are separated by commas. Newlines often serve as visual separators as well, but the comma logically separates the distinct `case` clauses.

### Semantics

1.  **Evaluation Order**: `SubjectExpression` is evaluated once. Clauses (`case ...`) are checked sequentially from top to bottom.
2.  **Matching**: The first `PatternX` that structurally matches the subject value is considered.
3.  **Guards**: If the matching `PatternX` has an `if GuardX`, the guard is evaluated. If it returns `false`, the clause is skipped, and matching continues with the next clause. If the guard returns `true` (or there's no guard), this clause is chosen.
4.  **Execution**: The `ResultExpressionListX` of the chosen clause is executed. Variables bound by `PatternX` are available within this scope.
5.  **Result**: The `match` expression evaluates to the value of the *last expression* in the executed `ResultExpressionListX`.
6.  **Exhaustiveness**:
    *   The compiler SHOULD attempt to check if the patterns cover all possible cases for the type of `SubjectExpression` (especially for union types like `Option` or custom unions).
    *   If the patterns are not exhaustive, the compiler SHOULD issue a warning or error. Including a `case _ => ...` clause usually satisfies exhaustiveness for most types.
    *   If no pattern matches at runtime (due to non-exhaustive patterns), this SHOULD result in a runtime error/panic.
7.  **Type Compatibility**: All `ResultExpressionListX` across all clauses must result in values of a compatible type. The type of the `match` expression is this common compatible type.

### Examples

```able
## Temperature Example
struct F { deg: f32 }
struct C { deg: f32 }
struct K { deg: f32 }
union Temp = F | C | K
value: Temp = F { deg: 70.0 }

temperature_desc = value match {
  case C { deg } => `${deg} Celsius`,
  case F { deg } => `${deg} Fahrenheit`,
  case K { deg } if deg > 0.0 => `${deg} Kelvin`,
  case _ => "Unknown or invalid temperature"
}
## temperature_desc is "70.0 Fahrenheit"

## Option Example
maybe_num: ?i32 = Some i32 { value: 0 }
description = maybe_num match {
  case Some { value: x } if x > 0 => `Positive: ${x}`,
  case Some { value: 0 } => "Zero",
  case Some { value: x } => `Negative: ${x}`,
  case nil => "Nothing" ## Matching the 'nil' part of ?i32
}
## description is "Zero"
```
