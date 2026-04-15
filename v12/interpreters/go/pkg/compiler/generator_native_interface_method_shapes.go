package compiler

import "able/interpreter-go/pkg/ast"

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
		goType, ok = g.recoverRepresentableCarrierType(impl.Info.Package, paramType, goType)
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
	returnGoType, ok = g.recoverRepresentableCarrierType(impl.Info.Package, returnTypeExpr, returnGoType)
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
		bindings, ok := g.nativeInterfaceImplBindings(impl, method)
		if !ok {
			continue
		}
		bindings, ok = g.nativeInterfaceMergeConcreteInfoBindings(info, bindings)
		if !ok {
			continue
		}
		expectsSelf := g.nativeInterfaceImplExpectsSelf(info, impl)
		if witnessGoType := g.nativeInterfaceImplWitnessGoType(info, impl, bindings); witnessGoType != goType {
			info, bindings, ok = g.nativeInterfaceConcreteImplInfo(goType, impl, bindings)
			if !ok || info == nil || g.nativeInterfaceImplWitnessGoType(info, impl, bindings) != goType {
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
		candidate.CompiledParamGoTypes = g.nativeInterfaceCompiledParamGoTypes(info, expectsSelf)
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
		bindings, ok := g.nativeInterfaceImplBindings(impl, method)
		if !ok {
			continue
		}
		bindings, ok = g.nativeInterfaceMergeConcreteInfoBindings(info, bindings)
		if !ok {
			continue
		}
		expectsSelf := g.nativeInterfaceImplExpectsSelf(info, impl)
		if witnessGoType := g.nativeInterfaceImplWitnessGoType(info, impl, bindings); witnessGoType != goType {
			info, bindings, ok = g.nativeInterfaceConcreteImplInfo(goType, impl, bindings)
			if !ok || info == nil || g.nativeInterfaceImplWitnessGoType(info, impl, bindings) != goType {
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
		candidate.CompiledParamGoTypes = g.nativeInterfaceCompiledParamGoTypes(info, expectsSelf)
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
