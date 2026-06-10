package compiler

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) preferImplSpecializationTemplate(pkgName string, base ast.TypeExpression, candidate ast.TypeExpression) bool {
	if g == nil || candidate == nil {
		return false
	}
	if base == nil {
		return true
	}
	base = normalizeTypeExprForPackage(g, pkgName, base)
	candidate = normalizeTypeExprForPackage(g, pkgName, candidate)
	if _, ok := base.(*ast.GenericTypeExpression); ok {
		if _, ok := candidate.(*ast.GenericTypeExpression); !ok {
			return false
		}
	}
	return true
}

func (g *generator) staticReceiverTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if g == nil {
		return nil, false
	}
	if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, goType); ok && inferred != nil {
		inferred = normalizeTypeExprForPackage(g, ctx.packageName, inferred)
		if preferred := g.preferConcreteTypeExprForGoType(ctx, inferred, goType); preferred != nil {
			if g.typeExprCompatibleWithCarrier(ctx, preferred, goType) {
				return preferred, true
			}
			return nil, false
		}
		if g.typeExprCompatibleWithCarrier(ctx, inferred, goType) {
			return inferred, true
		}
		return nil, false
	}
	if preferred := g.preferConcreteTypeExprForGoType(ctx, nil, goType); preferred != nil {
		if g.typeExprCompatibleWithCarrier(ctx, preferred, goType) {
			return preferred, true
		}
		return nil, false
	}
	return nil, false
}

func (g *generator) typeExprCompatibleWithCarrier(ctx *compileContext, expr ast.TypeExpression, goType string) bool {
	if g == nil || expr == nil {
		return false
	}
	if strings.TrimSpace(goType) == "" || goType == "runtime.Value" || goType == "any" {
		return true
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	expr = normalizeTypeExprForPackage(g, pkgName, expr)
	if expr == nil {
		return false
	}
	if !g.typeExprFullyBound(pkgName, expr) {
		return true
	}
	canonicalGoType, ok := g.lowerCarrierTypeInPackage(pkgName, expr)
	if !ok || canonicalGoType == "" || canonicalGoType == "runtime.Value" || canonicalGoType == "any" {
		return false
	}
	if canonicalGoType == goType {
		return true
	}
	if g.sameNominalStructFamily(goType, canonicalGoType) || g.sameNominalStructFamily(canonicalGoType, goType) {
		return true
	}
	if g.staticArrayCarrierCoercible(goType, canonicalGoType) || g.staticArrayCarrierCoercible(canonicalGoType, goType) {
		return true
	}
	return g.receiverGoTypeCompatible(canonicalGoType, goType) || g.receiverGoTypeCompatible(goType, canonicalGoType)
}

func (g *generator) staticTargetTypeExpr(ctx *compileContext, expr ast.Expression) (ast.TypeExpression, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil && ident.Name != "" && ctx != nil {
		if _, exists := ctx.lookup(ident.Name); !exists {
			if bound, ok := ctx.typeBindings[ident.Name]; ok && bound != nil {
				return normalizeTypeExprForPackage(g, ctx.packageName, bound), true
			}
		}
	}
	if inferred, ok := g.inferExpressionTypeExpr(ctx, expr, ""); ok && inferred != nil {
		inferred = normalizeTypeExprForPackage(g, ctx.packageName, inferred)
		if preferred := g.preferConcreteTypeExprForGoType(ctx, inferred, ""); preferred != nil {
			return preferred, true
		}
		return inferred, true
	}
	return nil, false
}

func (g *generator) preferConcreteTypeExprForGoType(ctx *compileContext, inferred ast.TypeExpression, goType string) ast.TypeExpression {
	if g == nil || strings.TrimSpace(goType) == "" {
		return inferred
	}
	concrete, ok := g.typeExprForGoType(goType)
	if !ok || concrete == nil {
		return inferred
	}
	pkgName := ""
	if ctx != nil {
		pkgName = ctx.packageName
	}
	concretePkg := g.resolvedTypeExprPackage(pkgName, concrete)
	concrete = normalizeTypeExprForPackage(g, concretePkg, concrete)
	if concrete == nil || (!g.typeExprFullyBound(pkgName, concrete) && !g.typeExprFullyBound(concretePkg, concrete)) {
		return inferred
	}
	if inferred == nil {
		return concrete
	}
	inferredPkg := g.resolvedTypeExprPackage(pkgName, inferred)
	if !g.typeExprFullyBound(pkgName, inferred) && !g.typeExprFullyBound(inferredPkg, inferred) {
		return concrete
	}
	if ctx != nil && g.typeExprHasGeneric(inferred, ctx.genericNames) {
		return concrete
	}
	return inferred
}

func (g *generator) inferExpressionTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if g == nil {
		return nil, false
	}
	if inferred, ok := g.inferLocalTypeExpr(ctx, expr, goType); ok && inferred != nil {
		return g.lowerNormalizedTypeExpr(ctx, inferred), true
	}
	if goType != "" {
		if inferred, ok := g.typeExprForGoType(goType); ok && inferred != nil {
			return g.lowerNormalizedTypeExpr(ctx, inferred), true
		}
	}
	return nil, false
}

func (g *generator) functionReturnTypeExpr(info *functionInfo) ast.TypeExpression {
	return g.functionReturnTypeExprWithBindings(info, g.compileContextTypeBindings(info))
}

func (g *generator) functionReturnTypeExprWithBindings(info *functionInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || info == nil || info.Definition == nil {
		return nil
	}
	retExpr := g.functionDeclaredOrInferredReturnTypeExpr(info)
	if retExpr == nil {
		return nil
	}
	if impl := g.implMethodByInfo[info]; impl != nil {
		concreteTarget := g.specializedImplTargetType(impl, bindings)
		if concreteTarget == nil {
			concreteTarget = impl.TargetType
		}
		interfaceBindings := g.implTypeBindings(info.Package, impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, concreteTarget)
		selfTarget := g.implSelfTargetType(info.Package, concreteTarget, interfaceBindings)
		allBindings := g.mergeImplSelfTargetBindings(info.Package, concreteTarget, selfTarget, interfaceBindings)
		for name, expr := range bindings {
			if expr == nil {
				continue
			}
			if name == "Self" && selfTarget != nil {
				continue
			}
			if allBindings == nil {
				allBindings = make(map[string]ast.TypeExpression)
			}
			allBindings[name] = normalizeTypeExprForPackage(g, info.Package, expr)
		}
		if selfTarget != nil {
			if allBindings == nil {
				allBindings = make(map[string]ast.TypeExpression)
			}
			allBindings["Self"] = normalizeTypeExprForPackage(g, info.Package, selfTarget)
		}
		retExpr = resolveSelfTypeExpr(retExpr, selfTarget)
		retExpr = substituteTypeParams(retExpr, allBindings)
		return normalizeTypeExprForPackage(g, info.Package, retExpr)
	}
	retExpr = substituteTypeParams(retExpr, bindings)
	return normalizeTypeExprForPackage(g, info.Package, retExpr)
}

func (g *generator) concreteCompileContextBindings(info *functionInfo, genericNames map[string]struct{}) map[string]ast.TypeExpression {
	return g.normalizeConcreteTypeBindings(info.Package, g.compileContextTypeBindings(info), genericNames)
}

func (g *generator) mergeConcreteTypeBindings(pkgName string, genericNames map[string]struct{}, base map[string]ast.TypeExpression, extra map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || len(extra) == 0 {
		return base
	}
	if base == nil {
		base = make(map[string]ast.TypeExpression)
	}
	for name, expr := range extra {
		if expr == nil {
			continue
		}
		if len(genericNames) > 0 {
			if _, ok := genericNames[name]; !ok {
				continue
			}
		}
		if _, exists := base[name]; exists {
			continue
		}
		base[name] = normalizeTypeExprForPackage(g, pkgName, expr)
	}
	return base
}

func (g *generator) normalizeConcreteTypeBindings(pkgName string, bindings map[string]ast.TypeExpression, genericNames map[string]struct{}) map[string]ast.TypeExpression {
	if g == nil || len(bindings) == 0 {
		return nil
	}
	out := make(map[string]ast.TypeExpression, len(bindings))
	for name, expr := range bindings {
		if len(genericNames) > 0 {
			if _, ok := genericNames[name]; !ok {
				continue
			}
		}
		if expr == nil {
			continue
		}
		resolvedPkg := g.resolvedTypeExprPackage(pkgName, expr)
		normalized := normalizeTypeExprForPackage(g, resolvedPkg, expr)
		normalized = g.recordResolvedTypeExprPackage(normalized, resolvedPkg)
		mapped, ok := g.lowerCarrierTypeInPackage(resolvedPkg, normalized)
		mapped, ok = g.recoverRepresentableCarrierType(resolvedPkg, normalized, mapped)
		if !g.typeExprFullyBound(pkgName, normalized) && !g.typeExprFullyBound(resolvedPkg, normalized) &&
			(!ok || mapped == "" || mapped == "runtime.Value" || mapped == "any") {
			continue
		}
		if simple, ok := normalized.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == name {
			continue
		}
		if g.typeExprHasGeneric(normalized, genericNames) {
			continue
		}
		out[name] = normalized
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneTypeBindings(bindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if len(bindings) == 0 {
		return nil
	}
	out := make(map[string]ast.TypeExpression, len(bindings))
	for name, expr := range bindings {
		out[name] = expr
	}
	return out
}

func (g *generator) specializedTypeTemplateMatches(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	if len(bindings) > 0 {
		template = substituteTypeParams(template, bindings)
		actual = substituteTypeParams(actual, bindings)
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	return g.specializedTypeTemplateMatchesNormalized(pkgName, template, actual, genericNames, bindings, seen)
}

func (g *generator) specializedSameBaseGenericBindings(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	templateGeneric, ok := template.(*ast.GenericTypeExpression)
	if !ok || templateGeneric == nil {
		return false
	}
	actualGeneric, ok := actual.(*ast.GenericTypeExpression)
	if !ok || actualGeneric == nil || len(templateGeneric.Arguments) != len(actualGeneric.Arguments) {
		return false
	}
	if normalizeTypeExprString(g, pkgName, templateGeneric.Base) != normalizeTypeExprString(g, pkgName, actualGeneric.Base) {
		return false
	}
	for idx := range templateGeneric.Arguments {
		if !g.specializedBindTemplateArg(pkgName, templateGeneric.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings) {
			return false
		}
	}
	return true
}

func (g *generator) specializedBindTemplateArg(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	template = g.normalizeTypeExprForSpecialization(pkgName, template, nil)
	actual = g.normalizeTypeExprForSpecialization(pkgName, actual, nil)
	switch tt := template.(type) {
	case *ast.SimpleTypeExpression:
		if tt == nil || tt.Name == nil || tt.Name.Name == "" {
			return false
		}
		if _, ok := genericNames[tt.Name.Name]; ok {
			if bound, exists := bindings[tt.Name.Name]; exists {
				if normalizeTypeExprString(g, pkgName, bound) == tt.Name.Name && normalizeTypeExprString(g, pkgName, actual) != tt.Name.Name {
					bindings[tt.Name.Name] = actual
					return true
				}
				boundKey := normalizeTypeExprIdentityKey(g, pkgName, bound)
				actualKey := normalizeTypeExprIdentityKey(g, pkgName, actual)
				if boundKey != "" && actualKey != "" && boundKey != actualKey &&
					normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual) &&
					g.typeExprFullyBound(pkgName, actual) {
					bindings[tt.Name.Name] = actual
					return true
				}
				return normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual)
			}
			bindings[tt.Name.Name] = actual
			return true
		}
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	case *ast.GenericTypeExpression:
		actualGeneric, ok := actual.(*ast.GenericTypeExpression)
		if !ok || actualGeneric == nil || len(tt.Arguments) != len(actualGeneric.Arguments) {
			return false
		}
		if normalizeTypeExprString(g, pkgName, tt.Base) != normalizeTypeExprString(g, pkgName, actualGeneric.Base) {
			return false
		}
		for idx := range tt.Arguments {
			if !g.specializedBindTemplateArg(pkgName, tt.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		actualNullable, ok := actual.(*ast.NullableTypeExpression)
		return ok && actualNullable != nil && g.specializedBindTemplateArg(pkgName, tt.InnerType, actualNullable.InnerType, genericNames, bindings)
	case *ast.ResultTypeExpression:
		actualResult, ok := actual.(*ast.ResultTypeExpression)
		return ok && actualResult != nil && g.specializedBindTemplateArg(pkgName, tt.InnerType, actualResult.InnerType, genericNames, bindings)
	case *ast.UnionTypeExpression:
		if actualUnion, ok := actual.(*ast.UnionTypeExpression); ok && actualUnion != nil {
			if len(tt.Members) != len(actualUnion.Members) {
				return false
			}
			for idx := range tt.Members {
				if !g.specializedBindTemplateArg(pkgName, tt.Members[idx], actualUnion.Members[idx], genericNames, bindings) {
					return false
				}
			}
			return true
		}
		for _, member := range tt.Members {
			candidate := cloneTypeBindings(bindings)
			if candidate == nil {
				candidate = make(map[string]ast.TypeExpression)
			}
			if !g.specializedBindTemplateArg(pkgName, member, actual, genericNames, candidate) {
				continue
			}
			applyTypeBindings(bindings, candidate)
			return true
		}
		return false
	case *ast.FunctionTypeExpression:
		actualFn, ok := actual.(*ast.FunctionTypeExpression)
		if !ok || actualFn == nil || len(tt.ParamTypes) != len(actualFn.ParamTypes) {
			return false
		}
		for idx := range tt.ParamTypes {
			if !g.specializedBindTemplateArg(pkgName, tt.ParamTypes[idx], actualFn.ParamTypes[idx], genericNames, bindings) {
				return false
			}
		}
		return g.specializedBindTemplateArg(pkgName, tt.ReturnType, actualFn.ReturnType, genericNames, bindings)
	default:
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
}

func (g *generator) specializedTypeTemplateMatchesNormalized(pkgName string, template ast.TypeExpression, actual ast.TypeExpression, genericNames map[string]struct{}, bindings map[string]ast.TypeExpression, seen map[string]struct{}) bool {
	if g == nil || template == nil || actual == nil {
		return false
	}
	if !g.typeExprHasGeneric(template, genericNames) && !g.typeExprHasGeneric(actual, genericNames) {
		if _, unionTemplate := template.(*ast.UnionTypeExpression); !unionTemplate {
			if _, unionActual := actual.(*ast.UnionTypeExpression); !unionActual {
				if normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual) {
					return true
				}
				if templateSimple, ok := template.(*ast.SimpleTypeExpression); ok && templateSimple != nil && templateSimple.Name != nil {
					if actualGeneric, ok := actual.(*ast.GenericTypeExpression); ok && actualGeneric != nil {
						if actualBase, ok := typeExprBaseName(actualGeneric.Base); ok && actualBase == templateSimple.Name.Name {
							return true
						}
					}
				}
				return false
			}
		}
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	key := specializedTypeTemplateMatchKey(template, actual)
	if _, ok := seen[key]; ok {
		return true
	}
	seen[key] = struct{}{}
	switch tt := template.(type) {
	case *ast.SimpleTypeExpression:
		if tt == nil || tt.Name == nil || tt.Name.Name == "" {
			return false
		}
		if _, ok := genericNames[tt.Name.Name]; ok {
			if bound, exists := bindings[tt.Name.Name]; exists {
				if normalizeTypeExprString(g, pkgName, bound) == tt.Name.Name && normalizeTypeExprString(g, pkgName, actual) != tt.Name.Name {
					bindings[tt.Name.Name] = actual
					return true
				}
				boundKey := normalizeTypeExprIdentityKey(g, pkgName, bound)
				actualKey := normalizeTypeExprIdentityKey(g, pkgName, actual)
				if boundKey != "" && actualKey != "" && boundKey != actualKey &&
					normalizeTypeExprString(g, pkgName, bound) == normalizeTypeExprString(g, pkgName, actual) &&
					g.typeExprFullyBound(pkgName, actual) {
					bindings[tt.Name.Name] = actual
					return true
				}
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
		if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Base, actualGeneric.Base, genericNames, bindings, seen) {
			return false
		}
		for idx := range tt.Arguments {
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Arguments[idx], actualGeneric.Arguments[idx], genericNames, bindings, seen) {
				return false
			}
		}
		return true
	case *ast.NullableTypeExpression:
		actualNullable, ok := actual.(*ast.NullableTypeExpression)
		return ok && actualNullable != nil && g.specializedTypeTemplateMatchesNormalized(pkgName, tt.InnerType, actualNullable.InnerType, genericNames, bindings, seen)
	case *ast.ResultTypeExpression:
		actualResult, ok := actual.(*ast.ResultTypeExpression)
		return ok && actualResult != nil && g.specializedTypeTemplateMatchesNormalized(pkgName, tt.InnerType, actualResult.InnerType, genericNames, bindings, seen)
	case *ast.UnionTypeExpression:
		if actualUnion, ok := actual.(*ast.UnionTypeExpression); ok && actualUnion != nil {
			if len(tt.Members) != len(actualUnion.Members) {
				return false
			}
			for idx := range tt.Members {
				if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.Members[idx], actualUnion.Members[idx], genericNames, bindings, seen) {
					return false
				}
			}
			return true
		}
		for _, member := range tt.Members {
			candidate := cloneTypeBindings(bindings)
			if candidate == nil {
				candidate = make(map[string]ast.TypeExpression)
			}
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, member, actual, genericNames, candidate, seen) {
				continue
			}
			applyTypeBindings(bindings, candidate)
			return true
		}
		return false
	case *ast.FunctionTypeExpression:
		actualFn, ok := actual.(*ast.FunctionTypeExpression)
		if !ok || actualFn == nil || len(tt.ParamTypes) != len(actualFn.ParamTypes) {
			return false
		}
		for idx := range tt.ParamTypes {
			if !g.specializedTypeTemplateMatchesNormalized(pkgName, tt.ParamTypes[idx], actualFn.ParamTypes[idx], genericNames, bindings, seen) {
				return false
			}
		}
		return g.specializedTypeTemplateMatchesNormalized(pkgName, tt.ReturnType, actualFn.ReturnType, genericNames, bindings, seen)
	default:
		return normalizeTypeExprString(g, pkgName, template) == normalizeTypeExprString(g, pkgName, actual)
	}
}

func (g *generator) normalizeTypeExprForSpecialization(pkgName string, expr ast.TypeExpression, seen map[string]struct{}) ast.TypeExpression {
	if g == nil || expr == nil {
		return expr
	}
	pkgName = g.resolvedTypeExprPackage(pkgName, expr)
	if seen == nil {
		seen = make(map[string]struct{})
	}
	switch t := expr.(type) {
	case *ast.SimpleTypeExpression:
		if t == nil || t.Name == nil || t.Name.Name == "" {
			return normalizeTypeExprForPackage(g, pkgName, expr)
		}
		key := pkgName + "|" + t.Name.Name
		if _, ok := seen[key]; ok {
			return expr
		}
		nextSeen := make(map[string]struct{}, len(seen)+1)
		for existing := range seen {
			nextSeen[existing] = struct{}{}
		}
		nextSeen[key] = struct{}{}
		if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil && expanded != expr {
			return g.normalizeTypeExprForSpecialization(pkgName, expanded, nextSeen)
		}
		return expr
	case *ast.GenericTypeExpression:
		if t == nil {
			return expr
		}
		if baseName, ok := typeExprBaseName(t.Base); ok && baseName != "" {
			key := pkgName + "|" + baseName + "<" + normalizeTypeExprListKey(g, pkgName, t.Arguments) + ">"
			if _, ok := seen[key]; ok {
				return expr
			}
			nextSeen := make(map[string]struct{}, len(seen)+1)
			for existing := range seen {
				nextSeen[existing] = struct{}{}
			}
			nextSeen[key] = struct{}{}
			if expanded := g.expandTypeAliasForPackage(pkgName, expr); expanded != nil && expanded != expr {
				return g.normalizeTypeExprForSpecialization(pkgName, expanded, nextSeen)
			}
		}
		basePkg := g.resolvedTypeExprPackage(pkgName, t.Base)
		base := g.normalizeTypeExprForSpecialization(basePkg, t.Base, seen)
		changed := base != t.Base
		args := make([]ast.TypeExpression, 0, len(t.Arguments))
		for _, arg := range t.Arguments {
			argPkg := g.resolvedTypeExprPackage(pkgName, arg)
			next := g.normalizeTypeExprForSpecialization(argPkg, arg, seen)
			args = append(args, next)
			if next != arg {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return g.recordResolvedTypeExprPackage(ast.NewGenericTypeExpression(base, args), pkgName)
	case *ast.NullableTypeExpression:
		if t == nil {
			return expr
		}
		innerPkg := g.resolvedTypeExprPackage(pkgName, t.InnerType)
		inner := g.normalizeTypeExprForSpecialization(innerPkg, t.InnerType, seen)
		if inner == t.InnerType {
			return expr
		}
		return g.recordResolvedTypeExprPackage(ast.NewNullableTypeExpression(inner), pkgName)
	case *ast.ResultTypeExpression:
		if t == nil {
			return expr
		}
		innerPkg := g.resolvedTypeExprPackage(pkgName, t.InnerType)
		inner := g.normalizeTypeExprForSpecialization(innerPkg, t.InnerType, seen)
		if inner == t.InnerType {
			return expr
		}
		return g.recordResolvedTypeExprPackage(ast.NewResultTypeExpression(inner), pkgName)
	case *ast.UnionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		members := make([]ast.TypeExpression, 0, len(t.Members))
		for _, member := range t.Members {
			memberPkg := g.resolvedTypeExprPackage(pkgName, member)
			next := g.normalizeTypeExprForSpecialization(memberPkg, member, seen)
			members = append(members, next)
			if next != member {
				changed = true
			}
		}
		if !changed {
			return expr
		}
		return g.recordResolvedTypeExprPackage(ast.NewUnionTypeExpression(members), pkgName)
	case *ast.FunctionTypeExpression:
		if t == nil {
			return expr
		}
		changed := false
		params := make([]ast.TypeExpression, 0, len(t.ParamTypes))
		for _, param := range t.ParamTypes {
			paramPkg := g.resolvedTypeExprPackage(pkgName, param)
			next := g.normalizeTypeExprForSpecialization(paramPkg, param, seen)
			params = append(params, next)
			if next != param {
				changed = true
			}
		}
		retPkg := g.resolvedTypeExprPackage(pkgName, t.ReturnType)
		ret := g.normalizeTypeExprForSpecialization(retPkg, t.ReturnType, seen)
		if ret != t.ReturnType {
			changed = true
		}
		if !changed {
			return g.recordResolvedTypeExprPackage(normalizeCallableSyntaxTypeExpr(expr), pkgName)
		}
		return g.recordResolvedTypeExprPackage(normalizeCallableSyntaxTypeExpr(ast.NewFunctionTypeExpression(params, ret)), pkgName)
	default:
		return normalizeTypeExprForPackage(g, pkgName, expr)
	}
}

func specializedTypeTemplateMatchKey(template ast.TypeExpression, actual ast.TypeExpression) string {
	return fmt.Sprintf("%T:%x|%T:%x", template, typeExprPointer(template), actual, typeExprPointer(actual))
}

func typeExprPointer(expr ast.TypeExpression) uintptr {
	if expr == nil {
		return 0
	}
	value := reflect.ValueOf(expr)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return 0
	}
	return value.Pointer()
}

func (g *generator) specializedImplFunctionKey(info *functionInfo, bindings map[string]ast.TypeExpression) string {
	if info == nil {
		return ""
	}
	if g != nil {
		bindings = g.canonicalImplSpecializationBindings(info, nil, bindings)
	}
	base := strings.TrimSpace(info.Name)
	if info.QualifiedName != "" {
		base = strings.TrimSpace(info.QualifiedName)
	}
	if base == "" {
		base = strings.TrimSpace(info.GoName)
	}
	if pkg := strings.TrimSpace(info.Package); pkg != "" {
		base = pkg + "::" + base
	}
	if g != nil {
		if impl := g.implMethodInfoForFunction(info); impl != nil {
			if impl.ImplName != "" {
				base += "|impl=" + impl.ImplName
			}
			if constraintKey := constraintSignature(collectConstraintSpecs(impl.ImplGenerics, impl.WhereClause)); constraintKey != "" && constraintKey != "<none>" {
				base += "|constraints=" + constraintKey
			}
		}
	}
	if len(bindings) == 0 {
		return base
	}
	names := make([]string, 0, len(bindings))
	for name := range bindings {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names)+1)
	parts = append(parts, base)
	for _, name := range names {
		bindingKey := normalizeTypeExprIdentityKey(g, info.Package, bindings[name])
		if bindingKey == "" {
			bindingKey = normalizeTypeExprString(g, info.Package, bindings[name])
		}
		parts = append(parts, name+"="+bindingKey)
	}
	return strings.Join(parts, "|")
}
