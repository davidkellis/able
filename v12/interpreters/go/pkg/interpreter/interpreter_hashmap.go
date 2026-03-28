package interpreter

import (
	"fmt"
	"strings"

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

	parseHandle := func(val runtime.Value) (int64, error) {
		intVal, ok := val.(runtime.IntegerValue)
		if !ok {
			return 0, fmt.Errorf("hash map handle must be an integer")
		}
		n, ok := intVal.ToInt64()
		if !ok {
			return 0, fmt.Errorf("hash map handle is out of range")
		}
		return n, nil
	}

	hashMapNew := runtime.NativeFunctionValue{
		Name:  "__able_hash_map_new",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("__able_hash_map_new expects no arguments")
			}
			handle := i.newHashMapHandle(0)
			return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
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
			return runtime.NewSmallInt(handle, runtime.IntegerI64), nil
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
			hash, err := i.HashMapHashValue(args[1])
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
			hash, err := i.HashMapHashValue(args[1])
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
			hash, err := i.HashMapHashValue(args[1])
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
			return runtime.NewSmallInt(int64(len(state.Entries)), runtime.IntegerI32), nil
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
			newHandle, err := runtime.HashMapStoreClone(handle)
			if err != nil {
				return nil, err
			}
			return runtime.NewSmallInt(newHandle, runtime.IntegerI64), nil
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
	return runtime.HashMapStoreNewWithCapacity(capacity)
}

func (i *Interpreter) hashMapStateForHandle(handle int64) (*runtime.HashMapValue, error) {
	return runtime.HashMapStoreState(handle)
}

func (i *Interpreter) hashMapFindEntryWithHash(hm *runtime.HashMapValue, hash uint64, key runtime.Value) (int, bool, error) {
	for idx, entry := range hm.Entries {
		if entry.Hash != hash {
			continue
		}
		equal, err := i.HashMapKeysEqual(entry.Key, key)
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
	hash, err := i.HashMapHashValue(key)
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

// HashMapKeysEqual compares hash map keys using Eq.eq.
func (i *Interpreter) HashMapKeysEqual(a, b runtime.Value) (bool, error) {
	receiver := unwrapInterfaceValue(a)
	other := unwrapInterfaceValue(b)
	method, err := i.resolveInterfaceMethod(receiver, "Eq", "eq")
	if err != nil {
		return false, err
	}
	if method == nil {
		return false, raiseSignal{value: runtime.ErrorValue{Message: fmt.Sprintf("HashMap key type %s does not implement Eq.eq", i.typeDescForValue(receiver))}}
	}
	result, err := i.callCallableValue(method, []runtime.Value{receiver, other}, nil, nil)
	if err != nil {
		return false, err
	}
	boolResult, ok := result.(runtime.BoolValue)
	if !ok {
		if ptr, okPtr := result.(*runtime.BoolValue); okPtr && ptr != nil {
			boolResult = *ptr
			ok = true
		}
	}
	if !ok {
		return false, raiseSignal{value: runtime.ErrorValue{Message: fmt.Sprintf("Eq.eq must return bool (got %s)", result.Kind())}}
	}
	return boolResult.Val, nil
}

// HashMapHashValue computes the hash for a map key using Hash.hash.
func (i *Interpreter) HashMapHashValue(val runtime.Value) (uint64, error) {
	receiver := unwrapInterfaceValue(val)
	method, err := i.resolveInterfaceMethod(receiver, "Hash", "hash")
	if err != nil {
		return 0, err
	}
	if method == nil {
		return 0, raiseSignal{value: runtime.ErrorValue{Message: fmt.Sprintf("HashMap key type %s does not implement Hash.hash", i.typeDescForValue(receiver))}}
	}
	hasher, err := i.newKernelHasher()
	if err != nil {
		return 0, err
	}
	result, err := i.callCallableValue(method, []runtime.Value{receiver, hasher}, nil, nil)
	if err != nil {
		return 0, err
	}
	if !isVoidOrNil(result) {
		return 0, raiseSignal{value: runtime.ErrorValue{Message: "Hash.hash must return void"}}
	}
	return i.finishKernelHasher(hasher)
}

func unwrapInterfaceValue(val runtime.Value) runtime.Value {
	for {
		switch v := val.(type) {
		case *runtime.InterfaceValue:
			if v == nil || v.Underlying == nil {
				return val
			}
			val = v.Underlying
		case runtime.InterfaceValue:
			if v.Underlying == nil {
				return val
			}
			val = v.Underlying
		default:
			return val
		}
	}
}

func (i *Interpreter) resolveInterfaceMethod(receiver runtime.Value, interfaceName string, methodName string) (runtime.Value, error) {
	info, ok := i.getTypeInfoForValue(receiver)
	if !ok {
		if i.interfaceMethodResolver != nil {
			if method, found := i.interfaceMethodResolver(receiver, interfaceName, methodName); found {
				return method, nil
			}
		}
		return nil, nil
	}
	method, err := i.findMethodCached(info, methodName, interfaceName)
	if err != nil {
		return nil, err
	}
	if method != nil {
		return method, nil
	}
	if i.interfaceMethodResolver != nil {
		if resolved, found := i.interfaceMethodResolver(receiver, interfaceName, methodName); found {
			return resolved, nil
		}
	}
	return nil, nil
}

func (i *Interpreter) newKernelHasher() (runtime.Value, error) {
	candidates := []string{"KernelHasher.new", "able.kernel.KernelHasher.new"}
	for _, name := range candidates {
		val, err := i.global.Get(name)
		if err != nil {
			if resolved, ok := i.lookupPackageRegistrySymbol("able.kernel", name); ok {
				val = resolved
				err = nil
			}
		}
		if err != nil {
			continue
		}
		result, err := i.callCallableValue(val, nil, i.global, nil)
		if err != nil {
			return nil, err
		}
		switch inst := result.(type) {
		case *runtime.StructInstanceValue:
			if inst != nil && structInstanceName(inst) == "KernelHasher" {
				return inst, nil
			}
		}
		return nil, fmt.Errorf("KernelHasher.new returned unexpected value")
	}
	return nil, fmt.Errorf("KernelHasher.new is not available")
}

func (i *Interpreter) lookupPackageRegistrySymbol(pkgName string, symbol string) (runtime.Value, bool) {
	if i == nil || pkgName == "" || symbol == "" {
		return nil, false
	}
	bucket, ok := i.packageRegistry[pkgName]
	if !ok || bucket == nil {
		return nil, false
	}
	if val, ok := bucket[symbol]; ok {
		return val, true
	}
	if idx := strings.LastIndex(symbol, "."); idx >= 0 && idx+1 < len(symbol) {
		if val, ok := bucket[symbol[idx+1:]]; ok {
			return val, true
		}
	}
	return nil, false
}

func (i *Interpreter) finishKernelHasher(hasher runtime.Value) (uint64, error) {
	method, err := i.resolveInterfaceMethod(hasher, "Hasher", "finish")
	if err != nil {
		return 0, err
	}
	if method == nil {
		return 0, fmt.Errorf("Hasher.finish is not available for KernelHasher")
	}
	result, err := i.callCallableValue(method, []runtime.Value{hasher}, nil, nil)
	if err != nil {
		return 0, err
	}
	return integerToUint64(result)
}

func (i *Interpreter) typeDescForValue(val runtime.Value) string {
	if info, ok := i.getTypeInfoForValue(val); ok {
		if desc := typeInfoToString(info); desc != "" && desc != "<unknown>" {
			return desc
		}
		if info.name != "" {
			return info.name
		}
	}
	return fmt.Sprintf("%T", val)
}

func isVoidOrNil(val runtime.Value) bool {
	switch v := val.(type) {
	case runtime.VoidValue, runtime.NilValue:
		return true
	case *runtime.VoidValue, *runtime.NilValue:
		return v != nil
	default:
		return false
	}
}
