package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_StoreSlotBinaryIntSlotConstDiscardResultReusesOwnedI32SlotOutsideCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	start := int32(bytecodeRawI32SlotCacheMax)
	vm.slots = []runtime.Value{runtime.NewSmallInt(int64(start), runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "+",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		intImmediateRaw: 2,
		hasIntImmediate: true,
		hasIntRaw:       true,
		discardResult:   true,
	}
	if err := vm.execStoreSlotBinaryIntSlotConst(instr, nil); err != nil {
		t.Fatalf("unexpected first out-of-cache store error: %v", err)
	}
	firstCell, ok := vm.slots[0].(*runtime.IntegerValue)
	if !ok || firstCell == nil {
		t.Fatalf("first stored slot = %#v, want owned small-int cell", vm.slots[0])
	}
	if !firstCell.IsSmallRef() || firstCell.TypeSuffix != runtime.IntegerI32 || firstCell.Int64FastRef() != int64(start+2) {
		t.Fatalf("first owned small-int cell = %#v, want i32 %d", firstCell, start+2)
	}
	if err := vm.execLoadSlotOpcode(&bytecodeInstruction{op: bytecodeOpLoadSlot, target: 0}); err != nil {
		t.Fatalf("load owned i32 slot: %v", err)
	}
	if _, ok := vm.stack[0].(*runtime.IntegerValue); ok {
		t.Fatalf("visible LoadSlot should materialize owned i32 cell, got %#v", vm.stack[0])
	}
	loaded, ok := bytecodeDirectSmallI32Value(vm.stack[0])
	if !ok || loaded != int64(start+2) {
		t.Fatalf("loaded value = %#v, want materialized i32 %d", vm.stack[0], start+2)
	}
	vm.stack = nil
	if err := vm.execStoreSlotBinaryIntSlotConst(instr, nil); err != nil {
		t.Fatalf("unexpected second out-of-cache store error: %v", err)
	}
	secondCell, ok := vm.slots[0].(*runtime.IntegerValue)
	if !ok || secondCell == nil {
		t.Fatalf("second stored slot = %#v, want owned small-int cell", vm.slots[0])
	}
	if firstCell != secondCell {
		t.Fatalf("expected repeated discarded out-of-cache stores to reuse owned i32 slot cell")
	}
	if !secondCell.IsSmallRef() || secondCell.Int64FastRef() != int64(start+4) {
		t.Fatalf("second owned small-int cell = %#v, want i32 %d", secondCell, start+4)
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstDiscardResultUpdatesExistingOwnedI32SlotInPlaceWithoutMap(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	existing := &runtime.IntegerValue{}
	existing.ResetSmall(int64(bytecodeRawI32SlotCacheMax+2), runtime.IntegerI32)
	vm.slots = []runtime.Value{existing}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "+",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		intImmediateRaw: 2,
		hasIntImmediate: true,
		hasIntRaw:       true,
		discardResult:   true,
	}
	if err := vm.execStoreSlotBinaryIntSlotConst(instr, nil); err != nil {
		t.Fatalf("unexpected discarded out-of-cache store error: %v", err)
	}
	got, ok := vm.slots[0].(*runtime.IntegerValue)
	if !ok || got == nil {
		t.Fatalf("stored slot = %#v, want owned small-int cell", vm.slots[0])
	}
	if got != existing {
		t.Fatalf("expected discarded out-of-cache store to update existing owned i32 slot cell in place")
	}
	if !got.IsSmallRef() || got.TypeSuffix != runtime.IntegerI32 || got.Int64FastRef() != int64(bytecodeRawI32SlotCacheMax+4) {
		t.Fatalf("updated owned small-int cell = %#v, want i32 %d", got, bytecodeRawI32SlotCacheMax+4)
	}
	if vm.ownedI32Slots != nil {
		t.Fatalf("expected direct in-place owned i32 slot update to avoid allocating i32 slot map")
	}
}
