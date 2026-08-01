//go:build js && wasm

package interpreter

import "able/interpreter-go/pkg/ast"

// The browser runtime accepts pre-parsed AST modules. It intentionally does
// not link the native loader/typechecker stack, so no inference facts are
// available for optional bytecode proof metadata.
type bytecodeInferenceFacts map[ast.Node]struct{}
type bytecodeMethodSelections map[ast.Node]struct{}

type interpreterTypechecker struct{}

type interpreterTypecheckDiagnostic struct {
	Message string
}
