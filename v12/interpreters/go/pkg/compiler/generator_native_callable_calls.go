package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileNativeCallableCall(ctx *compileContext, call *ast.FunctionCall, expected string, callableExpr string, callableType string, fnTypeExpr *ast.FunctionTypeExpression, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || callableExpr == "" {
		return nil, "", "", false
	}
	info := g.nativeCallableInfoForGoType(callableType)
	if info == nil && fnTypeExpr != nil {
		var ok bool
		info, ok = g.ensureNativeCallableInfo(ctx.packageName, fnTypeExpr)
		if !ok {
			return nil, "", "", false
		}
	}
	if info == nil {
		return nil, "", "", false
	}
	if len(call.Arguments) != len(info.ParamGoTypes) {
		ctx.setReason("call arity mismatch")
		return nil, "", "", false
	}

	lines := make([]string, 0, len(call.Arguments)*4+8)
	callableTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("var %s %s = %s", callableTemp, info.GoType, callableExpr))
	args := make([]string, 0, len(call.Arguments))
	for idx, arg := range call.Arguments {
		expectedType := info.ParamGoTypes[idx]
		var expectedTypeExpr ast.TypeExpression
		if idx < len(info.ParamTypeExprs) {
			expectedTypeExpr = info.ParamTypeExprs[idx]
		}
		argLines, expr, _, ok := g.compileExprLinesWithExpectedTypeExpr(ctx, arg, expectedType, expectedTypeExpr)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, argLines...)
		args = append(args, expr)
	}

	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("var %s %s", resultTemp, info.ReturnGoType))
	lines = append(lines, fmt.Sprintf("var %s *__ableControl", controlTemp))
	lines = append(lines, fmt.Sprintf("__able_push_call_frame(%s)", callNode))
	lines = append(lines, fmt.Sprintf("if %s == nil {", callableTemp))
	lines = append(lines, fmt.Sprintf("\t%s = %s", controlTemp, g.runtimeErrorControlExpr(callNode, `fmt.Errorf("missing callable value")`)))
	lines = append(lines, "} else {")
	lines = append(lines, fmt.Sprintf("\t%s, %s = %s(%s)", resultTemp, controlTemp, callableTemp, strings.Join(args, ", ")))
	lines = append(lines, "}")
	lines = append(lines, "__able_pop_call_frame()")
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	if g.isVoidType(expected) {
		if !g.isVoidType(info.ReturnGoType) {
			lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		}
		return lines, "struct{}{}", "struct{}", true
	}
	if expected == "runtime.Value" && info.ReturnGoType != "runtime.Value" {
		convLines, converted, ok := g.lowerRuntimeValue(ctx, resultTemp, info.ReturnGoType)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, "runtime.Value", true
	}
	if expected == "" || g.typeMatches(expected, info.ReturnGoType) {
		return lines, resultTemp, info.ReturnGoType, true
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" && g.canCoerceStaticExpr(expected, info.ReturnGoType) {
		return g.lowerCoerceExpectedStaticExpr(ctx, lines, resultTemp, info.ReturnGoType, expected)
	}
	ctx.setReason("call return type mismatch")
	return nil, "", "", false
}
