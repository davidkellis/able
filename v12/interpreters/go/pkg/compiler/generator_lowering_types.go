package compiler

import "able/interpreter-go/pkg/ast"

// lowerNormalizedTypeExpr is the canonical type-normalization entrypoint for
// compiler codegen. Emitters should not normalize or substitute type
// expressions ad hoc.
func (g *generator) lowerNormalizedTypeExpr(ctx *compileContext, expr ast.TypeExpression) ast.TypeExpression {
	return g.typeExprInContext(ctx, expr)
}

// lowerCarrierType is the canonical carrier-synthesis entrypoint for codegen
// sites that need a Go carrier for a type in the current compile context.
func (g *generator) lowerCarrierType(ctx *compileContext, expr ast.TypeExpression) (string, bool) {
	return g.mapTypeExpressionInContext(ctx, expr)
}

// lowerCarrierTypeInPackage is the canonical carrier-synthesis entrypoint for
// package-scoped synthesis that is not tied to a compile context.
func (g *generator) lowerCarrierTypeInPackage(pkgName string, expr ast.TypeExpression) (string, bool) {
	return g.mapTypeExpressionInPackage(pkgName, expr)
}
