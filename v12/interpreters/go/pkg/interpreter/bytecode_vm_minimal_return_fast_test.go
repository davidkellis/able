package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMMinimalSelfFastReturnNoCoerceRestoresCompactSlot0(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnSimpleCheck: bytecodeSimpleTypeCheckI32,
			slotKinds:         []bytecodeCellKind{bytecodeCellKindI32, bytecodeCellKindValue},
		},
	}

	vm.slots = []runtime.Value{
		runtime.NewSmallInt(10, runtime.IntegerI32),
		&runtime.FunctionValue{},
	}
	vm.setSelfFastSlot0I32Raw(10)
	if !vm.pushSelfFastSlot0CallFrame(7) {
		t.Fatalf("expected compact self-fast frame push to succeed")
	}
	vm.slots[0] = runtime.NewSmallInt(9, runtime.IntegerI32)
	vm.setSelfFastSlot0I32Raw(9)

	returnVal := runtime.NewSmallInt(1, runtime.IntegerI32)
	instr := &bytecodeInstruction{op: bytecodeOpReturnConstIfIntLessEqualSlotConst}
	if !vm.tryFinishMinimalSelfFastReturnNoCoerce(program, instr, returnVal, bytecodeSimpleTypeCheckUnknown) {
		t.Fatalf("expected compact i32 return to finish directly")
	}
	if vm.ip != 7 {
		t.Fatalf("expected return ip 7, got %d", vm.ip)
	}
	if len(vm.stack) != 1 || !valuesEqual(vm.stack[0], returnVal) {
		t.Fatalf("expected return value on stack, got %#v", vm.stack)
	}
	if got, ok := bytecodeRawI32Value(vm.slots[0]); !ok || got != 10 {
		t.Fatalf("expected caller slot0 restored to 10, got %d ok=%v", got, ok)
	}
	if !vm.selfFastSlot0I32Valid || vm.selfFastSlot0I32Raw != 10 {
		t.Fatalf("expected raw slot0 lane restored to 10, valid=%v raw=%d", vm.selfFastSlot0I32Valid, vm.selfFastSlot0I32Raw)
	}
	if vm.selfFastMinimalSuffix != 0 || len(vm.selfFastMinimal) != 0 {
		t.Fatalf("expected compact frame stack to be empty, suffix=%d frames=%d", vm.selfFastMinimalSuffix, len(vm.selfFastMinimal))
	}
}

func TestBytecodeVMMinimalSelfFastReturnNoCoerceRejectsGenericInt(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnSimpleCheck: bytecodeSimpleTypeCheckAnyInteger,
			slotKinds:         []bytecodeCellKind{bytecodeCellKindI32},
		},
	}

	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerI32)}
	vm.setSelfFastSlot0I32Raw(10)
	if !vm.pushSelfFastSlot0CallFrame(11) {
		t.Fatalf("expected compact self-fast frame push to succeed")
	}

	instr := &bytecodeInstruction{op: bytecodeOpReturnBinaryIntAddI32}
	if vm.tryFinishMinimalSelfFastReturnNoCoerce(program, instr, runtime.NewSmallInt(1, runtime.IntegerI32), bytecodeSimpleTypeCheckI32) {
		t.Fatalf("expected generic Int return to stay on normal coercion path")
	}
	if vm.selfFastMinimalSuffix != 1 || len(vm.selfFastMinimal) != 1 {
		t.Fatalf("expected compact frame to remain untouched, suffix=%d frames=%d", vm.selfFastMinimalSuffix, len(vm.selfFastMinimal))
	}
}
