package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeIndexMethodCacheEntry struct {
	globalRevision     uint64
	receiverKind       bytecodeMemberReceiverKind
	arrayElemType      uint16
	methodCacheVersion uint64
	method             runtime.Value
	hasMethod          bool
}

type bytecodeIndexMethodCacheTable struct {
	get []bytecodeIndexMethodCacheEntry
	set []bytecodeIndexMethodCacheEntry
}

type bytecodeInlineIndexMethodCacheEntry struct {
	valid              bool
	program            *bytecodeProgram
	ip                 int
	method             string
	globalRevision     uint64
	receiverKind       bytecodeMemberReceiverKind
	arrayElemType      uint16
	methodCacheVersion uint64
	resolvedMethod     runtime.Value
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

func bytecodeArrayElementTypeTokenFromValues(values []runtime.Value) (uint16, bool) {
	if len(values) == 0 {
		return bytecodeIndexTypeUnknown, true
	}
	return bytecodeIndexValueTypeToken(values[0])
}

func bytecodeArrayElementTypeToken(arr *runtime.ArrayValue) (uint16, bool) {
	if arr == nil {
		return bytecodeIndexTypeUnknown, false
	}
	if arr.State != nil && arr.State.ElementTypeTokenKnown {
		return arr.State.ElementTypeToken, true
	}
	return bytecodeArrayElementTypeTokenFromValues(arr.Elements)
}

func (vm *bytecodeVM) indexMethodCacheIdentity(receiver runtime.Value) (bytecodeMemberReceiverKind, uint16, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, false
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, false
	}
	elemType, ok := bytecodeArrayElementTypeToken(arr)
	if !ok {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, false
	}
	return bytecodeMemberReceiverArray, elemType, true
}

func (vm *bytecodeVM) indexMethodCacheIdentityKey(receiver runtime.Value) (bytecodeMemberReceiverKind, uint16, uint64, uint64, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, 0, 0, false
	}
	receiverKind, elemType, ok := vm.indexMethodCacheIdentity(receiver)
	if !ok {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, 0, 0, false
	}
	globalRevision := vm.bytecodeGlobalRevision()
	methodCacheVersion := vm.bytecodeMethodCacheVersion()
	return receiverKind, elemType, globalRevision, methodCacheVersion, true
}

func (vm *bytecodeVM) indexMethodCacheEntry(program *bytecodeProgram, ip int, methodName string, create bool) (*bytecodeIndexMethodCacheEntry, bool) {
	if vm == nil || program == nil || ip < 0 || ip >= len(program.instructions) {
		return nil, false
	}
	table, ok := vm.indexMethodCache[program]
	if !ok {
		if !create {
			return nil, false
		}
		table = &bytecodeIndexMethodCacheTable{}
		if vm.indexMethodCache == nil {
			vm.indexMethodCache = make(map[*bytecodeProgram]*bytecodeIndexMethodCacheTable, 8)
		}
		vm.indexMethodCache[program] = table
	}
	switch methodName {
	case "get":
		if table.get == nil {
			if !create {
				return nil, false
			}
			table.get = make([]bytecodeIndexMethodCacheEntry, len(program.instructions))
		}
		return &table.get[ip], true
	case "set":
		if table.set == nil {
			if !create {
				return nil, false
			}
			table.set = make([]bytecodeIndexMethodCacheEntry, len(program.instructions))
		}
		return &table.set[ip], true
	default:
		return nil, false
	}
}

func (vm *bytecodeVM) lookupCachedIndexMethod(program *bytecodeProgram, ip int, methodName string, receiverKind bytecodeMemberReceiverKind, elemType uint16, globalRevision uint64, methodCacheVersion uint64) (runtime.Value, bool, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return nil, false, false
	}
	if hot := vm.indexMethodHot; hot.valid &&
		hot.program == program &&
		hot.ip == ip &&
		hot.method == methodName &&
		hot.globalRevision == globalRevision &&
		hot.receiverKind == receiverKind &&
		hot.arrayElemType == elemType &&
		hot.methodCacheVersion == methodCacheVersion {
		return hot.resolvedMethod, true, hot.hasMethod
	}
	entry, ok := vm.indexMethodCacheEntry(program, ip, methodName, false)
	if !ok {
		return nil, false, false
	}
	if entry.globalRevision != globalRevision {
		return nil, false, false
	}
	if entry.methodCacheVersion != methodCacheVersion {
		return nil, false, false
	}
	if entry.receiverKind != receiverKind || entry.arrayElemType != elemType {
		return nil, false, false
	}
	vm.indexMethodHot = bytecodeInlineIndexMethodCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		method:             methodName,
		globalRevision:     entry.globalRevision,
		receiverKind:       receiverKind,
		arrayElemType:      elemType,
		methodCacheVersion: entry.methodCacheVersion,
		resolvedMethod:     entry.method,
		hasMethod:          entry.hasMethod,
	}
	return entry.method, true, entry.hasMethod
}

func (vm *bytecodeVM) storeCachedIndexMethod(program *bytecodeProgram, ip int, methodName string, receiverKind bytecodeMemberReceiverKind, elemType uint16, globalRevision uint64, methodCacheVersion uint64, method runtime.Value, hasMethod bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return
	}
	if hasMethod && method == nil {
		return
	}
	entry, ok := vm.indexMethodCacheEntry(program, ip, methodName, true)
	if !ok {
		return
	}
	*entry = bytecodeIndexMethodCacheEntry{
		globalRevision:     globalRevision,
		receiverKind:       receiverKind,
		arrayElemType:      elemType,
		methodCacheVersion: methodCacheVersion,
		method:             method,
		hasMethod:          hasMethod,
	}
	vm.indexMethodHot = bytecodeInlineIndexMethodCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		method:             methodName,
		globalRevision:     globalRevision,
		receiverKind:       receiverKind,
		arrayElemType:      elemType,
		methodCacheVersion: methodCacheVersion,
		resolvedMethod:     method,
		hasMethod:          hasMethod,
	}
}

func (vm *bytecodeVM) lookupHotDirectArrayIndexSite(methodName string, receiver runtime.Value) (*runtime.ArrayValue, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return nil, false
	}
	hot := vm.indexMethodHot
	if !hot.valid ||
		hot.program != vm.currentProgram ||
		hot.ip != vm.ip ||
		hot.method != methodName ||
		hot.hasMethod ||
		hot.receiverKind != bytecodeMemberReceiverArray {
		return nil, false
	}
	if hot.globalRevision != vm.bytecodeGlobalRevision() || hot.methodCacheVersion != vm.bytecodeMethodCacheVersion() {
		return nil, false
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return nil, false
	}
	elemType, ok := bytecodeArrayElementTypeToken(arr)
	if !ok || elemType != hot.arrayElemType {
		return nil, false
	}
	return arr, true
}

func (vm *bytecodeVM) resolveIndexGet(obj runtime.Value, idxVal runtime.Value) (runtime.Value, error) {
	if vm != nil && vm.interp != nil && vm.interp.canUseDirectArrayIndexGetFastPath() {
		if arr, ok := bytecodeArrayReceiverForIndexCache(obj); ok {
			if result, handled, err := vm.resolveDirectArrayIndexGet(arr, idxVal); handled {
				return result, err
			}
		}
	}
	if arr, ok := vm.lookupHotDirectArrayIndexSite("get", obj); ok {
		result, _, err := vm.resolveDirectArrayIndexGet(arr, idxVal)
		return result, err
	}
	method, hasMethod, cacheable, err := vm.resolveCachedIndexMethod(vm.currentProgram, vm.ip, obj, "get", "Index")
	if err != nil {
		return nil, err
	}
	if cacheable {
		if hasMethod {
			return vm.interp.CallFunction(method, []runtime.Value{obj, idxVal})
		}
		if arr, ok := obj.(*runtime.ArrayValue); ok {
			if result, handled, err := vm.resolveDirectArrayIndexGet(arr, idxVal); handled {
				return result, err
			}
		}
		return vm.interp.indexGetWithoutMethod(obj, idxVal)
	}
	return vm.interp.indexGet(obj, idxVal)
}

func (vm *bytecodeVM) resolveIndexSet(obj runtime.Value, idxVal runtime.Value, value runtime.Value, op ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, error) {
	if vm != nil && vm.interp != nil && vm.interp.canUseDirectArrayIndexSetFastPath() {
		if arr, ok := bytecodeArrayReceiverForIndexCache(obj); ok {
			if result, handled, err := vm.resolveDirectArrayIndexSet(arr, idxVal, value, op, binaryOp, isCompound); handled {
				return result, err
			}
		}
	}
	if arr, ok := vm.lookupHotDirectArrayIndexSite("set", obj); ok {
		result, _, err := vm.resolveDirectArrayIndexSet(arr, idxVal, value, op, binaryOp, isCompound)
		return result, err
	}
	setMethod, hasSetMethod, cacheable, err := vm.resolveCachedIndexMethod(vm.currentProgram, vm.ip, obj, "set", "IndexMut")
	if err != nil {
		return nil, err
	}
	if !cacheable {
		return vm.interp.assignIndex(obj, idxVal, value, op, binaryOp, isCompound)
	}
	if !hasSetMethod {
		if arr, ok := obj.(*runtime.ArrayValue); ok {
			if result, handled, err := vm.resolveDirectArrayIndexSet(arr, idxVal, value, op, binaryOp, isCompound); handled {
				return result, err
			}
		}
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
	receiverKind, elemType, globalRevision, methodCacheVersion, cacheable := vm.indexMethodCacheIdentityKey(receiver)
	if !cacheable {
		return nil, false, false, nil
	}
	if method, cached, hasMethod := vm.lookupCachedIndexMethod(program, ip, methodName, receiverKind, elemType, globalRevision, methodCacheVersion); cached {
		return method, hasMethod, true, nil
	}
	method, err := vm.interp.findIndexMethod(receiver, methodName, iface)
	if err != nil {
		return nil, false, true, err
	}
	if method != nil {
		vm.storeCachedIndexMethod(program, ip, methodName, receiverKind, elemType, globalRevision, methodCacheVersion, method, true)
		return method, true, true, nil
	}
	vm.storeCachedIndexMethod(program, ip, methodName, receiverKind, elemType, globalRevision, methodCacheVersion, nil, false)
	return nil, false, true, nil
}

func bytecodeDirectArrayIndex(idxVal runtime.Value) (int, bool, error) {
	switch idx := idxVal.(type) {
	case runtime.IntegerValue:
		if idx.IsSmall() {
			raw := idx.Int64Fast()
			if raw < math.MinInt || raw > math.MaxInt {
				return 0, true, fmt.Errorf("Array index must be within int range")
			}
			return int(raw), true, nil
		}
	case *runtime.IntegerValue:
		if idx != nil && idx.IsSmall() {
			raw := idx.Int64Fast()
			if raw < math.MinInt || raw > math.MaxInt {
				return 0, true, fmt.Errorf("Array index must be within int range")
			}
			return int(raw), true, nil
		}
	}
	idxInt, ok := bytecodeDirectIntegerValue(idxVal)
	if !ok {
		idxInt, ok = bytecodeIntegerValue(idxVal)
	}
	if !ok {
		return 0, false, nil
	}
	raw, fits := idxInt.ToInt64()
	if !fits {
		return 0, true, fmt.Errorf("Array index must be within int range")
	}
	if raw < math.MinInt || raw > math.MaxInt {
		return 0, true, fmt.Errorf("Array index must be within int range")
	}
	return int(raw), true, nil
}

func bytecodeTrackedArrayState(arr *runtime.ArrayValue) (*runtime.ArrayState, bool) {
	if arr == nil || arr.State == nil || arr.Handle == 0 || arr.TrackedHandle != arr.Handle {
		return nil, false
	}
	return arr.State, true
}

func (vm *bytecodeVM) resolveDirectArrayIndexGet(arr *runtime.ArrayValue, idxVal runtime.Value) (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil || arr == nil {
		return nil, false, nil
	}
	idx, ok, err := bytecodeDirectArrayIndex(idxVal)
	if err != nil || !ok {
		return nil, ok, err
	}
	state, tracked := bytecodeTrackedArrayState(arr)
	if !tracked {
		var err error
		state, err = vm.interp.ensureArrayState(arr, 0)
		if err != nil {
			return nil, true, err
		}
	}
	if idx < 0 || idx >= len(state.Values) {
		return vm.interp.makeIndexErrorValue(idx, len(state.Values)), true, nil
	}
	val := state.Values[idx]
	if val == nil {
		return vm.interp.makeIndexErrorValue(idx, len(state.Values)), true, nil
	}
	return val, true, nil
}

func (vm *bytecodeVM) resolveDirectArrayIndexSet(arr *runtime.ArrayValue, idxVal runtime.Value, value runtime.Value, op ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil || arr == nil {
		return nil, false, nil
	}
	if op == ast.AssignmentDeclare {
		return nil, true, fmt.Errorf("Cannot use := on index assignment")
	}
	idx, ok, err := bytecodeDirectArrayIndex(idxVal)
	if err != nil || !ok {
		return nil, ok, err
	}
	state, tracked := bytecodeTrackedArrayState(arr)
	if !tracked {
		var err error
		state, err = vm.interp.ensureArrayState(arr, 0)
		if err != nil {
			return nil, true, err
		}
	}
	if idx < 0 || idx >= len(state.Values) {
		return nil, true, fmt.Errorf("Array index out of bounds")
	}
	if op == ast.AssignmentAssign {
		state.Values[idx] = value
		vm.interp.syncArrayValues(arr.Handle, state)
		return value, true, nil
	}
	if !isCompound {
		return nil, true, fmt.Errorf("unsupported assignment operator %s", op)
	}
	current := state.Values[idx]
	computed, err := applyBinaryOperator(vm.interp, binaryOp, current, value)
	if err != nil {
		return nil, true, err
	}
	state.Values[idx] = computed
	vm.interp.syncArrayValues(arr.Handle, state)
	return computed, true, nil
}
