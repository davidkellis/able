package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type concreteNativeInterfaceMethodCall struct {
	method *nativeInterfaceMethod
	impl   *nativeInterfaceAdapterMethod
}

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
	if method.DefaultDefinition != nil && !g.nativeInterfaceReceiverUsesConcreteAdapterWrapper(receiverType, receiverExpr) {
		defaultMethod := &nativeInterfaceGenericMethod{
			Name:              method.Name,
			GoName:            method.GoName,
			InterfaceName:     method.InterfaceName,
			InterfacePackage:  method.InterfacePackage,
			InterfaceArgs:     method.InterfaceArgs,
			ParamTypeExprs:    method.ParamTypeExprs,
			ReturnTypeExpr:    method.ReturnTypeExpr,
			DefaultDefinition: method.DefaultDefinition,
		}
		if directLines, expr, retType, ok := g.compileStaticNativeInterfaceGenericDefaultMethodCall(
			ctx,
			call,
			expected,
			receiverExpr,
			receiverType,
			defaultMethod,
			method.ParamTypeExprs,
			method.ParamGoTypes,
			method.ReturnTypeExpr,
			method.ReturnGoType,
			nil,
			callNode,
		); ok {
			return directLines, expr, retType, true
		}
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
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	if expected == "" || g.typeMatches(expected, method.ReturnGoType) {
		return lines, resultTemp, method.ReturnGoType, true
	}
	if expected != "runtime.Value" && method.ReturnGoType == "runtime.Value" {
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if expected == "runtime.Value" && method.ReturnGoType != "runtime.Value" {
		convLines, converted, ok := g.lowerRuntimeValue(ctx, resultTemp, method.ReturnGoType)
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
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, anyTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" && g.canCoerceStaticExpr(expected, method.ReturnGoType) {
		return g.lowerCoerceExpectedStaticExpr(ctx, lines, resultTemp, method.ReturnGoType, expected)
	}
	ctx.setReason("call return type mismatch")
	return nil, "", "", false
}

func (g *generator) nativeInterfaceReceiverUsesConcreteAdapterWrapper(receiverType string, receiverExpr string) bool {
	if g == nil || receiverType == "" || receiverExpr == "" {
		return false
	}
	info := g.nativeInterfaceInfoForGoType(receiverType)
	if info == nil || len(info.Adapters) == 0 {
		return false
	}
	trimmed := strings.TrimSpace(receiverExpr)
	for _, adapter := range info.Adapters {
		if adapter == nil || adapter.WrapHelper == "" {
			continue
		}
		if strings.HasPrefix(trimmed, adapter.WrapHelper+"(") {
			return true
		}
	}
	return false
}

func (g *generator) compileConcreteNativeInterfaceMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, methodName string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || receiverExpr == "" || receiverType == "" || methodName == "" {
		return nil, "", "", false
	}
	if g.nativeInterfaceInfoForGoType(receiverType) != nil {
		return nil, "", "", false
	}
	candidate, ok := g.concreteNativeInterfaceMethodForReceiver(receiverType, methodName, len(call.Arguments))
	if !ok || candidate == nil || candidate.method == nil || candidate.impl == nil || candidate.impl.Info == nil {
		return nil, "", "", false
	}
	method := candidate.method
	impl := candidate.impl
	lines := make([]string, 0, len(call.Arguments)*4+8)
	args := make([]string, 0, len(call.Arguments)+1)
	receiverLines, coercedReceiverExpr, _, ok := g.prepareConcreteNativeInterfaceReceiver(ctx, receiverExpr, receiverType, impl)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, receiverLines...)
	args = append(args, coercedReceiverExpr)
	for idx, arg := range call.Arguments {
		expectedType := method.ParamGoTypes[idx]
		if idx < len(impl.CompiledParamGoTypes) && impl.CompiledParamGoTypes[idx] != "" {
			expectedType = impl.CompiledParamGoTypes[idx]
		}
		var expectedTypeExpr ast.TypeExpression
		if idx < len(method.ParamTypeExprs) {
			expectedTypeExpr = method.ParamTypeExprs[idx]
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
	lines = append(lines, fmt.Sprintf("__able_push_call_frame(%s)", callNode))
	lines = append(lines, fmt.Sprintf("%s, %s := __able_compiled_%s(%s)", resultTemp, controlTemp, impl.Info.GoName, strings.Join(args, ", ")))
	lines = append(lines, "__able_pop_call_frame()")
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	resultType := impl.CompiledReturnGoType
	if resultType == "" {
		resultType = impl.Info.ReturnType
	}
	return g.finishNativeInterfaceGenericCallReturn(ctx, lines, resultTemp, resultType, expected)
}

func (g *generator) concreteNativeInterfaceMethodForReceiver(receiverType string, methodName string, argCount int) (*concreteNativeInterfaceMethodCall, bool) {
	if g == nil || receiverType == "" || methodName == "" {
		return nil, false
	}
	var found *concreteNativeInterfaceMethodCall
	considerCandidate := func(method *nativeInterfaceMethod, impl *nativeInterfaceAdapterMethod) bool {
		if method == nil || impl == nil || impl.Info == nil {
			return false
		}
		candidate := &concreteNativeInterfaceMethodCall{method: method, impl: impl}
		if found == nil {
			found = candidate
			return true
		}
		if found.impl.Info == candidate.impl.Info {
			return true
		}
		if equivalentFunctionInfoSignature(found.impl.Info, candidate.impl.Info) {
			if g.nativeInterfaceAdapterMethodSpecificity(candidate.impl) >= g.nativeInterfaceAdapterMethodSpecificity(found.impl) {
				found = candidate
			}
			return true
		}
		foundScore := g.nativeInterfaceAdapterMethodSpecificity(found.impl)
		candidateScore := g.nativeInterfaceAdapterMethodSpecificity(candidate.impl)
		switch {
		case candidateScore > foundScore:
			found = candidate
			return true
		case candidateScore < foundScore:
			return true
		default:
			return false
		}
	}
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || impl.ImplName != "" || impl.MethodName != methodName {
			continue
		}
		concreteInfo, bindings, ok := g.nativeInterfaceConcreteImplInfo(receiverType, impl, info.TypeBindings)
		if !ok || concreteInfo == nil {
			continue
		}
		bindings, ok = g.nativeInterfaceMergeConcreteInfoBindings(concreteInfo, bindings)
		if !ok {
			continue
		}
		paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, optionalLast, ok := g.nativeInterfaceMethodImplSignature(impl, bindings)
		if !ok {
			continue
		}
		paramCount := len(paramGoTypes)
		if argCount != paramCount && !(optionalLast && argCount == paramCount-1) {
			continue
		}
		targetName, _ := typeExprBaseName(impl.TargetType)
		methodInfo := &methodInfo{
			TargetName:   targetName,
			TargetType:   impl.TargetType,
			MethodName:   impl.MethodName,
			ReceiverType: concreteInfo.Params[0].GoType,
			ExpectsSelf:  len(concreteInfo.Params) > 0,
			Info:         concreteInfo,
		}
		if specialized, ok := g.ensureSpecializedImplMethod(methodInfo, impl, bindings); ok && specialized != nil && specialized.Info != nil {
			methodInfo = specialized
			concreteInfo = specialized.Info
		}
		method := &nativeInterfaceMethod{
			Name:             impl.MethodName,
			GoName:           sanitizeIdent(impl.MethodName),
			InterfaceName:    impl.InterfaceName,
			InterfacePackage: impl.Info.Package,
			InterfaceArgs:    cloneTypeExprSlice(impl.InterfaceArgs),
			ParamGoTypes:     paramGoTypes,
			ParamTypeExprs:   paramTypeExprs,
			ReturnGoType:     returnGoType,
			ReturnTypeExpr:   returnTypeExpr,
			OptionalLast:     optionalLast,
		}
		adapterMethod := &nativeInterfaceAdapterMethod{
			Info:                 concreteInfo,
			ParamGoTypes:         paramGoTypes,
			ReturnGoType:         returnGoType,
			CompiledReturnGoType: concreteInfo.ReturnType,
		}
		if len(concreteInfo.Params) > 1 {
			adapterMethod.CompiledParamGoTypes = make([]string, 0, len(concreteInfo.Params)-1)
			for idx := 1; idx < len(concreteInfo.Params); idx++ {
				adapterMethod.CompiledParamGoTypes = append(adapterMethod.CompiledParamGoTypes, concreteInfo.Params[idx].GoType)
			}
		}
		if !considerCandidate(method, adapterMethod) {
			return nil, false
		}
	}
	return found, found != nil
}
