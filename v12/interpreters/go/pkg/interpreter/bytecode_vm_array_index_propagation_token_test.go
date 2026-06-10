package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ArrayGetPrimitiveNoErrorTokenCacheTracksMethodVersion(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	if !vm.arrayGetPrimitiveNoErrorToken(bytecodeIndexTypeChar) {
		t.Fatalf("expected fresh char primitive no-error cache hit")
	}

	interp.implMethods["char"] = []implEntry{{interfaceName: "Error"}}
	interp.invalidateMethodCache()
	if vm.arrayGetPrimitiveNoErrorToken(bytecodeIndexTypeChar) {
		t.Fatalf("expected char primitive no-error cache to observe method version invalidation")
	}
}

func TestBytecodeVM_CanSkipSuccessPropagationUsesPrimitiveTokenCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpIndexGet},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0

	if !vm.canSkipSuccessPropagation(runtime.CharValue{Val: 'a'}) {
		t.Fatalf("expected char propagation skip before Error impl is registered")
	}

	interp.implMethods["char"] = []implEntry{{interfaceName: "Error"}}
	interp.invalidateMethodCache()
	if vm.canSkipSuccessPropagation(runtime.CharValue{Val: 'a'}) {
		t.Fatalf("expected char propagation skip cache to observe method version invalidation")
	}
}

func TestBytecodeVM_CanSkipExactPrimitiveArrayGetSuccessPropagationTracksMethodVersion(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayIndexGetSlot},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0

	if !vm.canSkipExactPrimitiveArrayGetSuccessPropagation(runtime.CharValue{Val: 'a'}, bytecodeIndexTypeChar, true) {
		t.Fatalf("expected exact char array get propagation skip before Error impl is registered")
	}

	interp.implMethods["char"] = []implEntry{{interfaceName: "Error"}}
	interp.invalidateMethodCache()
	if vm.canSkipExactPrimitiveArrayGetSuccessPropagation(runtime.CharValue{Val: 'a'}, bytecodeIndexTypeChar, true) {
		t.Fatalf("expected exact char array get propagation skip cache to observe method version invalidation")
	}
}

func TestBytecodeVM_CanSkipSuccessPropagationUsesArrayNoErrorCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpIndexGet},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0
	arr := interp.newArrayValue([]runtime.Value{runtime.NewSmallInt(1, runtime.IntegerI32)}, 1)

	if !vm.canSkipSuccessPropagation(arr) {
		t.Fatalf("expected array propagation skip before Error impl is registered")
	}

	interp.implMethods["Array"] = []implEntry{{interfaceName: "Error"}}
	interp.invalidateMethodCache()
	if vm.canSkipSuccessPropagation(arr) {
		t.Fatalf("expected array propagation skip cache to observe method version invalidation")
	}
}

func TestBytecodeVM_ArrayIndexGetSlotDoesNotSkipPropagationWhenCharMayBeError(t *testing.T) {
	interp := NewBytecode()
	interp.implMethods["char"] = []implEntry{{interfaceName: "Error"}}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayIndexGetSlot},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(2, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("mono char array index slot propagation check failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after mono char array index slot with Error impl = %d, want propagation opcode at 1", vm.ip)
	}
}

func TestBytecodeVM_ArrayIndexGetSlotDoesNotSkipPropagationForOutOfBoundsMonoPrimitive(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b')
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayIndexGetSlot},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(8, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("out-of-bounds mono char array index slot propagation check failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after out-of-bounds mono char array index slot = %d, want propagation opcode at 1", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack after out-of-bounds mono char array index slot = %#v, want one error value", vm.stack)
	}
	if _, ok := vm.stack[0].(runtime.ErrorValue); !ok {
		t.Fatalf("stack after out-of-bounds mono char array index slot = %#v, want ErrorValue", vm.stack)
	}
}

func TestBytecodeVM_ArrayIndexGetSlotDoesNotSkipPropagationForStalePrimitiveToken(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	errValue := runtime.ErrorValue{Message: "bad"}
	arr := interp.newArrayValue([]runtime.Value{errValue}, 1)
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	state.ElementTypeToken = bytecodeIndexTypeChar
	state.ElementTypeTokenKnown = true
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpArrayIndexGetSlot},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0
	vm.slots = []runtime.Value{
		arr,
		runtime.NewSmallInt(0, runtime.IntegerI32),
	}
	instr := &bytecodeInstruction{
		op:        bytecodeOpArrayIndexGetSlot,
		argCount:  0,
		loopBreak: 1,
	}

	if err := vm.execArrayIndexGetSlot(instr); err != nil {
		t.Fatalf("stale-token array index slot propagation check failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after stale-token array index slot = %d, want propagation opcode at 1", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack after stale-token array index slot = %#v, want one error value", vm.stack)
	}
	got, ok := vm.stack[0].(runtime.ErrorValue)
	if !ok || got.Message != errValue.Message {
		t.Fatalf("stack after stale-token array index slot = %#v, want %#v", vm.stack, errValue)
	}
}
