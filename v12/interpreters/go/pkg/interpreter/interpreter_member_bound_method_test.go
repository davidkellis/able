package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestMemberAccessOnZeroArgBoundMethodUsesMethodResult(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()

	target := &runtime.StructInstanceValue{
		Fields: map[string]runtime.Value{
			"ok": runtime.BoolValue{Val: true},
		},
	}
	bound := runtime.NativeBoundMethodValue{
		Receiver: target,
		Method: runtime.NativeFunctionValue{
			Name:  "bytes",
			Arity: 0,
			Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
				return args[0], nil
			},
		},
	}

	got, err := interp.MemberGet(bound, runtime.StringValue{Val: "ok"}, env)
	if err != nil {
		t.Fatalf("member get on bound method: %v", err)
	}
	boolVal, ok := got.(runtime.BoolValue)
	if !ok {
		t.Fatalf("expected boolean field, got %T (%#v)", got, got)
	}
	if !boolVal.Val {
		t.Fatalf("expected boolean true, got %#v", got)
	}
}
