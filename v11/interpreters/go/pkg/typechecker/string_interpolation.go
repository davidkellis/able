package typechecker

import "able/interpreter10-go/pkg/ast"

func (c *Checker) checkStringInterpolation(env *Environment, expr *ast.StringInterpolation) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, PrimitiveType{Kind: PrimitiveString}
	}

	var diags []Diagnostic
	resultType := PrimitiveType{Kind: PrimitiveString}

	for _, part := range expr.Parts {
		if part == nil {
			continue
		}
		partDiags, _ := c.checkExpression(env, part)
		diags = append(diags, partDiags...)
	}

	c.infer.set(expr, resultType)
	return diags, resultType
}
