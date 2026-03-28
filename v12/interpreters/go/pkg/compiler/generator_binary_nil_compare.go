package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

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

func (g *generator) compileStaticNullableEqualityComparison(expr *ast.BinaryExpression, left string, leftType string, right string, rightType string, expected string) (string, string, bool) {
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
	innerType, ok := g.nativeNullableValueInnerType(leftType)
	if !ok || !g.isEqualityComparable(innerType) {
		return "", "", false
	}
	eqExpr := fmt.Sprintf("((%s == nil && %s == nil) || (%s != nil && %s != nil && (*%s == *%s)))", left, right, left, right, left, right)
	if expr.Operator == "!=" {
		return "(!" + eqExpr + ")", "bool", true
	}
	return eqExpr, "bool", true
}

func (g *generator) compileRuntimeNilComparison(ctx *compileContext, expr *ast.BinaryExpression, left string, leftType string, right string, rightType string, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || expr == nil {
		return nil, "", "", false
	}
	if expr.Operator != "==" && expr.Operator != "!=" {
		return nil, "", "", false
	}
	if expected != "" && expected != "bool" {
		return nil, "", "", false
	}
	runtimeExpr := ""
	runtimeType := ""
	switch {
	case isNilLiteralExpression(expr.Left) && (rightType == "runtime.Value" || rightType == "any"):
		runtimeExpr = right
		runtimeType = rightType
	case isNilLiteralExpression(expr.Right) && (leftType == "runtime.Value" || leftType == "any"):
		runtimeExpr = left
		runtimeType = leftType
	default:
		return nil, "", "", false
	}
	if runtimeExpr == "" || runtimeType == "" {
		return nil, "", "", false
	}
	checkExpr := runtimeExpr
	if runtimeType == "any" {
		checkExpr = "__able_any_to_value(" + runtimeExpr + ")"
	}
	if expr.Operator == "!=" {
		return nil, "(!__able_is_nil(" + checkExpr + "))", "bool", true
	}
	return nil, "__able_is_nil(" + checkExpr + ")", "bool", true
}
