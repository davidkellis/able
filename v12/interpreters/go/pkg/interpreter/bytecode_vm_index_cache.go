package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeIndexMethodCacheKey struct {
	program       *bytecodeProgram
	ip            int
	method        string
	receiverKind  bytecodeMemberReceiverKind
	arrayElemType uint16
}

type bytecodeIndexMethodCacheEntry struct {
	globalRevision     uint64
	methodCacheVersion uint64
	method             runtime.Value
	hasMethod          bool
}

func bytecodeArrayReceiverForIndexCache(value runtime.Value) (*runtime.ArrayValue, bool) {
	switch v := value.(type) {
	case *runtime.ArrayValue:
		return v, v != nil
	case runtime.InterfaceValue:
		if arr, ok := v.Underlying.(*runtime.ArrayValue); ok && arr != nil {
			return arr, true
		}
	case *runtime.InterfaceValue:
		if v != nil {
			if arr, ok := v.Underlying.(*runtime.ArrayValue); ok && arr != nil {
				return arr, true
			}
		}
	}
	return nil, false
}

const (
	bytecodeIndexTypeUnknown uint16 = iota
	bytecodeIndexTypeI8
	bytecodeIndexTypeI16
	bytecodeIndexTypeI32
	bytecodeIndexTypeI64
	bytecodeIndexTypeI128
	bytecodeIndexTypeU8
	bytecodeIndexTypeU16
	bytecodeIndexTypeU32
	bytecodeIndexTypeU64
	bytecodeIndexTypeU128
	bytecodeIndexTypeIsize
	bytecodeIndexTypeUsize
	bytecodeIndexTypeF32
	bytecodeIndexTypeF64
	bytecodeIndexTypeString
	bytecodeIndexTypeBool
	bytecodeIndexTypeChar
	bytecodeIndexTypeNil
	bytecodeIndexTypeVoid
)

func bytecodeIntegerTypeToken(suffix runtime.IntegerType) uint16 {
	switch suffix {
	case runtime.IntegerI8:
		return bytecodeIndexTypeI8
	case runtime.IntegerI16:
		return bytecodeIndexTypeI16
	case runtime.IntegerI32:
		return bytecodeIndexTypeI32
	case runtime.IntegerI64:
		return bytecodeIndexTypeI64
	case runtime.IntegerI128:
		return bytecodeIndexTypeI128
	case runtime.IntegerU8:
		return bytecodeIndexTypeU8
	case runtime.IntegerU16:
		return bytecodeIndexTypeU16
	case runtime.IntegerU32:
		return bytecodeIndexTypeU32
	case runtime.IntegerU64:
		return bytecodeIndexTypeU64
	case runtime.IntegerU128:
		return bytecodeIndexTypeU128
	case runtime.IntegerIsize:
		return bytecodeIndexTypeIsize
	case runtime.IntegerUsize:
		return bytecodeIndexTypeUsize
	default:
		return bytecodeIndexTypeUnknown
	}
}

func bytecodeFloatTypeToken(suffix runtime.FloatType) uint16 {
	switch suffix {
	case runtime.FloatF32:
		return bytecodeIndexTypeF32
	case runtime.FloatF64:
		return bytecodeIndexTypeF64
	default:
		return bytecodeIndexTypeUnknown
	}
}

func bytecodeIndexValueTypeToken(value runtime.Value) (uint16, bool) {
	normalized := unwrapInterfaceValue(value)
	switch v := normalized.(type) {
	case runtime.IntegerValue:
		token := bytecodeIntegerTypeToken(v.TypeSuffix)
		return token, token != bytecodeIndexTypeUnknown
	case *runtime.IntegerValue:
		if v == nil {
			return bytecodeIndexTypeUnknown, false
		}
		token := bytecodeIntegerTypeToken(v.TypeSuffix)
		return token, token != bytecodeIndexTypeUnknown
	case runtime.FloatValue:
		token := bytecodeFloatTypeToken(v.TypeSuffix)
		return token, token != bytecodeIndexTypeUnknown
	case *runtime.FloatValue:
		if v == nil {
			return bytecodeIndexTypeUnknown, false
		}
		token := bytecodeFloatTypeToken(v.TypeSuffix)
		return token, token != bytecodeIndexTypeUnknown
	case runtime.StringValue, *runtime.StringValue:
		return bytecodeIndexTypeString, true
	case runtime.BoolValue, *runtime.BoolValue:
		return bytecodeIndexTypeBool, true
	case runtime.CharValue, *runtime.CharValue:
		return bytecodeIndexTypeChar, true
	case runtime.NilValue:
		return bytecodeIndexTypeNil, true
	case runtime.VoidValue:
		return bytecodeIndexTypeVoid, true
	default:
		return bytecodeIndexTypeUnknown, false
	}
}

func bytecodeArrayElementTypeToken(arr *runtime.ArrayValue) (uint16, bool) {
	if arr == nil {
		return bytecodeIndexTypeUnknown, false
	}
	if len(arr.Elements) == 0 {
		return bytecodeIndexTypeUnknown, true
	}
	return bytecodeIndexValueTypeToken(arr.Elements[0])
}

func (vm *bytecodeVM) indexMethodCacheKey(program *bytecodeProgram, ip int, methodName string, receiver runtime.Value) (bytecodeIndexMethodCacheKey, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || methodName == "" {
		return bytecodeIndexMethodCacheKey{}, false
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return bytecodeIndexMethodCacheKey{}, false
	}
	elemType, ok := bytecodeArrayElementTypeToken(arr)
	if !ok {
		return bytecodeIndexMethodCacheKey{}, false
	}
	return bytecodeIndexMethodCacheKey{
		program:       program,
		ip:            ip,
		method:        methodName,
		receiverKind:  bytecodeMemberReceiverArray,
		arrayElemType: elemType,
	}, true
}

func (vm *bytecodeVM) lookupCachedIndexMethod(program *bytecodeProgram, ip int, methodName string, receiver runtime.Value) (runtime.Value, bool, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return nil, false, false
	}
	key, ok := vm.indexMethodCacheKey(program, ip, methodName, receiver)
	if !ok || vm.indexMethodCache == nil {
		return nil, false, false
	}
	entry, ok := vm.indexMethodCache[key]
	if !ok {
		return nil, false, false
	}
	if entry.globalRevision != vm.interp.global.Revision() {
		return nil, false, false
	}
	if entry.methodCacheVersion != vm.interp.currentMethodCacheVersion() {
		return nil, false, false
	}
	return entry.method, true, entry.hasMethod
}

func (vm *bytecodeVM) storeCachedIndexMethod(program *bytecodeProgram, ip int, methodName string, receiver runtime.Value, method runtime.Value, hasMethod bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return
	}
	key, ok := vm.indexMethodCacheKey(program, ip, methodName, receiver)
	if !ok {
		return
	}
	if hasMethod && method == nil {
		return
	}
	if vm.indexMethodCache == nil {
		vm.indexMethodCache = make(map[bytecodeIndexMethodCacheKey]bytecodeIndexMethodCacheEntry, 16)
	}
	vm.indexMethodCache[key] = bytecodeIndexMethodCacheEntry{
		globalRevision:     vm.interp.global.Revision(),
		methodCacheVersion: vm.interp.currentMethodCacheVersion(),
		method:             method,
		hasMethod:          hasMethod,
	}
}

func (vm *bytecodeVM) resolveIndexGet(obj runtime.Value, idxVal runtime.Value) (runtime.Value, error) {
	method, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(vm.currentProgram, vm.ip, obj, "get", "Index")
	if err != nil {
		return nil, err
	}
	if cacheable {
		if hasMethod {
			return vm.interp.CallFunction(method, []runtime.Value{obj, idxVal})
		}
		return vm.interp.indexGetWithoutMethod(obj, idxVal)
	}
	return vm.interp.indexGet(obj, idxVal)
}

func (vm *bytecodeVM) resolveIndexSet(obj runtime.Value, idxVal runtime.Value, value runtime.Value, op ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, error) {
	setMethod, hasSetMethod, cacheable, err := vm.resolveCachedIndexMethod(vm.currentProgram, vm.ip, obj, "set", "IndexMut")
	if err != nil {
		return nil, err
	}
	if !cacheable {
		return vm.interp.assignIndex(obj, idxVal, value, op, binaryOp, isCompound)
	}
	if !hasSetMethod {
		return vm.interp.assignIndexWithoutMethods(obj, idxVal, value, op, binaryOp, isCompound)
	}
	if op == ast.AssignmentDeclare {
		return nil, fmt.Errorf("Cannot use := on index assignment")
	}
	if op == ast.AssignmentAssign {
		setResult, err := vm.interp.CallFunction(setMethod, []runtime.Value{obj, idxVal, value})
		if err != nil {
			return nil, err
		}
		if isErrorResult(vm.interp, setResult) {
			return setResult, nil
		}
		return value, nil
	}
	if !isCompound {
		return nil, fmt.Errorf("unsupported assignment operator %s", op)
	}
	getMethod, hasGetMethod, _, err := vm.resolveCachedIndexMethod(vm.currentProgram, vm.ip, obj, "get", "Index")
	if err != nil {
		return nil, err
	}
	if !hasGetMethod {
		return nil, fmt.Errorf("Compound index assignment requires readable Index implementation")
	}
	current, err := vm.interp.CallFunction(getMethod, []runtime.Value{obj, idxVal})
	if err != nil {
		return nil, err
	}
	computed, err := applyBinaryOperator(vm.interp, binaryOp, current, value)
	if err != nil {
		return nil, err
	}
	setResult, err := vm.interp.CallFunction(setMethod, []runtime.Value{obj, idxVal, computed})
	if err != nil {
		return nil, err
	}
	if isErrorResult(vm.interp, setResult) {
		return setResult, nil
	}
	return computed, nil
}

func (vm *bytecodeVM) resolveCachedIndexMethod(program *bytecodeProgram, ip int, receiver runtime.Value, methodName string, iface string) (runtime.Value, bool, bool, error) {
	if method, cached, hasMethod := vm.lookupCachedIndexMethod(program, ip, methodName, receiver); cached {
		return method, hasMethod, true, nil
	}
	if _, cacheable := vm.indexMethodCacheKey(program, ip, methodName, receiver); !cacheable {
		return nil, false, false, nil
	}
	method, err := vm.interp.findIndexMethod(receiver, methodName, iface)
	if err != nil {
		return nil, false, true, err
	}
	if method != nil {
		vm.storeCachedIndexMethod(program, ip, methodName, receiver, method, true)
		return method, true, true, nil
	}
	vm.storeCachedIndexMethod(program, ip, methodName, receiver, nil, false)
	return nil, false, true, nil
}
