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
						"to_string",
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
			ast.CallExpr(ast.Member(ast.ID("arr"), ast.ID("to_string"))),
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

func TestAliasTypeQualifiedMethodsUseCanonicalExport(t *testing.T) {
	coreModule := ast.Mod(
		[]ast.Statement{
			ast.StructDef(
				"Widget",
				[]*ast.StructFieldDefinition{
					ast.FieldDef(ast.Ty("i32"), "value"),
				},
				ast.StructKindNamed,
				nil,
				nil,
				false,
			),
		},
		nil,
		ast.Pkg([]interface{}{"corealias"}, false),
	)

	coreChecker := New()
	if diags, err := coreChecker.CheckModule(coreModule); err != nil {
		t.Fatalf("core CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("core diagnostics: %v", diags)
	}

	extModule := ast.Mod(
		[]ast.Statement{
			ast.NewTypeAliasDefinition(ast.ID("AliasWidget"), ast.Ty("Widget"), nil, nil, true),
			ast.Methods(
				ast.Ty("AliasWidget"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"make",
						[]*ast.FunctionParameter{
							ast.Param("value", ast.Ty("i32")),
						},
						[]ast.Statement{
							ast.Ret(
								ast.StructLit(
									[]*ast.StructFieldInitializer{
										ast.FieldInit(ast.ID("value"), "value"),
									},
									false,
									"Widget",
									nil,
									nil,
								),
							),
						},
						ast.Ty("Widget"),
						nil,
						nil,
						false,
						false,
					),
				},
				nil,
				nil,
			),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"corealias"}, true, nil, nil),
		},
		ast.Pkg([]interface{}{"extalias"}, false),
	)

	extChecker := New()
	extPrelude := typeEnvForImports(t, "corealias", coreChecker.ExportedSymbols(), nil, wildcardBinding(coreChecker.ExportedSymbols()), false, "", true)
	extChecker.SetPrelude(extPrelude, coreChecker.ModuleImplementations(), coreChecker.ModuleMethodSets())
	if diags, err := extChecker.CheckModule(extModule); err != nil {
		t.Fatalf("ext CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("ext diagnostics: %v", diags)
	}

	extSymbols := make(map[string]Type, len(extChecker.ExportedSymbols()))
	for _, sym := range extChecker.ExportedSymbols() {
		extSymbols[sym.Name] = sym.Type
	}
	if _, ok := extSymbols["Widget.make"]; !ok {
		t.Fatalf("expected ext exports to include Widget.make, got %v", extSymbols)
	}

	coreSymbols := make(map[string]Type, len(coreChecker.ExportedSymbols()))
	for _, sym := range coreChecker.ExportedSymbols() {
		coreSymbols[sym.Name] = sym.Type
	}

	prelude := NewEnvironment(nil)
	prelude.Define("corealias", PackageType{Package: "corealias", Symbols: coreSymbols})
	prelude.Define("extalias", PackageType{Package: "extalias", Symbols: extSymbols})
	for name, typ := range coreSymbols {
		prelude.Define(name, typ)
	}
	for name, typ := range extSymbols {
		prelude.Define(name, typ)
	}

	consumer := New()
	consumer.SetPrelude(
		prelude,
		append(coreChecker.ModuleImplementations(), extChecker.ModuleImplementations()...),
		append(coreChecker.ModuleMethodSets(), extChecker.ModuleMethodSets()...),
	)

	consumerModule := ast.Mod(
		[]ast.Statement{
			ast.Assign(
				ast.ID("inst"),
				ast.CallExpr(ast.Member(ast.ID("Widget"), ast.ID("make")), ast.Int(5)),
			),
			ast.Member(ast.ID("inst"), ast.ID("value")),
		},
		[]*ast.ImportStatement{
			ast.Imp([]interface{}{"corealias"}, true, nil, nil),
			ast.Imp([]interface{}{"extalias"}, true, nil, nil),
		},
		ast.Pkg([]interface{}{"appalias"}, false),
	)

	if diags, err := consumer.CheckModule(consumerModule); err != nil {
		t.Fatalf("consumer CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("consumer diagnostics: %v", diags)
	}
}
