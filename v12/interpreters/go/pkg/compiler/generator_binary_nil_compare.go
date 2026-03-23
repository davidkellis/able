package compiler

import "able/interpreter-go/pkg/ast"

func isNilLiteralExpression(expr ast.Expression) bool {
	_, ok := expr.(*ast.NilLiteral)
	return ok
}

func (g *generator) compileStaticNilComparison(expr *ast.BinaryExpression, left string, leftType string, right string, rightType string, expected string) (string, string, bool) {
	if g == nil || expr == nil {
		return "", "", false
	}
	if expr.Operator != "==" && expr.Operator != "!=" {
		return "", "", false
	}
	if expected != "" && expected != "bool" {
		return "", "", false
	}
	if leftType == "" || rightType == "" || leftType != rightType {
		return "", "", false
	}
	if !g.goTypeHasNilZeroValue(leftType) {
		return "", "", false
	}
	if !isNilLiteralExpression(expr.Left) && !isNilLiteralExpression(expr.Right) {
		if typedNil, ok := g.typedNilExpr(leftType); !ok || (left != typedNil && right != typedNil) {
			return "", "", false
		}
	}
	return "(" + left + " " + expr.Operator + " " + right + ")", "bool", true
}
