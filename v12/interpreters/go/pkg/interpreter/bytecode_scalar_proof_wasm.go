//go:build js && wasm

package interpreter

import "able/interpreter-go/pkg/ast"

func bytecodeInferenceSimpleCheck(_ bytecodeInferenceFacts, _ ast.Node) bytecodeSimpleTypeCheck {
	return bytecodeSimpleTypeCheckUnknown
}
