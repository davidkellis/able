package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildTargetFromManifest(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "package.yml"), `
name: demo
targets:
  app: main.able
`)
	writeFile(t, filepath.Join(projectDir, "main.able"), `
extern go fn __able_os_exit(code: i32) -> void {}

fn main() -> void {
  __able_os_exit(0)
}
`)

	enterWorkingDir(t, projectDir)

	_ = runCLIExpectSuccess(t, "build", "app")

	outDir := filepath.Join(projectDir, "target", "compiled", "app")
	binPath := filepath.Join(outDir, "app")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected binary at %s: %v", binPath, err)
	}
}

func TestBuildOutputOutsideModuleRoot(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	buildRoot := t.TempDir()
	if isWithinDir(buildRoot, moduleRoot) {
		t.Skip("temp dir is within module root; unable to validate external output")
	}

	outDir := filepath.Join(buildRoot, "out")
	binPath := filepath.Join(buildRoot, "compiled-bin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	projectDir := t.TempDir()
	entryPath := filepath.Join(projectDir, "main.able")
	writeFile(t, entryPath, `
extern go fn __able_os_exit(code: i32) -> void {}

fn main() -> void {
  __able_os_exit(0)
}
`)

	enterTempWorkingDir(t)
	_ = runCLIExpectSuccess(t, "build", "--out", outDir, "--bin", binPath, entryPath)
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
	// Stdlib source copying is opportunistic and only happens when a cached
	// canonical stdlib is available to the build helper.
	if _, err := ensureCachedStdlib(); err == nil {
		stdlibCopyDir := filepath.Join(outDir, "v12", "stdlib", "src")
		if info, err := os.Stat(stdlibCopyDir); err != nil || !info.IsDir() {
			t.Fatalf("expected stdlib sources directory at %s: %v", stdlibCopyDir, err)
		}
	}
	kernelCopy := filepath.Join(outDir, "v12", "kernel", "src", "kernel.able")
	if _, err := os.Stat(kernelCopy); err != nil {
		t.Fatalf("expected kernel sources at %s: %v", kernelCopy, err)
	}

	relocated := filepath.Join(buildRoot, "relocated")
	if err := os.Rename(outDir, relocated); err != nil {
		t.Fatalf("relocate output: %v", err)
	}
	build := exec.Command("go", "build", ".")
	build.Dir = relocated
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed in relocated output: %v\n%s", err, string(output))
	}
}

func TestBuildNoFallbacksFlagFailsWhenFallbackRequired(t *testing.T) {
	enterTempWorkingDir(t)
	entryPath := writeFallbackBuildEntry(t)
	_, _, stderr := runCLIExpectFailure(t, "build", "--no-fallbacks", entryPath)
	assertTextContainsAll(t, stderr, "fallback not allowed")
}

func TestParseBuildArgumentsDefaultsToStaticNoFallbacks(t *testing.T) {
	if err := os.Unsetenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	if err := os.Unsetenv("ABLE_EXPERIMENTAL_MONO_ARRAYS"); err != nil {
		t.Fatalf("unset mono env: %v", err)
	}

	config, remaining, err := parseBuildArguments([]string{"main.able"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if len(remaining) != 1 || remaining[0] != "main.able" {
		t.Fatalf("unexpected remaining args: %#v", remaining)
	}
	if config.RequireNoFallbacks {
		t.Fatalf("expected strict no-fallback mode disabled by default")
	}
	if !config.RequireNoStaticFallbacks {
		t.Fatalf("expected static fallback guard enabled by default")
	}
	if !config.ExperimentalMonoArrays {
		t.Fatalf("expected mono array experiment enabled by default")
	}
}

func TestParseBuildArgumentsMonoArraysFromEnv(t *testing.T) {
	t.Setenv("ABLE_EXPERIMENTAL_MONO_ARRAYS", "true")
	config, _, err := parseBuildArguments([]string{"main.able"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if !config.ExperimentalMonoArrays {
		t.Fatalf("expected mono array experiment enabled from env")
	}
}

func TestParseBuildArgumentsMonoArraysDisabledFromEnv(t *testing.T) {
	t.Setenv("ABLE_EXPERIMENTAL_MONO_ARRAYS", "false")
	config, _, err := parseBuildArguments([]string{"main.able"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if config.ExperimentalMonoArrays {
		t.Fatalf("expected mono array experiment disabled from env")
	}
}

func TestParseBuildArgumentsMonoArraysFlagOverridesEnv(t *testing.T) {
	t.Setenv("ABLE_EXPERIMENTAL_MONO_ARRAYS", "true")
	config, _, err := parseBuildArguments([]string{"--no-experimental-mono-arrays", "main.able"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if config.ExperimentalMonoArrays {
		t.Fatalf("expected --no-experimental-mono-arrays to override env")
	}

	config, _, err = parseBuildArguments([]string{"--experimental-mono-arrays", "main.able"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if !config.ExperimentalMonoArrays {
		t.Fatalf("expected --experimental-mono-arrays to enable flag")
	}
}

func TestParseBuildArgumentsMonoArraysInvalidEnv(t *testing.T) {
	t.Setenv("ABLE_EXPERIMENTAL_MONO_ARRAYS", "maybe")
	if _, _, err := parseBuildArguments([]string{"main.able"}); err == nil {
		t.Fatalf("expected invalid ABLE_EXPERIMENTAL_MONO_ARRAYS to fail parsing")
	} else if !strings.Contains(err.Error(), "invalid ABLE_EXPERIMENTAL_MONO_ARRAYS value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildNoFallbacksEnvFailsWhenFallbackRequired(t *testing.T) {
	enterTempWorkingDir(t)
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")
	entryPath := writeFallbackBuildEntry(t)
	_, _, stderr := runCLIExpectFailure(t, "build", entryPath)
	assertTextContainsAll(t, stderr, "fallback not allowed")
}

func TestBuildNoFallbacksInvalidEnvFailsArgumentParsing(t *testing.T) {
	enterTempWorkingDir(t)
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "sometimes")
	entryPath := writeFallbackBuildEntry(t)
	_, _, stderr := runCLIExpectFailure(t, "build", entryPath)
	assertTextContainsAll(t, stderr, "invalid ABLE_COMPILER_REQUIRE_NO_FALLBACKS value")
}

func TestBuildAllowFallbacksOverridesEnv(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "true")
	entryPath := writeFallbackBuildEntry(t)
	outRoot := t.TempDir()
	outDir := filepath.Join(outRoot, "out")
	binPath := filepath.Join(outRoot, "app")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	enterTempWorkingDir(t)
	_, stderr := runCLIExpectSuccessAllowingStderr(t,
		"build",
		"--allow-fallbacks",
		"--out", outDir,
		"--bin", binPath,
		entryPath,
	)
	assertTextContainsAll(t, stderr, "typechecker:")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected compiled binary at %s: %v", binPath, err)
	}
}

func TestBuildEnvFalseAllowsFallbacks(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "false")
	entryPath := writeFallbackBuildEntry(t)
	outRoot := t.TempDir()
	outDir := filepath.Join(outRoot, "out")
	binPath := filepath.Join(outRoot, "app")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	enterTempWorkingDir(t)
	_, stderr := runCLIExpectSuccessAllowingStderr(t,
		"build",
		"--out", outDir,
		"--bin", binPath,
		entryPath,
	)
	assertTextContainsAll(t, stderr, "typechecker:")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected compiled binary at %s: %v", binPath, err)
	}
}

func TestParseBuildArgumentsEnvFalseDisablesStaticNoFallbacks(t *testing.T) {
	t.Setenv("ABLE_COMPILER_REQUIRE_NO_FALLBACKS", "false")

	config, _, err := parseBuildArguments([]string{"main.able"})
	if err != nil {
		t.Fatalf("parse args: %v", err)
	}
	if config.RequireNoFallbacks {
		t.Fatalf("expected strict no-fallback mode disabled when env=false")
	}
	if config.RequireNoStaticFallbacks {
		t.Fatalf("expected static fallback guard disabled when env=false")
	}
}

func writeFallbackBuildEntry(t *testing.T) string {
	t.Helper()
	projectDir := t.TempDir()
	entryPath := filepath.Join(projectDir, "main.able")
	writeFile(t, entryPath, `
extern go fn __able_os_exit(code: i32) -> void {}

fn needs_fallback(x: i64, y: i64) -> i64 {
  x / y
}

fn main() -> void {
  needs_fallback(4, 2)
  __able_os_exit(0)
}
`)
	return entryPath
}
