package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) nativeInterfaceMergeConcreteInfoBindings(info *functionInfo, bindings map[string]ast.TypeExpression) (map[string]ast.TypeExpression, bool) {
	if g == nil || info == nil {
		return bindings, info != nil
	}
	if len(info.TypeBindings) == 0 {
		return cloneTypeBindings(bindings), true
	}
	merged := cloneTypeBindings(bindings)
	if merged == nil {
		merged = make(map[string]ast.TypeExpression, len(info.TypeBindings))
	}
	skipNames := g.nonTransferableConcreteInfoBindingNames(info)
	for name, expr := range info.TypeBindings {
		if expr == nil {
			continue
		}
		if _, skip := skipNames[name]; skip {
			continue
		}
		normalized := normalizeTypeExprForPackage(g, info.Package, expr)
		if existing, ok := merged[name]; ok && existing != nil {
			existing = normalizeTypeExprForPackage(g, info.Package, existing)
			if simple, ok := normalized.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == name {
				continue
			}
			if normalizeTypeExprString(g, info.Package, existing) != normalizeTypeExprString(g, info.Package, normalized) {
				return nil, false
			}
			continue
		}
		merged[name] = normalized
	}
	return merged, true
}

func (g *generator) nonTransferableConcreteInfoBindingNames(info *functionInfo) map[string]struct{} {
	skip := map[string]struct{}{
		"Self":     {},
		"SelfType": {},
	}
	if g == nil || info == nil || g.implMethodByInfo == nil {
		return skip
	}
	impl := g.implMethodByInfo[info]
	if impl == nil {
		return skip
	}
	if iface, _, ok := g.interfaceDefinitionForImpl(impl); ok && iface != nil {
		for name := range g.interfaceSelfBindingNames(iface) {
			skip[name] = struct{}{}
		}
	}
	return skip
}

func (g *generator) nativeInterfaceImplBindings(impl *implMethodInfo, method *nativeInterfaceMethod) (map[string]ast.TypeExpression, bool) {
	if g == nil || impl == nil || method == nil {
		return nil, false
	}
	actualPkg := impl.Info.Package
	if actualPkg == "" {
		actualPkg = method.InterfacePackage
	}
	genericNames := mergeGenericNameSets(nativeInterfaceGenericNameSet(impl.InterfaceGenerics), genericParamNameSet(impl.ImplGenerics))
	genericNames = mergeGenericNameSets(genericNames, g.callableGenericNames(impl.Info))
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
			goType, ok = g.recoverRepresentableCarrierType(ifacePkg, substituted, goType)
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
		returnGoType, ok = g.recoverRepresentableCarrierType(ifacePkg, returnExpr, returnGoType)
		if !ok || returnGoType == "" {
			continue
		}
		method := &nativeInterfaceMethod{
			Name:             sig.Name.Name,
			GoName:           sanitizeIdent(sig.Name.Name),
			InterfaceName:    ifaceName,
			InterfacePackage: ifacePkg,
			InterfaceArgs:    ifaceArgs,
			ExpectsSelf:      expectsSelf,
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
		if actualGeneric, ok := actual.(*ast.GenericTypeExpression); ok && actualGeneric != nil {
			if actualBase, ok := typeExprBaseName(actualGeneric.Base); ok && actualBase == tt.Name.Name {
				return true
			}
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
		if g.nativeInterfaceTypeTemplateMatches(pkgName, tt.ReturnType, actualFn.ReturnType, genericNames, bindings) {
			return true
		}
		return g.nativeInterfaceCallableReturnTemplateCompatible(pkgName, tt.ReturnType, actualFn.ReturnType, bindings)
	default:
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
}

func (g *generator) nativeInterfaceCallableReturnTemplateCompatible(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(template, bindings))
	actual = normalizeTypeExprForPackage(g, pkgName, substituteTypeParams(actual, bindings))
	templateGo, ok := g.lowerCarrierTypeInPackage(pkgName, template)
	if !ok || templateGo == "" {
		return false
	}
	actualGo, ok := g.lowerCarrierTypeInPackage(pkgName, actual)
	if !ok || actualGo == "" {
		return false
	}
	if g.isVoidType(templateGo) || templateGo == "runtime.Value" || templateGo == "any" {
		return true
	}
	return g.canCoerceStaticExpr(templateGo, actualGo)
}
