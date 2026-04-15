package compiler

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

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

func (g *generator) preferImplSpecializationTemplate(pkgName string, base ast.TypeExpression, candidate ast.TypeExpression) bool {
	if g == nil || candidate == nil {
		return false
	}
	if base == nil {
		return true
	}
	base = normalizeTypeExprForPackage(g, pkgName, base)
	candidate = normalizeTypeExprForPackage(g, pkgName, candidate)
	if _, ok := base.(*ast.GenericTypeExpression); ok {
		if _, ok := candidate.(*ast.GenericTypeExpression); !ok {
			return false
		}
	}
	return true
}

func (g *generator) staticReceiverTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if g == nil {
		return nil, false
	}
	if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, goType); ok && inferred != nil {
		inferred = normalizeTypeExprForPackage(g, ctx.packageName, inferred)
		if preferred := g.preferConcreteTypeExprForGoType(ctx, inferred, goType); preferred != nil {
			if g.typeExprCompatibleWithCarrier(ctx, preferred, goType) {
				return preferred, true
			}
			return nil, false
		}
		if g.typeExprCompatibleWithCarrier(ctx, inferred, goType) {
			return inferred, true
		}
		return nil, false
	}
	if preferred := g.preferConcreteTypeExprForGoType(ctx, nil, goType); preferred != nil {
		if g.typeExprCompatibleWithCarrier(ctx, preferred, goType) {
			return preferred, true
		}
		return nil, false
	}
	return nil, false
}

func (g *generator) typeExprCompatibleWithCarrier(ctx *compileContext, expr ast.TypeExpression, goType string) bool {
	if g == nil || expr == nil {
		return false
	}
	if strings.TrimSpace(goType) == "" || goType == "runtime.Value" || goType == "any" {
		return true
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	if expr == nil {
		return false
	}
	if !g.typeExprFullyBound(pkgName, expr) {
		return true
	}
	canonicalGoType, ok := g.lowerCarrierTypeInPackage(pkgName, expr)
	if !ok || canonicalGoType == "" || canonicalGoType == "runtime.Value" || canonicalGoType == "any" {
		return false
	}
	if canonicalGoType == goType {
		return true
	}
	if g.sameNominalStructFamily(goType, canonicalGoType) || g.sameNominalStructFamily(canonicalGoType, goType) {
		return true
	}
	if g.staticArrayCarrierCoercible(goType, canonicalGoType) || g.staticArrayCarrierCoercible(canonicalGoType, goType) {
		return true
	}
	return g.receiverGoTypeCompatible(canonicalGoType, goType) || g.receiverGoTypeCompatible(goType, canonicalGoType)
}

func (g *generator) staticTargetTypeExpr(ctx *compileContext, expr ast.Expression) (ast.TypeExpression, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil && ident.Name != "" && ctx != nil {
		if _, exists := ctx.lookup(ident.Name); !exists {
			if bound, ok := ctx.typeBindings[ident.Name]; ok && bound != nil {
				return normalizeTypeExprForPackage(g, ctx.packageName, bound), true
			}
		}
	}
	if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, ""); ok && inferred != nil {
		inferred = normalizeTypeExprForPackage(g, ctx.packageName, inferred)
		if preferred := g.preferConcreteTypeExprForGoType(ctx, inferred, ""); preferred != nil {
			return preferred, true
		}
		return inferred, true
	}
	return nil, false
}

func (g *generator) preferConcreteTypeExprForGoType(ctx *compileContext, inferred ast.TypeExpression, goType string) ast.TypeExpression {
	if g == nil || strings.TrimSpace(goType) == "" {
		return inferred
	}
	concrete, ok := g.typeExprForGoType(goType)
	if !ok || concrete == nil {
		return inferred
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	concretePkg := g.resolvedTypeExprPackage(pkgName, concrete)
	concrete = normalizeTypeExprForPackage(g, concretePkg, concrete)
	if concrete == nil || (!g.typeExprFullyBound(pkgName, concrete) && !g.typeExprFullyBound(concretePkg, concrete)) {
		return inferred
	}
	if inferred == nil {
		return concrete
	}
	inferredPkg := g.resolvedTypeExprPackage(pkgName, inferred)
	if !g.typeExprFullyBound(pkgName, inferred) && !g.typeExprFullyBound(inferredPkg, inferred) {
		return concrete
	}
	if ctx != nil && g.typeExprHasGeneric(inferred, ctx.genericNames) {
		return concrete
	}
	return inferred
}

func (g *generator) inferExpressionTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if g == nil {
		return nil, false
	}
	if inferred, ok := g.inferLocalTypeExpr(ctx, expr, goType); ok && inferred != nil {
		return g.lowerNormalizedTypeExpr(ctx, inferred), true
	}
	if goType != "" {
		if inferred, ok := g.typeExprForGoType(goType); ok && inferred != nil {
			return g.lowerNormalizedTypeExpr(ctx, inferred), true
		}
	}
	return nil, false
}

func (g *generator) functionReturnTypeExpr(info *functionInfo) ast.TypeExpression {
	return g.functionReturnTypeExprWithBindings(info, g.compileContextTypeBindings(info))
}

func (g *generator) functionReturnTypeExprWithBindings(info *functionInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || info == nil || info.Definition == nil {
		return nil
	}
	retExpr := g.functionDeclaredOrInferredReturnTypeExpr(info)
	if retExpr == nil {
		return nil
	}
	if impl := g.implMethodByInfo[info]; impl != nil {
		concreteTarget := g.specializedImplTargetType(impl, bindings)
		if concreteTarget == nil {
			concreteTarget = impl.TargetType
		}
		interfaceBindings := g.implTypeBindings(info.Package, impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, concreteTarget)
		selfTarget := g.implSelfTargetType(info.Package, concreteTarget, interfaceBindings)
		allBindings := g.mergeImplSelfTargetBindings(info.Package, concreteTarget, selfTarget, interfaceBindings)
		for name, expr := range bindings {
			if expr == nil {
				continue
			}
			if name == "Self" && selfTarget != nil {
				continue
			}
			if allBindings == nil {
				allBindings = make(map[string]ast.TypeExpression)
			}
			allBindings[name] = normalizeTypeExprForPackage(g, info.Package, expr)
		}
		if selfTarget != nil {
			if allBindings == nil {
				allBindings = make(map[string]ast.TypeExpression)
			}
			allBindings["Self"] = normalizeTypeExprForPackage(g, info.Package, selfTarget)
		}
		retExpr = resolveSelfTypeExpr(retExpr, selfTarget)
		retExpr = substituteTypeParams(retExpr, allBindings)
		return normalizeTypeExprForPackage(g, info.Package, retExpr)
	}
	retExpr = substituteTypeParams(retExpr, bindings)
	return normalizeTypeExprForPackage(g, info.Package, retExpr)
}

func (g *generator) concreteCompileContextBindings(info *functionInfo, genericNames map[string]struct{}) map[string]ast.TypeExpression {
	return g.normalizeConcreteTypeBindings(info.Package, g.compileContextTypeBindings(info), genericNames)
}

func (g *generator) mergeConcreteTypeBindings(pkgName string, genericNames map[string]struct{}, base map[string]ast.TypeExpression, extra map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || len(extra) == 0 {
		return base
	}
	if base == nil {
		base = make(map[string]ast.TypeExpression)
	}
	for name, expr := range extra {
		if expr == nil {
			continue
		}
		if len(genericNames) > 0 {
			if _, ok := genericNames[name]; !ok {
				continue
			}
		}
		if _, exists := base[name]; exists {
			continue
		}
		base[name] = normalizeTypeExprForPackage(g, pkgName, expr)
	}
	return base
}

func (g *generator) normalizeConcreteTypeBindings(pkgName string, bindings map[string]ast.TypeExpression, genericNames map[string]struct{}) map[string]ast.TypeExpression {
	if g == nil || len(bindings) == 0 {
		return nil
	}
	out := make(map[string]ast.TypeExpression, len(bindings))
	for name, expr := range bindings {
		if len(genericNames) > 0 {
			if _, ok := genericNames[name]; !ok {
				continue
			}
		}
		if expr == nil {
			continue
		}
		resolvedPkg := g.resolvedTypeExprPackage(pkgName, expr)
		normalized := normalizeTypeExprForPackage(g, resolvedPkg, expr)
		normalized = g.recordResolvedTypeExprPackage(normalized, resolvedPkg)
		mapped, ok := g.lowerCarrierTypeInPackage(resolvedPkg, normalized)
		mapped, ok = g.recoverRepresentableCarrierType(resolvedPkg, normalized, mapped)
		if !g.typeExprFullyBound(pkgName, normalized) && !g.typeExprFullyBound(resolvedPkg, normalized) &&
			(!ok || mapped == "" || mapped == "runtime.Value" || mapped == "any") {
			continue
		}
		if simple, ok := normalized.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == name {
			continue
		}
		if g.typeExprHasGeneric(normalized, genericNames) {
			continue
		}
		out[name] = normalized
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneTypeBindings(bindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if len(bindings) == 0 {
		return nil
	}
	out := make(map[string]ast.TypeExpression, len(bindings))
	for name, expr := range bindings {
		out[name] = expr
	}
	return out
}

func (g *generator) specializedTypeTemplateMatches(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	if len(bindings) > 0 {
		template = substituteTypeParams(template, bindings)
		actual = substituteTypeParams(actual, bindings)
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	return g.specializedTypeTemplateMatchesNormalized(pkgName, template, actual, genericNames, bindings, seen)
}

func (g *generator) specializedSameBaseGenericBindings(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	templateGeneric, ok := template.(*ast.GenericTypeExpression)
	if !ok || templateGeneric == nil {
		return false
	}
	actualGeneric, ok := actual.(*ast.GenericTypeExpression)
	if !ok || actualGeneric == nil || len(templateGeneric.Arguments) != len(actualGeneric.Arguments) {
		return false
	}
	if normalizeTypeExprString(g, pkgName, templateGeneric.Base) != normalizeTypeExprString(g, pkgName, actualGeneric.Base) {
		return false
	}
	for idx := range templateGeneric.Arguments {
		if !g.specializedBindTemplateArg(pkgName, templateGeneric.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings) {
			return false
		}
	}
	return true
}

func (g *generator) specializedBindTemplateArg(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	switch tt := template.(type) {
	case *ast.SimpleTypeExpression:
		if tt == nil || tt.Name == nil || tt.Name.Name == "" {
			return false
		}
		if _, ok := genericNames[tt.Name.Name]; ok {
			if bound, exists := bindings[tt.Name.Name]; exists {
				if normalizeTypeExprString(g, pkgName, bound) == tt.Name.Name && normalizeTypeExprString(g, pkgName, actual) != tt.Name.Name {
					bindings[tt.Name.Name] = actual
					return true
				}
				boundKey := normalizeTypeExprIdentityKey(g, pkgName, bound)
				actualKey := normalizeTypeExprIdentityKey(g, pkgName, actual)
				if boundKey != "" && actualKey != "" && boundKey != actualKey &&
					normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual) &&
					g.typeExprFullyBound(pkgName, actual) {
					bindings[tt.Name.Name] = actual
					return true
				}
				return normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual)
			}
			bindings[tt.Name.Name] = actual
			return true
		}
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	case *ast.GenericTypeExpression:
		actualGeneric, ok := actual.(*ast.GenericTypeExpression)
		if !ok || actualGeneric == nil || len(tt.Arguments) != len(actualGeneric.Arguments) {
			return false
		}
		if normalizeTypeExprString(g, pkgName, tt.Base) != normalizeTypeExprString(g, pkgName, actualGeneric.Base) {
			return false
		}
		for idx := range tt.Arguments {
			if !g.specializedBindTemplateArg(pkgName, tt.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		actualNullable, ok := actual.(*ast.NullableTypeExpression)
		return ok && actualNullable != nil && g.specializedBindTemplateArg(pkgName, tt.InnerType, actualNullable.InnerType, genericNames, bindings)
	case *ast.ResultTypeExpression:
		actualResult, ok := actual.(*ast.ResultTypeExpression)
		return ok && actualResult != nil && g.specializedBindTemplateArg(pkgName, tt.InnerType, actualResult.InnerType, genericNames, bindings)
	case *ast.UnionTypeExpression:
		if actualUnion, ok := actual.(*ast.UnionTypeExpression); ok && actualUnion != nil {
			if len(tt.Members) != len(actualUnion.Members) {
				return false
			}
			for idx := range tt.Members {
				if !g.specializedBindTemplateArg(pkgName, tt.Members[idx], actualUnion.Members[idx], genericNames, bindings) {
					return false
				}
			}
			return true
		}
		for _, member := range tt.Members {
			candidate := cloneTypeBindings(bindings)
			if candidate == nil {
				candidate = make(map[string]ast.TypeExpression)
			}
			if !g.specializedBindTemplateArg(pkgName, member, actual, genericNames, candidate) {
				continue
			}
			applyTypeBindings(bindings, candidate)
			return true
		}
		return false
	case *ast.FunctionTypeExpression:
		actualFn, ok := actual.(*ast.FunctionTypeExpression)
		if !ok || actualFn == nil || len(tt.ParamTypes) != len(actualFn.ParamTypes) {
			return false
		}
		for idx := range tt.ParamTypes {
			if !g.specializedBindTemplateArg(pkgName, tt.ParamTypes[idx], actualFn.ParamTypes[idx], genericNames, bindings) {
				return false
			}
		}
		return g.specializedBindTemplateArg(pkgName, tt.ReturnType, actualFn.ReturnType, genericNames, bindings)
	default:
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
}

func (g *generator) specializedTypeTemplateMatchesNormalized(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	if !g.typeExprHasGeneric(template, genericNames) && !g.typeExprHasGeneric(actual, genericNames) {
		if _, unionTemplate := template.(*ast.UnionTypeExpression); !unionTemplate {
			if _, unionActual := actual.(*ast.UnionTypeExpression); !unionActual {
				if normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual) {
					return true
				}
				if templateSimple, ok := template.(*ast.SimpleTypeExpression); ok && templateSimple != nil && templateSimple.Name != nil {
					if actualGeneric, ok := actual.(*ast.GenericTypeExpression); ok && actualGeneric != nil {
						if actualBase, ok := typeExprBaseName(actualGeneric.Base); ok && actualBase == templateSimple.Name.Name {
							return true
						}
					}
				}
				return false
			}
		}
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	key := specializedTypeTemplateMatchKey(template, actual)
	if _, ok := seen[key]; ok {
		return true
	}
	seen[key] = struct{}{}
	switch tt := template.(type) {
	case *ast.SimpleTypeExpression:
		if tt == nil || tt.Name == nil || tt.Name.Name == "" {
			return false
		}
		if _, ok := genericNames[tt.Name.Name]; ok {
			if bound, exists := bindings[tt.Name.Name]; exists {
				if normalizeTypeExprString(g, pkgName, bound) == tt.Name.Name && normalizeTypeExprString(g, pkgName, actual) != tt.Name.Name {
					bindings[tt.Name.Name] = actual
					return true
				}
				boundKey := normalizeTypeExprIdentityKey(g, pkgName, bound)
				actualKey := normalizeTypeExprIdentityKey(g, pkgName, actual)
				if boundKey != "" && actualKey != "" && boundKey != actualKey &&
					normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual) &&
					g.typeExprFullyBound(pkgName, actual) {
					bindings[tt.Name.Name] = actual
					return true
				}
				return normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual)
			}
			bindings[tt.Name.Name] = actual
			return true
		}
		if actualGeneric, ok := actual.(*ast.GenericTypeExpression); ok && actualGeneric != nil {
			if actualBase, ok := typeExprBaseName(actualGeneric.Base); ok && actualBase == tt.Name.Name {
				return true
			}
		}
		actualSimple, ok := actual.(*ast.SimpleTypeExpression)
		return ok && actualSimple != nil && actualSimple.Name != nil && actualSimple.Name.Name == tt.Name.Name
	case *ast.GenericTypeExpression:
		actualGeneric, ok := actual.(*ast.GenericTypeExpression)
		if !ok || actualGeneric == nil || len(tt.Arguments) != len(actualGeneric.Arguments) {
			return false
		}
		if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Base, actualGeneric.Base, genericNames, bindings, seen) {
			return false
		}
		for idx := range tt.Arguments {
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings, seen) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		actualNullable, ok := actual.(*ast.NullableTypeExpression)
		return ok && actualNullable != nil && g.specializedTypeTemplateMatchesNormalized(pkgName, tt.InnerType, actualNullable.InnerType, genericNames, bindings, seen)
	case *ast.ResultTypeExpression:
		actualResult, ok := actual.(*ast.ResultTypeExpression)
		return ok && actualResult != nil && g.specializedTypeTemplateMatchesNormalized(pkgName, tt.InnerType, actualResult.InnerType, genericNames, bindings, seen)
	case *ast.UnionTypeExpression:
		if actualUnion, ok := actual.(*ast.UnionTypeExpression); ok && actualUnion != nil {
			if len(tt.Members) != len(actualUnion.Members) {
				return false
			}
			for idx := range tt.Members {
				if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Members[idx], actualUnion.Members[idx], genericNames, bindings, seen) {
					return false
				}
			}
			return true
		}
		for _, member := range tt.Members {
			candidate := cloneTypeBindings(bindings)
			if candidate == nil {
				candidate = make(map[string]ast.TypeExpression)
			}
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, member, actual, genericNames, candidate, seen) {
				continue
			}
			applyTypeBindings(bindings, candidate)
			return true
		}
		return false
	case *ast.FunctionTypeExpression:
		actualFn, ok := actual.(*ast.FunctionTypeExpression)
		if !ok || actualFn == nil || len(tt.ParamTypes) != len(actualFn.ParamTypes) {
			return false
		}
		for idx := range tt.ParamTypes {
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.ParamTypes[idx], actualFn.ParamTypes[idx], genericNames, bindings, seen) {
				return false
			}
		}
		return g.specializedTypeTemplateMatchesNormalized(pkgName, tt.ReturnType, actualFn.ReturnType, genericNames, bindings, seen)
	default:
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
}

func (g *generator) normalizeTypeExprForSpecialization(pkgName string, expr ast.TypeExpression, seen map[string]struct{}) ast.TypeExpression {
	if g == nil || expr == nil {
		return expr
	}
	pkgName = g.resolvedTypeExprPackage(pkgName, expr)
	if seen == nil {
		seen = make(map[string]struct{})
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			return normalizeTypeExprForPackage(g, pkgName, expr)
		}
		key := pkgName + "|" + t.Name.Name
		if _, ok := seen[key]; ok {
			return expr
		}
		nextSeen := make(map[string]struct{}, len(seen)+1)
		for existing := range seen {
			nextSeen[existing] = struct{}{}
		}
		nextSeen[key] = struct{}{}
		if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil && expanded != expr {
			return g.normalizeTypeExprForSpecialization(pkgName, expanded, nextSeen)
		}
		return expr
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		if baseName, ok := typeExprBaseName(t.Base); ok && baseName != "" {
			key := pkgName + "|" + baseName + "<" + normalizeTypeExprListKey(g, pkgName, t.Arguments) + ">"
			if _, ok := seen[key]; ok {
				return expr
			}
			nextSeen := make(map[string]struct{}, len(seen)+1)
			for existing := range seen {
				nextSeen[existing] = struct{}{}
			}
			nextSeen[key] = struct{}{}
			if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil && expanded != expr {
				return g.normalizeTypeExprForSpecialization(pkgName, expanded, nextSeen)
			}
		}
		basePkg := g.resolvedTypeExprPackage(pkgName, t.Base)
		base := g.normalizeTypeExprForSpecialization(basePkg, t.Base, seen)
		changed := base != t.Base
		args := make([]ast.TypeExpression, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			argPkg := g.resolvedTypeExprPackage(pkgName, arg)
			next := g.normalizeTypeExprForSpecialization(argPkg, arg, seen)
			args = append(args, next)
			if next != arg {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return g.recordResolvedTypeExprPackage(ast.NewGenericTypeExpression(base, args), pkgName)
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		innerPkg := g.resolvedTypeExprPackage(pkgName, t.InnerType)
		inner := g.normalizeTypeExprForSpecialization(innerPkg, t.InnerType, seen)
		if inner == t.InnerType {
			return expr
		}
		return g.recordResolvedTypeExprPackage(ast.NewNullableTypeExpression(inner), pkgName)
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		innerPkg := g.resolvedTypeExprPackage(pkgName, t.InnerType)
		inner := g.normalizeTypeExprForSpecialization(innerPkg, t.InnerType, seen)
		if inner == t.InnerType {
			return expr
		}
		return g.recordResolvedTypeExprPackage(ast.NewResultTypeExpression(inner), pkgName)
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		members := make([]ast.TypeExpression, 0, len(t.Members))
		for _, member := range t.Members {
			memberPkg := g.resolvedTypeExprPackage(pkgName, member)
			next := g.normalizeTypeExprForSpecialization(memberPkg, member, seen)
			members = append(members, next)
			if next != member {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return g.recordResolvedTypeExprPackage(ast.NewUnionTypeExpression(members), pkgName)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		params := make([]ast.TypeExpression, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			paramPkg := g.resolvedTypeExprPackage(pkgName, param)
			next := g.normalizeTypeExprForSpecialization(paramPkg, param, seen)
			params = append(params, next)
			if next != param {
				changed = true
			}
		}
		retPkg := g.resolvedTypeExprPackage(pkgName, t.ReturnType)
		ret := g.normalizeTypeExprForSpecialization(retPkg, t.ReturnType, seen)
		if ret != t.ReturnType {
			changed = true
		}
		if !changed {
			return g.recordResolvedTypeExprPackage(normalizeCallableSyntaxTypeExpr(expr), pkgName)
		}
		return g.recordResolvedTypeExprPackage(normalizeCallableSyntaxTypeExpr(ast.NewFunctionTypeExpression(params, ret)), pkgName)
	default:
		return normalizeTypeExprForPackage(g, pkgName, expr)
	}
}

func specializedTypeTemplateMatchKey(template ast.TypeExpression, actual ast.TypeExpression) string {
	return fmt.Sprintf("%T:%x|%T:%x", template, typeExprPointer(template), actual, typeExprPointer(actual))
}

func typeExprPointer(expr ast.TypeExpression) uintptr {
	if expr == nil {
		return 0
	}
	value := reflect.ValueOf(expr)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return 0
	}
	return value.Pointer()
}

func (g *generator) specializedImplFunctionKey(info *functionInfo, bindings map[string]ast.TypeExpression) string {
	if info == nil {
		return ""
	}
	if g != nil {
		bindings = g.canonicalImplSpecializationBindings(info, nil, bindings)
	}
	base := strings.TrimSpace(info.Name)
	if info.QualifiedName != "" {
		base = strings.TrimSpace(info.QualifiedName)
	}
	if base == "" {
		base = strings.TrimSpace(info.GoName)
	}
	if pkg := strings.TrimSpace(info.Package); pkg != "" {
		base = pkg + "::" + base
	}
	if g != nil {
		if impl := g.implMethodInfoForFunction(info); impl != nil {
			if impl.ImplName != "" {
				base += "|impl=" + impl.ImplName
			}
			if constraintKey := constraintSignature(collectConstraintSpecs(impl.ImplGenerics, impl.WhereClause)); constraintKey != "" && constraintKey != "<none>" {
				base += "|constraints=" + constraintKey
			}
		}
	}
	if len(bindings) == 0 {
		return base
	}
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names)+1)
	parts = append(parts, base)
	for _, name := range names {
		bindingKey := normalizeTypeExprIdentityKey(g, info.Package, bindings[name])
		if bindingKey == "" {
			bindingKey = normalizeTypeExprString(g, info.Package, bindings[name])
		}
		parts = append(parts, name+"="+bindingKey)
	}
	return strings.Join(parts, "|")
}
