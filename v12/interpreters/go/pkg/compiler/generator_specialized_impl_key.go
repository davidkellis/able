package compiler

import "able/interpreter-go/pkg/ast"

func (g *generator) canonicalImplSpecializationBindings(info *functionInfo, impl *implMethodInfo, bindings map[string]ast.TypeExpression) map[string]ast.TypeExpression {
	if g == nil || info == nil || len(bindings) == 0 {
		return bindings
	}
	if impl == nil {
		impl = g.implMethodInfoForFunction(info)
	}
	if impl == nil {
		return bindings
	}
	method := &methodInfo{
		TargetType:  impl.TargetType,
		MethodName:  impl.MethodName,
		ExpectsSelf: methodDefinitionExpectsSelf(info.Definition),
		Info:        info,
	}
	allowed := g.implSpecializationGenericNames(method)
	out := make(map[string]ast.TypeExpression)
	for name, expr := range bindings {
		if expr == nil {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[name]; !ok {
				continue
			}
		}
		out[name] = expr
	}
	delete(out, "Self")
	delete(out, "SelfType")
	if iface := g.interfaces[impl.InterfaceName]; iface != nil {
		for name := range g.interfaceSelfBindingNames(iface) {
			delete(out, name)
		}
	}
	concreteTarget := g.specializedImplTargetType(impl, bindings)
	if concreteTarget == nil {
		concreteTarget = impl.TargetType
	}
	if concreteTarget == nil {
		return out
	}
	concreteTarget = normalizeTypeExprForPackage(g, info.Package, concreteTarget)
	if seeded := cloneTypeBindings(out); seeded != nil {
		if g.seedImplBindingsFromConcreteTarget(method, impl, concreteTarget, seeded) {
			for name, expr := range seeded {
				if expr == nil {
					continue
				}
				if len(allowed) > 0 {
					if _, ok := allowed[name]; !ok {
						continue
					}
				}
				out[name] = normalizeTypeExprForPackage(g, info.Package, expr)
			}
		}
	}
	out["Self"] = concreteTarget
	if iface := g.interfaces[impl.InterfaceName]; iface != nil {
		interfaceBindings := g.implTypeBindings(info.Package, impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, concreteTarget)
		selfTarget := g.implSelfTargetType(info.Package, concreteTarget, interfaceBindings)
		for name, expr := range g.interfaceSelfTypeBindings(iface, selfTarget) {
			if expr == nil {
				continue
			}
			out[name] = normalizeTypeExprForPackage(g, info.Package, expr)
		}
	}
	return out
}
