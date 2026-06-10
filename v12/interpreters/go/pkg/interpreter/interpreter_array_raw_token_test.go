package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeIndexValueTypeTokenRecognizesRawPrimitiveCarriers(t *testing.T) {
	tests := []struct {
		name  string
		value runtime.Value
		want  uint16
	}{
		{name: "raw_i32", value: bytecodeRawI32SlotCachedValue(7), want: bytecodeIndexTypeI32},
		{name: "raw_u32", value: bytecodeRawIntegerValue{Raw: 11, TypeSuffix: runtime.IntegerU32}, want: bytecodeIndexTypeU32},
		{name: "raw_f64", value: bytecodeRawF64SlotValue(3.5), want: bytecodeIndexTypeF64},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := bytecodeIndexValueTypeToken(tc.value)
			if !ok {
				t.Fatalf("bytecodeIndexValueTypeToken(%T) reported unknown", tc.value)
			}
			if got != tc.want {
				t.Fatalf("bytecodeIndexValueTypeToken(%T) = %d, want %d", tc.value, got, tc.want)
			}
		})
	}
}

func TestInterpreterSyncTrackedRawI32AppendSeedsTokenAndCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue(nil, 0)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}

	vm.appendTrackedArrayValueFast(arr, state, bytecodeRawI32SlotCachedValue(7))

	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		t.Fatalf("tracked raw append token = known:%v token:%d, want known i32", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if len(state.CachedI32Values) != 1 || len(state.CachedI32ValuesValid) != 1 {
		t.Fatalf("tracked raw append cache lengths = %d/%d, want 1/1", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if !state.CachedI32ValuesValid[0] || state.CachedI32Values[0] != 7 {
		t.Fatalf("tracked raw append cache entry = (%v, %d), want (true, 7)", state.CachedI32ValuesValid[0], state.CachedI32Values[0])
	}
	if !state.CachedI32ValuesKnown {
		t.Fatalf("tracked raw append should mark cached i32 values known")
	}
}

func TestResizeTrackedArrayI32RawCacheUsesValueBackingCapacity(t *testing.T) {
	state := &arrayState{
		Values:                make([]runtime.Value, 1, 16),
		ElementTypeToken:      bytecodeIndexTypeI32,
		ElementTypeTokenKnown: true,
	}

	if !resizeTrackedArrayI32RawCacheLength(state, 1) {
		t.Fatalf("resizeTrackedArrayI32RawCacheLength returned false")
	}
	if len(state.CachedI32Values) != 1 || len(state.CachedI32ValuesValid) != 1 {
		t.Fatalf("cache lengths = %d/%d, want 1/1", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if cap(state.CachedI32Values) != cap(state.Values) || cap(state.CachedI32ValuesValid) != cap(state.Values) {
		t.Fatalf("cache caps = %d/%d, value cap %d", cap(state.CachedI32Values), cap(state.CachedI32ValuesValid), cap(state.Values))
	}
}

func TestInterpreterSyncArrayHandleLengthKeepsRawI32TokenAndCache(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := interp.newArrayValue(nil, 0)

	state, err := interp.ensureArrayState(arr, 0)
	if err != nil {
		t.Fatalf("ensure array state: %v", err)
	}

	vm.appendTrackedArrayValueFast(arr, state, bytecodeRawI32SlotCachedValue(7))
	vm.appendTrackedArrayValueFast(arr, state, bytecodeRawI32SlotCachedValue(9))

	setArrayLength(state, 1)
	interp.syncArrayHandleLength(arr.Handle, state)

	if !state.ElementTypeTokenKnown || state.ElementTypeToken != bytecodeIndexTypeI32 {
		t.Fatalf("length sync token = known:%v token:%d, want known i32", state.ElementTypeTokenKnown, state.ElementTypeToken)
	}
	if len(state.CachedI32Values) != 1 || len(state.CachedI32ValuesValid) != 1 {
		t.Fatalf("length sync cache lengths = %d/%d, want 1/1", len(state.CachedI32Values), len(state.CachedI32ValuesValid))
	}
	if !state.CachedI32ValuesValid[0] || state.CachedI32Values[0] != 7 {
		t.Fatalf("length sync cache entry = (%v, %d), want (true, 7)", state.CachedI32ValuesValid[0], state.CachedI32Values[0])
	}
	if !state.CachedI32ValuesKnown {
		t.Fatalf("length sync should keep cached i32 values known")
	}
}
