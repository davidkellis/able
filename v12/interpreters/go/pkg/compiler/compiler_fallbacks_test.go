package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

const compilerFallbackEnv = "ABLE_COMPILER_FALLBACK_AUDIT"

func TestCompilerExecFixtureFallbacks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler fallback audit in short mode")
	}
	if strings.TrimSpace(os.Getenv(compilerFallbackEnv)) == "" {
		t.Skip("compiler fallback audit disabled")
	}

	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveCompilerFixtures(t, root)
	if len(fixtures) == 0 {
		t.Skip("no compiler fixtures configured")
	}

	type fallbackEntry struct {
		fixture string
		name    string
		reason  string
	}
	var failures []fallbackEntry

	for _, rel := range fixtures {
		dir := filepath.Join(root, filepath.FromSlash(rel))
		manifest, err := interpreter.LoadFixtureManifest(dir)
		if err != nil {
			t.Fatalf("read manifest: %v", err)
		}
		if shouldSkipTarget(manifest.SkipTargets, "go") {
			continue
		}
		expectedTypecheck := manifest.Expect.TypecheckDiagnostics
		if expectedTypecheck != nil && len(expectedTypecheck) > 0 {
			continue
		}
		entry := manifest.Entry
		if entry == "" {
			entry = "main.able"
		}
		entryPath := filepath.Join(dir, entry)

		searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
		if err != nil {
			t.Fatalf("exec search paths: %v", err)
		}
		loader, err := driver.NewLoader(searchPaths)
		if err != nil {
			t.Fatalf("loader init: %v", err)
		}
		program, err := loader.Load(entryPath)
		loader.Close()
		if err != nil {
			t.Fatalf("load program: %v", err)
		}

		comp := New(Options{PackageName: "main"})
		result, err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		for _, fb := range result.Fallbacks {
			failures = append(failures, fallbackEntry{
				fixture: rel,
				name:    fb.Name,
				reason:  fb.Reason,
			})
		}
	}

	if len(failures) > 0 {
		lines := make([]string, 0, len(failures))
		for _, entry := range failures {
			lines = append(lines, fmt.Sprintf("%s: %s (%s)", entry.fixture, entry.name, entry.reason))
		}
		t.Fatalf("compiler fallbacks detected:\n%s", strings.Join(lines, "\n"))
	}
}
