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
		return "Unknown"
	}

	switch val := t.(type) {
	case UnknownType:
		return "Unknown"
	case PrimitiveType:
		switch val.Kind {
		case PrimitiveBool:
			return "bool"
		case PrimitiveChar:
			return "char"
		case PrimitiveString:
			return "String"
		case PrimitiveIoHandle:
			return "IoHandle"
		case PrimitiveProcHandle:
			return "ProcHandle"
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
			return "Unknown"
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
	case FunctionOverloadType:
		return "<function overload>"
	case ImplementationNamespaceType:
		if val.Impl != nil && val.Impl.ImplName != "" {
			return val.Impl.ImplName
		}
		return "implementation"
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

func isUnknownFunctionSignature(fn FunctionType) bool {
	return len(fn.Params) == 0 && isUnknownType(fn.Return)
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

func typeCanBeNil(t Type) bool {
	if t == nil {
		return true
	}
	switch val := t.(type) {
	case UnknownType:
		return true
	case PrimitiveType:
		return val.Kind == PrimitiveNil
	case NullableType:
		return true
	case UnionLiteralType:
		for _, member := range val.Members {
			if typeCanBeNil(member) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func isNumericType(t Type) bool {
	if isRatioType(t) {
		return true
	}
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

func isRatioType(t Type) bool {
	switch v := t.(type) {
	case StructType:
		return v.StructName == "Ratio"
	case StructInstanceType:
		return v.StructName == "Ratio"
	case AppliedType:
		return isRatioType(v.Base)
	default:
		return false
	}
}

func isResultType(t Type) bool {
	if t == nil {
		return false
	}
	if name, ok := unionName(t); ok && name == "Result" {
		return true
	}
	if applied, ok := t.(AppliedType); ok {
		if name, ok := structName(applied.Base); ok && name == "Result" {
			return true
		}
	}
	return false
}
