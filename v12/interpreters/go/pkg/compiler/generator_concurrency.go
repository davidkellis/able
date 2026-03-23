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

func (g *generator) compileSpawnExpression(ctx *compileContext, expr *ast.SpawnExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Expression == nil {
		ctx.setReason("missing spawn expression")
		return nil, "", "", false
	}
	if expected != "" && expected != "runtime.Value" {
		ctx.setReason("spawn return type mismatch")
		return nil, "", "", false
	}
	child := ctx.child()
	child.loopDepth = 0
	child.breakpoints = make(map[string]int)
	child.rethrowVar = ""
	child.controlMode = compileControlModeRuntimeValueError
	bodyLines, bodyExpr, bodyType, ok := g.compileTailExpression(child, "", expr.Expression)
	if !ok {
		return nil, "", "", false
	}
	resultExpr := ""
	var convLines []string
	if g.isVoidType(bodyType) {
		voidChild := ctx.child()
		voidChild.loopDepth = 0
		voidChild.breakpoints = make(map[string]int)
		voidChild.rethrowVar = ""
		voidChild.controlMode = compileControlModeRuntimeValueError
		voidLines, stmtOK := g.compileStatement(voidChild, expr.Expression)
		if stmtOK {
			bodyLines = voidLines
		}
		resultExpr = "runtime.VoidValue{}"
	} else {
		cl, runtimeExpr, ok := g.lowerRuntimeValue(child, bodyExpr, bodyType)
		if !ok {
			ctx.setReason("spawn body unsupported")
			return nil, "", "", false
		}
		convLines = cl
		resultExpr = runtimeExpr
	}
	taskLines := append([]string{}, bodyLines...)
	taskLines = append(taskLines, convLines...)
	taskLines = append(taskLines, fmt.Sprintf("return %s, nil", resultExpr))
	taskBody := strings.Join(taskLines, "; ")
	taskExpr := fmt.Sprintf("func(_ *runtime.Environment) (runtime.Value, error) { %s }", taskBody)
	return nil, fmt.Sprintf("__able_spawn(%s)", taskExpr), "runtime.Value", true
}

func (g *generator) compileAwaitExpression(ctx *compileContext, expr *ast.AwaitExpression, expected string) ([]string, string, string, bool) {
	if expr == nil || expr.Expression == nil {
		ctx.setReason("missing await expression")
		return nil, "", "", false
	}
	if !g.isStaticallyKnownExpectedType(expected) {
		ctx.setReason("await return type mismatch")
		return nil, "", "", false
	}
	iterLines, iterExpr, iterType, ok := g.compileExprLines(ctx, expr.Expression, "")
	if !ok {
		return nil, "", "", false
	}
	iterConvLines, iterRuntime, ok := g.lowerRuntimeValue(ctx, iterExpr, iterType)
	if !ok {
		ctx.setReason("await iterable unsupported")
		return nil, "", "", false
	}
	var lines []string
	lines = append(lines, iterLines...)
	lines = append(lines, iterConvLines...)
	awaitName := g.awaitExprName(expr)
	awaitExpr := fmt.Sprintf("__able_await(%s, %s)", awaitName, iterRuntime)
	if expected != "" && expected != "runtime.Value" {
		expectLines, converted, ok := g.lowerExpectRuntimeValue(ctx, awaitExpr, expected)
		if !ok {
			ctx.setReason("await return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, expectLines...)
		return lines, converted, expected, true
	}
	return lines, awaitExpr, "runtime.Value", true
}
