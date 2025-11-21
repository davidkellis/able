package typechecker

import "able/interpreter10-go/pkg/ast"

func (c *Checker) checkArrayLiteral(env *Environment, expr *ast.ArrayLiteral) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	var (
		diags       []Diagnostic
		elementType Type = UnknownType{}
	)

	for _, element := range expr.Elements {
		if element == nil {
			continue
		}
		elemDiags, elemType := c.checkExpression(env, element)
		diags = append(diags, elemDiags...)

		if elemType == nil || isUnknownType(elemType) {
			continue
		}
		if isUnknownType(elementType) {
			elementType = elemType
			continue
		}
		elementType = mergeTypesAllowUnion(elementType, elemType)
	}

	arrayType := ArrayType{Element: elementType}
	c.infer.set(expr, arrayType)
	return diags, arrayType
}
