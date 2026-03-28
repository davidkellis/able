package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) nominalMethodSpecializationGenericNames(method *methodInfo) map[string]struct{} {
	if method == nil {
		return nil
	}
	return mergeGenericNameSets(g.callableGenericNames(method.Info), g.methodGenericNames(method))
}

func (g *generator) nominalTargetGenericParams(method *methodInfo) []*ast.GenericParameter {
	if g == nil || method == nil || method.TargetType == nil || method.Info == nil {
		return nil
	}
	baseName, ok := typeExprBaseName(method.TargetType)
	if !ok || baseName == "" {
		return nil
	}
	if info, ok := g.structInfoForTypeName(method.Info.Package, baseName); ok && info != nil && info.Node != nil {
		return info.Node.GenericParams
	}
	if iface, ok := g.interfaces[baseName]; ok && iface != nil {
		return iface.GenericParams
	}
	return nil
}

func (g *generator) bindGenericTypeArguments(pkgName string, bindings map[string]ast.TypeExpression, params []*ast.GenericParameter, args []ast.TypeExpression) bool {
	if len(params) != len(args) {
		return false
	}
	for idx, arg := range args {
		gp := params[idx]
		if gp == nil || gp.Name == nil || gp.Name.Name == "" || arg == nil {
			return false
		}
		normalized := normalizeTypeExprForPackage(g, pkgName, arg)
		if existing, exists := bindings[gp.Name.Name]; exists {
			if normalizeTypeExprString(g, pkgName, existing) != normalizeTypeExprString(g, pkgName, normalized) {
				return false
			}
			continue
		}
		bindings[gp.Name.Name] = normalized
	}
	return true
}

func (g *generator) applyNominalCallTypeArgumentBindings(method *methodInfo, call *ast.FunctionCall, bindings map[string]ast.TypeExpression) bool {
	if g == nil || method == nil || method.Info == nil || call == nil || len(call.TypeArguments) == 0 {
		return true
	}
	methodParams := method.Info.Definition.GenericParams
	targetParams := g.nominalTargetGenericParams(method)
	args := call.TypeArguments
	switch {
	case len(methodParams) > 0 && len(args) == len(methodParams):
		return g.bindGenericTypeArguments(method.Info.Package, bindings, methodParams, args)
	case len(targetParams) > 0 && len(args) == len(targetParams):
		return g.bindGenericTypeArguments(method.Info.Package, bindings, targetParams, args)
	case len(targetParams) > 0 && len(methodParams) > 0 && len(args) == len(targetParams)+len(methodParams):
		if !g.bindGenericTypeArguments(method.Info.Package, bindings, targetParams, args[:len(targetParams)]) {
			return false
		}
		return g.bindGenericTypeArguments(method.Info.Package, bindings, methodParams, args[len(targetParams):])
	case len(methodParams) == 0 && len(targetParams) == 0:
		return false
	default:
		return false
	}
}

func (g *generator) concreteNominalMethodTargetTypeExpr(ctx *compileContext, method *methodInfo, target ast.Expression, expected string) (ast.TypeExpression, bool) {
	if g == nil || method == nil {
		return nil, false
	}
	if expr, ok := g.staticTargetTypeExpr(ctx, target); ok && expr != nil {
		expr = g.refineStaticTargetTypeExprWithExpected(ctx, target, method.Info.Package, expr, expected)
		return normalizeTypeExprForPackage(g, method.Info.Package, expr), true
	}
	ident, ok := target.(*ast.Identifier)
	if !ok || ident == nil || ident.Name != method.TargetName || method.TargetType == nil {
		return nil, false
	}
	expr := normalizeTypeExprForPackage(g, method.Info.Package, substituteTypeParams(method.TargetType, ctx.typeBindings))
	if expr == nil {
		return nil, false
	}
	if g.typeExprHasGeneric(expr, g.nominalMethodSpecializationGenericNames(method)) {
		if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil {
			if expectedBase, ok := typeExprBaseName(expectedExpr); ok && expectedBase == method.TargetName {
				return normalizeTypeExprForPackage(g, method.Info.Package, expectedExpr), true
			}
		}
	}
	return expr, true
}

func (g *generator) refineStaticTargetTypeExprWithExpected(ctx *compileContext, target ast.Expression, pkgName string, targetExpr ast.TypeExpression, expected string) ast.TypeExpression {
	if g == nil || targetExpr == nil {
		return targetExpr
	}
	expectedExpr := g.specializationExpectedTypeExpr(ctx, expected)
	if expectedExpr == nil {
		return targetExpr
	}
	targetBase, ok := typeExprBaseName(targetExpr)
	if !ok || targetBase == "" {
		return targetExpr
	}
	expectedBase, ok := typeExprBaseName(expectedExpr)
	if !ok || expectedBase == "" || expectedBase != targetBase {
		return targetExpr
	}
	if _, ok := targetExpr.(*ast.GenericTypeExpression); ok {
		return targetExpr
	}
	if ident, ok := target.(*ast.Identifier); ok && ident != nil && ident.Name != "" && ctx != nil {
		if bound, exists := ctx.typeBindings[ident.Name]; exists && bound != nil {
			if _, boundGeneric := bound.(*ast.GenericTypeExpression); boundGeneric {
				return targetExpr
			}
		}
	}
	return normalizeTypeExprForPackage(g, pkgName, expectedExpr)
}

func (g *generator) concreteMethodCallInfo(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, receiver ast.Expression, receiverType string, expected string) *methodInfo {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil {
		return method
	}
	impl := g.implMethodByInfo[method.Info]
	if impl != nil {
		specialized, ok := g.specializeConcreteImplMethod(ctx, call, method, impl, receiver, receiverType, expected)
		if ok && specialized != nil {
			return specialized
		}
		return method
	}
	specialized, ok := g.specializeConcreteNominalMethod(ctx, call, method, receiver, receiverType, expected)
	if !ok || specialized == nil {
		return method
	}
	return specialized
}

func (g *generator) concreteStaticMethodCallInfo(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, target ast.Expression, expected string) *methodInfo {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || method.ExpectsSelf {
		return method
	}
	impl := g.implMethodByInfo[method.Info]
	if impl != nil {
		specialized, ok := g.specializeConcreteStaticImplMethod(ctx, call, method, impl, target, expected)
		if ok && specialized != nil {
			return specialized
		}
		return method
	}
	specialized, ok := g.specializeConcreteStaticNominalMethod(ctx, call, method, target, expected)
	if !ok || specialized == nil {
		return method
	}
	return specialized
}

func (g *generator) specializeConcreteNominalMethod(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, receiver ast.Expression, receiverType string, expected string) (*methodInfo, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil {
		return nil, false
	}
	genericNames := g.nominalMethodSpecializationGenericNames(method)
	if len(genericNames) == 0 {
		return nil, false
	}
	receiverTypeExpr, ok := g.staticReceiverTypeExpr(ctx, receiver, receiverType)
	if !ok || receiverTypeExpr == nil || g.typeExprHasGeneric(receiverTypeExpr, genericNames) {
		return nil, false
	}
	bindings, ok := g.specializedNominalMethodBindings(ctx, call, method, receiverTypeExpr, expected)
	if !ok || len(bindings) == 0 {
		return nil, false
	}
	return g.ensureSpecializedNominalMethod(method, bindings)
}

func (g *generator) specializeConcreteStaticNominalMethod(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, target ast.Expression, expected string) (*methodInfo, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || method.ExpectsSelf {
		return nil, false
	}
	genericNames := g.nominalMethodSpecializationGenericNames(method)
	if len(genericNames) == 0 {
		return nil, false
	}
	targetTypeExpr, ok := g.concreteNominalMethodTargetTypeExpr(ctx, method, target, expected)
	if !ok || targetTypeExpr == nil {
		return nil, false
	}
	bindings, ok := g.specializedStaticNominalMethodBindings(ctx, call, method, targetTypeExpr, expected)
	if !ok || len(bindings) == 0 {
		return nil, false
	}
	return g.ensureSpecializedNominalMethod(method, bindings)
}

func (g *generator) ensureSpecializedNominalMethod(method *methodInfo, bindings map[string]ast.TypeExpression) (*methodInfo, bool) {
	if g == nil || method == nil || method.Info == nil || len(bindings) == 0 {
		return nil, false
	}
	key := g.specializedImplFunctionKey(method.Info, bindings)
	if existing, ok := g.reusableSpecializedFunctionInfo(key, method.Info); ok {
		return g.specializedNominalMethodInfo(method, existing, bindings), true
	}
	specialized := &functionInfo{
		Name:           method.Info.Name,
		Package:        method.Info.Package,
		QualifiedName:  method.Info.QualifiedName,
		GoName:         g.mangler.unique(method.Info.GoName + "_spec"),
		TypeBindings:   cloneTypeBindings(bindings),
		Definition:     method.Info.Definition,
		HasOriginal:    method.Info.HasOriginal,
		InternalOnly:   true,
		SupportedTypes: method.Info.SupportedTypes,
	}
	mapper := NewTypeMapper(g, specialized.Package)
	concreteTarget := normalizeTypeExprForPackage(g, specialized.Package, substituteTypeParams(method.TargetType, bindings))
	if concreteTarget == nil {
		concreteTarget = method.TargetType
	}
	g.fillMethodInfo(specialized, mapper, concreteTarget, method.ExpectsSelf)
	if !specialized.SupportedTypes {
		return nil, false
	}
	specialized.Compileable = true
	g.specializedFunctions = append(g.specializedFunctions, specialized)
	g.touchNativeInterfaceAdapters()
	g.specializedFunctionIndex[key] = specialized
	if g.bodyCompileable(specialized, specialized.ReturnType) {
		specialized.Compileable = true
		specialized.Reason = ""
	}
	return g.specializedNominalMethodInfo(method, specialized, bindings), true
}

func (g *generator) specializedNominalMethodInfo(method *methodInfo, info *functionInfo, bindings map[string]ast.TypeExpression) *methodInfo {
	if method == nil || info == nil {
		return nil
	}
	receiverType := method.ReceiverType
	if method.ExpectsSelf && len(info.Params) > 0 {
		receiverType = info.Params[0].GoType
	}
	targetType := method.TargetType
	if len(bindings) > 0 {
		targetType = normalizeTypeExprForPackage(g, info.Package, substituteTypeParams(targetType, bindings))
	}
	return &methodInfo{
		TargetName:   method.TargetName,
		TargetType:   targetType,
		MethodName:   method.MethodName,
		ReceiverType: receiverType,
		ExpectsSelf:  method.ExpectsSelf,
		Info:         info,
	}
}

func (g *generator) specializedNominalMethodBindings(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, receiverTypeExpr ast.TypeExpression, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || receiverTypeExpr == nil {
		return nil, false
	}
	genericNames := g.nominalMethodSpecializationGenericNames(method)
	bindings := g.concreteCompileContextBindings(method.Info, genericNames)
	bindings = g.mergeConcreteTypeBindings(method.Info.Package, genericNames, bindings, ctx.typeBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if method.TargetType != nil {
		if !g.specializedTypeTemplateMatches(method.Info.Package, method.TargetType, receiverTypeExpr, genericNames, bindings, make(map[string]struct{})) {
			return nil, false
		}
	}
	bindings = g.normalizeConcreteTypeBindings(method.Info.Package, bindings, genericNames)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	return g.finishSpecializedNominalMethodBindings(ctx, call, method, genericNames, bindings, expected)
}

func (g *generator) specializedStaticNominalMethodBindings(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, targetTypeExpr ast.TypeExpression, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || targetTypeExpr == nil {
		return nil, false
	}
	genericNames := g.nominalMethodSpecializationGenericNames(method)
	bindings := g.concreteCompileContextBindings(method.Info, genericNames)
	bindings = g.mergeConcreteTypeBindings(method.Info.Package, genericNames, bindings, ctx.typeBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if method.TargetType != nil {
		if !g.specializedTypeTemplateMatches(method.Info.Package, method.TargetType, targetTypeExpr, genericNames, bindings, make(map[string]struct{})) {
			return nil, false
		}
	}
	bindings = g.normalizeConcreteTypeBindings(method.Info.Package, bindings, genericNames)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	return g.finishSpecializedNominalMethodBindings(ctx, call, method, genericNames, bindings, expected)
}

func (g *generator) preferredNominalMethodTargetTypeExpr(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, receiverTypeExpr ast.TypeExpression, expected string) ast.TypeExpression {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || method.TargetType == nil {
		return nil
	}
	genericNames := g.nominalMethodSpecializationGenericNames(method)
	if len(genericNames) == 0 {
		return nil
	}
	bindings := g.concreteCompileContextBindings(method.Info, genericNames)
	bindings = g.mergeConcreteTypeBindings(method.Info.Package, genericNames, bindings, ctx.typeBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if receiverTypeExpr != nil {
		receiverTypeExpr = normalizeTypeExprForPackage(g, method.Info.Package, receiverTypeExpr)
		if method.TargetType != nil {
			_ = g.applySpecializedTypeTemplateMatch(method.Info.Package, method.TargetType, receiverTypeExpr, genericNames, bindings)
		}
	}
	bindings, ok := g.finishSpecializedNominalMethodBindings(ctx, call, method, genericNames, bindings, expected)
	if !ok || len(bindings) == 0 {
		return nil
	}
	targetTypeExpr := normalizeTypeExprForPackage(g, method.Info.Package, substituteTypeParams(method.TargetType, bindings))
	if targetTypeExpr == nil || g.typeExprHasGeneric(targetTypeExpr, genericNames) {
		return nil
	}
	return targetTypeExpr
}

func (g *generator) finishSpecializedNominalMethodBindings(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil {
		return nil, false
	}
	if len(call.TypeArguments) > 0 {
		if method.Info.Definition == nil {
			return nil, false
		}
		if !g.applyNominalCallTypeArgumentBindings(method, call, bindings) {
			return nil, false
		}
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil && method.Info.Definition != nil && method.Info.Definition.ReturnType != nil {
		returnExpr := g.functionReturnTypeExprWithBindings(method.Info, bindings)
		if returnExpr != nil {
			_ = g.applySpecializedTypeTemplateMatch(method.Info.Package, returnExpr, expectedExpr, genericNames, bindings)
		}
	}
	paramOffset := 0
	if method.ExpectsSelf {
		paramOffset = 1
	}
	for idx, arg := range call.Arguments {
		paramIdx := paramOffset + idx
		if paramIdx >= len(method.Info.Params) {
			break
		}
		paramTypeExpr := method.Info.Params[paramIdx].TypeExpr
		if paramTypeExpr == nil {
			continue
		}
		actualExpr, actualGoType, ok := g.specializedCallActualTypeExpr(ctx, method.Info.Package, arg, paramTypeExpr, bindings)
		actualExpr, ok = g.specializationConcreteArgTypeExprForParam(method.Info.Package, paramTypeExpr, actualExpr, actualGoType)
		if !ok || actualExpr == nil {
			continue
		}
		_ = g.bindSpecializedCallArgument(method.Info.Package, paramTypeExpr, actualExpr, genericNames, bindings)
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil && method.Info.Definition != nil && method.Info.Definition.ReturnType != nil {
		returnExpr := g.functionReturnTypeExprWithBindings(method.Info, bindings)
		if returnExpr != nil {
			_ = g.applySpecializedTypeTemplateMatch(method.Info.Package, returnExpr, expectedExpr, genericNames, bindings)
		}
	}
	bindings = g.normalizeConcreteTypeBindings(method.Info.Package, bindings, genericNames)
	if len(bindings) == 0 {
		return nil, false
	}
	return bindings, true
}
