package compiler

import (
	"fmt"
	"reflect"
	"sort"
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

func (g *generator) compileResolvedMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, method *methodInfo, receiverExpr string, receiverType string, callNode string) ([]string, string, string, bool) {
	if call == nil || method == nil || method.Info == nil {
		ctx.setReason("missing method call")
		return nil, "", "", false
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
	if method.ExpectsSelf {
		if receiverExpr == "" {
			ctx.setReason("method receiver missing")
			return nil, "", "", false
		}
		selfType := ""
		if len(info.Params) > 0 {
			selfType = info.Params[0].GoType
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
		if lastType != "runtime.Value" && lastType != "any" {
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
		if lastType == "any" {
			args = append(args, "nil")
		} else {
			args = append(args, "runtime.NilValue{}")
		}
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

func (g *generator) specializeConcreteImplMethod(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, impl *implMethodInfo, receiver ast.Expression, receiverType string, expected string) (*methodInfo, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || impl == nil {
		return nil, false
	}
	receiverTypeExpr, ok := g.staticReceiverTypeExpr(ctx, receiver, receiverType)
	if !ok || receiverTypeExpr == nil {
		return nil, false
	}
	genericNames := g.callableGenericNames(method.Info)
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
	genericNames := g.callableGenericNames(method.Info)
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

func (g *generator) ensureSpecializedImplMethod(method *methodInfo, impl *implMethodInfo, bindings map[string]ast.TypeExpression) (*methodInfo, bool) {
	if g == nil || method == nil || method.Info == nil || impl == nil || len(bindings) == 0 {
		return nil, false
	}
	key := g.specializedImplFunctionKey(method.Info, bindings)
	if existing, ok := g.specializedFunctionIndex[key]; ok && existing != nil && (existing.Compileable || existing == method.Info) {
		receiverType := method.ReceiverType
		if method.ExpectsSelf && len(existing.Params) > 0 {
			receiverType = existing.Params[0].GoType
		}
		return &methodInfo{
			TargetName:   method.TargetName,
			TargetType:   g.specializedImplTargetType(impl, bindings),
			MethodName:   method.MethodName,
			ReceiverType: receiverType,
			ExpectsSelf:  method.ExpectsSelf,
			Info:         existing,
		}, true
	}
	specialized := &functionInfo{
		Name:           method.Info.Name,
		Package:        method.Info.Package,
		QualifiedName:  method.Info.QualifiedName,
		GoName:         g.mangler.unique(method.Info.GoName + "_spec"),
		TypeBindings:   cloneTypeBindings(bindings),
		Definition:     method.Info.Definition,
		HasOriginal:    method.Info.HasOriginal,
		InternalOnly:   true,
		SupportedTypes: method.Info.SupportedTypes,
	}
	mapper := NewTypeMapper(g, specialized.Package)
	concreteTarget := g.specializedImplTargetType(impl, bindings)
	fillBindings := g.implTypeBindings(impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, concreteTarget)
	if fillBindings == nil {
		fillBindings = make(map[string]ast.TypeExpression, len(bindings))
	}
	for name, expr := range bindings {
		if expr == nil {
			continue
		}
		fillBindings[name] = normalizeTypeExprForPackage(g, specialized.Package, expr)
	}
	specialized.TypeBindings = cloneTypeBindings(fillBindings)
	g.fillImplMethodInfo(specialized, mapper, concreteTarget, fillBindings)
	if !specialized.SupportedTypes {
		return nil, false
	}
	specialized.Compileable = true
	g.implMethodByInfo[specialized] = impl
	g.specializedFunctions = append(g.specializedFunctions, specialized)
	g.touchNativeInterfaceAdapters()
	g.specializedFunctionIndex[key] = specialized
	if g.bodyCompileable(specialized, specialized.ReturnType) {
		specialized.Compileable = true
		specialized.Reason = ""
	}
	receiverType := method.ReceiverType
	if method.ExpectsSelf && len(specialized.Params) > 0 {
		receiverType = specialized.Params[0].GoType
	}
	return &methodInfo{
		TargetName:   method.TargetName,
		TargetType:   concreteTarget,
		MethodName:   method.MethodName,
		ReceiverType: receiverType,
		ExpectsSelf:  method.ExpectsSelf,
		Info:         specialized,
	}, true
}

func (g *generator) specializedImplTargetType(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || impl == nil || impl.TargetType == nil {
		return nil
	}
	target := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(impl.TargetType, bindings))
	if target == nil {
		target = impl.TargetType
	}
	if _, ok := target.(*ast.GenericTypeExpression); ok {
		return target
	}
	baseName, ok := typeExprBaseName(target)
	if !ok || baseName == "" {
		return target
	}
	if baseName == "Array" && len(impl.InterfaceArgs) == 1 {
		arg := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(impl.InterfaceArgs[0], bindings))
		if arg != nil {
			return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), []ast.TypeExpression{arg}))
		}
	}
	info, ok := g.structInfoForTypeName(impl.Info.Package, baseName)
	if !ok || info == nil || info.Node == nil || len(info.Node.GenericParams) == 0 {
		return target
	}
	if len(impl.InterfaceArgs) == len(info.Node.GenericParams) {
		args := make([]ast.TypeExpression, 0, len(impl.InterfaceArgs))
		for _, arg := range impl.InterfaceArgs {
			if arg == nil {
				return target
			}
			concreteArg := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(arg, bindings))
			if concreteArg == nil {
				return target
			}
			args = append(args, concreteArg)
		}
		return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
	}
	args := make([]ast.TypeExpression, 0, len(info.Node.GenericParams))
	for _, gp := range info.Node.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			return target
		}
		bound, ok := bindings[gp.Name.Name]
		if !ok || bound == nil {
			return target
		}
		args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, bound))
	}
	return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
}

func (g *generator) specializedImplTargetTemplate(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || impl == nil || impl.TargetType == nil {
		return nil
	}
	target := normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(impl.TargetType, bindings))
	if target == nil {
		target = impl.TargetType
	}
	if _, ok := target.(*ast.GenericTypeExpression); ok {
		return target
	}
	baseName, ok := typeExprBaseName(target)
	if !ok || baseName == "" {
		return target
	}
	info, ok := g.structInfoForTypeName(impl.Info.Package, baseName)
	if !ok || info == nil || info.Node == nil || len(info.Node.GenericParams) == 0 {
		return target
	}
	if len(impl.InterfaceArgs) == len(info.Node.GenericParams) {
		args := make([]ast.TypeExpression, 0, len(impl.InterfaceArgs))
		for _, arg := range impl.InterfaceArgs {
			if arg == nil {
				return target
			}
			args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(arg, bindings)))
		}
		return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
	}
	args := make([]ast.TypeExpression, 0, len(info.Node.GenericParams))
	for _, gp := range info.Node.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			return target
		}
		bound, ok := bindings[gp.Name.Name]
		if !ok || bound == nil {
			return target
		}
		args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, bound))
	}
	return normalizeTypeExprForPackage(g, impl.Info.Package, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
}

func removeSpecializedFunction(list []*functionInfo, target *functionInfo) []*functionInfo {
	for idx, info := range list {
		if info != target {
			continue
		}
		copy(list[idx:], list[idx+1:])
		list[len(list)-1] = nil
		return list[:len(list)-1]
	}
	return list
}

func (g *generator) implSiblingsForFunction(info *functionInfo) map[string]implSiblingInfo {
	if g == nil || info == nil {
		return nil
	}
	implInfo := g.implMethodByInfo[info]
	if implInfo == nil || !implInfo.IsDefault {
		return nil
	}
	siblings := g.implSiblingsForDefault(implInfo)
	if len(siblings) == 0 || len(info.TypeBindings) == 0 {
		return siblings
	}
	currentBindings := g.compileContextTypeBindings(info)
	if len(currentBindings) == 0 {
		return siblings
	}
	out := make(map[string]implSiblingInfo, len(siblings))
	for name, sibling := range siblings {
		if sibling.Info == nil {
			out[name] = sibling
			continue
		}
		siblingImpl := g.implMethodByInfo[sibling.Info]
		if siblingImpl == nil {
			out[name] = sibling
			continue
		}
		methodName := siblingImpl.MethodName
		if methodName == "" {
			methodName = name
		}
		specializedBindings := g.implSiblingBindingsForFunction(info, implInfo, siblingImpl, currentBindings)
		if len(specializedBindings) == 0 {
			out[name] = sibling
			continue
		}
		specialized, ok := g.ensureSpecializedImplMethod(&methodInfo{
			MethodName:  methodName,
			ExpectsSelf: methodDefinitionExpectsSelf(sibling.Info.Definition),
			Info:        sibling.Info,
		}, siblingImpl, specializedBindings)
		if !ok || specialized == nil || specialized.Info == nil {
			out[name] = sibling
			continue
		}
		sibling.GoName = specialized.Info.GoName
		sibling.Arity = specialized.Info.Arity
		sibling.Info = specialized.Info
		out[name] = sibling
	}
	return out
}

func (g *generator) implSiblingBindingsForFunction(info *functionInfo, currentImpl *implMethodInfo, siblingImpl *implMethodInfo, currentBindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || info == nil || currentImpl == nil || siblingImpl == nil || siblingImpl.Info == nil {
		return nil
	}
	genericNames := g.implSpecializationGenericNames(&methodInfo{
		TargetType:  siblingImpl.TargetType,
		MethodName:  siblingImpl.MethodName,
		ExpectsSelf: methodDefinitionExpectsSelf(siblingImpl.Info.Definition),
		Info:        siblingImpl.Info,
	})
	bindings := g.mergeConcreteTypeBindings(siblingImpl.Info.Package, genericNames, nil, currentBindings)
	var receiverBindings map[string]ast.TypeExpression
	actualSelfType := g.implConcreteSelfTypeExpr(info, currentImpl, currentBindings)
	if actualSelfType != nil && siblingImpl.TargetType != nil {
		targetTemplate := g.specializedImplTargetTemplate(siblingImpl, bindings)
		if targetTemplate == nil {
			targetTemplate = siblingImpl.TargetType
		}
		selfBindings := make(map[string]ast.TypeExpression)
		matched := g.specializedTypeTemplateMatches(
			siblingImpl.Info.Package,
			targetTemplate,
			actualSelfType,
			genericNames,
			selfBindings,
			make(map[string]struct{}),
		)
		if !matched {
			_ = g.specializedSameBaseGenericBindings(
				siblingImpl.Info.Package,
				targetTemplate,
				actualSelfType,
				genericNames,
				selfBindings,
			)
		}
		if len(selfBindings) > 0 {
			if bindings == nil {
				bindings = make(map[string]ast.TypeExpression, len(selfBindings))
			}
			for name, expr := range selfBindings {
				if expr == nil {
					continue
				}
				if _, ok := genericNames[name]; len(genericNames) > 0 && !ok {
					continue
				}
				if _, exists := bindings[name]; exists {
					continue
				}
				bindings[name] = normalizeTypeExprForPackage(g, siblingImpl.Info.Package, expr)
			}
		}
		receiverBindings = g.normalizeConcreteTypeBindings(siblingImpl.Info.Package, selfBindings, genericNames)
		bindings = g.mergeConcreteTypeBindings(siblingImpl.Info.Package, genericNames, bindings, receiverBindings)
	}
	actualInterfaceExpr := g.implConcreteInterfaceExpr(currentImpl, currentBindings)
	if actualInterfaceExpr != nil {
		interfaceBindings, ok := g.nativeInterfaceImplBindingsForTarget(
			currentImpl.Info.Package,
			actualInterfaceExpr,
			genericParamNameSet(siblingImpl.InterfaceGenerics),
			siblingImpl.Info.Package,
			siblingImpl.InterfaceName,
			siblingImpl.InterfaceArgs,
			make(map[string]struct{}),
		)
		if ok {
			bindings = g.mergeConcreteTypeBindings(siblingImpl.Info.Package, genericNames, bindings, interfaceBindings)
		}
	}
	normalized := g.normalizeConcreteTypeBindings(siblingImpl.Info.Package, bindings, genericNames)
	if len(normalized) == 0 && len(receiverBindings) > 0 {
		return receiverBindings
	}
	return normalized
}

func (g *generator) implSpecializationGenericNames(method *methodInfo) map[string]struct{} {
	if method == nil || method.Info == nil {
		return nil
	}
	return mergeGenericNameSets(g.callableGenericNames(method.Info), g.methodGenericNames(method))
}

func (g *generator) implConcreteSelfTypeExpr(info *functionInfo, impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || info == nil {
		return nil
	}
	if len(info.Params) > 0 && info.Params[0].TypeExpr != nil {
		return normalizeTypeExprForPackage(g, info.Package, info.Params[0].TypeExpr)
	}
	if impl == nil || impl.TargetType == nil {
		return nil
	}
	selfType := g.implSelfTargetType(impl.TargetType, bindings)
	selfType = substituteTypeParams(selfType, bindings)
	return normalizeTypeExprForPackage(g, info.Package, selfType)
}

func (g *generator) implConcreteInterfaceExpr(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || impl == nil || impl.InterfaceName == "" {
		return nil
	}
	if len(impl.InterfaceArgs) == 0 {
		return ast.Ty(impl.InterfaceName)
	}
	args := make([]ast.TypeExpression, 0, len(impl.InterfaceArgs))
	for _, arg := range impl.InterfaceArgs {
		if arg == nil {
			return nil
		}
		args = append(args, normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(arg, bindings)))
	}
	return nativeInterfaceInstantiationExpr(impl.InterfaceName, args)
}

func (g *generator) specializedImplMethodBindings(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, impl *implMethodInfo, receiverTypeExpr ast.TypeExpression, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || impl == nil {
		return nil, false
	}
	genericNames := g.implSpecializationGenericNames(method)
	bindings := g.concreteCompileContextBindings(method.Info, genericNames)
	bindings = g.mergeConcreteTypeBindings(method.Info.Package, genericNames, bindings, ctx.typeBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if method.ExpectsSelf && len(method.Info.Params) > 0 && method.Info.Params[0].TypeExpr != nil {
		receiverTemplate := method.Info.Params[0].TypeExpr
		if targetTemplate := g.specializedImplTargetTemplate(impl, bindings); targetTemplate != nil {
			receiverTemplate = targetTemplate
		}
		if !g.specializedTypeTemplateMatches(method.Info.Package, receiverTemplate, receiverTypeExpr, genericNames, bindings, make(map[string]struct{})) {
			return nil, false
		}
	}
	if len(call.TypeArguments) > 0 {
		if method.Info.Definition == nil || len(method.Info.Definition.GenericParams) != len(call.TypeArguments) {
			return nil, false
		}
		for idx, arg := range call.TypeArguments {
			if arg == nil {
				return nil, false
			}
			gp := method.Info.Definition.GenericParams[idx]
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				return nil, false
			}
			bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, method.Info.Package, arg)
		}
	}
	paramOffset := 0
	if method.ExpectsSelf {
		paramOffset = 1
	}
	for idx, arg := range call.Arguments {
		paramIdx := paramOffset + idx
		if paramIdx >= len(method.Info.Params) {
			break
		}
		paramTypeExpr := method.Info.Params[paramIdx].TypeExpr
		if paramTypeExpr == nil {
			continue
		}
		actualExpr, ok := g.inferExpressionTypeExpr(ctx, arg, "")
		if !ok || actualExpr == nil {
			continue
		}
		_ = g.specializedTypeTemplateMatches(method.Info.Package, paramTypeExpr, actualExpr, genericNames, bindings, make(map[string]struct{}))
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil && method.Info.Definition != nil && method.Info.Definition.ReturnType != nil {
		returnExpr := g.functionReturnTypeExprWithBindings(method.Info, bindings)
		if returnExpr != nil {
			_ = g.specializedTypeTemplateMatches(method.Info.Package, returnExpr, expectedExpr, genericNames, bindings, make(map[string]struct{}))
		}
	}
	bindings = g.normalizeConcreteTypeBindings(method.Info.Package, bindings, genericNames)
	if len(bindings) == 0 {
		return nil, false
	}
	return bindings, true
}

func (g *generator) specializedStaticImplMethodBindings(ctx *compileContext, call *ast.FunctionCall, method *methodInfo, impl *implMethodInfo, targetTypeExpr ast.TypeExpression, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.Info == nil || impl == nil || targetTypeExpr == nil {
		return nil, false
	}
	genericNames := g.implSpecializationGenericNames(method)
	bindings := g.concreteCompileContextBindings(method.Info, genericNames)
	bindings = g.mergeConcreteTypeBindings(method.Info.Package, genericNames, bindings, ctx.typeBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if genericTarget, ok := targetTypeExpr.(*ast.GenericTypeExpression); ok && genericTarget != nil {
		if targetParams := g.nominalTargetGenericParams(method); len(targetParams) > 0 {
			_ = g.bindGenericTypeArguments(method.Info.Package, bindings, targetParams, genericTarget.Arguments)
		}
	}
	targetTemplate := g.specializedImplTargetTemplate(impl, bindings)
	if targetTemplate == nil {
		targetTemplate = impl.TargetType
	}
	if targetTemplate != nil {
		if !g.specializedTypeTemplateMatches(method.Info.Package, targetTemplate, targetTypeExpr, genericNames, bindings, make(map[string]struct{})) {
			return nil, false
		}
	}
	if len(call.TypeArguments) > 0 {
		if method.Info.Definition == nil || len(method.Info.Definition.GenericParams) != len(call.TypeArguments) {
			return nil, false
		}
		for idx, arg := range call.TypeArguments {
			if arg == nil {
				return nil, false
			}
			gp := method.Info.Definition.GenericParams[idx]
			if gp == nil || gp.Name == nil || gp.Name.Name == "" {
				return nil, false
			}
			bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, method.Info.Package, arg)
		}
	}
	for idx, arg := range call.Arguments {
		if idx >= len(method.Info.Params) {
			break
		}
		paramTypeExpr := method.Info.Params[idx].TypeExpr
		if paramTypeExpr == nil {
			continue
		}
		actualExpr, ok := g.inferExpressionTypeExpr(ctx, arg, "")
		if !ok || actualExpr == nil {
			continue
		}
		_ = g.specializedTypeTemplateMatches(method.Info.Package, paramTypeExpr, actualExpr, genericNames, bindings, make(map[string]struct{}))
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil && method.Info.Definition != nil && method.Info.Definition.ReturnType != nil {
		returnExpr := g.functionReturnTypeExprWithBindings(method.Info, bindings)
		if returnExpr != nil {
			_ = g.specializedTypeTemplateMatches(method.Info.Package, returnExpr, expectedExpr, genericNames, bindings, make(map[string]struct{}))
		}
	}
	bindings = g.normalizeConcreteTypeBindings(method.Info.Package, bindings, genericNames)
	if len(bindings) == 0 {
		return nil, false
	}
	return bindings, true
}

func (g *generator) staticReceiverTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if g == nil {
		return nil, false
	}
	if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, goType); ok && inferred != nil {
		return normalizeTypeExprForPackage(g, ctx.packageName, inferred), true
	}
	return nil, false
}

func (g *generator) staticTargetTypeExpr(ctx *compileContext, expr ast.Expression) (ast.TypeExpression, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil && ident.Name != "" && ctx != nil {
		if _, exists := ctx.lookup(ident.Name); !exists {
			if bound, ok := ctx.typeBindings[ident.Name]; ok && bound != nil {
				return normalizeTypeExprForPackage(g, ctx.packageName, bound), true
			}
		}
	}
	if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, ""); ok && inferred != nil {
		return normalizeTypeExprForPackage(g, ctx.packageName, inferred), true
	}
	return nil, false
}

func (g *generator) inferExpressionTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if g == nil {
		return nil, false
	}
	if inferred, ok := g.inferLocalTypeExpr(ctx, expr, goType); ok && inferred != nil {
		return g.lowerNormalizedTypeExpr(ctx, inferred), true
	}
	if goType != "" {
		if inferred, ok := g.typeExprForGoType(goType); ok && inferred != nil {
			return g.lowerNormalizedTypeExpr(ctx, inferred), true
		}
	}
	return nil, false
}

func (g *generator) functionReturnTypeExpr(info *functionInfo) ast.TypeExpression {
	return g.functionReturnTypeExprWithBindings(info, g.compileContextTypeBindings(info))
}

func (g *generator) functionReturnTypeExprWithBindings(info *functionInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || info == nil || info.Definition == nil || info.Definition.ReturnType == nil {
		return nil
	}
	retExpr := info.Definition.ReturnType
	if impl := g.implMethodByInfo[info]; impl != nil {
		retExpr = resolveSelfTypeExpr(retExpr, impl.TargetType)
	}
	retExpr = substituteTypeParams(retExpr, bindings)
	return normalizeTypeExprForPackage(g, info.Package, retExpr)
}

func (g *generator) concreteCompileContextBindings(info *functionInfo, genericNames map[string]struct{}) map[string]ast.TypeExpression {
	return g.normalizeConcreteTypeBindings(info.Package, g.compileContextTypeBindings(info), genericNames)
}

func (g *generator) mergeConcreteTypeBindings(pkgName string, genericNames map[string]struct{}, base map[string]ast.TypeExpression, extra map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || len(extra) == 0 {
		return base
	}
	if base == nil {
		base = make(map[string]ast.TypeExpression)
	}
	for name, expr := range extra {
		if expr == nil {
			continue
		}
		if len(genericNames) > 0 {
			if _, ok := genericNames[name]; !ok {
				continue
			}
		}
		if _, exists := base[name]; exists {
			continue
		}
		base[name] = normalizeTypeExprForPackage(g, pkgName, expr)
	}
	return base
}

func (g *generator) normalizeConcreteTypeBindings(pkgName string, bindings map[string]ast.TypeExpression, genericNames map[string]struct{}) map[string]ast.TypeExpression {
	if g == nil || len(bindings) == 0 {
		return nil
	}
	out := make(map[string]ast.TypeExpression, len(bindings))
	for name, expr := range bindings {
		if len(genericNames) > 0 {
			if _, ok := genericNames[name]; !ok {
				continue
			}
		}
		if expr == nil {
			continue
		}
		normalized := normalizeTypeExprForPackage(g, pkgName, expr)
		if simple, ok := normalized.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == name {
			continue
		}
		if g.typeExprHasGeneric(normalized, genericNames) {
			continue
		}
		out[name] = normalized
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneTypeBindings(bindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if len(bindings) == 0 {
		return nil
	}
	out := make(map[string]ast.TypeExpression, len(bindings))
	for name, expr := range bindings {
		out[name] = expr
	}
	return out
}

func (g *generator) specializedTypeTemplateMatches(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	if len(bindings) > 0 {
		template = substituteTypeParams(template, bindings)
		actual = substituteTypeParams(actual, bindings)
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	return g.specializedTypeTemplateMatchesNormalized(pkgName, template, actual, genericNames, bindings, seen)
}

func (g *generator) specializedSameBaseGenericBindings(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	templateGeneric, ok := template.(*ast.GenericTypeExpression)
	if !ok || templateGeneric == nil {
		return false
	}
	actualGeneric, ok := actual.(*ast.GenericTypeExpression)
	if !ok || actualGeneric == nil || len(templateGeneric.Arguments) != len(actualGeneric.Arguments) {
		return false
	}
	if normalizeTypeExprString(g, pkgName, templateGeneric.Base) != normalizeTypeExprString(g, pkgName, actualGeneric.Base) {
		return false
	}
	for idx := range templateGeneric.Arguments {
		if !g.specializedBindTemplateArg(pkgName, templateGeneric.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings) {
			return false
		}
	}
	return true
}

func (g *generator) specializedBindTemplateArg(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	switch tt := template.(type) {
	case *ast.SimpleTypeExpression:
		if tt == nil || tt.Name == nil || tt.Name.Name == "" {
			return false
		}
		if _, ok := genericNames[tt.Name.Name]; ok {
			if bound, exists := bindings[tt.Name.Name]; exists {
				if normalizeTypeExprString(g, pkgName, bound) == tt.Name.Name && normalizeTypeExprString(g, pkgName, actual) != tt.Name.Name {
					bindings[tt.Name.Name] = actual
					return true
				}
				return normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual)
			}
			bindings[tt.Name.Name] = actual
			return true
		}
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	case *ast.GenericTypeExpression:
		actualGeneric, ok := actual.(*ast.GenericTypeExpression)
		if !ok || actualGeneric == nil || len(tt.Arguments) != len(actualGeneric.Arguments) {
			return false
		}
		if normalizeTypeExprString(g, pkgName, tt.Base) != normalizeTypeExprString(g, pkgName, actualGeneric.Base) {
			return false
		}
		for idx := range tt.Arguments {
			if !g.specializedBindTemplateArg(pkgName, tt.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		actualNullable, ok := actual.(*ast.NullableTypeExpression)
		return ok && actualNullable != nil && g.specializedBindTemplateArg(pkgName, tt.InnerType, actualNullable.InnerType, genericNames, bindings)
	case *ast.ResultTypeExpression:
		actualResult, ok := actual.(*ast.ResultTypeExpression)
		return ok && actualResult != nil && g.specializedBindTemplateArg(pkgName, tt.InnerType, actualResult.InnerType, genericNames, bindings)
	case *ast.UnionTypeExpression:
		actualUnion, ok := actual.(*ast.UnionTypeExpression)
		if !ok || actualUnion == nil || len(tt.Members) != len(actualUnion.Members) {
			return false
		}
		for idx := range tt.Members {
			if !g.specializedBindTemplateArg(pkgName, tt.Members[idx], actualUnion.Members[idx], genericNames, bindings) {
				return false
			}
		}
		return true
	case *ast.FunctionTypeExpression:
		actualFn, ok := actual.(*ast.FunctionTypeExpression)
		if !ok || actualFn == nil || len(tt.ParamTypes) != len(actualFn.ParamTypes) {
			return false
		}
		for idx := range tt.ParamTypes {
			if !g.specializedBindTemplateArg(pkgName, tt.ParamTypes[idx], actualFn.ParamTypes[idx], genericNames, bindings) {
				return false
			}
		}
		return g.specializedBindTemplateArg(pkgName, tt.ReturnType, actualFn.ReturnType, genericNames, bindings)
	default:
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
}

func (g *generator) specializedTypeTemplateMatchesNormalized(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	if !g.typeExprHasGeneric(template, genericNames) && !g.typeExprHasGeneric(actual, genericNames) {
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	key := specializedTypeTemplateMatchKey(template, actual)
	if _, ok := seen[key]; ok {
		return true
	}
	seen[key] = struct{}{}
	switch tt := template.(type) {
	case *ast.SimpleTypeExpression:
		if tt == nil || tt.Name == nil || tt.Name.Name == "" {
			return false
		}
		if _, ok := genericNames[tt.Name.Name]; ok {
			if bound, exists := bindings[tt.Name.Name]; exists {
				if normalizeTypeExprString(g, pkgName, bound) == tt.Name.Name && normalizeTypeExprString(g, pkgName, actual) != tt.Name.Name {
					bindings[tt.Name.Name] = actual
					return true
				}
				return normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual)
			}
			bindings[tt.Name.Name] = actual
			return true
		}
		actualSimple, ok := actual.(*ast.SimpleTypeExpression)
		return ok && actualSimple != nil && actualSimple.Name != nil && actualSimple.Name.Name == tt.Name.Name
	case *ast.GenericTypeExpression:
		actualGeneric, ok := actual.(*ast.GenericTypeExpression)
		if !ok || actualGeneric == nil || len(tt.Arguments) != len(actualGeneric.Arguments) {
			return false
		}
		if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Base, actualGeneric.Base, genericNames, bindings, seen) {
			return false
		}
		for idx := range tt.Arguments {
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings, seen) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		actualNullable, ok := actual.(*ast.NullableTypeExpression)
		return ok && actualNullable != nil && g.specializedTypeTemplateMatchesNormalized(pkgName, tt.InnerType, actualNullable.InnerType, genericNames, bindings, seen)
	case *ast.ResultTypeExpression:
		actualResult, ok := actual.(*ast.ResultTypeExpression)
		return ok && actualResult != nil && g.specializedTypeTemplateMatchesNormalized(pkgName, tt.InnerType, actualResult.InnerType, genericNames, bindings, seen)
	case *ast.UnionTypeExpression:
		actualUnion, ok := actual.(*ast.UnionTypeExpression)
		if !ok || actualUnion == nil || len(tt.Members) != len(actualUnion.Members) {
			return false
		}
		for idx := range tt.Members {
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Members[idx], actualUnion.Members[idx], genericNames, bindings, seen) {
				return false
			}
		}
		return true
	case *ast.FunctionTypeExpression:
		actualFn, ok := actual.(*ast.FunctionTypeExpression)
		if !ok || actualFn == nil || len(tt.ParamTypes) != len(actualFn.ParamTypes) {
			return false
		}
		for idx := range tt.ParamTypes {
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.ParamTypes[idx], actualFn.ParamTypes[idx], genericNames, bindings, seen) {
				return false
			}
		}
		return g.specializedTypeTemplateMatchesNormalized(pkgName, tt.ReturnType, actualFn.ReturnType, genericNames, bindings, seen)
	default:
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
}

func (g *generator) normalizeTypeExprForSpecialization(pkgName string, expr ast.TypeExpression, seen map[string]struct{}) ast.TypeExpression {
	if g == nil || expr == nil {
		return expr
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			return normalizeTypeExprForPackage(g, pkgName, expr)
		}
		key := pkgName + "|" + t.Name.Name
		if _, ok := seen[key]; ok {
			return expr
		}
		nextSeen := make(map[string]struct{}, len(seen)+1)
		for existing := range seen {
			nextSeen[existing] = struct{}{}
		}
		nextSeen[key] = struct{}{}
		if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil && expanded != expr {
			return g.normalizeTypeExprForSpecialization(pkgName, expanded, nextSeen)
		}
		return expr
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		if baseName, ok := typeExprBaseName(t.Base); ok && baseName != "" {
			key := pkgName + "|" + baseName + "<" + normalizeTypeExprListKey(g, pkgName, t.Arguments) + ">"
			if _, ok := seen[key]; ok {
				return expr
			}
			nextSeen := make(map[string]struct{}, len(seen)+1)
			for existing := range seen {
				nextSeen[existing] = struct{}{}
			}
			nextSeen[key] = struct{}{}
			if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil && expanded != expr {
				return g.normalizeTypeExprForSpecialization(pkgName, expanded, nextSeen)
			}
		}
		base := g.normalizeTypeExprForSpecialization(pkgName, t.Base, seen)
		changed := base != t.Base
		args := make([]ast.TypeExpression, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			next := g.normalizeTypeExprForSpecialization(pkgName, arg, seen)
			args = append(args, next)
			if next != arg {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewGenericTypeExpression(base, args)
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		inner := g.normalizeTypeExprForSpecialization(pkgName, t.InnerType, seen)
		if inner == t.InnerType {
			return expr
		}
		return ast.NewNullableTypeExpression(inner)
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		inner := g.normalizeTypeExprForSpecialization(pkgName, t.InnerType, seen)
		if inner == t.InnerType {
			return expr
		}
		return ast.NewResultTypeExpression(inner)
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		members := make([]ast.TypeExpression, 0, len(t.Members))
		for _, member := range t.Members {
			next := g.normalizeTypeExprForSpecialization(pkgName, member, seen)
			members = append(members, next)
			if next != member {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		params := make([]ast.TypeExpression, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			next := g.normalizeTypeExprForSpecialization(pkgName, param, seen)
			params = append(params, next)
			if next != param {
				changed = true
			}
		}
		ret := g.normalizeTypeExprForSpecialization(pkgName, t.ReturnType, seen)
		if ret != t.ReturnType {
			changed = true
		}
		if !changed {
			return normalizeCallableSyntaxTypeExpr(expr)
		}
		return normalizeCallableSyntaxTypeExpr(ast.NewFunctionTypeExpression(params, ret))
	default:
		return normalizeTypeExprForPackage(g, pkgName, expr)
	}
}

func specializedTypeTemplateMatchKey(template ast.TypeExpression, actual ast.TypeExpression) string {
	return fmt.Sprintf("%T:%x|%T:%x", template, typeExprPointer(template), actual, typeExprPointer(actual))
}

func typeExprPointer(expr ast.TypeExpression) uintptr {
	if expr == nil {
		return 0
	}
	value := reflect.ValueOf(expr)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return 0
	}
	return value.Pointer()
}

func (g *generator) specializedImplFunctionKey(info *functionInfo, bindings map[string]ast.TypeExpression) string {
	if info == nil {
		return ""
	}
	base := strings.TrimSpace(info.Name)
	if info.QualifiedName != "" {
		base = strings.TrimSpace(info.QualifiedName)
	}
	if base == "" {
		base = strings.TrimSpace(info.GoName)
	}
	if pkg := strings.TrimSpace(info.Package); pkg != "" {
		base = pkg + "::" + base
	}
	if len(bindings) == 0 {
		return base
	}
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names)+1)
	parts = append(parts, base)
	for _, name := range names {
		parts = append(parts, name+"="+normalizeTypeExprString(g, info.Package, bindings[name]))
	}
	return strings.Join(parts, "|")
}
