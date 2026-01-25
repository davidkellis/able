package typechecker

func mergeFunctionOverload(existing Type, next FunctionType) (Type, bool, bool) {
	switch current := existing.(type) {
	case FunctionType:
		if functionSignaturesEquivalent(current, next) {
			return existing, true, true
		}
		return FunctionOverloadType{Overloads: []FunctionType{current, next}}, true, false
	case FunctionOverloadType:
		for _, overload := range current.Overloads {
			if functionSignaturesEquivalent(overload, next) {
				return existing, true, true
			}
		}
		merged := make([]FunctionType, 0, len(current.Overloads)+1)
		merged = append(merged, current.Overloads...)
		merged = append(merged, next)
		return FunctionOverloadType{Overloads: merged}, true, false
	default:
		return existing, false, false
	}
}

func hasExactFunctionSignature(existing Type, next FunctionType) bool {
	switch current := existing.(type) {
	case FunctionType:
		return typesEquivalentForSignature(current, next)
	case FunctionOverloadType:
		for _, overload := range current.Overloads {
			if typesEquivalentForSignature(overload, next) {
				return true
			}
		}
	}
	return false
}
