package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestArrayIndexFromValueAcceptsBytecodeRawI32(t *testing.T) {
	got, err := arrayIndexFromValue(bytecodeRawI32SlotCachedValue(7))
	if err != nil {
		t.Fatalf("arrayIndexFromValue(raw i32) error = %v", err)
	}
	if got != 7 {
		t.Fatalf("arrayIndexFromValue(raw i32) = %d, want 7", got)
	}
}

func TestBytecodeDirectArrayIndexAcceptsRawIntegerCarriers(t *testing.T) {
	tests := []struct {
		name  string
		value runtime.Value
		want  int
	}{
		{name: "i32 slot", value: bytecodeRawI32SlotCachedValue(13), want: 13},
		{name: "i32 stack cell", value: &bytecodeRawI32StackCell{Val: 17}, want: 17},
		{name: "i64 slot cell", value: &bytecodeRawI64SlotCell{Val: 19}, want: 19},
		{name: "u32 result", value: bytecodeRawU32ResultValue(23), want: 23},
	}
	for _, tc := range tests {
		got, ok, err := bytecodeDirectArrayIndex(tc.value)
		if err != nil {
			t.Fatalf("%s: bytecodeDirectArrayIndex error = %v", tc.name, err)
		}
		if !ok || got != tc.want {
			t.Fatalf("%s: bytecodeDirectArrayIndex = (%d, %v), want (%d, true)", tc.name, got, ok, tc.want)
		}
		got, ok = bytecodeDirectSmallArrayIndex(tc.value)
		if !ok || got != tc.want {
			t.Fatalf("%s: bytecodeDirectSmallArrayIndex = (%d, %v), want (%d, true)", tc.name, got, ok, tc.want)
		}
	}
}

func TestBytecodeVM_ArrayPopParity(t *testing.T) {
	module := mustParseModuleSource(t, `
fn main() -> i32 {
  handle := __able_array_new()
  __able_array_write(handle, 0, 3)
  idx := 1 - 0
  __able_array_set_len(handle, idx)
  __able_array_read(handle, 0)
}

main()
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode array pop mismatch: got=%#v want=%#v", got, want)
	}
}
