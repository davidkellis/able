package interpreter

import "testing"

func TestBytecodeVM_IndexMethodCacheUsesActiveProgramTable(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: make([]bytecodeInstruction, 3)}
	other := &bytecodeProgram{instructions: make([]bytecodeInstruction, 2)}

	vm.setActiveLookupProgram(program)
	if vm.activeLookup.indexMethodTable != nil {
		t.Fatalf("new active program should not have an index method table")
	}

	entry, ok := vm.indexMethodCacheEntry(program, 1, "get", true)
	if !ok || entry == nil {
		t.Fatalf("expected active index method cache entry")
	}
	if vm.activeLookup.indexMethodTable == nil || vm.activeLookup.indexMethodTable != vm.indexMethodCache[program] {
		t.Fatalf("expected active index method table to point at program cache")
	}
	if len(vm.activeLookup.indexMethodGetEntries) != len(program.instructions) {
		t.Fatalf("active get entries length = %d, want %d", len(vm.activeLookup.indexMethodGetEntries), len(program.instructions))
	}
	entry.globalRevision = 17

	again, ok := vm.indexMethodCacheEntry(program, 1, "get", false)
	if !ok || again != entry {
		t.Fatalf("active index method cache entry = %p/%v, want %p/true", again, ok, entry)
	}
	if again.globalRevision != 17 {
		t.Fatalf("cached entry global revision = %d, want 17", again.globalRevision)
	}

	vm.setActiveLookupProgram(other)
	if vm.activeLookup.indexMethodTable != nil {
		t.Fatalf("switching to a program without a table should clear the active table")
	}
	if vm.activeLookup.indexMethodGetEntries != nil || vm.activeLookup.indexMethodSetEntries != nil {
		t.Fatalf("switching active program should clear active index entry slices")
	}
	otherEntry, ok := vm.indexMethodCacheEntry(other, 0, "set", true)
	if !ok || otherEntry == nil {
		t.Fatalf("expected active index method cache entry for second program")
	}
	if vm.activeLookup.indexMethodTable == nil || vm.activeLookup.indexMethodTable != vm.indexMethodCache[other] {
		t.Fatalf("expected second program to become the active index method table")
	}
	if len(vm.activeLookup.indexMethodSetEntries) != len(other.instructions) {
		t.Fatalf("active set entries length = %d, want %d", len(vm.activeLookup.indexMethodSetEntries), len(other.instructions))
	}

	vm.setActiveLookupProgram(program)
	if vm.activeLookup.indexMethodTable != nil {
		t.Fatalf("switching back should leave the active index method table lazy until first access")
	}
	if vm.activeLookup.indexMethodGetEntries != nil || vm.activeLookup.indexMethodSetEntries != nil {
		t.Fatalf("switching back should leave active index entry slices lazy")
	}
	again, ok = vm.indexMethodCacheEntry(program, 1, "get", false)
	if !ok || again != entry {
		t.Fatalf("reactivated entry = %p/%v, want %p/true", again, ok, entry)
	}
	if vm.activeLookup.indexMethodTable == nil || vm.activeLookup.indexMethodTable != vm.indexMethodCache[program] {
		t.Fatalf("lazy index method access should reactivate the original table")
	}
	if len(vm.activeLookup.indexMethodGetEntries) != len(program.instructions) {
		t.Fatalf("reactivated active get entries length = %d, want %d", len(vm.activeLookup.indexMethodGetEntries), len(program.instructions))
	}
	if again.globalRevision != 17 {
		t.Fatalf("reactivated entry global revision = %d, want 17", again.globalRevision)
	}
}
