# Able v11 AST Contract (Go)

Date: 2025‑10‑19  
Maintainer: Able Agents

## Purpose

This note captures the v11 AST shape implemented in `interpreter-go/pkg/ast`
and highlights the metadata that downstream tooling (typechecker, parser,
compiler) can rely on. The AST is shared across interpreters; changes must be
treated as breaking for all runtimes and fixture exporters.

## Testing separation requirements

To keep semantics well-defined and detect regressions early, the language
project maintains three independent test layers around the AST contract:

1. **AST evaluation in isolation** — Interpreter suites construct `ast.*`
   nodes directly (Go: `pkg/interpreter/*_test.go`, TS: `test/{basics,control_flow,…}`)
   and execute them without invoking parsers. These guarantee that evaluation
   honours the AST contract independently of any surface syntax.
2. **Parser (CST) conformance** — The tree-sitter grammar corpus (`parser10/tree-sitter-able/test/corpus`
   and Go `pkg/parser/parser_test.go`) exercises pure parsing, asserting that
   Able source text yields the expected concrete syntax trees.
3. **Parser→AST mapping** — Mapper tests (Go: `pkg/parser/parser_test.go`,
   TS: `test/parser/fixtures_mapper.test.ts`) convert parse trees into AST
   nodes and compare them against the canonical `module.json` fixtures.

These layers must remain independent: AST semantics do not rely on parser
output, and parser regressions cannot silently skew interpreter behaviour.
End-to-end fixture runs (`pkg/interpreter/fixtures_parity_test.go`,
`v11/interpreters/ts/scripts/run-fixtures.ts`) sit on top of the three layers to
verify the full pipeline but do not replace the isolated suites.
Fixture manifests may include setup modules (for example, `package.json`) that
now point at dedicated `<name>.able` sources; loader helpers fall back to
`source.able` only for the primary module, keeping multi-module fixtures in
sync across interpreters.

## Top-level module structure

| Node                     | Notes |
|--------------------------|-------|
| `Module`                 | Owns optional `PackageStatement`, `PreludeStatement`, `Imports`, `DynImport` statements, `Body []Statement`. |
| `PackageStatement`       | `NamePath []*Identifier`, `IsPrivate`. |
| `ImportStatement`        | `Selectors []ImportSelector`, destination env is resolved by interpreter. |
| `DynImportStatement`     | Captures dynamic import target and alias. |

## Statements & declarations

- `FunctionDefinition` – carries `GenericParams`, `Params` (pattern + type),
  optional `ReturnType`, and optional `WhereClause`.
- `StructDefinition`, `UnionDefinition`, `InterfaceDefinition`,
  `ImplementationDefinition`, `MethodsDefinition` – all include optional
  generics, where clauses, and privacy flags surfaced in the spec.
- `ReturnStatement`, `RaiseStatement`, `BreakStatement`, `ContinueStatement`,
  `WhileLoop`, `ForLoop` – match the evaluation semantics required by the spec.
- `AssignmentExpression` doubles as a statement; `Operator` enumerates `:=`
  declaration vs. mutation and compound assignments.

Patterns (`Pattern` interface) cover wildcard, literal, struct, array, typed,
and identifier patterns. Function parameters and destructuring reuse the same
structures.

## Expressions

Key nodes and metadata:

- Literals: carry raw values plus numeric suffix (`IntegerType`, `FloatType`).
- `FunctionCall`: supports positional args, optional type arguments, and UFCS via
  member access.
- Control flow: `IfExpression` with `ElseIfClause` + optional else body, `MatchExpression` with clause
  patterns & guards, `RescueExpression`, `EnsureExpression`, `OrElseExpression`,
  `PropagationExpression`.
- Async: `SpawnExpression`, `BreakpointExpression`.
- Structs: `StructLiteral` with named/positional initialisers, optional type
  arguments, and functional update base expression.

## Type expressions

Every syntax form surfaced in the spec is present as its own node:

- `SimpleTypeExpression`, `GenericTypeExpression`, `FunctionTypeExpression`,
  `NullableTypeExpression`, `ResultTypeExpression`, `UnionTypeExpression`,
  `WildcardTypeExpression`.
- Constraints: `GenericParameter` + associated `InterfaceConstraint`, and
  `WhereClauseConstraint` allow the typechecker to enforce bounds without AST
  mutation.

## Metadata relied upon by tooling

- **Generics** – All declaration nodes expose `GenericParams` and per-clause
  `WhereClause`.
- **Privacy** – `IsPrivate` flags exist on package, struct, union, interface,
  impl, and method containers.
- **Type annotations** – Present on parameters, struct fields, function return
  types, and typed patterns. Local bindings rely on `AssignmentDeclare` combined
  with optional typed patterns.
- **Async helpers** – `SpawnExpression` nodes are unique and carry their body
  expressions directly, so schedulers and typecheckers can recognise async
  contexts.
- **Implementations** – `ImplementationDefinition` includes interface name,
  generic arguments, target type, and body definitions; `MethodsDefinition`
  covers inherent impl blocks.

## Gaps / open questions

1. **Source locations** – Nodes currently lack span/position information, which
   limits diagnostic quality for the future typechecker and parser error
   reporting. If we decide to add locations, prefer an optional `Span` struct
   on `nodeImpl` so the contract stays uniform across languages.
2. **Doc comments / attributes** – Not yet represented. If the spec formalises
   attributes, we will need dedicated fields before parser work.
3. **Mutable metadata** – Downstream tools should attach inferred types via
   side tables rather than mutating the AST; this keeps fixtures serialisable.
4. **Pattern exhaustiveness annotations** – The AST does not currently carry
   compiler hints; any future data should live in auxiliary maps.

## Next steps

1. Keep this document updated whenever a node gains new fields or semantics.
2. Audit spec sections for soon-to-land features (e.g., traits or modules) and
   confirm the AST already models the necessary constructs.
3. Before parser or typechecker milestones, run a cross-check between this
   contract and the JSON fixtures under `fixtures/ast` to ensure serialisation
   remains lossless.
