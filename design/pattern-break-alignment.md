# Break & Pattern Semantics Alignment (v10 Interpreters)

_Date: 2025-03-xx_

## Overview
Both Able v10 interpreters now share the same observable behaviour for:

1. `break` statements (optional label/value) inside loops and breakpoint expressions.
2. Pattern-driven assignments and loop bindings (identifier, wildcard, literal, array with rest, typed patterns).
3. Typed pattern failures producing a consistent error message (`"Typed pattern mismatch in assignment"`).

The TypeScript interpreter remains the behavioural reference, and the Go interpreter now mirrors these semantics. JSON fixtures under `fixtures/ast/patterns/` assert parity across runtimes.

## Key Decisions
- **Optional break fields**: The AST allows `label`/`value` to be omitted. Breaks target active breakpoint labels; loop breaks carry a value returned from the loop. Labeled loop breaks remain unsupported in this milestone.
- **Pattern assignment**: Go evaluator now respects the same pattern semantics as TypeScript: literal checks, array rest capture, and typed pattern coercion (currently a no-op pending interface support).
- **Error messaging**: Typed pattern mismatches must report the exact string expected by existing TS tests to keep parity fixtures deterministic.

## Fixtures & Tests
- Added fixtures `patterns/array_destructuring`, `patterns/for_array_pattern`, `patterns/typed_assignment`, and `patterns/typed_assignment_error` to exercise successful and failing scenarios.
- Go parity tests hydrate every fixture (`go test ./pkg/interpreter`), while `bun run scripts/run-fixtures.ts` keeps the TS interpreter honest.

## Follow-up
- Struct-pattern fixtures are blocked until Go implements struct literals/member access.
- Once the spec chapter for patterns is drafted, reference this document and embed the canonical error strings/examples.
