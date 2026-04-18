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

const compilerInterfaceLookupFixturesEnv = "ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES"
const compilerInterfaceLookupStrictTotalEnv = "ABLE_COMPILER_INTERFACE_LOOKUP_STRICT_TOTAL"
const compilerGlobalLookupStrictTotalEnv = "ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL"

func TestCompilerInterfaceLookupBypassForStaticFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler interface-lookup audit in short mode")
	}
	raw := strings.TrimSpace(os.Getenv(compilerInterfaceLookupFixturesEnv))
	if raw == "" {
		t.Skip("default fixture set is split across TestCompilerInterfaceLookupBypassForStaticFixturesBatch{1,2,3,4}")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveInterfaceLookupAuditFixtures(t, root)
	if len(fixtures) == 0 {
		t.Skip("no interface-lookup fixtures configured")
	}
	runCompilerInterfaceLookupAuditFixtureList(t, root, fixtures)
}

func TestCompilerInterfaceLookupBypassForStaticFixturesBatch1(t *testing.T) {
	runCompilerInterfaceLookupAuditFixtureDefaultBatch(t, 0)
}

func TestCompilerInterfaceLookupBypassForStaticFixturesBatch2(t *testing.T) {
	runCompilerInterfaceLookupAuditFixtureDefaultBatch(t, 1)
}

func TestCompilerInterfaceLookupBypassForStaticFixturesBatch3(t *testing.T) {
	runCompilerInterfaceLookupAuditFixtureDefaultBatch(t, 2)
}

func TestCompilerInterfaceLookupBypassForStaticFixturesBatch4(t *testing.T) {
	runCompilerInterfaceLookupAuditFixtureDefaultBatch(t, 3)
}

func TestCompilerInterfaceLookupReportersFixtureRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler interface-lookup reporters regression in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	runCompilerInterfaceLookupAuditFixture(t, root, "06_12_26_stdlib_test_harness_reporters")
}

func resolveInterfaceLookupAuditFixtures(t *testing.T, root string) []string {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv(compilerInterfaceLookupFixturesEnv))
	fixtures := []string(nil)
	if raw == "" {
		fixtures = defaultCompilerInterfaceLookupAuditFixtures()
		return applyCompilerFixtureBatch(t, fixtures, compilerInterfaceLookupBatchIndexEnv, compilerInterfaceLookupBatchCountEnv)
	}
	if strings.EqualFold(raw, "all") {
		fixtures = collectExecFixtures(t, root)
		return applyCompilerFixtureBatch(t, fixtures, compilerInterfaceLookupBatchIndexEnv, compilerInterfaceLookupBatchCountEnv)
	}
	fixtures = []string{}
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
	return applyCompilerFixtureBatch(t, fixtures, compilerInterfaceLookupBatchIndexEnv, compilerInterfaceLookupBatchCountEnv)
}

func runCompilerInterfaceLookupAuditFixtureDefaultBatch(t *testing.T, batch int) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping compiler interface-lookup audit in short mode")
	}
	if strings.TrimSpace(os.Getenv(compilerInterfaceLookupFixturesEnv)) != "" {
		t.Skip("explicit interface-lookup fixture selection requested")
	}
	batches := defaultCompilerInterfaceLookupAuditFixtureBatches()
	if batch < 0 || batch >= len(batches) {
		t.Fatalf("invalid interface-lookup audit batch %d", batch)
	}
	fixtures := batches[batch]
	if len(fixtures) == 0 {
		t.Skip("no interface-lookup fixtures configured for batch")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	runCompilerInterfaceLookupAuditFixtureList(t, root, fixtures)
}

func runCompilerInterfaceLookupAuditFixtureList(t *testing.T, root string, fixtures []string) {
	t.Helper()
	for _, rel := range fixtures {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			t.Parallel()
			runCompilerInterfaceLookupAuditFixture(t, root, rel)
		})
	}
}

func defaultCompilerInterfaceLookupAuditFixtures() []string {
	return []string{
		"06_01_compiler_type_qualified_method",
		"06_03_operator_overloading_interfaces",
		"06_10_dynamic_metaprogramming_package_object",
		"06_12_20_stdlib_math_core_numeric",
		"06_12_21_stdlib_fs_path",
		"06_12_22_stdlib_io_temp",
		"06_12_23_stdlib_os",
		"06_12_24_stdlib_process",
		"06_12_25_stdlib_term",
		"06_12_26_stdlib_test_harness_reporters",
		"07_04_apply_callable_interface",
		"10_01_interface_defaults_composites",
		"10_02_impl_specificity_named_overrides",
		"10_02_impl_where_clause",
		"10_03_interface_type_dynamic_dispatch",
		"10_04_interface_dispatch_defaults_generics",
		"10_05_interface_named_impl_defaults",
		"10_06_interface_generic_param_dispatch",
		"10_07_interface_default_chain",
		"10_08_interface_default_override",
		"10_09_interface_named_impl_inherent",
		"10_10_interface_inheritance_defaults",
		"10_11_interface_generic_args_dispatch",
		"10_12_interface_union_target_dispatch",
		"10_13_interface_param_generic_args",
		"10_14_interface_return_generic_args",
		"10_15_interface_default_generic_method",
		"10_16_interface_value_storage",
		"10_17_interface_overload_dispatch",
		"13_04_import_alias_selective_dynimport",
		"14_01_language_interfaces_index_apply_iterable",
		"14_01_operator_interfaces_arithmetic_comparison",
	}
}

func defaultCompilerInterfaceLookupAuditFixtureBatches() [][]string {
	fixtures := defaultCompilerInterfaceLookupAuditFixtures()
	const batchCount = 4
	batches := make([][]string, batchCount)
	for idx, rel := range fixtures {
		batch := idx % batchCount
		batches[batch] = append(batches[batch], rel)
	}
	return batches
}

func runCompilerInterfaceLookupAuditFixture(t *testing.T, root, rel string) {
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

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-interface-lookup-fixture")

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
	cmdEnv := withEnv(os.Environ(), "ABLE_COMPILER_STRICT_DISPATCH_MARKER", "1")
	cmdEnv = withEnv(cmdEnv, "ABLE_COMPILER_INTERFACE_LOOKUP_MARKER", "1")
	cmdEnv = withEnv(cmdEnv, "ABLE_COMPILER_GLOBAL_LOOKUP_MARKER", "1")
	cmdEnv = withEnv(cmdEnv, "ABLE_COMPILER_BOUNDARY_MARKER", "1")
	cmd.Env = applyFixtureEnv(cmdEnv, manifest.Env)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("compiled fixture timed out after 60s")
	}
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("run error: %v", runErr)
		}
	}

	strict := ""
	rawFallbackCount := ""
	rawLookupCalls := ""
	rawInterfaceLookupCalls := ""
	rawGlobalLookupCalls := ""
	rawGlobalLookupEnvCalls := ""
	rawGlobalLookupRegistryCalls := ""
	rawGlobalLookupNames := ""
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
		if strings.HasPrefix(line, "__ABLE_BOUNDARY_FALLBACK_CALLS=") {
			rawFallbackCount = strings.TrimPrefix(line, "__ABLE_BOUNDARY_FALLBACK_CALLS=")
			continue
		}
		if strings.HasPrefix(line, "__ABLE_BOUNDARY_EXPLICIT_CALLS=") {
			continue
		}
		if strings.HasPrefix(line, "__ABLE_BOUNDARY_FALLBACK_NAMES=") {
			continue
		}
		if strings.HasPrefix(line, "__ABLE_BOUNDARY_EXPLICIT_NAMES=") {
			continue
		}
		if strings.HasPrefix(line, "__ABLE_MEMBER_LOOKUP_CALLS=") {
			rawLookupCalls = strings.TrimPrefix(line, "__ABLE_MEMBER_LOOKUP_CALLS=")
			continue
		}
		if strings.HasPrefix(line, "__ABLE_GLOBAL_LOOKUP_FALLBACK_CALLS=") {
			rawGlobalLookupCalls = strings.TrimPrefix(line, "__ABLE_GLOBAL_LOOKUP_FALLBACK_CALLS=")
			continue
		}
		if strings.HasPrefix(line, "__ABLE_GLOBAL_LOOKUP_FALLBACK_ENV_CALLS=") {
			rawGlobalLookupEnvCalls = strings.TrimPrefix(line, "__ABLE_GLOBAL_LOOKUP_FALLBACK_ENV_CALLS=")
			continue
		}
		if strings.HasPrefix(line, "__ABLE_GLOBAL_LOOKUP_FALLBACK_REGISTRY_CALLS=") {
			rawGlobalLookupRegistryCalls = strings.TrimPrefix(line, "__ABLE_GLOBAL_LOOKUP_FALLBACK_REGISTRY_CALLS=")
			continue
		}
		if strings.HasPrefix(line, "__ABLE_GLOBAL_LOOKUP_FALLBACK_NAMES=") {
			rawGlobalLookupNames = strings.TrimPrefix(line, "__ABLE_GLOBAL_LOOKUP_FALLBACK_NAMES=")
			continue
		}
		if strings.HasPrefix(line, "__ABLE_MEMBER_LOOKUP_INTERFACE_CALLS=") {
			rawInterfaceLookupCalls = strings.TrimPrefix(line, "__ABLE_MEMBER_LOOKUP_INTERFACE_CALLS=")
			continue
		}
		if strings.HasPrefix(line, "__ABLE_MEMBER_LOOKUP_NAMES=") {
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
	if rawFallbackCount == "" || rawLookupCalls == "" || rawInterfaceLookupCalls == "" || rawGlobalLookupCalls == "" || rawGlobalLookupEnvCalls == "" || rawGlobalLookupRegistryCalls == "" {
		t.Fatalf("audit markers missing; fallback=%q lookup=%q interface=%q global_lookup=%q global_env=%q global_registry=%q stderr=%q stdout=%q", rawFallbackCount, rawLookupCalls, rawInterfaceLookupCalls, rawGlobalLookupCalls, rawGlobalLookupEnvCalls, rawGlobalLookupRegistryCalls, stderr.String(), stdout.String())
	}
	fallbackCount, err := strconv.ParseInt(rawFallbackCount, 10, 64)
	if err != nil {
		t.Fatalf("invalid fallback marker %q: %v", rawFallbackCount, err)
	}
	if fallbackCount != 0 {
		t.Fatalf("expected no fallback calls for static fixture, got fallbackCalls=%d stderr=%q stdout=%q", fallbackCount, stderr.String(), stdout.String())
	}
	lookupCalls, err := strconv.ParseInt(rawLookupCalls, 10, 64)
	if err != nil {
		t.Fatalf("invalid lookup call marker %q: %v", rawLookupCalls, err)
	}
	interfaceLookupCalls, err := strconv.ParseInt(rawInterfaceLookupCalls, 10, 64)
	if err != nil {
		t.Fatalf("invalid interface lookup call marker %q: %v", rawInterfaceLookupCalls, err)
	}
	globalLookupCalls, err := strconv.ParseInt(rawGlobalLookupCalls, 10, 64)
	if err != nil {
		t.Fatalf("invalid global lookup call marker %q: %v", rawGlobalLookupCalls, err)
	}
	globalLookupEnvCalls, err := strconv.ParseInt(rawGlobalLookupEnvCalls, 10, 64)
	if err != nil {
		t.Fatalf("invalid global lookup env call marker %q: %v", rawGlobalLookupEnvCalls, err)
	}
	globalLookupRegistryCalls, err := strconv.ParseInt(rawGlobalLookupRegistryCalls, 10, 64)
	if err != nil {
		t.Fatalf("invalid global lookup registry call marker %q: %v", rawGlobalLookupRegistryCalls, err)
	}
	if strictTotalInterfaceLookupAudit() && lookupCalls != 0 {
		t.Fatalf("expected no interpreter member lookup calls for static fixture, got totalCalls=%d interfaceCalls=%d stderr=%q stdout=%q", lookupCalls, interfaceLookupCalls, stderr.String(), stdout.String())
	}
	if interfaceLookupCalls != 0 {
		t.Fatalf("expected no interpreter interface member lookup calls, got interfaceCalls=%d totalCalls=%d stderr=%q stdout=%q", interfaceLookupCalls, lookupCalls, stderr.String(), stdout.String())
	}
	if globalLookupEnvCalls != 0 {
		t.Fatalf("expected no bridge global-env lookup fallback calls for static fixture, got envCalls=%d registryCalls=%d names=%q stderr=%q stdout=%q", globalLookupEnvCalls, globalLookupRegistryCalls, rawGlobalLookupNames, stderr.String(), stdout.String())
	}
	if strictTotalGlobalLookupAudit() && globalLookupCalls != 0 {
		t.Fatalf("expected no bridge global lookup fallback calls for static fixture, got globalCalls=%d envCalls=%d registryCalls=%d names=%q stderr=%q stdout=%q", globalLookupCalls, globalLookupEnvCalls, globalLookupRegistryCalls, rawGlobalLookupNames, stderr.String(), stdout.String())
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

func strictTotalInterfaceLookupAudit() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(compilerInterfaceLookupStrictTotalEnv))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func strictTotalGlobalLookupAudit() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(compilerGlobalLookupStrictTotalEnv))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
