# Able Language Specification: Functions

This section defines the syntax and semantics for function definition, invocation, partial application, and related concepts like closures and anonymous functions in Able. Functions are first-class values.

## 1. Named Function Definition

Defines a function with a specific identifier in the current scope.

### Syntax
```able
fn Identifier[<GenericParamList>] ([ParameterList]) [-> ReturnType] {
  ExpressionList
}
```

-   **`fn`**: Keyword introducing a function definition.
-   **`Identifier`**: The function name (e.g., `add`, `process_data`).
-   **`[<GenericParamList>]`**: Optional space-delimited generic parameters and constraints (e.g., `<T>`, `<T: Display>`). Use `<>` delimiters for the list.
-   **`([ParameterList])`**: Required parentheses enclosing the parameter list.
    -   **`ParameterList`**: Comma-separated list of parameters, each defined as `Identifier: Type` (e.g., `a: i32`, `user: User`). Type annotations are generally required unless future inference rules allow omission.
    -   May be empty: `()`.
-   **`[-> ReturnType]`**: Optional return type annotation. If omitted, the return type is inferred from the body's final expression. If the body's last expression evaluates to `nil` (e.g., assignment, loop), the return type is `Nil`.
-   **`{ ExpressionList }`**: The function body block. Contains one or more expressions separated by newlines or semicolons.
    -   **Return Value**: The value of the *last expression* in `ExpressionList` is implicitly returned. There is no explicit `return` keyword.

### Examples
```able
## Simple function
fn add(a: i32, b: i32) -> i32 { a + b }

## Generic function
fn identity<T>(val: T) -> T { val }

## Function with side effects and inferred Nil return type
fn greet(name: String) {
  message = `Hello, ${name}!`
  print(message) ## Assuming print returns nil
}

## Multi-expression body
fn process(x: i32) -> String {
  y = x * 2
  z = y + 1
  `Result: ${z}` ## Last expression is the return value
}
```

### Semantics
-   Introduces the `Identifier` into the current scope, bound to the defined function value.
-   Parameters are bound to argument values during invocation and are local to the function body scope.
-   The function body executes sequentially.
-   The type of a function is `(ParamType1, ParamType2, ...) -> ReturnType`.

## 2. Anonymous Functions and Closures

Functions can be created without being bound to a name at definition time. They capture their lexical environment (forming closures).

### 2.1. Verbose Anonymous Function Syntax

Mirrors named function definition but omits the identifier. Useful for complex lambdas or when explicit generics are needed.

#### Syntax
```able
fn[<GenericParamList>] ([ParameterList]) [-> ReturnType] { ExpressionList }
```

#### Example
```able
mapper = fn(x: i32) -> String { `Value: ${x}` }
generic_fn = fn<T: Display>(item: T) { print(item.to_string()) }
```

### 2.2. Lambda Expression Syntax

Concise syntax, primarily for single-expression bodies.

#### Syntax
```able
{ [LambdaParameterList] [-> ReturnType] => Expression }
```
-   **`{ ... }`**: Lambda delimiters.
-   **`[LambdaParameterList]`**: Comma-separated identifiers, optional types (`ident: Type`). No parentheses used. Zero parameters represented by empty list before `=>`.
-   **`[-> ReturnType]`**: Optional return type.
-   **`=>`**: Separator.
-   **`Expression`**: Single expression defining the return value.

#### Examples
```able
increment = { x => x + 1 }
adder = { x: i32, y: i32 => x + y }
get_zero = { => 0 }
complex_lambda = { x, y => do { temp = x + y; temp * temp } } ## Using a block expression
```

### 2.3. Closures

Both anonymous function forms create closures. They capture variables from the scope where they are defined. Captured variables are accessed according to the mutability rules of the original binding (currently mutable by default).

```able
fn make_adder(amount: i32) -> (i32 -> i32) {
  adder_lambda = { value => value + amount } ## Captures 'amount'
  return adder_lambda
}
add_5 = make_adder(5)
result = add_5(10) ## result is 15
```

## 3. Function Invocation

### 3.1. Standard Call

Parentheses enclose comma-separated arguments.
```able
Identifier ( ArgumentList )
```
```able
add(5, 3)
identity<String>("hello") ## Explicit generic argument
```

### 3.2. Trailing Lambda Syntax

```able
Function ( [OtherArgs] ) LambdaExpr
Function LambdaExpr ## If lambda is only argument
```

If the last argument is a lambda, it can follow the closing parenthesis. If it's the *only* argument, parentheses can be omitted.
```able
items.reduce(0) { acc, x => acc + x }
items.map { item => item.process() }
```

### 3.3. Method Call Syntax

Allows calling functions (both inherent/interface methods and qualifying free functions) using dot notation on the first argument.

#### Syntax
```able
ReceiverExpression . FunctionOrMethodName ( RemainingArgumentList )
```

#### Semantics
When `receiver.name(args...)` is encountered:

1.  **Field Access:** Check if `name` is a field of `receiver`'s type. If yes, treat as field access.
2.  **Method Resolution:** Search for a function or method named `name` applicable to `receiver`:
    *   Check for **inherent methods** defined in a `methods Type { ... }` block for `receiver`'s type.
    *   Check for **interface methods** defined via `impl Interface for Type { ... }`. Specificity rules apply if multiple interfaces provide `name`. Named implementations can disambiguate (`receiver.ImplName.method(...)` - TBC).
    *   Check for **free functions** (`fn name(...)`) where the type of the *first parameter* matches `receiver`'s type.
3.  **Invocation:**
    *   If a matching inherent or interface method is found, it's called with `receiver` as the implicit `self` instance.
    *   If a matching free function is found, the call is desugared to `name(receiver, args...)`.
4.  **Ambiguity/Error:** If multiple candidates match (e.g., inherent method and free function, or methods from different interfaces without clear specificity) and cannot be disambiguated, it's a compile-time error. If no match is found, it's a compile-time error.

#### Example (Method Call Syntax on Free Function)
```able
fn add(a: i32, b: i32) -> i32 { a + b }
res = 4.add(5) ## Resolved via Method Call Syntax to add(4, 5) -> 9
```

### 3.4. Callable Value Invocation (`Apply` Interface)

If `value` implements the `Apply` interface, `value(args...)` desugars to `value.apply(args...)`.
```able
impl Apply for Integer { fn apply(self: Integer, a: Integer) -> Integer { self * a } }
thirty = 5(6) ## Calls 5.apply(6)
```

## 4. Partial Function Application

Create a new function by providing some arguments and using `_` as a placeholder for others.

### Syntax
Use `_` in place of arguments in a function or method call expression.
```able
function_name(Arg1, _, Arg3, ...)
instance.method_name(_, Arg2, ...)
```

### Syntax & Semantics
-   `function_name(Arg1, _, ...)` creates a closure.
-   `receiver.method_name(_, Arg2, ...)` creates a closure capturing `receiver`.
-   `TypeName::method_name(_, Arg2, ...)` (if static access is needed/allowed) creates a closure expecting `self` as the first argument.
-   `receiver.free_function_name` (using Method Call Syntax access without `()`) creates a closure equivalent to `free_function_name(receiver, _, ...)`.

### Examples
```able
add_10 = add(_, 10)      ## Function expects one arg: add(arg, 10)
result = add_10(5)       ## result is 15

prefix_hello = prepend("Hello, ", _) ## Function expects one arg
msg = prefix_hello("World")          ## msg is "Hello, World"

## method call syntax access creates partially applied function
add_five = 5.add ## Creates function add(5, _) via Method Call Syntax access
result_pa = add_five(20)  ## result_pa is 25
```

## 5. Shorthand Notations

### 5.1. Implicit First Argument Access (`#member`)

Within the body of any function (named, anonymous, lambda, or method), the syntax `#Identifier` provides shorthand access to a field or method of the function's *first parameter*.

#### Syntax
```able
#Identifier
```

#### Semantics
-   Syntactic sugar for `param1.Identifier`, where `param1` is the **first parameter** of the function the `#member` expression appears within.
-   If the function has *no* parameters, using `#member` is a compile-time error.
-   Inside a function `fn func_name(param1: Type1, param2: Type2, ...) { ... }`, an expression `#member` within the function body is syntactic sugar for `param1.member`.
-   This relies on the *convention* that the first parameter often represents the primary object or context (`self`).
-   The `param1` value must have a field or method named `member` accessible via the dot (`.`) operator.
-   This applies regardless of whether the first parameter is explicitly named `self`.

#### Example
```able
struct Data { value: i32, name: string }
impl Data {
    ## Inside an instance method, #value means self.value
    fn display(self: Self) {
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

### 5.2. Implicit Self Parameter Definition (`fn #method`)

**Allowed only when defining functions within a `methods TypeName { ... }` block or an `impl Interface for Type { ... }` block.**

#### Syntax
```able
fn #method_name ([param2: Type2, ...]) [-> ReturnType] { ... }
```

#### Semantics
-   Syntactic sugar for defining an **instance method**. Automatically adds `self: Self` as the first parameter.
-   `fn #method(p2) { ... }` is equivalent to `fn method(self: Self, p2) { ... }`.
-   `Self` refers to the type the `methods` or `impl` block is for.

#### Example
```able
struct Counter { value: i32 }
impl Counter {
  ## Define increment using shorthand
  fn #increment() {
    #value = #value + 1 ## #value means self.value
  }

  ## Equivalent explicit definition:
  # fn increment(self: Self) {
  #  self.value = self.value + 1
  # }

  ## Define add using shorthand
  fn #add(amount: i32) {
    #value = #value + amount
  }
}

c = Counter { value: 5 }
c.increment() ## c.value becomes 6
c.add(10)     ## c.value becomes 16
```
