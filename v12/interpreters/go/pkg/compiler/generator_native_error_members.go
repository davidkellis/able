package compiler

import "fmt"

func (g *generator) compileNativeErrorMemberAccess(ctx *compileContext, objectExpr string, memberName string, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || objectExpr == "" || memberName == "" {
		return nil, "", "", false
	}
	if memberName != "value" {
		return nil, "", "", false
	}
	resultTemp := ctx.newTemp()
	lines := []string{
		fmt.Sprintf("%s := runtime.Value(runtime.NilValue{})", resultTemp),
		fmt.Sprintf("if %s.Payload != nil {", objectExpr),
		fmt.Sprintf("\tif payloadValue, ok := %s.Payload[\"value\"]; ok {", objectExpr),
		fmt.Sprintf("\t\t%s = payloadValue", resultTemp),
		"\t}",
		"}",
	}
	if expected == "" || expected == "runtime.Value" || expected == "any" {
		return lines, resultTemp, "runtime.Value", true
	}
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
	if !ok {
		ctx.setReason("member access type mismatch")
		return nil, "", "", false
	}
	lines = append(lines, convLines...)
	return lines, converted, expected, true
}
