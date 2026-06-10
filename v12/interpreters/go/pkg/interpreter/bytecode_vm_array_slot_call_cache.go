package interpreter

import "able/interpreter-go/pkg/runtime"

const (
	bytecodeArraySlotCallHotEntries    = 8
	bytecodeArraySlotCallDirectEntries = 16
)

type bytecodeArraySlotCallCacheEntry struct {
	globalRevision     uint64
	methodCacheVersion uint64
	fastPath           bytecodeMemberMethodFastPathKind
}

type bytecodeInlineArraySlotCallCacheEntry struct {
	valid              bool
	program            *bytecodeProgram
	ip                 int
	globalRevision     uint64
	methodCacheVersion uint64
	fastPath           bytecodeMemberMethodFastPathKind
}

func bytecodeArraySlotCallShape(name string, argCount int) bool {
	_, ok := bytecodeArraySlotCallFastPathForName(name, argCount)
	return ok
}

func bytecodeArraySlotCallFastPathForName(name string, argCount int) (bytecodeMemberMethodFastPathKind, bool) {
	switch name {
	case "len":
		if argCount == 0 {
			return bytecodeMemberMethodFastPathArrayLen, true
		}
	case "read_slot":
		if argCount == 1 {
			return bytecodeMemberMethodFastPathArrayReadSlot, true
		}
	case "write_slot":
		if argCount == 2 {
			return bytecodeMemberMethodFastPathArrayWriteSlot, true
		}
	case "push":
		if argCount == 1 {
			return bytecodeMemberMethodFastPathArrayPush, true
		}
	}
	return bytecodeMemberMethodFastPathNone, false
}

func bytecodeArraySlotCallFastPathForInstruction(instr *bytecodeInstruction) (bytecodeMemberMethodFastPathKind, bool) {
	if instr == nil || instr.safe {
		return bytecodeMemberMethodFastPathNone, false
	}
	switch instr.memberFastPath {
	case bytecodeMemberMethodFastPathArrayLen, bytecodeMemberMethodFastPathArrayReadSlot, bytecodeMemberMethodFastPathArrayWriteSlot, bytecodeMemberMethodFastPathArrayPush:
		return instr.memberFastPath, true
	}
	return bytecodeArraySlotCallFastPathForName(instr.name, instr.argCount)
}

func bytecodeMemberMethodFastPathIsArraySlot(kind bytecodeMemberMethodFastPathKind) bool {
	return kind == bytecodeMemberMethodFastPathArrayLen ||
		kind == bytecodeMemberMethodFastPathArrayReadSlot ||
		kind == bytecodeMemberMethodFastPathArrayWriteSlot ||
		kind == bytecodeMemberMethodFastPathArrayReadWriteSlot ||
		kind == bytecodeMemberMethodFastPathArrayPush
}

func (vm *bytecodeVM) canUseCanonicalArraySlotCallCache(instr bytecodeInstruction, receiver runtime.Value, kind bytecodeMemberMethodFastPathKind) bool {
	if vm == nil || vm.interp == nil || vm.env == nil || !bytecodeMemberMethodFastPathIsArraySlot(kind) {
		return false
	}
	expected, ok := bytecodeArraySlotCallFastPathForInstruction(&instr)
	if !ok || expected != kind {
		return false
	}
	arr, ok := receiver.(*runtime.ArrayValue)
	if !ok || arr == nil {
		return false
	}
	return !vm.hasRuntimeData()
}

func (vm *bytecodeVM) canUseCanonicalArraySlotCallCacheForArray(arr *runtime.ArrayValue) bool {
	return vm != nil &&
		vm.interp != nil &&
		arr != nil &&
		!vm.hasRuntimeData()
}

func (vm *bytecodeVM) promoteCanonicalArraySlotCallHot(entry bytecodeInlineArraySlotCallCacheEntry) {
	if vm == nil || !entry.valid {
		return
	}
	for i := 0; i < len(vm.arraySlotCallHot); i++ {
		hot := &vm.arraySlotCallHot[i]
		if hot.valid &&
			hot.program == entry.program &&
			hot.ip == entry.ip {
			copy(vm.arraySlotCallHot[1:i+1], vm.arraySlotCallHot[0:i])
			vm.arraySlotCallHot[0] = entry
			return
		}
	}
	copy(vm.arraySlotCallHot[1:], vm.arraySlotCallHot[:len(vm.arraySlotCallHot)-1])
	vm.arraySlotCallHot[0] = entry
}

func bytecodeArraySlotCallDirectIndex(ip int) int {
	return int(uint(ip) & uint(bytecodeArraySlotCallDirectEntries-1))
}

func (vm *bytecodeVM) canonicalArraySlotCallVersions() (uint64, uint64) {
	return vm.bytecodeGlobalAndMethodVersions()
}

func (vm *bytecodeVM) storeCanonicalArraySlotCallDirect(entry bytecodeInlineArraySlotCallCacheEntry) {
	if vm == nil || !entry.valid {
		return
	}
	vm.arraySlotCallDirect[bytecodeArraySlotCallDirectIndex(entry.ip)] = entry
}

func (vm *bytecodeVM) lookupCachedCanonicalArraySlotCall(program *bytecodeProgram, ip int, instr bytecodeInstruction, receiver runtime.Value) (bytecodeMemberMethodFastPathKind, bool) {
	kind, ok := bytecodeArraySlotCallFastPathForInstruction(&instr)
	if !ok || program == nil || !vm.canUseCanonicalArraySlotCallCache(instr, receiver, kind) {
		return bytecodeMemberMethodFastPathNone, false
	}
	if !vm.lookupCachedCanonicalArraySlotCallForArrayValidated(program, ip, kind) {
		return bytecodeMemberMethodFastPathNone, false
	}
	return kind, true
}

func (vm *bytecodeVM) lookupCachedCanonicalArraySlotCallForArray(program *bytecodeProgram, ip int, kind bytecodeMemberMethodFastPathKind) bool {
	if program == nil || vm == nil || !bytecodeMemberMethodFastPathIsArraySlot(kind) || vm.hasRuntimeData() {
		return false
	}
	return vm.lookupCachedCanonicalArraySlotCallForArrayValidated(program, ip, kind)
}

func (vm *bytecodeVM) lookupCachedCanonicalArraySlotCallForArrayValidated(program *bytecodeProgram, ip int, kind bytecodeMemberMethodFastPathKind) bool {
	globalRev, methodVersion := vm.canonicalArraySlotCallVersions()
	return vm.lookupCachedCanonicalArraySlotCallForArrayValidatedWithVersions(program, ip, kind, globalRev, methodVersion)
}

func (vm *bytecodeVM) lookupCachedCanonicalArraySlotCallForArrayValidatedWithVersions(program *bytecodeProgram, ip int, kind bytecodeMemberMethodFastPathKind, globalRev uint64, methodVersion uint64) bool {
	direct := &vm.arraySlotCallDirect[bytecodeArraySlotCallDirectIndex(ip)]
	if direct.valid &&
		direct.program == program &&
		direct.ip == ip &&
		direct.fastPath == kind {
		if direct.globalRevision != globalRev ||
			direct.methodCacheVersion != methodVersion {
			return false
		}
		return true
	}
	for i := 0; i < len(vm.arraySlotCallHot); i++ {
		hot := &vm.arraySlotCallHot[i]
		if !hot.valid ||
			hot.program != program ||
			hot.ip != ip ||
			hot.fastPath != kind {
			continue
		}
		if hot.globalRevision != globalRev ||
			hot.methodCacheVersion != methodVersion {
			return false
		}
		vm.storeCanonicalArraySlotCallDirect(*hot)
		return true
	}
	if vm.arraySlotCallCache == nil {
		return false
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	entry, ok := vm.arraySlotCallCache[key]
	if !ok ||
		entry.globalRevision != globalRev ||
		entry.methodCacheVersion != methodVersion ||
		entry.fastPath != kind {
		return false
	}
	inlineEntry := bytecodeInlineArraySlotCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
		fastPath:           entry.fastPath,
	}
	vm.promoteCanonicalArraySlotCallHot(inlineEntry)
	vm.storeCanonicalArraySlotCallDirect(inlineEntry)
	return true
}

func (vm *bytecodeVM) storeCachedCanonicalArraySlotCall(program *bytecodeProgram, ip int, instr bytecodeInstruction, receiver runtime.Value, kind bytecodeMemberMethodFastPathKind) {
	if program == nil || !vm.canUseCanonicalArraySlotCallCache(instr, receiver, kind) {
		return
	}
	arr, _ := receiver.(*runtime.ArrayValue)
	vm.storeCachedCanonicalArraySlotCallForArray(program, ip, arr, kind)
}

func (vm *bytecodeVM) storeCachedCanonicalArraySlotCallForArray(program *bytecodeProgram, ip int, arr *runtime.ArrayValue, kind bytecodeMemberMethodFastPathKind) {
	if program == nil || !vm.canUseCanonicalArraySlotCallCacheForArray(arr) || !bytecodeMemberMethodFastPathIsArraySlot(kind) {
		return
	}
	globalRev, methodVersion := vm.canonicalArraySlotCallVersions()
	entry := bytecodeArraySlotCallCacheEntry{
		globalRevision:     globalRev,
		methodCacheVersion: methodVersion,
		fastPath:           kind,
	}
	if vm.arraySlotCallCache == nil {
		vm.arraySlotCallCache = make(map[bytecodeGlobalLookupCacheKey]bytecodeArraySlotCallCacheEntry, 8)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	vm.arraySlotCallCache[key] = entry
	inlineEntry := bytecodeInlineArraySlotCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
		fastPath:           entry.fastPath,
	}
	vm.promoteCanonicalArraySlotCallHot(inlineEntry)
	vm.storeCanonicalArraySlotCallDirect(inlineEntry)
}
