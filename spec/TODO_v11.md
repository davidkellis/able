# Able v11 Spec TODOs

This list tracks language features and spec work deferred to the v11 cycle.

- [ ] **Map literals**: Specify a literal syntax (`#{ key: value, ... }`), typing/inference rules (empty literal requirements, key/value unification), evaluation order, duplicate-key behaviour, and optional spread/update forms. Update AST contract and interpreter semantics accordingly.
- [ ] **Struct functional update semantics**: v10 already supports `Struct { ...source, field: value }` for named structs (same struct type only; later fields overwrite earlier ones; positional structs disallow spreads). The spec mentions the pattern but doesn’t fully spell out semantics or generic cases. For v11, tighten the documentation and add shared fixtures/tests that cover generic updates and edge cases.
- [ ] **Type alias declarations**: Define syntax (`type Name = ...`), generic forms, visibility, and evaluation semantics (including interaction with `import`/`methods`/`impl`).
- [ ] **Safe member access (`?.`)**: Nail down grammar, typing (nullable receiver propagation), evaluation semantics, and AST/runtime updates for the safe navigation operator.
- [ ] **Optional generic type parameter declarations**: Function definitions
like `fn choose_first<T, U>(first: T, second: U) -> T where T: Display + Clone, U: Display { ... }` may omit the explicit declaration of `<T, U>` when those types are not already defined in the current scope, and the compiler should infer that since they are used as types in the function signature and within the function body, that they are generic type parameters in the scope of the function definition. Similar rules apply to struct and interface definitions.


## Candidate features from other popular languages (undecided if we should add them)

- **Operator overloading and custom operators**: No user-defined operator overloads or new operator definitions (present in Rust, Scala, Ruby).
- **Function/method overloading**: No ad‑hoc overloading by arity or parameter types; one signature per name (common in Scala/Ruby).
- **Named and default arguments**: Calls are positional-only; no default parameter values or keyword/named args (common in Scala/Ruby).
- **Variadic parameters and argument splats**: No varargs in function definitions and no splat/spread in call sites (common in Scala/Ruby).
- **Tuple types and tuple literals**: Not part of the core; you use positional structs instead (native tuples are common in Rust/Scala).
- **Immutable bindings/const and field immutability**: No `const`/`val` or per-field immutability; everything is mutable by design conventions (Rust/Scala provide this).
- **Macro/annotation/derive system**: No compile-time macros, annotations/attributes, or derive-like auto-impls (Rust/Scala have these; Ruby has rich runtime metaprogramming).
- **Comprehension syntax**: No list/set/for‑yield comprehensions; iteration is via `for`/iterators (Scala has for‑comprehensions).
- **Labeled loop control**: Labeled `break`/`continue` on loops aren’t supported; there’s a separate `breakpoint 'label` construct instead (Rust supports labeled loop control).
- **Block comments**: Only line comments are specified; block comment syntax is marked TBD (most languages support both).

Notes:
- Exceptions, pattern matching, ADTs (via `union` + singleton structs), traits/interfaces with defaults, HKTs at the interface level, iterators/generators, ranges, string interpolation, and concurrency (`proc`/`spawn`) are already in v10.
- Ownership/borrowing and lifetimes (Rust-specific) are intentionally not in Able v10 (GC-based).

- In short: the big missing “usually expected” niceties are overloading (operators/functions), named/default/variadic parameters, tuples, const/val immutability, and a macro/annotation/derive story.

- I reviewed the v10 spec and compiled gaps beyond the existing v11 TODO list; the highlights are overloading, call-site conveniences (named/default/varargs), tuples, const/val immutability, and macro/annotation/derive support.
