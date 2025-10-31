package interpreter

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/typechecker"
)

// runFixtureWithExecutor replays a fixture directory using the provided executor.
func runFixtureWithExecutor(t testingT, dir string, executor Executor) {
	t.Helper()
	underlying, ok := t.(*testing.T)
	if !ok {
		panic("runFixtureWithExecutor expects *testing.T")
	}
	manifest := readManifest(underlying, dir)
	if shouldSkipTarget(manifest.SkipTargets, "go") {
		return
	}
	entry := manifest.Entry
	if entry == "" {
		entry = "module.json"
	}
	modulePath := filepath.Join(dir, entry)
	module := readModule(underlying, modulePath)

	interp := NewWithExecutor(executor)
	mode := configureFixtureTypechecker(interp)
	var stdout []string
	registerPrint(interp, &stdout)

	var programModules []*driver.Module
	for _, setup := range manifest.Setup {
		setupPath := filepath.Join(dir, setup)
		setupModule := readModule(underlying, setupPath)
		programModules = append(programModules, fixtureDriverModule(setupModule, setupPath))
	}

	entryModule := fixtureDriverModule(module, modulePath)
	programModules = append(programModules, entryModule)
	program := &driver.Program{
		Entry:   entryModule,
		Modules: programModules,
	}

	value, _, moduleDiags, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    mode == typecheckModeOff,
		AllowDiagnostics: mode == typecheckModeWarn,
	})

	checkerDiags := extractDiagnostics(moduleDiags)
	checkFixtureTypecheckDiagnostics(underlying, mode, manifest.Expect.TypecheckDiagnostics, checkerDiags)

	if len(moduleDiags) > 0 && mode != typecheckModeWarn {
		// Diagnostics prevented evaluation; nothing further to assert.
		return
	}

	if len(manifest.Expect.Errors) > 0 {
		if err == nil {
			t.Fatalf("expected evaluation error")
		}
		msg := extractErrorMessage(err)
		if !contains(manifest.Expect.Errors, msg) {
			t.Fatalf("expected error in %v, got %s", manifest.Expect.Errors, msg)
		}
		return
	}
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}
	assertResult(underlying, dir, manifest, value, stdout)
}

func shouldSkipTarget(skip []string, target string) bool {
	if len(skip) == 0 {
		return false
	}
	target = strings.ToLower(target)
	for _, entry := range skip {
		if strings.ToLower(strings.TrimSpace(entry)) == target {
			return true
		}
	}
	return false
}

// testingT captures the subset of testing.T used by fixture helpers.
type testingT interface {
	Helper()
	Fatalf(format string, args ...interface{})
}

func fixtureDriverModule(module *ast.Module, file string) *driver.Module {
	var pkgName string
	if module != nil && module.Package != nil {
		parts := make([]string, 0, len(module.Package.NamePath))
		for _, id := range module.Package.NamePath {
			if id == nil || id.Name == "" {
				continue
			}
			parts = append(parts, id.Name)
		}
		pkgName = strings.Join(parts, ".")
	}
	importSet := make(map[string]struct{})
	for _, imp := range module.Imports {
		if imp == nil {
			continue
		}
		parts := make([]string, 0, len(imp.PackagePath))
		for _, id := range imp.PackagePath {
			if id == nil || id.Name == "" {
				continue
			}
			parts = append(parts, id.Name)
		}
		if len(parts) == 0 {
			continue
		}
		importSet[strings.Join(parts, ".")] = struct{}{}
	}
	imports := make([]string, 0, len(importSet))
	for name := range importSet {
		imports = append(imports, name)
	}
	sort.Strings(imports)

	files := []string{}
	if file != "" {
		files = []string{file}
	}
	return &driver.Module{
		Package: pkgName,
		AST:     module,
		Files:   files,
		Imports: imports,
	}
}

func extractDiagnostics(diags []ModuleDiagnostic) []typechecker.Diagnostic {
	if len(diags) == 0 {
		return nil
	}
	out := make([]typechecker.Diagnostic, len(diags))
	for i, diag := range diags {
		out[i] = diag.Diagnostic
	}
	return out
}
