//go:build !(js && wasm)

package interpreter

import (
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
	"able/interpreter-go/pkg/typechecker"
)

func (i *Interpreter) contextualNumericLiteralValue(literal *ast.IntegerLiteral) (runtime.Value, bool, error) {
	if i == nil || literal == nil || literal.IntegerType != nil || literal.Value == nil {
		return nil, false, nil
	}
	inferred := i.runtimeInferenceFactsSnapshot()[literal]
	switch target := inferred.(type) {
	case typechecker.IntegerType:
		if target.Suffix == "" || target.Suffix == string(runtime.IntegerI32) {
			return nil, false, nil
		}
		kind := runtime.IntegerType(target.Suffix)
		info, err := getIntegerInfo(kind)
		if err != nil {
			return nil, true, err
		}
		value := new(big.Int).Set(literal.Value)
		if err := ensureFitsInteger(info, value); err != nil {
			return nil, true, err
		}
		return runtime.NewBigIntValue(value, kind), true, nil
	case typechecker.FloatType:
		value, _ := new(big.Float).SetInt(literal.Value).Float64()
		kind := runtime.FloatType(target.Suffix)
		if kind == runtime.FloatF32 {
			value = float64(float32(value))
		}
		return runtime.FloatValue{Val: value, TypeSuffix: kind}, true, nil
	default:
		return nil, false, nil
	}
}
