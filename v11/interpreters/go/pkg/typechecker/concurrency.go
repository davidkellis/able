package typechecker

import "able/interpreter-go/pkg/ast"

func (c *Checker) checkProcExpression(env *Environment, expr *ast.ProcExpression) ([]Diagnostic, Type) {
	c.pushAsyncContext()
	bodyDiags, bodyType := c.checkExpression(env, expr.Expression)
	c.popAsyncContext()
	result := ProcType{Result: bodyType}
	c.infer.set(expr, result)
	return bodyDiags, result
}

func (c *Checker) checkSpawnExpression(env *Environment, expr *ast.SpawnExpression) ([]Diagnostic, Type) {
	c.pushAsyncContext()
	bodyDiags, bodyType := c.checkExpression(env, expr.Expression)
	c.popAsyncContext()
	result := FutureType{Result: bodyType}
	c.infer.set(expr, result)
	return bodyDiags, result
}
