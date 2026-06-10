package interpreter

import "testing"

func TestBytecodeVM_ResetForRunPreservesLookupCaches(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}

	vm := newBytecodeVM(interp, env)
	vm.globalLookupCache = map[*bytecodeProgram][]bytecodeGlobalLookupCacheEntry{
		program: []bytecodeGlobalLookupCacheEntry{{valid: true}},
	}
	vm.scopeLookupCache = map[*bytecodeProgram][]bytecodeScopeLookupCacheEntry{
		program: []bytecodeScopeLookupCacheEntry{{env: env, owner: env}},
	}
	vm.nameLookupHot = bytecodeInlineNameLookupCacheEntry{valid: true, program: program, ip: 1, entry: &bytecodeScopeLookupCacheEntry{env: env, owner: env}}
	vm.callNameCache = map[*bytecodeProgram][]*bytecodeCallNameCacheEntry{
		program: []*bytecodeCallNameCacheEntry{{name: "call", env: env, owner: env}, {name: "call", env: env, owner: env}},
	}
	vm.callNameHot = bytecodeInlineCallNameCacheEntry{valid: true, program: program, ip: 1, entry: &bytecodeCallNameCacheEntry{name: "call", env: env, owner: env}}
	vm.memberMethodCache = map[bytecodeMemberMethodCacheKey]bytecodeMemberMethodCacheEntry{
		{program: program, ip: 1, member: "len"}: {},
	}
	vm.memberMethodHot = bytecodeInlineMemberMethodCacheEntry{valid: true, program: program, ip: 1, member: "len"}
	vm.memberMethodDirect[bytecodeMemberMethodDirectIndex(1)] = vm.memberMethodHot
	vm.indexMethodCache = map[*bytecodeProgram]*bytecodeIndexMethodCacheTable{
		program: {get: make([]bytecodeIndexMethodCacheEntry, 2)},
	}
	vm.indexMethodHot = bytecodeInlineIndexMethodCacheEntry{valid: true, program: program, ip: 1, methodKind: bytecodeIndexMethodCacheGet}
	vm.activeLookup.program = program
	vm.activeLookup.indexMethodTable = vm.indexMethodCache[program]

	vm.resetForRun(interp, env)

	if vm.globalLookupCache == nil || len(vm.globalLookupCache) != 1 {
		t.Fatalf("expected global lookup cache to persist across reset, got %#v", vm.globalLookupCache)
	}
	if vm.scopeLookupCache == nil || len(vm.scopeLookupCache) != 1 {
		t.Fatalf("expected scope lookup cache to persist across reset, got %#v", vm.scopeLookupCache)
	}
	if vm.callNameCache == nil || len(vm.callNameCache) != 1 {
		t.Fatalf("expected call-name cache to persist across reset, got %#v", vm.callNameCache)
	}
	if vm.memberMethodCache == nil || len(vm.memberMethodCache) != 1 {
		t.Fatalf("expected member-method cache to persist across reset, got %#v", vm.memberMethodCache)
	}
	if vm.indexMethodCache == nil || len(vm.indexMethodCache) != 1 {
		t.Fatalf("expected index-method cache to persist across reset, got %#v", vm.indexMethodCache)
	}
	if !vm.nameLookupHot.valid || !vm.callNameHot.valid || !vm.memberMethodHot.valid || !vm.indexMethodHot.valid {
		t.Fatalf("expected inline hot caches to persist across reset")
	}
	if !vm.memberMethodDirect[bytecodeMemberMethodDirectIndex(1)].valid {
		t.Fatalf("expected direct member-method cache to persist across reset")
	}
	if vm.activeLookup.program != nil || vm.activeLookup.indexMethodTable != nil {
		t.Fatalf("expected active cache pointers to clear across reset")
	}
}
