package interpreter

import (
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
		if instr.op == bytecodeOpCallMember {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode call-member opcode not emitted")
	}

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode method call mismatch: got=%#v want=%#v", got, want)
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
		if instr.op == bytecodeOpCallMember {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode call-member opcode not emitted for callable field fallback")
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
