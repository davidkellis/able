
# Able Language Specification: Control Flow

This section details the constructs Able uses to control the flow of execution, including conditional branching, pattern matching, looping, range expressions, and non-local jumps.

## 1. Branching Constructs

Branching constructs allow choosing different paths of execution based on conditions or patterns. Both `if/or` and `match` are expressions.

### 1.1. Conditional Chain (`if`/`or`)

This construct evaluates conditions sequentially and executes the block associated with the first true condition. It replaces traditional `if/else if/else`.

#### Syntax

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

#### Semantics

1.  **Sequential Evaluation**: Conditions are evaluated strictly in order.
2.  **First True Wins**: Execution stops at the first true `ConditionX`. The corresponding `ExpressionListX` is executed, and its result becomes the value of the `if/or` chain.
3.  **Default Clause**: Executes if no conditions are true and the clause exists.
4.  **Result Value**: The `if/or` chain evaluates to the result of the executed block. If no block executes (no conditions true and no default `or {}`), it evaluates to `nil`.
5.  **Type Compatibility**:
    *   If a default `or {}` guarantees execution, all result expressions must have compatible types. The chain's type is this common type.
    *   If no default `or {}` exists, non-`nil` results must be compatible. The chain's type is `?CompatibleType`.

#### Example

```able
grade = if score >= 90 { "A" }
        or score >= 80 { "B" }
        or { "C or lower" } ## Guarantees String result
```

### 1.2. Pattern Matching Expression (`match`)

Selects a branch by matching a subject expression against a series of patterns, executing the code associated with the first successful match. `match` is an expression.

#### Syntax

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
    *   **`PatternX`**: Pattern to match (Literal, Identifier, `_`, Type/Variant, Struct `{}`, Array `[]`). Bound variables are local to this clause.
    *   **`[if GuardX]`**: Optional `bool` guard expression using pattern variables.
    *   **`=>`**: Separator.
    *   **`ResultExpressionListX`**: Expressions executed if clause chosen; last expression's value is the result.

#### Semantics

1.  **Sequential Evaluation**: `SubjectExpression` evaluated once. `case` clauses checked top-to-bottom.
2.  **First Match Wins**: The first `PatternX` that matches *and* whose `GuardX` (if present) is true selects the clause.
3.  **Execution & Result**: The chosen `ResultExpressionListX` is executed. The `match` expression evaluates to the value of the last expression in that list.
4.  **Exhaustiveness**: Compiler SHOULD check for exhaustiveness (especially for unions). Non-exhaustive matches MAY warn/error at compile time and SHOULD panic at runtime. A `case _ => ...` usually ensures exhaustiveness.
5.  **Type Compatibility**: All `ResultExpressionListX` must yield compatible types. The `match` expression's type is this common type.

#### Example

```able
description = maybe_num match {
  case Some { value: x } if x > 0 => `Positive: ${x}`,
  case Some { value: 0 } => "Zero",
  case Some { value: x } => `Negative: ${x}`,
  case nil => "Nothing"
}
```

## 2. Looping Constructs

Loops execute blocks of code repeatedly. Loop expressions (`while`, `for`) evaluate to `nil`.

### 2.1. While Loop (`while`)

Repeats execution as long as a condition is true.

#### Syntax

```able
while Condition {
  BodyExpressionList
}
```

-   **`while`**: Keyword.
-   **`Condition`**: `bool` expression evaluated before each iteration.
-   **`{ BodyExpressionList }`**: Loop body executed if `Condition` is true.

#### Semantics

-   `Condition` checked. If `true`, body executes. Loop repeats. If `false`, loop terminates.
-   Always evaluates to `nil`.
-   Loop exit occurs when `Condition` is false or via a non-local jump (`break`).

#### Example

```able
counter = 0
while counter < 3 {
  print(counter)
  counter = counter + 1
} ## Prints 0, 1, 2. Result is nil.
```

### 2.2. For Loop (`for`)

Iterates over a sequence produced by an expression whose type implements the `Iterable` interface.

#### Syntax

```able
for Pattern in IterableExpression {
  BodyExpressionList
}
```

-   **`for`**: Keyword.
-   **`Pattern`**: Pattern to bind/deconstruct the current element yielded by the iterator.
-   **`in`**: Keyword.
-   **`IterableExpression`**: Expression evaluating to a value implementing `Iterable`.
-   **`{ BodyExpressionList }`**: Loop body executed for each element. Pattern bindings are available.

#### Semantics

-   The `IterableExpression` produces an iterator (details governed by the `Iterable` interface implementation).
-   The body executes once per element yielded by the iterator, matching the element against `Pattern`.
-   Always evaluates to `nil`.
-   Loop terminates when the iterator is exhausted or via a non-local jump (`break`).

#### Example

```able
items = ["a", "b"] ## Array implements Iterable
for item in items { print(item) } ## Prints "a", "b"

total = 0
for i in 1..3 { ## Range 1..3 implements Iterable
  total = total + i
} ## total becomes 6 (1+2+3)
```

### 2.3. Range Expressions

Provide a concise way to create iterable sequences of integers.

#### Syntax

```able
StartExpr .. EndExpr   // Inclusive range [StartExpr, EndExpr]
StartExpr ... EndExpr  // Exclusive range [StartExpr, EndExpr)
```

-   **`StartExpr`, `EndExpr`**: Integer expressions.
-   **`..` / `...`**: Operators creating range values.

#### Semantics

-   Syntactic sugar for creating values (e.g., via `Range.inclusive`/`Range.exclusive`) that implement the `Iterable` interface (and likely a `Range` interface).

## 3. Non-Local Jumps (`breakpoint` / `break`)

Provides a mechanism for early exit from a designated block (`breakpoint`), returning a value and unwinding the call stack if necessary. Replaces traditional `break`/`continue`.

### 3.1. Defining an Exit Point (`breakpoint`)

Marks a block that can be exited early. `breakpoint` is an expression.

#### Syntax

```able
breakpoint 'LabelName {
  ExpressionList
}
```

-   **`breakpoint`**: Keyword.
-   **`'LabelName`**: A label identifier (single quote prefix) uniquely naming this point within its lexical scope.
-   **`{ ExpressionList }`**: The block of code associated with the breakpoint.

### 3.2. Performing the Jump (`break`)

Initiates an early exit targeting a labeled `breakpoint` block.

#### Syntax

```able
break 'LabelName ValueExpression
```

-   **`break`**: Keyword.
-   **`'LabelName`**: The label identifying the target `breakpoint` block. Must match a lexically enclosing `breakpoint`. Compile error if not found.
-   **`ValueExpression`**: Expression whose result becomes the value of the exited `breakpoint` block.

### 3.3. Semantics

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

### Example

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
