package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/parser"
)

func TestExportFixturesWritesModuleJSON(t *testing.T) {
	root := t.TempDir()
	fixtureDir := filepath.Join(root, "sample")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	sourcePath := filepath.Join(fixtureDir, "source.able")
	writeFixtureSource(t, sourcePath, `
package sample

fn main() {
  print("hello")
}
`)

	var stdout bytes.Buffer
	if err := exportFixtures(root, false, nil, &stdout); err != nil {
		t.Fatalf("exportFixtures: %v", err)
	}
	if !strings.Contains(stdout.String(), "Wrote 1 fixture module(s)") {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}

	got, err := os.ReadFile(filepath.Join(fixtureDir, "module.json"))
	if err != nil {
		t.Fatalf("read module.json: %v", err)
	}
	want := exportedFixtureJSON(t, sourcePath)
	if !bytes.Equal(got, want) {
		t.Fatalf("module.json mismatch\nwant:\n%s\ngot:\n%s", string(want), string(got))
	}
}

func TestExportFixturesCheckPassesForCurrentFixtures(t *testing.T) {
	root := t.TempDir()
	fixtureDir := filepath.Join(root, "sample")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	sourcePath := filepath.Join(fixtureDir, "source.able")
	writeFixtureSource(t, sourcePath, `
package sample

fn main() {
  print("hello")
}
`)

	if err := exportFixtures(root, false, nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("initial exportFixtures: %v", err)
	}

	var stdout bytes.Buffer
	if err := exportFixtures(root, true, nil, &stdout); err != nil {
		t.Fatalf("check exportFixtures: %v", err)
	}
	if !strings.Contains(stdout.String(), "Verified 1 fixture module(s)") {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}
}

func TestExportFixturesCheckRejectsStaleModuleJSON(t *testing.T) {
	root := t.TempDir()
	fixtureDir := filepath.Join(root, "sample")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	writeFixtureSource(t, filepath.Join(fixtureDir, "source.able"), `
package sample

fn main() {
  print("hello")
}
`)
	modulePath := filepath.Join(fixtureDir, "module.json")
	if err := os.WriteFile(modulePath, []byte("{\"stale\":true}\n"), 0o644); err != nil {
		t.Fatalf("write stale module.json: %v", err)
	}

	err := exportFixtures(root, true, nil, &bytes.Buffer{})
	if err == nil {
		t.Fatalf("expected stale fixture failure")
	}
	if !strings.Contains(err.Error(), "stale fixture module(s)") || !strings.Contains(err.Error(), modulePath) {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestExportFixturesWritesOnlyRequestedTargets(t *testing.T) {
	root := t.TempDir()
	fixtureA := filepath.Join(root, "alpha")
	fixtureB := filepath.Join(root, "beta")
	for _, dir := range []string{fixtureA, fixtureB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir fixture: %v", err)
		}
	}
	sourceA := filepath.Join(fixtureA, "source.able")
	sourceB := filepath.Join(fixtureB, "source.able")
	writeFixtureSource(t, sourceA, "package alpha\n")
	writeFixtureSource(t, sourceB, "package beta\n")

	var stdout bytes.Buffer
	if err := exportFixtures(root, false, []string{"alpha"}, &stdout); err != nil {
		t.Fatalf("exportFixtures target: %v", err)
	}
	if !strings.Contains(stdout.String(), "Wrote 1 fixture module(s)") {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(fixtureA, "module.json")); err != nil {
		t.Fatalf("expected alpha module.json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(fixtureB, "module.json")); !os.IsNotExist(err) {
		t.Fatalf("expected beta module.json to remain absent, got %v", err)
	}
}

func TestExportFixturesCheckRequestedSourcePath(t *testing.T) {
	root := t.TempDir()
	fixtureDir := filepath.Join(root, "sample")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	sourcePath := filepath.Join(fixtureDir, "source.able")
	writeFixtureSource(t, sourcePath, "package sample\n")

	if err := exportFixtures(root, false, nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("initial exportFixtures: %v", err)
	}

	var stdout bytes.Buffer
	if err := exportFixtures(root, true, []string{sourcePath}, &stdout); err != nil {
		t.Fatalf("check exportFixtures source target: %v", err)
	}
	if !strings.Contains(stdout.String(), "Verified 1 fixture module(s)") {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}
}

func TestExportFixturesRejectsTargetOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "source.able")
	writeFixtureSource(t, outside, "package outside\n")

	err := exportFixtures(root, false, []string{outside}, &bytes.Buffer{})
	if err == nil {
		t.Fatalf("expected outside-root target failure")
	}
	if !strings.Contains(err.Error(), "not found under") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestRepositoryRootFromFindsGitAncestor(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	start := filepath.Join(root, "v12", "interpreters", "go", "cmd", "fixture-exporter")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatalf("mkdir start: %v", err)
	}
	if got := repositoryRootFrom(start); got != root {
		t.Fatalf("repositoryRootFrom(%q) = %q, want %q", start, got, root)
	}
}

func TestRepositoryRootFromFixtureExporterSuffixFallback(t *testing.T) {
	root := t.TempDir()
	start := filepath.Join(root, "v12", "interpreters", "go", "cmd", "fixture-exporter")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatalf("mkdir start: %v", err)
	}
	if got := repositoryRootFrom(start); got != root {
		t.Fatalf("repositoryRootFrom(%q) = %q, want %q", start, got, root)
	}
}

func exportedFixtureJSON(t *testing.T, sourcePath string) []byte {
	t.Helper()
	module, err := parseModule(sourcePath)
	if err != nil {
		t.Fatalf("parseModule: %v", err)
	}
	parser.NormalizeFixtureModule(module)
	data, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture module: %v", err)
	}
	return append(data, '\n')
}

func writeFixtureSource(t *testing.T, path, source string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(source)+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture source %s: %v", path, err)
	}
}
