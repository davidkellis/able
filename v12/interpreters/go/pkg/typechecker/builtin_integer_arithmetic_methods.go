package typechecker

import "strings"

func builtinFixedIntegerArithmeticMethod(object Type, name string) (FunctionType, bool) {
	integer, ok := object.(IntegerType)
	if !ok || !isFixedIntegerSuffix(integer.Suffix) {
		return FunctionType{}, false
	}
	mode, _, ok := splitFixedIntegerArithmeticMethod(name)
	if !ok {
		return FunctionType{}, false
	}

	valueType := IntegerType{Suffix: integer.Suffix}
	returnType := Type(valueType)
	if mode == "checked" {
		returnType = NullableType{Inner: valueType}
	}
	return FunctionType{
		Params: []Type{valueType},
		Return: returnType,
	}, true
}

func splitFixedIntegerArithmeticMethod(name string) (mode string, operation string, ok bool) {
	for _, candidateMode := range []string{"wrapping", "saturating", "checked"} {
		prefix := candidateMode + "_"
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		candidateOperation := strings.TrimPrefix(name, prefix)
		switch candidateOperation {
		case "add", "sub", "mul":
			return candidateMode, candidateOperation, true
		default:
			return "", "", false
		}
	}
	return "", "", false
}

func isFixedIntegerSuffix(suffix string) bool {
	switch suffix {
	case "i8", "i16", "i32", "i64", "i128",
		"u8", "u16", "u32", "u64", "u128":
		return true
	default:
		return false
	}
}
