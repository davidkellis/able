# Able Programming Language

Able is a programming language that aims to be pleasant to use, relatively small, and relatively fast (in the same ballpark as Java or Go). You can kind of view Able as a hybrid between Go and Kotlin. Another reasonable characterization would be Rust + GC - ownership.

Able has the feel of a hybrid OO/functional language.

Here are a few sample programs to give you a feel for the language:

**Hello World**

```
fn main() {
  puts("Hello world.")
}

fn main()
  puts("Hello world.")
end
```

**Factorial**

```
# recursive
fn factorial(n: u32) => n < 2 ? 1 : n * factorial(n - 1)
```

```
# tail recursive
fn factorial(n) => factorial(n, 1)
fn factorial(n: u32, product: u32) -> u32 {
  if n < 2 then product else factorial(n - 1, n * product)
}
```

```
# iterative
fn factorial(n: u32) {
  return 1 if n < 2
  for i in 2...n { n *= i }
  n
}
```

```
# reduce
fn factorial(n: u32) {
  return 1 if n <= 1
  (2..n).reduce(*)
}
```

**Primes**

```
fn primes_less_than(max) => max < 2 ? List() : primes_less_than(max, 2, List())
fn primes_less_than(max: u32, i: u32, primes_found: List u32) -> List u32 {
  return primes_found if i > max
  if primes_found.any? { p => i % p == 0 }
    primes_less_than(max, i + 1, primes_found)
  else
    primes_less_than(max, i + 1, primes_found + i)
}
```

## Language Specification

The full language specification may be found in [language_spec.md](language_spec.md).

What follows is a summary of the language specification, and should be enough quickly get up to speed with Able.

## Identifiers

All identifiers must conform to the pattern `[a-zA-Z0-9][a-zA-Z0-9_]*`

There are two namespaces for identifiers: a type namespace and a value namespace.

### Naming Conventions

- Prefer snake_case for file names, package names, variable names, and function names.
- Prefer PascalCase for type names.

## Comments

```
# comments follow the pound sign to the end of the line
```

## Packages

todo: figure out how to determine root package

Every source file should start with a package definition of the form `package <unqualified name>`, for example `package io`; otherwise, the file is interpreted as if `package main` were specified.

### Importing Packages

- Package import: `import io`
- Wildcard import: `import io.*`
- Selective import: `import io.{puts, gets, SomeType}`
- Aliased import: `import internationalization as i18n.{transform as tf}`

Imports may occur in package scope or any local scope.

## Variables

```
<variable name>: <type name>                        # assigned the zero value for <type name>
<variable name>: <type name> = <value expression>   # assigned the specified initial value
<variable name> = <value expression>                # assigned the specified initial value and the type is inferred from the type of <value expression> and the use of <variable name>
```

## Types

### Type Expressions

A type is a name given to a set of values, and every value has an associated type. For example, `bool` is the name given to the set `{true, false}`, and since the value `true` is a member of the set `{true, false}`, it is of type `bool`. `TwoBitUnsignedInt` might be the type name we give to the set `{0, 1, 2, 3}`, such that `3` would be a value of type `TwoBitUnsignedInt`.

A type is denoted by a type expression. A type expression is a string. All types are parametric types, in that all types have zero or more type parameters.

Type parameters may be bound or unbound. A bound type parameter is a type parameter for which either a named type variable or a preexisting type name has been substituted. An unbound type parameter is a type parameter that is either unspecified or substituted by the placeholder type variable, denoted by `_`.

A type that has all its type parameters bound is called a concrete type. A type that has any of its type parameters unbound is called a polymorphic type, and a type expression that represents a polymorphic type is called a type constructor.

### Type Constraints

There is one supported type constraint, the "implements" constraint:

`T: I` is read as type T implements interface I.

`T: I + J + K` is read as type T implements interface I and J and K.

### Built-In Types

- void
- Boolean - `Bool`
- Integer types - `i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`
- Floating point types - `f32`, `f64`
- String
- Array
- Map
- Range
- Tuple
- Struct
- Union
- Function
- Interface
- Thunk
- Lazy
- CallStackLocal
- Channel

### void

The void type, named `void`, has a single literal value, `()`.

### Boolean

The boolean type, named `Bool`, has two values: `true` and `false`

### Integer Types

Decimal literal (denoted by `:decimal_literal:`) conforms to the pattern: `[0-9][0-9_]*`
Binary literal conforms to the pattern: `0b[01_]+`
Hex literal conforms to the pattern: `0x[0-9a-fA-F_]+`
Octal literal conforms to the pattern: `0o[0-8_]+`

Any of the integer literals may be suffixed with `_?((i|u)(8|16|32|64))?`

Examples:

```
123
0123
0x123
0b11001
0o123
123i8
123u8
123i32
123_456
123_456u32
123_456_u32
0x123f_u32
```

### Floating Point Types

A floating point literal conforms to the pattern: `(decimal_literal (\. decimal_literal)? exp?) | (\. decimal_literal exp?)`

An exponent (denoted by `exp`) conforms to the pattern: `('e'|'E') ( '+' | '-' )? decimal_literal`, e.g. `e9`, `E3`, `E+9`, `e-9`, `E-3`.

Any floating point literal may be suffixed with `_?f(32|64)`

Examples:

```
12.0
5.4e3
5.4e+3
5.4e-3
5.4e3f32
5.4e+3f32
5.4e-3f64
5.4e3f_32
5.4e+3_f32
5.4e-3_f64
123e3
123f64
123e-3f32
123_e3
123_f64
123_e-3_f32
.4e3
.4e+3
.4e-3
.4e3f32
.4e+3f32
.4e-3f64
.4e3f_32
.4e+3_f32
.4e-3_f64
```

### String

All string literals are UTF-8 encoded strings wrapped in double quotes.

String interpolation is notated with either $variable_name or ${expression}

```
name = "Herbert"
greeting = "Hello $name"
greeting2 = "Hello ${name}"
```

### Array

There is no array literal syntax. Rather, arrays are created with the `Array` constructor function.

```
x = Array(1,2,3)
x: Array f32 = Array(5.6, 5.7, 8)
```

### Map

There is no map literal syntax. Maps are created with the `Map` constructor function.

```
m = Map(1 : "Foo", 2 : "Bar")
m = Map(
  1 : "Foo",
  2 : "Bar"
)
m(3) = "Baz"
m << 4:"Qux"
```

### Range

Ranges take the following form:

```
# inclusive range
<expression>..<expression>

# exclusive range
<expression>...<expression>
```

For example:

```
# inclusive range
1..5 |> to_a   # Array(1, 2, 3, 4, 5)

# exclusive range
1...5 |> to_a  # Array(1, 2, 3, 4)
```

## Tuples

```
record = (1, "Foo", :bar)    # this has type (i32, String, Symbol)
if record._1 == 1 then puts("you're in first place!")
```

Single element tuples are a special case notated by `( <expression> , )`:

```
(100, )    # this has type (i32)
```

### Pairs

Pair syntax is just syntactic sugar for expressing 2-tuples.

`(1, "Foo")` can be written as `1 : "Foo"` or `1:"Foo"`, all of which have type (i32, String).

## Structs

Struct definitions can only appear in package scope.

### Singleton Structs

Structs that have a single value are called singleton structs; their type name name and their single value share the same identifier.

```
struct Red
```

### Definition with Positional Fields

Structs defined with positional fields might also be called named tuples.

```
# non-generic definition
struct Point { i32, i32 }

# generic definitions
struct Foo T U { i32, f32, T, U }
struct Foo T U { i32, f32, T, U, }
struct Foo T U [T: Iterable, U: Stringable] { i32, f32, T, U }
struct Foo T:Iterable U:Stringable {
  i32,
  f32,
  T,
  U,
}
struct Foo T {
  i32
  f32
  T
}
```

### Definition with Named Fields

```
# non-generic definition
struct Point { x: i32, y: i32 }

# generic definitions
struct Foo T { x: i32, y: f32, z: T }
struct Foo T { x: i32, y: f32, z: T, }
struct Foo T:Iterable { x: i32, y: f32, z: T }
struct Foo T {
  x: i32,
  x: f32,
  y: T,
}
struct Foo T {
  x: i32
  y: f32
  z: T
}
```

### Instantiate structs

```
Foo { 1 }               # struct literal with fields supplied by position
Foo { 1, 2, t1 }        # struct literal with fields supplied by position
Foo { x: 1 }             # struct literal with fields supplied by name
Foo { x: 1, y: 2, z: t1 }  # struct literal with fields supplied by name
Foo {
  x: 1,
  y: 2,
  z: t1
}
Foo {
  x: 1
  y: 2
  z: t1
}
```

Struct literal expressions do **not** require that every field be supplied.
Fields that are omitted are given a zero value appropriate for their type.

### Using struct instances

```
struct Foo T { i32, f32, T }
struct Bar T { a: i32, b: f32, c: T }

foo = Foo { 1, 2.0, "foo" }
bar = Bar { a=1, b=2.0, c="bar" }

puts("foo is a Foo String { ${foo._1}, ${foo._2}, ${foo._3} }")
puts("bar is a Bar String { a: ${bar.a}, b: ${bar.b}, c: ${foo.c} }")
```

## Unions

Union types are similar to discriminated unions in F#, enums in Rust, and algebraic data types in Haskell. A union type is a type made up of two or more member types, each of which can be any other type, even other union types.

### Union definitions referencing predefined types

```
union String? = String | Nil

struct Red { level: i32 }
struct Green { level: i32 }
struct Blue { level: i32 }
union Color = Red | Green | Blue
```

### Union definitions referencing new struct types

```
union SmallInt = One | Two | Three

union House = SmallHouse { sqft: f32 }
  | MediumHouse { sqft: f32 }
  | LargeHouse { sqft: f32 }

union House =
  SmallHouse { sqft: f32 }
  | MediumHouse { sqft: f32 }
  | LargeHouse { sqft: f32 }

union House =
  | SmallHouse { sqft: f32 }
  | MediumHouse { sqft: f32 }
  | LargeHouse { sqft: f32 }
```

### Generic Unions

Union types may be defined generically, as in the following examples.

**With Singleton Struct Alternatives**

```
union Option T = Some T { val: T } | None
```

**With named fields:**

```
union Tree T = Leaf T { value: T } | Node T { value: T, left: Tree T, right: Tree T }

# would be internally translated into
struct Leaf T { value: T }
struct Node T { value: T, left: Tree T, right: Tree T }
union Tree T = Leaf T | Node T

# other examples:

union Foo T [T: Blah] =
  | Bar A [A: Stringable] { a: A, t: T }
  | Baz B [B: Qux] { b: B, t: T }
```

**With positional fields:**

```
union Tree T = Leaf T { T } | Node T { T, Tree T, Tree T }

# would be internally translated into
struct Leaf T { T }
struct Node T { T, Tree T, Tree T }
union Tree T = Leaf T | Node T

# other examples:

union Option A = Some A {A} | None A {}
union Result A B = Success A {A} | Failure B {B}
union ContrivedResult A B [A: Fooable, B: Barable] =
  | Success A X [X: Stringable] {A, X}
  | Failure B Y [Y: Serializable] {B, Y}
```

### Any Type

The `Any` type is a special union type that is the union of all types known to the compiler at compile time.

## Blocks

A block is a sequence of expressions enclosed in a set of curly braces, prefixed with the `do` keyword.

For example:

```
do {
  a = 5 + 7
  puts(a)
  a
}
```

or

```
do { a = 5 + 7; puts(a); a }
```

## Functions

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

### Named function syntax

`fn <function name>[<optional type paramter list>](<parameter list>) -> <optional return type> { <function body> }`
or
`fn <function name>[<optional type paramter list>](<parameter list>) -> <optional return type> => <function body>`

This style allows type constraints to be captured in the type parameter list (e.g. `fn convert_to_string[T: Parsable & Stringable](a: T) => to_string(a)` )

#### Point-free style

Point-free syntax may only be used with named functions, and takes the following form:

`fn <function name>[<optional type paramter list>] = <function object OR partial application expression OR placeholder lambda expression>`

Partial application is documented in the section [Partial application](#partial-application), and placeholder lambda expressions are covered in the section [Implicit Lambda Expressions via Placeholder Expression Syntax (a.k.a. Placeholder Lambdas)](#implicit-lambda-expressions-via-placeholder-expression-syntax-aka-placeholder-lambdas).

For example:

```
# puts is an alias for the println function
fn puts = println

# given:
fn add(a: i32, b: i32) -> i32 => a + b
fn fold_right(it: Iterator T, initial: V, accumulate: (T, V) -> V) -> V {
  it.reverse.fold_left(initial) { acc, val => accumulate(val, acc) }
}

# add1, add5, and add10 are partial applications of the add function
# sum is the partially applied fold_right function
fn add1 = add(1)
fn add5 = add(5,)
fn add10 = add(, 5)
fn sum = fold_right(, 0, +)

# definitions in terms of placeholder lambda expressions
fn add10_and_double = 2 * add(_, 10)
fn println_if_debug = println(_) if APP_CFG.debug
fn sum = _.fold_right(0, +)
fn sum = fold_right(_, 0, +)
```

### Anonymous function syntax

`fn[<optional type paramter list>](<parameter list>) -> <optional return type> { <function body> }`
or
`fn[<optional type paramter list>](<parameter list>) -> <optional return type> => <function body>`

You may omit the leading `fn` at the expense of not being able to capture constraints on generic type parameters:

`(<parameter list>) -> <optional return type> { <function body> }`
or
`(<parameter list>) -> <optional return type> => <function body>`

The style with the explicit free type parameter list is the only way to capture type constraints (e.g. `fn[T: Parsable & Stringable](a: T) => toString(a)` )

When the optional return type is omitted, then the `->` return type delimiter that immediately follows the parameter list and preceeds the function body should also be omitted.

### Lambda expression syntax

Lambda expressions are defined in the lambda expression syntax:
`{ <paramter list> -> <return type> => expression }`

The parameter list and the return type declarations are optional.

#### Implicit Lambda Expressions via Placeholder Expression Syntax (a.k.a. Placeholder Lambdas)

Lambda expressions may also be represented implicitly with a special convenience syntax called placeholder expression syntax.

An expression may be parameterized with placeholder symbols - each of which may be either an underscore, \_, or a numbered underscore, e.g. \_1, \_2, etc.

A placeholder symbol may be used in place of any value-representing subexpression.

The scope of a lambda expression implicitly defined through the use of placeholder symbols extends to the smallest subexpression that, if considered as the body of an anonymous function, would represent a function with the same function signature as what the containing expression expects.

For example:

`_ + 1` is treated as `{ x => x + 1 }`

`_ * _` is treated as `{ x, y => x * y }`

`if _ x else y` is treated as `{ c => if c x else y }`

`_.map(_ * 5)` is treated as `{ x => x.map( { y => y * 5 } ) }` where the subexpression `_ * 5` is treated as an inner lambda expression, and the full expression is treated as lambda expression

`_.reduce(_, _ * _)` is treated as `{ x, y => x.reduce(y, { a, b => a * b }) }` where `_ * _` is treated as an inner lambda expression, and the full expression is treated as a lambda expression

##### Special case

If a placeholder lambda is enclosed within a block, and the scope of the placeholder lambda expression extends to the boundaries of the block's delimiters, then the full block is considered a lambda expression having the same parameters as that of the contained placeholder lambda.

For example:
`5.times { puts("Greeting number $_") }` is treated as `5.times { i => puts("Greeting number $_") }`

### Variadic functions

```
Given:
fn puts(vals: Any*)

puts()
puts("Greetings!")
puts("Hello!", "Goodbye!")
puts("Hello!", *Array("A", "B", "C"), "Goodbye!")
```

#### Variadic disambiguation rule

If two functions are defined, one variadic and one non-variadic, and their function signatures conflict/overlap,
then an invocation of the ambiguous signature will dipatch to the non-variadic function.

For example:

```
Given:
fn log(label: String)
fn log(label: String, val: Any)
fn log(label: String, vals: Any*)

log("foo")                    # ambiguous; invokes non-variadic signature, log(String)
log("foo", 123)               # ambiguous; invokes non-variadic signature, log(String, Any)
log("foo", 123, "bar")        # not ambiguous; invokes variadic signature, puts(String, Any*)
```

### Function definition examples

#### Define non-generic functions

With explicit return type:

1. fn createPair(a, b: i32) -> Pair i32 { Pair{a, b} }<br>
   fn createPair(a, b: i32) -> Pair i32 => Pair{a, b}
2. createPair = fn(a, b: i32) -> Pair i32 { Pair{a, b} }<br>
   createPair = fn(a, b: i32) -> Pair i32 => Pair{a, b}

3. createPair = (a, b: i32) -> Pair i32 { Pair{a, b} }<br>
   createPair = (a, b: i32) -> Pair i32 => Pair{a, b}
4. createPair = { a, b: i32 -> Pair i32 => Pair{a, b} }

With inferred return type:

1. fn createPair(a, b: i32) { Pair{a, b} }<br>
   fn createPair(a, b: i32) => Pair{a, b}

2. createPair = fn(a, b: i32) { Pair{a, b} }<br>
   createPair = fn(a, b: i32) => Pair{a, b}

3. createPair = (a, b: i32) { Pair{a, b} }<br>
   createPair = (a, b: i32) => Pair{a, b}

4. createPair = { a, b: i32 => Pair{a, b} }

All the functions defined above are of type:
`(i32, i32) -> Pair i32`

#### Define generic functions

With explicit free type parameter and explicit return type:

1. fn createPair[T](a, b: T) -> Pair T { Pair{a, b} }<br>
   fn createPair[T](a, b: T) -> Pair T => Pair{a, b}

2. createPair = fn[T](a, b: T) -> Pair T { Pair{a, b} }<br>
   createPair = fn[T](a, b: T) -> Pair T => Pair{a, b}

With explicit free type parameter and inferred return type:

1. fn createPair[T](a, b: T) { Pair{a, b} }<br>
   fn createPair[T](a, b: T) => Pair{a, b}

2. createPair = fn[T](a, b: T) { Pair{a, b} }<br>
   createPair = fn[T](a, b: T) => Pair{a, b}

With implied free type parameter and explicit return type:

1. createPair = (a, b: T) -> Pair T { Pair{a, b} }<br>
   createPair = (a, b: T) -> Pair T => Pair{a, b}

2. createPair = { a, b: T -> Pair T => Pair{a, b} }

With implied free type parameter and inferred return type:

1. createPair = (a, b: T) { Pair{a, b} }<br>
   createPair = (a, b: T) => Pair{a, b}

2. createPair = { a, b: T => Pair{a, b} }

All the functions defined above are of type:
`(T, T) -> Pair T`

#### Lambda Expressions

```
createPair = { a, b: i32 -> Pair i32 => Pair{a, b} }
createPair = { a, b: i32 => Pair{a, b} }
createPair = { a, b: T -> Pair T => Pair{a, b} }
createPair = { a, b: T => Pair{a, b} }
```

### Function Application

#### Normal application

Functions may be invoked with the normal function application syntax `<function object>(<arg>, <arg>, ...)`, as in:

```
sum(Array(1,2,3))
f(1,2,*xs)
buildPair(1,2)
map(Array(1,2,3), square)
```

#### Named argument application

Given the following function definition:

```
fn create_person(age: i32, name: String, address: Address) -> Person { ... }
```

the `create_person` function may be invoked in any of the following ways:

```
create_person(@name: "Bob", @age: 20, @address: Address("123 Wide Street"))

create_person(
  @name: "Bob",
  @age: 20,
  @address: Address("123 Wide Street")
)

create_person(
  @name: "Bob"
  @age: 20
  @address: Address("123 Wide Street"))
)

create_person(
  @name: "Bob", @age: 20
  @address: Address("123 Wide Street"),
)
```

#### Partial application

Given:

```
f = fn(a: i32, b: i32, c: i32, d: i32, e: i32) => (a,b,c,d,e)
```

Partial application of the trailing arguments

```
f(,1,2,3)
```

Explicit partial application of the leading arguments

```
f(1,2,3,)
```

Implied partial application of the leading arguments

```
f(1,2,3)
```

Partial application of the first, third, and fifth arguments; returns a function with the signature (i32, i32) -> (i32, i32, i32, i32, i32)

```
f(1, _, 2, _, 3)
```

Partial application of the first, third, and fifth arguments; returns a function with the signature (i32) -> (i32, i32, i32, i32, i32)

```
f(1, _1, 2, _1, 3)
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

#### Method application

`obj.method(1, 2, 3)`
is equivalent to
`method(obj, 1, 2, 3)`

`obj.m1(1, 2, 3).m2(4, 5, 6)`
is equivalent to
`m2(m1(obj, 1, 2, 3), 4, 5, 6)`

#### Pipeline application

Given:
`struct Coord {i32, i32}`

`5 |> Coord {_1, _1}`
is equivalent to
`Coord{5, 5}`

`1 |> plus_two |> f(1, 2, _, 4, 5)`
is equivalent to
`f(1, 2, plus_two(1), 4, 5)`

#### Operator application

Operators are functions too. The only distinction between operator functions and non-operator functions is that operator functions additionally have an associated operator precedence and are associative. Non-operator functions are non-associative, and have no operator precedence.

By default, operators are left-associative. For example, `a + b + c` is evaluated as `((a + b) + c)`

Operators that have a trailing apostrophe are right-associative. For example, `a +' b +' c` is evaluated as `(a +' (b +' c))`

##### Right-associative operators

1. Exponentiation operator: `**`
   The exponentiation operator is a right-associative operator with precedence that is higher than addition, subtraction, multiplication, and division.

##### Special operators

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

#### Special application cases

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

## Interfaces

An interface is a set of functions that are guaranteed to be defined for a given type.

An interface can be used to represent either (1) a type, or (2) a constraint on a type or type variable.

When used as a type, the interface functions as a lens through which we see and interact with some underlying type that implements the interface. This form of polymorphism allows us to use different types as if they were the same type. For example, if the concrete types `i32` and `f32` both implement an interface called `Numeric`, then one could create an `Array Numeric` consisting of both `i32` and `f32` values.

When used as a constraint on a type, the interpretation is that some context needs to reference the underlying type, while also making a claim or guarantee that the underlying type implements the interface.

### Interface Definitions

Interfaces are defined with the following syntax:

```
interface A [B C ...] [for D [E F ...]] [where <type constraints>]{
	fn foo(T1, T2, ...) -> T3
	...
}
```

where A is the name of the interface optionally parameterized with type parameters, B, C, etc.

The `for` clause is optional. When present, the `for` clause specifies that a type variable, D, is deemed to implement interface A. If no `for` clause is specified, then the interface serves as a type constraint on types B, C, etc.

The type expression in the for clause is called the interface's <a name="self-type"></a>_self type_. The self type is a concrete or polymorphic type representing the type that will implement the interface. In the syntax above, `D [E F ...]` represents the self type.

D may represent either a concrete type or a polymorphic type depending on whether D's type parameters (if applicable) are bound or unbound (see [Type Expressions](#type-expressions) for more on bound vs. unbound type parameters).

Here are a few example interface definitions:

```
# Stringable interface
interface Stringable for T {
  fn to_s(T) -> String
}

# Comparable interface
interface Comparable for T {
  fn compare(T, T) -> i32
}

# Iterable interface
interface Iterable T for E {
  fn each(E, T -> void) -> void
  fn iterator(e: E) -> Iterator T => Iterator { gen => e.each(gen.yield) }
}

# Mappable interface (something approximating Functor in Haskell-land - see https://wiki.haskell.org/Typeclassopedia#Functor)
interface Mappable for M _ {
  fn map(m: M A, convert: A -> B) -> M B
}

# Types that satisfy the Range interface may be used with range literal syntax (e.g. S..E, or S...E) to represent a range of values.
interface Range S E Out {
  fn inclusive_range(start: S, end: E) -> Iterable Out
  fn exclusive_range(start: S, end: E) -> Iterable Out
}
```

#### Self Type

There is a special type named `Self` that may be used within an interface definition or an implementation of an interface to reference the [self type](#self-type) captured in the for clause of the interface or implementation.

For example:

```
# Stringable interface
interface Stringable for T {
  fn to_s(T) -> String
}
```

could also be written as:

```
interface Stringable for T {
  fn to_s(Self) -> String
}
```

In cases where the [self type](#self-type) is polymorphic, i.e. the self type has one or more unbound type parameters, the special `Self` type may be used as a type constructor and supplied with type arguments.

For example:

```
# Mappable interface (something approximating Functor in Haskell-land - see https://wiki.haskell.org/Typeclassopedia#Functor)
interface Mappable for M _ {
  fn map(m: M A, convert: A -> B) -> M B
}
```

could be written as:

```
# Mappable interface (something approximating Functor in Haskell-land - see https://wiki.haskell.org/Typeclassopedia#Functor)
interface Mappable for M _ {
  fn map(m: Self A, convert: A -> B) -> Self B
}
```

#### Interface Aliases and Interface Intersection Types

Intersection interfaces may be defined with the following syntax:

```
interface A [B C ...] = D [E F ...] [+ G [H I ...] + J [K L ...] ...]
```

where A is the name of the intersection interface and D, G, and J are the members of the intersection set.

Some contrived examples of intersection types:

```
interface WebEncodable T = XmlEncodable T + JsonEncodable T
interface Repository T = Collection T + Searchable T
interface Collection T = Countable T + Addable T + Removable T + Getable T
```

### Interface Implementations

Interface implementations are defined with the following syntax:

```
[ImplName =] impl [X, Y, Z, ...] A <B C ...> [for D <E F ...>] [where <type constraints>] {
	fn foo(T1, T2, ...) -> T3 { ... }
	...
}
```

Here are some example interface implementations:

```
struct Person { name: String, address: String }
impl Stringable for Person {
  fn to_s(p: Person) -> String => "Name: ${p.Name}\nAddress: ${p.Address}"
}

impl Comparable for Person {
  fn compare(p1, p2) => compare(p1.Name, p2.Name)
}

impl Iterable T for Array T {
  fn each(array: Array T, visit: T -> void) -> void {
    i = 0
    length = array.length()
    while i < length {
      visit(array(i))
      i += 1
    }
  }
}

struct OddNumbers { start: i32, end: i32 }
impl Iterable i32 for OddNumbers {
  fn each(odd_numbers_range: OddNumbers, visit: i32 -> void) {
    start = odd_numbers_range.start.even? ? odd_numbers_range.start + 1 : odd_numbers_range.start
    end = odd_numbers_range.end.even? ? odd_numbers_range.end - 1 : odd_numbers_range.end
    while start <= end {
      visit(start)
      start += 2
    }
  }
}

impl Mappable A for Array {
  fn map(arr: Array A, convert: A -> B) -> Array B {
    ab = ArrayBuilder(0, arr.length)  # length, capacity
    arr.each { element => convert(element) |> ab.add }
    ab.to_a
  }
}

# implement Range for f32 start and end values (e.g. 1.2..4.5, and 1.2...4.5)
impl Range i32 for (f32, f32) {
  fn inclusive_range(start: f32, end: f32) -> Iterable i32 => start.ceil..end.floor
  fn exclusive_range(start: f32, end: f32) -> Iterable i32 => start.ceil...end.floor
}
```

Implementations may be named in order to cope with situations where the same type implements the same interface with differing semantics. When an implementation is named, function calls that might otherwise be ambiguous can be fully qualified in order to disambiguate the function call. For example:

```
# given
interface Monoid for T {
  fn id() -> T
  fn op(T, T) -> T
}
Sum = impl Monoid for i32 {
  fn id() -> i32 = 0
  fn op(x: i32, y: i32) -> i32 = x + y
}
Product = impl Monoid for i32 {
  fn id() -> i32 = 1
  fn op(x: i32, y: i32) -> i32 = x * y
}

# since both of the following are ambiguous
id()
Monoid.id()
# you have to qualify the function call with the name of the implementation, like
Sum.id()
```

### Interface Usage

When calling a function defined in an interface, a client may supply either (1) a value of a type that implements the interface - the interface's [self type](#self-type) - or (2) a value of the interface type to the function in any argument that is typed as the interface's self type.

For example, given:

```
interface Iterable T for E {
  fn each(E, T -> void) -> void
  ...
}
```

a client may call the `each` function by supplying a value of type `Iterable T` as the first argument to `each`, like this:

```
# assuming List T implements the Iterable T interface
e: Iterable i32 = List(1, 2, 3)    # e has type Iterable i32
each(e, puts)
```

or `each` may be called by supplying a value of type `Array T` as the first argument, like this:

```
# assuming Array T implements the Iterable T interface
e = Array(1, 2, 3)    # e has type Array i32
each(e, puts)
```

In cases where the interface was defined without a `for` clause, then the interface may only be used as a type constraint. For example:

```
fn build_int_range[S, E, Range S E i32](start: S, end: E) -> Iterable i32 => start..end
```

Less frequently, a function in an interface may be defined to accept two or more parameters of the interface's self type. If the self type is polymorphic, and the function is invoked with a value of the interface type as an argument in place of one of the self type parameters, then the other self type parameters must also reference the same underlying type (either directly as a value of the underlying type, or indirectly through an interface value that represents the underlying type), so that the multiple self type parameters all remain consistent with respect to one another - i.e. so that they all reference the same underlying type.

### Overlapping Implementations / Impl Specificity

Since it is possible to define interface implementations that conflict, or overlap, with one another, we need some means to decide which one to use. In some cases, competing implementations are not clearly more specific than one another, but a good deal of the time, we can clearly define one implemention to be more specific than another.

Able's specificity logic is a derivative of Rust's impl specialization proposal at https://github.com/nox/rust-rfcs/blob/master/text/1210-impl-specialization.md. Much of this specification is taken from Rust's specialization RFC.

The following implementations, written without their implementation body, conflict with one another:

```
# overlap when T = i32
impl Foo for T {}
impl Foo for i32 {}

# overlap when T = i32
impl Foo for Array T {}
impl Foo for Array i32 {}

# overlap because Array T implements Iterable T
impl Foo for Iterable T {}
impl Foo for Array T {}

# overlap when T implements Bar
impl Foo for T {}
impl Foo for T: Bar {}

# overlap when T implements Bar and Baz
impl Foo for T: Bar {}
impl Foo for T: Baz {}

# overlap when T = f32
impl Foo for i32 | f32 {}
impl Foo for i32 | T {}

# overlap when two type unions implement the same interface, s.t. one type union is a proper subset of the other
impl Foo for i32 | f32 | f64 {}
impl Foo for i32 | f32 {}
```

The following specificity/disambiguation rules resolve the patterns of overlap/conflict that are most common:

1. Concrete types are more specific than type variables. For example:
   - `i32` is more specific than `T`
   - `Array i32` is more specific than `Array T`
2. Concrete types are more specific than interfaces. For example:
   - `Array T` is more specific than `Iterable T`
3. Interface types are more specific than unconstrained type variables. For example:
   - `Iterable T` is more specific than `T`
4. Constrained type variables are more specific than unconstrained type variables. For example:
   - `T: Iterable` is more specific than `T`
5. A union type that is a proper subset of another union type is more specific than its superset. For example:
   - `i32 | f32` is more specific than `i32 | f32 | f64`
6. Given two unions that differ only in a single type member, such that the member is concrete in one union and generic in the other union, the union that references the concrete member is more specific than the union that references the generic member.
   - `i32 | f32` is more specific than `i32 | T`
7. Given two unions that differ only in a single type member, such that the member is concrete in one union and an interface in the other union, the untion that references the concrete member is more specific than the union that references the interface member.
   - `i32 | f32` is more specific than `i32 | FloatingPointNumeric`
8. Given two sets of comparable type constraints, ConstraintSet1 and ConstraintSet2, if ConstraintSet1 is a proper subset of ConstraintSet2, then ConstraintSet1 imposes fewer constraints, and is therefore less specific than ConstraintSet2. In other words, the constraint set that is a proper superset of the other is the most specific constraint set. If the constraint sets are not a proper subset/superset of one another, then the constraint sets are not comparable, which means that the competing interface implementations are not comparable either. This rule captures rule (4).
   For example:
   - `T: Iterable` is more specific than `T` (with no type constraints), because `{Iterable}` is a proper superset of `{}` (note: `{}` represents the set of no type constraints).
   - `T: Iterable & Stringable` is more specific than `T: Iterable` because `{Iterable, Stringable}` is a proper superset of `{Iterable}`

There are many potential cases where the type constraints in competing interface implementations can't be compared. In those cases, the compiler will emit an error saying that the competing interface implementations are not comparable, and so the expression is ambiguous.

Ambiguities that cannot be resolved by these specialization rules must be resolved manually, by qualifying the relevant function call with either the name of the interface or the name of the implementation. For example:

```
# assuming Array T implements both Iterable T and BetterIterable T, and both define an each function
impl Iterable T for Array T { ... }
impl BetterIterable T for Array T { ... }

# an expression like
Array(1,2,3).each { elem => puts(elem) }
# would be ambiguous

# however, the expressions
BetterIterable.each(Array(1,2,3)) { elem => puts(elem) }
Array(1,2,3) |> BetterIterable.each { elem => puts(elem) }
# are not ambiguous
```

## Type Aliases

```
type Int32 = i32
type BinaryOperator A = (A, A) -> A
```

## Zero Values

Every type has a zero value. The following sections document the zero value for every type.

### Zero values for primitive types

- Nil - `nil`
- void - `()`
- Boolean - `false`
- Integer types - `0`
- Floating point types - `0.0`
- String - `""`
- Array - `Array[T]()` or `Array T{}`
- Map - `Map[K,V]()` or `Map K V{}`

### Zero values for compound types

#### Range

A range is an interface with three type parameters, and therefore has the zero value of an interface with three type parameters.

#### Tuple

A tuple of type `(A, B, C, ...)` with type parameters `A`, `B`, `C`, etc. has a zero value of
`(A', B', C', ...)` where `A'` is the zero value for type `A`, `B'` is the zero value for type `B`, `C'` is
the zero value for type `C`, etc.

#### Struct

A struct of type `struct S { a: A, b: B, c: C, ... }` with type parameters `A`, `B`, `C`, etc. has a zero value
of `S { a = A', b = B', c = C', ... }` where `A'` is the zero value for type `A`, `B'` is the zero value for type `B`, `C'` is
the zero value for type `C`, etc.

#### Union

Any union type in which `Nil` is a member has a zero value of `nil`. The `Any` type has a zero value of `nil`.

The zero value of any other union is arbitrarily chosen as the zero value of one of the member types of the union. It is undefined which member of the union will be chosen as the type used to produce a zero value. For example, for the type `A | B | C`, the zero value is arbitrarily/randomly chosen to be either the zero value for A, the zero value for B, or the zero value for C.

#### Function

The zero type for a function with signature `(A, B, C, ...)->Z` with type parameters `A` through `Z` is
the function `fn[A,B,C,...,Z](a:A, b:B, ...)->Z { Z' }`, where `Z'` represents the zero value for type `Z`.

#### Interface/Impl

The zero value for an interface with signature:

```
interface Foo for T {
  fn bar[A,B](a: A, b:B, ...) -> C
  fn baz[D,E](d:D, e:E, ...) -> F
}
```

is the value `nil`. The reason why is that all interfaces define a default implementation for the Nil type, such that each function returns a zero value of the function's return type, for example, `Nil` implicitly implements the previously defined `Foo` interface with a randomly generated impl of the form:

```
Foo1887678 = impl Foo for Nil {
  fn bar[A,B](a: A, b:B, ...) -> C { C' }
  fn baz[D,E](d:D, e:E, ...) -> F { F' }
}
```

where `C'` and `F'` represent the zero values for types `C` and `F` respectively.

## Control Flow Expressions

### If Statement

All `if` statements are expressions, evaluating to the last expression in the branch taken. In cases that an if expression
doesn't define a possible branch and that branch is taken, the if expression evaluates to nil.

```
if <condition> {
  ...
}
```

```
if <condition> {
  ...
} else {
  ...
}
```

```
if <condition> {
  ...
} elsif {
  ...
} else {
  ...
}
```

#### If Suffix Syntax

```
<expression> if <condition>
```

#### Ternary Syntax

```
<condition> ? <true branch expression> : <false branch expression>
```

### Unless Suffix Syntax

```
<expression> unless <condition>
```

### While Suffix Syntax

```
<expression> while <condition>
```

### Pattern Matching

Pattern matching expressions take the following general form:

```
<expression> match {
  <destructuring pattern> => <expression>
  <destructuring pattern> => <expression>
  <destructuring pattern> => <expression>
  ...
}

# OR

<expression> match { <destructuring pattern> => <expression> | <destructuring pattern> => <expression> | ... }
```

For example:

```
union House =
  | TinyHouse { sqft: f32 }
  | SmallHouse { sqft: f32 }
  | MediumHouse { sqft: f32 }
  | LargeHouse { sqft: f32 }
  | HugeHouse { sqft: f32, pools: i32 }

fn buildHouse(h: House) => h match {
  TinyHouse | SmallHouse => puts("build a small house")
  m: MediumHouse => puts("build a modest house - $m")
  l: LargeHouse{area} => puts ("build a large house of $area sq. ft. - $l")
  HugeHouse{_, poolCount} => puts ("build a huge house with $poolCount pools!")
}
```

### Looping

There are two looping constructs: `while`, and `for`.

#### while

```
while <condition> {
  ...
}
```

#### for

```
for <destructuring expression> in <Iterable value> {
  ...
}
```

Examples:

```
for i in 1..10 { puts(i) }

for (i, ch) in (1..10).zip("a".."j") { puts("$i -> $ch") }

for p: Person(n=name, a=address, z=zip) in personRepo.all() { puts("name: $n, address: $a, zip: $z") }
```

#### break/continue

`break` and `continue` may be used within any looping construct in order to break out of the loop or continue to
the next iteration of the loop, just as in most C-like languages.

### Jump-points / Non-local Return

Jump-points are labels that correspond to a position in the call stack, such that, if a `jump <label> <value>` statement is evaluated, the call stack is unwound up to the point at which the jump-point was established, and the `jumppoint` expression resolves to the value supplied to the `jump` statement.

Jump-point expressions take the following form:

```
jumppoint <label> {
  ...
  jump <label> <value>
  ...
}
```

To following example demonstrates an implementation of short circuiting logic, achieved via jump-points. The example implements the `all?` function defined within the `Iterable` interface. The combination of the `jumppoint` and the `jump` within the `each` block implements immediate return from `all?` upon the first observation of a false evaluation of the predicate function. This short-circuiting logic prevents further unnecessary evaluations of the predicate function once the first observation of a false return value is observed:

```
fn all?(iterable: I, predicate: T -> bool) -> bool {
  jumppoint :stop {
    iterable.each { t => jump :stop false unless predicate(t) }
    true
  }
}
```

### Generators

Able supports generators through the `Iterator[T]( Generator T -> void ) -> Iterator T` function. The `Iterator` function accepts a value-producing function - the generator function - and returns an `Iterator T` that iterates over the values produced by the generator.

The following are examples of generator expressions:

```
Iterator { gen => for i in 1..10 { gen.yield(i) } }
Iterator { gen => (1..10).each(gen.yield) }
```

In order to signal the end of the iterator, the generator function must either explicitly or implicitly return. The returned value is irrelevant. If the generator function never returns, then the generator produces an infinite Iterator.

The Generator T interface is defined as:

```
interface Generator A for T {
  fn yield(T, A) -> T
}
```

## Destructuring

Destructuring is the binding of variable names to positional or named values within a data structure. Destructuring may be used in assignment expressions, function parameter lists, and pattern matching alternatives.

There are three destructuring forms, depending on the type of thing being destructured. These three forms support the destructuring of structs, tuples, pairs, and types implementing the Iterable interface all support destructuring. Unions do not support destructuring.

There are three different contexts in which destructuring may be used.

The following sections demonstrate the three destructuring forms and how they can be used in each of the three different destructuring contexts. The examples assume the following Person and Address structure definitions:

```
struct Person {
  name: String
  height: f32
  weight: f32
  age: i32
  address: Address
}

struct Address {
  street: String
  state: String
  zip: i32
}
```

### Destructuring Forms

#### Struct Destructuring

Structs may be destructured positionally with the syntax:

```
p: Person { n, h, w, a, addr: Address {street, state, zip} }
```

In positional destructuring expressions, local variable identifiers may be bound to field values within the struct. Bindings are established by matching up local variable identifiers with fields based on the relative position of those variable identifiers and corresponding fields. For example, if a struct has three fields - named or not - then the fields may be bound to local variable identifiers by inserting variable names at the position within the destructuring expression at which those fields appear in the struct definition.

Structs may also be destructured via named field destructuring, which takes the following form:

```
# if it is desirable to reference p.address via the local identifier addr, then do the following:
p: Person { a=age, h=height, w=weight, addr: Address{z=zip}=address }

p Person{ age, h as height, w as weight, addr Address{z as zip} as address, Profile{billing}}
p Person{ age, h = height, w = weight, addr Address{z as zip} = address}
p Person{ age, h = height, w = weight, addr Address{z as zip} = address, Profile{billing} = profile}

p Person{@age: age, h as height, w as weight, addr Address{z as zip} as address}

# or

# if it isn't necessary to reference p.address via a local identifier, then do the following:
p: Person { a=age, h=height, w=weight, Address{z=zip}=address }
```

Named field destructuring uses the assignment operator to denote that a local variable identifier should be bound to a particular named field within the struct; the syntax takes the form `local_variable_identifier = field_name_from_struct`.

In cases where named field destructuring expressions need to be recursively destructured, the left hand side of the assignment operator may take one of two forms, (1) `local_identifier: AnotherStruct {...}`, or (2) `AnotherStruct {...}`. The first form is used when it is desirable to bind the full value that is being recursively destructured to a local identifier, while the second form is used when it is unnecessary to reference the full value that is being recursively destructured.

#### Tuple Destructuring

Tuples are destructured in the same way that structs are, except that (1) tuples are destructured without a leading type name and (2) tuples are destructured with a set of surrounding parenthesis instead of curly braces.

Destructuring a tuple positionally takes the form:

```
triple: ( a, b, c: Address { street, state, zip } )
```

Destructuring a tuple via "named" field destructuring takes the form:

```
triple: ( a=_0, b=_1, c: Address { z=zip }=_2 )
```

NOTE: In some contexts, the type of the positional fields may be necessary for type disambiguation purposes. In those cases, the type may be notated immediately following the name of the local variable identifier, as in the following two examples:

1. Destructuring a tuple positionally with explicit positional field types:

```
triple: ( a: i32, b: i32, c: Address { street, state, zip } )
```

2. Destructuring a tuple via "named" field destructuring with explicit named field types:

```
triple: ( a: i32=_0, b: i32=_1, c: Address { z=zip }=_2 )
```

#### Sequence Destructuring

Sequence destructuring may be used anywhere that struct destructuring may be used, and the syntax is the same
with the exception that square brackets are used instead of curly braces.

The syntax for sequence destructuring is:

```
<sequence struct that implements Iterable>[<var1>, <var2>, <...>, <varN>, *<remainder sequence>]
```

where `<var1>`, `<var2>`, `<...>`, and `<varN>` represent variable bindings and `*<remainder sequence>` represents
an optional variable binding capturing an iterator over the remaining part of the sequence that was not consumed
by the leading variable bindings.

For example:

```
Array[a, b, c] = Array(1..100)
Array[a, b, c, *rest] = Array(1..100)
List[x, y, z, *rest] = List(1..100)
```

Sequence destructuring is available to any struct that implements the Iterable interface.

### Destructuring Contexts

#### Assignment Destructuring

```
# positional destructuring form; assignment context
fn printPersonInfo(p: Person) {
  Person{n,h,w,a} = p
  puts("name=$n   height=$h   weight=$w   age=$a")
}

# or

# named field destructuring form; assignment context
fn printPersonInfo(p: Person) {
  Person{a=age, h=height, w=weight, n=name} = p
  puts("name=$n   height=$h   weight=$w   age=$a")
}

printPersonInfo(Person("Jim", 6.0833, 170, 25))


# positional tuple destructuring form; assignment context
fn printPair(pair: (i32, String)) {
  (i, str) = pair
  puts("i=$i   str=$str")
}

printPair( (4, "Bill") )
printPair( 4:"Bill" )
```

#### Function Parameter Destructuring

```
# 3 equivalent parameter destructuring examples;
fn printPersonInfo(p: Person{n, h, w, a}) {
  puts("name=$n   height=$h   weight=$w   age=$a")
}
# or
fn printPersonInfo(_: Person{n, h, w=weight, a}) {
  puts("name=$n   height=$h   weight=$w   age=$a")
}
# or
fn printPersonInfo(Person{n, h, w, a}) {
  puts("name=$n   height=$h   weight=$w   age=$a")
}

printPersonInfo(Person("Jim", 6.0833, 170, 25))


fn printPair(pair: (i: i32, str: String)) {
  puts("i=$i   str=$str")
}

printPair( (4, "Bill") )
printPair( 4:"Bill" )
```

#### Pattern Matching Destructuring

```
# assuming h is of type House
h match {
  TinyHouse | SmallHouse => puts("build a small house")
  m: MediumHouse => puts("build a modest house - $m")
  LargeHouse{area} => puts ("build a large house of $area sq. ft.")
  HugeHouse{poolCount=pools} => puts ("build a huge house with $poolCount pools!")
}
```

## Exceptions

```
struct Foo {code: i32, msg: String}

fn pp(w: Widget) {
  raise Foo(msg = "", code=5) if w.type.undefined?
}

fn foo() {
  bar()
  do {
    baz()
  catch:
    Foo(m=msg) => "puts foo error!"
  ensure:
    puts("done!")
  }
  qux()
}

fn foo() => baz() catch: Foo(m=msg) => "puts foo error!" ensure: puts("done!")

fn main() {
  pp(Widget(name = "card shuffler"))
catch:
  f: Foo(m=msg) => puts "foo error = $f"
  _ => puts "other error"
ensure:
  puts("ensure block called!")
}

fn main() => pp(Widget(name = "card shuffler")) catch: f: Foo(m=msg) => puts "foo error = $f"; _ => puts "other error" ensure: puts("ensure block called!")
```

Expressions that need to be evaluated prior to function return should be placed in an ensure block. Ensure blocks may be used to suffix try/catch expression, or they may be used to suffix function definitions.

How should ensure blocks affect the return value of the function?

## Lazy Evaluation

According to http://matt.might.net/articles/implementing-laziness/, "there are two reasonable interpretations of laziness" - by-name and by-need.

**In both interpretations, the evaluation of an expression is delayed until the expression's value is accessed, the difference is in whether the resulting value is memoized or not.**

### By-name Evaluation (not memoized)

By-name evaluation re-evaluates the expression each time the value is needed.

By-name evaluation is supported through the `Thunk` type.

A thunk represents an unevaluated value.

A thunk can be created in one of three ways:

1. assignment to a variable of type Thunk T
2. supplying an expression to a function parameter of type Thunk T
3. explicitly with the Thunk function.

For example

```
# 1. assignment to variable of type Thunk T
mail: Thunk Mail = sendAndReceiveMail()

# 2. supplying an expression to a function parameter of type Thunk T
fn runAfterDelay[T](expr: Thunk T) -> T {
  sleep(10.seconds)
  expr
}
runAfterDelay(sendAndReceiveMail())

# 3. using the Thunk function
Thunk(sendAndReceiveMail())
```

References:

- https://en.wikipedia.org/wiki/Thunk
- https://wiki.haskell.org/Thunk

### By-need Evaluation (memoized)

In by-need evaluation, the expression is evaluated once, on the first occasion that the expression's value is needed, and the resulting value is memoized for future use.

By-need evaluation is supported through the `Lazy` type.

The `Lazy` type has the same semantics as the `Thunk` type, with the exception that the value produced by the first access is memoized, such that subsequent accesses of the value return the same instance of the value.

## Special Evaluation Rules

### Value Discarding

If an expression, e, is used in a context that expects a value of type void, then regardless of the expression's type, the expression will be transformed into the expression { e; () }.

This evaluation rule is very similar to Scala's Value Discarding rule. See section 6.26.1 of the Scala language reference at http://www.scala-lang.org/docu/files/ScalaReference.pdf.

## Macros

Able supports meta-programming in the form of compile-time macros.

Macros are special code-emitting functions that capture a code template and emit a filled in template at each call site. Macro function parameters all have type `AstNode`. When a macro is invoked, the expressions supplied as arguments remain unevaluated, and are passed in as AST nodes. Within the body of the macro function, any placeholders in the code template are filled in, or replaced, by interpolating the function's arguments into the template at the placeholder locations.

Macro definitions take the form:

```
macro <name of macro function>(<parameters>) -> AstNode {
  <pre-template logic goes here>
  `<template goes here>`
}
```

All macro functions return a value of type `AstNode`. The backtick-enclosed template is a syntax-literal representation of an `AstNode` - when a backtick-enclosed template is evaluated at a call site, the backtick expression evaluates to an `AstNode` value. The resulting `AstNode` value may be bound to a local variable identifier and inspected/printed if desired, and/or may be immediately returned from the macro function.

Templates may include jinja-style placeholders/"tags" that inject values into the template or evaluate arbitrary code when the template is realized/evaluated. The two placeholder notations are:

- `{{ expression }}` - Value or expression insertion
  {% raw %}
- `{% expression %}` - Code evaluation
  {% endraw %}

The invocation of a macro function injects the returned `AstNode` into the AST - or equivalently, substitutes the returned expression into the source code - at the call site.

For example:
{% raw %}

```
macro defineJsonEncoder(type) {
  `
  fn encode(val: {{ type }}) -> String {
    b = StringBuilder()
    {% for (fieldName, fieldType) in typeof(type).fields { %}
      b << json.encode{{ fieldType }}Field("{{ fieldName }}", val.{{ fieldName }})
    {% } %}
    b.toString
  }
  `
}
{% endraw %}

struct Person { name: String, age: i32 }
defineJsonEncoder(Person)
# this call emits the following
# fn encode(val: Person) -> String {
#   b = StringBuilder()
#   b << json.encodeStringField("name", val.name)
#   b << json.encodeIntField("age", val.age)
#   b.toString
# }

struct Address { name: String, address1: String, address2: String, city: String, state: String, zip: String }
defineJsonEncoder(Address)
# this call emits the following
# fn encode(val: Address) -> String {
#   b = StringBuilder()
#   b << json.encodeStringField("name", val.name)
#   b << json.encodeStringField("address1", val.address1)
#   b << json.encodeStringField("address2", val.address2)
#   b << json.encodeStringField("city", val.city)
#   b << json.encodeStringField("state", val.state)
#   b << json.encodeStringField("zip", val.zip)
#   b.toString
# }
```

## Concurrency

Able supports the following concurrency mechanisms:

1. call stacks
2. channels
3. futures
4. async/await

Futures and async/await are implemented in terms of call stacks and channels.

### Call stacks (threads of execution)

A [call stack](https://en.wikipedia.org/wiki/Call_stack) is the machinery that underlies a running function. The call stack, in conjunction with the [program counter](https://en.wikipedia.org/wiki/Program_counter) captures the running state of a function, as well as the state of the function's caller, the caller's caller, etc.

Conceptually a call stack is a thread of execution that is initiated with a single function call.

Concurrently running call stacks are Able's primary concurrency primitive.

Depending on the compilation target, a call stack may be implemented by a goroutine, a fiber/coroutine, a green thread, or a kernel thread. Though the implementation differences seem significant, call stacks maintain consistent usage semantics across all compilation targets.

#### Semantics

Call stack semantics closely follow Goroutine semantics. The following references explain Goroutine semantics:

- https://golang.org/doc/effective_go.html#goroutines
- http://blog.nindalf.com/how-goroutines-work/
- https://softwareengineering.stackexchange.com/a/222694

Like Goroutines, call stacks...

1. are quick to create.
2. are lightweight, initially using only a few KB of memory per call stack.
3. run concurrently with other call stacks.
4. are scheduled cooperatively rather than preemptively.
5. are scheduled automatically by the language runtime.

Depending on how many CPU cores are available, concurrently running call stacks may be interleaved onto the same CPU core, or they may run in parallel on multiple CPU cores.

New call stacks may be created with the `spawn` function. `spawn` has the function signature `fn spawn(() -> T) -> Future T` and may be used in the style of:

```
futureResult = spawn(fireMissiles)
```

or

```
futureResult = spawn { doThis(); fireMissiles(); doThat() }
```

For example:

```
fn fireMissiles() -> i32 {
  missiles = launchAllMissiles()
  missiles.count
}

numberOfMissiles = spawn(fireMissiles)
```

#### Call stack local variables

Similar to thread local variables in Scala, Java, Clojure, or C#, Able supports call stack local variables.

A call stack local variable is a variable that may be bound to a different value per call stack.

All call stack local variables must be declared at the package level, as they are treated like regular global/package-level variables, with the exception that their value is scoped per call stack, rather than scoped per process.

Call stack local variables are supported through the use of the `CallStackLocal` type.

There are two ways to use the `CallStackLocal` type.

1. Call stack local variables may be declared, and implicitly given an appropriate zero value, as in:

   ```
   <variable name>: CallStackLocal <type>
   ```

   Since a call stack local variable causes the variable to exist in every call stack and none of them will have been assigned an initial value, the call stack local variable is given a default initial value of the zero value of the appropriate type. For example, if the call stack variable is of type `CallStackLocal A`, then the initial value is the zero value for the type A.

2. Call stack local variables may be declared _and_ defined with the `CallStackLocal` constructor function, as in:

   ```
   <variable name> = CallStackLocal(<value expression>)
   ```

   The `CallStackLocal` constructor function has the function signature `CallStackLocal(Thunk T) -> CallStackLocal T`.

   By defining a call stack local variable to have a specified value, the value-generating expression is captured in a `Thunk` value. When the call stack local variable is accessed for the first time from within a call stack, the thunk is evaluated to produce an initial value for the call stack local variable. The generated value is assigned to the call-stack-specific memory location for that call stack local variable. After the initial assignment of a value to a call stack local variable, the variable may be used just like any other variable.

There are special semantics around accessing and assigning to a `CallStackLocal` variable.

1. Assignment to a call stack local variable requires that the type of the right-hand-side of an assignment expression matches the type wrapped by the `CallStackLocal` wrapper type. For example, if `x` is of type `CallStackLocal i32`, then `x = 5` is what an assignment expression should look like, rather than `x = CallStackLocal(5)`.

2. Assigning a value to a call stack local variable only binds the value to that variable name within the scope of the call stack that evaluated the assignment expression.

3. Accessing a call stack local variable returns whatever value is bound to that variable name within the scope of the call stack that evaluated the access expression.

The following examples demonstrate how to use CallStackLocal variables.

```
name: CallStackLocal (String | Nil)   # declaration only; default initial value of nil

name1 = spawn { name = "Bob" }
name2 = spawn { name = "Tom" }
name3 = name
puts("name1 = $name1")
puts("name2 = $name2")
puts("name3 = $name3")

# prints
# name1 = Bob
# name2 = Tom
# name3 = nil
```

and

```
nameCache = CallStackLocal(Map[i32, String](1:"Bob"))   # declare and define

names1 = spawn { nameCache[2] = "Tom"; nameCache.values.join(", ") }
names2 = spawn { nameCache[1] = "Jim"; nameCache.values.join(", ") }
names3 = nameCache.values.join(", ")
puts("names1 = $names1")
puts("names2 = $names2")
puts("names3 = $names3")

# prints
# names1 = Bob, Tom
# names2 = Jim
# names3 = Bob
```

### Channels

Channels provide a means for call stacks to communicate with one another. Channels come in two flavors: unbuffered and buffered.

Unbuffered channels are synchronous - if one side sends, then the other side must receive before the sender can send anything else. In other words, if a sender tries to send a value across the channel, the send operation blocks until the receiving end of the channel receives the value. Similarly, if a receiving call stack is trying to receive on a channel, then the receiver blocks until a sender sends something.

Buffered channels are asynchronous until their buffer is filled, at which point they become synchronous. If a message is sent across the channel and not immediately received, then the message will be buffered until the receiver receives them. Once the receive buffer is full, senders may not send anything else across the channel until the receiver pulls messages out of the buffer. Once the buffer of a buffered channel is full, the channel begins acting like an unbuffered channel in that any additional attempts to send something across the channel will block until the receiver pulls a message out of the channel.

Channels are implemented with the `Channel` type and support two operations: `send` and `receive`.

For example:

```
c = Channel[i32]()	# this is an unbuffered channel
spawn { c.send(5) }
spawn { c.receive |> puts }
```

and

```
c = Channel[i32](1)	# this is a buffered channel, with a buffer of one message
spawn { c.send(5) }
spawn { c.send(10) }
spawn { c.receive |> puts }
# at this point, both send operations will have completed,
# and the receive buffer is full
```

## Reference

### Unary Prefix Operators

- `-` - negation operator performs arithmetic negation
- `!` - logical not operator performs logical negation of its Boolean-valued operand
- `~` - bitwise not operator produces the bitwise complement of its Boolean or Integer-valued operand

### Binary Operators

- `+` - addition
- `-` - subtraction
- `*` - multiplication
- `^` - exponentiation
- `/` - floating point division
- `//` - integer division
- `%` - modulus operator
- `/%` - divmod operator returns a 2-tuple (pair) consisting of the (quotient, remainder)

## Unsolved Problems

- https://github.com/matthiasn/talk-transcripts/blob/master/Hickey_Rich/EffectivePrograms.md
  - ~~ability to cope with sparse data/composable information constructs (heterogeneous lists and maps)~~
- Support something like Rust's questionmark operator: https://m4rw3r.github.io/rust-questionmark-operator ?
- Combined Option/Result type as in V (see https://vlang.io/docs#option) ?

## To do

- Use `&.` as safe navigation operator
- Coroutines will be implemented with call stacks and channels.
- Handle integer overflow like https://golang.org/ref/spec#Integer_overflow

## Not going to do

- Support the implementation of anonymous impls, for example, the ability to implement Iterable by supplying an anonymous impl of Iterator T interface.

  - Something like this?

    ```
    impl Iterable i32 for i32 {
      fn iterator(i: i32) -> impl Iterator i32 for struct { state: i32 } {
        fn next(it: this) -> Option T {

        }
      }
    }
    ```

# Style Guide

- strive for readability
- 2 or 4 space indentation (2 space preferred); no tabs; no mixed indentation in a single file
- open curly brace on same line as if/while/for expressions, function definitions, etc.

# Able Tooling

A single tool, `able`, handles builds, dependency management, package management, running tests, etc.

A single file, `project.cf`, defines a project's build configuration, dependencies, package metadata, etc.

A primary goal of having a single tool is to enable the quick spin-up of a development environment. Download a single binary and start working.

## Building

`able build`

## Read-Eval-Print-Loop (REPL)

`able repl`

## Running

`able run <file>`

## Testing

`able test`

## Package Management

- `able pkg build`
- `able pkg publish`

## Known Issues

These issues are on my todo list, but I don't have a fix for them yet.

- The parser doesn't currently detect empty rule bodies, and the application of a rule that has no body will never terminate.
