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
		if info.Definition != nil {
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
	if g.implMethodByInfo != nil {
		if impl := g.implMethodByInfo[info]; impl != nil {
			bindings := g.implTypeBindings(impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, impl.TargetType)
			for name, expr := range bindings {
				if expr != nil {
					merged[name] = normalizeTypeExprForPackage(g, info.Package, expr)
				}
			}
		}
	}
	for name, expr := range info.TypeBindings {
		if expr != nil {
			merged[name] = normalizeTypeExprForPackage(g, info.Package, expr)
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}
