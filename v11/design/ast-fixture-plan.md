# Shared AST Fixture Plan

Goal: validate that the Go and TypeScript interpreters consume identical Able v11 AST structures and produce the same observable behaviour.

These fixtures are **language implementation tests** and are not related to the
user-facing `able test` framework or its conventions.

## 1. Fixture Format

- Use JSON as the interchange format. Every node already carries `node_type` and snake_case fields that serialize naturally.
- Store canonical fixtures under `fixtures/ast/` with descriptive names (e.g., `literals/basic.json`). Each file contains a single `Module` node.
- When a scenario requires multiple related programs (e.g., import/export), use subdirectories with `main.json`, `dep.json`, etc.

## 2. Generation

1. Author fixtures using the TypeScript AST DSL (`v11/interpreters/ts/src/ast.ts` helpers) to maintain a single source of truth.
2. Add a Bun script (`v11/interpreters/ts/scripts/export-fixtures.ts`) that constructs ASTs with the DSL and writes JSON into `fixtures/ast/` using a stable serialization order (sorted object keys, arrays preserved).
3. Commit the generated JSON alongside a manifest describing expected outcomes (see below).

## 3. Manifest Structure

Each fixture directory includes an optional `manifest.json`:

```json
{
  "description": "Short explanation",
  "entry": "main.json",
  "expect": {
    "stdout": ["Hello, world"],
    "result": "i32:5",
    "errors": []
  }
}
```

- `entry` names the `Module` to evaluate; defaults to the only file present.
- `expect` defines optional checks (stdout lines, returned value, raised errors). Implementations can ignore unsupported keys but should report unknown ones.

## 4. Interpreter Harnesses

- **TypeScript**: add a Bun test runner that reads the manifest, loads the JSON into AST objects (simple recursive builder), evaluates the program, and asserts expectations.
- **Go**: mirror the harness using Go’s testing framework. Deserialize JSON into the Go AST structs, run `InterpretModule`, and compare results with the manifest.

## 5. Coverage Targets

- Literals & operators
- Control flow (`if`, `while`, `for`, `break label`)
- Pattern matching & destructuring
- Structs/unions/interfaces/impls
- Error handling (`raise`, `rescue`, `ensure`, `expr!`)
- Concurrency (`proc`, `spawn`, cancellation, yield)
- Modules/imports (selector, wildcard, alias, privacy)
- Host interop placeholders (ensure nodes round-trip even if not executed)

## 6. Workflow

1. When adding a new language feature, create or update fixtures in the TypeScript project.
2. Regenerate fixture JSON, run both interpreters’ fixture suites, and commit the updated files.
3. Fail CI if either interpreter diverges from the manifest expectations.

This plan ensures future interpreters can reuse the same fixture set while giving immediate cross-runtime parity checks.
