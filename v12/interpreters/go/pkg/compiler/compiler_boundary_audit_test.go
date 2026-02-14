package compiler

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerBoundaryFallbackMarkerForStaticFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler boundary audit in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveBoundaryAuditFixtures(t)
	if len(fixtures) == 0 {
		t.Skip("no boundary-audit fixtures configured")
	}
	for _, rel := range fixtures {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			runCompilerBoundaryAuditFixture(t, root, rel)
		})
	}
}

func resolveBoundaryAuditFixtures(t *testing.T) []string {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv("ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES"))
	if raw == "" {
		return defaultCompilerHeavyAuditFixtures()
	}
	if strings.EqualFold(raw, "all") {
		root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
		if _, err := os.Stat(root); os.IsNotExist(err) {
			root = filepath.Join("..", "..", "fixtures", "exec")
		}
		return collectExecFixtures(t, root)
	}
	seen := map[string]struct{}{}
	fixtures := []string{}
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

func runCompilerBoundaryAuditFixture(t *testing.T, root, rel string) {
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
	if manifest.Expect.Exit != nil && *manifest.Expect.Exit != 0 {
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
	workDir, err := os.MkdirTemp(tmpRoot, "ablec-boundary-fixture-")
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
	cmd.Env = applyFixtureEnv(withEnv(os.Environ(), "ABLE_COMPILER_BOUNDARY_MARKER", "1"), manifest.Env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runErr != nil {
		exitCode := 0
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		t.Fatalf("run failed: exit=%d err=%v stderr=%q stdout=%q", exitCode, runErr, stderr.String(), stdout.String())
	}

	rawCount := ""
	for _, line := range strings.Split(stderr.String(), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "__ABLE_BOUNDARY_FALLBACK_CALLS=") {
			rawCount = strings.TrimPrefix(line, "__ABLE_BOUNDARY_FALLBACK_CALLS=")
			break
		}
	}
	if rawCount == "" {
		t.Fatalf("boundary marker missing; stderr=%q stdout=%q", stderr.String(), stdout.String())
	}
	count, err := strconv.ParseInt(rawCount, 10, 64)
	if err != nil {
		t.Fatalf("invalid boundary marker %q: %v", rawCount, err)
	}
	if count != 0 {
		t.Fatalf("unexpected compiled->interpreter fallback calls: count=%d stderr=%q stdout=%q", count, stderr.String(), stdout.String())
	}
}
