package interpreter

import "able/interpreter-go/pkg/runtime"

const bytecodeArraySlotCallHotEntries = 4

type bytecodeArraySlotCallCacheEntry struct {
	env                *runtime.Environment
	envVersion         uint64
	globalRevision     uint64
	methodCacheVersion uint64
	fastPath           bytecodeMemberMethodFastPathKind
}

type bytecodeInlineArraySlotCallCacheEntry struct {
	valid              bool
	program            *bytecodeProgram
	ip                 int
	env                *runtime.Environment
	envVersion         uint64
	globalRevision     uint64
	methodCacheVersion uint64
	fastPath           bytecodeMemberMethodFastPathKind
}

func bytecodeArraySlotCallShape(name string, argCount int) bool {
	return (name == "read_slot" && argCount == 1) ||
		(name == "write_slot" && argCount == 2)
}

func bytecodeArraySlotCallFastPathForInstruction(instr bytecodeInstruction) (bytecodeMemberMethodFastPathKind, bool) {
	if instr.safe {
		return bytecodeMemberMethodFastPathNone, false
	}
	switch instr.name {
	case "read_slot":
		if instr.argCount == 1 {
			return bytecodeMemberMethodFastPathArrayReadSlot, true
		}
	case "write_slot":
		if instr.argCount == 2 {
			return bytecodeMemberMethodFastPathArrayWriteSlot, true
		}
	}
	return bytecodeMemberMethodFastPathNone, false
}

func bytecodeMemberMethodFastPathIsArraySlot(kind bytecodeMemberMethodFastPathKind) bool {
	return kind == bytecodeMemberMethodFastPathArrayReadSlot ||
		kind == bytecodeMemberMethodFastPathArrayWriteSlot
}

func (vm *bytecodeVM) canUseCanonicalArraySlotCallCache(instr bytecodeInstruction, receiver runtime.Value, kind bytecodeMemberMethodFastPathKind) bool {
	if vm == nil || vm.interp == nil || vm.env == nil || !bytecodeMemberMethodFastPathIsArraySlot(kind) {
		return false
	}
	expected, ok := bytecodeArraySlotCallFastPathForInstruction(instr)
	if !ok || expected != kind {
		return false
	}
	arr, ok := receiver.(*runtime.ArrayValue)
	if !ok || arr == nil {
		return false
	}
	return vm.env.RuntimeData() == nil
}

func (vm *bytecodeVM) canUseCanonicalArraySlotCallCacheForArray(arr *runtime.ArrayValue) bool {
	return vm != nil &&
		vm.interp != nil &&
		vm.env != nil &&
		arr != nil &&
		vm.env.RuntimeData() == nil
}

func (entry bytecodeInlineArraySlotCallCacheEntry) matchesCanonicalArraySlotCallIdentity(program *bytecodeProgram, ip int, env *runtime.Environment) bool {
	return entry.valid &&
		entry.program == program &&
		entry.ip == ip &&
		entry.env == env
}

func (entry bytecodeInlineArraySlotCallCacheEntry) matchesCanonicalArraySlotCallVersions(envVersion uint64, globalRev uint64, methodVersion uint64) bool {
	return entry.envVersion == envVersion &&
		entry.globalRevision == globalRev &&
		entry.methodCacheVersion == methodVersion
}

func (vm *bytecodeVM) promoteCanonicalArraySlotCallHot(entry bytecodeInlineArraySlotCallCacheEntry) {
	if vm == nil || !entry.valid {
		return
	}
	for i := 0; i < len(vm.arraySlotCallHot); i++ {
		if vm.arraySlotCallHot[i].matchesCanonicalArraySlotCallIdentity(entry.program, entry.ip, entry.env) {
			copy(vm.arraySlotCallHot[1:i+1], vm.arraySlotCallHot[0:i])
			vm.arraySlotCallHot[0] = entry
			return
		}
	}
	copy(vm.arraySlotCallHot[1:], vm.arraySlotCallHot[:len(vm.arraySlotCallHot)-1])
	vm.arraySlotCallHot[0] = entry
}

func (vm *bytecodeVM) canonicalArraySlotCallVersions(env *runtime.Environment) (uint64, uint64, uint64) {
	return vm.bytecodeEnvRevision(env), vm.bytecodeGlobalRevision(), vm.bytecodeMethodCacheVersion()
}

func (vm *bytecodeVM) lookupCachedCanonicalArraySlotCall(program *bytecodeProgram, ip int, instr bytecodeInstruction, receiver runtime.Value) (bytecodeMemberMethodFastPathKind, bool) {
	kind, ok := bytecodeArraySlotCallFastPathForInstruction(instr)
	if !ok || program == nil || !vm.canUseCanonicalArraySlotCallCache(instr, receiver, kind) {
		return bytecodeMemberMethodFastPathNone, false
	}
	if !vm.lookupCachedCanonicalArraySlotCallForArray(program, ip, kind) {
		return bytecodeMemberMethodFastPathNone, false
	}
	return kind, true
}

func (vm *bytecodeVM) lookupCachedCanonicalArraySlotCallForArray(program *bytecodeProgram, ip int, kind bytecodeMemberMethodFastPathKind) bool {
	if program == nil || vm == nil || vm.env == nil || !bytecodeMemberMethodFastPathIsArraySlot(kind) {
		return false
	}
	env := vm.env
	var (
		envVersion    uint64
		globalRev     uint64
		methodVersion uint64
		haveVersions  bool
	)
	for i := 0; i < len(vm.arraySlotCallHot); i++ {
		hot := vm.arraySlotCallHot[i]
		if !hot.matchesCanonicalArraySlotCallIdentity(program, ip, env) || hot.fastPath != kind {
			continue
		}
		if !haveVersions {
			envVersion, globalRev, methodVersion = vm.canonicalArraySlotCallVersions(env)
			haveVersions = true
		}
		if !hot.matchesCanonicalArraySlotCallVersions(envVersion, globalRev, methodVersion) {
			return false
		}
		return true
	}
	if vm.arraySlotCallCache == nil {
		return false
	}
	if !haveVersions {
		envVersion, globalRev, methodVersion = vm.canonicalArraySlotCallVersions(env)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	entry, ok := vm.arraySlotCallCache[key]
	if !ok ||
		entry.env != env ||
		entry.envVersion != envVersion ||
		entry.globalRevision != globalRev ||
		entry.methodCacheVersion != methodVersion ||
		entry.fastPath != kind {
		return false
	}
	vm.promoteCanonicalArraySlotCallHot(bytecodeInlineArraySlotCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		env:                entry.env,
		envVersion:         entry.envVersion,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
		fastPath:           entry.fastPath,
	})
	return true
}

func (vm *bytecodeVM) storeCachedCanonicalArraySlotCall(program *bytecodeProgram, ip int, instr bytecodeInstruction, receiver runtime.Value, kind bytecodeMemberMethodFastPathKind) {
	if program == nil || !vm.canUseCanonicalArraySlotCallCache(instr, receiver, kind) {
		return
	}
	envVersion, globalRev, methodVersion := vm.canonicalArraySlotCallVersions(vm.env)
	entry := bytecodeArraySlotCallCacheEntry{
		env:                vm.env,
		envVersion:         envVersion,
		globalRevision:     globalRev,
		methodCacheVersion: methodVersion,
		fastPath:           kind,
	}
	if vm.arraySlotCallCache == nil {
		vm.arraySlotCallCache = make(map[bytecodeGlobalLookupCacheKey]bytecodeArraySlotCallCacheEntry, 8)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	vm.arraySlotCallCache[key] = entry
	vm.promoteCanonicalArraySlotCallHot(bytecodeInlineArraySlotCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		env:                entry.env,
		envVersion:         entry.envVersion,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
		fastPath:           entry.fastPath,
	})
}
