# Able v11 Spec TODOs

This list tracks language features and spec work deferred to the v11 cycle.

- [ ] **Map literals**: Specify a literal syntax (`#{ key: value, ... }`), typing/inference rules (empty literal requirements, key/value unification), evaluation order, duplicate-key behaviour, and optional spread/update forms. Update AST contract and interpreter semantics accordingly.
- [ ] **Struct functional update semantics**: v10 already supports `Struct { ...source, field: value }` for named structs (same struct type only; later fields overwrite earlier ones; positional structs disallow spreads). The spec mentions the pattern but doesn’t fully spell out semantics or generic cases. For v11, tighten the documentation and add shared fixtures/tests that cover generic updates and edge cases.
- [ ] **Type alias declarations**: Define syntax (`type Name = ...`), generic forms, visibility, and evaluation semantics (including interaction with `import`/`methods`/`impl`).
- [ ] **Safe member access (`?.`)**: Nail down grammar, typing (nullable receiver propagation), evaluation semantics, and AST/runtime updates for the safe navigation operator.
- [ ] **Typed `=` declarations**: Decide whether `name: Type = expr` should be first-class declaration sugar (without needing `:=`) as hinted in §5.2.7, and update the spec + interpreters/typecheckers if we bless it. Capture scope/shadowing rules and diagnostics for mixed declaration/reassignment cases. Perhaps we just use = or := for declarations and = for reassignment, with the understandign that := always introduces a new binding and = always reassigns an existing binding; I think this is the best way to go.
- [ ] **Contextual integer widening & literal typing**: Fully specify the numeric literal inference/promotion rules promised in §6.3.2 (P1), including how integer literals adopt the surrounding type, how mixed-width arithmetic widens operands, and what diagnostics fire when narrowing would occur. Ensure both interpreters and fixtures enforce the v11 rules consistently. We want to be able to assign an integer literal to a variable of a wider width than necessary to represent the value, and we want to be able to assign a narrower width integer to a variable of a wider width.
- [ ] **Optional generic type parameter declarations**: Function definitions
like `fn choose_first<T, U>(first: T, second: U) -> T where T: Display + Clone, U: Display { ... }` may omit the explicit declaration of `<T, U>` when those types are not already defined in the current scope, and the compiler should infer that since they are used as types in the function signature and within the function body, that they are generic type parameters in the scope of the function definition. Similar rules apply to struct and interface definitions.
- [ ] **Channel `select` syntax & semantics**: v10 punts on multiplexing (see §12); the spec needs to define the `select { case send/receive ... }` grammar, evaluation order, blocking/default clauses, tie-breaking fairness, and how helpers like `proc_cancelled`/`proc_flush` integrate. Both interpreters plus the shared fixtures/parity suite must enforce the same behavior (buffered vs. unbuffered channels, nil-channel handling, cancellation propagation).
- [ ] Specify how stdlib channel/mutex helpers surface error structs (`ChannelClosed`, `ChannelNil`, `ChannelSendOnClosed`) so host runtimes map native failures to the documented Able `Error` types.
- [ ] Specify the `able.text.regex` module, including default code-point semantics, optional grapheme-aware execution, result types, and error reporting.
- [ ] Document the standard library `String`, `Grapheme`, and iteration helpers (byte, char, grapheme views) together with the byte-oriented indexing rules referenced by the spec.


## Candidate features from other popular languages (undecided if we should add them)

- **Operator overloading and custom operators**: No user-defined operator overloads or new operator definitions (present in Rust, Scala, Ruby).
- **Function/method overloading**: No ad‑hoc overloading by arity or parameter types; one signature per name (common in Scala/Ruby).
- **Named and default arguments**: Calls are positional-only; no default parameter values or keyword/named args (common in Scala/Ruby).
- **Variadic parameters and argument splats**: No varargs in function definitions and no splat/spread in call sites (common in Scala/Ruby).
- **Tuple types and tuple literals**: Not part of the core; you use positional structs instead (native tuples are common in Rust/Scala).
- **Macro/annotation/derive system**: No compile-time macros, annotations/attributes, or derive-like auto-impls (Rust/Scala have these; Ruby has rich runtime metaprogramming).
- **Comprehension syntax**: No list/set/for‑yield comprehensions; iteration is via `for`/iterators (Scala has for‑comprehensions).

Notes:
- In short: the big missing “usually expected” niceties are overloading (operators/functions), named/default/variadic parameters, tuples, const/val immutability, and a macro/annotation/derive story.

- I reviewed the v10 spec and compiled gaps beyond the existing v11 TODO list; the highlights are overloading, call-site conveniences (named/default/varargs), tuples, and macro/annotation/derive support.
