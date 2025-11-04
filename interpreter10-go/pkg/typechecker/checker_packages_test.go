package typechecker

import (
	"testing"

	"able/interpreter10-go/pkg/ast"
)

func TestCheckerExportsRespectPrivacy(t *testing.T) {
	module := ast.Mod(
		[]ast.Statement{
			ast.StructDef("PublicStruct", nil, ast.StructKindNamed, nil, nil, false),
			ast.StructDef("PrivateStruct", nil, ast.StructKindNamed, nil, nil, true),
			ast.Fn("public_fn", nil, []ast.Statement{ast.Ret(ast.Str("ok"))}, ast.Ty("string"), nil, nil, false, false),
			ast.Fn("secret_fn", nil, []ast.Statement{ast.Ret(ast.Str("nope"))}, ast.Ty("string"), nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"dep"}, false),
	)

	checker := New()
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("CheckModule returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	symbols := checker.ExportedSymbols()
	if len(symbols) != 2 {
		t.Fatalf("expected 2 exported symbols, got %d (%v)", len(symbols), symbols)
	}
	for _, sym := range symbols {
		if sym.Name == "PrivateStruct" || sym.Name == "secret_fn" {
			t.Fatalf("private symbol %s should not be exported", sym.Name)
		}
	}
}

func TestCheckerPackageAliasMemberAccess(t *testing.T) {
	depModule := dependencyModule()
	depChecker := New()
	if diags, err := depChecker.CheckModule(depModule); err != nil {
		t.Fatalf("dependency CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("dependency diagnostics: %v", diags)
	}

	mainModule := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"use_dep",
				nil,
				[]ast.Statement{
					ast.Ret(ast.CallExpr(ast.Member(ast.ID("lib"), ast.ID("provide")))),
				},
				ast.Ty("string"),
				nil, nil, false, false,
			),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, false, nil, "lib")},
		ast.Pkg([]interface{}{"app"}, false),
	)

	env := typeEnvForImports(t, "dep", depChecker.ExportedSymbols(), map[string]string{}, false, "lib", true)
	mainChecker := New()
	mainChecker.SetPrelude(env, depChecker.ModuleImplementations(), depChecker.ModuleMethodSets())

	if diags, err := mainChecker.CheckModule(mainModule); err != nil {
		t.Fatalf("main CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("main diagnostics: %v", diags)
	}
}

func TestCheckerWildcardImportBindsSymbols(t *testing.T) {
	depModule := dependencyModule()
	depChecker := New()
	if diags, err := depChecker.CheckModule(depModule); err != nil {
		t.Fatalf("dependency CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("dependency diagnostics: %v", diags)
	}

	mainModule := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"build_item",
				nil,
				[]ast.Statement{
					ast.Ret(ast.StructLit(nil, false, "Item", nil, nil)),
				},
				ast.Ty("Item"),
				nil, nil, false, false,
			),
			ast.Fn(
				"use_provide",
				nil,
				[]ast.Statement{
					ast.Ret(ast.CallExpr(ast.ID("provide"))),
				},
				ast.Ty("string"),
				nil, nil, false, false,
			),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, true, nil, nil)},
		ast.Pkg([]interface{}{"app"}, false),
	)

	env := typeEnvForImports(t, "dep", depChecker.ExportedSymbols(), wildcardBinding(depChecker.ExportedSymbols()), false, "", false)
	mainChecker := New()
	mainChecker.SetPrelude(env, depChecker.ModuleImplementations(), depChecker.ModuleMethodSets())

	if diags, err := mainChecker.CheckModule(mainModule); err != nil {
		t.Fatalf("main CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("main diagnostics: %v", diags)
	}
}

func TestCheckerSelectiveImportBindsSymbol(t *testing.T) {
	depModule := dependencyModule()
	depChecker := New()
	if diags, err := depChecker.CheckModule(depModule); err != nil {
		t.Fatalf("dependency CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("dependency diagnostics: %v", diags)
	}

	mainModule := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"call_provide",
				nil,
				[]ast.Statement{
					ast.Ret(ast.CallExpr(ast.ID("alias"))),
				},
				ast.Ty("string"),
				nil, nil, false, false,
			),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, false, []*ast.ImportSelector{ast.ImpSel("provide", "alias")}, nil)},
		ast.Pkg([]interface{}{"app"}, false),
	)

	env := typeEnvForImports(t, "dep", depChecker.ExportedSymbols(), map[string]string{"provide": "alias"}, false, "", false)
	mainChecker := New()
	mainChecker.SetPrelude(env, depChecker.ModuleImplementations(), depChecker.ModuleMethodSets())

	if diags, err := mainChecker.CheckModule(mainModule); err != nil {
		t.Fatalf("main CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("main diagnostics: %v", diags)
	}
}

func TestCheckerPrivateSymbolsNotImported(t *testing.T) {
	depModule := ast.Mod(
		[]ast.Statement{
			ast.Fn("provide", nil, []ast.Statement{ast.Ret(ast.Str("dep"))}, ast.Ty("string"), nil, nil, false, false),
			ast.Fn("secret", nil, []ast.Statement{ast.Ret(ast.Str("hidden"))}, ast.Ty("string"), nil, nil, false, true),
		},
		nil,
		ast.Pkg([]interface{}{"dep"}, false),
	)
	depChecker := New()
	if diags, err := depChecker.CheckModule(depModule); err != nil {
		t.Fatalf("dependency CheckModule error: %v", err)
	} else if len(diags) != 0 {
		t.Fatalf("dependency diagnostics: %v", diags)
	}

	mainModule := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"use_hidden",
				nil,
				[]ast.Statement{
					ast.Ret(ast.CallExpr(ast.Member(ast.ID("lib"), ast.ID("secret")))),
				},
				ast.Ty("string"),
				nil, nil, false, false,
			),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, false, nil, "lib")},
		ast.Pkg([]interface{}{"app"}, false),
	)

	env := typeEnvForImports(t, "dep", depChecker.ExportedSymbols(), map[string]string{}, false, "lib", true)
	mainChecker := New()
	mainChecker.SetPrelude(env, depChecker.ModuleImplementations(), depChecker.ModuleMethodSets())

	diags, err := mainChecker.CheckModule(mainModule)
	if err != nil {
		t.Fatalf("main CheckModule error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for missing private symbol access, got %v", diags)
	}
}

func dependencyModule() *ast.Module {
	return ast.Mod(
		[]ast.Statement{
			ast.StructDef("Item", nil, ast.StructKindNamed, nil, nil, false),
			ast.Fn("provide", nil, []ast.Statement{ast.Ret(ast.Str("dep"))}, ast.Ty("string"), nil, nil, false, false),
		},
		nil,
		ast.Pkg([]interface{}{"dep"}, false),
	)
}

func typeEnvForImports(t *testing.T, pkgName string, exports []ExportedSymbol, bindings map[string]string, wildcard bool, alias string, includePackage bool) *Environment {
	t.Helper()
	env := NewEnvironment(nil)
	symbols := make(map[string]Type, len(exports))
	for _, sym := range exports {
		symbols[sym.Name] = sym.Type
	}
	if includePackage {
		pkgAlias := alias
		if pkgAlias == "" {
			pkgAlias = pkgName
		}
		env.Define(pkgAlias, PackageType{Package: pkgName, Symbols: symbols})
	}
	for name, aliasName := range bindings {
		typ, ok := symbols[name]
		if !ok {
			typ = UnknownType{}
		}
		env.Define(aliasName, typ)
	}
	if wildcard {
		for name, typ := range symbols {
			env.Define(name, typ)
		}
	}
	return env
}

func wildcardBinding(exports []ExportedSymbol) map[string]string {
	bind := make(map[string]string, len(exports))
	for _, sym := range exports {
		bind[sym.Name] = sym.Name
	}
	return bind
}
