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
