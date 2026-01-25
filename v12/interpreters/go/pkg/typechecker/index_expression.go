package typechecker

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (c *Checker) checkIndexExpression(env *Environment, expr *ast.IndexExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	var diags []Diagnostic
	wrapResult := func(inner Type) Type {
		if inner == nil {
			inner = UnknownType{}
		}
		return AppliedType{
			Base:      StructType{StructName: "Result"},
			Arguments: []Type{inner},
		}
	}

	objDiags, objectType := c.checkExpression(env, expr.Object)
	diags = append(diags, objDiags...)

	indexDiags, indexType := c.checkExpression(env, expr.Index)
	diags = append(diags, indexDiags...)

	switch ty := objectType.(type) {
	case ArrayType:
		elem := ty.Element
		if elem == nil {
			elem = UnknownType{}
		}
		if indexType != nil && !isUnknownType(indexType) && !isIntegerType(indexType) {
			diags = append(diags, Diagnostic{
				Message: "typechecker: index must be an integer",
				Node:    expr.Index,
			})
		}
		result := wrapResult(elem)
		c.infer.set(expr, result)
		return diags, result
	case MapType:
		if indexType != nil && ty.Key != nil && !isUnknownType(indexType) && !typeAssignable(indexType, ty.Key) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: index expects type %s, got %s", typeName(ty.Key), typeName(indexType)),
				Node:    expr.Index,
			})
		}
		val := ty.Value
		if val == nil {
			val = UnknownType{}
		}
		result := wrapResult(val)
		c.infer.set(expr, result)
		return diags, result
	case StructInstanceType:
		if ty.StructName == "Array" && len(ty.Positional) > 0 {
			elem := ty.Positional[0]
			if indexType != nil && !isUnknownType(indexType) && !isIntegerType(indexType) {
				diags = append(diags, Diagnostic{
					Message: "typechecker: index must be an integer",
					Node:    expr.Index,
				})
			}
			result := wrapResult(elem)
			c.infer.set(expr, result)
			return diags, result
		}
		if ty.StructName == "HashMap" && len(ty.TypeArgs) >= 2 {
			val := ty.TypeArgs[1]
			result := wrapResult(val)
			c.infer.set(expr, result)
			return diags, result
		}
	case StructType:
		if ty.StructName == "Array" && len(ty.Positional) > 0 {
			elem := ty.Positional[0]
			if indexType != nil && !isUnknownType(indexType) && !isIntegerType(indexType) {
				diags = append(diags, Diagnostic{
					Message: "typechecker: index must be an integer",
					Node:    expr.Index,
				})
			}
			result := wrapResult(elem)
			c.infer.set(expr, result)
			return diags, result
		}
		if ty.StructName == "HashMap" && len(ty.Positional) >= 2 {
			val := ty.Positional[1]
			result := wrapResult(val)
			c.infer.set(expr, result)
			return diags, result
		}
	case AppliedType:
		if elem, ok := arrayElementType(ty); ok {
			if indexType != nil && !isUnknownType(indexType) && !isIntegerType(indexType) {
				diags = append(diags, Diagnostic{
					Message: "typechecker: index must be an integer",
					Node:    expr.Index,
				})
			}
			if elem == nil {
				elem = UnknownType{}
			}
			result := wrapResult(elem)
			c.infer.set(expr, result)
			return diags, result
		}
		if name, ok := structName(ty.Base); ok && name == "HashMap" && len(ty.Arguments) >= 2 {
			val := ty.Arguments[1]
			result := wrapResult(val)
			c.infer.set(expr, result)
			return diags, result
		}
		if iface, ok := ty.Base.(InterfaceType); ok && iface.InterfaceName == "Index" {
			var keyType, valueType Type = UnknownType{}, UnknownType{}
			if len(ty.Arguments) > 0 {
				keyType = ty.Arguments[0]
			}
			if len(ty.Arguments) > 1 {
				valueType = ty.Arguments[1]
			}
			if keyType != nil && !isUnknownType(keyType) && indexType != nil && !isUnknownType(indexType) && !typeAssignable(indexType, keyType) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: index expects type %s, got %s", typeName(keyType), typeName(indexType)),
					Node:    expr.Index,
				})
			}
			result := wrapResult(valueType)
			c.infer.set(expr, result)
			return diags, result
		}
	case UnknownType:
		result := wrapResult(UnknownType{})
		c.infer.set(expr, result)
		return diags, result
	}

	if ok, _ := c.typeImplementsInterface(objectType, InterfaceType{InterfaceName: "Index"}, nil); ok {
		result := wrapResult(UnknownType{})
		c.infer.set(expr, result)
		return diags, result
	}
	if ok, _ := c.typeImplementsInterface(objectType, InterfaceType{InterfaceName: "IndexMut"}, nil); ok {
		result := wrapResult(UnknownType{})
		c.infer.set(expr, result)
		return diags, result
	}

	if !isUnknownType(objectType) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot index into type %s", typeName(objectType)),
			Node:    expr.Object,
		})
	}

	result := wrapResult(UnknownType{})
	c.infer.set(expr, result)
	return diags, result
}

func (c *Checker) checkIndexAssignment(env *Environment, expr *ast.IndexExpression, valueType Type, op ast.AssignmentOperator) []Diagnostic {
	if expr == nil {
		return nil
	}
	var diags []Diagnostic
	objDiags, objectType := c.checkExpression(env, expr.Object)
	diags = append(diags, objDiags...)

	indexDiags, indexType := c.checkExpression(env, expr.Index)
	diags = append(diags, indexDiags...)

	if op == ast.AssignmentDeclare {
		diags = append(diags, Diagnostic{
			Message: "typechecker: cannot use := on index assignment",
			Node:    expr,
		})
	}

	requireIntegerIndex := func() {
		if indexType != nil && !isUnknownType(indexType) && !isIntegerType(indexType) {
			diags = append(diags, Diagnostic{
				Message: "typechecker: index must be an integer",
				Node:    expr.Index,
			})
		}
	}
	checkValueAssignable := func(expected Type) {
		if expected == nil {
			expected = UnknownType{}
		}
		if valueType != nil && expected != nil && !isUnknownType(valueType) && !isUnknownType(expected) && !typeAssignable(valueType, expected) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: index assignment expects value type %s, got %s", typeName(expected), typeName(valueType)),
				Node:    expr,
			})
		}
	}

	switch ty := objectType.(type) {
	case ArrayType:
		requireIntegerIndex()
		elem := ty.Element
		checkValueAssignable(elem)
		return diags
	case MapType:
		if indexType != nil && ty.Key != nil && !isUnknownType(indexType) && !typeAssignable(indexType, ty.Key) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: index expects type %s, got %s", typeName(ty.Key), typeName(indexType)),
				Node:    expr.Index,
			})
		}
		checkValueAssignable(ty.Value)
		return diags
	case StructInstanceType:
		if ty.StructName == "Array" && len(ty.Positional) > 0 {
			requireIntegerIndex()
			checkValueAssignable(ty.Positional[0])
			return diags
		}
		if ty.StructName == "HashMap" && len(ty.TypeArgs) >= 2 {
			keyType := ty.TypeArgs[0]
			valType := ty.TypeArgs[1]
			if keyType != nil && indexType != nil && !isUnknownType(keyType) && !isUnknownType(indexType) && !typeAssignable(indexType, keyType) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: index expects type %s, got %s", typeName(keyType), typeName(indexType)),
					Node:    expr.Index,
				})
			}
			checkValueAssignable(valType)
			return diags
		}
	case StructType:
		if ty.StructName == "Array" && len(ty.Positional) > 0 {
			requireIntegerIndex()
			checkValueAssignable(ty.Positional[0])
			return diags
		}
		if ty.StructName == "HashMap" && len(ty.Positional) >= 2 {
			keyType := ty.Positional[0]
			valType := ty.Positional[1]
			if keyType != nil && indexType != nil && !isUnknownType(keyType) && !isUnknownType(indexType) && !typeAssignable(indexType, keyType) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: index expects type %s, got %s", typeName(keyType), typeName(indexType)),
					Node:    expr.Index,
				})
			}
			checkValueAssignable(valType)
			return diags
		}
	case AppliedType:
		if elem, ok := arrayElementType(ty); ok {
			requireIntegerIndex()
			checkValueAssignable(elem)
			return diags
		}
		if name, ok := structName(ty.Base); ok && name == "HashMap" && len(ty.Arguments) >= 2 {
			keyType := ty.Arguments[0]
			valType := ty.Arguments[1]
			if keyType != nil && indexType != nil && !isUnknownType(keyType) && !isUnknownType(indexType) && !typeAssignable(indexType, keyType) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: index expects type %s, got %s", typeName(keyType), typeName(indexType)),
					Node:    expr.Index,
				})
			}
			checkValueAssignable(valType)
			return diags
		}
		if iface, ok := ty.Base.(InterfaceType); ok && iface.InterfaceName == "IndexMut" {
			var keyType, valType Type = UnknownType{}, UnknownType{}
			if len(ty.Arguments) > 0 {
				keyType = ty.Arguments[0]
			}
			if len(ty.Arguments) > 1 {
				valType = ty.Arguments[1]
			}
			if keyType != nil && indexType != nil && !isUnknownType(keyType) && !isUnknownType(indexType) && !typeAssignable(indexType, keyType) {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: index expects type %s, got %s", typeName(keyType), typeName(indexType)),
					Node:    expr.Index,
				})
			}
			checkValueAssignable(valType)
			return diags
		}
		if iface, ok := ty.Base.(InterfaceType); ok && iface.InterfaceName == "Index" {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: cannot assign via [] without IndexMut implementation on type %s", typeName(objectType)),
				Node:    expr,
			})
			return diags
		}
	case InterfaceType:
		if ty.InterfaceName == "IndexMut" {
			checkValueAssignable(UnknownType{})
			return diags
		}
		if ty.InterfaceName == "Index" {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: cannot assign via [] without IndexMut implementation on type %s", typeName(objectType)),
				Node:    expr,
			})
			return diags
		}
	case UnknownType:
		return diags
	}

	if ok, _ := c.typeImplementsInterface(objectType, InterfaceType{InterfaceName: "IndexMut"}, nil); ok {
		return diags
	}
	if ok, _ := c.typeImplementsInterface(objectType, InterfaceType{InterfaceName: "Index"}, nil); ok {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot assign via [] without IndexMut implementation on type %s", typeName(objectType)),
			Node:    expr,
		})
		return diags
	}
	if iface, ok := objectType.(InterfaceType); ok && iface.InterfaceName == "Index" {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot assign via [] without IndexMut implementation on type %s", typeName(objectType)),
			Node:    expr,
		})
		return diags
	}
	if applied, ok := objectType.(AppliedType); ok {
		switch base := applied.Base.(type) {
		case InterfaceType:
			if base.InterfaceName == "Index" {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: cannot assign via [] without IndexMut implementation on type %s", typeName(objectType)),
					Node:    expr,
				})
				return diags
			}
		case StructType:
			if base.StructName == "Index" {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: cannot assign via [] without IndexMut implementation on type %s", typeName(objectType)),
					Node:    expr,
				})
				return diags
			}
		}
	}

	if !isUnknownType(objectType) {
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: cannot assign via [] without IndexMut implementation on type %s", typeName(objectType)),
			Node:    expr,
		})
	}
	return diags
}
