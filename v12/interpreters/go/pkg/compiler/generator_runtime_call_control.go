package compiler

import "fmt"

func (g *generator) appendRuntimeCallControlLines(ctx *compileContext, lines []string, callExpr string) ([]string, string, bool) {
	if g == nil || ctx == nil || callExpr == "" {
		return nil, "", false
	}
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := %s", resultTemp, controlTemp, callExpr))
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, "", false
	}
	lines = append(lines, controlLines...)
	return lines, resultTemp, true
}

func (g *generator) appendRuntimeMemberGetControlLines(ctx *compileContext, lines []string, objExpr string, memberExpr string) ([]string, string, bool) {
	return g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_member_get(%s, %s)", objExpr, memberExpr))
}

func (g *generator) appendRuntimeMemberSetControlLines(ctx *compileContext, lines []string, objExpr string, memberExpr string, valueExpr string) ([]string, string, bool) {
	return g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_member_set(%s, %s, %s)", objExpr, memberExpr, valueExpr))
}

func (g *generator) appendRuntimeMemberGetMethodControlLines(ctx *compileContext, lines []string, objExpr string, memberExpr string) ([]string, string, bool) {
	return g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_member_get_method(%s, %s)", objExpr, memberExpr))
}
