package compiler

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func (g *generator) programNeedsCallableExecutionContext(modules []*driver.Module) bool {
	if g == nil || !g.executionContextsEnabled() {
		return false
	}
	for _, module := range modules {
		if module == nil || module.AST == nil || module.Package != g.entryPackage {
			continue
		}
		found := false
		ast.Walk(module.AST, func(node ast.Node) bool {
			if _, ok := node.(*ast.AwaitExpression); ok {
				found = true
				return false
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}
