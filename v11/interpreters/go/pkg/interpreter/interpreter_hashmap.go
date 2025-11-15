package interpreter

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"unicode/utf8"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
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

	hashMapPkg := &runtime.PackageValue{
		Name:   "HashMap",
		Public: make(map[string]runtime.Value),
	}

	newFn := runtime.NativeFunctionValue{
		Name:  "HashMap.new",
		Arity: 0,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			capacity := 0
			if len(args) > 1 {
				return nil, fmt.Errorf("HashMap.new expects zero or one argument")
			}
			if len(args) == 1 {
				val, err := arrayIndexFromValue(args[0])
				if err != nil {
					return nil, fmt.Errorf("HashMap.new capacity must be a non-negative integer")
				}
				if val < 0 {
					return nil, fmt.Errorf("HashMap.new capacity must be non-negative")
				}
				capacity = val
			}
			if capacity < 0 {
				capacity = 0
			}
			return &runtime.HashMapValue{Entries: make([]runtime.HashMapEntry, 0, capacity)}, nil
		},
	}

	hashMapPkg.Public["new"] = newFn
	i.global.Define("HashMap", hashMapPkg)
	i.hashMapReady = true
}

func (i *Interpreter) hashMapMember(hm *runtime.HashMapValue, member ast.Expression) (runtime.Value, error) {
	if hm == nil {
		return nil, fmt.Errorf("hash map receiver is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("hash map member access expects identifier")
	}
	switch ident.Name {
	case "set":
		fn := runtime.NativeFunctionValue{
			Name:  "hash_map.set",
			Arity: 2,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 3 {
					return nil, fmt.Errorf("set expects a receiver, key, and value")
				}
				receiver, ok := args[0].(*runtime.HashMapValue)
				if !ok {
					return nil, fmt.Errorf("set receiver must be a hash map")
				}
				hash, err := i.hashMapHashValue(args[1])
				if err != nil {
					return nil, err
				}
				idx, found, err := i.hashMapFindEntryWithHash(receiver, hash, args[1])
				if err != nil {
					return nil, err
				}
				if found {
					receiver.Entries[idx].Hash = hash
					receiver.Entries[idx].Value = args[2]
				} else {
					receiver.Entries = append(receiver.Entries, runtime.HashMapEntry{Key: args[1], Value: args[2], Hash: hash})
				}
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hm, Method: fn}, nil
	case "get":
		fn := runtime.NativeFunctionValue{
			Name:  "hash_map.get",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("get expects a receiver and a key")
				}
				receiver, ok := args[0].(*runtime.HashMapValue)
				if !ok {
					return nil, fmt.Errorf("get receiver must be a hash map")
				}
				hash, err := i.hashMapHashValue(args[1])
				if err != nil {
					return nil, err
				}
				idx, found, err := i.hashMapFindEntryWithHash(receiver, hash, args[1])
				if err != nil {
					return nil, err
				}
				if found {
					return receiver.Entries[idx].Value, nil
				}
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hm, Method: fn}, nil
	case "remove":
		fn := runtime.NativeFunctionValue{
			Name:  "hash_map.remove",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("remove expects a receiver and a key")
				}
				receiver, ok := args[0].(*runtime.HashMapValue)
				if !ok {
					return nil, fmt.Errorf("remove receiver must be a hash map")
				}
				hash, err := i.hashMapHashValue(args[1])
				if err != nil {
					return nil, err
				}
				idx, found, err := i.hashMapFindEntryWithHash(receiver, hash, args[1])
				if err != nil {
					return nil, err
				}
				if found {
					val := receiver.Entries[idx].Value
					receiver.Entries = append(receiver.Entries[:idx], receiver.Entries[idx+1:]...)
					return val, nil
				}
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hm, Method: fn}, nil
	case "contains":
		fn := runtime.NativeFunctionValue{
			Name:  "hash_map.contains",
			Arity: 1,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("contains expects a receiver and a key")
				}
				receiver, ok := args[0].(*runtime.HashMapValue)
				if !ok {
					return nil, fmt.Errorf("contains receiver must be a hash map")
				}
				hash, err := i.hashMapHashValue(args[1])
				if err != nil {
					return nil, err
				}
				_, found, err := i.hashMapFindEntryWithHash(receiver, hash, args[1])
				if err != nil {
					return nil, err
				}
				return runtime.BoolValue{Val: found}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hm, Method: fn}, nil
	case "size":
		fn := runtime.NativeFunctionValue{
			Name:  "hash_map.size",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("size expects only a receiver")
				}
				receiver, ok := args[0].(*runtime.HashMapValue)
				if !ok {
					return nil, fmt.Errorf("size receiver must be a hash map")
				}
				return runtime.IntegerValue{
					Val:        big.NewInt(int64(len(receiver.Entries))),
					TypeSuffix: runtime.IntegerU64,
				}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hm, Method: fn}, nil
	case "clear":
		fn := runtime.NativeFunctionValue{
			Name:  "hash_map.clear",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("clear expects only a receiver")
				}
				receiver, ok := args[0].(*runtime.HashMapValue)
				if !ok {
					return nil, fmt.Errorf("clear receiver must be a hash map")
				}
				receiver.Entries = receiver.Entries[:0]
				return runtime.NilValue{}, nil
			},
		}
		return &runtime.NativeBoundMethodValue{Receiver: hm, Method: fn}, nil
	default:
		return nil, fmt.Errorf("unknown hash map method '%s'", ident.Name)
	}
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

	var method *runtime.FunctionValue
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
	result, err := i.invokeFunction(method, []runtime.Value{receiver, hasher}, nil)
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

	var method *runtime.FunctionValue
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

	result, err := i.invokeFunction(method, []runtime.Value{receiver, other}, nil)
	if err != nil {
		return false, false, err
	}

	boolResult, ok := result.(runtime.BoolValue)
	if !ok {
		return false, false, fmt.Errorf("eq() must return bool (got %s)", result.Kind())
	}
	return boolResult.Val, true, nil
}
