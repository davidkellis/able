package compiler

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCompilerFailsClosedForInvariantArrayMismatch(t *testing.T) {
	takes := ast.Fn(
		"takes",
		[]*ast.FunctionParameter{ast.Param("values", ast.Gen(ast.Ty("Array"), ast.Ty("i32")))},
		[]ast.Statement{ast.Nil()},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	values := ast.Assign(
		ast.TypedP(ast.ID("values"), ast.Gen(ast.Ty("Array"), ast.Ty("i8"))),
		ast.Arr(ast.Int(7)),
	)
	module := ast.NewModule(
		[]ast.Statement{takes, values, ast.Call("takes", ast.ID("values"))},
		nil,
		ast.NewPackageStatement([]*ast.Identifier{ast.ID("demo")}, false),
	)

	_, err := New(Options{PackageName: "compiled", RequireNoFallbacks: true}).
		Compile(testProgramFromModule("demo", module))
	if err == nil {
		t.Fatal("expected compiler to reject invariant Array mismatch")
	}
	if !strings.Contains(err.Error(), "invariant type argument mismatch rejected") {
		t.Fatalf("unexpected compiler error: %v", err)
	}
}

func TestCompilerFailsClosedForCallableSignatureMismatch(t *testing.T) {
	takes := ast.Fn(
		"takes",
		[]*ast.FunctionParameter{
			ast.Param("callable", ast.FnType([]ast.TypeExpression{ast.Ty("i32")}, ast.Ty("i32"))),
		},
		[]ast.Statement{ast.Nil()},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	small := ast.Fn(
		"small",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("i8"))},
		[]ast.Statement{ast.ID("value")},
		ast.Ty("i8"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.NewModule(
		[]ast.Statement{takes, small, ast.Call("takes", ast.ID("small"))},
		nil,
		ast.NewPackageStatement([]*ast.Identifier{ast.ID("demo")}, false),
	)

	_, err := New(Options{PackageName: "compiled", RequireNoFallbacks: true}).
		Compile(testProgramFromModule("demo", module))
	if err == nil {
		t.Fatal("expected compiler to reject callable signature mismatch")
	}
	if !strings.Contains(err.Error(), "callable signature mismatch rejected") {
		t.Fatalf("unexpected compiler error: %v", err)
	}
}
