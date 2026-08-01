//go:build js && wasm

package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) staticReceiverTypeForCall(_ *ast.FunctionCall, _ *runtime.Environment) ast.TypeExpression {
	return nil
}
