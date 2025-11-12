package interpreter

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) ensureArrayBuiltins() {
	if i.arrayReady {
		return
	}
	i.initArrayBuiltins()
}

func (i *Interpreter) initArrayBuiltins() {
	if i.arrayReady {
		return
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
