
# Interfaces

An interface is a set of functions that are guaranteed to be defined for a given type.

An interface can be used to represent either (1) a type, or (2) a constraint on a type or type variable.

When used as a type, the interface functions as a lens through which we see and interact with some underlying type that implements the interface. This form of polymorphism allows us to use different types as if they were the same type. For example, if the concrete types `i32` and `f32` both implement an interface called `Numeric`, then one could create an `Array Numeric` consisting of both `i32` and `f32` values, with the understanding that we can interact with either type of value with the common interface defined within the `Numeric` interface.

When used as a constraint on a type, the interpretation is that some context needs to reference the underlying type, while also making a claim or guarantee that the underlying type implements the interface.

## Interface Definitions

Interfaces are defined with the following syntax:

```
interface A [B C ...] [for D [E F ...]] [where <type constraints>]{
  fn foo(T1, T2, ...) -> T3
  ...
}
```

where A is the name of the interface optionally parameterized with type parameters, B, C, etc.

The `for` clause is optional. When present, the `for` clause specifies that the type D is deemed to implement interface A.

The type expression in the for clause is called the interface's <a name="self-type"></a>_self type_. The self type is a concrete or polymorphic type representing the type that will implement the interface. In the syntax above, `D [E F ...]` represents the self type.

D may represent either a concrete type or a polymorphic type depending on whether D's type parameters (if applicable) are bound or unbound.

Here are a few example interface definitions:

```
interface Stringable for T {
  fn to_s(T) -> string
}

interface Comparable for T {
  fn compare(T, T) -> i32
}

interface Iterable T for E {
  fn each(E, T -> void) -> void
  fn iterator(e: E) -> Iterator T { Iterator { gen => e.each(gen.yield) } }
}

## Mappable interface (something approximating Functor in Haskell-land - see https://wiki.haskell.org/Typeclassopedia#Functor)
interface Mappable A for M _ {
  fn map(m: M A, convert: A -> B) -> M B
}

```

The functions defined as members of an interfaces are called the interface's methods.

### Self Type

There is a special type named `Self` that may be used within an interface definition or an implementation of an interface to reference the [self type](#self-type) captured in the for clause of the interface or implementation.

For example:

```
interface Stringable for T {
  fn to_s(T) -> string
}
```

could also be written as:

```
interface Stringable for T {
  fn to_s(Self) -> string
}
```

In cases where the [self type](#self-type) is polymorphic, i.e. the self type has one or more unbound type parameters, the special `Self` type may be used as a type constructor and supplied with type arguments.

For example:

```
## Mappable interface (something approximating Functor in Haskell-land - see https://wiki.haskell.org/Typeclassopedia#Functor)
interface Mappable A for M _ {
  fn map(m: M A, convert: A -> B) -> M B
}
```

could be written as:

```
## Mappable interface (something approximating Functor in Haskell-land - see https://wiki.haskell.org/Typeclassopedia#Functor)
interface Mappable A for M _ {
  fn map<B>(m: Self A, convert: A -> B) -> Self B
}
```

### Interface Aliases and Composite Interface Types

Composite interfaces may be defined with the following syntax:

```
interface A [B C ...] = D [E F ...] [+ G [H I ...] + J [K L ...] ...]
```

where A is the name of the composite interface and D, G, and J are the component interfaces that compose together to constitute the composite interface.

Some contrived examples of composite interface types:

```
interface WebEncodable T = XmlEncodable T + JsonEncodable T
interface Repository T = Collection T + Searchable T
interface Collection T = Countable T + Addable T + Removable T + Getable T
```

## Interface Implementations

Interface implementations are defined with the following syntax:

```
[ImplName =] impl <X, Y, Z, ...> A <B C ...> [for D <E F ...>] [where <type constraints>] {
  fn foo(T1, T2, ...) -> T3 { ... }
  ...
}
```

Here are some example interface implementations:

```
struct Person { name: string, address: string }
impl Stringable for Person {
  fn to_s(p: Person) -> string => "Name: ${p.Name}\nAddress: ${p.Address}"
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
    ab = ArrayBuilder(0, arr.length)  ## length, capacity
    ## arr.each { element => convert(element) |> ab.add }
    arr.each { convert(_) |> ab.add }
    ab.to_a()
  }
}

```

Implementations may be named in order to cope with situations where the same type implements the same interface with differing semantics. When an implementation is named, function calls that might otherwise be ambiguous can be fully qualified in order to disambiguate the function call. For example:

```
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
```

Since `Monoid.id()` may be ambiguous, depending on the number of impls for a given type - it is in this case -  you have to qualify the function call with the name of the implementation, like `Sum.id()`, or `Sum.op(4, 5)`.

### Importing Interfaces and Implementations

Interfaces may be imported like any other type. An interface's methods may be imported as well, as if treated like importing a function from a package.

```
## assume Monoid is defined in the monoid package
interface Monoid for T {
  fn id() -> T
  fn op(T, T) -> T
}
```

Then Monoid may be imported and used like this:
```
import monoid.{ Monoid }
```

If one wishes to use the

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
## assuming List T implements the Iterable T interface
e: Iterable i32 = List(1, 2, 3)    ## e has type Iterable i32
each(e, puts)
```

or `each` may be called by supplying a value of type `Array T` as the first argument, like this:

```
## assuming Array T implements the Iterable T interface
e = Array(1, 2, 3)    ## e has type Array i32
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
## overlap when T = i32
impl Foo for T {}
impl Foo for i32 {}

## overlap when T = i32
impl Foo for Array T {}
impl Foo for Array i32 {}

## overlap because Array T implements Iterable T
impl Foo for Iterable T {}
impl Foo for Array T {}

## overlap when T implements Bar
impl Foo for T {}
impl Foo for T: Bar {}

## overlap when T implements Bar and Baz
impl Foo for T: Bar {}
impl Foo for T: Baz {}

## overlap when T = f32
impl Foo for i32 | f32 {}
impl Foo for i32 | T {}

## overlap when two type unions implement the same interface, s.t. one type union is a proper subset of the other
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
   - `T: Iterable + Stringable` is more specific than `T: Iterable` because `{Iterable, Stringable}` is a proper superset of `{Iterable}`

There are many potential cases where the type constraints in competing interface implementations can't be compared. In those cases, the compiler will emit an error saying that the competing interface implementations are not comparable, and so the expression is ambiguous.

Ambiguities that cannot be resolved by these specialization rules must be resolved manually, by qualifying the relevant function call with either the name of the interface or the name of the implementation. For example:

```
## assuming Array T implements both Iterable T and BetterIterable T, and both define an each function
impl Iterable T for Array T { ... }
impl BetterIterable T for Array T { ... }

## an expression like
Array(1,2,3).each { elem => puts(elem) }
## would be ambiguous

## however, the expressions
BetterIterable.each(Array(1,2,3)) { elem => puts(elem) }
Array(1,2,3) |> BetterIterable.each { elem => puts(elem) }
## are not ambiguous
```

## Static methods

If the first parameter of an interface method is not the `Self` type, then the method is deemed to be a static method.

For example, given the following interface and implementations:
```
interface Zeroable for T {
  fn zero() : T
}

## Array type constructor implements Zeroable
impl Zeroable for Array {
  fn zero<T>(): Array T {
    return [];
  }
}

impl Zeroable for i32 {
  fn zero(): i32 { 0 }
}

impl Zeroable for bool {
  fn zero(): bool { false }
}
```

the zero function would be callable like this:
```
i32.zero()
bool.zero()
```
