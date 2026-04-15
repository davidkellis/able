package compiler

func (g *generator) nativeInterfaceNewImplCandidate(impl *implMethodInfo, info *functionInfo) (nativeInterfaceImplCandidate, bool) {
	if g == nil {
		return nativeInterfaceImplCandidate{}, false
	}
	if info == nil && impl != nil {
		info = impl.Info
	}
	if impl == nil || info == nil {
		return nativeInterfaceImplCandidate{}, false
	}
	candidate := nativeInterfaceImplCandidate{
		impl:        impl,
		info:        info,
		specificity: g.nativeInterfaceImplCandidateSpecificity(nativeInterfaceImplCandidate{impl: impl, info: info}),
		goName:      info.GoName,
		name:        info.Name,
	}
	if len(info.Params) > 0 {
		candidate.receiver = info.Params[0].GoType
	}
	return candidate, true
}

func nativeInterfaceImplCandidateLess(left nativeInterfaceImplCandidate, right nativeInterfaceImplCandidate) bool {
	if left.specificity != right.specificity {
		return left.specificity > right.specificity
	}
	if left.receiver != right.receiver {
		return left.receiver < right.receiver
	}
	if left.goName != right.goName {
		return left.goName < right.goName
	}
	return left.name < right.name
}

func mergeNativeInterfaceImplCandidates(base []nativeInterfaceImplCandidate, delta []nativeInterfaceImplCandidate) []nativeInterfaceImplCandidate {
	if len(base) == 0 {
		return append([]nativeInterfaceImplCandidate(nil), delta...)
	}
	if len(delta) == 0 {
		return base
	}
	merged := make([]nativeInterfaceImplCandidate, 0, len(base)+len(delta))
	i := 0
	j := 0
	for i < len(base) && j < len(delta) {
		if nativeInterfaceImplCandidateLess(base[i], delta[j]) {
			merged = append(merged, base[i])
			i++
			continue
		}
		merged = append(merged, delta[j])
		j++
	}
	if i < len(base) {
		merged = append(merged, base[i:]...)
	}
	if j < len(delta) {
		merged = append(merged, delta[j:]...)
	}
	return merged
}
