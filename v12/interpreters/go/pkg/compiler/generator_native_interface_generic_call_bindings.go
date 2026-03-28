package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) nativeInterfaceConcreteGenericCallBindings(methodInfo *methodInfo, impl *implMethodInfo, receiverTypeExpr ast.TypeExpression, paramTypeExprs []ast.TypeExpression, returnTypeExpr ast.TypeExpression, callTypeArgs []ast.TypeExpression, initial map[string]ast.TypeExpression) (map[string]ast.TypeExpression, bool) {
	if g == nil || methodInfo == nil || methodInfo.Info == nil || impl == nil {
		return nil, false
	}
	genericNames := g.implSpecializationGenericNames(methodInfo)
	bindings := cloneTypeBindings(initial)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if receiverTypeExpr != nil && methodInfo.ExpectsSelf && len(methodInfo.Info.Params) > 0 && methodInfo.Info.Params[0].TypeExpr != nil {
		receiverTemplate := methodInfo.Info.Params[0].TypeExpr
		if targetTemplate := g.specializedImplTargetTemplate(impl, bindings); targetTemplate != nil {
			if g.preferImplSpecializationTemplate(methodInfo.Info.Package, receiverTemplate, targetTemplate) {
				receiverTemplate = targetTemplate
			}
		}
		if !g.specializedTypeTemplateMatches(methodInfo.Info.Package, receiverTemplate, receiverTypeExpr, genericNames, bindings, make(map[string]struct{})) {
			return nil, false
		}
	}
	if len(callTypeArgs) > 0 {
		if methodInfo.Info.Definition == nil || len(methodInfo.Info.Definition.GenericParams) != len(callTypeArgs) {
			return nil, false
		}
		for idx, arg := range callTypeArgs {
			if arg == nil {
				return nil, false
			}
			gp := methodInfo.Info.Definition.GenericParams[idx]
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				return nil, false
			}
			bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, methodInfo.Info.Package, arg)
		}
	}
	implParamTypeExprs, _, implReturnTypeExpr, _, optionalLast, ok := g.nativeInterfaceMethodImplSignature(impl, bindings)
	if !ok || optionalLast || len(implParamTypeExprs) != len(paramTypeExprs) {
		return nil, false
	}
	for idx, actual := range paramTypeExprs {
		if actual == nil {
			return nil, false
		}
		actual = normalizeTypeExprForPackage(g, methodInfo.Info.Package, actual)
		if !g.bindSpecializedCallArgument(methodInfo.Info.Package, implParamTypeExprs[idx], actual, genericNames, bindings) &&
			!g.specializedConcreteTemplateSatisfiesActual(methodInfo.Info.Package, implParamTypeExprs[idx], actual, bindings) {
			return nil, false
		}
	}
	if returnTypeExpr != nil {
		returnTypeExpr = normalizeTypeExprForPackage(g, methodInfo.Info.Package, returnTypeExpr)
		if !g.bindSpecializedCallArgument(methodInfo.Info.Package, implReturnTypeExpr, returnTypeExpr, genericNames, bindings) &&
			!g.specializedConcreteTemplateSatisfiesActual(methodInfo.Info.Package, implReturnTypeExpr, returnTypeExpr, bindings) {
			return nil, false
		}
	}
	bindings = g.normalizeConcreteTypeBindings(methodInfo.Info.Package, bindings, nil)
	if len(bindings) == 0 {
		return nil, false
	}
	return bindings, true
}

func (g *generator) specializedConcreteTemplateSatisfiesActual(pkgName string, templateExpr ast.TypeExpression, actualExpr ast.TypeExpression, bindings map[string]ast.TypeExpression) bool {
	if g == nil || templateExpr == nil || actualExpr == nil {
		return false
	}
	concreteExpr := normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(templateExpr, bindings))
	if concreteExpr == nil || !g.typeExprFullyBound(pkgName, concreteExpr) {
		return false
	}
	mapper := NewTypeMapper(g, pkgName)
	concreteGoType, ok := mapper.Map(concreteExpr)
	concreteGoType, ok = g.recoverRepresentableCarrierType(pkgName, concreteExpr, concreteGoType)
	if !ok || concreteGoType == "" || concreteGoType == "runtime.Value" || concreteGoType == "any" {
		return false
	}
	actualGoType, ok := mapper.Map(actualExpr)
	actualGoType, ok = g.recoverRepresentableCarrierType(pkgName, actualExpr, actualGoType)
	if !ok || actualGoType == "" || actualGoType == "runtime.Value" || actualGoType == "any" {
		return false
	}
	return g.canCoerceStaticExpr(actualGoType, concreteGoType)
}
