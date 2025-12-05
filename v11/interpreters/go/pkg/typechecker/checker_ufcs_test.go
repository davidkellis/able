package typechecker

import (
	"able/interpreter10-go/pkg/ast"
	"testing"
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
		ast.Ty("string"),
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
