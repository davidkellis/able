# Functions Definitions

Functions may be defined with one of three different function definition syntaxes:

1. Named function syntax
2. Anonymous function syntax
3. Lambda expression syntax

Function types are denoted with the syntax:

```
(<parameter type>, <parameter type>, <parameter type>, ...) -> <return type>
```

As a special case, if a function only has a single parameter, the function type may be denoted with:

```
<parameter type> -> <return type>
```

## Named function syntax

`fn <function name>[<optional type paramter list>](<parameter list>) -> <optional return type> { <function body> }`

This style allows type constraints to be captured in the type parameter list (e.g. `fn convert_to_string [T: Parsable & Stringable] (a: T) { to_string(a) }` )

Partial application is documented in the section [Partial application](#partial-application), and placeholder lambda expressions are covered in the section [Implicit Lambda Expressions via Placeholder Expression Syntax (a.k.a. Placeholder Lambdas)](#implicit-lambda-expressions-via-placeholder-expression-syntax-aka-placeholder-lambdas).

For example:

```
# puts is an alias for the println function
fn puts = println

# given:
fn add(a: i32, b: i32) -> i32 { a + b }
fn fold_right(it: Iterator T, initial: V, accumulate: (T, V) -> V) -> V {
  it.reverse.fold_left(initial) { acc, val => accumulate(val, acc) }
}

# add1, add5, and add10 are partial applications of the add function
# sum is the partially applied fold_right function
fn add1 = add(1)
fn add5 = add(5,)
fn add10 = add(, 5)
fn sum = fold_right(, 0, +)
```

## Anonymous function syntax

`fn[<optional type paramter list>](<parameter list>) -> <optional return type> { <function body> }`

When the optional return type is omitted, then the `->` return type delimiter that immediately follows the parameter list and preceeds the function body should also be omitted.

## Lambda expression syntax

Lambda expressions are defined in the lambda expression syntax:
`{ <paramter list> -> <return type> => expression }`

The parameter list and the return type declarations are optional.
When the optional return type is omitted, then the `->` return type delimiter that immediately follows the parameter list and preceeds the function body should also be omitted.

## Variadic functions

```
## Given:
fn puts(...vals: Array T)

## Use:
puts()
puts("Greetings!")
puts("Hello!", "Goodbye!")
puts("Hello!", ...["A", "B", "C"], "Goodbye!")
```

### Variadic disambiguation rule

If two functions are defined, one variadic and one non-variadic, and their function signatures conflict/overlap,
then an invocation of the ambiguous signature will dipatch to the non-variadic function.

For example:

```
## Given:
fn log(label: string)
fn log(label: string, val: Any)
fn log(label: string, ...vals: Array Any)

## Use:
log("foo")                    ## ambiguous; invokes non-variadic signature, log(String)
log("foo", 123)               ## ambiguous; invokes non-variadic signature, log(String, Any)
log("foo", 123, "bar")        ## not ambiguous; invokes variadic signature, log(String, Array Any)
```

## Function definition examples

### Define non-generic functions

With explicit return type:

1. `fn createPair(a, b: i32) -> Pair i32 { Pair{a, b} }`
2. `createPair = fn(a, b: i32) -> Pair i32 { Pair{a, b} }`
3. `createPair = { a, b: i32 -> Pair i32 => Pair{a, b} }`

With inferred return type:

1. `fn createPair(a, b: i32) { Pair{a, b} }`
2. `createPair = fn(a, b: i32) { Pair{a, b} }`
3. `createPair = { a, b: i32 => Pair{a, b} }`

All the functions defined above are of type:
`(i32, i32) -> Pair i32`

### Define generic functions

With explicit free type parameter and explicit return type:

1. fn createPair[T](a, b: T) -> Pair T { Pair{a, b} }
2. createPair = fn[T](a, b: T) -> Pair T { Pair{a, b} }

With explicit free type parameter and inferred return type:

1. fn createPair[T](a, b: T) { Pair{a, b} }
2. createPair = fn[T](a, b: T) { Pair{a, b} }

With implied free type parameter and explicit return type:

1. createPair = (a, b: T) -> Pair T { Pair{a, b} }
2. createPair = { a, b: T -> Pair T => Pair{a, b} }

With implied free type parameter and inferred return type:

1. createPair = (a, b: T) { Pair{a, b} }
2. createPair = { a, b: T => Pair{a, b} }

All the functions defined above are of type:
`(T, T) -> Pair T`

### Lambda Expressions

```
createPair = { a, b: i32 -> Pair i32 => Pair{ a, b } }
createPair = { a, b: i32 => Pair{ a, b } }
createPair = { a, b: T -> Pair T => Pair{ a, b } }
createPair = { a, b: T => Pair{ a, b } }
```


# Function Application

There are several ways to invoke a function:
- positional argument application
- named argument application
- partial application
- method application
- pipeline application
- operator application

## Positional argument application

Functions may be invoked with the normal function application syntax `<function object>(<arg>, <arg>, ...)`, as in:

```
sum(Array(1,2,3))
f(1,2,...xs)
buildPair(1,2)
map([1,2,3], square)
```

## Named argument application

Given the following function definition:
```
fn create_person(age: i32, name: String, address: Address) -> Person { ... }
```

the `create_person` function may be invoked as follows:
```
create_person(@name: "Bob", @age: 20, @address: Address("123 Wide Street"))
```

## Partial application

Given:
```
f = fn(a: i32, b: i32, c: i32, d: i32, e: i32) => (a,b,c,d,e)
```

Partial application of the trailing arguments:
```
f(,1,2,3)
```

Explicit partial application of the leading arguments:
```
f(1,2,3,)
```

Partial application of the first, third, and fifth arguments; returns a function with the signature (i32, i32) -> (i32, i32, i32, i32, i32)
```
f(1, #, 2, #, 3)
```

Partial application of the first, third, and fifth arguments; returns a function with the signature (i32) -> (i32, i32, i32, i32, i32)
```
f(1, #1, 2, #1, 3)
```

Partial application of all the arguments; returns a function with the signature () -> (i32, i32, i32, i32, i32)
```
f(1, 2, 3, 4, 5,)
or
f(,1, 2, 3, 4, 5)
```

Partial application of a named argument

```
f(@c: 30)
```

## Method application

`obj.method(1, 2, 3)`
is equivalent to
`method(obj, 1, 2, 3)`

`obj.m1(1, 2, 3).m2(4, 5, 6)`
is equivalent to
`m2(m1(obj, 1, 2, 3), 4, 5, 6)`

## Pipeline application

Given:
`struct Coord {i32, i32}`

`5 |> Coord {_1, _1}`
is equivalent to
`Coord{5, 5}`

`1 |> plus_two |> f(1, 2, _, 4, 5)`
is equivalent to
`f(1, 2, plus_two(1), 4, 5)`

## Operator application

Operators are functions too. The only distinction between operator functions and non-operator functions is that operator functions additionally have an associated operator precedence and are associative. Non-operator functions are non-associative, and have no operator precedence.

By default, operators are left-associative. For example, `a + b + c` is evaluated as `((a + b) + c)`

Operators that have a trailing apostrophe are right-associative. For example, `a +' b +' c` is evaluated as `(a +' (b +' c))`

### Right-associative operators

1. Exponentiation operator: `**`
   The exponentiation operator is a right-associative operator with precedence that is higher than addition, subtraction, multiplication, and division.

### Special operators

Some operators have special semantics:

1. Assignment operator (infix), e.g. `a = 5`:
   The assignment operator is used in assignment expressions to assign the value expressed in the right hand side (RHS) of the assignment expression to the identifier or identifiers specified in the left hand side (LHS) of the assignment expression. The assignment operator is right-associative. The assignment operator is not a function, and as a result, is not a first-class object that may be used as a value. The assignment operator may not be redefined or overridden.

   An assignment expression evaluates to the value on the RHS of the assignment operator.

2. Function application operator (suffix), e.g. `array(5)`:
   The function application operator is used to apply a function to a given set of arguments. The function application operator may be defined on an arbitrary type in order to be able to treat that type as a function of some set of arguments. For example, an array, a map, or a string value may be viewed as a function of an index argument that retrieves the given array element, map value, or character within an array, map, or string, respectively.

   In short, a value of type T may be treated as a function, and invoked with normal function application syntax, if a function named `apply`, defined with a first parameter of type T, is in scope at the call site. [Special application rule #5](#special-application-rule-5) explains further.

3. Index update expression, e.g. `array(5) = b`:
   An index update expression - an assignment expression with a function application expression on the left-hand-side (LHS) - is interpreted as an assignment to an index position of some structure. Assigning a value to an index position of an array, or assigning a value to a key in a map are the two most common uses of this facility.

   The index update expression is syntactic sugar for calling the `update` function where the first argument is the structure value being updated, the next arguments are the components of the index position (most commonly, this will be a single value), and the last argument is the value being assigned to the specified index of the structure. The `update` function is expected to return the updated structure, so its return type must match the type of the first argument.

   For example, the `update` function for the `Array` type is defined as:

   ```
   fn update(a: Array T, index: u64) -> Array T
   ```

   and may be used as:

   ```
   a = Array(1,2,3)
   a(0) = 5
   # a == Array(5,2,3)
   ```

## Special application rules

1. When invoking a function of arity-1 with method syntax, the empty parenthesis may be omitted.

   For example, either of the following are acceptable:

   - `Array(1,2,3).length()`
   - `Array(1,2,3).length`

   This only works with functions of arity-1.

   Attempting to invoke a function of arity-2+ using method syntax without supplying any parenthesis will yield a partially applied function.

   For example, given `fn add(a: i32, b: i32) -> i32 { a + b }`, the expression `1.add` represents a partially applied function object of type (i32)->i32.

2. If supplying a lambda expression to a function as its last argument, then the lambda may be supplied immediately following the closing paren surrounding the argument list.

   For example, rather than writing:

   ```
   Array(1,2,3).reduce(0, { sum, x => sum + x})
   ```

   you may write

   ```
   ​Array(1,2,3).reduce(0) { sum, x => sum + x }
   ```

3. In cases where a lambda expression is supplied immediately following the closing paren of a function call, if the argument list is empty, then the parenthesis may be omitted.

   For example,

   ```
   ​Array(1,2,3).map() { x => x**2 }
   ```

   may be shortened to

   ```
   Array(1,2,3).map { x => x**2 }
   ```

4. Functions of unions have the following invocation behavior:

   Given the following definitions:

   ```
   struct Red { level: i32 }
   struct Green { level: i32 }
   struct Blue { level: i32 }
   union Color = Red | Green | Blue
   ```

   1. If the following is defined:

      ```
      fn paint_house(Color) -> bool
      ```

      and the following are not defined:

      ```
      fn paint_house(Red) -> bool
      fn paint_house(Green) -> bool
      fn paint_house(Blue) -> bool
      ```

      then

      ```
      c: Color = Red{}
      paint_house(c)           # invokes paint_house(Color)
      paint_house(Red{})       # invokes paint_house(Color)
      paint_house(Green{})     # invokes paint_house(Color)
      paint_house(Blue{})      # invokes paint_house(Color)
      ```

   2. If the following are defined:

      ```
      fn paint_house(Color) -> bool
      fn paint_house(Red) -> bool
      fn paint_house(Green) -> bool
      fn paint_house(Blue) -> bool
      ```

      then

      ```
      c: Color = Red{}
      paint_house(c)           # invokes paint_house(Color)
      paint_house(Red{})       # invokes paint_house(Red)
      paint_house(Green{})     # invokes paint_house(Green)
      paint_house(Blue{})      # invokes paint_house(Blue)
      ```

   3. If the following are defined:
      ```
      fn paint_house(Red) -> bool
      fn paint_house(Green) -> bool
      fn paint_house(Blue) -> bool
      ```
      and the following is not defined:
      ```
      fn paint_house(Color) -> bool
      ```
      then
      ```
      c: Color = Red{}
      paint_house(c)           # is not a valid function invocation
      paint_house(Red{})       # invokes paint_house(Red)
      paint_house(Green{})     # invokes paint_house(Green)
      paint_house(Blue{})      # invokes paint_house(Blue)
      ```

5. <a name="special-application-rule-5"></a>A value of type T may be treated as a function, and invoked with normal function application syntax, if a function named `apply`, defined with a first parameter of type T, is in scope at the call site.

   For example:

   ```
   fn apply(key: String, map: Map String V) -> Option V => map.get(key)
   name_rank_pairs = Map("joe" : 1, "bob" : 2, "tom" : 3)
   tom_rank = "tom"(name_rank_pairs)   # returns Some(3)
   ```

   Another example:

   ```
   fn apply(i: i32, j: i32) -> i32 => i * j
   5(6)    # returns 30
   ```

   Apply functions may also be invoked with method syntax, as in the following:

   ```
   fn apply(i: i32, j: i32) -> i32 => i * j
   5.apply(6)    # returns 30
   ```
6. A value of type T may be treated as a function, and invoked with normal function application syntax, if the type T implements the Apply interface, defined as:
   ```
   interface Apply for T {
     fn apply<R>(self: T) -> R {}
     fn apply<A, R>(self: T, a: A) -> R {}
     fn apply<A, B, R>(self: T, a: A, b: B) -> R {}
     ## ...
     ## fn apply<A, B, C, D, E, F, G, H, I, J, K, L, M, N, O, P, Q, R>(self: T, a: A, b: B, c: C, d: D, e: E, f: F, ..., q: Q) -> R {}
   }
   ```
   for example
   ```
   impl Apply for Integer {
     fn apply(self: Integer, other: Integer) { self * other }
   }
   5.apply(6) ## 30
   5(6) ## 30
   ```
