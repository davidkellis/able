package interpreter

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"unicode/utf8"

	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) ensureHashMapBuiltins() {
	if i.hashMapReady {
		return
	}
	i.initHashMapBuiltins()
}

func (i *Interpreter) initHashMapBuiltins() {
	if i.hashMapReady {
		return
	}

	if i.hashMapStates == nil {
		i.hashMapStates = make(map[int64]*runtime.HashMapValue)
	}
	if i.nextHashMapHandle == 0 {
		i.nextHashMapHandle = 1
	}

	parseHandle := func(val runtime.Value) (int64, error) {
		intVal, ok := val.(runtime.IntegerValue)
		if !ok || intVal.Val == nil {
			return 0, fmt.Errorf("hash map handle must be an integer")
		}
		if !intVal.Val.IsInt64() {
			return 0, fmt.Errorf("hash map handle is out of range")
		}
		return intVal.Val.Int64(), nil
	}

	hashMapNew := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_new",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("__able_hash_map_new expects no arguments")
			}
			handle := i.newHashMapHandle(0)
			return runtime.IntegerValue{Val: big.NewInt(handle), TypeSuffix: runtime.IntegerI64}, nil
		},
	}

	hashMapWithCapacity := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_with_capacity",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_hash_map_with_capacity expects capacity argument")
			}
			capacity, err := arrayIndexFromValue(args[0])
			if err != nil {
				return nil, fmt.Errorf("capacity must be a non-negative integer")
			}
			handle := i.newHashMapHandle(capacity)
			return runtime.IntegerValue{Val: big.NewInt(handle), TypeSuffix: runtime.IntegerI64}, nil
		},
	}

	hashMapGet := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_get",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_hash_map_get expects handle and key")
			}
			handle, err := parseHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.hashMapStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			hash, err := i.hashMapHashValue(args[1])
			if err != nil {
				return nil, err
			}
			idx, found, err := i.hashMapFindEntryWithHash(state, hash, args[1])
			if err != nil {
				return nil, err
			}
			if found {
				return state.Entries[idx].Value, nil
			}
			return runtime.NilValue{}, nil
		},
	}

	hashMapSet := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_set",
		Arity: 3,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("__able_hash_map_set expects handle, key, and value")
			}
			handle, err := parseHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.hashMapStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			if err := i.hashMapInsertEntry(state, args[1], args[2]); err != nil {
				return nil, err
			}
			return runtime.NilValue{}, nil
		},
	}

	hashMapRemove := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_remove",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_hash_map_remove expects handle and key")
			}
			handle, err := parseHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.hashMapStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			hash, err := i.hashMapHashValue(args[1])
			if err != nil {
				return nil, err
			}
			idx, found, err := i.hashMapFindEntryWithHash(state, hash, args[1])
			if err != nil {
				return nil, err
			}
			if found {
				val := state.Entries[idx].Value
				state.Entries = append(state.Entries[:idx], state.Entries[idx+1:]...)
				return val, nil
			}
			return runtime.NilValue{}, nil
		},
	}

	hashMapContains := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_contains",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_hash_map_contains expects handle and key")
			}
			handle, err := parseHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.hashMapStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			hash, err := i.hashMapHashValue(args[1])
			if err != nil {
				return nil, err
			}
			_, found, err := i.hashMapFindEntryWithHash(state, hash, args[1])
			if err != nil {
				return nil, err
			}
			return runtime.BoolValue{Val: found}, nil
		},
	}

	hashMapSize := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_size",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_hash_map_size expects handle")
			}
			handle, err := parseHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.hashMapStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			return runtime.IntegerValue{Val: big.NewInt(int64(len(state.Entries))), TypeSuffix: runtime.IntegerI32}, nil
		},
	}

	hashMapClear := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_clear",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_hash_map_clear expects handle")
			}
			handle, err := parseHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.hashMapStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			state.Entries = state.Entries[:0]
			return runtime.NilValue{}, nil
		},
	}

	hashMapForEach := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_for_each",
		Arity: 2,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__able_hash_map_for_each expects handle and callback")
			}
			handle, err := parseHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.hashMapStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			callback := args[1]
			for _, entry := range state.Entries {
				if _, err := i.callCallableValue(callback, []runtime.Value{entry.Key, entry.Value}, ctx.Env, nil); err != nil {
					return nil, err
				}
			}
			return runtime.NilValue{}, nil
		},
	}

	hashMapClone := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_clone",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__able_hash_map_clone expects handle")
			}
			handle, err := parseHandle(args[0])
			if err != nil {
				return nil, err
			}
			state, err := i.hashMapStateForHandle(handle)
			if err != nil {
				return nil, err
			}
			cloned := make([]runtime.HashMapEntry, len(state.Entries))
			copy(cloned, state.Entries)
			newHandle := i.newHashMapHandle(len(cloned))
			i.hashMapStates[newHandle].Entries = cloned
			return runtime.IntegerValue{Val: big.NewInt(newHandle), TypeSuffix: runtime.IntegerI64}, nil
		},
	}

	i.global.Define("__able_hash_map_new", hashMapNew)
	i.global.Define("__able_hash_map_with_capacity", hashMapWithCapacity)
	i.global.Define("__able_hash_map_get", hashMapGet)
	i.global.Define("__able_hash_map_set", hashMapSet)
	i.global.Define("__able_hash_map_remove", hashMapRemove)
	i.global.Define("__able_hash_map_contains", hashMapContains)
	i.global.Define("__able_hash_map_size", hashMapSize)
	i.global.Define("__able_hash_map_clear", hashMapClear)
	i.global.Define("__able_hash_map_for_each", hashMapForEach)
	i.global.Define("__able_hash_map_clone", hashMapClone)
	i.hashMapReady = true
}

func (i *Interpreter) newHashMapHandle(capacity int) int64 {
	if i.hashMapStates == nil {
		i.hashMapStates = make(map[int64]*runtime.HashMapValue)
	}
	if i.nextHashMapHandle == 0 {
		i.nextHashMapHandle = 1
	}
	if capacity < 0 {
		capacity = 0
	}
	handle := i.nextHashMapHandle
	i.nextHashMapHandle++
	i.hashMapStates[handle] = &runtime.HashMapValue{Entries: make([]runtime.HashMapEntry, 0, capacity)}
	return handle
}

func (i *Interpreter) hashMapStateForHandle(handle int64) (*runtime.HashMapValue, error) {
	if i.hashMapStates == nil {
		return nil, fmt.Errorf("hash map state is not initialized")
	}
	state, ok := i.hashMapStates[handle]
	if !ok || state == nil {
		return nil, fmt.Errorf("hash map handle %d is not defined", handle)
	}
	return state, nil
}

func (i *Interpreter) hashMapFindEntryWithHash(hm *runtime.HashMapValue, hash uint64, key runtime.Value) (int, bool, error) {
	for idx, entry := range hm.Entries {
		if entry.Hash != hash {
			continue
		}
		equal, err := i.hashMapKeysEqual(entry.Key, key)
		if err != nil {
			return -1, false, err
		}
		if equal {
			return idx, true, nil
		}
	}
	return -1, false, nil
}

func (i *Interpreter) hashMapInsertEntry(hm *runtime.HashMapValue, key runtime.Value, value runtime.Value) error {
	hash, err := i.hashMapHashValue(key)
	if err != nil {
		return err
	}
	idx, found, err := i.hashMapFindEntryWithHash(hm, hash, key)
	if err != nil {
		return err
	}
	if found {
		hm.Entries[idx].Hash = hash
		hm.Entries[idx].Key = key
		hm.Entries[idx].Value = value
		return nil
	}
	hm.Entries = append(hm.Entries, runtime.HashMapEntry{Key: key, Value: value, Hash: hash})
	return nil
}

func (i *Interpreter) hashMapKeysEqual(a, b runtime.Value) (bool, error) {
	if a == b {
		return true, nil
	}
	if valuesEqual(a, b) {
		return true, nil
	}
	switch av := a.(type) {
	case *runtime.ArrayValue:
		other, ok := b.(*runtime.ArrayValue)
		if !ok {
			return false, nil
		}
		if len(av.Elements) != len(other.Elements) {
			return false, nil
		}
		for idx := range av.Elements {
			eq, err := i.hashMapKeysEqual(av.Elements[idx], other.Elements[idx])
			if err != nil {
				return false, err
			}
			if !eq {
				return false, nil
			}
		}
		return true, nil
	case *runtime.StructInstanceValue, *runtime.InterfaceValue, runtime.InterfaceValue:
		res, handled, err := i.tryInvokeEq(a, b)
		if err != nil {
			return false, err
		}
		if handled {
			return res, nil
		}
		res, handled, err = i.tryInvokeEq(b, a)
		if err != nil {
			return false, err
		}
		if handled {
			return res, nil
		}
		typeDesc := fmt.Sprintf("%T", a)
		if info, ok := i.getTypeInfoForValue(a); ok {
			if s := typeInfoToString(info); s != "" && s != "<unknown>" {
				typeDesc = s
			}
		}
		return false, fmt.Errorf("hash map key type %s does not implement eq()", typeDesc)
	default:
		return false, nil
	}
}

func (i *Interpreter) hashMapHashValue(val runtime.Value) (uint64, error) {
	switch v := val.(type) {
	case runtime.StringValue:
		return runtime.HashWithTag('s', []byte(v.Val)), nil
	case runtime.BoolValue:
		b := byte(0)
		if v.Val {
			b = 1
		}
		return runtime.HashWithTag('b', []byte{b}), nil
	case runtime.CharValue:
		buf := make([]byte, 4)
		n := utf8.EncodeRune(buf, v.Val)
		return runtime.HashWithTag('c', buf[:n]), nil
	case runtime.NilValue:
		return runtime.HashWithTag('n', []byte{0}), nil
	case runtime.IntegerValue:
		if v.Val == nil {
			return 0, fmt.Errorf("hash map key: integer missing value")
		}
		bytes := v.Val.Bytes()
		if len(bytes) == 0 {
			bytes = []byte{0}
		}
		sign := byte(0)
		if v.Val.Sign() < 0 {
			sign = 1
		}
		data := append([]byte{sign}, bytes...)
		return runtime.HashWithTag('i', data), nil
	case runtime.FloatValue:
		if math.IsNaN(v.Val) {
			return 0, fmt.Errorf("hash map key: NaN is not hashable")
		}
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(v.Val))
		return runtime.HashWithTag('f', buf), nil
	case *runtime.ArrayValue:
		hash := runtime.HashWithTag('a', nil)
		for _, elem := range v.Elements {
			elemHash, err := i.hashMapHashValue(elem)
			if err != nil {
				return 0, err
			}
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, elemHash)
			hash = runtime.HashBytes(hash, buf)
		}
		return hash, nil
	default:
		return i.hashCustomHashValue(val)
	}
}

func (i *Interpreter) hashCustomHashValue(val runtime.Value) (uint64, error) {
	switch v := val.(type) {
	case *runtime.StructInstanceValue:
		return i.invokeHashMethod(v)
	case *runtime.InterfaceValue:
		if v == nil || v.Underlying == nil {
			return 0, fmt.Errorf("hash map key: interface value has no underlying instance")
		}
		return i.hashCustomHashValue(v.Underlying)
	case runtime.InterfaceValue:
		return i.hashCustomHashValue(&v)
	default:
		return 0, fmt.Errorf("hash map key type %T is not supported", val)
	}
}

func (i *Interpreter) invokeHashMethod(receiver runtime.Value) (uint64, error) {
	if receiver == nil {
		return 0, fmt.Errorf("hash map key: receiver is nil")
	}
	info, ok := i.getTypeInfoForValue(receiver)
	if !ok {
		return 0, fmt.Errorf("hash map key type %T does not support hashing", receiver)
	}

	var method runtime.Value
	if bucket, ok := i.inherentMethods[info.name]; ok {
		if fn, exists := bucket["hash"]; exists {
			method = fn
		}
	}
	if method == nil {
		var err error
		method, err = i.findMethod(info, "hash", "")
		if err != nil {
			return 0, err
		}
	}
	if method == nil {
		typeDesc := typeInfoToString(info)
		if typeDesc == "<unknown>" {
			typeDesc = info.name
		}
		return 0, fmt.Errorf("hash map key type %s does not implement hash()", typeDesc)
	}

	hasher := runtime.NewHasherValue()
	bound := runtime.BoundMethodValue{Receiver: receiver, Method: method}
	result, err := i.callCallableValue(bound, []runtime.Value{hasher}, nil, nil)
	if err != nil {
		return 0, err
	}

	switch rv := result.(type) {
	case runtime.IntegerValue:
		if rv.Val == nil {
			return 0, fmt.Errorf("hash() returned nil integer")
		}
		if rv.Val.Sign() < 0 {
			return 0, fmt.Errorf("hash() must return a non-negative integer")
		}
		if !rv.Val.IsUint64() {
			return 0, fmt.Errorf("hash() result exceeds u64 range")
		}
		return rv.Val.Uint64(), nil
	case runtime.NilValue:
		return hasher.Finish(), nil
	default:
		return 0, fmt.Errorf("hash() must return u64 or nil (got %s)", rv.Kind())
	}
}

func (i *Interpreter) tryInvokeEq(receiver runtime.Value, other runtime.Value) (bool, bool, error) {
	if receiver == nil {
		return false, false, nil
	}
	info, ok := i.getTypeInfoForValue(receiver)
	if !ok {
		return false, false, nil
	}

	var method runtime.Value
	if bucket, ok := i.inherentMethods[info.name]; ok {
		if fn, exists := bucket["eq"]; exists {
			method = fn
		}
	}
	if method == nil {
		var err error
		method, err = i.findMethod(info, "eq", "")
		if err != nil {
			return false, false, err
		}
	}
	if method == nil {
		return false, false, nil
	}

	bound := runtime.BoundMethodValue{Receiver: receiver, Method: method}
	result, err := i.callCallableValue(bound, []runtime.Value{other}, nil, nil)
	if err != nil {
		return false, false, err
	}

	boolResult, ok := result.(runtime.BoolValue)
	if !ok {
		return false, false, fmt.Errorf("eq() must return bool (got %s)", result.Kind())
	}
	return boolResult.Val, true, nil
}
