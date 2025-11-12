package interpreter

import (
	"fmt"
	"math/big"

	"able/interpreter10-go/pkg/runtime"
)

type hasherState struct {
	hash uint32
}

const (
	fnvOffset32 = uint32(2166136261)
	fnvPrime32  = uint32(16777619)
)

func (i *Interpreter) ensureHasherBuiltins() {
	if i.hasherReady {
		return
	}
	i.initHasherBuiltins()
}

func (i *Interpreter) initHasherBuiltins() {
	if i.hasherReady {
		return
	}
	if i.hashers == nil {
		i.hashers = make(map[int64]*hasherState)
	}

	makeHandleValue := func(handle int64) runtime.IntegerValue {
		return runtime.IntegerValue{
			Val:        big.NewInt(handle),
			TypeSuffix: runtime.IntegerI64,
		}
	}

	int64FromValue := func(val runtime.Value, label string) (int64, error) {
		switch v := val.(type) {
		case runtime.IntegerValue:
			if !v.Val.IsInt64() {
				return 0, fmt.Errorf("%s must fit in 64-bit integer", label)
			}
			return v.Val.Int64(), nil
		case *runtime.IntegerValue:
			if v == nil || v.Val == nil {
				return 0, fmt.Errorf("%s is nil", label)
			}
			if !v.Val.IsInt64() {
				return 0, fmt.Errorf("%s must fit in 64-bit integer", label)
			}
			return v.Val.Int64(), nil
		default:
			return 0, fmt.Errorf("%s must be an integer", label)
		}
	}

	stringFromValue := func(val runtime.Value, label string) (string, error) {
		switch v := val.(type) {
		case runtime.StringValue:
			return v.Val, nil
		case *runtime.StringValue:
			if v == nil {
				return "", fmt.Errorf("%s is nil", label)
			}
			return v.Val, nil
		default:
			return "", fmt.Errorf("%s must be a string", label)
		}
	}

	hasherCreate := runtime.NativeFunctionValue{
		Name:  "__able_hasher_create",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			i.hasherMu.Lock()
			defer i.hasherMu.Unlock()
			if i.nextHasherHandle == 0 {
				i.nextHasherHandle = 1
			}
			handle := i.nextHasherHandle
			i.nextHasherHandle++
			i.hashers[handle] = &hasherState{hash: fnvOffset32}
			return makeHandleValue(handle), nil
		},
	}

	hasherWrite := runtime.NativeFunctionValue{
		Name:  "__able_hasher_write",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_hasher_write expects handle and bytes string")
			}
			handle, err := int64FromValue(args[0], "hasher handle")
			if err != nil {
				return nil, err
			}
			if handle <= 0 {
				return nil, fmt.Errorf("hasher handle must be positive")
			}
			chunk, err := stringFromValue(args[1], "bytes")
			if err != nil {
				return nil, err
			}
			i.hasherMu.Lock()
			state, ok := i.hashers[handle]
			if !ok {
				i.hasherMu.Unlock()
				return nil, fmt.Errorf("unknown hasher handle %d", handle)
			}
			for _, b := range []byte(chunk) {
				state.hash ^= uint32(b)
				state.hash *= fnvPrime32
			}
			i.hasherMu.Unlock()
			return runtime.NilValue{}, nil
		},
	}

	hasherFinish := runtime.NativeFunctionValue{
		Name:  "__able_hasher_finish",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_hasher_finish expects handle argument")
			}
			handle, err := int64FromValue(args[0], "hasher handle")
			if err != nil {
				return nil, err
			}
			if handle <= 0 {
				return nil, fmt.Errorf("hasher handle must be positive")
			}
			i.hasherMu.Lock()
			state, ok := i.hashers[handle]
			if !ok {
				i.hasherMu.Unlock()
				return nil, fmt.Errorf("unknown hasher handle %d", handle)
			}
			delete(i.hashers, handle)
			hash := state.hash
			i.hasherMu.Unlock()
			return runtime.IntegerValue{
				Val:        big.NewInt(int64(hash)),
				TypeSuffix: runtime.IntegerI64,
			}, nil
		},
	}

	i.global.Define("__able_hasher_create", hasherCreate)
	i.global.Define("__able_hasher_write", hasherWrite)
	i.global.Define("__able_hasher_finish", hasherFinish)
	i.hasherReady = true
}
