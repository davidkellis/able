# Able v11 Spec TODOs

This list tracks language features and spec work deferred to the v11 cycle.

- [x] **Mutable `=` assignment semantics**: plain `=` reassignments currently no-op in the TypeScript interpreter (`x = x - 1` leaves `x` unchanged), which makes every imperative loop spin forever. v11 needs a spec note + implementation work to guarantee reassignment mutates the existing binding per §5.3. (Documented in §5.3.1.)
- [x] **Map literals**: Specify a literal syntax (`#{ key: value, ... }`), typing/inference rules (empty literal requirements, key/value unification), evaluation order, duplicate-key behaviour, and optional spread/update forms. Update AST contract and interpreter semantics accordingly. (Documented in §6.1.9.)
- [x] **Struct functional update semantics**: v10 already supports `Struct { ...source, field: value }` for named structs (same struct type only; later fields overwrite earlier ones; positional structs disallow spreads). The spec mentions the pattern but doesn’t fully spell out semantics or generic cases. For v11, tighten the documentation and add shared fixtures/tests that cover generic updates and edge cases. (Documented in §4.5.2.)
- [x] **Type alias declarations**: Define syntax (`type Name = ...`), generic forms, visibility, and evaluation semantics (including interaction with `import`/`methods`/`impl`). (Documented in §4.7.)
- [x] **Safe member access (`?.`)**: Nail down grammar, typing (nullable receiver propagation), evaluation semantics, and AST/runtime updates for the safe navigation operator. (Documented in §6.3.4.)
- [x] **Typed `=` declarations**: Decide whether `name: Type = expr` should be first-class declaration sugar (without needing `:=`) as hinted in §5.2.7, and update the spec + interpreters/typecheckers if we bless it. Capture scope/shadowing rules and diagnostics for mixed declaration/reassignment cases. (Documented in §5.1–§5.1.1.)
- [x] **Contextual integer widening & literal typing**: Fully specify the numeric literal inference/promotion rules promised in §6.3.2 (P1), including how integer literals adopt the surrounding type, how mixed-width arithmetic widens operands, and what diagnostics fire when narrowing would occur. Ensure both interpreters and fixtures enforce the v11 rules consistently. We want to be able to assign an integer literal to a variable of a wider width than necessary to represent the value, and we want to be able to assign a narrower width integer to a variable of a wider width. (Documented in §6.1.1 and §6.3.2.)
- [x] **Optional generic type parameter declarations**: Function definitions
like `fn choose_first<T, U>(first: T, second: U) -> T where T: Display + Clone, U: Display { ... }` may omit the explicit declaration of `<T, U>` when those types are not already defined in the current scope, and the compiler should infer that since they are used as types in the function signature and within the function body, that they are generic type parameters in the scope of the function definition. Similar rules apply to struct and interface definitions.
- [x] **Await/multiplexed async facility**: v10 punted on language-level multiplexing. v11 needs a generalized `await` surface that covers channel send/receive plus other async operations (timers, sockets), specifies fairness, cancellation, and helper callbacks. (Documented in §12.6.)
- [x] Specify how stdlib channel/mutex helpers surface error structs (`ChannelClosed`, `ChannelNil`, `ChannelSendOnClosed`) so host runtimes map native failures to the documented Able `Error` types. (Documented in §12.7.)
- [x] **`loop { ... }` expression**: the parser/interpreters still reject the `loop` keyword used throughout the channel + iterator examples. Define the grammar, ensure `break`/`continue` work with it, and document its value semantics (loops evaluate to `void` unless broken with a value). (Documented in §8.2.3.)
- [x] **Array/String runtime API**: real code expects `Array.size()`, `.push()`, `.get()`, `.set()`, and string helpers such as `.len()`, `.substring()`, `.split()`, `.replace()`, `.starts_with()`, `.ends_with()`. Today arrays are inert literals, so every LeetCode/Rosetta sample aborts. v11 must specify the minimal builtin surface (likely via stdlib modules) and ensure the TypeScript + Go interpreters expose it. (Documented in §6.8 and §6.1.5.)
- [x] **Stdlib packaging & module search paths**: `able run` needs a versioned stdlib bundle plus module-resolution rules so `able.*` packages load without manual wiring. (Documented in §13.6–§13.7.)
- [x] Specify the `able.text.regex` module, including default code-point semantics, optional grapheme-aware execution, result types, and error reporting. (Documented in §14.2.)
- [x] Document the standard library `string`, `Grapheme`, and iteration helpers (byte, char, grapheme views) together with the byte-oriented indexing rules referenced by the spec. (Documented in §6.12.1.)


## Candidate features from other popular languages (undecided if we should add them)

- **Operator overloading and custom operators**: No user-defined operator overloads or new operator definitions (present in Rust, Scala, Ruby).
- **Function/method overloading**: No ad‑hoc overloading by arity or parameter types; one signature per name (common in Scala/Ruby).
- **Named and default arguments**: Calls are positional-only; no default parameter values or keyword/named args (common in Scala/Ruby).
- **Variadic parameters and argument splats**: No varargs in function definitions and no splat/spread in call sites (common in Scala/Ruby).
- **Tuple types and tuple literals**: Not part of the core; you use positional structs instead (native tuples are common in Rust/Scala).
- **Macro/annotation/derive system**: No compile-time macros, annotations/attributes, or derive-like auto-impls (Rust/Scala have these; Ruby has rich runtime metaprogramming).
- **Comprehension syntax**: No list/set/for‑yield comprehensions; iteration is via `for`/iterators (Scala has for‑comprehensions).

Notes:
- In short: the big missing “usually expected” niceties are overloading (operators/functions), named/default/variadic parameters, tuples, and a macro/annotation/derive story.

- I reviewed the v10 spec and compiled gaps beyond the existing v11 TODO list; the highlights are overloading, call-site conveniences (named/default/varargs), tuples, and macro/annotation/derive support.
