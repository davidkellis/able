package parser

import (
	"testing"

	"able/interpreter-go/pkg/ast"
)

func TestParseCompositeInterfacePreservesSelfPatternAndBaseArguments(t *testing.T) {
	source := `interface Mixed<T> for Array T = Named<T> + Greeter
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

	expected := ast.NewModule([]ast.Statement{
		ast.Iface(
			"Mixed",
			[]*ast.FunctionSignature{},
			[]*ast.GenericParameter{ast.GenericParam("T")},
			ast.Gen(ast.Ty("Array"), ast.Ty("T")),
			nil,
			[]ast.TypeExpression{
				ast.Gen(ast.Ty("Named"), ast.Ty("T")),
				ast.Ty("Greeter"),
			},
			false,
		),
	}, nil, nil)
	expected.Imports = []*ast.ImportStatement{}

	assertModulesEqual(t, expected, mod)
}
