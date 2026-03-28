package compiler

import (
	"fmt"
	"sort"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) collectImplDefinition(def *ast.ImplementationDefinition, mapper *TypeMapper, pkgName string) {
	if def == nil || mapper == nil {
		return
	}
	if def.InterfaceName == nil || def.InterfaceName.Name == "" {
		return
	}
	if def.TargetType == nil {
		return
	}
	targetType := normalizeTypeExprForPackage(g, pkgName, def.TargetType)
	if targetType == nil {
		targetType = def.TargetType
	}
	if g.implDefinitions == nil {
		g.implDefinitions = make([]*implDefinitionInfo, 0, 4)
	}
	g.implDefinitions = append(g.implDefinitions, &implDefinitionInfo{
		Definition: def,
		Package:    pkgName,
	})
	if len(def.Definitions) == 0 {
		return
	}
	if g.implMethodList == nil {
		g.implMethodList = make([]*implMethodInfo, 0, len(def.Definitions))
	}
	ifaceName := def.InterfaceName.Name
	targetDesc := typeExpressionToString(targetType)
	implName := ""
	if def.ImplName != nil {
		implName = def.ImplName.Name
	}
	var ifaceGenerics []*ast.GenericParameter
	if iface := g.interfaces[ifaceName]; iface != nil {
		ifaceGenerics = iface.GenericParams
	}
	interfaceBindings := g.implTypeBindings(ifaceName, ifaceGenerics, def.InterfaceArgs, targetType)
	for idx, fn := range def.Definitions {
		if fn == nil || fn.ID == nil || fn.ID.Name == "" {
			continue
		}
		methodName := fn.ID.Name
		info := &functionInfo{
			Name:        fmt.Sprintf("impl %s for %s.%s", ifaceName, targetDesc, methodName),
			Package:     pkgName,
			GoName:      g.mangler.unique(fmt.Sprintf("impl_%s_%s_%d", sanitizeIdent(ifaceName), sanitizeIdent(methodName), idx)),
			Definition:  fn,
			HasOriginal: false,
		}
		g.fillImplMethodInfo(info, mapper, targetType, interfaceBindings)
		implInfo := &implMethodInfo{
			InterfaceName:     ifaceName,
			InterfaceArgs:     def.InterfaceArgs,
			InterfaceGenerics: ifaceGenerics,
			TargetType:        targetType,
			ImplName:          implName,
			ImplGenerics:      def.GenericParams,
			WhereClause:       def.WhereClause,
			MethodName:        methodName,
			Info:              info,
			ImplDefinition:    def,
		}
		g.implMethodList = append(g.implMethodList, implInfo)
		g.touchNativeInterfaceAdapters()
		if g.implMethodByInfo != nil {
			g.implMethodByInfo[info] = implInfo
		}
		if g.implMethodsBySignature != nil {
			key := functionInfoSignatureKey(info)
			g.implMethodsBySignature[key] = append(g.implMethodsBySignature[key], implInfo)
		}
	}
}

func (g *generator) collectDefaultImplMethods() {
	if g == nil || len(g.implDefinitions) == 0 {
		return
	}
	for _, entry := range g.implDefinitions {
		if entry == nil || entry.Definition == nil {
			continue
		}
		def := entry.Definition
		if def.InterfaceName == nil || def.InterfaceName.Name == "" || def.TargetType == nil {
			continue
		}
		ifaceName := def.InterfaceName.Name
		iface := g.interfaces[ifaceName]
		if iface == nil || len(iface.Signatures) == 0 {
			continue
		}
		explicit := make(map[string]struct{}, len(def.Definitions))
		for _, fn := range def.Definitions {
			if fn == nil || fn.ID == nil || fn.ID.Name == "" {
				continue
			}
			explicit[fn.ID.Name] = struct{}{}
		}
		implName := ""
		if def.ImplName != nil {
			implName = def.ImplName.Name
		}
		pkgName := g.interfacePackages[ifaceName]
		if pkgName == "" {
			pkgName = entry.Package
		}
		mapper := NewTypeMapper(g, pkgName)
		targetType := normalizeTypeExprForPackage(g, entry.Package, def.TargetType)
		if targetType == nil {
			targetType = def.TargetType
		}
		if g.implMethodList == nil {
			g.implMethodList = make([]*implMethodInfo, 0, len(iface.Signatures))
		}
		for idx, sig := range iface.Signatures {
			if sig == nil || sig.Name == nil || sig.Name.Name == "" || sig.DefaultImpl == nil {
				continue
			}
			if _, ok := explicit[sig.Name.Name]; ok {
				continue
			}
			defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
			info := &functionInfo{
				Name:        fmt.Sprintf("impl %s for %s.%s", ifaceName, typeExpressionToString(targetType), sig.Name.Name),
				Package:     pkgName,
				GoName:      g.mangler.unique(fmt.Sprintf("impl_%s_%s_default_%d", sanitizeIdent(ifaceName), sanitizeIdent(sig.Name.Name), idx)),
				Definition:  defaultDef,
				HasOriginal: false,
			}
			interfaceBindings := g.implTypeBindings(ifaceName, iface.GenericParams, def.InterfaceArgs, targetType)
			g.fillImplMethodInfo(info, mapper, targetType, interfaceBindings)
			implInfo := &implMethodInfo{
				InterfaceName:     ifaceName,
				InterfaceArgs:     def.InterfaceArgs,
				InterfaceGenerics: iface.GenericParams,
				TargetType:        targetType,
				ImplName:          implName,
				IsDefault:         true,
				ImplGenerics:      def.GenericParams,
				WhereClause:       def.WhereClause,
				MethodName:        sig.Name.Name,
				Info:              info,
				ImplDefinition:    def,
			}
			g.implMethodList = append(g.implMethodList, implInfo)
			g.touchNativeInterfaceAdapters()
			if g.implMethodByInfo != nil {
				g.implMethodByInfo[info] = implInfo
			}
			if g.implMethodsBySignature != nil {
				key := functionInfoSignatureKey(info)
				g.implMethodsBySignature[key] = append(g.implMethodsBySignature[key], implInfo)
			}
		}
	}
}

func implInterfaceTypeBindings(params []*ast.GenericParameter, args []ast.TypeExpression) map[string]ast.TypeExpression {
	if len(params) == 0 || len(args) == 0 {
		return nil
	}
	bindings := make(map[string]ast.TypeExpression, len(args))
	for idx, gp := range params {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" || idx >= len(args) || args[idx] == nil {
			continue
		}
		bindings[gp.Name.Name] = args[idx]
	}
	if len(bindings) == 0 {
		return nil
	}
	return bindings
}

func (g *generator) implTypeBindings(interfaceName string, params []*ast.GenericParameter, args []ast.TypeExpression, target ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil {
		return implInterfaceTypeBindings(params, args)
	}
	bindings := implInterfaceTypeBindings(params, args)
	iface := g.interfaces[interfaceName]
	selfBindings := g.interfaceSelfTypeBindings(iface, target)
	if len(selfBindings) == 0 {
		return g.normalizeImplBindings(bindings)
	}
	if len(bindings) == 0 {
		return g.normalizeImplBindings(selfBindings)
	}
	merged := make(map[string]ast.TypeExpression, len(bindings)+len(selfBindings))
	for name, expr := range bindings {
		merged[name] = expr
	}
	for name, expr := range selfBindings {
		if _, exists := merged[name]; !exists && expr != nil {
			merged[name] = expr
		}
	}
	return g.normalizeImplBindings(merged)
}

func (g *generator) normalizeImplBindings(bindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if len(bindings) == 0 {
		return bindings
	}
	normalized := make(map[string]ast.TypeExpression, len(bindings))
	for name, expr := range bindings {
		if expr == nil {
			continue
		}
		normalized[name] = substituteTypeParams(expr, bindings)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func (g *generator) interfaceSelfTypeBindings(iface *ast.InterfaceDefinition, target ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || iface == nil || iface.SelfTypePattern == nil || target == nil {
		return nil
	}
	bindings := make(map[string]ast.TypeExpression)
	switch pattern := iface.SelfTypePattern.(type) {
	case *ast.SimpleTypeExpression:
		if pattern == nil || pattern.Name == nil || pattern.Name.Name == "" || pattern.Name.Name == "_" {
			return nil
		}
		switch pattern.Name.Name {
		case "Self", "SelfType":
		default:
			return nil
		}
		bindings[pattern.Name.Name] = target
	case *ast.GenericTypeExpression:
		if pattern == nil || pattern.Base == nil {
			return nil
		}
		base, ok := pattern.Base.(*ast.SimpleTypeExpression)
		if !ok || base == nil || base.Name == nil || base.Name.Name == "" || g.isConcreteTypeName(base.Name.Name) {
			return nil
		}
		targetBase := target
		var targetArgs []ast.TypeExpression
		if targetGeneric, ok := target.(*ast.GenericTypeExpression); ok && targetGeneric != nil {
			if targetGeneric.Base != nil {
				targetBase = targetGeneric.Base
			}
			targetArgs = targetGeneric.Arguments
		}
		bindings[base.Name.Name] = targetBase
		if len(targetArgs) == len(pattern.Arguments) {
			for idx, arg := range pattern.Arguments {
				simple, ok := arg.(*ast.SimpleTypeExpression)
				if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" || simple.Name.Name == "_" {
					continue
				}
				if g.isConcreteTypeName(simple.Name.Name) {
					continue
				}
				if targetArgs[idx] != nil {
					bindings[simple.Name.Name] = targetArgs[idx]
				}
			}
		}
	default:
		return nil
	}
	if len(bindings) == 0 {
		return nil
	}
	return bindings
}

func (g *generator) interfaceSelfBindingNames(iface *ast.InterfaceDefinition) map[string]struct{} {
	if g == nil || iface == nil || iface.SelfTypePattern == nil {
		return nil
	}
	names := make(map[string]struct{})
	addName := func(name string) {
		if name == "" || name == "_" || g.isConcreteTypeName(name) {
			return
		}
		names[name] = struct{}{}
	}
	switch pattern := iface.SelfTypePattern.(type) {
	case *ast.SimpleTypeExpression:
		if pattern != nil && pattern.Name != nil {
			addName(pattern.Name.Name)
		}
	case *ast.GenericTypeExpression:
		if pattern == nil {
			break
		}
		if base, ok := pattern.Base.(*ast.SimpleTypeExpression); ok && base != nil && base.Name != nil {
			addName(base.Name.Name)
		}
		for _, arg := range pattern.Arguments {
			simple, ok := arg.(*ast.SimpleTypeExpression)
			if !ok || simple == nil || simple.Name == nil {
				continue
			}
			addName(simple.Name.Name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func (g *generator) fillImplMethodInfo(info *functionInfo, mapper *TypeMapper, target ast.TypeExpression, interfaceBindings map[string]ast.TypeExpression) {
	if info == nil || info.Definition == nil || mapper == nil {
		return
	}
	g.invalidateFunctionDerivedInfo(info)
	def := info.Definition
	selfTarget := g.implSelfTargetType(info.Package, target, interfaceBindings)
	allBindings := g.mergeImplSelfTargetBindings(info.Package, target, selfTarget, interfaceBindings)
	if selfTarget != nil {
		if allBindings == nil {
			allBindings = make(map[string]ast.TypeExpression)
		}
		allBindings["Self"] = normalizeTypeExprForPackage(g, info.Package, selfTarget)
		if impl := g.implMethodByInfo[info]; impl != nil {
			if iface := g.interfaces[impl.InterfaceName]; iface != nil {
				for name, expr := range g.interfaceSelfTypeBindings(iface, selfTarget) {
					if expr == nil {
						continue
					}
					allBindings[name] = normalizeTypeExprForPackage(g, info.Package, expr)
				}
			}
		}
	}
	info.TypeBindings = cloneTypeBindings(allBindings)
	params := make([]paramInfo, 0, len(def.Params))
	supported := true
	if def.IsMethodShorthand {
		supported = false
	}
	for idx, param := range def.Params {
		if param == nil {
			supported = false
			continue
		}
		name := fmt.Sprintf("arg%d", idx)
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			name = ident.Name
		} else {
			supported = false
		}
		paramType := param.ParamType
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && (ident.Name == "self" || ident.Name == "Self") {
			paramType = selfTarget
		}
		if paramType == nil {
			if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil {
				if ident.Name == "self" || ident.Name == "Self" {
					paramType = selfTarget
				}
			}
		}
		paramType = resolveSelfTypeExpr(paramType, selfTarget)
		paramType = substituteTypeParams(paramType, allBindings)
		goType, ok := mapper.Map(paramType)
		goType, ok = g.recoverRepresentableCarrierType(info.Package, paramType, goType)
		if !ok {
			supported = false
		}
		params = append(params, paramInfo{
			Name:      name,
			GoName:    safeParamName(name, idx),
			GoType:    goType,
			TypeExpr:  paramType,
			Supported: ok,
		})
	}
	retExpr := resolveSelfTypeExpr(def.ReturnType, selfTarget)
	retExpr = substituteTypeParams(retExpr, allBindings)
	expectsSelf := methodDefinitionExpectsSelf(def)
	retType := ""
	ok := false
	if forcedType, forced := g.staticMethodNominalStructReturnType(info.Package, selfTarget, expectsSelf, retExpr); forced {
		retType = forcedType
		ok = true
	}
	if !ok {
		retType, ok = mapper.Map(retExpr)
		retType, ok = g.recoverRepresentableCarrierType(info.Package, retExpr, retType)
	}
	if !ok || retType == "" {
		supported = false
	}
	info.Params = params
	info.ReturnType = retType
	info.SupportedTypes = supported
	info.Arity = len(params)
	if !supported {
		info.Compileable = false
		info.Reason = "unsupported param or return type"
		info.Arity = -1
	}
}

func (g *generator) mergeImplSelfTargetBindings(pkgName string, target ast.TypeExpression, selfTarget ast.TypeExpression, bindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil {
		return bindings
	}
	merged := cloneTypeBindings(bindings)
	targetGeneric, ok := normalizeTypeExprForPackage(g, pkgName, target).(*ast.GenericTypeExpression)
	if !ok || targetGeneric == nil {
		return merged
	}
	selfGeneric, ok := normalizeTypeExprForPackage(g, pkgName, selfTarget).(*ast.GenericTypeExpression)
	if !ok || selfGeneric == nil || len(targetGeneric.Arguments) != len(selfGeneric.Arguments) {
		return merged
	}
	targetBase, ok := typeExprBaseName(targetGeneric.Base)
	if !ok || targetBase == "" {
		return merged
	}
	selfBase, ok := typeExprBaseName(selfGeneric.Base)
	if !ok || selfBase != targetBase {
		return merged
	}
	if merged == nil {
		merged = make(map[string]ast.TypeExpression, len(targetGeneric.Arguments))
	}
	for idx, arg := range targetGeneric.Arguments {
		simple, ok := arg.(*ast.SimpleTypeExpression)
		if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" || g.isConcreteTypeName(simple.Name.Name) {
			continue
		}
		if selfGeneric.Arguments[idx] == nil {
			continue
		}
		if _, exists := merged[simple.Name.Name]; exists {
			continue
		}
		merged[simple.Name.Name] = normalizeTypeExprForPackage(g, pkgName, selfGeneric.Arguments[idx])
	}
	return merged
}

func (g *generator) implSelfTargetType(pkgName string, target ast.TypeExpression, interfaceBindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || target == nil {
		return target
	}
	normalizedTarget := normalizeTypeExprForPackage(g, pkgName, target)
	if generic, ok := normalizedTarget.(*ast.GenericTypeExpression); ok && generic != nil && !g.typeExprHasWildcard(generic) {
		return normalizedTarget
	}
	if reconstructed := g.reconstructGenericStructTargetFromBindings(pkgName, target, interfaceBindings); reconstructed != nil {
		return reconstructed
	}
	return normalizedTarget
}

func (g *generator) reconstructGenericStructTargetFromBindings(pkgName string, target ast.TypeExpression, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || target == nil {
		return nil
	}
	normalizedTarget := normalizeTypeExprForPackage(g, pkgName, target)
	if normalizedTarget == nil {
		return nil
	}
	if generic, ok := normalizedTarget.(*ast.GenericTypeExpression); ok && generic != nil {
		return normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(generic, bindings))
	}
	info, ok := g.structInfoForTypeExpr(pkgName, normalizedTarget)
	if !ok || info == nil || info.Node == nil || len(info.Node.GenericParams) == 0 {
		return nil
	}
	available := make([]ast.TypeExpression, 0, len(bindings))
	for _, expr := range bindings {
		if expr == nil {
			continue
		}
		if _, ok := expr.(*ast.WildcardTypeExpression); ok {
			continue
		}
		available = append(available, normalizeTypeExprForPackage(g, pkgName, expr))
	}
	args := make([]ast.TypeExpression, 0, len(info.Node.GenericParams))
	for _, gp := range info.Node.GenericParams {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" {
			return nil
		}
		if expr := normalizeTypeExprForPackage(g, pkgName, bindings[gp.Name.Name]); expr != nil {
			args = append(args, expr)
			continue
		}
		if len(info.Node.GenericParams) == 1 && len(available) == 1 {
			args = append(args, available[0])
			continue
		}
		return nil
	}
	baseName, ok := typeExprBaseName(normalizedTarget)
	if !ok || baseName == "" {
		return nil
	}
	return normalizeTypeExprForPackage(g, pkgName, ast.NewGenericTypeExpression(ast.Ty(baseName), args))
}

func (g *generator) sortedImplMethodInfos() []*implMethodInfo {
	if g == nil || len(g.implMethodList) == 0 {
		return nil
	}
	infos := make([]*implMethodInfo, 0, len(g.implMethodList))
	infos = append(infos, g.implMethodList...)
	sort.Slice(infos, func(i, j int) bool {
		left := infos[i]
		right := infos[j]
		if left == nil || right == nil {
			return left != nil
		}
		if left.InterfaceName != right.InterfaceName {
			return left.InterfaceName < right.InterfaceName
		}
		if left.ImplName != right.ImplName {
			return left.ImplName < right.ImplName
		}
		if left.MethodName != right.MethodName {
			return left.MethodName < right.MethodName
		}
		if left.Info == nil || right.Info == nil {
			return left.Info != nil
		}
		return left.Info.GoName < right.Info.GoName
	})
	return infos
}

func (g *generator) implSiblingsForDefault(target *implMethodInfo) map[string]implSiblingInfo {
	if g == nil || target == nil {
		return nil
	}
	targetTypeStr := typeExpressionToString(target.TargetType)
	targetTypeName, targetTypeNameOK := g.methodTargetName(target.TargetType)
	siblings := make(map[string]implSiblingInfo)
	ambiguous := make(map[string]struct{})
	allowedInterfaces := g.interfaceFamilyNames(target.InterfaceName)
	for _, impl := range g.implMethodList {
		if impl == nil || impl.Info == nil || !impl.Info.Compileable {
			continue
		}
		if _, ok := allowedInterfaces[impl.InterfaceName]; !ok {
			continue
		}
		// For named impls, match by impl name; for unnamed, match by target type.
		if target.ImplName != "" {
			if impl.ImplName != target.ImplName {
				continue
			}
		}
		if implTypeName, ok := g.methodTargetName(impl.TargetType); targetTypeNameOK && ok {
			if implTypeName != targetTypeName {
				continue
			}
		} else if typeExpressionToString(impl.TargetType) != targetTypeStr {
			continue
		}
		if impl.MethodName == target.MethodName {
			continue
		}
		candidate := implSiblingInfo{
			GoName: impl.Info.GoName,
			Arity:  impl.Info.Arity,
			Info:   impl.Info,
		}
		if existing, ok := siblings[impl.MethodName]; ok && existing.Info != candidate.Info {
			delete(siblings, impl.MethodName)
			ambiguous[impl.MethodName] = struct{}{}
			continue
		}
		if _, blocked := ambiguous[impl.MethodName]; blocked {
			continue
		}
		siblings[impl.MethodName] = candidate
	}
	if len(siblings) == 0 {
		return nil
	}
	return siblings
}

func (g *generator) interfaceFamilyNames(interfaceName string) map[string]struct{} {
	if g == nil || interfaceName == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var visit func(string)
	visit = func(name string) {
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		iface := g.interfaces[name]
		if iface == nil {
			return
		}
		pkgName := g.interfacePackages[name]
		for _, baseExpr := range iface.BaseInterfaces {
			_, baseName, _, _, ok := interfaceExprInfo(g, pkgName, baseExpr)
			if !ok {
				continue
			}
			visit(baseName)
		}
	}
	visit(interfaceName)
	return seen
}
