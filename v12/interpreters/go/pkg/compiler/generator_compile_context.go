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
	if g == nil || info == nil || g.implMethodByInfo == nil {
		return nil
	}
	impl := g.implMethodByInfo[info]
	if impl == nil || len(impl.InterfaceGenerics) == 0 || len(impl.InterfaceArgs) == 0 {
		return nil
	}
	bindings := make(map[string]ast.TypeExpression, len(impl.InterfaceArgs))
	for idx, gp := range impl.InterfaceGenerics {
		if gp == nil || gp.Name == nil || gp.Name.Name == "" || idx >= len(impl.InterfaceArgs) || impl.InterfaceArgs[idx] == nil {
			continue
		}
		bindings[gp.Name.Name] = normalizeTypeExprForPackage(g, info.Package, impl.InterfaceArgs[idx])
	}
	if len(bindings) == 0 {
		return nil
	}
	return bindings
}
