package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (c *Checker) checkRangeExpression(env *Environment, expr *ast.RangeExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	var diags []Diagnostic

	startDiags, startType := c.checkExpression(env, expr.Start)
	diags = append(diags, startDiags...)

	endDiags, endType := c.checkExpression(env, expr.End)
	diags = append(diags, endDiags...)

	isStartInteger := isIntegerType(startType)
	isEndInteger := isIntegerType(endType)

	if startType != nil && !isUnknownType(startType) && !isStartInteger {
		diags = append(diags, Diagnostic{
			Message: "typechecker: range start must be numeric",
			Node:    expr.Start,
		})
	}
	if endType != nil && !isUnknownType(endType) && !isEndInteger {
		diags = append(diags, Diagnostic{
			Message: "typechecker: range end must be numeric",
			Node:    expr.End,
		})
	}

	elementType := Type(UnknownType{})
	if isStartInteger && isEndInteger {
		if typeAssignable(startType, endType) {
			elementType = startType
		} else if typeAssignable(endType, startType) {
			elementType = endType
		} else {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: range bounds must share a numeric type, got %s and %s", typeName(startType), typeName(endType)),
				Node:    expr,
			})
		}
	} else if isStartInteger {
		elementType = startType
	} else if isEndInteger {
		elementType = endType
	}

	var bounds []Type
	if isStartInteger && startType != nil && !isUnknownType(startType) {
		bounds = append(bounds, startType)
	}
	if isEndInteger && endType != nil && !isUnknownType(endType) {
		bounds = append(bounds, endType)
	}

	rangeType := RangeType{Element: elementType, Bounds: bounds}
	c.infer.set(expr, rangeType)
	return diags, rangeType
}
