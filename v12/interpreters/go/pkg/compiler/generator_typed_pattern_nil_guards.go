package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) staticTypedPatternRequiresNonNilGuard(ctx *compileContext, subjectType string, patternType ast.TypeExpression) bool {
	if g == nil || patternType == nil || subjectType == "" || subjectType == "runtime.Value" || subjectType == "any" {
		return false
	}
	if !g.goTypeHasNilZeroValue(subjectType) {
		return false
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
		patternType = g.lowerNormalizedTypeExpr(ctx, patternType)
	} else {
		patternType = normalizeTypeExprForPackage(g, "", patternType)
	}
	if ctx != nil && g.staticTypedPatternUsesContextGeneric(ctx, patternType) && g.staticTypedPatternSubjectAllowsNil(ctx, subjectType) {
		return false
	}
	return !g.typeExprAllowsNilInPackage(pkgName, patternType)
}

func (g *generator) staticTypedPatternUsesContextGeneric(ctx *compileContext, expr ast.TypeExpression) bool {
	if g == nil || ctx == nil || expr == nil {
		return false
	}
	return !g.typeExprIsConcreteInPackage(ctx.packageName, expr)
}

func (g *generator) staticTypedPatternSubjectAllowsNil(ctx *compileContext, subjectType string) bool {
	if g == nil || ctx == nil || subjectType == "" || ctx.matchSubjectTypeExpr == nil {
		return false
	}
	subjectExpr := g.lowerNormalizedTypeExpr(ctx, ctx.matchSubjectTypeExpr)
	if subjectExpr == nil || !g.typeExprAllowsNilInPackage(ctx.packageName, subjectExpr) {
		return false
	}
	mapped, ok := g.lowerCarrierType(ctx, subjectExpr)
	if !ok || mapped == "" {
		return false
	}
	return g.typeMatches(mapped, subjectType)
}

func (g *generator) guardStaticTypedPatternNonNil(ctx *compileContext, subjectTemp string, subjectType string, patternType ast.TypeExpression, innerLines []string, innerCond string) ([]string, string, bool) {
	if !g.staticTypedPatternRequiresNonNilGuard(ctx, subjectType, patternType) {
		return innerLines, innerCond, true
	}
	return g.guardMatchConditionWithPredicate(ctx, fmt.Sprintf("(%s != nil)", subjectTemp), innerLines, innerCond)
}
