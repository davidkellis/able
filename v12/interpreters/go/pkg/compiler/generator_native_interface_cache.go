package compiler

import (
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type nativeInterfaceImplBindingCacheEntry struct {
	Bindings map[string]ast.TypeExpression
	Matched  bool
}

func nativeInterfaceGenericNamesKey(names map[string]struct{}) string {
	if len(names) == 0 {
		return ""
	}
	keys := make([]string, 0, len(names))
	for name := range names {
		if strings.TrimSpace(name) == "" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return strings.Join(keys, "|")
}

func normalizeTypeExprCacheListKey(g *generator, pkgName string, exprs []ast.TypeExpression) string {
	if len(exprs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		parts = append(parts, normalizeTypeExprString(g, pkgName, expr))
	}
	return strings.Join(parts, "|")
}

func (g *generator) nativeInterfaceImplBindingsCacheKey(
	actualPkg string,
	actualExpr ast.TypeExpression,
	genericNames map[string]struct{},
	expectedPkg string,
	expectedName string,
	expectedArgs []ast.TypeExpression,
) string {
	if g == nil || actualExpr == nil || expectedName == "" {
		return ""
	}
	return strings.Join([]string{
		strings.TrimSpace(actualPkg),
		normalizeTypeExprString(g, actualPkg, actualExpr),
		nativeInterfaceGenericNamesKey(genericNames),
		strings.TrimSpace(expectedPkg),
		expectedName,
		normalizeTypeExprCacheListKey(g, expectedPkg, expectedArgs),
	}, "::")
}
