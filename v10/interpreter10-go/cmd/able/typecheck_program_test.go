package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/interpreter"
)

func TestRunProgramTypecheckDetectsCrossPackageMismatch(t *testing.T) {
	root := t.TempDir()
	dep := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir root src: %v", err)
	}
	writeFile(t, filepath.Join(root, "package.yml"), `
name: root_app
`)
	writeFile(t, filepath.Join(root, "src", "main.able"), `
import dep;

fn shout() -> string {
  dep.provide() + "!"
}

fn main() -> void {}
`)

	if err := os.MkdirAll(filepath.Join(dep, "src"), 0o755); err != nil {
		t.Fatalf("mkdir dep src: %v", err)
	}
	writeFile(t, filepath.Join(dep, "package.yml"), `
name: dep
`)
	writeFile(t, filepath.Join(dep, "src", "lib.able"), `
fn provide() -> i32 { 42 }
`)

	loader, err := driver.NewLoader([]string{filepath.Join(dep, "src")})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	entry := filepath.Join(root, "src", "main.able")
	program, err := loader.Load(entry)
	if err != nil {
		t.Fatalf("Load program: %v", err)
	}

	result, err := interpreter.TypecheckProgram(program)
	if err != nil {
		t.Fatalf("runProgramTypecheck returned error: %v", err)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics from cross-package mismatch, got none")
	}
	if !strings.HasPrefix(result.Diagnostics[0].Package, "root_app") {
		t.Fatalf("expected diagnostic for root_app, got %s", result.Diagnostics[0].Package)
	}
	if want := "requires both operands"; !strings.Contains(result.Diagnostics[0].Diagnostic.Message, want) {
		t.Fatalf("expected diagnostic containing %q, got %q", want, result.Diagnostics[0].Diagnostic.Message)
	}
	if summary, ok := result.Packages["dep"]; !ok {
		t.Fatalf("expected summary for dep")
	} else if summary.Functions == nil {
		t.Fatalf("expected functions map in summary")
	}
}

func TestRunProgramTypecheckSucceedsWithImports(t *testing.T) {
	root := t.TempDir()
	dep := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir root src: %v", err)
	}
	writeFile(t, filepath.Join(root, "package.yml"), `
name: root_app
`)
	writeFile(t, filepath.Join(root, "src", "main.able"), `
import dep;

fn shout() -> string {
  dep.provide()
}

fn main() -> void {}
`)

	if err := os.MkdirAll(filepath.Join(dep, "src"), 0o755); err != nil {
		t.Fatalf("mkdir dep src: %v", err)
	}
	writeFile(t, filepath.Join(dep, "package.yml"), `
name: dep
`)
	writeFile(t, filepath.Join(dep, "src", "lib.able"), `
fn provide() -> string { "ok" }
`)

	loader, err := driver.NewLoader([]string{filepath.Join(dep, "src")})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	entry := filepath.Join(root, "src", "main.able")
	program, err := loader.Load(entry)
	if err != nil {
		t.Fatalf("Load program: %v", err)
	}

	result, err := interpreter.TypecheckProgram(program)
	if err != nil {
		t.Fatalf("runProgramTypecheck returned error: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %v", result.Diagnostics)
	}
	if summary, ok := result.Packages["dep"]; !ok || summary.Functions == nil {
		t.Fatalf("expected exported function summary for dep")
	}
}
