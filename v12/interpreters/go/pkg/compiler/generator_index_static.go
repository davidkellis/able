package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileableInterfaceMethodForReceiverArity(goType string, interfaceName string, methodName string, arity int) *functionInfo {
	if g == nil || goType == "" || goType == "runtime.Value" || goType == "any" || interfaceName == "" || methodName == "" {
		return nil
	}
	var found *functionInfo
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable {
			continue
		}
		if impl.InterfaceName != interfaceName || impl.MethodName != methodName {
			continue
		}
		if len(impl.Info.Params) != arity+1 || len(impl.Info.Params) == 0 || impl.Info.Params[0].GoType == "" {
			continue
		}
		receiverType := impl.Info.Params[0].GoType
		if receiverType != goType && !g.typeMatches(receiverType, goType) {
			continue
		}
		if found != nil && found != impl.Info {
			return nil
		}
		found = impl.Info
	}
	return found
}

func (g *generator) prepareStaticCallArg(ctx *compileContext, expr string, actual string, expected string) ([]string, string, string, bool) {
	if expected == "" || expected == actual {
		return nil, expr, actual, true
	}
	if expected == "runtime.Value" {
		lines, converted, ok := g.runtimeValueLines(ctx, expr, actual)
		if !ok {
			return nil, "", "", false
		}
		return lines, converted, "runtime.Value", true
	}
	return g.coerceExpectedStaticExpr(ctx, nil, expr, actual, expected)
}

func (g *generator) compileStaticIndexGet(ctx *compileContext, expr *ast.IndexExpression, expected string, receiverExpr string, receiverType string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || expr == nil || receiverExpr == "" || receiverType == "" {
		return nil, "", "", false
	}
	synthetic := ast.NewFunctionCall(ast.NewIdentifier("get"), []ast.Expression{expr.Index}, nil, false)
	callNode := g.diagNodeName(synthetic, "*ast.FunctionCall", "call")
	if _, ok := g.nativeInterfaceMethodForGoType(receiverType, "get"); ok {
		return g.compileNativeInterfaceMethodCall(ctx, synthetic, expected, receiverExpr, receiverType, "get", callNode)
	}
	if info := g.compileableInterfaceMethodForReceiverArity(receiverType, "Index", "get", 1); info != nil {
		method := &methodInfo{MethodName: "get", ExpectsSelf: true, Info: info}
		return g.compileResolvedMethodCall(ctx, synthetic, expected, method, receiverExpr, receiverType, callNode)
	}
	return nil, "", "", false
}

func (g *generator) finalizeStaticIndexSetResult(ctx *compileContext, lines []string, setExpr string, setType string, assignedExpr string, assignedType string) ([]string, string, string, bool) {
	assignedLines, assignedValue, ok := g.runtimeValueLines(ctx, assignedExpr, assignedType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, assignedLines...)
	switch {
	case g.isVoidType(setType):
		return lines, assignedValue, "runtime.Value", true
	case setType == "runtime.Value" || setType == "any":
		probeExpr := setExpr
		if setType == "any" {
			probeTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", probeTemp, setExpr))
			probeExpr = probeTemp
		}
		errorTemp := ctx.newTemp()
		errorOKTemp := ctx.newTemp()
		nilPtrTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s, %s, %s := __able_runtime_error_value(%s)", errorTemp, errorOKTemp, nilPtrTemp, probeExpr),
			fmt.Sprintf("var %s runtime.Value", resultTemp),
			fmt.Sprintf("if %s && !%s { %s = %s } else { %s = %s }", errorOKTemp, nilPtrTemp, resultTemp, errorTemp, resultTemp, assignedValue),
		)
		return lines, resultTemp, "runtime.Value", true
	case setType == "runtime.ErrorValue":
		errorLines, errorValue, ok := g.runtimeValueLines(ctx, setExpr, setType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, errorLines...)
		return lines, errorValue, "runtime.Value", true
	default:
		successMember, failureMember, ok := g.nativeUnionOrElseMembers(setType)
		if !ok || successMember == nil || failureMember == nil {
			return nil, "", "", false
		}
		failureTemp := ctx.newTemp()
		failureOKTemp := ctx.newTemp()
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s, %s := %s(%s)", failureTemp, failureOKTemp, failureMember.UnwrapHelper, setExpr))
		failureLines, failureValue, ok := g.runtimeValueLines(ctx, failureTemp, failureMember.GoType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, failureLines...)
		lines = append(lines,
			fmt.Sprintf("var %s runtime.Value", resultTemp),
			fmt.Sprintf("if %s { %s = %s } else { %s = %s }", failureOKTemp, resultTemp, failureValue, resultTemp, assignedValue),
		)
		return lines, resultTemp, "runtime.Value", true
	}
}

func (g *generator) compileStaticIndexSet(ctx *compileContext, target *ast.IndexExpression, receiverExpr string, receiverType string, valueExpr string, valueType string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || target == nil || receiverExpr == "" || receiverType == "" || valueExpr == "" || valueType == "" {
		return nil, "", "", false
	}
	callNode := g.diagNodeName(ast.NewFunctionCall(ast.NewIdentifier("set"), nil, nil, false), "*ast.FunctionCall", "call")
	if method, ok := g.nativeInterfaceMethodForGoType(receiverType, "set"); ok && method != nil && len(method.ParamGoTypes) == 2 {
		idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, target.Index, method.ParamGoTypes[0])
		if !ok {
			return nil, "", "", false
		}
		_ = idxType
		valueLines, coercedValueExpr, coercedValueType, ok := g.prepareStaticCallArg(ctx, valueExpr, valueType, method.ParamGoTypes[1])
		if !ok {
			return nil, "", "", false
		}
		callExpr := fmt.Sprintf("%s.%s(%s, %s)", receiverExpr, method.GoName, idxExpr, coercedValueExpr)
		resultTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := append([]string{}, idxLines...)
		lines = append(lines, valueLines...)
		lines = append(lines,
			fmt.Sprintf("__able_push_call_frame(%s)", callNode),
			fmt.Sprintf("%s, %s := %s", resultTemp, controlTemp, callExpr),
			"__able_pop_call_frame()",
		)
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		return g.finalizeStaticIndexSetResult(ctx, lines, resultTemp, method.ReturnGoType, coercedValueExpr, coercedValueType)
	}
	info := g.compileableInterfaceMethodForReceiverArity(receiverType, "IndexMut", "set", 2)
	if info == nil || len(info.Params) < 3 {
		return nil, "", "", false
	}
	idxLines, idxExpr, idxType, ok := g.compileExprLines(ctx, target.Index, info.Params[1].GoType)
	if !ok {
		return nil, "", "", false
	}
	_ = idxType
	valueLines, coercedValueExpr, coercedValueType, ok := g.prepareStaticCallArg(ctx, valueExpr, valueType, info.Params[2].GoType)
	if !ok {
		return nil, "", "", false
	}
	callExpr := fmt.Sprintf("__able_compiled_%s(%s, %s, %s)", info.GoName, receiverExpr, idxExpr, coercedValueExpr)
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines := append([]string{}, idxLines...)
	lines = append(lines, valueLines...)
	lines = append(lines,
		fmt.Sprintf("__able_push_call_frame(%s)", callNode),
		fmt.Sprintf("%s, %s := %s", resultTemp, controlTemp, callExpr),
		"__able_pop_call_frame()",
	)
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	return g.finalizeStaticIndexSetResult(ctx, lines, resultTemp, info.ReturnType, coercedValueExpr, coercedValueType)
}
