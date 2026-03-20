package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) typeMatches(expected, actual string) bool {
	if expected == "" {
		return true
	}
	if expected == actual {
		return true
	}
	// any accepts all types (any is Go's interface{}, all types satisfy it).
	if expected == "any" {
		return true
	}
	if g != nil {
		if iface := g.nativeInterfaceInfoForGoType(expected); iface != nil {
			return g.nativeInterfaceAcceptsActual(iface, actual)
		}
		if union := g.nativeUnionInfoForGoType(expected); union != nil {
			if _, ok := g.nativeUnionMember(union, actual); ok {
				return true
			}
			for _, member := range union.Members {
				if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActual(iface, actual) {
					return true
				}
				if g.nativeUnionRuntimeMemberAcceptsActual(member, actual) {
					return true
				}
			}
		}
	}
	return false
}

func (g *generator) canCoerceStaticExpr(expected, actual string) bool {
	if expected == "" || expected == "any" || actual == "" {
		return true
	}
	if g.nativeNullableWraps(expected, actual) {
		return true
	}
	if g.typeMatches(expected, actual) {
		return true
	}
	if g != nil && g.isIntegerType(expected) && g.isIntegerType(actual) {
		return true
	}
	if expected == "runtime.ErrorValue" && g.isNativeErrorCarrierType(actual) {
		return true
	}
	if innerType, nullable := g.nativeNullableValueInnerType(expected); nullable && innerType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(actual) {
		return true
	}
	if iface := g.nativeInterfaceInfoForGoType(expected); iface != nil {
		return g.nativeInterfaceAcceptsActual(iface, actual)
	}
	if union := g.nativeUnionInfoForGoType(expected); union != nil {
		if _, ok := g.nativeUnionMember(union, actual); ok {
			return true
		}
		for _, member := range union.Members {
			if member == nil {
				continue
			}
			if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActual(iface, actual) {
				return true
			}
			if member.GoType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(actual) {
				return true
			}
			if g.nativeUnionRuntimeMemberAcceptsActual(member, actual) {
				return true
			}
		}
	}
	return false
}

func (g *generator) isStaticallyKnownExpectedType(goType string) bool {
	if goType == "" || goType == "runtime.Value" || g.isVoidType(goType) {
		return true
	}
	if g.isNativeNullableValueType(goType) {
		return true
	}
	if g.nativeInterfaceInfoForGoType(goType) != nil {
		return true
	}
	if g.nativeUnionInfoForGoType(goType) != nil {
		return true
	}
	return g.typeCategory(goType) != "unknown"
}

func (g *generator) compileExprLinesWithExpectedTypeExpr(ctx *compileContext, expr ast.Expression, expectedGoType string, expectedTypeExpr ast.TypeExpression) ([]string, string, string, bool) {
	if ctx == nil || expectedTypeExpr == nil {
		return g.compileExprLines(ctx, expr, expectedGoType)
	}
	previous := ctx.expectedTypeExpr
	ctx.expectedTypeExpr = expectedTypeExpr
	defer func() {
		ctx.expectedTypeExpr = previous
	}()
	return g.compileExprLines(ctx, expr, expectedGoType)
}

func (g *generator) nativeUnionWrapLines(ctx *compileContext, expected, actual, expr string) ([]string, string, bool) {
	if wrapped, ok := g.nativeUnionWrapExpr(expected, actual, expr); ok {
		return nil, wrapped, true
	}
	if g == nil || ctx == nil || expected == "" || actual == "" || expr == "" {
		return nil, "", false
	}
	union := g.nativeUnionInfoForGoType(expected)
	if union == nil {
		return nil, "", false
	}
	if !g.isNativeErrorCarrierType(actual) {
		for _, member := range union.Members {
			if member == nil {
				continue
			}
			if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActual(iface, actual) {
				ifaceLines, ifaceExpr, ok := g.nativeInterfaceWrapLines(ctx, member.GoType, actual, expr)
				if !ok {
					return nil, "", false
				}
				return ifaceLines, fmt.Sprintf("%s(%s)", member.WrapHelper, ifaceExpr), true
			}
		}
		return nil, "", false
	}
	member, ok := g.nativeUnionMember(union, "runtime.ErrorValue")
	if !ok || member == nil {
		for _, runtimeMember := range union.Members {
			if !g.nativeUnionRuntimeMemberAcceptsActual(runtimeMember, actual) {
				continue
			}
			if actual == "runtime.Value" {
				return nil, fmt.Sprintf("%s(%s)", runtimeMember.WrapHelper, expr), true
			}
			runtimeLines, runtimeExpr, ok := g.runtimeValueLines(ctx, expr, actual)
			if !ok {
				return nil, "", false
			}
			return runtimeLines, fmt.Sprintf("%s(%s)", runtimeMember.WrapHelper, runtimeExpr), true
		}
		return nil, "", false
	}
	lines, errorExpr, ok := g.nativeErrorValueLines(ctx, actual, expr)
	if !ok {
		return nil, "", false
	}
	return lines, fmt.Sprintf("%s(%s)", member.WrapHelper, errorExpr), true
}

func (g *generator) compileableInterfaceMethodForReceiver(goType string, interfaceName string, methodName string) *functionInfo {
	if g == nil || goType == "" || goType == "runtime.Value" || goType == "any" || interfaceName == "" || methodName == "" {
		return nil
	}
	var found *functionInfo
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable {
			continue
		}
		if impl.InterfaceName != interfaceName || impl.MethodName != methodName {
			continue
		}
		if len(impl.Info.Params) == 0 || impl.Info.Params[0].GoType == "" {
			continue
		}
		receiverType := impl.Info.Params[0].GoType
		if receiverType != goType && !g.typeMatches(receiverType, goType) {
			continue
		}
		if found != nil && found != impl.Info {
			return nil
		}
		found = impl.Info
	}
	return found
}

func (g *generator) compileableInterfaceMethodForConcreteReceiver(goType string, methodName string) *methodInfo {
	if g == nil || goType == "" || goType == "runtime.Value" || goType == "any" || methodName == "" {
		return nil
	}
	var found *methodInfo
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable || impl.ImplName != "" {
			continue
		}
		if impl.MethodName != methodName || len(impl.Info.Params) == 0 || impl.Info.Params[0].GoType == "" {
			continue
		}
		receiverType := impl.Info.Params[0].GoType
		if receiverType != goType && !g.typeMatches(receiverType, goType) {
			continue
		}
		candidate := &methodInfo{MethodName: methodName, ExpectsSelf: true, Info: impl.Info}
		if found != nil && found.Info != candidate.Info {
			return nil
		}
		found = candidate
	}
	return found
}

func (g *generator) nativeErrorValueLines(ctx *compileContext, actual string, expr string) ([]string, string, bool) {
	if g == nil || ctx == nil || actual == "" || expr == "" || !g.isNativeErrorCarrierType(actual) {
		return nil, "", false
	}
	if actual == "runtime.ErrorValue" {
		return nil, expr, true
	}
	var runtimeLines []string
	runtimeExpr := expr
	if g.typeCategory(actual) == "struct" {
		baseName, ok := g.structBaseName(actual)
		if !ok {
			baseName = strings.TrimPrefix(actual, "*")
		}
		runtimeTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		runtimeLines = []string{
			fmt.Sprintf("%s, %s := __able_struct_%s_to(__able_runtime, %s)", runtimeTemp, errTemp, baseName, expr),
		}
		controlTemp := ctx.newTemp()
		runtimeLines = append(runtimeLines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		runtimeLines = append(runtimeLines, controlLines...)
		runtimeExpr = runtimeTemp
	} else {
		var ok bool
		runtimeLines, runtimeExpr, ok = g.runtimeValueLines(ctx, expr, actual)
		if !ok {
			return nil, "", false
		}
	}
	errorTemp := ctx.newTemp()
	lines := append([]string{}, runtimeLines...)
	var messageInfo *functionInfo
	var causeInfo *functionInfo
	if method := g.methodForReceiver(actual, "message"); method != nil && method.Info != nil && method.Info.Compileable && method.ExpectsSelf && method.Info.ReturnType == "string" {
		messageInfo = method.Info
	} else if info := g.compileableInterfaceMethodForReceiver(actual, "Error", "message"); info != nil && info.ReturnType == "string" {
		messageInfo = info
	}
	if method := g.methodForReceiver(actual, "cause"); method != nil && method.Info != nil && method.Info.Compileable && method.ExpectsSelf {
		if innerType, nullable := g.nativeNullableValueInnerType(method.Info.ReturnType); nullable && innerType == "runtime.ErrorValue" {
			causeInfo = method.Info
		}
	} else if info := g.compileableInterfaceMethodForReceiver(actual, "Error", "cause"); info != nil {
		if innerType, nullable := g.nativeNullableValueInnerType(info.ReturnType); nullable && innerType == "runtime.ErrorValue" {
			causeInfo = info
		}
	}
	if messageInfo != nil {
		messageTemp := ctx.newTemp()
		payloadTemp := ctx.newTemp()
		messageControlTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s, %s := __able_compiled_%s(%s)", messageTemp, messageControlTemp, messageInfo.GoName, expr),
			fmt.Sprintf("%s := map[string]runtime.Value{\"value\": %s}", payloadTemp, runtimeExpr),
		)
		controlLines, ok := g.controlCheckLines(ctx, messageControlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		if causeInfo != nil {
			causeTemp := ctx.newTemp()
			causeControlTemp := ctx.newTemp()
			lines = append(lines,
				fmt.Sprintf("%s, %s := __able_compiled_%s(%s)", causeTemp, causeControlTemp, causeInfo.GoName, expr),
				fmt.Sprintf("if %s != nil { %s[\"cause\"] = __able_nullable_error_to_value(%s) }", causeTemp, payloadTemp, causeTemp),
			)
			controlLines, ok = g.controlCheckLines(ctx, causeControlTemp)
			if !ok {
				return nil, "", false
			}
			lines = append(lines, controlLines...)
		}
		lines = append(lines, fmt.Sprintf("%s := runtime.ErrorValue{Message: %s, Payload: %s}", errorTemp, messageTemp, payloadTemp))
	} else {
		lines = append(lines, fmt.Sprintf("%s := bridge.ErrorValue(__able_runtime, %s)", errorTemp, runtimeExpr))
	}
	return lines, errorTemp, true
}

func (g *generator) wrapLinesAsExpression(ctx *compileContext, lines []string, expr string, exprType string) (string, string, bool) {
	if len(lines) == 0 {
		return expr, exprType, true
	}
	if expr == "" || exprType == "" {
		ctx.setReason("missing expression")
		return "", "", false
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", exprType, strings.Join(lines, "; "), expr), exprType, true
}

func (g *generator) zeroValueExpr(goType string) (string, bool) {
	if g.isNativeNullableValueType(goType) {
		return "nil", true
	}
	if g.isMonoArrayType(goType) {
		return "nil", true
	}
	if g.nativeInterfaceInfoForGoType(goType) != nil {
		return "nil", true
	}
	if g.nativeCallableInfoForGoType(goType) != nil {
		return "nil", true
	}
	if g.nativeUnionInfoForGoType(goType) != nil {
		return "nil", true
	}
	switch goType {
	case "struct{}":
		return "struct{}{}", true
	case "runtime.ErrorValue":
		return "runtime.ErrorValue{}", true
	case "runtime.Value":
		return "runtime.NilValue{}", true
	case "any":
		return "nil", true
	case "bool":
		return "false", true
	case "string":
		return "\"\"", true
	case "rune":
		return "rune(0)", true
	case "float32":
		return "float32(0)", true
	case "float64":
		return "float64(0)", true
	case "int":
		return "int(0)", true
	case "uint":
		return "uint(0)", true
	case "int8":
		return "int8(0)", true
	case "int16":
		return "int16(0)", true
	case "int32":
		return "int32(0)", true
	case "int64":
		return "int64(0)", true
	case "uint8":
		return "uint8(0)", true
	case "uint16":
		return "uint16(0)", true
	case "uint32":
		return "uint32(0)", true
	case "uint64":
		return "uint64(0)", true
	}
	if g.typeCategory(goType) == "struct" {
		base := strings.TrimPrefix(goType, "*")
		if strings.HasPrefix(goType, "*") {
			return fmt.Sprintf("&%s{}", base), true
		}
		return fmt.Sprintf("%s{}", base), true
	}
	return "", false
}

func (g *generator) typedStringifyExpr(expr string, goType string) (string, bool) {
	switch goType {
	case "bool":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatBool(%s)", expr), true
	case "rune":
		return fmt.Sprintf("string(%s)", expr), true
	case "float32":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatFloat(float64(%s), 'f', -1, 32)", expr), true
	case "float64":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatFloat(%s, 'f', -1, 64)", expr), true
	case "int":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatInt(int64(%s), 10)", expr), true
	case "int8", "int16", "int32", "int64":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatInt(int64(%s), 10)", expr), true
	case "uint":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatUint(uint64(%s), 10)", expr), true
	case "uint8", "uint16", "uint32", "uint64":
		g.needsStrconv = true
		return fmt.Sprintf("strconv.FormatUint(uint64(%s), 10)", expr), true
	}
	return "", false
}

func (g *generator) interfaceArgExprLines(ctx *compileContext, argExpr string, ifaceType ast.TypeExpression, context string, genericNames map[string]struct{}) ([]string, string, bool) {
	if argExpr == "" {
		return nil, "", false
	}
	resultTemp := ctx.newTemp()
	if g.typeExprHasGeneric(ifaceType, genericNames) {
		lines := []string{
			fmt.Sprintf("var %s runtime.Value", resultTemp),
			fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} } else { %s = %s }", argExpr, resultTemp, resultTemp, argExpr),
		}
		return lines, resultTemp, true
	}
	rendered, ok := g.renderTypeExpression(ifaceType)
	if !ok {
		return nil, "", false
	}
	expected := typeExpressionToString(ifaceType)
	if context == "" {
		context = "<call>"
	}
	valTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }"),
		fmt.Sprintf("%s, %s, %s := bridge.MatchType(__able_runtime, %s, %s)", valTemp, okTemp, errTemp, rendered, argExpr),
		fmt.Sprintf("__able_panic_on_error(%s)", errTemp),
		fmt.Sprintf("if !%s { panic(fmt.Errorf(\"type mismatch calling %s: expected %s\")) }", okTemp, context, expected),
		fmt.Sprintf("var %s runtime.Value", resultTemp),
		fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} } else { %s = %s }", valTemp, resultTemp, resultTemp, valTemp),
	}
	return lines, resultTemp, true
}

func (g *generator) interfaceReturnExprLines(ctx *compileContext, valueExpr string, ifaceType ast.TypeExpression, genericNames map[string]struct{}) ([]string, string, bool) {
	if valueExpr == "" {
		return nil, "", false
	}
	resultTemp := ctx.newTemp()
	if g.typeExprHasGeneric(ifaceType, genericNames) {
		lines := []string{
			fmt.Sprintf("var %s runtime.Value", resultTemp),
			fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} } else { %s = %s }", valueExpr, resultTemp, resultTemp, valueExpr),
		}
		return lines, resultTemp, true
	}
	rendered, ok := g.renderTypeExpression(ifaceType)
	if !ok {
		return nil, "", false
	}
	expected := typeExpressionToString(ifaceType)
	valTemp := ctx.newTemp()
	okTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("if __able_runtime == nil { panic(fmt.Errorf(\"compiler: missing runtime\")) }"),
		fmt.Sprintf("%s, %s, %s := bridge.MatchType(__able_runtime, %s, %s)", valTemp, okTemp, errTemp, rendered, valueExpr),
		fmt.Sprintf("__able_panic_on_error(%s)", errTemp),
		fmt.Sprintf("if !%s { panic(fmt.Errorf(\"return type mismatch: expected %s\")) }", okTemp, expected),
		fmt.Sprintf("var %s runtime.Value", resultTemp),
		fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} } else { %s = %s }", valTemp, resultTemp, resultTemp, valTemp),
	}
	return lines, resultTemp, true
}

func (g *generator) structReturnConversionLines(resultName, goType, runtimeVar string) ([]string, bool) {
	if g == nil {
		return nil, false
	}
	baseName, ok := g.structBaseName(goType)
	if !ok {
		baseName = strings.TrimPrefix(goType, "*")
		if baseName == "" {
			return nil, false
		}
	}
	if strings.HasPrefix(goType, "*") {
		return []string{
			fmt.Sprintf("if %s == nil {", resultName),
			"\treturn runtime.NilValue{}, nil",
			"}",
			fmt.Sprintf("return __able_struct_%s_to(%s, %s)", baseName, runtimeVar, resultName),
		}, true
	}
	return []string{fmt.Sprintf("return __able_struct_%s_to(%s, &%s)", baseName, runtimeVar, resultName)}, true
}

func (g *generator) nativeArrayValuesExpr(subjectTemp, subjectType string) (string, bool) {
	if g == nil || subjectTemp == "" || !g.isStaticArrayType(subjectType) {
		return "", false
	}
	return fmt.Sprintf("%s.Elements", subjectTemp), true
}

func (g *generator) nativeArrayFromElementsLines(ctx *compileContext, arrayType string, elementsExpr string) ([]string, string, bool) {
	if ctx == nil || arrayType == "" || elementsExpr == "" {
		return nil, "", false
	}
	if spec, ok := g.monoArraySpecForGoType(arrayType); ok && spec != nil {
		valuesTemp := ctx.newTemp()
		arrayTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := append([]%s(nil), %s...)", valuesTemp, spec.ElemGoType, elementsExpr),
			fmt.Sprintf("%s := &%s{Elements: %s}", arrayTemp, spec.GoName, valuesTemp),
			fmt.Sprintf("%s(%s)", spec.SyncHelper, arrayTemp),
		}
		return lines, arrayTemp, true
	}
	if !g.isArrayStructType(arrayType) {
		return nil, "", false
	}
	valuesTemp := ctx.newTemp()
	arrayTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := append([]runtime.Value(nil), %s...)", valuesTemp, elementsExpr),
		fmt.Sprintf("%s := &Array{Elements: %s}", arrayTemp, valuesTemp),
		fmt.Sprintf("__able_struct_Array_sync(%s)", arrayTemp),
	}
	return lines, arrayTemp, true
}
