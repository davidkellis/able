// Package interpreter executes Able v12 programs using the canonical Go runtime.
// It evaluates the shared AST format produced by the v12 fixture exporter (Go-based)
// and mirrors the semantics captured in spec/full_spec_v12.md. The interpreter
// stays in parity with the bytecode engine via the shared fixtures suite and the
// run_all_tests.sh harness.
package interpreter
