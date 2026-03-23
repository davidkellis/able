package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) compileStructUpdateFallback(ctx *compileContext, lit *ast.StructLiteral, structType string, expected string) ([]string, string, string, bool, bool) {
	if g == nil || ctx == nil || lit == nil || len(lit.FunctionalUpdateSources) == 0 {
		return nil, "", "", false, false
	}
	prevReason := ctx.reason
	for _, source := range lit.FunctionalUpdateSources {
		if source == nil {
			ctx.setReason("functional update source missing")
			return nil, "", "", false, true
		}
		if _, _, ok := g.compileExpr(ctx, source, structType); ok {
			continue
		}
		failReason := ctx.reason
		ctx.reason = prevReason
		if _, _, ok := g.compileExpr(ctx, source, "runtime.Value"); ok {
			ctx.reason = prevReason
			lines, expr, exprType, ok := g.compileStructLiteralRuntime(ctx, lit)
			if !ok {
				return nil, "", "", false, true
			}
			if expected != "" && expected != "runtime.Value" {
				convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, expr, expected)
				if !ok {
					ctx.setReason("struct literal type mismatch")
					return nil, "", "", false, true
				}
				lines = append(lines, convLines...)
				return lines, converted, expected, true, true
			}
			return lines, expr, exprType, true, true
		}
		ctx.reason = failReason
		return nil, "", "", false, true
	}
	ctx.reason = prevReason
	return nil, "", "", false, false
}
