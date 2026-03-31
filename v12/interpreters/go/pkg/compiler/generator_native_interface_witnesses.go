package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) nativeInterfaceImplExpectsSelf(info *functionInfo, impl *implMethodInfo) bool {
	if info != nil && info.Definition != nil {
		return methodDefinitionExpectsSelf(info.Definition)
	}
	if impl != nil && impl.Info != nil && impl.Info.Definition != nil {
		return methodDefinitionExpectsSelf(impl.Info.Definition)
	}
	return false
}

func (g *generator) nativeInterfaceImplTargetExpr(info *functionInfo, impl *implMethodInfo, bindings map[string]ast.TypeExpression) ast.TypeExpression {
	if g == nil || impl == nil || impl.TargetType == nil {
		return nil
	}
	pkgName := ""
	if info != nil && info.Package != "" {
		pkgName = info.Package
	} else if impl.Info != nil {
		pkgName = impl.Info.Package
	}
	if len(bindings) > 0 {
		if concrete := g.specializedImplTargetType(impl, bindings); concrete != nil {
			return normalizeTypeExprForPackage(g, pkgName, concrete)
		}
	}
	if info != nil && len(info.TypeBindings) > 0 {
		if concrete := g.specializedImplTargetType(impl, info.TypeBindings); concrete != nil {
			return normalizeTypeExprForPackage(g, pkgName, concrete)
		}
	}
	return normalizeTypeExprForPackage(g, pkgName, impl.TargetType)
}

func (g *generator) nativeInterfaceImplWitnessGoType(info *functionInfo, impl *implMethodInfo, bindings map[string]ast.TypeExpression) string {
	if g == nil || impl == nil {
		return ""
	}
	if g.nativeInterfaceImplExpectsSelf(info, impl) {
		if info != nil && len(info.Params) > 0 {
			receiverType := info.Params[0].GoType
			if receiverType != "" && receiverType != "runtime.Value" && receiverType != "any" {
				return receiverType
			}
			paramTypeExpr := info.Params[0].TypeExpr
			if paramTypeExpr == nil {
				paramTypeExpr = g.nativeInterfaceImplTargetExpr(info, impl, bindings)
			}
			if paramTypeExpr == nil {
				return receiverType
			}
			mapper := NewTypeMapper(g, info.Package)
			mapped, ok := mapper.Map(paramTypeExpr)
			mapped, ok = g.recoverRepresentableCarrierType(info.Package, paramTypeExpr, mapped)
			if !ok || mapped == "" {
				return receiverType
			}
			return mapped
		}
		return ""
	}
	targetExpr := g.nativeInterfaceImplTargetExpr(info, impl, bindings)
	if targetExpr == nil {
		return ""
	}
	pkgName := ""
	if info != nil && info.Package != "" {
		pkgName = info.Package
	} else if impl.Info != nil {
		pkgName = impl.Info.Package
	}
	mapper := NewTypeMapper(g, pkgName)
	goType, ok := mapper.Map(targetExpr)
	goType, ok = g.recoverRepresentableCarrierType(pkgName, targetExpr, goType)
	if !ok || goType == "" {
		return ""
	}
	return goType
}

func (g *generator) nativeInterfaceCompiledParamGoTypes(info *functionInfo, expectsSelf bool) []string {
	if info == nil || len(info.Params) == 0 {
		return nil
	}
	start := 0
	if expectsSelf {
		start = 1
	}
	if start >= len(info.Params) {
		return nil
	}
	paramGoTypes := make([]string, 0, len(info.Params)-start)
	for idx := start; idx < len(info.Params); idx++ {
		paramGoTypes = append(paramGoTypes, info.Params[idx].GoType)
	}
	return paramGoTypes
}
