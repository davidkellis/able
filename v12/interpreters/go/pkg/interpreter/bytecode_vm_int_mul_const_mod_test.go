package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodePositiveIntMulConstModFast(t *testing.T) {
	rem, handled, err := bytecodePositiveIntMulConstModFast(42, 48271, 2147483647)
	if err != nil {
		t.Fatalf("unexpected positive mul/mod error: %v", err)
	}
	if !handled {
		t.Fatalf("expected positive mul/mod fast path to handle exact positive operands")
	}
	if rem != 2027382 {
		t.Fatalf("remainder = %d, want 2027382", rem)
	}
}

func TestBytecodePositiveIntMulConstModFastRejectsNegativeBase(t *testing.T) {
	if _, handled, err := bytecodePositiveIntMulConstModFast(-1, 48271, 2147483647); err != nil || handled {
		t.Fatalf("negative base should fall back, handled=%v err=%v", handled, err)
	}
}

func TestBytecodeDirectPositiveSmallI64ImmediateValue(t *testing.T) {
	if raw, ok := bytecodeDirectPositiveSmallI64ImmediateValue(runtime.NewSmallInt(2147483647, runtime.IntegerI64)); !ok || raw != 2147483647 {
		t.Fatalf("value-form positive i64 immediate = %d, %v; want 2147483647, true", raw, ok)
	}
	ptr := runtime.NewSmallInt(48271, runtime.IntegerI64)
	if raw, ok := bytecodeDirectPositiveSmallI64ImmediateValue(&ptr); !ok || raw != 48271 {
		t.Fatalf("pointer-form positive i64 immediate = %d, %v; want 48271, true", raw, ok)
	}
	if _, ok := bytecodeDirectPositiveSmallI64ImmediateValue(runtime.NewSmallInt(0, runtime.IntegerI64)); ok {
		t.Fatalf("zero immediate should not take positive small-i64 fast path")
	}
}

func TestBytecodeStoreSlotIntMulConstModConstImmediateRaws(t *testing.T) {
	instr := &bytecodeInstruction{
		intImmediate:     runtime.NewSmallInt(48271, runtime.IntegerI64),
		intImmediate2:    runtime.NewSmallInt(2147483647, runtime.IntegerI64),
		intImmediateRaw:  48271,
		intImmediate2Raw: 2147483647,
		hasIntImmediate:  true,
		hasIntImmediate2: true,
		hasIntRaw:        true,
		hasIntRaw2:       true,
	}
	mulRaw, modRaw, ok := bytecodeStoreSlotIntMulConstModConstImmediateRaws(instr, instr.intImmediate, instr.intImmediate2)
	if !ok || mulRaw != 48271 || modRaw != 2147483647 {
		t.Fatalf("raw immediate decode = (%d, %d, %v), want (48271, 2147483647, true)", mulRaw, modRaw, ok)
	}
}

func TestBytecodeDirectPositiveSmallIntOfKind(t *testing.T) {
	if raw, ok := bytecodeDirectPositiveSmallIntOfKind(runtime.NewSmallInt(42, runtime.IntegerI64), runtime.IntegerI64); !ok || raw != 42 {
		t.Fatalf("value-form positive i64 decode = %d, %v; want 42, true", raw, ok)
	}
	if raw, ok := bytecodeDirectPositiveSmallIntOfKind(bytecodeRawI32SlotValue(7), runtime.IntegerI32); !ok || raw != 7 {
		t.Fatalf("raw-slot positive i32 decode = %d, %v; want 7, true", raw, ok)
	}
	if raw, ok := bytecodeDirectPositiveSmallIntOfKind(&bytecodeRawI64SlotCell{Val: 42}, runtime.IntegerI64); !ok || raw != 42 {
		t.Fatalf("raw-slot positive i64 decode = %d, %v; want 42, true", raw, ok)
	}
	if _, ok := bytecodeDirectPositiveSmallIntOfKind(runtime.NewSmallInt(-1, runtime.IntegerI64), runtime.IntegerI64); ok {
		t.Fatalf("negative value should not take positive small-int fast path")
	}
}

func TestBytecodeVM_StoreSlotIntMulConstModConstDiscardUsesRawI64Slot(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(42, runtime.IntegerI64)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotIntMulConstModConst,
		target:          0,
		value:           runtime.NewSmallInt(2147483647, runtime.IntegerI64),
		intImmediate:    runtime.NewSmallInt(48271, runtime.IntegerI64),
		hasIntImmediate: true,
		discardResult:   true,
	}
	if err := vm.execStoreSlotIntMulConstModConst(instr); err != nil {
		t.Fatalf("unexpected discarded mul/mod store error: %v", err)
	}
	cell, ok := vm.slots[0].(*bytecodeRawI64SlotCell)
	if !ok || cell == nil || cell.Val != 2027382 {
		t.Fatalf("stored slot = %#v, want raw i64 slot cell 2027382", vm.slots[0])
	}
	firstCell := cell
	read := bytecodeSlotReadValue(vm.slots[0])
	kind, raw, ok := bytecodeRawIntegerValueInfo(read)
	if !ok || kind != runtime.IntegerI64 || raw != 2027382 {
		t.Fatalf("raw slot read = %#v, want raw i64 2027382", read)
	}
	boxed, ok := bytecodeMaterializeRawValue(read).(runtime.IntegerValue)
	if !ok {
		t.Fatalf("materialized slot read = %#v, want integer value", bytecodeMaterializeRawValue(read))
	}
	if got, ok := boxed.ToInt64(); !ok || got != 2027382 {
		t.Fatalf("materialized slot read = %#v, want 2027382_i64", boxed)
	}
	if err := vm.execStoreSlotIntMulConstModConst(instr); err != nil {
		t.Fatalf("unexpected repeated discarded mul/mod store error: %v", err)
	}
	secondCell, ok := vm.slots[0].(*bytecodeRawI64SlotCell)
	if !ok || secondCell == nil || secondCell.Val != 1226992407 {
		t.Fatalf("reused slot = %#v, want raw i64 slot cell 1226992407", vm.slots[0])
	}
	if firstCell != secondCell {
		t.Fatalf("expected repeated discarded i64 recurrence stores to reuse raw slot cell")
	}
}

func TestBytecodeVM_StoreSlotIntMulConstModConstDiscardSteadyStateFastSupportsPointerModuloImmediate(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	sourceCell := &bytecodeRawI64SlotCell{Val: 42}
	vm.slots = []runtime.Value{sourceCell}
	modImm := runtime.NewSmallInt(2147483647, runtime.IntegerI64)
	instr := &bytecodeInstruction{
		op:               bytecodeOpStoreSlotIntMulConstModConst,
		target:           0,
		value:            &modImm,
		intImmediate:     runtime.NewSmallInt(48271, runtime.IntegerI64),
		intImmediate2:    runtime.NewSmallInt(2147483647, runtime.IntegerI64),
		intImmediateRaw:  48271,
		intImmediate2Raw: 2147483647,
		hasIntImmediate:  true,
		hasIntImmediate2: true,
		hasIntRaw:        true,
		hasIntRaw2:       true,
		discardResult:    true,
	}
	if err := vm.execStoreSlotIntMulConstModConst(instr); err != nil {
		t.Fatalf("unexpected steady-state discarded mul/mod store error: %v", err)
	}
	reusedCell, ok := vm.slots[0].(*bytecodeRawI64SlotCell)
	if !ok || reusedCell == nil || reusedCell.Val != 2027382 {
		t.Fatalf("stored slot = %#v, want raw i64 slot cell 2027382", vm.slots[0])
	}
	if reusedCell != sourceCell {
		t.Fatalf("expected steady-state discarded mul/mod store to reuse existing raw i64 cell")
	}
}
