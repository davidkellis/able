package compiler

import (
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type nativeInterfaceMethod struct {
	Name              string
	GoName            string
	InterfaceName     string
	InterfacePackage  string
	InterfaceArgs     []ast.TypeExpression
	ParamGoTypes      []string
	ParamTypeExprs    []ast.TypeExpression
	ReturnGoType      string
	ReturnTypeExpr    ast.TypeExpression
	OptionalLast      bool
	DefaultDefinition *ast.FunctionDefinition
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
	Key                    string
	GoType                 string
	TypeExpr               ast.TypeExpression
	TypeString             string
	AdapterVersion         int
	MarkerMethod           string
	ToRuntimeMethod        string
	FromRuntimeHelper      string
	FromRuntimePanic       string
	ToRuntimeHelper        string
	ToRuntimePanic         string
	ApplyRuntimeHelper     string
	RuntimeAdapter         string
	RuntimeWrapHelper      string
	RuntimeIteratorAdapter string
	Methods                []*nativeInterfaceMethod
	GenericMethods         []*nativeInterfaceGenericMethod
	Adapters               []*nativeInterfaceAdapter
}

type nativeInterfaceImplCandidate struct {
	impl *implMethodInfo
	info *functionInfo
}

func (g *generator) nativeInterfaceInfoForGoType(goType string) *nativeInterfaceInfo {
	if g == nil || goType == "" || g.nativeInterfaces == nil {
		return nil
	}
	for _, info := range g.nativeInterfaces {
		if info == nil {
			continue
		}
		match := info.GoType == goType || info.RuntimeAdapter == goType || info.RuntimeIteratorAdapter == goType
		if !match {
			for _, adapter := range info.Adapters {
				if adapter != nil && adapter.AdapterType == goType {
					match = true
					break
				}
			}
		}
		if match {
			if _, building := g.nativeInterfaceBuilding[info.Key]; building {
				return info
			}
			if _, refreshing := g.nativeInterfaceRefreshing[info.Key]; refreshing {
				return info
			}
			if g.nativeInterfaceRefreshAllowed() && info.AdapterVersion != g.nativeInterfaceAdapterVersion {
				g.refreshNativeInterfaceAdapters(info)
			}
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
	for _, adapter := range g.nativeInterfaceKnownAdapters(info) {
		if adapter != nil && adapter.GoType == actual {
			g.recordNativeInterfaceExplicitAdapter(info, adapter)
			return adapter, true
		}
	}
	if strings.HasPrefix(actual, "__able_iface_") {
		return nil, false
	}
	methodImpls := make(map[string]*nativeInterfaceAdapterMethod, len(info.Methods))
	for _, method := range info.Methods {
		found := g.nativeInterfaceMethodImpl(actual, method)
		if found == nil {
			return nil, false
		}
		methodImpls[method.Name] = found
	}
	if len(info.Methods) == 0 {
		for _, method := range info.GenericMethods {
			if !g.nativeInterfaceGenericMethodImplExists(actual, method) {
				return nil, false
			}
		}
	}
	typeExpr, ok := g.typeExprForGoType(actual)
	if !ok || typeExpr == nil {
		return nil, false
	}
	adapter := &nativeInterfaceAdapter{
		GoType:      actual,
		TypeExpr:    typeExpr,
		Token:       g.nativeUnionTypeToken(actual),
		AdapterType: info.GoType + "_adapter_" + g.nativeUnionTypeToken(actual),
		WrapHelper:  info.GoType + "_wrap_" + g.nativeUnionTypeToken(actual),
		Methods:     methodImpls,
	}
	g.recordNativeInterfaceExplicitAdapter(info, adapter)
	return adapter, true
}

func (g *generator) recordNativeInterfaceExplicitAdapter(info *nativeInterfaceInfo, adapter *nativeInterfaceAdapter) {
	if g == nil || info == nil || adapter == nil || adapter.GoType == "" {
		return
	}
	if g.nativeInterfaceExplicitAdapters == nil {
		g.nativeInterfaceExplicitAdapters = make(map[string]map[string]*nativeInterfaceAdapter)
	}
	adapters := g.nativeInterfaceExplicitAdapters[info.Key]
	if adapters == nil {
		adapters = make(map[string]*nativeInterfaceAdapter)
		g.nativeInterfaceExplicitAdapters[info.Key] = adapters
	}
	adapters[adapter.GoType] = adapter
	for _, existing := range info.Adapters {
		if existing != nil && existing.GoType == adapter.GoType {
			return
		}
	}
	info.Adapters = append(info.Adapters, adapter)
}

func (g *generator) nativeInterfaceKnownAdapters(info *nativeInterfaceInfo) []*nativeInterfaceAdapter {
	if g == nil || info == nil {
		return nil
	}
	adapterMap := make(map[string]*nativeInterfaceAdapter)
	for _, adapter := range info.Adapters {
		if adapter == nil || adapter.GoType == "" {
			continue
		}
		adapterMap[adapter.GoType] = adapter
	}
	if extra := g.nativeInterfaceExplicitAdapters[info.Key]; extra != nil {
		for goType, adapter := range extra {
			if adapter == nil || goType == "" {
				continue
			}
			if _, exists := adapterMap[goType]; exists {
				continue
			}
			adapterMap[goType] = adapter
		}
	}
	if len(adapterMap) == 0 {
		return nil
	}
	adapters := make([]*nativeInterfaceAdapter, 0, len(adapterMap))
	for _, adapter := range adapterMap {
		adapters = append(adapters, adapter)
	}
	sort.Slice(adapters, func(i, j int) bool {
		return adapters[i].GoType < adapters[j].GoType
	})
	return adapters
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
	normalized := expr
	if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil {
		normalized = expanded
	}
	return normalizeCallableSyntaxTypeExpr(normalizeKernelBuiltinTypeExpr(normalized))
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
	genericNames := nativeInterfaceGenericNameSet(impl.InterfaceGenerics)
	actualExpr := nativeInterfaceInstantiationExpr(impl.InterfaceName, impl.InterfaceArgs)
	return g.nativeInterfaceImplBindingsForTarget(
		actualPkg,
		actualExpr,
		genericNames,
		method.InterfacePackage,
		method.InterfaceName,
		method.InterfaceArgs,
		make(map[string]struct{}),
	)
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
		if sig.DefaultImpl != nil {
			method.DefaultDefinition = ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
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

func (g *generator) nativeInterfaceMethodImplSignature(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ([]ast.TypeExpression, []string, ast.TypeExpression, string, bool, bool) {
	if g == nil || impl == nil || impl.Info == nil || impl.Info.Definition == nil {
		return nil, nil, nil, "", false, false
	}
	def := impl.Info.Definition
	mapper := NewTypeMapper(g, impl.Info.Package)
	expectsSelf := methodDefinitionExpectsSelf(def)
	paramStart := 0
	if expectsSelf {
		paramStart = 1
	}
	paramTypeExprs := make([]ast.TypeExpression, 0, len(def.Params)-paramStart)
	paramGoTypes := make([]string, 0, len(def.Params)-paramStart)
	optionalLast := false
	for idx := paramStart; idx < len(def.Params); idx++ {
		param := def.Params[idx]
		if param == nil || param.ParamType == nil {
			return nil, nil, nil, "", false, false
		}
		paramType := resolveSelfTypeExpr(param.ParamType, impl.TargetType)
		paramType = normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(paramType, bindings))
		goType, ok := mapper.Map(paramType)
		if !ok || goType == "" {
			return nil, nil, nil, "", false, false
		}
		paramTypeExprs = append(paramTypeExprs, paramType)
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
		return nil, nil, nil, "", false, false
	}
	return paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, optionalLast, true
}

func (g *generator) nativeInterfaceMethodImpl(goType string, method *nativeInterfaceMethod) *nativeInterfaceAdapterMethod {
	if g == nil || method == nil || goType == "" {
		return nil
	}
	if exact := g.nativeInterfaceMethodImplExact(goType, method); exact != nil {
		return exact
	}
	var found *nativeInterfaceAdapterMethod
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || !info.Compileable || impl.ImplName != "" {
			continue
		}
		g.refreshRepresentableFunctionInfo(info)
		if impl.MethodName != method.Name {
			continue
		}
		if len(info.Params) == 0 {
			continue
		}
		bindings, ok := g.nativeInterfaceImplBindings(impl, method)
		if !ok {
			continue
		}
		if info.Params[0].GoType != goType {
			info, bindings, ok = g.nativeInterfaceConcreteImplInfo(goType, impl, bindings)
			if !ok || info == nil || len(info.Params) == 0 || info.Params[0].GoType != goType {
				continue
			}
		}
		paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, optionalLast, ok := g.nativeInterfaceMethodImplSignature(impl, bindings)
		if !ok || optionalLast != method.OptionalLast || len(paramGoTypes) != len(method.ParamGoTypes) {
			continue
		}
		matches := true
		leftVars := make(map[string]string)
		rightVars := make(map[string]string)
		for idx := range method.ParamGoTypes {
			if g.canCoerceStaticExpr(method.ParamGoTypes[idx], paramGoTypes[idx]) {
				continue
			}
			if !g.typeExprEquivalentModuloGenerics(method.ParamTypeExprs[idx], paramTypeExprs[idx], leftVars, rightVars) {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		if !g.canCoerceStaticExpr(method.ReturnGoType, returnGoType) &&
			!g.typeExprEquivalentModuloGenerics(method.ReturnTypeExpr, returnTypeExpr, leftVars, rightVars) {
			continue
		}
		candidate := &nativeInterfaceAdapterMethod{
			Info:                 info,
			CompiledReturnGoType: info.ReturnType,
			ParamGoTypes:         paramGoTypes,
			ReturnGoType:         returnGoType,
		}
		if len(info.Params) > 1 {
			candidate.CompiledParamGoTypes = make([]string, 0, len(info.Params)-1)
			for idx := 1; idx < len(info.Params); idx++ {
				candidate.CompiledParamGoTypes = append(candidate.CompiledParamGoTypes, info.Params[idx].GoType)
			}
		}
		if found != nil && found.Info != candidate.Info {
			if equivalentFunctionInfoSignature(found.Info, candidate.Info) {
				continue
			}
			return nil
		}
		found = candidate
	}
	return found
}

func (g *generator) nativeInterfaceMethodImplExactOnly(goType string, method *nativeInterfaceMethod) *nativeInterfaceAdapterMethod {
	if g == nil || method == nil || goType == "" {
		return nil
	}
	var found *nativeInterfaceAdapterMethod
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || !info.Compileable || impl.ImplName != "" {
			continue
		}
		g.refreshRepresentableFunctionInfo(info)
		if impl.MethodName != method.Name {
			continue
		}
		if len(info.Params) == 0 || info.Params[0].GoType != goType {
			continue
		}
		bindings, ok := g.nativeInterfaceImplBindings(impl, method)
		if !ok {
			continue
		}
		paramTypeExprs, paramGoTypes, returnTypeExpr, returnGoType, optionalLast, ok := g.nativeInterfaceMethodImplSignature(impl, bindings)
		if !ok || optionalLast != method.OptionalLast || len(paramGoTypes) != len(method.ParamGoTypes) {
			continue
		}
		matches := true
		leftVars := make(map[string]string)
		rightVars := make(map[string]string)
		for idx := range method.ParamGoTypes {
			if g.canCoerceStaticExpr(method.ParamGoTypes[idx], paramGoTypes[idx]) {
				continue
			}
			if !g.typeExprEquivalentModuloGenerics(method.ParamTypeExprs[idx], paramTypeExprs[idx], leftVars, rightVars) {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		if !g.canCoerceStaticExpr(method.ReturnGoType, returnGoType) &&
			!g.typeExprEquivalentModuloGenerics(method.ReturnTypeExpr, returnTypeExpr, leftVars, rightVars) {
			continue
		}
		candidate := &nativeInterfaceAdapterMethod{
			Info:                 info,
			CompiledReturnGoType: info.ReturnType,
			ParamGoTypes:         paramGoTypes,
			ReturnGoType:         returnGoType,
		}
		if len(info.Params) > 1 {
			candidate.CompiledParamGoTypes = make([]string, 0, len(info.Params)-1)
			for idx := 1; idx < len(info.Params); idx++ {
				candidate.CompiledParamGoTypes = append(candidate.CompiledParamGoTypes, info.Params[idx].GoType)
			}
		}
		if found != nil && found.Info != candidate.Info {
			if equivalentFunctionInfoSignature(found.Info, candidate.Info) {
				continue
			}
			return nil
		}
		found = candidate
	}
	return found
}

func (g *generator) nativeInterfaceMethodImplExact(goType string, method *nativeInterfaceMethod) *nativeInterfaceAdapterMethod {
	if g == nil || method == nil || goType == "" {
		return nil
	}
	found := g.nativeInterfaceMethodImplExactOnly(goType, method)
	if found == nil {
		return g.nativeInterfaceDefaultAdapterMethod(goType, method)
	}
	return found
}

func (g *generator) nativeInterfaceConcreteImplInfo(goType string, impl *implMethodInfo, initialBindings map[string]ast.TypeExpression) (*functionInfo, map[string]ast.TypeExpression, bool) {
	if g == nil || impl == nil || impl.Info == nil || goType == "" {
		return nil, nil, false
	}
	if g.structInfoByGoName(goType) == nil {
		return nil, nil, false
	}
	bindings := cloneTypeBindings(initialBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if len(impl.Info.Params) > 0 && impl.Info.Params[0].GoType == goType {
		return impl.Info, bindings, true
	}
	actualExpr, ok := g.typeExprForGoType(goType)
	if !ok || actualExpr == nil || impl.TargetType == nil {
		return nil, nil, false
	}
	genericNames := mergeGenericNameSets(nativeInterfaceGenericNameSet(impl.InterfaceGenerics), g.callableGenericNames(impl.Info))
	targetTemplate := g.specializedImplTargetTemplate(impl, bindings)
	if targetTemplate == nil {
		targetTemplate = impl.TargetType
	}
	if !g.specializedTypeTemplateMatches(impl.Info.Package, targetTemplate, actualExpr, genericNames, bindings, make(map[string]struct{})) {
		return nil, nil, false
	}
	targetName, _ := typeExprBaseName(impl.TargetType)
	receiverType := ""
	if len(impl.Info.Params) > 0 {
		receiverType = impl.Info.Params[0].GoType
	}
	method := &methodInfo{
		TargetName:   targetName,
		TargetType:   impl.TargetType,
		MethodName:   impl.MethodName,
		ReceiverType: receiverType,
		ExpectsSelf:  len(impl.Info.Params) > 0,
		Info:         impl.Info,
	}
	specialized, ok := g.ensureSpecializedImplMethod(method, impl, bindings)
	if !ok || specialized == nil || specialized.Info == nil || len(specialized.Info.Params) == 0 || specialized.Info.Params[0].GoType != goType {
		return nil, nil, false
	}
	g.refreshRepresentableFunctionInfo(specialized.Info)
	return specialized.Info, bindings, true
}

func (g *generator) nativeInterfaceImplCandidates() []nativeInterfaceImplCandidate {
	if g == nil {
		return nil
	}
	candidates := make([]nativeInterfaceImplCandidate, 0, len(g.implMethodList)+len(g.specializedFunctions))
	seen := make(map[*functionInfo]struct{}, len(g.implMethodList)+len(g.specializedFunctions))
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil {
			continue
		}
		if _, ok := seen[impl.Info]; ok {
			continue
		}
		seen[impl.Info] = struct{}{}
		candidates = append(candidates, nativeInterfaceImplCandidate{impl: impl, info: impl.Info})
	}
	for _, info := range g.specializedFunctions {
		if info == nil {
			continue
		}
		impl := g.implMethodByInfo[info]
		if impl == nil {
			continue
		}
		if _, ok := seen[info]; ok {
			continue
		}
		seen[info] = struct{}{}
		candidates = append(candidates, nativeInterfaceImplCandidate{impl: impl, info: info})
	}
	return candidates
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
		if g.nativeInterfaceRefreshAllowed() && info.AdapterVersion != g.nativeInterfaceAdapterVersion {
			g.refreshNativeInterfaceAdapters(info)
		}
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
	if ifaceName == "Iterator" {
		info.RuntimeIteratorAdapter = baseToken + "_runtime_iterator"
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
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		fn := candidateInfo.info
		if impl == nil || fn == nil || !fn.Compileable || impl.ImplName != "" {
			continue
		}
		g.refreshRepresentableFunctionInfo(fn)
		goType := ""
		if len(fn.Params) > 0 {
			goType = fn.Params[0].GoType
		}
		if goType == "" || goType == "runtime.Value" || goType == "any" {
			continue
		}
		adapterTypeExpr := impl.TargetType
		if len(fn.Params) > 0 && fn.Params[0].TypeExpr != nil {
			adapterTypeExpr = fn.Params[0].TypeExpr
		}
		carriers := []struct {
			goType   string
			typeExpr ast.TypeExpression
		}{
			{goType: goType, typeExpr: adapterTypeExpr},
		}
		if concreteGoType, concreteTypeExpr, ok := g.refreshNativeInterfaceAdapterConcreteTarget(info, impl, fn, goType, adapterTypeExpr); ok && concreteGoType != "" && concreteTypeExpr != nil && concreteGoType != goType {
			carriers = append(carriers, struct {
				goType   string
				typeExpr ast.TypeExpression
			}{goType: concreteGoType, typeExpr: concreteTypeExpr})
		}
		for _, carrier := range carriers {
			if carrier.goType == "" || carrier.typeExpr == nil {
				continue
			}
			if carrier.goType != goType {
				_, _, _ = g.nativeInterfaceConcreteImplInfo(carrier.goType, impl, fn.TypeBindings)
			}
			complete := true
			methodImpls := make(map[string]*nativeInterfaceAdapterMethod, len(info.Methods))
			for _, method := range info.Methods {
				found := g.nativeInterfaceMethodImplExactOnly(carrier.goType, method)
				if found == nil {
					complete = false
					break
				}
				methodImpls[method.Name] = found
			}
			if complete && len(info.Methods) == 0 {
				for _, method := range info.GenericMethods {
					if !g.nativeInterfaceGenericMethodImplExists(carrier.goType, method) {
						complete = false
						break
					}
				}
			}
			if !complete {
				continue
			}
			if _, exists := adapterMap[carrier.goType]; exists {
				continue
			}
			token := g.nativeUnionTypeToken(carrier.goType)
			adapter := &nativeInterfaceAdapter{
				GoType:      carrier.goType,
				TypeExpr:    carrier.typeExpr,
				Token:       token,
				AdapterType: info.GoType + "_adapter_" + token,
				WrapHelper:  info.GoType + "_wrap_" + token,
				Methods:     methodImpls,
			}
			adapterMap[carrier.goType] = adapter
			if len(info.Methods) == 0 {
				g.recordNativeInterfaceExplicitAdapter(info, adapter)
			}
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
	info.AdapterVersion = g.nativeInterfaceAdapterVersion
}

func (g *generator) refreshNativeInterfaceAdapterConcreteTarget(info *nativeInterfaceInfo, impl *implMethodInfo, fn *functionInfo, goType string, adapterTypeExpr ast.TypeExpression) (string, ast.TypeExpression, bool) {
	if g == nil || info == nil || impl == nil || fn == nil || goType == "" || adapterTypeExpr == nil {
		return "", nil, false
	}
	infoGenericNames := make(map[string]struct{})
	if baseName, ok := typeExprBaseName(info.TypeExpr); ok {
		if iface := g.interfaces[baseName]; iface != nil {
			infoGenericNames = addGenericParams(infoGenericNames, iface.GenericParams)
		}
	}
	receiverGenericNames := g.callableGenericNames(fn)
	if baseName, ok := typeExprBaseName(adapterTypeExpr); ok {
		if def, ok := g.structInfoForTypeName(fn.Package, baseName); ok && def != nil && def.Node != nil {
			receiverGenericNames = addGenericParams(receiverGenericNames, def.Node.GenericParams)
		}
	}
	if g.typeExprHasGeneric(info.TypeExpr, infoGenericNames) || !g.typeExprHasGeneric(adapterTypeExpr, receiverGenericNames) {
		return goType, adapterTypeExpr, true
	}
	infoPkg, infoName, infoArgs, _, ok := interfaceExprInfo(g, fn.Package, info.TypeExpr)
	if !ok {
		return "", nil, false
	}
	actualExpr := g.implConcreteInterfaceExpr(impl, fn.TypeBindings)
	if actualExpr == nil {
		return "", nil, false
	}
	bindings, ok := g.nativeInterfaceImplBindingsForTarget(
		fn.Package,
		actualExpr,
		genericParamNameSet(impl.InterfaceGenerics),
		infoPkg,
		infoName,
		infoArgs,
		make(map[string]struct{}),
	)
	if !ok {
		return "", nil, false
	}
	for name, expr := range fn.TypeBindings {
		if expr == nil {
			continue
		}
		bindings[name] = normalizeTypeExprForPackage(g, fn.Package, expr)
	}
	concreteTarget := g.specializedImplTargetType(impl, bindings)
	if concreteTarget == nil {
		return "", nil, false
	}
	concreteGoType, ok := NewTypeMapper(g, fn.Package).Map(concreteTarget)
	if !ok || concreteGoType == "" {
		return "", nil, false
	}
	return concreteGoType, concreteTarget, true
}
