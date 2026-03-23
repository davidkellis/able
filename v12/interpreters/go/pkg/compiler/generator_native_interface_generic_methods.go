package compiler

import (
	"strings"

	"able/interpreter-go/pkg/ast"
)

type nativeInterfaceGenericMethod struct {
	Name             string
	GoName           string
	InterfaceName    string
	InterfacePackage string
	InterfaceArgs    []ast.TypeExpression
	GenericParams    []*ast.GenericParameter
	WhereClause      []*ast.WhereClauseConstraint
	ParamTypeExprs   []ast.TypeExpression
	ReturnTypeExpr   ast.TypeExpression
	DefaultDefinition *ast.FunctionDefinition
}

func (g *generator) nativeInterfaceGenericMethodForGoType(goType string, methodName string) (*nativeInterfaceGenericMethod, bool) {
	info := g.nativeInterfaceInfoForGoType(goType)
	if info == nil || methodName == "" {
		return nil, false
	}
	for _, method := range info.GenericMethods {
		if method != nil && method.Name == methodName {
			return method, true
		}
	}
	return nil, false
}

func (g *generator) collectNativeInterfaceGenericMethods(pkgName string, expr ast.TypeExpression, seen map[string]struct{}, methods map[string]*nativeInterfaceGenericMethod) bool {
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
		if !g.collectNativeInterfaceGenericMethods(ifacePkg, next, seen, methods) {
			return false
		}
	}
	for _, sig := range ifaceDef.Signatures {
		if sig == nil || sig.Name == nil || sig.Name.Name == "" {
			return false
		}
		if len(sig.GenericParams) == 0 && len(sig.WhereClause) == 0 {
			continue
		}
		expectsSelf := functionSignatureExpectsSelf(sig)
		paramStart := 0
		if expectsSelf {
			paramStart = 1
		}
		paramTypes := make([]ast.TypeExpression, 0, len(sig.Params)-paramStart)
		for idx := paramStart; idx < len(sig.Params); idx++ {
			param := sig.Params[idx]
			if param == nil || param.ParamType == nil {
				return false
			}
			substituted := substituteTypeParams(param.ParamType, bindings)
			paramTypes = append(paramTypes, normalizeTypeExprForPackage(g, ifacePkg, substituted))
		}
		returnExpr := normalizeTypeExprForPackage(g, ifacePkg, substituteTypeParams(sig.ReturnType, bindings))
		var defaultDef *ast.FunctionDefinition
		if sig.DefaultImpl != nil {
			defaultDef = ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
		}
		method := &nativeInterfaceGenericMethod{
			Name:             sig.Name.Name,
			GoName:           sanitizeIdent(sig.Name.Name),
			InterfaceName:    ifaceName,
			InterfacePackage: ifacePkg,
			InterfaceArgs:    ifaceArgs,
			GenericParams:    sig.GenericParams,
			WhereClause:      sig.WhereClause,
			ParamTypeExprs:   paramTypes,
			ReturnTypeExpr:   returnExpr,
			DefaultDefinition: defaultDef,
		}
		if existing, ok := methods[method.Name]; ok {
			if typeExpressionListKey(existing.ParamTypeExprs) != typeExpressionListKey(method.ParamTypeExprs) ||
				typeExpressionToString(existing.ReturnTypeExpr) != typeExpressionToString(method.ReturnTypeExpr) {
				return false
			}
			continue
		}
		methods[method.Name] = method
	}
	return true
}

func typeExpressionListKey(exprs []ast.TypeExpression) string {
	if len(exprs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		parts = append(parts, typeExpressionToString(expr))
	}
	return strings.Join(parts, "|")
}

func (g *generator) typeExprForGoType(goType string) (ast.TypeExpression, bool) {
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		innerExpr, ok := g.typeExprForGoType(spec.ElemGoType)
		if !ok {
			return nil, false
		}
		return ast.NewGenericTypeExpression(ast.Ty("Array"), []ast.TypeExpression{innerExpr}), true
	}
	switch goType {
	case "runtime.Value", "any":
		return ast.Ty("any"), true
	case "bool":
		return ast.Ty("bool"), true
	case "string":
		return ast.Ty("String"), true
	case "rune":
		return ast.Ty("char"), true
	case "int8":
		return ast.Ty("i8"), true
	case "int16":
		return ast.Ty("i16"), true
	case "int32":
		return ast.Ty("i32"), true
	case "int64":
		return ast.Ty("i64"), true
	case "uint8":
		return ast.Ty("u8"), true
	case "uint16":
		return ast.Ty("u16"), true
	case "uint32":
		return ast.Ty("u32"), true
	case "uint64":
		return ast.Ty("u64"), true
	case "int":
		return ast.Ty("isize"), true
	case "uint":
		return ast.Ty("usize"), true
	case "float32":
		return ast.Ty("f32"), true
	case "float64":
		return ast.Ty("f64"), true
	case "struct{}":
		return ast.Ty("void"), true
	case "runtime.ErrorValue":
		return ast.Ty("Error"), true
	case "*Array":
		return ast.NewGenericTypeExpression(ast.Ty("Array"), []ast.TypeExpression{ast.NewWildcardTypeExpression()}), true
	}
	if iface := g.nativeInterfaceInfoForGoType(goType); iface != nil {
		return iface.TypeExpr, true
	}
	if callable := g.nativeCallableInfoForGoType(goType); callable != nil && callable.TypeExpr != nil {
		return callable.TypeExpr, true
	}
	if spec, ok := nativeNullableSpecForPointer(goType); ok {
		innerExpr, ok := g.typeExprForGoType(spec.InnerType)
		if !ok {
			return nil, false
		}
		return ast.NewNullableTypeExpression(innerExpr), true
	}
	if g.typeCategory(goType) == "struct" {
		if info := g.structInfoByGoName(goType); info != nil && info.Name != "" {
			if info.TypeExpr != nil {
				return info.TypeExpr, true
			}
			return ast.Ty(info.Name), true
		}
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		return ast.Ty(baseName), true
	}
	return nil, false
}

func (g *generator) inferNativeInterfaceGenericMethodShape(ctx *compileContext, call *ast.FunctionCall, method *nativeInterfaceGenericMethod, expected string) ([]ast.TypeExpression, []string, ast.TypeExpression, string, map[string]ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || method == nil {
		return nil, nil, nil, "", nil, false
	}
	bindings := make(map[string]ast.TypeExpression, len(method.GenericParams))
	if len(call.TypeArguments) > 0 {
		if len(call.TypeArguments) != len(method.GenericParams) {
			ctx.setReason("generic call arity mismatch")
			return nil, nil, nil, "", nil, false
		}
		for idx, arg := range call.TypeArguments {
			if method.GenericParams[idx] == nil || method.GenericParams[idx].Name == nil || method.GenericParams[idx].Name.Name == "" || arg == nil {
				ctx.setReason("generic call type mismatch")
				return nil, nil, nil, "", nil, false
			}
			bindings[method.GenericParams[idx].Name.Name] = arg
		}
	}
	if len(call.Arguments) != len(method.ParamTypeExprs) {
		ctx.setReason("call arity mismatch")
		return nil, nil, nil, "", nil, false
	}
	genericNames := nativeInterfaceGenericNameSet(method.GenericParams)
	for idx, arg := range call.Arguments {
		inferCtx := ctx.child()
		_, _, argType, ok := g.compileExprLines(inferCtx, arg, "")
		if !ok {
			continue
		}
		actualExpr, ok := g.typeExprForGoType(argType)
		if !ok {
			continue
		}
		_ = g.nativeInterfaceTypeTemplateMatches(method.InterfacePackage, method.ParamTypeExprs[idx], actualExpr, genericNames, bindings)
	}
	if expected != "" && expected != "runtime.Value" && expected != "any" {
		if expectedExpr, ok := g.typeExprForGoType(expected); ok {
			_ = g.nativeInterfaceTypeTemplateMatches(method.InterfacePackage, method.ReturnTypeExpr, expectedExpr, genericNames, bindings)
		}
	}
	mapper := NewTypeMapper(g, method.InterfacePackage)
	paramExprs := make([]ast.TypeExpression, 0, len(method.ParamTypeExprs))
	paramGoTypes := make([]string, 0, len(method.ParamTypeExprs))
	for _, paramExpr := range method.ParamTypeExprs {
		inst := normalizeTypeExprForPackage(g, method.InterfacePackage, substituteTypeParams(paramExpr, bindings))
		goType, ok := mapper.Map(inst)
		if !ok || goType == "" {
			return nil, nil, nil, "", nil, false
		}
		paramExprs = append(paramExprs, inst)
		paramGoTypes = append(paramGoTypes, goType)
	}
	returnExpr := normalizeTypeExprForPackage(g, method.InterfacePackage, substituteTypeParams(method.ReturnTypeExpr, bindings))
	returnGoType, ok := mapper.Map(returnExpr)
	if !ok || returnGoType == "" {
		return nil, nil, nil, "", nil, false
	}
	return paramExprs, paramGoTypes, returnExpr, returnGoType, cloneTypeBindings(bindings), true
}

func (g *generator) nativeInterfaceGenericImplBindings(impl *implMethodInfo, method *nativeInterfaceGenericMethod) (map[string]ast.TypeExpression, bool) {
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

func (g *generator) nativeInterfaceGenericMethodImplExists(goType string, method *nativeInterfaceGenericMethod) bool {
	if g == nil || method == nil || goType == "" {
		return false
	}
	return g.nativeInterfaceGenericMethodImpl(goType, method, method.ParamTypeExprs, nil, method.ReturnTypeExpr, "") != nil
}

func (g *generator) nativeInterfaceGenericMethodImplExistsExact(goType string, method *nativeInterfaceGenericMethod) bool {
	if g == nil || method == nil || goType == "" {
		return false
	}
	found := false
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || !info.Compileable || impl.ImplName != "" {
			continue
		}
		if impl.MethodName != method.Name {
			continue
		}
		if len(info.Params) == 0 || info.Params[0].GoType != goType {
			continue
		}
		if _, ok := g.nativeInterfaceGenericImplBindings(impl, method); !ok {
			continue
		}
		if found {
			return false
		}
		found = true
	}
	return found
}

func (g *generator) nativeInterfaceGenericMethodImpl(goType string, method *nativeInterfaceGenericMethod, paramTypeExprs []ast.TypeExpression, paramGoTypes []string, returnTypeExpr ast.TypeExpression, returnGoType string) *nativeInterfaceAdapterMethod {
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
		if impl.MethodName != method.Name {
			continue
		}
		if len(info.Params) == 0 {
			continue
		}
		bindings, ok := g.nativeInterfaceGenericImplBindings(impl, method)
		if !ok {
			continue
		}
		if info.Params[0].GoType != goType {
			info, bindings, ok = g.nativeInterfaceConcreteImplInfo(goType, impl, bindings)
			if !ok || info == nil || len(info.Params) == 0 || info.Params[0].GoType != goType {
				continue
			}
		}
		implParamTypeExprs, implParamGoTypes, implReturnTypeExpr, implReturnGoType, optionalLast, ok := g.nativeInterfaceMethodImplSignature(impl, bindings)
		if !ok || optionalLast {
			continue
		}
		expectedParamTypeExprs := paramTypeExprs
		if len(expectedParamTypeExprs) == 0 {
			expectedParamTypeExprs = method.ParamTypeExprs
		}
		expectedReturnTypeExpr := returnTypeExpr
		if expectedReturnTypeExpr == nil {
			expectedReturnTypeExpr = method.ReturnTypeExpr
		}
		expectedParamGoTypes := paramGoTypes
		if len(expectedParamGoTypes) == 0 && len(expectedParamTypeExprs) > 0 {
			mapper := NewTypeMapper(g, method.InterfacePackage)
			expectedParamGoTypes = make([]string, 0, len(expectedParamTypeExprs))
			for _, expr := range expectedParamTypeExprs {
				inst := normalizeTypeExprForPackage(g, method.InterfacePackage, substituteTypeParams(expr, bindings))
				goType, ok := mapper.Map(inst)
				if !ok || goType == "" {
					expectedParamGoTypes = nil
					break
				}
				expectedParamGoTypes = append(expectedParamGoTypes, goType)
			}
		}
		expectedReturnGoType := returnGoType
		if expectedReturnGoType == "" && expectedReturnTypeExpr != nil {
			mapper := NewTypeMapper(g, method.InterfacePackage)
			expectedReturnGoType, _ = mapper.Map(expectedReturnTypeExpr)
		}
		if len(implParamGoTypes) != len(expectedParamTypeExprs) || (len(expectedParamGoTypes) > 0 && len(expectedParamGoTypes) != len(implParamGoTypes)) {
			continue
		}
		matches := true
		leftVars := make(map[string]string)
		rightVars := make(map[string]string)
		for idx := range implParamGoTypes {
			if len(expectedParamGoTypes) > 0 && g.canCoerceStaticExpr(expectedParamGoTypes[idx], implParamGoTypes[idx]) {
				continue
			}
			if !g.typeExprEquivalentModuloGenerics(expectedParamTypeExprs[idx], implParamTypeExprs[idx], leftVars, rightVars) {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		if expectedReturnGoType != "" && !g.canCoerceStaticExpr(expectedReturnGoType, implReturnGoType) &&
			!g.typeExprEquivalentModuloGenerics(expectedReturnTypeExpr, implReturnTypeExpr, leftVars, rightVars) {
			continue
		}
		candidate := &nativeInterfaceAdapterMethod{
			Info:                 info,
			CompiledReturnGoType: info.ReturnType,
			ParamGoTypes:         implParamGoTypes,
			ReturnGoType:         implReturnGoType,
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
