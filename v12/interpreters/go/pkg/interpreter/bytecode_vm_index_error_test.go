package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CanonicalIndexSetMethodFastPathReturnsIndexErrorValue(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b')

	got, err := vm.resolveIndexSet(arr, runtime.NewSmallInt(9, runtime.IntegerI32), runtime.CharValue{Val: 'z'}, ast.AssignmentAssign, "", false)
	if err != nil {
		t.Fatalf("resolveIndexSet canonical fast path failed: %v", err)
	}
	errVal, ok := got.(runtime.ErrorValue)
	if !ok {
		t.Fatalf("resolveIndexSet out-of-range result = %#v, want IndexError value", got)
	}
	payload, ok := errVal.Payload["value"].(*runtime.StructInstanceValue)
	if !ok || payload.Definition == nil || payload.Definition.Node == nil || payload.Definition.Node.ID == nil || payload.Definition.Node.ID.Name != "IndexError" {
		t.Fatalf("resolveIndexSet out-of-range payload = %#v, want IndexError payload", errVal.Payload)
	}
}
