package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) inlineCheckedIntegerBinaryExpression(ctx *compileContext, left string, right string, operandType string, op string, nodeName string, leftExpr ast.Expression, rightExpr ast.Expression) ([]string, string, bool) {
	if g == nil || ctx == nil || operandType == "" {
		return nil, "", false
	}
	switch op {
	case "+", "-":
	default:
		return nil, "", false
	}
	if g.isUnsignedIntegerType(operandType) {
		return g.inlineCheckedUnsignedIntegerBinaryExpression(ctx, left, right, operandType, op, nodeName)
	}
	if g.isSignedIntegerType(operandType) {
		return g.inlineCheckedSignedIntegerBinaryExpression(ctx, left, right, operandType, op, nodeName, leftExpr, rightExpr)
	}
	return nil, "", false
}

func (g *generator) inlineCheckedSignedIntegerBinaryExpression(ctx *compileContext, left string, right string, operandType string, op string, nodeName string, leftExpr ast.Expression, rightExpr ast.Expression) ([]string, string, bool) {
	lowerBound, upperBound, ok := g.fixedSignedIntegerBounds(operandType)
	if !ok {
		return nil, "", false
	}
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	wideTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	overflowTransfer, ok := g.controlTransferLines(ctx, fmt.Sprintf("__able_raise_overflow(%s)", nodeName))
	if !ok {
		return nil, "", false
	}
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
	}
	if g.provesSignedIntegerBinaryOperationSafe(ctx, op, operandType, leftExpr, rightExpr) {
		lines = append(lines, fmt.Sprintf("%s := %s %s %s", resultTemp, leftTemp, op, rightTemp))
		return lines, resultTemp, true
	}
	lines = append(lines,
		fmt.Sprintf("%s := int64(%s) %s int64(%s)", wideTemp, leftTemp, op, rightTemp),
		fmt.Sprintf("if %s < %s || %s > %s {", wideTemp, lowerBound, wideTemp, upperBound),
	)
	lines = append(lines, indentLines(overflowTransfer, 1)...)
	lines = append(lines,
		"}",
		fmt.Sprintf("%s := %s(%s)", resultTemp, operandType, wideTemp),
	)
	return lines, resultTemp, true
}

func (g *generator) inlineCheckedUnsignedIntegerBinaryExpression(ctx *compileContext, left string, right string, operandType string, op string, nodeName string) ([]string, string, bool) {
	maxBound, ok := g.fixedUnsignedIntegerMax(operandType)
	if !ok {
		return nil, "", false
	}
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	wideTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	overflowTransfer, ok := g.controlTransferLines(ctx, fmt.Sprintf("__able_raise_overflow(%s)", nodeName))
	if !ok {
		return nil, "", false
	}
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
	}
	switch op {
	case "+":
		lines = append(lines,
			fmt.Sprintf("%s := uint64(%s) + uint64(%s)", wideTemp, leftTemp, rightTemp),
			fmt.Sprintf("if %s > %s {", wideTemp, maxBound),
		)
	case "-":
		lines = append(lines,
			fmt.Sprintf("if uint64(%s) < uint64(%s) {", leftTemp, rightTemp),
		)
	default:
		return nil, "", false
	}
	lines = append(lines, indentLines(overflowTransfer, 1)...)
	lines = append(lines, "}")
	if op == "-" {
		lines = append(lines, fmt.Sprintf("%s := uint64(%s) - uint64(%s)", wideTemp, leftTemp, rightTemp))
	}
	lines = append(lines, fmt.Sprintf("%s := %s(%s)", resultTemp, operandType, wideTemp))
	return lines, resultTemp, true
}

func (g *generator) fixedSignedIntegerBounds(goType string) (string, string, bool) {
	switch goType {
	case "int8":
		return "-128", "127", true
	case "int16":
		return "-32768", "32767", true
	case "int32":
		return "-2147483648", "2147483647", true
	default:
		return "", "", false
	}
}

func (g *generator) provesSignedIntegerBinaryOperationSafe(ctx *compileContext, op string, operandType string, leftExpr ast.Expression, rightExpr ast.Expression) bool {
	if g == nil || ctx == nil || operandType == "" {
		return false
	}
	switch op {
	case "+":
		leftFact, leftOK := g.exprIntegerFact(ctx, leftExpr)
		rightFact, rightOK := g.exprIntegerFact(ctx, rightExpr)
		if !leftOK || !rightOK || !leftFact.NonNegative || !rightFact.NonNegative || !leftFact.HasMax || !rightFact.HasMax {
			return false
		}
		sum, ok := addInt64NoOverflow(leftFact.MaxInclusive, rightFact.MaxInclusive)
		if !ok {
			return false
		}
		upperBound, ok := g.signedIntegerUpperBound(operandType)
		return ok && sum <= upperBound
	case "-":
		return g.exprProvenNonNegative(ctx, leftExpr) && g.exprProvenNonNegative(ctx, rightExpr)
	default:
		return false
	}
}

func (g *generator) fixedUnsignedIntegerMax(goType string) (string, bool) {
	switch goType {
	case "uint8":
		return "255", true
	case "uint16":
		return "65535", true
	case "uint32":
		return "4294967295", true
	default:
		return "", false
	}
}
