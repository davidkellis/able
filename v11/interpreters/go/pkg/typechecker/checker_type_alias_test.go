package typechecker

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestTypeAliasTypedPattern(t *testing.T) {
	alias := ast.NewTypeAliasDefinition(ast.ID("UserID"), ast.Ty("i32"), nil, nil, false)
	binding := ast.Assign(ast.TypedP(ast.ID("value"), ast.Ty("UserID")), ast.Int(42))
	module := ast.Mod([]ast.Statement{alias, binding}, nil, ast.Pkg(nil, false))

	checker := New()
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeAliasGenericInstantiation(t *testing.T) {
	genericAlias := ast.NewTypeAliasDefinition(
		ast.ID("Box"),
		ast.Gen(ast.Ty("Array"), ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	arrayLiteral := ast.Arr(ast.Str("a"), ast.Str("b"))
	binding := ast.Assign(
		ast.TypedP(ast.ID("values"), ast.Gen(ast.Ty("Box"), ast.Ty("String"))),
		arrayLiteral,
	)
	module := ast.Mod([]ast.Statement{genericAlias, binding}, nil, ast.Pkg(nil, false))

	checker := New()
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestTypeAliasExportedSummary(t *testing.T) {
	alias := ast.NewTypeAliasDefinition(ast.ID("UserID"), ast.Ty("i32"), nil, nil, false)
	module := ast.Mod([]ast.Statement{alias}, nil, ast.Pkg([]interface{}{"alias_pkg"}, false))

	checker := New()
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	symbols := checker.ExportedSymbols()
	if len(symbols) != 1 {
		t.Fatalf("expected 1 exported symbol, got %d", len(symbols))
	}
	if symbols[0].Name != "UserID" {
		t.Fatalf("expected alias name 'UserID', got %s", symbols[0].Name)
	}
	if got := formatType(symbols[0].Type); got != "type alias -> i32" {
		t.Fatalf("expected alias summary 'type alias -> i32', got %q", got)
	}
}

func TestTypeAliasWhereClauseEnforced(t *testing.T) {
	alias := ast.NewTypeAliasDefinition(
		ast.ID("DisplayOnly"),
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		[]*ast.WhereClauseConstraint{ast.WhereConstraint("T", ast.InterfaceConstr(ast.Ty("Display")))},
		false,
	)
	structDef := ast.StructDef(
		"Plain",
		[]*ast.StructFieldDefinition{
			ast.NewStructFieldDefinition(ast.Ty("String"), ast.ID("label")),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	plainLiteral := ast.StructLit(
		[]*ast.StructFieldInitializer{
			ast.NewStructFieldInitializer(ast.Str("ok"), ast.ID("label"), false),
		},
		false,
		"Plain",
		nil,
		nil,
	)
	binding := ast.Assign(
		ast.TypedP(ast.ID("bad"), ast.Gen(ast.Ty("DisplayOnly"), ast.Ty("Plain"))),
		plainLiteral,
	)
	module := ast.Mod([]ast.Statement{alias, structDef, binding}, nil, ast.Pkg(nil, false))

	checker := New()
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic enforcing alias constraints")
	}
	if got := diags[0].Message; got == "" {
		t.Fatalf("expected diagnostic message, got empty String")
	}
}
