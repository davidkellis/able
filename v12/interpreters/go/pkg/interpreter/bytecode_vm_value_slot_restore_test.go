package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_RestoreEmptyValueSlotI32FrameReleasesActiveSidecar(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	vm.slots = vm.acquireSlotFrame(1)
	if !vm.ensureActiveValueSlotI32Frame() {
		t.Fatalf("expected value-slot i32 frame activation to succeed")
	}
	if !vm.setActiveValueSlotI32Raw(0, 17) {
		t.Fatalf("expected value-slot i32 raw store to succeed")
	}

	vm.restoreValueSlotI32Frame(vm.slots, nil, nil)

	if _, ok := vm.activeValueSlotI32Raw(0); ok {
		t.Fatalf("expected empty restore to release active i32 sidecar")
	}
	if vm.slotI32Owner != nil || len(vm.slotI32Values) != 0 || len(vm.slotI32Valid) != 0 {
		t.Fatalf("expected no active i32 sidecar after empty restore")
	}
}

func TestBytecodeVM_RestoreEmptyValueSlotFloatFrameReleasesActiveSidecar(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	vm.slots = vm.acquireSlotFrame(1)
	if !vm.ensureActiveValueSlotFloatFrame() {
		t.Fatalf("expected value-slot float frame activation to succeed")
	}
	if !vm.setActiveValueSlotFloatRaw(0, 9.5, runtime.FloatF64) {
		t.Fatalf("expected value-slot float raw store to succeed")
	}

	vm.restoreValueSlotFloatFrame(vm.slots, nil, nil, nil)

	if _, _, ok := vm.activeValueSlotFloatRaw(0); ok {
		t.Fatalf("expected empty restore to release active float sidecar")
	}
	if vm.slotFloatOwner != nil || len(vm.slotFloatValues) != 0 || len(vm.slotFloatKinds) != 0 || len(vm.slotFloatValid) != 0 {
		t.Fatalf("expected no active float sidecar after empty restore")
	}
}

func TestBytecodeVM_RestoreValueSlotSidecarFramesReleasesActiveSidecars(t *testing.T) {
	vm := newBytecodeVM(nil, nil)
	vm.slots = vm.acquireSlotFrame(1)
	if !vm.ensureActiveValueSlotI32Frame() {
		t.Fatalf("expected value-slot i32 frame activation to succeed")
	}
	if !vm.setActiveValueSlotI32Raw(0, 23) {
		t.Fatalf("expected value-slot i32 raw store to succeed")
	}
	if !vm.ensureActiveValueSlotFloatFrame() {
		t.Fatalf("expected value-slot float frame activation to succeed")
	}
	if !vm.setActiveValueSlotFloatRaw(0, 4.25, runtime.FloatF64) {
		t.Fatalf("expected value-slot float raw store to succeed")
	}

	vm.restoreValueSlotSidecarFrames(vm.slots, nil, nil, nil, nil, nil)

	if _, ok := vm.activeValueSlotI32Raw(0); ok {
		t.Fatalf("expected empty sidecar restore to release active i32 sidecar")
	}
	if _, _, ok := vm.activeValueSlotFloatRaw(0); ok {
		t.Fatalf("expected empty sidecar restore to release active float sidecar")
	}
}
