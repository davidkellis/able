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
	PackageName            string
	TypeExpr               ast.TypeExpression
	TypeString             string
	AdapterVersion         int
	MarkerMethod           string
	ToRuntimeMethod        string
	FromRuntimeHelper      string
	TryFromRuntimeHelper   string
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
	impl        *implMethodInfo
	info        *functionInfo
	specificity int
	receiver    string
	goName      string
	name        string
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
	return g.nativeInterfaceAdapterForActualSeen(info, actual, make(map[string]struct{}))
}

func (g *generator) nativeInterfaceAdapterForActualSeen(info *nativeInterfaceInfo, actual string, seen map[string]struct{}) (*nativeInterfaceAdapter, bool) {
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
	_ = g.nativeInterfaceConcreteActualMatchesSeen(info, actual, seen)
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
	return g.nativeInterfaceConcreteActualMatchesSeen(info, actual, make(map[string]struct{}))
}

func (g *generator) nativeInterfaceConcreteActualMatchesSeen(info *nativeInterfaceInfo, actual string, seen map[string]struct{}) bool {
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
		actualInfo, bindings, ok := g.nativeInterfaceConcreteImplInfoSeen(actual, impl, fn.TypeBindings, seen)
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
	if g.nativeInterfaceRefreshAllowed() && (info.AdapterVersion != g.nativeInterfaceAdapterVersion ||
		(len(info.Adapters) == 0 && len(g.nativeInterfaceExplicitAdapters[info.Key]) == 0)) {
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
	_, normalized := g.normalizeTypeExprContextForPackage(pkgName, expr)
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
	exprPkg, normalized := g.normalizeTypeExprContextForPackage(pkgName, expr)
	expr = normalized
	baseName, ok := typeExprBaseName(expr)
	if !ok || baseName == "" || !g.isInterfaceName(baseName) || baseName == "Error" {
		return "", "", nil, nil, false
	}
	iface, ifacePkg, ok := g.interfaceDefinitionForPackage(exprPkg, baseName)
	if !ok || iface == nil {
		return "", "", nil, nil, false
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
	return g.nativeInterfaceConcreteImplInfoSeen(goType, impl, initialBindings, make(map[string]struct{}))
}

func (g *generator) nativeInterfaceConcreteImplInfoSeen(goType string, impl *implMethodInfo, initialBindings map[string]ast.TypeExpression, seen map[string]struct{}) (*functionInfo, map[string]ast.TypeExpression, bool) {
	if g == nil || impl == nil || impl.Info == nil || goType == "" {
		return nil, nil, false
	}
	bindings := cloneTypeBindings(initialBindings)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if iface, _, ok := g.interfaceDefinitionForPackage(impl.Info.Package, impl.InterfaceName); ok && iface != nil {
		delete(bindings, "Self")
		delete(bindings, "SelfType")
		for name := range g.interfaceSelfBindingNames(iface) {
			delete(bindings, name)
		}
	}
	if witnessGoType := g.nativeInterfaceImplWitnessGoType(impl.Info, impl, bindings); witnessGoType != "" && witnessGoType == goType {
		return impl.Info, bindings, true
	}
	actualExpr, ok := g.typeExprForGoType(goType)
	if !ok || actualExpr == nil || impl.TargetType == nil {
		return nil, nil, false
	}
	actualExpr = normalizeTypeExprForPackage(g, impl.Info.Package, actualExpr)
	bindings["Self"] = actualExpr
	if iface, _, ok := g.interfaceDefinitionForPackage(impl.Info.Package, impl.InterfaceName); ok && iface != nil {
		if !g.mergeConcreteBindings(impl.Info.Package, bindings, g.interfaceSelfTypeBindings(iface, actualExpr)) {
			return nil, nil, false
		}
	}
	if !g.bindNominalTargetActualArgs(impl.Info.Package, impl.TargetType, impl.InterfaceArgs, actualExpr, bindings) {
		// keep existing bindings when the target is not a constructor-style
		// nominal generic; the normal template matcher handles those cases
	}
	genericNames := mergeGenericNameSets(nativeInterfaceGenericNameSet(impl.InterfaceGenerics), g.callableGenericNames(impl.Info))
	targetTemplate := g.specializedImplTargetTemplate(impl, bindings)
	if targetTemplate == nil {
		targetTemplate = impl.TargetType
	}
	if g.usesNominalStructCarrier(impl.Info.Package, targetTemplate) &&
		!g.nominalTargetTypeExprCompatible(impl.Info.Package, actualExpr, targetTemplate) {
		return nil, nil, false
	}
	if !g.specializedTypeTemplateMatches(impl.Info.Package, targetTemplate, actualExpr, genericNames, bindings, make(map[string]struct{})) {
		return nil, nil, false
	}
	if !g.implConstraintsSatisfiedSeen(impl.Info.Package, impl, bindings, seen) {
		return nil, nil, false
	}
	targetName, _ := typeExprBaseName(impl.TargetType)
	expectsSelf := g.nativeInterfaceImplExpectsSelf(impl.Info, impl)
	receiverType := ""
	if expectsSelf {
		receiverType = g.implReceiverGoType(impl)
	}
	method := &methodInfo{
		TargetName:   targetName,
		TargetType:   impl.TargetType,
		MethodName:   impl.MethodName,
		ReceiverType: receiverType,
		ExpectsSelf:  expectsSelf,
		Info:         impl.Info,
	}
	specialized, ok := g.ensureSpecializedImplMethod(method, impl, bindings)
	if !ok || specialized == nil || specialized.Info == nil {
		return nil, nil, false
	}
	if g.nativeInterfaceImplWitnessGoType(specialized.Info, impl, bindings) != goType {
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
	if g.nativeInterfaceImplCandidateCache != nil &&
		g.nativeInterfaceImplCandidateCounts[0] == currentCounts[0] &&
		g.nativeInterfaceImplCandidateCounts[1] >= 0 &&
		g.nativeInterfaceImplCandidateCounts[1] < currentCounts[1] {
		seen := make(map[*functionInfo]struct{}, len(g.nativeInterfaceImplCandidateCache)+currentCounts[1]-g.nativeInterfaceImplCandidateCounts[1])
		for _, candidate := range g.nativeInterfaceImplCandidateCache {
			if candidate.info != nil {
				seen[candidate.info] = struct{}{}
			}
		}
		delta := make([]nativeInterfaceImplCandidate, 0, currentCounts[1]-g.nativeInterfaceImplCandidateCounts[1])
		for _, info := range g.specializedFunctions[g.nativeInterfaceImplCandidateCounts[1]:] {
			candidate, ok := g.nativeInterfaceNewImplCandidate(g.implMethodByInfo[info], info)
			if !ok {
				continue
			}
			if _, exists := seen[candidate.info]; exists {
				continue
			}
			seen[candidate.info] = struct{}{}
			delta = append(delta, candidate)
		}
		if len(delta) > 0 {
			sort.Slice(delta, func(i, j int) bool {
				return nativeInterfaceImplCandidateLess(delta[i], delta[j])
			})
			g.nativeInterfaceImplCandidateCache = mergeNativeInterfaceImplCandidates(g.nativeInterfaceImplCandidateCache, delta)
		}
		g.nativeInterfaceImplCandidateCounts = currentCounts
		return g.nativeInterfaceImplCandidateCache
	}

	candidates := make([]nativeInterfaceImplCandidate, 0, len(g.implMethodList)+len(g.specializedFunctions))
	seen := make(map[*functionInfo]struct{}, len(g.implMethodList)+len(g.specializedFunctions))
	for _, impl := range g.implMethodList {
		candidate, ok := g.nativeInterfaceNewImplCandidate(impl, nil)
		if !ok {
			continue
		}
		if _, exists := seen[candidate.info]; exists {
			continue
		}
		seen[candidate.info] = struct{}{}
		candidates = append(candidates, candidate)
	}
	for _, info := range g.specializedFunctions {
		candidate, ok := g.nativeInterfaceNewImplCandidate(g.implMethodByInfo[info], info)
		if !ok {
			continue
		}
		if _, exists := seen[candidate.info]; exists {
			continue
		}
		seen[candidate.info] = struct{}{}
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return nativeInterfaceImplCandidateLess(candidates[i], candidates[j])
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
	keyParts := []string{ifacePkg, ifaceName}
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
	baseToken = g.mangler.unique(baseToken)
	info := &nativeInterfaceInfo{
		Key:                  key,
		GoType:               baseToken,
		PackageName:          ifacePkg,
		TypeExpr:             g.recordResolvedTypeExprPackage(expr, ifacePkg),
		TypeString:           typeExpressionToString(expr),
		MarkerMethod:         baseToken + "_marker",
		ToRuntimeMethod:      baseToken + "_to_value",
		FromRuntimeHelper:    baseToken + "_from_value",
		TryFromRuntimeHelper: baseToken + "_try_from_value",
		FromRuntimePanic:     baseToken + "_from_runtime_value_or_panic",
		ToRuntimeHelper:      baseToken + "_to_runtime_value",
		ToRuntimePanic:       baseToken + "_to_runtime_value_or_panic",
		ApplyRuntimeHelper:   baseToken + "_apply_runtime_value",
		RuntimeAdapter:       baseToken + "_runtime_adapter",
		RuntimeWrapHelper:    baseToken + "_wrap_runtime",
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
		goType := g.nativeInterfaceImplWitnessGoType(fn, impl, fn.TypeBindings)
		adapterTypeExpr := g.nativeInterfaceImplTargetExpr(fn, impl, fn.TypeBindings)
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
			carrierBindings := cloneTypeBindings(fn.TypeBindings)
			if infoPkg, infoName, infoArgs, _, ok := interfaceExprInfo(g, fn.Package, info.TypeExpr); ok {
				genericNames := g.callableGenericNames(fn)
				actualIfaceExpr := nativeInterfaceInstantiationExpr(impl.InterfaceName, impl.InterfaceArgs)
				matchedBindings, ok := g.nativeInterfaceImplBindingsForTarget(
					fn.Package,
					actualIfaceExpr,
					genericNames,
					infoPkg,
					infoName,
					infoArgs,
					make(map[string]struct{}),
				)
				if !ok {
					continue
				}
				if carrierBindings == nil {
					carrierBindings = make(map[string]ast.TypeExpression, len(matchedBindings))
				}
				if !g.mergeConcreteBindings(fn.Package, carrierBindings, matchedBindings) {
					continue
				}
			}
			if interfaceFullyBound && !g.typeExprFullyBound(fn.Package, carrier.typeExpr) {
				continue
			}
			if carrier.goType != goType {
				_, concreteBindings, ok := g.nativeInterfaceConcreteImplInfo(carrier.goType, impl, carrierBindings)
				if ok && len(concreteBindings) > 0 {
					carrierBindings = concreteBindings
				}
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
		if iface, _, ok := g.interfaceDefinitionForPackage(info.PackageName, baseName); ok && iface != nil {
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
	_ = g.bindNominalTargetActualArgs(fn.Package, impl.TargetType, impl.InterfaceArgs, actualExpr, bindings)
	concreteTarget := g.specializedImplTargetType(impl, bindings)
	if concreteTarget == nil {
		concreteTarget = g.reconstructGenericStructTargetFromBindings(fn.Package, impl.TargetType, bindings)
	}
	if concreteTarget == nil {
		return "", nil, false
	}
	concreteGoType, ok := NewTypeMapper(g, fn.Package).Map(concreteTarget)
	if !ok || concreteGoType == "" {
		return "", nil, false
	}
	return concreteGoType, concreteTarget, true
}
