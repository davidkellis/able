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
	if !strings.Contains(compiledSrc, "const __able_native_static_arrays = true") {
		t.Fatalf("expected native static Array contract constant to be true")
	}
}

func TestNewGeneratorMonoArraysDefaultEnabled(t *testing.T) {
	gen := newGenerator(Options{PackageName: "demo"})
	if !gen.monoArraysEnabled() {
		t.Fatalf("expected bare newGenerator options to inherit mono-array default")
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
	if !strings.Contains(compiledSrc, "const __able_native_static_arrays = true") {
		t.Fatalf("expected native static Array contract constant to be true")
	}
}

func TestCompilerLegacyMonoArrayOptOutCannotDisableNativeStaticArrays(t *testing.T) {
	result := compileNoFallbackSourceWithCompilerOptions(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  values := [1, 2, 3]",
		"  values[1]! as i32",
		"}",
		"",
	}, "\n"), Options{
		PackageName:               "main",
		ExperimentalMonoArrays:    false,
		ExperimentalMonoArraysSet: true,
	})

	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "const __able_native_static_arrays = true") {
		t.Fatalf("legacy opt-out must not disable the native static Array contract")
	}
	if strings.Contains(compiledSrc, "__able_experimental_mono_arrays") {
		t.Fatalf("generated output must no longer describe native static arrays as experimental")
	}
	mainBody, ok := findCompiledFunction(result, "__able_compiled_fn_main")
	if !ok {
		t.Fatalf("could not find compiled main function")
	}
	if !strings.Contains(mainBody, "&__able_array_i32{Elements: []int32{") {
		t.Fatalf("expected native i32 slice wrapper with legacy opt-out set:\n%s", mainBody)
	}
	for _, forbidden := range []string{"runtime.ArrayStore", "runtime.ArrayValue", "Storage_handle"} {
		if strings.Contains(mainBody, forbidden) {
			t.Fatalf("legacy opt-out leaked %q into the static function body:\n%s", forbidden, mainBody)
		}
	}
}
