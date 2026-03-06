package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVMExecStringInterpolationReusesPartsBuffer(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{op: bytecodeOpStringInterpolation, argCount: 2}

	vm.stack = append(vm.stack, runtime.StringValue{Val: "a"}, runtime.StringValue{Val: "b"})
	if err := vm.execStringInterpolation(instr); err != nil {
		t.Fatalf("execStringInterpolation first call failed: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("expected one interpolation result on stack, got %d", len(vm.stack))
	}
	firstPtr := &vm.stringInterpParts[0]
	for idx, val := range vm.stringInterpParts[:2] {
		if val != nil {
			t.Fatalf("expected cleared interpolation buffer slot %d, got %#v", idx, val)
		}
	}

	vm.stack = vm.stack[:0]
	vm.ip = 0
	vm.stack = append(vm.stack, runtime.StringValue{Val: "c"}, runtime.StringValue{Val: "d"})
	if err := vm.execStringInterpolation(instr); err != nil {
		t.Fatalf("execStringInterpolation second call failed: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("expected one interpolation result on second call, got %d", len(vm.stack))
	}
	if &vm.stringInterpParts[0] != firstPtr {
		t.Fatalf("expected interpolation parts buffer reuse across calls")
	}
	for idx, val := range vm.stringInterpParts[:2] {
		if val != nil {
			t.Fatalf("expected cleared interpolation buffer slot %d after second call, got %#v", idx, val)
		}
	}
}

func TestBytecodeVMExecArrayLiteralCopiesStackSegment(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	instr := &bytecodeInstruction{op: bytecodeOpArrayLiteral, argCount: 2}
	left := runtime.StringValue{Val: "x"}
	right := runtime.StringValue{Val: "y"}

	vm.stack = append(vm.stack, left, right)
	if err := vm.execArrayLiteral(instr); err != nil {
		t.Fatalf("execArrayLiteral failed: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("expected one array literal result on stack, got %d", len(vm.stack))
	}
	arr, ok := vm.stack[0].(*runtime.ArrayValue)
	if !ok || arr == nil {
		t.Fatalf("expected array literal result, got %#v", vm.stack[0])
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 array elements, got %d", len(arr.Elements))
	}
	if arr.Elements[0] != left || arr.Elements[1] != right {
		t.Fatalf("unexpected array literal elements: %#v", arr.Elements)
	}
}
