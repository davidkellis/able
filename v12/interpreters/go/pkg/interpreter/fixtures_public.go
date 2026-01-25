package interpreter

import (
	"encoding/json"
	"fmt"

	"able/interpreter-go/pkg/ast"
)

// DecodeModule constructs an AST module from the fixture JSON payload using the shared decoder.
func DecodeModule(data []byte) (*ast.Module, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decode module json: %w", err)
	}
	node, err := decodeNode(raw)
	if err != nil {
		return nil, err
	}
	mod, ok := node.(*ast.Module)
	if !ok {
		return nil, fmt.Errorf("decoded node is %T, want *ast.Module", node)
	}
	return mod, nil
}
