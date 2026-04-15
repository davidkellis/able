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
	name    string
	version uint64
	value   runtime.Value
}

type bytecodeScopeLookupCacheEntry struct {
	name         string
	env          *runtime.Environment
	envVersion   uint64
	owner        *runtime.Environment
	ownerVersion uint64
	value        runtime.Value
}

type bytecodeInlineNameLookupCacheEntry struct {
	valid        bool
	program      *bytecodeProgram
	ip           int
	name         string
	env          *runtime.Environment
	envVersion   uint64
	owner        *runtime.Environment
	ownerVersion uint64
	value        runtime.Value
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

func (vm *bytecodeVM) lookupCachedGlobalValue(program *bytecodeProgram, ip int, name string) (runtime.Value, bool) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.globalLookupCache == nil {
		return nil, false
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	entry, ok := vm.globalLookupCache[key]
	if !ok || entry.name != name {
		return nil, false
	}
	if entry.version != vm.bytecodeGlobalRevision() {
		return nil, false
	}
	return entry.value, true
}

func (vm *bytecodeVM) storeCachedGlobalValue(program *bytecodeProgram, ip int, name string, value runtime.Value) {
	if vm == nil || vm.interp == nil || vm.interp.global == nil {
		return
	}
	if vm.globalLookupCache == nil {
		vm.globalLookupCache = make(map[bytecodeGlobalLookupCacheKey]bytecodeGlobalLookupCacheEntry, 8)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	vm.globalLookupCache[key] = bytecodeGlobalLookupCacheEntry{
		name:    name,
		version: vm.bytecodeGlobalRevision(),
		value:   value,
	}
}

func (vm *bytecodeVM) lookupCachedScopeValue(program *bytecodeProgram, ip int, name string) (runtime.Value, bool) {
	if vm == nil || vm.scopeLookupCache == nil || vm.env == nil {
		return nil, false
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	entry, ok := vm.scopeLookupCache[key]
	if !ok || entry.name != name {
		return nil, false
	}
	if entry.env != vm.env {
		return nil, false
	}
	if entry.envVersion != vm.bytecodeEnvRevision(vm.env) {
		return nil, false
	}
	if entry.owner == nil {
		return nil, false
	}
	if entry.ownerVersion != vm.bytecodeEnvRevision(entry.owner) {
		return nil, false
	}
	return entry.value, true
}

func (vm *bytecodeVM) storeCachedScopeValue(program *bytecodeProgram, ip int, name string, owner *runtime.Environment, value runtime.Value) {
	if vm == nil || vm.env == nil || owner == nil {
		return
	}
	if vm.scopeLookupCache == nil {
		vm.scopeLookupCache = make(map[bytecodeGlobalLookupCacheKey]bytecodeScopeLookupCacheEntry, 8)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	entry := bytecodeScopeLookupCacheEntry{
		name:         name,
		env:          vm.env,
		envVersion:   vm.bytecodeEnvRevision(vm.env),
		owner:        owner,
		ownerVersion: vm.bytecodeEnvRevision(owner),
		value:        value,
	}
	vm.scopeLookupCache[key] = entry
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{
		valid:        true,
		program:      program,
		ip:           ip,
		name:         name,
		env:          entry.env,
		envVersion:   entry.envVersion,
		owner:        entry.owner,
		ownerVersion: entry.ownerVersion,
		value:        value,
	}
}

func (vm *bytecodeVM) lookupHotName(program *bytecodeProgram, ip int, name string) (runtime.Value, bool) {
	if vm == nil || vm.env == nil {
		return nil, false
	}
	hot := vm.nameLookupHot
	if !hot.valid || hot.program != program || hot.ip != ip || hot.name != name || hot.env != vm.env {
		return nil, false
	}
	if hot.envVersion != vm.bytecodeEnvRevision(vm.env) {
		return nil, false
	}
	if hot.owner == nil || hot.ownerVersion != vm.bytecodeEnvRevision(hot.owner) {
		return nil, false
	}
	return hot.value, true
}

func (vm *bytecodeVM) storeHotGlobalName(program *bytecodeProgram, ip int, name string, value runtime.Value, owner *runtime.Environment) {
	if vm == nil || vm.env == nil || owner == nil {
		return
	}
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{
		valid:        true,
		program:      program,
		ip:           ip,
		name:         name,
		env:          vm.env,
		envVersion:   vm.bytecodeEnvRevision(vm.env),
		owner:        owner,
		ownerVersion: vm.bytecodeEnvRevision(owner),
		value:        value,
	}
}

func (vm *bytecodeVM) lookupCachedIdentifierName(program *bytecodeProgram, ip int, name string) (runtime.Value, bool) {
	if cached, ok := vm.lookupHotName(program, ip, name); ok {
		return cached, true
	}
	if vm != nil && vm.interp != nil && vm.interp.global != nil && vm.env == vm.interp.global {
		if cached, ok := vm.lookupCachedGlobalValue(program, ip, name); ok {
			vm.storeHotGlobalName(program, ip, name, cached, vm.interp.global)
			return cached, true
		}
		if val, owner, ok := vm.env.LookupWithOwner(name); ok {
			vm.storeCachedGlobalValue(program, ip, name, val)
			vm.storeHotGlobalName(program, ip, name, val, owner)
			return val, true
		}
		return nil, false
	}

	if cached, ok := vm.lookupCachedScopeValue(program, ip, name); ok {
		return cached, true
	}
	if val, owner, ok := vm.env.LookupWithOwner(name); ok {
		vm.storeCachedScopeValue(program, ip, name, owner, val)
		return val, true
	}
	return vm.env.Lookup(name)
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
