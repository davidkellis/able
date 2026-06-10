package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_DirectArrayIndexGetMonoU32PreservesRawValue(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	value := uint32(bytecodeSmallIntBoxMax + 90000)
	arr := monoU32ArrayValueForTest(t, value)

	got, err := vm.resolveDirectArrayIndexGetAt(arr, 0)
	if err != nil {
		t.Fatalf("mono u32 direct get failed: %v", err)
	}
	raw, ok := got.(bytecodeRawU32ResultValue)
	if !ok || uint32(raw) != value {
		t.Fatalf("mono u32 direct get result = %#v, want raw u32 %d", got, value)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("mono u32 direct get should not materialize boxed state")
	}
}

func TestBytecodeVM_DirectArrayIndexGetValidatedHandleReadsSmallIntegerIndex(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'c')
	alias := &runtime.ArrayValue{TrackedHandle: arr.Handle}

	got, token, tokenKnown, handled, err := vm.resolveDirectArrayIndexGetWithValidatedHandleAndToken(alias, runtime.NewSmallInt(2, runtime.IntegerI32), arr.Handle)
	if err != nil || !handled {
		t.Fatalf("validated handle get handled=%v err=%v", handled, err)
	}
	if !tokenKnown || token != bytecodeIndexTypeChar {
		t.Fatalf("validated handle token = (%d, %v), want char/true", token, tokenKnown)
	}
	if charVal, ok := got.(runtime.CharValue); !ok || charVal.Val != 'c' {
		t.Fatalf("validated handle get = %#v, want char c", got)
	}
	if alias.State != nil || alias.Elements != nil {
		t.Fatalf("validated handle get should not materialize alias state")
	}
}
