package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestAnalyzeFrameLayoutAllowsNonCapturingLambda(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(
				ast.ID("mapper"),
				ast.Lam(
					[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i32"))},
					ast.Bin("+", ast.ID("value"), ast.Int(1)),
				),
			),
			ast.CallExpr(ast.ID("mapper"), ast.Int(41)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	interp := NewBytecode()
	if layout := analyzeFrameLayout(interp, def); layout == nil {
		t.Fatalf("expected non-capturing lambda to preserve slot layout")
	}
	module := ast.Mod([]ast.Statement{def, ast.Call("f")}, nil, nil)
	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("non-capturing lambda bytecode mismatch: got=%#v want=%#v", got, want)
	}
}

func TestAnalyzeFrameLayoutRejectsCapturingLambda(t *testing.T) {
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("base", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Assign(
				ast.ID("mapper"),
				ast.Lam(
					[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i32"))},
					ast.Bin("+", ast.ID("value"), ast.ID("base")),
				),
			),
			ast.Int(0),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayout(NewBytecode(), def); layout != nil {
		t.Fatalf("expected capturing lambda to keep function off slot layout")
	}
}

func TestAnalyzeFrameLayoutLambdaParamShadowsOuterLocal(t *testing.T) {
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Assign(
				ast.ID("mapper"),
				ast.Lam(
					[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i32"))},
					ast.Bin("+", ast.ID("value"), ast.Int(1)),
				),
			),
			ast.CallExpr(ast.ID("mapper"), ast.ID("value")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayout(NewBytecode(), def); layout == nil {
		t.Fatalf("expected lambda parameter to shadow outer slot local")
	}
	module := ast.Mod([]ast.Statement{def, ast.Call("f", ast.Int(41))}, nil, nil)
	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("shadowing lambda bytecode mismatch: got=%#v want=%#v", got, want)
	}
}

func TestAnalyzeFrameLayoutRejectsLambdaFreeNameAfterNestedBlock(t *testing.T) {
	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Assign(
				ast.ID("mapper"),
				ast.Lam(
					[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i32"))},
					ast.Block(
						ast.Block(ast.Assign(ast.ID("temp"), ast.ID("value"))),
						ast.ID("temp"),
					),
				),
			),
			ast.Int(0),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayout(NewBytecode(), def); layout != nil {
		t.Fatalf("expected nested-block local not to leak into lambda free-name analysis")
	}
}
