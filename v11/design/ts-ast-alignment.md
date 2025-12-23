# TypeScript AST Alignment Notes (Able v11)

This document compares the canonical schema (see `design/ast-schema-v11.md`) with the existing TypeScript AST implementation in `v11/interpreters/ts/src/ast.ts`.

## Summary

- The TypeScript AST already matches the canonical node set and field structure defined in the schema. No structural changes are required at this time.
- Naming conventions (`type` discriminant strings, field names) align with the schema, so downstream tooling can treat the TypeScript definitions as a faithful implementation of the canonical AST.

## Minor Observations

- `ArrayPattern.restPattern` in TypeScript is constrained to `Identifier | WildcardPattern`. The schema allows any `Pattern` but documents that only identifiers or wildcards are semantically valid; no code change necessary.
- `FunctionCall.isTrailingLambda` defaults to `false` in TypeScript; the schema records the flag but does not prescribe a default. Implementations should maintain the same behaviour.
- `ProcExpression` / `SpawnExpression` accept `FunctionCall | BlockExpression`, matching the schema’s narrow definition. Ensure future editors keep this restriction rather than widening to arbitrary expressions.

## Action Items

None required now. Re-run this comparison whenever the spec, schema, or TypeScript implementation evolves.
