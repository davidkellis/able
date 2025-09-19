# Able Language Specification: Structs (Revised and Expanded)

This section defines the syntax and semantics for declaring and using **structs** in Able. Structs are composite data types that group related data. Able supports three kinds of struct definitions:

1.  **Singleton Structs:** Have no fields; the type name also serves as its only value. Useful for simple enumeration variants or tags.
2.  **Structs with Named Fields:** The standard struct type, where each field has an explicit name and type.
3.  **Structs with Positional Fields:** Also known as "named tuples," where fields are defined only by their type and accessed by their zero-based index.

A single struct definition must be *exclusively* one of these kinds; mixing named and positional fields within one definition is not allowed. All fields (named or positional) are mutable.

## 1. Singleton Structs

These represent types with only a single possible value, which is identical to the type name itself.

### Declaration

Declare using the `struct` keyword followed only by the identifier. The body `{}` is empty or omitted.

```able
struct Identifier
## Or equivalently:
# struct Identifier {}
```

-   **`struct`**: Keyword.
-   **`Identifier`**: The name of the struct type and its unique value (e.g., `Red`, `Unit`, `NotFound`).

### Examples

```able
struct Red
struct Green
struct Blue
struct Processing ## Represents a state
```

### Instantiation and Usage

No explicit instantiation syntax is needed. The identifier itself represents the unique value of that type.

```able
## Given: union Color = Red | Green | Blue
c1: Color = Red    ## Assign the value 'Red'
c2: Color = Blue   ## Assign the value 'Blue'

status = Processing ## Assign the value 'Processing'

## Usage in pattern matching
match c1 {
  Red => print("It's Red!")
  Green => print("It's Green!")
  Blue => print("It's Blue!")
}
```

### Semantics

-   Introduces a type with exactly one value.
-   The type name and the value identifier are the same.
-   Often used as variants in `union` types for simple enumerations or states.

## 2. Structs with Named Fields

These group data under named fields.

### Declaration

Fields are defined using `FieldName: Type`.

```able
struct Identifier [GenericParamList] {
  FieldName1: Type1, ## Comma separation
  FieldName2: Type2

  FieldName3: Type3  ## Newline separation (comma optional)
  FieldName4: Type4, ## Trailing comma allowed
}
```

-   **`struct`**: Keyword.
-   **`Identifier`**: The name of the struct type.
-   **`[GenericParamList]`**: Optional space-delimited generic parameters (e.g., `T`, `K V`). Bounds use `Param: Bound1 + Bound2`.
-   **`{ FieldList }`**: Block containing field definitions.
    -   Each field is `Identifier: Type`.
    -   Fields separated by commas or newlines. Trailing comma allowed.
    -   **Constraint**: Field names within a single struct definition must be unique.

### Examples

```able
## Simple struct
struct Point { x: f64, y: f64 }

## Generic struct with bounds
struct Record K V: Display {
  key: K
  value: V
}

## Multi-line definition
struct User {
  id: u64
  username: String
  is_active: bool
}
```

### Instantiation

Use the type name followed by `{ FieldName: Value, ... }`.

```able
Identifier [GenericArgs] {
  FieldName1: Value1,
  FieldName2: Value2,
  ... ## Field init shorthand { name } also applies
}
## GenericArgs are space-delimited if needed and not inferred (e.g., MyType Arg1 Arg2 { ... })
```

-   Order of fields does not matter during instantiation.
-   All fields must be initialized.
-   Field init shorthand `{ name }` (equivalent to `name: name`) is supported.

### Examples

```able
p1 = Point { x: 1.0, y: 2.5 }
p2 = Point { y: 5.0, x: 0.0 } ## Order doesn't matter

user1 = User {
  id: 101,
  username: "alice",
  is_active: true
}

username = "bob" ## Using shorthand
user2 = User { id: 102, username, is_active: false }

record1 = Record String i32 { key: "age", value: 30 }
```

### Field Access

Use dot notation: `instance.FieldName`.

```able
x_coord = p1.x ## Accesses field 'x'
name = user1.username
```

### Functional Update

Use `...Source` syntax. Multiple sources are allowed, processed left-to-right, with later sources and explicit fields overriding earlier ones.

```able
StructType { ...Source1, ...Source2, FieldOverride: NewValue, ... }
```

```able
defaults = User { id: 0, username: "guest", is_active: false }
admin_base = User { ...defaults, username: "admin", is_active: true }
## admin_base is { id: 0, username: "admin", is_active: true }

mom_addr = Address { street: "1 Main", city: "Anytown", zip: "11111" }
dad_addr = Address { ...mom_addr, street: "2 Oak", zip: "22222" }
final_addr = Address { ...mom_addr, ...dad_addr, zip: "99999" }
## final_addr is { street: "2 Oak", city: "Anytown", zip: "99999" }
```

### Field Mutation

Use assignment with dot notation: `instance.FieldName = NewValue`.

```able
p1.x = 100.0 ## Modifies p1's x field in place
user1.is_active = false
```

## 3. Structs with Positional Fields (Named Tuples)

These define fields implicitly by their position and type.

### Declaration

Fields are defined only by their `Type`.

```able
struct Identifier [GenericParamList] {
  Type1, ## Comma separation
  Type2

  Type3  ## Newline separation (comma optional)
  Type4, ## Trailing comma allowed
}
```

-   **`struct`**: Keyword.
-   **`Identifier`**: The name of the struct type (e.g., `IntPair`, `Coord3D`).
-   **`[GenericParamList]`**: Optional space-delimited generics with bounds.
-   **`{ FieldList }`**: Block containing type definitions.
    -   Each field is just a `Type`.
    -   Types are separated by commas or newlines. Trailing comma allowed.

### Examples

```able
struct IntPair { i32, i32 }
struct Coord3D { f64, f64, f64 }

struct MixedTuple T U: Display {
  i32
  String
  T
  U
}
```

### Instantiation

Use the type name followed by `{ Value1, Value2, ... }`.

```able
Identifier [GenericArgs] { Value1, Value2, ... }
## GenericArgs are space-delimited if needed.
```

-   Values must be provided in the same order as the types were defined.
-   All fields must be initialized.

### Examples

```able
pair = IntPair { 10, 20 }
coord = Coord3D { 1.0, -2.5, 0.0 }

mt = MixedTuple Bool String { 1, "hello", true, "world" }
```

### Field Access

Use dot notation with a zero-based integer index: `instance.Index`.

```able
first_val = pair.0   ## Accesses the first field (10)
second_val = pair.1  ## Accesses the second field (20)
z_coord = coord.2    ## Accesses the third field (0.0)

mt_string = mt.1     ## Accesses the String field "hello"
mt_bool = mt.2       ## Accesses the Bool field true
```

**Note:** Accessing an out-of-bounds index should be a compile-time error if the index is a literal and the type is known, otherwise potentially a runtime error/panic.

### Functional Update

The `...Source` syntax is **not** supported for positional structs due to the ambiguity of merging based on position. Create a new instance explicitly if updates are needed.

### Field Mutation

Use assignment with index notation: `instance.Index = NewValue`.

```able
pair.0 = 99         ## Modifies the first field of pair in place
coord.1 = coord.1 + 10.0 ## Update the second field
```
