package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_ReplaceTopHelpers(t *testing.T) {
	vm := &bytecodeVM{
		stack: []runtime.Value{
			runtime.BoolValue{Val: true},
			runtime.NewSmallInt(2, runtime.IntegerI32),
			runtime.StringValue{Val: "x"},
			runtime.StringValue{Val: "y"},
		},
	}

	if err := vm.replaceTop1(nil); err != nil {
		t.Fatalf("replaceTop1 returned error: %v", err)
	}
	if len(vm.stack) != 4 {
		t.Fatalf("replaceTop1 changed stack depth: got=%d want=4", len(vm.stack))
	}
	if _, ok := vm.stack[3].(runtime.NilValue); !ok {
		t.Fatalf("replaceTop1 should normalize nil results, got %#v", vm.stack[3])
	}

	replaced := runtime.StringValue{Val: "done"}
	if err := vm.replaceTop2(replaced); err != nil {
		t.Fatalf("replaceTop2 returned error: %v", err)
	}
	if len(vm.stack) != 3 {
		t.Fatalf("replaceTop2 changed stack depth incorrectly: got=%d want=3", len(vm.stack))
	}
	if !valuesEqual(vm.stack[2], replaced) {
		t.Fatalf("replaceTop2 wrote wrong result: got=%#v want=%#v", vm.stack[2], replaced)
	}

	final := runtime.NewSmallInt(9, runtime.IntegerI32)
	if err := vm.replaceTop3(final); err != nil {
		t.Fatalf("replaceTop3 returned error: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("replaceTop3 changed stack depth incorrectly: got=%d want=1", len(vm.stack))
	}
	if !valuesEqual(vm.stack[0], final) {
		t.Fatalf("replaceTop3 wrote wrong result: got=%#v want=%#v", vm.stack[0], final)
	}
}
