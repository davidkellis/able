package compiler

func (g *generator) nativeInterfaceDispatchCandidateEligible(info *functionInfo) bool {
	if g == nil || info == nil || !info.SupportedTypes {
		return false
	}
	if info.Compileable {
		return true
	}
	// During body compileability, interface adapter/dispatch discovery must be
	// able to see supported static candidates before the fixed point has marked
	// them compileable, otherwise callers lock in a residual runtime fallback.
	return g.bodyCompilationDepth > 0
}
