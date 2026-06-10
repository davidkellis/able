package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeStackSnapshotValueCopiesSmallIntegerPointerAsValue(t *testing.T) {
	src := runtime.NewSmallInt(42, runtime.IntegerI32)
	got := bytecodeStackSnapshotValue(&src)

	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("snapshot type = %T, want runtime.IntegerValue", got)
	}
	if !intVal.IsSmall() || intVal.Int64Fast() != 42 || intVal.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("snapshot = %#v, want small i32 42", intVal)
	}
}

func TestBytecodeStackSnapshotValueReusesCachedSmallIntegerPointerValue(t *testing.T) {
	src := runtime.NewSmallInt(42, runtime.IntegerI32)
	want, ok := bytecodeBoxedIntegerValue(runtime.IntegerI32, 42)
	if !ok {
		t.Fatalf("expected boxed i32 cache entry")
	}

	got := bytecodeStackSnapshotValue(&src)
	if got != want {
		t.Fatalf("snapshot identity = %#v, want cached boxed value %#v", got, want)
	}
}

func TestBytecodeStackSnapshotValueCopiesFloatPointerAsValue(t *testing.T) {
	src := runtime.FloatValue{Val: 3.5, TypeSuffix: runtime.FloatF64}
	got := bytecodeStackSnapshotValue(&src)

	floatVal, ok := got.(runtime.FloatValue)
	if !ok {
		t.Fatalf("snapshot type = %T, want runtime.FloatValue", got)
	}
	if floatVal.Val != 3.5 || floatVal.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("snapshot = %#v, want f64 3.5", floatVal)
	}
}

func TestBytecodeStackSnapshotValueClonesBigIntegerPointer(t *testing.T) {
	src := runtime.NewBigIntValue(big.NewInt(1234), runtime.IntegerI128)
	got := bytecodeStackSnapshotValue(&src)

	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("snapshot type = %T, want runtime.IntegerValue", got)
	}
	src.Val.SetInt64(9999)
	if intVal.BigInt().Int64() != 1234 {
		t.Fatalf("snapshot big-int changed with source mutation: got=%d want=1234", intVal.BigInt().Int64())
	}
}

func TestBytecodeStackSnapshotValueCopiesRawIntegerSlotCellAsRawValue(t *testing.T) {
	src := &bytecodeRawIntegerSlotCell{Raw: 55, TypeSuffix: runtime.IntegerU32}
	got := bytecodeStackSnapshotValue(src)

	kind, raw, ok := bytecodeRawIntegerValueInfo(got)
	if !ok {
		t.Fatalf("snapshot type = %T, want raw integer carrier", got)
	}
	if kind != runtime.IntegerU32 || raw != 55 {
		t.Fatalf("snapshot = %#v, want raw u32 55", got)
	}
}

func TestBytecodeStackSnapshotValueReusesRawI32ValueIdentity(t *testing.T) {
	src := bytecodeRawI32SlotCachedValue(55)
	got := bytecodeStackSnapshotValue(src)

	if got != src {
		t.Fatalf("snapshot raw i32 identity = %#v, want %#v", got, src)
	}
}

func TestBytecodeStackSnapshotValueReusesRawIntegerValueIdentity(t *testing.T) {
	src := runtime.Value(bytecodeRawIntegerValue{Raw: 55, TypeSuffix: runtime.IntegerU32})
	got := bytecodeStackSnapshotValue(src)

	if got != src {
		t.Fatalf("snapshot raw integer identity = %#v, want %#v", got, src)
	}
}

func TestBytecodeStackSnapshotValueRawI32HotPathIsAllocationFree(t *testing.T) {
	src := bytecodeRawI32SlotCachedValue(55)

	allocs := testing.AllocsPerRun(1000, func() {
		got := bytecodeStackSnapshotValue(src)
		if got != src {
			panic("unexpected raw i32 snapshot result")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected raw i32 snapshot hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestBytecodeStackSnapshotValueSmallIntegerPointerHotPathIsAllocationFree(t *testing.T) {
	src := runtime.NewSmallInt(42, runtime.IntegerI32)

	allocs := testing.AllocsPerRun(1000, func() {
		got := bytecodeStackSnapshotValue(&src)
		if kind, raw, ok := bytecodeRawIntegerValueInfo(got); ok {
			if kind != runtime.IntegerI32 || raw != 42 {
				panic("unexpected raw snapshot result")
			}
			return
		}
		intVal, ok := got.(runtime.IntegerValue)
		if !ok || !intVal.IsSmall() || intVal.Int64Fast() != 42 || intVal.TypeSuffix != runtime.IntegerI32 {
			panic("unexpected small integer snapshot result")
		}
	})
	if allocs != 0 {
		t.Fatalf("expected small integer pointer snapshot hot path allocations to be zero, got %.2f", allocs)
	}
}

func TestBytecodeVMAppendSlotStackValueCheckedCopiesIntegerPointer(t *testing.T) {
	vm := &bytecodeVM{slots: make([]runtime.Value, 1)}
	src := runtime.NewBigIntValue(big.NewInt(1234), runtime.IntegerI128)
	vm.slots[0] = &src

	vm.appendSlotStackValueChecked(0)
	src.Val.SetInt64(9999)

	intVal, ok := vm.stack[0].(runtime.IntegerValue)
	if !ok {
		t.Fatalf("append snapshot type = %T, want runtime.IntegerValue", vm.stack[0])
	}
	if intVal.BigInt().Int64() != 1234 {
		t.Fatalf("append snapshot big-int changed with source mutation: got=%d want=1234", intVal.BigInt().Int64())
	}
}

func TestBytecodeVMAppendSlotStackValueCheckedPreservesStructPointer(t *testing.T) {
	vm := &bytecodeVM{slots: make([]runtime.Value, 1)}
	inst := &runtime.StructInstanceValue{Fields: map[string]runtime.Value{"x": runtime.NewSmallInt(1, runtime.IntegerI32)}}
	vm.slots[0] = inst

	vm.appendSlotStackValueChecked(0)

	if len(vm.stack) != 1 || vm.stack[0] != inst {
		t.Fatalf("append struct snapshot = %#v, want original pointer %#v", vm.stack, inst)
	}
}

func TestBytecodeVMSlotStackValueCheckedCopiesIntegerPointer(t *testing.T) {
	vm := &bytecodeVM{slots: make([]runtime.Value, 1)}
	src := runtime.NewBigIntValue(big.NewInt(1234), runtime.IntegerI128)
	vm.slots[0] = &src

	got := vm.slotStackValueChecked(0)
	src.Val.SetInt64(9999)

	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("slot stack snapshot type = %T, want runtime.IntegerValue", got)
	}
	if intVal.BigInt().Int64() != 1234 {
		t.Fatalf("slot stack snapshot big-int changed with source mutation: got=%d want=1234", intVal.BigInt().Int64())
	}
}

func TestBytecodeVMSlotStackValueCheckedPreservesStructPointer(t *testing.T) {
	vm := &bytecodeVM{slots: make([]runtime.Value, 1)}
	inst := &runtime.StructInstanceValue{Fields: map[string]runtime.Value{"x": runtime.NewSmallInt(1, runtime.IntegerI32)}}
	vm.slots[0] = inst

	got := vm.slotStackValueChecked(0)

	if got != inst {
		t.Fatalf("slot stack struct snapshot = %#v, want original pointer %#v", got, inst)
	}
}

func TestBytecodeSlotReadValueMaterializesRawFloatAndPreservesRawInteger(t *testing.T) {
	floatVal := bytecodeSlotReadValue(bytecodeRawF64SlotValue(3.5))
	boxedFloat, ok := floatVal.(runtime.FloatValue)
	if !ok {
		t.Fatalf("slot read float type = %T, want runtime.FloatValue", floatVal)
	}
	if boxedFloat.Val != 3.5 || boxedFloat.TypeSuffix != runtime.FloatF64 {
		t.Fatalf("slot read float = %#v, want f64 3.5", boxedFloat)
	}

	intVal := bytecodeSlotReadValue(bytecodeRawIntegerValue{Raw: 7, TypeSuffix: runtime.IntegerU16})
	rawInt, ok := intVal.(bytecodeRawIntegerValue)
	if !ok {
		t.Fatalf("slot read integer type = %T, want bytecodeRawIntegerValue", intVal)
	}
	if rawInt.Raw != 7 || rawInt.TypeSuffix != runtime.IntegerU16 {
		t.Fatalf("slot read integer = %#v, want raw u16 7", rawInt)
	}
}
