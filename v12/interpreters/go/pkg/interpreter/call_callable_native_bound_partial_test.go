package interpreter

import (
	"fmt"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestCallCallableValue_NativeBoundMethodPartialDoesNotDoubleInjectReceiver(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	receiver := runtime.NewSmallInt(7, runtime.IntegerI32)
	arg := runtime.NewSmallInt(11, runtime.IntegerI32)

	native := runtime.NativeFunctionValue{
		Name:  "pair",
		Arity: 2,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("expected 2 args, got %d", len(args))
			}
			if !valuesEqual(args[0], receiver) {
				return nil, fmt.Errorf("receiver mismatch: got=%#v want=%#v", args[0], receiver)
			}
			return args[1], nil
		},
	}

	bound := runtime.BoundMethodValue{Receiver: receiver, Method: native}
	partialVal, err := interp.callCallableValue(bound, nil, env, nil)
	if err != nil {
		t.Fatalf("partial creation failed: %v", err)
	}
	if _, ok := partialVal.(*runtime.PartialFunctionValue); !ok {
		t.Fatalf("expected partial function value, got %#v", partialVal)
	}

	result, err := interp.callCallableValue(partialVal, []runtime.Value{arg}, env, nil)
	if err != nil {
		t.Fatalf("partial invocation failed: %v", err)
	}
	if !valuesEqual(result, arg) {
		t.Fatalf("unexpected result: got=%#v want=%#v", result, arg)
	}
}

func TestCallCallableValue_NativeSkipContextPassesNilContext(t *testing.T) {
	interp := New()
	env := interp.GlobalEnvironment()
	called := false

	native := runtime.NativeFunctionValue{
		Name:        "capture_ctx",
		Arity:       0,
		SkipContext: true,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			called = true
			if ctx != nil {
				return nil, fmt.Errorf("expected nil context, got %#v", ctx)
			}
			if len(args) != 0 {
				return nil, fmt.Errorf("expected no args, got %d", len(args))
			}
			return runtime.NewSmallInt(1, runtime.IntegerI32), nil
		},
	}

	result, err := interp.callCallableValue(native, nil, env, nil)
	if err != nil {
		t.Fatalf("native call failed: %v", err)
	}
	if !called {
		t.Fatalf("expected native impl to be called")
	}
	want := runtime.NewSmallInt(1, runtime.IntegerI32)
	if !valuesEqual(result, want) {
		t.Fatalf("unexpected result: got=%#v want=%#v", result, want)
	}
}
