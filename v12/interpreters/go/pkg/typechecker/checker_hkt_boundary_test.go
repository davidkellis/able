package typechecker

import (
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestRuntimeFieldRejectsUnboundTypeConstructor(t *testing.T) {
	field := ast.NewStructFieldDefinition(
		ast.Gen(ast.Ty("Array"), ast.WildT()),
		ast.ID("values"),
	)
	module := ast.Mod(
		[]ast.Statement{
			ast.StructDef("Bad", []*ast.StructFieldDefinition{field}, ast.StructKindNamed, nil, nil, false),
		},
		nil,
		nil,
	)

	checker := New()
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	assertUnboundValueTypeDiagnostic(t, diags)
}

func TestTypeConstructorAliasCannotAnnotateRuntimeValue(t *testing.T) {
	alias := ast.NewTypeAliasDefinition(
		ast.ID("AnyArray"),
		ast.Gen(ast.Ty("Array"), ast.WildT()),
		nil,
		nil,
		false,
	)
	binding := ast.Assign(
		ast.TypedP(ast.ID("values"), ast.Ty("AnyArray")),
		ast.Arr(ast.Int(1), ast.Int(2)),
	)
	module := ast.Mod([]ast.Statement{alias, binding}, nil, nil)

	checker := New()
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	assertUnboundValueTypeDiagnostic(t, diags)
}

func TestTypeConstructorAliasDeclarationRemainsValid(t *testing.T) {
	alias := ast.NewTypeAliasDefinition(
		ast.ID("AnyArray"),
		ast.Gen(ast.Ty("Array"), ast.WildT()),
		nil,
		nil,
		false,
	)

	checker := New()
	diags, err := checker.CheckModule(ast.Mod([]ast.Statement{alias}, nil, nil))
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected constructor alias declaration to remain valid, got %v", diags)
	}
}

func TestOrdinaryGenericParameterCannotBeAppliedAsConstructor(t *testing.T) {
	def := ast.Fn(
		"keep",
		[]*ast.FunctionParameter{
			ast.Param("value", ast.Gen(ast.Ty("F"), ast.Ty("T"))),
		},
		[]ast.Statement{ast.ID("value")},
		ast.Gen(ast.Ty("F"), ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("F"), ast.GenericParam("T")},
		nil,
		false,
		false,
	)

	checker := New()
	diags, err := checker.CheckModule(ast.Mod([]ast.Statement{def}, nil, nil))
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	for _, diag := range diags {
		if strings.Contains(diag.Message, "ordinary generic parameters cannot be applied as type constructors") {
			return
		}
	}
	t.Fatalf("expected ordinary higher-kinded parameter diagnostic, got %v", diags)
}

func assertUnboundValueTypeDiagnostic(t *testing.T, diags []Diagnostic) {
	t.Helper()
	for _, diag := range diags {
		if strings.Contains(diag.Message, "runtime value types must bind every type argument") {
			return
		}
	}
	t.Fatalf("expected unbound runtime value type diagnostic, got %v", diags)
}
