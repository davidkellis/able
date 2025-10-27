# Able v11 Spec TODOs

This list tracks language features and spec work deferred to the v11 cycle.

- [ ] **Map literals**: Specify a literal syntax (`#{ key: value, ... }`), typing/inference rules (empty literal requirements, key/value unification), evaluation order, duplicate-key behaviour, and optional spread/update forms. Update AST contract and interpreter semantics accordingly.
- [ ] **Struct functional update semantics**: v10 already supports `Struct { ...source, field: value }` for named structs (same struct type only; later fields overwrite earlier ones; positional structs disallow spreads). The spec mentions the pattern but doesn’t fully spell out semantics or generic cases. For v11, tighten the documentation and add shared fixtures/tests that cover generic updates and edge cases.
- [ ] **Type alias declarations**: Define syntax (`type Name = ...`), generic forms, visibility, and evaluation semantics (including interaction with `import`/`methods`/`impl`).
- [ ] **Safe member access (`?.`)**: Nail down grammar, typing (nullable receiver propagation), evaluation semantics, and AST/runtime updates for the safe navigation operator.
