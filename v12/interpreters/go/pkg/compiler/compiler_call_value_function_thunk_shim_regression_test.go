package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCallValueFunctionThunkDispatch(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_call_function_thunk(value runtime.Value, args []runtime.Value, call *ast.FunctionCall) (runtime.Value, bool) {") {
		t.Fatalf("expected shared function-thunk helper for __able_call_value dispatch")
	}
	if !strings.Contains(compiledSrc, "fn, ok, nilPtr := __able_callable_function_value(value)") {
		t.Fatalf("expected function-thunk helper to unwrap function values through shared helper")
	}
	if !strings.Contains(compiledSrc, "if !ok || nilPtr {") {
		t.Fatalf("expected function-thunk helper to reject non-function and typed-nil function values")
	}
	if !strings.Contains(compiledSrc, "if val, err, ok := __able_call_compiled_thunk(fn.Bytecode, __able_runtime.Env(), args); ok {") {
		t.Fatalf("expected function-thunk helper to wrap compiled thunk invocation")
	}

	start := strings.Index(compiledSrc, "func __able_call_bound_method(")
	if start < 0 {
		t.Fatalf("expected __able_call_bound_method helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_call_value(")
	if end < 0 {
		t.Fatalf("expected __able_call_bound_method segment terminator")
	}
	segment = segment[:end]
	if strings.Contains(segment, "__able_call_compiled_thunk(method.Bytecode, __able_runtime.Env(), injected)") {
		t.Fatalf("expected inline method compiled-thunk dispatch block to be removed from __able_call_bound_method")
	}
	if strings.Contains(segment, "if methodThunk, ok := method.(*runtime.FunctionValue); ok {") || strings.Contains(segment, "if methodThunk, ok, nilPtr := __able_callable_function_value(method); ok || nilPtr {") {
		t.Fatalf("expected __able_call_bound_method to avoid local methodThunk unwrapping branches")
	}
	if !strings.Contains(segment, "if val, handled := __able_call_function_thunk(method, injected, call); handled {") {
		t.Fatalf("expected __able_call_bound_method to delegate thunk dispatch directly to shared helper")
	}

	start = strings.Index(compiledSrc, "func __able_call_value(")
	if start < 0 {
		t.Fatalf("expected __able_call_value helper to be emitted")
	}
	segment = compiledSrc[start:]
	end = strings.Index(segment, "type __able_compiled_call_entry struct {")
	if end < 0 {
		t.Fatalf("expected __able_call_value segment terminator")
	}
	segment = segment[:end]
	if strings.Contains(segment, "__able_call_compiled_thunk(v.Bytecode, __able_runtime.Env(), args)") {
		t.Fatalf("expected inline *runtime.FunctionValue compiled-thunk block to be removed from __able_call_value")
	}
	if strings.Contains(segment, "switch v := fn.(type)") {
		t.Fatalf("expected single-case *runtime.FunctionValue switch to be removed from __able_call_value")
	}
	if strings.Contains(segment, "if fnThunk, ok := fn.(*runtime.FunctionValue); ok {") || strings.Contains(segment, "if fnThunk, ok, nilPtr := __able_callable_function_value(fn); ok || nilPtr {") {
		t.Fatalf("expected __able_call_value to avoid local fnThunk unwrapping branches")
	}
	if !strings.Contains(segment, "if val, handled := __able_call_function_thunk(fn, args, call); handled {") {
		t.Fatalf("expected __able_call_value to delegate thunk dispatch directly to shared helper")
	}
}
