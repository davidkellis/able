package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ArrayMemberFastPathMonoF64GetSkipsBoxedState(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	arr := monoF64ArrayValueForTest(t, 1.25, 2.5)
	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}

	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayGet,
		bytecodeInstruction{name: "get", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("mono f64 array get fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono f64 array get fast path to handle call")
	}
	assertFloatValue(t, vm.stack[0], runtime.FloatF64, 2.5)
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono f64 array get should not materialize boxed state")
	}

	vm = newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(3, runtime.IntegerI32)}
	_, handled, err = vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayGet,
		bytecodeInstruction{name: "get", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("mono f64 out-of-bounds get fast path failed: %v", err)
	}
	if !handled || !isNilRuntimeValue(vm.stack[0]) {
		t.Fatalf("mono f64 out-of-bounds get result = %#v, want nil", vm.stack[0])
	}
}
