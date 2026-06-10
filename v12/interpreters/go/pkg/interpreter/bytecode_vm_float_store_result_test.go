package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_StoreReusableNormalizedFloatSlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		bytecodeRawFloatSlotValue(3.5, runtime.FloatF64),
	}
	cached := &runtime.FloatValue{Val: 9.25, TypeSuffix: runtime.FloatF64}
	vm.ownedFloatSlots = map[*runtime.Value]*runtime.FloatValue{
		&vm.slots[0]: cached,
	}

	stored := vm.storeReusableNormalizedFloatSlotRaw(0, 4.5, runtime.FloatF64)

	if stored != vm.slots[0] {
		t.Fatalf("stored value = %#v, want current slot value %#v", stored, vm.slots[0])
	}
	if _, ok := vm.slots[0].(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("slot after raw store = %#v, want visible raw f64 slot value", vm.slots[0])
	}
	assertFloatValue(t, vm.slots[0], runtime.FloatF64, 4.5)
	if cached.Val != 9.25 || cached.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("cached owned float cell mutated to %#v, want untouched stale cell", cached)
	}
}

func TestBytecodeVM_FinishStoreSlotFloatRawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	owned := &runtime.FloatValue{Val: 1.25, TypeSuffix: runtime.FloatF64}
	vm.slots = []runtime.Value{owned}

	instr := &bytecodeInstruction{target: 0}
	if err := vm.finishStoreSlotFloatRawResult(instr, 4.5, runtime.FloatF64); err != nil {
		t.Fatalf("finishStoreSlotFloatRawResult() error = %v", err)
	}
	if got := vm.slots[0]; got != owned {
		t.Fatalf("slot after store = %#v, want reused owned cell %#v", got, owned)
	}
	if owned.Val != 4.5 || owned.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("owned slot cell = %#v, want updated f64 4.5", owned)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack len = %d, want 1", len(vm.stack))
	}
	if _, ok := vm.stack[0].(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("stack result = %#v, want raw f64 carrier", vm.stack[0])
	}
	assertFloatValue(t, vm.stack[0], runtime.FloatF64, 4.5)
}

func TestBytecodeVM_FinishStoreSlotFloatRawResultDiscardKeepsStackEmpty(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		runtime.NilValue{},
		bytecodeRawFloatSlotValue(1.25, runtime.FloatF64),
	}

	instr := &bytecodeInstruction{target: 1, discardResult: true}
	if err := vm.finishStoreSlotFloatRawResult(instr, 4.5, runtime.FloatF64); err != nil {
		t.Fatalf("finishStoreSlotFloatRawResult() discard error = %v", err)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("discard stack len = %d, want 0", len(vm.stack))
	}
	if _, ok := vm.slots[1].(bytecodeRawF64SlotValue); !ok {
		t.Fatalf("slot after discard store = %#v, want raw f64 carrier", vm.slots[1])
	}
	assertFloatValue(t, vm.slots[1], runtime.FloatF64, 4.5)
}
