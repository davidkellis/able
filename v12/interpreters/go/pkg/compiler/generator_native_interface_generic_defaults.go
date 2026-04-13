package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileStaticNativeInterfaceGenericDefaultMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, receiverExpr string, receiverType string, method *nativeInterfaceGenericMethod, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string, bindings map[string]ast.TypeExpression, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || method == nil || method.DefaultDefinition == nil || method.DefaultDefinition.Body == nil {
		return nil, "", "", false
	}
	info, ok := g.ensureSpecializedNativeInterfaceGenericDefaultMethod(method, receiverType, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, bindings)
	if !ok || info == nil || !info.Compileable {
		return nil, "", "", false
	}
	if len(call.Arguments) != len(paramGoTypes) {
		ctx.setReason("call arity mismatch")
		return nil, "", "", false
	}
	lines := make([]string, 0, len(call.Arguments)*4+8)
	args := make([]string, 0, len(call.Arguments)+1)
	receiverLines, coercedReceiverExpr, _, ok := g.prepareStaticCallArg(ctx, receiverExpr, receiverType, info.Params[0].GoType)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, receiverLines...)
	args = append(args, coercedReceiverExpr)
	for idx, arg := range call.Arguments {
		expectedType := paramGoTypes[idx]
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
	lines = append(lines, fmt.Sprintf("%s, %s := __able_compiled_%s(%s)", resultTemp, controlTemp, info.GoName, strings.Join(args, ", ")))
	lines = append(lines, "__able_pop_call_frame()")
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	return g.finishNativeInterfaceGenericCallReturn(ctx, lines, resultTemp, info.ReturnType, expected)
}

func (g *generator) ensureSpecializedNativeInterfaceGenericDefaultMethod(method *nativeInterfaceGenericMethod, receiverType string, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string, bindings map[string]ast.TypeExpression) (*functionInfo, bool) {
	if g == nil || method == nil || method.DefaultDefinition == nil || method.DefaultDefinition.Body == nil || receiverType == "" {
		return nil, false
	}
	receiverTypeExpr, concreteReceiverGoType, mergedBindings, ok := g.nativeInterfaceDefaultReceiverInfo(receiverType, method, bindings)
	if !ok || receiverTypeExpr == nil || concreteReceiverGoType == "" {
		return nil, false
	}
	key := g.specializedNativeInterfaceGenericMethodKey(method, concreteReceiverGoType, mergedBindings)
	if existing, ok := g.specializedFunctionIndex[key]; ok && existing != nil {
		if _, building := g.nativeInterfaceSpecializing[key]; building {
			return existing, true
		}
		if existing.Compileable {
			return existing, true
		}
		g.invalidateFunctionDerivedInfo(existing)
		existing.Name = fmt.Sprintf("iface %s.%s", method.InterfaceName, method.Name)
		existing.Package = method.InterfacePackage
		existing.QualifiedName = fmt.Sprintf("iface %s.%s", method.InterfaceName, method.Name)
		existing.Definition = method.DefaultDefinition
		existing.TypeBindings = cloneTypeBindings(mergedBindings)
		existing.HasOriginal = false
		existing.InternalOnly = true
		existing.Compileable = false
		existing.Reason = ""
		if !g.fillNativeInterfaceGenericDefaultMethodInfo(existing, concreteReceiverGoType, receiverTypeExpr, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType) {
			return nil, false
		}
		g.registerNativeInterfaceDefaultMethodInfo(existing, method, concreteReceiverGoType)
		existing.Compileable = true
		g.nativeInterfaceSpecializing[key] = struct{}{}
		defer delete(g.nativeInterfaceSpecializing, key)
		g.touchNativeInterfaceAdapters()
		if g.bodyCompileable(existing, existing.ReturnType) {
			existing.Compileable = true
			existing.Reason = ""
		} else {
			g.discardSpecializedFunctionInfo(key, existing)
			return nil, false
		}
		return existing, true
	}
	info := &functionInfo{
		Name:          fmt.Sprintf("iface %s.%s", method.InterfaceName, method.Name),
		Package:       method.InterfacePackage,
		QualifiedName: fmt.Sprintf("iface %s.%s", method.InterfaceName, method.Name),
		GoName:        g.mangler.unique(fmt.Sprintf("iface_%s_%s_default", sanitizeIdent(method.InterfaceName), sanitizeIdent(method.Name))),
		Definition:    method.DefaultDefinition,
		TypeBindings:  cloneTypeBindings(mergedBindings),
		HasOriginal:   false,
		InternalOnly:  true,
		Compileable:   false,
	}
	if !g.fillNativeInterfaceGenericDefaultMethodInfo(info, concreteReceiverGoType, receiverTypeExpr, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType) {
		return nil, false
	}
	g.registerNativeInterfaceDefaultMethodInfo(info, method, concreteReceiverGoType)
	g.specializedFunctions = append(g.specializedFunctions, info)
	info.Compileable = true
	g.nativeInterfaceSpecializing[key] = struct{}{}
	defer delete(g.nativeInterfaceSpecializing, key)
	g.touchNativeInterfaceAdapters()
	g.specializedFunctionIndex[key] = info
	if g.bodyCompileable(info, info.ReturnType) {
		info.Compileable = true
		info.Reason = ""
	} else {
		g.discardSpecializedFunctionInfo(key, info)
		return nil, false
	}
	return info, true
}

func (g *generator) nativeInterfaceGenericDefaultMethodBindings(method *nativeInterfaceGenericMethod, bindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || method == nil {
		return cloneTypeBindings(bindings)
	}
	merged := cloneTypeBindings(bindings)
	iface, _, _ := g.interfaceDefinitionForPackage(method.InterfacePackage, method.InterfaceName)
	for name, expr := range nativeInterfaceBindings(iface, method.InterfaceArgs) {
		if expr == nil {
			continue
		}
		if merged == nil {
			merged = make(map[string]ast.TypeExpression, len(method.InterfaceArgs)+len(bindings))
		}
		if !g.mergeConcreteBindings(method.InterfacePackage, merged, map[string]ast.TypeExpression{name: expr}) {
			return nil
		}
	}
	return merged
}

func (g *generator) fillNativeInterfaceGenericDefaultMethodInfo(info *functionInfo, receiverType string, receiverTypeExpr ast.TypeExpression, method *nativeInterfaceGenericMethod, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string) bool {
	if g == nil || info == nil || info.Definition == nil || method == nil || receiverType == "" || receiverTypeExpr == nil || returnGoType == "" {
		return false
	}
	g.invalidateFunctionDerivedInfo(info)
	def := info.Definition
	expectsSelf := methodDefinitionExpectsSelf(def)
	params := make([]paramInfo, 0, len(def.Params))
	supported := true
	nonSelfIdx := 0
	for idx, param := range def.Params {
		if param == nil {
			supported = false
			continue
		}
		name := fmt.Sprintf("arg%d", idx)
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			name = ident.Name
		}
		paramTypeExpr := param.ParamType
		goType := ""
		if expectsSelf && len(params) == 0 {
			paramTypeExpr = cloneTypeExpr(receiverTypeExpr)
			goType = receiverType
		} else {
			if nonSelfIdx >= len(paramTypeExprs) || nonSelfIdx >= len(paramGoTypes) {
				return false
			}
			paramTypeExpr = paramTypeExprs[nonSelfIdx]
			goType = paramGoTypes[nonSelfIdx]
			nonSelfIdx++
		}
		if paramTypeExpr == nil || goType == "" {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      name,
			GoName:    safeParamName(name, idx),
			GoType:    goType,
			TypeExpr:  paramTypeExpr,
			Supported: paramTypeExpr != nil && goType != "",
		})
	}
	if len(paramTypeExprs) != nonSelfIdx {
		return false
	}
	info.Params = params
	info.ReturnType = returnGoType
	info.SupportedTypes = supported && returnTypeExpr != nil
	info.Arity = len(params)
	if !info.SupportedTypes {
		info.Compileable = false
		info.Reason = "unsupported param or return type"
		info.Arity = -1
		return false
	}
	return true
}

func (g *generator) specializedNativeInterfaceGenericMethodKey(method *nativeInterfaceGenericMethod, receiverType string, bindings map[string]ast.TypeExpression) string {
	if g == nil || method == nil {
		return ""
	}
	parts := []string{
		"iface-default-generic",
		method.InterfacePackage,
		method.InterfaceName,
		receiverType,
		method.Name,
	}
	if len(bindings) == 0 {
		return strings.Join(parts, "::")
	}
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		parts = append(parts, name+"="+normalizeTypeExprString(g, method.InterfacePackage, bindings[name]))
	}
	return strings.Join(parts, "::")
}
