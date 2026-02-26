//go:build js && wasm

package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func parseSourceModule(path string) (*ast.Module, error) {
	return nil, fmt.Errorf("parse source %s: unavailable on js/wasm runtime", path)
}
