package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileNativeInterfaceMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, methodName string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || receiverExpr == "" || receiverType == "" || methodName == "" {
		return nil, "", "", false
	}
	if g.nativeInterfaceInfoForGoType(receiverType) == nil {
		return nil, "", "", false
	}
	method, ok := g.nativeInterfaceMethodForGoType(receiverType, methodName)
	if !ok || method == nil {
		return nil, "", "", false
	}
	callArgCount := len(call.Arguments)
	paramCount := len(method.ParamGoTypes)
	if callArgCount != paramCount {
		if !(method.OptionalLast && callArgCount == paramCount-1) {
			ctx.setReason("call arity mismatch")
			return nil, "", "", false
		}
	}
	args := make([]string, 0, len(method.ParamGoTypes))
	var argLines []string
	for idx, arg := range call.Arguments {
		expectedType := method.ParamGoTypes[idx]
		var expectedTypeExpr ast.TypeExpression
		if idx < len(method.ParamTypeExprs) {
			expectedTypeExpr = method.ParamTypeExprs[idx]
		}
		nextLines, expr, _, ok := g.compileExprLinesWithExpectedTypeExpr(ctx, arg, expectedType, expectedTypeExpr)
		if !ok {
			return nil, "", "", false
		}
		argLines = append(argLines, nextLines...)
		args = append(args, expr)
	}
	if method.OptionalLast && len(call.Arguments) == len(method.ParamGoTypes)-1 {
		zeroExpr, ok := g.zeroValueExpr(method.ParamGoTypes[len(method.ParamGoTypes)-1])
		if !ok {
			ctx.setReason("call arity mismatch")
			return nil, "", "", false
		}
		args = append(args, zeroExpr)
	}
	callExpr := fmt.Sprintf("%s.%s(%s)", receiverExpr, method.GoName, strings.Join(args, ", "))
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines := append([]string{}, argLines...)
	lines = append(lines, fmt.Sprintf("__able_push_call_frame(%s)", callNode))
	lines = append(lines, fmt.Sprintf("%s, %s := %s", resultTemp, controlTemp, callExpr))
	lines = append(lines, "__able_pop_call_frame()")
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	if expected == "" || g.typeMatches(expected, method.ReturnGoType) {
		return lines, resultTemp, method.ReturnGoType, true
	}
	if expected != "runtime.Value" && method.ReturnGoType == "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if expected == "runtime.Value" && method.ReturnGoType != "runtime.Value" {
		convLines, converted, ok := g.runtimeValueLines(ctx, resultTemp, method.ReturnGoType)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, "runtime.Value", true
	}
	if expected != "" && expected != "any" && method.ReturnGoType == "any" {
		anyTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", anyTemp, resultTemp))
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, anyTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" && g.canCoerceStaticExpr(expected, method.ReturnGoType) {
		return g.coerceExpectedStaticExpr(ctx, lines, resultTemp, method.ReturnGoType, expected)
	}
	ctx.setReason("call return type mismatch")
	return nil, "", "", false
}
