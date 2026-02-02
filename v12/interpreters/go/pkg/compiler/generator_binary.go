package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileBinaryExpression(ctx *compileContext, expr *ast.BinaryExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing binary expression")
		return "", "", false
	}
	if expr.Operator == "|>" || expr.Operator == "|>>" {
		return g.compilePipeExpression(ctx, expr, expected)
	}
	if expr.Operator != "&&" && expr.Operator != "||" {
		leftExpr, leftType, ok := g.compileExpr(ctx, expr.Left, "")
		if !ok {
			return "", "", false
		}
		rightExpr, rightType, ok := g.compileExpr(ctx, expr.Right, "")
		if !ok {
			return "", "", false
		}
		if leftType == "runtime.Value" || rightType == "runtime.Value" {
			return g.compileRuntimeBinaryOperation(ctx, expr.Operator, leftExpr, leftType, rightExpr, rightType, expected)
		}
	}
	switch expr.Operator {
	case "==", "!=", "<", "<=", ">", ">=":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType {
			ctx.setReason("comparison operand type mismatch")
			return "", "", false
		}
		if expr.Operator == "==" || expr.Operator == "!=" {
			if !g.isEqualityComparable(leftType) {
				ctx.setReason("unsupported comparison type")
				return "", "", false
			}
		} else if !g.isOrderedComparable(leftType) {
			ctx.setReason("unsupported comparison type")
			return "", "", false
		}
		if expected != "" && expected != "bool" {
			ctx.setReason("comparison expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(%s %s %s)", left, expr.Operator, right), "bool", true
	case "&&", "||":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType || leftType != "bool" {
			ctx.setReason("logical operator requires bool operands")
			return "", "", false
		}
		if expected != "" && expected != "bool" {
			ctx.setReason("logical expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(%s %s %s)", left, expr.Operator, right), "bool", true
	case ".&", ".|", ".^", "&", "|", "^":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return "", "", false
		}
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported bitwise operator type")
			return "", "", false
		}
		if !g.typeMatches(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
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
		return fmt.Sprintf("(%s %s %s)", left, op, right), leftType, true
	case ".<<", ".>>", "<<", ">>":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return "", "", false
		}
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported shift operator type")
			return "", "", false
		}
		if !g.typeMatches(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		op := expr.Operator
		if op == ".<<" {
			op = "<<"
		} else if op == ".>>" {
			op = ">>"
		}
		expr := g.compileShiftExpression(ctx, left, right, leftType, op)
		return expr, leftType, true
	case "+":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return "", "", false
		}
		resultType := leftType
		if !g.isStringType(resultType) && !g.isNumericType(resultType) {
			ctx.setReason("unsupported + operand type")
			return "", "", false
		}
		if !g.typeMatches(expected, resultType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(%s %s %s)", left, expr.Operator, right), resultType, true
	case "-", "*":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return "", "", false
		}
		resultType := leftType
		if !g.isNumericType(resultType) {
			ctx.setReason("unsupported numeric operator type")
			return "", "", false
		}
		if !g.typeMatches(expected, resultType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(%s %s %s)", left, expr.Operator, right), resultType, true
	case "/":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return "", "", false
		}
		if !g.isNumericType(leftType) {
			ctx.setReason("unsupported numeric operator type")
			return "", "", false
		}
		resultType := leftType
		if g.isIntegerType(resultType) {
			resultType = "float64"
		}
		if !g.typeMatches(expected, resultType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		expr := g.compileDivisionExpression(ctx, left, right, leftType, resultType)
		return expr, resultType, true
	case "//", "%":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return "", "", false
		}
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported integer operator type")
			return "", "", false
		}
		if !g.typeMatches(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		expr := g.compileDivModExpression(ctx, left, right, leftType, expr.Operator)
		return expr, leftType, true
	case "/%":
		left, leftType, right, rightType, ok := g.compileBinaryOperands(ctx, expr.Left, expr.Right)
		if !ok {
			return "", "", false
		}
		if leftType != rightType {
			ctx.setReason("binary operand type mismatch")
			return "", "", false
		}
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported integer operator type")
			return "", "", false
		}
		if expected != "" && expected != "runtime.Value" {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		expr := g.compileDivModResultExpression(ctx, left, right, leftType)
		if expr == "" {
			ctx.setReason("unsupported /% operands")
			return "", "", false
		}
		return expr, "runtime.Value", true
	default:
		ctx.setReason("unsupported operator")
		return "", "", false
	}
}

func (g *generator) compilePipeExpression(ctx *compileContext, expr *ast.BinaryExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing pipe expression")
		return "", "", false
	}
	leftExpr, leftType, ok := g.compileExpr(ctx, expr.Left, "")
	if !ok {
		return "", "", false
	}
	subjectValue, ok := g.runtimeValueExpr(leftExpr, leftType)
	if !ok {
		ctx.setReason("pipe subject unsupported")
		return "", "", false
	}
	subjectTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("%s := %s", subjectTemp, subjectValue)}

	pipeCtx := ctx.child()
	subjectParam := paramInfo{Name: subjectTemp, GoName: subjectTemp, GoType: "runtime.Value"}
	pipeCtx.locals[subjectTemp] = subjectParam
	pipeCtx.implicitReceiver = subjectParam
	pipeCtx.hasImplicitReceiver = true

	if placeholderExpr, _, ok := g.compilePlaceholderLambda(pipeCtx, expr.Right); ok {
		rhsTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", rhsTemp, placeholderExpr))
		callTemp := ctx.newTemp()
		argsTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("var %s []runtime.Value", argsTemp))
		lines = append(lines, fmt.Sprintf("switch %s.(type) {", rhsTemp))
		lines = append(lines, "case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue:")
		lines = append(lines, fmt.Sprintf("\t%s = nil", argsTemp))
		lines = append(lines, "default:")
		lines = append(lines, fmt.Sprintf("\t%s = []runtime.Value{%s}", argsTemp, subjectTemp))
		lines = append(lines, "}")
		lines = append(lines, fmt.Sprintf("%s := __able_call_value(%s, %s)", callTemp, rhsTemp, argsTemp))
		return g.pipeResultExpression(ctx, expected, lines, callTemp)
	}

	if call, ok := expr.Right.(*ast.FunctionCall); ok {
		calleeExpr, calleeType, ok := g.compileExpr(pipeCtx, call.Callee, "")
		if !ok {
			return "", "", false
		}
		calleeValue, ok := g.runtimeValueExpr(calleeExpr, calleeType)
		if !ok {
			ctx.setReason("pipe call target unsupported")
			return "", "", false
		}
		calleeTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", calleeTemp, calleeValue))
		argTemps := make([]string, 0, len(call.Arguments))
		for _, arg := range call.Arguments {
			argExpr, argType, ok := g.compileExpr(pipeCtx, arg, "")
			if !ok {
				return "", "", false
			}
			argValue, ok := g.runtimeValueExpr(argExpr, argType)
			if !ok {
				ctx.setReason("pipe call argument unsupported")
				return "", "", false
			}
			argTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", argTemp, argValue))
			argTemps = append(argTemps, argTemp)
		}
		argsTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("var %s []runtime.Value", argsTemp))
		lines = append(lines, fmt.Sprintf("switch %s.(type) {", calleeTemp))
		lines = append(lines, "case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue:")
		lines = append(lines, fmt.Sprintf("\t%s = %s", argsTemp, runtimeValueSlice(argTemps)))
		lines = append(lines, "default:")
		lines = append(lines, fmt.Sprintf("\t%s = %s", argsTemp, runtimeValueSliceWithSubject(subjectTemp, argTemps)))
		lines = append(lines, "}")
		callTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_call_value(%s, %s)", callTemp, calleeTemp, argsTemp))
		return g.pipeResultExpression(ctx, expected, lines, callTemp)
	}

	rhsExpr, rhsType, ok := g.compileExpr(pipeCtx, expr.Right, "")
	if !ok {
		return "", "", false
	}
	rhsValue, ok := g.runtimeValueExpr(rhsExpr, rhsType)
	if !ok {
		ctx.setReason("pipe rhs unsupported")
		return "", "", false
	}
	rhsTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", rhsTemp, rhsValue))
	argsTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("var %s []runtime.Value", argsTemp))
	lines = append(lines, fmt.Sprintf("switch %s.(type) {", rhsTemp))
	lines = append(lines, "case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue:")
	lines = append(lines, fmt.Sprintf("\t%s = nil", argsTemp))
	lines = append(lines, "default:")
	lines = append(lines, fmt.Sprintf("\t%s = []runtime.Value{%s}", argsTemp, subjectTemp))
	lines = append(lines, "}")
	callTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := __able_call_value(%s, %s)", callTemp, rhsTemp, argsTemp))
	return g.pipeResultExpression(ctx, expected, lines, callTemp)
}

func (g *generator) pipeResultExpression(ctx *compileContext, expected string, lines []string, resultTemp string) (string, string, bool) {
	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		return fmt.Sprintf("func() struct{} { %s; return struct{}{} }()", strings.Join(lines, "; ")), "struct{}", true
	}
	resultType := "runtime.Value"
	resultExpr := resultTemp
	if expected != "" && expected != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(resultTemp, expected)
		if !ok {
			ctx.setReason("pipe result type mismatch")
			return "", "", false
		}
		resultType = expected
		resultExpr = converted
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr), resultType, true
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

func (g *generator) compileRuntimeBinaryOperation(ctx *compileContext, op string, leftExpr string, leftType string, rightExpr string, rightType string, expected string) (string, string, bool) {
	leftVal, ok := g.runtimeValueExpr(leftExpr, leftType)
	if !ok {
		ctx.setReason("binary operand unsupported")
		return "", "", false
	}
	rightVal, ok := g.runtimeValueExpr(rightExpr, rightType)
	if !ok {
		ctx.setReason("binary operand unsupported")
		return "", "", false
	}
	expr := fmt.Sprintf("__able_binary_op(%q, %s, %s)", op, leftVal, rightVal)
	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		if expected != "" && expected != "bool" {
			ctx.setReason("comparison expression type mismatch")
			return "", "", false
		}
		converted, ok := g.expectRuntimeValueExpr(expr, "bool")
		if !ok {
			ctx.setReason("comparison expression type mismatch")
			return "", "", false
		}
		return converted, "bool", true
	}
	if expected != "" && expected != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(expr, expected)
		if !ok {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		return converted, expected, true
	}
	return expr, "runtime.Value", true
}

func (g *generator) compileBinaryOperands(ctx *compileContext, leftExpr ast.Expression, rightExpr ast.Expression) (string, string, string, string, bool) {
	if g.isUntypedNumericLiteral(leftExpr) && !g.isUntypedNumericLiteral(rightExpr) {
		right, rightType, ok := g.compileExpr(ctx, rightExpr, "")
		if !ok {
			return "", "", "", "", false
		}
		left, leftType, ok := g.compileExpr(ctx, leftExpr, rightType)
		if !ok {
			return "", "", "", "", false
		}
		return left, leftType, right, rightType, true
	}
	left, leftType, ok := g.compileExpr(ctx, leftExpr, "")
	if !ok {
		return "", "", "", "", false
	}
	right, rightType, ok := g.compileExpr(ctx, rightExpr, leftType)
	if !ok {
		return "", "", "", "", false
	}
	return left, leftType, right, rightType, true
}

func (g *generator) bitSizeExpr(goType string) string {
	switch goType {
	case "int", "uint":
		return "bridge.NativeIntBits"
	default:
		return fmt.Sprintf("%d", g.intBits(goType))
	}
}

func (g *generator) compileDivisionExpression(ctx *compileContext, left string, right string, operandType string, resultType string) string {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	if g.isIntegerType(operandType) {
		return fmt.Sprintf("func() %s { %s := %s; %s := %s; if %s == 0 { __able_raise_division_by_zero() }; return float64(%s) / float64(%s) }()", resultType, leftTemp, left, rightTemp, right, rightTemp, leftTemp, rightTemp)
	}
	return fmt.Sprintf("func() %s { %s := %s; %s := %s; if %s == 0 { __able_raise_division_by_zero() }; return %s / %s }()", resultType, leftTemp, left, rightTemp, right, rightTemp, leftTemp, rightTemp)
}

func (g *generator) compileDivModExpression(ctx *compileContext, left string, right string, operandType string, op string) string {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	helper := "__able_divmod_signed"
	cast := "int64"
	if g.isUnsignedIntegerType(operandType) {
		helper = "__able_divmod_unsigned"
		cast = "uint64"
	}
	if op == "//" {
		return fmt.Sprintf("func() %s { %s := %s; %s := %s; q, _ := %s(%s(%s), %s(%s)); return %s(q) }()", operandType, leftTemp, left, rightTemp, right, helper, cast, leftTemp, cast, rightTemp, operandType)
	}
	return fmt.Sprintf("func() %s { %s := %s; %s := %s; _, r := %s(%s(%s), %s(%s)); return %s(r) }()", operandType, leftTemp, left, rightTemp, right, helper, cast, leftTemp, cast, rightTemp, operandType)
}

func (g *generator) compileDivModResultExpression(ctx *compileContext, left string, right string, operandType string) string {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	leftVal, ok := g.runtimeValueExpr(left, operandType)
	if !ok {
		return ""
	}
	rightVal, ok := g.runtimeValueExpr(right, operandType)
	if !ok {
		return ""
	}
	return fmt.Sprintf("func() runtime.Value { %s := %s; %s := %s; return __able_binary_op(\"/%%\", %s, %s) }()", leftTemp, leftVal, rightTemp, rightVal, leftTemp, rightTemp)
}

func (g *generator) compileShiftExpression(ctx *compileContext, left string, right string, operandType string, op string) string {
	leftTemp := ctx.newTemp()
	rightTemp := ctx.newTemp()
	bitsExpr := g.bitSizeExpr(operandType)
	if g.isUnsignedIntegerType(operandType) {
		helper := "__able_shift_left_unsigned"
		if op == ">>" {
			helper = "__able_shift_right_unsigned"
		}
		return fmt.Sprintf("func() %s { %s := %s; %s := %s; return %s(%s(uint64(%s), uint64(%s), %s)) }()", operandType, leftTemp, left, rightTemp, right, operandType, helper, leftTemp, rightTemp, bitsExpr)
	}
	helper := "__able_shift_left_signed"
	if op == ">>" {
		helper = "__able_shift_right_signed"
	}
	return fmt.Sprintf("func() %s { %s := %s; %s := %s; return %s(%s(int64(%s), int64(%s), %s)) }()", operandType, leftTemp, left, rightTemp, right, operandType, helper, leftTemp, rightTemp, bitsExpr)
}

func (g *generator) compileBinaryOperation(ctx *compileContext, op string, leftExpr string, leftType string, rightExpr string, rightType string, expected string) (string, string, bool) {
	if leftType == "runtime.Value" || rightType == "runtime.Value" {
		return g.compileRuntimeBinaryOperation(ctx, op, leftExpr, leftType, rightExpr, rightType, expected)
	}
	if leftType != rightType {
		ctx.setReason("binary operand type mismatch")
		return "", "", false
	}
	switch op {
	case "+":
		if !g.isStringType(leftType) && !g.isNumericType(leftType) {
			ctx.setReason("unsupported + operand type")
			return "", "", false
		}
		if !g.typeMatches(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(%s + %s)", leftExpr, rightExpr), leftType, true
	case "-", "*":
		if !g.isNumericType(leftType) {
			ctx.setReason("unsupported numeric operator type")
			return "", "", false
		}
		if !g.typeMatches(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(%s %s %s)", leftExpr, op, rightExpr), leftType, true
	case "/":
		if !g.isNumericType(leftType) {
			ctx.setReason("unsupported numeric operator type")
			return "", "", false
		}
		resultType := leftType
		if g.isIntegerType(resultType) {
			resultType = "float64"
		}
		if !g.typeMatches(expected, resultType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		expr := g.compileDivisionExpression(ctx, leftExpr, rightExpr, leftType, resultType)
		return expr, resultType, true
	case "//", "%":
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported integer operator type")
			return "", "", false
		}
		if !g.typeMatches(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		expr := g.compileDivModExpression(ctx, leftExpr, rightExpr, leftType, op)
		return expr, leftType, true
	case ".&", ".|", ".^", "&", "|", "^":
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported bitwise operator type")
			return "", "", false
		}
		if !g.typeMatches(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		switch op {
		case ".&":
			op = "&"
		case ".|":
			op = "|"
		case ".^":
			op = "^"
		}
		return fmt.Sprintf("(%s %s %s)", leftExpr, op, rightExpr), leftType, true
	case ".<<", ".>>", "<<", ">>":
		if !g.isIntegerType(leftType) {
			ctx.setReason("unsupported shift operator type")
			return "", "", false
		}
		if !g.typeMatches(expected, leftType) {
			ctx.setReason("binary expression type mismatch")
			return "", "", false
		}
		if op == ".<<" {
			op = "<<"
		} else if op == ".>>" {
			op = ">>"
		}
		expr := g.compileShiftExpression(ctx, leftExpr, rightExpr, leftType, op)
		return expr, leftType, true
	default:
		ctx.setReason("unsupported operator")
		return "", "", false
	}
}
