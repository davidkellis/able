package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestInterpreterCoerceValueToTypePromotesDynamicArrayHandleToMonoU32(t *testing.T) {
	interp := New()
	handle := runtime.ArrayStoreNewReservedCapacity(4)
	arr, err := interp.arrayValueFromHandle(handle, 0, 4)
	if err != nil {
		t.Fatalf("arrayValueFromHandle: %v", err)
	}
	if arr.State == nil {
		t.Fatalf("expected dynamic handle-backed array view before coercion")
	}

	coerced, err := interp.coerceValueToType(ast.Gen(ast.Ty("Array"), ast.Ty("u32")), arr)
	if err != nil {
		t.Fatalf("coerceValueToType(Array u32): %v", err)
	}
	got, ok := coerced.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("coerceValueToType(Array u32) returned %T", coerced)
	}
	if got.Handle != handle {
		t.Fatalf("coerced handle = %d, want %d", got.Handle, handle)
	}
	if got.State != nil || got.Elements != nil {
		t.Fatalf("coerced mono u32 array should invalidate boxed value view")
	}
	typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoElementTypeNameIfKnown: %v", err)
	}
	if !ok || typeName != string(runtime.IntegerU32) {
		t.Fatalf("mono element type = (%q, %v), want (u32, true)", typeName, ok)
	}
	size, err := runtime.ArrayStoreSize(handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 0 {
		t.Fatalf("size = %d, want 0", size)
	}
	capacity, err := runtime.ArrayStoreCapacity(handle)
	if err != nil {
		t.Fatalf("ArrayStoreCapacity: %v", err)
	}
	if capacity != 4 {
		t.Fatalf("capacity = %d, want 4", capacity)
	}
}
