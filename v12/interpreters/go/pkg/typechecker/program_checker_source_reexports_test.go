package typechecker

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func TestProgramCheckerPublishesNamedSourceReexport(t *testing.T) {
	source := ast.Mod(
		[]ast.Statement{ast.Fn("value", nil, []ast.Statement{ast.Ret(ast.Int(42))}, ast.Ty("i32"), nil, nil, false, false)},
		nil,
		ast.Pkg([]interface{}{"source"}, false),
	)
	wrapper := ast.NewModuleWithExports(
		nil,
		[]*ast.ImportStatement{ast.Imp([]interface{}{"source"}, false, []*ast.ImportSelector{ast.ImpSel("value", nil)}, nil)},
		[]*ast.ExportStatement{ast.Exp("value")},
		ast.Pkg([]interface{}{"wrapper"}, false),
	)
	app := ast.Mod(
		[]ast.Statement{ast.Fn("main", nil, []ast.Statement{ast.Ret(ast.Call("value"))}, ast.Ty("i32"), nil, nil, false, false)},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"wrapper"}, false, []*ast.ImportSelector{ast.ImpSel("value", nil)}, nil)},
		ast.Pkg([]interface{}{"app"}, false),
	)

	result, err := NewProgramChecker().Check(&driver.Program{
		Modules: []*driver.Module{
			annotatedModule("source", source, "source.able", nil),
			annotatedModule("wrapper", wrapper, "wrapper.able", []string{"source"}),
			annotatedModule("app", app, "app.able", []string{"wrapper"}),
		},
	})
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", result.Diagnostics)
	}
	if _, ok := result.Packages["wrapper"].Functions["value"]; !ok {
		t.Fatalf("named re-export missing from wrapper function summary: %#v", result.Packages["wrapper"])
	}
}

func TestProgramCheckerPublishesWildcardSourceReexport(t *testing.T) {
	source := ast.Mod(
		[]ast.Statement{ast.Fn("value", nil, []ast.Statement{ast.Ret(ast.Int(42))}, ast.Ty("i32"), nil, nil, false, false)},
		nil,
		ast.Pkg([]interface{}{"source"}, false),
	)
	wrapper := ast.NewModuleWithExports(
		nil,
		nil,
		[]*ast.ExportStatement{ast.ExpAll([]interface{}{"source"})},
		ast.Pkg([]interface{}{"wrapper"}, false),
	)
	app := ast.Mod(
		[]ast.Statement{ast.Fn("main", nil, []ast.Statement{ast.Ret(ast.Call("value"))}, ast.Ty("i32"), nil, nil, false, false)},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"wrapper"}, true, nil, nil)},
		ast.Pkg([]interface{}{"app"}, false),
	)

	result, err := NewProgramChecker().Check(&driver.Program{
		Modules: []*driver.Module{
			annotatedModule("source", source, "source.able", nil),
			annotatedModule("wrapper", wrapper, "wrapper.able", []string{"source"}),
			annotatedModule("app", app, "app.able", []string{"wrapper"}),
		},
	})
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", result.Diagnostics)
	}
	if _, ok := result.Packages["wrapper"].Symbols["value"]; !ok {
		t.Fatalf("wildcard re-export missing from wrapper summary: %#v", result.Packages["wrapper"])
	}
}

func TestProgramCheckerRejectsPrivateNamedSourceReexport(t *testing.T) {
	source := ast.Mod(
		[]ast.Statement{ast.Fn("hidden", nil, []ast.Statement{ast.Ret(ast.Int(42))}, ast.Ty("i32"), nil, nil, false, true)},
		nil,
		ast.Pkg([]interface{}{"source"}, false),
	)
	wrapper := ast.NewModuleWithExports(
		nil,
		[]*ast.ImportStatement{ast.Imp([]interface{}{"source"}, false, []*ast.ImportSelector{ast.ImpSel("hidden", nil)}, nil)},
		[]*ast.ExportStatement{ast.Exp("hidden")},
		ast.Pkg([]interface{}{"wrapper"}, false),
	)

	result, err := NewProgramChecker().Check(&driver.Program{Modules: []*driver.Module{
		annotatedModule("source", source, "source.able", nil),
		annotatedModule("wrapper", wrapper, "wrapper.able", []string{"source"}),
	}})
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if len(result.Diagnostics) != 2 {
		t.Fatalf("expected import and re-export privacy diagnostics, got %v", result.Diagnostics)
	}
	if _, leaked := result.Packages["wrapper"].Symbols["hidden"]; leaked {
		t.Fatalf("private source symbol leaked through wrapper summary: %#v", result.Packages["wrapper"])
	}
}
