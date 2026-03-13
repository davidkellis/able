package compiler

import (
	"fmt"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type nativeInterfaceMethod struct {
	Name             string
	GoName           string
	InterfaceName    string
	InterfacePackage string
	InterfaceArgs    []ast.TypeExpression
	ParamGoTypes     []string
	ParamTypeExprs   []ast.TypeExpression
	ReturnGoType     string
	ReturnTypeExpr   ast.TypeExpression
	OptionalLast     bool
}

type nativeInterfaceAdapter struct {
	GoType      string
	Token       string
	AdapterType string
	WrapHelper  string
	Methods     map[string]*functionInfo
}

type nativeInterfaceInfo struct {
	Key               string
	GoType            string
	TypeExpr          ast.TypeExpression
	TypeString        string
	MarkerMethod      string
	ToRuntimeMethod   string
	FromRuntimeHelper string
	FromRuntimePanic  string
	ToRuntimeHelper   string
	ToRuntimePanic    string
	RuntimeAdapter    string
	RuntimeWrapHelper string
	Methods           []*nativeInterfaceMethod
	Adapters          []*nativeInterfaceAdapter
}

func (g *generator) nativeInterfaceInfoForGoType(goType string) *nativeInterfaceInfo {
	if g == nil || goType == "" || g.nativeInterfaces == nil {
		return nil
	}
	for _, info := range g.nativeInterfaces {
		if info != nil && info.GoType == goType {
			return info
		}
	}
	return nil
}

func (g *generator) nativeInterfaceMethodForGoType(goType string, methodName string) (*nativeInterfaceMethod, bool) {
	info := g.nativeInterfaceInfoForGoType(goType)
	if info == nil || methodName == "" {
		return nil, false
	}
	for _, method := range info.Methods {
		if method != nil && method.Name == methodName {
			return method, true
		}
	}
	return nil, false
}

func (g *generator) nativeInterfaceAdapterForActual(info *nativeInterfaceInfo, actual string) (*nativeInterfaceAdapter, bool) {
	if info == nil || actual == "" {
		return nil, false
	}
	for _, adapter := range info.Adapters {
		if adapter != nil && adapter.GoType == actual {
			return adapter, true
		}
	}
	return nil, false
}

func (g *generator) sortedNativeInterfaceKeys() []string {
	if g == nil || len(g.nativeInterfaces) == 0 {
		return nil
	}
	keys := make([]string, 0, len(g.nativeInterfaces))
	for key := range g.nativeInterfaces {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func interfaceTypeArgs(expr ast.TypeExpression) []ast.TypeExpression {
	if generic, ok := expr.(*ast.GenericTypeExpression); ok && generic != nil {
		return generic.Arguments
	}
	return nil
}

func functionSignatureExpectsSelf(sig *ast.FunctionSignature) bool {
	if sig == nil || len(sig.Params) == 0 {
		return false
	}
	first := sig.Params[0]
	if first == nil || first.ParamType == nil {
		if ident, ok := first.Name.(*ast.Identifier); ok && ident != nil {
			return strings.EqualFold(ident.Name, "self")
		}
		return false
	}
	simple, ok := first.ParamType.(*ast.SimpleTypeExpression)
	return ok && simple != nil && simple.Name != nil && simple.Name.Name == "Self"
}

func typeExprUsesSelf(expr ast.TypeExpression) bool {
	switch t := expr.(type) {
	case nil:
		return false
	case *ast.SimpleTypeExpression:
		return t != nil && t.Name != nil && t.Name.Name == "Self"
	case *ast.GenericTypeExpression:
		if t == nil {
			return false
		}
		if typeExprUsesSelf(t.Base) {
			return true
		}
		for _, arg := range t.Arguments {
			if typeExprUsesSelf(arg) {
				return true
			}
		}
		return false
	case *ast.NullableTypeExpression:
		return typeExprUsesSelf(t.InnerType)
	case *ast.ResultTypeExpression:
		return typeExprUsesSelf(t.InnerType)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if typeExprUsesSelf(member) {
				return true
			}
		}
		return false
	case *ast.FunctionTypeExpression:
		if typeExprUsesSelf(t.ReturnType) {
			return true
		}
		for _, param := range t.ParamTypes {
			if typeExprUsesSelf(param) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func normalizeTypeExprForPackage(g *generator, pkgName string, expr ast.TypeExpression) ast.TypeExpression {
	if g == nil || expr == nil {
		return expr
	}
	expanded := g.expandTypeAliasForPackage(pkgName, expr)
	if expanded != nil {
		return expanded
	}
	return expr
}

func normalizeTypeExprString(g *generator, pkgName string, expr ast.TypeExpression) string {
	return typeExpressionToString(normalizeTypeExprForPackage(g, pkgName, expr))
}

func normalizeTypeExprListKey(g *generator, pkgName string, exprs []ast.TypeExpression) string {
	if len(exprs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		parts = append(parts, normalizeTypeExprString(g, pkgName, expr))
	}
	return strings.Join(parts, "|")
}

func interfaceExprInfo(g *generator, pkgName string, expr ast.TypeExpression) (string, string, []ast.TypeExpression, *ast.InterfaceDefinition, bool) {
	if g == nil || expr == nil {
		return "", "", nil, nil, false
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	baseName, ok := typeExprBaseName(expr)
	if !ok || baseName == "" || !g.isInterfaceName(baseName) || baseName == "Error" {
		return "", "", nil, nil, false
	}
	iface := g.interfaces[baseName]
	if iface == nil {
		return "", "", nil, nil, false
	}
	ifacePkg := pkgName
	if recorded, ok := g.interfacePackages[baseName]; ok && recorded != "" {
		ifacePkg = recorded
	}
	args := interfaceTypeArgs(expr)
	if len(iface.GenericParams) != len(args) {
		if len(iface.GenericParams) != 0 {
			return "", "", nil, nil, false
		}
		args = nil
	}
	return ifacePkg, baseName, args, iface, true
}

func nativeInterfaceBindings(iface *ast.InterfaceDefinition, args []ast.TypeExpression) map[string]ast.TypeExpression {
	if iface == nil || len(iface.GenericParams) == 0 || len(iface.GenericParams) != len(args) {
		return nil
	}
	bindings := make(map[string]ast.TypeExpression, len(args))
	for idx, gp := range iface.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" || args[idx] == nil {
			continue
		}
		bindings[gp.Name.Name] = args[idx]
	}
	return bindings
}

func (g *generator) collectNativeInterfaceMethods(pkgName string, expr ast.TypeExpression, seen map[string]struct{}, methods map[string]*nativeInterfaceMethod) bool {
	if g == nil || expr == nil {
		return false
	}
	ifacePkg, ifaceName, ifaceArgs, ifaceDef, ok := interfaceExprInfo(g, pkgName, expr)
	if !ok {
		return false
	}
	key := ifaceName + "<" + normalizeTypeExprListKey(g, ifacePkg, ifaceArgs) + ">"
	if _, exists := seen[key]; exists {
		return true
	}
	seen[key] = struct{}{}
	bindings := nativeInterfaceBindings(ifaceDef, ifaceArgs)
	for _, baseExpr := range ifaceDef.BaseInterfaces {
		if baseExpr == nil {
			return false
		}
		next := substituteTypeParams(baseExpr, bindings)
		next = normalizeTypeExprForPackage(g, ifacePkg, next)
		if !g.collectNativeInterfaceMethods(ifacePkg, next, seen, methods) {
			return false
		}
	}
	mapper := NewTypeMapper(g, ifacePkg)
	for _, sig := range ifaceDef.Signatures {
		if sig == nil || sig.Name == nil || sig.Name.Name == "" {
			return false
		}
		if len(sig.GenericParams) > 0 || len(sig.WhereClause) > 0 {
			return false
		}
		expectsSelf := functionSignatureExpectsSelf(sig)
		paramStart := 0
		if expectsSelf {
			paramStart = 1
		}
		capHint := 0
		if len(sig.Params) > paramStart {
			capHint = len(sig.Params) - paramStart
		}
		paramTypes := make([]ast.TypeExpression, 0, capHint)
		paramGoTypes := make([]string, 0, capHint)
		optionalLast := false
		for idx := paramStart; idx < len(sig.Params); idx++ {
			param := sig.Params[idx]
			if param == nil || param.ParamType == nil {
				return false
			}
			substituted := substituteTypeParams(param.ParamType, bindings)
			substituted = normalizeTypeExprForPackage(g, ifacePkg, substituted)
			if typeExprUsesSelf(substituted) {
				return false
			}
			goType, ok := mapper.Map(substituted)
			if !ok || goType == "" {
				return false
			}
			paramTypes = append(paramTypes, substituted)
			paramGoTypes = append(paramGoTypes, goType)
			if idx == len(sig.Params)-1 {
				if _, ok := substituted.(*ast.NullableTypeExpression); ok {
					optionalLast = true
				}
			}
		}
		returnExpr := normalizeTypeExprForPackage(g, ifacePkg, substituteTypeParams(sig.ReturnType, bindings))
		if typeExprUsesSelf(returnExpr) {
			return false
		}
		returnGoType, ok := mapper.Map(returnExpr)
		if !ok || returnGoType == "" {
			return false
		}
		method := &nativeInterfaceMethod{
			Name:             sig.Name.Name,
			GoName:           sanitizeIdent(sig.Name.Name),
			InterfaceName:    ifaceName,
			InterfacePackage: ifacePkg,
			InterfaceArgs:    ifaceArgs,
			ParamGoTypes:     paramGoTypes,
			ParamTypeExprs:   paramTypes,
			ReturnGoType:     returnGoType,
			ReturnTypeExpr:   returnExpr,
			OptionalLast:     optionalLast,
		}
		if existing, ok := methods[method.Name]; ok {
			if existing.ReturnGoType != method.ReturnGoType || existing.OptionalLast != method.OptionalLast || len(existing.ParamGoTypes) != len(method.ParamGoTypes) {
				return false
			}
			for idx := range existing.ParamGoTypes {
				if existing.ParamGoTypes[idx] != method.ParamGoTypes[idx] {
					return false
				}
			}
			continue
		}
		methods[method.Name] = method
	}
	return true
}

func (g *generator) nativeInterfaceMethodImpl(goType string, method *nativeInterfaceMethod) *functionInfo {
	if g == nil || method == nil || goType == "" {
		return nil
	}
	expectedArgs := normalizeTypeExprListKey(g, method.InterfacePackage, method.InterfaceArgs)
	var found *functionInfo
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable || impl.ImplName != "" {
			continue
		}
		if impl.InterfaceName != method.InterfaceName || impl.MethodName != method.Name {
			continue
		}
		if len(impl.Info.Params) == 0 || impl.Info.Params[0].GoType != goType {
			continue
		}
		actualPkg := impl.Info.Package
		if actualPkg == "" {
			actualPkg = method.InterfacePackage
		}
		if normalizeTypeExprListKey(g, actualPkg, impl.InterfaceArgs) != expectedArgs {
			continue
		}
		if len(impl.Info.Params)-1 != len(method.ParamGoTypes) || impl.Info.ReturnType != method.ReturnGoType {
			continue
		}
		matches := true
		for idx := range method.ParamGoTypes {
			if impl.Info.Params[idx+1].GoType != method.ParamGoTypes[idx] {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		if found != nil && found != impl.Info {
			return nil
		}
		found = impl.Info
	}
	return found
}

func (g *generator) ensureNativeInterfaceInfo(pkgName string, expr ast.TypeExpression) (*nativeInterfaceInfo, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	ifacePkg, ifaceName, ifaceArgs, _, ok := interfaceExprInfo(g, pkgName, expr)
	if !ok {
		return nil, false
	}
	keyParts := []string{ifaceName}
	if len(ifaceArgs) > 0 {
		keyParts = append(keyParts, normalizeTypeExprListKey(g, ifacePkg, ifaceArgs))
	}
	key := strings.Join(keyParts, "<")
	key = strings.TrimSuffix(key, "<")
	if info, ok := g.nativeInterfaces[key]; ok && info != nil {
		if _, building := g.nativeInterfaceBuilding[key]; building {
			return info, true
		}
		g.refreshNativeInterfaceAdapters(info)
		return info, true
	}
	baseToken := sanitizeIdent("__able_iface_" + ifaceName)
	if len(ifaceArgs) > 0 {
		baseToken = sanitizeIdent(baseToken + "_" + normalizeTypeExprListKey(g, ifacePkg, ifaceArgs))
	}
	info := &nativeInterfaceInfo{
		Key:               key,
		GoType:            baseToken,
		TypeExpr:          expr,
		TypeString:        typeExpressionToString(expr),
		MarkerMethod:      baseToken + "_marker",
		ToRuntimeMethod:   baseToken + "_to_value",
		FromRuntimeHelper: baseToken + "_from_value",
		FromRuntimePanic:  baseToken + "_from_runtime_value_or_panic",
		ToRuntimeHelper:   baseToken + "_to_runtime_value",
		ToRuntimePanic:    baseToken + "_to_runtime_value_or_panic",
		RuntimeAdapter:    baseToken + "_runtime_adapter",
		RuntimeWrapHelper: baseToken + "_wrap_runtime",
	}
	g.nativeInterfaces[key] = info
	g.nativeInterfaceBuilding[key] = struct{}{}
	defer delete(g.nativeInterfaceBuilding, key)
	methodMap := make(map[string]*nativeInterfaceMethod)
	if !g.collectNativeInterfaceMethods(ifacePkg, expr, make(map[string]struct{}), methodMap) {
		delete(g.nativeInterfaces, key)
		return nil, false
	}
	if len(methodMap) == 0 {
		delete(g.nativeInterfaces, key)
		return nil, false
	}
	methodNames := make([]string, 0, len(methodMap))
	for name := range methodMap {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames)
	methods := make([]*nativeInterfaceMethod, 0, len(methodNames))
	for _, name := range methodNames {
		methods = append(methods, methodMap[name])
	}
	info.Methods = methods
	g.refreshNativeInterfaceAdapters(info)
	return info, true
}

func (g *generator) refreshNativeInterfaceAdapters(info *nativeInterfaceInfo) {
	if g == nil || info == nil {
		return
	}
	adapterMap := make(map[string]*nativeInterfaceAdapter)
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable || impl.ImplName != "" {
			continue
		}
		goType := ""
		if len(impl.Info.Params) > 0 {
			goType = impl.Info.Params[0].GoType
		}
		if goType == "" || goType == "runtime.Value" || goType == "any" {
			continue
		}
		complete := true
		methodImpls := make(map[string]*functionInfo, len(info.Methods))
		for _, method := range info.Methods {
			found := g.nativeInterfaceMethodImpl(goType, method)
			if found == nil {
				complete = false
				break
			}
			methodImpls[method.Name] = found
		}
		if !complete {
			continue
		}
		if _, exists := adapterMap[goType]; exists {
			continue
		}
		token := g.nativeUnionTypeToken(goType)
		adapterMap[goType] = &nativeInterfaceAdapter{
			GoType:      goType,
			Token:       token,
			AdapterType: info.GoType + "_adapter_" + token,
			WrapHelper:  info.GoType + "_wrap_" + token,
			Methods:     methodImpls,
		}
	}
	adapters := make([]*nativeInterfaceAdapter, 0, len(adapterMap))
	for _, adapter := range adapterMap {
		adapters = append(adapters, adapter)
	}
	sort.Slice(adapters, func(i, j int) bool {
		return adapters[i].GoType < adapters[j].GoType
	})
	info.Adapters = adapters
}

func (g *generator) nativeInterfaceAssignable(actual string, expected string) bool {
	if g == nil || actual == "" || expected == "" {
		return false
	}
	actualInfo := g.nativeInterfaceInfoForGoType(actual)
	expectedInfo := g.nativeInterfaceInfoForGoType(expected)
	if actualInfo == nil || expectedInfo == nil {
		return false
	}
	for _, expectedMethod := range expectedInfo.Methods {
		found := false
		for _, actualMethod := range actualInfo.Methods {
			if actualMethod == nil || expectedMethod == nil || actualMethod.Name != expectedMethod.Name {
				continue
			}
			if actualMethod.ReturnGoType != expectedMethod.ReturnGoType || actualMethod.OptionalLast != expectedMethod.OptionalLast || len(actualMethod.ParamGoTypes) != len(expectedMethod.ParamGoTypes) {
				continue
			}
			same := true
			for idx := range actualMethod.ParamGoTypes {
				if actualMethod.ParamGoTypes[idx] != expectedMethod.ParamGoTypes[idx] {
					same = false
					break
				}
			}
			if same {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (g *generator) nativeInterfaceAcceptsActual(info *nativeInterfaceInfo, actual string) bool {
	if g == nil || info == nil || actual == "" {
		return false
	}
	if actual == info.GoType || actual == "runtime.Value" {
		return true
	}
	if actual == "any" {
		return true
	}
	if g.nativeInterfaceAssignable(actual, info.GoType) {
		return true
	}
	_, ok := g.nativeInterfaceAdapterForActual(info, actual)
	return ok
}

func (g *generator) nativeInterfaceWrapLines(ctx *compileContext, expected string, actual string, expr string) ([]string, string, bool) {
	info := g.nativeInterfaceInfoForGoType(expected)
	if info == nil || ctx == nil || actual == "" || expr == "" {
		return nil, "", false
	}
	if actual == expected || g.nativeInterfaceAssignable(actual, expected) {
		return nil, expr, true
	}
	if adapter, ok := g.nativeInterfaceAdapterForActual(info, actual); ok {
		return nil, fmt.Sprintf("%s(%s)", adapter.WrapHelper, expr), true
	}
	if actual == "any" {
		valueTemp := ctx.newTemp()
		lines := []string{fmt.Sprintf("%s := __able_any_to_value(%s)", valueTemp, expr)}
		moreLines, converted, ok := g.nativeInterfaceWrapLines(ctx, expected, "runtime.Value", valueTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, moreLines...)
		return lines, converted, true
	}
	if actual == "runtime.Value" {
		valueTemp := ctx.newTemp()
		convertedTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s := %s", valueTemp, expr),
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convertedTemp, errTemp, info.FromRuntimeHelper, valueTemp),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convertedTemp, true
	}
	return nil, "", false
}
