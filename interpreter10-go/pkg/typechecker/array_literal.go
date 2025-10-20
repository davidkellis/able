package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) checkArrayLiteral(env *Environment, expr *ast.ArrayLiteral) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	var (
		diags       []Diagnostic
		elementType Type = UnknownType{}
	)

	for idx, element := range expr.Elements {
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
		if !typeAssignable(elemType, elementType) && !typeAssignable(elementType, elemType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf(
					"typechecker: array element %d has type %s, expected %s",
					idx+1,
					typeName(elemType),
					typeName(elementType),
				),
				Node: element,
			})
			elementType = UnknownType{}
		}
	}

	arrayType := ArrayType{Element: elementType}
	c.infer.set(expr, arrayType)
	return diags, arrayType
}
