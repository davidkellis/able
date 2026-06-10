package interpreter

import (
	"errors"
	"testing"

	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_StructCallableFieldRawImplReceivesRawI64Arg(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	seenRaw := false
	native := runtime.NativeFunctionValue{
		Name:       "raw_probe",
		Arity:      1,
		BorrowArgs: true,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			t.Fatalf("materialized native implementation should not run when RawImpl is available")
			return nil, nil
		},
		RawImpl: func(_ *runtime.NativeCallContext, args []runtime.RawValue) (runtime.RawValue, error) {
			if len(args) != 1 {
				t.Fatalf("raw arg count = %d, want 1", len(args))
			}
			kind, raw, ok := args[0].Integer()
			if !ok || kind != runtime.IntegerI64 || raw != 123 {
				t.Fatalf("raw arg = (%s, %d, %t), want i64 123", kind, raw, ok)
			}
			seenRaw = true
			return runtime.NewRawIntegerValue(runtime.IntegerI64, 456), nil
		},
	}
	receiver := &runtime.StructInstanceValue{
		Fields: map[string]runtime.Value{
			"probe": native,
		},
	}
	vm.stack = []runtime.Value{receiver, &bytecodeRawI64SlotCell{Val: 123}}

	_, handled, err := vm.execCallMemberStructCallableField(
		bytecodeInstruction{name: "probe", argCount: 1},
		receiver,
		0,
		1,
		nil,
		nil,
		nil,
		false,
	)
	if err != nil {
		t.Fatalf("execCallMemberStructCallableField: %v", err)
	}
	if !handled {
		t.Fatalf("expected struct callable field raw native call to be handled")
	}
	if !seenRaw {
		t.Fatalf("raw native implementation was not invoked")
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack length = %d, want 1", len(vm.stack))
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI64 || raw != 456 {
		t.Fatalf("raw native result = %#v (%s, %d, %t), want raw i64 456", vm.stack[0], kind, raw, ok)
	}
	if _, boxed := vm.stack[0].(runtime.IntegerValue); boxed {
		t.Fatalf("raw native result should stay as a bytecode carrier, got boxed integer")
	}
}

func TestBytecodeVM_ExactNativeRawSmallArityUsesInlineScratch(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	receiver := runtime.StringValue{Val: "receiver"}
	called := false
	native := runtime.NativeFunctionValue{
		Name:        "raw_bound_probe",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		RawImpl: func(_ *runtime.NativeCallContext, args []runtime.RawValue) (runtime.RawValue, error) {
			called = true
			if len(args) != 2 {
				t.Fatalf("raw arg count = %d, want receiver plus one arg", len(args))
			}
			if got, ok := args[0].Materialize().(runtime.StringValue); !ok || got.Val != receiver.Val {
				t.Fatalf("receiver raw arg = %#v, want %#v", args[0].Materialize(), receiver)
			}
			kind, raw, ok := args[1].Integer()
			if !ok || kind != runtime.IntegerI64 || raw != 123 {
				t.Fatalf("explicit raw arg = (%s, %d, %t), want i64 123", kind, raw, ok)
			}
			return runtime.NewRawIntegerValue(runtime.IntegerI64, raw+1), nil
		},
	}

	result, err := vm.execExactNativeRawCall(nil, bytecodeExactNativeCallTarget{
		native:           native,
		injectedReceiver: receiver,
		hasReceiver:      true,
	}, []runtime.Value{&bytecodeRawI64SlotCell{Val: 123}})
	if err != nil {
		t.Fatalf("execExactNativeRawCall: %v", err)
	}
	if !called {
		t.Fatalf("raw native implementation was not invoked")
	}
	kind, raw, ok := result.Integer()
	if !ok || kind != runtime.IntegerI64 || raw != 124 {
		t.Fatalf("raw result = (%s, %d, %t), want i64 124", kind, raw, ok)
	}
	if vm.nativeRawArgsBusy || len(vm.nativeRawArgs) != 0 {
		t.Fatalf("small raw exact call should release VM raw arg scratch, busy=%t len=%d", vm.nativeRawArgsBusy, len(vm.nativeRawArgs))
	}
	if cap(vm.nativeRawArgs) < 2 {
		t.Fatalf("small raw exact call should retain inline raw arg scratch, cap=%d", cap(vm.nativeRawArgs))
	}
	args := vm.nativeRawArgs[:2]
	if &args[0] != &vm.nativeRawArgsInline[0] {
		t.Fatalf("small raw exact call should use VM inline raw arg scratch")
	}
}

func TestBytecodeVM_ExactNativeRawSmallArityReleasesInlineScratchOnError(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	errSentinel := errors.New("raw failure")
	native := runtime.NativeFunctionValue{
		Name:        "raw_error_probe",
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		RawImpl: func(_ *runtime.NativeCallContext, args []runtime.RawValue) (runtime.RawValue, error) {
			if len(args) != 1 {
				t.Fatalf("raw arg count = %d, want 1", len(args))
			}
			return runtime.RawValue{}, errSentinel
		},
	}

	_, err := vm.execExactNativeRawCall(nil, bytecodeExactNativeCallTarget{
		native: native,
	}, []runtime.Value{&bytecodeRawI64SlotCell{Val: 123}})
	if !errors.Is(err, errSentinel) {
		t.Fatalf("execExactNativeRawCall error = %v, want %v", err, errSentinel)
	}
	if vm.nativeRawArgsBusy || len(vm.nativeRawArgs) != 0 {
		t.Fatalf("erroring raw exact call should release VM raw arg scratch, busy=%t len=%d", vm.nativeRawArgsBusy, len(vm.nativeRawArgs))
	}
	args := vm.nativeRawArgs[:1]
	if &args[0] != &vm.nativeRawArgsInline[0] {
		t.Fatalf("erroring raw exact call should keep VM inline raw arg scratch")
	}
}
