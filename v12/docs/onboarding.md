# Able v11 Onboarding Guide

*A concise introduction to the Able programming language for experienced developers.*

## Philosophy

Able is a hybrid functional-imperative language with:
- Static typing with extensive type inference
- First-class functions and algebraic data types
- Lightweight Go-inspired concurrency
- Minimal, expressive syntax

## Basic Syntax

### Variables and Mutability

```able
# Declaration with :=
name := "Alice"           # New binding
count := 42               # Type inferred as i32

# Reassignment with =
count = 100               # Rebind existing variable

# Destructuring assignment
point := {x: 10, y: 20}
{x, y} := point           # Extract fields
```

### Primitive Types

```able
# Numbers
int_val := 42             # i32 by default
float_val := 3.14         # f64 by default
large_num := 1_000_000_i64 # Explicit type with suffix

# Strings and Characters
text := "Hello, Able!"
char_val := 'π'
interpolated := `Value: ${int_val}`

# Other primitives
is_ready := true          # bool
nothing := nil            # Absence of value
unit := void              # Unit type
```

## Data Structures

### Structs

```able
# Named fields
struct Person {
  name: String
  age: i32
}

person := Person{name: "Bob", age: 30}
print(person.name)        # Field access

# Functional update
older_person := person{age: 31}

# Positional fields (named tuples)
struct Point i32 i32
p := Point(10, 20)
x, y := Point(p)          # Destructure

# Singleton struct
struct Success
result := Success
```

### Union Types (ADTs)

```able
# Define union
union Result {
  Ok(String)
  Error(String)
}

# Construct variants
success := Result.Ok("data")
failure := Result.Error("failed")

# Pattern matching
message := match result {
  Result.Ok(value) => `Success: ${value}`
  Result.Error(msg) => `Error: ${msg}`
}

# Nullable shorthand
?String                   # Equivalent to: nil | String
```

### Arrays

```able
# Creation
numbers := Array[1, 2, 3, 4]
empty := Array i32        # Empty array with type annotation

# Operations
numbers.push(5)           # Mutate
first := numbers[0]       # Index access
length := numbers.len()   # Method call
```

## Functions

### Definition

```able
# Named function
fn add(a: i32, b: i32) -> i32 {
  a + b                   # Implicit return
}

# Generic function
fn identity<T>(value: T) -> T {
  value
}

# With explicit return
fn process(x: i32) -> String {
  if x < 0 {
    return "negative"     # Early return
  }
  "positive"
}
```

### Anonymous Functions

```able
# Lambda syntax
double := |x| x * 2

# Verbose syntax
multiply := fn(a: i32, b: i32) -> i32 { a * b }

# Closures capture outer scope
factor := 3
scale := |x| x * factor
```

### Function Calls

```able
# Standard call
result := add(10, 20)

# Method call syntax (if method defined)
length := "hello".len()

# Trailing lambda
numbers.each(|item| print(item))
```

## Control Flow

### Conditional Expressions

```able
# if/elsif/else chain
grade := if score >= 90 { "A" }
         elsif score >= 80 { "B" }
         else { "C" }

# All branches are expressions
value := if condition { 10 } else { 20 }
```

### Pattern Matching

```able
# Match on values
description := match value {
  0 => "zero"
  1 | 2 => "small"
  n if n < 10 => "single digit"
  _ => "large"
}

# Match on types
process := match input {
  Person{name, age} => `Person: ${name}, age ${age}`
  Point(x, y) => `Point: (${x}, ${y})`
  nil => "null value"
}
```

### Loops

```able
# While loop
while condition {
  # loop body
}

# For loop over ranges
for i in 0..10 {
  print(i)               # 0 through 9
}

# For loop over arrays
for item in numbers {
  print(item)
}

# Continue and break
for i in 0..100 {
  if i % 2 == 0 { continue }
  if i > 50 { break }
  print(i)
}
```

### Non-local Jumps

```able
# Define exit point
breakpoint exit_loop

# Loop with break to exit point
for i in 0..1000 {
  for j in 0..1000 {
    if should_exit(i, j) {
      break exit_loop     # Jump to breakpoint
    }
  }
}

# Code after break executes
print("Exited nested loops")
```

## Interfaces and Implementations

### Interface Definition

```able
# Interface for displayable types
interface Display for T {
  fn to_string(self: Self) -> String
}

# Generic interface
interface Mappable M {
  fn map<M, R>(self: Self, fn(M) -> R) -> M R
}
```

### Implementation

```able
# Implement for specific type
impl Display for Person {
  fn to_string(self: Self) -> String {
    `Person: ${self.name} (${self.age})`
  }
}

# Generic implementation
impl Mappable Array for T {
  fn map<T, R>(self: Self, mapper: fn(T) -> R) -> Array R {
    result := Array R
    for item in self {
      result.push(mapper(item))
    }
    result
  }
}
```

### Usage

```able
# As constraint
fn print_item<T: Display>(item: T) {
  print(item.to_string())
}

# As existential type (dynamic dispatch)
items: Array Display = [person1, point1]
for item in items {
  print(item.to_string())
}
```

## Error Handling

### Option and Result Types

```able
# Nullable type shorthand
?String                   # nil | String

# Result type shorthand  
!String                   # Error | String

# Error propagation with !
data := risky_operation()!
# Equivalent to:
# data := match risky_operation() {
#   Ok(value) => value
#   Error(err) => return Error(err)
# }

# Error handling with else
result := (risky_operation() else { |err|
  print(`Operation failed: ${err}`)
  return nil
})
```

### Exceptions

```able
# Raise exception
fn validate(age: i32) -> void {
  if age < 0 {
    raise "Age cannot be negative"
  }
}

# Rescue exceptions
process := fn() -> String {
  validate(user_age)
  "success"
} rescue {
  case err => `Validation failed: ${err}`
}
```

## Concurrency

### Async Tasks

```able
# Start async task
handle := spawn fetch_data("http://example.com")

# Wait for result (handle errors with else)
data := handle.value() else { err => `failed: ${err.message()}` }

# Async block
result := spawn do {
  part1 := compute_part1()
  part2 := compute_part2()
  combine(part1, part2)
}

# Inspect status or cancel
status := handle.status()
if match status { case Pending => true } { print("Still running") }
handle.cancel()
```

## Modules and Imports

### Package Structure

```able
# package.yml
name: my_app
version: 1.0.0
dependencies:
  collections:
    github: davidkellis/able-collections
    version: "~>0.16.0"
```

### Imports

```able
# Import package
import collections

# Import specific items
import collections.{List, Map}

# Dynamic import
dynimport "runtime_config"
```

## Advanced Features

### Type Constraints

```able
# Function with constraints
fn process_all<T: Display + Clone>(items: Array T) -> Array String {
  items.map(|item| item.to_string().clone())
}
```

### Higher-Kinded Types

```able
# Interface for type constructors
interface Functor F {
  fn map<F, A, B>(self: F A, fn(A) -> B) -> F B
}
```

### Method Definitions

```able
struct Point { x: i32, y: i32 }

methods Point {
  fn distance(self: Self, other: Point) -> f64 {
    dx := self.x - other.x
    dy := self.y - other.y
    (dx*dx + dy*dy).sqrt()
  }
}
```

## Quick Reference

| Feature | Syntax | Example |
|---------|--------|---------|
| Declaration | `name := value` | `x := 42` |
| Reassignment | `name = value` | `x = 100` |
| Function | `fn name(params) -> type { body }` | `fn add(a: i32, b: i32) -> i32 { a + b }` |
| Lambda | `|params| expression` | `|x| x * 2` |
| Conditional | `if cond { expr } or { expr }` | `if x > 0 { "pos" } or { "neg" }` |
| Pattern Match | `match value { pattern => expr }` | `match x { 0 => "zero" }` |
| Struct | `struct Name { field: type }` | `struct Person { name: String }` |
| Union | `union Name { Variant(type) }` | `union Result { Ok(String) }` |
| Interface | `interface Name for T { fn sig }` | `interface Display for T { fn to_string(self: Self) -> String }` |
| Implementation | `impl Interface for Type { body }` | `impl Display for Person { ... }` |
| Async | `spawn expression` | `handle := spawn fetch_data()` |
| Nullable | `?Type` | `?String` |
| Result | `!Type` | `!String` |

## Getting Started

1. Create a `package.yml` file
2. Write your main function in `main.abl`
3. Run with: `able run main.abl`

```able
# main.abl
package main

fn main() -> void {
  message := "Hello, Able!"
  print(message)
}
```

Welcome to Able! 🚀
