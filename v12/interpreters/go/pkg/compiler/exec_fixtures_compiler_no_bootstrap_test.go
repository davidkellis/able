package compiler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

func TestCompilerNoBootstrapExecFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping no-bootstrap compiler fixtures in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveNoBootstrapFixtures(t, root)
	if len(fixtures) == 0 {
		t.Skip("no no-bootstrap compiler fixtures configured (set ABLE_COMPILER_NO_BOOTSTRAP_FIXTURES)")
	}
	for _, rel := range fixtures {
		rel := rel
		dir := filepath.Join(root, filepath.FromSlash(rel))
		t.Run(filepath.ToSlash(rel), func(t *testing.T) {
			runCompilerNoBootstrapExecFixture(t, dir, rel)
		})
	}
}

func resolveNoBootstrapFixtures(t *testing.T, root string) []string {
	raw := strings.TrimSpace(os.Getenv(compilerNoBootstrapFixtureEnv))
	if raw == "" {
		return nil
	}
	if strings.EqualFold(raw, "all") {
		return collectExecFixtures(t, root)
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
	})
	fixtures := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		fixtures = append(fixtures, trimmed)
	}
	return fixtures
}

func runCompilerNoBootstrapExecFixture(t *testing.T, dir string, rel string) {
	t.Helper()
	manifest, err := interpreter.LoadFixtureManifest(dir)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if shouldSkipTarget(manifest.SkipTargets, "go") {
		return
	}
	// Diag-only fixtures expect typechecker diagnostics, not runtime output.
	// Skip them in no-bootstrap mode (same as the bootstrap harness does).
	if len(manifest.Expect.TypecheckDiagnostics) > 0 {
		t.Skipf("skipping diag-only fixture in no-bootstrap mode")
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
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	moduleRoot, workDir := compilerTestWorkDirNoCleanup(t, "ablec-noboot")
	passed := false
	t.Cleanup(func() {
		if passed {
			_ = os.RemoveAll(workDir)
		} else {
			t.Logf("preserved work dir: %s", workDir)
		}
	})

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

	harness := compilerHarnessSourceNoBootstrap(manifest.Executor)
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte(harness), 0o600); err != nil {
		t.Fatalf("write harness: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-fixture")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = workDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath)
	cmd.Env = applyFixtureEnv(os.Environ(), manifest.Env)
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

	actualStdout := splitLines(stdout.String())
	actualStderr := splitLines(stderr.String())
	expected := manifest.Expect
	if expected.Stdout != nil {
		expectedStdout := expandFixtureLines(expected.Stdout)
		if !reflect.DeepEqual(actualStdout, expectedStdout) {
			t.Fatalf("stdout mismatch: expected %v, got %v\nstderr: %s", expectedStdout, actualStdout, stderr.String())
		}
	}
	if expected.Stderr != nil {
		expectedStderr := expandFixtureLines(expected.Stderr)
		if !reflect.DeepEqual(actualStderr, expectedStderr) {
			t.Fatalf("stderr mismatch: expected %v, got %v", expectedStderr, actualStderr)
		}
	}
	if expected.Exit != nil {
		if exitCode != *expected.Exit {
			t.Fatalf("exit code mismatch: expected %d, got %d\nstderr: %s", *expected.Exit, exitCode, stderr.String())
		}
	} else if exitCode != 0 {
		t.Fatalf("exit code mismatch: expected 0, got %d\nstderr: %s", exitCode, stderr.String())
	}
	passed = true
}

func compilerHarnessSourceNoBootstrap(executorName string) string {
	var buf strings.Builder
	buf.WriteString("package main\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n")
	buf.WriteString("\t\"os\"\n")
	buf.WriteString("\t\"strings\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/compiler/bridge\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/interpreter\"\n")
	buf.WriteString("\t\"able/interpreter-go/pkg/runtime\"\n")
	buf.WriteString(")\n\n")
	buf.WriteString("func main() {\n")
	buf.WriteString(fmt.Sprintf("\texecutor := selectFixtureExecutor(%q)\n", executorName))
	buf.WriteString("\tinterp := interpreter.NewWithExecutor(executor)\n")
	buf.WriteString("\tinterp.SetArgs(os.Args[1:])\n")
	buf.WriteString("\tregisterPrint(interp)\n")
	buf.WriteString("\tentryEnv := interp.GlobalEnvironment()\n")
	buf.WriteString("\trt, err := RegisterIn(interp, entryEnv)\n")
	buf.WriteString("\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tprintBoundaryMarkers := func() {\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_BOUNDARY_MARKER\") == \"\" {\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tcount := int64(0)\n")
	buf.WriteString("\t\texplicitCount := int64(0)\n")
	buf.WriteString("\t\tif rt != nil {\n")
	buf.WriteString("\t\t\tcount = __able_boundary_fallback_count_get()\n")
	buf.WriteString("\t\t\texplicitCount = __able_boundary_explicit_count_get()\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_BOUNDARY_FALLBACK_CALLS=%d\\n\", count)\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_BOUNDARY_EXPLICIT_CALLS=%d\\n\", explicitCount)\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE\") != \"\" {\n")
	buf.WriteString("\t\t\tfmt.Fprintf(os.Stderr, \"__ABLE_BOUNDARY_FALLBACK_NAMES=%s\\n\", __able_boundary_fallback_snapshot())\n")
	buf.WriteString("\t\t\tfmt.Fprintf(os.Stderr, \"__ABLE_BOUNDARY_EXPLICIT_NAMES=%s\\n\", __able_boundary_explicit_snapshot())\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tprintGlobalLookupMarkers := func() {\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_GLOBAL_LOOKUP_MARKER\") == \"\" {\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tcalls := bridge.GlobalLookupFallbackStats()\n")
	buf.WriteString("\t\tenvCalls, registryCalls := bridge.GlobalLookupFallbackBucketStats()\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_GLOBAL_LOOKUP_FALLBACK_CALLS=%d\\n\", calls)\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_GLOBAL_LOOKUP_FALLBACK_ENV_CALLS=%d\\n\", envCalls)\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_GLOBAL_LOOKUP_FALLBACK_REGISTRY_CALLS=%d\\n\", registryCalls)\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE\") != \"\" {\n")
	buf.WriteString("\t\t\tfmt.Fprintf(os.Stderr, \"__ABLE_GLOBAL_LOOKUP_FALLBACK_NAMES=%s\\n\", bridge.GlobalLookupFallbackSnapshot())\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tprintInterfaceLookupMarkers := func() {\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_INTERFACE_LOOKUP_MARKER\") == \"\" {\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tcalls, ifaceCalls := bridge.MemberGetPreferMethodsStats()\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_MEMBER_LOOKUP_CALLS=%d\\n\", calls)\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"__ABLE_MEMBER_LOOKUP_INTERFACE_CALLS=%d\\n\", ifaceCalls)\n")
	buf.WriteString("\t\tif os.Getenv(\"ABLE_COMPILER_BOUNDARY_MARKER_VERBOSE\") != \"\" {\n")
	buf.WriteString("\t\t\tfmt.Fprintf(os.Stderr, \"__ABLE_MEMBER_LOOKUP_NAMES=%s\\n\", bridge.MemberGetPreferMethodsSnapshot())\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tbridge.ResetMemberGetPreferMethodsCounters()\n")
	buf.WriteString("\tbridge.ResetGlobalLookupFallbackCounters()\n")
	buf.WriteString("\tif err := RunRegisteredMain(rt, interp, entryEnv); err != nil {\n")
	buf.WriteString("\t\tprintBoundaryMarkers()\n")
	buf.WriteString("\t\tprintGlobalLookupMarkers()\n")
	buf.WriteString("\t\tprintInterfaceLookupMarkers()\n")
	buf.WriteString("\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n\t\t\tos.Exit(code)\n\t\t}\n")
	buf.WriteString("\t\tfmt.Fprintln(os.Stderr, interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(err)))\n")
	buf.WriteString("\t\tos.Exit(1)\n\t}\n")
	buf.WriteString("\tprintBoundaryMarkers()\n")
	buf.WriteString("\tprintGlobalLookupMarkers()\n")
	buf.WriteString("\tprintInterfaceLookupMarkers()\n")
	buf.WriteString("}\n\n")
	buf.WriteString("func selectFixtureExecutor(name string) interpreter.Executor {\n")
	buf.WriteString("\tswitch strings.ToLower(strings.TrimSpace(name)) {\n")
	buf.WriteString("\tcase \"\", \"serial\":\n")
	buf.WriteString("\t\treturn interpreter.NewSerialExecutor(nil)\n")
	buf.WriteString("\tcase \"goroutine\":\n")
	buf.WriteString("\t\treturn interpreter.NewGoroutineExecutor(nil)\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\tfmt.Fprintf(os.Stderr, \"unknown fixture executor %q\\n\", name)\n")
	buf.WriteString("\t\tos.Exit(1)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn nil\n")
	buf.WriteString("}\n\n")
	buf.WriteString("func registerPrint(interp *interpreter.Interpreter) {\n")
	buf.WriteString("\tprintFn := runtime.NativeFunctionValue{\n")
	buf.WriteString("\t\tName:  \"print\",\n\t\tArity: 1,\n")
	buf.WriteString("\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	buf.WriteString("\t\t\tvar parts []string\n")
	buf.WriteString("\t\t\tfor _, arg := range args {\n\t\t\t\tparts = append(parts, formatRuntimeValue(arg))\n\t\t\t}\n")
	buf.WriteString("\t\t\tfmt.Fprintln(os.Stdout, strings.Join(parts, \" \"))\n")
	buf.WriteString("\t\t\treturn runtime.NilValue{}, nil\n\t\t},\n\t}\n")
	buf.WriteString("\tinterp.GlobalEnvironment().Define(\"print\", printFn)\n}\n\n")
	buf.WriteString("func formatRuntimeValue(val runtime.Value) string {\n")
	buf.WriteString("\tswitch v := val.(type) {\n")
	buf.WriteString("\tcase runtime.StringValue:\n\t\treturn v.Val\n")
	buf.WriteString("\tcase runtime.BoolValue:\n\t\tif v.Val { return \"true\" }; return \"false\"\n")
	buf.WriteString("\tcase runtime.VoidValue:\n\t\treturn \"void\"\n")
	buf.WriteString("\tcase runtime.IntegerValue:\n\t\treturn v.String()\n")
	buf.WriteString("\tcase runtime.FloatValue:\n\t\treturn fmt.Sprintf(\"%g\", v.Val)\n")
	buf.WriteString("\tdefault:\n\t\treturn fmt.Sprintf(\"[%s]\", v.Kind())\n\t}\n}\n")
	return buf.String()
}
