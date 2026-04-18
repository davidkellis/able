package compiler

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerBoundaryFallbackMarkerForStaticFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler boundary audit in short mode")
	}
	raw := strings.TrimSpace(os.Getenv("ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES"))
	if raw == "" {
		t.Skip("default fixture set is split across TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch{1,2,3,4}")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveBoundaryAuditFixtures(t)
	if len(fixtures) == 0 {
		t.Skip("no boundary-audit fixtures configured")
	}
	runCompilerBoundaryAuditFixtureList(t, root, fixtures)
}

func TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch1(t *testing.T) {
	runCompilerBoundaryAuditFixtureDefaultBatch(t, 0)
}

func TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch2(t *testing.T) {
	runCompilerBoundaryAuditFixtureDefaultBatch(t, 1)
}

func TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch3(t *testing.T) {
	runCompilerBoundaryAuditFixtureDefaultBatch(t, 2)
}

func TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch4(t *testing.T) {
	runCompilerBoundaryAuditFixtureDefaultBatch(t, 3)
}

func TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch5(t *testing.T) {
	runCompilerBoundaryAuditFixtureDefaultBatch(t, 4)
}

func TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch6(t *testing.T) {
	runCompilerBoundaryAuditFixtureDefaultBatch(t, 5)
}

func TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch7(t *testing.T) {
	runCompilerBoundaryAuditFixtureDefaultBatch(t, 6)
}

func TestCompilerBoundaryFallbackMarkerForStaticFixturesBatch8(t *testing.T) {
	runCompilerBoundaryAuditFixtureDefaultBatch(t, 7)
}

func runCompilerBoundaryAuditFixtureDefaultBatch(t *testing.T, batch int) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping compiler boundary audit in short mode")
	}
	if strings.TrimSpace(os.Getenv("ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES")) != "" {
		t.Skip("explicit boundary-audit fixture selection requested")
	}
	batches := defaultCompilerBoundaryAuditFixtureBatches()
	if batch < 0 || batch >= len(batches) {
		t.Fatalf("invalid boundary-audit batch %d", batch)
	}
	fixtures := batches[batch]
	if len(fixtures) == 0 {
		t.Skip("no boundary-audit fixtures configured for batch")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	runCompilerBoundaryAuditFixtureList(t, root, fixtures)
}

func runCompilerBoundaryAuditFixtureList(t *testing.T, root string, fixtures []string) {
	t.Helper()
	for _, rel := range fixtures {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			t.Parallel()
			runCompilerBoundaryAuditFixture(t, root, rel)
		})
	}
}

func defaultCompilerBoundaryAuditFixtureBatches() [][]string {
	return [][]string{
		{"12_08_blocking_io_concurrency"},
		{"13_06_stdlib_package_resolution", "06_01_compiler_struct_positional"},
		{"14_02_regex_core_match_streaming", "10_17_interface_overload_dispatch"},
		{"14_01_language_interfaces_index_apply_iterable"},
		{"06_12_04_stdlib_numbers_bigint", "06_12_20_stdlib_math_core_numeric"},
		{"06_12_10_stdlib_collections_list_vector", "06_12_21_stdlib_fs_path"},
		{"06_12_18_stdlib_collections_array_range"},
		{"06_12_19_stdlib_concurrency_channel_mutex_queue", "06_12_26_stdlib_test_harness_reporters"},
	}
}

func resolveBoundaryAuditFixtures(t *testing.T) []string {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv("ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES"))
	fixtures := []string(nil)
	if raw == "" {
		fixtures = defaultCompilerHeavyAuditFixtures()
		return applyCompilerFixtureBatch(t, fixtures, compilerBoundaryAuditBatchIndexEnv, compilerBoundaryAuditBatchCountEnv)
	}
	if strings.EqualFold(raw, "all") {
		root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
		if _, err := os.Stat(root); os.IsNotExist(err) {
			root = filepath.Join("..", "..", "fixtures", "exec")
		}
		fixtures = collectExecFixtures(t, root)
		return applyCompilerFixtureBatch(t, fixtures, compilerBoundaryAuditBatchIndexEnv, compilerBoundaryAuditBatchCountEnv)
	}
	seen := map[string]struct{}{}
	fixtures = []string{}
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
	return applyCompilerFixtureBatch(t, fixtures, compilerBoundaryAuditBatchIndexEnv, compilerBoundaryAuditBatchCountEnv)
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

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-boundary-fixture")

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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath)
	cmd.Env = applyFixtureEnv(withEnv(os.Environ(), "ABLE_COMPILER_BOUNDARY_MARKER", "1"), manifest.Env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("compiled fixture timed out after 60s")
	}
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
