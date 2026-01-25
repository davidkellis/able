# Idiomatic Able Code Guide

This guide captures conventions for writing idiomatic Able v11. The spec
(`spec/full_spec_v11.md`) remains authoritative; this document focuses on
readability, predictable control flow, and consistency across projects.

## Core Principles

- Favor expression-oriented code
- Prefer library helpers (`each`, `map`, `filter`, `reduce`) over raw loops when
  possible.
- Make error and optional flows explicit with `!`, `?`, `else`, and `match`.
- Use patterns and destructuring to keep data shapes front-and-center.
- Keep naming consistent: types in `CamelCase`, functions in `snake_case`.

## Bindings and Expression Flow

Use `:=` for new bindings and `=` for reassignment. Keep functions expression
oriented; reach for `return` only for early exit.

```able
fn describe(score: i32) -> String {
  label := if score >= 90 { "A" }
           elsif score >= 80 { "B" }
           else { "C or lower" }

  `grade: ${label}`
}
```

When a value needs a few steps, use `do` instead of temporary mutation.

```able
summary := do {
  subtotal := invoice.subtotal()
  tax := subtotal * 0.085
  `total: ${subtotal + tax}`
}
```

## Collections and Iteration

### Prefer higher-order helpers

When you are transforming or visiting every element, favor `each`/`map`/`filter`
instead of explicit `for` loops.

```able
# Preferred: side effects
users.each { user =>
  log(user.email)
}

# Preferred: transform
emails := users.map { user => user.email }

# Preferred: filter
active := users.filter { user => user.is_active() }

# Preferred: reduce
total := invoices.reduce(0) { acc, invoice => acc + invoice.total }
```

```able
# Avoid when possible
for user in users {
  log(user.email)
}
```

### Reach for `find`/`any`/`all` before manual loops

```able
has_errors := results.any { res => res.is_error() }
all_ready := workers.all { w => w.is_ready() }
first_failure := results.find { res => res.is_error() }
```

### Use `for` when control flow matters

`for` is still idiomatic when you need `break`, `continue`, or multi-step state.

```able
idx := 0
first_even := loop {
  if idx >= numbers.len() { break nil }
  value := numbers.get(idx) else { break nil }
  if value % 2 == 0 { break value }
  idx = idx + 1
}
```

```able
for entry in entries {
  if entry.is_skippable() { continue }
  if entry.is_final() { break }
  process(entry)
}
```

## Pipelines and Lambdas

Use `|>` and placeholders to make transformation pipelines readable.

```able
result := value
  |> normalize
  |> clamp(@, 0, 1)
  |> (@ * 100)
```

Favor short lambdas and placeholders when the meaning stays clear.

```able
scaled := numbers.map(@ * 2)
valid := numbers.filter(@ >= 0)
```

Use trailing lambdas for multi-step transforms.

```able
summary := orders.reduce(0) { acc, order =>
  acc + order.total
}
```

## Options and Results (`?` / `!`)

Prefer `?T` and `!T` for recoverable flows and use `or` to handle failure.

```able
fn load_config(path: String) -> !Config { ... }

fn boot(path: String) -> !void {
  config := load_config(path)!
  port := config.port or { 8080 }
  start_server(config.host, port)!
}
```

Capture errors when you need context.

```able
payload := read_file(path) or { |err|
  log(`fallback: ${err.message()}`)
  default_payload()
}
```

Avoid relying on truthiness for error-aware values; `Error` is falsy.

```able
# Prefer explicit handling
value := maybe_value or { default_value() }
```

## Pattern Matching and Destructuring

Use `match` to make union handling explicit and exhaustive.

```able
shape_area := shape match {
  case Circle { radius } => 3.141592653589793 * radius * radius,
  case Rectangle { width, height } => width * height,
  case Triangle { a, b, c } => {
    s := (a + b + c) / 2.0
    (s * (s - a) * (s - b) * (s - c)).sqrt()
  }
}
```

Destructure early to keep names local and meaningful.

```able
User { name, email } := load_user(id)!
print(`user ${name} <${email}>`)
```

## Struct Construction and Updates

Prefer named fields for clarity, and use functional updates when data should
remain stable by convention.

```able
next := Point { ...current, y: current.y + 1 }
```

## Methods, Interfaces, and `self`

Use `methods` for data-local behavior and `#` shorthand for the receiver.

```able
struct Counter { value: i32 }

methods Counter {
  fn #inc(by: i32 = 1) { #value = #value + by }
}
```

Use interfaces to express shared capabilities instead of ad-hoc conditionals.

```able
interface Renderable for T {
  fn render(self: Self) -> String
}
```

## Naming and Formatting

- Types, structs, and unions: `CamelCase`.
- Functions, fields, locals: `snake_case`.
- Predicates: `is_`, `has_`, `can_`.
- Conversions: `to_string`, `to_array`, `to_json`.
- Use trailing commas in multi-line literals for clean diffs.

```able
user := User {
  id: id,
  name: name,
  email: email,
}
```

## When Exceptions Are Appropriate

Use exceptions (`raise`/`rescue`) for truly exceptional conditions. Use `!T`
and `?T` for expected, recoverable outcomes.

```able
result := do { risky() } rescue {
  case _: ParseError => default_value(),
  case err: Error => {
    log(`unexpected error: ${err.message()}`)
    rethrow
  }
}
```
