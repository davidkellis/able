package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCallCompiledThunkDispatch(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_compiled_thunk_value(bytecode any) (interpreter.CompiledThunk, bool, bool) {") {
		t.Fatalf("expected shared compiled-thunk unwrapping helper to be emitted")
	}

	start := strings.Index(compiledSrc, "func __able_call_compiled_thunk(bytecode any, env *runtime.Environment, args []runtime.Value) (runtime.Value, error, bool) {")
	if start < 0 {
		t.Fatalf("expected __able_call_compiled_thunk helper to be emitted")
	}
	segment := compiledSrc[start:]
	end := strings.Index(segment, "func __able_call_native_function(fn runtime.NativeFunctionValue, partialTarget runtime.Value, providedArgs []runtime.Value, invokeArgs []runtime.Value, call *ast.FunctionCall, ctx *runtime.NativeCallContext) runtime.Value {")
	if end < 0 {
		t.Fatalf("expected __able_call_compiled_thunk segment terminator")
	}
	segment = segment[:end]

	if strings.Contains(segment, "if thunk, ok := bytecode.(func(*runtime.Environment, []runtime.Value) (runtime.Value, error)); ok && thunk != nil {") {
		t.Fatalf("expected legacy direct function-type assertion branch to be removed from __able_call_compiled_thunk")
	}
	if strings.Contains(segment, "if thunk, ok := bytecode.(interpreter.CompiledThunk); ok && thunk != nil {") {
		t.Fatalf("expected legacy direct CompiledThunk assertion branch to be removed from __able_call_compiled_thunk")
	}
	if !strings.Contains(segment, "thunk, ok, nilPtr := __able_compiled_thunk_value(bytecode)") {
		t.Fatalf("expected __able_call_compiled_thunk to use shared compiled-thunk unwrapping helper")
	}
	if !strings.Contains(segment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_call_compiled_thunk to honor shared helper typed-nil signaling")
	}
}
