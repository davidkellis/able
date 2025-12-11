package typechecker

import (
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestMethodFunctionsCallableByName(t *testing.T) {
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
	methods := ast.Methods(
		ast.Ty("Point"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"norm",
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
	)
	call := ast.CallExpr(
		ast.ID("norm"),
		ast.StructLit(
			[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(1), "x")},
			false,
			"Point",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{point, methods, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestMethodFunctionRequiresReceiverArgument(t *testing.T) {
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
	methods := ast.Methods(
		ast.Ty("Point"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"norm",
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
	)
	call := ast.CallExpr(ast.ID("norm"))
	module := ast.NewModule([]ast.Statement{point, methods, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected partial application without diagnostics, got %v", diags)
	}
}

func TestMethodFunctionRequiresReceiverType(t *testing.T) {
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
	methods := ast.Methods(
		ast.Ty("Point"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"norm",
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
	)
	call := ast.CallExpr(ast.ID("norm"), ast.Int(3))
	module := ast.NewModule([]ast.Statement{point, methods, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for missing receiver argument type")
	}
	if msg := diags[0].Message; !strings.Contains(msg, "expected Point") {
		t.Fatalf("expected receiver type diagnostic, got %v", diags)
	}
}

func TestMethodFunctionEnforcesMethodSetConstraints(t *testing.T) {
	checker := New()
	display := ast.Iface(
		"Display",
		[]*ast.FunctionSignature{
			ast.FnSig("show", []*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))}, ast.Ty("String"), nil, nil, nil),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	doc := ast.StructDef(
		"Doc",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("String"), "body"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	methods := ast.Methods(
		ast.Ty("Doc"),
		[]*ast.FunctionDefinition{
			ast.Fn(
				"title",
				nil,
				[]ast.Statement{ast.Ret(ast.Str("hi"))},
				ast.Ty("String"),
				nil,
				nil,
				true,
				false,
			),
		},
		nil,
		[]*ast.WhereClauseConstraint{
			ast.WhereConstraint("Self", ast.InterfaceConstr(ast.Ty("Display"))),
		},
	)
	call := ast.CallExpr(
		ast.ID("title"),
		ast.StructLit(
			[]*ast.StructFieldInitializer{ast.FieldInit(ast.Str("body"), "body")},
			false,
			"Doc",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{display, doc, methods, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected constraint diagnostic for method function")
	}
	foundConstraint := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "Display") {
			foundConstraint = true
			break
		}
	}
	if !foundConstraint {
		t.Fatalf("expected Display constraint diagnostic, got %v", diags)
	}
}
