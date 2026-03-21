package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileStaticIteratorControllerCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, methodName string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || receiverExpr == "" || receiverType != "*__able_generator" || methodName == "" {
		return nil, "", "", false
	}
	switch methodName {
	case "yield":
		if len(call.Arguments) > 1 {
			ctx.setReason("call arity mismatch")
			return nil, "", "", false
		}
		lines := []string{}
		valueExpr := "runtime.NilValue{}"
		if len(call.Arguments) == 1 {
			argLines, argExpr, argType, ok := g.compileExprLines(ctx, call.Arguments[0], "")
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, argLines...)
			argConvLines, argValueExpr, ok := g.runtimeValueLines(ctx, argExpr, argType)
			if !ok {
				ctx.setReason("call argument unsupported")
				return nil, "", "", false
			}
			lines = append(lines, argConvLines...)
			valueExpr = argValueExpr
		}
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s := %s.emit(%s)", errTemp, receiverExpr, valueExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		)
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		return lines, "runtime.NilValue{}", "runtime.Value", true
	case "stop":
		if len(call.Arguments) != 0 {
			ctx.setReason("call arity mismatch")
			return nil, "", "", false
		}
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s.stop()", errTemp, receiverExpr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		return lines, "runtime.NilValue{}", "runtime.Value", true
	case "close":
		if len(call.Arguments) != 0 {
			ctx.setReason("call arity mismatch")
			return nil, "", "", false
		}
		lines := []string{fmt.Sprintf("%s.close()", receiverExpr)}
		return lines, "runtime.NilValue{}", "runtime.Value", true
	default:
		return nil, "", "", false
	}
}

func (g *generator) compileStaticIteratorControllerBoundMethodValue(ctx *compileContext, receiverExpr string, receiverType string, methodName string, expected string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || receiverExpr == "" || receiverType != "*__able_generator" || methodName == "" || expected == "" {
		return nil, "", "", false
	}
	callableInfo := g.nativeCallableInfoForGoType(expected)
	if callableInfo == nil {
		return nil, "", "", false
	}
	receiverTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("var %s %s = %s", receiverTemp, receiverType, receiverExpr)}
	switch methodName {
	case "yield":
		if len(callableInfo.ParamGoTypes) != 1 || !g.isVoidType(callableInfo.ReturnGoType) {
			return nil, "", "", false
		}
		argName := "arg0"
		argValueExpr, ok := g.runtimeValueExpr(argName, callableInfo.ParamGoTypes[0])
		if !ok {
			return nil, "", "", false
		}
		callableExpr := fmt.Sprintf("%s(func(%s %s) (%s, *__ableControl) { __able_yield_err := %s.emit(%s); __able_yield_control := __able_control_from_error(__able_yield_err); if __able_yield_control != nil { return struct{}{}, __able_yield_control }; return struct{}{}, nil })", callableInfo.GoType, argName, callableInfo.ParamGoTypes[0], callableInfo.ReturnGoType, receiverTemp, argValueExpr)
		return lines, callableExpr, callableInfo.GoType, true
	case "stop":
		if len(callableInfo.ParamGoTypes) != 0 || !g.isVoidType(callableInfo.ReturnGoType) {
			return nil, "", "", false
		}
		callableExpr := fmt.Sprintf("%s(func() (%s, *__ableControl) { __able_stop_err := %s.stop(); __able_stop_control := __able_control_from_error(__able_stop_err); if __able_stop_control != nil { return struct{}{}, __able_stop_control }; return struct{}{}, nil })", callableInfo.GoType, callableInfo.ReturnGoType, receiverTemp)
		return lines, callableExpr, callableInfo.GoType, true
	case "close":
		if len(callableInfo.ParamGoTypes) != 0 || !g.isVoidType(callableInfo.ReturnGoType) {
			return nil, "", "", false
		}
		callableExpr := fmt.Sprintf("%s(func() (%s, *__ableControl) { %s.close(); return struct{}{}, nil })", callableInfo.GoType, callableInfo.ReturnGoType, receiverTemp)
		return lines, callableExpr, callableInfo.GoType, true
	default:
		return nil, "", "", false
	}
}
