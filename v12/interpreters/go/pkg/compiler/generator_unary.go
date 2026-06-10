package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileUnaryExpression(ctx *compileContext, expr *ast.UnaryExpression, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing unary expression")
		return nil, "", "", false
	}
	switch expr.Operator {
	case ast.UnaryOperatorNegate:
		operandExpected := expected
		if operandExpected != "" && !g.isNumericType(operandExpected) {
			// The surrounding expression can expect a union or interface even
			// though negation itself still operates on a concrete primitive.
			// Preserve that primitive here and let the shared outer coercion
			// wrap the completed result.
			operandExpected = ""
		}
		operandLines, operand, operandType, ok := g.compileExprLines(ctx, expr.Operand, operandExpected)
		if !ok {
			return nil, "", "", false
		}
		if g.isIntegerType(operandType) {
			nodeName := g.diagNodeName(expr, "*ast.UnaryExpression", "unary")
			temp := ctx.newTemp()
			bitsExpr := g.bitSizeExpr(operandType)
			lines := append([]string{}, operandLines...)
			lines = append(lines, fmt.Sprintf("%s := %s", temp, operand))
			if g.isWideIntegerType(operandType) {
				resultTemp := ctx.newTemp()
				okTemp := ctx.newTemp()
				controlTemp := ctx.newTemp()
				if operandType == "runtime.Uint128" {
					lines = append(lines, fmt.Sprintf("%s, %s := (runtime.Uint128{}).SubChecked(%s)", resultTemp, okTemp, temp))
				} else {
					lines = append(lines, fmt.Sprintf("%s, %s := (%s).NegateChecked()", resultTemp, okTemp, temp))
				}
				lines = append(lines,
					fmt.Sprintf("var %s *__ableControl", controlTemp),
					fmt.Sprintf("if !%s { %s = __able_raise_overflow(%s) }", okTemp, controlTemp, nodeName),
				)
				controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, controlLines...)
				return lines, resultTemp, operandType, true
			}
			if g.isUnsignedIntegerType(operandType) {
				resultTemp := ctx.newTemp()
				controlTemp := ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s, %s := __able_checked_sub_unsigned(uint64(0), uint64(%s), %s, %s)", resultTemp, controlTemp, temp, bitsExpr, nodeName))
				controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, controlLines...)
				return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp), operandType, true
			}
			resultTemp := ctx.newTemp()
			controlTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s, %s := __able_checked_sub_signed(int64(0), int64(%s), %s, %s)", resultTemp, controlTemp, temp, bitsExpr, nodeName))
			controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, controlLines...)
			return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp), operandType, true
		}
		if !g.isNumericType(operandType) {
			opConvLines, operandRuntime, ok := g.lowerRuntimeValue(ctx, operand, operandType)
			if !ok {
				ctx.setReason("unsupported unary operand type")
				return nil, "", "", false
			}
			operandLines = append(operandLines, opConvLines...)
			resultTemp := ctx.newTemp()
			controlTemp := ctx.newTemp()
			operandLines = append(operandLines, fmt.Sprintf("%s, %s := __able_unary_op(%q, %s)", resultTemp, controlTemp, string(expr.Operator), operandRuntime))
			controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
			if !ok {
				return nil, "", "", false
			}
			operandLines = append(operandLines, controlLines...)
			unaryExpr := resultTemp
			if expected == "" || expected == "runtime.Value" {
				return operandLines, unaryExpr, "runtime.Value", true
			}
			convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, unaryExpr, expected)
			if !ok {
				ctx.setReason("unary expression type mismatch")
				return nil, "", "", false
			}
			lines := append([]string{}, operandLines...)
			lines = append(lines, convLines...)
			return lines, converted, expected, true
		}
		return operandLines, fmt.Sprintf("(-%s)", operand), operandType, true
	case ast.UnaryOperatorNot:
		if expected != "" && expected != "bool" {
			ctx.setReason("unary expression type mismatch")
			return nil, "", "", false
		}
		operandLines, operand, operandType, ok := g.compileExprLines(ctx, expr.Operand, "")
		if !ok {
			return nil, "", "", false
		}
		if operandType == "bool" {
			return operandLines, fmt.Sprintf("(!%s)", operand), "bool", true
		}
		operandRuntime := operand
		if operandType != "runtime.Value" {
			convLines, converted, ok := g.lowerRuntimeValue(ctx, operand, operandType)
			if !ok {
				ctx.setReason("unsupported unary operand type")
				return nil, "", "", false
			}
			operandLines = append(operandLines, convLines...)
			operandRuntime = converted
		}
		return operandLines, fmt.Sprintf("!__able_truthy(%s)", operandRuntime), "bool", true
	case ast.UnaryOperatorBitNot:
		operandExpected := expected
		if operandExpected != "" && !g.isIntegerType(operandExpected) {
			// As with numeric negation, an outer union/interface expectation
			// must not erase the integer carrier before the operator runs.
			operandExpected = ""
		}
		operandLines, operand, operandType, ok := g.compileExprLines(ctx, expr.Operand, operandExpected)
		if !ok {
			return nil, "", "", false
		}
		if !g.isIntegerType(operandType) {
			opConvLines, operandRuntime, ok := g.lowerRuntimeValue(ctx, operand, operandType)
			if !ok {
				ctx.setReason("unsupported bitwise operand type")
				return nil, "", "", false
			}
			operandLines = append(operandLines, opConvLines...)
			resultTemp := ctx.newTemp()
			controlTemp := ctx.newTemp()
			operandLines = append(operandLines, fmt.Sprintf("%s, %s := __able_unary_op(%q, %s)", resultTemp, controlTemp, string(expr.Operator), operandRuntime))
			controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
			if !ok {
				return nil, "", "", false
			}
			operandLines = append(operandLines, controlLines...)
			unaryExpr := resultTemp
			if expected == "" || expected == "runtime.Value" {
				return operandLines, unaryExpr, "runtime.Value", true
			}
			convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, unaryExpr, expected)
			if !ok {
				ctx.setReason("unary expression type mismatch")
				return nil, "", "", false
			}
			lines := append([]string{}, operandLines...)
			lines = append(lines, convLines...)
			return lines, converted, expected, true
		}
		if g.isWideIntegerType(operandType) {
			return operandLines, fmt.Sprintf("(%s).Not()", operand), operandType, true
		}
		return operandLines, fmt.Sprintf("(^%s)", operand), operandType, true
	default:
		ctx.setReason("unsupported unary operator")
		return nil, "", "", false
	}
}
