package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileIteratorLiteral(ctx *compileContext, expr *ast.IteratorLiteral, expected string) (string, string, bool) {
	if expr == nil {
		ctx.setReason("missing iterator literal")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" {
		ctx.setReason("iterator literal type mismatch")
		return "", "", false
	}
	binding := "gen"
	if expr.Binding != nil && expr.Binding.Name != "" {
		binding = expr.Binding.Name
	}
	genParam := "__able_gen"
	bodyCtx := ctx.child()
	bodyCtx.locals[binding] = paramInfo{Name: binding, GoName: genParam, GoType: "runtime.Value"}
	if binding != "gen" {
		bodyCtx.locals["gen"] = paramInfo{Name: "gen", GoName: genParam, GoType: "runtime.Value"}
	}
	lines := make([]string, 0, len(expr.Body)+1)
	for _, stmt := range expr.Body {
		stmtLines, ok := g.compileStatement(bodyCtx, stmt)
		if !ok {
			return "", "", false
		}
		lines = append(lines, stmtLines...)
	}
	lines = append(lines, "return nil")
	run := fmt.Sprintf("func(%s runtime.Value) error { %s }", genParam, strings.Join(lines, "; "))
	g.needsIterator = true
	return fmt.Sprintf("__able_new_iterator(%s)", run), "runtime.Value", true
}
