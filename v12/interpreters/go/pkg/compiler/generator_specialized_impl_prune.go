package compiler

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
	if len(info.TypeBindings) == 0 {
		return nil, "", false
	}
	key := g.specializedImplFunctionKey(info, info.TypeBindings)
	if key == "" {
		return nil, "", false
	}
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
