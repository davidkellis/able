package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestEvaluateStringLiteral(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{ast.Str("hello")}, nil, nil)
	val, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "hello" {
		t.Fatalf("unexpected value %#v", val)
	}
}

func TestEvaluateIdentifierLookup(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()
	global.Define("greeting", runtime.StringValue{Val: "hello"})

	val, err := interp.evaluateExpression(ast.ID("greeting"), global)
	if err != nil {
		t.Fatalf("identifier lookup failed: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "hello" {
		t.Fatalf("unexpected value %#v", val)
	}
}

func TestEvaluateBlockCreatesScope(t *testing.T) {
	interp := New()
	global := interp.GlobalEnvironment()
	block := ast.Block(
		ast.Assign(ast.ID("x"), ast.Str("inner")),
		ast.ID("x"),
	)

	val, err := interp.evaluateExpression(block, global)
	if err != nil {
		t.Fatalf("block evaluation failed: %v", err)
	}
	str, ok := val.(runtime.StringValue)
	if !ok || str.Val != "inner" {
		t.Fatalf("unexpected block result %#v", val)
	}

	if _, err := global.Get("x"); err == nil {
		t.Fatalf("expected inner binding to stay scoped")
	}
}

func TestEvaluateBinaryAddition(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("a"), ast.Int(1)),
		ast.Assign(ast.ID("b"), ast.Int(2)),
		ast.Bin("+", ast.ID("a"), ast.ID("b")),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("module evaluation failed: %v", err)
	}
	iv, ok := result.(runtime.IntegerValue)
	if !ok || iv.Val.Cmp(bigInt(3)) != 0 {
		t.Fatalf("expected integer 3, got %#v", result)
	}
}

func TestStringInterpolationEvaluatesParts(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("x"), ast.Int(2)),
		ast.Interp(
			ast.Str("x = "),
			ast.ID("x"),
			ast.Str(", sum = "),
			ast.Bin("+", ast.Int(3), ast.Int(4)),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok {
		t.Fatalf("expected string result, got %#v", result)
	}
	if str.Val != "x = 2, sum = 7" {
		t.Fatalf("unexpected interpolation output: %q", str.Val)
	}
}

func TestStringInterpolationUsesToStringMethod(t *testing.T) {
	interp := New()
	module := ast.Mod([]ast.Statement{
		ast.StructDef(
			"Point",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.Ty("i32"), "x"),
				ast.FieldDef(ast.Ty("i32"), "y"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"to_string",
					[]*ast.FunctionParameter{ast.Param("self", nil)},
					[]ast.Statement{
						ast.Ret(ast.Interp(
							ast.Str("Point("),
							ast.Member(ast.ID("self"), "x"),
							ast.Str(","),
							ast.Member(ast.ID("self"), "y"),
							ast.Str(")"),
						)),
					},
					nil,
					nil,
					nil,
					false,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Assign(
			ast.ID("p"),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
					ast.FieldInit(ast.Int(2), "y"),
				},
				false,
				"Point",
				nil,
				nil,
			),
		),
		ast.Interp(
			ast.Str("P= "),
			ast.ID("p"),
		),
	}, nil, nil)

	result, _, err := interp.EvaluateModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	str, ok := result.(runtime.StringValue)
	if !ok || str.Val != "P= Point(1,2)" {
		t.Fatalf("unexpected interpolation output: %q", str.Val)
	}
}
