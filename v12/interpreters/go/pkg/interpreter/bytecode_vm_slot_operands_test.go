package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_SlotDirectFloatValueCoversRawOwnedAndActiveFrame(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	vm.slots = []runtime.Value{
		bytecodeRawFloatSlotValue(3.5, runtime.FloatF64),
		&runtime.FloatValue{Val: 1.25, TypeSuffix: runtime.FloatF32},
		nil,
	}
	if !vm.ensureActiveValueSlotFloatFrame() {
		t.Fatalf("expected active float frame")
	}
	if !vm.setActiveValueSlotFloatRaw(2, 9.5, runtime.FloatF64) {
		t.Fatalf("expected active float slot write")
	}

	rawVal, rawKind, ok := vm.slotDirectFloatValue(0)
	if !ok || rawKind != runtime.FloatF64 || rawVal != 3.5 {
		t.Fatalf("raw slot direct float = (%v, %v, %v), want (3.5, f64, true)", rawVal, rawKind, ok)
	}

	ownedVal, ownedKind, ok := vm.slotDirectFloatValue(1)
	if !ok || ownedKind != runtime.FloatF32 || ownedVal != 1.25 {
		t.Fatalf("owned slot direct float = (%v, %v, %v), want (1.25, f32, true)", ownedVal, ownedKind, ok)
	}

	activeVal, activeKind, ok := vm.slotDirectFloatValue(2)
	if !ok || activeKind != runtime.FloatF64 || activeVal != 9.5 {
		t.Fatalf("active slot direct float = (%v, %v, %v), want (9.5, f64, true)", activeVal, activeKind, ok)
	}
}

func TestBytecodeVM_SlotDirectF64ValueRejectsNonF64(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{
		runtime.FloatValue{Val: 2.5, TypeSuffix: runtime.FloatF32},
		runtime.FloatValue{Val: 4.5, TypeSuffix: runtime.FloatF64},
		nil,
	}
	if !vm.ensureActiveValueSlotFloatFrame() {
		t.Fatalf("expected active float frame")
	}
	if !vm.setActiveValueSlotFloatRaw(2, 7.5, runtime.FloatF32) {
		t.Fatalf("expected active float slot write")
	}

	if got, ok := vm.slotDirectF64Value(0); ok || got != 2.5 {
		t.Fatalf("slotDirectF64Value(f32 boxed) = (%v, %v), want (2.5, false)", got, ok)
	}
	if got, ok := vm.slotDirectF64Value(1); !ok || got != 4.5 {
		t.Fatalf("slotDirectF64Value(f64 boxed) = (%v, %v), want (4.5, true)", got, ok)
	}
	if got, ok := vm.slotDirectF64Value(2); ok || got != 7.5 {
		t.Fatalf("slotDirectF64Value(f32 active) = (%v, %v), want (7.5, false)", got, ok)
	}
}
