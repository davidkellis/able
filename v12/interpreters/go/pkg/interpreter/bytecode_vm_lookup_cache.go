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
	name    string
	scope   *runtime.Environment
	version uint64
	value   runtime.Value
}

func (vm *bytecodeVM) canUseGlobalLookupCache(name string) bool {
	if vm == nil || vm.interp == nil || vm.interp.global == nil || vm.env != vm.interp.global {
		return false
	}
	return name != "" && !strings.Contains(name, ".")
}

func (vm *bytecodeVM) canUseScopeLookupCache(name string) bool {
	if vm == nil || vm.env == nil || vm.interp == nil {
		return false
	}
	if vm.canUseGlobalLookupCache(name) {
		return false
	}
	return name != "" && !strings.Contains(name, ".")
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
	if entry.version != vm.interp.global.Revision() {
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
		version: vm.interp.global.Revision(),
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
	if entry.scope != vm.env {
		return nil, false
	}
	if entry.version != vm.env.Revision() {
		return nil, false
	}
	return entry.value, true
}

func (vm *bytecodeVM) storeCachedScopeValue(program *bytecodeProgram, ip int, name string, value runtime.Value) {
	if vm == nil || vm.env == nil {
		return
	}
	if vm.scopeLookupCache == nil {
		vm.scopeLookupCache = make(map[bytecodeGlobalLookupCacheKey]bytecodeScopeLookupCacheEntry, 8)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	vm.scopeLookupCache[key] = bytecodeScopeLookupCacheEntry{
		name:    name,
		scope:   vm.env,
		version: vm.env.Revision(),
		value:   value,
	}
}

func (vm *bytecodeVM) lookupCachedName(program *bytecodeProgram, ip int, name string) (runtime.Value, bool) {
	if vm.canUseGlobalLookupCache(name) {
		if cached, ok := vm.lookupCachedGlobalValue(program, ip, name); ok {
			return cached, true
		}
		if val, ok := vm.env.Lookup(name); ok {
			vm.storeCachedGlobalValue(program, ip, name, val)
			return val, true
		}
		return nil, false
	}

	if vm.canUseScopeLookupCache(name) {
		if cached, ok := vm.lookupCachedScopeValue(program, ip, name); ok {
			return cached, true
		}
		if current, ok := vm.env.LookupInCurrentScope(name); ok {
			vm.storeCachedScopeValue(program, ip, name, current)
			return current, true
		}
	}
	return vm.env.Lookup(name)
}

func (vm *bytecodeVM) resolveCachedName(program *bytecodeProgram, ip int, name string) (runtime.Value, error) {
	if val, ok := vm.lookupCachedName(program, ip, name); ok {
		return val, nil
	}
	return nil, fmt.Errorf("Undefined variable '%s'", name)
}
