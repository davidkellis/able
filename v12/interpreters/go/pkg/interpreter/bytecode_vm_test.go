package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_AssignmentAndBinary(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(1)),
		ast.Assign(ast.ID("y"), ast.Int(2)),
		ast.Bin("+", ast.ID("x"), ast.ID("y")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode result mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_AssignmentPattern(t *testing.T) {
	pattern := ast.ArrP([]ast.Pattern{ast.ID("a"), ast.ID("b")}, nil)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1), ast.Int(2))),
		ast.Assign(pattern, ast.ID("arr")),
		ast.Bin("+", ast.ID("a"), ast.ID("b")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	byteInterp := NewBytecode()
	program, err := byteInterp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpAssignPattern {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode assign pattern opcode not emitted")
	}
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode assignment pattern mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_CompoundAssignmentName(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(1)),
		ast.AssignOp(
			ast.AssignmentAdd,
			ast.ID("x"),
			ast.AssignOp(ast.AssignmentAssign, ast.ID("x"), ast.Int(5)),
		),
		ast.ID("x"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	byteInterp := NewBytecode()
	program, err := byteInterp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpAssignNameCompound {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode compound assignment opcode not emitted")
	}
	got := runBytecodeModuleWithInterpreter(t, byteInterp, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode compound assignment mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_CompoundAssignmentPattern(t *testing.T) {
	pattern := ast.ArrP([]ast.Pattern{ast.ID("a"), ast.ID("b")}, nil)
	module := ast.Mod([]ast.Statement{
		ast.AssignOp(ast.AssignmentAdd, pattern, ast.Arr(ast.Int(1), ast.Int(2))),
	}, nil, nil)

	byteInterp := NewBytecode()
	program, err := byteInterp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	found := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpAssignPattern {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bytecode assign pattern opcode not emitted")
	}
	_, _, err = byteInterp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected compound assignment error for pattern")
	}
	if err.Error() != "compound assignment not supported with patterns" {
		t.Fatalf("unexpected pattern compound assignment error: %v", err)
	}
}

func TestBytecodeVM_ImportStatement(t *testing.T) {
	pkgModule := ast.Mod([]ast.Statement{
		ast.Fn("value", nil, []ast.Statement{ast.Ret(ast.Int(7))}, nil, nil, nil, false, false),
	}, nil, ast.Pkg([]interface{}{"bytecode_pkg"}, false))

	byteInterp := NewBytecode()
	if _, _, err := byteInterp.EvaluateModule(pkgModule); err != nil {
		t.Fatalf("bytecode package module failed: %v", err)
	}

	entryModule := ast.Mod([]ast.Statement{
		ast.Imp([]interface{}{"bytecode_pkg"}, false, []*ast.ImportSelector{ast.ImpSel("value", nil)}, nil),
		ast.Call("value"),
	}, nil, nil)

	program, err := byteInterp.lowerModuleToBytecode(entryModule)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	sawImport := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpImport {
			sawImport = true
			break
		}
	}
	if !sawImport {
		t.Fatalf("bytecode import opcode not emitted")
	}

	got, _, err := byteInterp.EvaluateModule(entryModule)
	if err != nil {
		t.Fatalf("bytecode entry module failed: %v", err)
	}

	treeInterp := New()
	if _, _, err := treeInterp.EvaluateModule(pkgModule); err != nil {
		t.Fatalf("tree-walker package module failed: %v", err)
	}
	want, _, err := treeInterp.EvaluateModule(entryModule)
	if err != nil {
		t.Fatalf("tree-walker entry module failed: %v", err)
	}

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode import mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_DefinitionOpcodes(t *testing.T) {
	alias := ast.NewTypeAliasDefinition(ast.ID("Alias"), ast.Ty("i32"), nil, nil, false)
	union := ast.UnionDef("Maybe", []ast.TypeExpression{ast.Ty("i32"), ast.Ty("String")}, nil, nil, false)
	module := ast.Mod([]ast.Statement{
		alias,
		union,
	}, nil, nil)

	byteInterp := NewBytecode()
	program, err := byteInterp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	var sawAlias, sawUnion bool
	for _, instr := range program.instructions {
		switch instr.op {
		case bytecodeOpDefineTypeAlias:
			sawAlias = true
		case bytecodeOpDefineUnion:
			sawUnion = true
		}
	}
	if !sawAlias {
		t.Fatalf("bytecode type alias opcode not emitted")
	}
	if !sawUnion {
		t.Fatalf("bytecode union opcode not emitted")
	}
	_ = runBytecodeModuleWithInterpreter(t, byteInterp, module)
}

func TestBytecodeVM_UnaryNegate(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Un(ast.UnaryOperatorNegate, ast.Int(3)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode unary negate mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_UnaryNot(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Un(ast.UnaryOperatorNot, ast.Bool(true)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode unary not mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_RangeExpression(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("r"), ast.Range(ast.Int(1), ast.Int(4), true)),
		ast.Index(ast.ID("r"), ast.Int(2)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode range mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_TypeCast(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.NewTypeCastExpression(ast.Flt(3.7), ast.Ty("i32")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode type cast mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_StringInterpolation(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Interp(ast.Str("hello "), ast.Int(42), ast.Str("!")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode string interpolation mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_BlockScope(t *testing.T) {
	block := ast.Block(
		ast.Assign(ast.ID("x"), ast.Int(2)),
		ast.ID("x"),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(1)),
		ast.Assign(ast.ID("shadow"), block),
		ast.ID("x"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode block result mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_IfElse(t *testing.T) {
	ifExpr := ast.NewIfExpression(
		ast.Bool(false),
		ast.Block(ast.Int(1)),
		[]*ast.ElseIfClause{ast.NewElseIfClause(ast.Block(ast.Int(2)), ast.Bool(false))},
		ast.Block(ast.Int(3)),
	)
	module := ast.Mod([]ast.Statement{ifExpr}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode if result mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ForLoopArraySum(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.ForLoopPattern(
			ast.ID("x"),
			ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("sum"),
					ast.Bin("+", ast.ID("sum"), ast.ID("x")),
				),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode for loop result mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_ForLoopBreakValue(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.ForLoopPattern(
			ast.ID("x"),
			ast.Arr(ast.Int(1), ast.Int(2), ast.Int(3)),
			ast.Block(
				ast.IfExpr(
					ast.Bin("==", ast.ID("x"), ast.Int(2)),
					ast.Block(ast.Brk(nil, ast.ID("x"))),
				),
			),
		),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode for loop break mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_WhileLoop(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.While(
			ast.Bin("<", ast.ID("i"), ast.Int(3)),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("i"),
					ast.Bin("+", ast.ID("i"), ast.Int(1)),
				),
			),
		),
		ast.ID("i"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode while result mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_WhileLoopBreakValue(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.While(
			ast.Bool(true),
			ast.Block(
				ast.Brk(nil, ast.Int(7)),
			),
		),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode while break mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_WhileLoopContinue(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.Assign(ast.ID("sum"), ast.Int(0)),
		ast.While(
			ast.Bin("<", ast.ID("i"), ast.Int(3)),
			ast.Block(
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("i"),
					ast.Bin("+", ast.ID("i"), ast.Int(1)),
				),
				ast.IfExpr(
					ast.Bin("==", ast.ID("i"), ast.Int(2)),
					ast.Block(ast.Cont(nil)),
				),
				ast.AssignOp(
					ast.AssignmentAssign,
					ast.ID("sum"),
					ast.Bin("+", ast.ID("sum"), ast.ID("i")),
				),
			),
		),
		ast.ID("sum"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode while continue mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_LoopExpressionBreakValue(t *testing.T) {
	loopExpr := ast.Loop(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("i"),
			ast.Bin("+", ast.ID("i"), ast.Int(1)),
		),
		ast.IfExpr(
			ast.Bin("==", ast.ID("i"), ast.Int(3)),
			ast.Block(ast.Brk(nil, ast.ID("i"))),
		),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.Assign(ast.ID("result"), loopExpr),
		ast.ID("result"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode loop break mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_LoopExpressionContinue(t *testing.T) {
	loopExpr := ast.Loop(
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("i"),
			ast.Bin("+", ast.ID("i"), ast.Int(1)),
		),
		ast.IfExpr(
			ast.Bin("==", ast.ID("i"), ast.Int(2)),
			ast.Block(ast.Cont(nil)),
		),
		ast.IfExpr(
			ast.Bin("==", ast.ID("i"), ast.Int(4)),
			ast.Block(ast.Brk(nil, ast.ID("i"))),
		),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("i"), ast.Int(0)),
		ast.Assign(ast.ID("result"), loopExpr),
		ast.ID("result"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode loop continue mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_MatchLiteralPatterns(t *testing.T) {
	matchExpr := ast.Match(
		ast.Int(2),
		ast.Mc(ast.LitP(ast.Int(1)), ast.Int(10)),
		ast.Mc(ast.LitP(ast.Int(2)), ast.Int(20)),
		ast.Mc(ast.Wc(), ast.Int(30)),
	)
	module := ast.Mod([]ast.Statement{matchExpr}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode match literal mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_MatchGuard(t *testing.T) {
	matchExpr := ast.Match(
		ast.Int(3),
		ast.Mc(ast.ID("x"), ast.Int(1), ast.Bin("<", ast.ID("x"), ast.Int(3))),
		ast.Mc(ast.ID("x"), ast.Int(2), ast.Bin("==", ast.ID("x"), ast.Int(3))),
		ast.Mc(ast.Wc(), ast.Int(0)),
	)
	module := ast.Mod([]ast.Statement{matchExpr}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode match guard mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_RescueExpression(t *testing.T) {
	rescueExpr := ast.Rescue(
		ast.Block(ast.Raise(ast.Str("boom"))),
		ast.Mc(ast.Wc(), ast.Int(7)),
	)
	module := ast.Mod([]ast.Statement{rescueExpr}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode rescue mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_RaiseStatement(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Raise(ast.Str("boom")),
	}, nil, nil)

	interp := New()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	if _, err := vm.run(program); err == nil {
		t.Fatalf("expected raise error")
	} else if _, ok := err.(raiseSignal); !ok {
		t.Fatalf("expected raise signal, got %T", err)
	}
}

func TestBytecodeVM_EnsureExpression(t *testing.T) {
	ensureExpr := ast.Ensure(
		ast.Block(
			ast.AssignOp(
				ast.AssignmentAssign,
				ast.ID("x"),
				ast.Int(1),
			),
			ast.Int(5),
		),
		ast.AssignOp(
			ast.AssignmentAssign,
			ast.ID("x"),
			ast.Bin("+", ast.ID("x"), ast.Int(2)),
		),
	)
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(0)),
		ast.Assign(ast.ID("result"), ensureExpr),
		ast.Bin("+", ast.ID("x"), ast.ID("result")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode ensure mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_RethrowStatement(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Rethrow(),
	}, nil, nil)

	interp := New()
	program, err := interp.lowerModuleToBytecode(module)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	vm := newBytecodeVM(interp, interp.GlobalEnvironment())
	if _, err := vm.run(program); err == nil {
		t.Fatalf("expected rethrow error")
	} else if _, ok := err.(raiseSignal); !ok {
		t.Fatalf("expected raise signal, got %T", err)
	}
}

func TestBytecodeVM_PropagationExpression(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Prop(ast.Int(3)),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)

	if !valuesEqual(got, want) {
		t.Fatalf("bytecode propagation mismatch: got=%#v want=%#v", got, want)
	}
}

func TestBytecodeVM_OrElseExpression(t *testing.T) {
	errorVal := runtime.ErrorValue{Message: "boom"}
	errorModule := ast.Mod([]ast.Statement{
		ast.OrElse(ast.ID("err"), "e", ast.ID("e")),
	}, nil, nil)

	treeInterp := New()
	treeInterp.GlobalEnvironment().Define("err", errorVal)
	want := mustEvalModule(t, treeInterp, errorModule)

	byteInterp := NewBytecode()
	byteInterp.GlobalEnvironment().Define("err", errorVal)
	program, err := byteInterp.lowerModuleToBytecode(errorModule)
	if err != nil {
		t.Fatalf("bytecode lowering failed: %v", err)
	}
	sawOrElse := false
	for _, instr := range program.instructions {
		if instr.op == bytecodeOpOrElse {
			sawOrElse = true
			break
		}
	}
	if !sawOrElse {
		t.Fatalf("bytecode or-else opcode not emitted")
	}
	got := runBytecodeModuleWithInterpreter(t, byteInterp, errorModule)
	gotMsg, gotOK := asErrorValue(got)
	wantMsg, wantOK := asErrorValue(want)
	if !gotOK || !wantOK {
		t.Fatalf("bytecode or-else expected error value: got=%#v want=%#v", got, want)
	}
	if gotMsg.Message != wantMsg.Message {
		t.Fatalf("bytecode or-else error binding mismatch: got=%#v want=%#v", got, want)
	}

	nilModule := ast.Mod([]ast.Statement{
		ast.OrElse(ast.Nil(), nil, ast.Int(9)),
	}, nil, nil)
	nilTree := mustEvalModule(t, New(), nilModule)
	nilByte := runBytecodeModule(t, nilModule)
	if !valuesEqual(nilByte, nilTree) {
		t.Fatalf("bytecode or-else nil mismatch: got=%#v want=%#v", nilByte, nilTree)
	}
}

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
