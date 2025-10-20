package typechecker

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
)

func (c *Checker) checkUnaryExpression(env *Environment, expr *ast.UnaryExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}
	operandDiags, operandType := c.checkExpression(env, expr.Operand)
	var diags []Diagnostic
	diags = append(diags, operandDiags...)

	resultType := Type(UnknownType{})
	switch expr.Operator {
	case ast.UnaryOperatorNegate:
		if isUnknownType(operandType) {
			resultType = UnknownType{}
			break
		}
		if !isNumericType(operandType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: unary '%s' requires numeric operand (got %s)", expr.Operator, typeName(operandType)),
				Node:    expr,
			})
			resultType = UnknownType{}
			break
		}
		resultType = operandType
	case ast.UnaryOperatorNot:
		boolType := PrimitiveType{Kind: PrimitiveBool}
		if !typeAssignable(operandType, boolType) && !isUnknownType(operandType) {
			diags = append(diags, Diagnostic{
				Message: "typechecker: unary '!' requires boolean operand",
				Node:    expr,
			})
		}
		resultType = boolType
	case ast.UnaryOperatorBitNot:
		if isUnknownType(operandType) {
			resultType = UnknownType{}
			break
		}
		if !isIntegerType(operandType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: unary '~' requires integer operand (got %s)", typeName(operandType)),
				Node:    expr,
			})
			resultType = UnknownType{}
			break
		}
		resultType = operandType
	default:
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: unsupported unary operator %q", expr.Operator),
			Node:    expr,
		})
		resultType = UnknownType{}
	}

	c.infer.set(expr, resultType)
	return diags, resultType
}

func (c *Checker) checkBinaryExpression(env *Environment, expr *ast.BinaryExpression) ([]Diagnostic, Type) {
	if expr == nil {
		return nil, UnknownType{}
	}

	leftDiags, leftType := c.checkExpression(env, expr.Left)
	rightDiags, rightType := c.checkExpression(env, expr.Right)

	var diags []Diagnostic
	diags = append(diags, leftDiags...)
	diags = append(diags, rightDiags...)

	resultType := Type(UnknownType{})
	boolType := PrimitiveType{Kind: PrimitiveBool}

	switch expr.Operator {
	case "&&", "||":
		if !typeAssignable(leftType, boolType) && !isUnknownType(leftType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' left operand must be bool (got %s)", expr.Operator, typeName(leftType)),
				Node:    expr.Left,
			})
		}
		if !typeAssignable(rightType, boolType) && !isUnknownType(rightType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' right operand must be bool (got %s)", expr.Operator, typeName(rightType)),
				Node:    expr.Right,
			})
		}
		resultType = boolType
	case "+": // string concatenation or numeric addition
		switch {
		case isStringType(leftType) && isStringType(rightType):
			resultType = PrimitiveType{Kind: PrimitiveString}
		case isUnknownType(leftType) || isUnknownType(rightType):
			resultType = UnknownType{}
		case isNumericType(leftType) && isNumericType(rightType):
			resType, err := resolveNumericBinaryType(leftType, rightType)
			if err != "" {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: '+' %s", err),
					Node:    expr,
				})
				resultType = UnknownType{}
			} else {
				resultType = resType
			}
		default:
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '+' requires both operands to be numeric or string (got %s and %s)", typeName(leftType), typeName(rightType)),
				Node:    expr,
			})
			resultType = UnknownType{}
		}
	case "-", "*", "/", "%":
		if isUnknownType(leftType) || isUnknownType(rightType) {
			resultType = UnknownType{}
			break
		}
		if !isNumericType(leftType) || !isNumericType(rightType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' requires numeric operands (got %s and %s)", expr.Operator, typeName(leftType), typeName(rightType)),
				Node:    expr,
			})
			resultType = UnknownType{}
			break
		}
		resType, err := resolveNumericBinaryType(leftType, rightType)
		if err != "" {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' %s", expr.Operator, err),
				Node:    expr,
			})
			resultType = UnknownType{}
			break
		}
		resultType = resType
	case ">", "<", ">=", "<=":
		if isUnknownType(leftType) || isUnknownType(rightType) {
			resultType = boolType
			break
		}
		switch {
		case isNumericType(leftType) && isNumericType(rightType):
			if _, err := resolveNumericBinaryType(leftType, rightType); err != "" {
				diags = append(diags, Diagnostic{
					Message: fmt.Sprintf("typechecker: '%s' %s", expr.Operator, err),
					Node:    expr,
				})
			}
			resultType = boolType
		case isStringType(leftType) && isStringType(rightType):
			resultType = boolType
		default:
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' requires operands to share a numeric or string type (got %s and %s)", expr.Operator, typeName(leftType), typeName(rightType)),
				Node:    expr,
			})
			resultType = boolType
		}
	case "==", "!=":
		// Equality comparisons are defined for all types; we only assign bool.
		resultType = boolType
	case "&", "|", "^":
		if isUnknownType(leftType) || isUnknownType(rightType) {
			resultType = UnknownType{}
			break
		}
		if !isIntegerType(leftType) || !isIntegerType(rightType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' requires integer operands (got %s and %s)", expr.Operator, typeName(leftType), typeName(rightType)),
				Node:    expr,
			})
			resultType = UnknownType{}
			break
		}
		intType, err := resolveIntegerBinaryType(leftType, rightType)
		if err != "" {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' %s", expr.Operator, err),
				Node:    expr,
			})
			resultType = UnknownType{}
			break
		}
		resultType = intType
	case "<<", ">>":
		if isUnknownType(leftType) || isUnknownType(rightType) {
			resultType = UnknownType{}
			break
		}
		if !isIntegerType(leftType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' left operand must be integer (got %s)", expr.Operator, typeName(leftType)),
				Node:    expr.Left,
			})
		}
		if !isIntegerType(rightType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' right operand must be integer (got %s)", expr.Operator, typeName(rightType)),
				Node:    expr.Right,
			})
		}
		if intType, err := resolveIntegerBinaryType(leftType, rightType); err == "" {
			resultType = intType
		} else {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' %s", expr.Operator, err),
				Node:    expr,
			})
			resultType = UnknownType{}
		}
	default:
		diags = append(diags, Diagnostic{
			Message: fmt.Sprintf("typechecker: unsupported binary operator %q", expr.Operator),
			Node:    expr,
		})
		resultType = UnknownType{}
	}

	c.infer.set(expr, resultType)
	return diags, resultType
}

func resolveNumericBinaryType(left, right Type) (Type, string) {
	if isUnknownType(left) || isUnknownType(right) {
		return UnknownType{}, ""
	}
	if !isNumericType(left) || !isNumericType(right) {
		return UnknownType{}, fmt.Sprintf("requires numeric operands, got %s and %s", typeName(left), typeName(right))
	}
	if isFloatType(left) || isFloatType(right) {
		return resolveFloatBinaryType(left, right)
	}
	return resolveIntegerBinaryType(left, right)
}

func resolveFloatBinaryType(left, right Type) (Type, string) {
	if !isNumericType(left) || !isNumericType(right) {
		return UnknownType{}, fmt.Sprintf("requires numeric operands, got %s and %s", typeName(left), typeName(right))
	}
	var result Type
	if isFloatType(left) {
		result = left
	}
	if isFloatType(right) {
		if result == nil {
			result = right
		} else if result.Name() != right.Name() {
			return UnknownType{}, fmt.Sprintf("operands use incompatible float types %s and %s", typeName(left), typeName(right))
		}
	}
	if result == nil {
		result = FloatType{Suffix: "f64"}
	}
	return result, ""
}

func resolveIntegerBinaryType(left, right Type) (Type, string) {
	if isUnknownType(left) || isUnknownType(right) {
		return UnknownType{}, ""
	}
	if !isIntegerType(left) || !isIntegerType(right) {
		return UnknownType{}, fmt.Sprintf("requires integer operands, got %s and %s", typeName(left), typeName(right))
	}

	if lInt, ok := left.(IntegerType); ok {
		if rInt, ok := right.(IntegerType); ok {
			if lInt.Suffix == rInt.Suffix {
				return lInt, ""
			}
			return UnknownType{}, fmt.Sprintf("operands use incompatible integer types %s and %s", typeName(left), typeName(right))
		}
		if isPrimitiveInt(right) {
			return lInt, ""
		}
		return lInt, ""
	}

	if rInt, ok := right.(IntegerType); ok {
		if isPrimitiveInt(left) {
			return rInt, ""
		}
		return rInt, ""
	}

	if isPrimitiveInt(left) && isPrimitiveInt(right) {
		return PrimitiveType{Kind: PrimitiveInt}, ""
	}

	return UnknownType{}, ""
}
