# Composite interface self-pattern contract

Status: canonical for Able v12.

## Decision

A composite interface has one effective self-pattern contract. Its admitted
target set must be a subset of the target set admitted by every explicit base
interface after substituting the base reference's generic arguments.

The effective forms are:

- no `for` clause and `for Self` are implicit;
- an implicit composite may contain only implicit bases;
- an explicit composite may contain implicit bases, which adopt the
  composite's pattern for that composition; and
- an explicit base must structurally accept the composite pattern after
  generic substitution.

For example, a base declared `interface Items<T> for Array T` accepts a
composite declared `interface Strings for Array String = Items<String>`.
A base declared `interface IntegerOnly for i32` does not accept a composite
declared `interface Text for String = IntegerOnly`.

## Why target-set containment

Equality would reject useful narrowing: a general `for T` base can safely
participate in a composite restricted to `String`. Merely checking for the
presence of `for` would accept impossible relationships such as `String`
composed from an `i32`-only interface. Target-set containment admits the first
case and rejects the second with one general structural rule.

Requiring an implicit composite to remain wholly implicit prevents an omitted
clause from silently choosing among explicit restrictions. Authors must state
the composite pattern when any base brings an explicit restriction.

## Implementation boundary

The parser and canonical AST already preserve the composite pattern, base
references, and base type arguments. Validation therefore runs in the shared
program typechecker after local type declarations are fully hydrated. This
ordering also handles bases declared later in the same module.

The checker reuses implementation self-pattern matching and substitutes
explicit base arguments before comparison. Unresolved or non-interface bases
remain owned by the existing base-interface diagnostics.

No tree-walker, bytecode VM, compiler generator, runtime carrier, bridge, or
canonical stdlib special case is required. Valid programs continue through
the existing shared nominal interface lowering. Invalid relationships stop
before adapter or dictionary synthesis.

## Required coverage

- parser preservation of an explicit generic self pattern and generic bases;
- checker acceptance of compatible explicit/implicit bases and substituted
  generic patterns;
- checker rejection of an implicit composite with an explicit base;
- checker rejection of incompatible explicit patterns;
- an AST diagnostic fixture for source-attributed incompatibility; and
- one valid composite fixture in tree-walker, bytecode, parity, and strict
  fallback-free compiled execution.
