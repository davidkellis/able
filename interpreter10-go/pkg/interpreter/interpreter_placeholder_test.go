package interpreter

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func TestPlaceholderSimplePartialApplication(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"add",
			[]*ast.FunctionParameter{
				ast.Param("left", nil),
				ast.Param("right", nil),
			},
			[]ast.Statement{
				ast.Ret(ast.Bin("+", ast.ID("left"), ast.ID("right"))),
			},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.CallExpr(
			ast.CallExpr(ast.ID("add"), ast.Placeholder(), ast.Int(10)),
			ast.Int(5),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("placeholder partial application failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(15)) != 0 {
		t.Fatalf("expected 15, got %#v", intResult.Val)
	}
}

func TestPlaceholderMixedIndices(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"combine",
			[]*ast.FunctionParameter{
				ast.Param("a", nil),
				ast.Param("b", nil),
				ast.Param("c", nil),
			},
			[]ast.Statement{
				ast.Ret(
					ast.Bin(
						"+",
						ast.Bin("*", ast.ID("a"), ast.Int(100)),
						ast.Bin(
							"+",
							ast.Bin("*", ast.ID("b"), ast.Int(10)),
							ast.ID("c"),
						),
					),
				),
			},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.CallExpr(
			ast.CallExpr(ast.ID("combine"), ast.PlaceholderN(1), ast.Placeholder(), ast.PlaceholderN(3)),
			ast.Int(7),
			ast.Int(8),
			ast.Int(9),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("placeholder mixed indices evaluation failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(789)) != 0 {
		t.Fatalf("expected 789, got %#v", intResult.Val)
	}
}

func TestPipeTopicReference(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Bin("|>", ast.Int(5), ast.Bin("+", ast.TopicRef(), ast.Int(3))),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("pipe topic evaluation failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(8)) != 0 {
		t.Fatalf("expected 8, got %#v", intResult.Val)
	}
}

func TestPipeCallableRhs(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Fn(
			"increment",
			[]*ast.FunctionParameter{
				ast.Param("value", nil),
			},
			[]ast.Statement{
				ast.Ret(ast.Bin("+", ast.ID("value"), ast.Int(1))),
			},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Bin("|>", ast.Int(41), ast.ID("increment")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("pipe callable evaluation failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(42)) != 0 {
		t.Fatalf("expected 42, got %#v", intResult.Val)
	}
}
