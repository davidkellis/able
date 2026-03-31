package compiler

func (g *generator) nativeInterfaceDispatchCandidateEligible(info *functionInfo) bool {
	if g == nil || info == nil {
		return false
	}
	g.refreshRepresentableFunctionInfo(info)
	if !info.SupportedTypes {
		return false
	}
	if info.Compileable {
		return true
	}
	// During body compileability, interface adapter/dispatch discovery must be
	// able to see supported static candidates before the fixed point has marked
	// them compileable, otherwise callers lock in a residual runtime fallback.
	if g.bodyCompilationDepth > 0 {
		return true
	}
	if info.Definition == nil {
		return false
	}
	g.refreshRepresentableFunctionInfo(info)
	if info.Compileable {
		return true
	}
	if g.nativeInterfaceRefreshAllowed() && g.bodyCompileable(info, info.ReturnType) {
		info.Compileable = true
		info.Reason = ""
		return true
	}
	return false
}
