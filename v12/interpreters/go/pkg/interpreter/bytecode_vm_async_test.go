package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_SpawnExpression(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(3)),
		ast.Assign(ast.ID("handle"), ast.Spawn(ast.Block(
			ast.Bin("+", ast.ID("x"), ast.Int(4)),
		))),
		ast.CallExpr(ast.Member(ast.ID("handle"), "value")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode spawn mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_AwaitExpressionManualWaker(t *testing.T) {
	manualAwaitableDef := ast.StructDef("ManualAwaitable", []*ast.StructFieldDefinition{
		ast.FieldDef(ast.Ty("bool"), "ready"),
		ast.FieldDef(ast.Ty("i32"), "value"),
		ast.FieldDef(ast.Nullable(ast.Ty("AwaitWaker")), "waker"),
	}, ast.StructKindNamed, nil, nil, false)

	manualRegistrationDef := ast.StructDef("ManualRegistration", nil, ast.StructKindNamed, nil, nil, false)

	isReadyFn := ast.Fn(
		"is_ready",
		nil,
		[]ast.Statement{
			ast.Ret(ast.ImplicitMember(ast.ID("ready"))),
		},
		ast.Ty("bool"),
		nil,
		nil,
		true,
		false,
	)

	registerFn := ast.Fn(
		"register",
		[]*ast.FunctionParameter{
			ast.Param("waker", ast.Ty("AwaitWaker")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ImplicitMember(ast.ID("waker")), ast.ID("waker")),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("last_waker"), ast.ID("waker")),
			ast.Ret(ast.StructLit(nil, false, "ManualRegistration", nil, nil)),
		},
		ast.Ty("ManualRegistration"),
		nil,
		nil,
		true,
		false,
	)

	commitFn := ast.Fn(
		"commit",
		nil,
		[]ast.Statement{
			ast.Ret(ast.ImplicitMember(ast.ID("value"))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		true,
		false,
	)

	setReadyFn := ast.Fn(
		"set_ready",
		[]*ast.FunctionParameter{
			ast.Param("ready", ast.Ty("bool")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ImplicitMember(ast.ID("ready")), ast.ID("ready")),
			ast.Ret(ast.Nil()),
		},
		ast.Ty("void"),
		nil,
		nil,
		true,
		false,
	)

	cancelFn := ast.Fn(
		"cancel",
		nil,
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ImplicitMember(ast.ID("waker")), ast.Nil()),
			ast.Ret(ast.Nil()),
		},
		ast.Ty("void"),
		nil,
		nil,
		true,
		false,
	)

	manualArm := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Bool(false), "ready"),
			ast.FieldInit(ast.Int(42), "value"),
			ast.FieldInit(ast.Nil(), "waker"),
		},
		false,
		"ManualAwaitable",
		nil,
		nil,
	)

	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("last_waker"), ast.Nil()),
		manualAwaitableDef,
		manualRegistrationDef,
		ast.Methods(ast.Ty("ManualAwaitable"), []*ast.FunctionDefinition{
			isReadyFn,
			registerFn,
			commitFn,
			setReadyFn,
		}, nil, nil),
		ast.Methods(ast.Ty("ManualRegistration"), []*ast.FunctionDefinition{
			cancelFn,
		}, nil, nil),
		ast.Assign(ast.ID("arm"), manualArm),
		ast.Assign(ast.ID("result"), ast.Int(0)),
		ast.Assign(ast.ID("handle"), ast.Spawn(ast.Block(
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("result"),
				ast.Await(ast.Arr(ast.ID("arm"))),
			),
			ast.Nil(),
		))),
		ast.Call("future_flush"),
		ast.CallExpr(ast.Member(ast.ID("arm"), "set_ready"), ast.Bool(true)),
		ast.CallExpr(ast.Member(ast.ID("last_waker"), "wake")),
		ast.Call("future_flush"),
		ast.ID("result"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode await mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ImplicitMemberExpression(t *testing.T) {
	pointDef := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	sumFn := ast.Fn(
		"sum",
		nil,
		[]ast.Statement{
			ast.Ret(ast.Bin("+", ast.ImplicitMember(ast.ID("x")), ast.ImplicitMember(ast.ID("y")))),
		},
		ast.Ty("i32"),
		nil,
		nil,
		true,
		false,
	)
	pointLit := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(2), "x"),
			ast.FieldInit(ast.Int(5), "y"),
		},
		false,
		"Point",
		nil,
		nil,
	)
	module := ast.Mod([]ast.Statement{
		pointDef,
		ast.Methods(ast.Ty("Point"), []*ast.FunctionDefinition{sumFn}, nil, nil),
		ast.Assign(ast.ID("p"), pointLit),
		ast.CallExpr(ast.Member(ast.ID("p"), "sum")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode implicit member mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_IteratorLiteral(t *testing.T) {
	iterLit := ast.IteratorLit(
		ast.Yield(ast.Int(1)),
		ast.Yield(ast.Int(2)),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("iter"), iterLit),
		ast.Assign(ast.ID("first"), ast.CallExpr(ast.Member(ast.ID("iter"), "next"))),
		ast.Assign(ast.ID("second"), ast.CallExpr(ast.Member(ast.ID("iter"), "next"))),
		ast.Bin("+", ast.ID("first"), ast.ID("second")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode iterator literal mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_BreakpointExpression(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Bp("exit", ast.Brk("exit", ast.Int(9)), ast.Int(0)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode breakpoint mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_PlaceholderLambda(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("swap"),
			ast.Bin("+", ast.Bin("*", ast.PlaceholderN(2), ast.Int(100)), ast.PlaceholderN(1)),
		),
		ast.CallExpr(ast.ID("swap"), ast.Int(7), ast.Int(9)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode placeholder lambda mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ShortCircuitAndOr(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(0)),
		ast.Bin("&&", ast.Bool(false), ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(1))),
		ast.Bin("||", ast.Bool(true), ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(2))),
		ast.ID("x"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode short-circuit mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_PipeOperator(t *testing.T) {
	add := ast.Fn(
		"add",
		[]*ast.FunctionParameter{
			ast.Param("a", nil),
			ast.Param("b", nil),
		},
		[]ast.Statement{
			ast.Bin("+", ast.ID("a"), ast.ID("b")),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{
		add,
		ast.Bin("|>", ast.Int(5), ast.Call("add", ast.Int(3))),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode pipe mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_FunctionDefinitionAndCall(t *testing.T) {
	fn := ast.Fn(
		"add",
		[]*ast.FunctionParameter{
			ast.Param("a", nil),
			ast.Param("b", nil),
		},
		[]ast.Statement{
			ast.Bin("+", ast.ID("a"), ast.ID("b")),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)
	module := ast.Mod([]ast.Statement{
		fn,
		ast.Call("add", ast.Int(2), ast.Int(3)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode function call mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_LambdaCall(t *testing.T) {
	lambda := ast.Lam(
		[]*ast.FunctionParameter{ast.Param("x", nil)},
		ast.Bin("+", ast.ID("x"), ast.Int(1)),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("inc"), lambda),
		ast.CallExpr(ast.ID("inc"), ast.Int(4)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode lambda call mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ReturnStatement(t *testing.T) {
	fn := ast.Fn("early", nil, []ast.Statement{
		ast.Ret(ast.Int(7)),
		ast.Int(9),
	}, nil, nil, nil, false, false)
	module := ast.Mod([]ast.Statement{
		fn,
		ast.Call("early"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode return mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ReturnOutsideFunction(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Ret(ast.Int(1)),
	}, nil, nil)

	interp := NewBytecode()
	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected return outside function error")
	}
	if err.Error() != "return outside function" {
		t.Fatalf("unexpected return error: %v", err)
	}
}

func TestBytecodeVM_MemberAccess(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Member(ast.ID("point"), "x"),
	}, nil, nil)

	makePoint := func() *runtime.StructInstanceValue {
		return &runtime.StructInstanceValue{
			Fields: map[string]runtime.Value{
				"x": runtime.IntegerValue{Val: big.NewInt(7), TypeSuffix: runtime.IntegerI32},
			},
		}
	}

	treeInterp := New()
	treeInterp.GlobalEnvironment().Define("point", makePoint())
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := New()
	byteInterp.GlobalEnvironment().Define("point", makePoint())
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode member access mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_MemberAssignment(t *testing.T) {
	pointDef := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	pointLit := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(1), "x"),
			ast.FieldInit(ast.Int(2), "y"),
		},
		false,
		"Point",
		nil,
		nil,
	)
	module := ast.Mod([]ast.Statement{
		pointDef,
		ast.Assign(ast.ID("p"), pointLit),
		ast.AssignOp(ast.AssignmentAssign, ast.Member(ast.ID("p"), "x"), ast.Int(10)),
		ast.AssignOp(ast.AssignmentAdd, ast.Member(ast.ID("p"), "y"), ast.Int(5)),
		ast.Bin("+", ast.Member(ast.ID("p"), "x"), ast.Member(ast.ID("p"), "y")),
	}, nil, nil)

	byteInterp := NewBytecode()
	program, err := byteInterp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpMemberSet {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode member set opcode not emitted")
	}

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode member assignment mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_SafeMemberCallSkipsArgs(t *testing.T) {
	member := ast.NewMemberAccessExpression(ast.Nil(), ast.ID("noop"))
	member.Safe = true
	call := ast.CallExpr(member, ast.ID("boom"))
	module := ast.Mod([]ast.Statement{call}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode safe call mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_StructLiteral(t *testing.T) {
	pointDef := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
			ast.FieldDef(ast.Ty("i32"), "y"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	pointLit := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(3), "x"),
			ast.FieldInit(ast.Int(5), "y"),
		},
		false,
		"Point",
		nil,
		nil,
	)
	module := ast.Mod([]ast.Statement{
		pointDef,
		ast.Member(pointLit, "x"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode struct literal mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_IndexGet(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Index(ast.ID("arr"), ast.Int(1)),
	}, nil, nil)

	makeArray := func() *runtime.ArrayValue {
		return &runtime.ArrayValue{Elements: []runtime.Value{
			runtime.IntegerValue{Val: big.NewInt(1), TypeSuffix: runtime.IntegerI32},
			runtime.IntegerValue{Val: big.NewInt(4), TypeSuffix: runtime.IntegerI32},
		}}
	}

	treeInterp := New()
	treeInterp.GlobalEnvironment().Define("arr", makeArray())
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := New()
	byteInterp.GlobalEnvironment().Define("arr", makeArray())
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode index get mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_IndexAssign(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.AssignIndex(ast.ID("arr"), ast.Int(1), ast.Int(9)),
		ast.Index(ast.ID("arr"), ast.Int(1)),
	}, nil, nil)

	makeArray := func() *runtime.ArrayValue {
		return &runtime.ArrayValue{Elements: []runtime.Value{
			runtime.IntegerValue{Val: big.NewInt(1), TypeSuffix: runtime.IntegerI32},
			runtime.IntegerValue{Val: big.NewInt(4), TypeSuffix: runtime.IntegerI32},
		}}
	}

	treeInterp := New()
	treeInterp.GlobalEnvironment().Define("arr", makeArray())
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := New()
	byteInterp.GlobalEnvironment().Define("arr", makeArray())
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode index assign mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ArrayLiteral(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(2), ast.Int(4))),
		ast.Index(ast.ID("arr"), ast.Int(0)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode array literal mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_MapLiteral(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.MapLit([]ast.MapLiteralElement{
			ast.MapEntry(ast.Str("a"), ast.Int(1)),
			ast.MapEntry(ast.Str("b"), ast.Int(2)),
		}),
	}, nil, nil)

	treeInterp := New()
	loadKernelModule(t, treeInterp)
	seedHashMapStruct(t, treeInterp, treeInterp.GlobalEnvironment())
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := New()
	loadKernelModule(t, byteInterp)
	seedHashMapStruct(t, byteInterp, byteInterp.GlobalEnvironment())
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	expected := runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32}
	wantEntry := mapLiteralValue(t, treeInterp, want, runtime.StringValue{Val: "b"})
	gotEntry := mapLiteralValue(t, byteInterp, got, runtime.StringValue{Val: "b"})
	if !valuesEqual(wantEntry, expected) {
		t.Fatalf("tree-walker map literal entry mismatch: got=%#v want=%#v", wantEntry, expected)
	}
	if !valuesEqual(gotEntry, expected) {
		t.Fatalf("bytecode map literal entry mismatch: got=%#v want=%#v", gotEntry, expected)
	}
}

func TestBytecodeVM_MapLiteralSpread(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("defaults"), ast.MapLit([]ast.MapLiteralElement{
			ast.MapEntry(ast.Str("a"), ast.Int(1)),
			ast.MapEntry(ast.Str("b"), ast.Int(2)),
		})),
		ast.MapLit([]ast.MapLiteralElement{
			ast.MapEntry(ast.Str("b"), ast.Int(3)),
			ast.MapSpread(ast.ID("defaults")),
		}),
	}, nil, nil)

	treeInterp := New()
	loadKernelModule(t, treeInterp)
	seedHashMapStruct(t, treeInterp, treeInterp.GlobalEnvironment())
	want := mustEvalModule(t, treeInterp, module)

	byteInterp := New()
	loadKernelModule(t, byteInterp)
	seedHashMapStruct(t, byteInterp, byteInterp.GlobalEnvironment())
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	expected := runtime.IntegerValue{Val: big.NewInt(2), TypeSuffix: runtime.IntegerI32}
	wantEntry := mapLiteralValue(t, treeInterp, want, runtime.StringValue{Val: "b"})
	gotEntry := mapLiteralValue(t, byteInterp, got, runtime.StringValue{Val: "b"})
	if !valuesEqual(wantEntry, expected) {
		t.Fatalf("tree-walker map literal spread mismatch: got=%#v want=%#v", wantEntry, expected)
	}
	if !valuesEqual(gotEntry, expected) {
		t.Fatalf("bytecode map literal spread mismatch: got=%#v want=%#v", gotEntry, expected)
	}
}

func runBytecodeModule(t *testing.T, module *ast.Module) runtime.Value {
	t.Helper()
	interp := NewBytecode()
	return runBytecodeModuleWithInterpreter(t, interp, module)
}

func runBytecodeModuleWithInterpreter(t *testing.T, interp *Interpreter, module *ast.Module) runtime.Value {
	t.Helper()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	val, err := vm.run(program)
	if err != nil {
		t.Fatalf("bytecode execution failed: %v", err)
	}
	return val
}

func mapLiteralValue(t *testing.T, interp *Interpreter, value runtime.Value, key runtime.Value) runtime.Value {
	t.Helper()
	inst, ok := value.(*runtime.StructInstanceValue)
	if !ok {
		t.Fatalf("expected HashMap instance, got %T", value)
	}
	handle, err := interp.hashMapHandleFromInstance(inst)
	if err != nil {
		t.Fatalf("missing hash map handle: %v", err)
	}
	state, err := interp.hashMapStateForHandle(handle)
	if err != nil {
		t.Fatalf("missing hash map state: %v", err)
	}
	return mapStateValue(t, interp, state, key)
}
