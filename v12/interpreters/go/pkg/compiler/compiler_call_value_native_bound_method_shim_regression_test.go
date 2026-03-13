package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCallValueNativeBoundMethodDispatchBranches(t *testing.T) {
	_, compiledSrc := compileOutputs(t, "demo", map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"fn main() -> void {",
			"  _ = 1",
			"}",
			"",
		}, "\n"),
	})

	if !strings.Contains(compiledSrc, "func __able_call_native_bound_method(bound runtime.NativeBoundMethodValue, partialTarget runtime.Value, args []runtime.Value, call *ast.FunctionCall, ctx *runtime.NativeCallContext) (runtime.Value, *__ableControl) {") {
		t.Fatalf("expected shared native-bound helper for __able_call_value dispatch")
	}

	start := strings.Index(compiledSrc, "func __able_call_value(")
	if start < 0 {
		t.Fatalf("expected __able_call_value helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "type __able_compiled_call_entry struct {")
	if end < 0 {
		t.Fatalf("expected __able_call_value segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "case runtime.NativeBoundMethodValue:\n\t\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n\t\t\treturn __able_call_native_function(v.Method, v, args, injected, call, ctx)") {
		t.Fatalf("expected inline runtime.NativeBoundMethodValue receiver-injection block to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case *runtime.NativeBoundMethodValue:\n\t\t\tif v != nil {\n\t\t\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n\t\t\t\treturn __able_call_native_function(v.Method, v, args, injected, call, ctx)") {
		t.Fatalf("expected inline *runtime.NativeBoundMethodValue receiver-injection block to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case runtime.NativeBoundMethodValue:") {
		t.Fatalf("expected runtime.NativeBoundMethodValue switch branch to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case *runtime.NativeBoundMethodValue:") {
		t.Fatalf("expected *runtime.NativeBoundMethodValue switch branch to be removed from __able_call_value")
	}
	if !strings.Contains(segment, "if nativeBound, ok, _ := __able_callable_native_bound_method_value(fn); ok {") {
		t.Fatalf("expected __able_call_value to use shared native-bound unwrapping helper path")
	}
	if !strings.Contains(segment, "return __able_call_native_bound_method(nativeBound, fn, args, call, ctx)") {
		t.Fatalf("expected __able_call_value to dispatch native-bound calls through shared helper using normalized nativeBound value")
	}
}
