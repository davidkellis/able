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
	TypeExpr    ast.TypeExpression
	Token       string
	AdapterType string
	WrapHelper  string
	Methods     map[string]*nativeInterfaceAdapterMethod
}

type nativeInterfaceAdapterMethod struct {
	Info                 *functionInfo
	CompiledParamGoTypes []string
	CompiledReturnGoType string
	ParamGoTypes         []string
	ReturnGoType         string
}

type nativeInterfaceInfo struct {
	Key                string
	GoType             string
	TypeExpr           ast.TypeExpression
	TypeString         string
	MarkerMethod       string
	ToRuntimeMethod    string
	FromRuntimeHelper  string
	FromRuntimePanic   string
	ToRuntimeHelper    string
	ToRuntimePanic     string
	ApplyRuntimeHelper string
	RuntimeAdapter     string
	RuntimeWrapHelper  string
	Methods            []*nativeInterfaceMethod
	GenericMethods     []*nativeInterfaceGenericMethod
	Adapters           []*nativeInterfaceAdapter
}

func (g *generator) nativeInterfaceInfoForGoType(goType string) *nativeInterfaceInfo {
	if g == nil || goType == "" || g.nativeInterfaces == nil {
		return nil
	}
	for _, info := range g.nativeInterfaces {
		if info != nil && info.GoType == goType {
			if _, building := g.nativeInterfaceBuilding[info.Key]; building {
				return info
			}
			if _, refreshing := g.nativeInterfaceRefreshing[info.Key]; refreshing {
				return info
			}
			g.refreshNativeInterfaceAdapters(info)
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

func nativeInterfaceGenericNameSet(params []*ast.GenericParameter) map[string]struct{} {
	if len(params) == 0 {
		return nil
	}
	names := make(map[string]struct{}, len(params))
	for _, param := range params {
		if param == nil || param.Name == nil || param.Name.Name == "" {
			continue
		}
		names[param.Name.Name] = struct{}{}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func (g *generator) nativeInterfaceTypeTemplateMatches(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = normalizeTypeExprForPackage(g, pkgName, template)
	actual = normalizeTypeExprForPackage(g, pkgName, actual)
	switch tt := template.(type) {
	case *ast.SimpleTypeExpression:
		if tt == nil || tt.Name == nil || tt.Name.Name == "" {
			return false
		}
		if _, ok := genericNames[tt.Name.Name]; ok {
			if bound, exists := bindings[tt.Name.Name]; exists {
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
		if !g.nativeInterfaceTypeTemplateMatches(pkgName, tt.Base, actualGeneric.Base, genericNames, bindings) {
			return false
		}
		for idx := range tt.Arguments {
			if !g.nativeInterfaceTypeTemplateMatches(pkgName, tt.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		actualNullable, ok := actual.(*ast.NullableTypeExpression)
		return ok && actualNullable != nil && g.nativeInterfaceTypeTemplateMatches(pkgName, tt.InnerType, actualNullable.InnerType, genericNames, bindings)
	case *ast.ResultTypeExpression:
		actualResult, ok := actual.(*ast.ResultTypeExpression)
		return ok && actualResult != nil && g.nativeInterfaceTypeTemplateMatches(pkgName, tt.InnerType, actualResult.InnerType, genericNames, bindings)
	case *ast.UnionTypeExpression:
		actualUnion, ok := actual.(*ast.UnionTypeExpression)
		if !ok || actualUnion == nil || len(tt.Members) != len(actualUnion.Members) {
			return false
		}
		for idx := range tt.Members {
			if !g.nativeInterfaceTypeTemplateMatches(pkgName, tt.Members[idx], actualUnion.Members[idx], genericNames, bindings) {
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
			if !g.nativeInterfaceTypeTemplateMatches(pkgName, tt.ParamTypes[idx], actualFn.ParamTypes[idx], genericNames, bindings) {
				return false
			}
		}
		return g.nativeInterfaceTypeTemplateMatches(pkgName, tt.ReturnType, actualFn.ReturnType, genericNames, bindings)
	default:
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
}

func (g *generator) nativeInterfaceImplBindings(impl *implMethodInfo, method *nativeInterfaceMethod) (map[string]ast.TypeExpression, bool) {
	if g == nil || impl == nil || method == nil {
		return nil, false
	}
	actualPkg := impl.Info.Package
	if actualPkg == "" {
		actualPkg = method.InterfacePackage
	}
	if len(impl.InterfaceArgs) != len(method.InterfaceArgs) {
		return nil, false
	}
	genericNames := nativeInterfaceGenericNameSet(impl.InterfaceGenerics)
	bindings := implInterfaceTypeBindings(impl.InterfaceGenerics, impl.InterfaceArgs)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression, len(method.InterfaceArgs))
	}
	for idx, template := range impl.InterfaceArgs {
		if !g.nativeInterfaceTypeTemplateMatches(actualPkg, template, method.InterfaceArgs[idx], genericNames, bindings) {
			return nil, false
		}
	}
	return bindings, true
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
			continue
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
				paramTypes = nil
				paramGoTypes = nil
				break
			}
			goType, ok := mapper.Map(substituted)
			if !ok || goType == "" {
				paramTypes = nil
				paramGoTypes = nil
				break
			}
			paramTypes = append(paramTypes, substituted)
			paramGoTypes = append(paramGoTypes, goType)
			if idx == len(sig.Params)-1 {
				if _, ok := substituted.(*ast.NullableTypeExpression); ok {
					optionalLast = true
				}
			}
		}
		if paramTypes == nil || paramGoTypes == nil {
			continue
		}
		returnExpr := normalizeTypeExprForPackage(g, ifacePkg, substituteTypeParams(sig.ReturnType, bindings))
		if typeExprUsesSelf(returnExpr) {
			continue
		}
		returnGoType, ok := mapper.Map(returnExpr)
		if !ok || returnGoType == "" {
			continue
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

func (g *generator) nativeInterfaceMethodImplSignature(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ([]string, string, bool, bool) {
	if g == nil || impl == nil || impl.Info == nil || impl.Info.Definition == nil {
		return nil, "", false, false
	}
	def := impl.Info.Definition
	mapper := NewTypeMapper(g, impl.Info.Package)
	expectsSelf := methodDefinitionExpectsSelf(def)
	paramStart := 0
	if expectsSelf {
		paramStart = 1
	}
	paramGoTypes := make([]string, 0, len(def.Params)-paramStart)
	optionalLast := false
	for idx := paramStart; idx < len(def.Params); idx++ {
		param := def.Params[idx]
		if param == nil || param.ParamType == nil {
			return nil, "", false, false
		}
		paramType := resolveSelfTypeExpr(param.ParamType, impl.TargetType)
		paramType = normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(paramType, bindings))
		goType, ok := mapper.Map(paramType)
		if !ok || goType == "" {
			return nil, "", false, false
		}
		paramGoTypes = append(paramGoTypes, goType)
		if idx == len(def.Params)-1 {
			if _, ok := paramType.(*ast.NullableTypeExpression); ok {
				optionalLast = true
			}
		}
	}
	returnTypeExpr := resolveSelfTypeExpr(def.ReturnType, impl.TargetType)
	returnTypeExpr = normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(returnTypeExpr, bindings))
	returnGoType, ok := mapper.Map(returnTypeExpr)
	if !ok || returnGoType == "" {
		return nil, "", false, false
	}
	return paramGoTypes, returnGoType, optionalLast, true
}

func (g *generator) nativeInterfaceMethodImpl(goType string, method *nativeInterfaceMethod) *nativeInterfaceAdapterMethod {
	if g == nil || method == nil || goType == "" {
		return nil
	}
	var found *nativeInterfaceAdapterMethod
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
		bindings, ok := g.nativeInterfaceImplBindings(impl, method)
		if !ok {
			continue
		}
		paramGoTypes, returnGoType, optionalLast, ok := g.nativeInterfaceMethodImplSignature(impl, bindings)
		if !ok || optionalLast != method.OptionalLast || len(paramGoTypes) != len(method.ParamGoTypes) {
			continue
		}
		matches := true
		for idx := range method.ParamGoTypes {
			if !g.canCoerceStaticExpr(paramGoTypes[idx], method.ParamGoTypes[idx]) {
				matches = false
				break
			}
		}
		if !matches || !g.canCoerceStaticExpr(method.ReturnGoType, returnGoType) {
			continue
		}
		candidate := &nativeInterfaceAdapterMethod{
			Info:                 impl.Info,
			CompiledReturnGoType: impl.Info.ReturnType,
			ParamGoTypes:         paramGoTypes,
			ReturnGoType:         returnGoType,
		}
		if len(impl.Info.Params) > 1 {
			candidate.CompiledParamGoTypes = make([]string, 0, len(impl.Info.Params)-1)
			for idx := 1; idx < len(impl.Info.Params); idx++ {
				candidate.CompiledParamGoTypes = append(candidate.CompiledParamGoTypes, impl.Info.Params[idx].GoType)
			}
		}
		if found != nil && found.Info != candidate.Info {
			return nil
		}
		found = candidate
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
		if _, refreshing := g.nativeInterfaceRefreshing[key]; refreshing {
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
		Key:                key,
		GoType:             baseToken,
		TypeExpr:           expr,
		TypeString:         typeExpressionToString(expr),
		MarkerMethod:       baseToken + "_marker",
		ToRuntimeMethod:    baseToken + "_to_value",
		FromRuntimeHelper:  baseToken + "_from_value",
		FromRuntimePanic:   baseToken + "_from_runtime_value_or_panic",
		ToRuntimeHelper:    baseToken + "_to_runtime_value",
		ToRuntimePanic:     baseToken + "_to_runtime_value_or_panic",
		ApplyRuntimeHelper: baseToken + "_apply_runtime_value",
		RuntimeAdapter:     baseToken + "_runtime_adapter",
		RuntimeWrapHelper:  baseToken + "_wrap_runtime",
	}
	g.nativeInterfaces[key] = info
	g.nativeInterfaceBuilding[key] = struct{}{}
	defer delete(g.nativeInterfaceBuilding, key)
	methodMap := make(map[string]*nativeInterfaceMethod)
	if !g.collectNativeInterfaceMethods(ifacePkg, expr, make(map[string]struct{}), methodMap) {
		delete(g.nativeInterfaces, key)
		return nil, false
	}
	genericMethodMap := make(map[string]*nativeInterfaceGenericMethod)
	if !g.collectNativeInterfaceGenericMethods(ifacePkg, expr, make(map[string]struct{}), genericMethodMap) {
		delete(g.nativeInterfaces, key)
		return nil, false
	}
	if len(methodMap) == 0 && len(genericMethodMap) == 0 {
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
	genericMethodNames := make([]string, 0, len(genericMethodMap))
	for name := range genericMethodMap {
		genericMethodNames = append(genericMethodNames, name)
	}
	sort.Strings(genericMethodNames)
	info.GenericMethods = make([]*nativeInterfaceGenericMethod, 0, len(genericMethodNames))
	for _, name := range genericMethodNames {
		info.GenericMethods = append(info.GenericMethods, genericMethodMap[name])
	}
	g.refreshNativeInterfaceAdapters(info)
	return info, true
}

func (g *generator) refreshNativeInterfaceAdapters(info *nativeInterfaceInfo) {
	if g == nil || info == nil {
		return
	}
	if _, refreshing := g.nativeInterfaceRefreshing[info.Key]; refreshing {
		return
	}
	g.nativeInterfaceRefreshing[info.Key] = struct{}{}
	defer delete(g.nativeInterfaceRefreshing, info.Key)
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
		methodImpls := make(map[string]*nativeInterfaceAdapterMethod, len(info.Methods))
		for _, method := range info.Methods {
			found := g.nativeInterfaceMethodImpl(goType, method)
			if found == nil {
				complete = false
				break
			}
			methodImpls[method.Name] = found
		}
		if complete {
			for _, method := range info.GenericMethods {
				if !g.nativeInterfaceGenericMethodImplExists(goType, method) {
					complete = false
					break
				}
			}
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
			TypeExpr:    impl.TargetType,
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
			if actualMethod.OptionalLast != expectedMethod.OptionalLast || len(actualMethod.ParamGoTypes) != len(expectedMethod.ParamGoTypes) {
				continue
			}
			if g.nativeInterfaceMethodShapeAssignable(actualMethod, expectedMethod) {
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

func (g *generator) nativeInterfaceMethodShapeAssignable(actualMethod, expectedMethod *nativeInterfaceMethod) bool {
	if g == nil || actualMethod == nil || expectedMethod == nil {
		return false
	}
	if actualMethod.ReturnGoType == expectedMethod.ReturnGoType && len(actualMethod.ParamGoTypes) == len(expectedMethod.ParamGoTypes) {
		same := true
		for idx := range actualMethod.ParamGoTypes {
			if actualMethod.ParamGoTypes[idx] != expectedMethod.ParamGoTypes[idx] {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	leftVars := make(map[string]string)
	rightVars := make(map[string]string)
	for idx := range actualMethod.ParamTypeExprs {
		if !g.typeExprEquivalentModuloGenerics(actualMethod.ParamTypeExprs[idx], expectedMethod.ParamTypeExprs[idx], leftVars, rightVars) {
			return false
		}
	}
	return g.typeExprEquivalentModuloGenerics(actualMethod.ReturnTypeExpr, expectedMethod.ReturnTypeExpr, leftVars, rightVars)
}

func (g *generator) typeExprEquivalentModuloGenerics(left ast.TypeExpression, right ast.TypeExpression, leftVars map[string]string, rightVars map[string]string) bool {
	switch l := left.(type) {
	case nil:
		return right == nil
	case *ast.SimpleTypeExpression:
		r, ok := right.(*ast.SimpleTypeExpression)
		if !ok || l == nil || r == nil || l.Name == nil || r.Name == nil {
			return false
		}
		return g.simpleTypeEquivalentModuloGenerics(l.Name.Name, r.Name.Name, leftVars, rightVars)
	case *ast.GenericTypeExpression:
		r, ok := right.(*ast.GenericTypeExpression)
		if !ok || l == nil || r == nil || len(l.Arguments) != len(r.Arguments) {
			return false
		}
		if !g.typeExprEquivalentModuloGenerics(l.Base, r.Base, leftVars, rightVars) {
			return false
		}
		for idx := range l.Arguments {
			if !g.typeExprEquivalentModuloGenerics(l.Arguments[idx], r.Arguments[idx], leftVars, rightVars) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		r, ok := right.(*ast.NullableTypeExpression)
		return ok && l != nil && r != nil && g.typeExprEquivalentModuloGenerics(l.InnerType, r.InnerType, leftVars, rightVars)
	case *ast.ResultTypeExpression:
		r, ok := right.(*ast.ResultTypeExpression)
		return ok && l != nil && r != nil && g.typeExprEquivalentModuloGenerics(l.InnerType, r.InnerType, leftVars, rightVars)
	case *ast.UnionTypeExpression:
		r, ok := right.(*ast.UnionTypeExpression)
		if !ok || l == nil || r == nil || len(l.Members) != len(r.Members) {
			return false
		}
		for idx := range l.Members {
			if !g.typeExprEquivalentModuloGenerics(l.Members[idx], r.Members[idx], leftVars, rightVars) {
				return false
			}
		}
		return true
	case *ast.FunctionTypeExpression:
		r, ok := right.(*ast.FunctionTypeExpression)
		if !ok || l == nil || r == nil || len(l.ParamTypes) != len(r.ParamTypes) {
			return false
		}
		for idx := range l.ParamTypes {
			if !g.typeExprEquivalentModuloGenerics(l.ParamTypes[idx], r.ParamTypes[idx], leftVars, rightVars) {
				return false
			}
		}
		return g.typeExprEquivalentModuloGenerics(l.ReturnType, r.ReturnType, leftVars, rightVars)
	default:
		return typeExpressionToString(left) == typeExpressionToString(right)
	}
}

func (g *generator) simpleTypeEquivalentModuloGenerics(leftName string, rightName string, leftVars map[string]string, rightVars map[string]string) bool {
	leftConcrete := g.isConcreteTypeName(leftName)
	rightConcrete := g.isConcreteTypeName(rightName)
	if leftConcrete || rightConcrete {
		return leftConcrete && rightConcrete && leftName == rightName
	}
	if mapped, ok := leftVars[leftName]; ok {
		return mapped == rightName
	}
	if mapped, ok := rightVars[rightName]; ok {
		return mapped == leftName
	}
	leftVars[leftName] = rightName
	rightVars[rightName] = leftName
	return true
}

func (g *generator) isConcreteTypeName(name string) bool {
	switch strings.TrimSpace(name) {
	case "", "bool", "Bool", "String", "string", "char", "Char", "i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64", "isize", "usize", "f32", "f64", "void", "Void", "Error", "Value", "nil":
		return true
	}
	if g == nil {
		return false
	}
	if g.isInterfaceName(name) {
		return true
	}
	if _, ok := g.structInfoByNameUnique(name); ok {
		return true
	}
	if _, ok := g.unionPackages[name]; ok {
		return true
	}
	for _, perPkg := range g.typeAliases {
		if perPkg == nil {
			continue
		}
		if _, ok := perPkg[name]; ok {
			return true
		}
	}
	return false
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
	if _, ok := g.nativeInterfaceAdapterForActual(info, actual); ok {
		return true
	}
	for _, adapter := range info.Adapters {
		if adapter == nil || adapter.GoType == "" {
			continue
		}
		union := g.nativeUnionInfoForGoType(adapter.GoType)
		if union == nil {
			continue
		}
		if _, ok := g.nativeUnionMember(union, actual); ok {
			return true
		}
		for _, member := range union.Members {
			if member == nil {
				continue
			}
			if iface := g.nativeInterfaceInfoForGoType(member.GoType); iface != nil && g.nativeInterfaceAcceptsActual(iface, actual) {
				return true
			}
			if member.GoType == "runtime.ErrorValue" && g.isNativeErrorCarrierType(actual) {
				return true
			}
			if g.nativeUnionRuntimeMemberAcceptsActual(member, actual) {
				return true
			}
		}
	}
	return false
}

func (g *generator) nativeInterfaceWrapLines(ctx *compileContext, expected string, actual string, expr string) ([]string, string, bool) {
	info := g.nativeInterfaceInfoForGoType(expected)
	if info == nil || ctx == nil || actual == "" || expr == "" {
		return nil, "", false
	}
	if actual == expected {
		return nil, expr, true
	}
	if g.nativeInterfaceAssignable(actual, expected) {
		actualInfo := g.nativeInterfaceInfoForGoType(actual)
		if actualInfo == nil {
			return nil, expr, true
		}
		runtimeTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		convertedTemp := ctx.newTemp()
		convertErrTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := []string{
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", runtimeTemp, errTemp, actualInfo.ToRuntimeHelper, expr),
			fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
		}
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines,
			fmt.Sprintf("%s, %s := %s(__able_runtime, %s)", convertedTemp, convertErrTemp, info.FromRuntimeHelper, runtimeTemp),
			fmt.Sprintf("%s = __able_control_from_error(%s)", controlTemp, convertErrTemp),
		)
		controlLines, ok = g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", false
		}
		lines = append(lines, controlLines...)
		return lines, convertedTemp, true
	}
	if adapter, ok := g.nativeInterfaceAdapterForActual(info, actual); ok {
		return nil, fmt.Sprintf("%s(%s)", adapter.WrapHelper, expr), true
	}
	for _, adapter := range info.Adapters {
		if adapter == nil || adapter.GoType == "" {
			continue
		}
		if g.nativeUnionInfoForGoType(adapter.GoType) == nil {
			continue
		}
		unionLines, unionExpr, ok := g.nativeUnionWrapLines(ctx, adapter.GoType, actual, expr)
		if !ok {
			continue
		}
		return unionLines, fmt.Sprintf("%s(%s)", adapter.WrapHelper, unionExpr), true
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
