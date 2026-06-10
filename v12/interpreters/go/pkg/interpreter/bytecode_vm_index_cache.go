package interpreter

import (
	"fmt"
	"math"
	"strconv"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) indexMethodCacheIdentity(receiver runtime.Value) (bytecodeMemberReceiverKind, uint16, string, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, "", false
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, "", false
	}
	elemType, typeKey, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, "", false
	}
	return bytecodeMemberReceiverArray, elemType, typeKey, true
}

func (vm *bytecodeVM) indexMethodCacheIdentityKey(receiver runtime.Value) (bytecodeMemberReceiverKind, uint16, string, uint64, uint64, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, "", 0, 0, false
	}
	receiverKind, elemType, typeKey, ok := vm.indexMethodCacheIdentity(receiver)
	if !ok {
		return bytecodeMemberReceiverUnknown, bytecodeIndexTypeUnknown, "", 0, 0, false
	}
	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	return receiverKind, elemType, typeKey, globalRevision, methodCacheVersion, true
}

func (vm *bytecodeVM) indexMethodCacheEntry(program *bytecodeProgram, ip int, methodName string, create bool) (*bytecodeIndexMethodCacheEntry, bool) {
	return vm.indexMethodCacheEntryForKind(program, ip, bytecodeIndexMethodCacheKindFor(methodName), create)
}

func (vm *bytecodeVM) indexMethodCacheEntryForKind(program *bytecodeProgram, ip int, methodKind bytecodeIndexMethodCacheKind, create bool) (*bytecodeIndexMethodCacheEntry, bool) {
	if vm == nil || program == nil || ip < 0 || ip >= len(program.instructions) {
		return nil, false
	}
	table := vm.activeLookup.indexMethodTable
	if vm.activeLookup.program != program || table == nil {
		var ok bool
		table, ok = vm.indexMethodCache[program]
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
		if vm.activeLookup.program == program {
			vm.activeLookup.indexMethodTable = table
			vm.activeLookup.indexMethodGetEntries = table.get
			vm.activeLookup.indexMethodSetEntries = table.set
		}
	}
	switch methodKind {
	case bytecodeIndexMethodCacheGet:
		if table.get == nil {
			if !create {
				return nil, false
			}
			table.get = make([]bytecodeIndexMethodCacheEntry, len(program.instructions))
			if vm.activeLookup.program == program {
				vm.activeLookup.indexMethodGetEntries = table.get
			}
		}
		return &table.get[ip], true
	case bytecodeIndexMethodCacheSet:
		if table.set == nil {
			if !create {
				return nil, false
			}
			table.set = make([]bytecodeIndexMethodCacheEntry, len(program.instructions))
			if vm.activeLookup.program == program {
				vm.activeLookup.indexMethodSetEntries = table.set
			}
		}
		return &table.set[ip], true
	default:
		return nil, false
	}
}

func (vm *bytecodeVM) lookupCachedIndexMethod(program *bytecodeProgram, ip int, methodName string, receiverKind bytecodeMemberReceiverKind, elemType uint16, receiverTypeKey string, globalRevision uint64, methodCacheVersion uint64) (runtime.Value, bytecodeIndexMethodFastPathKind, bool, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return nil, bytecodeIndexMethodFastPathNone, false, false
	}
	methodKind := bytecodeIndexMethodCacheKindFor(methodName)
	if methodKind == bytecodeIndexMethodCacheUnknown {
		return nil, bytecodeIndexMethodFastPathNone, false, false
	}
	if hot := &vm.indexMethodHot; hot.valid &&
		hot.program == program &&
		hot.ip == ip &&
		hot.methodKind == methodKind &&
		hot.globalRevision == globalRevision &&
		hot.receiverKind == receiverKind &&
		hot.arrayElemType == elemType &&
		hot.receiverTypeKey == receiverTypeKey &&
		hot.methodCacheVersion == methodCacheVersion {
		return hot.resolvedMethod, hot.fastPath, true, hot.hasMethod
	}
	entry, ok := vm.indexMethodCacheEntryForKind(program, ip, methodKind, false)
	if !ok {
		return nil, bytecodeIndexMethodFastPathNone, false, false
	}
	if entry.globalRevision != globalRevision {
		return nil, bytecodeIndexMethodFastPathNone, false, false
	}
	if entry.methodCacheVersion != methodCacheVersion {
		return nil, bytecodeIndexMethodFastPathNone, false, false
	}
	if entry.receiverKind != receiverKind || entry.arrayElemType != elemType || entry.receiverTypeKey != receiverTypeKey {
		return nil, bytecodeIndexMethodFastPathNone, false, false
	}
	hot := &vm.indexMethodHot
	hot.valid = true
	hot.program = program
	hot.ip = ip
	hot.methodKind = methodKind
	hot.globalRevision = entry.globalRevision
	hot.receiverKind = receiverKind
	hot.arrayElemType = elemType
	hot.receiverTypeKey = receiverTypeKey
	hot.receiverArrayHandle = entry.receiverArrayHandle
	hot.receiverArrayRev = entry.receiverArrayRev
	hot.receiverArrayRevOK = entry.receiverArrayRevOK
	hot.receiverArrayCursor = entry.receiverArrayCursor
	hot.methodCacheVersion = entry.methodCacheVersion
	hot.resolvedMethod = entry.method
	hot.hasMethod = entry.hasMethod
	hot.fastPath = entry.fastPath
	vm.storeIndexMethodDirect(*hot)
	return entry.method, entry.fastPath, true, entry.hasMethod
}

func (vm *bytecodeVM) storeCachedIndexMethod(program *bytecodeProgram, ip int, methodName string, receiver runtime.Value, receiverKind bytecodeMemberReceiverKind, elemType uint16, receiverTypeKey string, globalRevision uint64, methodCacheVersion uint64, method runtime.Value, hasMethod bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return
	}
	if hasMethod && method == nil {
		return
	}
	methodKind := bytecodeIndexMethodCacheKindFor(methodName)
	if methodKind == bytecodeIndexMethodCacheUnknown {
		return
	}
	entry, ok := vm.indexMethodCacheEntryForKind(program, ip, methodKind, true)
	if !ok {
		return
	}
	fastPath := bytecodeIndexMethodFastPathNone
	if hasMethod {
		fastPath = vm.indexMethodFastPathFor(methodName, method)
	}
	receiverArrayHandle, receiverArrayRev, receiverArrayCursor, receiverArrayRevOK := vm.indexMethodReceiverRevisionWithCursor(receiver)
	*entry = bytecodeIndexMethodCacheEntry{
		globalRevision:      globalRevision,
		receiverKind:        receiverKind,
		arrayElemType:       elemType,
		receiverTypeKey:     receiverTypeKey,
		receiverArrayHandle: receiverArrayHandle,
		receiverArrayRev:    receiverArrayRev,
		receiverArrayRevOK:  receiverArrayRevOK,
		receiverArrayCursor: receiverArrayCursor,
		methodCacheVersion:  methodCacheVersion,
		method:              method,
		hasMethod:           hasMethod,
		fastPath:            fastPath,
	}
	hot := &vm.indexMethodHot
	hot.valid = true
	hot.program = program
	hot.ip = ip
	hot.methodKind = methodKind
	hot.globalRevision = globalRevision
	hot.receiverKind = receiverKind
	hot.arrayElemType = elemType
	hot.receiverTypeKey = receiverTypeKey
	hot.receiverArrayHandle = receiverArrayHandle
	hot.receiverArrayRev = receiverArrayRev
	hot.receiverArrayRevOK = receiverArrayRevOK
	hot.receiverArrayCursor = receiverArrayCursor
	hot.methodCacheVersion = methodCacheVersion
	hot.resolvedMethod = method
	hot.hasMethod = hasMethod
	hot.fastPath = fastPath
	vm.storeIndexMethodDirect(*hot)
}

func (vm *bytecodeVM) resolveIndexGet(obj runtime.Value, idxVal runtime.Value) (runtime.Value, error) {
	result, _, _, err := vm.resolveIndexGetWithToken(obj, idxVal)
	return result, err
}

func (vm *bytecodeVM) resolveIndexGetWithToken(obj runtime.Value, idxVal runtime.Value) (runtime.Value, uint16, bool, error) {
	var (
		arr   *runtime.ArrayValue
		arrOK bool
	)
	if vm != nil && vm.interp != nil {
		arr, arrOK = bytecodeArrayReceiverForIndexCache(obj)
		if arrOK && vm.interp.canUseDirectArrayIndexGetFastPath() {
			if result, token, tokenKnown, handled, err := vm.resolveDirectArrayIndexGetWithHandleAndToken(arr, idxVal, 0); handled {
				return result, token, tokenKnown, err
			}
		}
	}
	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	if arrOK {
		if arr, handle, ok := vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersions(bytecodeIndexMethodCacheGet, arr, bytecodeIndexMethodFastPathCanonicalArrayGet, globalRevision, methodCacheVersion); ok {
			result, token, tokenKnown, _, err := vm.resolveDirectArrayIndexGetWithValidatedHandleAndToken(arr, idxVal, handle)
			return result, token, tokenKnown, err
		}
	} else if arr, handle, ok := vm.lookupDirectCompatibleHotArrayIndexSiteWithVersions(bytecodeIndexMethodCacheGet, obj, bytecodeIndexMethodFastPathCanonicalArrayGet, globalRevision, methodCacheVersion); ok {
		result, token, tokenKnown, _, err := vm.resolveDirectArrayIndexGetWithValidatedHandleAndToken(arr, idxVal, handle)
		return result, token, tokenKnown, err
	}
	var (
		method    runtime.Value
		fastPath  bytecodeIndexMethodFastPathKind
		hasMethod bool
		cacheable bool
		err       error
	)
	if arrOK {
		method, fastPath, hasMethod, cacheable, err = vm.resolveCachedArrayIndexMethodWithVersions(vm.currentProgram, vm.ip, obj, arr, "get", "Index", globalRevision, methodCacheVersion)
	} else {
		method, fastPath, hasMethod, cacheable, err = vm.resolveCachedIndexMethodWithVersions(vm.currentProgram, vm.ip, obj, "get", "Index", globalRevision, methodCacheVersion)
	}
	if err != nil {
		return nil, bytecodeIndexTypeUnknown, false, err
	}
	if cacheable {
		if hasMethod {
			if arrOK && fastPath == bytecodeIndexMethodFastPathCanonicalArrayGet {
				if result, token, tokenKnown, handled, err := vm.resolveDirectArrayIndexGetWithHandleAndToken(arr, idxVal, 0); handled {
					return result, token, tokenKnown, err
				}
			}
			result, err := vm.interp.CallFunction(method, []runtime.Value{obj, idxVal})
			return result, bytecodeIndexTypeUnknown, false, err
		}
		if arr, ok := obj.(*runtime.ArrayValue); ok {
			if result, token, tokenKnown, handled, err := vm.resolveDirectArrayIndexGetWithHandleAndToken(arr, idxVal, 0); handled {
				return result, token, tokenKnown, err
			}
		}
		result, err := vm.interp.indexGetWithoutMethod(obj, idxVal)
		return result, bytecodeIndexTypeUnknown, false, err
	}
	result, err := vm.interp.indexGet(obj, idxVal)
	return result, bytecodeIndexTypeUnknown, false, err
}

func (vm *bytecodeVM) execArrayIndexSetSlot(instr *bytecodeInstruction) error {
	if vm == nil || vm.interp == nil || instr == nil {
		return fmt.Errorf("bytecode array index set slot missing VM or instruction")
	}
	objSlot, idxSlot := instr.argCount, instr.loopBreak
	if objSlot < 0 || objSlot >= len(vm.slots) || idxSlot < 0 || idxSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode array index set slot out of range")
	}
	if vm.stackDepth() == 0 {
		return fmt.Errorf("bytecode stack underflow")
	}
	valueIdx := vm.stackDepth() - 1
	value := vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, vm.stackValue(valueIdx))
	obj := vm.slots[objSlot]
	if !vm.hasI32RegisterFrame() {
		idxVal := vm.slots[idxSlot]
		var (
			result  runtime.Value
			err     error
			handled bool
		)
		if vm.interp.canUseDirectArrayIndexSetFastPath() {
			if arr, ok := obj.(*runtime.ArrayValue); ok && arr != nil {
				if idx, small := bytecodeDirectSmallArrayIndex(idxVal); small {
					result, err = vm.resolveDirectArrayIndexSetAt(arr, idx, value)
					handled = true
				} else if directResult, directHandled, directErr := vm.resolveDirectArrayIndexSet(arr, idxVal, value, ast.AssignmentAssign, "", false); directHandled {
					result, err = directResult, directErr
					handled = true
				}
			}
		}
		if !handled && err == nil {
			result, err = vm.resolveIndexSet(obj, idxVal, value, ast.AssignmentAssign, "", false)
		}
		if err != nil {
			err = vm.interp.wrapStandardRuntimeError(err)
			if instr.node != nil {
				err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
			}
			return err
		}
		vm.setStackValue(valueIdx, bytecodeStackResultValue(result))
		vm.ip++
		return nil
	}
	var (
		result  runtime.Value
		err     error
		handled bool
		idxVal  runtime.Value
	)
	if vm.interp.canUseDirectArrayIndexSetFastPath() {
		if arr, ok := obj.(*runtime.ArrayValue); ok && arr != nil {
			if idx, small := vm.slotDirectSmallArrayIndexValidated(idxSlot); small {
				result, err = vm.resolveDirectArrayIndexSetAt(arr, idx, value)
				handled = true
			} else if idxVal = vm.slotMaterializedValue(idxSlot); idxVal != nil {
				if directResult, directHandled, directErr := vm.resolveDirectArrayIndexSet(arr, idxVal, value, ast.AssignmentAssign, "", false); directHandled {
					result, err = directResult, directErr
					handled = true
				}
			}
		}
	}
	if !handled && err == nil {
		if idxVal == nil {
			idxVal = vm.slotMaterializedValue(idxSlot)
		}
		result, err = vm.resolveIndexSet(obj, idxVal, value, ast.AssignmentAssign, "", false)
	}
	if err != nil {
		err = vm.interp.wrapStandardRuntimeError(err)
		if instr.node != nil {
			err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return err
	}
	vm.setStackValue(valueIdx, bytecodeStackResultValue(result))
	vm.ip++
	return nil
}

func (vm *bytecodeVM) resolveIndexSet(obj runtime.Value, idxVal runtime.Value, value runtime.Value, op ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, error) {
	if vm != nil && vm.interp != nil && vm.interp.canUseDirectArrayIndexSetFastPath() {
		if arr, ok := bytecodeArrayReceiverForIndexCache(obj); ok {
			if result, handled, err := vm.resolveDirectArrayIndexSet(arr, idxVal, value, op, binaryOp, isCompound); handled {
				return result, err
			}
		}
	}
	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	if arr, _, ok := vm.lookupDirectCompatibleHotArrayIndexSiteWithVersions(bytecodeIndexMethodCacheSet, obj, bytecodeIndexMethodFastPathCanonicalArraySet, globalRevision, methodCacheVersion); ok {
		result, _, err := vm.resolveDirectArrayIndexSet(arr, idxVal, value, op, binaryOp, isCompound)
		return result, err
	}
	setMethod, setFastPath, hasSetMethod, cacheable, err := vm.resolveCachedIndexMethodWithVersions(vm.currentProgram, vm.ip, obj, "set", "IndexMut", globalRevision, methodCacheVersion)
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
	if arr, ok := bytecodeArrayReceiverForIndexCache(obj); ok && setFastPath == bytecodeIndexMethodFastPathCanonicalArraySet {
		if result, handled, err := vm.resolveDirectArrayIndexSet(arr, idxVal, value, op, binaryOp, isCompound); handled {
			return result, err
		}
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
	getMethod, _, hasGetMethod, _, err := vm.resolveCachedIndexMethodWithVersions(vm.currentProgram, vm.ip, obj, "get", "Index", globalRevision, methodCacheVersion)
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

func (vm *bytecodeVM) resolveCachedIndexMethod(program *bytecodeProgram, ip int, receiver runtime.Value, methodName string, iface string) (runtime.Value, bytecodeIndexMethodFastPathKind, bool, bool, error) {
	receiverKind, elemType, receiverTypeKey, globalRevision, methodCacheVersion, cacheable := vm.indexMethodCacheIdentityKey(receiver)
	if !cacheable {
		return nil, bytecodeIndexMethodFastPathNone, false, false, nil
	}
	return vm.resolveCachedIndexMethodForIdentity(program, ip, receiver, methodName, iface, receiverKind, elemType, receiverTypeKey, globalRevision, methodCacheVersion)
}

func (vm *bytecodeVM) resolveCachedIndexMethodWithVersions(program *bytecodeProgram, ip int, receiver runtime.Value, methodName string, iface string, globalRevision uint64, methodCacheVersion uint64) (runtime.Value, bytecodeIndexMethodFastPathKind, bool, bool, error) {
	receiverKind, elemType, receiverTypeKey, cacheable := vm.indexMethodCacheIdentity(receiver)
	if !cacheable {
		return nil, bytecodeIndexMethodFastPathNone, false, false, nil
	}
	return vm.resolveCachedIndexMethodForIdentity(program, ip, receiver, methodName, iface, receiverKind, elemType, receiverTypeKey, globalRevision, methodCacheVersion)
}

func (vm *bytecodeVM) resolveCachedArrayIndexMethodWithVersions(program *bytecodeProgram, ip int, receiver runtime.Value, arr *runtime.ArrayValue, methodName string, iface string, globalRevision uint64, methodCacheVersion uint64) (runtime.Value, bytecodeIndexMethodFastPathKind, bool, bool, error) {
	elemType, receiverTypeKey, cacheable := vm.arrayIndexReceiverIdentity(arr)
	if !cacheable {
		return nil, bytecodeIndexMethodFastPathNone, false, false, nil
	}
	return vm.resolveCachedIndexMethodForIdentity(program, ip, receiver, methodName, iface, bytecodeMemberReceiverArray, elemType, receiverTypeKey, globalRevision, methodCacheVersion)
}

func (vm *bytecodeVM) resolveCachedIndexMethodForIdentity(program *bytecodeProgram, ip int, receiver runtime.Value, methodName string, iface string, receiverKind bytecodeMemberReceiverKind, elemType uint16, receiverTypeKey string, globalRevision uint64, methodCacheVersion uint64) (runtime.Value, bytecodeIndexMethodFastPathKind, bool, bool, error) {
	if method, fastPath, cached, hasMethod := vm.lookupCachedIndexMethod(program, ip, methodName, receiverKind, elemType, receiverTypeKey, globalRevision, methodCacheVersion); cached {
		return method, fastPath, hasMethod, true, nil
	}
	method, err := vm.interp.findIndexMethod(receiver, methodName, iface)
	if err != nil {
		return nil, bytecodeIndexMethodFastPathNone, false, true, err
	}
	if method != nil {
		vm.storeCachedIndexMethod(program, ip, methodName, receiver, receiverKind, elemType, receiverTypeKey, globalRevision, methodCacheVersion, method, true)
		return method, vm.indexMethodFastPathFor(methodName, method), true, true, nil
	}
	vm.storeCachedIndexMethod(program, ip, methodName, receiver, receiverKind, elemType, receiverTypeKey, globalRevision, methodCacheVersion, nil, false)
	return nil, bytecodeIndexMethodFastPathNone, false, true, nil
}

func (vm *bytecodeVM) indexMethodFastPathFor(methodName string, method runtime.Value) bytecodeIndexMethodFastPathKind {
	switch value := method.(type) {
	case runtime.BoundMethodValue:
		method = value.Method
	case *runtime.BoundMethodValue:
		if value == nil {
			return bytecodeIndexMethodFastPathNone
		}
		method = value.Method
	}
	fn, ok := bytecodeSingleFunction(method)
	if !ok || fn == nil {
		return bytecodeIndexMethodFastPathNone
	}
	def, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || def == nil || def.ID == nil || def.ID.Name != methodName || vm == nil || vm.interp == nil {
		return bytecodeIndexMethodFastPathNone
	}
	origin := vm.interp.nodeOrigins[def]
	if !isCanonicalAbleStdlibOrigin(origin, "collections/array.able") {
		return bytecodeIndexMethodFastPathNone
	}
	switch methodName {
	case "get":
		if len(def.Params) == 2 &&
			typeExpressionToString(def.Params[1].ParamType) == "i32" {
			if _, ok := def.ReturnType.(*ast.ResultTypeExpression); ok && fn.MethodPriority < 0 {
				return bytecodeIndexMethodFastPathCanonicalArrayGet
			}
		}
	case "set":
		if len(def.Params) == 3 &&
			typeExpressionToString(def.Params[1].ParamType) == "i32" &&
			fn.MethodPriority < 0 {
			if resultType, ok := def.ReturnType.(*ast.ResultTypeExpression); ok &&
				typeExpressionToString(resultType.InnerType) == "void" {
				return bytecodeIndexMethodFastPathCanonicalArraySet
			}
		}
	}
	return bytecodeIndexMethodFastPathNone
}

func bytecodeInt64ToArrayIndex(raw int64) (int, error) {
	if strconv.IntSize == 64 {
		return int(raw), nil
	}
	if raw < math.MinInt || raw > math.MaxInt {
		return 0, fmt.Errorf("Array index must be within int range")
	}
	return int(raw), nil
}

func bytecodeDirectArrayIndexFromInteger(idx runtime.IntegerValue) (int, bool, error) {
	if idx.IsSmall() {
		if strconv.IntSize == 64 {
			return int(idx.Int64Fast()), true, nil
		}
		value, err := bytecodeInt64ToArrayIndex(idx.Int64Fast())
		return value, true, err
	}
	raw, fits := idx.ToInt64()
	if !fits {
		return 0, true, fmt.Errorf("Array index must be within int range")
	}
	value, err := bytecodeInt64ToArrayIndex(raw)
	return value, true, err
}

func bytecodeDirectArrayIndexFromRaw(raw int64) (int, bool, error) {
	value, err := bytecodeInt64ToArrayIndex(raw)
	return value, true, err
}

func bytecodeDirectArrayIndex(idxVal runtime.Value) (int, bool, error) {
	switch idx := idxVal.(type) {
	case bytecodeRawI32SlotValue:
		return bytecodeDirectArrayIndexFromRaw(int64(idx))
	case *bytecodeRawI32StackCell:
		if idx == nil {
			return 0, false, nil
		}
		return bytecodeDirectArrayIndexFromRaw(int64(idx.Val))
	case bytecodeRawU8ResultValue:
		return bytecodeDirectArrayIndexFromRaw(int64(idx))
	case bytecodeRawU16ResultValue:
		return bytecodeDirectArrayIndexFromRaw(int64(idx))
	case bytecodeRawU32ResultValue:
		return bytecodeDirectArrayIndexFromRaw(int64(idx))
	case bytecodeRawU64ResultValue:
		return bytecodeDirectArrayIndexFromRaw(int64(idx))
	case bytecodeRawUsizeResultValue:
		return bytecodeDirectArrayIndexFromRaw(int64(idx))
	case bytecodeRawI64ResultValue:
		return bytecodeDirectArrayIndexFromRaw(int64(idx))
	case bytecodeRawIntegerValue:
		return bytecodeDirectArrayIndexFromRaw(idx.Raw)
	case *bytecodeRawIntegerSlotCell:
		if idx == nil {
			return 0, false, nil
		}
		return bytecodeDirectArrayIndexFromRaw(idx.Raw)
	case *bytecodeRawI64SlotCell:
		if idx == nil {
			return 0, false, nil
		}
		return bytecodeDirectArrayIndexFromRaw(idx.Val)
	case runtime.IntegerValue:
		if idx.IsSmall() && strconv.IntSize == 64 {
			return int(idx.Int64Fast()), true, nil
		}
		return bytecodeDirectArrayIndexFromInteger(idx)
	case *runtime.IntegerValue:
		if idx == nil {
			return 0, false, nil
		}
		if idx.IsSmall() && strconv.IntSize == 64 {
			return int(idx.Int64Fast()), true, nil
		}
		return bytecodeDirectArrayIndexFromInteger(*idx)
	}
	raw := unwrapScalarValue(unwrapInterfaceValue(idxVal))
	switch idx := raw.(type) {
	case runtime.IntegerValue:
		return bytecodeDirectArrayIndexFromInteger(idx)
	case *runtime.IntegerValue:
		if idx != nil {
			return bytecodeDirectArrayIndexFromInteger(*idx)
		}
	}
	return 0, false, nil
}

func bytecodeDirectSmallArrayIndex(idxVal runtime.Value) (int, bool) {
	if strconv.IntSize != 64 {
		return 0, false
	}
	switch idx := idxVal.(type) {
	case bytecodeRawI32SlotValue:
		return int(idx), true
	case *bytecodeRawI32StackCell:
		if idx != nil {
			return int(idx.Val), true
		}
	case bytecodeRawU8ResultValue:
		return int(idx), true
	case bytecodeRawU16ResultValue:
		return int(idx), true
	case bytecodeRawU32ResultValue:
		return int(idx), true
	case bytecodeRawU64ResultValue:
		return int(int64(idx)), true
	case bytecodeRawUsizeResultValue:
		return int(int64(idx)), true
	case bytecodeRawI64ResultValue:
		return int(idx), true
	case bytecodeRawIntegerValue:
		return int(idx.Raw), true
	case *bytecodeRawIntegerSlotCell:
		if idx != nil {
			return int(idx.Raw), true
		}
	case *bytecodeRawI64SlotCell:
		if idx != nil {
			return int(idx.Val), true
		}
	case runtime.IntegerValue:
		if idx.IsSmall() {
			return int(idx.Int64Fast()), true
		}
	case *runtime.IntegerValue:
		if idx != nil && idx.IsSmallRef() {
			return int(idx.Int64FastRef()), true
		}
	}
	return 0, false
}

func bytecodeTrackedArrayState(arr *runtime.ArrayValue) (*runtime.ArrayState, bool) {
	if arr == nil || arr.State == nil || arr.Handle == 0 || arr.TrackedHandle != arr.Handle {
		return nil, false
	}
	return arr.State, true
}

func bytecodeSyncUnaliasedTrackedArrayWrite(arr *runtime.ArrayValue, state *runtime.ArrayState, idx int, value runtime.Value) bool {
	if arr == nil || state == nil || arr.TrackedAliases || arr.Handle == 0 || arr.TrackedHandle != arr.Handle {
		return false
	}
	updateArrayElementTypeTokenForWrite(state, idx, value)
	arrayStateWriteKeepsMaterializedValues(state, value)
	arr.State = state
	arr.Elements = state.Values
	return true
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
	if op == ast.AssignmentAssign {
		result, err := vm.resolveDirectArrayIndexSetAt(arr, idx, value)
		return result, true, err
	}
	if !isCompound {
		return nil, true, fmt.Errorf("unsupported assignment operator %s", op)
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
	current := state.Values[idx]
	computed, err := applyBinaryOperator(vm.interp, binaryOp, current, value)
	if err != nil {
		return nil, true, err
	}
	state.Values[idx] = computed
	if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, idx, computed) {
		vm.interp.syncTrackedArrayWrite(arr, state, idx, computed)
	}
	return computed, true, nil
}

func (vm *bytecodeVM) resolveDirectArrayIndexSetAt(arr *runtime.ArrayValue, idx int, value runtime.Value) (runtime.Value, error) {
	if vm == nil || vm.interp == nil || arr == nil {
		return nil, nil
	}
	value = vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, value)
	state, tracked := bytecodeTrackedArrayState(arr)
	if tracked {
		if idx < 0 || idx >= len(state.Values) {
			return vm.interp.makeIndexErrorValue(idx, len(state.Values)), nil
		}
		state.Values[idx] = value
		if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, idx, value) {
			vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
		}
		return value, nil
	}
	if arr.State == nil {
		handle, ok, err := vm.arrayHandleFast(arr)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil
		}
		size, err := runtime.ArrayStoreSize(handle)
		if err != nil {
			return nil, err
		}
		if idx < 0 || idx >= size {
			return vm.interp.makeIndexErrorValue(idx, size), nil
		}
		storedValue := vm.materializePrimitiveValue(bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonCollection, value)
		if err := runtime.ArrayStoreWrite(handle, idx, storedValue); err != nil {
			return nil, err
		}
		vm.interp.syncArrayHandleWriteAfterStore(handle, idx, storedValue)
		return storedValue, nil
	}
	state, err := vm.interp.ensureArrayState(arr, 0)
	if err != nil {
		return nil, err
	}
	if idx < 0 || idx >= len(state.Values) {
		return vm.interp.makeIndexErrorValue(idx, len(state.Values)), nil
	}
	state.Values[idx] = value
	if !bytecodeSyncUnaliasedTrackedArrayWrite(arr, state, idx, value) {
		vm.interp.syncTrackedArrayWrite(arr, state, idx, value)
	}
	return value, nil
}
