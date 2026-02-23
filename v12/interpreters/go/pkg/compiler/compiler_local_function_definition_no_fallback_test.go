package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func compileNoFallbackSource(t *testing.T, source string) *Result {
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

	result, err := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: true,
	}).Compile(program)
	if err != nil {
		t.Fatalf("compile with no fallbacks: %v", err)
	}
	if len(result.Fallbacks) != 0 {
		t.Fatalf("expected no fallbacks, got %v", result.Fallbacks)
	}
	return result
}

func TestCompilerNoFallbacksForLocalFunctionDefinitionStatement(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  fn fact(n: i32) -> i32 {",
		"    if n <= 1 {",
		"      1",
		"    } else {",
		"      n * fact(n - 1)",
		"    }",
		"  }",
		"  fact(5)",
		"}",
		"",
	}, "\n"))
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "var fact runtime.Value = runtime.NativeFunctionValue") {
		t.Fatalf("expected local function definition to compile into a runtime function binding")
	}
	if strings.Contains(compiledSrc, "CallOriginal(\"demo.main\"") {
		t.Fatalf("expected main to stay compiled without call_original fallback")
	}
}

func TestCompilerNoFallbacksForLocalFunctionDefinitionShadowingTypedBinding(t *testing.T) {
	result := compileNoFallbackSource(t, strings.Join([]string{
		"package demo",
		"",
		"fn main() -> i32 {",
		"  value := 41",
		"  fn value(x: i32) -> i32 {",
		"    x + 1",
		"  }",
		"  value(41)",
		"}",
		"",
	}, "\n"))
	compiledSrc := string(result.Files["compiled.go"])
	if !strings.Contains(compiledSrc, "runtime.NativeFunctionValue") {
		t.Fatalf("expected local shadowing function definition to compile into a runtime callable")
	}
	if strings.Contains(compiledSrc, "CallOriginal(\"demo.main\"") {
		t.Fatalf("expected shadowing local function path to stay compiled without call_original fallback")
	}
}
