package interpreter

import "able/interpreter-go/pkg/runtime"

type bytecodeArrayIndexReceiverIdentityCacheEntry struct {
	revision uint64
	elemType uint16
	key      string
	ok       bool
}

func (vm *bytecodeVM) arrayIndexReceiverIdentity(arr *runtime.ArrayValue) (uint16, string, bool) {
	if vm == nil || vm.interp == nil || arr == nil {
		return bytecodeIndexTypeUnknown, "", false
	}
	handle := bytecodeArrayStorageHandle(arr)
	if handle == 0 {
		return vm.computeArrayIndexReceiverIdentity(arr)
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if state.ElementTypeTokenKnown && state.ElementTypeToken != bytecodeIndexTypeUnknown {
			return state.ElementTypeToken, "", true
		}
		revision := state.Revision
		if cached, ok := vm.lookupArrayIndexReceiverIdentityCache(handle, revision); ok {
			return cached.elemType, cached.key, cached.ok
		}
		elemType, key, identityOK := vm.computeArrayIndexReceiverIdentity(arr)
		vm.storeArrayIndexReceiverIdentityCache(handle, revision, elemType, key, identityOK)
		return elemType, key, identityOK
	}
	revision, ok, err := runtime.ArrayStoreRevisionIfAvailable(handle)
	if err != nil || !ok {
		if elemType, ok := vm.arrayIndexReceiverIdentityFromMonoHandle(handle); ok {
			return elemType, "", true
		}
		return vm.computeArrayIndexReceiverIdentity(arr)
	}
	if elemType, ok := vm.lookupArrayIndexReceiverMonoTokenHot(handle, revision); ok {
		return elemType, "", true
	}
	if elemType, ok := vm.arrayIndexReceiverIdentityFromMonoHandle(handle); ok {
		vm.storeArrayIndexReceiverMonoTokenHot(handle, revision, elemType)
		return elemType, "", true
	}
	if cached, ok := vm.lookupArrayIndexReceiverIdentityCache(handle, revision); ok {
		return cached.elemType, cached.key, cached.ok
	}
	elemType, key, identityOK := vm.computeArrayIndexReceiverIdentity(arr)
	vm.storeArrayIndexReceiverIdentityCache(handle, revision, elemType, key, identityOK)
	return elemType, key, identityOK
}

func (vm *bytecodeVM) arrayIndexReceiverIdentityFromMonoHandle(handle int64) (uint16, bool) {
	if handle == 0 {
		return bytecodeIndexTypeUnknown, false
	}
	typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(handle)
	if err != nil || !ok {
		return bytecodeIndexTypeUnknown, false
	}
	elemType, ok := bytecodeIndexTypeTokenFromTypeName(typeName)
	return elemType, ok && elemType != bytecodeIndexTypeUnknown
}

func (vm *bytecodeVM) lookupArrayIndexReceiverMonoTokenHot(handle int64, revision uint64) (uint16, bool) {
	if vm == nil || handle == 0 || !vm.arrayIndexReceiverMonoTokenHotOK {
		return bytecodeIndexTypeUnknown, false
	}
	if vm.arrayIndexReceiverMonoTokenHotHandle != handle || vm.arrayIndexReceiverMonoTokenHotRevision != revision {
		return bytecodeIndexTypeUnknown, false
	}
	return vm.arrayIndexReceiverMonoTokenHot, true
}

func (vm *bytecodeVM) storeArrayIndexReceiverMonoTokenHot(handle int64, revision uint64, elemType uint16) {
	if vm == nil || handle == 0 || elemType == bytecodeIndexTypeUnknown {
		return
	}
	vm.arrayIndexReceiverMonoTokenHotHandle = handle
	vm.arrayIndexReceiverMonoTokenHotRevision = revision
	vm.arrayIndexReceiverMonoTokenHot = elemType
	vm.arrayIndexReceiverMonoTokenHotOK = true
}

func (vm *bytecodeVM) arrayElementTypeTokenForPropagation(arr *runtime.ArrayValue) (uint16, bool) {
	if arr == nil {
		return bytecodeIndexTypeUnknown, false
	}
	if arr.State != nil && arr.State.ElementTypeTokenKnown {
		return arr.State.ElementTypeToken, true
	}
	handle := bytecodeArrayStorageHandle(arr)
	if handle != 0 {
		revision, revOK, err := runtime.ArrayStoreRevisionIfAvailable(handle)
		if err == nil && revOK {
			if elemType, ok := vm.lookupArrayIndexReceiverMonoTokenHot(handle, revision); ok {
				return elemType, true
			}
			if elemType, ok := vm.arrayIndexReceiverIdentityFromMonoHandle(handle); ok {
				vm.storeArrayIndexReceiverMonoTokenHot(handle, revision, elemType)
				return elemType, true
			}
		} else if elemType, ok := vm.arrayIndexReceiverIdentityFromMonoHandle(handle); ok {
			return elemType, true
		}
	}
	return bytecodeArrayElementTypeTokenFromValues(arr.Elements)
}

func (vm *bytecodeVM) arrayIndexMethodReceiverTypeKey(arr *runtime.ArrayValue) (string, bool) {
	_, key, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || key == "" {
		return "", false
	}
	return key, true
}

func (vm *bytecodeVM) lookupArrayIndexReceiverIdentityCache(handle int64, revision uint64) (bytecodeArrayIndexReceiverIdentityCacheEntry, bool) {
	if vm == nil || handle == 0 {
		return bytecodeArrayIndexReceiverIdentityCacheEntry{}, false
	}
	if vm.arrayIndexReceiverIdentityHotHandle == handle && vm.arrayIndexReceiverIdentityHot.revision == revision {
		return vm.arrayIndexReceiverIdentityHot, true
	}
	if vm.arrayIndexReceiverIdentityCache == nil {
		return bytecodeArrayIndexReceiverIdentityCacheEntry{}, false
	}
	entry, ok := vm.arrayIndexReceiverIdentityCache[handle]
	if !ok || entry.revision != revision {
		return bytecodeArrayIndexReceiverIdentityCacheEntry{}, false
	}
	vm.arrayIndexReceiverIdentityHotHandle = handle
	vm.arrayIndexReceiverIdentityHot = entry
	return entry, true
}

func (vm *bytecodeVM) storeArrayIndexReceiverIdentityCache(handle int64, revision uint64, elemType uint16, key string, ok bool) {
	if vm == nil || handle == 0 {
		return
	}
	if vm.arrayIndexReceiverIdentityCache == nil {
		vm.arrayIndexReceiverIdentityCache = make(map[int64]bytecodeArrayIndexReceiverIdentityCacheEntry, 8)
	}
	entry := bytecodeArrayIndexReceiverIdentityCacheEntry{
		revision: revision,
		elemType: elemType,
		key:      key,
		ok:       ok,
	}
	vm.arrayIndexReceiverIdentityCache[handle] = entry
	vm.arrayIndexReceiverIdentityHotHandle = handle
	vm.arrayIndexReceiverIdentityHot = entry
}

func (vm *bytecodeVM) computeArrayIndexReceiverIdentity(arr *runtime.ArrayValue) (uint16, string, bool) {
	if vm == nil || vm.interp == nil || arr == nil {
		return bytecodeIndexTypeUnknown, "", false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		if elemType, key, ok := vm.arrayIndexReceiverIdentityFromValues(state.ElementTypeToken, state.ElementTypeTokenKnown, state.Values); ok {
			return elemType, key, true
		}
	}
	if arr.State != nil {
		if elemType, key, ok := vm.arrayIndexReceiverIdentityFromValues(arr.State.ElementTypeToken, arr.State.ElementTypeTokenKnown, arr.State.Values); ok {
			return elemType, key, true
		}
	}
	handle := bytecodeArrayStorageHandle(arr)
	if handle != 0 {
		if typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(handle); err == nil && ok {
			if elemType, ok := bytecodeIndexTypeTokenFromTypeName(typeName); ok && elemType != bytecodeIndexTypeUnknown {
				return elemType, "", true
			}
		}
		size, err := runtime.ArrayStoreSize(handle)
		if err == nil && size > 0 {
			value, err := runtime.ArrayStoreRead(handle, 0)
			if err == nil {
				return vm.arrayIndexReceiverIdentityFromValue(value)
			}
		}
	}
	return vm.arrayIndexReceiverIdentityFromValues(bytecodeIndexTypeUnknown, false, arr.Elements)
}

func (vm *bytecodeVM) arrayIndexReceiverIdentityFromValues(elemType uint16, elemTypeKnown bool, values []runtime.Value) (uint16, string, bool) {
	if elemTypeKnown && elemType != bytecodeIndexTypeUnknown {
		return elemType, "", true
	}
	if len(values) == 0 {
		return bytecodeIndexTypeUnknown, "", false
	}
	return vm.arrayIndexReceiverIdentityFromValue(values[0])
}

func (vm *bytecodeVM) arrayIndexReceiverIdentityFromValue(value runtime.Value) (uint16, string, bool) {
	if elemType, ok := bytecodeIndexValueTypeToken(value); ok && elemType != bytecodeIndexTypeUnknown {
		return elemType, "", true
	}
	key, ok := vm.arrayIndexMethodReceiverTypeKeyFromValue(value)
	if !ok {
		return bytecodeIndexTypeUnknown, "", false
	}
	return bytecodeIndexTypeUnknown, key, true
}

func (vm *bytecodeVM) arrayIndexMethodReceiverTypeKeyFromValues(values []runtime.Value) (string, bool) {
	if len(values) == 0 {
		return "", false
	}
	return vm.arrayIndexMethodReceiverTypeKeyFromValue(values[0])
}

func (vm *bytecodeVM) arrayIndexMethodReceiverTypeKeyFromValue(value runtime.Value) (string, bool) {
	if vm == nil || vm.interp == nil || value == nil {
		return "", false
	}
	expr := vm.interp.typeExpressionFromValue(value)
	if expr == nil {
		return "", false
	}
	key := typeExpressionToString(expr)
	return key, key != "" && key != "<?>"
}
