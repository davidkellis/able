package interpreter

import "able/interpreter-go/pkg/runtime"

func inlineCoerceValueBySimpleType(typeName string, value runtime.Value) (runtime.Value, bool, error) {
	if typeName == "" {
		return nil, false, nil
	}
	typeName = normalizeKernelAliasName(typeName)

	targetInt := runtime.IntegerType(typeName)
	if _, ok := lookupIntegerInfo(targetInt); ok {
		if coerced, ok := coerceIntegerValueToTargetKindIfInRange(value, targetInt); ok {
			return coerced, true, nil
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
			f := integerValueToFloat64Fast(val)
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, true, nil
		case *runtime.IntegerValue:
			if val == nil {
				return nil, false, nil
			}
			f := integerRefToFloat64Fast(val)
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, true, nil
		}
	}

	return nil, false, nil
}
