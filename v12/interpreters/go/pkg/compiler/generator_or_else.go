package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compilePropagationExpression(ctx *compileContext, expr *ast.PropagationExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Expression == nil {
		ctx.setReason("missing propagation expression")
		return nil, "", "", false
	}
	if indexExpr, ok := expr.Expression.(*ast.IndexExpression); ok {
		if lines, fastExpr, fastType, fastOK := g.compilePropagationMonoArrayIndex(ctx, indexExpr, expected); fastOK {
			return lines, fastExpr, fastType, true
		}
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr.Expression)
	if !ok {
		return nil, "", "", false
	}
	resultType := expected
	if resultType == "" {
		resultType = valueType
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}
	if valueType != "runtime.Value" {
		if !g.typeMatches(resultType, valueType) {
			ctx.setReason("propagation type mismatch")
			return nil, "", "", false
		}
		return valueLines, valueExpr, valueType, true
	}
	valueTemp := ctx.newTemp()
	lines := append([]string{}, valueLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
	lines = append(lines, fmt.Sprintf("if __able_is_error(%s) { bridge.Raise(__able_error_value(%s)) }", valueTemp, valueTemp))
	resultExpr := valueTemp
	if resultType != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueTemp, resultType)
		if !ok {
			ctx.setReason("propagation type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		resultExpr = converted
	}
	return lines, resultExpr, resultType, true
}

func (g *generator) compilePropagationMonoArrayIndex(ctx *compileContext, expr *ast.IndexExpression, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || expr == nil {
		return nil, "", "", false
	}
	objExpr, objType, ok := g.compileExpr(ctx, expr.Object, "")
	if !ok || !g.isArrayStructType(objType) {
		return nil, "", "", false
	}
	monoKind, monoEnabled := g.monoArrayKindForObject(ctx, expr.Object, objType)
	if !monoEnabled {
		return nil, "", "", false
	}
	idxExpr, idxType, ok := g.compileExpr(ctx, expr.Index, "")
	if !ok {
		return nil, "", "", false
	}
	objTemp := ctx.newTemp()
	idxTemp := ctx.newTemp()
	indexTemp := ctx.newTemp()
	handleRawTemp := ctx.newTemp()
	handleTemp := ctx.newTemp()
	lengthTemp := ctx.newTemp()
	nativeTemp := ctx.newTemp()
	readExpr, readType, ok := g.monoArrayReadExpr(monoKind, handleTemp, indexTemp)
	if !ok {
		return nil, "", "", false
	}
	resultType := expected
	if resultType == "" {
		resultType = readType
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}
	lines := []string{
		fmt.Sprintf("%s := %s", objTemp, objExpr),
	}
	lines, ok = g.appendIndexIntLines(ctx, lines, idxExpr, idxType, idxTemp, indexTemp)
	if !ok {
		return nil, "", "", false
	}
	handleErrTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
		fmt.Sprintf("%s, %s := bridge.AsInt(%s, 64)", handleTemp, handleErrTemp, handleRawTemp),
		fmt.Sprintf("if %s != nil { panic(%s) }", handleErrTemp, handleErrTemp),
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("if %s < 0 || %s >= %s { bridge.Raise(__able_error_value(__able_index_error(%s, %s))) }", indexTemp, indexTemp, lengthTemp, indexTemp, lengthTemp),
		fmt.Sprintf("%s, err := %s", nativeTemp, readExpr),
		"if err != nil { panic(err) }",
	)
	switch {
	case resultType == readType:
		return lines, nativeTemp, readType, true
	case resultType == "runtime.Value":
		rtConvLines, runtimeExpr, ok := g.runtimeValueLines(ctx, nativeTemp, readType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, rtConvLines...)
		return lines, runtimeExpr, "runtime.Value", true
	default:
		if widenedExpr, ok := g.nativeIntegerWidenExpr(nativeTemp, readType, resultType); ok {
			return lines, widenedExpr, resultType, true
		}
		rtConvLines, runtimeExpr, ok := g.runtimeValueLines(ctx, nativeTemp, readType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, rtConvLines...)
		convLines, convertedExpr, ok := g.expectRuntimeValueExprLines(ctx, runtimeExpr, resultType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, convertedExpr, resultType, true
	}
}

func (g *generator) compileOrElseExpression(ctx *compileContext, expr *ast.OrElseExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Expression == nil || expr.Handler == nil {
		ctx.setReason("missing or-else expression")
		return nil, "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr.Expression)
	if !ok {
		return nil, "", "", false
	}
	handlerCtx := ctx.child()
	bindingName := ""
	if expr.ErrorBinding != nil && expr.ErrorBinding.Name != "" {
		bindingName = expr.ErrorBinding.Name
		handlerCtx.locals[bindingName] = paramInfo{Name: bindingName, GoName: sanitizeIdent(bindingName), GoType: "runtime.Value"}
	}

	preferredType := expected
	if preferredType == "" && valueType != "runtime.Value" {
		preferredType = valueType
	}
	handlerLines, handlerExpr, handlerType, ok := g.compileBlockExpression(handlerCtx, expr.Handler, preferredType)
	if !ok {
		return nil, "", "", false
	}

	resultType := expected
	if resultType == "" {
		switch {
		case valueType == "" && handlerType == "":
			resultType = "runtime.Value"
		case valueType == "":
			resultType = handlerType
		case handlerType == "":
			resultType = valueType
		case valueType == handlerType:
			resultType = valueType
		default:
			resultType = "runtime.Value"
		}
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}

	handlerResultExpr := handlerExpr
	var handlerCoerceLines []string
	switch {
	case handlerType == resultType:
	case handlerType == "runtime.Value" && resultType != "runtime.Value":
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, handlerExpr, resultType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return nil, "", "", false
		}
		handlerCoerceLines = convLines
		handlerResultExpr = converted
	case resultType == "runtime.Value" && handlerType != "runtime.Value":
		convLines, converted, ok := g.runtimeValueLines(ctx, handlerExpr, handlerType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return nil, "", "", false
		}
		handlerCoerceLines = convLines
		handlerResultExpr = converted
	default:
		ctx.setReason("or-else type mismatch")
		return nil, "", "", false
	}

	resultTemp := ctx.newTemp()
	failedTemp := ctx.newTemp()
	valueTemp := ctx.newTemp()
	failureTemp := ""
	errorTemp := ""
	if bindingName != "" {
		failureTemp = ctx.newTemp()
		errorTemp = ctx.newTemp()
	}

	lines := []string{
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("var %s bool", failedTemp),
	}
	if bindingName != "" {
		lines = append(lines, fmt.Sprintf("var %s runtime.Value", failureTemp))
		lines = append(lines, fmt.Sprintf("var %s bool", errorTemp))
	}
	lines = append(lines, "func() {")
	if bindingName != "" {
		lines = append(lines, fmt.Sprintf("\tdefer func() { if recovered := recover(); recovered != nil { switch v := recovered.(type) { case runtime.Value: %s = v; %s = true; %s = true; case error: if val, ok := interpreter.RaisedValue(v); ok { %s = val; %s = true; %s = true } else { panic(recovered) }; default: panic(recovered) } } }()", failureTemp, failedTemp, errorTemp, failureTemp, failedTemp, errorTemp))
	} else {
		lines = append(lines, fmt.Sprintf("\tdefer func() { if recovered := recover(); recovered != nil { switch v := recovered.(type) { case runtime.Value: %s = true; case error: if _, ok := interpreter.RaisedValue(v); ok { %s = true } else { panic(recovered) }; default: panic(recovered) } } }()", failedTemp, failedTemp))
	}
	lines = append(lines, indentLines(valueLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s := %s", valueTemp, valueExpr))
	successExpr := valueTemp
	var successConvLines []string
	switch {
	case valueType == resultType:
	case valueType == "runtime.Value" && resultType != "runtime.Value":
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, valueTemp, resultType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return nil, "", "", false
		}
		successConvLines = convLines
		successExpr = converted
	case resultType == "runtime.Value" && valueType != "runtime.Value":
		convLines, converted, ok := g.runtimeValueLines(ctx, valueTemp, valueType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return nil, "", "", false
		}
		successConvLines = convLines
		successExpr = converted
	default:
		ctx.setReason("or-else type mismatch")
		return nil, "", "", false
	}
	if valueType == "runtime.Value" || valueType == "any" {
		checkTemp := valueTemp
		if valueType == "any" {
			checkTemp = ctx.newTemp()
			lines = append(lines, fmt.Sprintf("\t%s := __able_any_to_value(%s)", checkTemp, valueTemp))
		}
		if bindingName != "" {
			lines = append(lines, fmt.Sprintf("\tif __able_is_nil(%s) { %s = runtime.NilValue{}; %s = true; return }", checkTemp, failureTemp, failedTemp))
			lines = append(lines, fmt.Sprintf("\tif __able_is_error(%s) { %s = %s; %s = true; %s = true; return }", checkTemp, failureTemp, checkTemp, failedTemp, errorTemp))
		} else {
			lines = append(lines, fmt.Sprintf("\tif __able_is_nil(%s) { %s = true; return }", checkTemp, failedTemp))
			lines = append(lines, fmt.Sprintf("\tif __able_is_error(%s) { %s = true; return }", checkTemp, failedTemp))
		}
	}
	lines = append(lines, indentLines(successConvLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, successExpr))
	lines = append(lines, "}()")

	lines = append(lines, fmt.Sprintf("if %s {", failedTemp))
	if bindingName != "" {
		goName := sanitizeIdent(bindingName)
		lines = append(lines, fmt.Sprintf("\tvar %s runtime.Value", goName))
		lines = append(lines, fmt.Sprintf("\tif %s { %s = %s } else { %s = runtime.NilValue{} }", errorTemp, goName, failureTemp, goName))
		lines = append(lines, fmt.Sprintf("\t_ = %s", goName))
	}
	lines = append(lines, indentLines(handlerLines, 1)...)
	lines = append(lines, indentLines(handlerCoerceLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, handlerResultExpr))
	lines = append(lines, "}")
	return lines, resultTemp, resultType, true
}
