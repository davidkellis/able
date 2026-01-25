// Package typechecker implements the Able v12 static semantics in Go. It validates
// declarations, expressions, patterns, and generic constraints using the shared
// AST definitions, producing diagnostics that are diffed across tree-walker and
// bytecode runs. The checker can run standalone or be wired
// into the Go interpreter/CLI to fail fast on specification violations.
package typechecker
