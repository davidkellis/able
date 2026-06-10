package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMProgramReturnNoCoercionFastPathMaterializesKnownSimple(t *testing.T) {
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:        ast.Ty("i32"),
			returnSimpleType:  "i32",
			returnSimpleCheck: bytecodeSimpleTypeCheckI32,
		},
	}

	got, ok := bytecodeTryMaterializedProgramReturnNoCoercion(nil, nil, program, nil, bytecodeRawI32SlotCachedValue(42), bytecodeSimpleTypeCheckI32)
	if !ok {
		t.Fatalf("expected known i32 return to use no-coercion fast path")
	}
	assertIntValue(t, got, runtime.IntegerI32, 42)
}

func TestBytecodeVMProgramReturnNoCoercionFastPathRejectsWidening(t *testing.T) {
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:        ast.Ty("u64"),
			returnSimpleType:  "u64",
			returnSimpleCheck: bytecodeSimpleTypeCheckU64,
		},
	}

	got, ok := bytecodeTryMaterializedProgramReturnNoCoercion(nil, nil, program, nil, bytecodeRawI32SlotCachedValue(42), bytecodeSimpleTypeCheckI32)
	if ok {
		t.Fatalf("expected i32 to u64 return to stay on coercion path")
	}
	assertIntValue(t, got, runtime.IntegerI32, 42)
}

func TestBytecodeVMProgramReturnNoCoercionFastPathRejectsGenericSimple(t *testing.T) {
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:             ast.Ty("T"),
			returnSimpleType:       "T",
			returnTypeUsesGenerics: true,
		},
	}

	if _, ok := bytecodeTryMaterializedProgramReturnNoCoercion(nil, NewBytecode(), program, nil, runtime.NewSmallInt(42, runtime.IntegerI32), bytecodeSimpleTypeCheckUnknown); ok {
		t.Fatalf("expected generic simple return to stay on generic coercion path")
	}
}

func TestBytecodeVMProgramReturnNoCoercionFastPathAcceptsNamedSimple(t *testing.T) {
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnType:       ast.Ty("String"),
			returnSimpleType: "String",
		},
	}

	got, ok := bytecodeTryMaterializedProgramReturnNoCoercion(nil, NewBytecode(), program, nil, runtime.StringValue{Val: "ok"}, bytecodeSimpleTypeCheckUnknown)
	if !ok {
		t.Fatalf("expected named simple return to use no-coercion fast path")
	}
	if value, ok := got.(runtime.StringValue); !ok || value.Val != "ok" {
		t.Fatalf("named simple return = %#v, want String ok", got)
	}
}

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

func TestBytecodeVMMinimalSelfFastReturnNoCoerceRestoresAmbientControlStacks(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{
		frameLayout: &bytecodeFrameLayout{
			returnSimpleCheck: bytecodeSimpleTypeCheckI32,
			slotKinds:         []bytecodeCellKind{bytecodeCellKindI32},
		},
	}

	vm.slots = []runtime.Value{runtime.NewSmallInt(10, runtime.IntegerI32)}
	vm.setSelfFastSlot0I32Raw(10)
	vm.iterStack = []forLoopIterator{{}, {}}
	vm.loopStack = []bytecodeLoopFrame{{}, {}}
	if !vm.pushSelfFastSlot0CallFrameWithBases(13, 1, 1) {
		t.Fatalf("expected compact self-fast frame push to succeed")
	}
	vm.slots[0] = runtime.NewSmallInt(9, runtime.IntegerI32)
	vm.setSelfFastSlot0I32Raw(9)
	vm.iterStack = append(vm.iterStack, forLoopIterator{})
	vm.loopStack = append(vm.loopStack, bytecodeLoopFrame{})

	returnVal := runtime.NewSmallInt(1, runtime.IntegerI32)
	instr := &bytecodeInstruction{op: bytecodeOpReturnConstIfIntLessEqualSlotConst}
	if !vm.tryFinishMinimalSelfFastReturnNoCoerce(program, instr, returnVal, bytecodeSimpleTypeCheckUnknown) {
		t.Fatalf("expected compact i32 return to finish directly")
	}
	if got, want := len(vm.iterStack), 1; got != want {
		t.Fatalf("expected iter stack depth %d after return, got %d", want, got)
	}
	if got, want := len(vm.loopStack), 1; got != want {
		t.Fatalf("expected loop stack depth %d after return, got %d", want, got)
	}
}
