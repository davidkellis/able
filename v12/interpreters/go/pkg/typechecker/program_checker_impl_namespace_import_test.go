package typechecker

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func TestProgramCheckerRejectsNamedImplementationImportBindingCollision(t *testing.T) {
	first := namedImplementationExportModule("first", "Fancy", false)
	second := namedImplementationExportModule("second", "Fancy", false)
	app := annotatedModule("app", ast.Mod(nil, []*ast.ImportStatement{
		ast.Imp([]interface{}{"first"}, false, []*ast.ImportSelector{ast.ImpSel("Fancy", nil)}, nil),
		ast.Imp([]interface{}{"second"}, false, []*ast.ImportSelector{ast.ImpSel("Fancy", nil)}, nil),
	}, ast.Pkg([]interface{}{"app"}, false)), "app.able", []string{"first", "second"})

	result, err := NewProgramChecker().Check(&driver.Program{Modules: []*driver.Module{first, second, app}, Entry: app})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one collision diagnostic, got %v", result.Diagnostics)
	}
	if got, want := result.Diagnostics[0].Diagnostic.Message, "typechecker: named implementation binding 'Fancy' conflicts with an existing import; use a selector alias"; got != want {
		t.Fatalf("unexpected diagnostic %q, want %q", got, want)
	}
}

func TestProgramCheckerAllowsNamedImplementationImportRename(t *testing.T) {
	first := namedImplementationExportModule("first", "Fancy", false)
	second := namedImplementationExportModule("second", "Fancy", false)
	app := annotatedModule("app", ast.Mod(nil, []*ast.ImportStatement{
		ast.Imp([]interface{}{"first"}, false, []*ast.ImportSelector{ast.ImpSel("Fancy", nil)}, nil),
		ast.Imp([]interface{}{"second"}, false, []*ast.ImportSelector{ast.ImpSel("Fancy", "OtherFancy")}, nil),
	}, ast.Pkg([]interface{}{"app"}, false)), "app.able", []string{"first", "second"})

	result, err := NewProgramChecker().Check(&driver.Program{Modules: []*driver.Module{first, second, app}, Entry: app})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected renamed imports to typecheck, got %v", result.Diagnostics)
	}
}

func TestProgramCheckerRejectsWildcardNamedImplementationImportBindingCollision(t *testing.T) {
	first := namedImplementationExportModule("first", "Fancy", false)
	second := namedImplementationExportModule("second", "Fancy", false)
	app := annotatedModule("app", ast.Mod(nil, []*ast.ImportStatement{
		ast.Imp([]interface{}{"first"}, true, nil, nil),
		ast.Imp([]interface{}{"second"}, true, nil, nil),
	}, ast.Pkg([]interface{}{"app"}, false)), "app.able", []string{"first", "second"})

	result, err := NewProgramChecker().Check(&driver.Program{Modules: []*driver.Module{first, second, app}, Entry: app})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one wildcard collision diagnostic, got %v", result.Diagnostics)
	}
	if got, want := result.Diagnostics[0].Diagnostic.Message, "typechecker: named implementation binding 'Fancy' conflicts with an existing import; use a selector alias"; got != want {
		t.Fatalf("unexpected diagnostic %q, want %q", got, want)
	}
}

func TestProgramCheckerRejectsLocalNamedImplementationCollisionWithImport(t *testing.T) {
	dep := namedImplementationExportModule("dep", "Fancy", false)
	appModule := ast.Mod([]ast.Statement{
		ast.Impl("Marker", ast.Ty("Widget"), nil, "Fancy", nil, nil, nil, false),
	}, []*ast.ImportStatement{
		ast.Imp([]interface{}{"dep"}, false, []*ast.ImportSelector{
			ast.ImpSel("Marker", nil),
			ast.ImpSel("Widget", nil),
			ast.ImpSel("Fancy", nil),
		}, nil),
	}, ast.Pkg([]interface{}{"app"}, false))
	app := annotatedModule("app", appModule, "app.able", []string{"dep"})

	result, err := NewProgramChecker().Check(&driver.Program{Modules: []*driver.Module{dep, app}, Entry: app})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one collision diagnostic, got %v", result.Diagnostics)
	}
	if got, want := result.Diagnostics[0].Diagnostic.Message, "typechecker: named implementation binding 'Fancy' conflicts with an imported binding; use a selector alias"; got != want {
		t.Fatalf("unexpected diagnostic %q, want %q", got, want)
	}
}

func TestProgramCheckerRejectsPrivateNamedImplementationImport(t *testing.T) {
	dep := namedImplementationExportModule("dep", "Hidden", true)
	app := annotatedModule("app", ast.Mod(nil, []*ast.ImportStatement{
		ast.Imp([]interface{}{"dep"}, false, []*ast.ImportSelector{ast.ImpSel("Hidden", nil)}, nil),
	}, ast.Pkg([]interface{}{"app"}, false)), "app.able", []string{"dep"})

	result, err := NewProgramChecker().Check(&driver.Program{Modules: []*driver.Module{dep, app}, Entry: app})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one privacy diagnostic, got %v", result.Diagnostics)
	}
	if got, want := result.Diagnostics[0].Diagnostic.Message, "typechecker: package 'dep' symbol 'Hidden' is private"; got != want {
		t.Fatalf("unexpected diagnostic %q, want %q", got, want)
	}
}

func namedImplementationExportModule(pkg, name string, private bool) *driver.Module {
	module := ast.Mod([]ast.Statement{
		ast.StructDef("Widget", nil, ast.StructKindNamed, nil, nil, false),
		ast.Iface("Marker", nil, nil, nil, nil, nil, false),
		ast.Impl("Marker", ast.Ty("Widget"), nil, name, nil, nil, nil, private),
	}, nil, ast.Pkg([]interface{}{pkg}, false))
	return annotatedModule(pkg, module, pkg+".able", nil)
}
