package runtime

import (
	"math/big"
	"testing"
)

func TestArrayStoreDynamicCapacityGrowthAmortized(t *testing.T) {
	handle := ArrayStoreNew()
	for idx := 0; idx <= 64; idx++ {
		if err := ArrayStoreWrite(handle, idx, BoolValue{Val: true}); err != nil {
			t.Fatalf("ArrayStoreWrite(%d): %v", idx, err)
		}
	}
	size, err := ArrayStoreSize(handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 65 {
		t.Fatalf("expected size 65, got %d", size)
	}
	capacity, err := ArrayStoreCapacity(handle)
	if err != nil {
		t.Fatalf("ArrayStoreCapacity: %v", err)
	}
	if capacity <= size {
		t.Fatalf("expected amortized growth capacity > size (size=%d capacity=%d)", size, capacity)
	}
}

func TestArrayStoreDynamicSparseWritePreservesNilGap(t *testing.T) {
	handle := ArrayStoreNew()
	if err := ArrayStoreWrite(handle, 2, BoolValue{Val: true}); err != nil {
		t.Fatalf("ArrayStoreWrite sparse: %v", err)
	}
	size, err := ArrayStoreSize(handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 3 {
		t.Fatalf("expected size 3 after sparse write, got %d", size)
	}
	first, err := ArrayStoreRead(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreRead(0): %v", err)
	}
	if _, ok := first.(NilValue); !ok {
		t.Fatalf("expected nil gap at index 0, got %#v", first)
	}
	second, err := ArrayStoreRead(handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreRead(1): %v", err)
	}
	if _, ok := second.(NilValue); !ok {
		t.Fatalf("expected nil gap at index 1, got %#v", second)
	}
	third, err := ArrayStoreRead(handle, 2)
	if err != nil {
		t.Fatalf("ArrayStoreRead(2): %v", err)
	}
	boolVal, ok := third.(BoolValue)
	if !ok || !boolVal.Val {
		t.Fatalf("expected bool true at index 2, got %#v", third)
	}
}

func TestArrayStoreReservedCapacityAllocatesDynamicBackingOnWrite(t *testing.T) {
	handle := ArrayStoreNewReservedCapacity(8)
	state := arrayStates[handle]
	if state == nil {
		t.Fatalf("reserved handle has no dynamic state")
	}
	if state.Capacity != 8 || cap(state.Values) != 0 {
		t.Fatalf("reserved state capacity=%d backing=%d, want capacity 8 backing 0", state.Capacity, cap(state.Values))
	}
	if err := ArrayStoreWrite(handle, 0, BoolValue{Val: true}); err != nil {
		t.Fatalf("ArrayStoreWrite: %v", err)
	}
	if state.Capacity != 8 || cap(state.Values) != 8 {
		t.Fatalf("write should allocate reserved backing, capacity=%d backing=%d", state.Capacity, cap(state.Values))
	}
}

func TestArrayStoreMonoBoolRoundTripAndDynamicFallback(t *testing.T) {
	handle := ArrayStoreMonoNewBool()
	if err := ArrayStoreMonoWriteBool(handle, 0, true); err != nil {
		t.Fatalf("ArrayStoreMonoWriteBool: %v", err)
	}
	if err := ArrayStoreMonoWriteBool(handle, 3, false); err != nil {
		t.Fatalf("ArrayStoreMonoWriteBool sparse extend: %v", err)
	}
	size, err := ArrayStoreSize(handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 4 {
		t.Fatalf("expected size 4, got %d", size)
	}

	value, err := ArrayStoreRead(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreRead: %v", err)
	}
	boolVal, ok := value.(BoolValue)
	if !ok || !boolVal.Val {
		t.Fatalf("expected bool true from generic read, got %#v", value)
	}

	if err := ArrayStoreWrite(handle, 0, BoolValue{Val: false}); err != nil {
		t.Fatalf("ArrayStoreWrite on mono handle: %v", err)
	}
	typedValue, err := ArrayStoreMonoReadBool(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadBool: %v", err)
	}
	if typedValue {
		t.Fatalf("expected typed bool false after generic write")
	}

	state, err := ArrayStoreState(handle)
	if err != nil {
		t.Fatalf("ArrayStoreState deopt: %v", err)
	}
	if len(state.Values) != 4 {
		t.Fatalf("expected deopt state length 4, got %d", len(state.Values))
	}

	if err := ArrayStoreMonoWriteBool(handle, 1, true); err != nil {
		t.Fatalf("ArrayStoreMonoWriteBool after deopt: %v", err)
	}
	value, err = ArrayStoreRead(handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreRead after deopt write: %v", err)
	}
	boolVal, ok = value.(BoolValue)
	if !ok || !boolVal.Val {
		t.Fatalf("expected bool true after deopt write, got %#v", value)
	}
}

func TestArrayStoreMonoI64RoundTripAndDynamicFallback(t *testing.T) {
	handle := ArrayStoreMonoNewI64()
	if err := ArrayStoreMonoWriteI64(handle, 0, 42); err != nil {
		t.Fatalf("ArrayStoreMonoWriteI64: %v", err)
	}
	if err := ArrayStoreMonoWriteI64(handle, 2, -9); err != nil {
		t.Fatalf("ArrayStoreMonoWriteI64 sparse extend: %v", err)
	}
	size, err := ArrayStoreSize(handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 3 {
		t.Fatalf("expected size 3, got %d", size)
	}

	value, err := ArrayStoreRead(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreRead: %v", err)
	}
	intVal, ok := value.(IntegerValue)
	if n, nOk := intVal.ToInt64(); !ok || !nOk || n != 42 {
		t.Fatalf("expected integer 42 from generic read, got %#v", value)
	}

	if err := ArrayStoreWrite(handle, 0, IntegerValue{Val: mustBigInt(t, 100), TypeSuffix: IntegerI64}); err != nil {
		t.Fatalf("ArrayStoreWrite on mono i64 handle: %v", err)
	}
	typedValue, err := ArrayStoreMonoReadI64(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreMonoReadI64: %v", err)
	}
	if typedValue != 100 {
		t.Fatalf("expected typed i64 100 after generic write, got %d", typedValue)
	}

	state, err := ArrayStoreState(handle)
	if err != nil {
		t.Fatalf("ArrayStoreState deopt: %v", err)
	}
	if len(state.Values) != 3 {
		t.Fatalf("expected deopt state length 3, got %d", len(state.Values))
	}

	if err := ArrayStoreMonoWriteI64(handle, 1, 77); err != nil {
		t.Fatalf("ArrayStoreMonoWriteI64 after deopt: %v", err)
	}
	value, err = ArrayStoreRead(handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreRead after deopt write: %v", err)
	}
	intVal, ok = value.(IntegerValue)
	if n, nOk := intVal.ToInt64(); !ok || !nOk || n != 77 {
		t.Fatalf("expected integer 77 after deopt write, got %#v", value)
	}
}

func TestArrayStoreMonoF64PromoteUsesReservedCapacity(t *testing.T) {
	handle := ArrayStoreNewReservedCapacity(4)
	ok, err := ArrayStoreAppendF64Promote(handle, 1.5)
	if err != nil {
		t.Fatalf("ArrayStoreAppendF64Promote: %v", err)
	}
	if !ok {
		t.Fatalf("expected reserved dynamic array to promote to mono f64")
	}
	values, mono, err := ArrayStoreMonoF64ValuesIfAvailable(handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	if !mono || len(values) != 1 || cap(values) != 4 || values[0] != 1.5 {
		t.Fatalf("mono f64 values=%#v len=%d cap=%d mono=%v, want [1.5] cap 4", values, len(values), cap(values), mono)
	}
	capacity, err := ArrayStoreCapacity(handle)
	if err != nil {
		t.Fatalf("ArrayStoreCapacity: %v", err)
	}
	if capacity != 4 {
		t.Fatalf("mono f64 capacity = %d, want 4", capacity)
	}
}

func TestArrayStoreMonoF64PromoteAppendRoundTripAndDynamicFallback(t *testing.T) {
	handle := ArrayStoreNewWithCapacity(4)
	ok, err := ArrayStoreAppendF64Promote(handle, 1.5)
	if err != nil {
		t.Fatalf("ArrayStoreAppendF64Promote: %v", err)
	}
	if !ok {
		t.Fatalf("expected dynamic array to promote to mono f64")
	}
	ok, err = ArrayStoreAppendF64Promote(handle, 2.5)
	if err != nil {
		t.Fatalf("ArrayStoreAppendF64Promote second: %v", err)
	}
	if !ok {
		t.Fatalf("expected mono f64 append to stay handled")
	}
	values, mono, err := ArrayStoreMonoF64ValuesIfAvailable(handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	if !mono || len(values) != 2 || values[0] != 1.5 || values[1] != 2.5 {
		t.Fatalf("mono f64 values = %#v mono=%v, want [1.5 2.5]", values, mono)
	}
	size, err := ArrayStoreSize(handle)
	if err != nil {
		t.Fatalf("ArrayStoreSize: %v", err)
	}
	if size != 2 {
		t.Fatalf("mono f64 size = %d, want 2", size)
	}
	capacity, err := ArrayStoreCapacity(handle)
	if err != nil {
		t.Fatalf("ArrayStoreCapacity: %v", err)
	}
	if capacity != 4 {
		t.Fatalf("mono f64 capacity = %d, want 4", capacity)
	}
	value, err := ArrayStoreRead(handle, 1)
	if err != nil {
		t.Fatalf("ArrayStoreRead: %v", err)
	}
	floatVal, ok := value.(FloatValue)
	if !ok || floatVal.TypeSuffix != FloatF64 || floatVal.Val != 2.5 {
		t.Fatalf("ArrayStoreRead mono f64 = %#v, want f64 2.5", value)
	}

	state, err := ArrayStoreState(handle)
	if err != nil {
		t.Fatalf("ArrayStoreState deopt: %v", err)
	}
	if len(state.Values) != 2 {
		t.Fatalf("deopt f64 state length = %d, want 2", len(state.Values))
	}
	if err := ArrayStoreMonoWriteF64(handle, 0, 4.5); err != nil {
		t.Fatalf("ArrayStoreMonoWriteF64 after deopt: %v", err)
	}
	value, err = ArrayStoreRead(handle, 0)
	if err != nil {
		t.Fatalf("ArrayStoreRead after deopt write: %v", err)
	}
	floatVal, ok = value.(FloatValue)
	if !ok || floatVal.TypeSuffix != FloatF64 || floatVal.Val != 4.5 {
		t.Fatalf("ArrayStoreRead after deopt write = %#v, want f64 4.5", value)
	}
}

func TestArrayStoreMonoF64PromoteBulkAppend(t *testing.T) {
	handle := ArrayStoreNewWithCapacity(4)
	ok, err := ArrayStoreAppendF64ValuesPromote(handle, []float64{1.5, 2.5, 3.5})
	if err != nil {
		t.Fatalf("ArrayStoreAppendF64ValuesPromote: %v", err)
	}
	if !ok {
		t.Fatalf("expected dynamic array to bulk-promote to mono f64")
	}
	values, mono, err := ArrayStoreMonoF64ValuesIfAvailable(handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable: %v", err)
	}
	if !mono || len(values) != 3 || cap(values) != 4 || values[0] != 1.5 || values[1] != 2.5 || values[2] != 3.5 {
		t.Fatalf("mono f64 values=%#v len=%d cap=%d mono=%v, want [1.5 2.5 3.5] cap 4", values, len(values), cap(values), mono)
	}
	ok, err = ArrayStoreAppendF64ValuesPromote(handle, []float64{4.5, 5.5})
	if err != nil {
		t.Fatalf("ArrayStoreAppendF64ValuesPromote second: %v", err)
	}
	if !ok {
		t.Fatalf("expected mono f64 bulk append to stay handled")
	}
	values, mono, err = ArrayStoreMonoF64ValuesIfAvailable(handle)
	if err != nil {
		t.Fatalf("ArrayStoreMonoF64ValuesIfAvailable second: %v", err)
	}
	if !mono || len(values) != 5 || values[3] != 4.5 || values[4] != 5.5 {
		t.Fatalf("mono f64 values after second append=%#v mono=%v, want suffix [4.5 5.5]", values, mono)
	}
}

func mustBigInt(t *testing.T, value int64) *big.Int {
	t.Helper()
	return big.NewInt(value)
}
