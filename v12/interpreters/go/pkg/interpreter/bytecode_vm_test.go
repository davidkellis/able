package interpreter

import (
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

func TestBytecodeVM_PropagationStillRaisesStructImplementingError(t *testing.T) {
	module := mustParseModuleSource(t, `
struct MyError { message: String }

impl Error for MyError {
  fn message(self: Self) -> String { self.message }
  fn cause(self: Self) -> ?Error { nil }
}

fn result_value() -> !String {
  MyError { message: "bad" }
}

fn fail() -> !String {
  result_value()!
  "after"
}

fail() or { err => err.message() }
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode propagation mismatch: got=%#v want=%#v", got, want)
	}
	if str, ok := got.(runtime.StringValue); !ok || str.Val != "bad" {
		t.Fatalf("expected handled error message, got %#v", got)
	}
}

func TestBytecodeVM_PropagationReturnsNilFromCurrentFunction(t *testing.T) {
	module := mustParseModuleSource(t, `
fn maybe_text(ok: bool) {
  if ok { "value" } else { nil }
}

fn marker(ok: bool) {
  maybe_text(ok)!
  "after"
}

first := marker(true) or { "nil" }
second := marker(false) or { "nil" }
first + ":" + second
`)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode propagation mismatch: got=%#v want=%#v", got, want)
	}
	if str, ok := got.(runtime.StringValue); !ok || str.Val != "after:nil" {
		t.Fatalf("expected nil propagation to return from marker, got %#v", got)
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
