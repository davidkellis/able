package compiler

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func (g *generator) programRequiresSchedulerExecutionContext(modules []*driver.Module) bool {
	if g == nil {
		return false
	}
	for _, module := range modules {
		if module == nil || module.AST == nil {
			continue
		}
		found := false
		ast.Walk(module.AST, func(node ast.Node) bool {
			switch node.(type) {
			case *ast.AwaitExpression, *ast.SpawnExpression:
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
