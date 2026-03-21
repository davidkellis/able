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
	ifacePkg, ifaceName, ifaceArgs, ifaceDef, ok := interfaceExprInfo(g, actualPkg, actualExpr)
	if !ok {
		return nil, false
	}
	key := ifacePkg + "::" + ifaceName + "<" + normalizeTypeExprListKey(g, ifacePkg, ifaceArgs) + ">"
	if _, exists := seen[key]; exists {
		return nil, false
	}
	seen[key] = struct{}{}
	if ifaceName == expectedName && (expectedPkg == "" || ifacePkg == expectedPkg) && len(ifaceArgs) == len(expectedArgs) {
		bindings := make(map[string]ast.TypeExpression, len(expectedArgs))
		for idx, template := range ifaceArgs {
			if !g.nativeInterfaceTypeTemplateMatches(ifacePkg, template, expectedArgs[idx], genericNames, bindings) {
				return nil, false
			}
		}
		return bindings, true
	}
	if ifaceDef == nil {
		return nil, false
	}
	bindings := nativeInterfaceBindings(ifaceDef, ifaceArgs)
	for _, baseExpr := range ifaceDef.BaseInterfaces {
		if baseExpr == nil {
			continue
		}
		next := substituteTypeParams(baseExpr, bindings)
		next = normalizeTypeExprForPackage(g, ifacePkg, next)
		if matched, ok := g.nativeInterfaceImplBindingsForTarget(ifacePkg, next, genericNames, expectedPkg, expectedName, expectedArgs, seen); ok {
			return matched, true
		}
	}
	return nil, false
}
