package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) importedSelectorAliasStartsTypeExpr(pkgName string, expr ast.TypeExpression) bool {
	if g == nil || expr == nil {
		return false
	}
	baseName, ok := typeExprBaseName(expr)
	if !ok || strings.TrimSpace(baseName) == "" {
		return false
	}
	sourcePkg, sourceName := g.importedSelectorSourceTypeAlias(pkgName, baseName)
	return strings.TrimSpace(sourcePkg) != "" && strings.TrimSpace(sourceName) != ""
}

func (g *generator) importedSelectorAliasAppearsInTypeExpr(pkgName string, expr ast.TypeExpression) bool {
	if g == nil || expr == nil {
		return false
	}
	if g.importedSelectorAliasStartsTypeExpr(pkgName, expr) {
		return true
	}
	switch t := expr.(type) {
	case *ast.GenericTypeExpression:
		if t == nil {
			return false
		}
		if g.importedSelectorAliasAppearsInTypeExpr(pkgName, t.Base) {
			return true
		}
		for _, arg := range t.Arguments {
			if g.importedSelectorAliasAppearsInTypeExpr(pkgName, arg) {
				return true
			}
		}
	case *ast.NullableTypeExpression:
		return t != nil && g.importedSelectorAliasAppearsInTypeExpr(pkgName, t.InnerType)
	case *ast.ResultTypeExpression:
		return t != nil && g.importedSelectorAliasAppearsInTypeExpr(pkgName, t.InnerType)
	case *ast.UnionTypeExpression:
		if t == nil {
			return false
		}
		for _, member := range t.Members {
			if g.importedSelectorAliasAppearsInTypeExpr(pkgName, member) {
				return true
			}
		}
	case *ast.FunctionTypeExpression:
		if t == nil {
			return false
		}
		for _, param := range t.ParamTypes {
			if g.importedSelectorAliasAppearsInTypeExpr(pkgName, param) {
				return true
			}
		}
		return g.importedSelectorAliasAppearsInTypeExpr(pkgName, t.ReturnType)
	}
	return false
}

func (g *generator) recordResolvedTypeExprPackage(expr ast.TypeExpression, pkgName string) ast.TypeExpression {
	if g == nil || expr == nil {
		return expr
	}
	pkgName = strings.TrimSpace(pkgName)
	if pkgName == "" || g.normalizedTypeExprPackagesByExpr == nil {
		return expr
	}
	g.normalizedTypeExprPackagesByExpr[expr] = pkgName
	return expr
}

func (g *generator) invalidateNormalizedTypeExprCaches() {
	if g == nil {
		return
	}
	g.normalizedTypeExprCache = make(map[string]ast.TypeExpression)
	g.normalizedTypeExprPackageCache = make(map[string]string)
}

func (g *generator) normalizeTypeExprContextForPackage(pkgName string, expr ast.TypeExpression) (string, ast.TypeExpression) {
	resolvedPkg := strings.TrimSpace(pkgName)
	if g == nil || expr == nil {
		return resolvedPkg, expr
	}
	if g.normalizedTypeExprPackagesByExpr != nil && !g.importedSelectorAliasAppearsInTypeExpr(resolvedPkg, expr) {
		if recorded := strings.TrimSpace(g.normalizedTypeExprPackagesByExpr[expr]); recorded != "" {
			resolvedPkg = recorded
		}
	}
	cacheKey := normalizeTypeExprCacheKey(g, resolvedPkg, expr)
	if cacheKey != "" {
		if cached, ok := g.normalizedTypeExprCache[cacheKey]; ok && cached != nil {
			if g.normalizedTypeExprPackageCache != nil {
				if recorded := strings.TrimSpace(g.normalizedTypeExprPackageCache[cacheKey]); recorded != "" {
					resolvedPkg = recorded
				}
			}
			if g.normalizedTypeExprPackagesByExpr != nil {
				if recorded := strings.TrimSpace(g.normalizedTypeExprPackagesByExpr[cached]); recorded != "" {
					resolvedPkg = recorded
				} else {
					g.normalizedTypeExprPackagesByExpr[cached] = resolvedPkg
				}
			}
			return resolvedPkg, cached
		}
	}
	resolvedPkg, normalized := g.expandTypeAliasContextForPackage(resolvedPkg, expr)
	normalized = normalizeBuiltinSemanticTypeExprInPackage(g, resolvedPkg, normalized)
	normalized = normalizeKernelBuiltinTypeExpr(normalized)
	normalized = normalizeCallableSyntaxTypeExpr(normalized)
	normalized = normalizeNestedTypeExprChildrenForPackage(g, resolvedPkg, normalized)
	if normalized != nil && g.normalizedTypeExprPackagesByExpr != nil {
		g.normalizedTypeExprPackagesByExpr[normalized] = resolvedPkg
	}
	if cacheKey != "" && normalized != nil {
		g.normalizedTypeExprCache[cacheKey] = normalized
		if g.normalizedTypeExprPackageCache != nil {
			g.normalizedTypeExprPackageCache[cacheKey] = resolvedPkg
		}
	}
	return resolvedPkg, normalized
}

func normalizeNestedTypeExprChildrenForPackage(g *generator, pkgName string, expr ast.TypeExpression) ast.TypeExpression {
	if g == nil || expr == nil {
		return expr
	}
	switch t := expr.(type) {
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		base := normalizeTypeExprForPackage(g, pkgName, t.Base)
		changed := base != t.Base
		args := make([]ast.TypeExpression, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			next := normalizeTypeExprForPackage(g, pkgName, arg)
			args = append(args, next)
			if next != arg {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		inner := normalizeTypeExprForPackage(g, pkgName, t.InnerType)
		if inner == t.InnerType {
			return expr
		}
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		inner := normalizeTypeExprForPackage(g, pkgName, t.InnerType)
		if inner == t.InnerType {
			return expr
		}
		return ast.NewResultTypeExpression(inner)
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		members := make([]ast.TypeExpression, 0, len(t.Members))
		for _, member := range t.Members {
			next := normalizeTypeExprForPackage(g, pkgName, member)
			members = append(members, next)
			if next != member {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		ret := normalizeTypeExprForPackage(g, pkgName, t.ReturnType)
		changed := ret != t.ReturnType
		params := make([]ast.TypeExpression, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			next := normalizeTypeExprForPackage(g, pkgName, param)
			params = append(params, next)
			if next != param {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewFunctionTypeExpression(params, ret)
	default:
		return expr
	}
}

func (g *generator) resolvedTypeExprPackage(pkgName string, expr ast.TypeExpression) string {
	if g == nil || expr == nil {
		return strings.TrimSpace(pkgName)
	}
	if g.normalizedTypeExprPackagesByExpr != nil {
		if recorded := strings.TrimSpace(g.normalizedTypeExprPackagesByExpr[expr]); recorded != "" {
			return recorded
		}
	}
	resolvedPkg, normalized := g.normalizeTypeExprContextForPackage(pkgName, expr)
	if normalized != nil && g.normalizedTypeExprPackagesByExpr != nil {
		if recorded := strings.TrimSpace(g.normalizedTypeExprPackagesByExpr[normalized]); recorded != "" {
			return recorded
		}
	}
	return resolvedPkg
}

func (g *generator) expandTypeAliasContextForPackage(pkgName string, expr ast.TypeExpression) (string, ast.TypeExpression) {
	currentPkg := strings.TrimSpace(pkgName)
	current := expr
	if g == nil || expr == nil {
		return currentPkg, current
	}
	seen := make(map[string]struct{})
	for range 32 {
		next, nextPkg, key, changed := g.expandTypeAliasOnceRawForPackage(currentPkg, current)
		if !changed || next == nil {
			return currentPkg, current
		}
		if strings.TrimSpace(nextPkg) == "" {
			nextPkg = currentPkg
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return currentPkg, current
		}
		seenKey := strings.TrimSpace(nextPkg) + "|" + key
		if _, exists := seen[seenKey]; exists {
			return currentPkg, current
		}
		seen[seenKey] = struct{}{}
		current = next
		currentPkg = strings.TrimSpace(nextPkg)
	}
	return currentPkg, current
}

func (g *generator) expandTypeAliasOnceRawForPackage(pkgName string, expr ast.TypeExpression) (ast.TypeExpression, string, string, bool) {
	if g == nil || expr == nil {
		return expr, strings.TrimSpace(pkgName), "", false
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || strings.TrimSpace(t.Name.Name) == "" {
			return expr, strings.TrimSpace(pkgName), "", false
		}
		aliasPkg, aliasName, target, params, ok := g.typeAliasTargetForPackage(pkgName, t.Name.Name)
		if !ok || target == nil {
			return expr, strings.TrimSpace(pkgName), "", false
		}
		if len(params) != 0 {
			return expr, strings.TrimSpace(pkgName), "", false
		}
		return target, aliasPkg, aliasName, true
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr, strings.TrimSpace(pkgName), "", false
		}
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base == nil || base.Name == nil || strings.TrimSpace(base.Name.Name) == "" {
			return expr, strings.TrimSpace(pkgName), "", false
		}
		aliasPkg, aliasName, target, params, ok := g.typeAliasTargetForPackage(pkgName, base.Name.Name)
		if !ok || target == nil {
			return expr, strings.TrimSpace(pkgName), "", false
		}
		if len(params) == 0 {
			args := make([]ast.TypeExpression, 0, len(t.Arguments))
			for _, arg := range t.Arguments {
				if arg == nil {
					return expr, strings.TrimSpace(pkgName), "", false
				}
				args = append(args, normalizeTypeExprForPackage(g, pkgName, arg))
			}
			expanded := ast.NewGenericTypeExpression(cloneTypeExpr(target), args)
			return normalizeTypeExprForPackage(g, aliasPkg, expanded), aliasPkg, aliasName + "<" + normalizeTypeExprListKey(g, pkgName, t.Arguments) + ">", true
		}
		if len(params) != len(t.Arguments) {
			return expr, strings.TrimSpace(pkgName), "", false
		}
		bindings := make(map[string]ast.TypeExpression, len(params))
		for idx, gp := range params {
			if gp == nil || gp.Name == nil || gp.Name.Name == "" || t.Arguments[idx] == nil {
				return expr, strings.TrimSpace(pkgName), "", false
			}
			bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, pkgName, t.Arguments[idx])
		}
		expanded := substituteTypeParams(target, bindings)
		return normalizeTypeExprForPackage(g, aliasPkg, expanded), aliasPkg, aliasName + "<" + normalizeTypeExprListKey(g, pkgName, t.Arguments) + ">", true
	default:
		return expr, strings.TrimSpace(pkgName), "", false
	}
}
