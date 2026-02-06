package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) compileStructUpdateFallback(ctx *compileContext, lit *ast.StructLiteral, structType string, expected string) (string, string, bool, bool) {
	if g == nil || ctx == nil || lit == nil || len(lit.FunctionalUpdateSources) == 0 {
		return "", "", false, false
	}
	prevReason := ctx.reason
	for _, source := range lit.FunctionalUpdateSources {
		if source == nil {
			ctx.setReason("functional update source missing")
			return "", "", false, true
		}
		if _, _, ok := g.compileExpr(ctx, source, structType); ok {
			continue
		}
		failReason := ctx.reason
		ctx.reason = prevReason
		if _, _, ok := g.compileExpr(ctx, source, "runtime.Value"); ok {
			ctx.reason = prevReason
			expr, exprType, ok := g.compileStructLiteralRuntime(ctx, lit)
			if !ok {
				return "", "", false, true
			}
			if expected != "" && expected != "runtime.Value" {
				converted, ok := g.expectRuntimeValueExpr(expr, expected)
				if !ok {
					ctx.setReason("struct literal type mismatch")
					return "", "", false, true
				}
				return converted, expected, true, true
			}
			return expr, exprType, true, true
		}
		ctx.reason = failReason
		return "", "", false, true
	}
	ctx.reason = prevReason
	return "", "", false, false
}
