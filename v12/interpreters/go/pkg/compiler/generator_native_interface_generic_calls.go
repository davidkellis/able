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
	actualReceiverType := receiverType
	var actualExpr ast.TypeExpression
	var info *nativeInterfaceInfo
	if member, ok := call.Callee.(*ast.MemberAccessExpression); ok && member != nil && member.Object != nil {
		if inferredExpr, ok := g.inferExpressionTypeExpr(ctx, member.Object, receiverType); ok && inferredExpr != nil {
			actualExpr = g.lowerNormalizedTypeExpr(ctx, inferredExpr)
			if actualInfo, ok := g.ensureNativeInterfaceInfo(ctx.packageName, actualExpr); ok && actualInfo != nil && actualInfo.GoType != "" {
				info = actualInfo
				receiverType = actualInfo.GoType
			}
		}
	}
	var method *nativeInterfaceGenericMethod
	if info == nil {
		info = g.nativeInterfaceInfoForGoType(receiverType)
	}
	if info != nil && info.GoType != "" {
		receiverType = info.GoType
		method, _ = g.nativeInterfaceGenericMethodForGoType(receiverType, methodName)
	}
	if method == nil {
		if concreteInfo, concreteMethod, ok := g.nativeInterfaceGenericMethodForConcreteReceiver(actualReceiverType, actualExpr, methodName); ok {
			info = concreteInfo
			method = concreteMethod
			if concreteInfo != nil && concreteInfo.GoType != "" {
				receiverType = concreteInfo.GoType
			} else {
				receiverType = actualReceiverType
			}
		}
	}
	if method == nil {
		return nil, "", "", false
	}
	shapeReceiverType := receiverType
	defaultReceiverType := receiverType
	if actualReceiverType != "" && g.nativeInterfaceInfoForGoType(actualReceiverType) == nil {
		shapeReceiverType = actualReceiverType
		defaultReceiverType = actualReceiverType
	}
	paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings, ok := g.inferNativeInterfaceGenericMethodShape(ctx, call, shapeReceiverType, method, expected)
	if !ok {
		return nil, "", "", false
	}
	if directLines, expr, retType, ok := g.compileConcreteNativeInterfaceGenericMethodCall(ctx, call, expected, receiverExpr, actualReceiverType, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings, callNode); ok {
		return directLines, expr, retType, true
	}
	if directLines, expr, retType, ok := g.compileStaticNativeInterfaceGenericDefaultMethodCall(ctx, call, expected, receiverExpr, defaultReceiverType, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings, callNode); ok {
		return directLines, expr, retType, true
	}
	if info == nil && method != nil {
		if ifaceInfo, ok := g.ensureNativeInterfaceInfo(method.InterfacePackage, nativeInterfaceInstantiationExpr(method.InterfaceName, method.InterfaceArgs)); ok && ifaceInfo != nil && ifaceInfo.GoType != "" {
			info = ifaceInfo
		}
	}
	if dispatchInfo, ok := g.ensureNativeInterfaceGenericDispatchInfo(ctx, call, expected, info, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings); ok && dispatchInfo != nil {
		lines := make([]string, 0, len(call.Arguments)*4+8)
		args := make([]string, 0, len(call.Arguments))
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
			args = append(args, argTemp)
		}
		resultTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		callArgs := append([]string{receiverExpr}, args...)
		callArgs = append(callArgs, callNode)
		lines = append(lines, fmt.Sprintf("%s, %s := __able_compiled_%s(%s)", resultTemp, controlTemp, dispatchInfo.GoName, strings.Join(callArgs, ", ")))
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		return g.finishNativeInterfaceGenericCallReturn(ctx, lines, resultTemp, returnGoType, expected)
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
	controlLines, ok := g.lowerControlCheck(ctx, receiverControlTemp)
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
		argValueLines, argValueExpr, ok := g.lowerRuntimeValue(ctx, argTemp, exprType)
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
		writebackLines, ok := g.lowerControlCheck(ctx, writebackControl)
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
		writebackLines, ok := g.lowerControlCheck(ctx, writebackControl)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, writebackLines...)
	}
	controlLines, ok = g.lowerControlCheck(ctx, controlTemp)
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
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, returnGoType)
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

func (g *generator) nativeInterfaceGenericMethodForConcreteReceiver(receiverType string, actualExpr ast.TypeExpression, methodName string) (*nativeInterfaceInfo, *nativeInterfaceGenericMethod, bool) {
	if g == nil || receiverType == "" || methodName == "" {
		return nil, nil, false
	}
	if actualExpr == nil {
		if inferred, ok := g.typeExprForGoType(receiverType); ok && inferred != nil {
			actualExpr = inferred
		}
	}
	if actualExpr != nil {
		actualExpr = normalizeTypeExprForPackage(g, "", actualExpr)
	}
	if actualExpr != nil {
		for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
			impl := candidateInfo.impl
			info := candidateInfo.info
			if impl == nil || info == nil || impl.ImplName != "" || impl.TargetType == nil {
				continue
			}
			bindings := cloneTypeBindings(info.TypeBindings)
			if bindings == nil {
				bindings = make(map[string]ast.TypeExpression)
			}
			if iface := g.interfaces[impl.InterfaceName]; iface != nil {
				if !g.mergeConcreteBindings(impl.Info.Package, bindings, g.interfaceSelfTypeBindings(iface, actualExpr)) {
					continue
				}
			}
			genericNames := mergeGenericNameSets(nativeInterfaceGenericNameSet(impl.InterfaceGenerics), g.callableGenericNames(info))
			targetTemplate := g.specializedImplTargetTemplate(impl, bindings)
			if targetTemplate == nil {
				targetTemplate = impl.TargetType
			}
			if !g.specializedTypeTemplateMatches(impl.Info.Package, targetTemplate, actualExpr, genericNames, bindings, make(map[string]struct{})) {
				continue
			}
			ifaceExpr := g.implConcreteInterfaceExpr(impl, bindings)
			if ifaceExpr == nil {
				continue
			}
			if concreteInfo, ok := g.ensureNativeInterfaceInfo(impl.Info.Package, ifaceExpr); ok && concreteInfo != nil {
				for _, candidate := range concreteInfo.GenericMethods {
					if candidate != nil && candidate.Name == methodName {
						return concreteInfo, candidate, true
					}
				}
			}
			methods := make(map[string]*nativeInterfaceGenericMethod)
			if !g.collectNativeInterfaceGenericMethods(impl.Info.Package, ifaceExpr, make(map[string]struct{}), methods) {
				continue
			}
			if candidate := methods[methodName]; candidate != nil {
				return nil, candidate, true
			}
		}
	}
	for _, key := range g.sortedNativeInterfaceKeys() {
		info := g.nativeInterfaces[key]
		if info == nil || len(info.GenericMethods) == 0 {
			continue
		}
		var method *nativeInterfaceGenericMethod
		for _, candidate := range info.GenericMethods {
			if candidate != nil && candidate.Name == methodName {
				method = candidate
				break
			}
		}
		if method == nil {
			continue
		}
		if _, ok := g.nativeInterfaceAdapterForActual(info, receiverType); ok {
			return info, method, true
		}
		if actualExpr != nil && g.nativeInterfaceConcreteActualMatches(info, receiverType) {
			return info, method, true
		}
	}
	return nil, nil, false
}

func (g *generator) compileConcreteNativeInterfaceGenericMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, method *nativeInterfaceGenericMethod, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string, bindings map[string]ast.TypeExpression, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || receiverExpr == "" || receiverType == "" || method == nil {
		return nil, "", "", false
	}
	if g.nativeInterfaceInfoForGoType(receiverType) != nil {
		return nil, "", "", false
	}
	impl := g.nativeInterfaceSpecializedGenericMethodImpl(ctx, call, expected, receiverType, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType)
	if impl == nil || impl.Info == nil || impl.Info.GoName == "" {
		impl = g.concreteNativeInterfaceGenericMethodImpl(receiverType, method, bindings)
		if impl == nil || impl.Info == nil || impl.Info.GoName == "" {
			return nil, "", "", false
		}
	}
	lines := make([]string, 0, len(call.Arguments)*4+8)
	args := make([]string, 0, len(call.Arguments)+1)
	receiverLines, coercedReceiverExpr, _, ok := g.prepareConcreteNativeInterfaceReceiver(ctx, receiverExpr, receiverType, impl)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, receiverLines...)
	args = append(args, coercedReceiverExpr)
	for idx, arg := range call.Arguments {
		expectedType := paramGoTypes[idx]
		if idx < len(impl.CompiledParamGoTypes) && impl.CompiledParamGoTypes[idx] != "" {
			expectedType = impl.CompiledParamGoTypes[idx]
		}
		expectedTypeExpr := paramTypeExprs[idx]
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

func (g *generator) concreteNativeInterfaceGenericMethodImpl(receiverType string, method *nativeInterfaceGenericMethod, bindings map[string]ast.TypeExpression) *nativeInterfaceAdapterMethod {
	if g == nil || receiverType == "" || method == nil {
		return nil
	}
	var found *nativeInterfaceAdapterMethod
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || impl.ImplName != "" || impl.MethodName != method.Name || impl.InterfaceName != method.InterfaceName {
			continue
		}
		merged := cloneTypeBindings(bindings)
		if merged == nil {
			merged = make(map[string]ast.TypeExpression)
		}
		if methodBindings, ok := g.nativeInterfaceGenericImplBindings(impl, method); ok {
			if !g.mergeConcreteBindings(impl.Info.Package, merged, methodBindings) {
				continue
			}
		}
		if iface := g.interfaces[impl.InterfaceName]; iface != nil {
			if concreteTarget := g.specializedImplTargetTemplate(impl, merged); concreteTarget != nil {
				if !g.mergeConcreteBindings(impl.Info.Package, merged, g.interfaceSelfTypeBindings(iface, concreteTarget)) {
					continue
				}
			}
		}
		if concreteInfo, concreteBindings, ok := g.nativeInterfaceConcreteImplInfo(receiverType, impl, merged); ok && concreteInfo != nil {
			merged = concreteBindings
			if merged, ok = g.nativeInterfaceMergeConcreteInfoBindings(concreteInfo, merged); !ok {
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
			specialized, ok := g.ensureSpecializedImplMethod(methodInfo, impl, merged)
			if !ok || specialized == nil || specialized.Info == nil {
				continue
			}
			candidate := &nativeInterfaceAdapterMethod{
				Info:                 specialized.Info,
				CompiledReturnGoType: specialized.Info.ReturnType,
			}
			if len(specialized.Info.Params) > 1 {
				candidate.CompiledParamGoTypes = make([]string, 0, len(specialized.Info.Params)-1)
				for idx := 1; idx < len(specialized.Info.Params); idx++ {
					candidate.CompiledParamGoTypes = append(candidate.CompiledParamGoTypes, specialized.Info.Params[idx].GoType)
				}
			}
			if found != nil && found.Info != candidate.Info {
				if equivalentFunctionInfoSignature(found.Info, candidate.Info) {
					if g.nativeInterfaceAdapterMethodSpecificity(candidate) >= g.nativeInterfaceAdapterMethodSpecificity(found) {
						found = candidate
					}
					continue
				}
				foundScore := g.nativeInterfaceAdapterMethodSpecificity(found)
				candidateScore := g.nativeInterfaceAdapterMethodSpecificity(candidate)
				switch {
				case candidateScore > foundScore:
					found = candidate
				case candidateScore < foundScore:
					continue
				default:
					return nil
				}
			} else {
				found = candidate
			}
		}
	}
	return found
}
