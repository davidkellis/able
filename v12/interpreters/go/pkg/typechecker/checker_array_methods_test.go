package typechecker

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestArrayMethodSetOverridesNativeSignatures(t *testing.T) {
	checker := New()

	call := ast.CallExpr(ast.Member(ast.ID("arr"), "size"))

	methods := ast.Methods(
		ast.Ty("Array"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"size",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("T"))),
				},
				[]ast.Statement{
					ast.Ret(ast.Str("ok")),
				},
				ast.Ty("String"),
				nil,
				nil,
				false,
				false,
			),
		},
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
	)

	probe := ast.Fn(
		"probe",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("String"))),
		},
		[]ast.Statement{
			ast.Ret(call),
		},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.NewModule([]ast.Statement{methods, probe}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if typ, ok := checker.infer[call]; !ok {
		t.Fatalf("expected inference entry for array size call")
	} else if typeName(typ) != "String" {
		t.Fatalf("expected array size to use method set return type, got %q", typeName(typ))
	}
}
