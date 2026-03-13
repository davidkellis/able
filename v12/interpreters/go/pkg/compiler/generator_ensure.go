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

	tryControlTemp := ctx.newTemp()
	tryDoneLabel := ctx.newTemp()
	tryCtx := ctx.child()
	tryCtx.controlCaptureVar = tryControlTemp
	tryCtx.controlCaptureLabel = tryDoneLabel
	tryLines, tryExpr, tryType, ok := g.compileTailExpression(tryCtx, expected, expr.TryExpression)
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

	ensureControlTemp := ctx.newTemp()
	ensureDoneLabel := ctx.newTemp()
	ensureCtx := ctx.child()
	ensureCtx.controlCaptureVar = ensureControlTemp
	ensureCtx.controlCaptureLabel = ensureDoneLabel
	ensureLines, ok := g.compileBlockStatement(ensureCtx, expr.EnsureBlock)
	if !ok {
		return nil, "", "", false
	}

	resultTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("var %s %s", resultTemp, resultType),
		fmt.Sprintf("var %s *__ableControl", tryControlTemp),
		fmt.Sprintf("var %s *__ableControl", ensureControlTemp),
		"{",
	}
	lines = append(lines, indentLines(tryLines, 1)...)
	lines = append(lines, fmt.Sprintf("\t%s = %s", resultTemp, tryExpr))
	lines = append(lines, "}")
	lines = append(lines, fmt.Sprintf("if false { goto %s }", tryDoneLabel))
	lines = append(lines, fmt.Sprintf("%s:", tryDoneLabel))
	lines = append(lines, "{")
	lines = append(lines, indentLines(ensureLines, 1)...)
	lines = append(lines, "}")
	lines = append(lines, fmt.Sprintf("if false { goto %s }", ensureDoneLabel))
	lines = append(lines, fmt.Sprintf("%s:", ensureDoneLabel))

	ensureTransferLines, ok := g.controlTransferLines(ctx, ensureControlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, fmt.Sprintf("if %s != nil {", ensureControlTemp))
	lines = append(lines, indentLines(ensureTransferLines, 1)...)
	lines = append(lines, "}")

	tryTransferLines, ok := g.controlTransferLines(ctx, tryControlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, fmt.Sprintf("if %s != nil {", tryControlTemp))
	lines = append(lines, indentLines(tryTransferLines, 1)...)
	lines = append(lines, "}")
	return lines, resultTemp, resultType, true
}
