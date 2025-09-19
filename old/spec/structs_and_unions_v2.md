Okay, here is the consolidated specification section covering Structs and Union Types based on our latest definitions.

---

# Able Language Specification: Structs and Union Types

This section defines Able's composite data types: **structs** (for grouping related data) and **union types** (for representing values that can be one of several types).

## 1. Structs

Structs aggregate named or positional data fields into a single type. Able supports three kinds of struct definitions: singleton, named-field, and positional-field. A single struct definition must be exclusively one kind. All fields are mutable.

### 1.1. Singleton Structs

Represent types with exactly one value, identical to the type name itself. Useful for simple enumeration variants or tags.

#### Declaration
```able
struct Identifier
```
*(Optionally `struct Identifier {}`)*

-   **`Identifier`**: Names the type and its unique value (e.g., `Red`, `EOF`, `Success`).

#### Instantiation & Usage
Use the identifier directly as the value.
```able
status = Success
color_val: Color = Red ## Assuming 'union Color = Red | ...'
```
Matched using the identifier in patterns: `case Red => ...`.

### 1.2. Structs with Named Fields

Group data under named fields.

#### Declaration
```able
struct Identifier [GenericParamList] {
  FieldName1: Type1,
  FieldName2: Type2
  FieldName3: Type3 ## Comma or newline separated, trailing comma ok
}
```
-   **`Identifier`**: Struct type name.
-   **`[GenericParamList]`**: Optional space-delimited generics (e.g., `T`, `K V: Display`).
-   **`FieldName: Type`**: Defines a field with a unique name within the struct.

#### Instantiation
Use `{ FieldName: Value, ... }`. Order doesn't matter. All fields must be initialized. Field init shorthand `{ name }` is supported.
```able
Identifier [GenericArgs] { Field1: Value1, Field2: Value2, ... }
## GenericArgs space-delimited if explicit, often inferred.
p = Point { x: 1.0, y: 2.0 }
u = User { id: 101, username, is_active: true } ## Shorthand
```

#### Field Access
Dot notation: `instance.FieldName`.
```able
x_coord = p.x
```

#### Functional Update
Create a new instance based on others using `...Source`. Later sources/fields override earlier ones.
```able
StructType { ...Source1, ...Source2, FieldOverride: NewValue, ... }
addr = Address { ...base_addr, zip: "90210" }
```

#### Field Mutation
Modify fields in-place using assignment.
```able
instance.FieldName = NewValue
p.x = p.x + 10.0
```

### 1.3. Structs with Positional Fields (Named Tuples)

Define fields by their position and type. Accessed by index.

#### Declaration
```able
struct Identifier [GenericParamList] {
  Type1,
  Type2
  Type3 ## Comma or newline separated, trailing comma ok
}
```
-   **`Identifier`**: Struct type name (e.g., `IntPair`, `Coord3D`).
-   **`[GenericParamList]`**: Optional space-delimited generics.
-   **`Type`**: Defines a field by its type at a specific zero-based position.

#### Instantiation
Use `{ Value1, Value2, ... }`. Values must be provided in the defined order. All fields must be initialized.
```able
Identifier [GenericArgs] { Value1, Value2, ... }
pair = IntPair { 10, 20 }
```

#### Field Access
Dot notation with zero-based integer index: `instance.Index`.
```able
first = pair.0 ## Accesses 10
second = pair.1 ## Accesses 20
```
Compile-time error preferred for invalid literal indices. Runtime error otherwise.

#### Functional Update
Not supported via `...Source` syntax for positional structs. Create new instances explicitly.

#### Field Mutation
Modify fields in-place using indexed assignment.
```able
instance.Index = NewValue
pair.0 = pair.0 + 5
```

### 1.4. Inherent Methods (`methods Type`)

Define methods (instance or static) directly associated with a specific struct type using a `methods` block. This is distinct from implementing interfaces.

#### Syntax
```able
methods [GenericParams] TypeName [GenericArgs] {
  [FunctionDefinitionList]
}
```
-   **`methods`**: Keyword initiating the block for defining inherent methods.
-   **`[GenericParams]`**: Optional generics `<...>` for the block itself (rare).
-   **`TypeName`**: The struct type name (e.g., `Point`, `User`).
-   **`[GenericArgs]`**: Generic arguments if `TypeName` is generic (e.g., `methods Pair A B { ... }`).
-   **`{ [FunctionDefinitionList] }`**: Contains standard `fn` definitions.

#### Method Definitions within `methods` block:

1.  **Instance Methods:** Operate on an instance. Defined using:
    *   Explicit `self`: `fn method_name(self: Self, ...) { ... }`
    *   Shorthand `fn #`: `fn #method_name(...) { ... }` (implicitly adds `self: Self`)
2.  **Static Methods:** Associated with the type itself, not an instance. Defined *without* `self` and *without* the `#` prefix.
    *   `fn static_name(...) { ... }`

#### Example: `methods` block for `Address`
```able
struct Address { house_number: u16, street: string, city: string, state: string, zip: u16 }

methods Address {
  ## Instance method using shorthand definition and access
  fn #to_s() -> string {
    `${#house_number} ${#street}\n${#city}, ${#state} ${#zip}`
  }

  ## Instance method using explicit self
  fn update_zip(self: Self, zip_code: u16) -> void {
    self.zip = zip_code ## Could also use #zip here
  }

  ## Static method (constructor pattern)
  fn from_parts(hn: u16, st: string, ct: string, sta: string, zp: u16) -> Self {
    Address { house_number: hn, street: st, city: ct, state: sta, zip: zp }
  }
}

## Usage
addr = Address.from_parts(...) ## Call static method
addr_string = addr.to_s()     ## Call instance method
addr.update_zip(90211)        ## Call instance method
```

## 2. Union Types (Sum Types / ADTs)

Represent values that can be one of several different types (variants). Essential for modeling alternatives (e.g., success/error, presence/absence, different kinds of related data).

### 2.1. Union Declaration

Define a new type as a composition of existing variant types using `|`.

#### Syntax
```able
union UnionTypeName [GenericParamList] = VariantType1 | VariantType2 | ... | VariantTypeN
```
-   **`union`**: Keyword.
-   **`UnionTypeName`**: The name of the new union type being defined.
-   **`[GenericParamList]`**: Optional space-delimited generic parameters applicable to the union itself.
-   **`=`**: Separator.
-   **`VariantType1 | VariantType2 | ...`**: List of one or more variant types separated by `|`.
    -   Each `VariantType` must be a pre-defined, valid type name (e.g., primitive, struct, another union, generic type application).

#### Examples
```able
## Simple enumeration using singleton structs
struct Red; struct Green; struct Blue;
union Color = Red | Green | Blue

## Option type (conceptual - assumes Some/None structs exist)
## struct Some T { value: T }
## struct None {} -- Note: 'None' conflicts with 'nil' literal/type. Rename needed? e.g., 'Empty'
## union Option T = Some T | Empty
union Option T = Some T | nil ## More direct using 'nil' type

## Result type (conceptual - assumes Ok/Err structs exist)
## struct Ok T { value: T }
## struct Err E { error: E }
## union Result T E = Ok T | Err E

## Mixing types
union IntOrString = i32 | string
```

### 2.2. Nullable Type Shorthand (`?`)

Provides concise syntax for types that can be either a specific type or `nil`.

#### Syntax
```able
?Type
```
-   **`?`**: Prefix operator indicating nullability.
-   **`Type`**: Any valid type expression.

#### Equivalence
`?Type` is syntactic sugar for the union `nil | Type`.
*(Note: Defined as `nil | Type` rather than `Type | nil`)*

#### Examples
```able
name: ?string = "Alice"
age: ?i32 = nil
maybe_user: ?User = find_user(id)
```

### 2.3. Constructing Union Values

Create a value of the union type by creating a value of one of its variant types.

```able
c: Color = Green
opt_val: Option i32 = Some i32 { value: 42 }
opt_nothing: Option i32 = nil

res_ok: Result string string = Ok string { value: "Data loaded" }
res_err: Result string string = Err string { error: "File not found" }

val: IntOrString = 100
val2: IntOrString = "hello"
```

### 2.4. Using Union Values

The primary way to interact with union values is via `match` expressions, which allow safely deconstructing the value based on its current variant.

```able
temp: Temp = F { deg: 32.0 }
desc = temp match {
  case F { deg } => `Fahrenheit: ${deg}`,
  case C { deg } => `Celsius: ${deg}`,
  case K { deg } => `Kelvin: ${deg}`
}

maybe_name: ?string = get_name_option()
display_name = maybe_name match {
  case s: string => s, ## Matches non-nil string
  case nil      => "Guest"
}
```

---
