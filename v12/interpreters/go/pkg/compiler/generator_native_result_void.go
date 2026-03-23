package compiler

import "fmt"

func (g *generator) nativeUnionNilExpr(expected string) (string, bool) {
	if g == nil || expected == "" {
		return "", false
	}
	info := g.nativeUnionInfoForGoType(expected)
	if info == nil {
		return "", false
	}
	for _, member := range info.Members {
		if member == nil || member.TypeExpr == nil {
			continue
		}
		if typeExpressionToString(member.TypeExpr) != "nil" {
			continue
		}
		switch member.GoType {
		case "runtime.Value":
			return fmt.Sprintf("%s(runtime.NilValue{})", member.WrapHelper), true
		case "any":
			return fmt.Sprintf("%s(any(nil))", member.WrapHelper), true
		default:
			if typedNil, ok := g.typedNilExpr(member.GoType); ok {
				return fmt.Sprintf("%s(%s)", member.WrapHelper, typedNil), true
			}
		}
	}
	return "", false
}

func (g *generator) nativeResultVoidSuccessExpr(ctx *compileContext, expected string) (string, bool) {
	if g == nil || expected == "" {
		return "", false
	}
	if info := g.nativeUnionInfoForGoType(expected); info != nil {
		if member, ok := g.nativeUnionMember(info, "struct{}"); ok && member != nil {
			return fmt.Sprintf("%s(struct{}{})", member.WrapHelper), true
		}
	}
	if expr, ok := g.nativeUnionNilExpr(expected); ok {
		return expr, true
	}
	if (expected == "runtime.Value" || expected == "any") && ctx != nil {
		if g.isResultVoidTypeExpr(ctx.expectedTypeExpr) || g.isResultVoidTypeExpr(ctx.returnTypeExpr) {
			return "runtime.VoidValue{}", true
		}
	}
	return "", false
}
