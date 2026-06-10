package compiler

import "able/interpreter-go/pkg/ast"

type typeExprPackageCacheKey struct {
	Expr        ast.TypeExpression
	PackageName string
}

type typeExprPackageCacheEntry struct {
	ImportedAliasKnown   bool
	ImportedAliasAppears bool
	NormalizedPackage    string
	NormalizedExpr       ast.TypeExpression
}

func (g *generator) recordNormalizedTypeExprSource(source ast.TypeExpression, pkgName string, normalized ast.TypeExpression) {
	if g == nil || source == nil || normalized == nil {
		return
	}
	key := typeExprPackageCacheKey{
		Expr:        source,
		PackageName: pkgName,
	}
	entry := g.typeExprPackageCache[key]
	entry.NormalizedPackage = pkgName
	entry.NormalizedExpr = normalized
	g.typeExprPackageCache[key] = entry
}
