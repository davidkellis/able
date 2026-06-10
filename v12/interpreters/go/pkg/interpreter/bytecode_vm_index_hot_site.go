package interpreter

import "able/interpreter-go/pkg/runtime"

func (vm *bytecodeVM) lookupHotCanonicalArrayIndexSite(methodName string, receiver runtime.Value, fastPath bytecodeIndexMethodFastPathKind) (*runtime.ArrayValue, bool) {
	methodKind := bytecodeIndexMethodCacheKindFor(methodName)
	if methodKind == bytecodeIndexMethodCacheUnknown {
		return nil, false
	}
	return vm.lookupHotArrayIndexSiteWithFastPath(methodKind, receiver, fastPath, true)
}

func (vm *bytecodeVM) lookupHotDirectArrayIndexSite(methodName string, receiver runtime.Value) (*runtime.ArrayValue, bool) {
	methodKind := bytecodeIndexMethodCacheKindFor(methodName)
	if methodKind == bytecodeIndexMethodCacheUnknown {
		return nil, false
	}
	return vm.lookupHotArrayIndexSiteWithFastPath(methodKind, receiver, bytecodeIndexMethodFastPathNone, false)
}

func (vm *bytecodeVM) lookupHotCanonicalArrayIndexGetSite(receiver runtime.Value) (*runtime.ArrayValue, bool) {
	return vm.lookupHotArrayIndexSiteWithFastPath(bytecodeIndexMethodCacheGet, receiver, bytecodeIndexMethodFastPathCanonicalArrayGet, true)
}

func (vm *bytecodeVM) lookupHotDirectArrayIndexGetSite(receiver runtime.Value) (*runtime.ArrayValue, bool) {
	return vm.lookupHotArrayIndexSiteWithFastPath(bytecodeIndexMethodCacheGet, receiver, bytecodeIndexMethodFastPathNone, false)
}

func (vm *bytecodeVM) lookupHotCanonicalArrayIndexSetSite(receiver runtime.Value) (*runtime.ArrayValue, bool) {
	return vm.lookupHotArrayIndexSiteWithFastPath(bytecodeIndexMethodCacheSet, receiver, bytecodeIndexMethodFastPathCanonicalArraySet, true)
}

func (vm *bytecodeVM) lookupHotDirectArrayIndexSetSite(receiver runtime.Value) (*runtime.ArrayValue, bool) {
	return vm.lookupHotArrayIndexSiteWithFastPath(bytecodeIndexMethodCacheSet, receiver, bytecodeIndexMethodFastPathNone, false)
}

func (vm *bytecodeVM) lookupHotArrayIndexSiteWithFastPath(methodKind bytecodeIndexMethodCacheKind, receiver runtime.Value, fastPath bytecodeIndexMethodFastPathKind, hasMethod bool) (*runtime.ArrayValue, bool) {
	globalRevision, methodCacheVersion := vm.bytecodeGlobalAndMethodVersions()
	return vm.lookupHotArrayIndexSiteWithVersions(methodKind, receiver, fastPath, hasMethod, globalRevision, methodCacheVersion)
}

func (vm *bytecodeVM) lookupHotArrayIndexSiteWithVersions(methodKind bytecodeIndexMethodCacheKind, receiver runtime.Value, fastPath bytecodeIndexMethodFastPathKind, hasMethod bool, globalRevision uint64, methodCacheVersion uint64) (*runtime.ArrayValue, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return nil, false
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return nil, false
	}
	hot := &vm.indexMethodHot
	if hot.valid &&
		hot.program == vm.currentProgram &&
		hot.ip == vm.ip &&
		hot.methodKind == methodKind &&
		hot.hasMethod == hasMethod &&
		hot.fastPath == fastPath &&
		hot.receiverKind == bytecodeMemberReceiverArray &&
		hot.globalRevision == globalRevision &&
		hot.methodCacheVersion == methodCacheVersion {
		if vm.indexMethodHotReceiverRevisionMatches(hot, arr) {
			return arr, true
		}
		elemType, typeKey, ok := vm.arrayIndexReceiverIdentity(arr)
		if !ok || elemType != hot.arrayElemType || typeKey != hot.receiverTypeKey {
			return nil, false
		}
		vm.setIndexMethodHotArrayRevision(arr)
		return arr, true
	}
	if arr, ok := vm.lookupExactIndexMethodHotAlt(methodKind, arr, fastPath, hasMethod, globalRevision, methodCacheVersion); ok {
		return arr, true
	}
	direct := &vm.indexMethodDirect[bytecodeIndexMethodDirectIndex(vm.ip)]
	if direct.matchesExactIndexMethod(vm.currentProgram, vm.ip, methodKind, fastPath, hasMethod, globalRevision, methodCacheVersion) {
		if indexMethodInlineReceiverRevisionMatchesReady(direct, arr) {
			vm.promoteIndexMethodHot(direct)
			return arr, true
		}
		if vm.lookupInlineArrayIndexSiteIntoHot(direct, arr) {
			return arr, true
		}
	}
	return vm.lookupCachedArrayIndexSiteEntry(methodKind, arr, fastPath, hasMethod, globalRevision, methodCacheVersion)
}

func (vm *bytecodeVM) lookupDirectCompatibleHotArrayIndexSiteWithVersions(methodKind bytecodeIndexMethodCacheKind, receiver runtime.Value, canonicalFastPath bytecodeIndexMethodFastPathKind, globalRevision uint64, methodCacheVersion uint64) (*runtime.ArrayValue, int64, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return nil, 0, false
	}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return nil, 0, false
	}
	return vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersions(methodKind, arr, canonicalFastPath, globalRevision, methodCacheVersion)
}

func (vm *bytecodeVM) lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersions(methodKind bytecodeIndexMethodCacheKind, arr *runtime.ArrayValue, canonicalFastPath bytecodeIndexMethodFastPathKind, globalRevision uint64, methodCacheVersion uint64) (*runtime.ArrayValue, int64, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || arr == nil {
		return nil, 0, false
	}
	return vm.lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersionsReady(methodKind, arr, canonicalFastPath, globalRevision, methodCacheVersion)
}

func (vm *bytecodeVM) lookupDirectCompatibleHotArrayIndexSiteForArrayWithVersionsReady(methodKind bytecodeIndexMethodCacheKind, arr *runtime.ArrayValue, canonicalFastPath bytecodeIndexMethodFastPathKind, globalRevision uint64, methodCacheVersion uint64) (*runtime.ArrayValue, int64, bool) {
	hot := &vm.indexMethodHot
	if hot.matchesDirectCompatibleIndexMethod(vm.currentProgram, vm.ip, methodKind, canonicalFastPath, globalRevision, methodCacheVersion) {
		if indexMethodInlineReceiverRevisionMatchesReady(hot, arr) {
			return arr, hot.receiverArrayHandle, true
		}
		elemType, typeKey, ok := vm.arrayIndexReceiverIdentity(arr)
		if !ok || elemType != hot.arrayElemType || typeKey != hot.receiverTypeKey {
			return nil, 0, false
		}
		setIndexMethodInlineArrayRevision(hot, arr)
		handle := int64(0)
		if hot.receiverArrayRevOK {
			handle = hot.receiverArrayHandle
		}
		return arr, handle, true
	}
	if arr, handle, ok := vm.lookupDirectCompatibleIndexMethodHotAlt(methodKind, arr, canonicalFastPath, globalRevision, methodCacheVersion); ok {
		return arr, handle, true
	}
	direct := &vm.indexMethodDirect[bytecodeIndexMethodDirectIndex(vm.ip)]
	if direct.matchesDirectCompatibleIndexMethod(vm.currentProgram, vm.ip, methodKind, canonicalFastPath, globalRevision, methodCacheVersion) {
		if indexMethodInlineReceiverRevisionMatchesReady(direct, arr) {
			vm.promoteIndexMethodHot(direct)
			return arr, direct.receiverArrayHandle, true
		}
		if vm.lookupInlineArrayIndexSiteIntoHot(direct, arr) {
			return arr, vm.indexMethodHot.receiverArrayHandle, true
		}
	}
	return vm.lookupDirectCompatibleCachedArrayIndexSiteEntry(methodKind, arr, canonicalFastPath, globalRevision, methodCacheVersion)
}

func (vm *bytecodeVM) lookupInlineArrayIndexSiteIntoHot(entry *bytecodeInlineIndexMethodCacheEntry, arr *runtime.ArrayValue) bool {
	handle, ok := vm.lookupInlineArrayIndexSiteReady(entry, arr)
	if !ok {
		return false
	}
	vm.promoteIndexMethodHot(entry)
	vm.indexMethodHot.receiverArrayHandle = handle
	return true
}

func (vm *bytecodeVM) lookupExactIndexMethodHotAlt(methodKind bytecodeIndexMethodCacheKind, arr *runtime.ArrayValue, fastPath bytecodeIndexMethodFastPathKind, hasMethod bool, globalRevision uint64, methodCacheVersion uint64) (*runtime.ArrayValue, bool) {
	alt := &vm.indexMethodHotAlt
	if !alt.matchesExactIndexMethod(vm.currentProgram, vm.ip, methodKind, fastPath, hasMethod, globalRevision, methodCacheVersion) {
		return nil, false
	}
	if indexMethodInlineReceiverRevisionMatchesReady(alt, arr) {
		return arr, true
	}
	elemType, typeKey, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != alt.arrayElemType || typeKey != alt.receiverTypeKey {
		return nil, false
	}
	setIndexMethodInlineArrayRevision(alt, arr)
	return arr, true
}

func (vm *bytecodeVM) lookupDirectCompatibleIndexMethodHotAlt(methodKind bytecodeIndexMethodCacheKind, arr *runtime.ArrayValue, canonicalFastPath bytecodeIndexMethodFastPathKind, globalRevision uint64, methodCacheVersion uint64) (*runtime.ArrayValue, int64, bool) {
	alt := &vm.indexMethodHotAlt
	if !alt.matchesDirectCompatibleIndexMethod(vm.currentProgram, vm.ip, methodKind, canonicalFastPath, globalRevision, methodCacheVersion) {
		return nil, 0, false
	}
	if indexMethodInlineReceiverRevisionMatchesReady(alt, arr) {
		return arr, alt.receiverArrayHandle, true
	}
	elemType, typeKey, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != alt.arrayElemType || typeKey != alt.receiverTypeKey {
		return nil, 0, false
	}
	setIndexMethodInlineArrayRevision(alt, arr)
	handle := int64(0)
	if alt.receiverArrayRevOK {
		handle = alt.receiverArrayHandle
	}
	return arr, handle, true
}

func (vm *bytecodeVM) promoteIndexMethodHot(entry *bytecodeInlineIndexMethodCacheEntry) {
	if vm == nil || entry == nil || !entry.valid {
		return
	}
	hot := vm.indexMethodHot
	if hot.valid &&
		(hot.program != entry.program || hot.ip != entry.ip || hot.methodKind != entry.methodKind) {
		vm.indexMethodHotAlt = hot
	}
	vm.indexMethodHot = *entry
}

func (vm *bytecodeVM) lookupInlineArrayIndexSiteReady(entry *bytecodeInlineIndexMethodCacheEntry, arr *runtime.ArrayValue) (int64, bool) {
	if entry == nil || arr == nil {
		return 0, false
	}
	if indexMethodInlineReceiverRevisionMatchesReady(entry, arr) {
		return entry.receiverArrayHandle, true
	}
	elemType, typeKey, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != entry.arrayElemType || typeKey != entry.receiverTypeKey {
		return 0, false
	}
	setIndexMethodInlineArrayRevision(entry, arr)
	if entry.receiverArrayRevOK {
		return entry.receiverArrayHandle, true
	}
	return 0, true
}

func (vm *bytecodeVM) lookupDirectCompatibleCachedArrayIndexSiteEntry(methodKind bytecodeIndexMethodCacheKind, arr *runtime.ArrayValue, canonicalFastPath bytecodeIndexMethodFastPathKind, globalRevision uint64, methodCacheVersion uint64) (*runtime.ArrayValue, int64, bool) {
	ip := vm.ip
	var entry *bytecodeIndexMethodCacheEntry
	if ip >= 0 && vm.activeLookup.program == vm.currentProgram {
		entry, _ = vm.indexMethodCacheEntryForKind(vm.currentProgram, ip, methodKind, false)
	} else {
		entry = vm.inactiveIndexMethodEntry(methodKind, ip)
	}
	if entry == nil ||
		entry.globalRevision != globalRevision ||
		entry.methodCacheVersion != methodCacheVersion ||
		entry.receiverKind != bytecodeMemberReceiverArray ||
		!bytecodeIndexMethodDirectArrayCompatible(entry.hasMethod, entry.fastPath, canonicalFastPath) {
		return nil, 0, false
	}
	if indexMethodEntryReceiverRevisionMatchesReady(entry, arr) {
		return arr, entry.receiverArrayHandle, true
	}
	elemType, typeKey, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != entry.arrayElemType || typeKey != entry.receiverTypeKey {
		return nil, 0, false
	}
	vm.setIndexMethodEntryArrayRevision(entry, arr)
	vm.setIndexMethodHotFromEntry(entry, methodKind)
	handle := int64(0)
	if entry.receiverArrayRevOK {
		handle = entry.receiverArrayHandle
	}
	return arr, handle, true
}

func bytecodeIndexMethodDirectArrayCompatible(hasMethod bool, fastPath bytecodeIndexMethodFastPathKind, canonicalFastPath bytecodeIndexMethodFastPathKind) bool {
	if hasMethod {
		return fastPath == canonicalFastPath
	}
	return fastPath == bytecodeIndexMethodFastPathNone
}

func bytecodeIndexMethodDirectIndex(ip int) int {
	return int(uint(ip) & uint(bytecodeIndexMethodDirectEntries-1))
}

func (entry *bytecodeInlineIndexMethodCacheEntry) matchesExactIndexMethod(program *bytecodeProgram, ip int, methodKind bytecodeIndexMethodCacheKind, fastPath bytecodeIndexMethodFastPathKind, hasMethod bool, globalRevision uint64, methodCacheVersion uint64) bool {
	return entry != nil &&
		entry.valid &&
		entry.program == program &&
		entry.ip == ip &&
		entry.methodKind == methodKind &&
		entry.hasMethod == hasMethod &&
		entry.fastPath == fastPath &&
		entry.receiverKind == bytecodeMemberReceiverArray &&
		entry.globalRevision == globalRevision &&
		entry.methodCacheVersion == methodCacheVersion
}

func (entry *bytecodeInlineIndexMethodCacheEntry) matchesDirectCompatibleIndexMethod(program *bytecodeProgram, ip int, methodKind bytecodeIndexMethodCacheKind, canonicalFastPath bytecodeIndexMethodFastPathKind, globalRevision uint64, methodCacheVersion uint64) bool {
	return entry != nil &&
		entry.valid &&
		entry.program == program &&
		entry.ip == ip &&
		entry.methodKind == methodKind &&
		entry.receiverKind == bytecodeMemberReceiverArray &&
		entry.globalRevision == globalRevision &&
		entry.methodCacheVersion == methodCacheVersion &&
		bytecodeIndexMethodDirectArrayCompatible(entry.hasMethod, entry.fastPath, canonicalFastPath)
}

func (vm *bytecodeVM) storeIndexMethodDirect(entry bytecodeInlineIndexMethodCacheEntry) {
	if vm == nil || !entry.valid {
		return
	}
	vm.indexMethodDirect[bytecodeIndexMethodDirectIndex(entry.ip)] = entry
}

func (vm *bytecodeVM) lookupCachedArrayIndexSiteEntry(methodKind bytecodeIndexMethodCacheKind, arr *runtime.ArrayValue, fastPath bytecodeIndexMethodFastPathKind, hasMethod bool, globalRevision uint64, methodCacheVersion uint64) (*runtime.ArrayValue, bool) {
	ip := vm.ip
	var entry *bytecodeIndexMethodCacheEntry
	if ip >= 0 && vm.activeLookup.program == vm.currentProgram {
		entry, _ = vm.indexMethodCacheEntryForKind(vm.currentProgram, ip, methodKind, false)
	} else {
		entry = vm.inactiveIndexMethodEntry(methodKind, ip)
	}
	if entry == nil ||
		entry.globalRevision != globalRevision ||
		entry.methodCacheVersion != methodCacheVersion ||
		entry.receiverKind != bytecodeMemberReceiverArray ||
		entry.hasMethod != hasMethod ||
		entry.fastPath != fastPath {
		return nil, false
	}
	if vm.indexMethodEntryReceiverRevisionMatches(entry, arr) {
		return arr, true
	}
	elemType, typeKey, ok := vm.arrayIndexReceiverIdentity(arr)
	if !ok || elemType != entry.arrayElemType || typeKey != entry.receiverTypeKey {
		return nil, false
	}
	vm.setIndexMethodEntryArrayRevision(entry, arr)
	vm.setIndexMethodHotFromEntry(entry, methodKind)
	return arr, true
}

func (vm *bytecodeVM) inactiveIndexMethodEntry(methodKind bytecodeIndexMethodCacheKind, ip int) *bytecodeIndexMethodCacheEntry {
	if ip < 0 {
		return nil
	}
	entry, ok := vm.indexMethodCacheEntryForKind(vm.currentProgram, ip, methodKind, false)
	if !ok {
		return nil
	}
	return entry
}

func (vm *bytecodeVM) setIndexMethodHotFromEntry(entry *bytecodeIndexMethodCacheEntry, methodKind bytecodeIndexMethodCacheKind) {
	if vm == nil || entry == nil {
		return
	}
	hot := &vm.indexMethodHot
	hot.valid = true
	hot.program = vm.currentProgram
	hot.ip = vm.ip
	hot.methodKind = methodKind
	hot.globalRevision = entry.globalRevision
	hot.receiverKind = entry.receiverKind
	hot.arrayElemType = entry.arrayElemType
	hot.receiverTypeKey = entry.receiverTypeKey
	hot.receiverArrayHandle = entry.receiverArrayHandle
	hot.receiverArrayRev = entry.receiverArrayRev
	hot.receiverArrayRevOK = entry.receiverArrayRevOK
	hot.receiverArrayCursor = entry.receiverArrayCursor
	hot.methodCacheVersion = entry.methodCacheVersion
	hot.resolvedMethod = entry.method
	hot.hasMethod = entry.hasMethod
	hot.fastPath = entry.fastPath
	vm.storeIndexMethodDirect(*hot)
}

func (vm *bytecodeVM) setIndexMethodHotReceiverRevision(receiver runtime.Value) {
	if vm == nil {
		return
	}
	hot := &vm.indexMethodHot
	if !hot.valid {
		return
	}
	hot.receiverArrayHandle = 0
	hot.receiverArrayRev = 0
	hot.receiverArrayRevOK = false
	hot.receiverArrayCursor = runtime.ArrayStoreRevisionCursor{}
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return
	}
	vm.setIndexMethodHotArrayRevision(arr)
}

func (vm *bytecodeVM) setIndexMethodHotArrayRevision(arr *runtime.ArrayValue) {
	if vm == nil {
		return
	}
	hot := &vm.indexMethodHot
	if !hot.valid {
		return
	}
	setIndexMethodInlineArrayRevision(hot, arr)
}

func setIndexMethodInlineArrayRevision(entry *bytecodeInlineIndexMethodCacheEntry, arr *runtime.ArrayValue) {
	if entry == nil {
		return
	}
	handle, revision, cursor, ok := indexMethodArrayReceiverRevisionWithCursor(arr)
	entry.receiverArrayHandle = 0
	entry.receiverArrayRev = 0
	entry.receiverArrayRevOK = false
	entry.receiverArrayCursor = runtime.ArrayStoreRevisionCursor{}
	if !ok {
		return
	}
	entry.receiverArrayHandle = handle
	entry.receiverArrayRev = revision
	entry.receiverArrayRevOK = true
	entry.receiverArrayCursor = cursor
}

func (vm *bytecodeVM) setIndexMethodEntryArrayRevision(entry *bytecodeIndexMethodCacheEntry, arr *runtime.ArrayValue) {
	if entry == nil {
		return
	}
	handle, revision, cursor, ok := indexMethodArrayReceiverRevisionWithCursor(arr)
	entry.receiverArrayHandle = 0
	entry.receiverArrayRev = 0
	entry.receiverArrayRevOK = false
	entry.receiverArrayCursor = runtime.ArrayStoreRevisionCursor{}
	if !ok {
		return
	}
	entry.receiverArrayHandle = handle
	entry.receiverArrayRev = revision
	entry.receiverArrayRevOK = true
	entry.receiverArrayCursor = cursor
}

func (vm *bytecodeVM) indexMethodHotReceiverRevisionMatches(hot *bytecodeInlineIndexMethodCacheEntry, arr *runtime.ArrayValue) bool {
	if hot == nil || arr == nil || !hot.receiverArrayRevOK || hot.receiverArrayHandle == 0 {
		return false
	}
	return indexMethodInlineReceiverRevisionMatchesReady(hot, arr)
}

func (vm *bytecodeVM) indexMethodEntryReceiverRevisionMatches(entry *bytecodeIndexMethodCacheEntry, arr *runtime.ArrayValue) bool {
	if entry == nil || arr == nil || !entry.receiverArrayRevOK || entry.receiverArrayHandle == 0 {
		return false
	}
	return indexMethodEntryReceiverRevisionMatchesReady(entry, arr)
}

func indexMethodInlineReceiverRevisionMatchesReady(entry *bytecodeInlineIndexMethodCacheEntry, arr *runtime.ArrayValue) bool {
	if !entry.receiverArrayRevOK || entry.receiverArrayHandle == 0 {
		return false
	}
	return indexMethodArrayReceiverRevisionMatchesWithCursor(arr, entry.receiverArrayHandle, entry.receiverArrayRev, entry.receiverArrayCursor)
}

func indexMethodEntryReceiverRevisionMatchesReady(entry *bytecodeIndexMethodCacheEntry, arr *runtime.ArrayValue) bool {
	if !entry.receiverArrayRevOK || entry.receiverArrayHandle == 0 {
		return false
	}
	return indexMethodArrayReceiverRevisionMatchesWithCursor(arr, entry.receiverArrayHandle, entry.receiverArrayRev, entry.receiverArrayCursor)
}

func (vm *bytecodeVM) indexMethodReceiverRevision(receiver runtime.Value) (int64, uint64, bool) {
	handle, revision, _, ok := vm.indexMethodReceiverRevisionWithCursor(receiver)
	return handle, revision, ok
}

func (vm *bytecodeVM) indexMethodReceiverRevisionWithCursor(receiver runtime.Value) (int64, uint64, runtime.ArrayStoreRevisionCursor, bool) {
	arr, ok := bytecodeArrayReceiverForIndexCache(receiver)
	if !ok {
		return 0, 0, runtime.ArrayStoreRevisionCursor{}, false
	}
	return indexMethodArrayReceiverRevisionWithCursor(arr)
}

func indexMethodArrayReceiverRevision(arr *runtime.ArrayValue) (int64, uint64, bool) {
	handle, revision, _, ok := indexMethodArrayReceiverRevisionWithCursor(arr)
	return handle, revision, ok
}

func indexMethodArrayReceiverRevisionWithCursor(arr *runtime.ArrayValue) (int64, uint64, runtime.ArrayStoreRevisionCursor, bool) {
	handle := bytecodeArrayStorageHandle(arr)
	if handle == 0 {
		return 0, 0, runtime.ArrayStoreRevisionCursor{}, false
	}
	if state, tracked := bytecodeTrackedArrayState(arr); tracked {
		cursor, revision, ok, err := runtime.ArrayStoreRevisionCursorIfAvailable(handle)
		if err == nil && ok {
			return handle, revision, cursor, true
		}
		return handle, state.Revision, runtime.ArrayStoreRevisionCursor{}, true
	}
	cursor, revision, ok, err := runtime.ArrayStoreRevisionCursorIfAvailable(handle)
	if err != nil || !ok {
		return 0, 0, runtime.ArrayStoreRevisionCursor{}, false
	}
	return handle, revision, cursor, true
}

func indexMethodArrayReceiverRevisionMatches(arr *runtime.ArrayValue, expectedHandle int64, expectedRevision uint64) bool {
	return indexMethodArrayReceiverRevisionMatchesWithCursor(arr, expectedHandle, expectedRevision, runtime.ArrayStoreRevisionCursor{})
}

func indexMethodArrayReceiverRevisionMatchesWithCursor(arr *runtime.ArrayValue, expectedHandle int64, expectedRevision uint64, cursor runtime.ArrayStoreRevisionCursor) bool {
	if arr == nil || expectedHandle == 0 {
		return false
	}
	if handle := arr.Handle; handle == expectedHandle {
		if arr.State != nil && arr.TrackedHandle == handle {
			return arr.State.Revision == expectedRevision
		}
		return indexMethodArrayStoreRevisionMatchesWithCursor(handle, expectedRevision, cursor)
	} else if handle != 0 || arr.TrackedHandle != expectedHandle {
		return false
	}
	return indexMethodArrayStoreRevisionMatchesWithCursor(expectedHandle, expectedRevision, cursor)
}

func indexMethodArrayStoreRevisionMatchesWithCursor(handle int64, expectedRevision uint64, cursor runtime.ArrayStoreRevisionCursor) bool {
	if cursor.MatchesKnownHandle(handle, expectedRevision) {
		return true
	}
	matches, ok, err := runtime.ArrayStoreRevisionMatchesIfAvailable(handle, expectedRevision)
	return err == nil && ok && matches
}
