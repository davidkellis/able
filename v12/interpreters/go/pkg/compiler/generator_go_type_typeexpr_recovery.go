package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) recoverKnownConcreteTypeExprForGoType(goType string) (ast.TypeExpression, string, bool) {
	if g == nil || strings.TrimSpace(goType) == "" {
		return nil, "", false
	}
	type recoveredTypeExpr struct {
		pkgName string
		expr    ast.TypeExpression
	}
	var found *recoveredTypeExpr
	consider := func(pkgName string, expr ast.TypeExpression) bool {
		normalized := normalizeTypeExprForPackage(g, pkgName, expr)
		if normalized == nil || !g.typeExprFullyBound(pkgName, normalized) {
			return true
		}
		if found == nil {
			found = &recoveredTypeExpr{pkgName: pkgName, expr: normalized}
			return true
		}
		if normalizeTypeExprString(g, pkgName, normalized) != normalizeTypeExprString(g, found.pkgName, found.expr) {
			return false
		}
		return true
	}
	considerFunction := func(info *functionInfo) bool {
		if info == nil || len(info.Params) == 0 {
			return true
		}
		param := info.Params[0]
		if param.GoType != goType || param.TypeExpr == nil {
			return true
		}
		return consider(info.Package, param.TypeExpr)
	}
	for _, method := range g.methodList {
		if method == nil || !considerFunction(method.Info) {
			return nil, "", false
		}
	}
	for _, impl := range g.implMethodList {
		if impl == nil || !considerFunction(impl.Info) {
			return nil, "", false
		}
	}
	for _, info := range g.specializedFunctions {
		if !considerFunction(info) {
			return nil, "", false
		}
	}
	if found == nil || found.expr == nil {
		return nil, "", false
	}
	return found.expr, found.pkgName, true
}
