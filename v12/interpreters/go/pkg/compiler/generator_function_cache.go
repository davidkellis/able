package compiler

func (g *generator) invalidateFunctionDerivedInfo(info *functionInfo) {
	if g == nil || info == nil {
		return
	}
	info.cachedBindings = nil
	info.cachedCarrier = false
}

func (g *generator) invalidateAllFunctionDerivedInfo() {
	if g == nil {
		return
	}
	for _, info := range g.allFunctionInfos() {
		g.invalidateFunctionDerivedInfo(info)
	}
}
