package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) nominalTargetTypeExprCompatible(pkgName string, actualExpr ast.TypeExpression, targetExpr ast.TypeExpression) bool {
	if g == nil || actualExpr == nil || targetExpr == nil {
		return true
	}
	actualExpr = normalizeTypeExprForPackage(g, pkgName, actualExpr)
	targetExpr = normalizeTypeExprForPackage(g, pkgName, targetExpr)
	if normalizeTypeExprString(g, pkgName, actualExpr) == normalizeTypeExprString(g, pkgName, targetExpr) {
		return true
	}
	actualBase, actualBaseOK := typeExprBaseName(normalizeTypeExprForPackage(g, pkgName, actualExpr))
	targetBase, targetBaseOK := typeExprBaseName(normalizeTypeExprForPackage(g, pkgName, targetExpr))
	if actualBaseOK && targetBaseOK {
		if actualBase == "" || targetBase == "" {
			return false
		}
		if actualBase != targetBase {
			return false
		}
	}
	actualInfo, actualOK := g.structInfoForTypeExpr(pkgName, actualExpr)
	targetInfo, targetOK := g.structInfoForTypeExpr(pkgName, targetExpr)
	if !actualOK || !targetOK || actualInfo == nil || targetInfo == nil {
		return false
	}
	return actualInfo.Package == targetInfo.Package && actualInfo.Name != "" && actualInfo.Name == targetInfo.Name
}

func (g *generator) usesNominalStructCarrier(pkgName string, expr ast.TypeExpression) bool {
	if g == nil || expr == nil {
		return false
	}
	info, ok := g.structInfoForTypeExpr(pkgName, normalizeTypeExprForPackage(g, pkgName, expr))
	return ok && info != nil
}

func (g *generator) nominalTargetTemplateMayBindLater(pkgName string, template ast.TypeExpression, actual ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = normalizeTypeExprForPackage(g, pkgName, template)
	actual = normalizeTypeExprForPackage(g, pkgName, actual)
	templateBase, templateOK := typeExprBaseName(template)
	actualBase, actualOK := typeExprBaseName(actual)
	if !templateOK || !actualOK || templateBase == "" || actualBase == "" || templateBase != actualBase {
		return false
	}
	return !g.typeExprFullyBound(pkgName, actual)
}

func (g *generator) specializedTargetMatchesOrDefers(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	if g.specializedTypeTemplateMatches(pkgName, template, actual, genericNames, bindings, make(map[string]struct{})) {
		return true
	}
	return g.nominalTargetTemplateMayBindLater(pkgName, template, actual)
}
