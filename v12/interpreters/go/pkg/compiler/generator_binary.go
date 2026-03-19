package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileBinaryExpression(ctx *compileContext, expr *ast.BinaryExpression, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing binary expression")
		return nil, "", "", false
	}
	if expr.Operator == "|>" || expr.Operator == "|>>" {
		return g.compilePipeExpression(ctx, expr, expected)
	}
	if expr.Operator == "&&" || expr.Operator == "||" {
		return g.compileLogicalBinaryExpression(ctx, expr, expected)
	}
	// Compile operands once for all non-logical operators.
	operandLines, left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
	if !ok {
		return nil, "", "", false
	}
	// If either operand is runtime.Value or any, use runtime binary operation.
	if leftType == "runtime.Value" || rightType == "runtime.Value" || leftType == "any" || rightType == "any" {
		rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, expr.Operator, left, leftType, right, rightType, expected)
		if !ok {
			return nil, "", "", false
		}
		return append(operandLines, rtLines...), rtExpr, rtType, true
	}
	switch expr.Operator {
	case "==", "!=", "<", "<=", ">", ">=":
		if leftType != rightType {
			ctx.setReason("comparison operand type mismatch")
			return nil, "", "", false
		}
		if expr.Operator == "==" || expr.Operator == "!=" {
			if !g.isEqualityComparable(leftType) {
				if expected != "" && expected != "bool" {
					ctx.setReason("comparison expression type mismatch")
					return nil, "", "", false
				}
				rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, expr.Operator, left, leftType, right, rightType, "bool")
				if !ok {
					return nil, "", "", false
				}
				return append(operandLines, rtLines...), rtExpr, rtType, true
			}
		} else if !g.isOrderedComparable(leftType) {
			if expected != "" && expected != "bool" {
				ctx.setReason("comparison expression type mismatch")
				return nil, "", "", false
			}
			rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, expr.Operator, left, leftType, right, rightType, "bool")
			if !ok {
				return nil, "", "", false
			}
			return append(operandLines, rtLines...), rtExpr, rtType, true
		}
		if expected != "" && expected != "bool" {
			ctx.setReason("comparison expression type mismatch")
			return nil, "", "", false
		}
		return operandLines, fmt.Sprintf("(%s %s %s)", left, expr.Operator, right), "bool", true
	case "^":
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return nil, "", "", false
		}
		if !g.isNumericType(leftType) {
			ctx.setReason("unsupported numeric operator type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		nodeName := g.diagNodeName(expr, "*ast.BinaryExpression", "binary")
		if g.isIntegerType(leftType) {
			opLines, opExpr := g.compilePowExpression(ctx, left, right, leftType, nodeName)
			return append(operandLines, opLines...), opExpr, leftType, true
		}
		if g.isFloatType(leftType) {
			opLines, opExpr := g.compileFloatPowExpression(ctx, left, right, leftType)
			return append(operandLines, opLines...), opExpr, leftType, true
		}
		rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, expr.Operator, left, leftType, right, rightType, leftType)
		if !ok {
			return nil, "", "", false
		}
		return append(operandLines, rtLines...), rtExpr, rtType, true
	case ".&", ".|", ".^", "&", "|":
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return nil, "", "", false
		}
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported bitwise operator type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		op := expr.Operator
		switch op {
		case ".&":
			op = "&"
		case ".|":
			op = "|"
		case ".^":
			op = "^"
		}
		return operandLines, fmt.Sprintf("(%s %s %s)", left, op, right), leftType, true
	case ".<<", ".>>", "<<", ">>":
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return nil, "", "", false
		}
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported shift operator type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		op := expr.Operator
		if op == ".<<" {
			op = "<<"
		} else if op == ".>>" {
			op = ">>"
		}
		nodeName := g.diagNodeName(expr, "*ast.BinaryExpression", "binary")
		opLines, opExpr := g.compileShiftExpression(ctx, left, right, leftType, op, nodeName)
		return append(operandLines, opLines...), opExpr, leftType, true
	case "+":
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return nil, "", "", false
		}
		resultType := leftType
		if !g.isStringType(resultType) && !g.isNumericType(resultType) {
			rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, expr.Operator, left, leftType, right, rightType, expected)
			if !ok {
				return nil, "", "", false
			}
			return append(operandLines, rtLines...), rtExpr, rtType, true
		}
		if !g.canCoerceStaticExpr(expected, resultType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		if g.isIntegerType(resultType) {
			nodeName := g.diagNodeName(expr, "*ast.BinaryExpression", "binary")
			opLines, opExpr := g.compileCheckedIntegerBinaryExpression(ctx, left, right, resultType, expr.Operator, nodeName)
			return append(operandLines, opLines...), opExpr, resultType, true
		}
		return operandLines, fmt.Sprintf("(%s %s %s)", left, expr.Operator, right), resultType, true
	case "-", "*":
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return nil, "", "", false
		}
		resultType := leftType
		if !g.isNumericType(resultType) {
			rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, expr.Operator, left, leftType, right, rightType, expected)
			if !ok {
				return nil, "", "", false
			}
			return append(operandLines, rtLines...), rtExpr, rtType, true
		}
		if !g.canCoerceStaticExpr(expected, resultType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		if g.isIntegerType(resultType) {
			nodeName := g.diagNodeName(expr, "*ast.BinaryExpression", "binary")
			opLines, opExpr := g.compileCheckedIntegerBinaryExpression(ctx, left, right, resultType, expr.Operator, nodeName)
			return append(operandLines, opLines...), opExpr, resultType, true
		}
		return operandLines, fmt.Sprintf("(%s %s %s)", left, expr.Operator, right), resultType, true
	case "/":
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return nil, "", "", false
		}
		if !g.isNumericType(leftType) {
			rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, expr.Operator, left, leftType, right, rightType, expected)
			if !ok {
				return nil, "", "", false
			}
			return append(operandLines, rtLines...), rtExpr, rtType, true
		}
		resultType := leftType
		if g.isIntegerType(resultType) {
			resultType = "float64"
		}
		if !g.canCoerceStaticExpr(expected, resultType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		nodeName := g.diagNodeName(expr, "*ast.BinaryExpression", "binary")
		opLines, opExpr := g.compileDivisionExpression(ctx, left, right, leftType, resultType, nodeName)
		return append(operandLines, opLines...), opExpr, resultType, true
	case "//", "%":
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return nil, "", "", false
		}
		if !g.isIntegerType(leftType) {
			if expr.Operator == "%" {
				rtLines, rtExpr, rtType, ok := g.compileRuntimeBinaryOperation(ctx, expr.Operator, left, leftType, right, rightType, expected)
				if !ok {
					return nil, "", "", false
				}
				return append(operandLines, rtLines...), rtExpr, rtType, true
			}
			ctx.setReason("unsupported integer operator type")
			return nil, "", "", false
		}
		if !g.canCoerceStaticExpr(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		nodeName := g.diagNodeName(expr, "*ast.BinaryExpression", "binary")
		opLines, opExpr := g.compileDivModExpression(ctx, left, right, leftType, expr.Operator, nodeName)
		return append(operandLines, opLines...), opExpr, leftType, true
	case "/%":
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return nil, "", "", false
		}
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported integer operator type")
			return nil, "", "", false
		}
		nodeName := g.diagNodeName(expr, "*ast.BinaryExpression", "binary")
		opLines, opExpr, ok := g.compileDivModResultExpression(ctx, left, right, leftType, nodeName)
		if !ok {
			ctx.setReason("unsupported /% operands")
			return nil, "", "", false
		}
		return append(operandLines, opLines...), opExpr, "runtime.Value", true
	default:
		ctx.setReason("unsupported operator")
		return nil, "", "", false
	}
}

func (g *generator) compileLogicalBinaryExpression(ctx *compileContext, expr *ast.BinaryExpression, expected string) ([]string, string, string, bool) {
	leftLines, leftExpr, leftType, ok := g.compileExprLines(ctx, expr.Left, "")
	if !ok {
		return nil, "", "", false
	}
	rightLines, rightExpr, rightType, ok := g.compileExprLines(ctx, expr.Right, "")
	if !ok {
		return nil, "", "", false
	}
	logicalLines := append([]string{}, leftLines...)
	logicalLines = append(logicalLines, rightLines...)
	if leftType == "bool" && rightType == "bool" {
		if expected != "" && expected != "bool" {
			ctx.setReason("logical expression type mismatch")
			return nil, "", "", false
		}
		return logicalLines, fmt.Sprintf("(%s %s %s)", leftExpr, expr.Operator, rightExpr), "bool", true
	}
	leftVal := leftExpr
	if leftType != "runtime.Value" {
		convLines, converted, ok := g.runtimeValueLines(ctx, leftExpr, leftType)
		if !ok {
			ctx.setReason("logical operand unsupported")
			return nil, "", "", false
		}
		logicalLines = append(logicalLines, convLines...)
		leftVal = converted
	}
	rightVal := rightExpr
	if rightType != "runtime.Value" {
		convLines, converted, ok := g.runtimeValueLines(ctx, rightExpr, rightType)
		if !ok {
			ctx.setReason("logical operand unsupported")
			return nil, "", "", false
		}
		logicalLines = append(logicalLines, convLines...)
		rightVal = converted
	}
	leftTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	lines := append(logicalLines,
		fmt.Sprintf("%s := %s", leftTemp, leftVal),
		fmt.Sprintf("var %s runtime.Value", resultTemp),
	)
	if expr.Operator == "&&" {
		lines = append(lines, fmt.Sprintf("if __able_truthy(%s) { %s = %s } else { %s = %s }", leftTemp, resultTemp, rightVal, resultTemp, leftTemp))
	} else {
		lines = append(lines, fmt.Sprintf("if __able_truthy(%s) { %s = %s } else { %s = %s }", leftTemp, resultTemp, leftTemp, resultTemp, rightVal))
	}
	if expected != "" && expected != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("logical expression type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	return lines, resultTemp, "runtime.Value", true
}

func (g *generator) compilePipeExpression(ctx *compileContext, expr *ast.BinaryExpression, expected string) ([]string, string, string, bool) {
	if expr == nil {
		ctx.setReason("missing pipe expression")
		return nil, "", "", false
	}
	leftLines, leftExpr, leftType, ok := g.compileExprLines(ctx, expr.Left, "")
	if !ok {
		return nil, "", "", false
	}
	subjectConvLines, subjectValue, ok := g.runtimeValueLines(ctx, leftExpr, leftType)
	if !ok {
		ctx.setReason("pipe subject unsupported")
		return nil, "", "", false
	}
	subjectTemp := ctx.newTemp()
	lines := append([]string{}, leftLines...)
	lines = append(lines, subjectConvLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", subjectTemp, subjectValue))

	pipeCtx := ctx.child()
	subjectParam := paramInfo{Name: subjectTemp, GoName: subjectTemp, GoType: "runtime.Value"}
	pipeCtx.locals[subjectTemp] = subjectParam
	pipeCtx.implicitReceiver = subjectParam
	pipeCtx.hasImplicitReceiver = true

	if placeholderExpr, _, ok := g.compilePlaceholderLambda(pipeCtx, expr.Right, ""); ok {
		rhsTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", rhsTemp, placeholderExpr))
		callTemp := ctx.newTemp()
		argsTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("var %s []runtime.Value", argsTemp))
		lines = append(lines, fmt.Sprintf("switch %s.(type) { case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue: %s = nil; default: %s = []runtime.Value{%s} }", rhsTemp, argsTemp, argsTemp, subjectTemp))
		var callOK bool
		lines, callTemp, callOK = g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_call_value(%s, %s, nil)", rhsTemp, argsTemp))
		if !callOK {
			return nil, "", "", false
		}
		return g.pipeResultLines(ctx, expected, lines, callTemp)
	}

	if call, ok := expr.Right.(*ast.FunctionCall); ok {
		calleeLines, calleeExpr, calleeType, ok := g.compileExprLines(pipeCtx, call.Callee, "")
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, calleeLines...)
		calleeConvLines, calleeValue, ok := g.runtimeValueLines(ctx, calleeExpr, calleeType)
		if !ok {
			ctx.setReason("pipe call target unsupported")
			return nil, "", "", false
		}
		calleeTemp := ctx.newTemp()
		lines = append(lines, calleeConvLines...)
		lines = append(lines, fmt.Sprintf("%s := %s", calleeTemp, calleeValue))
		argTemps := make([]string, 0, len(call.Arguments))
		for _, arg := range call.Arguments {
			argLines, argExpr, argType, ok := g.compileExprLines(pipeCtx, arg, "")
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, argLines...)
			argConvLines, argValue, ok := g.runtimeValueLines(ctx, argExpr, argType)
			if !ok {
				ctx.setReason("pipe call argument unsupported")
				return nil, "", "", false
			}
			argTemp := ctx.newTemp()
			lines = append(lines, argConvLines...)
			lines = append(lines, fmt.Sprintf("%s := %s", argTemp, argValue))
			argTemps = append(argTemps, argTemp)
		}
		argsTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("var %s []runtime.Value", argsTemp))
		lines = append(lines, fmt.Sprintf("switch %s.(type) { case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue: %s = %s; default: %s = %s }", calleeTemp, argsTemp, runtimeValueSlice(argTemps), argsTemp, runtimeValueSliceWithSubject(subjectTemp, argTemps)))
		callTemp := ctx.newTemp()
		callNode := g.diagNodeName(call, "*ast.FunctionCall", "call")
		var callOK bool
		lines, callTemp, callOK = g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_call_value(%s, %s, %s)", calleeTemp, argsTemp, callNode))
		if !callOK {
			return nil, "", "", false
		}
		return g.pipeResultLines(ctx, expected, lines, callTemp)
	}

	rhsLines, rhsExpr, rhsType, ok := g.compileExprLines(pipeCtx, expr.Right, "")
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, rhsLines...)
	rhsConvLines, rhsValue, ok := g.runtimeValueLines(ctx, rhsExpr, rhsType)
	if !ok {
		ctx.setReason("pipe rhs unsupported")
		return nil, "", "", false
	}
	rhsTemp := ctx.newTemp()
	lines = append(lines, rhsConvLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", rhsTemp, rhsValue))
	argsTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("var %s []runtime.Value", argsTemp))
	lines = append(lines, fmt.Sprintf("switch %s.(type) { case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue: %s = nil; default: %s = []runtime.Value{%s} }", rhsTemp, argsTemp, argsTemp, subjectTemp))
	callTemp := ctx.newTemp()
	var callOK bool
	lines, callTemp, callOK = g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_call_value(%s, %s, nil)", rhsTemp, argsTemp))
	if !callOK {
		return nil, "", "", false
	}
	return g.pipeResultLines(ctx, expected, lines, callTemp)
}

func (g *generator) pipeResultLines(ctx *compileContext, expected string, lines []string, resultTemp string) ([]string, string, string, bool) {
	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		return lines, "struct{}{}", "struct{}", true
	}
	if expected != "" && expected != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("pipe result type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	return lines, resultTemp, "runtime.Value", true
}

func runtimeValueSlice(args []string) string {
	if len(args) == 0 {
		return "nil"
	}
	return "[]runtime.Value{" + strings.Join(args, ", ") + "}"
}

func runtimeValueSliceWithSubject(subject string, args []string) string {
	if len(args) == 0 {
		return "[]runtime.Value{" + subject + "}"
	}
	all := append([]string{subject}, args...)
	return "[]runtime.Value{" + strings.Join(all, ", ") + "}"
}

func (g *generator) compileRuntimeBinaryOperation(ctx *compileContext, op string, leftExpr string, leftType string, rightExpr string, rightType string, expected string) ([]string, string, string, bool) {
	leftLines, leftVal, ok := g.runtimeValueLines(ctx, leftExpr, leftType)
	if !ok {
		ctx.setReason("binary operand unsupported")
		return nil, "", "", false
	}
	rightLines, rightVal, ok := g.runtimeValueLines(ctx, rightExpr, rightType)
	if !ok {
		ctx.setReason("binary operand unsupported")
		return nil, "", "", false
	}
	var lines []string
	lines = append(lines, leftLines...)
	lines = append(lines, rightLines...)
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := __able_binary_op(%q, %s, %s)", resultTemp, controlTemp, op, leftVal, rightVal))
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	expr := resultTemp
	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		if expected != "" && expected != "bool" {
			ctx.setReason("comparison expression type mismatch")
			return nil, "", "", false
		}
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, expr, "bool")
		if !ok {
			ctx.setReason("comparison expression type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, "bool", true
	}
	if expected != "" && expected != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, expr, expected)
		if !ok {
			ctx.setReason("binary expression type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	return lines, expr, "runtime.Value", true
}

func (g *generator) compileBinaryOperands(ctx *compileContext, leftExpr ast.Expression, rightExpr ast.Expression) ([]string, string, string, string, string, bool) {
	if g.isUntypedNumericLiteral(leftExpr) && g.isUntypedNumericLiteral(rightExpr) {
		if g.isUntypedFloatLiteral(leftExpr) || g.isUntypedFloatLiteral(rightExpr) {
			expected := "float64"
			leftLines, left, leftType, ok := g.compileExprLines(ctx, leftExpr, expected)
			if !ok {
				return nil, "", "", "", "", false
			}
			rightLines, right, rightType, ok := g.compileExprLines(ctx, rightExpr, expected)
			if !ok {
				return nil, "", "", "", "", false
			}
			lines := append([]string{}, leftLines...)
			lines = append(lines, rightLines...)
			return lines, left, leftType, right, rightType, true
		}
	}
	if g.isUntypedNumericLiteral(leftExpr) && !g.isUntypedNumericLiteral(rightExpr) {
		rightLines, right, rightType, ok := g.compileExprLines(ctx, rightExpr, "")
		if !ok {
			return nil, "", "", "", "", false
		}
		leftLines, left, leftType, ok := g.compileExprLines(ctx, leftExpr, rightType)
		if !ok {
			return nil, "", "", "", "", false
		}
		lines := append([]string{}, rightLines...)
		lines = append(lines, leftLines...)
		return lines, left, leftType, right, rightType, true
	}
	leftLines, left, leftType, ok := g.compileExprLines(ctx, leftExpr, "")
	if !ok {
		return nil, "", "", "", "", false
	}
	rightLines, right, rightType, ok := g.compileExprLines(ctx, rightExpr, leftType)
	if !ok {
		return nil, "", "", "", "", false
	}
	lines := append([]string{}, leftLines...)
	lines = append(lines, rightLines...)
	return lines, left, leftType, right, rightType, true
}

func (g *generator) isUntypedFloatLiteral(expr ast.Expression) bool {
	switch lit := expr.(type) {
	case *ast.FloatLiteral:
		return lit != nil && lit.FloatType == nil
	default:
		return false
	}
}

func (g *generator) bitSizeExpr(goType string) string {
	switch goType {
	case "int", "uint":
		return "bridge.NativeIntBits"
	default:
		return fmt.Sprintf("%d", g.intBits(goType))
	}
}

func (g *generator) compileDivisionExpression(ctx *compileContext, left string, right string, operandType string, resultType string, nodeName string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	resultTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
	}
	if g.isIntegerType(operandType) {
		transferLines, ok := g.controlTransferLines(ctx, fmt.Sprintf("control"))
		if !ok {
			return nil, ""
		}
		lines = append(lines, fmt.Sprintf("if %s == 0 {", rightTemp))
		lines = append(lines, indentLines([]string{
			fmt.Sprintf("if control := __able_raise_division_by_zero(%s); control != nil {", nodeName),
		}, 1)...)
		lines = append(lines, indentLines(transferLines, 2)...)
		lines = append(lines, indentLines([]string{
			"}",
		}, 1)...)
		lines = append(lines, "}")
		lines = append(lines, fmt.Sprintf("%s := float64(%s) / float64(%s)", resultTemp, leftTemp, rightTemp))
		return lines, resultTemp
	}
	lines = append(lines, fmt.Sprintf("%s := (%s / %s)", resultTemp, leftTemp, rightTemp))
	return lines, resultTemp
}

func (g *generator) compileDivModExpression(ctx *compileContext, left string, right string, operandType string, op string, nodeName string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	bitsExpr := g.bitSizeExpr(operandType)
	helper := "__able_divmod_signed"
	cast := "int64"
	if g.isUnsignedIntegerType(operandType) {
		helper = "__able_divmod_unsigned"
		cast = "uint64"
	}
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
	}
	if op == "//" {
		lines = append(lines, fmt.Sprintf("%s, _, %s := %s(%s(%s), %s(%s), %s, %s)", resultTemp, controlTemp, helper, cast, leftTemp, cast, rightTemp, bitsExpr, nodeName))
	} else {
		lines = append(lines, fmt.Sprintf("_, %s, %s := %s(%s(%s), %s(%s), %s, %s)", resultTemp, controlTemp, helper, cast, leftTemp, cast, rightTemp, bitsExpr, nodeName))
	}
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, ""
	}
	lines = append(lines, controlLines...)
	return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp)
}

func (g *generator) compileDivModResultExpression(ctx *compileContext, left string, right string, operandType string, nodeName string) ([]string, string, bool) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	bitsExpr := g.bitSizeExpr(operandType)
	helper := "__able_divmod_signed"
	cast := "int64"
	if g.isUnsignedIntegerType(operandType) {
		helper = "__able_divmod_unsigned"
		cast = "uint64"
	}
	suffix, ok := g.integerTypeSuffix(operandType)
	if !ok {
		return nil, "", false
	}
	quotTemp := ctx.newTemp()
	remTemp := ctx.newTemp()
	quotVal, ok := g.runtimeValueExpr(quotTemp, operandType)
	if !ok {
		return nil, "", false
	}
	remVal, ok := g.runtimeValueExpr(remTemp, operandType)
	if !ok {
		return nil, "", false
	}
	typeExpr := fmt.Sprintf("ast.Ty(%q)", suffix)
	qTemp := ctx.newTemp()
	rTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
		fmt.Sprintf("%s, %s, %s := %s(%s(%s), %s(%s), %s, %s)", qTemp, rTemp, controlTemp, helper, cast, leftTemp, cast, rightTemp, bitsExpr, nodeName),
		fmt.Sprintf("%s := %s(%s)", quotTemp, operandType, qTemp),
		fmt.Sprintf("%s := %s(%s)", remTemp, operandType, rTemp),
	}
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, controlLines...)
	return lines, fmt.Sprintf("__able_divmod_result(%s, %s, %s)", quotVal, remVal, typeExpr), true
}

func (g *generator) compileShiftExpression(ctx *compileContext, left string, right string, operandType string, op string, nodeName string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	bitsExpr := g.bitSizeExpr(operandType)
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
	}
	if g.isUnsignedIntegerType(operandType) {
		helper := "__able_shift_left_unsigned"
		if op == ">>" {
			helper = "__able_shift_right_unsigned"
		}
		resultTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s, %s := %s(uint64(%s), uint64(%s), %s, %s)", resultTemp, controlTemp, helper, leftTemp, rightTemp, bitsExpr, nodeName))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, ""
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp)
	}
	helper := "__able_shift_left_signed"
	if op == ">>" {
		helper = "__able_shift_right_signed"
	}
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := %s(int64(%s), int64(%s), %s, %s)", resultTemp, controlTemp, helper, leftTemp, rightTemp, bitsExpr, nodeName))
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, ""
	}
	lines = append(lines, controlLines...)
	return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp)
}

func (g *generator) compileCheckedIntegerBinaryExpression(ctx *compileContext, left string, right string, operandType string, op string, nodeName string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	bitsExpr := g.bitSizeExpr(operandType)
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
	}
	if g.isUnsignedIntegerType(operandType) {
		helper := "__able_checked_add_unsigned"
		switch op {
		case "-":
			helper = "__able_checked_sub_unsigned"
		case "*":
			helper = "__able_checked_mul_unsigned"
		}
		resultTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s, %s := %s(uint64(%s), uint64(%s), %s, %s)", resultTemp, controlTemp, helper, leftTemp, rightTemp, bitsExpr, nodeName))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, ""
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp)
	}
	helper := "__able_checked_add_signed"
	switch op {
	case "-":
		helper = "__able_checked_sub_signed"
	case "*":
		helper = "__able_checked_mul_signed"
	}
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := %s(int64(%s), int64(%s), %s, %s)", resultTemp, controlTemp, helper, leftTemp, rightTemp, bitsExpr, nodeName))
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, ""
	}
	lines = append(lines, controlLines...)
	return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp)
}

func (g *generator) compilePowExpression(ctx *compileContext, left string, right string, operandType string, nodeName string) ([]string, string) {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	bitsExpr := g.bitSizeExpr(operandType)
	lines := []string{
		fmt.Sprintf("%s := %s", leftTemp, left),
		fmt.Sprintf("%s := %s", rightTemp, right),
	}
	if g.isUnsignedIntegerType(operandType) {
		resultTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s, %s := __able_pow_unsigned(uint64(%s), uint64(%s), %s, %s)", resultTemp, controlTemp, leftTemp, rightTemp, bitsExpr, nodeName))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, ""
		}
		lines = append(lines, controlLines...)
		return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp)
	}
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := __able_pow_signed(int64(%s), int64(%s), %s, %s)", resultTemp, controlTemp, leftTemp, rightTemp, bitsExpr, nodeName))
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, ""
	}
	lines = append(lines, controlLines...)
	return lines, fmt.Sprintf("%s(%s)", operandType, resultTemp)
}

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
			opLines, opExpr := g.compileCheckedIntegerBinaryExpression(ctx, leftExpr, rightExpr, leftType, "+", nodeName)
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
			opLines, opExpr := g.compileCheckedIntegerBinaryExpression(ctx, leftExpr, rightExpr, leftType, op, nodeName)
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
