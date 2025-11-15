package typechecker

import (
	"fmt"
	"math/big"
	"strings"
)

// typeName returns a human-readable identifier for a type, tolerating nil.
func typeName(t Type) string {
	return formatType(t)
}

func formatType(t Type) string {
	if t == nil {
		return "unknown"
	}

	switch val := t.(type) {
	case UnknownType:
		return "unknown"
	case PrimitiveType:
		switch val.Kind {
		case PrimitiveBool:
			return "bool"
		case PrimitiveChar:
			return "char"
		case PrimitiveString:
			return "string"
		case PrimitiveNil:
			return "nil"
		case PrimitiveInt:
			return "int"
		case PrimitiveFloat:
			return "float"
		default:
			return strings.ToLower(string(val.Kind))
		}
	case IntegerType:
		if val.Suffix != "" {
			return val.Suffix
		}
		return "int"
	case FloatType:
		if val.Suffix != "" {
			return val.Suffix
		}
		return "float"
	case TypeParameterType:
		if val.ParameterName == "" {
			return "unknown"
		}
		return val.ParameterName
	case StructType:
		return val.StructName
	case StructInstanceType:
		return val.StructName
	case InterfaceType:
		return val.InterfaceName
	case AliasType:
		target := formatType(val.Target)
		if target == "" || target == "<unknown>" {
			target = typeName(val.Target)
		}
		if target == "" {
			target = "<unknown>"
		}
		return "type alias -> " + target
	case UnionType:
		return val.UnionName
	case ArrayType:
		elem := formatType(val.Element)
		return strings.TrimSpace("Array " + elem)
	case NullableType:
		return formatType(val.Inner) + "?"
	case RangeType:
		return strings.TrimSpace("Range " + formatType(val.Element))
	case IteratorType:
		return strings.TrimSpace("Iterator " + formatType(val.Element))
	case ProcType:
		return strings.TrimSpace("Proc " + formatType(val.Result))
	case FutureType:
		return strings.TrimSpace("Future " + formatType(val.Result))
	case AppliedType:
		base := formatType(val.Base)
		if len(val.Arguments) == 0 {
			return base
		}
		args := make([]string, len(val.Arguments))
		for i, arg := range val.Arguments {
			args[i] = formatType(arg)
		}
		return strings.TrimSpace(base + " " + strings.Join(args, " "))
	case UnionLiteralType:
		if len(val.Members) == 0 {
			return "Union"
		}
		members := make([]string, len(val.Members))
		for i, member := range val.Members {
			members[i] = formatType(member)
		}
		return strings.Join(members, " | ")
	case FunctionType:
		params := make([]string, len(val.Params))
		for i, param := range val.Params {
			params[i] = formatType(param)
		}
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(params, ", "), formatType(val.Return))
	}

	return t.Name()
}

type intBounds struct {
	min    *big.Int
	max    *big.Int
	bits   int
	signed bool
}

func signedBounds(bits int) intBounds {
	max := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(bits-1)), nil)
	max.Sub(max, big.NewInt(1))
	min := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(bits-1)), nil)
	min.Neg(min)
	return intBounds{min: min, max: max, bits: bits, signed: true}
}

func unsignedBounds(bits int) intBounds {
	max := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(bits)), nil)
	max.Sub(max, big.NewInt(1))
	return intBounds{min: big.NewInt(0), max: max, bits: bits, signed: false}
}

var signedIntegerOrder = []string{"i8", "i16", "i32", "i64", "i128"}
var unsignedIntegerOrder = []string{"u8", "u16", "u32", "u64", "u128"}

var integerBounds = map[string]intBounds{
	"i8":   signedBounds(8),
	"i16":  signedBounds(16),
	"i32":  signedBounds(32),
	"i64":  signedBounds(64),
	"i128": signedBounds(128),
	"u8":   unsignedBounds(8),
	"u16":  unsignedBounds(16),
	"u32":  unsignedBounds(32),
	"u64":  unsignedBounds(64),
	"u128": unsignedBounds(128),
}

func integerInfo(name string) (intBounds, bool) {
	info, ok := integerBounds[name]
	return info, ok
}

func isSignedInteger(name string) bool {
	info, ok := integerBounds[name]
	return ok && info.signed
}

func integerBitsFor(name string) (int, bool) {
	info, ok := integerBounds[name]
	if !ok {
		return 0, false
	}
	return info.bits, true
}

func smallestSignedFor(bits int) (string, bool) {
	for _, name := range signedIntegerOrder {
		info := integerBounds[name]
		if info.bits >= bits {
			return name, true
		}
	}
	return "", false
}

func smallestUnsignedFor(bits int) (string, bool) {
	for _, name := range unsignedIntegerOrder {
		info := integerBounds[name]
		if info.bits >= bits {
			return name, true
		}
	}
	return "", false
}

func isUnknownType(t Type) bool {
	if t == nil {
		return true
	}
	_, ok := t.(UnknownType)
	return ok
}

func isTypeParameter(t Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.(TypeParameterType)
	return ok
}

func isIntegerType(t Type) bool {
	if t == nil {
		return false
	}
	switch val := t.(type) {
	case IntegerType:
		return true
	case PrimitiveType:
		return val.Kind == PrimitiveInt
	default:
		return false
	}
}

func isFloatType(t Type) bool {
	if t == nil {
		return false
	}
	switch v := t.(type) {
	case FloatType:
		return true
	case PrimitiveType:
		return v.Kind == PrimitiveFloat
	}
	return false
}

func isNumericType(t Type) bool {
	return isIntegerType(t) || isFloatType(t)
}

func isBoolType(t Type) bool {
	if t == nil {
		return false
	}
	if prim, ok := t.(PrimitiveType); ok {
		return prim.Kind == PrimitiveBool
	}
	return false
}

func isStringType(t Type) bool {
	if t == nil {
		return false
	}
	if prim, ok := t.(PrimitiveType); ok {
		return prim.Kind == PrimitiveString
	}
	return false
}

func isInterfaceLikeType(t Type) bool {
	if t == nil {
		return false
	}
	switch v := t.(type) {
	case InterfaceType:
		return true
	case AppliedType:
		return isInterfaceLikeType(v.Base)
	default:
		return false
	}
}

func isPrimitiveInt(t Type) bool {
	if t == nil {
		return false
	}
	if prim, ok := t.(PrimitiveType); ok {
		return prim.Kind == PrimitiveInt
	}
	return false
}

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
		if nullable, ok := from.(NullableType); ok {
			return typeAssignable(nullable.Inner, target.Inner)
		}
		return typeAssignable(from, target.Inner)
	case UnionLiteralType:
		return unionAssignable(from, target)
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
		for _, member := range source.Members {
			if !typeAssignable(member, to) {
				return false
			}
		}
		return true
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
	if applied, ok := resultAppliedType(expected); ok {
		if typeAssignable(actual, expected) {
			return actual, true
		}
		success := applied.Arguments[0]
		if typeAssignable(actual, success) {
			return applied, true
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
	var members []Type
	for _, t := range types {
		if t == nil || isUnknownType(t) {
			continue
		}
		members = appendUnionMember(members, t)
	}
	if len(members) == 0 {
		return UnknownType{}
	}
	if len(members) == 1 {
		return members[0]
	}
	return UnionLiteralType{Members: members}
}

func appendUnionMember(existing []Type, candidate Type) []Type {
	if candidate == nil {
		return existing
	}
	switch v := candidate.(type) {
	case UnionLiteralType:
		for _, member := range v.Members {
			existing = appendUnionMember(existing, member)
		}
		return existing
	default:
		for _, member := range existing {
			if sameType(member, candidate) {
				return existing
			}
		}
		return append(existing, candidate)
	}
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

func iterableElementType(t Type) (Type, bool) {
	if t == nil {
		return UnknownType{}, true
	}
	if _, ok := t.(UnknownType); ok {
		return UnknownType{}, true
	}
	if elem, ok := arrayElementType(t); ok {
		if elem == nil || isUnknownType(elem) {
			return UnknownType{}, true
		}
		return elem, true
	}
	if rng, ok := t.(RangeType); ok {
		if rng.Element == nil || isUnknownType(rng.Element) {
			return UnknownType{}, true
		}
		return rng.Element, true
	}
	if iter, ok := t.(IteratorType); ok {
		if iter.Element == nil || isUnknownType(iter.Element) {
			return UnknownType{}, true
		}
		return iter.Element, true
	}
	if applied, ok := t.(AppliedType); ok {
		switch base := applied.Base.(type) {
		case InterfaceType:
			if base.InterfaceName == "Iterable" {
				if len(applied.Arguments) == 0 {
					return UnknownType{}, true
				}
				elem := applied.Arguments[0]
				if elem == nil || isUnknownType(elem) {
					return UnknownType{}, true
				}
				return elem, true
			}
		}
	}
	return UnknownType{}, false
}

func structName(t Type) (string, bool) {
	switch s := t.(type) {
	case StructType:
		return s.StructName, true
	case StructInstanceType:
		return s.StructName, true
	case AppliedType:
		if base, ok := s.Base.(StructType); ok {
			return base.StructName, true
		}
	}
	return "", false
}

func unionName(t Type) (string, bool) {
	switch u := t.(type) {
	case UnionType:
		return u.UnionName, u.UnionName != ""
	case AppliedType:
		return unionName(u.Base)
	}
	return "", false
}

func arrayElementType(t Type) (Type, bool) {
	switch arr := t.(type) {
	case ArrayType:
		return arr.Element, true
	case StructType:
		if arr.StructName == "Array" {
			if len(arr.Positional) > 0 {
				return arr.Positional[0], true
			}
			return UnknownType{}, true
		}
	case StructInstanceType:
		if arr.StructName == "Array" {
			if len(arr.Positional) > 0 {
				return arr.Positional[0], true
			}
			return UnknownType{}, true
		}
	case AppliedType:
		if name, ok := structName(arr.Base); ok && name == "Array" {
			if len(arr.Arguments) > 0 {
				return arr.Arguments[0], true
			}
			return UnknownType{}, true
		}
	}
	return nil, false
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

func typeAssignableToAny(from Type, targets []Type) bool {
	for _, target := range targets {
		if typeAssignable(from, target) {
			return true
		}
	}
	return false
}

func substituteFunctionType(fn FunctionType, subst map[string]Type) FunctionType {
	if len(subst) == 0 {
		return fn
	}
	params := make([]Type, len(fn.Params))
	for i, param := range fn.Params {
		params[i] = substituteType(param, subst)
	}
	ret := substituteType(fn.Return, subst)
	where := substituteWhereSpecs(fn.Where, subst)
	obligations := substituteObligations(fn.Obligations, subst)
	typeParams := fn.TypeParams
	if len(typeParams) > 0 {
		filtered := make([]GenericParamSpec, 0, len(typeParams))
		for _, param := range typeParams {
			if param.Name == "" {
				filtered = append(filtered, param)
				continue
			}
			if _, ok := subst[param.Name]; ok {
				continue
			}
			filtered = append(filtered, param)
		}
		typeParams = filtered
	}
	return FunctionType{
		Params:      params,
		Return:      ret,
		TypeParams:  typeParams,
		Where:       where,
		Obligations: obligations,
	}
}

func substituteType(t Type, subst map[string]Type) Type {
	if t == nil {
		return nil
	}
	switch v := t.(type) {
	case TypeParameterType:
		if replacement, ok := subst[v.ParameterName]; ok {
			return replacement
		}
		return v
	case FunctionType:
		return substituteFunctionType(v, subst)
	case ArrayType:
		return ArrayType{Element: substituteType(v.Element, subst)}
	case NullableType:
		return NullableType{Inner: substituteType(v.Inner, subst)}
	case RangeType:
		return RangeType{Element: substituteType(v.Element, subst)}
	case UnionType:
		params := make([]GenericParamSpec, len(v.TypeParams))
		for i, param := range v.TypeParams {
			constraints := make([]Type, len(param.Constraints))
			for j, constraint := range param.Constraints {
				constraints[j] = substituteType(constraint, subst)
			}
			params[i] = GenericParamSpec{
				Name:        param.Name,
				Constraints: constraints,
			}
		}
		where := substituteWhereSpecs(v.Where, subst)
		variants := make([]Type, len(v.Variants))
		for i, variant := range v.Variants {
			variants[i] = substituteType(variant, subst)
		}
		return UnionType{
			UnionName:  v.UnionName,
			TypeParams: params,
			Where:      where,
			Variants:   variants,
		}
	case AppliedType:
		base := substituteType(v.Base, subst)
		args := make([]Type, len(v.Arguments))
		for i, arg := range v.Arguments {
			args[i] = substituteType(arg, subst)
		}
		return AppliedType{Base: base, Arguments: args}
	case UnionLiteralType:
		members := make([]Type, len(v.Members))
		for i, member := range v.Members {
			members[i] = substituteType(member, subst)
		}
		return UnionLiteralType{Members: members}
	case ProcType:
		return ProcType{Result: substituteType(v.Result, subst)}
	case FutureType:
		return FutureType{Result: substituteType(v.Result, subst)}
	}
	return t
}

func substituteWhereSpecs(specs []WhereConstraintSpec, subst map[string]Type) []WhereConstraintSpec {
	if len(specs) == 0 || len(subst) == 0 {
		return specs
	}
	out := make([]WhereConstraintSpec, 0, len(specs))
	for _, spec := range specs {
		if spec.TypeParam != "" {
			if _, ok := subst[spec.TypeParam]; ok {
				// This where-clause references a type parameter that has been
				// substituted with a concrete type; drop the clause because the
				// obligation is now captured via the substituted constraints.
				continue
			}
		}
		constraints := make([]Type, len(spec.Constraints))
		for j, constraint := range spec.Constraints {
			constraints[j] = substituteType(constraint, subst)
		}
		out = append(out, WhereConstraintSpec{
			TypeParam:   spec.TypeParam,
			Constraints: constraints,
		})
	}
	return out
}

func substituteObligations(obligations []ConstraintObligation, subst map[string]Type) []ConstraintObligation {
	if len(obligations) == 0 || len(subst) == 0 {
		return obligations
	}
	out := make([]ConstraintObligation, len(obligations))
	for i, ob := range obligations {
		var subject Type
		if ob.Subject != nil {
			subject = substituteType(ob.Subject, subst)
		} else if replacement, ok := subst[ob.TypeParam]; ok {
			subject = replacement
		}
		out[i] = ConstraintObligation{
			Owner:      ob.Owner,
			TypeParam:  ob.TypeParam,
			Constraint: substituteType(ob.Constraint, subst),
			Subject:    subject,
			Context:    ob.Context,
			Node:       ob.Node,
		}
	}
	return out
}

func instantiateAlias(alias AliasType, args []Type) (Type, map[string]Type) {
	subst := make(map[string]Type, len(alias.TypeParams))
	for idx, param := range alias.TypeParams {
		if param.Name == "" {
			continue
		}
		if idx < len(args) && args[idx] != nil {
			subst[param.Name] = args[idx]
			continue
		}
		if _, exists := subst[param.Name]; !exists {
			subst[param.Name] = UnknownType{}
		}
	}
	target := alias.Target
	if len(subst) > 0 {
		target = substituteType(target, subst)
	}
	if target == nil {
		target = UnknownType{}
	}
	return target, subst
}

func populateObligationSubjects(obligations []ConstraintObligation, subject Type) []ConstraintObligation {
	if len(obligations) == 0 || subject == nil || isUnknownType(subject) {
		return obligations
	}
	out := make([]ConstraintObligation, len(obligations))
	for i, ob := range obligations {
		if ob.Subject != nil && !isUnknownType(ob.Subject) {
			out[i] = ob
			continue
		}
		out[i] = ob
		out[i].Subject = subject
	}
	return out
}
