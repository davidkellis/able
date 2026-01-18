package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) initRatioBuiltins() {
	if i.ratioReady {
		return
	}
	ratioFromFloat := runtime.NativeFunctionValue{
		Name:  "__able_ratio_from_float",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_ratio_from_float expects one argument")
			}
			val, ok := args[0].(runtime.FloatValue)
			if !ok {
				return nil, fmt.Errorf("__able_ratio_from_float expects float input")
			}
			parts, err := ratioFromFloatValue(val)
			if err != nil {
				return nil, err
			}
			return i.makeRatioValue(parts)
		},
	}

	f32Bits := runtime.NativeFunctionValue{
		Name:  "__able_f32_bits",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_f32_bits expects one argument")
			}
			switch val := args[0].(type) {
			case runtime.FloatValue:
				if val.TypeSuffix != runtime.FloatF32 {
					return nil, fmt.Errorf("__able_f32_bits expects f32 input")
				}
				bits := math.Float32bits(float32(val.Val))
				return runtime.IntegerValue{Val: new(big.Int).SetUint64(uint64(bits)), TypeSuffix: runtime.IntegerU32}, nil
			case *runtime.FloatValue:
				if val == nil || val.TypeSuffix != runtime.FloatF32 {
					return nil, fmt.Errorf("__able_f32_bits expects f32 input")
				}
				bits := math.Float32bits(float32(val.Val))
				return runtime.IntegerValue{Val: new(big.Int).SetUint64(uint64(bits)), TypeSuffix: runtime.IntegerU32}, nil
			default:
				return nil, fmt.Errorf("__able_f32_bits expects f32 input")
			}
		},
	}

	f64Bits := runtime.NativeFunctionValue{
		Name:  "__able_f64_bits",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_f64_bits expects one argument")
			}
			switch val := args[0].(type) {
			case runtime.FloatValue:
				if val.TypeSuffix != runtime.FloatF64 {
					return nil, fmt.Errorf("__able_f64_bits expects f64 input")
				}
				bits := math.Float64bits(val.Val)
				return runtime.IntegerValue{Val: new(big.Int).SetUint64(bits), TypeSuffix: runtime.IntegerU64}, nil
			case *runtime.FloatValue:
				if val == nil || val.TypeSuffix != runtime.FloatF64 {
					return nil, fmt.Errorf("__able_f64_bits expects f64 input")
				}
				bits := math.Float64bits(val.Val)
				return runtime.IntegerValue{Val: new(big.Int).SetUint64(bits), TypeSuffix: runtime.IntegerU64}, nil
			default:
				return nil, fmt.Errorf("__able_f64_bits expects f64 input")
			}
		},
	}

	u64Mul := runtime.NativeFunctionValue{
		Name:  "__able_u64_mul",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_u64_mul expects two arguments")
			}
			lhs, err := expectU64(args[0])
			if err != nil {
				return nil, err
			}
			rhs, err := expectU64(args[1])
			if err != nil {
				return nil, err
			}
			product := lhs * rhs
			return runtime.IntegerValue{Val: new(big.Int).SetUint64(product), TypeSuffix: runtime.IntegerU64}, nil
		},
	}
	i.global.Define("__able_ratio_from_float", ratioFromFloat)
	i.global.Define("__able_f32_bits", f32Bits)
	i.global.Define("__able_f64_bits", f64Bits)
	i.global.Define("__able_u64_mul", u64Mul)
	i.ratioReady = true
}

func expectU64(value runtime.Value) (uint64, error) {
	switch v := value.(type) {
	case runtime.IntegerValue:
		if v.TypeSuffix != runtime.IntegerU64 || v.Val == nil || !v.Val.IsUint64() {
			return 0, fmt.Errorf("__able_u64_mul expects u64 inputs")
		}
		return v.Val.Uint64(), nil
	case *runtime.IntegerValue:
		if v == nil || v.TypeSuffix != runtime.IntegerU64 || v.Val == nil || !v.Val.IsUint64() {
			return 0, fmt.Errorf("__able_u64_mul expects u64 inputs")
		}
		return v.Val.Uint64(), nil
	default:
		return 0, fmt.Errorf("__able_u64_mul expects u64 inputs")
	}
}
