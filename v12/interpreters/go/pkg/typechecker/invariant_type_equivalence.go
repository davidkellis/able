package typechecker

// invariantTypeEquivalent compares types used beneath a generic or callable
// boundary. Unknowns remain compatible while inference is incomplete, but
// concrete types must have the same representation rather than merely support
// a top-level value conversion.
func invariantTypeEquivalent(a, b Type) bool {
	return typesEquivalent(a, b, true)
}

func exactTypeEquivalent(a, b Type) bool {
	return typesEquivalent(a, b, false)
}

func typesEquivalent(a, b Type, allowUnknown bool) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	a = normalizeSpecialType(expandAliasForUnion(a))
	b = normalizeSpecialType(expandAliasForUnion(b))
	if isUnknownType(a) || isUnknownType(b) {
		return allowUnknown
	}
	if allowUnknown && (isTypeParameter(a) || isTypeParameter(b)) {
		return true
	}

	switch actual := a.(type) {
	case PrimitiveType:
		expected, ok := b.(PrimitiveType)
		return ok && actual.Kind == expected.Kind
	case IntegerType:
		expected, ok := b.(IntegerType)
		return ok && actual.Suffix == expected.Suffix
	case FloatType:
		expected, ok := b.(FloatType)
		return ok && actual.Suffix == expected.Suffix
	case TypeParameterType:
		expected, ok := b.(TypeParameterType)
		return ok && actual.ParameterName == expected.ParameterName
	case StructType:
		switch expected := b.(type) {
		case StructType:
			return actual.StructName == expected.StructName
		case StructInstanceType:
			return actual.StructName == expected.StructName && len(expected.TypeArgs) == 0
		default:
			return false
		}
	case StructInstanceType:
		switch expected := b.(type) {
		case StructInstanceType:
			return actual.StructName == expected.StructName &&
				invariantTypeArgumentsEquivalentWithUnknown(actual.TypeArgs, expected.TypeArgs, allowUnknown)
		case StructType:
			return actual.StructName == expected.StructName && len(actual.TypeArgs) == 0
		default:
			return false
		}
	case InterfaceType:
		expected, ok := b.(InterfaceType)
		return ok && actual.InterfaceName == expected.InterfaceName
	case UnionType:
		expected, ok := b.(UnionType)
		if !ok || actual.UnionName != expected.UnionName {
			return false
		}
		if allowUnknown && (len(actual.Variants) == 0 || len(expected.Variants) == 0) {
			return true
		}
		return unorderedTypesEquivalent(actual.Variants, expected.Variants, allowUnknown)
	case FunctionType:
		expected, ok := b.(FunctionType)
		return ok && functionTypesEquivalent(actual, expected, allowUnknown)
	case FunctionOverloadType:
		expected, ok := b.(FunctionOverloadType)
		if !ok || len(actual.Overloads) != len(expected.Overloads) {
			return false
		}
		for i := range actual.Overloads {
			if !functionTypesEquivalent(actual.Overloads[i], expected.Overloads[i], allowUnknown) {
				return false
			}
		}
		return true
	case FutureType:
		expected, ok := b.(FutureType)
		return ok && typesEquivalent(actual.Result, expected.Result, allowUnknown)
	case AppliedType:
		expected, ok := b.(AppliedType)
		return ok &&
			typesEquivalent(actual.Base, expected.Base, allowUnknown) &&
			invariantTypeArgumentsEquivalentWithUnknown(actual.Arguments, expected.Arguments, allowUnknown)
	case NullableType:
		expected, ok := b.(NullableType)
		return ok && typesEquivalent(actual.Inner, expected.Inner, allowUnknown)
	case UnionLiteralType:
		expected, ok := b.(UnionLiteralType)
		return ok && unorderedTypesEquivalent(actual.Members, expected.Members, allowUnknown)
	case ArrayType:
		expected, ok := b.(ArrayType)
		return ok && typesEquivalent(actual.Element, expected.Element, allowUnknown)
	case MapType:
		expected, ok := b.(MapType)
		return ok &&
			typesEquivalent(actual.Key, expected.Key, allowUnknown) &&
			typesEquivalent(actual.Value, expected.Value, allowUnknown)
	case RangeType:
		expected, ok := b.(RangeType)
		return ok && typesEquivalent(actual.Element, expected.Element, allowUnknown)
	case IteratorType:
		expected, ok := b.(IteratorType)
		return ok && typesEquivalent(actual.Element, expected.Element, allowUnknown)
	case PackageType:
		expected, ok := b.(PackageType)
		return ok && actual.Package == expected.Package
	case ImplementationNamespaceType:
		expected, ok := b.(ImplementationNamespaceType)
		return ok && actual.Name() == expected.Name()
	default:
		return a.Name() == b.Name()
	}
}

func invariantTypeArgumentsEquivalent(actual, expected []Type) bool {
	return invariantTypeArgumentsEquivalentWithUnknown(actual, expected, true)
}

func invariantTypeArgumentsEquivalentWithUnknown(actual, expected []Type, allowUnknown bool) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range actual {
		if !typesEquivalent(actual[i], expected[i], allowUnknown) {
			return false
		}
	}
	return true
}

func unorderedTypesEquivalent(actual, expected []Type, allowUnknown bool) bool {
	if len(actual) != len(expected) {
		return false
	}
	used := make([]bool, len(expected))
	for _, actualType := range actual {
		found := false
		for i, expectedType := range expected {
			if !used[i] && typesEquivalent(actualType, expectedType, allowUnknown) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func functionTypesEquivalent(actual, expected FunctionType, allowUnknown bool) bool {
	if len(actual.TypeParams) != len(expected.TypeParams) {
		return false
	}

	substitution := make(map[string]Type, len(expected.TypeParams))
	for i := range actual.TypeParams {
		actualParam := actual.TypeParams[i]
		expectedParam := expected.TypeParams[i]
		substitution[expectedParam.Name] = TypeParameterType{ParameterName: actualParam.Name}
		if actualParam.IsInferred != expectedParam.IsInferred ||
			!invariantTypeArgumentsEquivalentWithUnknown(
				actualParam.Constraints,
				substituteTypes(expectedParam.Constraints, substitution),
				allowUnknown,
			) {
			return false
		}
	}

	if len(actual.Params) != len(expected.Params) {
		return false
	}
	expectedParams := substituteTypes(expected.Params, substitution)
	for i := range actual.Params {
		if !typesEquivalent(actual.Params[i], expectedParams[i], allowUnknown) {
			return false
		}
	}
	if !typesEquivalent(actual.Return, substituteType(expected.Return, substitution), allowUnknown) {
		return false
	}
	return whereSpecsEquivalent(actual.Where, expected.Where, substitution, allowUnknown) &&
		obligationsEquivalent(actual.Obligations, expected.Obligations, substitution, allowUnknown)
}

func substituteTypes(types []Type, substitution map[string]Type) []Type {
	out := make([]Type, len(types))
	for i, typ := range types {
		out[i] = substituteType(typ, substitution)
	}
	return out
}

func whereSpecsEquivalent(actual, expected []WhereConstraintSpec, substitution map[string]Type, allowUnknown bool) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range actual {
		actualSpec := actual[i]
		expectedSpec := expected[i]
		if replacement, ok := substitution[expectedSpec.TypeParam].(TypeParameterType); ok {
			expectedSpec.TypeParam = replacement.ParameterName
		}
		if actualSpec.TypeParam != expectedSpec.TypeParam ||
			!typesEquivalent(actualSpec.Subject, substituteType(expectedSpec.Subject, substitution), allowUnknown) ||
			!invariantTypeArgumentsEquivalentWithUnknown(
				actualSpec.Constraints,
				substituteTypes(expectedSpec.Constraints, substitution),
				allowUnknown,
			) {
			return false
		}
	}
	return true
}

func obligationsEquivalent(actual, expected []ConstraintObligation, substitution map[string]Type, allowUnknown bool) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range actual {
		actualObligation := actual[i]
		expectedObligation := expected[i]
		if replacement, ok := substitution[expectedObligation.TypeParam].(TypeParameterType); ok {
			expectedObligation.TypeParam = replacement.ParameterName
		}
		if actualObligation.TypeParam != expectedObligation.TypeParam ||
			!typesEquivalent(actualObligation.Subject, substituteType(expectedObligation.Subject, substitution), allowUnknown) ||
			!typesEquivalent(actualObligation.Constraint, substituteType(expectedObligation.Constraint, substitution), allowUnknown) {
			return false
		}
	}
	return true
}
