# Able Language

## Identifiers

Variable identifiers may take the following form:
`[a-zA-Z0-9][a-zA-Z0-9_]*[a-zA-Z0-9_?]?`

## Packages

package io

There must be one, and only one, package definition per file, preferably at the top of the file.

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
  p("#{Unicode.abbreviation} is Unicode; #{i18n.Ascii.abbreviation} is Ascii")
  ```

## Type Expressions

A type is a name given to a set of values, and every value has an associated type. For example, `Boolean` is the name given to the set `{true, false}`, and the value `true` is of type `Boolean`.  `TwoBitUnsignedInt` might be the type name we give to the set `{0, 1, 2, 3}`, and `3` could be a value of type `TwoBitUnsignedInt`.

A type is denoted by a type expression. All types are parametric types, in that all types have zero or more type parameters.

Type parameters may be bound or unbound. A bound type parameter is a type parameter for which either a named type variable or a preexisting type name has been substituted. An unbound type parameter is a type parameter that is either unspecified or substituted by the placeholder type variable, denoted by `_`.

A type that has all its type parameters bound is called a concrete type. A type that has any of its type parameters unbound is called a polymorphic type, and a type expression that represents a polymorphic type is called a type constructor.

References:

- https://www.haskell.org/tutorial/goodies.html has a good article on types, type expressions, type variables, etc.

## Strings

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

// Note: need to define structs here.

Struct type definitions define both a type and a constructor function of the same name as the type)

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
`randomInt = { => Random.int() }`
or
`randomInt: fn()->Int = { Random.int() }`

#### Implicit Lambda Expressions via Placeholder Syntax (a.k.a. Placeholder Lambdas)

Lambda expressions may also be represented implicitly with a special convenience syntax called placeholder syntax.

An expression may be parameterized with placeholder symbols - each of which may be either an underscore, _, or a numbered underscore, e.g. _1, _2, etc.

A placeholder symbol may be used in place of any value-representing subexpression, such as an argument to a function, an operator argument, a conditional variable in an if statement, etc. Placeholder symbols may not be used in place of language keywords, e.g. in place of “if”, “package”, etc.

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

1. fn createPair(a, b: Int) -> Pair Int { Pair(a, b) }
   fn createPair(a, b: Int) -> Pair Int => Pair(a, b)
2. createPair = fn(a, b: Int) -> Pair Int { Pair(a, b) }
   createPair = fn(a, b: Int) -> Pair Int => Pair(a, b)
3. createPair = (a, b: Int) -> Pair Int { Pair(a,b) }
   createPair = (a, b: Int) -> Pair Int => Pair(a, b)
4. createPair = { a, b: Int -> Pair Int => Pair(a,b) }

With inferred return type:

1. fn createPair(a, b: Int) { Pair(a, b) }
   fn createPair(a, b: Int) => Pair(a, b)
2. createPair = fn(a, b: Int) { Pair(a, b) }
   createPair = fn(a, b: Int) => Pair(a, b)
3. createPair = (a, b: Int) { Pair(a,b) }
   createPair = (a, b: Int) => Pair(a, b)
4. createPair = { a, b: Int => Pair(a,b) }
5. createPair = Pair(_, 0) + Pair(0, _)

All the functions defined above are of type:
`(Int, Int) -> Pair Int`

#### Define generic functions

With explicit free type parameter and explicit return type:

1. fn createPair T(a, b: T) -> Pair T { Pair(a, b) }
   fn createPair T(a, b: T) -> Pair T => Pair(a, b)
2. createPair = fn[T](a, b: T) -> Pair T { Pair(a, b) }
   createPair = fn[T](a, b: T) -> Pair T => Pair(a, b)

With explicit free type parameter and inferred return type:

1. fn createPair T(a, b: T) { Pair(a, b) }
   fn createPair T(a, b: T) => Pair(a, b)
2. createPair = fn[T](a, b: T) { Pair(a, b) }
   createPair = fn[T](a, b: T) => Pair(a, b)

With implied free type parameter and explicit return type:

1. createPair = (a, b: T) -> Pair T { Pair(a,b) }
   createPair = (a, b: T) -> Pair T => Pair(a, b)
2. createPair = { a, b: T -> Pair T => Pair(a, b) }

With implied free type parameter and inferred return type:

1. createPair = (a, b: T) { Pair(a,b) }
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

#### Normal Application

```
f(1,2,3)
f(1,2,*xs)
buildPair(1,2)
map(Array(1,2,3), square)
```

##### Special cases

1. If supplying a lambda expression to a function as its last argument, then the lambda may be supplied immediately following the closing paren surrounding the argument list.

   For example:
`​Array(1,2,3).reduce(0) { sum, x => sum + x }`

2. In cases where a lambda expression is supplied immediately following the closing paren of a function call, if the argument list is empty, then the parenthesis may be omitted.

   For example,
`​Array(1,2,3).map() { x => x^2 }`
may be shortened to
`Array(1,2,3).map { x => x^2 }`

#### Named argument application

Given the following function definition:
`fn createPerson(age: Int, name: String, address: Address) -> Person { ... }`

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
`f = fn(a,b,c,d,e: Int) => Tuple5(a,b,c,d,e)`

Partial application of the trailing arguments
`f(,1,2,3)`

Explicit partial application of the leading arguments
`f(1,2,3,)`

Implied partial application of the leading arguments
`f(1,2,3)`

Partial application of the first, third, and fifth arguments; returns a function with the signature (Int, Int) -> Tuple5
`f(1, _, 2, _, 3)`

Partial application of the first, third, and fifth arguments; returns a function with the signature (Int) -> Tuple5
`f(1, _1, 2, _1, 3)`

Partial application of a named argument
`f(c = 30)`

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

By default, operators are left-associative. For example, `a + b + c` is evaluated as `(a + b) + c`

Just as in Scala, operators that have a trailing colon are right-associative.

##### Special operators

Some operators have special semantics, and may not be overridden.

1. Assignment operator: `=`
   The assignment operator is used in assignment expressions to assign the value expressed in the right hand side (RHS) of the assignment expression to the identifier or identifiers specified in the left hand side (LHS) of the assignment expression. The assignment operator is right-associative. The assignment operator is not a function, and as a result, is not a first-class object that may be used as a value.
2. Exponentiation operator: `**`
   The exponentiation operator is a right-associative operator with precedence that is higher than addition, subtraction, multiplication, and division.

## Interfaces

An interface is a set of functions that are guaranteed to be defined for a given type or type constructor.

An interface can be used as either a constraint on a type, or a type expression. When used as a constraint on a type, the type being constrained is deemed to implement the interface. When used as a type expression, the interface represents a constrained view of the type being represented.

Interfaces are defined with the following syntax:
```
interface A [B C ...] for D [E F ...] {
	fn foo(T1, T2, ...) -> T3
	...
}
```
where A is the name of the interface optionally parameterized with type parameters, B, C, etc., and where D is a type variable deemed to implement interface A.

The type expression in the for clause is called the interface's "self type variable". The self type variable is a type variable representing the type or type constructor that will implement the interface. In the syntax above, `D [E F ...]` is the self type variable.

D may represent either a concrete type or a type constructor depending on whether D's type parameters (if applicable) are specified with named types and type variables (e.g. E, F, Int, Float, etc.) or left unspecified with type placeholders (i.e. underscores, `_`). If D's type parameters are fully specified with named types and type variables, then D is considered to be a concrete type. If D's type parameters are specified with underscores, then D is considered to be a type constructor of N type parameters, where N is the number of type parameters that were left unspecified through the use of placeholder syntax (i.e. underscores supplied as type parameters).

When calling a function defined in an interface, a client may supply either (1) a value of a type that implements the interface or (2) a value of the interface type to the function in any argument that is typed as the interface's self type.
For example, given:
```
interface Enumerable T for E T {
  fn each(E T, T -> Unit) -> Unit
}
```
a client may call the `each` function by supplying a value of type `Enumerable T` as the first argument to `each`, like this:
```
// assuming Array T implements the Enumerable T interface
e: Enumerable Int = Array(1, 2, 3)
each(e, puts)
```
or `each` may be called by supplying a value of type `Array T` as the first argument, like this:
```
// assuming Array T implements the Enumerable T interface
e = Array(1, 2, 3)
each(e, puts)
```
