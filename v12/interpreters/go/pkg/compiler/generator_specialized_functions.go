package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) specializationConcreteArgTypeExpr(pkgName string, expr ast.TypeExpression) (ast.TypeExpression, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	if expr == nil || !g.typeExprFullyBound(pkgName, expr) {
		return nil, false
	}
	goType, ok := g.lowerCarrierTypeInPackage(pkgName, expr)
	if !ok || goType == "" || goType == "runtime.Value" || goType == "any" {
		return nil, false
	}
	return expr, true
}

func (g *generator) specializationConcreteArgTypeExprForParam(pkgName string, paramExpr ast.TypeExpression, actualExpr ast.TypeExpression, actualGoType string) (ast.TypeExpression, bool) {
	if g == nil || paramExpr == nil || actualExpr == nil {
		return nil, false
	}
	paramExpr = normalizeTypeExprForPackage(g, pkgName, paramExpr)
	actualExpr = normalizeTypeExprForPackage(g, pkgName, actualExpr)
	concreteExpr, ok := g.specializationConcreteArgTypeExpr(pkgName, actualExpr)
	if !ok || concreteExpr == nil {
		return nil, false
	}
	if actualGoType != "" && actualGoType != "runtime.Value" && actualGoType != "any" {
		mapped, ok := g.lowerCarrierTypeInPackage(pkgName, concreteExpr)
		mapped, ok = g.recoverRepresentableCarrierType(pkgName, concreteExpr, mapped)
		if !ok || mapped == "" || mapped == "runtime.Value" || mapped == "any" {
			return nil, false
		}
		if mapped != actualGoType {
			return nil, false
		}
	}
	return concreteExpr, true
}

func (g *generator) discardSpecializedFunctionInfo(key string, info *functionInfo) {
	if g == nil || info == nil {
		return
	}
	if key != "" {
		if existing, ok := g.specializedFunctionIndex[key]; ok && existing == info {
			delete(g.specializedFunctionIndex, key)
		}
	}
	filtered := g.specializedFunctions[:0]
	for _, candidate := range g.specializedFunctions {
		if candidate == nil || candidate == info {
			continue
		}
		filtered = append(filtered, candidate)
	}
	g.specializedFunctions = filtered
}

func (g *generator) reusableSpecializedFunctionInfo(key string, info *functionInfo) (*functionInfo, bool) {
	if g == nil || key == "" {
		return nil, false
	}
	existing, ok := g.specializedFunctionIndex[key]
	if !ok || existing == nil {
		return nil, false
	}
	if existing.Compileable || existing == info {
		return existing, true
	}
	// A same-key stale entry must be removed before replacement. Otherwise the
	// selection scans over `specializedFunctions` can observe both bodies and
	// keep an older, less concrete specialization.
	g.discardSpecializedFunctionInfo(key, existing)
	return nil, false
}

func (g *generator) specializationExpectedTypeExpr(ctx *compileContext, expected string) ast.TypeExpression {
	if g == nil {
		return nil
	}
	var contextual ast.TypeExpression
	if ctx != nil && ctx.expectedTypeExpr != nil {
		contextual = g.lowerNormalizedTypeExpr(ctx, ctx.expectedTypeExpr)
		if contextual != nil && !g.typeExprHasGeneric(contextual, ctx.genericNames) && g.typeExprFullyBound(ctx.packageName, contextual) {
			return contextual
		}
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" {
		expectedExpr, ok := g.typeExprForGoType(expected)
		if ok && expectedExpr != nil {
			expectedExpr = g.lowerNormalizedTypeExpr(ctx, expectedExpr)
			pkgName := ""
			if ctx != nil {
				pkgName = ctx.packageName
			}
			if g.typeExprFullyBound(pkgName, expectedExpr) {
				return expectedExpr
			}
		}
	}
	if ctx != nil && contextual != nil && g.typeExprFullyBound(ctx.packageName, contextual) {
		return contextual
	}
	return nil
}

func (g *generator) concreteFunctionCallInfo(ctx *compileContext, call *ast.FunctionCall, info *functionInfo, expected string) *functionInfo {
	if g == nil || ctx == nil || call == nil || info == nil || info.Definition == nil {
		return info
	}
	genericNames := g.compileContextGenericNames(info)
	if len(genericNames) == 0 {
		return info
	}
	if ctx.analysisOnly && expected == "" && ctx.expectedTypeExpr == nil && len(call.TypeArguments) == 0 {
		return info
	}
	bindings, ok := g.specializedFunctionBindings(ctx, call, info, expected)
	if !ok || len(bindings) == 0 {
		return info
	}
	specialized, ok := g.ensureSpecializedFunctionInfo(info, bindings)
	if !ok || specialized == nil {
		return info
	}
	return specialized
}

func (g *generator) applySpecializedTypeTemplateMatch(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil || bindings == nil {
		return false
	}
	candidate := cloneTypeBindings(bindings)
	if candidate == nil {
		candidate = make(map[string]ast.TypeExpression)
	}
	if !g.specializedTypeTemplateMatches(pkgName, template, actual, genericNames, candidate, make(map[string]struct{})) {
		return false
	}
	applyTypeBindings(bindings, candidate)
	return true
}

func (g *generator) applySpecializedConcreteInterfaceReturnBinding(pkgName string, returnTypeExpr ast.TypeExpression, expectedExpr ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || returnTypeExpr == nil || expectedExpr == nil || bindings == nil {
		return false
	}
	candidate := cloneTypeBindings(bindings)
	if candidate == nil {
		candidate = make(map[string]ast.TypeExpression)
	}
	for name := range genericNames {
		delete(candidate, name)
	}
	if !g.bindSpecializedConcreteInterfaceReturn(pkgName, returnTypeExpr, expectedExpr, genericNames, candidate) {
		return false
	}
	applyTypeBindings(bindings, candidate)
	return true
}

func (g *generator) specializedFunctionBindings(ctx *compileContext, call *ast.FunctionCall, info *functionInfo, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || info == nil || info.Definition == nil {
		return nil, false
	}
	genericNames := g.compileContextGenericNames(info)
	bindings := g.concreteCompileContextBindings(info, genericNames)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	// Do not seed free-function generic bindings from the enclosing compile
	// context by raw generic name. Function generic names are local to the
	// callee; inheriting outer bindings here lets unrelated generics with the
	// same spelling poison specialization selection.
	if len(call.TypeArguments) > 0 {
		if len(call.TypeArguments) != len(info.Definition.GenericParams) {
			return nil, false
		}
		for idx, arg := range call.TypeArguments {
			gp := info.Definition.GenericParams[idx]
			if gp == nil || gp.Name == nil || gp.Name.Name == "" || arg == nil {
				return nil, false
			}
			concreteArg, ok := g.specializationConcreteArgTypeExpr(info.Package, arg)
			if !ok || concreteArg == nil {
				continue
			}
			bindings[gp.Name.Name] = concreteArg
		}
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil {
		if _, _, _, _, ok := interfaceExprInfo(g, info.Package, expectedExpr); ok {
			returnBindings := cloneTypeBindings(bindings)
			if returnBindings == nil {
				returnBindings = make(map[string]ast.TypeExpression)
			}
			for name := range genericNames {
				delete(returnBindings, name)
			}
			if returnExpr := g.functionReturnTypeExprWithBindings(info, returnBindings); returnExpr != nil {
				_ = g.applySpecializedConcreteInterfaceReturnBinding(info.Package, returnExpr, expectedExpr, genericNames, bindings)
			}
		} else if returnExpr := g.functionReturnTypeExprWithBindings(info, bindings); returnExpr != nil {
			_ = g.applySpecializedTypeTemplateMatch(info.Package, returnExpr, expectedExpr, genericNames, bindings)
		}
	}
	for idx, arg := range call.Arguments {
		if idx >= len(info.Params) {
			break
		}
		paramType := g.functionParamTypeExpr(info, idx)
		if paramType == nil {
			paramType = info.Params[idx].TypeExpr
		}
		if paramType == nil {
			continue
		}
		actualExpr, actualGoType, ok := g.specializedCallActualTypeExpr(ctx, info.Package, arg, paramType, bindings)
		actualExpr, ok = g.specializationConcreteArgTypeExprForParam(info.Package, paramType, actualExpr, actualGoType)
		if !ok || actualExpr == nil {
			continue
		}
		_ = g.bindSpecializedCallArgument(info.Package, paramType, actualExpr, genericNames, bindings)
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil {
		if _, _, _, _, ok := interfaceExprInfo(g, info.Package, expectedExpr); ok {
			returnBindings := cloneTypeBindings(bindings)
			if returnBindings == nil {
				returnBindings = make(map[string]ast.TypeExpression)
			}
			for name := range genericNames {
				delete(returnBindings, name)
			}
			if returnExpr := g.functionReturnTypeExprWithBindings(info, returnBindings); returnExpr != nil {
				_ = g.applySpecializedConcreteInterfaceReturnBinding(info.Package, returnExpr, expectedExpr, genericNames, bindings)
			}
		} else if returnExpr := g.functionReturnTypeExprWithBindings(info, bindings); returnExpr != nil {
			_ = g.applySpecializedTypeTemplateMatch(info.Package, returnExpr, expectedExpr, genericNames, bindings)
		}
	}
	bindings = g.normalizeConcreteTypeBindings(info.Package, bindings, genericNames)
	if len(bindings) == 0 {
		return nil, false
	}
	return bindings, true
}

func (g *generator) bindSpecializedConcreteInterfaceReturn(pkgName string, returnTypeExpr ast.TypeExpression, expectedExpr ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || returnTypeExpr == nil || expectedExpr == nil || bindings == nil {
		return false
	}
	ifacePkg, ifaceName, ifaceArgs, _, ok := interfaceExprInfo(g, pkgName, expectedExpr)
	if !ok {
		return false
	}
	returnTemplate := normalizeTypeExprForPackage(g, pkgName, returnTypeExpr)
	if returnTemplate == nil {
		return false
	}
	var found map[string]ast.TypeExpression
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || impl.InterfaceName == "" || impl.ImplName != "" {
			continue
		}
		if impl.InterfaceName != ifaceName {
			continue
		}
		candidateGenericNames := mergeGenericNameSets(genericNames, nativeInterfaceGenericNameSet(impl.InterfaceGenerics))
		candidateGenericNames = mergeGenericNameSets(candidateGenericNames, g.callableGenericNames(info))
		candidateBindings := cloneTypeBindings(bindings)
		if candidateBindings == nil {
			candidateBindings = make(map[string]ast.TypeExpression)
		}
		targetTemplate := g.specializedImplTargetTemplate(impl, candidateBindings)
		if targetTemplate == nil {
			targetTemplate = impl.TargetType
		}
		if targetTemplate == nil {
			continue
		}
		targetTemplate = normalizeTypeExprForPackage(g, info.Package, targetTemplate)
		matchedTarget := g.specializedTypeTemplateMatches(pkgName, returnTemplate, targetTemplate, candidateGenericNames, candidateBindings, make(map[string]struct{}))
		if !matchedTarget {
			matchedTarget = g.specializedTypeTemplateMatches(pkgName, targetTemplate, returnTemplate, candidateGenericNames, candidateBindings, make(map[string]struct{}))
		}
		if !matchedTarget {
			matchedTarget = g.specializedSameBaseGenericBindings(pkgName, returnTemplate, targetTemplate, candidateGenericNames, candidateBindings)
		}
		if !matchedTarget {
			matchedTarget = g.specializedSameBaseGenericBindings(pkgName, targetTemplate, returnTemplate, candidateGenericNames, candidateBindings)
		}
		if !matchedTarget {
			continue
		}
		actualIfaceExpr := g.implConcreteInterfaceExpr(impl, candidateBindings)
		if actualIfaceExpr == nil {
			continue
		}
		matched, ok := g.nativeInterfaceImplBindingsForTarget(info.Package, actualIfaceExpr, candidateGenericNames, ifacePkg, ifaceName, ifaceArgs, make(map[string]struct{}))
		if !ok || !g.mergeConcreteBindings(pkgName, candidateBindings, matched) {
			continue
		}
		resolvedBindings := make(map[string]ast.TypeExpression, len(candidateBindings))
		for name, expr := range candidateBindings {
			if expr == nil {
				continue
			}
			resolvedBindings[name] = normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(expr, candidateBindings))
		}
		normalized := g.normalizeConcreteTypeBindings(pkgName, resolvedBindings, genericNames)
		if len(normalized) == 0 {
			continue
		}
		if found != nil {
			if !concreteBindingsEquivalent(g, pkgName, found, normalized) {
				return false
			}
			continue
		}
		found = normalized
	}
	if len(found) == 0 {
		return false
	}
	applyTypeBindings(bindings, found)
	return true
}

func (g *generator) ensureSpecializedFunctionInfo(info *functionInfo, bindings map[string]ast.TypeExpression) (*functionInfo, bool) {
	if g == nil || info == nil || info.Definition == nil || len(bindings) == 0 {
		return nil, false
	}
	if g.implMethodByInfo != nil {
		if impl := g.implMethodByInfo[info]; impl != nil {
			baseInfo := impl.Info
			if baseInfo == nil {
				baseInfo = info
			}
			targetName, _ := typeExprBaseName(impl.TargetType)
			method := &methodInfo{
				TargetName:   targetName,
				TargetType:   impl.TargetType,
				MethodName:   impl.MethodName,
				ExpectsSelf:  methodDefinitionExpectsSelf(baseInfo.Definition),
				Info:         baseInfo,
			}
			if method.ExpectsSelf && len(baseInfo.Params) > 0 {
				method.ReceiverType = baseInfo.Params[0].GoType
			}
			specialized, ok := g.ensureSpecializedImplMethod(method, impl, bindings)
			if !ok || specialized == nil || specialized.Info == nil {
				return nil, false
			}
			return specialized.Info, true
		}
	}
	name := strings.TrimSpace(info.Name)
	qualified := strings.TrimSpace(info.QualifiedName)
	if strings.HasPrefix(name, "impl ") || strings.HasPrefix(qualified, "impl ") {
		return nil, false
	}
	key := g.specializedImplFunctionKey(info, bindings)
	if existing, ok := g.reusableSpecializedFunctionInfo(key, info); ok {
		existing.TypeBindings = cloneTypeBindings(bindings)
		mapper := NewTypeMapper(g, existing.Package)
		g.fillSpecializedFunctionInfo(existing, mapper)
		if !existing.SupportedTypes {
			return nil, false
		}
		return existing, true
	}
	specialized := &functionInfo{
		Name:           info.Name,
		Package:        info.Package,
		QualifiedName:  info.QualifiedName,
		GoName:         g.mangler.unique(info.GoName + "_spec"),
		TypeBindings:   cloneTypeBindings(bindings),
		Definition:     info.Definition,
		HasOriginal:    info.HasOriginal,
		InternalOnly:   true,
		SupportedTypes: info.SupportedTypes,
	}
	mapper := NewTypeMapper(g, specialized.Package)
	g.fillSpecializedFunctionInfo(specialized, mapper)
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
	return specialized, true
}

func (g *generator) fillSpecializedFunctionInfo(info *functionInfo, mapper *TypeMapper) {
	if info == nil || info.Definition == nil || mapper == nil {
		return
	}
	g.invalidateFunctionDerivedInfo(info)
	def := info.Definition
	params := make([]paramInfo, 0, len(def.Params))
	supported := true
	if def.IsMethodShorthand {
		supported = false
	}
	for idx, param := range def.Params {
		name := "arg"
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			name = ident.Name
		} else {
			name = safeParamName(name, idx)
			supported = false
		}
		paramType := normalizeTypeExprForPackage(g, info.Package, substituteTypeParams(param.ParamType, info.TypeBindings))
		goType, ok := mapper.Map(paramType)
		goType, ok = g.recoverRepresentableCarrierType(info.Package, paramType, goType)
		if !ok || goType == "" {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      name,
			GoName:    safeParamName(name, idx),
			GoType:    goType,
			TypeExpr:  paramType,
			Supported: ok,
		})
	}
	retExpr := normalizeTypeExprForPackage(g, info.Package, substituteTypeParams(def.ReturnType, info.TypeBindings))
	retType, ok := mapper.Map(retExpr)
	retType, ok = g.recoverRepresentableCarrierType(info.Package, retExpr, retType)
	if !ok || retType == "" {
		supported = false
	}
	info.Params = params
	info.ReturnType = retType
	info.SupportedTypes = supported
	info.Arity = len(params)
	if !supported {
		info.Compileable = false
		info.Reason = "unsupported param or return type"
		info.Arity = -1
	}
}
