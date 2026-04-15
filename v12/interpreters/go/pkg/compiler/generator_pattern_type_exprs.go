package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) patternBindingTypeExpr(ctx *compileContext, goType string, explicit ast.TypeExpression) ast.TypeExpression {
	if g == nil {
		return explicit
	}
	typeExpr := g.lowerNormalizedTypeExpr(ctx, explicit)
	if typeExpr != nil {
		return typeExpr
	}
	typeExpr, _ = g.typeExprForGoType(goType)
	return g.lowerNormalizedTypeExpr(ctx, typeExpr)
}
