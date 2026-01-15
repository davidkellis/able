package typechecker

import (
	"fmt"
	"math/big"
)

// typeAssignable performs a shallow compatibility check between two types.
// It intentionally permits Unknown/TypeParam targets so later passes can refine them.
func typeAssignable(from, to Type) bool {
	if to != nil {
		if name, ok := structName(to); ok && name == "void" {
			return true
		}
	}
	from = normalizeSpecialType(from)
	to = normalizeSpecialType(to)
	if to == nil || isUnknownType(to) {
		return true
	}
	if isTypeParameter(to) {
		// Type parameters accept any argument for now; constraint solving happens later.
		return true
	}
	if from == nil {
		return false
	}
	if isUnknownType(from) {
		return true
	}
	if isTypeParameter(from) {
		return true
	}
	if literalAssignableTo(from, to) {
		return true
	}
	if sourceInt, ok := from.(IntegerType); ok {
		if targetInt, ok := to.(IntegerType); ok {
			if integerRangeWithin(sourceInt.Suffix, targetInt.Suffix) {
				return true
			}
		}
	}
	switch target := to.(type) {
	case StructType:
		if name, ok := structName(from); ok {
			return name == target.StructName
		}
		return false
	case StructInstanceType:
		if name, ok := structName(from); ok {
			return name == target.StructName
		}
		return false
	case ArrayType:
		if elem, ok := arrayElementType(from); ok {
			return typeAssignable(elem, target.Element)
		}
		return false
	case RangeType:
		if rng, ok := from.(RangeType); ok {
			return typeAssignable(rng.Element, target.Element)
		}
		return false
	case IteratorType:
		if iter, ok := from.(IteratorType); ok {
			return typeAssignable(iter.Element, target.Element)
		}
		return false
	case MapType:
		if sourceMap, ok := from.(MapType); ok {
			return typeAssignable(sourceMap.Key, target.Key) && typeAssignable(sourceMap.Value, target.Value)
		}
		return false
	case NullableType:
		if prim, ok := from.(PrimitiveType); ok && prim.Kind == PrimitiveNil {
			return true
		}
		if nullable, ok := from.(NullableType); ok {
			return typeAssignable(nullable.Inner, target.Inner)
		}
		return typeAssignable(from, target.Inner)
	case UnionLiteralType:
		return unionAssignable(from, target)
	case UnionType:
		if union, ok := from.(UnionType); ok {
			if union.UnionName != "" && target.UnionName != "" && union.UnionName == target.UnionName {
				return true
			}
			return unionLiteralAssignableToNamed(UnionLiteralType{Members: union.Variants}, target)
		}
		if literal, ok := from.(UnionLiteralType); ok {
			return unionLiteralAssignableToNamed(literal, target)
		}
		return typeAssignableToAny(from, target.Variants)
	case AppliedType:
		if applied, ok := from.(AppliedType); ok {
			return appliedTypesAssignable(applied, target)
		}
		if name, ok := structName(from); ok {
			if base, ok := target.Base.(StructType); ok && base.StructName == name {
				return true
			}
		}
		return false
	}

	switch source := from.(type) {
	case StructType:
		if name, ok := structName(to); ok {
			return source.StructName == name
		}
	case StructInstanceType:
		if name, ok := structName(to); ok {
			return source.StructName == name
		}
	case AppliedType:
		if targetApplied, ok := to.(AppliedType); ok {
			return appliedTypesAssignable(source, targetApplied)
		}
		if name, ok := structName(to); ok {
			if base, ok := source.Base.(StructType); ok {
				return base.StructName == name
			}
		}
	case MapType:
		if targetMap, ok := to.(MapType); ok {
			return typeAssignable(source.Key, targetMap.Key) && typeAssignable(source.Value, targetMap.Value)
		}
	case ArrayType:
		if elem, ok := arrayElementType(to); ok {
			return typeAssignable(source.Element, elem)
		}
	case RangeType:
		if rng, ok := to.(RangeType); ok {
			return typeAssignable(source.Element, rng.Element)
		}
	case IteratorType:
		if iter, ok := to.(IteratorType); ok {
			return typeAssignable(source.Element, iter.Element)
		}
	case NullableType:
		return typeAssignable(source.Inner, to)
	case UnionLiteralType:
		if targetUnion, ok := to.(UnionLiteralType); ok {
			return unionAssignable(source, targetUnion)
		}
		if targetNamed, ok := to.(UnionType); ok {
			return unionLiteralAssignableToNamed(source, targetNamed)
		}
		for _, member := range source.Members {
			if !typeAssignable(member, to) {
				return false
			}
		}
		return true
	case UnionType:
		if targetNamed, ok := to.(UnionType); ok {
			if source.UnionName != "" && targetNamed.UnionName != "" {
				return source.UnionName == targetNamed.UnionName
			}
			return unionLiteralAssignableToNamed(UnionLiteralType{Members: source.Variants}, targetNamed)
		}
		if targetUnion, ok := to.(UnionLiteralType); ok {
			return unionLiteralAssignableToNamed(targetUnion, source)
		}
		for _, variant := range source.Variants {
			if typeAssignable(variant, to) {
				return true
			}
		}
		return false
	}

	return from.Name() == to.Name()
}

func literalAssignableTo(from, to Type) bool {
	if from == nil || to == nil {
		return false
	}
	from = normalizeSpecialType(from)
	to = normalizeSpecialType(to)
	source, ok := from.(IntegerType)
	if !ok || source.Literal == nil {
		return false
	}
	if target, ok := to.(FloatType); ok {
		if source.Explicit {
			return false
		}
		return target.Suffix == "f32" || target.Suffix == "f64"
	}
	target, ok := to.(IntegerType)
	if !ok {
		return false
	}
	if source.Explicit {
		return source.Suffix == target.Suffix
	}
	bounds, ok := integerBounds[target.Suffix]
	if !ok {
		return source.Suffix == target.Suffix
	}
	value := new(big.Int).Set(source.Literal)
	return value.Cmp(bounds.min) >= 0 && value.Cmp(bounds.max) <= 0
}

func integerRangeWithin(sourceSuffix, targetSuffix string) bool {
	sourceBounds, ok := integerBounds[sourceSuffix]
	if !ok {
		return false
	}
	targetBounds, ok := integerBounds[targetSuffix]
	if !ok {
		return false
	}
	return sourceBounds.min.Cmp(targetBounds.min) >= 0 && sourceBounds.max.Cmp(targetBounds.max) <= 0
}

func literalMismatchMessage(from, to Type) (string, bool) {
	if from == nil || to == nil {
		return "", false
	}
	from = normalizeSpecialType(from)
	to = normalizeSpecialType(to)
	switch actual := from.(type) {
	case ArrayType:
		if expected, ok := to.(ArrayType); ok {
			return literalMismatchMessage(actual.Element, expected.Element)
		}
	case MapType:
		if expected, ok := to.(MapType); ok {
			if msg, ok := literalMismatchMessage(actual.Key, expected.Key); ok {
				return msg, true
			}
			return literalMismatchMessage(actual.Value, expected.Value)
		}
	case RangeType:
		if expected, ok := to.(RangeType); ok {
			if msg, ok := literalMismatchMessage(actual.Element, expected.Element); ok {
				return msg, true
			}
			for _, bound := range actual.Bounds {
				if msg, ok := literalMismatchMessage(bound, expected.Element); ok {
					return msg, true
				}
			}
			return "", false
		}
	case IteratorType:
		if expected, ok := to.(IteratorType); ok {
			return literalMismatchMessage(actual.Element, expected.Element)
		}
	case ProcType:
		if expected, ok := to.(ProcType); ok {
			return literalMismatchMessage(actual.Result, expected.Result)
		}
	case FutureType:
		if expected, ok := to.(FutureType); ok {
			return literalMismatchMessage(actual.Result, expected.Result)
		}
	case NullableType:
		if expected, ok := to.(NullableType); ok {
			return literalMismatchMessage(actual.Inner, expected.Inner)
		}
	case UnionLiteralType:
		if expected, ok := to.(UnionLiteralType); ok {
			limit := len(actual.Members)
			if len(expected.Members) < limit {
				limit = len(expected.Members)
			}
			for i := 0; i < limit; i++ {
				if msg, ok := literalMismatchMessage(actual.Members[i], expected.Members[i]); ok {
					return msg, true
				}
			}
			return "", false
		}
		for _, member := range actual.Members {
			if msg, ok := literalMismatchMessage(member, to); ok {
				return msg, true
			}
		}
		return "", false
	}
	if expectedUnion, ok := to.(UnionLiteralType); ok {
		for _, member := range expectedUnion.Members {
			if msg, ok := literalMismatchMessage(from, member); ok {
				return msg, true
			}
		}
		return "", false
	}
	source, ok := from.(IntegerType)
	if !ok || source.Literal == nil || source.Explicit {
		return "", false
	}
	target, ok := to.(IntegerType)
	if !ok {
		return "", false
	}
	bounds, ok := integerBounds[target.Suffix]
	if !ok {
		return "", false
	}
	value := new(big.Int).Set(source.Literal)
	if value.Cmp(bounds.min) < 0 || value.Cmp(bounds.max) > 0 {
		return fmt.Sprintf("literal %s does not fit in %s", value.String(), target.Suffix), true
	}
	return "", false
}

func normalizeResultReturn(actual, expected Type) (Type, bool) {
	if actual == nil {
		actual = PrimitiveType{Kind: PrimitiveNil}
	}
	if success, ok := resultSuccessType(expected); ok {
		if typeAssignable(actual, expected) {
			return actual, true
		}
		if typeAssignable(actual, success) {
			return expected, true
		}
		return actual, false
	}
	if typeAssignable(actual, expected) {
		return actual, true
	}
	return actual, false
}

func resultAppliedType(t Type) (AppliedType, bool) {
	applied, ok := t.(AppliedType)
	if !ok {
		return AppliedType{}, false
	}
	name, ok := structName(applied.Base)
	if !ok || name != "Result" {
		return AppliedType{}, false
	}
	if len(applied.Arguments) == 0 || applied.Arguments[0] == nil {
		return AppliedType{}, false
	}
	return applied, true
}

func resultSuccessType(t Type) (Type, bool) {
	if applied, ok := resultAppliedType(t); ok {
		return applied.Arguments[0], true
	}
	if union, ok := t.(UnionType); ok && union.UnionName == "Result" {
		for _, variant := range union.Variants {
			if isResultErrorVariant(variant) {
				continue
			}
			return variant, true
		}
		return UnknownType{}, true
	}
	return nil, false
}

func isResultErrorVariant(t Type) bool {
	if iface, _, ok := interfaceFromType(t); ok {
		return iface.InterfaceName == "Error"
	}
	if name, ok := structName(t); ok && name == "Error" {
		return true
	}
	return false
}

func mergeBranchTypes(types []Type) Type {
	var result Type = UnknownType{}
	for _, t := range types {
		if t == nil || isUnknownType(t) {
			continue
		}
		if isUnknownType(result) {
			result = t
			continue
		}
		if result.Name() != t.Name() {
			return UnknownType{}
		}
	}
	return result
}

func mergeCompatibleTypes(a, b Type) Type {
	if a == nil || isUnknownType(a) {
		if b == nil {
			return UnknownType{}
		}
		return b
	}
	if b == nil || isUnknownType(b) {
		return a
	}
	if typeAssignable(b, a) {
		return a
	}
	if typeAssignable(a, b) {
		return b
	}
	return UnknownType{}
}

func mergeCompatibleTypesSlice(types ...Type) Type {
	var result Type = UnknownType{}
	for _, t := range types {
		if t == nil {
			continue
		}
		result = mergeCompatibleTypes(result, t)
		if isUnknownType(result) && t != nil && !isUnknownType(t) {
			result = t
		}
	}
	if result == nil {
		return UnknownType{}
	}
	return result
}

func mergeTypesAllowUnion(a, b Type) Type {
	if a == nil || isUnknownType(a) {
		return b
	}
	if b == nil || isUnknownType(b) {
		return a
	}
	if typeAssignable(b, a) {
		return a
	}
	if typeAssignable(a, b) {
		return b
	}
	return buildUnionType(a, b)
}

func buildUnionType(types ...Type) Type {
	return normalizeUnionTypes(types)
}

func sameType(a, b Type) bool {
	if a == nil || b == nil {
		return false
	}
	a = normalizeSpecialType(a)
	b = normalizeSpecialType(b)
	if isUnknownType(a) || isUnknownType(b) {
		return false
	}
	if a.Name() == b.Name() {
		return true
	}
	switch av := a.(type) {
	case AppliedType:
		if bv, ok := b.(AppliedType); ok {
			if !sameType(av.Base, bv.Base) {
				return false
			}
			if len(av.Arguments) != len(bv.Arguments) {
				return false
			}
			for i := range av.Arguments {
				if !sameType(av.Arguments[i], bv.Arguments[i]) {
					return false
				}
			}
			return true
		}
	case ArrayType:
		if bv, ok := b.(ArrayType); ok {
			return sameType(av.Element, bv.Element)
		}
	case MapType:
		if bv, ok := b.(MapType); ok {
			return sameType(av.Key, bv.Key) && sameType(av.Value, bv.Value)
		}
	case RangeType:
		if bv, ok := b.(RangeType); ok {
			return sameType(av.Element, bv.Element)
		}
	case IteratorType:
		if bv, ok := b.(IteratorType); ok {
			return sameType(av.Element, bv.Element)
		}
	case ProcType:
		if bv, ok := b.(ProcType); ok {
			return sameType(av.Result, bv.Result)
		}
	case FutureType:
		if bv, ok := b.(FutureType); ok {
			return sameType(av.Result, bv.Result)
		}
	case NullableType:
		if bv, ok := b.(NullableType); ok {
			return sameType(av.Inner, bv.Inner)
		}
	case UnionLiteralType:
		if bv, ok := b.(UnionLiteralType); ok {
			if len(av.Members) != len(bv.Members) {
				return false
			}
			for i := range av.Members {
				if !sameType(av.Members[i], bv.Members[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func normalizeSpecialType(t Type) Type {
	if t == nil {
		return nil
	}
	switch v := t.(type) {
	case AppliedType:
		if name, ok := structName(v); ok {
			if converted, ok := convertSpecialAppliedType(name, v.Arguments); ok {
				return converted
			}
		}
	case StructType:
		if converted, ok := convertSpecialAppliedType(v.StructName, v.Positional); ok {
			return converted
		}
	case StructInstanceType:
		args := v.TypeArgs
		if len(args) == 0 {
			args = v.Positional
		}
		if converted, ok := convertSpecialAppliedType(v.StructName, args); ok {
			return converted
		}
	}
	return t
}

func convertSpecialAppliedType(name string, args []Type) (Type, bool) {
	switch name {
	case "Array":
		return ArrayType{Element: argumentOrUnknown(args, 0)}, true
	case "Iterator":
		return IteratorType{Element: argumentOrUnknown(args, 0)}, true
	case "Range":
		return RangeType{Element: argumentOrUnknown(args, 0)}, true
	case "Map":
		return MapType{Key: argumentOrUnknown(args, 0), Value: argumentOrUnknown(args, 1)}, true
	case "HashMap":
		return MapType{Key: argumentOrUnknown(args, 0), Value: argumentOrUnknown(args, 1)}, true
	case "Proc":
		return ProcType{Result: argumentOrUnknown(args, 0)}, true
	case "Future":
		return FutureType{Result: argumentOrUnknown(args, 0)}, true
	default:
		return nil, false
	}
}

func argumentOrUnknown(args []Type, idx int) Type {
	if idx >= 0 && idx < len(args) {
		if args[idx] != nil {
			return args[idx]
		}
	}
	return UnknownType{}
}

func appliedTypesAssignable(from, to AppliedType) bool {
	if !typeAssignable(from.Base, to.Base) {
		return false
	}
	if len(from.Arguments) != len(to.Arguments) {
		return false
	}
	for i := range from.Arguments {
		if !typeAssignable(from.Arguments[i], to.Arguments[i]) {
			return false
		}
	}
	return true
}

func unionAssignable(from Type, to UnionLiteralType) bool {
	if union, ok := from.(UnionLiteralType); ok {
		for _, member := range union.Members {
			if !typeAssignableToAny(member, to.Members) {
				return false
			}
		}
		return true
	}
	return typeAssignableToAny(from, to.Members)
}

func unionLiteralAssignableToNamed(source UnionLiteralType, target UnionType) bool {
	if len(target.Variants) == 0 {
		return true
	}
	for _, member := range source.Members {
		if !typeAssignableToAny(member, target.Variants) {
			return false
		}
	}
	return true
}

func typeAssignableToAny(from Type, targets []Type) bool {
	for _, target := range targets {
		if typeAssignable(from, target) {
			return true
		}
	}
	return false
}
