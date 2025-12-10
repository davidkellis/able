package interpreter

import (
	"strings"
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

func TestPlaceholderBareAtUsesFirstArgument(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("square"),
			ast.Bin("*", ast.Placeholder(), ast.Placeholder()),
		),
		ast.Call("square", ast.Int(6)),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("placeholder square evaluation failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(36)) != 0 {
		t.Fatalf("expected 36, got %#v", intResult.Val)
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
	if intResult.Val.Cmp(bigInt(779)) != 0 {
		t.Fatalf("expected 779, got %#v", intResult.Val)
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

func TestLambdaContainingPlaceholderRemainsExplicitFunction(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(
			ast.ID("builder"),
			ast.LamBlock(nil, ast.Block(ast.Placeholder())),
		),
		ast.Call("builder"),
	}, nil, nil)

	result, env, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("lambda placeholder module failed: %v", err)
	}
	builderVal, err := env.Get("builder")
	if err != nil {
		t.Fatalf("expected builder in environment: %v", err)
	}
	if _, ok := builderVal.(*runtime.FunctionValue); !ok {
		t.Fatalf("expected builder to be FunctionValue, got %#v", builderVal)
	}
	native, ok := result.(runtime.NativeFunctionValue)
	if !ok {
		t.Fatalf("expected builder() to return placeholder function, got %#v", result)
	}
	if native.Arity != 1 {
		t.Fatalf("expected placeholder arity 1, got %d", native.Arity)
	}
}

func TestPipeRhsMustBeCallableWithoutTopic(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Bin("|>", ast.Int(5), ast.Int(3)),
	}, nil, nil)

	_, _, err := interp.EvaluateModule(module)
	if err == nil {
		t.Fatalf("expected pipe to error when RHS is not callable")
	}
	if !strings.Contains(err.Error(), "pipe RHS must be callable") {
		t.Fatalf("expected error about pipe RHS, got %v", err)
	}
}

func TestPipeImplicitMethodShorthand(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Data",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "value"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Data"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"increment",
					nil,
					[]ast.Statement{
						ast.Ret(ast.Bin("+", ast.ImplicitMember("value"), ast.Int(1))),
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
			ast.ID("item"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(41), "value"),
				},
				false,
				"Data",
				nil,
				nil,
			),
		),
		ast.Bin("|>", ast.ID("item"), ast.ImplicitMember("increment")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("pipe implicit method failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(42)) != 0 {
		t.Fatalf("expected 42, got %#v", intResult.Val)
	}
}

func TestPipeUfcsFunction(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Fn(
			"translate",
			[]*ast.FunctionParameter{
				ast.Param("point", nil),
			},
			[]ast.Statement{
				ast.AssignMember(
					ast.ID("point"),
					"x",
					ast.Bin("+", ast.Member(ast.ID("point"), "x"), ast.Int(5)),
				),
				ast.Ret(ast.ID("point")),
			},
			nil,
			nil,
			nil,
			false,
			false,
		),
		ast.Assign(
			ast.ID("point"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(2), "x"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.Assign(
			ast.ID("result"),
			ast.Bin("|>", ast.ID("point"), ast.ID("translate")),
		),
		ast.Member(ast.ID("result"), "x"),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("pipe UFCS function failed: %v", err)
	}
	intResult, ok := result.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %#v", result)
	}
	if intResult.Val.Cmp(bigInt(7)) != 0 {
		t.Fatalf("expected 7, got %#v", intResult.Val)
	}
}
