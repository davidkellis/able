package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

func packageIndependentTypeIdentityName(name string) bool {
	switch name {
	case "", "bool", "Bool", "String", "string", "char", "Char",
		"i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64",
		"isize", "usize", "f32", "f64", "void", "Void",
		"Array",
		"Error", "Value", "nil":
		return true
	}
	return false
}

func normalizeTypeExprCacheKey(g *generator, pkgName string, expr ast.TypeExpression) string {
	if expr == nil {
		return ""
	}
	resolvedPkg := strings.TrimSpace(pkgName)
	if g != nil && g.normalizedTypeExprPackagesByExpr != nil && !g.importedSelectorAliasAppearsInTypeExpr(resolvedPkg, expr) {
		if recorded := strings.TrimSpace(g.normalizedTypeExprPackagesByExpr[expr]); recorded != "" {
			resolvedPkg = recorded
		}
	}
	prefix := resolvedPkg + "::"
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return ""
		}
		if packageIndependentTypeIdentityName(t.Name.Name) {
			prefix = ""
		}
		return prefix + "simple(" + t.Name.Name + ")"
	case *ast.GenericTypeExpression:
		if t == nil {
			return ""
		}
		parts := []string{normalizeTypeExprCacheKey(g, resolvedPkg, t.Base)}
		for _, arg := range t.Arguments {
			parts = append(parts, normalizeTypeExprCacheKey(g, resolvedPkg, arg))
		}
		return prefix + "generic(" + strings.Join(parts, ",") + ")"
	case *ast.NullableTypeExpression:
		if t == nil {
			return ""
		}
		return prefix + "nullable(" + normalizeTypeExprCacheKey(g, resolvedPkg, t.InnerType) + ")"
	case *ast.ResultTypeExpression:
		if t == nil {
			return ""
		}
		return prefix + "result(" + normalizeTypeExprCacheKey(g, resolvedPkg, t.InnerType) + ")"
	case *ast.UnionTypeExpression:
		if t == nil {
			return ""
		}
		parts := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			parts = append(parts, normalizeTypeExprCacheKey(g, resolvedPkg, member))
		}
		return prefix + "union(" + strings.Join(parts, ",") + ")"
	case *ast.FunctionTypeExpression:
		if t == nil {
			return ""
		}
		parts := make([]string, 0, len(t.ParamTypes)+1)
		for _, param := range t.ParamTypes {
			parts = append(parts, normalizeTypeExprCacheKey(g, resolvedPkg, param))
		}
		parts = append(parts, normalizeTypeExprCacheKey(g, resolvedPkg, t.ReturnType))
		return prefix + "fn(" + strings.Join(parts, ",") + ")"
	case *ast.WildcardTypeExpression:
		return prefix + "wildcard"
	default:
		return prefix + typeExpressionToString(expr)
	}
}

func normalizeTypeExprIdentityKey(g *generator, pkgName string, expr ast.TypeExpression) string {
	if expr == nil {
		return ""
	}
	resolvedPkg := strings.TrimSpace(pkgName)
	normalized := expr
	if g != nil {
		resolvedPkg, normalized = g.normalizeTypeExprContextForPackage(pkgName, expr)
	}
	if normalized == nil {
		return ""
	}
	prefix := resolvedPkg + "::"
	switch t := normalized.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil {
			return ""
		}
		if packageIndependentTypeIdentityName(t.Name.Name) {
			prefix = ""
		}
		return prefix + "simple(" + t.Name.Name + ")"
	case *ast.GenericTypeExpression:
		if t == nil {
			return ""
		}
		parts := []string{normalizeTypeExprIdentityKey(g, resolvedPkg, t.Base)}
		for _, arg := range t.Arguments {
			parts = append(parts, normalizeTypeExprIdentityKey(g, resolvedPkg, arg))
		}
		return prefix + "generic(" + strings.Join(parts, ",") + ")"
	case *ast.NullableTypeExpression:
		if t == nil {
			return ""
		}
		return prefix + "nullable(" + normalizeTypeExprIdentityKey(g, resolvedPkg, t.InnerType) + ")"
	case *ast.ResultTypeExpression:
		if t == nil {
			return ""
		}
		return prefix + "result(" + normalizeTypeExprIdentityKey(g, resolvedPkg, t.InnerType) + ")"
	case *ast.UnionTypeExpression:
		if t == nil {
			return ""
		}
		parts := make([]string, 0, len(t.Members))
		for _, member := range t.Members {
			parts = append(parts, normalizeTypeExprIdentityKey(g, resolvedPkg, member))
		}
		return prefix + "union(" + strings.Join(parts, ",") + ")"
	case *ast.FunctionTypeExpression:
		if t == nil {
			return ""
		}
		parts := make([]string, 0, len(t.ParamTypes)+1)
		for _, param := range t.ParamTypes {
			parts = append(parts, normalizeTypeExprIdentityKey(g, resolvedPkg, param))
		}
		parts = append(parts, normalizeTypeExprIdentityKey(g, resolvedPkg, t.ReturnType))
		return prefix + "fn(" + strings.Join(parts, ",") + ")"
	case *ast.WildcardTypeExpression:
		return prefix + "wildcard"
	default:
		return prefix + typeExpressionToString(normalized)
	}
}
