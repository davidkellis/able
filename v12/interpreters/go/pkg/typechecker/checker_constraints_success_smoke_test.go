package typechecker

import (
	"able/interpreter-go/pkg/ast"
	"testing"
)

func TestConstraintSolverAcceptsSatisfiedImpl(t *testing.T) {
	checker := New()
	showSig := ast.FnSig(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		ast.Ty("String"),
		nil,
		nil,
		nil,
	)
	displayIface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	wrapperStruct := ast.StructDef(
		"Wrapper",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("String"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	showMethod := ast.Fn(
		"show",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
		},
		[]ast.Statement{
			ast.Block(ast.Str("ok")),
		},
		ast.Ty("String"),
		nil,
		nil,
		false,
		false,
	)
	displayImpl := ast.Impl(
		"Display",
		ast.Ty("Wrapper"),
		[]*ast.FunctionDefinition{showMethod},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	genericParam := ast.GenericParam("T", ast.InterfaceConstr(ast.Ty("Display")))
	fn := ast.Fn(
		"useDisplay",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.Ret(ast.ID("value")),
		},
		ast.Ty("T"),
		[]*ast.GenericParameter{genericParam},
		nil,
		false,
		false,
	)
	call := ast.Call(
		"useDisplay",
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Str("hi"), "value"),
			},
			false,
			"Wrapper",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{displayIface, wrapperStruct, displayImpl, fn, call}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}
