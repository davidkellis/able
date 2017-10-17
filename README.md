# Able Language

Able is a programming language that aims to be pleasant to use, relatively small, and relatively fast (in the same ballpark as Java or Go).

Here are a few sample programs to give you a feel for the language:

**Hello World**

```
package main

fn main() {
  puts("Hello world.")
}
```

**Factorial**

```
// recursive
fn factorial(n: UInt) => n < 2 ? 1 : n * factorial(n - 1)

// tail recursive
fn factorial(n) => factorial(n, 1)
fn factorial(n: UInt, product: UInt) -> UInt {
  if n < 2 then product else factorial(n - 1, n * product)
}

// iterative
fn factorial(n: UInt) {
  return 1 if n < 2
  for i in 2...n { n *= i }
  n
}

// reduce
fn factorial(n: UInt) {
  return 1 if n <= 1
  (2..n).reduce(*)
}
```

**Primes**

```
fn primesLessThan(max) => max < 2 ? List() : primesLessThan(max, 2, List())
fn primesLessThan(max: Uint, i: UInt, primesFound: List UInt) -> List UInt {
  return primesFound if i > max
  if primesFound.any? { p => i % p == 0 }
    primesLessThan(max, i + 1, primesFound)
  else
    primesLessThan(max, i + 1, primesFound + i)
}
```

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Identifiers](#identifiers)
- [Packages](#packages)
  - [Importing Packages](#importing-packages)
- [Type Expressions](#type-expressions)
- [Variables](#variables)
  - [Built-In Types](#built-in-types)
- [Strings](#strings)
- [Arrays](#arrays)
- [Maps](#maps)
- [Tuples](#tuples)
  - [Pairs](#pairs)
- [Structs](#structs)
  - [Definition with Positional Fields](#definition-with-positional-fields)
  - [Definition with Named Fields](#definition-with-named-fields)
- [Unions](#unions)
  - [Union definitions referencing predefined types](#union-definitions-referencing-predefined-types)
  - [Union definitions referencing new struct types](#union-definitions-referencing-new-struct-types)
- [Blocks](#blocks)
- [Functions](#functions)
  - [Named function syntax](#named-function-syntax)
  - [Anonymous function syntax](#anonymous-function-syntax)
  - [Lambda expression syntax](#lambda-expression-syntax)
    - [Implicit Lambda Expressions via Placeholder Syntax (a.k.a. Placeholder Lambdas)](#implicit-lambda-expressions-via-placeholder-syntax-aka-placeholder-lambdas)
      - [Special case](#special-case)
  - [Function definition examples](#function-definition-examples)
    - [Define non-generic functions](#define-non-generic-functions)
    - [Define generic functions](#define-generic-functions)
    - [Lambda Expressions](#lambda-expressions)
  - [Function Application](#function-application)
    - [Normal application](#normal-application)
    - [Named argument application](#named-argument-application)
    - [Partial application](#partial-application)
    - [Method application](#method-application)
    - [Pipeline application](#pipeline-application)
    - [Operator application](#operator-application)
      - [Right-associative operators](#right-associative-operators)
      - [Special operators](#special-operators)
    - [Special cases](#special-cases)
- [Interfaces](#interfaces)
  - [Interface Definitions](#interface-definitions)
  - [Interface Implementations](#interface-implementations)
  - [Interface Usage](#interface-usage)
- [Pattern Matching](#pattern-matching)
- [Destructuring](#destructuring)
  - [Assignment Destructuring](#assignment-destructuring)
  - [Parameter Destructuring](#parameter-destructuring)
  - [Pattern Matching Destructuring](#pattern-matching-destructuring)
- [Breakpoints / Non-local Return](#breakpoints--non-local-return)
- [Exceptions](#exceptions)
- [Lazy Evaluation](#lazy-evaluation)
  - [By-name Evaluation (not memoized)](#by-name-evaluation-not-memoized)
  - [By-need Evaluation (memoized)](#by-need-evaluation-memoized)
- [Concurrency](#concurrency)
  - [Call stacks (threads of execution)](#call-stacks-threads-of-execution)
    - [Semantics](#semantics)
    - [Call stack local variables](#call-stack-local-variables)
  - [Channels](#channels)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Identifiers

Variable and function identifiers must conform to the pattern
`[a-zA-Z0-9][a-zA-Z0-9_]*[a-zA-Z0-9_?!]?`

Package identifiers must conform to the pattern
`[a-zA-Z0-9][a-zA-Z0-9_]*[a-zA-Z0-9_]?`

## Packages

There must be one, and only one, package definition per file, specified at the top of the file.

```
package io
```

A package introduces a new variable scope that forms the root scope for any other scopes introduced within the package.

### Importing Packages

1. Import package:

  ```
  import io
  io.puts("hi")
  ```

2. Wildcard import:

  ```
  import io.*
  puts("enter your name:")
  name = gets()
  ```

3. Import individual types/functions from a package:

  ```
  import io.{puts, gets}
  puts("enter your name:")
  name = gets()
  ```

4. Import package and/or individual types/functions from a package with a different local identifier:

  ```
  import io.{puts as p}
  import internationalization as i18n.{Unicode}
  p("${Unicode.abbreviation} is Unicode; ${i18n.Ascii.abbreviation} is Ascii")
  ```

## Type Expressions

A type is a name given to a set of values, and every value has an associated type. For example, `Boolean` is the name given to the set `{true, false}`, and the value `true` is of type `Boolean`.  `TwoBitUnsignedInt` might be the type name we give to the set `{0, 1, 2, 3}`, such that `3` would be a value of type `TwoBitUnsignedInt`.

A type is denoted by a type expression. All types are parametric types, in that all types have zero or more type parameters.

Type parameters may be bound or unbound. A bound type parameter is a type parameter for which either a named type variable or a preexisting type name has been substituted. An unbound type parameter is a type parameter that is either unspecified or substituted by the placeholder type variable, denoted by `_`.

A type that has all its type parameters bound is called a concrete type. A type that has any of its type parameters unbound is called a polymorphic type, and a type expression that represents a polymorphic type is called a type constructor.

References:

- https://www.haskell.org/tutorial/goodies.html has a good article on types, type expressions, type variables, etc.

## Variables

Variables are defined with the following syntax:
```
<variable name>: <type name> = <value expression>
```
and if the type can be inferred, then the definition may be shortened to:
```
<variable name> = <value expression>
```

### Built-In Types

- Boolean
- Integer types - Int8, Int16, Int32 (Int), Int64, UInt8 (Byte), UInt16, UInt32 (UInt), UInt64
- Floating point types - Float, Double
- String
- Array
- Map
- Tuple
- Struct
- Union
- Function
- Interface
- Thunk
- Lazy
- CallStackLocal
- Channel

## Strings

All string literals are double quoted.

```
name = "Herbert"
greeting = "Hello $name"
greeting2 = "Hello ${name}"
```

## Arrays

```
x = Array(1,2,3)
x: Array Float = Array(5.6, 5.7, 8)
```

## Maps

```
m = Map(1 :: "Foo", 2 :: "Bar")
m[3] = "Baz"
m << 4::"Qux"
```

## Tuples

```
record = (1, "Foo")
if record._1 == 1 then puts("you're in first place!")
```

`record` is of type (Int, String)

### Pairs

Pair syntax is just syntactic sugar for expressing 2-tuples.

`(1, "Foo")` can be written as `1 :: "Foo"` or `1::"Foo"`

## Structs

Struct type definitions define both a type and a constructor function of the same name as the type.

### Definition with Positional Fields
```
struct Foo T { Int, Float, T }
struct Foo T { Int, Float, T, }
struct Foo T {
  Int,
  Float,
  T,
}
struct Foo T {
  Int
  Float
  T
}
```

### Definition with Named Fields
```
struct Foo T { x: Int, y: Float, z: T }
struct Foo T { x: Int, y: Float, z: T, }
struct Foo T {
  x: Int,
  x: Float,
  y: T,
}
struct Foo T {
  x: Int
  y: Float
  z: T
}
```

Create instances of structs

```
Foo(1,2,t1)
Foo(x=1, y=2, z=t1)
Foo(
  x = 1,
  y = 2,
  z = t1
)
Foo(
  x = 1
  y = 2
  z = t1
)
```

Struct definitions can only appear in package scope.

## Unions

### Union definitions referencing predefined types

```
union String? = String | Nil

struct Red { level: Int }
struct Green { level: Int }
struct Blue { level: Int }
union Color = Red | Green | Blue
```

### Union definitions referencing new struct types

```
union House = SmallHouse { sqft: Float }
  | MediumHouse { sqft: Float }
  | LargeHouse { sqft: Float }

union House =
  | SmallHouse { sqft: Float }
  | MediumHouse { sqft: Float }
  | LargeHouse { sqft: Float }
```

Union definitions that define new types is syntactic sugar. For example:

```
union House = SmallHouse { sqft: Float }
  | MediumHouse { sqft: Float }
  | LargeHouse { sqft: Float }
```

is internally translated into:

```
struct SmallHouse { sqft: Float }
struct MediumHouse { sqft: Float }
struct LargeHouse { sqft: Float }
union House = SmallHouse | MediumHouse | LargeHouse
```

## Blocks

A block is a sequence of expressions enclosed in a set of curly braces. A block is itself an expression that evaluates to the value returned by the last expression in the block. A block introduces a new variable scope, but may reference identifiers from any enclosing scopes.

For example:
```
{
  a = 5 + 7
  puts(a)
  a
}
```
or
```
{ a = 5 + 7; puts(a); a }
```

## Functions

Functions may be defined with one of three different function definition syntaxes:

1. Named function syntax
2. Anonymous function syntax
3. Lambda expression syntax

All functions are first class objects, and each of the three syntaxes produces the same kind of function object.

In each of the three syntaxes, the return type may be omitted unless the return type is ambiguous and type inference is unable to determine what the return type is supposed to be. If type inference fails to determine the return type, then the return type must be explicitly provided.

In the two syntaxes that provide a means to capture an optional type parameter list, the type parameter list may also capture constraints on type parameters.

### Named function syntax

Named function syntax allows one to define a named function in the current lexical scope. Named function syntax take the following form:

`fn <function name>[<optional type paramter list>](<parameter list>) -> <optional return type> { <function body> }`
or
`fn <function name>[<optional type paramter list>](<parameter list>) -> <optional return type> => <function body>`

The style allows type constraints to be captured in the type parameter list (e.g. `fn convertToString[T: Parsable + Stringable](a: T) => toString(a)` )

### Anonymous function syntax

Anonymous function syntax allows one to define a function object that may be immediately supplied as an argument to a function, or assigned to a variable for later use. Anonymous function syntax takes the following form:

`fn[<optional type paramter list>](<parameter list>) -> <optional return type> { <function body> }`
or
`fn[<optional type paramter list>](<parameter list>) -> <optional return type> => <function body>`

Additionally, you can omit the leading `fn` at the expense of not being able to capture constraints on generic type parameters:

`(<parameter list>) -> <optional return type> { <function body> }`
or
`(<parameter list>) -> <optional return type> => <function body>`

The style with the explicit free type parameter list is the only way to capture type constraints (e.g. `fn[T: Parsable + Stringable](a: T) => toString(a)` )

### Lambda expression syntax

Lambda expressions are defined in the lambda expression syntax:
```{ <paramter list> -> <return type> => expression }```

Lambda expressions may be defined without any parameters, but depending on the context, the "fat arrow", `=>`, may be necessary to disambiguate whether the expression represents a block or a zero-argument lambda.

As an example of an ambiguous expression, `randomInt = { Random.int() }` may represent an expression of type `Int`, or it may represent an expression of type `()->Int`.

When the compiler can't infer the correct return type, the expression must be made more explicit. For example, to disambiguate `randomInt = { Random.int() }`, either of the following would work:
```
randomInt = { => Random.int() }
```
or
```
randomInt: fn()->Int = { Random.int() }
```

#### Implicit Lambda Expressions via Placeholder Syntax (a.k.a. Placeholder Lambdas)

Lambda expressions may also be represented implicitly with a special convenience syntax called placeholder syntax.

An expression may be parameterized with placeholder symbols - each of which may be either an underscore, _, or a numbered underscore, e.g. _1, _2, etc.

A placeholder symbol may be used in place of any value-representing subexpression, such as an argument to a function, an operator argument, a conditional variable in an if statement, etc. Placeholder symbols may not be used in place of language keywords, e.g. in place of "if", "package", etc.

The scope of a lambda expression implicitly defined through the use of placeholder symbols extends to the smallest subexpression that, if considered as the body of an anonymous function, would represent a function with the same function signature as what the containing expression expects.

For example:

`_ + 1` is treated as `{ x => x + 1 }`

`_ * _` is treated as `{ x, y => x * y }`

`if _ x else y` is treated as `{ c => if c x else y }`

`_.map(_ * 5)` is treated as `{ x => x.map( { y => y * 5 } ) }` where the subexpression `_ * 5` is treated as an inner lambda expression, and the full expression is treated as lambda expression

`_.reduce(_, _ * _)` is treated as `{ x, y => x.reduce(y, { a, b => a * b }) }` where `_ * _` is treated as an inner lambda expression, and the full expression is treated as a lambda expression

##### Special case

If a placeholder lambda is enclosed within a block, and the scope of the placeholder lambda extends to the boundaries of the block’s delimiters, then the full block is considered a lambda expression having the same parameters as that of the contained placeholder lambda.

For example:
`5.times { puts("Greeting number $_") }` is treated as `5.times { i => puts("Greeting number $_") }`


### Function definition examples

#### Define non-generic functions

With explicit return type:

1. fn createPair(a, b: Int) -> Pair Int { Pair(a, b) }<br>
   fn createPair(a, b: Int) -> Pair Int => Pair(a, b)

3. createPair = fn(a, b: Int) -> Pair Int { Pair(a, b) }<br>
   createPair = fn(a, b: Int) -> Pair Int => Pair(a, b)

4. createPair = (a, b: Int) -> Pair Int { Pair(a,b) }<br>
   createPair = (a, b: Int) -> Pair Int => Pair(a, b)

5. createPair = { a, b: Int -> Pair Int => Pair(a,b) }

With inferred return type:

1. fn createPair(a, b: Int) { Pair(a, b) }<br>
   fn createPair(a, b: Int) => Pair(a, b)

2. createPair = fn(a, b: Int) { Pair(a, b) }<br>
   createPair = fn(a, b: Int) => Pair(a, b)

3. createPair = (a, b: Int) { Pair(a,b) }<br>
   createPair = (a, b: Int) => Pair(a, b)

4. createPair = { a, b: Int => Pair(a,b) }

All the functions defined above are of type:
`(Int, Int) -> Pair Int`

#### Define generic functions

With explicit free type parameter and explicit return type:

1. fn createPair T(a, b: T) -> Pair T { Pair(a, b) }<br>
   fn createPair T(a, b: T) -> Pair T => Pair(a, b)

2. createPair = fn[T](a, b: T) -> Pair T { Pair(a, b) }<br>
   createPair = fn[T](a, b: T) -> Pair T => Pair(a, b)

With explicit free type parameter and inferred return type:

1. fn createPair T(a, b: T) { Pair(a, b) }<br>
   fn createPair T(a, b: T) => Pair(a, b)

2. createPair = fn[T](a, b: T) { Pair(a, b) }<br>
   createPair = fn[T](a, b: T) => Pair(a, b)

With implied free type parameter and explicit return type:

1. createPair = (a, b: T) -> Pair T { Pair(a,b) }<br>
   createPair = (a, b: T) -> Pair T => Pair(a, b)

2. createPair = { a, b: T -> Pair T => Pair(a, b) }

With implied free type parameter and inferred return type:

1. createPair = (a, b: T) { Pair(a,b) }<br>
   createPair = (a, b: T) => Pair(a, b)

2. createPair = { a, b: T => Pair(a, b) }

All the functions defined above are of type:
`(T, T) -> Pair T`

#### Lambda Expressions

```
createPair = { a, b: Int -> Pair Int => Pair(a, b) }
createPair = { a, b: Int => Pair(a, b) }
createPair = { a, b: T -> Pair T => Pair(a, b) }
createPair = { a, b: T => Pair(a, b) }
```

### Function Application

#### Normal application

```
f(1,2,3)
f(1,2,*xs)
buildPair(1,2)
map(Array(1,2,3), square)
```

#### Named argument application

Given the following function definition:
```
fn createPerson(age: Int, name: String, address: Address) -> Person { ... }
```

the `createPerson` function may be invoked in any of the following ways:

```
createPerson(name = "Bob", age = 20, address = Address("123 Wide Street"))

createPerson(
  name = "Bob",
  age = 20,
  address = Address("123 Wide Street")
)

createPerson(
  name = "Bob"
  age = 20
  address = Address("123 Wide Street"))
)

createPerson(
  name = "Bob", age = 20
  address = Address("123 Wide Street"),
)
```

#### Partial application

Given:
```
f = fn(a,b,c,d,e: Int) => Tuple5(a,b,c,d,e)
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

Partial application of the first, third, and fifth arguments; returns a function with the signature (Int, Int) -> Tuple5
```
f(1, _, 2, _, 3)
```

Partial application of the first, third, and fifth arguments; returns a function with the signature (Int) -> Tuple5
```
f(1, _1, 2, _1, 3)
```

Partial application of all the arguments; returns a function with the signature () -> Tuple5
```
f(1, 2, 3, 4, 5,)
or
f(,1, 2, 3, 4, 5)
```

Partial application of a named argument
```
f(c = 30)
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
`struct Coord {Int, Int}`

`5 |> Coord(_1, _1)`
is equivalent to
`Coord(5, 5)`

`1 |> plusTwo |> f(1, 2, _, 4, 5)`
is equivalent to
`f(1, 2, plusTwo(1), 4, 5)`

#### Operator application

Operators are functions too. The only distinction between operator functions and non-operator functions is that operator functions additionally have an associated operator precedence and are associative. Non-operator functions are non-associative, and have no operator precedence.

By default, operators are left-associative. For example, `a + b + c` is evaluated as `((a + b) + c)`

Operators that have a trailing apostrophe are right-associative. For example, `a +' b +' c` is evaluated as `(a +' (b +' c))`

##### Right-associative operators

1. Exponentiation operator: `**`
   The exponentiation operator is a right-associative operator with precedence that is higher than addition, subtraction, multiplication, and division.

##### Special operators

Some operators have special semantics, and may not be overridden.

1. Assignment operator: `=`
   The assignment operator is used in assignment expressions to assign the value expressed in the right hand side (RHS) of the assignment expression to the identifier or identifiers specified in the left hand side (LHS) of the assignment expression. The assignment operator is right-associative. The assignment operator is not a function, and as a result, is not a first-class object that may be used as a value.

#### Special cases

1. When invoking a function of arity-1 with method syntax, the empty parenthesis may be omitted.

   For example, either of the following are acceptable:
   - `Array(1,2,3).length()`
   - `Array(1,2,3).length`

   This only works with functions of arity-1.

   Attempting to invoke a function of arity-2+ using method syntax without supplying any parenthesis will yield a partially applied function.

   For example, given `fn add(a, b: Int) -> Int { a + b }`, the expression `1.add` represents a partially applied function object of type (Int)->Int.

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
   ​Array(1,2,3).map() { x => x^2 }
   ```
   may be shortened to
   ```
   Array(1,2,3).map { x => x^2 }
   ```

4. If a function is defined with a parameter of a union type, then the function may be called by suppling either a value whose type is the union type or a value whose type is any one of the members of the union.

   For example, given
   ```
   struct Red { level: Int }
   struct Green { level: Int }
   struct Blue { level: Int }
   union Color = Red | Green | Blue

   fn paintHouse(c: Color) -> Boolean { ... }
   ```
   the `paintHouse` function may be invoked with either of the following:
   ```
   paintHouse(Red(5))   // value of type Red
   ```
   OR
   ```
   color: Color = Red(5)
   paintHouse(color)   // value of type Color
   ```

5. When passing a value of a union type to a function, the function must either be defined in terms of a parameter of the union type, or be defined for each member of the union type. If both alternatives are fully defined, then the function defined in terms of the union type is preferred.

   For example, given
   ```
   struct Red { level: Int }
   struct Green { level: Int }
   struct Blue { level: Int }
   union Color = Red | Green | Blue
   ```
   the expression
   ```
   color: Color = Red
   paintHouse(color)
   ```
   is allowable only if:
   - `paintHouse(Color)` has been defined
   <br>OR
   - `paintHouse(Red)` has been defined, and<br>
   `paintHouse(Green)` has been defined, and<br>
   `paintHouse(Blue)` has been defined

   If `paintHouse(Color)` as well as `paintHouse(Red)`, `paintHouse(Green)`, and `paintHouse(Blue)` are all defined, then `paintHouse(Color)` is preferred.

6. If a function has been defined for both a union type, e.g. `union U = U1 | U2`, and an individual member of the union type, e.g. `U1`, and the function is called with a value of type `U1`, then the function that gets called is the one defined for the individual member of the union type.

   For example, given
   ```
   struct Red { level: Int }
   struct Green { level: Int }
   struct Blue { level: Int }
   union Color = Red | Green | Blue

   fn paintHouse(c: Color) -> Boolean { ... }
   fn paintHouse(c: Red) -> Boolean { ... }
   ```
   the expression `paintHouse(Red(5))` invokes the function defined by `fn paintHouse(c: Red) -> Boolean { ... }`.

## Interfaces

An interface is a set of functions that are guaranteed to be defined for a given type.

An interface can be used to represent either (1) an abstract type, or (2) a constraint on a type or type variable.

When used as an abstract type, the interface functions as a lens through which we see and interact with some underlying type that implements the interface. This form of polymorphism allows us to use different types as if they were the same type. For example, if the concrete types Int and Float both implement an interface called Numeric, then one could create an `Array Numeric` consisting of both `Int` and `Float` values.

When used as a constraint on a type, the interpretation is that some context needs to reference the underlying type, while also making a claim or guarantee that the underlying type implements the interface.

### Interface Definitions

Interfaces are defined with the following syntax:
```
interface A [B C ...] for D [E F ...] {
	fn foo(T1, T2, ...) -> T3
	...
}
```
where A is the name of the interface optionally parameterized with type parameters, B, C, etc., and where D is a type variable deemed to implement interface A.

The type expression in the for clause is called the interface's *self type*. The self type is a concrete or polymorphic type representing the type that will implement the interface. In the syntax above, `D [E F ...]` represents the self type.

D may represent either a concrete type or a polymorphic type depending on whether D's type parameters (if applicable) are bound or unbound (see [Type Expressions](#type-expressions) for more on bound vs. unbound type parameters).

Here are a few example interface definitions:

```
// Stringable interface
interface Stringable for T {
  fn toString(T) -> String
}

// Comparable interface
interface Comparable for T {
  fn compare(T, T) -> Int
}

// Enumerable interface
interface Enumerable T for E {
  fn each(E, T -> Unit) -> Unit
  fn iterator(e: E) -> Iterator T = buildIterator { e.each(yield) }
}

// Mappable interface (Functor to you Haskell folks)
interface Mappable A for M _ {
  fn map(m: M A, convert: A -> B) -> M B
}
```

### Interface Implementations

Interface implementations are defined with the following syntax:
```
impl [X, Y, Z, ...] A <B C ...> for D <E F ...> {
	fn foo(T1, T2, ...) -> T3 { ... }
	...
}
```
where X, Y, and Z are free type parameters and type constraints that are scoped to the implementation of the interface (i.e. the types bound to those type variables are constant within a specific implementation of the interface) and enclosed in square brackets, A is the name of the interface parameterized with type parameters B, C, etc. (if applicable), and where `D <E F ...>` is the type that is implementing interface A.

Here are some example interface implementations:

```
struct Person { name: String, address: String }
impl Stringable for Person {
  fn toString(p: Person) -> String => "Name: ${p.Name}\nAddress: ${p.Address}"
}

impl Comparable for Person {
  fn compare(p1, p2) => compare(p1.Name, p2.Name)
}

impl Enumerable T for Array T {
  fn each(array: Array T, visit: T -> Unit) -> Unit {
    i = 0
    length = array.length()
    while i < length {
      visit(array[i])
      i += 1
    }
  }
}

struct OddNumbers { start: Int, end: Int }
impl Enumerable Int for OddNumbers {
  fn each(oddNumbersRange: OddNumbers, visit: Int -> Unit) {
    for i in start..end { visit(i) }
  }
}

impl Mappable A for Array {
  fn map(arr: Array A, convert: A -> B) -> Array B {
    ab = ArrayBuilder(0, arr.length)  # length, capacity
    arr.each { element => convert(element) |> ab.add }
    ab.toArray
  }
}
```

### Interface Usage

When calling a function defined in an interface, a client may supply either (1) a value of a type that implements the interface or (2) a value of the interface type to the function in any argument that is typed as the interface's self type.

For example, given:
```
interface Enumerable T for E {
  fn each(E, T -> Unit) -> Unit
  ...
}
```
a client may call the `each` function by supplying a value of type `Enumerable T` as the first argument to `each`, like this:
```
// assuming Array T implements the Enumerable T interface
e: Enumerable Int = Array(1, 2, 3)    // e has type Enumerable Int
each(e, puts)
```
or `each` may be called by supplying a value of type `Array T` as the first argument, like this:
```
// assuming Array T implements the Enumerable T interface
e = Array(1, 2, 3)    // e has type Array Int
each(e, puts)
```

## Pattern Matching

Pattern matching expression take the following general form:

```
<expression0> match {
  case <pattern1> => <expression1>
  case <pattern2> => <expression2>
  case _ => <expression3>
}
```

For example:

```
union House =
  | TinyHouse { sqft: Float }
  | SmallHouse { sqft: Float }
  | MediumHouse { sqft: Float }
  | LargeHouse { sqft: Float }
  | HugeHouse { sqft: Float, pools: Int }

fn buildHouse(h: House) => h match {
  case TinyHouse | SmallHouse => puts("build a small house")
  case m: MediumHouse => puts("build a modest house - $m")
  case LargeHouse(area) => puts ("build a large house of $area sq. ft.")
  case HugeHouse(_, poolCount) => puts ("build a huge house with $poolCount pools!")
}
```

## Destructuring

Destructuring is the binding of variable names to positional or named values within a data structure. Destructuring may be used in assignment expressions, function parameter lists, and pattern matching case expressions.

Structs, tuples, pairs, arrays, and maps support destructuring. Unions do not support destructuring.

The following three sections demonstrate destructuring in three different contexts. The examples assume the following Person structure definition:

```
struct Person {
  name: String
  height: Float
  weight: Float
  age: Int
}
```

### Assignment Destructuring

```
fn printPersonInfo(p: Person) {
  Person(n,h,w,a) = p
  puts("name=$n   height=$h   weight=$w   age=$a")
}

printPersonInfo(Person("Jim", 6.0833, 170, 25))
```

### Parameter Destructuring

```
fn printPersonInfo(p: Person(n, h, w, a)) {
  puts("name=$n   height=$h   weight=$w   age=$a")
}

printPersonInfo(Person("Jim", 6.0833, 170, 25))
```

### Pattern Matching Destructuring

```
// assuming h is of type House
h match {
  case TinyHouse | SmallHouse => puts("build a small house")
  case m: MediumHouse => puts("build a modest house - $m")
  case LargeHouse(area) => puts ("build a large house of $area sq. ft.")
  case HugeHouse(poolCount=pools) => puts ("build a huge house with $poolCount pools!")
}
```

## Breakpoints / Non-local Return

Breakpoints are labels that correspond to a position in the call stack, such that, if a `break <label> <value>` statement is evaluated, the call stack is unwound up to the point at which the breakpoint was established, and the breakpoint expression resolves to the value supplied to the `break` statement.

Breakpoint expressions take the following form:

```
breakpoint <label> {
  ...
  break <label> <value>
  ...
}
```

To following example demonstrates an implementation of short circuiting logic, achieved via breakpoints. The example implements the `all?` function defined within the `Enumerable` interface. The combination of the breakpoint and the `break` within the `each` block implements immediate return from `all?` upon the first observation of a false evaluation of the predicate function. This short-circuiting logic prevents further unnecessary evaluations of the predicate function once the first observation of a false return value is observed:

```
fn all?(enumerable: E, predicate: T -> Boolean) -> Boolean {
  breakpoint :stop {
    enumerable.each { t => break :stop false unless predicate(t) }
    true
  }
}
```

## Exceptions

```
struct Foo {code: Int, msg: String}

fn pp(w: Widget) {
  raise Foo(msg = "", code=5) if w.type.undefined?
}

fn foo() {
  bar()
  try {
    baz()
  } catch {
    case Foo(m=msg) => "puts foo error!"
  }
  qux()
}

fn main() {
  pp(Widget(name = "card shuffler"))
} catch {
  case f: Foo(m=msg) => puts "foo error = $f"
  case _ => puts "other error"
} ensure {
  puts("ensure block called!")
}
```

Expressions that need to be evaluated prior to function return should be placed in an ensure block. Ensure blocks may be used to suffix try/catch expression, or they may be used to suffix function definitions.

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
// 1. assignment to variable of type Thunk T
mail: Thunk Mail = sendAndReceiveMail()

// 2. supplying an expression to a function parameter of type Thunk T
fn runAfterDelay[T](expr: Thunk T) -> T {
  sleep(10.seconds)
  expr
}

// 3. using the Thunk function
Thunk(sendAndReceiveMail())
```

References:

 - https://en.wikipedia.org/wiki/Thunk
 - https://wiki.haskell.org/Thunk

### By-need Evaluation (memoized)

In by-need evaluation, the expression is evaluated once, on the first occasion that the expression's value is needed, and the resulting value is memoized for future use.

By-need evaluation is supported through the `Lazy` marker type.

The `Lazy` marker type has the same semantics as the `Thunk` marker type, with the exception that the value produced by the first access is memoized, such that subsequent accesses of the value return the same instance of the value.

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

New call stacks may be created with the `spawn` function. `spawn` has the function signature `fn spawn(() -> T) -> Lazy T` and may be used in the style of:
```
lazyResult = spawn(fireMissiles)
```
or
```
lazyResult = spawn { doThis(); fireMissiles(); doThat() }
```

For example:
```
fn fireMissiles() -> Int {
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

1. Call stack local variables may be declared, but left undefined, so long as the type representing the call-stack-local value is a union type with `Nil` as one of its members, as in:
   ```
   <variable name>: CallStackLocal (<type> | Nil)
   ```

   `Nil` must be one of the members of the wrapped type union because the declaration of a call stack local variable causes the variable to exist in every call stack and none of them will have been assigned a value, so they are all implicitly assigned a default value of `nil`.

2. Call stack local variables may be declared *and* defined with the `CallStackLocal` constructor function, as in:
   ```
   <variable name> = CallStackLocal(<value expression>)
   ```

   The `CallStackLocal` constructor function has the function signature `CallStackLocal(Thunk T) -> CallStackLocal T`.

   By defining a call stack local variable to have a specified value, the value-generating expression is captured in a `Thunk` value. When the call stack local variable is accessed for the first time from within a call stack, the thunk is evaluated to produce an initial value for the call stack local variable. The generated value is assigned to the call-stack-specific memory location for that call stack local variable. After the initial assignment of a value to a call stack local variable, the variable may be used just like any other variable.

There are special semantics around accessing and assigning to a `CallStackLocal` variable.

1. Assignment to a call stack local variable requires that the type of the right-hand-side of an assignment expression matches the type wrapped by the `CallStackLocal` wrapper type. For example, if `x` is of type `CallStackLocal Int`, then `x = 5` is what an assignment expression should look like, rather than `x = CallStackLocal(5)`.

2. Assigning a value to a call stack local variable only binds the value to that variable name within the scope of the call stack that evaluated the assignment expression.

3. Accessing a call stack local variable returns whatever value is bound to that variable name within the scope of the call stack that evaluated the access expression.

The following examples demonstrate how to use CallStackLocal variables.
```
name: CallStackLocal (String | Nil)   // declaration only

name1 = spawn { name = "Bob" }
name2 = spawn { name = "Tom" }
name3 = name
puts("name1 = $name1")
puts("name2 = $name2")
puts("name3 = $name3")

// prints
// name1 = Bob
// name2 = Tom
// name3 = nil
```
and
```
nameCache = CallStackLocal(Map[Int, String](1::"Bob"))   // declare and define

names1 = spawn { nameCache[2] = "Tom"; nameCache.values.join(", ") }
names2 = spawn { nameCache[1] = "Jim"; nameCache.values.join(", ") }
names3 = nameCache.values.join(", ")
puts("names1 = $names1")
puts("names2 = $names2")
puts("names3 = $names3")

// prints
// names1 = Bob, Tom
// names2 = Jim
// names3 = Bob
```

### Channels

Channels provide a means for call stacks to communicate with one another. Channels come in two flavors: unbuffered and buffered.

Unbuffered channels are synchronous - if one side sends, then the other side must receive before the sender can send anything else. In other words, if a sender tries to send a value across the channel, the send operation blocks until the receiving end of the channel receives the value. Similarly, if a receiving call stack is trying to receive on a channel, then the receiver blocks until a sender sends something.

Buffered channels are asynchronous until their buffer is filled, at which point they become synchronous. If a message is sent across the channel and not immediately received, then the message will be buffered until the receiver receives them. Once the receive buffer is full, senders may not send anything else across the channel until the receiver pulls messages out of the buffer. Once the buffer of a buffered channel is full, the channel begins acting like an unbuffered channel in that any additional attempts to send something across the channel will block until the receiver pulls a message out of the channel.

Channels are implemented with the `Channel` type and support two operations: `send` and `receive`.

For example:

```
c = Channel[Int]()	// this is an unbuffered channel
spawn { c.send(5) }
spawn { c.receive |> puts }
```
and
```
c = Channel[Int](1)	// this is a buffered channel, with a buffer of one message
spawn { c.send(5) }
spawn { c.send(10) }
spawn { c.receive |> puts }
// at this point, both send operations will have completed,
// and the receive buffer is full
```
