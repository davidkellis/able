# Able v10 AST Contract (Go)

Date: 2025‑10‑19  
Maintainer: Able Agents

## Purpose

This note captures the v10 AST shape implemented in `interpreter10-go/pkg/ast`
and highlights the metadata that downstream tooling (typechecker, parser,
compiler) can rely on. The AST is shared across interpreters; changes must be
treated as breaking for all runtimes and fixture exporters.

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
- Control flow: `IfExpression` + `OrClause`, `MatchExpression` with clause
  patterns & guards, `RescueExpression`, `EnsureExpression`, `OrElseExpression`,
  `PropagationExpression`.
- Async: `ProcExpression`, `SpawnExpression`, `BreakpointExpression`.
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
- **Async helpers** – `ProcExpression`/`SpawnExpression` nodes are unique and
  carry their body expressions directly, so schedulers and typecheckers can
  recognise async contexts.
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
