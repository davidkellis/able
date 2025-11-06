package typechecker

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/driver"
)

func TestProgramCheckerResolvesDependencies(t *testing.T) {
	dep := testDependencyModule()
	app := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"use_dep",
				nil,
				[]ast.Statement{
					ast.Ret(
						ast.CallExpr(
							ast.Member(ast.ID("lib"), ast.ID("provide")),
						),
					),
				},
				ast.Ty("string"),
				nil, nil, false, false,
			),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"dep"}, false, nil, "lib")},
		ast.Pkg([]interface{}{"app"}, false),
	)

	depModule := annotatedModule("dep", dep, "dep.able", nil)
	appModule := annotatedModule("app", app, "app.able", []string{"dep"})
	program := &driver.Program{
		Modules: []*driver.Module{
			depModule,
			appModule,
		},
		Entry: appModule,
	}

	pc := NewProgramChecker()
	result, err := pc.Check(program)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %v", result.Diagnostics)
	}
}

func TestProgramCheckerReportsUnknownPackage(t *testing.T) {
	app := ast.Mod(
		[]ast.Statement{
			ast.Fn(
				"use_missing",
				nil,
				[]ast.Statement{
					ast.Ret(ast.Str("noop")),
				},
				ast.Ty("string"),
				nil, nil, false, false,
			),
		},
		[]*ast.ImportStatement{ast.Imp([]interface{}{"missing"}, false, nil, nil)},
		ast.Pkg([]interface{}{"app"}, false),
	)

	appModule := annotatedModule("app", app, "app.able", nil)
	program := &driver.Program{
		Modules: []*driver.Module{appModule},
		Entry:   appModule,
	}

	pc := NewProgramChecker()
	result, err := pc.Check(program)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d (%v)", len(result.Diagnostics), result.Diagnostics)
	}
	if got := result.Diagnostics[0].Diagnostic.Message; got != "typechecker: import references unknown package 'missing'" {
		t.Fatalf("unexpected diagnostic message %q", got)
	}
}

func TestProgramCheckerExposesPublicExports(t *testing.T) {
	dep := ast.Mod(
		[]ast.Statement{
			ast.StructDef("PublicStruct", nil, ast.StructKindNamed, nil, nil, false),
			ast.StructDef("PrivateStruct", nil, ast.StructKindNamed, nil, nil, true),
			ast.Fn("public_fn", nil, []ast.Statement{ast.Ret(ast.Str("ok"))}, ast.Ty("string"), nil, nil, false, false),
			ast.Fn("secret_fn", nil, []ast.Statement{ast.Ret(ast.Str("nope"))}, ast.Ty("string"), nil, nil, false, true),
			ast.Iface("Display", []*ast.FunctionSignature{
				ast.FnSig("show", []*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
				}, ast.Ty("string"), nil, nil, nil),
			}, nil, nil, nil, nil, false),
			ast.Methods(
				ast.Ty("PublicStruct"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"describe",
						[]*ast.FunctionParameter{
							ast.Param("self", ast.Ty("PublicStruct")),
						},
						[]ast.Statement{
							ast.Ret(ast.Str("desc")),
						},
						ast.Ty("string"),
						nil, nil, false, false,
					),
				},
				nil,
				nil,
			),
			ast.Impl(
				"Display",
				ast.Ty("PublicStruct"),
				[]*ast.FunctionDefinition{
					ast.Fn(
						"show",
						[]*ast.FunctionParameter{
							ast.Param("self", ast.Ty("PublicStruct")),
						},
						[]ast.Statement{
							ast.Ret(ast.Str("shown")),
						},
						ast.Ty("string"),
						nil, nil, false, false,
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
		ast.Pkg([]interface{}{"dep"}, false),
	)
	app := ast.Mod(nil, nil, ast.Pkg([]interface{}{"app"}, false))

	depModule := annotatedModule("dep", dep, "dep.able", nil)
	appModule := annotatedModule("app", app, "app.able", nil)
	program := &driver.Program{
		Modules: []*driver.Module{depModule, appModule},
		Entry:   appModule,
	}

	pc := NewProgramChecker()
	result, err := pc.Check(program)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", result.Diagnostics)
	}

	summary, ok := result.Packages["dep"]
	if !ok {
		t.Fatalf("expected package summary for dep")
	}
	if summary.Visibility != "public" {
		t.Fatalf("expected public visibility, got %q", summary.Visibility)
	}

	exports := summary.Symbols
	if exports == nil {
		t.Fatalf("expected exports for dep, got nil")
	}
	if _, ok := exports["PrivateStruct"]; ok {
		t.Fatalf("private struct should not be exported")
	}
	if _, ok := exports["secret_fn"]; ok {
		t.Fatalf("private function should not be exported")
	}
	if _, ok := exports["PublicStruct"]; !ok {
		t.Fatalf("public struct missing from exports")
	}
	if _, ok := exports["public_fn"]; !ok {
		t.Fatalf("public function missing from exports")
	}
	structs := summary.Structs
	if len(structs) != 1 {
		t.Fatalf("expected 1 public struct, got %d", len(structs))
	}
	if _, ok := structs["PublicStruct"]; !ok {
		t.Fatalf("PublicStruct missing from struct metadata")
	}
	interfaces := summary.Interfaces
	if len(interfaces) != 1 {
		t.Fatalf("expected 1 public interface, got %d", len(interfaces))
	}
	if _, ok := interfaces["Display"]; !ok {
		t.Fatalf("Display missing from interface metadata")
	}
	functions := summary.Functions
	if len(functions) != 1 {
		t.Fatalf("expected 1 public function, got %d", len(functions))
	}
	if _, ok := functions["public_fn"]; !ok {
		t.Fatalf("public_fn missing from function metadata")
	}
	impls := summary.Implementations
	if len(impls) != 1 {
		t.Fatalf("expected 1 implementation, got %d", len(impls))
	}
	if impls[0].InterfaceName != "Display" {
		t.Fatalf("unexpected implementation interface %q", impls[0].InterfaceName)
	}
	methodSets := summary.MethodSets
	if len(methodSets) != 1 {
		t.Fatalf("expected 1 method set, got %d", len(methodSets))
	}
	if _, ok := methodSets[0].Methods["describe"]; !ok {
		t.Fatalf("expected describe method in method set")
	}
}

func TestProgramCheckerCapturesPackageVisibility(t *testing.T) {
	privMod := annotatedModule("app.private_pkg", ast.Mod(nil, nil, ast.Pkg([]interface{}{"app", "private_pkg"}, true)), "priv.able", nil)
	program := &driver.Program{
		Modules: []*driver.Module{privMod},
		Entry:   privMod,
	}

	pc := NewProgramChecker()
	result, err := pc.Check(program)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	summary, ok := result.Packages["app.private_pkg"]
	if !ok {
		t.Fatalf("expected summary for private package")
	}
	if summary.Visibility != "private" {
		t.Fatalf("expected private visibility, got %q", summary.Visibility)
	}
}

func testDependencyModule() *ast.Module {
	return ast.Mod(
		[]ast.Statement{
			ast.StructDef("Item", nil, ast.StructKindNamed, nil, nil, false),
			ast.Fn("provide", nil, []ast.Statement{ast.Ret(ast.Str("dep"))}, ast.Ty("string"), nil, nil, false, false),
		},
		nil,
		ast.Pkg([]interface{}{"dep"}, false),
	)
}

func annotatedModule(pkg string, module *ast.Module, file string, imports []string) *driver.Module {
	files := []string{}
	if file != "" {
		files = []string{file}
	}
	origins := make(map[ast.Node]string)
	ast.AnnotateOrigins(module, file, origins)
	return &driver.Module{
		Package:     pkg,
		AST:         module,
		Files:       files,
		Imports:     imports,
		NodeOrigins: origins,
	}
}

func TestProgramCheckerDiagnosticIncludesSourceHint(t *testing.T) {
	tmpDir := t.TempDir()
	source := []byte(`package app

fn main() -> i32 {
  "hi"
}
`)
	entryPath := filepath.Join(tmpDir, "app.able")
	if err := os.WriteFile(entryPath, source, 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("loader.Load: %v", err)
	}

	pc := NewProgramChecker()
	result, err := pc.Check(program)
	if err != nil {
		t.Fatalf("ProgramChecker.Check: %v", err)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics for type mismatch, got none")
	}
	diag := result.Diagnostics[0]
	if diag.Source.Path == "" {
		t.Fatalf("expected diagnostic to include source path")
	}
	if diag.Source.Path != entryPath {
		t.Fatalf("expected diagnostic path %q, got %q", entryPath, diag.Source.Path)
	}
	if diag.Source.Line <= 0 || diag.Source.Column <= 0 {
		t.Fatalf("expected positive line/column, got %d:%d", diag.Source.Line, diag.Source.Column)
	}
}
