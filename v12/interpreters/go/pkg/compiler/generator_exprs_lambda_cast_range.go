package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileTypeCast(ctx *compileContext, expr *ast.TypeCastExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil || expr.TargetType == nil {
		ctx.setReason("missing type cast")
		return "", "", false
	}
	valueExpr, valueType, ok := g.compileExpr(ctx, expr.Expression, "")
	if !ok {
		return "", "", false
	}
	targetGoType := ""
	if mapped, mappedOK := g.mapTypeExpressionInPackage(ctx.packageName, expr.TargetType); mappedOK && mapped != "" {
		targetGoType = mapped
	}
	if targetGoType != "" && valueType != "runtime.Value" {
		if nativeCastExpr, castOK := g.nativeIntegerWidenExpr(valueExpr, valueType, targetGoType); castOK {
			if expected == "runtime.Value" {
				runtimeExpr, ok := g.runtimeValueExpr(nativeCastExpr, targetGoType)
				if !ok {
					ctx.setReason("cast type mismatch")
					return "", "", false
				}
				return runtimeExpr, "runtime.Value", true
			}
			if expected == "" || expected == targetGoType {
				return nativeCastExpr, targetGoType, true
			}
		}
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
	} else if targetGoType != "" {
		desiredType = targetGoType
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
			mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, param.ParamType)
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
	if len(params) > 0 {
		lambdaCtx.implicitReceiver = params[0]
		lambdaCtx.hasImplicitReceiver = true
	}
	genericValueVars := make(map[string]string)
	if len(expr.GenericParams) > 0 || len(expr.WhereClause) > 0 {
		generics := genericNameSet(expr.GenericParams)
		for idx, param := range expr.Params {
			if param == nil || param.ParamType == nil {
				continue
			}
			if simple, ok := param.ParamType.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil {
				if _, ok := generics[simple.Name.Name]; ok {
					genericValueVars[simple.Name.Name] = fmt.Sprintf("__able_lambda_arg_%d_value", idx)
				}
			}
		}
	}

	desiredReturn := ""
	if expr.ReturnType != nil {
		mapped, ok := g.mapTypeExpressionInPackage(ctx.packageName, expr.ReturnType)
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
		if ctx.reason == "" && lambdaCtx.reason != "" {
			ctx.setReason(lambdaCtx.reason)
		}
		return "", "", false
	}
	if desiredReturn != "" && !g.typeMatches(desiredReturn, bodyType) {
		ctx.setReason("lambda return type mismatch")
		return "", "", false
	}

	lambdaResultName := lambdaCtx.newTemp()
	lambdaErrName := lambdaCtx.newTemp()
	implLines := make([]string, 0, len(bodyLines)+len(params)*4+7)
	implLines = append(implLines, fmt.Sprintf("defer func() { if recovered := recover(); recovered != nil { switch recovered.(type) { case __able_break, __able_break_label_signal, __able_continue_signal, __able_continue_label_signal: panic(recovered) }; %s = nil; %s = bridge.Recover(__able_runtime, callCtx, recovered) } }()", lambdaResultName, lambdaErrName))
	implLines = append(implLines, "if __able_runtime != nil && callCtx != nil && callCtx.Env != nil { prevEnv := __able_runtime.SwapEnv(callCtx.Env); defer __able_runtime.SwapEnv(prevEnv) }")
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
		if param.GoName != "_" {
			implLines = append(implLines, fmt.Sprintf("_ = %s", param.GoName))
		}
	}
	if len(genericValueVars) > 0 {
		constraintLines, ok := g.lambdaConstraintLines(expr, genericValueVars)
		if !ok {
			ctx.setReason("unsupported lambda constraints")
			return "", "", false
		}
		implLines = append(implLines, constraintLines...)
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
	lambdaExpr := fmt.Sprintf("runtime.NativeFunctionValue{Name: %q, Arity: %d, Impl: func(callCtx *runtime.NativeCallContext, args []runtime.Value) (%s runtime.Value, %s error) { %s }}", "<lambda>", len(params), lambdaResultName, lambdaErrName, implBody)
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
			if expr, ok := stmt.(ast.Expression); ok && expr != nil {
				compileExpected := returnType
				if g.isVoidType(returnType) {
					compileExpected = ""
				}
				stmtLines, valueExpr, valueType, ok := g.compileTailExpression(ctx, compileExpected, expr)
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
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		return []string{
			fmt.Sprintf("%s, err := __able_struct_%s_from(%s)", target, baseName, valueVar),
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
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		return []string{fmt.Sprintf("return __able_struct_%s_to(__able_runtime, %s)", baseName, resultName)}, true
	default:
		return []string{fmt.Sprintf("return %s, nil", resultName)}, true
	}
}
