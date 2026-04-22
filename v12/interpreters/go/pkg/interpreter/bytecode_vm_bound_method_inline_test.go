package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestBytecodeVM_InlineBoundMethodCallStats(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	structDef := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	inc := ast.Fn(
		"inc",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("delta", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Bin("+", ast.Member(ast.ID("self"), "n"), ast.ID("delta")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	methods := ast.Methods(
		ast.Ty("S"),
		[]*ast.FunctionDefinition{inc},
		nil,
		nil,
	)

	module := ast.Mod([]ast.Statement{
		structDef,
		methods,
		ast.Assign(
			ast.ID("s"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(5), "n"),
			}, false, "S", nil, nil),
		),
		ast.Assign(ast.ID("a"), ast.CallExpr(ast.Member(ast.ID("s"), "inc"), ast.Int(3))),
		ast.Assign(ast.ID("b"), ast.CallExpr(ast.Member(ast.ID("s"), "inc"), ast.Int(4))),
		ast.Bin("+", ast.ID("a"), ast.ID("b")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode bound-method inline mismatch: got=%#v want=%#v", got, want)
	}

	stats := interp.BytecodeStats()
	if stats.InlineCallHits == 0 {
		t.Fatalf("expected inline call hits > 0 for bound method call sites")
	}
}

func TestBytecodeVM_InlineBoundGenericMethodCallStats(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("T"), "value"),
		},
		ast.StructKindNamed,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)

	setFn := ast.Fn(
		"set",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.Member(ast.ID("self"), "value"), ast.ID("value")),
			ast.Ret(ast.Member(ast.ID("self"), "value")),
		},
		ast.Ty("T"),
		nil,
		nil,
		false,
		false,
	)

	methods := ast.Methods(
		ast.Gen(ast.Ty("Box"), ast.Ty("T")),
		[]*ast.FunctionDefinition{setFn},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
	)

	module := ast.Mod([]ast.Statement{
		boxDef,
		methods,
		ast.Assign(
			ast.ID("box"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(0), "value"),
			}, false, "Box", nil, []ast.TypeExpression{ast.Ty("i32")}),
		),
		ast.Assign(ast.ID("a"), ast.CallExpr(ast.Member(ast.ID("box"), "set"), ast.Int(3))),
		ast.Assign(ast.ID("b"), ast.CallExpr(ast.Member(ast.ID("box"), "set"), ast.Int(4))),
		ast.Bin("+", ast.ID("a"), ast.ID("b")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode bound-generic-method inline mismatch: got=%#v want=%#v", got, want)
	}

	stats := interp.BytecodeStats()
	if stats.InlineCallHits == 0 {
		t.Fatalf("expected inline call hits > 0 for bound generic method call sites")
	}
}

func TestBytecodeVM_InlineOverloadedMemberMethodCallStats(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	structDef := ast.StructDef(
		"Printer",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	renderInt := ast.Fn(
		"render",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Bin("+", ast.Member(ast.ID("self"), "n"), ast.ID("value")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	renderString := ast.Fn(
		"render",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
			ast.Param("value", ast.Ty("String")),
		},
		[]ast.Statement{
			ast.Member(ast.ID("self"), "n"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	methods := ast.Methods(
		ast.Ty("Printer"),
		[]*ast.FunctionDefinition{renderInt, renderString},
		nil,
		nil,
	)

	module := ast.Mod([]ast.Statement{
		structDef,
		methods,
		ast.Assign(
			ast.ID("p"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(5), "n"),
			}, false, "Printer", nil, nil),
		),
		ast.Assign(ast.ID("a"), ast.CallExpr(ast.Member(ast.ID("p"), "render"), ast.Int(3))),
		ast.Assign(ast.ID("b"), ast.CallExpr(ast.Member(ast.ID("p"), "render"), ast.Int(4))),
		ast.Bin("+", ast.ID("a"), ast.ID("b")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode overloaded-member inline mismatch: got=%#v want=%#v", got, want)
	}

	stats := interp.BytecodeStats()
	if stats.InlineCallHits == 0 {
		t.Fatalf("expected inline call hits > 0 for overloaded member call sites")
	}
}
