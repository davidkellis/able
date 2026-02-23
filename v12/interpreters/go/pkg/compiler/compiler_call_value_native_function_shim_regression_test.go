package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCallValueNativeFunctionDispatchBranches(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_call_native_function_value(value runtime.Value, args []runtime.Value, call *ast.FunctionCall, ctx *runtime.NativeCallContext) (runtime.Value, bool) {") {
		t.Fatalf("expected shared native-function value helper for __able_call_value dispatch")
	}
	if !strings.Contains(compiledSrc, "native, ok, nilPtr := __able_callable_native_function_value(value)") {
		t.Fatalf("expected native-function helper to use shared native callable unwrapping")
	}
	if strings.Contains(compiledSrc, "if nilPtr || !ok {") {
		t.Fatalf("expected native-function helper guard to use normalized helper-order check")
	}
	if !strings.Contains(compiledSrc, "if !ok || nilPtr {") {
		t.Fatalf("expected native-function helper guard to use normalized helper-order check")
	}
	if !strings.Contains(compiledSrc, "return __able_call_native_function(native, partialTarget, args, args, call, ctx), true") {
		t.Fatalf("expected native-function helper to delegate through shared native-call helper")
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

	if strings.Contains(segment, "case runtime.NativeFunctionValue:") {
		t.Fatalf("expected inline runtime.NativeFunctionValue dispatch branch to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case *runtime.NativeFunctionValue:") {
		t.Fatalf("expected inline *runtime.NativeFunctionValue dispatch branch to be removed from __able_call_value")
	}
	if !strings.Contains(segment, "if val, handled := __able_call_native_function_value(fn, args, call, ctx); handled {") {
		t.Fatalf("expected __able_call_value to use shared native-function value helper")
	}
}
