# Language Specification 2025

## Features

The core elements of the Able language are intended to be compliment one another while being clearly distinct (i.e. orthogonal but complimentary).

- [Packages and imports](packages.md)
- Primitive types
  - nil
  - boolean
  - integers
  - floats
  - strings
  - arrays
  - maps
- Collections
  - Interfaces
    - Iterable ; see https://docs.scala-lang.org/overviews/collections-2.13/trait-iterable.html
      - Seq ; see https://docs.scala-lang.org/overviews/collections-2.13/seqs.html
        - IndexedSeq
        - LinearSeq
      - Sets ; see https://docs.scala-lang.org/overviews/collections-2.13/sets.html
      - Maps
  - Generators
- Semantics by Interface
  - Function Application
  - Equality
  - Inequality
  - Struct Instantiation
  - Indexed Assignment
- Blocks
- Structs
- Unions
- [Functions](functions.md)
- Interfaces
  - Implementations
- Control Flow
  - if
  - while
  - for
- Destructuring

## Identifiers

All identifiers must conform to the pattern `[a-zA-Z0-9][a-zA-Z0-9_]*`

### Naming Conventions

- Prefer snake_case for file names, package names, variable names, and function names.
- Prefer PascalCase for type names.
- Primitive built-in types are all lowercase and single word names (e.g. string, u8, i128)

## Comments

```
## comments follow the double pound sign, `##`, to the end of the line
```

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

