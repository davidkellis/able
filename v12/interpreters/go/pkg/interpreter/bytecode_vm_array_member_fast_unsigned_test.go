package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func monoU32ArrayValueForTest(t *testing.T, values ...uint32) *runtime.ArrayValue {
	t.Helper()
	handle := runtime.ArrayStoreMonoNewWithCapacityU32(len(values))
	for idx, value := range values {
		if err := runtime.ArrayStoreMonoWriteU32(handle, idx, value); err != nil {
			t.Fatalf("write mono u32 value %d: %v", idx, err)
		}
	}
	return &runtime.ArrayValue{Handle: handle, TrackedHandle: handle}
}

func monoU64ArrayValueForTest(t *testing.T, values ...uint64) *runtime.ArrayValue {
	t.Helper()
	handle := runtime.ArrayStoreMonoNewWithCapacityU64(len(values))
	for idx, value := range values {
		if err := runtime.ArrayStoreMonoWriteU64(handle, idx, value); err != nil {
			t.Fatalf("write mono u64 value %d: %v", idx, err)
		}
	}
	return &runtime.ArrayValue{Handle: handle, TrackedHandle: handle}
}

func TestBytecodeVM_ArrayPushMemberFastPromotesRawU8WithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue(nil, 1)
	vm.stack = []runtime.Value{arr, bytecodeRawU8ResultValue(23)}

	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayPush,
		bytecodeInstruction{name: "push", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("raw u8 array push fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw u8 array push fast path to handle call")
	}
	raw, ok, err := runtime.ArrayStoreMonoReadU8IfAvailable(arr.Handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadU8IfAvailable: %v", err)
	}
	if !ok || raw != 23 {
		t.Fatalf("mono u8 push read = (%d, %v), want (23, true)", raw, ok)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono u8 push should not materialize boxed state")
	}
}

func TestBytecodeVM_ArrayMemberFastPathMonoU32GetSkipsBoxedState(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoU32ArrayValueForTest(t, 7, 11)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}

	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayGet,
		bytecodeInstruction{name: "get", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("mono u32 array get fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono u32 array get fast path to handle call")
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerU32 || raw != 11 {
		t.Fatalf("mono u32 get numeric value = %#v, want 11_u32", vm.stack[0])
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono u32 array get should not materialize boxed state")
	}
}

func TestBytecodeVM_ArrayReadSlotMemberFastReadsMonoU32Handle(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoU32ArrayValueForTest(t, 13, 17)

	got, mode, handled, err := vm.readArraySlotValueFast(arr, runtime.NewSmallInt(0, runtime.IntegerI32))
	if err != nil {
		t.Fatalf("mono u32 read_slot fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono u32 read_slot fast path to handle call")
	}
	if mode != "array_read_slot_mono_u32_fast" {
		t.Fatalf("mono u32 read_slot mode = %q, want array_read_slot_mono_u32_fast", mode)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(got)
	if !ok || kind != runtime.IntegerU32 || raw != 13 {
		t.Fatalf("mono u32 read_slot numeric value = %#v, want 13_u32", got)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono u32 read_slot should not materialize boxed state")
	}
}

func TestBytecodeVM_ArrayPushMemberFastAppendsMonoU32WithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoU32ArrayValueForTest(t)
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(23, runtime.IntegerU32)}

	_, handled, err := vm.execCachedMemberMethodFastPath(
		bytecodeMemberMethodFastPathArrayPush,
		bytecodeInstruction{name: "push", argCount: 1},
		0,
		1,
		nil,
	)
	if err != nil {
		t.Fatalf("mono u32 array push fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono u32 array push fast path to handle call")
	}
	if _, ok := vm.stack[0].(runtime.VoidValue); !ok {
		t.Fatalf("push result = %#v, want void", vm.stack[0])
	}
	raw, ok, err := runtime.ArrayStoreMonoReadU32IfAvailable(arr.Handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadU32IfAvailable: %v", err)
	}
	if !ok || raw != 23 {
		t.Fatalf("mono u32 push read = (%d, %v), want (23, true)", raw, ok)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono u32 push should not materialize boxed state")
	}
}

func TestBytecodeVM_ArrayWriteSlotFastPreservesMonoU32HandleWithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoU32ArrayValueForTest(t, 5, 7)

	mode, handled, err := vm.writeArraySlotValueFast(
		arr,
		runtime.NewSmallInt(1, runtime.IntegerI32),
		bytecodeRawIntegerResultValue(runtime.IntegerU32, 29),
	)
	if err != nil {
		t.Fatalf("mono u32 write_slot fast path failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected mono u32 write_slot fast path to handle call")
	}
	if mode != "array_write_slot_fast" {
		t.Fatalf("mono u32 write_slot mode = %q, want array_write_slot_fast", mode)
	}
	raw, ok, err := runtime.ArrayStoreMonoReadU32IfAvailable(arr.Handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadU32IfAvailable: %v", err)
	}
	if !ok || raw != 29 {
		t.Fatalf("mono u32 write_slot read = (%d, %v), want (29, true)", raw, ok)
	}
	typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(arr.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoElementTypeNameIfKnown: %v", err)
	}
	if !ok || typeName != string(runtime.IntegerU32) {
		t.Fatalf("mono u32 write_slot type = (%q, %v), want (u32, true)", typeName, ok)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono u32 write_slot should not materialize boxed state")
	}
}

func TestBytecodeVM_DirectArrayIndexSetPreservesMonoU32HandleWithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoU32ArrayValueForTest(t, 3, 9)

	result, err := vm.resolveDirectArrayIndexSetAt(arr, 0, bytecodeRawIntegerResultValue(runtime.IntegerU32, 41))
	if err != nil {
		t.Fatalf("mono u32 direct index set failed: %v", err)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(result)
	if !ok || kind != runtime.IntegerU32 || raw != 41 {
		t.Fatalf("mono u32 direct index set result = %#v, want 41_u32", result)
	}
	stored, ok, err := runtime.ArrayStoreMonoReadU32IfAvailable(arr.Handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadU32IfAvailable: %v", err)
	}
	if !ok || stored != 41 {
		t.Fatalf("mono u32 direct index set read = (%d, %v), want (41, true)", stored, ok)
	}
	typeName, ok, err := runtime.ArrayStoreMonoElementTypeNameIfKnown(arr.Handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoElementTypeNameIfKnown: %v", err)
	}
	if !ok || typeName != string(runtime.IntegerU32) {
		t.Fatalf("mono u32 direct index set type = (%q, %v), want (u32, true)", typeName, ok)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono u32 direct index set should not materialize boxed state")
	}
}

func TestBytecodeVM_ExactNativeArrayReadFastReadsMonoU64WithoutMaterializingState(t *testing.T) {
	interp := NewBytecode()
	const largeValue = uint64(1)<<63 + 25
	arr := monoU64ArrayValueForTest(t, 4, largeValue)
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	result, handled, err := vm.tryExecExactNativeArrayReadFast("__able_array_read", []runtime.Value{
		runtime.NewSmallInt(arr.Handle, runtime.IntegerI64),
		runtime.NewSmallInt(1, runtime.IntegerI32),
	})
	if err != nil {
		t.Fatalf("exact native array read fast failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected exact native array read fast path to handle call")
	}
	intVal, ok := result.(runtime.IntegerValue)
	if !ok || intVal.TypeSuffix != runtime.IntegerU64 {
		t.Fatalf("exact native array read result = %#v, want u64", result)
	}
	if intVal.BigInt().Cmp(new(big.Int).SetUint64(largeValue)) != 0 {
		t.Fatalf("exact native array read numeric value = %#v, want %d_u64", result, largeValue)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("exact native array read should not materialize boxed state")
	}
}
