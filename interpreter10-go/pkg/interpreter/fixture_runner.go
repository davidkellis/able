package interpreter

import (
	"path/filepath"
	"testing"
)

// runFixtureWithExecutor replays a fixture directory using the provided executor.
func runFixtureWithExecutor(t testingT, dir string, executor Executor) {
	t.Helper()
	underlying, ok := t.(*testing.T)
	if !ok {
		panic("runFixtureWithExecutor expects *testing.T")
	}
	manifest := readManifest(underlying, dir)
	entry := manifest.Entry
	if entry == "" {
		entry = "module.json"
	}
	modulePath := filepath.Join(dir, entry)
	module := readModule(underlying, modulePath)

	interpreter := NewWithExecutor(executor)
	mode := configureFixtureTypechecker(interpreter)
	var stdout []string
	registerPrint(interpreter, &stdout)

	if len(manifest.Setup) > 0 {
		for _, setup := range manifest.Setup {
			setupModule := readModule(underlying, filepath.Join(dir, setup))
			if _, _, err := interpreter.EvaluateModule(setupModule); err != nil {
				t.Fatalf("setup module %s failed: %v", setup, err)
			}
		}
	}

	result, _, err := interpreter.EvaluateModule(module)
	diags := interpreter.TypecheckDiagnostics()
	if len(manifest.Expect.Errors) > 0 {
		if err == nil {
			t.Fatalf("expected evaluation error")
		}
		msg := extractErrorMessage(err)
		if !contains(manifest.Expect.Errors, msg) {
			t.Fatalf("expected error in %v, got %s", manifest.Expect.Errors, msg)
		}
		checkFixtureTypecheckDiagnostics(underlying, mode, manifest.Expect.TypecheckDiagnostics, diags)
		return
	}
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}
	checkFixtureTypecheckDiagnostics(underlying, mode, manifest.Expect.TypecheckDiagnostics, diags)
	assertResult(underlying, dir, manifest, result, stdout)
}

// testingT captures the subset of testing.T used by fixture helpers.
type testingT interface {
	Helper()
	Fatalf(format string, args ...interface{})
}
