package compiler

import (
	"sort"
	"strings"
)

func (g *generator) discardRedundantImplFallbackSpecializations() {
	if g == nil || len(g.specializedFunctions) == 0 {
		return
	}
	compileable := make(map[string]*functionInfo)
	for _, info := range g.specializedFunctions {
		if info == nil || !info.Compileable {
			continue
		}
		if _, key, ok := g.redundantImplSpecializationKey(info); ok {
			compileable[key] = info
		}
	}
	if len(compileable) == 0 {
		return
	}
	removed := false
	kept := g.specializedFunctions[:0]
	for _, info := range g.specializedFunctions {
		if info == nil {
			continue
		}
		_, key, ok := g.redundantImplSpecializationKey(info)
		if ok && !info.Compileable {
			if winner, exists := compileable[key]; exists && winner != nil && winner != info {
				g.dropSpecializedFunctionIndexValue(info)
				removed = true
				continue
			}
		}
		kept = append(kept, info)
	}
	if removed {
		g.specializedFunctions = kept
		g.touchNativeInterfaceAdapters()
	}
}

func (g *generator) redundantImplSpecializationKey(info *functionInfo) (*implMethodInfo, string, bool) {
	if g == nil || info == nil {
		return nil, "", false
	}
	impl := g.implMethodByInfo[info]
	if impl == nil {
		return nil, "", false
	}
	baseInfo := impl.Info
	if baseInfo == nil {
		baseInfo = info
	}
	method := &methodInfo{
		TargetType:  impl.TargetType,
		MethodName:  impl.MethodName,
		ExpectsSelf: methodDefinitionExpectsSelf(baseInfo.Definition),
		Info:        baseInfo,
	}
	concreteTarget := g.specializedImplTargetType(impl, info.TypeBindings)
	if concreteTarget == nil {
		concreteTarget = impl.TargetType
	}
	if concreteTarget == nil {
		return nil, "", false
	}
	parts := []string{
		g.implMethodCanonicalKey(impl),
		normalizeTypeExprString(g, baseInfo.Package, concreteTarget),
	}
	genericNames := g.implSpecializationGenericNames(method)
	if iface, _, ok := g.interfaceDefinitionForImpl(impl); ok && iface != nil {
		for name := range g.interfaceSelfBindingNames(iface) {
			delete(genericNames, name)
		}
	}
	delete(genericNames, "Self")
	delete(genericNames, "SelfType")
	if len(genericNames) > 0 {
		names := make([]string, 0, len(genericNames))
		for name := range genericNames {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			expr := info.TypeBindings[name]
			if expr == nil {
				continue
			}
			parts = append(parts, name+"="+normalizeTypeExprString(g, baseInfo.Package, expr))
		}
	}
	key := strings.Join(parts, "|")
	return impl, key, true
}

func (g *generator) dropSpecializedFunctionIndexValue(target *functionInfo) {
	if g == nil || target == nil || g.specializedFunctionIndex == nil {
		return
	}
	for key, info := range g.specializedFunctionIndex {
		if info == target {
			delete(g.specializedFunctionIndex, key)
		}
	}
}
