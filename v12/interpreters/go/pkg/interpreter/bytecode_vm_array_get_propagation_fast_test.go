package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ArrayGetF64SuccessSkipsFollowingPropagation(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 2.5, TypeSuffix: runtime.FloatF64},
	}, 1)
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeF64 {
		t.Fatalf("array element token = known %v token %d, want f64", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}

	program := &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpCallMemberArrayGet, name: "get", argCount: 1},
		{op: bytecodeOpPropagation},
	}}
	vm.currentProgram = program
	vm.ip = 0
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(0, runtime.IntegerI32)}

	_, handled, err := vm.finishArrayGetMemberFast(program.instructions[0], arr, 0, 0, nil)
	if err != nil {
		t.Fatalf("finish Array.get fast: %v", err)
	}
	if !handled {
		t.Fatalf("expected Array.get fast path to handle call")
	}
	if vm.ip != 2 {
		t.Fatalf("ip after fused Array.get propagation = %d, want 2", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack length = %d, want 1", len(vm.stack))
	}
	got, ok := vm.stack[0].(runtime.FloatValue)
	if !ok || got.Val != 2.5 || got.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("Array.get result = %#v, want f64 2.5", vm.stack[0])
	}
}

func TestBytecodeVM_ArrayGetF64NilKeepsPropagationOpcode(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{}, 0)
	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}
	state.ElementTypeToken = bytecodeIndexTypeF64
	state.ElementTypeTokenKnown = true

	program := &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpCallMemberArrayGet, name: "get", argCount: 1},
		{op: bytecodeOpPropagation},
	}}
	vm.currentProgram = program
	vm.ip = 0
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(0, runtime.IntegerI32)}

	_, handled, err := vm.finishArrayGetMemberFast(program.instructions[0], arr, 0, 0, nil)
	if err != nil {
		t.Fatalf("finish Array.get fast: %v", err)
	}
	if !handled {
		t.Fatalf("expected Array.get fast path to handle call")
	}
	if vm.ip != 1 {
		t.Fatalf("ip after nil Array.get = %d, want propagation opcode at 1", vm.ip)
	}
	if len(vm.stack) != 1 || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("stack after nil Array.get = %#v, want nil", vm.stack)
	}
}

func TestBytecodeVM_ArrayGetF64DoesNotSkipPropagationWhenF64MayBeError(t *testing.T) {
	interp := NewBytecode()
	interp.implMethods["f64"] = []implEntry{{interfaceName: "Error"}}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 7.5, TypeSuffix: runtime.FloatF64},
	}, 1)

	program := &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpCallMemberArrayGet, name: "get", argCount: 1},
		{op: bytecodeOpPropagation},
	}}
	vm.currentProgram = program
	vm.ip = 0
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(0, runtime.IntegerI32)}

	_, handled, err := vm.finishArrayGetMemberFast(program.instructions[0], arr, 0, 0, nil)
	if err != nil {
		t.Fatalf("finish Array.get fast: %v", err)
	}
	if !handled {
		t.Fatalf("expected Array.get fast path to handle call")
	}
	if vm.ip != 1 {
		t.Fatalf("ip with f64 Error impl = %d, want propagation opcode at 1", vm.ip)
	}
}

func TestBytecodeVM_ArrayGetF64TokenDoesNotSkipNonFloatResult(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue([]runtime.Value{
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64},
		runtime.ErrorValue{Message: "bad"},
	}, 2)

	program := &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpCallMemberArrayGet, name: "get", argCount: 1},
		{op: bytecodeOpPropagation},
	}}
	vm.currentProgram = program
	vm.ip = 0
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}

	_, handled, err := vm.finishArrayGetMemberFast(program.instructions[0], arr, 1, 0, nil)
	if err != nil {
		t.Fatalf("finish Array.get fast: %v", err)
	}
	if !handled {
		t.Fatalf("expected Array.get fast path to handle call")
	}
	if vm.ip != 1 {
		t.Fatalf("ip with non-float result under f64 token = %d, want propagation opcode at 1", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack length = %d, want 1", len(vm.stack))
	}
	if _, ok := vm.stack[0].(runtime.ErrorValue); !ok {
		t.Fatalf("stack result = %#v, want ErrorValue", vm.stack[0])
	}
}

func TestBytecodeVM_ArrayGetPrimitiveNoErrorCacheTracksMethodVersion(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	if !vm.arrayGetPrimitiveNoError("f64") {
		t.Fatalf("expected fresh f64 primitive no-error cache hit")
	}

	interp.implMethods["f64"] = []implEntry{{interfaceName: "Error"}}
	interp.invalidateMethodCache()
	if vm.arrayGetPrimitiveNoError("f64") {
		t.Fatalf("expected f64 primitive no-error cache to observe method version invalidation")
	}
}
