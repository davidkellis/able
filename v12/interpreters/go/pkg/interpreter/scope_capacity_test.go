package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestBlockLocalBindingCapacityCountsImmediateDeclarations(t *testing.T) {
	block := ast.Block(
		ast.Assign(ast.ID("x"), ast.Int(1)),
		ast.Assign(ast.TypedP(ast.ID("y"), ast.Ty("i32")), ast.Int(2)),
		ast.Fn("helper", nil, []ast.Statement{ast.Int(0)}, ast.Ty("i32"), nil, nil, false, false),
		ast.StructDef("Point", nil, ast.StructKindNamed, nil, nil, false),
		ast.Bin("+", ast.ID("x"), ast.ID("y")),
	)

	if got := blockLocalBindingCapacity(block); got != 4 {
		t.Fatalf("blockLocalBindingCapacity(...) = %d, want 4", got)
	}
}

func TestAssignmentTargetBindingCapacityCountsNestedPatterns(t *testing.T) {
	target := ast.NewStructPattern(
		[]*ast.StructPatternField{
			ast.NewStructPatternField(ast.ID("x"), ast.ID("left"), nil, nil),
			ast.NewStructPatternField(ast.NewTypedPattern(ast.ID("y"), ast.Ty("i32")), ast.ID("right"), nil, nil),
			ast.NewStructPatternField(ast.NewWildcardPattern(), ast.ID("skip"), ast.ID("alias"), nil),
		},
		false,
		ast.ID("Pair"),
	)

	if got := assignmentTargetBindingCapacity(target); got != 3 {
		t.Fatalf("assignmentTargetBindingCapacity(...) = %d, want 3", got)
	}
}

func TestFunctionLocalBindingCapacityIncludesParamsLocalsAndGenerics(t *testing.T) {
	fn := ast.Fn(
		"map_one",
		[]*ast.FunctionParameter{
			ast.Param("x", ast.Ty("T")),
			ast.NewFunctionParameter(ast.NewTypedPattern(ast.ID("y"), ast.Ty("i32")), ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Assign(ast.ID("sum"), ast.Bin("+", ast.ID("x"), ast.ID("y"))),
			ast.Ret(ast.ID("sum")),
		},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	call := ast.CallExpr(ast.ID("map_one"), ast.Int(1), ast.Int(2))
	call.TypeArguments = []ast.TypeExpression{ast.Ty("i32")}

	if got := functionLocalBindingCapacity(fn, call); got != 5 {
		t.Fatalf("functionLocalBindingCapacity(...) = %d, want 5", got)
	}
}
