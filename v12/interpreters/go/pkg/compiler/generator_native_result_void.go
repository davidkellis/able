package compiler

import "fmt"

func (g *generator) nativeResultVoidSuccessExpr(ctx *compileContext, expected string) (string, bool) {
	if g == nil || expected == "" {
		return "", false
	}
	if info := g.nativeUnionInfoForGoType(expected); info != nil {
		if member, ok := g.nativeUnionMember(info, "struct{}"); ok && member != nil {
			return fmt.Sprintf("%s(struct{}{})", member.WrapHelper), true
		}
	}
	if (expected == "runtime.Value" || expected == "any") && ctx != nil {
		if g.isResultVoidTypeExpr(ctx.expectedTypeExpr) || g.isResultVoidTypeExpr(ctx.returnTypeExpr) {
			return "runtime.VoidValue{}", true
		}
	}
	return "", false
}
