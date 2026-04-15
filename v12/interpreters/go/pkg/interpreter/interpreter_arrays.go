package interpreter

import (
	"fmt"
	"math"
	"sync"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type arrayState = runtime.ArrayState

const arrayMetadataBoxU64Max int64 = 16384

var (
	arrayMetadataBoxU64Once sync.Once
	arrayMetadataBoxedU64   []runtime.Value
)

func initArrayMetadataU64BoxCache() {
	size := int(arrayMetadataBoxU64Max) + 1
	arrayMetadataBoxedU64 = make([]runtime.Value, size)
	for idx := range arrayMetadataBoxedU64 {
		arrayMetadataBoxedU64[idx] = runtime.NewSmallInt(int64(idx), runtime.IntegerU64)
	}
}

func boxedArrayMetadataU64Value(value int64) (runtime.Value, bool) {
	if value < 0 || value > arrayMetadataBoxU64Max {
		return nil, false
	}
	arrayMetadataBoxU64Once.Do(initArrayMetadataU64BoxCache)
	return arrayMetadataBoxedU64[int(value)], true
}

func (i *Interpreter) trackArrayValue(handle int64, arr *runtime.ArrayValue) {
	if arr == nil || handle == 0 {
		return
	}
	if arr.Handle != 0 && arr.Handle != handle {
		i.untrackArrayValue(arr.Handle, arr)
	}
	if i.arraysByHandle == nil {
		i.arraysByHandle = make(map[int64]arrayHandleTracking)
	}
	arr.Handle = handle
	arr.TrackedHandle = handle
	tracking := i.arraysByHandle[handle]
	switch {
	case tracking.single == arr:
	case tracking.many != nil:
		tracking.many[arr] = struct{}{}
	case tracking.single == nil:
		tracking.single = arr
	default:
		tracking.many = map[*runtime.ArrayValue]struct{}{
			tracking.single: {},
			arr:             {},
		}
		tracking.single = nil
	}
	i.arraysByHandle[handle] = tracking
}

func (i *Interpreter) untrackArrayValue(handle int64, arr *runtime.ArrayValue) {
	if i == nil || arr == nil || handle == 0 || i.arraysByHandle == nil {
		return
	}
	tracking, ok := i.arraysByHandle[handle]
	if !ok {
		return
	}
	switch {
	case tracking.single == arr:
		tracking.single = nil
	case tracking.many != nil:
		delete(tracking.many, arr)
		if len(tracking.many) == 1 {
			for only := range tracking.many {
				tracking.single = only
			}
			tracking.many = nil
		}
	}
	if tracking.single == nil && len(tracking.many) == 0 {
		if arr.TrackedHandle == handle {
			arr.TrackedHandle = 0
		}
		delete(i.arraysByHandle, handle)
		return
	}
	i.arraysByHandle[handle] = tracking
}

func (i *Interpreter) syncArrayValues(handle int64, state *arrayState) {
	if state == nil || i.arraysByHandle == nil {
		return
	}
	token, ok := bytecodeArrayElementTypeTokenFromValues(state.Values)
	state.ElementTypeToken = token
	state.ElementTypeTokenKnown = ok
	tracking, ok := i.arraysByHandle[handle]
	if !ok {
		return
	}
	if tracking.single != nil {
		tracking.single.Handle = handle
		tracking.single.TrackedHandle = handle
		tracking.single.State = state
		tracking.single.Elements = state.Values
		return
	}
	for arr := range tracking.many {
		if arr == nil {
			continue
		}
		arr.Handle = handle
		arr.TrackedHandle = handle
		arr.State = state
		arr.Elements = state.Values
	}
}

func (i *Interpreter) ensureArrayBuiltins() {
	if i.arrayReady {
		return
	}
	i.initArrayBuiltins()
}

func (i *Interpreter) arrayStateForHandle(handle int64) (*arrayState, error) {
	return runtime.ArrayStoreState(handle)
}

func ensureArrayCapacity(state *arrayState, minimum int) bool {
	return runtime.ArrayEnsureCapacity(state, minimum)
}

func setArrayLength(state *arrayState, length int) {
	runtime.ArraySetLength(state, length)
}

func (i *Interpreter) ensureArrayState(arr *runtime.ArrayValue, capacityHint int) (*arrayState, error) {
	if arr == nil {
		return nil, fmt.Errorf("array receiver is nil")
	}
	i.ensureArrayBuiltins()
	if arr.State != nil && arr.Handle != 0 && arr.TrackedHandle == arr.Handle && capacityHint <= arr.State.Capacity {
		return arr.State, nil
	}
	state, handle, err := runtime.ArrayStoreEnsure(arr, capacityHint)
	if err != nil {
		return nil, err
	}
	arr.State = state
	i.trackArrayValue(handle, arr)
	i.syncArrayValues(handle, state)
	return state, nil
}

// ArrayElements exposes array state access for compiled interop.
func (i *Interpreter) ArrayElements(arr *runtime.ArrayValue) ([]runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("interpreter: nil interpreter")
	}
	state, err := i.ensureArrayState(arr, 0)
	if err != nil {
		return nil, err
	}
	return state.Values, nil
}

func (i *Interpreter) arrayValueFromHandle(handle int64, lengthHint int, capacityHint int) (*runtime.ArrayValue, error) {
	if handle == 0 {
		return nil, fmt.Errorf("array handle must be non-zero")
	}
	i.ensureArrayBuiltins()
	arr, state, err := runtime.ArrayStoreValueFromHandle(handle, lengthHint, capacityHint)
	if err != nil {
		return nil, err
	}
	arr.State = state
	i.syncArrayValues(handle, state)
	i.trackArrayValue(handle, arr)
	return arr, nil
}

func (i *Interpreter) newArrayValue(elements []runtime.Value, capacityHint int) *runtime.ArrayValue {
	if capacityHint < len(elements) {
		capacityHint = len(elements)
	}
	arr := &runtime.ArrayValue{Elements: elements}
	if _, err := i.ensureArrayState(arr, capacityHint); err != nil {
		return arr
	}
	return arr
}

func (i *Interpreter) arrayValueFromStructFields(fields map[string]runtime.Value) (*runtime.ArrayValue, error) {
	var handle int64
	var length int
	var capacity int
	if fields != nil {
		if hv, ok := fields["storage_handle"]; ok {
			if intVal, ok := hv.(runtime.IntegerValue); ok {
				if h, ok := intVal.ToInt64(); ok {
					handle = h
				}
			}
		}
		if lv, ok := fields["length"]; ok {
			if l, err := arrayIndexFromValue(lv); err == nil {
				length = l
			}
		}
		if cv, ok := fields["capacity"]; ok {
			if c, err := arrayIndexFromValue(cv); err == nil {
				capacity = c
			}
		}
	}
	if capacity < length {
		capacity = length
	}
	if handle != 0 {
		return i.arrayValueFromHandle(handle, length, capacity)
	}
	return i.newArrayValue(make([]runtime.Value, length, capacity), capacity), nil
}

func (i *Interpreter) initArrayBuiltins() {
	if i.arrayReady {
		return
	}
	if i.arraysByHandle == nil {
		i.arraysByHandle = make(map[int64]arrayHandleTracking)
	}

	parseArrayHandle := func(val runtime.Value) (int64, error) {
		intVal, ok := val.(runtime.IntegerValue)
		if !ok {
			return 0, fmt.Errorf("array handle must be an integer")
		}
		n, ok := intVal.ToInt64()
		if !ok {
			return 0, fmt.Errorf("array handle is out of range")
		}
		return n, nil
	}

	arrayNewHandle := runtime.NativeFunctionValue{
		Name:       "__able_array_new",
		Arity:      0,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("__able_array_new expects no arguments")
			}
			handle := runtime.ArrayStoreNew()
			return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
		},
	}

	arrayWithCapacity := runtime.NativeFunctionValue{
		Name:       "__able_array_with_capacity",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_with_capacity expects capacity argument")
			}
			capacity, err := arrayIndexFromValue(args[0])
			if err != nil {
				return nil, fmt.Errorf("capacity must be a non-negative integer")
			}
			if capacity < 0 {
				capacity = 0
			}
			handle := runtime.ArrayStoreNewWithCapacity(capacity)
			return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
		},
	}

	arraySize := runtime.NativeFunctionValue{
		Name:       "__able_array_size",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_size expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			size, err := runtime.ArrayStoreSize(handle)
			if err != nil {
				return nil, err
			}
			sizeVal := int64(size)
			if boxed, ok := boxedArrayMetadataU64Value(sizeVal); ok {
				return boxed, nil
			}
			return runtime.NewSmallInt(sizeVal, runtime.IntegerU64), nil
		},
	}

	arrayCapacity := runtime.NativeFunctionValue{
		Name:       "__able_array_capacity",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_capacity expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			capacity, err := runtime.ArrayStoreCapacity(handle)
			if err != nil {
				return nil, err
			}
			capacityVal := int64(capacity)
			if boxed, ok := boxedArrayMetadataU64Value(capacityVal); ok {
				return boxed, nil
			}
			return runtime.NewSmallInt(capacityVal, runtime.IntegerU64), nil
		},
	}

	arraySetLen := runtime.NativeFunctionValue{
		Name:       "__able_array_set_len",
		Arity:      2,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_array_set_len expects handle and length")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			length, err := arrayIndexFromValue(args[1])
			if err != nil {
				return nil, fmt.Errorf("length must be a non-negative integer")
			}
			if err := runtime.ArrayStoreSetLength(handle, length); err != nil {
				return nil, err
			}
			if state, err := runtime.ArrayStoreState(handle); err == nil {
				i.syncArrayValues(handle, state)
			}
			return runtime.NilValue{}, nil
		},
	}

	arrayRead := runtime.NativeFunctionValue{
		Name:       "__able_array_read",
		Arity:      2,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_array_read expects handle and index")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			idx, err := arrayIndexFromValue(args[1])
			if err != nil {
				return nil, err
			}
			val, err := runtime.ArrayStoreRead(handle, idx)
			if err != nil {
				return nil, err
			}
			return val, nil
		},
	}

	arrayWrite := runtime.NativeFunctionValue{
		Name:       "__able_array_write",
		Arity:      3,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("__able_array_write expects handle, index, and value")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			idx, err := arrayIndexFromValue(args[1])
			if err != nil {
				return nil, err
			}
			if idx < 0 {
				return nil, fmt.Errorf("index must be non-negative")
			}
			if err := runtime.ArrayStoreWrite(handle, idx, args[2]); err != nil {
				return nil, err
			}
			if state, err := runtime.ArrayStoreState(handle); err == nil {
				i.syncArrayValues(handle, state)
			}
			return runtime.NilValue{}, nil
		},
	}

	arrayReserve := runtime.NativeFunctionValue{
		Name:       "__able_array_reserve",
		Arity:      2,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_array_reserve expects handle and capacity")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			minCapacity, err := arrayIndexFromValue(args[1])
			if err != nil {
				return nil, fmt.Errorf("capacity must be a non-negative integer")
			}
			if err := runtime.ArrayStoreReserve(handle, minCapacity); err != nil {
				return nil, err
			}
			if state, err := runtime.ArrayStoreState(handle); err == nil {
				i.syncArrayValues(handle, state)
			}
			return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
		},
	}

	arrayClone := runtime.NativeFunctionValue{
		Name:       "__able_array_clone",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_clone expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			newHandle, err := runtime.ArrayStoreClone(handle)
			if err != nil {
				return nil, err
			}
			return runtime.NewSmallInt(newHandle, runtime.IntegerI64), nil
		},
	}

	arrayPkg := &runtime.PackageValue{
		Name:   "Array",
		Public: make(map[string]runtime.Value),
	}

	arrayNew := runtime.NativeFunctionValue{
		Name:       "Array.new",
		Arity:      0,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			capacity := 0
			if len(args) > 1 {
				return nil, fmt.Errorf("Array.new expects zero or one argument")
			}
			if len(args) == 1 {
				val, err := arrayIndexFromValue(args[0])
				if err != nil {
					return nil, fmt.Errorf("Array.new capacity must be a non-negative integer")
				}
				if val < 0 {
					return nil, fmt.Errorf("Array.new capacity must be non-negative")
				}
				capacity = val
			}
			if capacity < 0 {
				capacity = 0
			}
			return i.newArrayValue(make([]runtime.Value, 0, capacity), capacity), nil
		},
	}

	arrayPkg.Public["new"] = arrayNew
	i.global.Define("__able_array_new", arrayNewHandle)
	i.global.Define("__able_array_with_capacity", arrayWithCapacity)
	i.global.Define("__able_array_size", arraySize)
	i.global.Define("__able_array_capacity", arrayCapacity)
	i.global.Define("__able_array_set_len", arraySetLen)
	i.global.Define("__able_array_read", arrayRead)
	i.global.Define("__able_array_write", arrayWrite)
	i.global.Define("__able_array_reserve", arrayReserve)
	i.global.Define("__able_array_clone", arrayClone)
	i.global.Define("Array", arrayPkg)
	i.arrayReady = true
}

func (i *Interpreter) arrayMember(arr *runtime.ArrayValue, member ast.Expression) (runtime.Value, error) {
	if arr == nil {
		return nil, fmt.Errorf("array receiver is nil")
	}
	state, err := i.ensureArrayState(arr, 0)
	if err != nil {
		return nil, err
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("array member access expects identifier")
	}
	switch ident.Name {
	case "storage_handle":
		if boxed, ok := boxedSmallIntValue(runtime.IntegerI64, arr.Handle); ok {
			return boxed, nil
		}
		return runtime.NewSmallInt(arr.Handle, runtime.IntegerI64), nil
	case "length":
		length := int64(len(state.Values))
		if boxed, ok := boxedSmallIntValue(runtime.IntegerI32, length); ok {
			return boxed, nil
		}
		return runtime.NewSmallInt(length, runtime.IntegerI32), nil
	case "capacity":
		capacity := int64(state.Capacity)
		if boxed, ok := boxedSmallIntValue(runtime.IntegerI32, capacity); ok {
			return boxed, nil
		}
		return runtime.NewSmallInt(capacity, runtime.IntegerI32), nil
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

func arrayIndexFromValue(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
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
	default:
		return 0, fmt.Errorf("array index must be an integer")
	}
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
