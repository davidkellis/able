package driver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestBasic(t *testing.T) {
	path := writeManifest(t, `
name: able-cli
version: "0.1.0"
license: MIT
authors:
  - David
  - Ada
targets:
  app: src/main.able
dependencies:
  stdlib: "~> 1.0.0"
  logging:
    version: "~> 2.0"
    features: ["core", "ansi"]
dev_dependencies:
  testkit:
    path: ../testkit
build_dependencies:
  builder:
    git: https://github.com/example/builder.git
    rev: abc123
`)

	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest returned error: %v", err)
	}

	if got, want := manifest.Name, "able_cli"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
	if got := manifest.Version; got != "0.1.0" {
		t.Fatalf("Version = %q, want 0.1.0", got)
	}
	if len(manifest.Authors) != 2 || manifest.Authors[0] != "David" || manifest.Authors[1] != "Ada" {
		t.Fatalf("Authors unexpected: %#v", manifest.Authors)
	}

	target, ok := manifest.Targets["app"]
	if !ok {
		t.Fatalf("Targets missing app entry: %#v", manifest.Targets)
	}
	if target.Main != "src/main.able" {
		t.Fatalf("target.Main = %q, want src/main.able", target.Main)
	}

	stdlib := manifest.Dependencies["stdlib"]
	if stdlib == nil || stdlib.Version != "~> 1.0.0" {
		t.Fatalf("stdlib dependency not parsed: %#v", stdlib)
	}

	logging := manifest.Dependencies["logging"]
	if logging == nil {
		t.Fatal("logging dependency missing")
	}
	if got := strings.Join(logging.Features, ","); got != "ansi,core" {
		t.Fatalf("logging features sorted/dedup failed, got %q", got)
	}

	testkit := manifest.DevDependencies["testkit"]
	if testkit == nil || testkit.Path != "../testkit" {
		t.Fatalf("dev dependency path override missing: %#v", testkit)
	}

	builder := manifest.BuildDependencies["builder"]
	if builder == nil || builder.Git == "" || builder.Rev != "abc123" {
		t.Fatalf("build dependency not captured: %#v", builder)
	}

	if got := strings.Join(manifest.TargetOrder, ","); got != "app" {
		t.Fatalf("TargetOrder unexpected: %s", got)
	}
}

func TestLoadManifestDependencyShorthand(t *testing.T) {
	path := writeManifest(t, `
name: lib
dependencies:
  stdlib: "~> 1.2.3"
  utils:
    git: https://example.com/utils.git
    tag: v1.0.0
  local:
    path: ../local
`)
	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest returned error: %v", err)
	}

	if manifest.Dependencies["stdlib"].Version != "~> 1.2.3" {
		t.Fatalf("stdlib version mismatch: %#v", manifest.Dependencies["stdlib"])
	}
	if manifest.Dependencies["utils"].Git == "" || manifest.Dependencies["utils"].Tag != "v1.0.0" {
		t.Fatalf("git dependency not parsed: %#v", manifest.Dependencies["utils"])
	}
	if manifest.Dependencies["local"].Path != "../local" {
		t.Fatalf("path dependency missing: %#v", manifest.Dependencies["local"])
	}
}

func TestLoadManifestValidation(t *testing.T) {
	path := writeManifest(t, `
name: ""
targets:
  cli: src/main.able
dependencies:
  util: {}
`)

	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	msg := err.Error()
	wantFragments := []string{
		"name must be provided",
		"dependencies.util: must specify version, git, or path",
	}
	for _, fragment := range wantFragments {
		if !strings.Contains(msg, fragment) {
			t.Fatalf("validation error missing fragment %q: %s", fragment, msg)
		}
	}
}

func TestLoadManifestTargetEntrypointRequired(t *testing.T) {
	path := writeManifest(t, `
name: demo
targets:
  cli: ""
`)

	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("expected error for empty target entrypoint, got nil")
	}
	if !strings.Contains(err.Error(), `target "cli" requires an entrypoint path`) {
		t.Fatalf("expected entrypoint error, got %v", err)
	}
}

func TestManifestDefaultTarget(t *testing.T) {
	path := writeManifest(t, `
name: demo
targets:
  app-server: src/app.able
  lint: spec/lint.able
  Worker: src/worker.able
`)

	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}

	target, err := manifest.DefaultTarget()
	if err != nil {
		t.Fatalf("DefaultTarget returned error: %v", err)
	}
	if target.OriginalName != "app-server" {
		t.Fatalf("DefaultTarget = %q, want app-server", target.OriginalName)
	}
	if target.Main != "src/app.able" {
		t.Fatalf("Default target main mismatch: %s", target.Main)
	}

	wantOrder := []string{"app_server", "lint", "Worker"}
	if got := manifest.TargetOrder; len(got) != len(wantOrder) {
		t.Fatalf("TargetOrder length = %d, want %d (%v)", len(got), len(wantOrder), wantOrder)
	} else {
		for i := range wantOrder {
			if got[i] != wantOrder[i] {
				t.Fatalf("TargetOrder[%d] = %q, want %q", i, got[i], wantOrder[i])
			}
		}
	}
}

func TestManifestFindTarget(t *testing.T) {
	path := writeManifest(t, `
name: demo
targets:
  app-server: src/app.able
  helper: src/helper.able
`)

	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}

	if target, ok := manifest.FindTarget("app-server"); !ok || target == nil || target.OriginalName != "app-server" {
		t.Fatalf("FindTarget app-server failed: %#v", target)
	}
	if target, ok := manifest.FindTarget("app_server"); !ok || target == nil || target.OriginalName != "app-server" {
		t.Fatalf("FindTarget sanitized app_server failed: %#v", target)
	}
	if target, ok := manifest.FindTarget("APP-SERVER"); !ok || target == nil || target.OriginalName != "app-server" {
		t.Fatalf("FindTarget case-insensitive lookup failed: %#v", target)
	}
	if target, ok := manifest.FindTarget("missing"); ok || target != nil {
		t.Fatalf("FindTarget missing should be nil, got %#v", target)
	}
}

func writeManifest(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "package.yml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}
