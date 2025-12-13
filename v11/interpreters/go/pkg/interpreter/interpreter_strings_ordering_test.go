package interpreter

import (
	"path/filepath"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/runtime"
)

func TestStringCmpReturnsOrderingInstances(t *testing.T) {
	loader, err := driver.NewLoader([]driver.SearchPath{
		{Path: filepath.Join("..", "..", "..", "..", "stdlib", "src"), Kind: driver.RootStdlib},
		{Path: filepath.Join("..", "..", "..", "..", "kernel", "src"), Kind: driver.RootStdlib},
	})
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	stdlibProgram, err := loader.Load(filepath.Join("..", "..", "..", "..", "stdlib", "src", "text", "string.able"))
	if err != nil {
		t.Fatalf("load stdlib string: %v", err)
	}

	interp := New()
	if _, _, _, err := interp.EvaluateProgram(stdlibProgram, ProgramEvaluationOptions{SkipTypecheck: true}); err != nil {
		t.Fatalf("evaluate stdlib string: %v", err)
	}
	ctx := &runtime.NativeCallContext{Env: interp.GlobalEnvironment()}

	boundValue, err := interp.stringMemberWithOverrides(runtime.StringValue{Val: "a"}, ast.NewIdentifier("cmp"), interp.global)
	if err != nil {
		t.Fatalf("cmp lookup failed: %v", err)
	}
	var (
		receiver runtime.Value
		call     func(receiver runtime.Value, other runtime.Value) (runtime.Value, error)
	)
	switch b := boundValue.(type) {
	case *runtime.NativeBoundMethodValue:
		receiver = b.Receiver
		call = func(receiver runtime.Value, other runtime.Value) (runtime.Value, error) {
			return b.Method.Impl(ctx, []runtime.Value{receiver, other})
		}
	case *runtime.BoundMethodValue:
		receiver = b.Receiver
		call = func(receiver runtime.Value, other runtime.Value) (runtime.Value, error) {
			bound := runtime.BoundMethodValue{Receiver: receiver, Method: b.Method}
			return interp.callCallableValue(bound, []runtime.Value{other}, nil, nil)
		}
	default:
		t.Fatalf("expected bound method, got %T", boundValue)
	}

	less, err := call(receiver, runtime.StringValue{Val: "b"})
	if err != nil {
		t.Fatalf("cmp(a,b) failed: %v", err)
	}
	equal, err := call(receiver, runtime.StringValue{Val: "a"})
	if err != nil {
		t.Fatalf("cmp(a,a) failed: %v", err)
	}
	greater, err := call(receiver, runtime.StringValue{Val: "A"})
	if err != nil {
		t.Fatalf("cmp(a,A) failed: %v", err)
	}

	assertOrderingTag := func(val runtime.Value, expected string) {
		inst, ok := val.(*runtime.StructInstanceValue)
		if !ok || inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
			t.Fatalf("expected ordering struct instance, got %#v", val)
		}
		if got := inst.Definition.Node.ID.Name; got != expected {
			t.Fatalf("expected tag %s, got %s", expected, got)
		}
	}

	assertOrderingTag(less, "Less")
	assertOrderingTag(equal, "Equal")
	assertOrderingTag(greater, "Greater")
}
