# Able v12 Manual

Able couples a concise, block-oriented syntax with static types, algebraic data types, async/runtime helpers, and host interop. This manual mirrors the Julia language manual’s progression (getting started → basics → control flow → types → methods/interfaces → modules → async/parallelism → metaprogramming/interop) while staying faithful to the v12 specification (`spec/full_spec_v12.md`). Treat the spec as the authority; the manual is the definitive learning resource and bridges directly to runnable interpreter behaviour.

## Contents
- 1. Getting Started
- 2. Syntax & Expressions
- 3. Bindings & Patterns
- 4. Types & Data Structures
- 5. Functions & Functional Tools
- 6. Control Flow & Pattern Matching
- 7. Methods, Interfaces, & Dispatch
- 8. Error Handling
- 9. Concurrency & Async
- 10. Modules & Packages
- 11. Standard Library Essentials
- 12. Dynamic Metaprogramming
- 13. Host Interop
- 14. Program Entry & Tooling

## 1. Getting Started

### 1.1 Hello, Able

Able source lives in packages. A package with a public `fn main() -> void` produces an executable:

```able
package helloworld

fn main() {
  print("Hello, Able!")
}
```

Blocks are expressions; the last expression is the value. Braces delimit blocks—indentation is cosmetic.

### 1.2 Resources

- Spec (authoritative semantics): `spec/full_spec_v12.md`
- Roadmap/status: `PLAN.md`
- Interpreters: `v12/interpreters/go` (tree-walker + bytecode)
- Fixtures/examples: `v12/fixtures`, `v12/examples`
- Design notes: `design/` (e.g., parity plans, typechecker plan, regex plan)

### 1.3 Toolchain Setup

```bash
# Go interpreters
cd v12/interpreters/go
go test ./...

# Go interpreter
cd ../go
go test ./...
```

Unified harness (runs Go suites and fixtures):

```bash
./run_all_tests.sh --version=v12
```

### 1.4 Running Able Code

- Execute fixtures via Go harness: `go test ./pkg/interpreter -run ExecFixtures`
- Export fixtures after edits: run the Go exporter via `v12/export_fixtures.sh`
- Go parity: `go test ./pkg/interpreter`
- Add ad-hoc `.able` files under `examples/` or a package and run through the interpreter CLIs.

### 1.5 Project Layout

- `v12/`: active workspace (interpreters, parser, stdlib, fixtures)
- `spec/`: language specs (v1–v12). Use only v12 for behaviour.
- Archived workspace at the repo root: frozen—do not edit unless explicitly requested.
- Standard library ships as `able.*` packages; the namespace is reserved.
- Kernel builtins are minimal: host-provided globals (print, scheduler, channels/mutexes/hashers), array helpers, and String byte view only (`len_bytes`, `bytes`). Higher-level string/regex/collection helpers live in the Able stdlib layered on top.

## 2. Syntax & Expressions

### 2.1 Lexical Basics

- Line comments: `## comment`. Block comment syntax is TBD.
- Identifiers: ASCII letters, digits, `_`; cannot start with a digit.
- Statement termination: newline or `;`. Trailing commas are allowed in lists/params.
- Reserved tokens include `:=`, `=`, `->`, `_`, `?`, `!`, `@`/`@n` placeholders, `#`, `|`, `|>`, `...`, `..`, and backtick-string delimiters.

### 2.2 Literals

- Numbers: `123`, `1_000`, `3.14`, `2e10`, `0xff`, `0b1010`.
- Booleans: `true`, `false`; Nil: `nil`; Void literal is the absence of value in `void` contexts.
- Characters: `'a'`, `'\\n'`. Strings: double quotes for plain literals (`"hello"`), backticks for interpolation `` `Hello ${name}` `` (UTF-8; kernel exposes byte length/iterator, stdlib adds char/grapheme helpers).
- Arrays: `[1, 2, 3]` with element type `Array i32`. Structs: `Point { x: 1, y: 2 }`. Map literal support is via stdlib helpers.

### 2.3 Blocks as Expressions

`do { ... }` evaluates to its final expression (or `void` if empty):

```able
msg := do {
  total := compute()
  `Total: ${total}`
}
```

### 2.4 Truthiness

Only `false`, `nil`, and values of type `Error` are falsy. Everything else is truthy (`void` is truthy).

### 2.5 Operators & Precedence

Usual arithmetic/logical operators plus:

- Assignment: `:=` (declare), `=` (reassign).
- Range: `a..b` (inclusive), `a...b` (exclusive).
- Safe navigation: `expr?.field` / `expr?.fn()` short-circuits to `nil` on `nil`.
- Pipe forward `|>` (callable-only) and low-precedence `|>>`: `value |> normalize` or apply a placeholder callable `value |> (@ + 1)`; `|>>` runs after assignment without extra parens.
- Placeholders: `@` or numbered `@2` inside lambdas to build callables; `%` is modulo.
- Safe navigation and indexing rely on stdlib interfaces (Index/IndexMut, etc.).

Consult the spec’s precedence table; `|>>` is the loosest, `|>` sits between `||/&&` and assignment.

### 2.6 String Interpolation

Backtick strings interpolate with `${expr}` and preserve UTF-8. Escape backticks with `` `\`` `` and `${` with `${"{"}`.

## 3. Bindings & Patterns

### 3.1 Declarations vs Assignment

- `name := expr` declares and initializes (type inferred unless annotated).
- `name = expr` reassigns an existing binding.

Bindings are mutable; structs/arrays/maps have mutable fields/elements.

### 3.2 Patterns

Patterns work in `:=`, `=`, `for`, `match`, parameter lists, and `else` handlers:

- Identifier: `x`
- Wildcard: `_`
- Typed identifier: `count: i32`
- Struct (named fields): `Point { x, y }`
- Struct (positional): `{ x, y }`
- Array: `[first, ...rest]`
- Nested and typed patterns: `case Err { msg: String }`

Pattern mismatch yields an `Error` (falsy). Typed patterns act as guards and fail if the runtime type does not conform.

### 3.3 Functional Updates & Mutation

Struct fields are mutable by default: `point.x = 4`. Functional copy-with-update:

```able
point = Point { ...point, y: point.y + 1 }
```

## 4. Types & Data Structures

### 4.1 Type Expressions & Generics

Type arguments are space-delimited: `Array String`, `Map String (Array i32)`. Generic parameters use `T`, `U`, etc.; constraints use interfaces (`T: Display + Clone`). Unused generic parameters can be inferred.

### 4.2 Primitive & Composite Types

- Primitives: `i8/i16/i32/i64/i128`, unsigned variants, `f32/f64`, `bool`, `char`, `String`, `nil`, `void`.
- Arrays: typed, mutable, iterable; see stdlib helpers (§11).
- Ranges: `Range` values from `a..b` (inclusive) or `a...b` (exclusive), usable in `for`.
- Maps: constructed via stdlib; indexing uses `Index`/`IndexMut`.
- Generator literal: `Iterator { gen => ... gen.yield(v) ... }` exposes an `Iterator T`.

### 4.3 Structs

Forms:

- Singleton: `struct Active`
- Named fields: `struct User { id: i64, name: String, email: ?String }`
- Positional (named-tuple style): `struct Pair T U (T, U)`

Instantiation matches the declaration form. Fields are mutable unless an API restricts them.

### 4.4 Unions & Nullability

`union` introduces sum types:

```able
union Shape =
    Circle { radius: f64 }
  | Rectangle { width: f64, height: f64 }
  | Triangle { a: f64, b: f64, c: f64 }
```

`?T` ≡ `nil | T`. `!T` is shorthand used with propagation (`Error | T` or `nil | Error | T` depending on nesting). Construct variants with their struct syntax; match or type guards consume them.

### 4.5 Type Aliases

`type PairStr = Pair String String`. Aliases may be generic and recursive; they don’t create nominal distinctions. Aliases can be used in `methods`/`impl` definitions and across modules.

## 5. Functions & Functional Tools

### 5.1 Named Functions

```able
fn greet(name: String) -> String { `Hello ${name}` }
fn add<T: Numeric>(a: T, b: T) -> T { a + b }
```

- The return type defaults to the last expression; `-> void` is explicit.
- Early return: `return expr`.
- Generic parameters may be explicit or inferred.

### 5.2 Anonymous Functions & Lambdas

- Verbose: `fn(x: i32) -> String { x.to_string() }`
- Lambda shorthand: `{ x, y => x + y }`
- Trailing lambda: `items.map { x => x * 2 }`
- Closures capture lexical bindings by reference; mutation uses outer mutability rules.

### 5.3 Shorthand Notation

- Implicit first param (`#member`): inside functions/methods, `#field` means `self.field`.
- Implicit self parameter: `fn #increment(by: i32) { #value = #value + by }`
- Placeholder lambdas: `numbers.map(@ * 2)`, `add_prefix = add("hi", @2)`
- Pipe forward: `value |> (@.trim().to_string()) |> (@ + "!")`

### 5.4 Calls & Callable Values

Standard call syntax plus method-call sugar (`value.method(args)`) that resolves through `methods`/`impl`. Any value implementing `Apply` can be invoked like `callable(arg)`; interfaces drive operator overloading (§7).

## 6. Control Flow & Pattern Matching

### 6.1 Conditionals

`if/elsif/else` chains yield expression values:

```able
grade = if score >= 90 { "A" }
        elsif score >= 80 { "B" }
        else { "C or lower" }
```

### 6.2 Pattern Matching

`match` is an expression. Patterns may bind, destructure, and guard:

```able
desc = shape match {
  case Circle { radius } => radius * radius * 3.1415926535,
  case Rectangle { width, height } => width * height,
  case Triangle { a, b, c } if a + b > c => {
    s = (a + b + c) / 2.0
    (s * (s - a) * (s - b) * (s - c)).sqrt()
  }
}
```

Non-exhaustive matches raise at runtime. Matching on interface types uses runtime dictionaries.

### 6.3 Loops

- `while cond { ... }`
- `for pattern in iterable { ... }` uses `Iterable`/`Iterator`.
- `loop { ... }` is an expression; `break value` yields `value`.
- `continue` skips to the next iteration.
- Range loops: `for i in 0...len { ... }`

### 6.4 Breakpoints & Non-Local Break

`breakpoint 'label { ... }` establishes an exit; `break 'label value` unwinds to it. Plain `break` targets the innermost loop.

### 6.5 Safe Navigation

`value?.field` and `value?.fn()` short-circuit to `nil` if the left side is `nil`; otherwise they behave like standard access/call. Chaining is allowed.

## 7. Methods, Interfaces, & Dispatch

### 7.1 Inherent Methods (`methods`)

Define methods directly on a type:

```able
methods Counter {
  fn new(start: i32) -> Self { Counter { value: start } }
  fn #inc(by: i32 = 1) { #value = #value + by }
}
```

Method-call syntax (`counter.inc(2)`) is sugar for calling these functions with an implicit receiver argument.

### 7.2 Interfaces

Interfaces describe behaviour; `Self` refers to the implementing type when omitted from `for`:

```able
interface Display for T { fn to_string(self: Self) -> String }
interface Iterable T { fn iterator(self: Self) -> (Iterator T) }
```

Interfaces may have default bodies (including static methods).

### 7.3 Implementations (`impl`)

```able
impl Display for User {
  fn to_string(self: Self) -> String { `User(${self.id}, ${self.name})` }
}
```

- Generic and higher-kinded `impl` forms are allowed (`impl Hashable A for Array`).
- Overlapping impls pick the most specific; ambiguity is an error until imports are pruned.
- `impl` visibility follows normal rules; interface-typed values carry the impl dictionary for dynamic dispatch even when the impl is not in scope.
- Named impls provide explicit disambiguation (`Sum = impl Monoid for i32 { ... }`).

### 7.4 Method Call Resolution

Method calls search inherent methods (`methods`), then applicable `impl` methods in scope, using named impls only when explicitly qualified. Operator syntax (`+`, `[]`, `call`, comparison, Display, Error) lowers to the standard interface catalogue (§11.4).

## 8. Error Handling

### 8.1 Option/Result Unions and Propagation

- `?T` is `nil | T`; `!T` is `Error | T` (or nested with `nil`).
- `expr!` unwraps and propagates `nil`/`Error` early from `!` or `?` expressions.

```able
fn boot() -> !void {
  config := read_config(path)!       ## returns early on Error
  port := config.port else { 8080 }  ## handle nil/err inline
  start_server(config.host, port)!
}
```

### 8.2 `else` Handlers

`value else { fallback }` handles `nil` or `Error` payloads. The handler may capture the error:

```able
name = maybe_name else { "guest" }
content = read_file(path) else { |err| raise FileError { msg: err.message() } }
```

### 8.3 Exceptions

`raise` throws an `Error` value; `rescue` catches:

```able
ratio = do { divide(total, count) } rescue {
  case _: DivideByZeroError => 0,
  case e: Error => {
    log(`failed: ${e.message()}`)
    -1
  }
}
```

Use `ensure` for finally-style cleanup and `rethrow` to propagate within `rescue`.

## 9. Concurrency & Async

Able mirrors Go-style semantics with shared interfaces across runtimes.

### 9.1 `spawn` (Futures)

`spawn expr` starts an asynchronous task and returns `Future T`:

```able
handle: Future String = spawn fetch(url)
future_flush(32) ## advance cooperative scheduler (TS runtime exposes this helper)

handle.status() match {
  case Pending => log("waiting..."),
  case Failed { error } => log(error.message()),
  case Resolved => void,
  case Cancelled => log("cancelled")
}

body = handle.value() else { "fallback" }
```

`Future` exposes `status()`, `value() -> !T`, and `cancel()`. Helpers inside async bodies: `future_yield()`, `future_cancelled()`, `future_flush(limit?)`, `future_pending_tasks()` (diagnostic).

### 9.2 Future handle vs value view

`Future T` acts as a handle in non-`T` contexts and as a value in `T` contexts. Evaluating it (or calling `value()`) blocks until the task resolves, re-raising task exceptions. Futures memoize results and expose `status()`/`value()` via the handle when lifecycle control is needed.

### 9.3 `await` & `Awaitable`

`await [arms...]` selects the first ready `Awaitable`, applying its callback and returning that result. Arms expose `is_ready`, `register(waker)`, and `commit` to integrate with the scheduler. `Await.default` provides a single default arm. Runtimes must choose fairly when multiple arms are ready and honour cancellation by waking blocked tasks.

### 9.4 Channels & Mutexes

- `Channel T`: Go semantics (rendezvous when unbuffered, FIFO buffering). `send`, `receive -> ?T`, `try_send`, `try_receive`, `close`, `is_closed`. Iteration drains until closed. Await arms `try_recv/try_send` integrate with `await`. Errors: `ChannelClosed`, `ChannelNil`, `ChannelSendOnClosed`, `ChannelTimeout`.
- `Mutex`: non-reentrant mutual exclusion. `lock`, `unlock`; helper patterns ensure unlock on every path. Errors include `MutexUnlocked` (and reserved `MutexPoisoned`).

### 9.5 Cancellation & Re-entrancy

Blocking operations must observe task cancellation. `value()` calls are re-entrant: nested waits must continue to make progress in both runtimes.

## 10. Modules & Packages

### 10.1 Package Names & Layout

- Root package name comes from `package.yml` (`name` field, hyphen → underscore).
- Directory structure maps to package paths; hyphenated directories become underscores.
- `package foo.bar` inside a file appends segments; omitting `package` uses the directory-derived package.

### 10.2 Imports

```able
import math
import math.{sqrt, pow}
import io::console
import io.{print::println}
import math.*
```

Use selective imports to avoid collisions. Imports may appear in any scope.

### 10.3 Visibility

Top-level items are public unless prefixed with `private`. An `impl` participates at a call site only if in scope; interface-typed values carry their impl dictionaries for dynamic dispatch even when the impl is not imported.

### 10.4 Dynamic Imports (`dynimport`)

Binds names from runtime-defined packages (interpreted mode). Resolution mirrors static imports but happens at runtime; missing packages/names raise `Error`.

### 10.5 Module Search Order & Standard Library

Loader search order (deduplicated):

1. Workspace roots (entry manifest directory and ancestors)
2. Current working directory / CLI-provided roots
3. `ABLE_PATH` entries
4. `ABLE_MODULE_PATHS` entries (preferred override for fixtures/REPL)
5. Bundled kernel/stdlib roots discovered near the workspace/executable (or pinned via the resolver)

Collisions of the same package path across roots are errors. `able.*` is reserved for the stdlib chosen by the resolver; tooling treats the bundled copy as read-only and uses `ABLE_MODULE_PATHS`/lockfile pins for overrides (no stdlib-specific env knob).

## 11. Standard Library Essentials

### 11.1 Strings & Graphemes

- Length helpers: `len_bytes`, `len_chars`, `len_graphemes`
- `substring`, `split`, `replace`, `starts_with`, `ends_with`
- Iterators over bytes/chars/graphemes; interpolation uses UTF-8 data

### 11.2 Arrays

Helpers: `size`, `push`, `pop`, `get`, `set`, `clear` with `IndexError` on invalid indices. Arrays satisfy `Iterable`.

### 11.3 Text Regex (`able.text.regex`)

Deterministic (RE2-style) regexes with compile/match/split/replace helpers, regex sets, streaming scanner, and grapheme-aware options. Literal-only compile/match/find_all are available; metacharacters and the advanced helpers still raise `RegexUnsupportedFeature` until the full engine lands.

### 11.4 Interface Catalogue (language-supported)

Syntax sugar lowers to stdlib interfaces:

- Indexing: `Index`, `IndexMut`
- Iteration: `Iterable`, `Iterator`, `Range`
- Callable values: `Apply`
- Operators: arithmetic/bitwise, comparison, hashing
- Display & Error: `Display`, `Error`
- Object utilities: `Clone`, `Default`
- Concurrency: `Future`, `FutureError`, `Awaitable`

Collection and async types in stdlib implement the relevant interfaces.

### 11.5 Numeric & Utility Modules

Math helpers (`abs`, `min/max`, `clamp`, `pow`, `sqrt`, rounding), numeric interfaces/impls, rational and 128-bit integer helpers, plus other bundled modules under `able.*`.

## 12. Dynamic Metaprogramming

Interpreted execution can create/extend dynamic packages and import them via `dynimport`. Package objects expose late-bound namespaces. Dynamic calls are shape-checked at runtime; static code may adapt dynamic values to interfaces when required.

## 13. Host Interop

Embed host-language code via package-scope `prelude <target> { ... }` and `extern <target> fn ... { ... }` bodies (Go, Crystal, JavaScript, Python, Ruby). Only core primitive/container mappings are supported (copy-in/copy-out for arrays). `host_error(message)` converts host failures to Able `Error`s. Extern bodies execute in-place and may block; dynamic packages cannot contain extern bodies. Multiple extern targets may back a function; Able bodies act as fallback when present.

## 14. Program Entry & Tooling

- Any package with public `fn main() -> void` produces a binary named after the package path. Multiple binaries are allowed.
- `os.args()` provides CLI args; returning from `main` exits 0; unhandled exceptions exit 1; use `os.exit(code)` for custom codes.
- Background `spawn` tasks are not awaited when `main` returns—join explicitly if needed.
- Language implementation testing (fixtures/parity): keep tree-walker + bytecode interpreters in sync. Run `go test ./...`, `go test ./pkg/interpreter`, and `./run_all_tests.sh --version=v12` before landing changes.
- Fixtures: update `v12/fixtures`, export via the Go exporter (TODO), and ensure tree-walker/bytecode parity stays green.
- User-facing testing (Able programs): `able test` plus the `able.spec` DSL (backed by `able.test.*`) are planned for end-user test suites and are separate from fixture/parity work.

Able emphasises clarity, parity between interpreters, and spec-first behaviour. When in doubt, check the v12 spec, add fixtures, and validate in both runtimes.
