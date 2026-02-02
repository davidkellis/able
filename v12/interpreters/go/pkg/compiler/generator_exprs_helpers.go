package compiler

import (
	"fmt"
	"strings"
)

func (g *generator) typeMatches(expected, actual string) bool {
	if expected == "" {
		return true
	}
	return expected == actual
}

func (g *generator) wrapLinesAsExpression(ctx *compileContext, lines []string, expr string, exprType string) (string, string, bool) {
	if len(lines) == 0 {
		return expr, exprType, true
	}
	if expr == "" || exprType == "" {
		ctx.setReason("missing expression")
		return "", "", false
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", exprType, strings.Join(lines, "; "), expr), exprType, true
}
