package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/runtime"
)

type bytecodeGlobalLookupCacheKey struct {
	program *bytecodeProgram
	ip      int
}

type bytecodeGlobalLookupCacheEntry struct {
	valid   bool
	version uint64
	value   runtime.Value
}

type bytecodeScopeLookupCacheEntry struct {
	name                string
	env                 *runtime.Environment
	envVersion          uint64
	nameShapeStateID    uint64
	bindingShapeVersion uint64
	nameShapeVersion    uint64
	owner               *runtime.Environment
	ownerVersion        uint64
	value               runtime.Value
}

type bytecodeInlineNameLookupCacheEntry struct {
	valid   bool
	program *bytecodeProgram
	ip      int
	entry   *bytecodeScopeLookupCacheEntry
}

type bytecodeResolvedIdentifierLookup struct {
	value        runtime.Value
	env          *runtime.Environment
	envVersion   uint64
	owner        *runtime.Environment
	ownerVersion uint64
}

func (vm *bytecodeVM) seedNameLookupHot(program *bytecodeProgram, ip int, entry *bytecodeScopeLookupCacheEntry) {
	vm.nameLookupHot.valid = true
	vm.nameLookupHot.program = program
	vm.nameLookupHot.ip = ip
	vm.nameLookupHot.entry = entry
}

func (vm *bytecodeVM) resolvedIdentifierLookup(value runtime.Value, owner *runtime.Environment) bytecodeResolvedIdentifierLookup {
	if vm == nil || vm.env == nil || owner == nil {
		return bytecodeResolvedIdentifierLookup{}
	}
	envVersion := vm.bytecodeEnvRevision(vm.env)
	ownerVersion := envVersion
	if owner != vm.env {
		ownerVersion = vm.bytecodeEnvRevision(owner)
	}
	return vm.resolvedIdentifierLookupWithVersions(value, vm.env, envVersion, owner, ownerVersion)
}

func (vm *bytecodeVM) resolvedIdentifierLookupWithVersions(value runtime.Value, env *runtime.Environment, envVersion uint64, owner *runtime.Environment, ownerVersion uint64) bytecodeResolvedIdentifierLookup {
	if env == nil || owner == nil {
		return bytecodeResolvedIdentifierLookup{}
	}
	return bytecodeResolvedIdentifierLookup{
		value:        value,
		env:          env,
		envVersion:   envVersion,
		owner:        owner,
		ownerVersion: ownerVersion,
	}
}

func bytecodeResolvedIdentifierLookupFromScopeEntry(entry *bytecodeScopeLookupCacheEntry) bytecodeResolvedIdentifierLookup {
	if entry == nil {
		return bytecodeResolvedIdentifierLookup{}
	}
	return bytecodeResolvedIdentifierLookup{
		value:        entry.value,
		env:          entry.env,
		envVersion:   entry.envVersion,
		owner:        entry.owner,
		ownerVersion: entry.ownerVersion,
	}
}

func (entry *bytecodeScopeLookupCacheEntry) lexicalNameShapeValid(name string, env *runtime.Environment) bool {
	if entry == nil || env == nil || name == "" || entry.name != name {
		return false
	}
	if entry.nameShapeStateID != env.BindingShapeStateID() {
		return false
	}
	bindingShapeVersion := env.BindingShapeRevision()
	if entry.bindingShapeVersion == bindingShapeVersion {
		return true
	}
	if entry.nameShapeVersion != env.BindingNameRevision(name) {
		return false
	}
	entry.bindingShapeVersion = bindingShapeVersion
	return true
}

func (vm *bytecodeVM) canUseLexicalLookupCache(name string) bool {
	if vm == nil || vm.env == nil || vm.interp == nil {
		return false
	}
	return name != "" && !strings.Contains(name, ".")
}

func (vm *bytecodeVM) canUseGlobalLookupCache(name string) bool {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.env != vm.interp.global {
		return false
	}
	return vm.canUseLexicalLookupCache(name)
}

func (vm *bytecodeVM) canUseScopeLookupCache(name string) bool {
	if !vm.canUseLexicalLookupCache(name) {
		return false
	}
	if vm.canUseGlobalLookupCache(name) {
		return false
	}
	return true
}

func (vm *bytecodeVM) globalLookupCacheEntries(program *bytecodeProgram, create bool) ([]bytecodeGlobalLookupCacheEntry, bool) {
	if vm == nil || program == nil {
		return nil, false
	}
	if ipCount := len(program.instructions); ipCount == 0 {
		return nil, false
	}
	if vm.globalLookupHotProgram == program && vm.globalLookupHotEntries != nil {
		return vm.globalLookupHotEntries, true
	}
	entries, ok := vm.globalLookupCache[program]
	if !ok {
		if !create {
			return nil, false
		}
		entries = make([]bytecodeGlobalLookupCacheEntry, len(program.instructions))
		if vm.globalLookupCache == nil {
			vm.globalLookupCache = make(map[*bytecodeProgram][]bytecodeGlobalLookupCacheEntry, 8)
		}
		vm.globalLookupCache[program] = entries
	}
	vm.globalLookupHotProgram = program
	vm.globalLookupHotEntries = entries
	return entries, true
}

func (vm *bytecodeVM) activeGlobalLookupCacheEntries(program *bytecodeProgram, create bool) ([]bytecodeGlobalLookupCacheEntry, bool) {
	if vm == nil || program == nil {
		return nil, false
	}
	if vm.activeLookup.program == program && vm.activeLookup.globalLookupEntries != nil {
		return vm.activeLookup.globalLookupEntries, true
	}
	entries, ok := vm.globalLookupCacheEntries(program, create)
	if !ok {
		return nil, false
	}
	if vm.activeLookup.program == program {
		vm.activeLookup.globalLookupEntries = entries
	}
	return entries, true
}

func (vm *bytecodeVM) scopeLookupCacheEntries(program *bytecodeProgram, create bool) ([]bytecodeScopeLookupCacheEntry, bool) {
	if vm == nil || program == nil {
		return nil, false
	}
	if ipCount := len(program.instructions); ipCount == 0 {
		return nil, false
	}
	if vm.scopeLookupHotProgram == program && vm.scopeLookupHotEntries != nil {
		return vm.scopeLookupHotEntries, true
	}
	entries, ok := vm.scopeLookupCache[program]
	if !ok {
		if !create {
			return nil, false
		}
		entries = make([]bytecodeScopeLookupCacheEntry, len(program.instructions))
		if vm.scopeLookupCache == nil {
			vm.scopeLookupCache = make(map[*bytecodeProgram][]bytecodeScopeLookupCacheEntry, 8)
		}
		vm.scopeLookupCache[program] = entries
	}
	vm.scopeLookupHotProgram = program
	vm.scopeLookupHotEntries = entries
	return entries, true
}

func (vm *bytecodeVM) activeScopeLookupCacheEntries(program *bytecodeProgram, create bool) ([]bytecodeScopeLookupCacheEntry, bool) {
	if vm == nil || program == nil {
		return nil, false
	}
	if vm.activeLookup.program == program && vm.activeLookup.scopeLookupEntries != nil {
		return vm.activeLookup.scopeLookupEntries, true
	}
	entries, ok := vm.scopeLookupCacheEntries(program, create)
	if !ok {
		return nil, false
	}
	if vm.activeLookup.program == program {
		vm.activeLookup.scopeLookupEntries = entries
	}
	return entries, true
}

func (vm *bytecodeVM) setActiveLookupProgram(program *bytecodeProgram) {
	if vm == nil {
		return
	}
	if program != nil && vm.activeLookup.program == program {
		return
	}
	vm.activeLookup = bytecodeActiveLookupProgramState{program: program}
}

func (vm *bytecodeVM) captureActiveLookupProgramState(program *bytecodeProgram) bytecodeActiveLookupProgramState {
	if vm == nil || program == nil || vm.activeLookup.program != program {
		return bytecodeActiveLookupProgramState{}
	}
	return vm.activeLookup
}

func (vm *bytecodeVM) restoreActiveLookupProgramState(state bytecodeActiveLookupProgramState) bool {
	if vm == nil || state.program == nil {
		return false
	}
	vm.activeLookup = state
	return true
}

func (vm *bytecodeVM) lookupCachedGlobalValue(program *bytecodeProgram, ip int, _ string) (runtime.Value, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || program == nil || ip < 0 || ip >= len(program.instructions) {
		return nil, false
	}
	entries, ok := vm.activeGlobalLookupCacheEntries(program, false)
	if !ok {
		return nil, false
	}
	entry := entries[ip]
	if !entry.valid {
		return nil, false
	}
	if entry.version != vm.bytecodeGlobalRevision() {
		return nil, false
	}
	return entry.value, true
}

func (vm *bytecodeVM) storeCachedGlobalValue(program *bytecodeProgram, ip int, _ string, value runtime.Value) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || program == nil || ip < 0 || ip >= len(program.instructions) {
		return
	}
	entries, ok := vm.activeGlobalLookupCacheEntries(program, true)
	if !ok {
		return
	}
	entries[ip] = bytecodeGlobalLookupCacheEntry{
		valid:   true,
		version: vm.bytecodeGlobalRevision(),
		value:   value,
	}
}

func (vm *bytecodeVM) lookupCachedScopeEntry(program *bytecodeProgram, ip int, name string) (*bytecodeScopeLookupCacheEntry, bool) {
	if vm == nil || vm.env == nil {
		return nil, false
	}
	currentEnv := vm.env
	singleThread := vm.bytecodeSingleThread()
	return vm.lookupCachedScopeEntryWithVersion(program, ip, name, currentEnv, bytecodeEnvironmentRevision(currentEnv, singleThread), singleThread)
}

func (vm *bytecodeVM) lookupCachedScopeEntryWithVersion(program *bytecodeProgram, ip int, name string, currentEnv *runtime.Environment, currentEnvVersion uint64, singleThread bool) (*bytecodeScopeLookupCacheEntry, bool) {
	return vm.lookupCachedScopeEntryWithVersionInternal(program, ip, name, currentEnv, currentEnvVersion, singleThread, false)
}

func (vm *bytecodeVM) lookupCachedOrRefreshScopeEntryWithVersion(program *bytecodeProgram, ip int, name string, currentEnv *runtime.Environment, currentEnvVersion uint64, singleThread bool) (*bytecodeScopeLookupCacheEntry, bool) {
	return vm.lookupCachedScopeEntryWithVersionInternal(program, ip, name, currentEnv, currentEnvVersion, singleThread, true)
}

func (vm *bytecodeVM) lookupCachedScopeEntryWithVersionInternal(program *bytecodeProgram, ip int, name string, currentEnv *runtime.Environment, currentEnvVersion uint64, singleThread bool, refreshSameEnv bool) (*bytecodeScopeLookupCacheEntry, bool) {
	if vm == nil || currentEnv == nil || program == nil || ip < 0 || ip >= len(program.instructions) {
		return nil, false
	}
	entries, ok := vm.activeScopeLookupCacheEntries(program, false)
	if !ok {
		return nil, false
	}
	entry := &entries[ip]
	if entry.env == nil {
		return nil, false
	}
	if entry.name != "" && name != "" && entry.name != name {
		return nil, false
	}
	if entry.env == currentEnv {
		if entry.envVersion != currentEnvVersion {
			if entry.owner != currentEnv &&
				vm.scopeLookupEntryOuterEnvValid(entry, currentEnv, name) &&
				bytecodeEnvironmentRevision(entry.owner, singleThread) == entry.ownerVersion {
				entry.envVersion = currentEnvVersion
				vm.seedNameLookupHot(program, ip, entry)
				return entry, true
			}
			if refreshSameEnv {
				return vm.refreshSameEnvScopeEntryCheckedWithVersion(entry, program, ip, name, currentEnv, currentEnvVersion, singleThread)
			}
			return nil, false
		}
	} else if !vm.scopeLookupEntryOuterEnvValid(entry, currentEnv, name) {
		return nil, false
	}
	if entry.owner == nil {
		return nil, false
	}
	if entry.owner == currentEnv {
		if entry.ownerVersion != currentEnvVersion {
			if refreshSameEnv {
				return vm.refreshSameEnvScopeEntryCheckedWithVersion(entry, program, ip, name, currentEnv, currentEnvVersion, singleThread)
			}
			return nil, false
		}
	} else if entry.ownerVersion != bytecodeEnvironmentRevision(entry.owner, singleThread) {
		return nil, false
	}
	vm.seedNameLookupHot(program, ip, entry)
	return entry, true
}

func (vm *bytecodeVM) refreshSameEnvScopeEntryWithVersion(program *bytecodeProgram, ip int, name string, currentEnv *runtime.Environment, currentEnvVersion uint64, singleThread bool) (*bytecodeScopeLookupCacheEntry, bool) {
	if vm == nil || !singleThread || currentEnv == nil || program == nil || ip < 0 || ip >= len(program.instructions) {
		return nil, false
	}
	entries, ok := vm.activeScopeLookupCacheEntries(program, false)
	if !ok {
		return nil, false
	}
	entry := &entries[ip]
	if entry.env != currentEnv || entry.owner != currentEnv {
		return nil, false
	}
	if entry.name != "" && name != "" && entry.name != name {
		return nil, false
	}
	return vm.refreshSameEnvScopeEntryCheckedWithVersion(entry, program, ip, name, currentEnv, currentEnvVersion, singleThread)
}

func (vm *bytecodeVM) refreshSameEnvScopeEntryCheckedWithVersion(entry *bytecodeScopeLookupCacheEntry, program *bytecodeProgram, ip int, name string, currentEnv *runtime.Environment, currentEnvVersion uint64, singleThread bool) (*bytecodeScopeLookupCacheEntry, bool) {
	if vm == nil || !singleThread || entry == nil || currentEnv == nil {
		return nil, false
	}
	if entry.env != currentEnv || entry.owner != currentEnv {
		return nil, false
	}
	value, ok := currentEnv.LookupInCurrentScope(name)
	if !ok {
		return nil, false
	}
	entry.name = name
	entry.envVersion = currentEnvVersion
	entry.nameShapeStateID = 0
	entry.bindingShapeVersion = 0
	entry.nameShapeVersion = 0
	entry.ownerVersion = currentEnvVersion
	entry.value = value
	vm.seedNameLookupHot(program, ip, entry)
	return entry, true
}

func (vm *bytecodeVM) scopeLookupEntryOuterEnvValid(entry *bytecodeScopeLookupCacheEntry, currentEnv *runtime.Environment, name string) bool {
	if vm == nil || entry == nil || currentEnv == nil {
		return false
	}
	if vm.interp == nil || vm.interp.global == nil || entry.owner == nil {
		return false
	}
	if entry.owner != vm.interp.global && currentEnv.Parent() != entry.owner {
		return false
	}
	return entry.lexicalNameShapeValid(name, currentEnv)
}

func (vm *bytecodeVM) scopeLookupEntryShapeMetadata(name string, currentEnv *runtime.Environment, owner *runtime.Environment) (uint64, uint64, uint64) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || name == "" || currentEnv == nil || owner == nil {
		return 0, 0, 0
	}
	if currentEnv == owner {
		return 0, 0, 0
	}
	if owner != vm.interp.global && currentEnv.Parent() != owner {
		return 0, 0, 0
	}
	return currentEnv.BindingShapeStateID(), currentEnv.BindingShapeRevision(), currentEnv.BindingNameRevision(name)
}

func (vm *bytecodeVM) lookupCachedScopeValue(program *bytecodeProgram, ip int, name string) (bytecodeResolvedIdentifierLookup, bool) {
	entry, ok := vm.lookupCachedScopeEntry(program, ip, name)
	if !ok {
		return bytecodeResolvedIdentifierLookup{}, false
	}
	return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
}

func (vm *bytecodeVM) storeCachedScopeEntry(program *bytecodeProgram, ip int, name string, owner *runtime.Environment, value runtime.Value) *bytecodeScopeLookupCacheEntry {
	if vm == nil || vm.env == nil || owner == nil {
		return nil
	}
	currentEnvVersion := vm.bytecodeEnvRevision(vm.env)
	ownerVersion := currentEnvVersion
	if owner != vm.env {
		ownerVersion = vm.bytecodeEnvRevision(owner)
	}
	return vm.storeCachedScopeEntryWithVersions(program, ip, name, vm.env, currentEnvVersion, owner, ownerVersion, value)
}

func (vm *bytecodeVM) storeCachedScopeEntryWithVersions(program *bytecodeProgram, ip int, name string, currentEnv *runtime.Environment, currentEnvVersion uint64, owner *runtime.Environment, ownerVersion uint64, value runtime.Value) *bytecodeScopeLookupCacheEntry {
	if vm == nil || currentEnv == nil || owner == nil || program == nil || ip < 0 || ip >= len(program.instructions) {
		return nil
	}
	entries, ok := vm.activeScopeLookupCacheEntries(program, true)
	if !ok {
		return nil
	}
	entry := &entries[ip]
	if owner == currentEnv {
		entry.name = name
		entry.env = currentEnv
		entry.envVersion = currentEnvVersion
		entry.nameShapeStateID = 0
		entry.bindingShapeVersion = 0
		entry.nameShapeVersion = 0
		entry.owner = owner
		entry.ownerVersion = ownerVersion
		entry.value = value
		vm.seedNameLookupHot(program, ip, entry)
		return entry
	}
	nameShapeStateID, bindingShapeVersion, nameShapeVersion := vm.scopeLookupEntryShapeMetadata(name, currentEnv, owner)
	entry.name = name
	entry.env = currentEnv
	entry.envVersion = currentEnvVersion
	entry.nameShapeStateID = nameShapeStateID
	entry.bindingShapeVersion = bindingShapeVersion
	entry.nameShapeVersion = nameShapeVersion
	entry.owner = owner
	entry.ownerVersion = ownerVersion
	entry.value = value
	vm.seedNameLookupHot(program, ip, entry)
	return entry
}

func (vm *bytecodeVM) storeCachedScopeValue(program *bytecodeProgram, ip int, name string, owner *runtime.Environment, value runtime.Value) bytecodeResolvedIdentifierLookup {
	entry := vm.storeCachedScopeEntry(program, ip, name, owner, value)
	if entry == nil {
		return bytecodeResolvedIdentifierLookup{}
	}
	return bytecodeResolvedIdentifierLookupFromScopeEntry(entry)
}

func (vm *bytecodeVM) storeHotGlobalNameEntry(program *bytecodeProgram, ip int, name string, value runtime.Value, owner *runtime.Environment) *bytecodeScopeLookupCacheEntry {
	if vm == nil || vm.env == nil || owner == nil {
		return nil
	}
	currentEnvVersion := vm.bytecodeEnvRevision(vm.env)
	ownerVersion := currentEnvVersion
	if owner != vm.env {
		ownerVersion = vm.bytecodeEnvRevision(owner)
	}
	return vm.storeHotGlobalNameEntryWithVersions(program, ip, name, vm.env, currentEnvVersion, owner, ownerVersion, value)
}

func (vm *bytecodeVM) storeHotGlobalNameEntryWithVersions(program *bytecodeProgram, ip int, name string, currentEnv *runtime.Environment, currentEnvVersion uint64, owner *runtime.Environment, ownerVersion uint64, value runtime.Value) *bytecodeScopeLookupCacheEntry {
	if vm == nil || currentEnv == nil || owner == nil {
		return nil
	}
	entry := &vm.nameLookupHotEntry
	nameShapeStateID, bindingShapeVersion, nameShapeVersion := vm.scopeLookupEntryShapeMetadata(name, currentEnv, owner)
	entry.name = name
	entry.env = currentEnv
	entry.envVersion = currentEnvVersion
	entry.nameShapeStateID = nameShapeStateID
	entry.bindingShapeVersion = bindingShapeVersion
	entry.nameShapeVersion = nameShapeVersion
	entry.owner = owner
	entry.ownerVersion = ownerVersion
	entry.value = value
	vm.seedNameLookupHot(program, ip, entry)
	return entry
}

func (vm *bytecodeVM) storeHotGlobalName(program *bytecodeProgram, ip int, name string, value runtime.Value, owner *runtime.Environment) bytecodeResolvedIdentifierLookup {
	entry := vm.storeHotGlobalNameEntry(program, ip, name, value, owner)
	if entry == nil {
		return bytecodeResolvedIdentifierLookup{}
	}
	return bytecodeResolvedIdentifierLookupFromScopeEntry(entry)
}

func (vm *bytecodeVM) lookupCachedIdentifierName(program *bytecodeProgram, ip int, name string) (runtime.Value, bool) {
	if vm == nil || vm.env == nil {
		return nil, false
	}
	currentEnv := vm.env
	singleThread := vm.bytecodeSingleThread()
	currentEnvVersion := bytecodeEnvironmentRevision(currentEnv, singleThread)
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	if entry, ok := vm.lookupHotNameEntryWithVersion(program, ip, name, currentEnv, currentEnvVersion, singleThread); ok {
		if statsEnabled {
			vm.interp.recordBytecodeLoadNameHotHit()
		}
		return entry.value, true
	}
	if vm.interp != nil && vm.interp.global != nil && currentEnv == vm.interp.global {
		if cached, ok := vm.lookupCachedGlobalValue(program, ip, name); ok {
			if statsEnabled {
				vm.interp.recordBytecodeLoadNameGlobalCacheHit()
			}
			vm.storeHotGlobalNameEntryWithVersions(program, ip, name, currentEnv, currentEnvVersion, vm.interp.global, currentEnvVersion, cached)
			return cached, true
		}
		if val, owner, ownerVersion, ok := currentEnv.LookupWithOwnerAndRevisionHint(name, singleThread); ok {
			if statsEnabled {
				vm.interp.recordBytecodeLoadNameDirectResolve(owner == currentEnv)
				vm.interp.recordBytecodeLoadNameGlobalStore()
			}
			vm.storeCachedGlobalValue(program, ip, name, val)
			vm.storeHotGlobalNameEntryWithVersions(program, ip, name, currentEnv, currentEnvVersion, owner, ownerVersion, val)
			return val, true
		}
		return nil, false
	}
	if entry, ok := vm.lookupCachedOrRefreshScopeEntryWithVersion(program, ip, name, currentEnv, currentEnvVersion, singleThread); ok {
		if statsEnabled {
			vm.interp.recordBytecodeLoadNameScopeCacheHit()
		}
		return entry.value, true
	}
	if val, owner, ownerVersion, ok := currentEnv.LookupWithOwnerAndRevisionHint(name, singleThread); ok {
		if statsEnabled {
			vm.interp.recordBytecodeLoadNameDirectResolve(owner == currentEnv)
			vm.interp.recordBytecodeLoadNameScopeStore()
		}
		if entry := vm.storeCachedScopeEntryWithVersions(program, ip, name, currentEnv, currentEnvVersion, owner, ownerVersion, val); entry != nil {
			return entry.value, true
		}
		return val, true
	}
	return nil, false
}

func (vm *bytecodeVM) lookupIdentifierNameForCallCache(program *bytecodeProgram, ip int, name string) (bytecodeResolvedIdentifierLookup, bool) {
	if vm == nil || vm.env == nil {
		return bytecodeResolvedIdentifierLookup{}, false
	}
	currentEnv := vm.env
	singleThread := vm.bytecodeSingleThread()
	currentEnvVersion := bytecodeEnvironmentRevision(currentEnv, singleThread)
	if entry, ok := vm.lookupHotNameEntryWithVersion(program, ip, name, currentEnv, currentEnvVersion, singleThread); ok {
		return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
	}
	if vm.interp != nil && vm.interp.global != nil && currentEnv == vm.interp.global {
		if cached, ok := vm.lookupCachedGlobalValue(program, ip, name); ok {
			if entry := vm.storeHotGlobalNameEntryWithVersions(program, ip, name, currentEnv, currentEnvVersion, vm.interp.global, currentEnvVersion, cached); entry != nil {
				return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
			}
			return bytecodeResolvedIdentifierLookup{}, false
		}
		if val, owner, ownerVersion, ok := currentEnv.LookupWithOwnerAndRevisionHint(name, singleThread); ok {
			return vm.resolvedIdentifierLookupWithVersions(val, currentEnv, currentEnvVersion, owner, ownerVersion), true
		}
		return bytecodeResolvedIdentifierLookup{}, false
	}
	if entry, ok := vm.lookupCachedOrRefreshScopeEntryWithVersion(program, ip, name, currentEnv, currentEnvVersion, singleThread); ok {
		return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
	}
	if val, owner, ownerVersion, ok := currentEnv.LookupWithOwnerAndRevisionHint(name, singleThread); ok {
		return vm.resolvedIdentifierLookupWithVersions(val, currentEnv, currentEnvVersion, owner, ownerVersion), true
	}
	return bytecodeResolvedIdentifierLookup{}, false
}

func (vm *bytecodeVM) lookupHotNameEntry(program *bytecodeProgram, ip int, name string) (*bytecodeScopeLookupCacheEntry, bool) {
	if vm == nil || vm.env == nil {
		return nil, false
	}
	currentEnv := vm.env
	singleThread := vm.bytecodeSingleThread()
	return vm.lookupHotNameEntryWithVersion(program, ip, name, currentEnv, bytecodeEnvironmentRevision(currentEnv, singleThread), singleThread)
}

func (vm *bytecodeVM) lookupHotNameEntryWithVersion(program *bytecodeProgram, ip int, name string, currentEnv *runtime.Environment, currentEnvVersion uint64, singleThread bool) (*bytecodeScopeLookupCacheEntry, bool) {
	if vm == nil || currentEnv == nil {
		return nil, false
	}
	hot := vm.nameLookupHot
	if !hot.valid || hot.program != program || hot.ip != ip || hot.entry == nil {
		return nil, false
	}
	if hot.entry.name != "" && name != "" && hot.entry.name != name {
		return nil, false
	}
	if hot.entry.env == currentEnv {
		if hot.entry.envVersion != currentEnvVersion {
			return nil, false
		}
	} else if !vm.scopeLookupEntryOuterEnvValid(hot.entry, currentEnv, name) {
		return nil, false
	}
	if hot.entry.owner == nil {
		return nil, false
	}
	if hot.entry.owner == currentEnv {
		if hot.entry.ownerVersion != currentEnvVersion {
			return nil, false
		}
	} else if hot.entry.ownerVersion != bytecodeEnvironmentRevision(hot.entry.owner, singleThread) {
		return nil, false
	}
	return hot.entry, true
}

func (vm *bytecodeVM) lookupCachedIdentifierNameEntry(program *bytecodeProgram, ip int, name string) (bytecodeResolvedIdentifierLookup, bool) {
	if vm == nil || vm.env == nil {
		return bytecodeResolvedIdentifierLookup{}, false
	}
	currentEnv := vm.env
	singleThread := vm.bytecodeSingleThread()
	currentEnvVersion := bytecodeEnvironmentRevision(currentEnv, singleThread)
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	if entry, ok := vm.lookupHotNameEntryWithVersion(program, ip, name, currentEnv, currentEnvVersion, singleThread); ok {
		if statsEnabled {
			vm.interp.recordBytecodeLoadNameHotHit()
		}
		return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
	}
	if vm.interp != nil && vm.interp.global != nil && currentEnv == vm.interp.global {
		if cached, ok := vm.lookupCachedGlobalValue(program, ip, name); ok {
			if statsEnabled {
				vm.interp.recordBytecodeLoadNameGlobalCacheHit()
			}
			if entry := vm.storeHotGlobalNameEntryWithVersions(program, ip, name, currentEnv, currentEnvVersion, vm.interp.global, currentEnvVersion, cached); entry != nil {
				return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
			}
			return bytecodeResolvedIdentifierLookup{}, false
		}
		if val, owner, ownerVersion, ok := currentEnv.LookupWithOwnerAndRevisionHint(name, singleThread); ok {
			if statsEnabled {
				vm.interp.recordBytecodeLoadNameDirectResolve(owner == currentEnv)
				vm.interp.recordBytecodeLoadNameGlobalStore()
			}
			vm.storeCachedGlobalValue(program, ip, name, val)
			if entry := vm.storeHotGlobalNameEntryWithVersions(program, ip, name, currentEnv, currentEnvVersion, owner, ownerVersion, val); entry != nil {
				return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
			}
			return bytecodeResolvedIdentifierLookup{}, false
		}
		return bytecodeResolvedIdentifierLookup{}, false
	}

	if entry, ok := vm.lookupCachedOrRefreshScopeEntryWithVersion(program, ip, name, currentEnv, currentEnvVersion, singleThread); ok {
		if statsEnabled {
			vm.interp.recordBytecodeLoadNameScopeCacheHit()
		}
		return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
	}
	if val, owner, ownerVersion, ok := currentEnv.LookupWithOwnerAndRevisionHint(name, singleThread); ok {
		if statsEnabled {
			vm.interp.recordBytecodeLoadNameDirectResolve(owner == currentEnv)
			vm.interp.recordBytecodeLoadNameScopeStore()
		}
		if entry := vm.storeCachedScopeEntryWithVersions(program, ip, name, currentEnv, currentEnvVersion, owner, ownerVersion, val); entry != nil {
			return bytecodeResolvedIdentifierLookupFromScopeEntry(entry), true
		}
		return bytecodeResolvedIdentifierLookup{}, false
	}
	return bytecodeResolvedIdentifierLookup{}, false
}

func (vm *bytecodeVM) lookupCachedName(program *bytecodeProgram, ip int, name string) (runtime.Value, bool) {
	if !vm.canUseLexicalLookupCache(name) {
		return vm.env.Lookup(name)
	}
	return vm.lookupCachedIdentifierName(program, ip, name)
}

func (vm *bytecodeVM) resolveCachedIdentifierName(program *bytecodeProgram, ip int, name string) (runtime.Value, error) {
	if val, ok := vm.lookupCachedIdentifierName(program, ip, name); ok {
		return val, nil
	}
	return nil, fmt.Errorf("Undefined variable '%s'", name)
}

func (vm *bytecodeVM) resolveCachedName(program *bytecodeProgram, ip int, name string) (runtime.Value, error) {
	if val, ok := vm.lookupCachedName(program, ip, name); ok {
		return val, nil
	}
	return nil, fmt.Errorf("Undefined variable '%s'", name)
}
