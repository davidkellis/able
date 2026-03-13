package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileCharLiteral(ctx *compileContext, lit *ast.CharLiteral, expected string) (string, string, bool) {
	if lit == nil {
		ctx.setReason("missing char literal")
		return "", "", false
	}
	actual := "rune"
	runes := []rune(lit.Value)
	if len(runes) != 1 {
		ctx.setReason("invalid char literal")
		return "", "", false
	}
	if g.nativeNullableWraps(expected, actual) {
		return fmt.Sprintf("__able_ptr(rune(%q))", runes[0]), expected, true
	}
	if !g.typeMatches(expected, actual) {
		ctx.setReason("unsupported char literal type")
		return "", "", false
	}
	return fmt.Sprintf("rune(%q)", runes[0]), actual, true
}
