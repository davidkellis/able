package compiler

func (g *generator) touchNativeInterfaceAdapters() {
	if g == nil {
		return
	}
	g.nativeInterfaceAdapterVersion++
	if g.nativeInterfaceAdapterVersion <= 0 {
		g.nativeInterfaceAdapterVersion = 1
	}
}

func (g *generator) nativeInterfaceRefreshAllowed() bool {
	if g == nil {
		return false
	}
	return len(g.nativeInterfaceSpecializing) == 0 && g.bodyCompilationDepth == 0
}
