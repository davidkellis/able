package interpreter

import (
	"math/big"

	"able/interpreter-go/pkg/runtime"
)

func inlineCoerceValueBySimpleType(typeName string, value runtime.Value) (runtime.Value, bool, error) {
	if typeName == "" {
		return nil, false, nil
	}
	typeName = normalizeKernelAliasName(typeName)

	targetInt := runtime.IntegerType(typeName)
	if _, ok := lookupIntegerInfo(targetInt); ok {
		switch val := value.(type) {
		case runtime.IntegerValue:
			if val.TypeSuffix == targetInt {
				return value, true, nil
			}
			if integerRangeWithinKinds(val.TypeSuffix, targetInt) {
				if intVal, ok := val.ToInt64(); ok {
					return boxedOrSmallIntegerValue(targetInt, intVal), true, nil
				}
				return runtime.NewBigIntValue(new(big.Int).Set(val.BigInt()), targetInt), true, nil
			}
		case *runtime.IntegerValue:
			if val == nil {
				return nil, false, nil
			}
			if val.TypeSuffix == targetInt {
				return value, true, nil
			}
			if integerRangeWithinKinds(val.TypeSuffix, targetInt) {
				if intVal, ok := val.ToInt64(); ok {
					return boxedOrSmallIntegerValue(targetInt, intVal), true, nil
				}
				return runtime.NewBigIntValue(new(big.Int).Set(val.BigInt()), targetInt), true, nil
			}
		}
		return nil, false, nil
	}

	if typeName == "f32" || typeName == "f64" {
		targetFloat := runtime.FloatType(typeName)
		switch val := value.(type) {
		case runtime.FloatValue:
			if val.TypeSuffix == targetFloat {
				return value, true, nil
			}
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, val.Val), TypeSuffix: targetFloat}, true, nil
		case *runtime.FloatValue:
			if val == nil {
				return nil, false, nil
			}
			if val.TypeSuffix == targetFloat {
				return value, true, nil
			}
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, val.Val), TypeSuffix: targetFloat}, true, nil
		case runtime.IntegerValue:
			f := bigIntToFloat(val.BigInt())
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, true, nil
		case *runtime.IntegerValue:
			if val == nil {
				return nil, false, nil
			}
			f := bigIntToFloat(val.BigInt())
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, true, nil
		}
	}

	return nil, false, nil
}
