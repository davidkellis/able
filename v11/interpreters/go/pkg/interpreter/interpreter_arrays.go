package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

type arrayState struct {
	values   []runtime.Value
	capacity int
}

func (i *Interpreter) ensureArrayBuiltins() {
	if i.arrayReady {
		return
	}
	i.initArrayBuiltins()
}

func (i *Interpreter) arrayStateForHandle(handle int64) (*arrayState, error) {
	if i.arrayStates == nil {
		i.arrayStates = make(map[int64]*arrayState)
	}
	state, ok := i.arrayStates[handle]
	if !ok {
		return nil, fmt.Errorf("array handle %d is not defined", handle)
	}
	return state, nil
}

func ensureArrayCapacity(state *arrayState, minimum int) {
	if minimum <= state.capacity {
		return
	}
	newValues := make([]runtime.Value, len(state.values), minimum)
	copy(newValues, state.values)
	state.values = newValues
	state.capacity = minimum
}

func setArrayLength(state *arrayState, length int) {
	if length < 0 {
		return
	}
	if length <= len(state.values) {
		state.values = state.values[:length]
		return
	}
	for len(state.values) < length {
		state.values = append(state.values, runtime.NilValue{})
	}
}

func (i *Interpreter) initArrayBuiltins() {
	if i.arrayReady {
		return
	}

	if i.arrayStates == nil {
		i.arrayStates = make(map[int64]*arrayState)
	}
	if i.nextArrayHandle == 0 {
		i.nextArrayHandle = 1
	}

	parseArrayHandle := func(val runtime.Value) (int64, error) {
		intVal, ok := val.(runtime.IntegerValue)
		if !ok || intVal.Val == nil {
			return 0, fmt.Errorf("array handle must be an integer")
		}
		if !intVal.Val.IsInt64() {
			return 0, fmt.Errorf("array handle is out of range")
		}
		return intVal.Val.Int64(), nil
	}

	arrayNewHandle := runtime.NativeFunctionValue{
		Name:  "__able_array_new",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("__able_array_new expects no arguments")
			}
			handle := i.nextArrayHandle
			i.nextArrayHandle++
			i.arrayStates[handle] = &arrayState{values: make([]runtime.Value, 0), capacity: 0}
			return runtime.IntegerValue{Val: big.NewInt(handle), TypeSuffix: runtime.IntegerI64}, nil
		},
	}

	arrayWithCapacity := runtime.NativeFunctionValue{
		Name:  "__able_array_with_capacity",
		Arity: 1,
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
			handle := i.nextArrayHandle
			i.nextArrayHandle++
			i.arrayStates[handle] = &arrayState{values: make([]runtime.Value, 0, capacity), capacity: capacity}
			return runtime.IntegerValue{Val: big.NewInt(handle), TypeSuffix: runtime.IntegerI64}, nil
		},
	}

	arraySize := runtime.NativeFunctionValue{
		Name:  "__able_array_size",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_size expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.arrayStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			return runtime.IntegerValue{Val: big.NewInt(int64(len(state.values))), TypeSuffix: runtime.IntegerU64}, nil
		},
	}

	arrayCapacity := runtime.NativeFunctionValue{
		Name:  "__able_array_capacity",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_capacity expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.arrayStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			return runtime.IntegerValue{Val: big.NewInt(int64(state.capacity)), TypeSuffix: runtime.IntegerU64}, nil
		},
	}

	arraySetLen := runtime.NativeFunctionValue{
		Name:  "__able_array_set_len",
		Arity: 2,
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
			state, err := i.arrayStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			ensureArrayCapacity(state, length)
			setArrayLength(state, length)
			return runtime.NilValue{}, nil
		},
	}

	arrayRead := runtime.NativeFunctionValue{
		Name:  "__able_array_read",
		Arity: 2,
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
			state, err := i.arrayStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			if idx < 0 || idx >= len(state.values) {
				return runtime.NilValue{}, nil
			}
			return state.values[idx], nil
		},
	}

	arrayWrite := runtime.NativeFunctionValue{
		Name:  "__able_array_write",
		Arity: 3,
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
			state, err := i.arrayStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			ensureArrayCapacity(state, idx+1)
			if idx >= len(state.values) {
				setArrayLength(state, idx+1)
			}
			state.values[idx] = args[2]
			return runtime.NilValue{}, nil
		},
	}

	arrayReserve := runtime.NativeFunctionValue{
		Name:  "__able_array_reserve",
		Arity: 2,
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
			state, err := i.arrayStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			ensureArrayCapacity(state, minCapacity)
			return runtime.IntegerValue{Val: big.NewInt(int64(state.capacity)), TypeSuffix: runtime.IntegerU64}, nil
		},
	}

	arrayClone := runtime.NativeFunctionValue{
		Name:  "__able_array_clone",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_array_clone expects handle")
			}
			handle, err := parseArrayHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.arrayStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			cloned := make([]runtime.Value, len(state.values))
			copy(cloned, state.values)
			newHandle := i.nextArrayHandle
			i.nextArrayHandle++
			i.arrayStates[newHandle] = &arrayState{values: cloned, capacity: state.capacity}
			return runtime.IntegerValue{Val: big.NewInt(newHandle), TypeSuffix: runtime.IntegerI64}, nil
		},
	}

	arrayPkg := &runtime.PackageValue{
		Name:   "Array",
		Public: make(map[string]runtime.Value),
	}

	arrayNew := runtime.NativeFunctionValue{
		Name:  "Array.new",
		Arity: 0,
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
			arr := &runtime.ArrayValue{Elements: make([]runtime.Value, 0, capacity)}
			return arr, nil
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
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("array member access expects identifier")
	}
	switch ident.Name {
	case "push":
		fn := runtime.NativeFunctionValue{
			Name:  "array.push",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("push expects a receiver and one argument")
				}
				receiver, ok := args[0].(*runtime.ArrayValue)
				if !ok {
					return nil, fmt.Errorf("push receiver must be an array")
				}
				receiver.Elements = append(receiver.Elements, args[1])
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: arr, Method: fn}, nil
	case "pop":
		fn := runtime.NativeFunctionValue{
			Name:  "array.pop",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("pop expects only a receiver")
				}
				receiver, ok := args[0].(*runtime.ArrayValue)
				if !ok {
					return nil, fmt.Errorf("pop receiver must be an array")
				}
				length := len(receiver.Elements)
				if length == 0 {
					return runtime.NilValue{}, nil
				}
				last := receiver.Elements[length-1]
				receiver.Elements = receiver.Elements[:length-1]
				return last, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: arr, Method: fn}, nil
	case "clone":
		fn := runtime.NativeFunctionValue{
			Name:  "array.clone",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("clone expects only a receiver")
				}
				receiver, ok := args[0].(*runtime.ArrayValue)
				if !ok {
					return nil, fmt.Errorf("clone receiver must be an array")
				}
				length := len(receiver.Elements)
				copied := make([]runtime.Value, length)
				copy(copied, receiver.Elements)
				return &runtime.ArrayValue{Elements: copied}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: arr, Method: fn}, nil
	case "size":
		fn := runtime.NativeFunctionValue{
			Name:  "array.size",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("size expects only a receiver")
				}
				receiver, ok := args[0].(*runtime.ArrayValue)
				if !ok {
					return nil, fmt.Errorf("size receiver must be an array")
				}
				length := len(receiver.Elements)
				return runtime.IntegerValue{
					Val:        big.NewInt(int64(length)),
					TypeSuffix: runtime.IntegerU64,
				}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: arr, Method: fn}, nil
	case "get":
		fn := runtime.NativeFunctionValue{
			Name:  "array.get",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("get expects a receiver and an index")
				}
				receiver, ok := args[0].(*runtime.ArrayValue)
				if !ok {
					return nil, fmt.Errorf("get receiver must be an array")
				}
				idx, err := arrayIndexFromValue(args[1])
				if err != nil {
					return nil, err
				}
				if idx < 0 || idx >= len(receiver.Elements) {
					return runtime.NilValue{}, nil
				}
				return receiver.Elements[idx], nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: arr, Method: fn}, nil
	case "set":
		fn := runtime.NativeFunctionValue{
			Name:  "array.set",
			Arity: 2,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 3 {
					return nil, fmt.Errorf("set expects a receiver, index, and value")
				}
				receiver, ok := args[0].(*runtime.ArrayValue)
				if !ok {
					return nil, fmt.Errorf("set receiver must be an array")
				}
				idx, err := arrayIndexFromValue(args[1])
				if err != nil {
					return nil, err
				}
				if idx < 0 || idx >= len(receiver.Elements) {
					return makeIndexError(idx, len(receiver.Elements)), nil
				}
				receiver.Elements[idx] = args[2]
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: arr, Method: fn}, nil
	case "clear":
		fn := runtime.NativeFunctionValue{
			Name:  "array.clear",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("clear expects only a receiver")
				}
				receiver, ok := args[0].(*runtime.ArrayValue)
				if !ok {
					return nil, fmt.Errorf("clear receiver must be an array")
				}
				receiver.Elements = receiver.Elements[:0]
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: arr, Method: fn}, nil
	case "iterator":
		fn := runtime.NativeFunctionValue{
			Name:  "array.iterator",
			Arity: 0,
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
					if index >= len(receiver.Elements) {
						return runtime.IteratorEnd, true, nil
					}
					val := receiver.Elements[index]
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
		return nil, fmt.Errorf("unknown array method '%s'", ident.Name)
	}
}

func arrayIndexFromValue(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		if v.Val == nil {
			return 0, fmt.Errorf("array index must be an integer")
		}
		if v.Val.Sign() < 0 {
			return 0, fmt.Errorf("array index must be non-negative")
		}
		if !v.Val.IsInt64() {
			return 0, fmt.Errorf("array index out of range")
		}
		res := v.Val.Int64()
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
		"index": runtime.IntegerValue{
			Val:        big.NewInt(int64(index)),
			TypeSuffix: runtime.IntegerI64,
		},
		"length": runtime.IntegerValue{
			Val:        big.NewInt(int64(length)),
			TypeSuffix: runtime.IntegerI64,
		},
	}
	message := fmt.Sprintf("index %d out of bounds for length %d", index, length)
	return runtime.ErrorValue{
		TypeName: ast.NewIdentifier("IndexError"),
		Payload:  payload,
		Message:  message,
	}
}
