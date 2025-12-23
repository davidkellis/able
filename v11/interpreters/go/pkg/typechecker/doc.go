// Package typechecker implements the Able v11 static semantics in Go. It validates
// declarations, expressions, patterns, and generic constraints using the shared
// AST definitions, producing diagnostics that are diffed against the Bun checker
// during the fixture and parity runs. The checker can run standalone or be wired
// into the Go interpreter/CLI to fail fast on specification violations.
package typechecker
