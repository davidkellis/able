package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func captureCLI(t *testing.T, args []string) (int, string, string) {
	t.Helper()

	stdout := os.Stdout
	stderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	code := run(args)

	if err := wOut.Close(); err != nil {
		t.Fatalf("stdout close: %v", err)
	}
	if err := wErr.Close(); err != nil {
		t.Fatalf("stderr close: %v", err)
	}

	os.Stdout = stdout
	os.Stderr = stderr

	outBytes, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("stdout read: %v", err)
	}
	errBytes, err := io.ReadAll(rErr)
	if err != nil {
		t.Fatalf("stderr read: %v", err)
	}

	if err := rOut.Close(); err != nil {
		t.Fatalf("stdout pipe close: %v", err)
	}
	if err := rErr.Close(); err != nil {
		t.Fatalf("stderr pipe close: %v", err)
	}

	return code, string(outBytes), string(errBytes)
}

func TestAblecBuildEmitsGoMod(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	buildRoot := t.TempDir()

	outDir := filepath.Join(buildRoot, "out")
	binPath := filepath.Join(buildRoot, "compiled-bin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	projectDir := t.TempDir()
	entryPath := filepath.Join(projectDir, "main.able")
	writeFile(t, entryPath, `
fn main() -> void {
}
`)

	code, _, stderr := captureCLI(t, []string{"-build", "-o", outDir, "-bin", binPath, entryPath})
	if code != 0 {
		t.Fatalf("ablec build returned exit code %d, stderr: %q", code, stderr)
	}

	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected binary at %s: %v", binPath, err)
	}

	goModPath := filepath.Join(outDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	content := string(data)

	modulePath, err := readGoModulePath(moduleRoot)
	if err != nil {
		t.Fatalf("read module path: %v", err)
	}

	if !strings.Contains(content, "module able/compiled") {
		t.Fatalf("go.mod missing module header: %q", content)
	}
	if !strings.Contains(content, "require "+modulePath+" v0.0.0") {
		t.Fatalf("go.mod missing require for %s: %q", modulePath, content)
	}
	if !strings.Contains(content, "replace "+modulePath+" => ./v12/interpreters/go") {
		t.Fatalf("go.mod missing replace for %s: %q", modulePath, content)
	}

	moduleCopy := filepath.Join(outDir, "v12", "interpreters", "go", "go.mod")
	if _, err := os.Stat(moduleCopy); err != nil {
		t.Fatalf("expected interpreter module copy at %s: %v", moduleCopy, err)
	}

	parserCopy := filepath.Join(outDir, "v12", "parser", "tree-sitter-able", "src", "parser.c")
	if _, err := os.Stat(parserCopy); err != nil {
		t.Fatalf("expected parser sources at %s: %v", parserCopy, err)
	}
}

func TestResolveAblecExperimentalMonoArraysDefaultEnabled(t *testing.T) {
	prev, hadPrev := os.LookupEnv("ABLE_EXPERIMENTAL_MONO_ARRAYS")
	if err := os.Unsetenv("ABLE_EXPERIMENTAL_MONO_ARRAYS"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	defer func() {
		if hadPrev {
			_ = os.Setenv("ABLE_EXPERIMENTAL_MONO_ARRAYS", prev)
		} else {
			_ = os.Unsetenv("ABLE_EXPERIMENTAL_MONO_ARRAYS")
		}
	}()

	enabled, err := resolveAblecExperimentalMonoArraysFromEnv()
	if err != nil {
		t.Fatalf("resolve env: %v", err)
	}
	if !enabled {
		t.Fatalf("expected mono arrays enabled by default")
	}
}

func TestResolveAblecExperimentalMonoArraysDisabledFromEnv(t *testing.T) {
	t.Setenv("ABLE_EXPERIMENTAL_MONO_ARRAYS", "false")

	enabled, err := resolveAblecExperimentalMonoArraysFromEnv()
	if err != nil {
		t.Fatalf("resolve env: %v", err)
	}
	if enabled {
		t.Fatalf("expected mono arrays disabled from env")
	}
}

func TestResolveAblecExperimentalMonoArraysInvalidEnv(t *testing.T) {
	t.Setenv("ABLE_EXPERIMENTAL_MONO_ARRAYS", "maybe")

	if _, err := resolveAblecExperimentalMonoArraysFromEnv(); err == nil {
		t.Fatalf("expected invalid env value error")
	}
}

func TestAblecMonoArraysDefaultEnabledInGeneratedOutput(t *testing.T) {
	projectDir := t.TempDir()
	entryPath := filepath.Join(projectDir, "main.able")
	writeFile(t, entryPath, `
fn main() -> void {}
`)
	outDir := filepath.Join(projectDir, "out")

	code, _, stderr := captureCLI(t, []string{"-o", outDir, entryPath})
	if code != 0 {
		t.Fatalf("ablec compile returned exit code %d, stderr: %q", code, stderr)
	}

	compiledPath := filepath.Join(outDir, "compiled.go")
	data, err := os.ReadFile(compiledPath)
	if err != nil {
		t.Fatalf("read compiled.go: %v", err)
	}
	if !strings.Contains(string(data), "const __able_experimental_mono_arrays = true") {
		t.Fatalf("expected mono arrays constant enabled by default")
	}
}

func TestAblecNoExperimentalMonoArraysFlagDisablesGeneratedOutput(t *testing.T) {
	projectDir := t.TempDir()
	entryPath := filepath.Join(projectDir, "main.able")
	writeFile(t, entryPath, `
fn main() -> void {}
`)
	outDir := filepath.Join(projectDir, "out")

	code, _, stderr := captureCLI(t, []string{"-o", outDir, "--no-experimental-mono-arrays", entryPath})
	if code != 0 {
		t.Fatalf("ablec compile returned exit code %d, stderr: %q", code, stderr)
	}

	compiledPath := filepath.Join(outDir, "compiled.go")
	data, err := os.ReadFile(compiledPath)
	if err != nil {
		t.Fatalf("read compiled.go: %v", err)
	}
	if !strings.Contains(string(data), "const __able_experimental_mono_arrays = false") {
		t.Fatalf("expected mono arrays constant disabled via --no-experimental-mono-arrays")
	}
}

func TestCollectStdlibPathsPrefersCachedAbleHome(t *testing.T) {
	cacheHome := filepath.Join(t.TempDir(), ".able")
	cacheSrc := filepath.Join(cacheHome, "pkg", "src", "able", "0.1.0", "src")
	repoRoot := t.TempDir()
	base := filepath.Join(repoRoot, "workspace")
	siblingSrc := filepath.Join(repoRoot, "able-stdlib", "src")

	if err := os.MkdirAll(cacheSrc, 0o755); err != nil {
		t.Fatalf("mkdir cache src: %v", err)
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	if err := os.MkdirAll(siblingSrc, 0o755); err != nil {
		t.Fatalf("mkdir sibling src: %v", err)
	}

	t.Setenv("ABLE_STDLIB_ROOT", "")
	t.Setenv("ABLE_HOME", cacheHome)
	t.Setenv("ABLE_PATH", "")
	t.Setenv("ABLE_MODULE_PATHS", "")

	paths := collectStdlibPaths(base)
	if len(paths) == 0 || paths[0] != cacheSrc {
		t.Fatalf("collectStdlibPaths() = %v, want first entry %q", paths, cacheSrc)
	}
}
