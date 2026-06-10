package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CallMemberUsesGenericIteratorNextFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	step := 0
	iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
		switch step {
		case 0:
			step++
			return runtime.NewSmallInt(7, runtime.IntegerI32), false, nil
		default:
			return nil, true, nil
		}
	}, nil)

	vm.stack = []runtime.Value{iter}
	_, err := vm.execCallMemberNext(bytecodeInstruction{name: "next", argCount: 0}, nil)
	if err != nil {
		t.Fatalf("generic iterator next fast path failed: %v", err)
	}
	if !valuesEqual(vm.stack[0], runtime.NewSmallInt(7, runtime.IntegerI32)) {
		t.Fatalf("next result = %#v, want i32 7", vm.stack[0])
	}

	vm.stack = []runtime.Value{iter}
	_, err = vm.execCallMemberNext(bytecodeInstruction{name: "next", argCount: 0}, nil)
	if err != nil {
		t.Fatalf("generic iterator next end fast path failed: %v", err)
	}
	if _, ok := vm.stack[0].(runtime.IteratorEndValue); !ok {
		t.Fatalf("next end result = %#v, want IteratorEnd", vm.stack[0])
	}
}

func TestBytecodeVM_CallMemberGenericIteratorNextFastPathWrapsNilValue(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	iter := runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
		return nil, false, nil
	}, nil)

	vm.stack = []runtime.Value{iter}
	_, err := vm.execCallMemberNext(bytecodeInstruction{name: "next", argCount: 0}, nil)
	if err != nil {
		t.Fatalf("generic iterator nil fast path failed: %v", err)
	}
	if _, ok := vm.stack[0].(runtime.NilValue); !ok {
		t.Fatalf("next nil result = %#v, want NilValue", vm.stack[0])
	}
}

func TestBytecodeVM_CallMemberIteratorNextRawI64KeepsBytecodeCarrier(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	iter := runtime.NewIteratorValueWithRaw(func() (runtime.RawValue, bool, error) {
		return runtime.NewRawIntegerValue(runtime.IntegerI64, 99), false, nil
	}, nil)

	vm.stack = []runtime.Value{iter}
	_, err := vm.execCallMemberNext(bytecodeInstruction{name: "next", argCount: 0}, nil)
	if err != nil {
		t.Fatalf("generic iterator raw next fast path failed: %v", err)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI64 || raw != 99 {
		t.Fatalf("next raw result = %#v (%s, %d, %t), want raw i64 99", vm.stack[0], kind, raw, ok)
	}
	if _, boxed := vm.stack[0].(runtime.IntegerValue); boxed {
		t.Fatalf("next raw result should stay as a bytecode carrier, got boxed integer")
	}
}

func TestBytecodeVM_CallMemberInterfaceIteratorNextUsesFastPath(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	iter := runtime.NewIteratorValueWithRaw(func() (runtime.RawValue, bool, error) {
		return runtime.NewRawIntegerValue(runtime.IntegerI64, 77), false, nil
	}, nil)
	iface := &runtime.InterfaceValue{
		Underlying:    iter,
		SharedMethods: map[string]runtime.Value{"next": iteratorNextNativeMethod()},
	}

	vm.stack = []runtime.Value{iface}
	_, err := vm.execCallMemberNext(bytecodeInstruction{name: "next", argCount: 0}, nil)
	if err != nil {
		t.Fatalf("interface iterator next fast path failed: %v", err)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI64 || raw != 77 {
		t.Fatalf("interface next raw result = %#v (%s, %d, %t), want raw i64 77", vm.stack[0], kind, raw, ok)
	}
}

func TestBytecodeVM_CallMemberInterfaceIteratorNextFastPathRequiresIteratorNative(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	iter := runtime.NewIteratorValueWithRaw(func() (runtime.RawValue, bool, error) {
		return runtime.NewRawIntegerValue(runtime.IntegerI64, 77), false, nil
	}, nil)
	iface := &runtime.InterfaceValue{
		Underlying: iter,
		SharedMethods: map[string]runtime.Value{
			"next": runtime.NativeFunctionValue{
				Name:       "custom.next",
				Arity:      0,
				BorrowArgs: true,
				RawImpl: func(_ *runtime.NativeCallContext, _ []runtime.RawValue) (runtime.RawValue, error) {
					return runtime.NewRawIntegerValue(runtime.IntegerI64, 1), nil
				},
			},
		},
	}

	vm.stack = []runtime.Value{iface}
	_, handled, err := vm.execIteratorNextCallMemberFast(bytecodeInstruction{name: "next", argCount: 0}, 0, nil)
	if err != nil {
		t.Fatalf("interface iterator next fast-path rejection failed: %v", err)
	}
	if handled {
		t.Fatalf("interface iterator next fast path handled non-iterator native override")
	}
}

func TestBytecodeVM_CallMemberNextUsesCachedDirectMember(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{}
	receiver := bytecodeNextCacheTestReceiver(runtime.NewSmallInt(0, runtime.IntegerI32))
	native := runtime.NativeFunctionValue{
		Name:       "box.next",
		Arity:      0,
		BorrowArgs: true,
		RawImpl: func(_ *runtime.NativeCallContext, args []runtime.RawValue) (runtime.RawValue, error) {
			if len(args) != 1 {
				t.Fatalf("cached next raw args = %d, want receiver only", len(args))
			}
			return runtime.NewRawIntegerValue(runtime.IntegerI64, 5), nil
		},
	}
	if _, ok := vm.storeCachedMemberMethod(
		program,
		3,
		"next",
		true,
		receiver,
		runtime.NativeBoundMethodValue{Receiver: receiver, Method: native},
	); !ok {
		t.Fatalf("expected next member-method cache store to succeed")
	}

	vm.ip = 3
	vm.stack = []runtime.Value{receiver}
	_, err := vm.execCallMemberNext(bytecodeInstruction{name: "next", argCount: 0}, program)
	if err != nil {
		t.Fatalf("cached direct next fast path failed: %v", err)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI64 || raw != 5 {
		t.Fatalf("cached next result = %#v (%s, %d, %t), want raw i64 5", vm.stack[0], kind, raw, ok)
	}
}

func TestBytecodeVM_CallMemberNextCacheHonorsCallableFieldShadow(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	program := &bytecodeProgram{}
	receiver := bytecodeNextCacheTestReceiver(runtime.NewSmallInt(0, runtime.IntegerI32))
	cachedNative := runtime.NativeFunctionValue{
		Name:       "box.next.cached",
		Arity:      0,
		BorrowArgs: true,
		RawImpl: func(_ *runtime.NativeCallContext, _ []runtime.RawValue) (runtime.RawValue, error) {
			return runtime.NewRawIntegerValue(runtime.IntegerI64, 5), nil
		},
	}
	if _, ok := vm.storeCachedMemberMethod(
		program,
		4,
		"next",
		true,
		receiver,
		runtime.NativeBoundMethodValue{Receiver: receiver, Method: cachedNative},
	); !ok {
		t.Fatalf("expected next member-method cache store to succeed")
	}
	receiver.Fields["next"] = runtime.NativeFunctionValue{
		Name:       "box.next.field",
		Arity:      0,
		BorrowArgs: true,
		RawImpl: func(_ *runtime.NativeCallContext, args []runtime.RawValue) (runtime.RawValue, error) {
			if len(args) != 0 {
				t.Fatalf("callable field next raw args = %d, want no explicit args", len(args))
			}
			return runtime.NewRawIntegerValue(runtime.IntegerI64, 9), nil
		},
	}

	vm.ip = 4
	vm.stack = []runtime.Value{receiver}
	_, err := vm.execCallMemberNext(bytecodeInstruction{name: "next", argCount: 0}, program)
	if err != nil {
		t.Fatalf("callable field next fallback failed: %v", err)
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerI64 || raw != 9 {
		t.Fatalf("callable field next result = %#v (%s, %d, %t), want raw i64 9", vm.stack[0], kind, raw, ok)
	}
}

func bytecodeNextCacheTestReceiver(nextField runtime.Value) *runtime.StructInstanceValue {
	structDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "next"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	return &runtime.StructInstanceValue{
		Definition: structDef,
		Fields: map[string]runtime.Value{
			"next": nextField,
		},
	}
}

func TestRuntimeIteratorNextMaterializesRawI64ForOrdinaryCallers(t *testing.T) {
	iter := runtime.NewIteratorValueWithRaw(func() (runtime.RawValue, bool, error) {
		return runtime.NewRawIntegerValue(runtime.IntegerI64, 123), false, nil
	}, nil)

	value, done, err := iter.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if done {
		t.Fatalf("Next returned done for raw value")
	}
	intVal, ok := value.(runtime.IntegerValue)
	if !ok || !intVal.IsSmall() || intVal.TypeSuffix != runtime.IntegerI64 || intVal.Int64Fast() != 123 {
		t.Fatalf("ordinary Next value = %#v, want materialized small i64 123", value)
	}
}
