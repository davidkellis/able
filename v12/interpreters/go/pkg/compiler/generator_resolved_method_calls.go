package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) resolveStaticMethodCall(ctx *compileContext, object ast.Expression, memberName string) (*methodInfo, bool) {
	if g == nil || object == nil || memberName == "" {
		return nil, false
	}
	ident, ok := object.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return nil, false
	}
	if ctx != nil {
		if _, ok := ctx.lookup(ident.Name); ok {
			return nil, false
		}
	}
	targetName := ident.Name
	targetTypeExpr := ast.TypeExpression(ast.Ty(targetName))
	if ctx != nil && len(ctx.typeBindings) > 0 {
		if bound, ok := ctx.typeBindings[ident.Name]; ok && bound != nil {
			targetTypeExpr = bound
			if boundName, ok := typeExprBaseName(bound); ok && boundName != "" {
				targetName = boundName
			}
		}
	}
	info, ok := g.structInfoForTypeName(ctx.packageName, targetName)
	if (!ok || info == nil) && g != nil {
		info, _ = g.structInfoByNameUnique(targetName)
	}
	method := g.methodForTypeName(targetName, memberName, false)
	if method == nil && info != nil {
		if typeBucket := g.methods[targetName]; len(typeBucket) > 0 {
			entries := typeBucket[memberName]
			var found *methodInfo
			for _, candidate := range entries {
				if candidate == nil || candidate.Info == nil || !candidate.Info.Compileable || candidate.ExpectsSelf {
					continue
				}
				if found != nil && found.Info != candidate.Info {
					if equivalentFunctionInfoSignature(found.Info, candidate.Info) {
						found = candidate
						continue
					}
					found = nil
					break
				}
				found = candidate
			}
			method = found
		}
	}
	if method == nil && info != nil {
		method = g.methodForStruct(info, memberName, false)
	}
	if method == nil {
		var found *methodInfo
		for _, candidate := range g.methodList {
			if candidate == nil || candidate.Info == nil || !candidate.Info.Compileable || candidate.ExpectsSelf {
				continue
			}
			if candidate.TargetName != targetName || candidate.MethodName != memberName {
				continue
			}
			if found != nil && found.Info != candidate.Info {
				if equivalentFunctionInfoSignature(found.Info, candidate.Info) {
					found = candidate
					continue
				}
				found = nil
				break
			}
			found = candidate
		}
		method = found
	}
	if method == nil {
		method = g.compileableInterfaceStaticMethodForConcreteTarget(targetTypeExpr, memberName)
	}
	if method == nil {
		return nil, false
	}
	return method, true
}

func (g *generator) staticMethodCallCandidates(ctx *compileContext, object ast.Expression, memberName string) []*methodInfo {
	if g == nil || object == nil || memberName == "" {
		return nil
	}
	ident, ok := object.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return nil
	}
	if ctx != nil {
		if _, ok := ctx.lookup(ident.Name); ok {
			return nil
		}
	}
	targetName := ident.Name
	targetTypeExpr := ast.TypeExpression(ast.Ty(targetName))
	if ctx != nil && len(ctx.typeBindings) > 0 {
		if bound, ok := ctx.typeBindings[ident.Name]; ok && bound != nil {
			targetTypeExpr = bound
			if boundName, ok := typeExprBaseName(bound); ok && boundName != "" {
				targetName = boundName
			}
		}
	}
	info, ok := g.structInfoForTypeName(ctx.packageName, targetName)
	if (!ok || info == nil) && g != nil {
		info, _ = g.structInfoByNameUnique(targetName)
	}
	var candidates []*methodInfo
	appendCandidate := func(candidate *methodInfo) {
		if candidate == nil || candidate.Info == nil || !candidate.Info.Compileable || candidate.ExpectsSelf {
			return
		}
		for _, existing := range candidates {
			if existing == candidate || (existing.Info == candidate.Info && existing.MethodName == candidate.MethodName) {
				return
			}
		}
		candidates = append(candidates, candidate)
	}
	if typeBucket := g.methods[targetName]; len(typeBucket) > 0 {
		for _, candidate := range typeBucket[memberName] {
			if candidate == nil || candidate.Info == nil || !candidate.Info.Compileable || candidate.ExpectsSelf {
				continue
			}
			if info != nil && candidate.Info.Package != info.Package {
				continue
			}
			appendCandidate(candidate)
		}
	}
	if len(candidates) == 0 {
		for _, candidate := range g.methodList {
			if candidate == nil || candidate.Info == nil || !candidate.Info.Compileable || candidate.ExpectsSelf {
				continue
			}
			if candidate.TargetName != targetName || candidate.MethodName != memberName {
				continue
			}
			if info != nil && candidate.Info.Package != info.Package {
				continue
			}
			appendCandidate(candidate)
		}
	}
	if len(candidates) == 0 {
		if method := g.compileableInterfaceStaticMethodForConcreteTarget(targetTypeExpr, memberName); method != nil {
			appendCandidate(method)
		}
	}
	return candidates
}

func staticMethodCallArityCompatible(call *ast.FunctionCall, method *methodInfo) bool {
	if call == nil || method == nil || method.Info == nil {
		return false
	}
	callArgCount := len(call.Arguments)
	paramCount := len(method.Info.Params)
	optionalLast := hasOptionalLastMethodParam(method)
	if callArgCount == paramCount {
		return true
	}
	return optionalLast && callArgCount == paramCount-1
}

func (g *generator) resolveStaticMethodCallForCall(ctx *compileContext, call *ast.FunctionCall, object ast.Expression, memberName string) (*methodInfo, bool) {
	if g == nil || call == nil {
		return nil, false
	}
	if method, ok := g.resolveStaticMethodCall(ctx, object, memberName); ok && method != nil {
		return method, true
	}
	candidates := g.staticMethodCallCandidates(ctx, object, memberName)
	if len(candidates) == 0 {
		return nil, false
	}
	compatible := make([]*methodInfo, 0, len(candidates))
	for _, candidate := range candidates {
		if !staticMethodCallArityCompatible(call, candidate) {
			continue
		}
		compatible = append(compatible, candidate)
	}
	if len(compatible) == 1 {
		return compatible[0], true
	}
	if len(compatible) == 0 {
		return nil, false
	}
	found := compatible[0]
	for _, candidate := range compatible[1:] {
		if !equivalentFunctionInfoSignature(found.Info, candidate.Info) {
			return nil, false
		}
	}
	return found, true
}

func hasOptionalLastMethodParam(method *methodInfo) bool {
	if method == nil || method.Info == nil || method.Info.Definition == nil || len(method.Info.Definition.Params) == 0 {
		return false
	}
	return isNullableParam(method.Info.Definition.Params[len(method.Info.Definition.Params)-1])
}

func (g *generator) compileResolvedMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, method *methodInfo, receiverExpr string, receiverType string, callNode string) ([]string, string, string, bool) {
	if call == nil || method == nil || method.Info == nil {
		ctx.setReason("missing method call")
		return nil, "", "", false
	}
	if member, ok := call.Callee.(*ast.MemberAccessExpression); ok && member != nil {
		if method.ExpectsSelf {
			method = g.concreteMethodCallInfo(ctx, call, method, member.Object, receiverType, expected)
		} else {
			method = g.concreteStaticMethodCallInfo(ctx, call, method, member.Object, expected)
		}
	}
	info := method.Info
	if !info.Compileable {
		ctx.setReason("unsupported method call")
		return nil, "", "", false
	}
	needsIntCast := false
	needsRuntimeConv := expected != "" && expected != "runtime.Value" && info.ReturnType == "runtime.Value"
	needsAnyConv := expected != "" && expected != "any" && info.ReturnType == "any"
	needsStaticCoerce := expected != "" && expected != "runtime.Value" && expected != "any" && expected != info.ReturnType && g.canCoerceStaticExpr(expected, info.ReturnType)
	if !g.typeMatches(expected, info.ReturnType) && !needsRuntimeConv && !needsAnyConv && !needsStaticCoerce {
		if g.isIntegerType(expected) && g.isIntegerType(info.ReturnType) {
			needsIntCast = true
		} else {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
	}
	paramOffset := 0
	args := make([]string, 0, len(call.Arguments)+1)
	var argPreLines []string
	receiverWritebackNeeded := false
	receiverWritebackTargetExpr := ""
	receiverWritebackTargetType := ""
	if method.ExpectsSelf {
		if receiverExpr == "" {
			ctx.setReason("method receiver missing")
			return nil, "", "", false
		}
		selfType := ""
		if len(info.Params) > 0 {
			selfType = g.canonicalMethodReceiverGoType(info, receiverType)
		}
		if g.staticNominalReceiverWritebackAllowed(call) {
			if origin, ok := ctx.nominalCoercionOrigin(receiverExpr); ok {
				receiverWritebackNeeded = true
				receiverWritebackTargetExpr = origin.Expr
				receiverWritebackTargetType = origin.GoType
			} else if receiverExpr != "" &&
				receiverType != "" &&
				selfType != "" &&
				receiverType != selfType &&
				(g.sameNominalStructFamily(receiverType, selfType) || g.staticArrayCarrierCoercible(receiverType, selfType)) {
				receiverWritebackNeeded = true
				receiverWritebackTargetExpr = receiverExpr
				receiverWritebackTargetType = receiverType
			}
		}
		receiverLines, coercedReceiver, _, ok := g.prepareStaticCallArg(ctx, receiverExpr, receiverType, selfType)
		if !ok {
			ctx.setReason("method receiver type mismatch")
			return nil, "", "", false
		}
		argPreLines = append(argPreLines, receiverLines...)
		receiverExpr = coercedReceiver
		args = append(args, receiverExpr)
		paramOffset = 1
	}
	params := info.Params
	if paramOffset > len(params) {
		ctx.setReason("method params missing")
		return nil, "", "", false
	}
	callArgCount := len(call.Arguments)
	paramCount := len(params) - paramOffset
	optionalLast := g.hasOptionalLastParam(info)
	if callArgCount != paramCount {
		if !(optionalLast && callArgCount == paramCount-1) {
			ctx.setReason("call arity mismatch")
			return nil, "", "", false
		}
	}
	missingOptional := optionalLast && callArgCount == paramCount-1
	if missingOptional && paramCount > 0 {
		lastType := params[len(params)-1].GoType
		if _, ok := g.zeroValueExpr(lastType); !ok {
			ctx.setReason("call arity mismatch")
			return nil, "", "", false
		}
	}
	for idx, arg := range call.Arguments {
		param := params[paramOffset+idx]
		paramGoType := param.GoType
		expectedArgType := g.staticParamCarrierType(ctx, param)
		compileExpectedArgType := expectedArgType
		if g.nativeUnionInfoForGoType(expectedArgType) != nil {
			compileExpectedArgType = ""
		}
		if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok && (expectedArgType == "" || expectedArgType == "runtime.Value" || expectedArgType == "any") {
			if ifaceInfo, ok := g.ensureNativeInterfaceInfo(ctx.packageName, ifaceType); ok && ifaceInfo != nil && ifaceInfo.GoType != "" {
				expectedArgType = ifaceInfo.GoType
			}
		}
		argLines, expr, exprType, ok := g.compileExprLinesWithExpectedTypeExpr(ctx, arg, compileExpectedArgType, param.TypeExpr)
		if !ok {
			return nil, "", "", false
		}
		argPreLines = append(argPreLines, argLines...)
		argExpr := expr
		argType := exprType
		if paramGoType == "runtime.Value" && argType != "runtime.Value" {
			convLines, valueExpr, ok := g.lowerRuntimeValue(ctx, argExpr, argType)
			if !ok {
				ctx.setReason("call argument unsupported")
				return nil, "", "", false
			}
			argPreLines = append(argPreLines, convLines...)
			argExpr = valueExpr
			argType = "runtime.Value"
		}
		if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok && paramGoType == "runtime.Value" {
			ifaceLines, coerced, ok := g.interfaceArgExprLines(ctx, argExpr, ifaceType, info.Name, ctx.genericNames)
			if !ok {
				ctx.setReason("interface argument unsupported")
				return nil, "", "", false
			}
			argPreLines = append(argPreLines, ifaceLines...)
			argExpr = coerced
		} else if paramGoType != "" && paramGoType != "any" && argType != paramGoType {
			coerceLines, coercedExpr, coercedType, ok := g.prepareStaticCallArg(ctx, argExpr, argType, paramGoType)
			if !ok {
				ctx.setReason("call argument type mismatch")
				return nil, "", "", false
			}
			argPreLines = append(argPreLines, coerceLines...)
			argExpr = coercedExpr
			argType = coercedType
		}
		args = append(args, argExpr)
	}
	if missingOptional {
		lastType := params[len(params)-1].GoType
		zeroExpr, ok := g.zeroValueExpr(lastType)
		if !ok {
			ctx.setReason("call arity mismatch")
			return nil, "", "", false
		}
		args = append(args, zeroExpr)
	}
	callExpr := fmt.Sprintf("__able_compiled_%s(%s)", info.GoName, strings.Join(args, ", "))
	resultTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	lines := append(argPreLines, []string{
		fmt.Sprintf("__able_push_call_frame(%s)", callNode),
		fmt.Sprintf("%s, %s := %s", resultTemp, controlTemp, callExpr),
		"__able_pop_call_frame()",
	}...)
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	if receiverWritebackNeeded {
		writebackLines, ok := g.appendStaticNominalReceiverWriteback(ctx, receiverWritebackTargetExpr, receiverWritebackTargetType, receiverExpr, info.Params[0].GoType)
		if !ok {
			ctx.setReason("same-family receiver writeback failed")
			return nil, "", "", false
		}
		lines = append(lines, writebackLines...)
	}
	if needsIntCast {
		castTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s(%s)", castTemp, expected, resultTemp))
		return lines, castTemp, expected, true
	}
	if needsRuntimeConv {
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	if needsAnyConv {
		if expected == "runtime.Value" {
			convTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, resultTemp))
			return lines, convTemp, "runtime.Value", true
		}
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
	if needsStaticCoerce {
		return g.lowerCoerceExpectedStaticExpr(ctx, lines, resultTemp, info.ReturnType, expected)
	}
	return lines, resultTemp, info.ReturnType, true
}

func (g *generator) staticNominalReceiverWritebackAllowed(call *ast.FunctionCall) bool {
	if g == nil || call == nil {
		return false
	}
	member, ok := call.Callee.(*ast.MemberAccessExpression)
	if !ok || member == nil || member.Safe {
		return false
	}
	return g.isAddressableMemberObject(member.Object)
}

func (g *generator) appendStaticNominalReceiverWriteback(ctx *compileContext, actualExpr string, actualType string, coercedExpr string, coercedType string) ([]string, bool) {
	if g == nil || ctx == nil || actualExpr == "" || actualType == "" || coercedExpr == "" || coercedType == "" {
		return nil, false
	}
	lines := []string{}
	convertedExpr := coercedExpr
	switch {
	case actualType == coercedType:
		// No coercion required; write back the mutated value directly.
	case g.staticArrayCarrierCoercible(actualType, coercedType):
		coerceLines, converted, ok := g.coerceStaticArrayCarrierLines(ctx, coercedExpr, coercedType, actualType)
		if !ok {
			return nil, false
		}
		lines = append(lines, coerceLines...)
		convertedExpr = converted
	case g.sameNominalStructFamily(actualType, coercedType):
		coerceLines, converted, ok := g.coerceNominalStructFamilyLines(ctx, coercedExpr, coercedType, actualType)
		if !ok {
			return nil, false
		}
		lines = append(lines, coerceLines...)
		convertedExpr = converted
	default:
		return nil, false
	}
	if strings.HasPrefix(actualType, "*") {
		lines = append(lines, fmt.Sprintf("*%s = *%s", actualExpr, convertedExpr))
	} else {
		lines = append(lines, fmt.Sprintf("%s = *%s", actualExpr, convertedExpr))
	}
	return lines, true
}

func (g *generator) specializeConcreteImplMethod(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, impl *implMethodInfo, receiver ast.Expression, receiverType string, expected string) (*methodInfo, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || impl == nil {
		return nil, false
	}
	receiverTypeExpr, ok := g.staticReceiverTypeExpr(ctx, receiver, receiverType)
	if !ok || receiverTypeExpr == nil {
		return nil, false
	}
	if concreteReceiverType, ok := g.lowerCarrierTypeInPackage(method.Info.Package, receiverTypeExpr); ok && concreteReceiverType != "" && concreteReceiverType != "runtime.Value" && concreteReceiverType != "any" && concreteReceiverType != receiverType {
		if !(g.sameNominalStructFamily(receiverType, concreteReceiverType) ||
			g.staticArrayCarrierCoercible(receiverType, concreteReceiverType) ||
			g.receiverGoTypeCompatible(concreteReceiverType, receiverType) ||
			g.receiverGoTypeCompatible(receiverType, concreteReceiverType)) {
			return nil, false
		}
	}
	genericNames := g.implSpecializationGenericNames(&methodInfo{
		TargetType:  impl.TargetType,
		MethodName:  method.MethodName,
		ExpectsSelf: method.ExpectsSelf,
		Info:        method.Info,
	})
	if g.typeExprHasGeneric(receiverTypeExpr, genericNames) {
		return nil, false
	}
	bindings, ok := g.specializedImplMethodBindings(ctx, call, method, impl, receiverTypeExpr, expected)
	if !ok || len(bindings) == 0 {
		return nil, false
	}
	return g.ensureSpecializedImplMethod(method, impl, bindings)
}

func (g *generator) specializeConcreteStaticImplMethod(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, impl *implMethodInfo, target ast.Expression, expected string) (*methodInfo, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || impl == nil {
		return nil, false
	}
	targetTypeExpr, ok := g.staticTargetTypeExpr(ctx, target)
	if !ok || targetTypeExpr == nil {
		return nil, false
	}
	targetTypeExpr = g.refineStaticTargetTypeExprWithExpected(ctx, target, method.Info.Package, targetTypeExpr, expected)
	genericNames := g.implSpecializationGenericNames(&methodInfo{
		TargetType:  impl.TargetType,
		MethodName:  method.MethodName,
		ExpectsSelf: method.ExpectsSelf,
		Info:        method.Info,
	})
	if g.typeExprHasGeneric(targetTypeExpr, genericNames) {
		return nil, false
	}
	bindings, ok := g.specializedStaticImplMethodBindings(ctx, call, method, impl, targetTypeExpr, expected)
	if !ok || len(bindings) == 0 {
		return nil, false
	}
	specialized, ok := g.ensureSpecializedImplMethod(method, impl, bindings)
	return specialized, ok
}
