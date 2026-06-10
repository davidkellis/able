package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) arrayMember(arr *runtime.ArrayValue, member ast.Expression) (runtime.Value, error) {
	if arr == nil {
		return nil, fmt.Errorf("array receiver is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("array member access expects identifier")
	}
	switch ident.Name {
	case "storage_handle":
		handle := arrayValueHandle(arr)
		if boxed, ok := boxedSmallIntValue(runtime.IntegerI64, handle); ok {
			return boxed, nil
		}
		return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
	case "length":
		return arrayLengthMemberValue(arr)
	case "capacity":
		return arrayCapacityMemberValue(arr)
	case "iterator":
		fn := runtime.NativeFunctionValue{
			Name:       "array.iterator",
			Arity:      0,
			BorrowArgs: true,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("iterator expects only a receiver")
				}
				receiver, ok := args[0].(*runtime.ArrayValue)
				if !ok {
					return nil, fmt.Errorf("iterator receiver must be an array")
				}
				index := 0
				iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
					current, err := i.ensureArrayState(receiver, 0)
					if err != nil {
						return nil, true, err
					}
					if index >= len(current.Values) {
						return runtime.IteratorEnd, true, nil
					}
					val := current.Values[index]
					index++
					if val == nil {
						return runtime.NilValue{}, false, nil
					}
					return val, false, nil
				}, nil)
				return iter, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: arr, Method: fn}, nil
	default:
		return nil, fmt.Errorf("array has no member '%s' (import able.collections.array for stdlib helpers)", ident.Name)
	}
}

func arrayValueHandle(arr *runtime.ArrayValue) int64 {
	if arr == nil {
		return 0
	}
	if arr.Handle != 0 {
		return arr.Handle
	}
	return arr.TrackedHandle
}

func arrayCurrentTrackedState(arr *runtime.ArrayValue) *arrayState {
	if arr == nil || arr.State == nil {
		return nil
	}
	if arr.Handle == 0 {
		return arr.State
	}
	if arr.TrackedHandle == arr.Handle {
		return arr.State
	}
	return nil
}

func boxedArrayMetadataI32Value(value int) runtime.Value {
	raw := int64(value)
	if boxed, ok := boxedSmallIntValue(runtime.IntegerI32, raw); ok {
		return boxed
	}
	if boxed, ok := boxedExtendedI32Value(raw); ok {
		return boxed
	}
	return runtime.NewSmallInt(raw, runtime.IntegerI32)
}

func arrayLengthMemberValue(arr *runtime.ArrayValue) (runtime.Value, error) {
	if state := arrayCurrentTrackedState(arr); state != nil {
		return state.BoxedLengthValue(), nil
	}
	if handle := arrayValueHandle(arr); handle != 0 {
		size, err := runtime.ArrayStoreSize(handle)
		if err != nil {
			return nil, err
		}
		return boxedArrayMetadataI32Value(size), nil
	}
	return boxedArrayMetadataI32Value(len(arr.Elements)), nil
}

func arrayCapacityMemberValue(arr *runtime.ArrayValue) (runtime.Value, error) {
	if state := arrayCurrentTrackedState(arr); state != nil {
		return state.BoxedCapacityValue(), nil
	}
	if handle := arrayValueHandle(arr); handle != 0 {
		capacity, err := runtime.ArrayStoreCapacity(handle)
		if err != nil {
			return nil, err
		}
		return boxedArrayMetadataI32Value(capacity), nil
	}
	return boxedArrayMetadataI32Value(cap(arr.Elements)), nil
}

func isDirectArrayMemberName(name string) bool {
	switch name {
	case "storage_handle", "length", "capacity", "iterator":
		return true
	default:
		return false
	}
}

func arrayIndexFromValue(val runtime.Value) (int, error) {
	v, ok := hostIntegerValue(val)
	if !ok {
		return 0, fmt.Errorf("array index must be an integer")
	}
	if v.Sign() < 0 {
		return 0, fmt.Errorf("array index must be non-negative")
	}
	res, ok := v.ToInt64()
	if !ok {
		return 0, fmt.Errorf("array index out of range")
	}
	if res > math.MaxInt {
		return 0, fmt.Errorf("array index out of range")
	}
	return int(res), nil
}

func makeIndexError(index int, length int) runtime.Value {
	payload := map[string]runtime.Value{
		"index":  runtime.NewSmallInt(int64(index), runtime.IntegerI64),
		"length": runtime.NewSmallInt(int64(length), runtime.IntegerI64),
	}
	message := fmt.Sprintf("index %d out of bounds for length %d", index, length)
	return runtime.ErrorValue{
		TypeName: ast.NewIdentifier("IndexError"),
		Payload:  payload,
		Message:  message,
	}
}
