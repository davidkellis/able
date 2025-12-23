package parser

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestParseTypeAliasDefinitions(t *testing.T) {
	source := `type UserID = u64

type Box T = Array T

private type ResultValue E T where E: Error + Display = E | T
`

	p, err := NewModuleParser()
	if err != nil {
		t.Fatalf("NewModuleParser error: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule([]byte(source))
	if err != nil {
		t.Fatalf("ParseModule error: %v", err)
	}

	aliasID := ast.NewTypeAliasDefinition(ast.ID("UserID"), ast.Ty("u64"), nil, nil, false)
	aliasBox := ast.NewTypeAliasDefinition(
		ast.ID("Box"),
		ast.Gen(ast.Ty("Array"), ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
	)
	aliasResult := ast.NewTypeAliasDefinition(
		ast.ID("ResultValue"),
		ast.UnionT(ast.Ty("E"), ast.Ty("T")),
		[]*ast.GenericParameter{
			ast.GenericParam("E"),
			ast.GenericParam("T"),
		},
		[]*ast.WhereClauseConstraint{
			ast.WhereConstraint(
				"E",
				ast.InterfaceConstr(ast.Ty("Error")),
				ast.InterfaceConstr(ast.Ty("Display")),
			),
		},
		true,
	)

	expected := ast.NewModule([]ast.Statement{aliasID, aliasBox, aliasResult}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}
