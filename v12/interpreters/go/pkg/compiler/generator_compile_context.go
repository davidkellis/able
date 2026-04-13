package compiler

import "able/interpreter-go/pkg/ast"

func newCompileContext(gen *generator, info *functionInfo, functions map[string]*functionInfo, overloads map[string]*overloadInfo, packageName string, genericNames map[string]struct{}) *compileContext {
	counter := 0
	ctx := &compileContext{
		params:       make(map[string]paramInfo),
		locals:       make(map[string]paramInfo),
		functions:    functions,
		overloads:    overloads,
		packageName:  packageName,
		temps:        &counter,
		loopDepth:    0,
		breakpoints:  make(map[string]int),
		genericNames: genericNames,
	}
	if info != nil {
		ctx.returnType = info.ReturnType
		if gen != nil {
			ctx.returnTypeExpr = gen.functionReturnTypeExpr(info)
		} else if info.Definition != nil {
			ctx.returnTypeExpr = info.Definition.ReturnType
		}
		for _, param := range info.Params {
			if param.Name == "" {
				continue
			}
			ctx.params[param.Name] = param
			if fact, ok := info.ParamFacts[param.Name]; ok && param.GoName != "" && fact.hasUsefulFact() {
				ctx.setIntegerFact(param.GoName, fact)
			}
		}
		if len(info.Params) > 0 {
			ctx.implicitReceiver = info.Params[0]
			ctx.hasImplicitReceiver = true
		}
	}
	if gen != nil {
		ctx.typeBindings = gen.compileContextTypeBindings(info)
	}
	return ctx
}

func (g *generator) compileContextTypeBindings(info *functionInfo) map[string]ast.TypeExpression {
	if g == nil || info == nil {
		return nil
	}
	merged := make(map[string]ast.TypeExpression)
	var currentImpl *implMethodInfo
	var currentTarget ast.TypeExpression
	if g.implMethodByInfo != nil {
		if impl := g.implMethodByInfo[info]; impl != nil {
			currentImpl = impl
			currentTarget = g.compileContextImplTargetType(info, impl)
			bindings := g.implTypeBindings(info.Package, impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, currentTarget)
			selfTarget := g.implSelfTargetType(info.Package, currentTarget, bindings)
			merged = g.mergeImplSelfTargetBindings(info.Package, currentTarget, selfTarget, bindings)
			if merged == nil {
				merged = make(map[string]ast.TypeExpression)
			}
			skipNames := map[string]struct{}{}
			if selfTarget != nil {
				merged["Self"] = normalizeTypeExprForPackage(g, info.Package, selfTarget)
				skipNames["Self"] = struct{}{}
				skipNames["SelfType"] = struct{}{}
				if iface, _, ok := g.interfaceDefinitionForImpl(impl); ok && iface != nil {
					for name, expr := range g.interfaceSelfTypeBindings(iface, selfTarget) {
						if expr == nil {
							continue
						}
						merged[name] = normalizeTypeExprForPackage(g, info.Package, expr)
						skipNames[name] = struct{}{}
					}
					for name := range g.interfaceSelfBindingNames(iface) {
						skipNames[name] = struct{}{}
					}
				}
			}
			for name, expr := range info.TypeBindings {
				if expr == nil {
					continue
				}
				if _, skip := skipNames[name]; skip {
					continue
				}
				merged[name] = normalizeTypeExprForPackage(g, info.Package, expr)
			}
		} else {
			for name, expr := range info.TypeBindings {
				if expr != nil {
					merged[name] = normalizeTypeExprForPackage(g, info.Package, expr)
				}
			}
		}
	}
	if len(info.Params) > 0 {
		selfParam := info.Params[0]
		if currentImpl == nil && (selfParam.Name == "self" || selfParam.Name == "Self") {
			var candidate ast.TypeExpression
			if selfParam.TypeExpr != nil {
				candidate = normalizeTypeExprForPackage(g, info.Package, selfParam.TypeExpr)
			} else if selfExpr, ok := g.typeExprForGoType(selfParam.GoType); ok && selfExpr != nil {
				candidate = normalizeTypeExprForPackage(g, info.Package, selfExpr)
			}
			if candidate != nil {
				if existing, ok := merged["Self"]; !ok || existing == nil {
					merged["Self"] = candidate
				}
			}
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func (g *generator) compileContextImplTargetType(info *functionInfo, impl *implMethodInfo) ast.TypeExpression {
	if g == nil || info == nil || impl == nil {
		return nil
	}
	currentTarget := g.specializedImplTargetType(impl, info.TypeBindings)
	if len(info.Params) > 0 {
		selfParam := info.Params[0]
		if selfParam.Name == "self" || selfParam.Name == "Self" {
			if selfParam.TypeExpr != nil {
				if candidate := normalizeTypeExprForPackage(g, info.Package, selfParam.TypeExpr); candidate != nil {
					if currentTarget == nil {
						return candidate
					}
					if normalizeTypeExprString(g, info.Package, candidate) == normalizeTypeExprString(g, info.Package, currentTarget) ||
						g.nominalTargetTypeExprCompatible(info.Package, candidate, currentTarget) ||
						g.nominalTargetTypeExprCompatible(info.Package, currentTarget, candidate) {
						return candidate
					}
					return currentTarget
				}
			}
			if selfParam.GoType != "" {
				if candidate, ok := g.typeExprForGoType(selfParam.GoType); ok && candidate != nil {
					if candidate = normalizeTypeExprForPackage(g, info.Package, candidate); candidate != nil {
						if currentTarget == nil {
							return candidate
						}
						if normalizeTypeExprString(g, info.Package, candidate) == normalizeTypeExprString(g, info.Package, currentTarget) ||
							g.nominalTargetTypeExprCompatible(info.Package, candidate, currentTarget) ||
							g.nominalTargetTypeExprCompatible(info.Package, currentTarget, candidate) {
							return candidate
						}
						return currentTarget
					}
				}
			}
		}
	}
	if currentTarget != nil {
		return currentTarget
	}
	return impl.TargetType
}
