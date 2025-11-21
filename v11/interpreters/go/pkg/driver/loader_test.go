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
	loader, err := NewLoader([]SearchPath{{Path: depSrc}})
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

func TestLoaderRejectsPackageCollisions(t *testing.T) {
	root := t.TempDir()
	primary := filepath.Join(root, "primary")
	secondary := filepath.Join(root, "secondary")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("mkdir primary: %v", err)
	}
	if err := os.MkdirAll(secondary, 0o755); err != nil {
		t.Fatalf("mkdir secondary: %v", err)
	}

	writeFile(t, filepath.Join(primary, "package.yml"), "name: collide\n")
	mainPath := filepath.Join(primary, "main.able")
	writeFile(t, mainPath, `
package main

fn main() -> void {}
`)
	writeFile(t, filepath.Join(primary, "shared.able"), `
package shared

fn value() -> string { "primary" }
`)

	writeFile(t, filepath.Join(secondary, "package.yml"), "name: collide\n")
	writeFile(t, filepath.Join(secondary, "shared.able"), `
package shared

fn value() -> string { "secondary" }
`)

	loader, err := NewLoader([]SearchPath{{Path: secondary}})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	if _, err := loader.Load(mainPath); err == nil {
		t.Fatalf("expected collision error, got nil")
	} else if !strings.Contains(err.Error(), "package collide.shared") {
		t.Fatalf("unexpected collision error: %v", err)
	}
}

func TestLoaderRejectsAbleNamespace(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.yml"), "name: able\n")
	entry := filepath.Join(root, "main.able")
	writeFile(t, entry, `
package main

fn main() -> void {}
`)

	loader, err := NewLoader(nil)
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	if _, err := loader.Load(entry); err == nil {
		t.Fatalf("expected namespace validation failure")
	} else if !strings.Contains(err.Error(), "able.*") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoaderDynImportDependencies(t *testing.T) {
	root := t.TempDir()
	depRoot := filepath.Join(root, "dep")
	appRoot := filepath.Join(root, "app")

	if err := os.MkdirAll(depRoot, 0o755); err != nil {
		t.Fatalf("mkdir dep root: %v", err)
	}
	if err := os.MkdirAll(appRoot, 0o755); err != nil {
		t.Fatalf("mkdir app root: %v", err)
	}

	writeFile(t, filepath.Join(depRoot, "package.yml"), "name: extras\n")
	writeFile(t, filepath.Join(depRoot, "tools.able"), `
package tools

fn message() -> string { "dynimport" }
`)

	entryPath := filepath.Join(appRoot, "main.able")
	writeFile(t, entryPath, `
package main

dynimport extras.tools.{message}

fn main() -> void { message() }
`)

	loader, err := NewLoader([]SearchPath{{Path: depRoot}})
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("loader.Load returned error: %v", err)
	}

	found := false
	for _, mod := range program.Modules {
		if mod != nil && mod.Package == "extras.tools" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected dynimport target extras.tools to be loaded; modules: %#v", program.Modules)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
