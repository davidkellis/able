# Able v10 Canonical AST Schema

This document defines the language-agnostic Abstract Syntax Tree for Able v10 programs. It is derived directly from `spec/full_spec_v10.md` and is intended to be implemented verbatim by every interpreter, compiler, or tooling project targeting Able v10.

The schema uses neutral terminology so it can map cleanly to TypeScript, Go, Crystal, or any other host language. Each implementation is responsible for preserving the structure and semantics described here.

---
## 1. Conventions

- Every node exposes a discriminant `node_type` (string/enum) that matches the names listed below.
- Collections appear in the order that source programs declare them (e.g., function parameters).
- Optional fields are noted explicitly. Absence means the semantic default described in the spec.
- Identifiers are stored as bare names; source location data is out of scope for this document but may be attached by tooling.
- Unless otherwise stated, all references to `Expression`, `Statement`, `TypeExpression`, and `Pattern` refer to the union types defined in the following sections.

---
## 2. Identifiers & Literals

| Node | Description | Fields |
|------|-------------|--------|
| `Identifier` | Lexical identifier | `name: string`
| `StringLiteral` | String literal (`"..."`, supports interpolation via `StringInterpolation`) | `value: string`
| `IntegerLiteral` | Integer literal with optional type suffix | `value: arbitrary precision integer`, `integer_type?: enum {i8,i16,i32,i64,i128,u8,u16,u32,u64,u128}` |
| `FloatLiteral` | Floating-point literal | `value: double precision`, `float_type?: enum {f32,f64}` |
| `BooleanLiteral` | `true`/`false` | `value: bool`
| `NilLiteral` | `nil` | *(no value field; always nil)*
| `CharLiteral` | Character literal | `value: string (single scalar)`
| `ArrayLiteral` | Array literal | `elements: Expression[]`

Literals participate both as expressions and, where allowed, as patterns.

---
## 3. Type Expressions & Generics

These nodes model the surface type system (§4 of the spec).

| Node | Fields |
|------|--------|
| `SimpleTypeExpression` | `name: Identifier`
| `GenericTypeExpression` | `base: TypeExpression`, `arguments: TypeExpression[]`
| `FunctionTypeExpression` | `param_types: TypeExpression[]`, `return_type: TypeExpression`
| `NullableTypeExpression` | `inner_type: TypeExpression`
| `ResultTypeExpression` | `inner_type: TypeExpression` (represents `!T` in type position)
| `UnionTypeExpression` | `members: TypeExpression[]`
| `WildcardTypeExpression` | *(no additional data)*

Generic constraints:

| Node | Fields |
|------|--------|
| `InterfaceConstraint` | `interface_type: TypeExpression`
| `GenericParameter` | `name: Identifier`, `constraints?: InterfaceConstraint[]`
| `WhereClauseConstraint` | `type_param: Identifier`, `constraints: InterfaceConstraint[]`

---
## 4. Patterns

Patterns cover destructuring, match clauses, and loop bindings (§5, §8).

| Node | Fields |
|------|--------|
| `WildcardPattern` | *(no fields)*
| `LiteralPattern` | `literal: Literal`
| `StructPatternField` | `field_name?: Identifier`, `pattern: Pattern`, `binding?: Identifier`
| `StructPattern` | `struct_type?: Identifier`, `fields: StructPatternField[]`, `is_positional: bool`
| `ArrayPattern` | `elements: Pattern[]`, `rest_pattern?: Pattern` *(identifier or wildcard)*
| `TypedPattern` | `pattern: Pattern`, `type_annotation: TypeExpression`

`Identifier` nodes can also serve as patterns and assignment targets.

---
## 5. Expressions

Core expression union:

```
Expression :=
  Identifier | Literal | UnaryExpression | BinaryExpression |
  FunctionCall | BlockExpression | AssignmentExpression |
  RangeExpression | StringInterpolation | MemberAccessExpression |
  IndexExpression | ImplicitMemberExpression | LambdaExpression |
  ProcExpression | SpawnExpression | PropagationExpression |
  OrElseExpression | BreakpointExpression | PlaceholderExpression |
  TopicReferenceExpression | IfExpression | MatchExpression |
  StructLiteral | RescueExpression | EnsureExpression
```

Node definitions:

| Node | Fields |
|------|--------|
| `UnaryExpression` | `operator: enum {- , ! , ~}`, `operand: Expression`
| `BinaryExpression` | `operator: string`, `left: Expression`, `right: Expression`
| `FunctionCall` | `callee: Expression`, `arguments: Expression[]`, `type_arguments?: TypeExpression[]`, `is_trailing_lambda: bool`
| `BlockExpression` | `body: Statement[]`
| `AssignmentExpression` | `operator: enum {:=, =, +=, -=, *=, /=, %=, &=, |=, ^=, <<=, >>=}`, `left: AssignmentTarget`, `right: Expression`
| `RangeExpression` | `start: Expression`, `end: Expression`, `inclusive: bool`
| `StringInterpolation` | `parts: (StringLiteral or Expression)[]`
| `MemberAccessExpression` | `object: Expression`, `member: Identifier | IntegerLiteral`
| `IndexExpression` | `object: Expression`, `index: Expression`
| `ImplicitMemberExpression` | `member: Identifier` *(shorthand `#member`; treated as an `Expression` and valid assignment target)*
| `LambdaExpression` | `generic_params?: GenericParameter[]`, `params: FunctionParameter[]`, `return_type?: TypeExpression`, `body: Expression | BlockExpression`, `where_clause?: WhereClauseConstraint[]`, `is_verbose_syntax: bool`
| `ProcExpression` | `expression: FunctionCall | BlockExpression`
| `SpawnExpression` | `expression: FunctionCall | BlockExpression`
| `PropagationExpression` | `expression: Expression` (postfix `!`)
| `OrElseExpression` | `expression: Expression`, `handler: BlockExpression`, `error_binding?: Identifier`
| `BreakpointExpression` | `label: Identifier`, `body: BlockExpression`
| `PlaceholderExpression` | `index?: integer` *(un-numbered `@` omits `index`; numbered `@n` stores 1-based `index`)* — detection is expression-local; placeholders nested inside explicit lambdas, iterator literals, `proc`, or `spawn` remain scoped to those constructs per §7.6.3. 
| `TopicReferenceExpression` | *(represents pipe-topic `%` while piping expressions)*

Assignment targets are any `Pattern`, `MemberAccessExpression`, `ImplicitMemberExpression`, `IndexExpression`, or `Identifier`.

---
## 6. Control Flow

| Node | Fields |
|------|--------|
| `OrClause` | `condition?: Expression`, `body: BlockExpression`
| `IfExpression` | `if_condition: Expression`, `if_body: BlockExpression`, `or_clauses: OrClause[]`
| `MatchClause` | `pattern: Pattern`, `guard?: Expression`, `body: Expression`
| `MatchExpression` | `subject: Expression`, `clauses: MatchClause[]`
| `WhileLoop` | `condition: Expression`, `body: BlockExpression`
| `ForLoop` | `pattern: Pattern`, `iterable: Expression`, `body: BlockExpression`
| `BreakStatement` | `label: Identifier`, `value: Expression`

---
## 7. Error Handling

| Node | Fields |
|------|--------|
| `RaiseStatement` | `expression: Expression`
| `RescueExpression` | `monitored_expression: Expression`, `clauses: MatchClause[]`
| `EnsureExpression` | `try_expression: Expression`, `ensure_block: BlockExpression`
| `RethrowStatement` | *(no additional fields)*

`PropagationExpression` (`expr!`) is listed within expressions above.

---
## 8. Declarations & Definitions

### Structs & Struct Literals

| Node | Fields |
|------|--------|
| `StructFieldDefinition` | `name?: Identifier`, `field_type: TypeExpression`
| `StructDefinition` | `id: Identifier`, `generic_params?: GenericParameter[]`, `fields: StructFieldDefinition[]`, `where_clause?: WhereClauseConstraint[]`, `kind: enum {singleton, named, positional}`, `is_private?: bool`
| `StructFieldInitializer` | `name?: Identifier`, `value: Expression`, `is_shorthand: bool`
| `StructLiteral` | `struct_type?: Identifier`, `fields: StructFieldInitializer[]`, `is_positional: bool`, `functional_update_source?: Expression`, `type_arguments?: TypeExpression[]`

### Unions

| Node | Fields |
|------|--------|
| `UnionDefinition` | `id: Identifier`, `generic_params?: GenericParameter[]`, `variants: TypeExpression[]`, `where_clause?: WhereClauseConstraint[]`, `is_private?: bool`

### Functions & Methods

| Node | Fields |
|------|--------|
| `FunctionParameter` | `name: Pattern`, `param_type?: TypeExpression`
| `FunctionDefinition` | `id: Identifier`, `generic_params?: GenericParameter[]`, `params: FunctionParameter[]`, `return_type?: TypeExpression`, `body: BlockExpression`, `where_clause?: WhereClauseConstraint[]`, `is_method_shorthand: bool`, `is_private: bool`
| `FunctionSignature` | `name: Identifier`, `generic_params?: GenericParameter[]`, `params: FunctionParameter[]`, `return_type?: TypeExpression`, `where_clause?: WhereClauseConstraint[]`, `default_impl?: BlockExpression`

### Interfaces & Implementations

| Node | Fields |
|------|--------|
| `InterfaceDefinition` | `id: Identifier`, `generic_params?: GenericParameter[]`, `self_type_pattern?: TypeExpression`, `signatures: FunctionSignature[]`, `where_clause?: WhereClauseConstraint[]`, `base_interfaces?: TypeExpression[]`, `is_private: bool`
| `ImplementationDefinition` | `impl_name?: Identifier`, `generic_params?: GenericParameter[]`, `interface_name: Identifier`, `interface_args?: TypeExpression[]`, `target_type: TypeExpression`, `definitions: FunctionDefinition[]`, `where_clause?: WhereClauseConstraint[]`, `is_private?: bool`
| `MethodsDefinition` | `target_type: TypeExpression`, `generic_params?: GenericParameter[]`, `definitions: FunctionDefinition[]`, `where_clause?: WhereClauseConstraint[]`

---
## 9. Modules, Imports & Host Interop

| Node | Fields |
|------|--------|
| `PackageStatement` | `name_path: Identifier[]`, `is_private?: bool`
| `ImportSelector` | `name: Identifier`, `alias?: Identifier`
| `ImportStatement` | `package_path: Identifier[]`, `is_wildcard: bool`, `selectors?: ImportSelector[]`, `alias?: Identifier`
| `DynImportStatement` | `package_path: Identifier[]`, `is_wildcard: bool`, `selectors?: ImportSelector[]`, `alias?: Identifier`
| `Module` | `package?: PackageStatement`, `imports: ImportStatement[]`, `body: Statement[]`

Host interop constructs (§14) are surfaced via:

| Node | Fields |
|------|--------|
| `PreludeStatement` | `target: enum {go, crystal, typescript, python, ruby}`, `code: string`
| `ExternFunctionBody` | `target: HostTarget`, `signature: FunctionDefinition`, `body: string`

---
## 10. Statement Union

```
Statement :=
  Expression | FunctionDefinition | StructDefinition | UnionDefinition |
  InterfaceDefinition | ImplementationDefinition | MethodsDefinition |
  ImportStatement | PackageStatement | ReturnStatement | RaiseStatement |
  RethrowStatement | BreakStatement | WhileLoop | ForLoop |
  PreludeStatement | ExternFunctionBody | DynImportStatement
```

`ReturnStatement` carries an optional `argument: Expression`.

---
## 11. Alignment with v10 Spec

- **Core expressions** map to the evaluation rules in §6 (Expressions) and §7 (Operators).
- **Control flow** covers `if/or`, loops, labeled breaks (§8).
- **Pattern system** matches destructuring/`match` semantics (§5, §9).
- **Error handling** implements `raise`, `rescue`, `ensure`, `expr!`, `rethrow` (§10).
- **Interface & implementation** nodes encode the structure from §11 (Interfaces & Methods).
- **Concurrency** is represented by `ProcExpression`, `SpawnExpression`, and associated runtime expectations (§12). The AST carries no scheduler logic; implementations must provide the behaviors defined in the spec.
- **Modules/imports** follow §13 (Packages & Imports). Dynamic imports are kept explicit to support `dynimport` semantics.
- **Host interop** nodes are placeholders for §14; interpreters may ignore code bodies on unsupported targets but must preserve the structures for tooling.

---
## 12. Implementation Guidance

1. **Node Names**: Implementations should mirror `node_type` string values exactly. Enumerations may become typed constants but must remain stable across languages.
2. **Optional Fields**: Omit when semantically absent (e.g., `error_binding` on `OrElseExpression` when not supplied).
3. **Pattern-As-Target**: Any `Pattern` can appear on the left side of declarations (`:=`) and assignments (`=`) consistent with the spec’s destructuring rules.
4. **Functional Update**: `StructLiteral.functional_update_source` holds the expression whose fields are copied when using `{ ..existing }` syntax.
5. **Type Arguments**: Both `FunctionCall` and `StructLiteral` attach `type_arguments` to support explicit generics.
6. **Trailing Lambdas**: `FunctionCall.is_trailing_lambda` signals syntax where the final argument is provided using trailing block syntax; interpreters may treat this as sugar but the flag ensures round-tripping.

The TypeScript and Go interpreters must adapt their internal representations to this schema. Any divergence should be documented and resolved as a priority.
