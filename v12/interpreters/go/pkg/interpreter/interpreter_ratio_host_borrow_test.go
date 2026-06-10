package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestPrimitiveKernelNativesBorrowArguments(t *testing.T) {
	interp := NewBytecode()
	for _, name := range []string{
		"__able_ratio_from_float",
		"__able_f32_bits",
		"__able_f64_bits",
		"__able_f64_sqrt",
		"__able_u64_mul",
	} {
		value, err := interp.GlobalEnvironment().Get(name)
		if err != nil {
			t.Fatalf("lookup %s: %v", name, err)
		}
		native, ok := value.(runtime.NativeFunctionValue)
		if !ok || !native.BorrowArgs {
			t.Fatalf("%s = %#v, want argument-borrowing native", name, value)
		}
		if native.RawImpl != nil {
			t.Fatalf("%s unexpectedly has an ordinary raw implementation", name)
		}
	}
}

func TestBytecodeVM_PrimitiveKernelBorrowedArgsAreNotRetained(t *testing.T) {
	interp := NewBytecode()
	callee, err := interp.GlobalEnvironment().Get("__able_f64_sqrt")
	if err != nil {
		t.Fatalf("sqrt lookup: %v", err)
	}
	target, ok := bytecodeResolveExactNativeCallTarget(callee, 1)
	if !ok || !target.native.BorrowArgs || target.native.RawImpl != nil {
		t.Fatalf("sqrt target = %#v, %t; want borrowed ordinary native", target, ok)
	}
	args := []runtime.Value{runtime.FloatValue{Val: 6.25, TypeSuffix: runtime.FloatF64}}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	result, handled, err := vm.execExactNativeCall(target, args, nil)
	if err != nil || !handled {
		t.Fatalf("exact sqrt = %#v, %t, %v", result, handled, err)
	}
	root, ok := result.(runtime.FloatValue)
	if !ok || root.TypeSuffix != runtime.FloatF64 || root.Val != 2.5 {
		t.Fatalf("sqrt result = %#v, want f64 2.5", result)
	}
	args[0] = runtime.StringValue{Val: "reused"}
	if root.Val != 2.5 {
		t.Fatalf("sqrt result changed after argument slice reuse: %#v", root)
	}
}
