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
				Message: fmt.Sprintf("typechecker: unary '%s' requires integer operand (got %s)", expr.Operator, typeName(operandType)),
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

	if expr.Operator == "|>" || expr.Operator == "|>>" {
		pipeCall := buildPipeCall(expr)
		if pipeCall == nil {
			return []Diagnostic{{Message: "typechecker: invalid pipe expression", Node: expr}}, UnknownType{}
		}
		pipeDiags, pipeType := c.checkFunctionCallExpression(env, pipeCall)
		c.infer.set(expr, pipeType)
		return pipeDiags, pipeType
	}

	leftDiags, leftType := c.checkExpression(env, expr.Left)

	var diags []Diagnostic
	diags = append(diags, leftDiags...)

	rightDiags, rightType := c.checkExpression(env, expr.Right)
	diags = append(diags, rightDiags...)

	resultType := Type(UnknownType{})
	boolType := PrimitiveType{Kind: PrimitiveBool}

	switch expr.Operator {
	case "&&", "||":
		if !typeAssignable(leftType, boolType) && !isUnknownType(leftType) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' left operand must be bool (got %s)", expr.Operator, typeName(leftType)),
				Node:    expr.Right,
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
		default:
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
		}
	case "-", "*", "^":
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
	case "/":
		resType, err := resolveDivisionBinaryType(leftType, rightType)
		if err != "" {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' %s", expr.Operator, err),
				Node:    expr,
			})
			resultType = UnknownType{}
			break
		}
		resultType = resType
	case "//", "%":
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
	case "/%":
		intType, err := resolveIntegerBinaryType(leftType, rightType)
		if err != "" {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' %s", expr.Operator, err),
				Node:    expr,
			})
			resultType = UnknownType{}
			break
		}
		resultType = AppliedType{
			Base: StructType{
				StructName: "DivMod",
				TypeParams: []GenericParamSpec{{Name: "T"}},
			},
			Arguments: []Type{intType},
		}
	case ">", "<", ">=", "<=":
		if isUnknownType(leftType) || isUnknownType(rightType) {
			resultType = boolType
			break
		}
		if isStringType(leftType) && isStringType(rightType) {
			resultType = boolType
			break
		}
		_, err := resolveNumericBinaryType(leftType, rightType)
		if err != "" && !(isUnknownType(leftType) || isUnknownType(rightType)) {
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("typechecker: '%s' %s", expr.Operator, err),
				Node:    expr,
			})
		}
		resultType = boolType
	case "==", "!=":
		// Equality comparisons are defined for all types; we only assign bool.
		resultType = boolType
	case ".&", "&", ".|", "|", ".^":
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
	case ".<<", "<<", ".>>", ">>":
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
	case "|>", "|>>":
		// Pipe expressions are desugared by the interpreter; the checker currently
		// treats them as opaque and propagates the right-hand side type.
		resultType = rightType
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

func buildPipeCall(expr *ast.BinaryExpression) *ast.FunctionCall {
	if expr == nil {
		return nil
	}
	if _, ok := placeholderFunctionPlan(expr.Right); ok {
		return ast.NewFunctionCall(expr.Right, []ast.Expression{expr.Left}, nil, false)
	}
	if call, ok := expr.Right.(*ast.FunctionCall); ok && call != nil {
		args := append([]ast.Expression{expr.Left}, call.Arguments...)
		typeArgs := call.TypeArguments
		if len(typeArgs) > 0 {
			copied := make([]ast.TypeExpression, len(typeArgs))
			copy(copied, typeArgs)
			typeArgs = copied
		}
		return ast.NewFunctionCall(call.Callee, args, typeArgs, call.IsTrailingLambda)
	}
	return ast.NewFunctionCall(expr.Right, []ast.Expression{expr.Left}, nil, false)
}

func resolveNumericBinaryType(left, right Type) (Type, string) {
	if isUnknownType(left) || isUnknownType(right) {
		return UnknownType{}, ""
	}
	if isFloatType(left) || isFloatType(right) {
		if !isNumericType(left) || !isNumericType(right) {
			return UnknownType{}, fmt.Sprintf("requires numeric operands, got %s and %s", typeName(left), typeName(right))
		}
		return resolveFloatBinaryType(left, right)
	}
	if !isNumericType(left) || !isNumericType(right) {
		return UnknownType{}, fmt.Sprintf("requires numeric operands, got %s and %s", typeName(left), typeName(right))
	}
	return resolveIntegerBinaryType(left, right)
}

func resolveDivisionBinaryType(left, right Type) (Type, string) {
	if isUnknownType(left) || isUnknownType(right) {
		return UnknownType{}, ""
	}
	if isFloatType(left) || isFloatType(right) {
		if !isNumericType(left) || !isNumericType(right) {
			return UnknownType{}, fmt.Sprintf("requires numeric operands, got %s and %s", typeName(left), typeName(right))
		}
		return resolveFloatBinaryType(left, right)
	}
	if !isNumericType(left) || !isNumericType(right) {
		return UnknownType{}, fmt.Sprintf("requires numeric operands, got %s and %s", typeName(left), typeName(right))
	}
	return FloatType{Suffix: "f64"}, ""
}

func resolveFloatBinaryType(left, right Type) (Type, string) {
	result := "f32"
	if lFloat, ok := left.(FloatType); ok && lFloat.Suffix == "f64" {
		result = "f64"
	}
	if rFloat, ok := right.(FloatType); ok && rFloat.Suffix == "f64" {
		result = "f64"
	}
	return FloatType{Suffix: result}, ""
}

func resolveIntegerBinaryType(left, right Type) (Type, string) {
	if isUnknownType(left) || isUnknownType(right) {
		return UnknownType{}, ""
	}
	leftSuffix, ok := integerSuffixForType(left)
	if !ok {
		return UnknownType{}, fmt.Sprintf("requires integer operands, got %s and %s", typeName(left), typeName(right))
	}
	rightSuffix, ok := integerSuffixForType(right)
	if !ok {
		return UnknownType{}, fmt.Sprintf("requires integer operands, got %s and %s", typeName(left), typeName(right))
	}
	resultSuffix, errMsg := promoteIntegerSuffixes(leftSuffix, rightSuffix)
	if errMsg != "" {
		return UnknownType{}, errMsg
	}
	return IntegerType{Suffix: resultSuffix}, ""
}

func integerSuffixForType(t Type) (string, bool) {
	switch val := t.(type) {
	case IntegerType:
		if val.Suffix != "" {
			return val.Suffix, true
		}
		return "i32", true
	case PrimitiveType:
		if val.Kind == PrimitiveInt {
			return "i32", true
		}
	}
	return "", false
}

func promoteIntegerSuffixes(left, right string) (string, string) {
	leftInfo, ok := integerInfo(left)
	if !ok {
		return "", fmt.Sprintf("requires integer operands, got %s", left)
	}
	rightInfo, ok := integerInfo(right)
	if !ok {
		return "", fmt.Sprintf("requires integer operands, got %s", right)
	}
	if leftInfo.signed == rightInfo.signed {
		targetBits := leftInfo.bits
		if rightInfo.bits > targetBits {
			targetBits = rightInfo.bits
		}
		if leftInfo.signed {
			if suffix, ok := smallestSignedFor(targetBits); ok {
				return suffix, ""
			}
		} else {
			if suffix, ok := smallestUnsignedFor(targetBits); ok {
				return suffix, ""
			}
		}
		return "", fmt.Sprintf("integer operands %s and %s require %d bits, exceeding available widths", left, right, targetBits)
	}
	needed := leftInfo.bits + 1
	if rightInfo.bits+1 > needed {
		needed = rightInfo.bits + 1
	}
	if suffix, ok := smallestSignedFor(needed); ok {
		return suffix, ""
	}
	var unsignedCandidate *intBounds
	var unsignedName string
	if !leftInfo.signed {
		unsignedCandidate = &leftInfo
		unsignedName = left
	}
	if !rightInfo.signed && (unsignedCandidate == nil || rightInfo.bits > unsignedCandidate.bits) {
		unsignedCandidate = &rightInfo
		unsignedName = right
	}
	if unsignedCandidate != nil && unsignedCandidate.bits >= max(leftInfo.bits, rightInfo.bits) {
		return unsignedName, ""
	}
	return "", fmt.Sprintf("integer operands %s and %s require %d bits, exceeding available widths", left, right, needed)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
