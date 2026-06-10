package interpreter

import (
	"math"
	"strings"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_StoreSlotBinaryIntSlotConstSubtractFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(4, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "-",
		intImmediate:    runtime.NewSmallInt(3, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	if err := vm.execStoreSlotBinaryIntSlotConst(instr, nil); err != nil {
		t.Fatalf("unexpected subtract store-slot fast-path error: %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != 1 {
		t.Fatalf("stored slot = %#v, want small i32 1", vm.slots[0])
	}
	stackGot, ok := bytecodeDirectSmallI32Value(vm.stack[0])
	if !ok || stackGot != 1 {
		t.Fatalf("stack result = %#v, want small i32 1", vm.stack[0])
	}
	if !vm.selfFastSlot0I32Valid || vm.selfFastSlot0I32Raw != 1 {
		t.Fatalf("expected slot0 raw lane to refresh to 1, valid=%v raw=%d", vm.selfFastSlot0I32Valid, vm.selfFastSlot0I32Raw)
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstFastPathOverflow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(math.MaxInt32, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "+",
		intImmediate:    runtime.NewSmallInt(1, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	err := vm.execStoreSlotBinaryIntSlotConst(instr, nil)
	if err == nil || !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("expected integer overflow, got %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != math.MaxInt32 {
		t.Fatalf("overflow should leave slot unchanged, got %#v", vm.slots[0])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("overflow should not push assignment result, got len=%d", len(vm.stack))
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstSubtractFastPathOverflow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(math.MinInt32, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "-",
		intImmediate:    runtime.NewSmallInt(1, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	err := vm.execStoreSlotBinaryIntSlotConst(instr, nil)
	if err == nil || !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("expected integer overflow, got %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != math.MinInt32 {
		t.Fatalf("overflow should leave slot unchanged, got %#v", vm.slots[0])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("overflow should not push assignment result, got len=%d", len(vm.stack))
	}
}

func TestBytecodeVM_StoreSlotBinaryIntSlotConstMultiplyFastPathOverflow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(math.MaxInt32, runtime.IntegerI32)}
	instr := &bytecodeInstruction{
		op:              bytecodeOpStoreSlotBinaryIntSlotConst,
		target:          0,
		operator:        "*",
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		hasIntImmediate: true,
	}
	err := vm.execStoreSlotBinaryIntSlotConst(instr, nil)
	if err == nil || !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("expected integer overflow, got %v", err)
	}
	got, ok := bytecodeDirectSmallI32Value(vm.slots[0])
	if !ok || got != math.MaxInt32 {
		t.Fatalf("overflow should leave slot unchanged, got %#v", vm.slots[0])
	}
	if len(vm.stack) != 0 {
		t.Fatalf("overflow should not push assignment result, got len=%d", len(vm.stack))
	}
}
