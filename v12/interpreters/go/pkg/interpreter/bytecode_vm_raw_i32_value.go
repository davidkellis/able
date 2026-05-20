package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeRawI32SlotValue int32

func (v bytecodeRawI32SlotValue) Kind() runtime.Kind {
	return runtime.KindInteger
}

func bytecodeBoxRawI32Value(value bytecodeRawI32SlotValue) runtime.Value {
	return bytecodeBoxedIntegerI32Value(int64(value))
}
