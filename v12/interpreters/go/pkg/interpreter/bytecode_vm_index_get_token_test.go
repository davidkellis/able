package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CanonicalIndexGetWithTokenCarriesPrimitiveToken(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{{op: bytecodeOpIndexGet}}}
	vm.ip = 0
	arr := monoCharArrayValueForTest(t, 'a', 'b', 'l', 'e')

	got, token, tokenKnown, err := vm.resolveIndexGetWithToken(arr, runtime.NewSmallInt(2, runtime.IntegerI32))
	if err != nil {
		t.Fatalf("resolveIndexGetWithToken canonical fast path failed: %v", err)
	}
	if charVal, ok := got.(runtime.CharValue); !ok || charVal.Val != 'l' {
		t.Fatalf("resolveIndexGetWithToken result = %#v, want char 'l'", got)
	}
	if !tokenKnown || token != bytecodeIndexTypeChar {
		t.Fatalf("resolveIndexGetWithToken token = (%d, %v), want char/true", token, tokenKnown)
	}
	if arr.State != nil || arr.Elements != nil {
		t.Fatalf("resolveIndexGetWithToken should not materialize boxed state")
	}
}

func TestBytecodeVM_MonoPrimitiveArrayTokenTableCoversRuntimeKinds(t *testing.T) {
	cases := []struct {
		kind runtime.ArrayStoreMonoPrimitiveReadKind
		want uint16
	}{
		{runtime.ArrayStoreMonoPrimitiveReadI32, bytecodeIndexTypeI32},
		{runtime.ArrayStoreMonoPrimitiveReadI64, bytecodeIndexTypeI64},
		{runtime.ArrayStoreMonoPrimitiveReadBool, bytecodeIndexTypeBool},
		{runtime.ArrayStoreMonoPrimitiveReadChar, bytecodeIndexTypeChar},
		{runtime.ArrayStoreMonoPrimitiveReadU8, bytecodeIndexTypeU8},
		{runtime.ArrayStoreMonoPrimitiveReadU32, bytecodeIndexTypeU32},
		{runtime.ArrayStoreMonoPrimitiveReadU64, bytecodeIndexTypeU64},
		{runtime.ArrayStoreMonoPrimitiveReadF64, bytecodeIndexTypeF64},
	}
	for _, tc := range cases {
		got, ok := bytecodeMonoPrimitiveArrayToken(tc.kind)
		if !ok || got != tc.want {
			t.Fatalf("bytecodeMonoPrimitiveArrayToken(%d) = (%d, %v), want (%d, true)", tc.kind, got, ok, tc.want)
		}
	}
	if got, ok := bytecodeMonoPrimitiveArrayToken(runtime.ArrayStoreMonoPrimitiveReadNone); ok || got != bytecodeIndexTypeUnknown {
		t.Fatalf("bytecodeMonoPrimitiveArrayToken(None) = (%d, %v), want unknown/false", got, ok)
	}
	if got, ok := bytecodeMonoPrimitiveArrayToken(runtime.ArrayStoreMonoPrimitiveReadKind(len(bytecodeMonoPrimitiveArrayTokenTable))); ok || got != bytecodeIndexTypeUnknown {
		t.Fatalf("bytecodeMonoPrimitiveArrayToken(out of range) = (%d, %v), want unknown/false", got, ok)
	}
}

func TestBytecodeVM_IndexGetWithTokenDoesNotSkipPropagationWhenCharMayBeError(t *testing.T) {
	interp := NewBytecode()
	preloadArrayStdlibForTest(t, interp)
	interp.implMethods["char"] = []implEntry{{interfaceName: "Error"}}
	interp.invalidateMethodCache()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	arr := monoCharArrayValueForTest(t, 'a', 'b')
	vm.currentProgram = &bytecodeProgram{instructions: []bytecodeInstruction{
		{op: bytecodeOpIndexGet},
		{op: bytecodeOpPropagation},
	}}
	vm.ip = 0
	vm.stack = []runtime.Value{arr, runtime.NewSmallInt(1, runtime.IntegerI32)}

	if err := vm.execIndexGet(bytecodeInstruction{op: bytecodeOpIndexGet}); err != nil {
		t.Fatalf("generic index get propagation guard failed: %v", err)
	}
	if vm.ip != 1 {
		t.Fatalf("ip after char Error impl index get = %d, want propagation opcode at 1", vm.ip)
	}
}
