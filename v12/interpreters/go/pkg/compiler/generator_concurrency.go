package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) awaitExprName(expr *ast.AwaitExpression) string {
	if g.awaitNames == nil {
		g.awaitNames = make(map[*ast.AwaitExpression]string)
	}
	if expr != nil {
		if name, ok := g.awaitNames[expr]; ok {
			return name
		}
	}
	name := fmt.Sprintf("__able_await_expr_%d", len(g.awaitExprs))
	g.awaitExprs = append(g.awaitExprs, name)
	if expr != nil {
		g.awaitNames[expr] = name
	}
	g.needsAst = true
	return name
}

func (g *generator) compileSpawnExpression(ctx *compileContext, expr *ast.SpawnExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil {
		ctx.setReason("missing spawn expression")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" {
		ctx.setReason("spawn return type mismatch")
		return "", "", false
	}
	child := ctx.child()
	child.loopDepth = 0
	child.breakpoints = make(map[string]int)
	child.rethrowVar = ""
	bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(child, "", expr.Expression)
	if !ok {
		return "", "", false
	}
	resultExpr := ""
	if g.isVoidType(bodyType) {
		resultExpr = "runtime.VoidValue{}"
	} else {
		runtimeExpr, ok := g.runtimeValueExpr(bodyExpr, bodyType)
		if !ok {
			ctx.setReason("spawn body unsupported")
			return "", "", false
		}
		resultExpr = runtimeExpr
	}
	taskLines := append([]string{}, bodyLines...)
	taskLines = append(taskLines, fmt.Sprintf("return %s, nil", resultExpr))
	taskBody := strings.Join(taskLines, "; ")
	taskExpr := fmt.Sprintf("func(_ *runtime.Environment) (runtime.Value, error) { %s }", taskBody)
	return fmt.Sprintf("__able_spawn(%s)", taskExpr), "runtime.Value", true
}

func (g *generator) compileAwaitExpression(ctx *compileContext, expr *ast.AwaitExpression, expected string) (string, string, bool) {
	if expr == nil || expr.Expression == nil {
		ctx.setReason("missing await expression")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" && !g.isVoidType(expected) && g.typeCategory(expected) == "unknown" {
		ctx.setReason("await return type mismatch")
		return "", "", false
	}
	iterExpr, iterType, ok := g.compileExpr(ctx, expr.Expression, "")
	if !ok {
		return "", "", false
	}
	iterRuntime, ok := g.runtimeValueExpr(iterExpr, iterType)
	if !ok {
		ctx.setReason("await iterable unsupported")
		return "", "", false
	}
	awaitName := g.awaitExprName(expr)
	awaitExpr := fmt.Sprintf("__able_await(%s, %s)", awaitName, iterRuntime)
	resultType := "runtime.Value"
	if expected != "" && expected != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(awaitExpr, expected)
		if !ok {
			ctx.setReason("await return type mismatch")
			return "", "", false
		}
		return converted, expected, true
	}
	return awaitExpr, resultType, true
}
