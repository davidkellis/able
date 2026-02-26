package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
)

func TestDetectDynamicFeaturesDynImportBindings(t *testing.T) {
	root := t.TempDir()
	writePackageConfig(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "dynimport demo.dynamic.{answer}\n\nfn main() -> i32 { answer() }\nfn helper() -> i32 { 1 }\n"
	writeFile(t, entryPath, source)

	program := loadProgram(t, entryPath)
	report, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("detect dynamic features: %v", err)
	}

	moduleUsage := report.Modules["demo"]
	if !moduleUsage.HasDynImports {
		t.Fatalf("expected dynimports on demo module")
	}
	if moduleUsage.HasDynImportWildcard {
		t.Fatalf("expected no wildcard dynimport")
	}
	if !moduleUsage.HasDynamicCalls {
		t.Fatalf("expected dynamic usage in demo module")
	}

	if _, ok := report.DynBindings["demo"]["answer"]; !ok {
		t.Fatalf("expected dynimport binding for answer")
	}

	mainUsage, ok := findFunctionUsage(report, "main")
	if !ok || !mainUsage.UsesDynamic {
		t.Fatalf("expected main to use dynamic features")
	}
	helperUsage, ok := findFunctionUsage(report, "helper")
	if !ok || helperUsage.UsesDynamic {
		t.Fatalf("expected helper to remain static")
	}
}

func TestDetectDynamicFeaturesWildcardDynImport(t *testing.T) {
	root := t.TempDir()
	writePackageConfig(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "dynimport demo.dynamic.*\n\nfn main() -> i32 { 1 }\n"
	writeFile(t, entryPath, source)

	program := loadProgram(t, entryPath)
	report, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("detect dynamic features: %v", err)
	}

	moduleUsage := report.Modules["demo"]
	if !moduleUsage.HasDynImports || !moduleUsage.HasDynImportWildcard {
		t.Fatalf("expected wildcard dynimport on demo module")
	}
	mainUsage, ok := findFunctionUsage(report, "main")
	if !ok || !mainUsage.UsesDynamic {
		t.Fatalf("expected wildcard dynimport to mark main as dynamic")
	}
}

func TestDetectDynamicFeaturesDynMemberCall(t *testing.T) {
	root := t.TempDir()
	writePackageConfig(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void { dyn.def_package(\"demo.dynamic\") }\n"
	writeFile(t, entryPath, source)

	program := loadProgram(t, entryPath)
	report, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("detect dynamic features: %v", err)
	}

	moduleUsage := report.Modules["demo"]
	if moduleUsage.HasDynImports {
		t.Fatalf("expected no dynimports on demo module")
	}
	if !moduleUsage.HasDynamicCalls {
		t.Fatalf("expected dynamic usage due to dyn.def_package")
	}
	mainUsage, ok := findFunctionUsage(report, "main")
	if !ok || !mainUsage.UsesDynamic {
		t.Fatalf("expected main to use dynamic features")
	}
}

func TestDetectDynamicFeaturesIncludesEntryWhenModulesSliceEmpty(t *testing.T) {
	root := t.TempDir()
	writePackageConfig(t, root, "demo")
	entryPath := filepath.Join(root, "main.able")
	source := "fn main() -> void { dyn.def_package(\"demo.dynamic\") }\n"
	writeFile(t, entryPath, source)

	program := loadProgram(t, entryPath)
	program.Modules = nil

	report, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("detect dynamic features: %v", err)
	}
	if !report.UsesDynamic() {
		t.Fatalf("expected entry module dynamic usage to be detected")
	}
	entryUsage, ok := report.Modules[program.Entry.Package]
	if !ok {
		t.Fatalf("expected entry package %q in module usage map", program.Entry.Package)
	}
	if !entryUsage.UsesDynamic() {
		t.Fatalf("expected entry package %q to be marked dynamic", program.Entry.Package)
	}
}

func TestDetectDynamicFeaturesUsesDynamicIgnoresUnreachableModules(t *testing.T) {
	entryMain := ast.Fn(
		"main",
		nil,
		[]ast.Statement{ast.Ret(nil)},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	entryAst := ast.Mod(
		[]ast.Statement{entryMain},
		nil,
		ast.Pkg([]interface{}{"app"}, false),
	)
	entry := annotatedModule("app", entryAst, "app.able", nil)

	dynFn := ast.Fn(
		"probe",
		nil,
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("dyn"), "def_package"), ast.Str("demo.dynamic")),
			ast.Ret(nil),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)
	dynAst := ast.Mod(
		[]ast.Statement{dynFn},
		nil,
		ast.Pkg([]interface{}{"tools"}, false),
	)
	tools := annotatedModule("tools", dynAst, "tools.able", nil)

	program := &driver.Program{
		Entry:   entry,
		Modules: []*driver.Module{entry, tools},
	}
	report, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("detect dynamic features: %v", err)
	}
	if report.UsesDynamic() {
		t.Fatalf("expected dynamic usage to ignore unreachable modules")
	}
	if _, ok := report.ReachablePackages["app"]; !ok {
		t.Fatalf("expected entry package to be reachable")
	}
	if _, ok := report.ReachablePackages["tools"]; ok {
		t.Fatalf("did not expect tools package to be reachable")
	}
	if usage, ok := report.Modules["tools"]; !ok || !usage.UsesDynamic() {
		t.Fatalf("expected tools module to be marked dynamic in raw module usage")
	}
}

func writePackageConfig(t *testing.T, root, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: "+name+"\n"), 0o644); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
}

func writeFile(t *testing.T, path, source string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func loadProgram(t *testing.T, entryPath string) *driver.Program {
	t.Helper()
	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}
	return program
}

func findFunctionUsage(report *DynamicFeatureReport, name string) (DynamicFunctionUsage, bool) {
	for _, usage := range report.Functions {
		if usage.Name == name {
			return usage, true
		}
	}
	return DynamicFunctionUsage{}, false
}
