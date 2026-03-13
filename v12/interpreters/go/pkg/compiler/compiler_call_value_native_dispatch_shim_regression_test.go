package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCallValueNativeDispatchBranches(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_call_native_function(fn runtime.NativeFunctionValue, partialTarget runtime.Value, providedArgs []runtime.Value, invokeArgs []runtime.Value, call *ast.FunctionCall, ctx *runtime.NativeCallContext) (runtime.Value, *__ableControl) {") {
		t.Fatalf("expected shared native-call helper for __able_call_value dispatch")
	}
	if !strings.Contains(compiledSrc, "func __able_call_native_function_value(value runtime.Value, args []runtime.Value, call *ast.FunctionCall, ctx *runtime.NativeCallContext) (runtime.Value, *__ableControl, bool) {") {
		t.Fatalf("expected shared native-function value helper for __able_call_value dispatch")
	}
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

	if strings.Contains(segment, "case runtime.NativeFunctionValue:\n\t\t\tif v.Arity >= 0 {") {
		t.Fatalf("expected inline runtime.NativeFunctionValue dispatch block to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case runtime.NativeBoundMethodValue:\n\t\t\tif v.Method.Arity >= 0 {") {
		t.Fatalf("expected inline runtime.NativeBoundMethodValue dispatch block to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case runtime.NativeFunctionValue:\n\t\t\t\tif method.Arity >= 0 {") {
		t.Fatalf("expected inline bound-method native dispatch block to be removed from __able_call_value")
	}

	if !strings.Contains(segment, "if val, control, handled := __able_call_native_function_value(fn, args, call, ctx); handled {") {
		t.Fatalf("expected __able_call_value to use shared native-function value helper")
	}
	if strings.Contains(segment, "case runtime.NativeBoundMethodValue:") {
		t.Fatalf("expected runtime.NativeBoundMethodValue switch dispatch branch to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case *runtime.NativeBoundMethodValue:") {
		t.Fatalf("expected *runtime.NativeBoundMethodValue switch dispatch branch to be removed from __able_call_value")
	}
	if !strings.Contains(segment, "if nativeBound, ok, _ := __able_callable_native_bound_method_value(fn); ok {") {
		t.Fatalf("expected __able_call_value to use normalized native-bound unwrapping helper path")
	}
	if !strings.Contains(segment, "return __able_call_native_bound_method(nativeBound, fn, args, call, ctx)") {
		t.Fatalf("expected __able_call_value native-bound dispatch to use shared native-bound helper")
	}
	if strings.Contains(segment, "case runtime.BoundMethodValue:") {
		t.Fatalf("expected runtime.BoundMethodValue switch dispatch branch to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case *runtime.BoundMethodValue:") {
		t.Fatalf("expected *runtime.BoundMethodValue switch dispatch branch to be removed from __able_call_value")
	}
	if !strings.Contains(segment, "if bound, ok, _ := __able_callable_bound_method_value(fn); ok {") {
		t.Fatalf("expected __able_call_value to use normalized bound-method unwrapping helper path")
	}
	if !strings.Contains(segment, "if val, control, handled := __able_call_bound_method(bound, fn, args, call, ctx); handled {") {
		t.Fatalf("expected __able_call_value bound-method dispatch to use shared helper")
	}
	if !strings.Contains(compiledSrc, "return __able_call_native_function(bound.Method, partialTarget, args, injected, call, ctx)") {
		t.Fatalf("expected native-bound helper to use shared native-call helper")
	}
	if !strings.Contains(compiledSrc, "if native, ok, nilPtr := __able_callable_native_function_value(method); ok && !nilPtr {") {
		t.Fatalf("expected bound-method native dispatch to use normalized native-function unwrapping helper after interface unwrapping")
	}
	if !strings.Contains(compiledSrc, "val, control := __able_call_native_function(native, partialTarget, args, injected, call, ctx)") {
		t.Fatalf("expected bound-method native dispatch to remain normalized through shared native-call helper")
	}
}
