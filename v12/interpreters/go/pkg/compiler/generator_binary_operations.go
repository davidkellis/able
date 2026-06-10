package compiler

import "fmt"

func (g *generator) compileFloatPowExpression(ctx *compileContext, left string, right string, operandType string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
	}
	if operandType == "float32" {
		return lines, fmt.Sprintf("__able_pow_float32(%s, %s)", leftTemp, rightTemp)
	}
	return lines, fmt.Sprintf("__able_pow_float64(%s, %s)", leftTemp, rightTemp)
}

func (g *generator) compileBinaryOperation(ctx *compileContext, op string, leftExpr string, leftType string, rightExpr string, rightType string, expected string, nodeName string) ([]string, string, string, bool) {
	if leftType == "runtime.Value" || rightType == "runtime.Value" {
		rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, op, leftExpr, leftType, rightExpr, rightType, expected)
		if !ok {
			return nil, "", "", false
		}
		return rtLines, rtExpr, rtType, true
	}
	if leftType != rightType {
		ctx.setReason("binary operand type mismatch")
		return nil, "", "", false
	}
	switch op {
	case "+":
		if !g.isStringType(leftType) && !g.isNumericType(leftType) {
			ctx.setReason("unsupported + operand type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		if g.isIntegerType(leftType) {
			opLines, opExpr := g.compileCheckedIntegerBinaryExpression(ctx, leftExpr, rightExpr, leftType, "+", nodeName, nil, nil)
			return opLines, opExpr, leftType, true
		}
		return nil, fmt.Sprintf("(%s + %s)", leftExpr, rightExpr), leftType, true
	case "-", "*":
		if !g.isNumericType(leftType) {
			ctx.setReason("unsupported numeric operator type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		if g.isIntegerType(leftType) {
			opLines, opExpr := g.compileCheckedIntegerBinaryExpression(ctx, leftExpr, rightExpr, leftType, op, nodeName, nil, nil)
			return opLines, opExpr, leftType, true
		}
		return nil, fmt.Sprintf("(%s %s %s)", leftExpr, op, rightExpr), leftType, true
	case "/":
		if !g.isNumericType(leftType) {
			ctx.setReason("unsupported numeric operator type")
			return nil, "", "", false
		}
		resultType := leftType
		if g.isIntegerType(resultType) {
			resultType = "float64"
		}
		if !g.canCoerceStaticExpr(expected, resultType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		opLines, opExpr := g.compileDivisionExpression(ctx, leftExpr, rightExpr, leftType, resultType, nodeName)
		return opLines, opExpr, resultType, true
	case "//", "%":
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported integer operator type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		opLines, opExpr := g.compileDivModExpression(ctx, leftExpr, rightExpr, leftType, op, nodeName)
		return opLines, opExpr, leftType, true
	case ".&", ".|", ".^", "&", "|", "^":
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported bitwise operator type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		switch op {
		case ".&":
			op = "&"
		case ".|":
			op = "|"
		case ".^":
			op = "^"
		}
		return nil, fmt.Sprintf("(%s %s %s)", leftExpr, op, rightExpr), leftType, true
	case ".<<", ".>>", "<<", ">>":
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported shift operator type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		if op == ".<<" {
			op = "<<"
		} else if op == ".>>" {
			op = ">>"
		}
		opLines, opExpr := g.compileShiftExpression(ctx, leftExpr, rightExpr, leftType, op, nodeName)
		return opLines, opExpr, leftType, true
	default:
		ctx.setReason("unsupported operator")
		return nil, "", "", false
	}
}
