package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func compileNoFallbackSourceWithCompilerOptions(t *testing.T, source string, opts Options) *Result {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	opts.RequireNoFallbacks = true
	result, err := New(opts).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}
	return result
}

func TestCompilerFeatureFlagMonoArraysDefaultEnabled(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 { 1 }",
		"",
	}, "\n"))
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "const __able_experimental_mono_arrays = true") {
		t.Fatalf("expected mono-array feature flag constant to default to true")
	}
}

func TestCompilerFeatureFlagMonoArraysEnabledViaOptions(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 { 1 }",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		ExperimentalMonoArrays: true,
	})
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "const __able_experimental_mono_arrays = true") {
		t.Fatalf("expected mono-array feature flag constant to be enabled")
	}
}
