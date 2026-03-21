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
	return g.compileResolvedMethodCall(ctx, call, expected, &methodInfo{MethodName: method.Name, ExpectsSelf: true, Info: info}, receiverExpr, receiverType, callNode)
}

func (g *generator) ensureSpecializedNativeInterfaceGenericDefaultMethod(method *nativeInterfaceGenericMethod, receiverType string, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string, bindings map[string]ast.TypeExpression) (*functionInfo, bool) {
	if g == nil || method == nil || method.DefaultDefinition == nil || method.DefaultDefinition.Body == nil || receiverType == "" {
		return nil, false
	}
	mergedBindings := g.nativeInterfaceGenericDefaultMethodBindings(method, bindings)
	key := g.specializedNativeInterfaceGenericMethodKey(method, receiverType, mergedBindings)
	if existing, ok := g.specializedFunctionIndex[key]; ok && existing != nil && existing.Compileable {
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
		Compileable:   true,
	}
	if !g.fillNativeInterfaceGenericDefaultMethodInfo(info, receiverType, method, paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType) {
		return nil, false
	}
	g.specializedFunctions = append(g.specializedFunctions, info)
	g.specializedFunctionIndex[key] = info
	if !g.bodyCompileable(info, info.ReturnType) {
		delete(g.specializedFunctionIndex, key)
		g.specializedFunctions = removeSpecializedFunction(g.specializedFunctions, info)
		return nil, false
	}
	info.Compileable = true
	info.Reason = ""
	return info, true
}

func (g *generator) nativeInterfaceGenericDefaultMethodBindings(method *nativeInterfaceGenericMethod, bindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || method == nil {
		return cloneTypeBindings(bindings)
	}
	merged := cloneTypeBindings(bindings)
	iface := g.interfaces[method.InterfaceName]
	for name, expr := range nativeInterfaceBindings(iface, method.InterfaceArgs) {
		if expr == nil {
			continue
		}
		if merged == nil {
			merged = make(map[string]ast.TypeExpression, len(method.InterfaceArgs)+len(bindings))
		}
		if _, exists := merged[name]; exists {
			continue
		}
		merged[name] = normalizeTypeExprForPackage(g, method.InterfacePackage, expr)
	}
	return merged
}

func (g *generator) fillNativeInterfaceGenericDefaultMethodInfo(info *functionInfo, receiverType string, method *nativeInterfaceGenericMethod, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string) bool {
	if g == nil || info == nil || info.Definition == nil || method == nil || receiverType == "" || returnGoType == "" {
		return false
	}
	receiverInfo := g.nativeInterfaceInfoForGoType(receiverType)
	if receiverInfo == nil || receiverInfo.TypeExpr == nil {
		return false
	}
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
			paramTypeExpr = receiverInfo.TypeExpr
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
