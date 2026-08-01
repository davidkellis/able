package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

func primitiveFixedIntegerArithmeticMethod(
	typeName string,
	methodName string,
) (runtime.FixedIntegerMode, runtime.FixedIntegerOperation, bool) {
	var mode runtime.FixedIntegerMode
	switch {
	case strings.HasPrefix(methodName, "wrapping_"):
		mode = runtime.FixedIntegerWrapping
	case strings.HasPrefix(methodName, "saturating_"):
		mode = runtime.FixedIntegerSaturating
	case strings.HasPrefix(methodName, "checked_"):
		mode = runtime.FixedIntegerChecked
	default:
		return 0, 0, false
	}

	var operation runtime.FixedIntegerOperation
	switch methodName[strings.IndexByte(methodName, '_')+1:] {
	case "add":
		operation = runtime.FixedIntegerAdd
	case "sub":
		operation = runtime.FixedIntegerSub
	case "mul":
		operation = runtime.FixedIntegerMul
	default:
		return 0, 0, false
	}

	switch typeName {
	case "i8", "i16", "i32", "i64", "i128",
		"u8", "u16", "u32", "u64", "u128":
		return mode, operation, true
	default:
		return 0, 0, false
	}
}

func (i *Interpreter) primitiveFixedIntegerArithmeticNativeMethod(
	typeName string,
	methodName string,
) (runtime.NativeFunctionValue, bool) {
	mode, operation, ok := primitiveFixedIntegerArithmeticMethod(typeName, methodName)
	if !ok {
		return runtime.NativeFunctionValue{}, false
	}
	return runtime.NativeFunctionValue{
		Name:        fmt.Sprintf("%s.%s", typeName, methodName),
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("%s expects a receiver and one argument", methodName)
			}
			left, ok := primitiveCanonicalValue(args[0]).(runtime.IntegerValue)
			if !ok || string(left.TypeSuffix) != typeName {
				return nil, fmt.Errorf("%s receiver must be %s", methodName, typeName)
			}
			right, ok := primitiveCanonicalValue(args[1]).(runtime.IntegerValue)
			if !ok || right.TypeSuffix != left.TypeSuffix {
				return nil, fmt.Errorf("%s argument must be %s", methodName, typeName)
			}
			result, present, err := runtime.FixedIntegerArithmetic(left, right, operation, mode)
			if err != nil {
				return nil, err
			}
			if !present {
				return runtime.NilValue{}, nil
			}
			return result, nil
		},
	}, true
}
