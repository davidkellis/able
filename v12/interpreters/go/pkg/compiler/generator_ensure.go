package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileEnsureExpression(ctx *compileContext, expr *ast.EnsureExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.TryExpression == nil {
		ctx.setReason("missing ensure expression")
		return nil, "", "", false
	}
	if expr.EnsureBlock == nil {
		ctx.setReason("missing ensure block")
		return nil, "", "", false
	}
	tryLines, tryExpr, tryType, ok := g.compileTailExpression(ctx, expected, expr.TryExpression)
	if !ok {
		return nil, "", "", false
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
		return nil, "", "", false
	}
	ensureLines, ok := g.compileBlockStatement(ctx.child(), expr.EnsureBlock)
	if !ok {
		return nil, "", "", false
	}
	resultTemp := ctx.newTemp()
	recoveredTemp := ctx.newTemp()
	recoveredOkTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("var %s any", recoveredTemp),
		fmt.Sprintf("var %s bool", recoveredOkTemp),
		"func() {",
		fmt.Sprintf("\tdefer func() { if recovered := recover(); recovered != nil { %s = recovered; %s = true } }()", recoveredTemp, recoveredOkTemp),
	}
	lines = append(lines, indentLines(tryLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, tryExpr))
	lines = append(lines, "}()")
	lines = append(lines, ensureLines...)
	lines = append(lines, fmt.Sprintf("if %s { panic(%s) }", recoveredOkTemp, recoveredTemp))
	return lines, resultTemp, resultType, true
}
