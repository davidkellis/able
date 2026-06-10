package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ArrayIndexGetSlotUsesCanonicalIndexCache(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	if interp.canUseDirectArrayIndexGetFastPath() {
		t.Fatalf("expected stdlib bootstrap to install canonical Array Index impl")
	}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{instructions: []bytecodeInstruction{{op: bytecodeOpArrayIndexGetSlot}}}
	vm.currentProgram = program
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')

	seed, err := vm.resolveIndexGet(arr, runtime.NewSmallInt(2, runtime.IntegerI32))
	if err != nil {
		t.Fatalf("seed canonical index cache: %v", err)
	}
	if charVal, ok := seed.(runtime.CharValue); !ok || charVal.Val != 'l' {
		t.Fatalf("seed canonical index result = %#v, want char 'l'", seed)
	}

	interp.bytecodeStatsEnabled = true
	interp.ResetBytecodeStats()
	vm.stack = nil
	vm.slots = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}
	vm.ip = 0
	if err := vm.execArrayIndexGetSlot(&bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}); err != nil {
		t.Fatalf("array index slot canonical cache path failed: %v", err)
	}
	if got, ok := vm.stack[0].(runtime.CharValue); !ok || got.Val != 'b' {
		t.Fatalf("array index slot canonical cache result = %#v, want char 'b'", vm.stack[0])
	}
	stats := interp.BytecodeStats()
	if stats.ArrayIndexSlotLookups != 1 || stats.ArrayIndexSlotDirectHits != 1 || stats.ArrayIndexSlotFallbacks != 0 {
		t.Fatalf("array index slot stats lookups/direct/fallback = %d/%d/%d, want 1/1/0", stats.ArrayIndexSlotLookups, stats.ArrayIndexSlotDirectHits, stats.ArrayIndexSlotFallbacks)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("array index slot canonical cache path should not materialize boxed state")
	}
}
