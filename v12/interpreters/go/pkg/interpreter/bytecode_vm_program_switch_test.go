package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_InlineReturnRestoresCallerActiveLookupCaches(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	caller := &bytecodeProgram{instructions: make([]bytecodeInstruction, 3)}
	callee := &bytecodeProgram{instructions: []bytecodeInstruction{{op: bytecodeOpReturn}}}
	vm := newBytecodeVM(interp, env)

	vm.setActiveLookupProgram(caller)
	fn := &runtime.FunctionValue{Closure: env}
	env.Define("f", fn)
	entry := bytecodeBuildCallNameCacheEntry("f", vm.resolvedIdentifierLookup(fn, env), fn, 0, nil)
	cached := vm.storeCachedCallName(caller, 1, entry)
	if cached == nil {
		t.Fatalf("expected cached call-name entry")
	}
	callerCallNameEntries := vm.activeLookup.callNameEntries
	if len(callerCallNameEntries) != len(caller.instructions) {
		t.Fatalf("active call-name entries length = %d, want %d", len(callerCallNameEntries), len(caller.instructions))
	}

	callerSlots := []runtime.Value{runtime.NewSmallInt(1, runtime.IntegerI32)}
	calleeSlots := []runtime.Value{runtime.NewSmallInt(2, runtime.IntegerI32)}
	vm.slots = callerSlots
	vm.pushCallFrame(2, caller, callerSlots, env, nil, 0, 0, false, false)
	if len(vm.callFrames) != 1 || vm.callFrames[0].activeLookup.program != caller {
		t.Fatalf("expected full call frame to capture caller active lookup state")
	}

	delete(vm.callNameCache, caller)
	vm.callNameHotProgram = nil
	vm.callNameHotEntries = nil

	program := caller
	instructions := caller.instructions
	validatedIntConsts := []bool(nil)
	slotConstIntImmTable := (*bytecodeSlotConstIntImmediateTable)(nil)
	vm.slots = calleeSlots
	vm.switchRunProgram(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, callee)
	if vm.activeLookup.program != callee {
		t.Fatalf("active lookup program = %p, want callee %p", vm.activeLookup.program, callee)
	}

	err := vm.finishInlineReturn(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, nil, runtime.NilValue{}, bytecodeSimpleTypeCheckUnknown)
	if err != nil {
		t.Fatalf("finishInlineReturn failed: %v", err)
	}
	if program != caller || vm.currentProgram != caller {
		t.Fatalf("returned program/current = %p/%p, want caller %p", program, vm.currentProgram, caller)
	}
	if vm.activeLookup.program != caller {
		t.Fatalf("active lookup program = %p, want caller %p", vm.activeLookup.program, caller)
	}
	if len(vm.activeLookup.callNameEntries) != len(callerCallNameEntries) || vm.activeLookup.callNameEntries[1] != cached {
		t.Fatalf("caller active call-name entries were not restored from the saved frame state")
	}
	if !sameSlotFrame(vm.slots, callerSlots) {
		t.Fatalf("caller slot frame was not restored")
	}
	if len(vm.stack) != 1 || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("return stack = %#v, want single nil value", vm.stack)
	}
}

func TestBytecodeVM_SwitchRunProgramPreservesAlreadyActiveLookupCaches(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	caller := &bytecodeProgram{instructions: []bytecodeInstruction{{op: bytecodeOpConst}}}
	target := &bytecodeProgram{instructions: []bytecodeInstruction{{op: bytecodeOpConst}, {op: bytecodeOpReturn}}}
	globalEntries := []bytecodeGlobalLookupCacheEntry{{valid: true, version: 3, value: runtime.BoolValue{Val: true}}}
	scopeEntries := []bytecodeScopeLookupCacheEntry{{name: "value", value: runtime.NewSmallInt(7, runtime.IntegerI32)}}
	callEntry := &bytecodeCallNameCacheEntry{name: "f"}
	callEntries := []*bytecodeCallNameCacheEntry{callEntry}
	indexTable := &bytecodeIndexMethodCacheTable{}
	indexGetEntries := []bytecodeIndexMethodCacheEntry{{}}
	indexSetEntries := []bytecodeIndexMethodCacheEntry{{}}
	vm.activeLookup.program = target
	vm.activeLookup.globalLookupEntries = globalEntries
	vm.activeLookup.scopeLookupEntries = scopeEntries
	vm.activeLookup.callNameEntries = callEntries
	vm.activeLookup.indexMethodTable = indexTable
	vm.activeLookup.indexMethodGetEntries = indexGetEntries
	vm.activeLookup.indexMethodSetEntries = indexSetEntries

	program := caller
	instructions := caller.instructions
	vm.switchRunProgram(&program, &instructions, nil, nil, target)

	if program != target || vm.currentProgram != target {
		t.Fatalf("program/current = %p/%p, want target %p", program, vm.currentProgram, target)
	}
	if len(instructions) != len(target.instructions) {
		t.Fatalf("instruction slice length = %d, want %d", len(instructions), len(target.instructions))
	}
	if len(vm.activeLookup.globalLookupEntries) != len(globalEntries) || &vm.activeLookup.globalLookupEntries[0] != &globalEntries[0] {
		t.Fatalf("active global cache backing slice was not preserved")
	}
	if len(vm.activeLookup.scopeLookupEntries) != len(scopeEntries) || &vm.activeLookup.scopeLookupEntries[0] != &scopeEntries[0] {
		t.Fatalf("active scope cache backing slice was not preserved")
	}
	if len(vm.activeLookup.callNameEntries) != len(callEntries) || vm.activeLookup.callNameEntries[0] != callEntry {
		t.Fatalf("active call-name cache was not preserved")
	}
	if vm.activeLookup.indexMethodTable != indexTable {
		t.Fatalf("active index method table was not preserved")
	}
	if len(vm.activeLookup.indexMethodGetEntries) != len(indexGetEntries) || &vm.activeLookup.indexMethodGetEntries[0] != &indexGetEntries[0] {
		t.Fatalf("active index get cache backing slice was not preserved")
	}
	if len(vm.activeLookup.indexMethodSetEntries) != len(indexSetEntries) || &vm.activeLookup.indexMethodSetEntries[0] != &indexSetEntries[0] {
		t.Fatalf("active index set cache backing slice was not preserved")
	}
}

func TestBytecodeVM_SwitchRunProgramUsesFinalizedProgramMetadata(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	caller := &bytecodeProgram{instructions: []bytecodeInstruction{{op: bytecodeOpConst}}}
	target := finalizeBytecodeProgramMetadata(&bytecodeProgram{instructions: []bytecodeInstruction{
		{
			op:              bytecodeOpBinaryIntAddSlotConst,
			intImmediate:    runtime.NewSmallInt(11, runtime.IntegerI32),
			hasIntImmediate: true,
		},
		{op: bytecodeOpReturn},
	}})
	if target.slotConstIntImmTable == nil {
		t.Fatalf("expected finalized target to carry a slot-const immediate table")
	}

	program := caller
	instructions := caller.instructions
	validatedIntConsts := []bool{true}
	var slotConstIntImmTable *bytecodeSlotConstIntImmediateTable
	vm.switchRunProgram(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, target)

	if program != target || vm.currentProgram != target {
		t.Fatalf("program/current = %p/%p, want target %p", program, vm.currentProgram, target)
	}
	if len(instructions) != len(target.instructions) {
		t.Fatalf("instruction slice length = %d, want %d", len(instructions), len(target.instructions))
	}
	if validatedIntConsts != nil {
		t.Fatalf("finalized target without integer const op should set validation slots to nil")
	}
	if slotConstIntImmTable != target.slotConstIntImmTable {
		t.Fatalf("slot const immediate table = %p, want finalized table %p", slotConstIntImmTable, target.slotConstIntImmTable)
	}
	if vm.validatedIntConsts != nil {
		t.Fatalf("finalized metadata fast path should not allocate validation map")
	}
}

func TestBytecodeVM_SwitchRunProgramI32RegisterFrameActivatesAndReleases(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	plainProgram := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{slotCount: 1}}

	vm.switchRunProgramI32RegisterFrame(plainProgram)
	if vm.i32RegisterProgram != nil || len(vm.i32Registers) != 0 || len(vm.i32RegisterValid) != 0 {
		t.Fatalf("plain program without active frame should not activate i32 registers")
	}

	i32Program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{
		slotCount:        1,
		slotKinds:        []bytecodeCellKind{bytecodeCellKindI32},
		i32RegisterFrame: true,
	}}
	vm.slots = []runtime.Value{runtime.NewSmallInt(42, runtime.IntegerI32)}
	vm.switchRunProgramI32RegisterFrame(i32Program)
	if vm.i32RegisterProgram != i32Program {
		t.Fatalf("i32 register program = %p, want %p", vm.i32RegisterProgram, i32Program)
	}
	if raw, ok := vm.i32RegisterRaw(0); !ok || raw != 42 {
		t.Fatalf("active i32 raw = (%d, %t), want (42, true)", raw, ok)
	}

	vm.switchRunProgramI32RegisterFrame(plainProgram)
	if vm.i32RegisterProgram != nil || len(vm.i32Registers) != 0 || len(vm.i32RegisterValid) != 0 {
		t.Fatalf("switching to plain program should release active i32 registers")
	}
}

func TestBytecodeVM_SwitchRunProgramI32RegisterFrameIfNeededSkipsPlainInactive(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	plainProgram := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{slotCount: 1}}

	vm.switchRunProgramI32RegisterFrameIfNeeded(plainProgram)
	if vm.i32RegisterProgram != nil || len(vm.i32Registers) != 0 || len(vm.i32RegisterValid) != 0 {
		t.Fatalf("plain inactive switch should not activate i32 registers")
	}

	i32Program := &bytecodeProgram{frameLayout: &bytecodeFrameLayout{
		slotCount:        1,
		slotKinds:        []bytecodeCellKind{bytecodeCellKindI32},
		i32RegisterFrame: true,
	}}
	vm.slots = []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}
	vm.switchRunProgramI32RegisterFrameIfNeeded(i32Program)
	if vm.i32RegisterProgram != i32Program {
		t.Fatalf("i32 register program = %p, want %p", vm.i32RegisterProgram, i32Program)
	}
	if raw, ok := vm.i32RegisterRaw(0); !ok || raw != 7 {
		t.Fatalf("active i32 raw = (%d, %t), want (7, true)", raw, ok)
	}

	vm.switchRunProgramI32RegisterFrameIfNeeded(plainProgram)
	if vm.i32RegisterProgram != nil || len(vm.i32Registers) != 0 || len(vm.i32RegisterValid) != 0 {
		t.Fatalf("plain switch should release active i32 registers")
	}
}
