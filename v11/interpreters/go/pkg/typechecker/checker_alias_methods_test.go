package typechecker

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestAliasMethodsPropagateAcrossPackages(t *testing.T) {
	libModule := ast.Mod(
		[]ast.Statement{
			ast.NewTypeAliasDefinition(
				ast.ID("Bag"),
				ast.Gen(ast.Ty("Array"), ast.Ty("T")),
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
				true,
			),
			ast.NewTypeAliasDefinition(
				ast.ID("StrList"),
				ast.Gen(ast.Ty("Array"), ast.Ty("String")),
				nil,
				nil,
				true,
			),
			ast.Methods(
				ast.Gen(ast.Ty("Bag"), ast.Ty("T")),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"head",
						[]*ast.FunctionParameter{
							ast.Param("self", ast.Ty("Self")),
						},
						[]ast.Statement{
							ast.Ret(ast.Index(ast.ID("self"), ast.Int(0))),
						},
						ast.Nullable(ast.Ty("T")),
						[]*ast.GenericParameter{ast.GenericParam("T")},
						nil,
						false,
						false,
					),
				},
				[]*ast.GenericParameter{ast.GenericParam("T")},
				nil,
			),
			ast.Impl(
				ast.ID("Display"),
				ast.Ty("StrList"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"to_String",
						[]*ast.FunctionParameter{
							ast.Param("self", ast.Ty("Self")),
						},
						[]ast.Statement{
							ast.Ret(ast.Str("<strlist>")),
						},
						ast.Ty("String"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				nil,
				nil,
				nil,
				false,
			),
		},
		nil,
		ast.Pkg([]interface{}{"pkg"}, false),
	)

	libChecker := New()
	if diags, err := libChecker.CheckModule(libModule); err != nil {
		t.Fatalf("library CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("library diagnostics: %v", diags)
	}

	env := typeEnvForImports(t, "pkg", libChecker.ExportedSymbols(), nil, map[string]string{}, false, "", true)
	consumer := New()
	consumer.SetPrelude(env, libChecker.ModuleImplementations(), libChecker.ModuleMethodSets())

	consumerModule := ast.Mod(
		[]ast.Statement{
			ast.Assign(ast.ID("arr"), ast.Arr(ast.Str("a"), ast.Str("b"))),
			ast.CallExpr(ast.Member(ast.ID("arr"), ast.ID("head"))),
			ast.CallExpr(ast.Member(ast.ID("arr"), ast.ID("to_String"))),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"pkg"}, false, nil, nil)},
		ast.Pkg([]interface{}{"app"}, false),
	)

	if diags, err := consumer.CheckModule(consumerModule); err != nil {
		t.Fatalf("consumer CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("consumer diagnostics: %v", diags)
	}
}

