package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeRawF32SlotValue float32

func (v bytecodeRawF32SlotValue) Kind() runtime.Kind { return runtime.KindFloat }

type bytecodeRawF64SlotValue float64

func (v bytecodeRawF32SlotValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v bytecodeRawF64SlotValue) MaterializeRuntimeValue() runtime.Value {
	return bytecodeMaterializeRawValue(v)
}

func (v bytecodeRawF64SlotValue) Kind() runtime.Kind { return runtime.KindFloat }

func bytecodeNormalizedRawFloatSlotValue(value float64, kind runtime.FloatType) runtime.Value {
	switch kind {
	case runtime.FloatF32:
		return bytecodeRawF32SlotValue(float32(value))
	default:
		return bytecodeRawF64SlotValue(value)
	}
}

func bytecodeSetNormalizedRawFloatValue(dst *runtime.Value, value float64, kind runtime.FloatType) {
	if dst == nil {
		return
	}
	*dst = bytecodeNormalizedRawFloatSlotValue(value, kind)
}

func bytecodeRawFloatSlotValue(value float64, kind runtime.FloatType) runtime.Value {
	return bytecodeNormalizedRawFloatSlotValue(normalizeFloat(kind, value), kind)
}

func bytecodeDirectRawFloatValue(value runtime.Value) (float64, runtime.FloatType, bool) {
	switch fv := value.(type) {
	case bytecodeRawF32SlotValue:
		return float64(fv), runtime.FloatF32, true
	case bytecodeRawF64SlotValue:
		return float64(fv), runtime.FloatF64, true
	}
	return 0, runtime.FloatF64, false
}

func bytecodeMaterializeRawFloatValue(value runtime.Value) runtime.Value {
	switch fv := value.(type) {
	case bytecodeRawF32SlotValue:
		return runtime.FloatValue{Val: float64(fv), TypeSuffix: runtime.FloatF32}
	case bytecodeRawF64SlotValue:
		return runtime.FloatValue{Val: float64(fv), TypeSuffix: runtime.FloatF64}
	default:
		return value
	}
}
