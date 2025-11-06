# Able v10 Manual

This manual mirrors the structure of the Julia language manual, providing an accelerated walkthrough of Able v10 for experienced programmers. The language specification in `spec/full_spec_v10.md` remains authoritative; the examples here focus on practical usage. Check the Go (`interpreter10-go/`) and TypeScript (`interpreter10/`) interpreters for runnable artefacts.

## 1. Getting Started

Able programs live inside package directories. Top-level `fn main() -> void` definitions produce executables, mirroring `examples/hello_world.able`.

```able
package helloworld

fn main() {
  print("Hello, Able!")
}
```

An interactive REPL is not yet part of the toolchain, but both interpreters accept module inputs and can run fixtures or tests from the command line.

### 1.1 Resources

- Specification: `spec/full_spec_v10.md`
- TypeScript interpreter: `interpreter10/README.md`
- Go interpreter overview: `interpreter10-go/` and `interpreter10-go/PARITY.md`
- Design notes: `design/` (e.g., `pattern-break-alignment.md`, `typechecker-plan.md`)
- Shared fixtures: `fixtures/ast` and `interpreter10/scripts/run-fixtures.ts`

### 1.2 Installation

Install both interpreters to keep behaviour aligned:

```bash
# TypeScript interpreter
bun install
bun test

# Go interpreter
go mod tidy      # initial setup
go test ./pkg/interpreter

# Unified test runner
./run_all_tests.sh --typecheck-fixtures=strict
```

Use Bun ≥ 1.2 and Go ≥ 1.22. Windows users can run via WSL2.

### 1.3 Running Able Code

Harness existing interpreters:

```bash
# Execute shared AST fixtures with the TypeScript interpreter
bun run scripts/run-fixtures.ts --filter hello_world

# Run Go interpreter parity tests (executes Able fixtures and typechecks)
go test ./pkg/interpreter
```

For ad-hoc exploration, add `.able` files under `examples/` or create new fixtures, then export with `scripts/export-fixtures.ts`.

## 2. Variables

Able separates declaration (`:=`) from assignment (`=`). Bindings are mutable by default and values may be mutated in-place.

```able
x := 5
y := x * 2
x = 8

point := Point { x: 1, y: 2 }
point.x = point.x + 1
```

Patterns can appear in declarations, assignment, parameter lists, `for`, and `match`.

```able
Point { x, y } := point
{x, y} = { x + 1, y * 2 }

[first, ...rest] := numbers
Container { items: [ Data { id, ... }, ...others ] } := payload

if user := maybe_lookup() {
  greet(user)
} or {
  log("not found")
}
```

Underscore `_` ignores a position. Pattern mismatches yield `Error` values (falsy per spec §6.11).

## 3. Control Flow

Blocks are expressions; the last expression yields the value. Use `do { ... }` in expression position.

```able
message := do {
  total := compute_total()
  `Total: ${total}`
}
```

### 3.1 Conditionals

`if … or …` chains behave like Julia’s `if/elseif/else`, returning the branch value. Only `false`, `nil`, and values of type `Error` are falsy.

```able
grade = if score >= 90 { "A" }
        or score >= 80 { "B" }
        or { "C or lower" }
```

### 3.2 Loops

`while` and `for` loops evaluate to `void`.

```able
while counter < 3 {
  print(counter)
  counter = counter + 1
}

for {item, idx} in items.enumerate() {
  print(`items[${idx}] = ${item}`)
}

for n in 0...len { total = total + n }   ## exclusive upper bound
```

### 3.3 Breakpoints

`breakpoint 'label { ... }` establishes an exit point; `break 'label value` unwinds to it. Plain `break` targets the innermost loop.

```able
winner = breakpoint 'scan {
  for cand in candidates {
    if cand.score > threshold {
      break 'scan cand
    }
  }
  nil
}
```

## 4. Functions

Functions are first-class. Implicit returns use the last expression; `return` exits early.

```able
fn greet(name: string) -> string {
  `Hello ${name}`
}

fn make_adder(delta: i32) -> (i32 -> i32) {
  { value => value + delta }
}

fn find_first_negative(items: Array i32) -> ?i32 {
  for item in items {
    if item < 0 { return item }
  }
  nil
}
```

### 4.1 Lambdas & Placeholders

- Lambda shorthand: `{ x, y => x + y }`
- Verbose anonymous function: `fn(x: i32) -> string { x.to_string() }`
- Placeholder lambdas: `numbers.map(@ * 2)`, `add_10 = add(@, 10)`

Trailing lambdas mirror Julia’s `do` blocks:

```able
sum = numbers.reduce(0) { acc, n => acc + n }
```

### 4.2 Pipelines

`|>` forwards values. The RHS must reference `%` (topic) or evaluate to a unary callable.

```able
result = value
  |> %.normalize()
  |> clamp(@, 0, 1)
  |> (@ * 100)
```

## 5. Types

Able is statically typed with inference. Annotations use `name: Type`. Runtime values must have concrete types.

### 5.1 Primitive & Collection Types

- Numbers (`i32`, `u64`, `f64`), `bool`, `char`, `string`, `nil`, `void`
- Arrays: `[1, 2, 3]`, `numbers.size()`, `numbers[0]`
- Ranges: `1..10` (inclusive), `0...len` (exclusive)
- Maps: `scores["Ada"] = 42` via stdlib helpers

### 5.2 Structs

```able
struct User {
  id: i64,
  name: string,
  email: ?string
}

struct Pair T U (T, U)
struct Active

user := User { id: 1, name: "Ada", email: nil }
```

Fields are mutable. Structural updates copy and override fields:

```able
user = User { ...user, email: "ada@example.com" }
```

### 5.3 Generics

```able
struct Box T { value: T }

fn wrap<T>(value: T) -> Box T {
  Box { value }
}

fn describe<T: Display + Clone>(value: T) -> string {
  copy := value.clone()
  copy.to_string()
}
```

Type expressions use space-separated arguments (`Array string`, `Map string User`). Parenthesize for grouping (`Map string (Array i32)`).

## 6. Methods

### 6.1 Inherent Methods

```able
struct Counter { value: i32 }

methods Counter {
  fn new(start: i32) -> Self { Counter { value: start } }

  fn #increment(by: i32 = 1) -> void {
    #value = #value + by
  }
}
```

`#field` is shorthand for accessing the first parameter’s member (`self` by convention).

### 6.2 Interfaces & Implementations

Interfaces capture behaviour; `impl` provides concrete bodies.

```able
interface Display for T {
  fn to_string(self: Self) -> string
}

impl Display for User {
  fn to_string(self: Self) -> string {
    `User(${self.id}, ${self.name})`
  }
}
```

Specificity rules prefer concrete implementations over generic ones. Name an `impl` to disambiguate:

```able
Sum = impl Monoid for i32 {
  fn id() -> Self { 0 }
  fn op(self: Self, other: Self) -> Self { self + other }
}

total = Sum.op(2, 3)
```

Values typed as interfaces perform dynamic dispatch:

```able
fn render(value: Display) -> string {
  value.to_string()
}
```

## 7. Constructors & Updates

Struct literals require all fields. Positional structs use tuple syntax. Singleton structs carry one value equal to their type name.

```able
origin := Point { x: 0, y: 0 }
polar := Polar { radius: 1.0, theta: 0.0 }
polar = Polar { ...polar, theta: pi / 2.0 }

state := Active
```

Arrays, ranges, channels, and maps rely on stdlib builders (see `stdlib/` and `design/channels-mutexes.md`).

## 8. Pattern Matching & Unions

`union` declares algebraic data types. `?T` (`nil | T`) and `!T` (`Error | T`) are shorthands.

```able
union Shape =
    Circle { radius: f64 }
  | Rectangle { width: f64, height: f64 }
  | Triangle { a: f64, b: f64, c: f64 }

area = shape match {
  case Circle { radius } => 3.141592653589793 * radius * radius,
  case Rectangle { width, height } => width * height,
  case Triangle { a, b, c } => {
    s = (a + b + c) / 2.0
    (s * (s - a) * (s - b) * (s - c)).sqrt()
  }
}
```

Patterns support guards (`case v: i32 if v > 0`) and interface types. Non-exhaustive matches raise exceptions at runtime.

`else` bridges optional/Result values:

```able
username = maybe_name else { "guest" }
config = load_config() else { |err|
  log(`using anonymous mode: ${err.message()}`)
  default_config()
}
```

`expr!` unwraps `?T`, `!T`, or `nil | Error | T`, propagating failures.

## 9. Modules

Package paths mirror directory structure (hyphen → underscore). Declare subpackages with `package analytics.reports`. Top-level definitions are public unless marked `private`.

```able
import math
import math.{sqrt, pow}
import io as console
import io.{print as println}
```

`import pkg.*` brings every public name—prefer selective imports. `dynimport` resolves runtime-defined packages in interpreter builds.

Executables arise from packages defining `fn main()`. Use `os.args()` for CLI arguments and `os.exit(code)` for non-zero termination.

### 9.1 Standard Math Helpers

The bundled `math` module includes common constants and helpers so you can keep numeric code concise:

- Constants — `math.pi()`, `math.tau()`, `math.half_pi()`, `math.e()`.
- Basic helpers — `math.abs`, `math.min`, `math.max`, `math.clamp`, integer counterparts (`abs_i64`, `clamp_i64`), and `math.sign`.
- Transformations — `math.deg_to_rad`, `math.rad_to_deg`, `math.lerp`, `math.approx_eq` (with tolerance).
- Core functions — `math.sqrt` (with UFCS so `value.sqrt()` works), integer-exponent `math.pow`, and `math.hypot`.
- Rounding & range helpers — `math.floor`, `math.ceil`, `math.round`, `math.trunc`, `math.fract`, `math.clamp01`, `math.inverse_lerp`, `math.remap`, `math.remap_clamped`, `math.wrap`, `math.wrap_angle_radians`, plus integer helpers `math.gcd` / `math.lcm`.

Import what you need:

```able
import math.{sqrt, pow, pi, hypot}
radius := 12.5
area := pi() * pow(radius, 2)
diag := hypot(3.0, 4.0)
```

### 9.2 Numeric Interfaces

Import `able.core.numeric` to access the algebraic interfaces (`Numeric`, `Integral`, `Signed`, `Unsigned`, `NumericConversions`, etc.). The `able.numbers.primitives` module provides implementations of those interfaces for the built-in integer/float types, so once you import it you can call helpers directly on literals/values:

```able
import able.core.numeric
import able.numbers.primitives

expect((-5).abs()).to(eq(5))
result := 17.div_mod(5)
expect(result.remainder).to(eq(2))
expect(12.0.to_i32()).to(eq(12))
expect(3.75.floor()).to(eq(3.0))
expect((-2.25).fract()).to(eq(0.75))
```

These traits back higher-level helpers like `sum`, `product`, and upcoming collection reducers.

### 9.3 Rational Numbers

`able.numbers.rational` introduces an exact `Rational` type implemented purely in Able. Rationals always stay in lowest terms, keep denominators positive, and participate in the numeric/typeclass hierarchy (`Numeric`, `Signed`, `Fractional`, `NumericConversions`, `Display`, `Eq`, `Ord`).

```able
import able.numbers.rational.{Rational}

half := Rational.new(1, 2)
third := Rational.new(1, 3)
sum := half.add(third)        ## => 5/6
scaled := sum.mul(Rational.new(7, 5))

expect(sum.round().to_i64()).to(eq(1))
expect(scaled.to_f64()).to(be_within(1e-12, 1.1666666667))
```

Use `Rational.from_i64` for integers, `reciprocal`/`abs`/`clamp` for helpers, and the conversion APIs (`to_i32`, `to_f64`, …) when an exact fraction needs to become a primitive. All operations raise the usual numeric errors (`DivisionByZeroError`, `OverflowError`, or `NumericConversionError`) when invariants are violated.

### 9.4 128-bit Integers

The `able.numbers.int128` module provides an `Int128` struct that stores signed 128-bit values as two `u64` chunks. Even though the interpreters already expose primitive `i128` literals, `Int128` is useful when you need explicit control over the representation (serialization, bit fiddling, deterministic arithmetic across runtimes).

```able
import able.numbers.int128.{Int128}

total := Int128.from_i128(0_i128)
value := Int128.from_i128(2_i128.pow(96)!)
total = total.add(value)
total = total.sub(Int128.from_i64(1))

expect(total.to_string()).to(eq("79228162514264337589248983039"))
expect(total.to_i64()).to(raise_error()) ## overflows native i64
```

The companion `able.numbers.uint128` module exposes an unsigned variant (`UInt128`) that spans `[0, 2^128 - 1]` while still storing the value as `(high: u64, low: u64)`. `UInt128` implements the unsigned numeric stack (`Numeric`, `Unsigned`, `NumericConversions`, `Eq`, `Ord`) and keeps its operations checked:

```able
import able.numbers.uint128.{UInt128}

mask := UInt128.from_u128(0xffffffffffffffffffffffffffffffffffff_u128)
half := mask.div(UInt128.from_u64(2))

expect(mask.add(UInt128.one())).to(raise_error()) ## overflow
expect(half.rem(UInt128.from_u64(3)).to_u64()).to(eq(1_u64))
```

Core helpers for both structs mirror each other: `add`, `sub`, `mul`, `div`, `rem`, `negate` (signed only), `abs`, `clamp`, `compare`, plus the numeric conversions (`to_i32`, `to_u32`, `to_i64`, `to_u64`, `to_f64`). Division/remainder raise `DivisionByZeroError`; out-of-range conversions raise `NumericConversionError`; arithmetic that exceeds 128 bits raises `OverflowError`.

## 10. Error Handling

Able blends Option/Result unions with exceptions.

### 10.1 Propagation with `!`

```able
fn read_config(path: string) -> !Config { ... }

fn boot() -> !void {
  config := read_config(path)!
  port := config.port else { 8080 }
  start_server(config.host, port)!
}
```

`expr!` unwraps successes and returns early on `nil` or `Error`. The enclosing function must return a compatible union.

### 10.2 Exceptions

`raise` throws errors; `rescue` handles them.

```able
fn divide(a: i32, b: i32) -> i32 {
  if b == 0 { raise DivideByZeroError {} }
  a / b
}

ratio = do { divide(total, count) } rescue {
  case _: DivideByZeroError => 0,
  case e: Error => {
    log(`failed: ${e.message()}`)
    -1
  }
}
```

Use `ensure`/`rethrow` (spec §11.3) for finally-style cleanup and propagation.

## 11. Concurrency

Able mirrors Go-style concurrency with `proc`, `spawn`, channels, and mutexes.

### 11.1 `proc`

```able
handle: Proc string = proc fetch_data(url)

proc_flush(32)

handle.status() match {
  case Pending => log("still fetching..."),
  case Failed { error } => log(`fetch failed: ${error.message()}`),
  case Cancelled => log("cancelled"),
  case Resolved => void
}

content = handle.value() else { |err|
  log(`background fetch failed: ${err.message()}`)
  "fallback"
}
```

`Proc T` exposes `status()`, `value() -> !T`, and `cancel()`. Inside async bodies, use `proc_yield()`, `proc_cancelled()`, and `proc_flush(limit?)`.

### 11.2 `spawn`

```able
future_total: Future i64 = spawn sum_big_array(data)

print("Computing...")
total = future_total
print(`Total: ${total}`)
```

Evaluating a `Future T` blocks and yields `T`, re-raising exceptions from the task.

### 11.3 Channels & Mutexes

`Channel T` and `Mutex` mirror Go semantics.

```able
ch: Channel string = Channel.new(0)

producer = proc do {
  for name in names { ch.send(name) }
  ch.close()
}

for value in ch {
  print(`hello ${value}`)
}
```

Channels block on send/receive, support `try_send`/`try_receive`, and return `nil` when closed and drained. Mutexes are non-reentrant; use helpers that guarantee `unlock()` on every path (see `design/channels-mutexes.md`).

## 12. Tooling & Next Steps

- **Spec-first:** align with `spec/full_spec_v10.md`; update it when behaviour changes. Track open items in `spec/todo.md`.
- **Shared fixtures:** edit `.able` sources, run `bun run scripts/export-fixtures.ts`, then execute both interpreters (`bun run scripts/run-fixtures.ts`, `go test ./pkg/interpreter`).
- **Design alignment:** capture decisions in `design/` and update `PLAN.md`/`interpreter10/PLAN.md`.
- **Testing:** `./run_all_tests.sh` orchestrates TypeScript tests, fixture runs, and Go parity checks; use `--typecheck-fixtures=strict` before landing runtime changes.
- **Further reading:** explore `examples/*.able`, interpreter tests (`interpreter10/test/`), and concurrency notes in `design/`.

Able emphasises clarity, expressive types, and cross-runtime parity. When semantics are unclear, exercise fixtures in both interpreters, compare with the Go implementation, and document findings for future contributors.
