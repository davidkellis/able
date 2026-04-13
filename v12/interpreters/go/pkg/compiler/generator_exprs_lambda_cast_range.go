package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileTypeCast(ctx *compileContext, expr *ast.TypeCastExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Expression == nil || expr.TargetType == nil {
		ctx.setReason("missing type cast")
		return nil, "", "", false
	}
	valueLines, valueExpr, valueType, ok := g.compileExprLines(ctx, expr.Expression, "")
	if !ok {
		return nil, "", "", false
	}
	targetGoType := ""
	if mapped, mappedOK := g.lowerCarrierType(ctx, expr.TargetType); mappedOK && mapped != "" {
		targetGoType = mapped
	}
	if targetGoType != "" && valueType != "runtime.Value" {
		nodeName := g.diagNodeName(expr, "*ast.TypeCastExpression", "cast")
		if nativeCastLines, nativeCastExpr, nativeCastType, castOK := g.nativePrimitiveCastLines(ctx, nodeName, valueExpr, valueType, targetGoType); castOK {
			lines := append([]string{}, valueLines...)
			lines = append(lines, nativeCastLines...)
			if expected == "runtime.Value" {
				convLines, runtimeExpr, ok := g.lowerRuntimeValue(ctx, nativeCastExpr, nativeCastType)
				if !ok {
					ctx.setReason("cast type mismatch")
					return nil, "", "", false
				}
				lines = append(lines, convLines...)
				return lines, runtimeExpr, "runtime.Value", true
			}
			if expected == "" || expected == nativeCastType {
				return lines, nativeCastExpr, nativeCastType, true
			}
		}
	}
	convLines, valueRuntime, ok := g.lowerRuntimeValue(ctx, valueExpr, valueType)
	if !ok {
		ctx.setReason("cast operand unsupported")
		return nil, "", "", false
	}
	lines := append([]string{}, valueLines...)
	lines = append(lines, convLines...)
	targetExpr, ok := g.renderTypeExpression(expr.TargetType)
	if !ok {
		ctx.setReason("unsupported cast type")
		return nil, "", "", false
	}
	castTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := __able_cast(%s, %s)", castTemp, controlTemp, valueRuntime, targetExpr))
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	castExpr := castTemp
	if expected == "runtime.Value" {
		return lines, castExpr, "runtime.Value", true
	}
	desiredType := "runtime.Value"
	if expected != "" && expected != "runtime.Value" {
		desiredType = expected
	} else if targetGoType != "" {
		desiredType = targetGoType
	}
	if desiredType == "struct{}" {
		ctx.setReason("cast to void unsupported")
		return nil, "", "", false
	}
	if desiredType == "runtime.Value" {
		return lines, castExpr, "runtime.Value", true
	}
	expectLines, converted, ok := g.lowerExpectRuntimeValue(ctx, castExpr, desiredType)
	if !ok {
		ctx.setReason("cast type mismatch")
		return nil, "", "", false
	}
	lines = append(lines, expectLines...)
	return lines, converted, desiredType, true
}

func (g *generator) compileRangeExpression(ctx *compileContext, expr *ast.RangeExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Start == nil || expr.End == nil {
		ctx.setReason("missing range expression")
		return nil, "", "", false
	}
	startLines, startExpr, startType, ok := g.compileExprLines(ctx, expr.Start, "")
	if !ok {
		return nil, "", "", false
	}
	endLines, endExpr, endType, ok := g.compileExprLines(ctx, expr.End, "")
	if !ok {
		return nil, "", "", false
	}
	startConvLines, startRuntime, ok := g.lowerRuntimeValue(ctx, startExpr, startType)
	if !ok {
		ctx.setReason("range start unsupported")
		return nil, "", "", false
	}
	endConvLines, endRuntime, ok := g.lowerRuntimeValue(ctx, endExpr, endType)
	if !ok {
		ctx.setReason("range end unsupported")
		return nil, "", "", false
	}
	startTemp := ctx.newTemp()
	endTemp := ctx.newTemp()
	var lines []string
	lines = append(lines, startLines...)
	lines = append(lines, startConvLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", startTemp, startRuntime))
	lines = append(lines, endLines...)
	lines = append(lines, endConvLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", endTemp, endRuntime))
	rangeExpr := fmt.Sprintf("__able_range(%s, %s, %t)", startTemp, endTemp, expr.Inclusive)
	if expected != "" && expected != "runtime.Value" {
		if info := g.structInfoByGoName(expected); info != nil && info.Name == "Range" {
			return lines, rangeExpr, "runtime.Value", true
		}
		expectLines, converted, ok := g.lowerExpectRuntimeValue(ctx, rangeExpr, expected)
		if !ok {
			ctx.setReason("range expression type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, expectLines...)
		return lines, converted, expected, true
	}
	return lines, rangeExpr, "runtime.Value", true
}

func (g *generator) compileLambdaExpression(ctx *compileContext, expr *ast.LambdaExpression, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing lambda expression")
		return "", "", false
	}
	expectedCallable := g.nativeCallableInfoForGoType(expected)
	if expected != "" && expected != "runtime.Value" && expected != "any" && expectedCallable == nil {
		ctx.setReason("lambda expression type mismatch")
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
	lambdaCtx.controlMode = ""
	controlTemp := lambdaCtx.newTemp()
	controlDoneLabel := lambdaCtx.newTemp()
	lambdaCtx.controlCaptureVar = controlTemp
	lambdaCtx.controlCaptureLabel = controlDoneLabel
	lambdaCtx.controlCaptureBreak = true

	expectedFnType := g.expectedLambdaFunctionType(ctx)
	if expectedFnType != nil && len(expectedFnType.ParamTypes) != len(expr.Params) {
		expectedFnType = nil
	}

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
		paramTypeExpr := param.ParamType
		if paramTypeExpr == nil && expectedFnType != nil && idx < len(expectedFnType.ParamTypes) {
			paramTypeExpr = expectedFnType.ParamTypes[idx]
		}
		if paramTypeExpr != nil {
			mapped, ok := g.lowerCarrierType(ctx, paramTypeExpr)
			if !ok {
				ctx.setReason("unsupported lambda parameter type")
				return "", "", false
			}
			goType = mapped
		}
		info := paramInfo{Name: ident.Name, GoName: goName, GoType: goType, TypeExpr: paramTypeExpr}
		lambdaCtx.locals[ident.Name] = info
		params = append(params, info)
	}
	if len(params) > 0 {
		lambdaCtx.implicitReceiver = params[0]
		lambdaCtx.hasImplicitReceiver = true
	}
	genericValueVars := make(map[string]string)
	if len(expr.GenericParams) > 0 || len(expr.WhereClause) > 0 {
		generics := genericNameSet(expr.GenericParams)
		for idx, param := range params {
			if param.TypeExpr == nil {
				continue
			}
			if simple, ok := param.TypeExpr.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil {
				if _, ok := generics[simple.Name.Name]; ok {
					genericValueVars[simple.Name.Name] = fmt.Sprintf("__able_lambda_arg_%d_value", idx)
				}
			}
		}
	}

	desiredReturn := ""
	lambdaReturnTypeExpr := expr.ReturnType
	if lambdaReturnTypeExpr == nil && expectedFnType != nil {
		lambdaReturnTypeExpr = expectedFnType.ReturnType
	}
	if lambdaReturnTypeExpr != nil {
		mapped, ok := g.lowerCarrierType(ctx, lambdaReturnTypeExpr)
		if !ok {
			ctx.setReason("unsupported lambda return type")
			return "", "", false
		}
		desiredReturn = mapped
	}
	if desiredReturn != "" {
		lambdaCtx.returnType = desiredReturn
	}
	if lambdaReturnTypeExpr != nil {
		lambdaCtx.returnTypeExpr = lambdaReturnTypeExpr
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
		if ctx.reason == "" && lambdaCtx.reason != "" {
			ctx.setReason(lambdaCtx.reason)
		}
		return "", "", false
	}
	if desiredReturn != "" && !g.typeMatches(desiredReturn, bodyType) {
		ctx.setReason("lambda return type mismatch")
		return "", "", false
	}
	paramExprs := make([]ast.TypeExpression, 0, len(params))
	paramGoTypes := make([]string, 0, len(params))
	for _, param := range params {
		paramExpr := param.TypeExpr
		if paramExpr == nil {
			if fallback, ok := g.typeExprForGoType(param.GoType); ok {
				paramExpr = fallback
			}
		}
		if paramExpr == nil {
			ctx.setReason("unsupported lambda parameter type")
			return "", "", false
		}
		paramExprs = append(paramExprs, paramExpr)
		paramGoTypes = append(paramGoTypes, param.GoType)
	}
	returnTypeExpr := lambdaReturnTypeExpr
	if returnTypeExpr == nil {
		if expectedFnType != nil {
			returnTypeExpr = expectedFnType.ReturnType
		} else if fallback, ok := g.typeExprForGoType(bodyType); ok {
			returnTypeExpr = fallback
		}
	}
	if returnTypeExpr == nil {
		ctx.setReason("unsupported lambda return type")
		return "", "", false
	}
	callableInfo, ok := g.ensureNativeCallableInfoFromSignatureInPackage(ctx.packageName, paramExprs, paramGoTypes, returnTypeExpr, bodyType)
	if !ok || callableInfo == nil {
		ctx.setReason("unsupported lambda type")
		return "", "", false
	}
	implLines := make([]string, 0, len(bodyLines)+len(params)*2+3)
	implLines = append(implLines, g.inlineRuntimeEnvSwapLinesForPackage(ctx.packageName)...)
	implLines = append(implLines, fmt.Sprintf("var %s *__ableControl", controlTemp))
	zeroExpr, zeroOK := g.zeroValueExpr(callableInfo.ReturnGoType)
	if !zeroOK {
		implLines = append(implLines, fmt.Sprintf("var __able_zero %s", callableInfo.ReturnGoType))
		zeroExpr = "__able_zero"
	}
	if len(genericValueVars) > 0 {
		for _, param := range params {
			if param.TypeExpr == nil {
				continue
			}
			simple, ok := param.TypeExpr.(*ast.SimpleTypeExpression)
			if !ok || simple == nil || simple.Name == nil {
				continue
			}
			valueVar, ok := genericValueVars[simple.Name.Name]
			if !ok || valueVar == "" {
				continue
			}
			runtimeExpr, ok := g.runtimeValueExpr(param.GoName, param.GoType)
			if !ok {
				ctx.setReason("unsupported lambda generic constraint type")
				return "", "", false
			}
			implLines = append(implLines, fmt.Sprintf("%s := %s", valueVar, runtimeExpr))
		}
		constraintLines, ok := g.lambdaConstraintLines(expr, genericValueVars, zeroExpr)
		if !ok {
			ctx.setReason("unsupported lambda constraints")
			return "", "", false
		}
		implLines = append(implLines, constraintLines...)
	}
	if g.isVoidType(bodyType) {
		implLines = append(implLines, fmt.Sprintf("%s: for {", controlDoneLabel))
		implLines = append(implLines, bodyLines...)
		if bodyExpr != "" {
			implLines = append(implLines, fmt.Sprintf("_ = %s", bodyExpr))
		}
		implLines = append(implLines, fmt.Sprintf("break %s", controlDoneLabel))
		implLines = append(implLines, "}")
		implLines = append(implLines, fmt.Sprintf("if %s != nil { return %s, %s }", controlTemp, zeroExpr, controlTemp))
		implLines = append(implLines, "return struct{}{}, nil")
	} else {
		resultTemp := lambdaCtx.newTemp()
		implLines = append(implLines, fmt.Sprintf("var %s %s", resultTemp, callableInfo.ReturnGoType))
		implLines = append(implLines, fmt.Sprintf("%s: for {", controlDoneLabel))
		implLines = append(implLines, bodyLines...)
		implLines = append(implLines, fmt.Sprintf("%s = %s", resultTemp, bodyExpr))
		implLines = append(implLines, fmt.Sprintf("break %s", controlDoneLabel))
		implLines = append(implLines, "}")
		implLines = append(implLines, fmt.Sprintf("if %s != nil { return %s, %s }", controlTemp, zeroExpr, controlTemp))
		implLines = append(implLines, fmt.Sprintf("return %s, nil", resultTemp))
	}
	lambdaExpr := fmt.Sprintf("%s(func(", callableInfo.GoType)
	for idx, param := range params {
		if idx > 0 {
			lambdaExpr += ", "
		}
		lambdaExpr += fmt.Sprintf("%s %s", param.GoName, param.GoType)
	}
	lambdaExpr += fmt.Sprintf(") (%s, *__ableControl) { %s })", callableInfo.ReturnGoType, strings.Join(implLines, "; "))
	return lambdaExpr, callableInfo.GoType, true
}

func (g *generator) expectedLambdaFunctionType(ctx *compileContext) *ast.FunctionTypeExpression {
	if g == nil || ctx == nil || ctx.expectedTypeExpr == nil {
		return nil
	}
	expectedExpr := g.lowerNormalizedTypeExpr(ctx, ctx.expectedTypeExpr)
	fnType, ok := expectedExpr.(*ast.FunctionTypeExpression)
	if !ok || fnType == nil {
		return nil
	}
	return fnType
}

func (g *generator) lambdaTypeExprCompatible(ctx *compileContext, actual ast.TypeExpression, expected ast.TypeExpression) bool {
	if g == nil || actual == nil || expected == nil {
		return false
	}
	actualExpr := g.lowerNormalizedTypeExpr(ctx, actual)
	expectedExpr := g.lowerNormalizedTypeExpr(ctx, expected)
	actualKey := normalizeTypeExprIdentityKey(g, ctx.packageName, actualExpr)
	expectedKey := normalizeTypeExprIdentityKey(g, ctx.packageName, expectedExpr)
	if actualKey != "" && actualKey == expectedKey {
		return true
	}
	actualType, actualOK := g.lowerCarrierType(ctx, actualExpr)
	expectedType, expectedOK := g.lowerCarrierType(ctx, expectedExpr)
	if !actualOK || !expectedOK || actualType == "" || expectedType == "" {
		return false
	}
	return g.typeMatches(expectedType, actualType) && g.typeMatches(actualType, expectedType)
}

func (g *generator) lambdaExpressionMatchesExpectedFunctionType(ctx *compileContext, expr *ast.LambdaExpression, expected *ast.FunctionTypeExpression) bool {
	if g == nil || ctx == nil || expr == nil || expected == nil {
		return false
	}
	if len(expr.Params) != len(expected.ParamTypes) {
		return false
	}
	for idx, param := range expr.Params {
		if param == nil {
			return false
		}
		if param.ParamType != nil && !g.lambdaTypeExprCompatible(ctx, param.ParamType, expected.ParamTypes[idx]) {
			return false
		}
	}
	if expr.ReturnType != nil && !g.lambdaTypeExprCompatible(ctx, expr.ReturnType, expected.ReturnType) {
		return false
	}
	return true
}

func (g *generator) expectedCallableTypeForLambda(ctx *compileContext, expected string, expr *ast.LambdaExpression) (string, ast.TypeExpression, bool) {
	if g == nil || ctx == nil || expr == nil {
		return "", nil, false
	}
	if callableInfo := g.nativeCallableInfoForGoType(expected); callableInfo != nil && callableInfo.TypeExpr != nil {
		return expected, callableInfo.TypeExpr, true
	}
	union := g.nativeUnionInfoForGoType(expected)
	if union == nil {
		return "", nil, false
	}
	var matchedType string
	var matchedExpr ast.TypeExpression
	for _, member := range union.Members {
		if member == nil || member.TypeExpr == nil {
			continue
		}
		memberExpr := g.lowerNormalizedTypeExpr(ctx, member.TypeExpr)
		fnType, ok := memberExpr.(*ast.FunctionTypeExpression)
		if !ok || fnType == nil || !g.lambdaExpressionMatchesExpectedFunctionType(ctx, expr, fnType) {
			continue
		}
		if matchedType != "" && matchedType != member.GoType {
			return "", nil, false
		}
		matchedType = member.GoType
		matchedExpr = fnType
	}
	if matchedType == "" {
		return "", nil, false
	}
	return matchedType, matchedExpr, true
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
			if ret.Argument == nil {
				if returnType != "" && !g.isVoidType(returnType) {
					ctx.setReason("missing return value")
					return nil, "", "", false
				}
				return lines, "struct{}{}", "struct{}", true
			}
			compileExpected := returnType
			if g.isVoidType(returnType) {
				compileExpected = ""
			}
			stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, compileExpected, ret.Argument)
			if !ok {
				return nil, "", "", false
			}
			if returnType != "" && !g.isVoidType(returnType) && !g.typeMatches(returnType, valueType) {
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
		if isLast {
			if g.isVoidType(returnType) {
				stmtLines, ok := g.compileStatement(ctx, stmt)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, stmtLines...)
				return lines, "struct{}{}", "struct{}", true
			}
			if expr, ok := stmt.(ast.Expression); ok && expr != nil {
				compileExpected := returnType
				stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, compileExpected, expr)
				if !ok {
					return nil, "", "", false
				}
				if returnType != "" && !g.isVoidType(returnType) && !g.typeMatches(returnType, valueType) {
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

func (g *generator) lambdaArgConversionLines(pkgName string, argVar string, valueVar string, goType string, typeExpr ast.TypeExpression, target string) ([]string, bool) {
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		return []string{
			fmt.Sprintf("%s, err := %s(__able_runtime, %s)", target, iface.FromRuntimeHelper, valueVar),
			"if err != nil { return nil, err }",
		}, true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		return []string{
			fmt.Sprintf("%s, err := %s(__able_runtime, %s)", target, callable.FromRuntimeHelper, valueVar),
			"if err != nil { return nil, err }",
		}, true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		return []string{
			fmt.Sprintf("%s, err := %s(__able_runtime, %s)", target, union.FromRuntimeHelper, valueVar),
			"if err != nil { return nil, err }",
		}, true
	}
	if helper, ok := g.nativeNullableFromRuntimeHelper(goType); ok {
		return []string{
			fmt.Sprintf("%s, err := %s(%s)", target, helper, valueVar),
			"if err != nil { return nil, err }",
		}, true
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		lines := []string{
			fmt.Sprintf("%s, err := %s(%s)", target, spec.FromRuntimeHelper, valueVar),
			"if err != nil { return nil, err }",
		}
		if g.staticArrayArgRequiresValue(pkgName, typeExpr, goType) {
			lines = append(lines, fmt.Sprintf("if %s == nil { return nil, fmt.Errorf(\"type mismatch calling lambda: expected %s\") }", target, typeExpressionToString(typeExpr)))
		}
		return lines, true
	}
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
		if g.isArrayStructType(goType) {
			errVar := argVar + "_err"
			lines := []string{
				fmt.Sprintf("var %s *Array", target),
				fmt.Sprintf("var %s error", errVar),
			}
			lines = append(lines, g.runtimeValueToGenericArrayBoundaryLines(target, errVar, valueVar, true)...)
			lines = append(lines, fmt.Sprintf("if %s != nil { return nil, %s }", errVar, errVar))
			if g.structArgRequiresValue(pkgName, typeExpr, goType) {
				lines = append(lines, fmt.Sprintf("if %s == nil { return nil, fmt.Errorf(\"type mismatch calling lambda: expected Array\") }", target))
			}
			return lines, true
		}
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		lines := []string{
			fmt.Sprintf("%s, err := __able_struct_%s_from(%s)", target, baseName, valueVar),
			"if err != nil { return nil, err }",
		}
		if g.structArgRequiresValue(pkgName, typeExpr, goType) {
			lines = append(lines, fmt.Sprintf("if %s == nil { return nil, fmt.Errorf(\"type mismatch calling lambda: expected %s\") }", target, baseName))
		}
		return lines, true
	default:
		return []string{fmt.Sprintf("%s := %s", target, valueVar)}, true
	}
}

func (g *generator) structArgRequiresValue(pkgName string, typeExpr ast.TypeExpression, goType string) bool {
	if g == nil || g.typeCategory(goType) != "struct" {
		return false
	}
	if !strings.HasPrefix(goType, "*") {
		return true
	}
	if typeExpr == nil {
		return true
	}
	if _, ok := typeExpr.(*ast.NullableTypeExpression); ok {
		return false
	}
	if _, members, ok := g.expandedUnionMembersInPackage(pkgName, typeExpr); ok {
		if _, nullable := nativeUnionNullableInnerTypeExpr(members); nullable {
			return false
		}
	}
	return true
}

func (g *generator) lambdaReturnLines(resultName string, goType string) ([]string, bool) {
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		return []string{fmt.Sprintf("return %s(__able_runtime, %s)", spec.ToRuntimeHelper, resultName)}, true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		return []string{fmt.Sprintf("return %s(__able_runtime, %s)", iface.ToRuntimeHelper, resultName)}, true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil {
		return []string{fmt.Sprintf("return %s(__able_runtime, %s)", callable.ToRuntimeHelper, resultName)}, true
	}
	if union := g.nativeUnionInfoForGoType(goType); union != nil {
		return []string{fmt.Sprintf("return %s(__able_runtime, %s)", union.ToRuntimeHelper, resultName)}, true
	}
	if helper, ok := g.nativeNullableToRuntimeHelper(goType); ok {
		return []string{fmt.Sprintf("return %s(%s), nil", helper, resultName)}, true
	}
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
		lines, ok := g.structReturnConversionLines(resultName, goType, "__able_runtime")
		if !ok {
			return []string{fmt.Sprintf("return %s, nil", resultName)}, true
		}
		return lines, true
	default:
		return []string{fmt.Sprintf("return %s, nil", resultName)}, true
	}
}
