package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ExecStoreSlotUntypedI32UsesValueSlotSidecar(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	vm.slots = vm.acquireSlotFrame(1)
	vm.prepareValueSlotI32Frame(nil)
	vm.stack = []runtime.Value{runtime.NewSmallInt(41, runtime.IntegerI32)}

	instr := &bytecodeInstruction{op: bytecodeOpStoreSlot, target: 0, storeRawI32Sidecar: true}
	if err := vm.execStoreSlot(instr); err != nil {
		t.Fatalf("execStoreSlot returned error: %v", err)
	}
	if vm.slots[0] != nil {
		t.Fatalf("untyped i32 store should keep value in sidecar, got slot %#v", vm.slots[0])
	}
	if raw, ok := vm.activeValueSlotI32Raw(0); !ok || raw != 41 {
		t.Fatalf("active value-slot raw = (%d, %t), want (41, true)", raw, ok)
	}
	if got, ok := vm.slotDirectSmallI32Value(0); !ok || got != 41 {
		t.Fatalf("slot direct raw = (%d, %t), want (41, true)", got, ok)
	}
	if raw, ok := bytecodeRawI32Value(vm.slotRuntimeValue(0)); !ok || raw != 41 {
		t.Fatalf("materialized slot raw = (%d, %t), want (41, true)", raw, ok)
	}
	if raw, ok := bytecodeRawI32Value(vm.stack[0]); !ok || raw != 41 {
		t.Fatalf("stack value raw = (%d, %t), want (41, true)", raw, ok)
	}
}

func TestBytecodeVM_CallFrameRestoresValueSlotI32Sidecar(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	callerSlots := vm.acquireSlotFrame(1)
	vm.slots = callerSlots
	vm.prepareValueSlotI32Frame(nil)
	if !vm.storeActiveValueSlotI32Raw(0, 33) {
		t.Fatalf("expected caller slot sidecar store to succeed")
	}

	vm.pushCallFrame(7, nil, vm.slots, nil, nil, 0, 0, false, false)
	if _, ok := vm.activeValueSlotI32Raw(0); ok {
		t.Fatalf("detaching the caller frame should clear the active sidecar")
	}

	calleeSlots := vm.acquireSlotFrame(1)
	vm.slots = calleeSlots
	vm.prepareValueSlotI32Frame(nil)

	returnIP, _, returnSlots, _, _, _, _, _, _, ok := vm.popCallFrameFields()
	if !ok {
		t.Fatalf("expected saved call frame")
	}
	if returnIP != 7 {
		t.Fatalf("return IP = %d, want 7", returnIP)
	}
	vm.slots = returnSlots
	if got, ok := vm.slotDirectSmallI32Value(0); !ok || got != 33 {
		t.Fatalf("restored caller raw = (%d, %t), want (33, true)", got, ok)
	}
}

func TestBytecodeVM_SelfFastFullCallFrameRestoresEnvironment(t *testing.T) {
	env := runtime.NewEnvironment(nil)
	vm := newBytecodeVM(nil, env)
	callerSlots := vm.acquireSlotFrame(1)
	vm.slots = callerSlots

	returnGenerics := map[string]struct{}{"T": {}}
	vm.pushCallFrame(9, nil, vm.slots, env, returnGenerics, 0, 0, false, true)

	returnIP, _, returnSlots, returnEnv, _, _, _, selfFast, _, ok := vm.popCallFrameFields()
	if !ok {
		t.Fatalf("expected saved self-fast frame")
	}
	if !selfFast {
		t.Fatalf("expected self-fast frame result")
	}
	if returnIP != 9 {
		t.Fatalf("return IP = %d, want 9", returnIP)
	}
	if returnEnv != env {
		t.Fatalf("return env = %p, want %p", returnEnv, env)
	}
	if len(returnSlots) == 0 || &returnSlots[0] != &callerSlots[0] {
		t.Fatalf("expected caller slots to be restored")
	}
}

func TestBytecodeVM_SelfFastSlot0FrameRestoresValueSlotI32Sidecar(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	vm.slots = vm.acquireSlotFrame2()
	vm.prepareValueSlotI32Frame(nil)
	if !vm.storeActiveValueSlotI32Raw(0, 12) {
		t.Fatalf("expected caller slot0 sidecar store to succeed")
	}
	vm.setSelfFastSlot0I32Raw(12)
	vm.slots[1] = runtime.StringValue{Val: "self"}

	if !vm.pushSelfFastSlot0CallFrameWithBases(5, 0, 0) {
		t.Fatalf("expected self-fast slot0 frame push to succeed")
	}
	if !vm.storeActiveValueSlotI32Raw(0, 44) {
		t.Fatalf("expected callee slot0 sidecar store to succeed")
	}

	returnIP, _, returnSlots, _, _, _, _, selfFast, _, ok := vm.popCallFrameFields()
	if !ok {
		t.Fatalf("expected saved self-fast frame")
	}
	if !selfFast {
		t.Fatalf("expected self-fast frame result")
	}
	if returnIP != 5 {
		t.Fatalf("return IP = %d, want 5", returnIP)
	}
	vm.slots = returnSlots
	if got, ok := vm.slotDirectSmallI32Value(0); !ok || got != 12 {
		t.Fatalf("restored slot0 raw = (%d, %t), want (12, true)", got, ok)
	}
}
