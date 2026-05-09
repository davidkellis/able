package interpreter

import (
	"able/interpreter-go/pkg/runtime"
)

const bytecodeArrayNewCallHotEntries = 4

type bytecodeArrayNewCallCacheEntry struct {
	env                *runtime.Environment
	envVersion         uint64
	globalRevision     uint64
	methodCacheVersion uint64
}

type bytecodeInlineArrayNewCallCacheEntry struct {
	valid              bool
	program            *bytecodeProgram
	ip                 int
	env                *runtime.Environment
	envVersion         uint64
	globalRevision     uint64
	methodCacheVersion uint64
}

func (vm *bytecodeVM) canUseCanonicalArrayNewCallCache(instr bytecodeInstruction, receiver runtime.Value) bool {
	if vm == nil || vm.interp == nil || vm.env == nil || instr.name != "new" || instr.argCount != 0 || instr.safe {
		return false
	}
	if !bytecodeCanonicalArrayDefinitionReceiver(vm.interp, receiver) {
		return false
	}
	return vm.env.RuntimeData() == nil
}

func (entry bytecodeInlineArrayNewCallCacheEntry) matchesCanonicalArrayNewCallIdentity(program *bytecodeProgram, ip int, env *runtime.Environment) bool {
	return entry.valid &&
		entry.program == program &&
		entry.ip == ip &&
		entry.env == env
}

func (entry bytecodeInlineArrayNewCallCacheEntry) matchesCanonicalArrayNewCallVersions(envVersion uint64, globalRev uint64, methodVersion uint64) bool {
	return entry.envVersion == envVersion &&
		entry.globalRevision == globalRev &&
		entry.methodCacheVersion == methodVersion
}

func (vm *bytecodeVM) promoteCanonicalArrayNewCallHot(entry bytecodeInlineArrayNewCallCacheEntry) {
	if vm == nil || !entry.valid {
		return
	}
	for i := 0; i < len(vm.arrayNewCallHot); i++ {
		if vm.arrayNewCallHot[i].matchesCanonicalArrayNewCallIdentity(entry.program, entry.ip, entry.env) {
			copy(vm.arrayNewCallHot[1:i+1], vm.arrayNewCallHot[0:i])
			vm.arrayNewCallHot[0] = entry
			return
		}
	}
	copy(vm.arrayNewCallHot[1:], vm.arrayNewCallHot[:len(vm.arrayNewCallHot)-1])
	vm.arrayNewCallHot[0] = entry
}

func (vm *bytecodeVM) canonicalArrayNewCallVersions(env *runtime.Environment) (uint64, uint64, uint64) {
	return vm.bytecodeEnvRevision(env), vm.bytecodeGlobalRevision(), vm.bytecodeMethodCacheVersion()
}

func (vm *bytecodeVM) lookupCachedCanonicalArrayNewCall(program *bytecodeProgram, ip int, instr bytecodeInstruction, receiver runtime.Value) bool {
	if program == nil || !vm.canUseCanonicalArrayNewCallCache(instr, receiver) {
		return false
	}
	env := vm.env
	var (
		envVersion    uint64
		globalRev     uint64
		methodVersion uint64
		haveVersions  bool
	)
	for i := 0; i < len(vm.arrayNewCallHot); i++ {
		hot := vm.arrayNewCallHot[i]
		if !hot.matchesCanonicalArrayNewCallIdentity(program, ip, env) {
			continue
		}
		if !haveVersions {
			envVersion, globalRev, methodVersion = vm.canonicalArrayNewCallVersions(env)
			haveVersions = true
		}
		if !hot.matchesCanonicalArrayNewCallVersions(envVersion, globalRev, methodVersion) {
			return false
		}
		return true
	}
	if vm.arrayNewCallCache == nil {
		return false
	}
	if !haveVersions {
		envVersion, globalRev, methodVersion = vm.canonicalArrayNewCallVersions(env)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	entry, ok := vm.arrayNewCallCache[key]
	if !ok ||
		entry.env != env ||
		entry.envVersion != envVersion ||
		entry.globalRevision != globalRev ||
		entry.methodCacheVersion != methodVersion {
		return false
	}
	vm.promoteCanonicalArrayNewCallHot(bytecodeInlineArrayNewCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		env:                entry.env,
		envVersion:         entry.envVersion,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
	})
	return true
}

func (vm *bytecodeVM) storeCachedCanonicalArrayNewCall(program *bytecodeProgram, ip int, instr bytecodeInstruction, receiver runtime.Value) {
	if program == nil || !vm.canUseCanonicalArrayNewCallCache(instr, receiver) {
		return
	}
	envVersion, globalRev, methodVersion := vm.canonicalArrayNewCallVersions(vm.env)
	entry := bytecodeArrayNewCallCacheEntry{
		env:                vm.env,
		envVersion:         envVersion,
		globalRevision:     globalRev,
		methodCacheVersion: methodVersion,
	}
	if vm.arrayNewCallCache == nil {
		vm.arrayNewCallCache = make(map[bytecodeGlobalLookupCacheKey]bytecodeArrayNewCallCacheEntry, 4)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	vm.arrayNewCallCache[key] = entry
	vm.promoteCanonicalArrayNewCallHot(bytecodeInlineArrayNewCallCacheEntry{
		valid:              true,
		program:            program,
		ip:                 ip,
		env:                entry.env,
		envVersion:         entry.envVersion,
		globalRevision:     entry.globalRevision,
		methodCacheVersion: entry.methodCacheVersion,
	})
}
