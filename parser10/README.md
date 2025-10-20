# Able v10 Parser Workspace

This directory hosts the tree-sitter grammar for the Able v10 language. The
initial scaffold lives in `tree-sitter-able/` and will evolve until it can
round-trip programs into the shared AST (`design/ast-contract.md`).

## Immediate goals
- Capture the v10 lexical surface (identifiers, literals, comments, keywords).
- Encode the top-level grammar (package/import statements, declarations,
  expression statements) without the v11 safe-navigation operator.
- Map grammar productions to canonical AST constructors so the Go/TS runtimes
  can consume parser output directly.
- Wire CI/dev scripts to build and test the grammar via the tree-sitter CLI.

## Status
- `tree-sitter init` scaffold committed.
- `grammar.js` now includes the base structure for packages, imports, function
  and struct definitions, literal expressions, and supporting helpers.
- Next steps: flesh out the full statement/expression set, derive precedence
  tables from `spec/full_spec_v10.md`, and add fixtures + corpus tests.
