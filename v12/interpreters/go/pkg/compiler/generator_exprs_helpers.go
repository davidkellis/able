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
			return g.nativeInterfaceAcceptsActualShallow(iface, actual)
		}
		if union := g.nativeUnionInfoForGoType(expected); union != nil {
			return g.nativeUnionAcceptsActual(union, actual)
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
	if g.staticArrayCarrierCoercible(expected, actual) {
		return true
	}
	if g.typeMatches(expected, actual) {
		return true
	}
	if g != nil && g.nominalStructCarrierCoercible(expected, actual) {
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
		if g.nativeInterfaceAcceptsActual(iface, actual) {
			return true
		}
		switch g.typeCategory(actual) {
		case "struct", "interface", "union", "callable", "monoarray", "runtime", "runtime_error":
			return true
		}
		return false
	}
	if g.nativeCallableInfoForGoType(expected) != nil {
		if actual == expected || actual == "runtime.Value" || actual == "any" {
			return true
		}
		if g.nativeCallableInfoForGoType(actual) != nil {
			return true
		}
	}
	if union := g.nativeUnionInfoForGoType(expected); union != nil {
		if g.nativeUnionAcceptsActual(union, actual) {
			return true
		}
	}
	return false
}

func (g *generator) nativeCallableWrapLines(ctx *compileContext, expected string, actual string, expr string) ([]string, string, bool) {
	info := g.nativeCallableInfoForGoType(expected)
	if info == nil || ctx == nil || actual == "" || expr == "" {
		return nil, "", false
	}
	if actual == expected {
		return nil, expr, true
	}
	if actual == "any" {
		valueTemp := ctx.newTemp()
		lines := []string{fmt.Sprintf("%s := __able_any_to_value(%s)", valueTemp, expr)}
		moreLines, converted, ok := g.nativeCallableWrapLines(ctx, expected, "runtime.Value", valueTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, moreLines...)
		return lines, converted, true
	}
	if actual == "runtime.Value" {
		convertedTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convertedTemp, errTemp, info.FromRuntimeHelper, expr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convertedTemp, true
	}
	if actualInfo := g.nativeCallableInfoForGoType(actual); actualInfo != nil {
		runtimeTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		convertedTemp := ctx.newTemp()
		convertErrTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", runtimeTemp, errTemp, actualInfo.ToRuntimeHelper, expr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines,
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convertedTemp, convertErrTemp, info.FromRuntimeHelper, runtimeTemp),
			fmt.Sprintf("%s = __able_control_from_error(%s)", controlTemp, convertErrTemp),
		)
		controlLines, ok = g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convertedTemp, true
	}
	if valueLines, valueExpr, ok := g.lowerRuntimeValue(ctx, expr, actual); ok {
		convertedTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := append([]string{}, valueLines...)
		lines = append(lines,
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convertedTemp, errTemp, info.FromRuntimeHelper, valueExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		)
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convertedTemp, true
	}
	return nil, "", false
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
	if ctx == nil {
		return g.compileExprLines(ctx, expr, expectedGoType)
	}
	expectedTypeExpr = g.concretizedExpectedTypeExpr(ctx, expectedGoType, expectedTypeExpr)
	if expectedTypeExpr == nil {
		return g.compileExprLines(ctx, expr, expectedGoType)
	}
	previous := ctx.expectedTypeExpr
	ctx.expectedTypeExpr = expectedTypeExpr
	defer func() {
		ctx.expectedTypeExpr = previous
	}()
	return g.compileExprLines(ctx, expr, expectedGoType)
}

func (g *generator) concretizedExpectedTypeExpr(ctx *compileContext, expectedGoType string, expectedTypeExpr ast.TypeExpression) ast.TypeExpression {
	if g == nil {
		return expectedTypeExpr
	}
	if expectedTypeExpr != nil {
		expectedTypeExpr = g.lowerNormalizedTypeExpr(ctx, expectedTypeExpr)
		if expectedTypeExpr != nil {
			if expectedGoType == "" || expectedGoType == "runtime.Value" || expectedGoType == "any" || g.typeExprCompatibleWithCarrier(ctx, expectedTypeExpr, expectedGoType) {
				return expectedTypeExpr
			}
		}
	}
	if expectedGoType != "" && expectedGoType != "runtime.Value" && expectedGoType != "any" {
		if expr, ok := g.typeExprForGoType(expectedGoType); ok && expr != nil {
			return g.lowerNormalizedTypeExpr(ctx, expr)
		}
	}
	return nil
}

func (g *generator) staticParamCarrierType(ctx *compileContext, param paramInfo) string {
	expectedType := param.GoType
	if g == nil || ctx == nil || param.TypeExpr == nil {
		return expectedType
	}
	if expectedType != "" && expectedType != "runtime.Value" && expectedType != "any" {
		return expectedType
	}
	if recovered, ok := g.joinCarrierTypeFromTypeExpr(ctx, param.TypeExpr); ok && recovered != "" {
		return recovered
	}
	recovered, ok := g.lowerCarrierType(ctx, param.TypeExpr)
	if !ok || recovered == "" {
		return expectedType
	}
	return recovered
}

func (g *generator) nativeUnionWrapLines(ctx *compileContext, expected, actual, expr string) ([]string, string, bool) {
	if g == nil || ctx == nil || expected == "" || actual == "" || expr == "" {
		return nil, "", false
	}
	union := g.nativeUnionInfoForGoType(expected)
	if union == nil {
		return nil, "", false
	}
	return g.nativeUnionWrapLinesSeen(ctx, union, actual, expr, make(map[string]struct{}))
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
		if !g.receiverGoTypeCompatible(receiverType, goType) {
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
	return g.compileableInterfaceMethodForConcreteReceiverExpr(nil, nil, goType, methodName)
}

func (g *generator) receiverGoTypeCompatible(expectedReceiverType string, actualGoType string) bool {
	if g == nil || expectedReceiverType == "" || actualGoType == "" {
		return false
	}
	if expectedReceiverType == actualGoType || g.typeMatches(expectedReceiverType, actualGoType) {
		return true
	}
	if expectedInfo := g.structInfoByGoName(expectedReceiverType); expectedInfo != nil {
		if actualInfo := g.structInfoByGoName(actualGoType); actualInfo != nil && expectedInfo.Name != "" && expectedInfo.Name == actualInfo.Name {
			return g.receiverNominalFamilyCompatible(expectedReceiverType, actualGoType)
		}
	}
	// Compiler-owned static array wrappers should reuse the shared Array impl
	// surface for specialization instead of dropping to dynamic dispatch.
	if g.isArrayStructType(expectedReceiverType) && g.isStaticArrayType(actualGoType) {
		return true
	}
	return false
}

func (g *generator) receiverNominalFamilyCompatible(expectedReceiverType string, actualGoType string) bool {
	if g == nil || expectedReceiverType == "" || actualGoType == "" {
		return false
	}
	expectedInfo := g.structInfoByGoName(expectedReceiverType)
	actualInfo := g.structInfoByGoName(actualGoType)
	if expectedInfo == nil || actualInfo == nil || expectedInfo.Name == "" || expectedInfo.Name != actualInfo.Name {
		return false
	}
	expectedExpr, ok := g.typeExprForGoType(expectedReceiverType)
	if !ok || expectedExpr == nil {
		return false
	}
	actualExpr, ok := g.typeExprForGoType(actualGoType)
	if !ok || actualExpr == nil {
		return false
	}
	expectedExpr = normalizeTypeExprForPackage(g, expectedInfo.Package, expectedExpr)
	actualExpr = normalizeTypeExprForPackage(g, actualInfo.Package, actualExpr)
	if expectedExpr == nil || actualExpr == nil {
		return false
	}
	return g.typeExprFullyBound(actualInfo.Package, actualExpr) && !g.typeExprFullyBound(expectedInfo.Package, expectedExpr)
}

func (g *generator) compileableInterfaceStaticMethodForConcreteTarget(target ast.TypeExpression, methodName string) *methodInfo {
	if g == nil || target == nil || methodName == "" {
		return nil
	}
	targetKey := typeExpressionToString(target)
	targetName, targetNameOK := g.methodTargetName(target)
	var found *methodInfo
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable || impl.ImplName != "" || impl.TargetType == nil {
			continue
		}
		if impl.MethodName != methodName || methodDefinitionExpectsSelf(impl.Info.Definition) {
			continue
		}
		if implTypeName, ok := g.methodTargetName(impl.TargetType); targetNameOK && ok {
			if implTypeName != targetName {
				continue
			}
		} else if typeExpressionToString(impl.TargetType) != targetKey {
			continue
		}
		candidate := &methodInfo{
			TargetName:   typeExpressionToString(impl.TargetType),
			TargetType:   impl.TargetType,
			MethodName:   methodName,
			ExpectsSelf:  false,
			Info:         impl.Info,
			ReceiverType: "",
		}
		if found != nil && found.Info != candidate.Info {
			if equivalentFunctionInfoSignature(found.Info, candidate.Info) {
				found = candidate
				continue
			}
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
		baseName, ok := g.structHelperName(actual)
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
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		runtimeLines = append(runtimeLines, controlLines...)
		runtimeExpr = runtimeTemp
	} else {
		var ok bool
		runtimeLines, runtimeExpr, ok = g.lowerRuntimeValue(ctx, expr, actual)
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
		controlLines, ok := g.lowerControlCheck(ctx, messageControlTemp)
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
			controlLines, ok = g.lowerControlCheck(ctx, causeControlTemp)
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

func (g *generator) goTypeHasNilZeroValue(goType string) bool {
	if goType == "" || goType == "runtime.Value" || goType == "runtime.ErrorValue" {
		return false
	}
	if goType == "any" {
		return true
	}
	if strings.HasPrefix(goType, "*") || strings.HasPrefix(goType, "[]") {
		return true
	}
	if g.isMonoArrayType(goType) {
		return true
	}
	if g.nativeInterfaceInfoForGoType(goType) != nil {
		return true
	}
	if g.nativeCallableInfoForGoType(goType) != nil {
		return true
	}
	if g.nativeUnionInfoForGoType(goType) != nil {
		return true
	}
	return false
}

func (g *generator) typedNilExpr(goType string) (string, bool) {
	if !g.goTypeHasNilZeroValue(goType) {
		return "", false
	}
	if strings.HasPrefix(goType, "*") || strings.HasPrefix(goType, "[]") {
		return fmt.Sprintf("(%s)(nil)", goType), true
	}
	return fmt.Sprintf("%s(nil)", goType), true
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
	baseName, ok := g.structHelperName(goType)
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
