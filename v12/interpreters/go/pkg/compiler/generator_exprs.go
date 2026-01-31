package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) typeMatches(expected, actual string) bool {
	if expected == "" {
		return true
	}
	return expected == actual
}

func (g *generator) compileExpr(ctx *compileContext, expr ast.Expression, expected string) (string, string, bool) {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		actual := "string"
		if !g.typeMatches(expected, actual) {
			ctx.setReason("expected string literal")
			return "", "", false
		}
		return strconv.Quote(e.Value), actual, true
	case *ast.BooleanLiteral:
		actual := "bool"
		if !g.typeMatches(expected, actual) {
			ctx.setReason("expected bool literal")
			return "", "", false
		}
		return strconv.FormatBool(e.Value), actual, true
	case *ast.IntegerLiteral:
		return g.compileIntegerLiteral(ctx, e, expected)
	case *ast.FloatLiteral:
		return g.compileFloatLiteral(ctx, e, expected)
	case *ast.CharLiteral:
		return g.compileCharLiteral(ctx, e, expected)
	case *ast.StructLiteral:
		return g.compileStructLiteral(ctx, e, expected)
	case *ast.ArrayLiteral:
		return g.compileArrayLiteral(ctx, e, expected)
	case *ast.MapLiteral:
		return g.compileMapLiteral(ctx, e, expected)
	case *ast.Identifier:
		return g.compileIdentifier(ctx, e, expected)
	case *ast.UnaryExpression:
		return g.compileUnaryExpression(ctx, e, expected)
	case *ast.BinaryExpression:
		return g.compileBinaryExpression(ctx, e, expected)
	case *ast.FunctionCall:
		return g.compileFunctionCall(ctx, e, expected)
	case *ast.MemberAccessExpression:
		return g.compileMemberAccess(ctx, e, expected)
	case *ast.IndexExpression:
		return g.compileIndexExpression(ctx, e, expected)
	default:
		ctx.setReason("unsupported expression")
		return "", "", false
	}
}

func (g *generator) compileIntegerLiteral(ctx *compileContext, lit *ast.IntegerLiteral, expected string) (string, string, bool) {
	if lit == nil || lit.Value == nil {
		ctx.setReason("missing integer literal")
		return "", "", false
	}
	actual := g.inferIntegerLiteralType(lit)
	explicit := lit.IntegerType != nil
	if expected == "" {
		expected = actual
	}
	if explicit && expected != actual {
		ctx.setReason("integer literal type mismatch")
		return "", "", false
	}
	if g.isFloatType(expected) {
		if explicit {
			ctx.setReason("integer literal type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("%s(%s)", expected, lit.Value.String()), expected, true
	}
	if !g.typeMatches(expected, actual) && !g.isIntegerType(expected) {
		ctx.setReason("unsupported integer literal type")
		return "", "", false
	}
	return fmt.Sprintf("%s(%s)", expected, lit.Value.String()), expected, true
}

func (g *generator) compileFloatLiteral(ctx *compileContext, lit *ast.FloatLiteral, expected string) (string, string, bool) {
	if lit == nil {
		ctx.setReason("missing float literal")
		return "", "", false
	}
	actual := g.inferFloatLiteralType(lit)
	explicit := lit.FloatType != nil
	if expected == "" {
		expected = actual
	}
	if explicit && expected != actual {
		ctx.setReason("float literal type mismatch")
		return "", "", false
	}
	if !g.typeMatches(expected, actual) && !g.isFloatType(expected) {
		ctx.setReason("unsupported float literal type")
		return "", "", false
	}
	return fmt.Sprintf("%s(%s)", expected, strconv.FormatFloat(lit.Value, 'g', -1, 64)), expected, true
}

func (g *generator) compileCharLiteral(ctx *compileContext, lit *ast.CharLiteral, expected string) (string, string, bool) {
	if lit == nil {
		ctx.setReason("missing char literal")
		return "", "", false
	}
	actual := "rune"
	if !g.typeMatches(expected, actual) {
		ctx.setReason("unsupported char literal type")
		return "", "", false
	}
	runes := []rune(lit.Value)
	if len(runes) != 1 {
		ctx.setReason("invalid char literal")
		return "", "", false
	}
	return fmt.Sprintf("rune(%q)", runes[0]), actual, true
}

func (g *generator) compileStructLiteral(ctx *compileContext, lit *ast.StructLiteral, expected string) (string, string, bool) {
	if lit == nil || lit.StructType == nil || lit.IsPositional || len(lit.FunctionalUpdateSources) > 0 || len(lit.TypeArguments) > 0 {
		ctx.setReason("unsupported struct literal")
		return "", "", false
	}
	info, ok := g.structs[lit.StructType.Name]
	if !ok || info == nil {
		ctx.setReason("unknown struct literal type")
		return "", "", false
	}
	if expected != "" && info.GoName != expected {
		ctx.setReason("struct literal type mismatch")
		return "", "", false
	}
	if !info.Supported {
		ctx.setReason("unsupported struct type")
		return "", "", false
	}
	fieldValues := make(map[string]string, len(lit.Fields))
	for _, field := range lit.Fields {
		if field == nil || field.Name == nil || field.Value == nil {
			ctx.setReason("unsupported struct field")
			return "", "", false
		}
		fieldName := field.Name.Name
		fieldInfo := g.fieldInfo(info, fieldName)
		if fieldInfo == nil {
			ctx.setReason("unknown struct field")
			return "", "", false
		}
		expr, _, ok := g.compileExpr(ctx, field.Value, fieldInfo.GoType)
		if !ok {
			return "", "", false
		}
		fieldValues[fieldInfo.GoName] = expr
	}
	if len(fieldValues) != len(info.Fields) {
		ctx.setReason("struct literal missing fields")
		return "", "", false
	}
	parts := make([]string, 0, len(info.Fields))
	for _, field := range info.Fields {
		value, ok := fieldValues[field.GoName]
		if !ok {
			ctx.setReason("struct literal missing field values")
			return "", "", false
		}
		parts = append(parts, fmt.Sprintf("%s: %s", field.GoName, value))
	}
	return fmt.Sprintf("%s{%s}", info.GoName, strings.Join(parts, ", ")), info.GoName, true
}

func (g *generator) compileIdentifier(ctx *compileContext, ident *ast.Identifier, expected string) (string, string, bool) {
	if ident == nil || ident.Name == "" {
		ctx.setReason("missing identifier")
		return "", "", false
	}
	param, ok := ctx.lookup(ident.Name)
	if !ok {
		ctx.setReason("unknown identifier")
		return "", "", false
	}
	if !g.typeMatches(expected, param.GoType) {
		ctx.setReason("identifier type mismatch")
		return "", "", false
	}
	return param.GoName, param.GoType, true
}

func (g *generator) compileUnaryExpression(ctx *compileContext, expr *ast.UnaryExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing unary expression")
		return "", "", false
	}
	switch expr.Operator {
	case ast.UnaryOperatorNegate:
		operand, operandType, ok := g.compileExpr(ctx, expr.Operand, expected)
		if !ok {
			return "", "", false
		}
		if !g.isNumericType(operandType) {
			ctx.setReason("unsupported unary operand type")
			return "", "", false
		}
		if !g.typeMatches(expected, operandType) {
			ctx.setReason("unary expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(-%s)", operand), operandType, true
	case ast.UnaryOperatorNot:
		if expected != "" && expected != "bool" {
			ctx.setReason("unary expression type mismatch")
			return "", "", false
		}
		operand, _, ok := g.compileExpr(ctx, expr.Operand, "bool")
		if !ok {
			return "", "", false
		}
		return fmt.Sprintf("(!%s)", operand), "bool", true
	case ast.UnaryOperatorBitNot:
		operand, operandType, ok := g.compileExpr(ctx, expr.Operand, expected)
		if !ok {
			return "", "", false
		}
		if !g.isIntegerType(operandType) {
			ctx.setReason("unsupported bitwise operand type")
			return "", "", false
		}
		if !g.typeMatches(expected, operandType) {
			ctx.setReason("unary expression type mismatch")
			return "", "", false
		}
		return fmt.Sprintf("(^%s)", operand), operandType, true
	default:
		ctx.setReason("unsupported unary operator")
		return "", "", false
	}
}

func (g *generator) compileBinaryExpression(ctx *compileContext, expr *ast.BinaryExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing binary expression")
		return "", "", false
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

func (g *generator) compileFunctionCall(ctx *compileContext, call *ast.FunctionCall, expected string) (string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return "", "", false
	}
	if len(call.TypeArguments) > 0 {
		ctx.setReason("generic calls unsupported")
		return "", "", false
	}
	callee, ok := call.Callee.(*ast.Identifier)
	if !ok || callee == nil {
		ctx.setReason("unsupported call target")
		return "", "", false
	}
	info, ok := ctx.functions[callee.Name]
	if !ok || info == nil || !info.Compileable {
		ctx.setReason("call target not compileable")
		return "", "", false
	}
	if !g.typeMatches(expected, info.ReturnType) {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	if len(call.Arguments) != len(info.Params) {
		ctx.setReason("call arity mismatch")
		return "", "", false
	}
	args := make([]string, 0, len(call.Arguments))
	for idx, arg := range call.Arguments {
		param := info.Params[idx]
		expr, _, ok := g.compileExpr(ctx, arg, param.GoType)
		if !ok {
			return "", "", false
		}
		args = append(args, expr)
	}
	return fmt.Sprintf("__able_compiled_%s(%s)", info.GoName, strings.Join(args, ", ")), info.ReturnType, true
}

func (g *generator) fieldInfo(info *structInfo, name string) *fieldInfo {
	for idx := range info.Fields {
		if info.Fields[idx].Name == name {
			return &info.Fields[idx]
		}
	}
	return nil
}

func safeParamName(name string, idx int) string {
	candidate := sanitizeIdent(name)
	if candidate == "" || candidate == "err" || candidate == "args" || candidate == "rt" || candidate == "ctx" {
		return fmt.Sprintf("p%d", idx)
	}
	return candidate
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
