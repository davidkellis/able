// Package interpreter executes Able v11 programs using the canonical Go runtime.
// It evaluates the shared AST format produced by v11/interpreters/ts/scripts/export-fixtures.ts
// and mirrors the semantics captured in spec/full_spec_v11.md. The interpreter
// is kept in parity with the Bun implementation via the shared fixtures suite,
// parity CLI, and run_all_tests.sh harness.
package interpreter
