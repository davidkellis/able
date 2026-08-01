//go:build js && wasm

package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) contextualNumericLiteralValue(_ *ast.IntegerLiteral) (runtime.Value, bool, error) {
	return nil, false, nil
}
