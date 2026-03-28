package compiler

import (
	"sort"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) specializedCallActualTypeExpr(ctx *compileContext, pkgName string, arg ast.Expression, paramTypeExpr ast.TypeExpression, bindings map[string]ast.TypeExpression) (ast.TypeExpression, string, bool) {
	if g == nil || ctx == nil || arg == nil {
		return nil, "", false
	}
	expectedTypeExpr := normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(paramTypeExpr, bindings))
	expectedGoType := ""
	if expectedTypeExpr != nil && g.typeExprFullyBound(pkgName, expectedTypeExpr) {
		if mapped, ok := g.lowerCarrierTypeInPackage(pkgName, expectedTypeExpr); ok && mapped != "" {
			if recovered, ok := g.recoverRepresentableCarrierType(pkgName, expectedTypeExpr, mapped); ok && recovered != "" {
				expectedGoType = recovered
			} else {
				expectedGoType = mapped
			}
		}
	}
	argCtx := ctx.child()
	actualGoType := ""
	var actualGoTypeOK bool
	if expectedGoType != "" && expectedGoType != "runtime.Value" && expectedGoType != "any" {
		_, _, actualGoType, actualGoTypeOK = g.compileExprLinesWithExpectedTypeExpr(argCtx, arg, expectedGoType, expectedTypeExpr)
	} else {
		_, _, actualGoType, actualGoTypeOK = g.compileExprLines(argCtx, arg, "")
	}
	if expectedTypeExpr != nil && expectedGoType != "" && expectedGoType != "runtime.Value" && expectedGoType != "any" {
		if !actualGoTypeOK || actualGoType == "" || actualGoType == expectedGoType || g.canCoerceStaticExpr(expectedGoType, actualGoType) {
			return expectedTypeExpr, expectedGoType, true
		}
	}
	actualExpr, ok := g.inferExpressionTypeExpr(argCtx, arg, actualGoType)
	if ok && actualExpr != nil {
		actualExpr = g.preferConcreteTypeExprForGoType(argCtx, actualExpr, actualGoType)
	}
	if !ok || actualExpr == nil {
		return nil, "", false
	}
	return actualExpr, actualGoType, true
}

func (g *generator) bindSpecializedCallArgument(pkgName string, paramTypeExpr ast.TypeExpression, actualExpr ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || paramTypeExpr == nil || actualExpr == nil || bindings == nil {
		return false
	}
	if g.bindSpecializedCallTemplate(pkgName, paramTypeExpr, actualExpr, genericNames, bindings) {
		return true
	}
	return g.bindSpecializedConcreteInterfaceArgument(pkgName, paramTypeExpr, actualExpr, genericNames, bindings)
}

func (g *generator) bindSpecializedCallTemplate(pkgName string, paramTypeExpr ast.TypeExpression, actualExpr ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	candidate := cloneTypeBindings(bindings)
	if candidate == nil {
		candidate = make(map[string]ast.TypeExpression)
	}
	if !g.specializedTypeTemplateMatches(pkgName, paramTypeExpr, actualExpr, genericNames, candidate, make(map[string]struct{})) {
		return false
	}
	applyTypeBindings(bindings, candidate)
	return true
}

func applyTypeBindings(dst map[string]ast.TypeExpression, src map[string]ast.TypeExpression) {
	if len(dst) == 0 || len(src) == 0 {
		for name, expr := range src {
			dst[name] = expr
		}
		return
	}
	for name, expr := range src {
		dst[name] = expr
	}
}

func (g *generator) bindSpecializedConcreteInterfaceArgument(pkgName string, paramTypeExpr ast.TypeExpression, actualExpr ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || paramTypeExpr == nil || actualExpr == nil || bindings == nil {
		return false
	}
	ifacePkg, ifaceName, ifaceArgs, _, ok := interfaceExprInfo(g, pkgName, paramTypeExpr)
	if !ok {
		return false
	}
	actualGoType, ok := g.recoverRepresentableCarrierType(pkgName, actualExpr, "")
	if !ok || actualGoType == "" || actualGoType == "runtime.Value" || actualGoType == "any" {
		return false
	}
	var found map[string]ast.TypeExpression
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		if impl == nil || impl.Info == nil || !impl.Info.Compileable || impl.ImplName != "" {
			continue
		}
		candidateBindings := cloneTypeBindings(bindings)
		if candidateBindings == nil {
			candidateBindings = make(map[string]ast.TypeExpression)
		}
		info, implBindings, ok := g.nativeInterfaceConcreteImplInfo(actualGoType, impl, candidateBindings)
		if !ok || info == nil {
			continue
		}
		actualIfaceExpr := g.implConcreteInterfaceExpr(impl, implBindings)
		if actualIfaceExpr == nil {
			continue
		}
		matched, ok := g.nativeInterfaceImplBindingsForTarget(info.Package, actualIfaceExpr, genericNames, ifacePkg, ifaceName, ifaceArgs, make(map[string]struct{}))
		if !ok {
			continue
		}
		if !g.mergeConcreteBindings(pkgName, implBindings, matched) {
			continue
		}
		normalized := g.normalizeConcreteTypeBindings(pkgName, implBindings, genericNames)
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

func (g *generator) mergeConcreteBindings(pkgName string, dst map[string]ast.TypeExpression, extra map[string]ast.TypeExpression) bool {
	for name, expr := range extra {
		if expr == nil {
			continue
		}
		normalized := normalizeTypeExprForPackage(g, pkgName, expr)
		if existing, exists := dst[name]; exists {
			existing = normalizeTypeExprForPackage(g, pkgName, existing)
			if simple, ok := existing.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == name {
				dst[name] = normalized
				continue
			}
			existingConcrete := g.typeExprFullyBound(pkgName, existing)
			normalizedConcrete := g.typeExprFullyBound(pkgName, normalized)
			if !existingConcrete && normalizedConcrete {
				dst[name] = normalized
				continue
			}
			if existingConcrete && !normalizedConcrete {
				continue
			}
			if normalizeTypeExprString(g, pkgName, existing) != normalizeTypeExprString(g, pkgName, normalized) {
				return false
			}
			continue
		}
		dst[name] = normalized
	}
	return true
}

func concreteBindingsEquivalent(g *generator, pkgName string, left map[string]ast.TypeExpression, right map[string]ast.TypeExpression) bool {
	if len(left) != len(right) {
		return false
	}
	names := make([]string, 0, len(left))
	for name := range left {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if normalizeTypeExprString(g, pkgName, left[name]) != normalizeTypeExprString(g, pkgName, right[name]) {
			return false
		}
	}
	return true
}
