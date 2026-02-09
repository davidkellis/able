package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestAnalyzeProgramBuildsGraphAndTypecheck(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "foo"), 0o755); err != nil {
		t.Fatalf("mkdir foo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "foo", "bar.able"), []byte("fn bar() -> i32 { 1 }\n"), 0o644); err != nil {
		t.Fatalf("write bar.able: %v", err)
	}
	entryPath := filepath.Join(root, "main.able")
	source := "import demo.foo.{bar}\n\nfn main() { bar() }\n"
	if err := os.WriteFile(entryPath, []byte(source), 0o644); err != nil {
		t.Fatalf("write main.able: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	analysis, err := AnalyzeProgram(program)
	if err != nil {
		t.Fatalf("analyze program: %v", err)
	}
	if analysis.Program == nil || analysis.Program.Entry == nil {
		t.Fatalf("expected program and entry module")
	}
	if analysis.Graph.Modules["demo"] == nil {
		t.Fatalf("missing demo module in graph")
	}
	if analysis.Graph.Modules["demo.foo"] == nil {
		t.Fatalf("missing demo.foo module in graph")
	}
	deps := analysis.Graph.StaticEdges["demo"]
	if len(deps) != 1 || deps[0] != "demo.foo" {
		t.Fatalf("unexpected static deps for demo: %v", deps)
	}
	if len(analysis.Typecheck.Packages) == 0 {
		t.Fatalf("expected typecheck package summaries")
	}
	if _, ok := analysis.Typecheck.Packages["demo"]; !ok {
		t.Fatalf("expected typecheck summary for demo")
	}
	if _, ok := analysis.Typecheck.Packages["demo.foo"]; !ok {
		t.Fatalf("expected typecheck summary for demo.foo")
	}
	if analysis.Typecheck.Inferred == nil {
		t.Fatalf("expected inferred types map")
	}
	if _, ok := analysis.Typecheck.Inferred["demo"]; !ok {
		t.Fatalf("expected inferred types for demo")
	}
}
