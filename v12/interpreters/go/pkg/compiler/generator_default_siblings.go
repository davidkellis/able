package compiler

import "strings"

func (g *generator) defaultSiblingsForFunction(info *functionInfo) map[string]implSiblingInfo {
	if g == nil || info == nil {
		return nil
	}
	if implInfo := g.implMethodByInfo[info]; implInfo != nil && implInfo.IsDefault {
		return g.implSiblingsForFunction(info)
	}
	if meta := g.nativeInterfaceDefaultByInfo[info]; meta != nil {
		return g.nativeInterfaceDefaultSiblingsForFunction(meta)
	}
	return nil
}

func (g *generator) registerNativeInterfaceDefaultMethodInfo(info *functionInfo, method *nativeInterfaceGenericMethod, receiverGoType string) {
	if g == nil || info == nil || method == nil || receiverGoType == "" {
		return
	}
	g.nativeInterfaceDefaultByInfo[info] = &nativeInterfaceDefaultMethodInfo{
		InterfacePackage: method.InterfacePackage,
		InterfaceName:    method.InterfaceName,
		InterfaceArgs:    cloneTypeExprSlice(method.InterfaceArgs),
		ReceiverGoType:   receiverGoType,
		MethodName:       method.Name,
	}
}

func (g *generator) nativeInterfaceDefaultSiblingsForFunction(meta *nativeInterfaceDefaultMethodInfo) map[string]implSiblingInfo {
	if g == nil || meta == nil || meta.InterfaceName == "" || meta.ReceiverGoType == "" {
		return nil
	}
	cacheKey := g.nativeInterfaceDefaultSiblingCacheKey(meta)
	if cached, ok := g.nativeInterfaceDefaultSiblingCache[cacheKey]; ok {
		if len(cached) == 0 {
			return nil
		}
		return cached
	}
	info, ok := g.ensureNativeInterfaceInfo(meta.InterfacePackage, nativeInterfaceInstantiationExpr(meta.InterfaceName, meta.InterfaceArgs))
	if !ok || info == nil {
		return nil
	}
	siblings := make(map[string]implSiblingInfo)
	ambiguous := make(map[string]struct{})
	for _, method := range info.Methods {
		if method == nil || method.Name == "" || method.Name == meta.MethodName {
			continue
		}
		// Default interface methods may need to call either an explicit sibling
		// impl or another derived default sibling on the same receiver. Use the
		// full shared method resolution path here instead of exact-only matching.
		impl := g.nativeInterfaceMethodImpl(meta.ReceiverGoType, method)
		if impl == nil || impl.Info == nil || !impl.Info.Compileable {
			continue
		}
		candidate := implSiblingInfo{GoName: impl.Info.GoName, Arity: impl.Info.Arity, Info: impl.Info}
		if existing, ok := siblings[method.Name]; ok && existing.Info != candidate.Info {
			delete(siblings, method.Name)
			ambiguous[method.Name] = struct{}{}
			continue
		}
		if _, blocked := ambiguous[method.Name]; blocked {
			continue
		}
		siblings[method.Name] = candidate
	}
	if len(siblings) == 0 {
		g.nativeInterfaceDefaultSiblingCache[cacheKey] = map[string]implSiblingInfo{}
		return nil
	}
	g.nativeInterfaceDefaultSiblingCache[cacheKey] = siblings
	return siblings
}

func (g *generator) nativeInterfaceDefaultSiblingCacheKey(meta *nativeInterfaceDefaultMethodInfo) string {
	if g == nil || meta == nil {
		return ""
	}
	parts := []string{
		meta.InterfacePackage,
		meta.InterfaceName,
		normalizeTypeExprListKey(g, meta.InterfacePackage, meta.InterfaceArgs),
		meta.ReceiverGoType,
		meta.MethodName,
	}
	return strings.Join(parts, "::")
}
