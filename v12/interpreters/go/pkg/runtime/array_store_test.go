package runtime

import "testing"

func TestArrayStateBoxedLengthUsesSharedMetadataCache(t *testing.T) {
	values := make([]Value, 20000)
	allocs := testing.AllocsPerRun(1000, func() {
		state := ArrayState{Values: values, Capacity: len(values)}
		got := state.BoxedLengthValue()
		intVal, ok := got.(IntegerValue)
		if !ok {
			t.Fatalf("BoxedLengthValue type = %T, want IntegerValue", got)
		}
		if intVal.Int64Fast() != int64(len(values)) || intVal.TypeSuffix != IntegerI32 {
			t.Fatalf("BoxedLengthValue = (%d, %s), want (%d, %s)", intVal.Int64Fast(), intVal.TypeSuffix, len(values), IntegerI32)
		}
	})
	if allocs > 0.1 {
		t.Fatalf("BoxedLengthValue allocs = %.2f, want <= 0.1", allocs)
	}
}

func TestArrayStateBoxedCapacityUsesSharedMetadataCache(t *testing.T) {
	values := make([]Value, 10)
	allocs := testing.AllocsPerRun(1000, func() {
		state := ArrayState{Values: values, Capacity: 20000}
		got := state.BoxedCapacityValue()
		intVal, ok := got.(IntegerValue)
		if !ok {
			t.Fatalf("BoxedCapacityValue type = %T, want IntegerValue", got)
		}
		if intVal.Int64Fast() != 20000 || intVal.TypeSuffix != IntegerI32 {
			t.Fatalf("BoxedCapacityValue = (%d, %s), want (%d, %s)", intVal.Int64Fast(), intVal.TypeSuffix, 20000, IntegerI32)
		}
	})
	if allocs > 0.1 {
		t.Fatalf("BoxedCapacityValue allocs = %.2f, want <= 0.1", allocs)
	}
}

func TestBoxedArrayMetadataU64ValueUsesSharedCache(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		got, ok := BoxedArrayMetadataU64Value(20000)
		if !ok {
			t.Fatalf("BoxedArrayMetadataU64Value should cache 20000")
		}
		intVal, ok := got.(IntegerValue)
		if !ok {
			t.Fatalf("BoxedArrayMetadataU64Value type = %T, want IntegerValue", got)
		}
		if intVal.Int64Fast() != 20000 || intVal.TypeSuffix != IntegerU64 {
			t.Fatalf("BoxedArrayMetadataU64Value = (%d, %s), want (%d, %s)", intVal.Int64Fast(), intVal.TypeSuffix, 20000, IntegerU64)
		}
	})
	if allocs > 0.1 {
		t.Fatalf("BoxedArrayMetadataU64Value allocs = %.2f, want <= 0.1", allocs)
	}
}

func TestAdjustArrayI32CacheLengthShrinkPreservesKnownEntries(t *testing.T) {
	state := ArrayState{
		CachedI32Values:      []int32{1, 2, 3},
		CachedI32ValuesValid: []bool{true, true, false},
		CachedI32ValuesCount: 2,
		CachedI32ValuesKnown: false,
	}

	adjustArrayI32CacheLength(&state, 3, 2)

	if len(state.CachedI32Values) != 2 || len(state.CachedI32ValuesValid) != 2 {
		t.Fatalf("shrunk cache lengths = (%d, %d), want (2, 2)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if state.CachedI32ValuesCount != 2 || !state.CachedI32ValuesKnown {
		t.Fatalf("shrunk cache count/known = (%d, %v), want (2, true)", state.CachedI32ValuesCount, state.CachedI32ValuesKnown)
	}
	if state.CachedI32Values[0] != 1 || state.CachedI32Values[1] != 2 {
		t.Fatalf("shrunk cache values = %v, want [1 2]", state.CachedI32Values)
	}
}

func TestAdjustArrayI32CacheLengthGrowAppendsInvalidSlots(t *testing.T) {
	values := make([]int32, 2, 4)
	values[0], values[1] = 4, 5
	valid := make([]bool, 2, 4)
	valid[0], valid[1] = true, true
	state := ArrayState{
		CachedI32Values:      values,
		CachedI32ValuesValid: valid,
		CachedI32ValuesCount: 2,
		CachedI32ValuesKnown: true,
	}

	adjustArrayI32CacheLength(&state, 2, 4)

	if len(state.CachedI32Values) != 4 || len(state.CachedI32ValuesValid) != 4 {
		t.Fatalf("grown cache lengths = (%d, %d), want (4, 4)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if state.CachedI32ValuesCount != 2 || state.CachedI32ValuesKnown {
		t.Fatalf("grown cache count/known = (%d, %v), want (2, false)", state.CachedI32ValuesCount, state.CachedI32ValuesKnown)
	}
	if state.CachedI32ValuesValid[2] || state.CachedI32ValuesValid[3] {
		t.Fatalf("grown cache new validity flags = %v, want false slots", state.CachedI32ValuesValid)
	}
}

func TestArrayStoreWriteDynamicAppendPreservesCachedI32Values(t *testing.T) {
	handle := ArrayStoreNew()
	if err := ArrayStoreWrite(handle, 0, NewSmallInt(4, IntegerI32)); err != nil {
		t.Fatalf("ArrayStoreWrite(0): %v", err)
	}
	state := arrayStates[handle]
	state.CachedI32Values = []int32{4}
	state.CachedI32ValuesValid = []bool{true}
	state.CachedI32ValuesCount = 1
	state.CachedI32ValuesKnown = true

	if err := ArrayStoreWrite(handle, 1, NewSmallInt(5, IntegerI32)); err != nil {
		t.Fatalf("ArrayStoreWrite(1): %v", err)
	}

	if len(state.CachedI32Values) != 2 || len(state.CachedI32ValuesValid) != 2 {
		t.Fatalf("appended cache lengths = (%d, %d), want (2, 2)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if !state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 2 {
		t.Fatalf("appended cache count/known = (%d, %v), want (2, true)", state.CachedI32ValuesCount, state.CachedI32ValuesKnown)
	}
	if state.CachedI32Values[0] != 4 || state.CachedI32Values[1] != 5 {
		t.Fatalf("appended cache values = %v, want [4 5]", state.CachedI32Values)
	}
}

func TestArrayStoreWriteDynamicSparsePreservesCachedI32Values(t *testing.T) {
	handle := ArrayStoreNew()
	if err := ArrayStoreWrite(handle, 0, NewSmallInt(4, IntegerI32)); err != nil {
		t.Fatalf("ArrayStoreWrite(0): %v", err)
	}
	state := arrayStates[handle]
	state.CachedI32Values = []int32{4}
	state.CachedI32ValuesValid = []bool{true}
	state.CachedI32ValuesCount = 1
	state.CachedI32ValuesKnown = true

	if err := ArrayStoreWrite(handle, 3, NewSmallInt(9, IntegerI32)); err != nil {
		t.Fatalf("ArrayStoreWrite(3): %v", err)
	}

	if len(state.CachedI32Values) != 4 || len(state.CachedI32ValuesValid) != 4 {
		t.Fatalf("sparse cache lengths = (%d, %d), want (4, 4)", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if state.CachedI32ValuesKnown || state.CachedI32ValuesCount != 2 {
		t.Fatalf("sparse cache count/known = (%d, %v), want (2, false)", state.CachedI32ValuesCount, state.CachedI32ValuesKnown)
	}
	if state.CachedI32Values[0] != 4 || state.CachedI32Values[3] != 9 {
		t.Fatalf("sparse cache values = %v, want slot 0=4 slot 3=9", state.CachedI32Values)
	}
	if state.CachedI32ValuesValid[1] || state.CachedI32ValuesValid[2] {
		t.Fatalf("sparse cache gap validity = %v, want invalid gap", state.CachedI32ValuesValid)
	}
}

func TestArrayStoreWriteDynamicNonI32ClearsCachedI32Values(t *testing.T) {
	handle := ArrayStoreNew()
	if err := ArrayStoreWrite(handle, 0, NewSmallInt(4, IntegerI32)); err != nil {
		t.Fatalf("ArrayStoreWrite(0): %v", err)
	}
	state := arrayStates[handle]
	state.CachedI32Values = []int32{4}
	state.CachedI32ValuesValid = []bool{true}
	state.CachedI32ValuesCount = 1
	state.CachedI32ValuesKnown = true

	if err := ArrayStoreWrite(handle, 1, BoolValue{Val: true}); err != nil {
		t.Fatalf("ArrayStoreWrite(1): %v", err)
	}

	if state.CachedI32Values != nil || state.CachedI32ValuesValid != nil || state.CachedI32ValuesCount != 0 || state.CachedI32ValuesKnown {
		t.Fatalf("mixed-type write should clear cached i32 values, got values=%v valid=%v count=%d known=%v", state.CachedI32Values, state.CachedI32ValuesValid, state.CachedI32ValuesCount, state.CachedI32ValuesKnown)
	}
}
