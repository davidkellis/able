# Able v10 Go Interpreter

This package hosts the Go reference interpreter for the Able v10 language. The Go implementation is the canonical runtime that must match `spec/full_spec_v10.md` exactly and stay in lockstep with other interpreters on the shared AST and semantics.

Current focus:
- Define the canonical AST in Go (`pkg/ast`).
- Mirror the TypeScript AST helpers so shared fixtures remain compatible.
- Build the evaluator and runtime using Go-native concurrency primitives once the AST is complete.
- Keep the JSON fixtures under `fixtures/ast` green by running `go test ./pkg/interpreter` after any change to `interpreter10/scripts/export-fixtures.ts` (the tests hydrate every fixture to ensure parity with the TypeScript interpreter).

Development checklist and milestones live in the workspace root `PLAN.md` until this package grows its own detailed plan.
