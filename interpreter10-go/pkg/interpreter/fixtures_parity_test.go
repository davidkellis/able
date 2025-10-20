package interpreter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureParityStringLiteral(t *testing.T) {
	root := filepath.Join("..", "..", "..", "fixtures", "ast")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("reading fixtures: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fixtureDir := filepath.Join(root, entry.Name())
		walkFixtures(t, fixtureDir, func(dir string) {
			interp := New()
			mode := configureFixtureTypechecker(interp)
			var stdout []string
			registerPrint(interp, &stdout)
			manifest := readManifest(t, dir)
			entryFile := manifest.Entry
			if entryFile == "" {
				entryFile = "module.json"
			}
			modulePath := filepath.Join(dir, entryFile)
			module := readModule(t, modulePath)
			if len(manifest.Setup) > 0 {
				for _, setupFile := range manifest.Setup {
					setupPath := filepath.Join(dir, setupFile)
					setupModule := readModule(t, setupPath)
					if _, _, err := interp.EvaluateModule(setupModule); err != nil {
						t.Fatalf("fixture %s setup module %s failed: %v", dir, setupFile, err)
					}
				}
			}
			result, _, err := interp.EvaluateModule(module)
			if len(manifest.Expect.Errors) > 0 {
				if err == nil {
					t.Fatalf("fixture %s expected evaluation error", dir)
				}
				msg := extractErrorMessage(err)
				if !contains(manifest.Expect.Errors, msg) {
					t.Fatalf("fixture %s expected error message in %v, got %s", dir, manifest.Expect.Errors, msg)
				}
				checkFixtureTypecheckDiagnostics(t, mode, manifest.Expect.TypecheckDiagnostics, interp.TypecheckDiagnostics())
				return
			}
			if err != nil {
				t.Fatalf("fixture %s evaluation error: %v", dir, err)
			}
			checkFixtureTypecheckDiagnostics(t, mode, manifest.Expect.TypecheckDiagnostics, interp.TypecheckDiagnostics())
			assertResult(t, dir, manifest, result, stdout)
		})
	}
}

func walkFixtures(t *testing.T, dir string, fn func(string)) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	hasModule := false
	for _, entry := range entries {
		if entry.Type().IsRegular() && entry.Name() == "module.json" {
			hasModule = true
		}
	}
	if hasModule {
		fn(dir)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			walkFixtures(t, filepath.Join(dir, entry.Name()), fn)
		}
	}
}
