package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_SlotLoweringEmitsTypedPrimitiveMatchOpcode(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.Assign(ast.ID("x"), ast.NewTypeCastExpression(ast.Int(7), ast.Ty("u8"))),
			ast.Match(
				ast.ID("x"),
				ast.Mc(
					ast.TypedP(ast.ID("byte"), ast.Ty("u8")),
					ast.Bin("+", ast.NewTypeCastExpression(ast.ID("byte"), ast.Ty("i32")), ast.Int(1)),
				),
				ast.Mc(ast.LitP(ast.Nil()), ast.Int(0)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("slot-lowered primitive match mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered typed-pattern match opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotNil) {
		t.Fatalf("expected slot-lowered nil-pattern match opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered primitive match should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringEmitsTypedNominalMatchOpcode(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Node", nil, ast.StructKindNamed, nil, nil, false),
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Nullable(ast.Ty("Node"))),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.LitP(ast.Nil()), ast.Int(0)),
				ast.Mc(ast.TypedP(ast.ID("node"), ast.Ty("Node")), ast.Int(1)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.StructLit(nil, false, "Node", nil, nil)),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if intResult, ok := got.(runtime.IntegerValue); !ok {
		t.Fatalf("expected integer result, got %T (%#v)", got, got)
	} else if val, ok := intResult.ToInt64(); !ok || val != 1 {
		t.Fatalf("unexpected nominal match result: got=%#v want=1", got)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered nominal typed-pattern match opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered nominal match should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringEmitsZeroFieldNamedStructMatchWithoutGenericMatch(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Node", []*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		}, ast.StructKindNamed, nil, nil, false),
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Ty("Node")),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.StructP(nil, false, "Node"), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(7), "value"),
		}, false, "Node", nil, nil)),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("slot-lowered zero-field named struct match mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered zero-field named struct match opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered zero-field named struct match should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringEmitsEmptyNamedStructMatchWithoutGenericMatch(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Empty", nil, ast.StructKindNamed, nil, nil, false),
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Ty("Empty")),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.StructP(nil, false, "Empty"), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.StructLit(nil, false, "Empty", nil, nil)),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("slot-lowered empty named struct match mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered empty named struct match opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered empty named struct match should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringEmitsIteratorEndStructPatternWithoutGenericMatch(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", nil, []ast.Statement{
			ast.Match(
				ast.ID("it_end"),
				ast.Mc(ast.StructP(nil, false, "IteratorEnd"), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	interp := NewBytecode()
	interp.GlobalEnvironment().Define("it_end", runtime.IteratorEnd)
	got := runBytecodeModuleWithInterpreter(t, interp, module)

	tree := New()
	tree.GlobalEnvironment().Define("it_end", runtime.IteratorEnd)
	want := mustEvalModule(t, tree, module)
	if !valuesEqual(got, want) {
		t.Fatalf("slot-lowered IteratorEnd struct pattern mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered IteratorEnd struct-pattern opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered IteratorEnd struct pattern should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringEmitsIdentifierBindingMatchWithoutGenericMatch(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Ty("i32")),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.ID("matched"), ast.Bin("+", ast.ID("matched"), ast.Int(1))),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.Int(41)),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("slot-lowered identifier match mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered identifier match should not emit generic Match opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpEnterScope) {
		t.Fatalf("slot-lowered identifier match should not emit runtime EnterScope")
	}
}

func TestBytecodeVM_SlotLoweringEmitsNamedStructFieldPatternWithoutGenericMatch(t *testing.T) {
	body := ast.NewIfExpression(
		ast.ID("flag"),
		ast.Block(ast.Bin("+", ast.ID("v"), ast.Int(1))),
		nil,
		ast.Block(ast.Bin("-", ast.ID("v"), ast.Int(1))),
	)
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Node", []*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
			ast.FieldDef(ast.Ty("bool"), "flag"),
		}, ast.StructKindNamed, nil, nil, false),
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Nullable(ast.Ty("Node"))),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.LitP(ast.Nil()), ast.Int(0)),
				ast.Mc(ast.StructP([]*ast.StructPatternField{
					ast.FieldP(ast.ID("v"), "value", nil),
					ast.FieldP(ast.ID("flag"), "flag", nil),
				}, false, "Node"), body),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(7), "value"),
			ast.FieldInit(ast.Bool(true), "flag"),
		}, false, "Node", nil, nil)),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("slot-lowered named struct field match mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if program.frameLayout == nil {
		t.Fatalf("expected function f to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered named struct field match opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpLoadSlotStructField) {
		t.Fatalf("expected named struct field match to use slot struct-field load bytecode")
	}
	if len(program.namedStructMembers) == 0 {
		t.Fatalf("expected named struct field match to record member fast-path plans")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered named struct field match should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringParsedGenericNamedStructFieldPattern(t *testing.T) {
	module := mustParseModuleSource(t, `
struct TreeEmpty {}
struct TreeNode T {
  value: T,
  left: Tree T,
  height: i32,
}
union Tree T = TreeEmpty | TreeNode T

fn tree_height<T>(tree: Tree T) -> i32 {
  tree match {
    case TreeEmpty {} => 0,
    case TreeNode { height } => height,
  }
}

tree_height(TreeNode { value: 1, left: TreeEmpty {}, height: 3 })
`)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("parsed generic named struct field match mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "tree_height")
	if program.frameLayout == nil {
		t.Fatalf("expected parsed generic tree_height to use slot frame layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected parsed generic named struct field match opcode")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpLoadSlotStructField) {
		t.Fatalf("expected parsed generic named struct field match to use slot struct-field load bytecode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered parsed generic named struct field match should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringTypedGenericPatternCarriesPlan(t *testing.T) {
	module := mustParseModuleSource(t, `
struct Box T {
  value: T
}

fn accepts<T>(box: Box T) -> i32 {
  box match {
    case matched: Box T => 1,
    case _ => 0,
  }
}

accepts(Box { value: 9 })
`)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("typed generic match mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "accepts")
	if program.frameLayout == nil {
		t.Fatalf("expected accepts to use slot frame layout")
	}
	foundPlan := false
	for idx := range program.instructions {
		instr := &program.instructions[idx]
		if instr.op == bytecodeOpJumpIfNotTypedPattern && instr.genericStructMatch != nil {
			foundPlan = true
			break
		}
	}
	if !foundPlan {
		t.Fatalf("expected typed generic match opcode to carry generic struct plan")
	}
}

func TestBytecodeVM_SlotLoweringSingletonIdentifierPatternFallsBackToGenericMatch(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Done", nil, ast.StructKindSingleton, nil, nil, false),
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Nullable(ast.Ty("Done"))),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.ID("Done"), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.Nil()),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("singleton identifier pattern mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if !bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("singleton identifier pattern should stay on generic Match lowering")
	}
}

func TestBytecodeVM_SlotLoweringTypedSingletonIdentifierPatternFallsBackToGenericMatch(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Done", nil, ast.StructKindSingleton, nil, nil, false),
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Nullable(ast.Ty("Done"))),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.TypedP(ast.ID("Done"), ast.Nullable(ast.Ty("Done"))), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.Nil()),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("typed singleton identifier pattern mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if !bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("typed singleton identifier pattern should stay on generic Match lowering")
	}
}

func TestBytecodeVM_SlotLoweringZeroFieldSingletonStructPatternWithoutGenericMatch(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Done", nil, ast.StructKindSingleton, nil, nil, false),
		ast.Fn("f", nil, []ast.Statement{
			ast.Match(
				ast.ID("Done"),
				ast.Mc(ast.StructP(nil, false, "Done"), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f"),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("zero-field singleton struct pattern mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered zero-field singleton struct pattern opcode")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("slot-lowered zero-field singleton struct pattern should not emit generic Match opcode")
	}
}

func TestBytecodeVM_SlotLoweringZeroFieldNonStructPatternFallsBackToGenericMatch(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Ty("String")),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.StructP(nil, false, "String"), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.Str("hello")),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("zero-field non-struct pattern mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if !bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("zero-field non-struct pattern should stay on generic Match lowering")
	}
}

func TestBytecodeVM_SlotLoweringTypedNominalMatchWithFields(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Node", []*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		}, ast.StructKindNamed, nil, nil, false),
		ast.Fn("f", []*ast.FunctionParameter{
			ast.Param("value", ast.Nullable(ast.Ty("Node"))),
		}, []ast.Statement{
			ast.Match(
				ast.ID("value"),
				ast.Mc(ast.LitP(ast.Nil()), ast.Int(0)),
				ast.Mc(ast.TypedP(ast.ID("node"), ast.Ty("Node")), ast.Member(ast.ID("node"), "value")),
			),
		}, ast.Ty("i32"), nil, nil, false, false),
		ast.Call("f", ast.StructLit([]*ast.StructFieldInitializer{
			ast.FieldInit(ast.Int(7), "value"),
		}, false, "Node", nil, nil)),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("slot-lowered nominal match with fields mismatch: got=%#v want=%#v", got, want)
	}

	program := mustBytecodeFunctionProgram(t, interp, "f")
	if !bytecodeProgramContainsOpcode(program, bytecodeOpJumpIfNotTypedPattern) {
		t.Fatalf("expected slot-lowered nominal typed-pattern match opcode")
	}
}

func mustBytecodeFunctionProgram(t *testing.T, interp *Interpreter, name string) *bytecodeProgram {
	t.Helper()
	raw, err := interp.GlobalEnvironment().Get(name)
	if err != nil {
		t.Fatalf("lookup function %s: %v", name, err)
	}
	fn, ok := raw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("expected function %s, got %T", name, raw)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("expected bytecode program for %s", name)
	}
	return program
}
