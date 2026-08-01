package typechecker

import "able/interpreter-go/pkg/ast"

func assignabilityDiagnostic(message string, node ast.Node, actual, expected Type) Diagnostic {
	return Diagnostic{
		Severity: SeverityError,
		Code:     assignabilityDiagnosticCode(actual, expected),
		Message:  message,
		Node:     node,
	}
}

func assignabilityDiagnosticCode(actual, expected Type) DiagnosticCode {
	actual = normalizeSpecialType(expandAliasForUnion(actual))
	expected = normalizeSpecialType(expandAliasForUnion(expected))
	if actual == nil || expected == nil {
		return ""
	}
	if _, ok := actual.(FunctionType); ok {
		if _, ok := expected.(FunctionType); ok {
			return DiagnosticCodeCallableSignatureMismatch
		}
	}
	if sameInvariantConstructor(actual, expected) &&
		!invariantTypeEquivalent(actual, expected) {
		return DiagnosticCodeInvariantTypeArgument
	}
	return ""
}

func sameInvariantConstructor(actual, expected Type) bool {
	switch actualType := actual.(type) {
	case ArrayType:
		_, ok := expected.(ArrayType)
		return ok
	case MapType:
		_, ok := expected.(MapType)
		return ok
	case RangeType:
		_, ok := expected.(RangeType)
		return ok
	case IteratorType:
		_, ok := expected.(IteratorType)
		return ok
	case FutureType:
		_, ok := expected.(FutureType)
		return ok
	case NullableType:
		_, ok := expected.(NullableType)
		return ok
	case AppliedType:
		expectedType, ok := expected.(AppliedType)
		return ok && exactTypeEquivalent(actualType.Base, expectedType.Base)
	case StructInstanceType:
		expectedType, ok := expected.(StructInstanceType)
		return ok && actualType.StructName == expectedType.StructName
	case UnionType:
		expectedType, ok := expected.(UnionType)
		return ok && actualType.UnionName != "" && actualType.UnionName == expectedType.UnionName
	default:
		return false
	}
}
