package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) nominalTargetTypeExprCompatible(pkgName string, actualExpr ast.TypeExpression, targetExpr ast.TypeExpression) bool {
	if g == nil || actualExpr == nil || targetExpr == nil {
		return true
	}
	actualInfo, actualOK := g.structInfoForTypeExpr(pkgName, actualExpr)
	targetInfo, targetOK := g.structInfoForTypeExpr(pkgName, targetExpr)
	if !actualOK || !targetOK || actualInfo == nil || targetInfo == nil {
		return true
	}
	return actualInfo.Package == targetInfo.Package && actualInfo.Name != "" && actualInfo.Name == targetInfo.Name
}
