package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) specializationExpectedTypeExpr(ctx *compileContext, expected string) ast.TypeExpression {
	if g == nil {
		return nil
	}
	if ctx != nil && ctx.expectedTypeExpr != nil {
		return g.lowerNormalizedTypeExpr(ctx, ctx.expectedTypeExpr)
	}
	if expected == "" || expected == "runtime.Value" || expected == "any" {
		return nil
	}
	expectedExpr, ok := g.typeExprForGoType(expected)
	if !ok || expectedExpr == nil {
		return nil
	}
	return g.lowerNormalizedTypeExpr(ctx, expectedExpr)
}

func (g *generator) concreteFunctionCallInfo(ctx *compileContext, call *ast.FunctionCall, info *functionInfo, expected string) *functionInfo {
	if g == nil || ctx == nil || call == nil || info == nil || info.Definition == nil {
		return info
	}
	genericNames := g.callableGenericNames(info)
	if len(genericNames) == 0 {
		return info
	}
	bindings, ok := g.specializedFunctionBindings(ctx, call, info, expected)
	if !ok || len(bindings) == 0 {
		return info
	}
	specialized, ok := g.ensureSpecializedFunctionInfo(info, bindings)
	if !ok || specialized == nil {
		return info
	}
	return specialized
}

func (g *generator) specializedFunctionBindings(ctx *compileContext, call *ast.FunctionCall, info *functionInfo, expected string) (map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || info == nil || info.Definition == nil {
		return nil, false
	}
	genericNames := g.callableGenericNames(info)
	bindings := g.normalizeConcreteTypeBindings(info.Package, cloneTypeBindings(info.TypeBindings), genericNames)
	if bindings == nil {
		bindings = make(map[string]ast.TypeExpression)
	}
	if len(call.TypeArguments) > 0 {
		if len(call.TypeArguments) != len(info.Definition.GenericParams) {
			return nil, false
		}
		for idx, arg := range call.TypeArguments {
			gp := info.Definition.GenericParams[idx]
			if gp == nil || gp.Name == nil || gp.Name.Name == "" || arg == nil {
				return nil, false
			}
			bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, info.Package, arg)
		}
	}
	for idx, arg := range call.Arguments {
		if idx >= len(info.Params) {
			break
		}
		paramType := info.Params[idx].TypeExpr
		if paramType == nil {
			continue
		}
		actualExpr, ok := g.inferExpressionTypeExpr(ctx, arg, "")
		if !ok || actualExpr == nil {
			argCtx := ctx.child()
			_, _, argType, ok := g.compileExprLines(argCtx, arg, "")
			if !ok {
				continue
			}
			actualExpr, ok = g.typeExprForGoType(argType)
			if !ok || actualExpr == nil {
				continue
			}
		}
		g.specializedTypeTemplateMatches(info.Package, paramType, actualExpr, genericNames, bindings, make(map[string]struct{}))
	}
	if expectedExpr := g.specializationExpectedTypeExpr(ctx, expected); expectedExpr != nil {
		g.specializedTypeTemplateMatches(info.Package, info.Definition.ReturnType, expectedExpr, genericNames, bindings, make(map[string]struct{}))
	}
	bindings = g.normalizeConcreteTypeBindings(info.Package, bindings, genericNames)
	if len(bindings) == 0 {
		return nil, false
	}
	return bindings, true
}

func (g *generator) ensureSpecializedFunctionInfo(info *functionInfo, bindings map[string]ast.TypeExpression) (*functionInfo, bool) {
	if g == nil || info == nil || info.Definition == nil || len(bindings) == 0 {
		return nil, false
	}
	key := g.specializedImplFunctionKey(info, bindings)
	if existing, ok := g.specializedFunctionIndex[key]; ok && existing != nil && existing.Compileable {
		return existing, true
	}
	specialized := &functionInfo{
		Name:           info.Name,
		Package:        info.Package,
		QualifiedName:  info.QualifiedName,
		GoName:         g.mangler.unique(info.GoName + "_spec"),
		TypeBindings:   cloneTypeBindings(bindings),
		Definition:     info.Definition,
		HasOriginal:    info.HasOriginal,
		InternalOnly:   true,
		SupportedTypes: info.SupportedTypes,
	}
	mapper := NewTypeMapper(g, specialized.Package)
	g.fillSpecializedFunctionInfo(specialized, mapper)
	if !specialized.SupportedTypes {
		return nil, false
	}
	specialized.Compileable = true
	g.specializedFunctions = append(g.specializedFunctions, specialized)
	g.specializedFunctionIndex[key] = specialized
	if !g.bodyCompileable(specialized, specialized.ReturnType) {
		delete(g.specializedFunctionIndex, key)
		g.specializedFunctions = removeSpecializedFunction(g.specializedFunctions, specialized)
		return nil, false
	}
	specialized.Compileable = true
	specialized.Reason = ""
	return specialized, true
}

func (g *generator) fillSpecializedFunctionInfo(info *functionInfo, mapper *TypeMapper) {
	if info == nil || info.Definition == nil || mapper == nil {
		return
	}
	def := info.Definition
	params := make([]paramInfo, 0, len(def.Params))
	supported := true
	if def.IsMethodShorthand {
		supported = false
	}
	for idx, param := range def.Params {
		name := "arg"
		if ident, ok := param.Name.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			name = ident.Name
		} else {
			name = safeParamName(name, idx)
			supported = false
		}
		paramType := normalizeTypeExprForPackage(g, info.Package, substituteTypeParams(param.ParamType, info.TypeBindings))
		goType, ok := mapper.Map(paramType)
		if !ok || goType == "" {
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
	retExpr := normalizeTypeExprForPackage(g, info.Package, substituteTypeParams(def.ReturnType, info.TypeBindings))
	retType, ok := mapper.Map(retExpr)
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
