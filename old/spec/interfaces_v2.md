
# Able Language Specification: Interfaces and Implementations (Second Revision)

This section defines **interfaces**, which specify shared functionality (contracts), and **implementations** (`impl`), which provide the concrete behavior for specific types or type constructors conforming to an interface. This system enables polymorphism, code reuse, and abstraction.

## 1. Interfaces

An interface defines a contract consisting of a set of function signatures. A type or type constructor fulfills this contract by providing implementations for these functions.

### 1.1. Interface Usage Models

Interfaces serve two primary purposes:

1.  **As Constraints:** They restrict the types allowable for generic parameters, ensuring those types provide the necessary functionality defined by the interface. This is compile-time polymorphism.
    ```able
    fn print_item<T: Display>(item: T) { ... } ## T must implement Display
    ```
2.  **As Types (Lens/Existential Type):** An interface name can be used as a type itself, representing *any* value whose concrete type implements the interface. This allows treating different concrete types uniformly through the common interface. This typically involves dynamic dispatch.
    ```able
    ## Assuming Circle and Square implement Display
    shapes: Array (Display) = [Circle{...}, Square{...}] ## Array holds values seen through the Display "lens"
    for shape in shapes {
        print(shape.to_string()) ## Calls the appropriate to_string via dynamic dispatch
    }
    ```
    *(Note: The exact syntax and mechanism for interface types like `Array (Display)` or `dyn Display` need further specification, but the concept is adopted.)*

### 1.2. Interface Definition

Interfaces are defined using the `interface` keyword, specifying required function signatures and the "self type" structure via a mandatory `for` clause.

#### Syntax
```able
interface InterfaceName [GenericParamList] for SelfTypePattern [where <ConstraintList>] {
  [FunctionSignatureList]
}
```

-   **`interface`**: Keyword.
-   **`InterfaceName`**: The identifier naming the interface (e.g., `Display`, `Mappable`).
-   **`[GenericParamList]`**: Optional space-delimited generic parameters for the interface itself (e.g., `T`, `K V`). Constraints use `Param: Bound`.
-   **`for`**: Keyword introducing the self type pattern (mandatory).
-   **`SelfTypePattern`**: Specifies the structure of the type(s) that can implement this interface. This defines the meaning of `Self`.
    *   **Concrete Type:** `TypeName [TypeArguments]` (e.g., `for Point`, `for Array i32`). `Self` refers to this specific concrete type.
    *   **Generic Type Variable:** `TypeVariable` (e.g., `for T`). `Self` refers to the type variable `T`. Used when the interface itself is generic *over* the implementing type.
    *   **Type Constructor (HKT):** `TypeConstructor _ ...` (e.g., `for M _`, `for Map K _`). `Self` refers to the type constructor (`M` or `Map` in examples). `_` denotes unbound parameters.
    *   **Generic Type Constructor:** `TypeConstructor TypeVariable ...` (e.g., `for Array T`). `Self` refers to the specific application `Array T`.
-   **`[where <ConstraintList>]`**: Optional constraints on generic parameters used in `GenericParamList` or `SelfTypePattern`.
-   **`{ [FunctionSignatureList] }`**: Block containing function signatures (methods).
    *   Each signature follows `fn name[<MethodGenerics>]([Param1: Type1, ...]) -> ReturnType;`.
    *   Methods can be instance methods (typically taking `self: Self` as the first parameter) or static methods (no `self: Self` parameter).

#### `Self` Keyword Interpretation
Within the interface definition (and corresponding `impl` blocks):
*   If `SelfTypePattern` is `MyType Arg1`, `Self` means `MyType Arg1`.
*   If `SelfTypePattern` is `T`, `Self` means `T`.
*   If `SelfTypePattern` is `MyConstructor _`, `Self` means the constructor `MyConstructor`. `Self Arg` means `MyConstructor Arg`.
*   If `SelfTypePattern` is `MyConstructor T`, `Self` means `MyConstructor T`.

#### Examples
```able
## Interface for specific concrete type (less common)
# interface PointExt for Point { fn extra_op(self: Self) -> i32; }

## Interface generic over the implementing type
interface Display for T {
  fn to_string(self: Self) -> String;
}

interface Clone for T {
  fn clone(self: Self) -> Self;
}

## Interface for a specific generic type application
interface IntArrayOps for Array i32 {
  fn sum(self: Self) -> i32;
}

## Interface for a type constructor (HKT)
interface Mappable A for M _ {
  fn map<B>(self: Self A, f: (A -> B)) -> Self B; ## Self=M, Self A = M A
}

## Interface with static method
interface Zeroable for T {
  fn zero() -> Self; ## Static method, returns an instance of the implementing type T
}
```

### 1.3. Composite Interfaces (Interface Aliases)

Define an interface as a combination of other interfaces.

#### Syntax
```able
interface NewInterfaceName [GenericParams] = Interface1 [Args] + Interface2 [Args] + ...
```
-   Implementing `NewInterfaceName` requires implementing all constituent interfaces (`Interface1`, `Interface2`, etc.).

#### Example
```able
interface ReadWrite = Reader + Writer ## Assuming Reader, Writer interfaces exist
interface DisplayClone T = Display for T + Clone for T ## Needs for clause? TBD
## Refined Example: Assuming Display/Clone are defined 'for T'
interface DisplayClone for T = Display + Clone
```
*(Note: Interaction with `for` clause in composite definitions needs clarification. Does the composite inherit/require consistent self types? Assume for now it applies to the same `T`)*.

## 2. Implementations

Implementations provide the concrete function bodies for an interface for a specific type or type constructor.

### 2.1. Implementation Declaration

Provides bodies for interface methods. Can use `fn #method` shorthand if desired.

#### Syntax
```able
[ImplName =] impl [<ImplGenericParams>] InterfaceName [InterfaceArgs] for Target [where <ConstraintList>] {
  [ConcreteFunctionDefinitions]
}
```

-   **`[ImplName =]`**: Optional name for the implementation, used for disambiguation. Followed by `=`.
-   **`impl`**: Keyword.
-   **`[<ImplGenericParams>]`**: Optional comma-separated generics for the implementation itself (e.g., `<T: Numeric>`). Use `<>` delimiters.
-   **`InterfaceName`**: The name of the interface being implemented.
-   **`[InterfaceArgs]`**: Space-delimited type arguments for the interface's generic parameters (if any).
-   **`for`**: Keyword (mandatory).
-   **`Target`**: The specific type or type constructor implementing the interface. This must structurally match the `SelfTypePattern` defined in the interface's `for` clause.
    *   If interface is `for Point`, `Target` must be `Point`.
    *   If interface is `for T`, `Target` can be any type `SomeType` (or a type variable `T` in generic impls).
    *   If interface is `for M _`, `Target` must be `TypeConstructor` (e.g., `Array`). The implementation needs access to the constructor's type parameter (see HKT Impl syntax).
    *   If interface is `for Array T`, `Target` must be `Array U` (where `U` might be constrained).
-   **`[where <ConstraintList>]`**: Optional constraints on `ImplGenericParams`.
-   **`{ [ConcreteFunctionDefinitions] }`**: Block containing the full definitions (`fn name(...) { body }`) for all functions required by the interface. Signatures must match.

-   **Distinction from methods declaration:** `methods Type { ... }` defines *inherent* methods (part of the type's own API). `impl Interface for Type { ... }` fulfills an *external contract* defined by the interface. An inherent method defined in `methods Type` may be used to explicitly satisfy an interface requirement in an `impl` block, but the `impl Interface for Type` block is still needed to declare the conformance.

#### HKT Implementation Syntax (Refined)
To implement an interface `for M _` for a concrete constructor like `Array`:
```able
[ImplName =] impl [<ImplGenerics>] InterfaceName TypeParamName for TypeConstructor [where ...] {
  ## 'TypeParamName' (e.g., A) is bound here and usable below.
  fn method<B>(self: TypeConstructor TypeParamName, ...) -> TypeConstructor B { ... }
}
## Example:
impl Mappable A for Array {
  fn map<B>(self: Array A, f: (A -> B)) -> Array B { ... }
}

union Option T = T | nil
impl Mappable A for Option {
  fn map<B>(self: Self A, f: A -> B): Self B { self match { case a: A  => f(a), case nil => nil } }
}

impl Mappable A for _ | nil {
  fn map<B>(self: Self A, f: A -> B): Self B { self match { case a: A  => f(a), case nil => nil } }
}
```

#### Examples
```able
## Implementing Display for Point
impl Display for Point {
  fn to_string(self: Self) -> String { `Point({self.x}, {self.y})` }
}

## Implementing Zeroable (static method)
impl Zeroable for i32 {
  fn zero() -> Self { 0 } ## Self is i32 here
}
impl Zeroable for Array { ## Implementing for the constructor
  fn zero<T>() -> Array T { [] } ## Needs generic param on method
}

## Named Monoid Implementations for i32
Sum = impl Monoid for i32 { ## Assuming interface Monoid for T exists
  fn id() -> Self { 0 }
  fn op(self: Self, other: Self) -> Self { self + other }
}
Product = impl Monoid for i32 {
  fn id() -> Self { 1 }
  fn op(self: Self, other: Self) -> Self { self * other }
}
```

### 2.2. Overlapping Implementations and Specificity

When multiple `impl` blocks could apply to a given type and interface, Able uses specificity rules to choose the *most specific* implementation. If no single implementation is more specific, it results in a compile-time ambiguity error. Rules (derived from Rust RFC 1210, simplified):

1.  **Concrete vs. Generic:** Implementations for concrete types (`impl ... for i32`) are more specific than implementations for type variables (`impl ... for T`). (`Array i32` is more specific than `Array T`).
2.  **Concrete vs. Interface Bound:** Implementations for concrete types (`impl ... for Array T`) are more specific than implementations for types bound by an interface (`impl ... for T: Iterable`).
3.  **Interface Bound vs. Unconstrained:** Implementations for constrained type variables (`impl ... for T: Iterable`) are more specific than for unconstrained variables (`impl ... for T`).
4.  **Subset Unions:** Implementations for union types that are proper subsets are more specific (`impl ... for i32 | f32` is more specific than `impl ... for i32 | f32 | f64`).
5.  **Constraint Set Specificity:** An `impl` whose type parameters have a constraint set that is a proper superset of another `impl`'s constraints is more specific (`impl<T: A + B> ...` is more specific than `impl<T: A> ...`).

Ambiguities must be resolved manually, typically by qualifying the method call.

## 3. Usage

### 3.1. Instance Method Calls

Use dot notation on a value whose type implements the interface.
```able
p = Point { x: 1, y: 2 }
s = p.to_string() ## Calls Point's impl of Display.to_string

arr = [1, 2, 3]
arr_mapped = arr.map({ x => x * 2 }) ## Calls Array's impl of Mappable.map
```

### 3.2. Static Method Calls

Use `TypeName.static_method(...)` notation. The `TypeName` must have an `impl` for the interface containing the static method.
```able
zero_int = i32.zero()          ## Calls i32's impl of Zeroable.zero
empty_f64_array = Array.zero<f64>() ## Calls Array's impl of Zeroable.zero, needs type arg
```

### 3.3. Disambiguation (Named Impls)

If multiple implementations exist (e.g., `Sum` and `Product` for `Monoid for i32`), qualify the call with the implementation name:
```able
sum_id = Sum.id()             ## 0
prod_id = Product.id()         ## 1
res = Sum.op(5, 6)          ## 11 (Calls Sum's op) - Method style TBD: 5.Sum.op(6)? Needs clarification.
res2 = Product.op(5, 6)       ## 30
```
*(Note: Interaction between named impls and method call syntax (`value.ImplName.method(...)`) needs confirmation).*

### 3.4. Interface Types (Dynamic Dispatch)

Using an interface as a type allows storing heterogeneous values implementing the same interface. Method calls typically use dynamic dispatch.
```able
displayables: Array (Display) = [1, "hello", Point{...}]
for item in displayables {
  print(item.to_string()) ## Dynamic dispatch selects correct to_string impl
}
```

