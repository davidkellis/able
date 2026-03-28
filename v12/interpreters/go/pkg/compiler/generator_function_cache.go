package compiler

func (g *generator) invalidateFunctionDerivedInfo(info *functionInfo) {
	if g == nil || info == nil {
		return
	}
	info.cachedBindings = nil
	info.cachedCarrier = false
}

