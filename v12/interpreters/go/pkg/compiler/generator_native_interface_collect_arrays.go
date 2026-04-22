package compiler

import (
	"bytes"
	"fmt"
	"sort"
)

import "able/interpreter-go/pkg/ast"

type iteratorCollectMonoArrayInfo struct {
	Key               string
	GoName            string
	Package           string
	ReceiverType      string
	ReturnType        string
	ElemGoType        string
	IteratorEndUnwrap string
	ValueUnwrap       string
	DefaultGoName     string
	ExtendGoName      string
}

func (g *generator) compileStaticIteratorCollectMonoArrayCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, method *nativeInterfaceGenericMethod, returnGoType string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Name != "collect" || receiverExpr == "" || receiverType == "" {
		return nil, "", "", false
	}
	info, ok := g.ensureIteratorCollectMonoArrayInfo(method, receiverType, returnGoType)
	if !ok || info == nil {
		return nil, "", "", false
	}
	lines := make([]string, 0, 6)
	receiverLines, coercedReceiverExpr, _, ok := g.prepareStaticCallArg(ctx, receiverExpr, receiverType, info.ReceiverType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, receiverLines...)
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := %s(%s)", resultTemp, controlTemp, g.compiledCallTargetNameForPackage(ctx.packageName, info.Package, info.GoName), coercedReceiverExpr))
	controlLines, ok := g.compiledControlCheckWithCallFrameLines(ctx, controlTemp, callNode)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	return g.finishNativeInterfaceGenericCallReturn(ctx, lines, resultTemp, info.ReturnType, expected)
}

func (g *generator) ensureIteratorCollectMonoArrayInfo(method *nativeInterfaceGenericMethod, receiverType string, returnGoType string) (*iteratorCollectMonoArrayInfo, bool) {
	if g == nil || method == nil || method.Name != "collect" || receiverType == "" || returnGoType == "" {
		return nil, false
	}
	spec, ok := g.monoArraySpecForGoType(returnGoType)
	if !ok || spec == nil {
		return nil, false
	}
	nextMethod, ok := g.nativeInterfaceMethodForGoType(receiverType, "next")
	if !ok || nextMethod == nil || nextMethod.ReturnGoType == "" {
		return nil, false
	}
	nextUnion := g.nativeUnionInfoForGoType(nextMethod.ReturnGoType)
	if nextUnion == nil {
		return nil, false
	}
	iterEndGoType, ok := g.lowerCarrierTypeInPackage(method.InterfacePackage, ast.Ty("IteratorEnd"))
	if !ok || iterEndGoType == "" {
		return nil, false
	}
	iterEndMember, ok := g.nativeUnionMember(nextUnion, iterEndGoType)
	if !ok || iterEndMember == nil {
		return nil, false
	}
	valueMember, ok := g.nativeUnionMember(nextUnion, spec.ElemGoType)
	if !ok || valueMember == nil {
		return nil, false
	}
	key := fmt.Sprintf("%s::%s::%s", method.InterfacePackage, receiverType, returnGoType)
	if existing, ok := g.iteratorCollectMonoArrays[key]; ok && existing != nil {
		return existing, true
	}
	info := &iteratorCollectMonoArrayInfo{
		Key:               key,
		GoName:            g.mangler.unique(fmt.Sprintf("iface_%s_%s_default", sanitizeIdent(method.InterfaceName), sanitizeIdent(method.Name))),
		Package:           method.InterfacePackage,
		ReceiverType:      receiverType,
		ReturnType:        returnGoType,
		ElemGoType:        spec.ElemGoType,
		IteratorEndUnwrap: iterEndMember.UnwrapHelper,
		ValueUnwrap:       valueMember.UnwrapHelper,
	}
	elemExpr, ok := g.typeExprForGoType(spec.ElemGoType)
	if !ok || elemExpr == nil {
		return nil, false
	}
	arrayExpr := ast.NewGenericTypeExpression(ast.Ty("Array"), []ast.TypeExpression{
		normalizeTypeExprForPackage(g, method.InterfacePackage, elemExpr),
	})
	defaultMethod, ok := g.specializedBuiltinArrayInterfaceMethod("Default", "default", arrayExpr, nil)
	if !ok || defaultMethod == nil || defaultMethod.Info == nil {
		return nil, false
	}
	extendMethod, ok := g.specializedBuiltinArrayInterfaceMethod("Extend", "extend", arrayExpr, map[string]ast.TypeExpression{"T": elemExpr})
	if !ok || extendMethod == nil || extendMethod.Info == nil {
		return nil, false
	}
	info.DefaultGoName = defaultMethod.Info.GoName
	info.ExtendGoName = extendMethod.Info.GoName
	g.iteratorCollectMonoArrays[key] = info
	return info, true
}

func (g *generator) specializedBuiltinArrayInterfaceMethod(interfaceName string, methodName string, arrayExpr ast.TypeExpression, extraBindings map[string]ast.TypeExpression) (*methodInfo, bool) {
	if g == nil || interfaceName == "" || methodName == "" || arrayExpr == nil {
		return nil, false
	}
	concreteArrayExpr := normalizeTypeExprForPackage(g, "", arrayExpr)
	bindings := map[string]ast.TypeExpression{
		"Self": concreteArrayExpr,
	}
	var elemExpr ast.TypeExpression
	if generic, ok := concreteArrayExpr.(*ast.GenericTypeExpression); ok && generic != nil && len(generic.Arguments) == 1 {
		if baseName, ok := typeExprBaseName(generic.Base); ok && baseName == "Array" {
			elemExpr = normalizeTypeExprForPackage(g, "", generic.Arguments[0])
			bindings["T"] = elemExpr
		}
	}
	for name, expr := range extraBindings {
		if expr == nil {
			continue
		}
		bindings[name] = normalizeTypeExprForPackage(g, "", expr)
	}
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil {
			continue
		}
		if impl.InterfaceName != interfaceName || impl.MethodName != methodName {
			continue
		}
		baseName, ok := typeExprBaseName(impl.TargetType)
		if !ok || baseName != "Array" {
			continue
		}
		baseInfo := impl.Info
		method := &methodInfo{
			TargetName:  baseName,
			TargetType:  impl.TargetType,
			MethodName:  methodName,
			ExpectsSelf: methodDefinitionExpectsSelf(baseInfo.Definition),
			Info:        baseInfo,
		}
		if method.ExpectsSelf && len(baseInfo.Params) > 0 {
			method.ReceiverType = baseInfo.Params[0].GoType
		}
		implBindings := cloneTypeBindings(bindings)
		for name, expr := range implBindings {
			implBindings[name] = normalizeTypeExprForPackage(g, baseInfo.Package, expr)
		}
		specialized, ok := g.ensureSpecializedImplMethod(method, impl, implBindings)
		expectedSelfGoType, _ := g.lowerCarrierTypeInPackage(baseInfo.Package, concreteArrayExpr)
		expectedElemGoType := ""
		if elemExpr != nil {
			expectedElemGoType, _ = g.lowerCarrierTypeInPackage(baseInfo.Package, elemExpr)
		}
		if ok && specialized != nil && specialized.Info != nil && g.builtinArrayMethodInfoMatchesCarrier(specialized.Info, methodName, expectedSelfGoType, expectedElemGoType) {
			return specialized, true
		}
		if matched, ok := g.findBuiltinArraySpecializedMethodInfo(impl, methodName, concreteArrayExpr, expectedSelfGoType, expectedElemGoType); ok && matched != nil {
			return matched, true
		}
		if synthesized, ok := g.ensureBuiltinArrayConcreteImplMethod(method, impl, concreteArrayExpr, implBindings, expectedSelfGoType, expectedElemGoType); ok && synthesized != nil {
			return synthesized, true
		}
	}
	return nil, false
}

func (g *generator) builtinArrayMethodInfoMatchesCarrier(info *functionInfo, methodName string, expectedSelfGoType string, expectedElemGoType string) bool {
	if g == nil || info == nil || expectedSelfGoType == "" {
		return false
	}
	switch methodName {
	case "default":
		return info.ReturnType == expectedSelfGoType
	case "extend":
		if len(info.Params) < 2 || info.Params[0].GoType != expectedSelfGoType || info.ReturnType != expectedSelfGoType {
			return false
		}
		return expectedElemGoType == "" || info.Params[1].GoType == expectedElemGoType
	default:
		if len(info.Params) > 0 && info.Params[0].GoType != expectedSelfGoType {
			return false
		}
		return true
	}
}

func (g *generator) findBuiltinArraySpecializedMethodInfo(impl *implMethodInfo, methodName string, concreteArrayExpr ast.TypeExpression, expectedSelfGoType string, expectedElemGoType string) (*methodInfo, bool) {
	if g == nil || impl == nil || expectedSelfGoType == "" {
		return nil, false
	}
	for _, info := range g.specializedFunctions {
		if info == nil || !info.Compileable || !g.builtinArrayMethodInfoMatchesCarrier(info, methodName, expectedSelfGoType, expectedElemGoType) {
			continue
		}
		currentImpl := g.implMethodByInfo[info]
		if currentImpl != impl {
			continue
		}
		targetType := g.specializedImplTargetType(impl, info.TypeBindings)
		if targetType == nil {
			targetType = concreteArrayExpr
		}
		receiverType := ""
		if len(info.Params) > 0 {
			receiverType = info.Params[0].GoType
		}
		return &methodInfo{
			TargetName:   "Array",
			TargetType:   targetType,
			MethodName:   methodName,
			ReceiverType: receiverType,
			ExpectsSelf:  methodDefinitionExpectsSelf(info.Definition),
			Info:         info,
		}, true
	}
	return nil, false
}

func (g *generator) ensureBuiltinArrayConcreteImplMethod(method *methodInfo, impl *implMethodInfo, concreteArrayExpr ast.TypeExpression, bindings map[string]ast.TypeExpression, expectedSelfGoType string, expectedElemGoType string) (*methodInfo, bool) {
	if g == nil || method == nil || method.Info == nil || impl == nil || concreteArrayExpr == nil {
		return nil, false
	}
	baseInfo := impl.Info
	if baseInfo == nil {
		baseInfo = method.Info
	}
	fillBindings := cloneTypeBindings(bindings)
	if fillBindings == nil {
		fillBindings = make(map[string]ast.TypeExpression)
	}
	fillBindings["Self"] = normalizeTypeExprForPackage(g, baseInfo.Package, concreteArrayExpr)
	if canonical := g.canonicalImplSpecializationBindings(method.Info, impl, fillBindings); len(canonical) > 0 {
		fillBindings = canonical
	}
	key := g.specializedImplFunctionKey(method.Info, fillBindings)
	if existing, ok := g.reusableSpecializedFunctionInfo(key, method.Info); ok && existing != nil {
		receiverType := ""
		if len(existing.Params) > 0 {
			receiverType = existing.Params[0].GoType
		}
		if g.builtinArrayMethodInfoMatchesCarrier(existing, method.MethodName, expectedSelfGoType, expectedElemGoType) {
			return &methodInfo{
				TargetName:   "Array",
				TargetType:   concreteArrayExpr,
				MethodName:   method.MethodName,
				ReceiverType: receiverType,
				ExpectsSelf:  methodDefinitionExpectsSelf(existing.Definition),
				Info:         existing,
			}, true
		}
	}
	mapper := NewTypeMapper(g, baseInfo.Package)
	specialized := &functionInfo{
		Name:          baseInfo.Name,
		Package:       baseInfo.Package,
		QualifiedName: baseInfo.QualifiedName,
		GoName:        g.mangler.unique(baseInfo.GoName + "_spec"),
		TypeBindings:  cloneTypeBindings(fillBindings),
		Definition:    baseInfo.Definition,
		HasOriginal:   baseInfo.HasOriginal,
		InternalOnly:  true,
	}
	g.fillImplMethodInfo(specialized, mapper, concreteArrayExpr, fillBindings)
	g.invalidateFunctionDerivedInfo(specialized)
	g.refreshRepresentableFunctionInfo(specialized)
	if !specialized.SupportedTypes || !g.builtinArrayMethodInfoMatchesCarrier(specialized, method.MethodName, expectedSelfGoType, expectedElemGoType) {
		return nil, false
	}
	specialized.Compileable = true
	g.implMethodByInfo[specialized] = impl
	g.specializedFunctions = append(g.specializedFunctions, specialized)
	g.touchNativeInterfaceAdapters()
	g.specializedFunctionIndex[key] = specialized
	if g.bodyCompileable(specialized, specialized.ReturnType) {
		specialized.Compileable = true
		specialized.Reason = ""
	}
	receiverType := ""
	if len(specialized.Params) > 0 {
		receiverType = specialized.Params[0].GoType
	}
	return &methodInfo{
		TargetName:   "Array",
		TargetType:   concreteArrayExpr,
		MethodName:   method.MethodName,
		ReceiverType: receiverType,
		ExpectsSelf:  methodDefinitionExpectsSelf(specialized.Definition),
		Info:         specialized,
	}, true
}

func (g *generator) renderIteratorCollectMonoArrayHelpers(buf *bytes.Buffer) {
	if g == nil || buf == nil || len(g.iteratorCollectMonoArrays) == 0 {
		return
	}
	keys := make([]string, 0, len(g.iteratorCollectMonoArrays))
	for key := range g.iteratorCollectMonoArrays {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		info := g.iteratorCollectMonoArrays[key]
		if info == nil {
			continue
		}
		g.renderIteratorCollectMonoArrayHelper(buf, info)
	}
}

func (g *generator) renderIteratorCollectMonoArrayHelper(buf *bytes.Buffer, info *iteratorCollectMonoArrayInfo) {
	if g == nil || buf == nil || info == nil {
		return
	}
	spec, ok := g.monoArraySpecForGoType(info.ReturnType)
	if !ok || spec == nil {
		return
	}
	zeroExpr, ok := g.zeroValueExpr(info.ReturnType)
	if !ok {
		return
	}
	if envVar, ok := g.packageEnvVar(info.Package); ok {
		fmt.Fprintf(buf, "func __able_compiled_%s(self %s) (%s, *__ableControl) {\n", info.GoName, info.ReceiverType, info.ReturnType)
		writeRuntimeEnvSwapIfNeeded(buf, "\t", "__able_runtime", envVar, "")
	} else {
		fmt.Fprintf(buf, "func __able_compiled_%s(self %s) (%s, *__ableControl) {\n", info.GoName, info.ReceiverType, info.ReturnType)
	}
	fmt.Fprintf(buf, "\tacc, control := %s()\n", g.compiledCallTargetNameForPackage(info.Package, info.Package, info.DefaultGoName))
	fmt.Fprintf(buf, "\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\treturn %s, control\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\titer := self\n")
	fmt.Fprintf(buf, "\tfor {\n")
	fmt.Fprintf(buf, "\t\tnext, control := iter.next()\n")
	fmt.Fprintf(buf, "\t\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn %s, control\n", zeroExpr)
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tendValue, endOK := %s(next)\n", info.IteratorEndUnwrap)
	fmt.Fprintf(buf, "\t\tendMatch := false\n")
	fmt.Fprintf(buf, "\t\tif endOK {\n")
	fmt.Fprintf(buf, "\t\t\tendMatch = (endValue != nil)\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tif endMatch {\n")
	fmt.Fprintf(buf, "\t\t\tbreak\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tvalue, valueOK := %s(next)\n", info.ValueUnwrap)
	fmt.Fprintf(buf, "\t\tif valueOK {\n")
	fmt.Fprintf(buf, "\t\t\tnextAcc, control := %s(acc, value)\n", g.compiledCallTargetNameForPackage(info.Package, info.Package, info.ExtendGoName))
	fmt.Fprintf(buf, "\t\t\tif control != nil {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn %s, control\n", zeroExpr)
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\tacc = nextAcc\n")
	fmt.Fprintf(buf, "\t\t\tcontinue\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn %s, __able_runtime_error_control(nil, fmt.Errorf(\"Non-exhaustive match\"))\n", zeroExpr)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn acc, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) finishNativeInterfaceGenericCallReturn(ctx *compileContext, lines []string, resultExpr string, resultType string, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil {
		return nil, "", "", false
	}
	if expected == "" || g.typeMatches(expected, resultType) {
		return lines, resultExpr, resultType, true
	}
	if expected != "runtime.Value" && resultType == "runtime.Value" {
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultExpr, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if expected == "runtime.Value" && resultType != "runtime.Value" {
		convLines, converted, ok := g.lowerRuntimeValue(ctx, resultExpr, resultType)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, "runtime.Value", true
	}
	if expected != "" && expected != "any" && resultType == "any" {
		anyTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", anyTemp, resultExpr))
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, anyTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" && g.canCoerceStaticExpr(expected, resultType) {
		return g.lowerCoerceExpectedStaticExpr(ctx, lines, resultExpr, resultType, expected)
	}
	ctx.setReason("call return type mismatch")
	return nil, "", "", false
}
