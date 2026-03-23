package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) nativeInterfaceDefaultAdapterMethod(goType string, method *nativeInterfaceMethod) *nativeInterfaceAdapterMethod {
	if g == nil || method == nil || method.DefaultDefinition == nil || method.DefaultDefinition.Body == nil || goType == "" {
		return nil
	}
	defaultMethod := &nativeInterfaceGenericMethod{
		Name:              method.Name,
		GoName:            method.GoName,
		InterfaceName:     method.InterfaceName,
		InterfacePackage:  method.InterfacePackage,
		InterfaceArgs:     method.InterfaceArgs,
		ParamTypeExprs:    cloneTypeExprSlice(method.ParamTypeExprs),
		ReturnTypeExpr:    cloneTypeExpr(method.ReturnTypeExpr),
		DefaultDefinition: method.DefaultDefinition,
	}
	info, ok := g.ensureSpecializedNativeInterfaceDefaultMethod(defaultMethod, goType, method.ParamGoTypes, method.ReturnGoType)
	if !ok || info == nil || !info.Compileable {
		return nil
	}
	g.refreshRepresentableFunctionInfo(info)
	adapterMethod := &nativeInterfaceAdapterMethod{
		Info:                 info,
		CompiledReturnGoType: info.ReturnType,
		ParamGoTypes:         append([]string(nil), method.ParamGoTypes...),
		ReturnGoType:         method.ReturnGoType,
	}
	if len(info.Params) > 1 {
		adapterMethod.CompiledParamGoTypes = make([]string, 0, len(info.Params)-1)
		for idx := 1; idx < len(info.Params); idx++ {
			adapterMethod.CompiledParamGoTypes = append(adapterMethod.CompiledParamGoTypes, info.Params[idx].GoType)
		}
	}
	return adapterMethod
}

func (g *generator) ensureSpecializedNativeInterfaceDefaultMethod(method *nativeInterfaceGenericMethod, receiverGoType string, paramGoTypes []string, returnGoType string) (*functionInfo, bool) {
	if g == nil || method == nil || method.DefaultDefinition == nil || method.DefaultDefinition.Body == nil || receiverGoType == "" || returnGoType == "" {
		return nil, false
	}
	receiverTypeExpr, ok := g.typeExprForGoType(receiverGoType)
	if !ok || receiverTypeExpr == nil {
		return nil, false
	}
	mergedBindings := g.nativeInterfaceGenericDefaultMethodBindings(method, nil)
	key := g.specializedNativeInterfaceDefaultMethodKey(method, receiverGoType, mergedBindings)
	if existing, ok := g.specializedFunctionIndex[key]; ok && existing != nil {
		if _, building := g.nativeInterfaceSpecializing[key]; building {
			return existing, true
		}
		existing.Name = fmt.Sprintf("iface %s.%s", method.InterfaceName, method.Name)
		existing.Package = method.InterfacePackage
		existing.QualifiedName = fmt.Sprintf("iface %s.%s", method.InterfaceName, method.Name)
		existing.Definition = method.DefaultDefinition
		existing.TypeBindings = cloneTypeBindings(mergedBindings)
		existing.HasOriginal = false
		existing.InternalOnly = true
		existing.Compileable = false
		existing.Reason = ""
		if !g.fillNativeInterfaceDefaultMethodInfo(existing, receiverGoType, receiverTypeExpr, method, paramGoTypes, returnGoType) {
			return nil, false
		}
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
	if !g.fillNativeInterfaceDefaultMethodInfo(info, receiverGoType, receiverTypeExpr, method, paramGoTypes, returnGoType) {
		return nil, false
	}
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

func (g *generator) fillNativeInterfaceDefaultMethodInfo(info *functionInfo, receiverGoType string, receiverTypeExpr ast.TypeExpression, method *nativeInterfaceGenericMethod, paramGoTypes []string, returnGoType string) bool {
	if g == nil || info == nil || info.Definition == nil || method == nil || receiverGoType == "" || receiverTypeExpr == nil || returnGoType == "" {
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
			paramTypeExpr = cloneTypeExpr(receiverTypeExpr)
			goType = receiverGoType
		} else {
			if nonSelfIdx >= len(method.ParamTypeExprs) || nonSelfIdx >= len(paramGoTypes) {
				return false
			}
			paramTypeExpr = cloneTypeExpr(method.ParamTypeExprs[nonSelfIdx])
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
	if len(method.ParamTypeExprs) != nonSelfIdx {
		return false
	}
	info.Params = params
	info.ReturnType = returnGoType
	info.SupportedTypes = supported && method.ReturnTypeExpr != nil
	info.Arity = len(params)
	if !info.SupportedTypes {
		info.Compileable = false
		info.Reason = "unsupported param or return type"
		info.Arity = -1
		return false
	}
	return true
}

func (g *generator) specializedNativeInterfaceDefaultMethodKey(method *nativeInterfaceGenericMethod, receiverType string, bindings map[string]ast.TypeExpression) string {
	if g == nil || method == nil {
		return ""
	}
	parts := []string{
		"iface-default",
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

func cloneTypeExprSlice(exprs []ast.TypeExpression) []ast.TypeExpression {
	if len(exprs) == 0 {
		return nil
	}
	cloned := make([]ast.TypeExpression, 0, len(exprs))
	for _, expr := range exprs {
		cloned = append(cloned, cloneTypeExpr(expr))
	}
	return cloned
}

func cloneTypeExpr(expr ast.TypeExpression) ast.TypeExpression {
	switch t := expr.(type) {
	case nil:
		return nil
	case *ast.SimpleTypeExpression:
		if t == nil {
			return nil
		}
		return ast.NewSimpleTypeExpression(ast.NewIdentifier(t.Name.Name))
	case *ast.GenericTypeExpression:
		if t == nil {
			return nil
		}
		args := make([]ast.TypeExpression, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			args = append(args, cloneTypeExpr(arg))
		}
		return ast.NewGenericTypeExpression(cloneTypeExpr(t.Base), args)
	case *ast.NullableTypeExpression:
		if t == nil {
			return nil
		}
		return ast.NewNullableTypeExpression(cloneTypeExpr(t.InnerType))
	case *ast.ResultTypeExpression:
		if t == nil {
			return nil
		}
		return ast.NewResultTypeExpression(cloneTypeExpr(t.InnerType))
	case *ast.UnionTypeExpression:
		if t == nil {
			return nil
		}
		members := make([]ast.TypeExpression, 0, len(t.Members))
		for _, member := range t.Members {
			members = append(members, cloneTypeExpr(member))
		}
		return ast.NewUnionTypeExpression(members)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return nil
		}
		params := make([]ast.TypeExpression, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			params = append(params, cloneTypeExpr(param))
		}
		return ast.NewFunctionTypeExpression(params, cloneTypeExpr(t.ReturnType))
	default:
		return expr
	}
}
