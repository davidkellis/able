# Able v11 Parser Workspace

This directory hosts the tree-sitter grammar for the Able v11 language. The
initial scaffold lives in `tree-sitter-able/` and will evolve until it can
round-trip programs into the shared AST (`design/ast-contract.md`).

## Immediate goals
- Capture the v11 lexical surface (identifiers, literals, comments, keywords).
- Encode the top-level grammar (package/import statements, declarations,
  expression statements) without the v11 safe-navigation operator.
- Map grammar productions to canonical AST constructors so the Go/TS runtimes
  can consume parser output directly.
- Wire CI/dev scripts to build and test the grammar via the tree-sitter CLI.

## Status
- `tree-sitter init` scaffold committed.
- `grammar.js` now tracks the v11 declaration surface: spec-style generics for
  structs/unions/interfaces/methods, interface compositions (`=` + `+`), and
  `impl` headers that accept space-delimited interface arguments.
- Host interop is parsed via dedicated `prelude <target> { ... }` and
  `extern <target> fn ... { ... }` rules that treat the body as raw host code
  while keeping signatures aligned with Able syntax.
- Async/error flow now matches the spec: `proc`/`spawn` accept only function
  calls or blocks, `rescue` requires match-style clauses, `ensure` wraps any
  expression (including `rescue`), and `rethrow` stays a standalone statement.
- Type expressions cover union pipes, function arrows, nullable/result
  shorthands, wildcard placeholders, and space-delimited applications (e.g.,
  `Array string`, `Self A`), while expressions cover callable-only pipes and
  placeholder lambdas (`@`, `@n`).
- Corpus directory stubbed at `tree-sitter-able/test/corpus`; add cases once the
  grammar stabilises. Use `npm run test` (alias for `tree-sitter test`) to drive
  regression suites as fixtures land.
- Next steps: clarify any outstanding `type` alias semantics once the spec
  nails them down, plan the future safe member access operator once it is
  specced, thread placeholder handling into AST generation, populate the
  corpus, and hook the parser output into AST builders for integration smoke
  tests.
