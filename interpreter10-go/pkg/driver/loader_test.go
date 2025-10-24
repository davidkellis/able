package driver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoaderIncludesSearchPathPackages(t *testing.T) {
	root := t.TempDir()

	// Dependency project with its own manifest.
	depRoot := filepath.Join(root, "dep")
	if err := os.MkdirAll(filepath.Join(depRoot, "src"), 0o755); err != nil {
		t.Fatalf("mkdir dep src: %v", err)
	}
	writeFile(t, filepath.Join(depRoot, "package.yml"), `
name: dep
version: 0.1.0
`)
writeFile(t, filepath.Join(depRoot, "src", "core.able"), `
package core

fn value() -> string {
  "dep"
}
`)

	// Entry project importing the dependency.
	appRoot := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(appRoot, "src"), 0o755); err != nil {
		t.Fatalf("mkdir app src: %v", err)
	}
	writeFile(t, filepath.Join(appRoot, "package.yml"), `
name: app
version: 0.1.0
`)
mainPath := filepath.Join(appRoot, "src", "main.able")
writeFile(t, mainPath, `
package main

import dep.core

fn main() -> void {
  core.value()
}
`)

	depSrc := filepath.Join(depRoot, "src")
	loader, err := NewLoader([]string{depSrc})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(mainPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if program == nil || program.Entry == nil {
		t.Fatalf("program entry missing: %#v", program)
	}
	if program.Entry.Package != "app.src.main" {
		t.Fatalf("entry package = %q, want app.src.main", program.Entry.Package)
	}

	foundDep := false
	for _, mod := range program.Modules {
		if mod != nil && mod.Package == "dep.core" {
			foundDep = true
			break
		}
	}
	if !foundDep {
		t.Fatalf("expected dep.core module to be loaded; modules: %#v", program.Modules)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
