package compiler

import "able/interpreter-go/pkg/ast"

type joinBranchInfo struct {
	GoType   string
	Expr     ast.Expression
	TypeExpr ast.TypeExpression
	SawNil   bool
}

func (g *generator) joinResultType(ctx *compileContext, types ...string) (string, bool) {
	if g == nil || len(types) == 0 {
		return "", false
	}
	unique := make([]string, 0, len(types))
	seen := make(map[string]struct{}, len(types))
	for _, goType := range types {
		if goType == "" || goType == "runtime.Value" || goType == "any" || g.isVoidType(goType) {
			return "", false
		}
		if _, ok := seen[goType]; ok {
			continue
		}
		seen[goType] = struct{}{}
		unique = append(unique, goType)
	}
	if len(unique) == 0 {
		return "", false
	}
	if len(unique) == 1 {
		return unique[0], true
	}
	for _, candidate := range unique {
		if !g.isJoinCarrierType(candidate) {
			continue
		}
		compatible := true
		for _, actual := range unique {
			if actual == candidate {
				continue
			}
			if !g.canCoerceStaticExpr(candidate, actual) {
				compatible = false
				break
			}
		}
		if compatible {
			return candidate, true
		}
	}
	if errorType, ok := g.commonJoinErrorType(unique); ok {
		return errorType, true
	}
	if ifaceType, ok := g.commonJoinInterfaceType(ctx, unique); ok {
		return ifaceType, true
	}
	memberExprs := make([]ast.TypeExpression, 0, len(unique))
	for _, goType := range unique {
		typeExpr, ok := g.typeExprForGoType(goType)
		if !ok || typeExpr == nil {
			return "", false
		}
		memberExprs = append(memberExprs, typeExpr)
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	info, ok := g.ensureNativeUnionInfo(pkgName, memberExprs)
	if !ok || info == nil {
		return "", false
	}
	return info.GoType, true
}

func (g *generator) joinCarrierTypeFromTypeExpr(ctx *compileContext, expr ast.TypeExpression) (string, bool) {
	if g == nil || expr == nil {
		return "", false
	}
	expr = g.lowerNormalizedTypeExpr(ctx, expr)
	mapped, ok := g.lowerCarrierType(ctx, expr)
	if ok && mapped != "" && mapped != "runtime.Value" && mapped != "any" && !g.isVoidType(mapped) {
		return mapped, true
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	if _, _, _, _, ok := interfaceExprInfo(g, pkgName, expr); ok {
		if info, ok := g.ensureNativeInterfaceInfo(pkgName, expr); ok && info != nil && info.GoType != "" {
			return info.GoType, true
		}
	}
	if _, members, ok := g.expandedUnionMembersInPackage(pkgName, expr); ok {
		if info, ok := g.ensureNativeUnionInfo(pkgName, members); ok && info != nil && info.GoType != "" {
			return info.GoType, true
		}
	}
	if !ok || mapped == "" || mapped == "runtime.Value" || mapped == "any" || g.isVoidType(mapped) {
		return "", false
	}
	return mapped, true
}

func (g *generator) recoverJoinBranchType(ctx *compileContext, branch joinBranchInfo) (string, bool) {
	if g == nil {
		return "", false
	}
	if branch.SawNil {
		return "", true
	}
	if branch.GoType != "" && branch.GoType != "runtime.Value" && branch.GoType != "any" && !g.isVoidType(branch.GoType) {
		return branch.GoType, true
	}
	if recovered, ok := g.joinCarrierTypeFromTypeExpr(ctx, branch.TypeExpr); ok {
		return recovered, true
	}
	if branch.Expr != nil {
		if inferred, ok := g.inferExpressionTypeExpr(ctx, branch.Expr, branch.GoType); ok && inferred != nil {
			if recovered, ok := g.joinCarrierTypeFromTypeExpr(ctx, inferred); ok {
				return recovered, true
			}
		}
	}
	return "", false
}

func (g *generator) joinResultTypeFromBranches(ctx *compileContext, branches []joinBranchInfo) (string, bool) {
	if g == nil || len(branches) == 0 {
		return "", false
	}
	concreteTypes := make([]string, 0, len(branches))
	sawNil := false
	for _, branch := range branches {
		if branch.SawNil {
			sawNil = true
			continue
		}
		goType, ok := g.recoverJoinBranchType(ctx, branch)
		if !ok || goType == "" {
			return "", false
		}
		concreteTypes = append(concreteTypes, goType)
	}
	return g.joinResultTypeWithSawNil(ctx, concreteTypes, sawNil)
}

func (g *generator) joinResultTypeWithSawNil(ctx *compileContext, types []string, sawNil bool) (string, bool) {
	if g == nil {
		return "", false
	}
	concreteTypes := make([]string, 0, len(types))
	for _, goType := range types {
		if goType == "" || goType == "runtime.Value" || goType == "any" || g.isVoidType(goType) {
			return "", false
		}
		concreteTypes = append(concreteTypes, goType)
	}
	if len(concreteTypes) == 0 {
		return "", false
	}
	joinedType, ok := g.joinResultType(ctx, concreteTypes...)
	if !ok {
		return "", false
	}
	if !sawNil {
		return joinedType, true
	}
	if g.goTypeHasNilZeroValue(joinedType) {
		return joinedType, true
	}
	if nullableType, ok := g.nativeNullablePointerType(joinedType); ok {
		return nullableType, true
	}
	return "", false
}

func (g *generator) joinResultTypeAllowNil(ctx *compileContext, types []string, nilFlags []bool) (string, bool) {
	if g == nil || len(types) == 0 {
		return "", false
	}
	concreteTypes := make([]string, 0, len(types))
	sawNil := false
	for idx, goType := range types {
		if idx < len(nilFlags) && nilFlags[idx] {
			sawNil = true
			continue
		}
		concreteTypes = append(concreteTypes, goType)
	}
	return g.joinResultTypeWithSawNil(ctx, concreteTypes, sawNil)
}

func (g *generator) joinBranchIsNilExpr(expr string, exprType string) bool {
	if g == nil || expr == "" || exprType == "" {
		return false
	}
	switch {
	case exprType == "any" && expr == "any(nil)":
		return true
	case exprType == "runtime.Value" && expr == "runtime.NilValue{}":
		return true
	}
	if typedNil, ok := g.typedNilExpr(exprType); ok && expr == typedNil {
		return true
	}
	return false
}

func (g *generator) commonJoinErrorType(types []string) (string, bool) {
	if g == nil || len(types) == 0 {
		return "", false
	}
	for _, goType := range types {
		if goType != "runtime.ErrorValue" && !g.isNativeErrorCarrierType(goType) {
			return "", false
		}
	}
	return "runtime.ErrorValue", true
}

func (g *generator) commonJoinInterfaceType(ctx *compileContext, types []string) (string, bool) {
	if g == nil || len(types) == 0 {
		return "", false
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	for _, goType := range types {
		g.materializeJoinInterfaceCandidatesForActual(pkgName, goType)
	}
	candidates := make([]*nativeInterfaceInfo, 0, len(g.nativeInterfaces))
	seen := make(map[string]struct{}, len(g.nativeInterfaces))
	for _, info := range g.nativeInterfaces {
		if info == nil || info.GoType == "" {
			continue
		}
		if _, ok := seen[info.GoType]; ok {
			continue
		}
		seen[info.GoType] = struct{}{}
		compatible := true
		for _, goType := range types {
			if !g.nativeInterfaceAcceptsActual(info, goType) {
				compatible = false
				break
			}
		}
		if compatible {
			candidates = append(candidates, info)
		}
	}
	if len(candidates) == 0 {
		return "", false
	}
	maximal := make([]*nativeInterfaceInfo, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		dominated := false
		for _, other := range candidates {
			if other == nil || other == candidate {
				continue
			}
			if g.nativeInterfaceAssignable(other.GoType, candidate.GoType) &&
				!g.nativeInterfaceAssignable(candidate.GoType, other.GoType) {
				dominated = true
				break
			}
		}
		if !dominated {
			maximal = append(maximal, candidate)
		}
	}
	if len(maximal) != 1 || maximal[0] == nil || maximal[0].GoType == "" {
		return "", false
	}
	return maximal[0].GoType, true
}

func (g *generator) materializeJoinInterfaceCandidatesForActual(pkgName string, goType string) {
	if g == nil || goType == "" {
		return
	}
	if info := g.nativeInterfaceInfoForGoType(goType); info != nil && info.TypeExpr != nil {
		g.ensureJoinInterfaceInfoTree(pkgName, info.TypeExpr, make(map[string]struct{}))
	}
	for _, candidateInfo := range g.nativeInterfaceImplCandidates() {
		impl := candidateInfo.impl
		info := candidateInfo.info
		if impl == nil || info == nil || !info.Compileable || impl.InterfaceName == "" || impl.ImplName != "" {
			continue
		}
		_, bindings, ok := g.nativeInterfaceConcreteImplInfo(goType, impl, info.TypeBindings)
		if !ok {
			continue
		}
		ifaceExpr := g.implConcreteInterfaceExpr(impl, bindings)
		if ifaceExpr == nil {
			continue
		}
		ifacePkg := g.interfacePackages[impl.InterfaceName]
		if ifacePkg == "" {
			ifacePkg = info.Package
		}
		g.ensureJoinInterfaceInfoTree(ifacePkg, ifaceExpr, make(map[string]struct{}))
	}
}

func (g *generator) ensureJoinInterfaceInfoTree(pkgName string, expr ast.TypeExpression, seen map[string]struct{}) {
	if g == nil || expr == nil {
		return
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	ifacePkg, ifaceName, ifaceArgs, ifaceDef, ok := interfaceExprInfo(g, pkgName, expr)
	if !ok {
		return
	}
	key := ifacePkg + "::" + ifaceName + "<" + normalizeTypeExprListKey(g, ifacePkg, ifaceArgs) + ">"
	if _, exists := seen[key]; exists {
		return
	}
	seen[key] = struct{}{}
	_, _ = g.ensureNativeInterfaceInfo(ifacePkg, expr)
	if ifaceDef == nil {
		return
	}
	bindings := nativeInterfaceBindings(ifaceDef, ifaceArgs)
	for _, baseExpr := range ifaceDef.BaseInterfaces {
		if baseExpr == nil {
			continue
		}
		next := substituteTypeParams(baseExpr, bindings)
		next = normalizeTypeExprForPackage(g, ifacePkg, next)
		g.ensureJoinInterfaceInfoTree(ifacePkg, next, seen)
	}
}

func (g *generator) isJoinCarrierType(goType string) bool {
	if g == nil || goType == "" {
		return false
	}
	if g.nativeUnionInfoForGoType(goType) != nil {
		return true
	}
	if g.nativeInterfaceInfoForGoType(goType) != nil {
		return true
	}
	if g.nativeCallableInfoForGoType(goType) != nil {
		return true
	}
	if g.isNativeNullableValueType(goType) {
		return true
	}
	return goType == "runtime.ErrorValue"
}

func (g *generator) coerceJoinBranch(ctx *compileContext, resultType string, expr string, exprType string) ([]string, string, bool) {
	if resultType == "" || exprType == "" || expr == "" {
		return nil, "", false
	}
	if resultType == "runtime.Value" {
		if exprType == "runtime.Value" {
			return nil, expr, true
		}
		convLines, converted, ok := g.lowerRuntimeValue(ctx, expr, exprType)
		if !ok {
			return nil, "", false
		}
		return convLines, converted, true
	}
	if resultType == "any" {
		return nil, expr, true
	}
	if exprType == "runtime.Value" {
		return g.lowerExpectRuntimeValue(ctx, expr, resultType)
	}
	if exprType == "any" {
		if expr == "any(nil)" {
			if typedNil, ok := g.typedNilExpr(resultType); ok {
				return nil, typedNil, true
			}
		}
		return g.lowerExpectRuntimeValue(ctx, expr, resultType)
	}
	lines, coercedExpr, coercedType, ok := g.lowerCoerceExpectedStaticExpr(ctx, nil, expr, exprType, resultType)
	if !ok {
		return nil, "", false
	}
	if coercedType == resultType || g.typeMatches(resultType, coercedType) {
		return lines, coercedExpr, true
	}
	return nil, "", false
}
