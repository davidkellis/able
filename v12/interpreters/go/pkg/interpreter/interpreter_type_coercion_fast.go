package interpreter

import (
	"fmt"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) coerceValueToTypeWouldBeNoOp(typeExpr ast.TypeExpression) bool {
	if typeExpr == nil {
		return true
	}
	switch t := typeExpr.(type) {
	case *ast.SimpleTypeExpression:
		return false
	case *ast.GenericTypeExpression:
		base, ok := t.Base.(*ast.SimpleTypeExpression)
		if !ok || base == nil || base.Name == nil {
			return true
		}
		_, isInterface := i.interfaces[normalizeKernelAliasName(base.Name.Name)]
		return !isInterface
	default:
		return true
	}
}

func castValueToCanonicalSimpleTypeFast(typeName string, rawValue runtime.Value) (runtime.Value, bool, error) {
	switch val := rawValue.(type) {
	case runtime.IntegerValue:
		if string(val.TypeSuffix) == typeName {
			return rawValue, true, nil
		}
	case *runtime.IntegerValue:
		if val != nil && string(val.TypeSuffix) == typeName {
			return rawValue, true, nil
		}
	case runtime.FloatValue:
		if string(val.TypeSuffix) == typeName {
			return rawValue, true, nil
		}
	case *runtime.FloatValue:
		if val != nil && string(val.TypeSuffix) == typeName {
			return rawValue, true, nil
		}
	}

	switch typeName {
	case "String":
		switch rawValue.(type) {
		case runtime.StringValue, *runtime.StringValue:
			return rawValue, true, nil
		}
		return nil, false, nil
	case "bool", "Bool":
		switch rawValue.(type) {
		case runtime.BoolValue, *runtime.BoolValue:
			return rawValue, true, nil
		}
		return nil, false, nil
	case "char":
		switch rawValue.(type) {
		case runtime.CharValue, *runtime.CharValue:
			return rawValue, true, nil
		}
		return nil, false, nil
	case "Error":
		switch rawValue.(type) {
		case runtime.ErrorValue, *runtime.ErrorValue:
			return rawValue, true, nil
		}
		return nil, false, nil
	}

	targetKind := runtime.IntegerType(typeName)
	if info, ok := lookupIntegerInfo(targetKind); ok {
		switch val := rawValue.(type) {
		case runtime.IntegerValue:
			if val.TypeSuffix == targetKind {
				return rawValue, true, nil
			}
			wrapped := patternToInteger(bitPattern(val.BigInt(), info), info)
			if wrapped.IsInt64() {
				return boxedOrSmallIntegerValue(targetKind, wrapped.Int64()), true, nil
			}
			return runtime.NewBigIntValue(new(big.Int).Set(wrapped), targetKind), true, nil
		case *runtime.IntegerValue:
			if val == nil {
				return nil, true, fmt.Errorf("cannot cast <nil> to %s", targetKind)
			}
			if val.TypeSuffix == targetKind {
				return rawValue, true, nil
			}
			wrapped := patternToInteger(bitPattern(val.BigInt(), info), info)
			if wrapped.IsInt64() {
				return boxedOrSmallIntegerValue(targetKind, wrapped.Int64()), true, nil
			}
			return runtime.NewBigIntValue(new(big.Int).Set(wrapped), targetKind), true, nil
		case runtime.FloatValue:
			casted, err := castFloatValueToInteger(targetKind, info, val.Val)
			return casted, true, err
		case *runtime.FloatValue:
			if val == nil {
				return nil, true, fmt.Errorf("cannot cast <nil> to %s", targetKind)
			}
			casted, err := castFloatValueToInteger(targetKind, info, val.Val)
			return casted, true, err
		}
		return nil, false, nil
	}

	if typeName == "f32" || typeName == "f64" {
		targetFloat := runtime.FloatType(typeName)
		switch val := rawValue.(type) {
		case runtime.FloatValue:
			if val.TypeSuffix == targetFloat {
				return rawValue, true, nil
			}
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, val.Val), TypeSuffix: targetFloat}, true, nil
		case *runtime.FloatValue:
			if val == nil {
				return nil, true, fmt.Errorf("cannot cast <nil> to %s", typeName)
			}
			if val.TypeSuffix == targetFloat {
				return rawValue, true, nil
			}
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, val.Val), TypeSuffix: targetFloat}, true, nil
		case runtime.IntegerValue:
			f := bigIntToFloat(val.BigInt())
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, true, nil
		case *runtime.IntegerValue:
			if val == nil {
				return nil, true, fmt.Errorf("cannot cast <nil> to %s", typeName)
			}
			f := bigIntToFloat(val.BigInt())
			return runtime.FloatValue{Val: normalizeFloat(targetFloat, f), TypeSuffix: targetFloat}, true, nil
		}
		return nil, false, nil
	}

	return nil, false, nil
}
