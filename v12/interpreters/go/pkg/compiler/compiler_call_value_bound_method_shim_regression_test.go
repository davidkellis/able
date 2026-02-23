package compiler

import (
	"strings"
	"testing"
)

func TestCompilerNormalizesCallValueBoundMethodDispatchBranches(t *testing.T) {
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

	if !strings.Contains(compiledSrc, "func __able_call_bound_method(bound runtime.BoundMethodValue, partialTarget runtime.Value, args []runtime.Value, call *ast.FunctionCall, ctx *runtime.NativeCallContext) (runtime.Value, bool) {") {
		t.Fatalf("expected shared bound-method dispatch helper for __able_call_value")
	}
	helperStart := strings.Index(compiledSrc, "func __able_call_bound_method(")
	if helperStart < 0 {
		t.Fatalf("expected __able_call_bound_method helper to be emitted")
	}
	helperSegment := compiledSrc[helperStart:]
	helperEnd := strings.Index(helperSegment, "func __able_call_value(")
	if helperEnd < 0 {
		t.Fatalf("expected __able_call_bound_method helper terminator")
	}
	helperSegment = helperSegment[:helperEnd]
	if strings.Contains(helperSegment, "switch method := bound.Method.(type)") {
		t.Fatalf("expected __able_call_bound_method to remove inline bound.Method switch dispatch")
	}
	if strings.Contains(helperSegment, "if iface, ok, nilPtr := __able_callable_interface_value(method); nilPtr {") {
		t.Fatalf("expected legacy nilPtr-first interface unwrapping branch to be removed from __able_call_bound_method")
	}
	if !strings.Contains(helperSegment, "if iface, ok, nilPtr := __able_callable_interface_value(method); ok || nilPtr {") {
		t.Fatalf("expected __able_call_bound_method to use normalized interface helper guard")
	}
	if !strings.Contains(helperSegment, "if !ok || nilPtr {") {
		t.Fatalf("expected __able_call_bound_method interface helper branch to enforce normalized typed-nil handling")
	}
	if strings.Contains(helperSegment, "if methodThunk, ok := method.(*runtime.FunctionValue); ok {") || strings.Contains(helperSegment, "if methodThunk, ok, nilPtr := __able_callable_function_value(method); ok || nilPtr {") {
		t.Fatalf("expected local methodThunk unwrapping branches to be removed from __able_call_bound_method")
	}
	if !strings.Contains(helperSegment, "if val, handled := __able_call_function_thunk(method, injected, call); handled {") {
		t.Fatalf("expected __able_call_bound_method thunk dispatch to delegate through shared helper directly")
	}
	if !strings.Contains(helperSegment, "if native, ok, nilPtr := __able_callable_native_function_value(method); ok && !nilPtr {") {
		t.Fatalf("expected __able_call_bound_method to use normalized native-function unwrapping helper path")
	}
	if !strings.Contains(helperSegment, "return __able_call_native_function(native, partialTarget, args, injected, call, ctx), true") {
		t.Fatalf("expected __able_call_bound_method native dispatch to route through shared native-call helper")
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

	if strings.Contains(segment, "case runtime.BoundMethodValue:\n\t\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n\t\t\tswitch method := v.Method.(type) {") {
		t.Fatalf("expected inline runtime.BoundMethodValue dispatch block to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case *runtime.BoundMethodValue:\n\t\t\tif v != nil {\n\t\t\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n\t\t\t\tswitch method := v.Method.(type) {") {
		t.Fatalf("expected inline *runtime.BoundMethodValue dispatch block to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case runtime.BoundMethodValue:") {
		t.Fatalf("expected runtime.BoundMethodValue switch branch to be removed from __able_call_value")
	}
	if strings.Contains(segment, "case *runtime.BoundMethodValue:") {
		t.Fatalf("expected *runtime.BoundMethodValue switch branch to be removed from __able_call_value")
	}
	if !strings.Contains(segment, "if bound, ok, _ := __able_callable_bound_method_value(fn); ok {") {
		t.Fatalf("expected __able_call_value to use normalized bound-method unwrapping helper path")
	}
	if !strings.Contains(segment, "if val, handled := __able_call_bound_method(bound, fn, args, call, ctx); handled {") {
		t.Fatalf("expected normalized bound-method dispatch to use shared helper")
	}
}
