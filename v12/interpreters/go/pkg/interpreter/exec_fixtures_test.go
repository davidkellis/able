package interpreter

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

func TestExecFixtures(t *testing.T) {
	root := filepath.Join(repositoryRoot(), "v12", "fixtures", "exec")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = filepath.Join("..", "..", "fixtures", "exec")
	}
	dirs := collectExecFixtures(t, root)
	for _, dir := range dirs {
		dir := dir
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			t.Fatalf("relative path for %s: %v", dir, err)
		}
		t.Run(filepath.ToSlash(rel), func(t *testing.T) {
			runExecFixture(t, dir)
		})
	}
}

func collectExecFixtures(t *testing.T, root string) []string {
	t.Helper()
	if root == "" {
		return nil
	}
	var dirs []string
	var walk func(string)
	walk = func(current string) {
		entries, err := os.ReadDir(current)
		if err != nil {
			return
		}
		hasManifest := false
		for _, entry := range entries {
			if entry.Type().IsRegular() && entry.Name() == "manifest.json" {
				hasManifest = true
				break
			}
		}
		if hasManifest {
			dirs = append(dirs, current)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				walk(filepath.Join(current, entry.Name()))
			}
		}
	}
	walk(root)
	return dirs
}

func runExecFixture(t *testing.T, dir string) {
	t.Helper()

	manifest := readManifest(t, dir)
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

	expectedTypecheck := manifest.Expect.TypecheckDiagnostics
	if expectedTypecheck != nil {
		check, err := TypecheckProgram(program)
		if err != nil {
			t.Fatalf("typecheck program: %v", err)
		}
		formatted := formatModuleDiagnostics(check.Diagnostics)
		if len(expectedTypecheck) == 0 {
			if len(formatted) != 0 {
				t.Fatalf("typecheck diagnostics mismatch: expected none, got %v", formatted)
			}
		} else {
			expectedKeys := diagnosticKeys(expectedTypecheck)
			actualKeys := diagnosticKeys(formatted)
			if len(expectedKeys) != len(actualKeys) {
				t.Fatalf("typecheck diagnostics mismatch: expected %v, got %v", expectedTypecheck, formatted)
			}
			for i := range expectedKeys {
				if expectedKeys[i] != actualKeys[i] {
					t.Fatalf("typecheck diagnostics mismatch: expected %v, got %v", expectedTypecheck, formatted)
				}
			}
		}
	}
	if expectedTypecheck != nil && len(expectedTypecheck) > 0 {
		return
	}

	executor := selectFixtureExecutor(t, manifest.Executor)
	interp := NewWithExecutor(executor)
	mode := configureFixtureTypechecker(interp)
	var stdout []string
	registerPrint(interp, &stdout)

	exitCode := 0
	var runtimeErr error
	exitSignaled := false

	entryEnv := interp.GlobalEnvironment()
	_, entryEnv, _, err = interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    mode == typecheckModeOff,
		AllowDiagnostics: mode != typecheckModeOff,
	})
	if err != nil {
		if code, ok := ExitCodeFromError(err); ok {
			exitCode = code
			exitSignaled = true
		} else {
			runtimeErr = err
			exitCode = 1
		}
	}

	var mainValue runtime.Value
	if runtimeErr == nil {
		env := entryEnv
		if env == nil {
			env = interp.GlobalEnvironment()
		}
		val, err := env.Get("main")
		if err != nil {
			runtimeErr = err
			exitCode = 1
		} else {
			mainValue = val
		}
	}

	if runtimeErr == nil {
		if _, err := interp.CallFunction(mainValue, nil); err != nil {
			if code, ok := ExitCodeFromError(err); ok {
				exitCode = code
				exitSignaled = true
			} else {
				runtimeErr = err
				exitCode = 1
			}
		}
	}

	expected := manifest.Expect

	if runtimeErr != nil {
		if expected.Exit == nil || exitCode != *expected.Exit {
			t.Fatalf("runtime error: %v", runtimeErr)
		}
	}

	if expected.Stdout != nil {
		if !reflect.DeepEqual(stdout, expected.Stdout) {
			t.Fatalf("stdout mismatch: expected %v, got %v", expected.Stdout, stdout)
		}
	}

	if expected.Stderr != nil {
		actualErrs := []string{}
		if runtimeErr != nil {
			actualErrs = append(actualErrs, DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(runtimeErr)))
		}
		if !reflect.DeepEqual(actualErrs, expected.Stderr) {
			t.Fatalf("stderr mismatch: expected %v, got %v", expected.Stderr, actualErrs)
		}
	}

	if expected.Exit != nil {
		if exitCode != *expected.Exit {
			t.Fatalf("exit code mismatch: expected %d, got %d", *expected.Exit, exitCode)
		}
	} else if exitSignaled {
		t.Fatalf("exit code mismatch: expected default exit, got %d", exitCode)
	} else if runtimeErr != nil {
		t.Fatalf("runtime error: %v", runtimeErr)
	}
}

func selectFixtureExecutor(t *testing.T, name string) Executor {
	t.Helper()
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "serial":
		return NewSerialExecutor(nil)
	case "goroutine":
		return NewGoroutineExecutor(nil)
	default:
		t.Fatalf("unknown fixture executor %q", name)
		return nil
	}
}

func buildExecSearchPaths(entryPath string, fixtureDir string, manifest fixtureManifest) ([]driver.SearchPath, error) {
	entryAbs, err := filepath.Abs(entryPath)
	if err != nil {
		return nil, err
	}
	entryDir := filepath.Dir(entryAbs)

	manifestRoot := findFixtureManifestRoot(entryDir)
	ablePathEnv := resolveFixtureEnv("ABLE_PATH", manifest.Env, os.Getenv("ABLE_PATH"))
	ableModulePathsEnv := resolveFixtureEnv("ABLE_MODULE_PATHS", manifest.Env, os.Getenv("ABLE_MODULE_PATHS"))

	var paths []driver.SearchPath
	seen := map[string]struct{}{}
	add := func(candidate string, kind driver.RootKind) {
		if candidate == "" {
			return
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			return
		}
		if _, ok := seen[abs]; ok {
			return
		}
		seen[abs] = struct{}{}
		paths = append(paths, driver.SearchPath{Path: abs, Kind: kind})
	}

	for _, extra := range []string{manifestRoot, entryDir} {
		add(extra, driver.RootUser)
	}
	if cwd, err := os.Getwd(); err == nil {
		add(cwd, driver.RootUser)
	}
	for _, entry := range resolveFixturePathList(ablePathEnv, fixtureDir) {
		add(entry, driver.RootUser)
	}
	for _, entry := range resolveFixturePathList(ableModulePathsEnv, fixtureDir) {
		add(entry, driver.RootUser)
	}
	for _, entry := range findKernelRoots(entryDir) {
		add(entry, driver.RootStdlib)
	}
	for _, entry := range findStdlibRoots(entryDir) {
		add(entry, driver.RootStdlib)
	}
	if cwd, err := os.Getwd(); err == nil {
		for _, entry := range findKernelRoots(cwd) {
			add(entry, driver.RootStdlib)
		}
		for _, entry := range findStdlibRoots(cwd) {
			add(entry, driver.RootStdlib)
		}
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, entry := range findKernelRoots(exeDir) {
			add(entry, driver.RootStdlib)
		}
		for _, entry := range findStdlibRoots(exeDir) {
			add(entry, driver.RootStdlib)
		}
	}
	return paths, nil
}

func resolveFixtureEnv(key string, env map[string]string, fallback string) string {
	if env == nil {
		return fallback
	}
	if value, ok := env[key]; ok {
		return value
	}
	return fallback
}

func resolveFixturePathList(raw string, baseDir string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	rawParts := strings.Split(raw, string(os.PathListSeparator))
	parts := make([]string, 0, len(rawParts))
	for _, entry := range rawParts {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		if !filepath.IsAbs(trimmed) {
			trimmed = filepath.Join(baseDir, filepath.FromSlash(trimmed))
		}
		parts = append(parts, trimmed)
	}
	return parts
}

func findFixtureManifestRoot(start string) string {
	dir := start
	for {
		candidate := filepath.Join(dir, "package.yml")
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func findKernelRoots(start string) []string {
	var roots []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}
	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "kernel", "src"),
			filepath.Join(dir, "v12", "kernel", "src"),
			filepath.Join(dir, "ablekernel", "src"),
			filepath.Join(dir, "able_kernel", "src"),
		} {
			add(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}

func findStdlibRoots(start string) []string {
	var roots []string
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}
	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "stdlib", "src"),
			filepath.Join(dir, "v12", "stdlib", "src"),
			filepath.Join(dir, "stdlib", "v12", "src"),
			filepath.Join(dir, "able-stdlib", "src"),
			filepath.Join(dir, "able_stdlib", "src"),
		} {
			add(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return roots
}
