package compiler

func equivalentFunctionInfoSignature(left, right *functionInfo) bool {
	if left == nil || right == nil {
		return false
	}
	if left == right {
		return true
	}
	if left.Package != right.Package || left.Name != right.Name || left.ReturnType != right.ReturnType {
		return false
	}
	if left.Definition == nil || right.Definition == nil || left.Definition != right.Definition {
		return false
	}
	if len(left.Params) != len(right.Params) {
		return false
	}
	for idx := range left.Params {
		if left.Params[idx].GoType != right.Params[idx].GoType {
			return false
		}
	}
	return true
}

