package driver

import (
	"path/filepath"
	"testing"
)

func TestWriteAndLoadLockfile(t *testing.T) {
	lock := &Lockfile{
		Root:      "app-kit",
		Tool:      "able-cli 0.0.0-dev",
		Generated: "2025-01-01T00:00:00Z",
		Packages: []*LockedPackage{
			{
				Name:     "util-strings",
				Version:  " 2.0.0 ",
				Source:   " registry://core ",
				Checksum: " SHA256:abc ",
				Dependencies: []LockedDependency{
					{Name: "core-lib", Version: " ~> 1.0 "},
					{Name: "core-lib", Version: " ~> 1.1 "},
				},
			},
			{
				Name:    "core-lib",
				Version: "1.2.3",
				Source:  "registry://core",
			},
		},
	}

	path := filepath.Join(t.TempDir(), "package.lock")
	if err := WriteLockfile(lock, path); err != nil {
		t.Fatalf("WriteLockfile error: %v", err)
	}

	loaded, err := LoadLockfile(path)
	if err != nil {
		t.Fatalf("LoadLockfile error: %v", err)
	}

	if loaded.Root != "app_kit" {
		t.Fatalf("Root = %q, want app_kit", loaded.Root)
	}
	if loaded.Tool != "able-cli 0.0.0-dev" {
		t.Fatalf("Tool = %q", loaded.Tool)
	}
	if len(loaded.Packages) != 2 {
		t.Fatalf("Packages length = %d, want 2", len(loaded.Packages))
	}
	if loaded.Packages[0].Name != "core_lib" {
		t.Fatalf("First package = %q, want core_lib", loaded.Packages[0].Name)
	}
	if loaded.Packages[1].Name != "util_strings" {
		t.Fatalf("Second package = %q, want util_strings", loaded.Packages[1].Name)
	}
	if got := loaded.Packages[1].Dependencies[0].Name; got != "core_lib" {
		t.Fatalf("Dependency name = %q, want core_lib", got)
	}
	if got := loaded.Packages[1].Dependencies[0].Version; got != "~> 1.0" {
		t.Fatalf("Dependency version = %q, want ~> 1.0", got)
	}
	if loaded.Path != path {
		t.Fatalf("Path = %q, want %q", loaded.Path, path)
	}
}

func TestLoadLockfileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.lock")
	if _, err := LoadLockfile(path); err == nil {
		t.Fatal("expected error for missing lockfile, got nil")
	}
}
