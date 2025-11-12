package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) checkIndexExpression(env *Environment, expr *ast.IndexExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	var diags []Diagnostic

	objDiags, objectType := c.checkExpression(env, expr.Object)
	diags = append(diags, objDiags...)

	indexDiags, indexType := c.checkExpression(env, expr.Index)
	diags = append(diags, indexDiags...)

	if indexType != nil && !isUnknownType(indexType) && !isIntegerType(indexType) {
		diags = append(diags, Diagnostic{
			Message: "typechecker: index must be an integer",
			Node:    expr.Index,
		})
	}

	switch ty := objectType.(type) {
	case ArrayType:
		elem := ty.Element
		if elem == nil {
			elem = UnknownType{}
		}
		c.infer.set(expr, elem)
		return diags, elem
	case StructInstanceType:
		if ty.StructName == "Array" && len(ty.Positional) > 0 {
			elem := ty.Positional[0]
			c.infer.set(expr, elem)
			return diags, elem
		}
	case StructType:
		if ty.StructName == "Array" && len(ty.Positional) > 0 {
			elem := ty.Positional[0]
			c.infer.set(expr, elem)
			return diags, elem
		}
	case UnknownType:
		c.infer.set(expr, UnknownType{})
		return diags, UnknownType{}
	}

	if !isUnknownType(objectType) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot index into type %s", typeName(objectType)),
			Node:    expr.Object,
		})
	}

	c.infer.set(expr, UnknownType{})
	return diags, UnknownType{}
}
