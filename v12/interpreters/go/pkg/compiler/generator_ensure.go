package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileEnsureExpression(ctx *compileContext, expr *ast.EnsureExpression, expected string) (string, string, bool) {
	if expr == nil || expr.TryExpression == nil {
		ctx.setReason("missing ensure expression")
		return "", "", false
	}
	if expr.EnsureBlock == nil {
		ctx.setReason("missing ensure block")
		return "", "", false
	}
	tryLines, tryExpr, tryType, ok := g.compileTailExpression(ctx, expected, expr.TryExpression)
	if !ok {
		return "", "", false
	}
	resultType := expected
	if resultType == "" {
		resultType = tryType
	}
	if resultType == "" {
		resultType = "runtime.Value"
	}
	if !g.typeMatches(resultType, tryType) {
		ctx.setReason("ensure type mismatch")
		return "", "", false
	}
	ensureLines, ok := g.compileBlockStatement(ctx.child(), expr.EnsureBlock)
	if !ok {
		return "", "", false
	}
	resultTemp := ctx.newTemp()
	recoveredTemp := ctx.newTemp()
	recoveredOkTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("var %s runtime.Value", recoveredTemp),
		fmt.Sprintf("var %s bool", recoveredOkTemp),
		"func() {",
		fmt.Sprintf("\tdefer func() { if recovered := recover(); recovered != nil { if val, ok := recovered.(runtime.Value); ok { %s = val; %s = true } else { panic(recovered) } } }()", recoveredTemp, recoveredOkTemp),
	}
	lines = append(lines, indentLines(tryLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, tryExpr))
	lines = append(lines, "}()")
	lines = append(lines, ensureLines...)
	lines = append(lines, fmt.Sprintf("if %s { panic(%s) }", recoveredOkTemp, recoveredTemp))
	lines = append(lines, fmt.Sprintf("return %s", resultTemp))
	exprValue := fmt.Sprintf("func() %s { %s }()", resultType, strings.Join(lines, "; "))
	return exprValue, resultType, true
}
