package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileNativeInterfaceGenericMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, methodName string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || receiverExpr == "" || receiverType == "" || methodName == "" {
		return nil, "", "", false
	}
	info := g.nativeInterfaceInfoForGoType(receiverType)
	if info == nil {
		return nil, "", "", false
	}
	method, ok := g.nativeInterfaceGenericMethodForGoType(receiverType, methodName)
	if !ok || method == nil {
		return nil, "", "", false
	}
	paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings, ok := g.inferNativeInterfaceGenericMethodShape(ctx, call, method, expected)
	if !ok {
		return nil, "", "", false
	}
	if directLines, expr, retType, ok := g.compileStaticNativeInterfaceGenericDefaultMethodCall(ctx, call, expected, receiverExpr, receiverType, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings, callNode); ok {
		return directLines, expr, retType, true
	}
	// Built-in Array remains an explicit language/kernel container boundary.
	// If the shared generic default-method path cannot materialize the bound
	// accumulator statically, fall back to the dedicated Array collect helper
	// rather than widening the generic path with more nominal-type branches.
	if directLines, expr, retType, ok := g.compileStaticIteratorCollectMonoArrayCall(ctx, call, expected, receiverExpr, receiverType, method, returnGoType, callNode); ok {
		return directLines, expr, retType, true
	}
	lines := make([]string, 0, len(call.Arguments)*5+10)
	receiverTemp := ctx.newTemp()
	receiverValueTemp := ctx.newTemp()
	receiverErrTemp := ctx.newTemp()
	receiverControlTemp := ctx.newTemp()
	lines = append(lines,
		fmt.Sprintf("var %s %s = %s", receiverTemp, receiverType, receiverExpr),
		fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", receiverValueTemp, receiverErrTemp, info.ToRuntimeHelper, receiverTemp),
		fmt.Sprintf("%s := __able_control_from_error(%s)", receiverControlTemp, receiverErrTemp),
	)
	controlLines, ok := g.controlCheckLines(ctx, receiverControlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)

	args := make([]string, 0, len(call.Arguments))
	argTemps := make([]string, 0, len(call.Arguments))
	argValueTemps := make([]string, 0, len(call.Arguments))
	argTypes := make([]string, 0, len(call.Arguments))
	for idx, arg := range call.Arguments {
		expectedType := paramGoTypes[idx]
		expectedTypeExpr := paramTypeExprs[idx]
		argLines, expr, exprType, ok := g.compileExprLinesWithExpectedTypeExpr(ctx, arg, expectedType, expectedTypeExpr)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, argLines...)
		argTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("var %s %s = %s", argTemp, exprType, expr))
		argValueLines, argValueExpr, ok := g.runtimeValueLines(ctx, argTemp, exprType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return nil, "", "", false
		}
		lines = append(lines, argValueLines...)
		args = append(args, argValueExpr)
		argTemps = append(argTemps, argTemp)
		argValueTemps = append(argValueTemps, argValueExpr)
		argTypes = append(argTypes, exprType)
	}

	argList := "nil"
	if len(args) > 0 {
		argList = "[]runtime.Value{" + strings.Join(args, ", ") + "}"
	}
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s, %s := __able_method_call_node(%s, %q, %s, %s)", resultTemp, controlTemp, receiverValueTemp, method.Name, argList, callNode))

	if info.ApplyRuntimeHelper != "" {
		writebackErr := ctx.newTemp()
		writebackControl := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s := %s(%s, %s)", writebackErr, info.ApplyRuntimeHelper, receiverTemp, receiverValueTemp),
			fmt.Sprintf("%s := __able_control_from_error(%s)", writebackControl, writebackErr),
		)
		writebackLines, ok := g.controlCheckLines(ctx, writebackControl)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, writebackLines...)
	}
	for idx, argType := range argTypes {
		iface := g.nativeInterfaceInfoForGoType(argType)
		if iface == nil || iface.ApplyRuntimeHelper == "" {
			continue
		}
		writebackErr := ctx.newTemp()
		writebackControl := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s := %s(%s, %s)", writebackErr, iface.ApplyRuntimeHelper, argTemps[idx], argValueTemps[idx]),
			fmt.Sprintf("%s := __able_control_from_error(%s)", writebackControl, writebackErr),
		)
		writebackLines, ok := g.controlCheckLines(ctx, writebackControl)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, writebackLines...)
	}
	controlLines, ok = g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)

	resultExpr := resultTemp
	resultType := "runtime.Value"
	switch {
	case g.isVoidType(returnGoType):
		lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		resultExpr = "struct{}{}"
		resultType = "struct{}"
	case returnGoType == "runtime.Value":
		resultExpr = resultTemp
		resultType = "runtime.Value"
	case returnGoType == "any":
		resultExpr = resultTemp
		resultType = "any"
	default:
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, returnGoType)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		resultExpr = converted
		resultType = returnGoType
	}
	return g.finishNativeInterfaceGenericCallReturn(ctx, lines, resultExpr, resultType, expected)
}
