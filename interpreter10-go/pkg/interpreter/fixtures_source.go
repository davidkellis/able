package interpreter

import (
	"fmt"
	"os"

	"able/interpreter10-go/pkg/ast"
	goParser "able/interpreter10-go/pkg/parser"
)

func parseSourceModule(path string) (*ast.Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read source %s: %w", path, err)
	}
	parser, err := goParser.NewModuleParser()
	if err != nil {
		return nil, fmt.Errorf("module parser init: %w", err)
	}
	defer parser.Close()
	module, err := parser.ParseModule(data)
	if err != nil {
		return nil, fmt.Errorf("parse module %s: %w", path, err)
	}
	return module, nil
}
