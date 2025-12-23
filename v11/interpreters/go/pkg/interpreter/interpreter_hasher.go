package interpreter

import (
	"fmt"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) hasherMember(hasher *runtime.HasherValue, member ast.Expression) (runtime.Value, error) {
	if hasher == nil {
		return nil, fmt.Errorf("hasher receiver is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("hasher member access expects identifier")
	}

	switch ident.Name {
	case "finish":
		fn := runtime.NativeFunctionValue{
			Name:  "hasher.finish",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("finish expects only a receiver")
				}
				ptr, ok := args[0].(*runtime.HasherValue)
				if !ok {
					return nil, fmt.Errorf("finish receiver must be a hasher")
				}
				value := ptr.Finish()
				return runtime.IntegerValue{
					Val:        new(big.Int).SetUint64(value),
					TypeSuffix: runtime.IntegerU64,
				}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hasher, Method: fn}, nil
	case "write_bytes":
		fn := runtime.NativeFunctionValue{
			Name:  "hasher.write_bytes",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("write_bytes expects a receiver and a string")
				}
				ptr, ok := args[0].(*runtime.HasherValue)
				if !ok {
					return nil, fmt.Errorf("write_bytes receiver must be a hasher")
				}
				str, ok := args[1].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("write_bytes expects a string argument")
				}
				ptr.WriteBytes([]byte(str.Val))
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hasher, Method: fn}, nil
	case "write_string":
		fn := runtime.NativeFunctionValue{
			Name:  "hasher.write_string",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("write_string expects a receiver and a string")
				}
				ptr, ok := args[0].(*runtime.HasherValue)
				if !ok {
					return nil, fmt.Errorf("write_string receiver must be a hasher")
				}
				str, ok := args[1].(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("write_string expects a string argument")
				}
				ptr.WriteString(str.Val)
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hasher, Method: fn}, nil
	case "write_bool":
		fn := runtime.NativeFunctionValue{
			Name:  "hasher.write_bool",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("write_bool expects a receiver and a bool")
				}
				ptr, ok := args[0].(*runtime.HasherValue)
				if !ok {
					return nil, fmt.Errorf("write_bool receiver must be a hasher")
				}
				val, ok := args[1].(runtime.BoolValue)
				if !ok {
					return nil, fmt.Errorf("write_bool expects a bool argument")
				}
				ptr.WriteBool(val.Val)
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hasher, Method: fn}, nil
	case "write_u64":
		fn := runtime.NativeFunctionValue{
			Name:  "hasher.write_u64",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("write_u64 expects a receiver and an integer")
				}
				ptr, ok := args[0].(*runtime.HasherValue)
				if !ok {
					return nil, fmt.Errorf("write_u64 receiver must be a hasher")
				}
				intVal, err := integerToUint64(args[1])
				if err != nil {
					return nil, err
				}
				ptr.WriteUint64(intVal)
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hasher, Method: fn}, nil
	case "write_i64":
		fn := runtime.NativeFunctionValue{
			Name:  "hasher.write_i64",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("write_i64 expects a receiver and an integer")
				}
				ptr, ok := args[0].(*runtime.HasherValue)
				if !ok {
					return nil, fmt.Errorf("write_i64 receiver must be a hasher")
				}
				intVal, err := integerToInt64(args[1])
				if err != nil {
					return nil, err
				}
				ptr.WriteInt64(intVal)
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hasher, Method: fn}, nil
	default:
		return nil, fmt.Errorf("unknown hasher method '%s'", ident.Name)
	}
}

func integerToUint64(val runtime.Value) (uint64, error) {
	iv, ok := val.(runtime.IntegerValue)
	if !ok {
		return 0, fmt.Errorf("expected unsigned integer value")
	}
	if iv.Val == nil || iv.Val.Sign() < 0 {
		return 0, fmt.Errorf("expected non-negative integer")
	}
	if !iv.Val.IsUint64() {
		return 0, fmt.Errorf("integer out of range for u64")
	}
	return iv.Val.Uint64(), nil
}

func integerToInt64(val runtime.Value) (int64, error) {
	iv, ok := val.(runtime.IntegerValue)
	if !ok {
		return 0, fmt.Errorf("expected integer value")
	}
	if iv.Val == nil || !iv.Val.IsInt64() {
		return 0, fmt.Errorf("integer out of range for i64")
	}
	return iv.Val.Int64(), nil
}
