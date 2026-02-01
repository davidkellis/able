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
	case *ast.StringInterpolation:
		return g.compileStringInterpolation(ctx, e, expected)
	case *ast.BooleanLiteral:
		actual := "bool"
		if !g.typeMatches(expected, actual) {
			ctx.setReason("expected bool literal")
			return "", "", false
		}
		return strconv.FormatBool(e.Value), actual, true
	case *ast.NilLiteral:
		actual := "runtime.Value"
		if expected != "" && expected != actual {
			ctx.setReason("nil literal type mismatch")
			return "", "", false
		}
		return "runtime.NilValue{}", actual, true
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
	case *ast.RangeExpression:
		return g.compileRangeExpression(ctx, e, expected)
	case *ast.TypeCastExpression:
		return g.compileTypeCast(ctx, e, expected)
	case *ast.LambdaExpression:
		return g.compileLambdaExpression(ctx, e, expected)
	case *ast.RescueExpression:
		return g.compileRescueExpression(ctx, e, expected)
	case *ast.EnsureExpression:
		return g.compileEnsureExpression(ctx, e, expected)
	case *ast.MatchExpression:
		return g.compileMatchExpression(ctx, e, expected)
	case *ast.LoopExpression:
		return g.compileLoopExpression(ctx, e, expected)
	default:
		ctx.setReason("unsupported expression")
		return "", "", false
	}
}

func (g *generator) compileStringInterpolation(ctx *compileContext, expr *ast.StringInterpolation, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing string interpolation")
		return "", "", false
	}
	actual := "string"
	if !g.typeMatches(expected, actual) {
		ctx.setReason("string interpolation type mismatch")
		return "", "", false
	}
	if len(expr.Parts) == 0 {
		return "\"\"", actual, true
	}
	lines := make([]string, 0, len(expr.Parts))
	parts := make([]string, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if part == nil {
			ctx.setReason("string interpolation missing part")
			return "", "", false
		}
		if lit, ok := part.(*ast.StringLiteral); ok {
			temp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", temp, strconv.Quote(lit.Value)))
			parts = append(parts, temp)
			continue
		}
		exprValue, exprType, ok := g.compileExpr(ctx, part, "")
		if !ok {
			return "", "", false
		}
		runtimeValue, ok := g.runtimeValueExpr(exprValue, exprType)
		if !ok {
			ctx.setReason("string interpolation part unsupported")
			return "", "", false
		}
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_stringify(%s)", temp, runtimeValue))
		parts = append(parts, temp)
	}
	concat := strings.Join(parts, " + ")
	if len(lines) == 0 {
		return concat, actual, true
	}
	return fmt.Sprintf("func() string { %s; return %s }()", strings.Join(lines, "; "), concat), actual, true
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

func (g *generator) compileFunctionCall(ctx *compileContext, call *ast.FunctionCall, expected string) (string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return "", "", false
	}
	if len(call.TypeArguments) > 0 {
		ctx.setReason("generic calls unsupported")
		return "", "", false
	}
	if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil {
		if info, ok := ctx.functions[callee.Name]; ok && info != nil && info.Compileable {
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
		if _, ok := ctx.lookup(callee.Name); !ok {
			return g.compileDynamicCall(ctx, call, expected, callee.Name)
		}
	}
	return g.compileDynamicCall(ctx, call, expected, "")
}

func (g *generator) compileDynamicCall(ctx *compileContext, call *ast.FunctionCall, expected string, calleeName string) (string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" && !g.isVoidType(expected) && g.typeCategory(expected) == "unknown" {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	lines := make([]string, 0, len(call.Arguments)+2)
	args := make([]string, 0, len(call.Arguments))
	calleeTemp := ""
	if calleeName == "" {
		switch callee := call.Callee.(type) {
		case *ast.MemberAccessExpression:
			objExpr, objType, ok := g.compileExpr(ctx, callee.Object, "")
			if !ok {
				return "", "", false
			}
			if callee.Safe && g.typeCategory(objType) == "runtime" {
				return g.compileSafeMemberCall(ctx, call, callee, expected, objExpr, objType)
			}
			objValue, ok := g.runtimeValueExpr(objExpr, objType)
			if !ok {
				ctx.setReason("method call receiver unsupported")
				return "", "", false
			}
			memberValue, ok := g.memberAssignmentRuntimeValue(ctx, callee.Member)
			if !ok {
				ctx.setReason("method call target unsupported")
				return "", "", false
			}
			objTemp := ctx.newTemp()
			memberTemp := ctx.newTemp()
			calleeTemp = ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
			lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberValue))
			lines = append(lines, fmt.Sprintf("%s := __able_member_get_method(%s, %s)", calleeTemp, objTemp, memberTemp))
		default:
			calleeExpr, calleeType, ok := g.compileExpr(ctx, call.Callee, "")
			if !ok {
				return "", "", false
			}
			calleeValue, ok := g.runtimeValueExpr(calleeExpr, calleeType)
			if !ok {
				ctx.setReason("call target unsupported")
				return "", "", false
			}
			calleeTemp = ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", calleeTemp, calleeValue))
		}
	}

	for _, arg := range call.Arguments {
		expr, goType, ok := g.compileExpr(ctx, arg, "")
		if !ok {
			return "", "", false
		}
		valueExpr, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return "", "", false
		}
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
		args = append(args, temp)
	}

	argList := strings.Join(args, ", ")
	if argList != "" {
		argList = "[]runtime.Value{" + argList + "}"
	} else {
		argList = "nil"
	}

	callExpr := ""
	if calleeName != "" {
		callExpr = fmt.Sprintf("__able_call_named(%q, %s)", calleeName, argList)
	} else {
		callExpr = fmt.Sprintf("__able_call_value(%s, %s)", calleeTemp, argList)
	}

	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
		return fmt.Sprintf("func() struct{} { %s; return struct{}{} }()", strings.Join(lines, "; ")), "struct{}", true
	}

	resultExpr := callExpr
	resultType := "runtime.Value"
	if expected != "" && expected != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(callExpr, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return "", "", false
		}
		resultExpr = converted
		resultType = expected
	}
	if len(lines) == 0 {
		return resultExpr, resultType, true
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr), resultType, true
}

func (g *generator) compileSafeMemberCall(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, expected string, objExpr string, objType string) (string, string, bool) {
	if call == nil || callee == nil {
		ctx.setReason("missing safe member call")
		return "", "", false
	}
	objValue, ok := g.runtimeValueExpr(objExpr, objType)
	if !ok {
		ctx.setReason("method call receiver unsupported")
		return "", "", false
	}
	memberValue, ok := g.memberAssignmentRuntimeValue(ctx, callee.Member)
	if !ok {
		ctx.setReason("method call target unsupported")
		return "", "", false
	}
	lines := make([]string, 0, len(call.Arguments)+4)
	args := make([]string, 0, len(call.Arguments))
	objTemp := ctx.newTemp()
	memberTemp := ctx.newTemp()
	calleeTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
	lines = append(lines, fmt.Sprintf("if __able_is_nil(%s) { return %s }", objTemp, safeNilReturnExpr(expected)))
	lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberValue))
	lines = append(lines, fmt.Sprintf("%s := __able_member_get_method(%s, %s)", calleeTemp, objTemp, memberTemp))
	for _, arg := range call.Arguments {
		expr, goType, ok := g.compileExpr(ctx, arg, "")
		if !ok {
			return "", "", false
		}
		valueExpr, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return "", "", false
		}
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
		args = append(args, temp)
	}
	argList := strings.Join(args, ", ")
	if argList != "" {
		argList = "[]runtime.Value{" + argList + "}"
	} else {
		argList = "nil"
	}
	callExpr := fmt.Sprintf("__able_call_value(%s, %s)", calleeTemp, argList)
	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
		lines = append(lines, "return struct{}{}")
		return fmt.Sprintf("func() struct{} { %s }()", strings.Join(lines, "; ")), "struct{}", true
	}
	lines = append(lines, fmt.Sprintf("return %s", callExpr))
	baseExpr := fmt.Sprintf("func() runtime.Value { %s }()", strings.Join(lines, "; "))
	if expected == "" || expected == "runtime.Value" {
		return baseExpr, "runtime.Value", true
	}
	converted, ok := g.expectRuntimeValueExpr(baseExpr, expected)
	if !ok {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	return converted, expected, true
}

func safeNilReturnExpr(expected string) string {
	if expected == "struct{}" {
		return "struct{}{}"
	}
	return "runtime.NilValue{}"
}

func (g *generator) compileTypeCast(ctx *compileContext, expr *ast.TypeCastExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil || expr.TargetType == nil {
		ctx.setReason("missing type cast")
		return "", "", false
	}
	valueExpr, valueType, ok := g.compileExpr(ctx, expr.Expression, "")
	if !ok {
		return "", "", false
	}
	valueRuntime, ok := g.runtimeValueExpr(valueExpr, valueType)
	if !ok {
		ctx.setReason("cast operand unsupported")
		return "", "", false
	}
	targetExpr, ok := g.renderTypeExpression(expr.TargetType)
	if !ok {
		ctx.setReason("unsupported cast type")
		return "", "", false
	}
	castExpr := fmt.Sprintf("__able_cast(%s, %s)", valueRuntime, targetExpr)
	if expected == "runtime.Value" {
		return castExpr, "runtime.Value", true
	}
	desiredType := "runtime.Value"
	if expected != "" && expected != "runtime.Value" {
		desiredType = expected
	} else if mapped, ok := g.mapTypeExpression(expr.TargetType); ok && mapped != "" {
		desiredType = mapped
	}
	if desiredType == "struct{}" {
		ctx.setReason("cast to void unsupported")
		return "", "", false
	}
	if desiredType == "runtime.Value" {
		return castExpr, "runtime.Value", true
	}
	converted, ok := g.expectRuntimeValueExpr(castExpr, desiredType)
	if !ok {
		ctx.setReason("cast type mismatch")
		return "", "", false
	}
	return converted, desiredType, true
}

func (g *generator) compileRangeExpression(ctx *compileContext, expr *ast.RangeExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Start == nil || expr.End == nil {
		ctx.setReason("missing range expression")
		return "", "", false
	}
	startExpr, startType, ok := g.compileExpr(ctx, expr.Start, "")
	if !ok {
		return "", "", false
	}
	endExpr, endType, ok := g.compileExpr(ctx, expr.End, "")
	if !ok {
		return "", "", false
	}
	startRuntime, ok := g.runtimeValueExpr(startExpr, startType)
	if !ok {
		ctx.setReason("range start unsupported")
		return "", "", false
	}
	endRuntime, ok := g.runtimeValueExpr(endExpr, endType)
	if !ok {
		ctx.setReason("range end unsupported")
		return "", "", false
	}
	startTemp := ctx.newTemp()
	endTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := %s", startTemp, startRuntime),
		fmt.Sprintf("%s := %s", endTemp, endRuntime),
	}
	rangeExpr := fmt.Sprintf("__able_range(%s, %s, %t)", startTemp, endTemp, expr.Inclusive)
	resultType := "runtime.Value"
	resultExpr := rangeExpr
	if expected != "" && expected != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(rangeExpr, expected)
		if !ok {
			ctx.setReason("range expression type mismatch")
			return "", "", false
		}
		resultExpr = converted
		resultType = expected
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr), resultType, true
}

func (g *generator) compileLambdaExpression(ctx *compileContext, expr *ast.LambdaExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing lambda expression")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" {
		ctx.setReason("lambda expression type mismatch")
		return "", "", false
	}
	if len(expr.GenericParams) > 0 || len(expr.WhereClause) > 0 {
		ctx.setReason("lambda generics unsupported")
		return "", "", false
	}
	if expr.Body == nil {
		ctx.setReason("missing lambda body")
		return "", "", false
	}

	lambdaCtx := ctx.child()
	if lambdaCtx == nil {
		ctx.setReason("missing lambda context")
		return "", "", false
	}
	lambdaCtx.loopDepth = 0

	params := make([]paramInfo, 0, len(expr.Params))
	for idx, param := range expr.Params {
		if param == nil {
			ctx.setReason("missing lambda parameter")
			return "", "", false
		}
		ident, ok := param.Name.(*ast.Identifier)
		if !ok || ident == nil || ident.Name == "" {
			ctx.setReason("unsupported lambda parameter")
			return "", "", false
		}
		goName := safeParamName(ident.Name, idx)
		goType := "runtime.Value"
		if param.ParamType != nil {
			mapped, ok := g.mapTypeExpression(param.ParamType)
			if !ok {
				ctx.setReason("unsupported lambda parameter type")
				return "", "", false
			}
			goType = mapped
		}
		info := paramInfo{Name: ident.Name, GoName: goName, GoType: goType}
		lambdaCtx.locals[ident.Name] = info
		params = append(params, info)
	}

	desiredReturn := ""
	if expr.ReturnType != nil {
		mapped, ok := g.mapTypeExpression(expr.ReturnType)
		if !ok {
			ctx.setReason("unsupported lambda return type")
			return "", "", false
		}
		desiredReturn = mapped
	}

	var bodyLines []string
	var bodyExpr string
	var bodyType string
	var ok bool
	if expr.IsVerboseSyntax {
		block, isBlock := expr.Body.(*ast.BlockExpression)
		if !isBlock || block == nil {
			ctx.setReason("verbose lambda requires block body")
			return "", "", false
		}
		bodyLines, bodyExpr, bodyType, ok = g.compileLambdaBlockBody(lambdaCtx, desiredReturn, block)
	} else {
		bodyLines, bodyExpr, bodyType, ok = g.compileTailExpression(lambdaCtx, desiredReturn, expr.Body)
	}
	if !ok {
		return "", "", false
	}
	if desiredReturn != "" && !g.typeMatches(desiredReturn, bodyType) {
		ctx.setReason("lambda return type mismatch")
		return "", "", false
	}

	implLines := make([]string, 0, len(bodyLines)+len(params)*4+4)
	implLines = append(implLines, fmt.Sprintf("if len(args) != %d { return nil, fmt.Errorf(\"lambda expects %d arguments, got %%d\", len(args)) }", len(params), len(params)))
	for idx, param := range params {
		argVar := fmt.Sprintf("__able_lambda_arg_%d", idx)
		valueVar := argVar + "_value"
		implLines = append(implLines, fmt.Sprintf("%s := args[%d]", valueVar, idx))
		convLines, ok := g.lambdaArgConversionLines(argVar, valueVar, param.GoType, param.GoName)
		if !ok {
			ctx.setReason("unsupported lambda parameter type")
			return "", "", false
		}
		implLines = append(implLines, convLines...)
	}

	implLines = append(implLines, bodyLines...)
	if g.isVoidType(bodyType) {
		if bodyExpr != "" {
			implLines = append(implLines, fmt.Sprintf("_ = %s", bodyExpr))
		}
		implLines = append(implLines, "return runtime.VoidValue{}, nil")
	} else {
		resultTemp := lambdaCtx.newTemp()
		implLines = append(implLines, fmt.Sprintf("%s := %s", resultTemp, bodyExpr))
		returnLines, ok := g.lambdaReturnLines(resultTemp, bodyType)
		if !ok {
			ctx.setReason("unsupported lambda return type")
			return "", "", false
		}
		implLines = append(implLines, returnLines...)
	}

	implBody := strings.Join(implLines, "; ")
	lambdaExpr := fmt.Sprintf("runtime.NativeFunctionValue{Name: %q, Arity: %d, Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) { %s }}", "<lambda>", len(params), implBody)
	return lambdaExpr, "runtime.Value", true
}

func (g *generator) compileLambdaBlockBody(ctx *compileContext, returnType string, block *ast.BlockExpression) ([]string, string, string, bool) {
	if block == nil {
		ctx.setReason("missing lambda body")
		return nil, "", "", false
	}
	statements := block.Body
	if len(statements) == 0 {
		if returnType == "" || g.isVoidType(returnType) {
			return nil, "struct{}{}", "struct{}", true
		}
		ctx.setReason("empty lambda body requires void return")
		return nil, "", "", false
	}
	lines := make([]string, 0, len(statements))
	for idx, stmt := range statements {
		isLast := idx == len(statements)-1
		if ret, ok := stmt.(*ast.ReturnStatement); ok {
			if !isLast {
				ctx.setReason("return must be final statement")
				return nil, "", "", false
			}
			if ret.Argument == nil {
				if returnType != "" && !g.isVoidType(returnType) {
					ctx.setReason("missing return value")
					return nil, "", "", false
				}
				return lines, "struct{}{}", "struct{}", true
			}
			stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, returnType, ret.Argument)
			if !ok {
				return nil, "", "", false
			}
			if returnType != "" && !g.typeMatches(returnType, valueType) {
				ctx.setReason("lambda return type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, stmtLines...)
			finalType := valueType
			if returnType != "" {
				finalType = returnType
			}
			return lines, valueExpr, finalType, true
		}
		if isLast {
			if expr, ok := stmt.(ast.Expression); ok && expr != nil {
				stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, returnType, expr)
				if !ok {
					return nil, "", "", false
				}
				if returnType != "" && !g.typeMatches(returnType, valueType) {
					ctx.setReason("lambda return type mismatch")
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
				if g.isVoidType(returnType) {
					if valueExpr != "" {
						lines = append(lines, fmt.Sprintf("_ = %s", valueExpr))
					}
					return lines, "struct{}{}", "struct{}", true
				}
				finalType := valueType
				if returnType != "" {
					finalType = returnType
				}
				return lines, valueExpr, finalType, true
			}
			if returnType == "" || g.isVoidType(returnType) {
				stmtLines, ok := g.compileStatement(ctx, stmt)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
				return lines, "struct{}{}", "struct{}", true
			}
			ctx.setReason("missing return statement")
			return nil, "", "", false
		}
		stmtLines, ok := g.compileStatement(ctx, stmt)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, stmtLines...)
	}
	ctx.setReason("missing return statement")
	return nil, "", "", false
}

func (g *generator) lambdaArgConversionLines(argVar string, valueVar string, goType string, target string) ([]string, bool) {
	switch g.typeCategory(goType) {
	case "runtime":
		return []string{fmt.Sprintf("%s := %s", target, valueVar)}, true
	case "bool":
		return []string{
			fmt.Sprintf("%s, err := bridge.AsBool(%s)", target, valueVar),
			"if err != nil { return nil, err }",
		}, true
	case "string":
		return []string{
			fmt.Sprintf("%s, err := bridge.AsString(%s)", target, valueVar),
			"if err != nil { return nil, err }",
		}, true
	case "rune":
		return []string{
			fmt.Sprintf("%s, err := bridge.AsRune(%s)", target, valueVar),
			"if err != nil { return nil, err }",
		}, true
	case "float32":
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsFloat(%s)", raw, valueVar),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := float32(%s)", target, raw),
		}, true
	case "float64":
		return []string{
			fmt.Sprintf("%s, err := bridge.AsFloat(%s)", target, valueVar),
			"if err != nil { return nil, err }",
		}, true
	case "int":
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsInt(%s, bridge.NativeIntBits)", raw, valueVar),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := int(%s)", target, raw),
		}, true
	case "uint":
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsUint(%s, bridge.NativeIntBits)", raw, valueVar),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := uint(%s)", target, raw),
		}, true
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(goType)
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsInt(%s, %d)", raw, valueVar, bits),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := %s(%s)", target, goType, raw),
		}, true
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(goType)
		raw := argVar + "_raw"
		return []string{
			fmt.Sprintf("%s, err := bridge.AsUint(%s, %d)", raw, valueVar, bits),
			"if err != nil { return nil, err }",
			fmt.Sprintf("%s := %s(%s)", target, goType, raw),
		}, true
	case "struct":
		return []string{
			fmt.Sprintf("%s, err := __able_struct_%s_from(%s)", target, goType, valueVar),
			"if err != nil { return nil, err }",
		}, true
	default:
		return []string{fmt.Sprintf("%s := %s", target, valueVar)}, true
	}
}

func (g *generator) lambdaReturnLines(resultName string, goType string) ([]string, bool) {
	switch g.typeCategory(goType) {
	case "runtime":
		return []string{fmt.Sprintf("return %s, nil", resultName)}, true
	case "void":
		return []string{
			fmt.Sprintf("_ = %s", resultName),
			"return runtime.VoidValue{}, nil",
		}, true
	case "bool":
		return []string{fmt.Sprintf("return bridge.ToBool(%s), nil", resultName)}, true
	case "string":
		return []string{fmt.Sprintf("return bridge.ToString(%s), nil", resultName)}, true
	case "rune":
		return []string{fmt.Sprintf("return bridge.ToRune(%s), nil", resultName)}, true
	case "float32":
		return []string{fmt.Sprintf("return bridge.ToFloat32(%s), nil", resultName)}, true
	case "float64":
		return []string{fmt.Sprintf("return bridge.ToFloat64(%s), nil", resultName)}, true
	case "int":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\")), nil", resultName)}, true
	case "uint":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\")), nil", resultName)}, true
	case "int8":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\")), nil", resultName)}, true
	case "int16":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\")), nil", resultName)}, true
	case "int32":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\")), nil", resultName)}, true
	case "int64":
		return []string{fmt.Sprintf("return bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\")), nil", resultName)}, true
	case "uint8":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\")), nil", resultName)}, true
	case "uint16":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\")), nil", resultName)}, true
	case "uint32":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\")), nil", resultName)}, true
	case "uint64":
		return []string{fmt.Sprintf("return bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\")), nil", resultName)}, true
	case "struct":
		return []string{fmt.Sprintf("return __able_struct_%s_to(__able_runtime, %s)", goType, resultName)}, true
	default:
		return []string{fmt.Sprintf("return %s, nil", resultName)}, true
	}
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
