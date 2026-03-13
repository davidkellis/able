package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileNativeErrorMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, methodName string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || receiverExpr == "" || receiverType != "runtime.ErrorValue" {
		return nil, "", "", false
	}
	if len(call.Arguments) != 0 {
		return nil, "", "", false
	}
	switch methodName {
	case "message":
		resultExpr := fmt.Sprintf("%s.Message", receiverExpr)
		resultType := "string"
		if expected != "" && !g.typeMatches(expected, resultType) {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		return nil, resultExpr, resultType, true
	case "cause":
		resultType := "*runtime.ErrorValue"
		if expected != "" && !g.typeMatches(expected, resultType) {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		resultTemp := ctx.newTemp()
		runtimeTemp := ctx.newTemp()
		okTemp := ctx.newTemp()
		nilPtrTemp := ctx.newTemp()
		copyTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := (*runtime.ErrorValue)(nil)", resultTemp),
			fmt.Sprintf("if %s.Payload != nil {", receiverExpr),
			fmt.Sprintf("\tif causeValue, ok := %s.Payload[\"cause\"]; ok && causeValue != nil {", receiverExpr),
		}
		convTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("\t\t%s, %s, %s := __able_runtime_error_value(causeValue)", runtimeTemp, okTemp, nilPtrTemp),
			fmt.Sprintf("\t\tif %s || %s {", okTemp, nilPtrTemp),
			fmt.Sprintf("\t\t\tif %s && !%s {", okTemp, nilPtrTemp),
			fmt.Sprintf("\t\t\t\t%s := %s", copyTemp, runtimeTemp),
			fmt.Sprintf("\t\t\t\t%s = &%s", resultTemp, copyTemp),
			"\t\t\t}",
			"\t\t} else {",
			fmt.Sprintf("\t\t\t%s, %s := __able_nullable_error_from_value(causeValue)", convTemp, errTemp),
			fmt.Sprintf("\t\t\tif %s != nil { bridge.RaiseRuntimeErrorWithContext(__able_runtime, %s, %s) }", errTemp, callNode, errTemp),
			fmt.Sprintf("\t\t\t%s = %s", resultTemp, convTemp),
			"\t\t}",
			"\t}",
			"}",
		)
		return lines, resultTemp, resultType, true
	default:
		return nil, "", "", false
	}
}
