package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerNoFallbacksStringDefaultImplStaticEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	source := strings.Join([]string{
		"package demo",
		"",
		"interface Default {",
		"  fn default() -> Self",
		"}",
		"",
		"struct String {",
		"  n: i32",
		"}",
		"",
		"methods String {",
		"  fn empty() -> String { String { n: 0 } }",
		"}",
		"",
		"impl Default for String {",
		"  fn default() -> String { String.empty() }",
		"}",
		"",
		"fn main() -> void {}",
		"",
	}, "\n")
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
	if len(result.Files) == 0 {
		t.Fatalf("expected generated output files")
	}
}
