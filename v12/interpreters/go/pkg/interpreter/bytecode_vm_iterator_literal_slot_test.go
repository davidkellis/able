package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestBytecodeVM_IteratorLiteralGenBindingUsesSlotFrame(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")
	iterLit := ast.IteratorLit(
		ast.CallExpr(ast.Member(ast.ID("gen"), "yield"), ast.Int(1)),
		ast.CallExpr(ast.Member(ast.ID("gen"), "yield"), ast.Int(2)),
	)
	iterLit.Binding = ast.ID("gen")
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("iter"), iterLit),
		ast.Assign(ast.ID("first"), ast.CallExpr(ast.Member(ast.ID("iter"), "next"))),
		ast.Assign(ast.ID("second"), ast.CallExpr(ast.Member(ast.ID("iter"), "next"))),
		ast.Bin("+", ast.ID("first"), ast.ID("second")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode iterator literal gen binding mismatch: got=%#v want=%#v", got, want)
	}
	if got := interp.BytecodeStats().LoadNameLookupsByName["gen"]; got != 0 {
		t.Fatalf("iterator generator binding should use slots, got %d gen LoadName lookups", got)
	}
}
