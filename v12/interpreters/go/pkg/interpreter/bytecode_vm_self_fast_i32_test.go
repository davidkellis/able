package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMSelfFastSlot0RawLaneRestoresCallerSlot(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	self := &runtime.FunctionValue{}
	vm.slots = []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32), self}
	vm.setSelfFastSlot0I32Raw(7)
	if !vm.pushSelfFastSlot0CallFrame(11) {
		t.Fatalf("expected compact slot0 frame push to succeed")
	}
	if vm.selfFastSlot0I32Valid {
		t.Fatalf("expected active raw slot0 lane to clear until the callee writes slot0")
	}
	vm.slots[0] = runtime.NewSmallInt(6, runtime.IntegerI32)
	vm.setSelfFastSlot0I32Raw(6)

	returnIP, _, returnSlots, _, _, _, _, _, ok := vm.popCallFrameFields()
	if !ok {
		t.Fatalf("expected compact self-fast frame to pop")
	}
	if returnIP != 11 {
		t.Fatalf("expected returnIP 11, got %d", returnIP)
	}
	if len(returnSlots) == 0 || &returnSlots[0] != &vm.slots[0] {
		t.Fatalf("expected compact frame to reuse current slot slice")
	}
	if got, ok := bytecodeRawI32Value(vm.slots[0]); !ok || got != 7 {
		t.Fatalf("expected caller slot0 to be restored to 7, got %d ok=%v", got, ok)
	}
	if !vm.selfFastSlot0I32Valid || vm.selfFastSlot0I32Raw != 7 {
		t.Fatalf("expected caller raw slot0 lane to restore to 7, valid=%v raw=%d", vm.selfFastSlot0I32Valid, vm.selfFastSlot0I32Raw)
	}
}

func TestBytecodeVMFusedSelfCallSlot0RawLaneTracksCallee(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	vm := newBytecodeVM(interp, env)

	layout := &bytecodeFrameLayout{
		slotCount:          2,
		paramSlots:         1,
		selfCallSlot:       1,
		selfCallOneArgFast: true,
	}
	program := &bytecodeProgram{
		frameLayout:              layout,
		returnGenericNamesCached: true,
	}
	self := &runtime.FunctionValue{Closure: env, Bytecode: program}
	vm.slots = vm.acquireSlotFrame2()
	vm.slots[0] = runtime.NewSmallInt(10, runtime.IntegerI32)
	vm.slots[1] = self
	vm.setSelfFastSlot0I32Raw(10)

	instr := &bytecodeInstruction{
		op:              bytecodeOpCallSelfIntSubSlotConst,
		target:          1,
		argCount:        0,
		intImmediate:    runtime.NewSmallInt(1, runtime.IntegerI32),
		intImmediateRaw: 1,
		hasIntImmediate: true,
		hasIntRaw:       true,
	}
	newProgram, err := vm.execCallSelfIntSubSlotConst(instr, nil, program)
	if err != nil {
		t.Fatalf("fused self-call failed: %v", err)
	}
	if newProgram != program {
		t.Fatalf("expected fused self-call to stay on current program")
	}
	if !vm.selfFastSlot0I32Valid || vm.selfFastSlot0I32Raw != 9 {
		t.Fatalf("expected active raw slot0 lane to track callee value 9, valid=%v raw=%d", vm.selfFastSlot0I32Valid, vm.selfFastSlot0I32Raw)
	}
	if len(vm.selfFastMinimal) != 1 || !vm.selfFastMinimal[0].slot0I32Valid || vm.selfFastMinimal[0].slot0I32Raw != 10 {
		t.Fatalf("expected compact frame to save caller raw slot0 10, frame=%#v", vm.selfFastMinimal)
	}
}

func TestBytecodeVMReturnConstIfSlot0UsesRawLane(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	returnVal := runtime.NewSmallInt(1, runtime.IntegerI32)
	vm.slots = []runtime.Value{runtime.NewSmallInt(2, runtime.IntegerI32)}
	vm.setSelfFastSlot0I32Raw(2)
	instr := &bytecodeInstruction{
		op:              bytecodeOpReturnConstIfIntLessEqualSlotConst,
		argCount:        0,
		value:           returnVal,
		intImmediate:    runtime.NewSmallInt(2, runtime.IntegerI32),
		intImmediateRaw: 2,
		hasIntImmediate: true,
		hasIntRaw:       true,
	}

	got, returned, err := vm.execReturnConstIfIntLessEqualSlotConst(instr, nil)
	if err != nil {
		t.Fatalf("unexpected return-const-if raw-lane error: %v", err)
	}
	if !returned || !valuesEqual(got, returnVal) {
		t.Fatalf("expected raw-lane return, got=%#v returned=%v", got, returned)
	}

	vm.ip = 7
	vm.setSelfFastSlot0I32Raw(3)
	got, returned, err = vm.execReturnConstIfIntLessEqualSlotConst(instr, nil)
	if err != nil {
		t.Fatalf("unexpected false return-const-if raw-lane error: %v", err)
	}
	if returned || got != nil {
		t.Fatalf("expected raw-lane false branch not to return, got=%#v returned=%v", got, returned)
	}
	if vm.ip != 8 {
		t.Fatalf("expected raw-lane false branch to advance ip, got %d", vm.ip)
	}
}

func TestBytecodeVMSlot0StoreRefreshesRawLane(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.slots = []runtime.Value{runtime.NewSmallInt(2, runtime.IntegerI32)}
	vm.stack = []runtime.Value{runtime.NewSmallInt(5, runtime.IntegerI32)}
	vm.setSelfFastSlot0I32Raw(2)

	if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlot, target: 0}); err != nil {
		t.Fatalf("slot0 store failed: %v", err)
	}
	if !vm.selfFastSlot0I32Valid || vm.selfFastSlot0I32Raw != 5 {
		t.Fatalf("expected slot0 i32 store to refresh raw lane to 5, valid=%v raw=%d", vm.selfFastSlot0I32Valid, vm.selfFastSlot0I32Raw)
	}

	vm.stack = []runtime.Value{runtime.StringValue{Val: "x"}}
	if err := vm.execStoreSlot(&bytecodeInstruction{op: bytecodeOpStoreSlot, target: 0}); err != nil {
		t.Fatalf("slot0 string store failed: %v", err)
	}
	if vm.selfFastSlot0I32Valid {
		t.Fatalf("expected non-i32 slot0 store to clear raw lane")
	}
}
