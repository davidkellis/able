package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestCompositeInterfaceSelfPatternAllowsCompatibleBases(t *testing.T) {
	tests := []struct {
		name       string
		statements []ast.Statement
	}{
		{
			name: "explicit composite combines explicit and implicit bases",
			statements: []ast.Statement{
				ast.Iface("Named", nil, []*ast.GenericParameter{ast.GenericParam("T")}, ast.Ty("T"), nil, nil, false),
				ast.Iface("Greeter", nil, nil, nil, nil, nil, false),
				ast.Iface(
					"NamedGreeter",
					nil,
					[]*ast.GenericParameter{ast.GenericParam("T")},
					ast.Ty("T"),
					nil,
					[]ast.TypeExpression{ast.Gen(ast.Ty("Named"), ast.Ty("T")), ast.Ty("Greeter")},
					false,
				),
			},
		},
		{
			name: "base arguments are substituted before comparison",
			statements: []ast.Statement{
				ast.Iface(
					"ArrayOnly",
					nil,
					[]*ast.GenericParameter{ast.GenericParam("T")},
					ast.Gen(ast.Ty("Array"), ast.Ty("T")),
					nil,
					nil,
					false,
				),
				ast.Iface(
					"StringArray",
					nil,
					nil,
					ast.Gen(ast.Ty("Array"), ast.Ty("String")),
					nil,
					[]ast.TypeExpression{ast.Gen(ast.Ty("ArrayOnly"), ast.Ty("String"))},
					false,
				),
			},
		},
		{
			name: "interface alias arguments use the resolved base application",
			statements: []ast.Statement{
				ast.Iface(
					"ArrayOnly",
					nil,
					[]*ast.GenericParameter{ast.GenericParam("T")},
					ast.Gen(ast.Ty("Array"), ast.Ty("T")),
					nil,
					nil,
					false,
				),
				ast.NewTypeAliasDefinition(
					ast.ID("NestedArrayOnly"),
					ast.Gen(ast.Ty("ArrayOnly"), ast.Gen(ast.Ty("Array"), ast.Ty("T"))),
					[]*ast.GenericParameter{ast.GenericParam("T")},
					nil,
					false,
				),
				ast.Iface(
					"NestedStringArray",
					nil,
					nil,
					ast.Gen(ast.Ty("Array"), ast.Gen(ast.Ty("Array"), ast.Ty("String"))),
					nil,
					[]ast.TypeExpression{ast.Gen(ast.Ty("NestedArrayOnly"), ast.Ty("String"))},
					false,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := New()
			diags, err := checker.CheckModule(ast.NewModule(tt.statements, nil, nil))
			if err != nil {
				t.Fatalf("CheckModule error: %v", err)
			}
			if len(diags) != 0 {
				t.Fatalf("expected no diagnostics, got %v", diags)
			}
		})
	}
}

func TestCompositeInterfaceSelfPatternRejectsImplicitCompositeWithExplicitBase(t *testing.T) {
	composite := ast.Iface("Mixed", nil, nil, nil, nil, []ast.TypeExpression{ast.Ty("Named")}, false)
	base := ast.Iface("Named", nil, []*ast.GenericParameter{ast.GenericParam("T")}, ast.Ty("T"), nil, nil, false)

	checker := New()
	diags, err := checker.CheckModule(ast.NewModule([]ast.Statement{composite, base}, nil, nil))
	if err != nil {
		t.Fatalf("CheckModule error: %v", err)
	}
	assertCompositeSelfPatternDiagnostic(
		t,
		diags,
		"composite interface 'Mixed' must declare a self type because base interface 'Named' declares self type 'T'",
	)
}

func TestCompositeInterfaceSelfPatternRejectsIncompatibleExplicitBase(t *testing.T) {
	base := ast.Iface("IntegerOnly", nil, nil, ast.Ty("i32"), nil, nil, false)
	composite := ast.Iface(
		"StringComposite",
		nil,
		nil,
		ast.Ty("String"),
		nil,
		[]ast.TypeExpression{ast.Ty("IntegerOnly")},
		false,
	)

	checker := New()
	diags, err := checker.CheckModule(ast.NewModule([]ast.Statement{base, composite}, nil, nil))
	if err != nil {
		t.Fatalf("CheckModule error: %v", err)
	}
	assertCompositeSelfPatternDiagnostic(
		t,
		diags,
		"composite interface 'StringComposite' self type 'String' is incompatible with base interface 'IntegerOnly' self type 'i32'",
	)
}

func assertCompositeSelfPatternDiagnostic(t *testing.T, diags []Diagnostic, want string) {
	t.Helper()
	for _, diag := range diags {
		if strings.Contains(diag.Message, want) {
			return
		}
	}
	t.Fatalf("expected diagnostic containing %q, got %v", want, diags)
}
