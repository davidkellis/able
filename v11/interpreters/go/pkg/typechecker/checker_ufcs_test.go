package typechecker

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestUfcsMemberCallBindsFreeFunction(t *testing.T) {
	checker := New()
	point := ast.StructDef(
		"Point",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "x"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	tag := ast.Fn(
		"tag",
		[]*ast.FunctionParameter{
			ast.Param("p", ast.Ty("Point")),
		},
		[]ast.Statement{
			ast.Ret(ast.Str("point")),
		},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	call := ast.CallExpr(
		ast.Member(
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "x"),
			}, false, "Point", nil, nil),
			"tag",
		),
	)
	module := ast.NewModule([]ast.Statement{point, tag, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestCallableFieldPreferredOverMethod(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{
		ast.StructDef(
			"Box",
			[]*ast.StructFieldDefinition{
				ast.FieldDef(ast.FnType([]ast.TypeExpression{ast.Ty("String")}, ast.Ty("String")), "action"),
			},
			ast.StructKindNamed,
			nil,
			nil,
			false,
		),
		ast.Methods(
			ast.Ty("Box"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"action",
					nil,
					[]ast.Statement{ast.Ret(ast.Int(1))},
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
		ast.CallExpr(
			ast.Member(
				ast.StructLit([]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Lam([]*ast.FunctionParameter{ast.Param("msg", ast.Ty("String"))}, ast.Str("ok")), "action"),
				}, false, "Box", nil, nil),
				"action",
			),
			ast.Str("hi"),
		),
	}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestAmbiguousCallablePoolReported(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{
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
		ast.Methods(
			ast.Ty("Point"),
			[]*ast.FunctionDefinition{
				ast.Fn(
					"describe",
					nil,
					[]ast.Statement{ast.Ret(ast.Str("method"))},
					ast.Ty("String"),
					nil,
					nil,
					true,
					false,
				),
			},
			nil,
			nil,
		),
		ast.Fn(
			"describe",
			[]*ast.FunctionParameter{
				ast.Param("p", ast.Ty("Point")),
			},
			[]ast.Statement{ast.Ret(ast.Str("free"))},
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		ast.CallExpr(
			ast.Member(
				ast.StructLit([]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Int(1), "x"),
				}, false, "Point", nil, nil),
				"describe",
			),
		),
	}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected inherent method to win without ambiguity, got %v", diags)
	}
}
