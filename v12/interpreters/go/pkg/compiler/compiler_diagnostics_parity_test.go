package compiler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

const compilerDiagnosticsFixtureEnv = "ABLE_COMPILER_DIAGNOSTICS_FIXTURES"

type compilerFixtureOutcome struct {
	Stdout []string
	Stderr []string
	Exit   int
}

func TestCompilerDiagnosticsParityFixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiler diagnostics parity in short mode")
	}
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	fixtures := resolveCompilerDiagnosticsFixtures()
	if len(fixtures) == 0 {
		t.Skip("no diagnostics fixtures configured")
	}
	for _, rel := range fixtures {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			t.Parallel()
			dir := filepath.Join(root, filepath.FromSlash(rel))
			manifest, err := interpreter.LoadFixtureManifest(dir)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if shouldSkipTarget(manifest.SkipTargets, "go") {
				return
			}
			if manifest.Expect.TypecheckDiagnostics != nil && len(manifest.Expect.TypecheckDiagnostics) > 0 {
				return
			}
			tree := runTreewalkerFixtureOutcome(t, dir, manifest)
			compiled := runCompiledFixtureOutcome(t, dir, manifest)

			if tree.Exit != compiled.Exit {
				t.Fatalf("exit mismatch: treewalker=%d compiled=%d", tree.Exit, compiled.Exit)
			}
			if !reflect.DeepEqual(tree.Stdout, compiled.Stdout) {
				t.Fatalf("stdout mismatch: treewalker=%v compiled=%v", tree.Stdout, compiled.Stdout)
			}
			if !reflect.DeepEqual(tree.Stderr, compiled.Stderr) {
				t.Fatalf("stderr mismatch: treewalker=%v compiled=%v", tree.Stderr, compiled.Stderr)
			}
		})
	}
}

func resolveCompilerDiagnosticsFixtures() []string {
	if raw := strings.TrimSpace(os.Getenv(compilerDiagnosticsFixtureEnv)); raw != "" {
		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
		})
		fixtures := make([]string, 0, len(parts))
		seen := map[string]struct{}{}
		for _, part := range parts {
			name := strings.TrimSpace(part)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			fixtures = append(fixtures, name)
		}
		return fixtures
	}
	return []string{
		"04_02_primitives_truthiness_numeric_diag",
		"06_01_compiler_division_by_zero",
		"06_01_compiler_integer_overflow",
		"06_01_compiler_integer_overflow_sub",
		"06_01_compiler_integer_overflow_mul",
		"06_01_compiler_unary_overflow",
		"06_01_compiler_divmod_overflow",
		"06_01_compiler_pow_overflow",
		"06_01_compiler_pow_negative_exponent",
		"06_01_compiler_shift_out_of_range",
		"06_01_compiler_compound_assignment_overflow",
	}
}

func runTreewalkerFixtureOutcome(t *testing.T, dir string, manifest interpreter.FixtureManifest) compilerFixtureOutcome {
	t.Helper()
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

	var executor interpreter.Executor = interpreter.NewSerialExecutor(nil)
	switch strings.ToLower(strings.TrimSpace(manifest.Executor)) {
	case "", "serial":
	case "goroutine":
		executor = interpreter.NewGoroutineExecutor(nil)
	default:
		t.Fatalf("unknown fixture executor %q", manifest.Executor)
	}
	interp := interpreter.NewWithExecutor(executor)
	interp.SetArgs(nil)
	stdout := make([]string, 0)
	registerCompilerDiagnosticsPrint(interp, &stdout)

	mode := resolveCompilerDiagnosticsTypecheckMode()
	var entryEnv *runtime.Environment
	var runtimeErr error
	exitCode := 0
	withFixtureEnvOverrides(manifest.Env, func() {
		_, env, _, err := interp.EvaluateProgram(program, interpreter.ProgramEvaluationOptions{
			SkipTypecheck:    mode == diagnosticsTypecheckModeOff,
			AllowDiagnostics: mode != diagnosticsTypecheckModeOff,
		})
		if err != nil {
			if code, ok := interpreter.ExitCodeFromError(err); ok {
				exitCode = code
				return
			}
			runtimeErr = err
			exitCode = 1
			return
		}
		entryEnv = env
		envToUse := entryEnv
		if envToUse == nil {
			envToUse = interp.GlobalEnvironment()
		}
		mainValue, getErr := envToUse.Get("main")
		if getErr != nil {
			runtimeErr = getErr
			exitCode = 1
			return
		}
		if _, callErr := interp.CallFunction(mainValue, nil); callErr != nil {
			if code, ok := interpreter.ExitCodeFromError(callErr); ok {
				exitCode = code
				return
			}
			runtimeErr = callErr
			exitCode = 1
		}
	})

	stderr := []string{}
	if runtimeErr != nil {
		diag := interpreter.DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(runtimeErr))
		stderr = expandFixtureLines([]string{diag})
	}
	return compilerFixtureOutcome{
		Stdout: stdout,
		Stderr: stderr,
		Exit:   exitCode,
	}
}

func runCompiledFixtureOutcome(t *testing.T, dir string, manifest interpreter.FixtureManifest) compilerFixtureOutcome {
	t.Helper()
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

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-diag-parity")

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
	return compilerFixtureOutcome{
		Stdout: splitLines(stdout.String()),
		Stderr: splitLines(stderr.String()),
		Exit:   exitCode,
	}
}

type diagnosticsTypecheckMode int

const (
	diagnosticsTypecheckModeOff diagnosticsTypecheckMode = iota
	diagnosticsTypecheckModeWarn
	diagnosticsTypecheckModeStrict
)

func resolveCompilerDiagnosticsTypecheckMode() diagnosticsTypecheckMode {
	if _, ok := os.LookupEnv("ABLE_TYPECHECK_FIXTURES"); !ok {
		return diagnosticsTypecheckModeStrict
	}
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("ABLE_TYPECHECK_FIXTURES")))
	switch mode {
	case "", "0", "off", "false":
		return diagnosticsTypecheckModeOff
	case "strict", "fail", "error", "1", "true":
		return diagnosticsTypecheckModeStrict
	case "warn", "warning":
		return diagnosticsTypecheckModeWarn
	default:
		return diagnosticsTypecheckModeWarn
	}
}

func withFixtureEnvOverrides(overrides map[string]string, fn func()) {
	if len(overrides) == 0 {
		fn()
		return
	}
	originals := make(map[string]*string, len(overrides))
	for key, value := range overrides {
		if current, ok := os.LookupEnv(key); ok {
			copyVal := current
			originals[key] = &copyVal
		} else {
			originals[key] = nil
		}
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key, original := range originals {
			if original == nil {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, *original)
			}
		}
	}()
	fn()
}

func registerCompilerDiagnosticsPrint(interp *interpreter.Interpreter, stdout *[]string) {
	if interp == nil {
		return
	}
	printFn := runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			parts := make([]string, 0, len(args))
			for _, arg := range args {
				parts = append(parts, formatCompilerDiagnosticsValue(arg))
			}
			*stdout = append(*stdout, strings.Join(parts, " "))
			return runtime.NilValue{}, nil
		},
	}
	interp.GlobalEnvironment().Define("print", printFn)
}

func formatCompilerDiagnosticsValue(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.VoidValue:
		return "void"
	case runtime.IntegerValue:
		return v.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	default:
		return fmt.Sprintf("[%s]", v.Kind())
	}
}
