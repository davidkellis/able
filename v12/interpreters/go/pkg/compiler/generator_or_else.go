package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compilePropagationExpression(ctx *compileContext, expr *ast.PropagationExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil {
		ctx.setReason("missing propagation expression")
		return "", "", false
	}
	if indexExpr, ok := expr.Expression.(*ast.IndexExpression); ok {
		if fastExpr, fastType, fastOK := g.compilePropagationMonoArrayIndex(ctx, indexExpr, expected); fastOK {
			return fastExpr, fastType, true
		}
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr.Expression)
	if !ok {
		return "", "", false
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
			return "", "", false
		}
		if len(valueLines) == 0 {
			return valueExpr, valueType, true
		}
		exprValue := fmt.Sprintf("func() %s { %s; return %s }()", valueType, strings.Join(valueLines, "; "), valueExpr)
		return exprValue, valueType, true
	}
	valueTemp := ctx.newTemp()
	lines := append([]string{}, valueLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", valueTemp, valueExpr))
	lines = append(lines, fmt.Sprintf("if __able_is_error(%s) { bridge.Raise(__able_error_value(%s)) }", valueTemp, valueTemp))
	resultExpr := valueTemp
	if resultType != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(valueTemp, resultType)
		if !ok {
			ctx.setReason("propagation type mismatch")
			return "", "", false
		}
		resultExpr = converted
	}
	exprValue := fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr)
	return exprValue, resultType, true
}

func (g *generator) compilePropagationMonoArrayIndex(ctx *compileContext, expr *ast.IndexExpression, expected string) (string, string, bool) {
	if g == nil || ctx == nil || expr == nil {
		return "", "", false
	}
	objExpr, objType, ok := g.compileExpr(ctx, expr.Object, "")
	if !ok || !g.isArrayStructType(objType) {
		return "", "", false
	}
	monoKind, monoEnabled := g.monoArrayKindForObject(ctx, expr.Object, objType)
	if !monoEnabled {
		return "", "", false
	}
	idxExpr, idxType, ok := g.compileExpr(ctx, expr.Index, "")
	if !ok {
		return "", "", false
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
		return "", "", false
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
		return "", "", false
	}
	lines = append(lines,
		fmt.Sprintf("%s := %s.Storage_handle", handleRawTemp, objTemp),
		fmt.Sprintf("%s := func() int64 { raw, err := bridge.AsInt(%s, 64); if err != nil { panic(err) }; return raw }()", handleTemp, handleRawTemp),
		fmt.Sprintf("%s, err := runtime.ArrayStoreSize(%s)", lengthTemp, handleTemp),
		"if err != nil { panic(err) }",
		fmt.Sprintf("if %s < 0 || %s >= %s { bridge.Raise(__able_error_value(__able_index_error(%s, %s))) }", indexTemp, indexTemp, lengthTemp, indexTemp, lengthTemp),
		fmt.Sprintf("%s, err := %s", nativeTemp, readExpr),
		"if err != nil { panic(err) }",
	)
	switch {
	case resultType == readType:
		return g.wrapLinesAsExpression(ctx, lines, nativeTemp, readType)
	case resultType == "runtime.Value":
		runtimeExpr, ok := g.runtimeValueExpr(nativeTemp, readType)
		if !ok {
			return "", "", false
		}
		return g.wrapLinesAsExpression(ctx, lines, runtimeExpr, "runtime.Value")
	default:
		if widenedExpr, ok := g.nativeIntegerWidenExpr(nativeTemp, readType, resultType); ok {
			return g.wrapLinesAsExpression(ctx, lines, widenedExpr, resultType)
		}
		runtimeExpr, ok := g.runtimeValueExpr(nativeTemp, readType)
		if !ok {
			return "", "", false
		}
		convertedExpr, ok := g.expectRuntimeValueExpr(runtimeExpr, resultType)
		if !ok {
			return "", "", false
		}
		return g.wrapLinesAsExpression(ctx, lines, convertedExpr, resultType)
	}
}

func (g *generator) compileOrElseExpression(ctx *compileContext, expr *ast.OrElseExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil || expr.Handler == nil {
		ctx.setReason("missing or-else expression")
		return "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, "", expr.Expression)
	if !ok {
		return "", "", false
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
		return "", "", false
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
	switch {
	case handlerType == resultType:
	case handlerType == "runtime.Value" && resultType != "runtime.Value":
		converted, ok := g.expectRuntimeValueExpr(handlerExpr, resultType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return "", "", false
		}
		handlerResultExpr = converted
	case resultType == "runtime.Value" && handlerType != "runtime.Value":
		converted, ok := g.runtimeValueExpr(handlerExpr, handlerType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return "", "", false
		}
		handlerResultExpr = converted
	default:
		ctx.setReason("or-else type mismatch")
		return "", "", false
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
	switch {
	case valueType == resultType:
	case valueType == "runtime.Value" && resultType != "runtime.Value":
		converted, ok := g.expectRuntimeValueExpr(valueTemp, resultType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return "", "", false
		}
		successExpr = converted
	case resultType == "runtime.Value" && valueType != "runtime.Value":
		converted, ok := g.runtimeValueExpr(valueTemp, valueType)
		if !ok {
			ctx.setReason("or-else type mismatch")
			return "", "", false
		}
		successExpr = converted
	default:
		ctx.setReason("or-else type mismatch")
		return "", "", false
	}
	if valueType == "runtime.Value" {
		if bindingName != "" {
			lines = append(lines, fmt.Sprintf("\tif __able_is_nil(%s) { %s = runtime.NilValue{}; %s = true; return }", valueTemp, failureTemp, failedTemp))
			lines = append(lines, fmt.Sprintf("\tif __able_is_error(%s) { %s = %s; %s = true; %s = true; return }", valueTemp, failureTemp, valueTemp, failedTemp, errorTemp))
		} else {
			lines = append(lines, fmt.Sprintf("\tif __able_is_nil(%s) { %s = true; return }", valueTemp, failedTemp))
			lines = append(lines, fmt.Sprintf("\tif __able_is_error(%s) { %s = true; return }", valueTemp, failedTemp))
		}
	}
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
	lines = append(lines, fmt.Sprintf("\treturn %s", handlerResultExpr))
	lines = append(lines, "}")
	lines = append(lines, fmt.Sprintf("return %s", resultTemp))

	exprValue := fmt.Sprintf("func() %s { %s }()", resultType, strings.Join(lines, "; "))
	return exprValue, resultType, true
}
