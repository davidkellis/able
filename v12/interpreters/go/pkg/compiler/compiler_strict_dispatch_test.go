package compiler

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerStrictDispatchForStdlibHeavyFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping strict dispatch fixture audit in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveStrictDispatchFixtures(t, root)
	if len(fixtures) == 0 {
		t.Skip("no strict-dispatch fixtures configured")
	}
	for _, rel := range fixtures {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			runCompilerStrictDispatchFixture(t, root, rel)
		})
	}
}

func resolveStrictDispatchFixtures(t *testing.T, root string) []string {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv("ABLE_COMPILER_STRICT_DISPATCH_FIXTURES"))
	if raw == "" {
		return defaultCompilerHeavyAuditFixtures()
	}
	if strings.EqualFold(raw, "all") {
		return collectExecFixtures(t, root)
	}
	fixtures := []string{}
	seen := make(map[string]struct{})
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
	}) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		fixtures = append(fixtures, part)
	}
	return fixtures
}

func runCompilerStrictDispatchFixture(t *testing.T, root, rel string) {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(rel))
	manifest, err := interpreter.LoadFixtureManifest(dir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if shouldSkipTarget(manifest.SkipTargets, "go") {
		return
	}
	expectedTypecheck := manifest.Expect.TypecheckDiagnostics
	if expectedTypecheck != nil && len(expectedTypecheck) > 0 {
		return
	}
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(dir, entry)

	searchPaths, err := buildExecSearchPaths(entryPath, dir, manifest)
	if err != nil {
		t.Fatalf("exec search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	program, err := loader.Load(entryPath)
	loader.Close()
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	tmpRoot := filepath.Join(moduleRoot, "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir tmp: %v", err)
	}
	workDir, err := os.MkdirTemp(tmpRoot, "ablec-strict-fixture-")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(workDir) })

	comp := New(Options{
		PackageName:        "main",
		RequireNoFallbacks: requireNoFallbacksForFixtureGates(t),
	})
	result, err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(workDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	harness := compilerHarnessSource(entryPath, searchPaths, manifest.Executor)
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte(harness), 0o600); err != nil {
		t.Fatalf("write harness: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-fixture")
	if goruntime.GOOS == "windows" {
		binPath += ".exe"
	}
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = workDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}

	cmd := exec.Command(binPath)
	cmd.Env = applyFixtureEnv(withEnv(os.Environ(), "ABLE_COMPILER_STRICT_DISPATCH_MARKER", "1"), manifest.Env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("run error: %v", runErr)
		}
	}

	strict := ""
	remainingStderrLines := make([]string, 0)
	for _, line := range splitLines(stderr.String()) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "__ABLE_STRICT=") {
			strict = strings.TrimPrefix(line, "__ABLE_STRICT=")
			continue
		}
		remainingStderrLines = append(remainingStderrLines, line)
	}
	if strict == "" {
		t.Fatalf("strict marker missing; stderr=%q stdout=%q", stderr.String(), stdout.String())
	}
	if strict != "true" {
		t.Fatalf("strict dispatch disabled; marker=%q stderr=%q stdout=%q", strict, stderr.String(), stdout.String())
	}

	actualStdout := splitLines(stdout.String())
	expected := manifest.Expect
	if expected.Stdout != nil {
		expectedStdout := expandFixtureLines(expected.Stdout)
		if !bytes.Equal([]byte(strings.Join(actualStdout, "\n")), []byte(strings.Join(expectedStdout, "\n"))) {
			t.Fatalf("stdout mismatch: expected %v, got %v", expectedStdout, actualStdout)
		}
	}
	if expected.Stderr != nil {
		expectedStderr := expandFixtureLines(expected.Stderr)
		if !bytes.Equal([]byte(strings.Join(remainingStderrLines, "\n")), []byte(strings.Join(expectedStderr, "\n"))) {
			t.Fatalf("stderr mismatch: expected %v, got %v", expectedStderr, remainingStderrLines)
		}
	}
	if expected.Exit != nil {
		if exitCode != *expected.Exit {
			t.Fatalf("exit code mismatch: expected %d, got %d (stderr=%q stdout=%q)", *expected.Exit, exitCode, stderr.String(), stdout.String())
		}
	} else if exitCode != 0 {
		t.Fatalf("exit code mismatch: expected 0, got %d (stderr=%q stdout=%q)", exitCode, stderr.String(), stdout.String())
	}
}
