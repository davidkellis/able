package interpreter

import (
	"fmt"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_CallMemberOpcodeExecutesMethodCall(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Counter",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Counter"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"add",
					[]*ast.FunctionParameter{
						ast.Param("delta", ast.Ty("i32")),
					},
					[]ast.Statement{
						ast.Ret(ast.Bin("+", ast.ImplicitMember("value"), ast.ID("delta"))),
					},
					ast.Ty("i32"),
					nil,
					nil,
					true,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("c"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(3), "value"),
				},
				false,
				"Counter",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("c"), "add"), ast.Int(4)),
	}, nil, nil)

	byteInterp := NewBytecode()
	program, err := byteInterp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpCallStaticMember {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode static-candidate member opcode not emitted")
	}

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode method call mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_CallMemberArraySlotFallsBackForNonArrayPushMethod(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Box"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"push",
					[]*ast.FunctionParameter{
						ast.Param("delta", ast.Ty("i32")),
					},
					[]ast.Statement{
						ast.Ret(ast.Bin("+", ast.ImplicitMember("value"), ast.ID("delta"))),
					},
					ast.Ty("i32"),
					nil,
					nil,
					true,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("box"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(3), "value"),
				},
				false,
				"Box",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("box"), "push"), ast.Int(4)),
	}, nil, nil)

	byteInterp := NewBytecode()
	program, err := byteInterp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpCallMemberArraySlot && instr.name == "push" && instr.argCount == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode array-slot candidate opcode not emitted for push-shaped member call")
	}

	want := mustEvalModule(t, New(), module)
	byteInterp.bytecodeStatsEnabled = true
	byteInterp.ResetBytecodeStats()
	vm := newBytecodeVM(byteInterp, byteInterp.GlobalEnvironment())
	got, err := vm.run(program)
	if err != nil {
		t.Fatalf("bytecode execution failed: %v", err)
	}
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode non-array push fallback mismatch: got=%#v want=%#v", got, want)
	}
	stats := byteInterp.BytecodeStats()
	if stats.ArrayMemberSlotLookups != 1 || stats.ArrayMemberSlotReceiverMiss != 1 || stats.ArrayMemberSlotFallbacks != 1 {
		t.Fatalf("array slot non-array fallback stats = lookups %d receiver %d fallbacks %d, want 1/1/1",
			stats.ArrayMemberSlotLookups,
			stats.ArrayMemberSlotReceiverMiss,
			stats.ArrayMemberSlotFallbacks,
		)
	}
}

func TestBytecodeVM_CallMemberFallsBackToCallableField(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("String"), "name"),
				ast.FieldDef(ast.FnType(nil, ast.Ty("String")), "action"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Box"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"action",
					nil,
					[]ast.Statement{ast.Ret(ast.Str("method"))},
					ast.Ty("String"),
					nil,
					nil,
					true,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("b"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Str("ok"), "name"),
					ast.FieldInit(ast.Lam(nil, ast.Str("field")), "action"),
				},
				false,
				"Box",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("b"), "action")),
	}, nil, nil)

	byteInterp := NewBytecode()
	program, err := byteInterp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpCallStaticMember {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode static-candidate member opcode not emitted for callable field fallback")
	}

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode callable field fallback mismatch: got=%#v want=%#v", got, want)
	}
	str, ok := got.(runtime.StringValue)
	if !ok || str.Val != "field" {
		t.Fatalf("expected callable field to win, got %T (%#v)", got, got)
	}
}

func TestBytecodeVM_CallMemberNativeCallableFieldDoesNotInjectReceiver(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	receiver := &runtime.StructInstanceValue{
		Fields: map[string]runtime.Value{
			"probe": runtime.NativeFunctionValue{
				Name:       "probe",
				Arity:      1,
				BorrowArgs: true,
				Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("probe args = %d, want 1", len(args))
					}
					delta, ok := args[0].(runtime.IntegerValue)
					if !ok {
						return nil, fmt.Errorf("delta type = %T", args[0])
					}
					value, ok := delta.ToInt64()
					if !ok {
						return nil, fmt.Errorf("delta not small int: %#v", delta)
					}
					return runtime.NewSmallInt(value+1, runtime.IntegerI32), nil
				},
			},
		},
	}
	vm.stack = []runtime.Value{receiver, runtime.NewSmallInt(5, runtime.IntegerI32)}

	newProg, err := vm.execCallMember(bytecodeInstruction{name: "probe", argCount: 1}, &bytecodeProgram{})
	if err != nil {
		t.Fatalf("bytecode native callable-field call failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("unexpected inline program for native callable-field call")
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack len = %d, want 1", len(vm.stack))
	}
	if !valuesEqual(vm.stack[0], runtime.NewSmallInt(6, runtime.IntegerI32)) {
		t.Fatalf("stack result = %#v, want 6", vm.stack[0])
	}
}

func TestBytecodeVM_CallMemberHandlesOptionalMethodArity(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Greeter",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("String"), "name"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Greeter"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"suffix",
					[]*ast.FunctionParameter{
						ast.Param("value", ast.Nullable(ast.Ty("String"))),
					},
					[]ast.Statement{ast.Ret(ast.Str("ok"))},
					ast.Ty("String"),
					nil,
					nil,
					true,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("g"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Str("hi"), "name"),
				},
				false,
				"Greeter",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("g"), "suffix")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, NewBytecode(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode optional method call mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_FinishCompletedCallSnapshotsRawIntegerResult(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())

	source := &bytecodeRawIntegerSlotCell{Raw: 7, TypeSuffix: runtime.IntegerU32}
	if _, err := vm.finishCompletedCall(source, nil, nil, nil); err != nil {
		t.Fatalf("finishCompletedCall failed: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack len = %d, want 1", len(vm.stack))
	}
	source.Raw = 99
	kind, raw, ok := bytecodeRawIntegerValueInfo(vm.stack[0])
	if !ok || kind != runtime.IntegerU32 || raw != 7 {
		t.Fatalf("completed call result = %#v, want raw u32 7 snapshot", vm.stack[0])
	}
	got, err := arrayIndexFromValue(vm.stack[0])
	if err != nil || got != 7 {
		t.Fatalf("completed call result = %#v (%v), want 7", vm.stack[0], err)
	}
}

func TestBytecodeVM_CallMemberSupportsPartialMethodApplication(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Adder",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "base"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Adder"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"add",
					[]*ast.FunctionParameter{
						ast.Param("self", ast.Ty("Self")),
						ast.Param("delta", ast.Ty("i32")),
						ast.Param("extra", ast.Ty("i32")),
					},
					[]ast.Statement{
						ast.Ret(
							ast.Bin(
								"+",
								ast.Bin("+", ast.Member(ast.ID("self"), "base"), ast.ID("delta")),
								ast.ID("extra"),
							),
						),
					},
					ast.Ty("i32"),
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("a"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(3), "base"),
				},
				false,
				"Adder",
				nil,
				nil,
			),
		),
		ast.Assign(ast.ID("step"), ast.CallExpr(ast.Member(ast.ID("a"), "add"), ast.Int(4))),
		ast.CallExpr(ast.ID("step"), ast.Int(5)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, NewBytecode(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode partial method call mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_CallMemberHandlesOverloadedMethods(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Printer",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("String"), "name"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Printer"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"render",
					[]*ast.FunctionParameter{
						ast.Param("value", ast.Ty("i32")),
					},
					[]ast.Statement{ast.Ret(ast.Str("int"))},
					ast.Ty("String"),
					nil,
					nil,
					true,
					false,
				),
				ast.Fn(
					"render",
					[]*ast.FunctionParameter{
						ast.Param("value", ast.Ty("String")),
					},
					[]ast.Statement{ast.Ret(ast.Str("string"))},
					ast.Ty("String"),
					nil,
					nil,
					true,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("p"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Str("hi"), "name"),
				},
				false,
				"Printer",
				nil,
				nil,
			),
		),
		ast.CallExpr(ast.Member(ast.ID("p"), "render"), ast.Int(4)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, NewBytecode(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode overloaded method call mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeResolveExactInjectedNativeCallTargetAcceptsNativeBoundMethod(t *testing.T) {
	receiver := runtime.StringValue{Val: "word"}
	native := runtime.NativeFunctionValue{
		Name:       "len_like",
		Arity:      0,
		BorrowArgs: true,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return runtime.NewSmallInt(int64(len(args)), runtime.IntegerI32), nil
		},
	}
	target, ok := bytecodeResolveExactInjectedNativeCallTarget(
		runtime.NativeBoundMethodValue{Receiver: receiver, Method: native},
		runtime.NilValue{},
		0,
	)
	if !ok {
		t.Fatalf("expected native bound method to resolve as exact injected native target")
	}
	if !target.hasReceiver {
		t.Fatalf("expected exact target to carry injected receiver")
	}
	if !valuesEqual(target.injectedReceiver, receiver) {
		t.Fatalf("unexpected injected receiver: got=%#v want=%#v", target.injectedReceiver, receiver)
	}
	if target.native.Name != native.Name || target.native.Arity != native.Arity || !target.native.BorrowArgs {
		t.Fatalf("unexpected native target: %#v", target.native)
	}
}

func TestBytecodeVM_CallMemberOnInterfaceValueSkipsBoundWrapper(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironmentWithValueCapacity(interp.GlobalEnvironment(), 0)
	receiver := runtime.StringValue{Val: "beta"}
	iface := &runtime.InterfaceValue{
		Interface: &runtime.InterfaceDefinitionValue{
			Node: ast.Iface(
				"Probe",
				[]*ast.FunctionSignature{
					ast.FnSig(
						"probe",
						[]*ast.FunctionParameter{ast.Param("delta", ast.Ty("i32"))},
						ast.Ty("i32"),
						nil,
						nil,
						nil,
					),
				},
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
		Underlying: receiver,
		Methods: map[string]runtime.Value{
			"probe": runtime.NativeFunctionValue{
				Name:       "probe",
				Arity:      1,
				BorrowArgs: true,
				Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("probe args = %d, want 2", len(args))
					}
					if !valuesEqual(args[0], receiver) {
						return nil, fmt.Errorf("receiver = %#v, want %#v", args[0], receiver)
					}
					delta, ok := args[1].(runtime.IntegerValue)
					if !ok {
						return nil, fmt.Errorf("delta type = %T", args[1])
					}
					value, ok := delta.ToInt64()
					if !ok {
						return nil, fmt.Errorf("delta not small int: %#v", delta)
					}
					return runtime.NewSmallInt(value+1, runtime.IntegerI32), nil
				},
			},
		},
	}
	vm := newBytecodeVM(interp, env)
	vm.stack = []runtime.Value{iface, runtime.NewSmallInt(5, runtime.IntegerI32)}

	newProg, err := vm.execCallMember(bytecodeInstruction{name: "probe", argCount: 1}, &bytecodeProgram{})
	if err != nil {
		t.Fatalf("bytecode direct interface member call failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("unexpected inline program for direct interface member call")
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack len = %d, want 1", len(vm.stack))
	}
	if !valuesEqual(vm.stack[0], runtime.NewSmallInt(6, runtime.IntegerI32)) {
		t.Fatalf("stack result = %#v, want 6", vm.stack[0])
	}
	if iface.BoundMethod != nil || iface.BoundMethodName != "" {
		t.Fatalf("bytecode direct interface member call should not materialize bound cache, got %q %#v", iface.BoundMethodName, iface.BoundMethod)
	}
}

func TestBytecodeVM_ExecCachedResolvedMemberCallInlinesDirectTemplateWithTypeArgs(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	structDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	receiver := &runtime.StructInstanceValue{
		Definition: structDef,
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	methodDef := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Box")),
			ast.Param("x", ast.Ty("T")),
		},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	methodProgram, err := interp.lowerFunctionDefinitionBytecode(methodDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	methodFn := &runtime.FunctionValue{
		Declaration: methodDef,
		Closure:     env,
	}
	setFunctionBytecodeProgram(methodFn, methodProgram)

	vm := newBytecodeVM(interp, env)
	vm.ip = 4
	vm.stack = append(vm.stack, receiver, runtime.NewSmallInt(41, runtime.IntegerI32))
	callNode := ast.NewFunctionCall(ast.Member(ast.ID("box"), "apply"), nil, []ast.TypeExpression{ast.Ty("i32")}, false)

	newProg, err := vm.execCachedResolvedMemberCall(
		bytecodeCachedMemberMethod{
			template: methodFn,
			dispatch: bytecodeMemberMethodDispatchInline,
			inlineFn: methodFn,
		},
		"apply",
		0,
		1,
		1,
		callNode,
		nil,
	)
	if err != nil {
		t.Fatalf("cached resolved member call failed: %v", err)
	}
	if newProg != methodProgram {
		t.Fatalf("expected resolved-function member call to inline, got program %#v want %#v", newProg, methodProgram)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("stack size after inline setup = %d, want 0", len(vm.stack))
	}
	if _, err := vm.runResumable(newProg, true); err != nil {
		t.Fatalf("resumed inline member call failed: %v", err)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack size after resumed inline member call = %d, want 1", len(vm.stack))
	}
	gotValue := vm.stack[0]
	got, ok := gotValue.(runtime.IntegerValue)
	gotVal, gotOK := got.ToInt64()
	if !ok || !gotOK || got.TypeSuffix != runtime.IntegerI32 || gotVal != 41 {
		t.Fatalf("inline member result = %#v, want i32 41", gotValue)
	}
	stats := interp.BytecodeStats()
	if stats.InlineCallHits != 1 {
		t.Fatalf("expected one inline hit for cached direct-template member call, got %d", stats.InlineCallHits)
	}
	if stats.InlineCallMisses != 0 {
		t.Fatalf("expected no inline misses for cached direct-template member call, got %d", stats.InlineCallMisses)
	}
	if stats.CallMemberResolvedInlineHits != 1 {
		t.Fatalf("expected one resolved-member inline hit, got %d", stats.CallMemberResolvedInlineHits)
	}
}

func TestBytecodeVM_ExecCachedResolvedMemberCallDirectTemplatePreservesPartialApplication(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	structDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	receiver := &runtime.StructInstanceValue{
		Definition: structDef,
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	methodDef := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Box")),
			ast.Param("x", ast.Ty("T")),
			ast.Param("y", ast.Ty("i32")),
		},
		[]ast.Statement{ast.Int(0)},
		ast.Ty("i32"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	methodProgram, err := interp.lowerFunctionDefinitionBytecode(methodDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	methodFn := &runtime.FunctionValue{
		Declaration: methodDef,
		Closure:     env,
	}
	setFunctionBytecodeProgram(methodFn, methodProgram)

	vm := newBytecodeVM(interp, env)
	vm.ip = 4
	explicitArg := runtime.NewSmallInt(41, runtime.IntegerI32)
	vm.stack = append(vm.stack, receiver, explicitArg)

	newProg, err := vm.execCachedResolvedMemberCall(
		bytecodeCachedMemberMethod{
			template: methodFn,
			dispatch: bytecodeMemberMethodDispatchInline,
			inlineFn: methodFn,
		},
		"apply",
		0,
		1,
		1,
		ast.NewFunctionCall(ast.Member(ast.ID("box"), "apply"), nil, nil, false),
		nil,
	)
	if err != nil {
		t.Fatalf("cached resolved member call failed: %v", err)
	}
	if newProg != nil {
		t.Fatalf("expected partial-application fallback to complete without switching programs")
	}
	if vm.ip != 5 {
		t.Fatalf("vm ip = %d, want 5", vm.ip)
	}
	if len(vm.stack) != 1 {
		t.Fatalf("stack size = %d, want 1", len(vm.stack))
	}
	partial, ok := vm.stack[0].(*runtime.PartialFunctionValue)
	if !ok || partial == nil {
		t.Fatalf("stack result = %#v, want partial function", vm.stack[0])
	}
	if partial.Target != methodFn {
		t.Fatalf("partial target = %#v, want %#v", partial.Target, methodFn)
	}
	if len(partial.BoundArgs) != 2 {
		t.Fatalf("partial bound arg count = %d, want 2", len(partial.BoundArgs))
	}
	if !valuesEqual(partial.BoundArgs[0], receiver) {
		t.Fatalf("partial receiver = %#v, want %#v", partial.BoundArgs[0], receiver)
	}
	if !valuesEqual(partial.BoundArgs[1], explicitArg) {
		t.Fatalf("partial explicit arg = %#v, want %#v", partial.BoundArgs[1], explicitArg)
	}
}

func TestBytecodeVM_TryCallResolvedCallableFromMemberStackHandlesDirectFunction(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	structDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	receiver := &runtime.StructInstanceValue{
		Definition: structDef,
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	methodDef := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Box")),
			ast.Param("x", ast.Ty("i32")),
		},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	methodProgram, err := interp.lowerFunctionDefinitionBytecode(methodDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	methodFn := &runtime.FunctionValue{
		Declaration: methodDef,
		Closure:     env,
	}
	setFunctionBytecodeProgram(methodFn, methodProgram)
	explicitArg := runtime.NewSmallInt(41, runtime.IntegerI32)
	vm := newBytecodeVM(interp, env)
	vm.stack = make([]runtime.Value, 0, 2)

	vm.stack = append(vm.stack, receiver, explicitArg)
	result, handled, err := vm.tryCallResolvedCallableFromMemberStack(methodFn, receiver, 0, 1, 1, nil)
	if err != nil {
		t.Fatalf("warmup direct member-stack call failed: %v", err)
	}
	if !handled || !valuesEqual(result, explicitArg) {
		t.Fatalf("direct member-stack call = (%#v, %t), want (%#v, true)", result, handled, explicitArg)
	}
	if len(vm.stack) != 0 {
		t.Fatalf("stack size after handled member-stack call = %d, want 0", len(vm.stack))
	}
}

func TestBytecodeVM_TryCallResolvedCallableFromMemberStackSkipsMismatchedInjectedReceiver(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	structDef := &runtime.StructDefinitionValue{
		Node: ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
	}
	receiver := &runtime.StructInstanceValue{
		Definition: structDef,
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(7, runtime.IntegerI32),
		},
	}
	otherReceiver := &runtime.StructInstanceValue{
		Definition: structDef,
		Fields: map[string]runtime.Value{
			"value": runtime.NewSmallInt(9, runtime.IntegerI32),
		},
	}
	methodDef := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Box")),
			ast.Param("x", ast.Ty("i32")),
		},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	methodProgram, err := interp.lowerFunctionDefinitionBytecode(methodDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	methodFn := &runtime.FunctionValue{
		Declaration: methodDef,
		Closure:     env,
	}
	setFunctionBytecodeProgram(methodFn, methodProgram)
	callable := runtime.BoundMethodValue{Receiver: otherReceiver, Method: methodFn}
	explicitArg := runtime.NewSmallInt(41, runtime.IntegerI32)
	vm := newBytecodeVM(interp, env)
	vm.stack = append(vm.stack, receiver, explicitArg)

	result, handled, err := vm.tryCallResolvedCallableFromMemberStack(callable, receiver, 0, 1, 1, nil)
	if err != nil {
		t.Fatalf("mismatched direct member-stack call failed: %v", err)
	}
	if handled {
		t.Fatalf("expected mismatched injected receiver to skip stack-window path, got result %#v", result)
	}
	if len(vm.stack) != 2 {
		t.Fatalf("stack size after skipped member-stack call = %d, want 2", len(vm.stack))
	}
	if !valuesEqual(vm.stack[0], receiver) || !valuesEqual(vm.stack[1], explicitArg) {
		t.Fatalf("stack contents changed after skipped member-stack call: %#v", vm.stack)
	}
}

func TestBytecodeVM_TryCallResolvedCallableFromMemberStackPreservesRawI32Arg(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	methodDef := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("i32")),
			ast.Param("x", ast.Ty("i32")),
		},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	methodProgram, err := interp.lowerFunctionDefinitionBytecode(methodDef)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	methodFn := &runtime.FunctionValue{
		Declaration: methodDef,
		Closure:     env,
	}
	setFunctionBytecodeProgram(methodFn, methodProgram)
	vm := newBytecodeVM(interp, env)
	rawReceiver := vm.stackRawI32Value(0, 7)
	rawArg := vm.stackRawI32Value(1, int32(bytecodeRawI32SlotCacheMax+17))
	vm.stack = append(vm.stack, rawReceiver, rawArg)

	result, handled, err := vm.tryCallResolvedCallableFromMemberStack(methodFn, rawReceiver, 0, 1, 1, nil)
	if err != nil {
		t.Fatalf("raw member-stack call failed: %v", err)
	}
	if !handled {
		t.Fatalf("expected raw member-stack call to be handled")
	}
	kind, raw, ok := bytecodeRawIntegerValueInfo(result)
	if !ok || kind != runtime.IntegerI32 || raw != int64(bytecodeRawI32SlotCacheMax+17) {
		t.Fatalf("raw member-stack result = %#v, want raw i32 arg passthrough", result)
	}
}

func TestBytecodeVM_ResolveConcreteMemberOverload_BoundDirectFunctionSkipsOverloadView(t *testing.T) {
	interp := NewBytecode()
	env := runtime.NewEnvironment(nil)
	vm := newBytecodeVM(interp, env)
	methodDef := ast.Fn(
		"apply",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Box")),
			ast.Param("x", ast.Ty("i32")),
		},
		[]ast.Statement{ast.ID("x")},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	methodFn := &runtime.FunctionValue{Declaration: methodDef, Closure: env}
	boundReceiver := &runtime.StructInstanceValue{}
	otherReceiver := &runtime.StructInstanceValue{}
	callable := runtime.BoundMethodValue{Receiver: boundReceiver, Method: methodFn}

	selected, injectedReceiver, ok, err := vm.resolveConcreteMemberOverload(
		callable,
		otherReceiver,
		[]runtime.Value{runtime.NewSmallInt(41, runtime.IntegerI32)},
		nil,
	)
	if err != nil {
		t.Fatalf("resolve concrete member overload failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected direct bound function overload resolution")
	}
	if selected != methodFn {
		t.Fatalf("selected function = %#v, want %#v", selected, methodFn)
	}
	if injectedReceiver != boundReceiver {
		t.Fatalf("injected receiver = %#v, want bound receiver %#v", injectedReceiver, boundReceiver)
	}
}
