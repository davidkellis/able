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
	ExpectsSelf       bool
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

func (g *generator) nativeInterfaceImplCandidateSpecificity(candidate nativeInterfaceImplCandidate) int {
	if g == nil || candidate.info == nil {
		return 0
	}
	score := 0
	for _, expr := range candidate.info.TypeBindings {
		if expr == nil {
			continue
		}
		score += 2
		if g.typeExprFullyBound(candidate.info.Package, normalizeTypeExprForPackage(g, candidate.info.Package, expr)) {
			score += 3
		}
	}
	if candidate.info.ReturnType != "" && candidate.info.ReturnType != "runtime.Value" && candidate.info.ReturnType != "any" {
		score += 2
	}
	for _, param := range candidate.info.Params {
		if param.GoType != "" && param.GoType != "runtime.Value" && param.GoType != "any" {
			score += 1
		}
		if param.TypeExpr != nil && g.typeExprFullyBound(candidate.info.Package, normalizeTypeExprForPackage(g, candidate.info.Package, param.TypeExpr)) {
			score += 1
		}
	}
	if candidate.impl != nil && candidate.impl.TargetType != nil {
		target := g.specializedImplTargetType(candidate.impl, candidate.info.TypeBindings)
		if target == nil {
			target = candidate.impl.TargetType
		}
		if target != nil && g.typeExprFullyBound(candidate.info.Package, normalizeTypeExprForPackage(g, candidate.info.Package, target)) {
			score += 5
		}
	}
	return score
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
	interfaceFullyBound := true
	if ifacePkg, _, _, _, ok := interfaceExprInfo(g, "", info.TypeExpr); ok {
		interfaceFullyBound = g.typeExprFullyBound(ifacePkg, info.TypeExpr)
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
	// Prime inherited/specialized impl metadata before probing method bodies.
	// Concrete nominal types like LinkedList<T> can satisfy a parent interface
	// via another native interface family (for example Enumerable -> Iterable),
	// but the specialized impl carrier may not be materialized until the first
	// concrete match walk.
	_ = g.nativeInterfaceConcreteActualMatches(info, actual)
	methodImpls := make(map[string]*nativeInterfaceAdapterMethod, len(info.Methods))
	for _, method := range info.Methods {
		found := g.nativeInterfaceMethodImpl(actual, method)
		if found == nil {
			return nil, false
		}
		found = g.canonicalBuiltinMappedDefaultAdapterMethod(actual, method, found)
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
	actualPkg := ""
	if structInfo := g.structInfoByGoName(actual); structInfo != nil {
		actualPkg = structInfo.Package
	}
	if interfaceFullyBound && !g.typeExprFullyBound(actualPkg, typeExpr) {
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

func (g *generator) nativeInterfaceConcreteActualMatches(info *nativeInterfaceInfo, actual string) bool {
	if g == nil || info == nil || actual == "" {
		return false
	}
	infoPkg, infoName, infoArgs, _, ok := interfaceExprInfo(g, "", info.TypeExpr)
	if !ok {
		return false
	}
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		fn := candidateInfo.info
		if impl == nil || fn == nil || impl.ImplName != "" {
			continue
		}
		actualInfo, bindings, ok := g.nativeInterfaceConcreteImplInfo(actual, impl, fn.TypeBindings)
		if !ok || actualInfo == nil {
			continue
		}
		actualExpr := g.implConcreteInterfaceExpr(impl, bindings)
		if actualExpr == nil {
			continue
		}
		genericNames := g.callableGenericNames(fn)
		if _, ok := g.nativeInterfaceImplBindingsForTarget(actualInfo.Package, actualExpr, genericNames, infoPkg, infoName, infoArgs, make(map[string]struct{})); ok {
			return true
		}
	}
	return false
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
	if info.AdapterVersion != g.nativeInterfaceAdapterVersion {
		g.refreshNativeInterfaceAdapters(info)
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
	if first != nil {
		if ident, ok := first.Name.(*ast.Identifier); ok && ident != nil && strings.EqualFold(ident.Name, "self") {
			return true
		}
	}
	if first == nil || first.ParamType == nil {
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
	cacheKey := normalizeTypeExprCacheKey(pkgName, expr)
	if cacheKey != "" {
		if cached, ok := g.normalizedTypeExprCache[cacheKey]; ok && cached != nil {
			return cached
		}
	}
	normalized := expr
	if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil {
		normalized = expanded
	}
	normalized = normalizeBuiltinSemanticTypeExprInPackage(g, pkgName, normalized)
	normalized = normalizeKernelBuiltinTypeExpr(normalized)
	normalized = normalizeCallableSyntaxTypeExpr(normalized)
	if cacheKey != "" && normalized != nil {
		g.normalizedTypeExprCache[cacheKey] = normalized
	}
	return normalized
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

func (g *generator) nativeInterfaceMethodImplSignature(impl *implMethodInfo, bindings map[string]ast.TypeExpression) ([]ast.TypeExpression, []string, ast.TypeExpression, string, bool, bool) {
	if g == nil || impl == nil || impl.Info == nil || impl.Info.Definition == nil {
		return nil, nil, nil, "", false, false
	}
	def := impl.Info.Definition
	mapper := NewTypeMapper(g, impl.Info.Package)
	expectsSelf := methodDefinitionExpectsSelf(def)
	if !expectsSelf && len(def.Params) > 0 && len(impl.Info.Params) > 0 {
		first := def.Params[0]
		if first != nil && first.ParamType != nil {
			firstType := resolveSelfTypeExpr(first.ParamType, impl.TargetType)
			firstType = normalizeTypeExprForPackage(g, impl.Info.Package, substituteTypeParams(firstType, bindings))
			if firstGoType, ok := mapper.Map(firstType); ok && firstGoType != "" {
				if recovered, ok := g.recoverRepresentableCarrierType(impl.Info.Package, firstType, firstGoType); ok && recovered != "" {
					firstGoType = recovered
				}
				if g.canCoerceStaticExpr(firstGoType, impl.Info.Params[0].GoType) {
					expectsSelf = true
				}
			}
		}
	}
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
		bindings, ok = g.nativeInterfaceMergeConcreteInfoBindings(info, bindings)
		if !ok {
			continue
		}
		if g.isNativeStructPointerType(goType) || info.Params[0].GoType != goType {
			info, bindings, ok = g.nativeInterfaceConcreteImplInfo(goType, impl, bindings)
			if !ok || info == nil || len(info.Params) == 0 || info.Params[0].GoType != goType {
				continue
			}
			bindings, ok = g.nativeInterfaceMergeConcreteInfoBindings(info, bindings)
			if !ok {
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
			paramWildcard := method.ParamGoTypes[idx] == "runtime.Value" || method.ParamGoTypes[idx] == "any"
			if paramWildcard {
				continue
			}
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
		returnWildcard := method.ReturnGoType == "runtime.Value" || method.ReturnGoType == "any"
		if !returnWildcard &&
			!g.canCoerceStaticExpr(method.ReturnGoType, returnGoType) &&
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
				if g.nativeInterfaceAdapterMethodSpecificity(candidate) >= g.nativeInterfaceAdapterMethodSpecificity(found) {
					found = candidate
				}
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
		if len(info.Params) == 0 {
			continue
		}
		bindings, ok := g.nativeInterfaceImplBindings(impl, method)
		if !ok {
			continue
		}
		bindings, ok = g.nativeInterfaceMergeConcreteInfoBindings(info, bindings)
		if !ok {
			continue
		}
		if g.isNativeStructPointerType(goType) || info.Params[0].GoType != goType {
			info, bindings, ok = g.nativeInterfaceConcreteImplInfo(goType, impl, bindings)
			if !ok || info == nil || len(info.Params) == 0 || info.Params[0].GoType != goType {
				continue
			}
			bindings, ok = g.nativeInterfaceMergeConcreteInfoBindings(info, bindings)
			if !ok {
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
			paramWildcard := method.ParamGoTypes[idx] == "runtime.Value" || method.ParamGoTypes[idx] == "any"
			if paramWildcard {
				continue
			}
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
		returnWildcard := method.ReturnGoType == "runtime.Value" || method.ReturnGoType == "any"
		if !returnWildcard &&
			!g.canCoerceStaticExpr(method.ReturnGoType, returnGoType) &&
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
				if g.nativeInterfaceAdapterMethodSpecificity(candidate) >= g.nativeInterfaceAdapterMethodSpecificity(found) {
					found = candidate
				}
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
	bindings := cloneTypeBindings(initialBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if iface := g.interfaces[impl.InterfaceName]; iface != nil {
		for name := range g.interfaceSelfBindingNames(iface) {
			delete(bindings, name)
		}
		if concreteTarget := g.specializedImplTargetTemplate(impl, bindings); concreteTarget != nil {
			if !g.mergeConcreteBindings(impl.Info.Package, bindings, g.interfaceSelfTypeBindings(iface, concreteTarget)) {
				return nil, nil, false
			}
		}
	}
	if receiverType := g.implReceiverGoType(impl); receiverType != "" && receiverType == goType {
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
	if !g.nominalTargetTypeExprCompatible(impl.Info.Package, actualExpr, targetTemplate) {
		return nil, nil, false
	}
	if !g.specializedTypeTemplateMatches(impl.Info.Package, targetTemplate, actualExpr, genericNames, bindings, make(map[string]struct{})) {
		return nil, nil, false
	}
	if !g.implConstraintsSatisfied(impl.Info.Package, impl, bindings) {
		return nil, nil, false
	}
	targetName, _ := typeExprBaseName(impl.TargetType)
	receiverType := g.implReceiverGoType(impl)
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
	currentCounts := [2]int{len(g.implMethodList), len(g.specializedFunctions)}
	if g.nativeInterfaceImplCandidateCache != nil && g.nativeInterfaceImplCandidateCounts == currentCounts {
		return g.nativeInterfaceImplCandidateCache
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
	sort.Slice(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]
		leftScore := g.nativeInterfaceImplCandidateSpecificity(left)
		rightScore := g.nativeInterfaceImplCandidateSpecificity(right)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		leftReceiver := ""
		rightReceiver := ""
		if left.info != nil && len(left.info.Params) > 0 {
			leftReceiver = left.info.Params[0].GoType
		}
		if right.info != nil && len(right.info.Params) > 0 {
			rightReceiver = right.info.Params[0].GoType
		}
		if leftReceiver != rightReceiver {
			return leftReceiver < rightReceiver
		}
		leftGoName := ""
		rightGoName := ""
		if left.info != nil {
			leftGoName = left.info.GoName
		}
		if right.info != nil {
			rightGoName = right.info.GoName
		}
		if leftGoName != rightGoName {
			return leftGoName < rightGoName
		}
		leftName := ""
		rightName := ""
		if left.info != nil {
			leftName = left.info.Name
		}
		if right.info != nil {
			rightName = right.info.Name
		}
		return leftName < rightName
	})
	g.nativeInterfaceImplCandidateCache = candidates
	g.nativeInterfaceImplCandidateCounts = currentCounts
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
	interfaceFullyBound := true
	if ifacePkg, _, _, _, ok := interfaceExprInfo(g, "", info.TypeExpr); ok {
		interfaceFullyBound = g.typeExprFullyBound(ifacePkg, info.TypeExpr)
	}
	adapterMap := make(map[string]*nativeInterfaceAdapter)
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		fn := candidateInfo.info
		if impl == nil || fn == nil || !g.nativeInterfaceDispatchCandidateEligible(fn) || impl.ImplName != "" {
			continue
		}
		g.refreshRepresentableFunctionInfo(fn)
		goType := ""
		if len(fn.Params) > 0 {
			goType = fn.Params[0].GoType
		}
		adapterTypeExpr := impl.TargetType
		if len(fn.Params) > 0 && fn.Params[0].TypeExpr != nil {
			adapterTypeExpr = fn.Params[0].TypeExpr
		}
		if (goType == "" || goType == "runtime.Value" || goType == "any") && adapterTypeExpr != nil {
			mapper := NewTypeMapper(g, fn.Package)
			if mapped, ok := mapper.Map(adapterTypeExpr); ok && mapped != "" && mapped != "runtime.Value" && mapped != "any" {
				goType = mapped
			}
		}
		if goType == "" || goType == "runtime.Value" || goType == "any" {
			continue
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
			var carrierBindings map[string]ast.TypeExpression
			if infoPkg, infoName, infoArgs, _, ok := interfaceExprInfo(g, fn.Package, info.TypeExpr); ok {
				genericNames := g.callableGenericNames(fn)
				var ok bool
				carrierBindings, ok = g.nativeInterfaceImplBindingsForTarget(
					fn.Package,
					carrier.typeExpr,
					genericNames,
					infoPkg,
					infoName,
					infoArgs,
					make(map[string]struct{}),
				)
				if !ok {
					continue
				}
			}
			if interfaceFullyBound && !g.typeExprFullyBound(fn.Package, carrier.typeExpr) {
				continue
			}
			if carrier.goType != goType {
				_, _, _ = g.nativeInterfaceConcreteImplInfo(carrier.goType, impl, carrierBindings)
			}
			complete := true
			methodImpls := make(map[string]*nativeInterfaceAdapterMethod, len(info.Methods))
			for _, method := range info.Methods {
				found := g.nativeInterfaceMethodImpl(carrier.goType, method)
				if found == nil {
					complete = false
					break
				}
				found = g.canonicalBuiltinMappedDefaultAdapterMethod(carrier.goType, method, found)
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
	genericNames := g.callableGenericNames(fn)
	bindings, ok := g.nativeInterfaceImplBindingsForTarget(
		fn.Package,
		actualExpr,
		genericNames,
		infoPkg,
		infoName,
		infoArgs,
		make(map[string]struct{}),
	)
	if !ok {
		return "", nil, false
	}
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression, len(fn.TypeBindings))
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
