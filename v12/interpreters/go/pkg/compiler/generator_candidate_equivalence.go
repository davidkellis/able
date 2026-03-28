package compiler

func equivalentFunctionInfoSignature(left, right *functionInfo) bool {
	if left == nil || right == nil {
		return false
	}
	if left == right {
		return true
	}
	// Specialized compiled variants of the same source definition may receive
	// different synthetic function names and may even be reconstructed from
	// separate parsed AST instances while still representing the same
	// carrier-level callable shape. Candidate selection should treat those as
	// equivalent when the source-level function name and mapped Go signature
	// match.
	if left.Package != right.Package || left.ReturnType != right.ReturnType {
		return false
	}
	if functionInfoSourceName(left) != functionInfoSourceName(right) {
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

func functionInfoSourceName(info *functionInfo) string {
	if info == nil {
		return ""
	}
	if info.Definition != nil && info.Definition.ID != nil && info.Definition.ID.Name != "" {
		return info.Definition.ID.Name
	}
	return info.Name
}
