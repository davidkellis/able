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

func mustBigInt(t *testing.T, value int64) *big.Int {
	t.Helper()
	return big.NewInt(value)
}
