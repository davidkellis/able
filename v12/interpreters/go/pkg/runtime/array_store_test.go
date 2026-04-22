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
