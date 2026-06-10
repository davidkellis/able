package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ArrayIndexGetSlotUnwrapsInterfaceArrayReceiver(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')
	vm.slots = []runtime.Value{
		&runtime.InterfaceValue{Underlying: arr},
		runtime.NewSmallInt(3, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("interface array index slot opcode failed: %v", err)
	}
	if got, ok := vm.stack[0].(runtime.CharValue); !ok || got.Val != 'e' {
		t.Fatalf("interface array index slot result = %#v, want char 'e'", vm.stack[0])
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("interface array index slot should not materialize boxed state")
	}
}
