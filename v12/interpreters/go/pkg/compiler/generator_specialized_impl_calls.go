package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) ensureSpecializedImplMethod(method *methodInfo, impl *implMethodInfo, bindings map[string]ast.TypeExpression) (*methodInfo, bool) {
	if g == nil || method == nil || method.Info == nil || impl == nil || len(bindings) == 0 {
		return nil, false
	}
	baseInfo := impl.Info
	if baseInfo == nil {
		baseInfo = method.Info
	}
	methodCopy := *method
	methodCopy.Info = baseInfo
	methodCopy.ExpectsSelf = method.ExpectsSelf || methodDefinitionExpectsSelf(baseInfo.Definition)
	if methodCopy.ReceiverType == "" && methodCopy.ExpectsSelf && len(baseInfo.Params) > 0 {
		methodCopy.ReceiverType = baseInfo.Params[0].GoType
	}
	method = &methodCopy
	if canonical := g.canonicalImplSpecializationBindings(method.Info, impl, bindings); len(canonical) > 0 {
		bindings = canonical
	}
	expectsSelf := method.ExpectsSelf || methodDefinitionExpectsSelf(method.Info.Definition)
	genericNames := g.implSpecializationGenericNames(method)
	if selfExpr, ok := bindings["Self"]; ok && selfExpr != nil && impl.TargetType != nil {
		targetBindings := cloneTypeBindings(bindings)
		if iface, _, ok := g.interfaceDefinitionForImpl(impl); ok && iface != nil {
			for name := range g.interfaceSelfBindingNames(iface) {
				delete(targetBindings, name)
			}
		}
		delete(targetBindings, "Self")
		delete(targetBindings, "SelfType")
		expectedSelf := normalizeTypeExprForPackage(g, method.Info.Package, substituteTypeParams(impl.TargetType, targetBindings))
		normalizedSelf := normalizeTypeExprForPackage(g, method.Info.Package, substituteTypeParams(selfExpr, bindings))
		if expectedSelf != nil && normalizedSelf != nil {
			if normalizeTypeExprString(g, method.Info.Package, normalizedSelf) != normalizeTypeExprString(g, method.Info.Package, expectedSelf) &&
				!g.nominalTargetTypeExprCompatible(method.Info.Package, normalizedSelf, expectedSelf) {
				return nil, false
			}
		}
	}
	concreteTarget := g.specializedImplTargetType(impl, bindings)
	if concreteTarget != nil && impl.TargetType != nil {
		targetBindings := cloneTypeBindings(bindings)
		if iface, _, ok := g.interfaceDefinitionForImpl(impl); ok && iface != nil {
			for name := range g.interfaceSelfBindingNames(iface) {
				delete(targetBindings, name)
			}
		}
		delete(targetBindings, "Self")
		delete(targetBindings, "SelfType")
		expectedTarget := normalizeTypeExprForPackage(g, method.Info.Package, substituteTypeParams(impl.TargetType, targetBindings))
		if expectedTarget != nil &&
			normalizeTypeExprString(g, method.Info.Package, concreteTarget) != normalizeTypeExprString(g, method.Info.Package, expectedTarget) &&
			!g.nominalTargetTypeExprCompatible(method.Info.Package, concreteTarget, expectedTarget) &&
			!g.nominalTargetTypeExprCompatible(method.Info.Package, expectedTarget, concreteTarget) {
			return nil, false
		}
	}
	if concreteTarget != nil {
		if selfExpr, ok := bindings["Self"]; ok && selfExpr != nil {
			normalizedSelf := normalizeTypeExprForPackage(g, method.Info.Package, substituteTypeParams(selfExpr, bindings))
			normalizedTarget := normalizeTypeExprForPackage(g, method.Info.Package, concreteTarget)
			if normalizeTypeExprString(g, method.Info.Package, normalizedSelf) != normalizeTypeExprString(g, method.Info.Package, normalizedTarget) {
				return nil, false
			}
		}
	}
	fillBindings := g.implTypeBindings(method.Info.Package, impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, concreteTarget)
	if fillBindings == nil {
		fillBindings = make(map[string]ast.TypeExpression, len(bindings))
	}
	if iface, _, ok := g.interfaceDefinitionForImpl(impl); ok && iface != nil && concreteTarget != nil {
		for name, expr := range g.interfaceSelfTypeBindings(iface, concreteTarget) {
			if expr == nil {
				continue
			}
			fillBindings[name] = normalizeTypeExprForPackage(g, method.Info.Package, expr)
		}
	}
	if concreteTarget != nil {
		if !g.seedImplBindingsFromConcreteTarget(method, impl, concreteTarget, fillBindings) {
			return nil, false
		}
	}
	for name, expr := range bindings {
		if expr == nil {
			continue
		}
		normalized := normalizeTypeExprForPackage(g, method.Info.Package, expr)
		if existing, ok := fillBindings[name]; ok && existing != nil {
			existing = normalizeTypeExprForPackage(g, method.Info.Package, existing)
			if simple, ok := existing.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == name {
				fillBindings[name] = normalized
				continue
			}
			existingHasGeneric := g.typeExprHasGeneric(existing, genericNames)
			normalizedHasGeneric := g.typeExprHasGeneric(normalized, genericNames)
			if existingHasGeneric && !normalizedHasGeneric {
				fillBindings[name] = normalized
				continue
			}
			if !existingHasGeneric && normalizedHasGeneric {
				continue
			}
			if normalizeTypeExprString(g, method.Info.Package, existing) != normalizeTypeExprString(g, method.Info.Package, normalized) {
				return nil, false
			}
			continue
		}
		fillBindings[name] = normalized
	}
	if concreteTarget != nil {
		fillBindings["Self"] = normalizeTypeExprForPackage(g, method.Info.Package, concreteTarget)
	}
	if !g.specializedImplBindingsAreConcrete(method.Info.Package, method, fillBindings) {
		return nil, false
	}
	if g.specializedImplSignatureUsesUnresolvedNominalStructs(method, impl, concreteTarget, fillBindings) {
		return nil, false
	}
	key := g.specializedImplFunctionKey(method.Info, fillBindings)
	if existing, ok := g.reusableSpecializedFunctionInfo(key, method.Info); ok {
		fillBindings = g.preserveConcreteImplSpecializationBindings(existing.Package, g.implSpecializationGenericNames(method), existing.TypeBindings, fillBindings)
		existing.TypeBindings = cloneTypeBindings(fillBindings)
		mapper := NewTypeMapper(g, existing.Package)
		g.fillImplMethodInfo(existing, mapper, concreteTarget, fillBindings)
		g.invalidateFunctionDerivedInfo(existing)
		g.refreshRepresentableFunctionInfo(existing)
		if !existing.SupportedTypes {
			return nil, false
		}
		receiverType := method.ReceiverType
		if expectsSelf && len(existing.Params) > 0 {
			receiverType = existing.Params[0].GoType
		}
		return &methodInfo{
			TargetName:   method.TargetName,
			TargetType:   g.specializedImplTargetType(impl, bindings),
			MethodName:   method.MethodName,
			ReceiverType: receiverType,
			ExpectsSelf:  expectsSelf,
			Info:         existing,
		}, true
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
	specialized.TypeBindings = cloneTypeBindings(fillBindings)
	g.fillImplMethodInfo(specialized, mapper, concreteTarget, fillBindings)
	g.invalidateFunctionDerivedInfo(specialized)
	g.refreshRepresentableFunctionInfo(specialized)
	if !specialized.SupportedTypes {
		return nil, false
	}
	if expectsSelf && concreteTarget != nil {
		if len(specialized.Params) == 0 || specialized.Params[0].TypeExpr == nil {
			return nil, false
		}
		if normalizeTypeExprString(g, method.Info.Package, specialized.Params[0].TypeExpr) != normalizeTypeExprString(g, method.Info.Package, concreteTarget) {
			return nil, false
		}
		if expectedReceiverGoType, ok := g.lowerCarrierTypeInPackage(method.Info.Package, concreteTarget); ok && expectedReceiverGoType != "" && expectedReceiverGoType != "runtime.Value" && expectedReceiverGoType != "any" {
			if specialized.Params[0].GoType != expectedReceiverGoType {
				return nil, false
			}
		}
		if selfExpr, ok := specialized.TypeBindings["Self"]; ok && selfExpr != nil {
			if normalizeTypeExprString(g, method.Info.Package, selfExpr) != normalizeTypeExprString(g, method.Info.Package, concreteTarget) {
				return nil, false
			}
		}
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
	receiverType := method.ReceiverType
	if expectsSelf && len(specialized.Params) > 0 {
		receiverType = specialized.Params[0].GoType
	}
	return &methodInfo{
		TargetName:   method.TargetName,
		TargetType:   concreteTarget,
		MethodName:   method.MethodName,
		ReceiverType: receiverType,
		ExpectsSelf:  expectsSelf,
		Info:         specialized,
	}, true
}

func (g *generator) preserveConcreteImplSpecializationBindings(pkgName string, genericNames map[string]struct{}, existing map[string]ast.TypeExpression, candidate map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || len(existing) == 0 {
		return candidate
	}
	if candidate == nil {
		candidate = make(map[string]ast.TypeExpression)
	}
	for name, existingExpr := range existing {
		if existingExpr == nil {
			continue
		}
		normalizedExisting := normalizeTypeExprForPackage(g, pkgName, existingExpr)
		if normalizedExisting == nil || !g.typeExprFullyBound(pkgName, normalizedExisting) {
			continue
		}
		candidateExpr := normalizeTypeExprForPackage(g, pkgName, candidate[name])
		if candidateExpr == nil || !g.typeExprFullyBound(pkgName, candidateExpr) || g.typeExprHasGeneric(candidateExpr, genericNames) {
			candidate[name] = normalizedExisting
		}
	}
	if len(candidate) == 0 {
		return nil
	}
	return candidate
}

func (g *generator) specializedImplBindingsAreConcrete(pkgName string, method *methodInfo, bindings map[string]ast.TypeExpression) bool {
	if g == nil || method == nil || len(bindings) == 0 {
		return false
	}
	genericNames := g.implSpecializationGenericNames(method)
	if len(genericNames) == 0 {
		return false
	}
	for name := range genericNames {
		expr, ok := bindings[name]
		if !ok || expr == nil {
			continue
		}
		normalized := normalizeTypeExprForPackage(g, pkgName, expr)
		if g.typeExprHasGeneric(normalized, genericNames) {
			return false
		}
		if !g.typeExprFullyBound(pkgName, normalized) {
			return false
		}
	}
	return true
}

func (g *generator) specializedImplTargetType(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || impl == nil || impl.TargetType == nil {
		return nil
	}
	var target ast.TypeExpression
	if generic, ok := impl.TargetType.(*ast.GenericTypeExpression); ok && generic != nil {
		args := make([]ast.TypeExpression, 0, len(generic.Arguments))
		for _, arg := range generic.Arguments {
			if arg == nil {
				target = impl.TargetType
				break
			}
			args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(arg, bindings)))
		}
		if target == nil {
			target = normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(generic.Base, args))
		}
	} else {
		target = normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(impl.TargetType, bindings))
	}
	if target == nil {
		target = impl.TargetType
	}
	if selfExpr, ok := bindings["Self"]; ok && selfExpr != nil {
		normalizedSelf := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(selfExpr, bindings))
		if normalizedSelf != nil {
			if target == nil {
				return normalizedSelf
			}
			if normalizeTypeExprString(g, impl.Info.Package, normalizedSelf) == normalizeTypeExprString(g, impl.Info.Package, target) ||
				g.nominalTargetTypeExprCompatible(impl.Info.Package, normalizedSelf, target) ||
				g.nominalTargetTypeExprCompatible(impl.Info.Package, target, normalizedSelf) {
				return normalizedSelf
			}
		}
	}
	if impl.Info != nil && len(impl.Info.Params) > 0 && impl.Info.Params[0].TypeExpr != nil {
		receiverExpr := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(impl.Info.Params[0].TypeExpr, bindings))
		if receiverExpr != nil {
			if target == nil {
				return receiverExpr
			}
			if normalizeTypeExprString(g, impl.Info.Package, receiverExpr) == normalizeTypeExprString(g, impl.Info.Package, target) ||
				g.nominalTargetTypeExprCompatible(impl.Info.Package, receiverExpr, target) ||
				g.nominalTargetTypeExprCompatible(impl.Info.Package, target, receiverExpr) {
				return receiverExpr
			}
		}
	}
	baseName, ok := typeExprBaseName(target)
	if !ok || baseName == "" {
		return target
	}
	if baseName == "Array" {
		if concrete := g.specializedBuiltinArrayTargetType(impl, bindings); concrete != nil {
			return concrete
		}
	}
	info, ok := g.structInfoForTypeName(impl.Info.Package, baseName)
	if !ok || info == nil || info.Node == nil || len(info.Node.GenericParams) == 0 {
		return target
	}
	if len(impl.InterfaceArgs) == len(info.Node.GenericParams) {
		args := make([]ast.TypeExpression, 0, len(impl.InterfaceArgs))
		for _, arg := range impl.InterfaceArgs {
			if arg == nil {
				return target
			}
			concreteArg := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(arg, bindings))
			if concreteArg == nil {
				return target
			}
			args = append(args, concreteArg)
		}
		return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
	}
	args := make([]ast.TypeExpression, 0, len(info.Node.GenericParams))
	for _, gp := range info.Node.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			return target
		}
		bound, ok := bindings[gp.Name.Name]
		if !ok || bound == nil {
			return target
		}
		args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, bound))
	}
	return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
}

func (g *generator) specializedImplTargetTemplate(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || impl == nil || impl.TargetType == nil {
		return nil
	}
	var target ast.TypeExpression
	if generic, ok := impl.TargetType.(*ast.GenericTypeExpression); ok && generic != nil {
		args := make([]ast.TypeExpression, 0, len(generic.Arguments))
		for _, arg := range generic.Arguments {
			if arg == nil {
				target = impl.TargetType
				break
			}
			args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(arg, bindings)))
		}
		if target == nil {
			target = normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(generic.Base, args))
		}
	} else {
		target = normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(impl.TargetType, bindings))
	}
	if target == nil {
		target = impl.TargetType
	}
	if selfExpr, ok := bindings["Self"]; ok && selfExpr != nil {
		normalizedSelf := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(selfExpr, bindings))
		if normalizedSelf != nil {
			if target == nil {
				return normalizedSelf
			}
			if normalizeTypeExprString(g, impl.Info.Package, normalizedSelf) == normalizeTypeExprString(g, impl.Info.Package, target) ||
				g.nominalTargetTypeExprCompatible(impl.Info.Package, normalizedSelf, target) ||
				g.nominalTargetTypeExprCompatible(impl.Info.Package, target, normalizedSelf) {
				return normalizedSelf
			}
		}
	}
	if impl.Info != nil && len(impl.Info.Params) > 0 && impl.Info.Params[0].TypeExpr != nil {
		receiverExpr := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(impl.Info.Params[0].TypeExpr, bindings))
		if receiverExpr != nil {
			if target == nil {
				return receiverExpr
			}
			if normalizeTypeExprString(g, impl.Info.Package, receiverExpr) == normalizeTypeExprString(g, impl.Info.Package, target) ||
				g.nominalTargetTypeExprCompatible(impl.Info.Package, receiverExpr, target) ||
				g.nominalTargetTypeExprCompatible(impl.Info.Package, target, receiverExpr) {
				return receiverExpr
			}
		}
	}
	baseName, ok := typeExprBaseName(target)
	if !ok || baseName == "" {
		return target
	}
	if baseName == "Array" {
		if concrete := g.specializedBuiltinArrayTargetType(impl, bindings); concrete != nil {
			return concrete
		}
	}
	info, ok := g.structInfoForTypeName(impl.Info.Package, baseName)
	if !ok || info == nil || info.Node == nil || len(info.Node.GenericParams) == 0 {
		return target
	}
	if len(impl.InterfaceArgs) == len(info.Node.GenericParams) {
		args := make([]ast.TypeExpression, 0, len(impl.InterfaceArgs))
		for _, arg := range impl.InterfaceArgs {
			if arg == nil {
				return target
			}
			args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(arg, bindings)))
		}
		return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
	}
	args := make([]ast.TypeExpression, 0, len(info.Node.GenericParams))
	for _, gp := range info.Node.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			return target
		}
		bound, ok := bindings[gp.Name.Name]
		if !ok || bound == nil {
			return target
		}
		args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, bound))
	}
	return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
}

func (g *generator) specializedBuiltinArrayTargetType(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || impl == nil {
		return nil
	}
	baseName, ok := typeExprBaseName(impl.TargetType)
	if !ok || baseName != "Array" {
		return nil
	}
	if len(impl.InterfaceArgs) == 1 {
		arg := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(impl.InterfaceArgs[0], bindings))
		if arg != nil {
			return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty("Array"), []ast.TypeExpression{arg}))
		}
	}
	if len(impl.ImplGenerics) == 1 {
		gp := impl.ImplGenerics[0]
		if gp != nil && gp.Name != nil && gp.Name.Name != "" {
			if bound, ok := bindings[gp.Name.Name]; ok && bound != nil {
				return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty("Array"), []ast.TypeExpression{
					normalizeTypeExprForPackage(g, impl.Info.Package, bound),
				}))
			}
		}
	}
	return nil
}

func removeSpecializedFunction(list []*functionInfo, target *functionInfo) []*functionInfo {
	for idx, info := range list {
		if info != target {
			continue
		}
		copy(list[idx:], list[idx+1:])
		list[len(list)-1] = nil
		return list[:len(list)-1]
	}
	return list
}

func (g *generator) implSiblingsForFunction(info *functionInfo) map[string]implSiblingInfo {
	if g == nil || info == nil {
		return nil
	}
	implInfo := g.implMethodByInfo[info]
	if implInfo == nil || !implInfo.IsDefault {
		return nil
	}
	siblings := g.implSiblingsForDefault(implInfo)
	if len(siblings) == 0 || len(info.TypeBindings) == 0 {
		return siblings
	}
	currentBindings := g.compileContextTypeBindings(info)
	if len(currentBindings) == 0 {
		return siblings
	}
	selfConcrete := false
	if actualSelfType := g.implConcreteSelfTypeExpr(info, implInfo, currentBindings); actualSelfType != nil {
		normalizedSelf := normalizeTypeExprForPackage(g, info.Package, actualSelfType)
		selfConcrete = g.typeExprFullyBound(info.Package, normalizedSelf)
		if selfConcrete && len(info.Params) > 0 && info.Params[0].GoType != "" {
			if canonicalSelfGoType, ok := g.lowerCarrierTypeInPackage(info.Package, normalizedSelf); ok && canonicalSelfGoType != "" && canonicalSelfGoType != "runtime.Value" && canonicalSelfGoType != "any" {
				selfConcrete = canonicalSelfGoType == info.Params[0].GoType
			}
		}
	}
	out := make(map[string]implSiblingInfo, len(siblings))
	for name, sibling := range siblings {
		if sibling.Info == nil {
			out[name] = sibling
			continue
		}
		siblingImpl := g.implMethodByInfo[sibling.Info]
		if siblingImpl == nil {
			out[name] = sibling
			continue
		}
		methodName := siblingImpl.MethodName
		if methodName == "" {
			methodName = name
		}
		if !selfConcrete {
			out[name] = sibling
			continue
		}
		specializedBindings := g.implSiblingBindingsForFunction(info, implInfo, siblingImpl, currentBindings)
		if len(specializedBindings) == 0 {
			out[name] = sibling
			continue
		}
		specialized, ok := g.ensureSpecializedImplMethod(&methodInfo{
			MethodName:  methodName,
			ExpectsSelf: methodDefinitionExpectsSelf(sibling.Info.Definition),
			Info:        sibling.Info,
		}, siblingImpl, specializedBindings)
		if !ok || specialized == nil || specialized.Info == nil {
			out[name] = sibling
			continue
		}
		sibling.GoName = specialized.Info.GoName
		sibling.Arity = specialized.Info.Arity
		sibling.Info = specialized.Info
		out[name] = sibling
	}
	return out
}

func (g *generator) implSiblingBindingsForFunction(info *functionInfo, currentImpl *implMethodInfo, siblingImpl *implMethodInfo, currentBindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || info == nil || currentImpl == nil || siblingImpl == nil || siblingImpl.Info == nil {
		return nil
	}
	genericNames := g.implSpecializationGenericNames(&methodInfo{
		TargetType:  siblingImpl.TargetType,
		MethodName:  siblingImpl.MethodName,
		ExpectsSelf: methodDefinitionExpectsSelf(siblingImpl.Info.Definition),
		Info:        siblingImpl.Info,
	})
	bindings := g.mergeConcreteTypeBindings(siblingImpl.Info.Package, genericNames, nil, currentBindings)
	var receiverBindings map[string]ast.TypeExpression
	actualSelfType := g.implConcreteSelfTypeExpr(info, currentImpl, currentBindings)
	if actualSelfType != nil && siblingImpl.TargetType != nil {
		targetTemplate := g.specializedImplTargetTemplate(siblingImpl, bindings)
		if targetTemplate == nil {
			targetTemplate = siblingImpl.TargetType
		}
		selfBindings := make(map[string]ast.TypeExpression)
		matched := g.specializedTypeTemplateMatches(
			siblingImpl.Info.Package,
			targetTemplate,
			actualSelfType,
			genericNames,
			selfBindings,
			make(map[string]struct{}),
		)
		if !matched {
			_ = g.specializedSameBaseGenericBindings(
				siblingImpl.Info.Package,
				targetTemplate,
				actualSelfType,
				genericNames,
				selfBindings,
			)
		}
		if len(selfBindings) > 0 {
			if bindings == nil {
				bindings = make(map[string]ast.TypeExpression, len(selfBindings))
			}
			for name, expr := range selfBindings {
				if expr == nil {
					continue
				}
				if _, ok := genericNames[name]; len(genericNames) > 0 && !ok {
					continue
				}
				if _, exists := bindings[name]; exists {
					continue
				}
				bindings[name] = normalizeTypeExprForPackage(g, siblingImpl.Info.Package, expr)
			}
		}
		receiverBindings = g.normalizeConcreteTypeBindings(siblingImpl.Info.Package, selfBindings, genericNames)
		bindings = g.mergeConcreteTypeBindings(siblingImpl.Info.Package, genericNames, bindings, receiverBindings)
	}
	actualInterfaceExpr := g.implConcreteInterfaceExpr(currentImpl, currentBindings)
	if actualInterfaceExpr != nil {
		interfaceBindings, ok := g.nativeInterfaceImplBindingsForTarget(
			currentImpl.Info.Package,
			actualInterfaceExpr,
			genericParamNameSet(siblingImpl.InterfaceGenerics),
			siblingImpl.Info.Package,
			siblingImpl.InterfaceName,
			siblingImpl.InterfaceArgs,
			make(map[string]struct{}),
		)
		if ok {
			bindings = g.mergeConcreteTypeBindings(siblingImpl.Info.Package, genericNames, bindings, interfaceBindings)
		}
	}
	normalized := g.normalizeConcreteTypeBindings(siblingImpl.Info.Package, bindings, genericNames)
	if len(normalized) == 0 && len(receiverBindings) > 0 {
		return receiverBindings
	}
	return normalized
}

func (g *generator) implSpecializationGenericNames(method *methodInfo) map[string]struct{} {
	if method == nil || method.Info == nil {
		return nil
	}
	return mergeGenericNameSets(g.callableGenericNames(method.Info), g.methodGenericNames(method))
}

func (g *generator) implConcreteSelfTypeExpr(info *functionInfo, impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || info == nil {
		return nil
	}
	if len(info.Params) > 0 && info.Params[0].TypeExpr != nil {
		return normalizeTypeExprForPackage(g, info.Package, info.Params[0].TypeExpr)
	}
	if impl == nil || impl.TargetType == nil {
		return nil
	}
	selfType := g.implSelfTargetType(info.Package, impl.TargetType, bindings)
	selfType = substituteTypeParams(selfType, bindings)
	return normalizeTypeExprForPackage(g, info.Package, selfType)
}

func (g *generator) implConcreteInterfaceExpr(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || impl == nil || impl.InterfaceName == "" {
		return nil
	}
	if len(impl.InterfaceArgs) == 0 {
		return ast.Ty(impl.InterfaceName)
	}
	args := make([]ast.TypeExpression, 0, len(impl.InterfaceArgs))
	for _, arg := range impl.InterfaceArgs {
		if arg == nil {
			return nil
		}
		args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(arg, bindings)))
	}
	return nativeInterfaceInstantiationExpr(impl.InterfaceName, args)
}

func (g *generator) seedImplBindingsFromConcreteTarget(method *methodInfo, impl *implMethodInfo, actualTypeExpr ast.TypeExpression, bindings map[string]ast.TypeExpression) bool {
	if g == nil || method == nil || method.Info == nil || impl == nil || actualTypeExpr == nil {
		return false
	}
	if bindings == nil {
		return false
	}
	if g.bindNominalTargetActualArgs(method.Info.Package, impl.TargetType, impl.InterfaceArgs, actualTypeExpr, bindings) {
		return true
	}
	genericNames := g.implSpecializationGenericNames(method)
	targetTemplate := g.specializedImplTargetTemplate(impl, bindings)
	if targetTemplate == nil {
		targetTemplate = impl.TargetType
	}
	if targetTemplate == nil {
		return false
	}
	targetTemplate = g.normalizeTypeExprForSpecialization(method.Info.Package, targetTemplate, nil)
	actualTypeExpr = g.normalizeTypeExprForSpecialization(method.Info.Package, actualTypeExpr, nil)
	if normalizeTypeExprString(g, method.Info.Package, targetTemplate) == normalizeTypeExprString(g, method.Info.Package, actualTypeExpr) {
		return true
	}
	if templateGeneric, ok := targetTemplate.(*ast.GenericTypeExpression); ok && templateGeneric != nil {
		if actualGeneric, ok := actualTypeExpr.(*ast.GenericTypeExpression); ok && actualGeneric != nil && len(templateGeneric.Arguments) == len(actualGeneric.Arguments) {
			if templateBase, ok := typeExprBaseName(templateGeneric.Base); ok && templateBase != "" {
				if actualBase, ok := typeExprBaseName(actualGeneric.Base); ok && actualBase == templateBase {
					for idx := range templateGeneric.Arguments {
						if !g.specializedBindTemplateArg(method.Info.Package, templateGeneric.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings) {
							return false
						}
					}
					return true
				}
			}
		}
	}
	return g.specializedSameBaseGenericBindings(method.Info.Package, targetTemplate, actualTypeExpr, genericNames, bindings)
}

func (g *generator) specializedImplMethodBindings(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, impl *implMethodInfo, receiverTypeExpr ast.TypeExpression, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || impl == nil {
		return nil, false
	}
	genericNames := g.implSpecializationGenericNames(method)
	bindings := g.concreteCompileContextBindings(method.Info, genericNames)
	bindings = g.mergeConcreteTypeBindings(method.Info.Package, genericNames, bindings, ctx.typeBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if method.ExpectsSelf && len(method.Info.Params) > 0 && method.Info.Params[0].TypeExpr != nil {
		receiverTemplate := method.Info.Params[0].TypeExpr
		if targetTemplate := g.specializedImplTargetTemplate(impl, bindings); targetTemplate != nil {
			if g.preferImplSpecializationTemplate(method.Info.Package, receiverTemplate, targetTemplate) {
				receiverTemplate = targetTemplate
			}
		}
		if !g.specializedTargetMatchesOrDefers(method.Info.Package, receiverTemplate, receiverTypeExpr, genericNames, bindings) {
			seeded := cloneTypeBindings(bindings)
			if seeded == nil {
				seeded = make(map[string]ast.TypeExpression)
			}
			for name := range genericNames {
				delete(seeded, name)
			}
			if !g.seedImplBindingsFromConcreteTarget(method, impl, receiverTypeExpr, seeded) {
				return nil, false
			}
			bindings = seeded
		} else {
			// A bare nominal template can match a concrete instantiated receiver
			// without introducing the receiver's hidden type arguments. Refresh the
			// bindings from the actual receiver so the specialized impl stays on the
			// fully bound carrier instead of falling back to the generic base type.
			seeded := cloneTypeBindings(bindings)
			if seeded == nil {
				seeded = make(map[string]ast.TypeExpression)
			}
			_ = g.seedImplBindingsFromConcreteTarget(method, impl, receiverTypeExpr, seeded)
			bindings = seeded
		}
	}
	if len(call.TypeArguments) > 0 {
		if method.Info.Definition == nil || len(method.Info.Definition.GenericParams) != len(call.TypeArguments) {
			return nil, false
		}
		for idx, arg := range call.TypeArguments {
			if arg == nil {
				return nil, false
			}
			gp := method.Info.Definition.GenericParams[idx]
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				return nil, false
			}
			bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, method.Info.Package, arg)
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
		actualExpr, ok := g.inferExpressionTypeExpr(ctx, arg, "")
		argCtx := ctx.child()
		_, _, actualGoType, actualGoTypeOK := g.compileExprLines(argCtx, arg, "")
		if !ok || actualExpr == nil {
			if !actualGoTypeOK {
				continue
			}
			actualExpr, ok = g.typeExprForGoType(actualGoType)
			if !ok || actualExpr == nil {
				continue
			}
		}
		actualExpr, ok = g.specializationConcreteArgTypeExprForParam(method.Info.Package, paramTypeExpr, actualExpr, actualGoType)
		if !ok || actualExpr == nil {
			continue
		}
		_ = g.applySpecializedTypeTemplateMatch(method.Info.Package, paramTypeExpr, actualExpr, genericNames, bindings)
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil && method.Info.Definition != nil && method.Info.Definition.ReturnType != nil {
		returnExpr := g.functionReturnTypeExprWithBindings(method.Info, bindings)
		if returnExpr != nil {
			_ = g.applySpecializedTypeTemplateMatch(method.Info.Package, returnExpr, expectedExpr, genericNames, bindings)
		}
	}
	bindings = g.normalizeConcreteTypeBindings(method.Info.Package, bindings, nil)
	if len(bindings) == 0 {
		return nil, false
	}
	return bindings, true
}

func (g *generator) specializedStaticImplMethodBindings(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, impl *implMethodInfo, targetTypeExpr ast.TypeExpression, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || impl == nil || targetTypeExpr == nil {
		return nil, false
	}
	genericNames := g.implSpecializationGenericNames(method)
	bindings := g.concreteCompileContextBindings(method.Info, genericNames)
	bindings = g.mergeConcreteTypeBindings(method.Info.Package, genericNames, bindings, ctx.typeBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if genericTarget, ok := targetTypeExpr.(*ast.GenericTypeExpression); ok && genericTarget != nil {
		if targetParams := g.nominalTargetGenericParams(method); len(targetParams) > 0 {
			_ = g.bindGenericTypeArguments(method.Info.Package, bindings, targetParams, genericTarget.Arguments)
		}
	}
	targetTemplate := g.specializedImplTargetTemplate(impl, bindings)
	if targetTemplate == nil {
		targetTemplate = impl.TargetType
	}
	if method.ExpectsSelf && len(method.Info.Params) > 0 && method.Info.Params[0].TypeExpr != nil {
		if !g.preferImplSpecializationTemplate(method.Info.Package, method.Info.Params[0].TypeExpr, targetTemplate) {
			targetTemplate = method.Info.Params[0].TypeExpr
		}
	}
	if targetTemplate != nil {
		if !g.specializedTargetMatchesOrDefers(method.Info.Package, targetTemplate, targetTypeExpr, genericNames, bindings) {
			seeded := cloneTypeBindings(bindings)
			if seeded == nil {
				seeded = make(map[string]ast.TypeExpression)
			}
			for name := range genericNames {
				delete(seeded, name)
			}
			if !g.seedImplBindingsFromConcreteTarget(method, impl, targetTypeExpr, seeded) {
				return nil, false
			}
			bindings = seeded
		} else {
			seeded := cloneTypeBindings(bindings)
			if seeded == nil {
				seeded = make(map[string]ast.TypeExpression)
			}
			_ = g.seedImplBindingsFromConcreteTarget(method, impl, targetTypeExpr, seeded)
			bindings = seeded
		}
	}
	if len(call.TypeArguments) > 0 {
		if method.Info.Definition == nil || len(method.Info.Definition.GenericParams) != len(call.TypeArguments) {
			return nil, false
		}
		for idx, arg := range call.TypeArguments {
			if arg == nil {
				return nil, false
			}
			gp := method.Info.Definition.GenericParams[idx]
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				return nil, false
			}
			bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, method.Info.Package, arg)
		}
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil && method.Info.Definition != nil && method.Info.Definition.ReturnType != nil {
		returnExpr := g.functionReturnTypeExprWithBindings(method.Info, bindings)
		if returnExpr != nil {
			_ = g.applySpecializedTypeTemplateMatch(method.Info.Package, returnExpr, expectedExpr, genericNames, bindings)
		}
	}
	for idx, arg := range call.Arguments {
		if idx >= len(method.Info.Params) {
			break
		}
		paramTypeExpr := method.Info.Params[idx].TypeExpr
		if paramTypeExpr == nil {
			continue
		}
		actualExpr, ok := g.inferExpressionTypeExpr(ctx, arg, "")
		argCtx := ctx.child()
		_, _, actualGoType, actualGoTypeOK := g.compileExprLines(argCtx, arg, "")
		if !ok || actualExpr == nil {
			if !actualGoTypeOK {
				continue
			}
			actualExpr, ok = g.typeExprForGoType(actualGoType)
			if !ok || actualExpr == nil {
				continue
			}
		}
		actualExpr, ok = g.specializationConcreteArgTypeExprForParam(method.Info.Package, paramTypeExpr, actualExpr, actualGoType)
		if !ok || actualExpr == nil {
			continue
		}
		_ = g.applySpecializedTypeTemplateMatch(method.Info.Package, paramTypeExpr, actualExpr, genericNames, bindings)
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil && method.Info.Definition != nil && method.Info.Definition.ReturnType != nil {
		returnExpr := g.functionReturnTypeExprWithBindings(method.Info, bindings)
		if returnExpr != nil {
			_ = g.applySpecializedTypeTemplateMatch(method.Info.Package, returnExpr, expectedExpr, genericNames, bindings)
		}
	}
	bindings = g.normalizeConcreteTypeBindings(method.Info.Package, bindings, nil)
	if len(bindings) == 0 {
		return nil, false
	}
	return bindings, true
}
