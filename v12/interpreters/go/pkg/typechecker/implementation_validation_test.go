package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestImplementationLabelCanonicalizesAliasTargets(t *testing.T) {
	checker := New()
	alias := ast.NewTypeAliasDefinition(
		ast.ID("Fancy"),
		ast.Gen(ast.Ty("Array"), ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
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
	iface := ast.Iface("Display", []*ast.FunctionSignature{showSig}, nil, nil, nil, nil, false)
	impl := ast.Impl(
		"Display",
		ast.Gen(ast.Ty("Fancy"), ast.Ty("T")),
		nil,
		nil,
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		nil,
		false,
	)
	module := ast.NewModule([]ast.Statement{alias, iface, impl}, nil, nil)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "impl Display for Array _") &&
			strings.Contains(diag.Message, "missing method 'show'") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected canonicalized impl label diagnostic, got %v", diags)
	}
}
