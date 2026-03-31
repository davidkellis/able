package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) specializedImplSignatureUsesUnresolvedNominalStructs(method *methodInfo, impl *implMethodInfo, target ast.TypeExpression, bindings map[string]ast.TypeExpression) bool {
	if g == nil || method == nil || method.Info == nil || method.Info.Definition == nil || impl == nil {
		return false
	}
	pkgName := method.Info.Package
	genericNames := g.implSpecializationGenericNames(method)
	if len(genericNames) == 0 {
		return false
	}
	interfaceBindings := g.implTypeBindings(pkgName, impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, target)
	selfTarget := g.implSelfTargetType(pkgName, target, interfaceBindings)
	allBindings := g.mergeImplSelfTargetBindings(pkgName, target, selfTarget, interfaceBindings)
	if allBindings == nil {
		allBindings = make(map[string]ast.TypeExpression)
	}
	for name, expr := range bindings {
		if expr == nil {
			continue
		}
		allBindings[name] = normalizeTypeExprForPackage(g, pkgName, expr)
	}
	if selfTarget != nil {
		allBindings["Self"] = normalizeTypeExprForPackage(g, pkgName, selfTarget)
	}
	for _, param := range method.Info.Definition.Params {
		if param == nil {
			continue
		}
		if g.typeExprNeedsConcreteNominalStructCarrier(pkgName, substituteTypeParams(resolveSelfTypeExpr(param.ParamType, selfTarget), allBindings), genericNames) {
			return true
		}
	}
	if g.typeExprNeedsConcreteNominalStructCarrier(pkgName, substituteTypeParams(resolveSelfTypeExpr(method.Info.Definition.ReturnType, selfTarget), allBindings), genericNames) {
		return true
	}
	return false
}

func (g *generator) typeExprNeedsConcreteNominalStructCarrier(pkgName string, expr ast.TypeExpression, genericNames map[string]struct{}) bool {
	if g == nil || expr == nil || len(genericNames) == 0 {
		return false
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	if expr == nil || !g.typeExprHasGeneric(expr, genericNames) {
		return false
	}
	generic, ok := expr.(*ast.GenericTypeExpression)
	if !ok || generic == nil {
		return false
	}
	baseName, ok := typeExprBaseName(generic)
	if !ok || baseName == "" {
		return false
	}
	_, ok = g.structInfoForTypeName(pkgName, baseName)
	return ok
}
