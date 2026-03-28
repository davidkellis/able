package compiler

import "able/interpreter-go/pkg/ast"

func nativeInterfaceInstantiationExpr(name string, args []ast.TypeExpression) ast.TypeExpression {
	if name == "" {
		return nil
	}
	base := ast.Ty(name)
	if len(args) == 0 {
		return base
	}
	return ast.Gen(base, args...)
}

func (g *generator) nativeInterfaceImplBindingsForTarget(
	actualPkg string,
	actualExpr ast.TypeExpression,
	genericNames map[string]struct{},
	expectedPkg string,
	expectedName string,
	expectedArgs []ast.TypeExpression,
	seen map[string]struct{},
) (map[string]ast.TypeExpression, bool) {
	if g == nil || actualExpr == nil || expectedName == "" {
		return nil, false
	}
	cacheKey := g.nativeInterfaceImplBindingsCacheKey(actualPkg, actualExpr, genericNames, expectedPkg, expectedName, expectedArgs)
	if cacheKey != "" {
		if cached, ok := g.nativeInterfaceImplBindingCache[cacheKey]; ok {
			return cloneTypeBindings(cached.Bindings), cached.Matched
		}
	}
	store := func(bindings map[string]ast.TypeExpression, matched bool) (map[string]ast.TypeExpression, bool) {
		if cacheKey != "" {
			g.nativeInterfaceImplBindingCache[cacheKey] = nativeInterfaceImplBindingCacheEntry{
				Bindings: cloneTypeBindings(bindings),
				Matched:  matched,
			}
		}
		return bindings, matched
	}
	ifacePkg, ifaceName, ifaceArgs, ifaceDef, ok := interfaceExprInfo(g, actualPkg, actualExpr)
	if !ok {
		return store(nil, false)
	}
	key := ifacePkg + "::" + ifaceName + "<" + normalizeTypeExprListKey(g, ifacePkg, ifaceArgs) + ">"
	if _, exists := seen[key]; exists {
		return store(nil, false)
	}
	seen[key] = struct{}{}
	if ifaceName == expectedName && (expectedPkg == "" || ifacePkg == expectedPkg) && len(ifaceArgs) == len(expectedArgs) {
		bindings := make(map[string]ast.TypeExpression, len(expectedArgs))
		if g.nativeInterfaceTypeBindingListMatches(ifacePkg, ifaceArgs, expectedArgs, genericNames, bindings) ||
			g.nativeInterfaceTypeBindingListMatches(ifacePkg, expectedArgs, ifaceArgs, genericNames, bindings) {
			return store(bindings, true)
		}
		leftVars := make(map[string]string)
		rightVars := make(map[string]string)
		if g.typeExprEquivalentModuloGenerics(
			nativeInterfaceInstantiationExpr(ifaceName, ifaceArgs),
			nativeInterfaceInstantiationExpr(expectedName, expectedArgs),
			leftVars,
			rightVars,
		) {
			return store(bindings, true)
		}
		return store(nil, false)
	}
	if ifaceDef == nil {
		return store(nil, false)
	}
	bindings := nativeInterfaceBindings(ifaceDef, ifaceArgs)
	for _, baseExpr := range ifaceDef.BaseInterfaces {
		if baseExpr == nil {
			continue
		}
		next := substituteTypeParams(baseExpr, bindings)
		next = normalizeTypeExprForPackage(g, ifacePkg, next)
		if matched, ok := g.nativeInterfaceImplBindingsForTarget(ifacePkg, next, genericNames, expectedPkg, expectedName, expectedArgs, seen); ok {
			return store(matched, true)
		}
	}
	return store(nil, false)
}

func (g *generator) nativeInterfaceTypeBindingListMatches(
	pkgName string,
	templates []ast.TypeExpression,
	actuals []ast.TypeExpression,
	genericNames map[string]struct{},
	bindings map[string]ast.TypeExpression,
) bool {
	if g == nil || len(templates) != len(actuals) {
		return false
	}
	candidate := cloneTypeBindings(bindings)
	if candidate == nil {
		candidate = make(map[string]ast.TypeExpression, len(actuals))
	}
	for idx, template := range templates {
		if !g.nativeInterfaceTypeTemplateMatches(pkgName, template, actuals[idx], genericNames, candidate) {
			return false
		}
	}
	applyTypeBindings(bindings, candidate)
	return true
}
