package interpreter

import "able/interpreter-go/pkg/ast"

func (vm *bytecodeVM) canonicalRuntimeTypeExpression(expr ast.TypeExpression) ast.TypeExpression {
	if vm == nil || vm.interp == nil || expr == nil {
		return expr
	}
	return vm.interp.canonicalizeTypeExpressionCached(expr, vm.env, vm.interp.typeExpressionReferencesAliasCached(expr))
}
