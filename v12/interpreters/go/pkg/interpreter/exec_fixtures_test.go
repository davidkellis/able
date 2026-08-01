package interpreter

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
	"able/interpreter-go/pkg/stdlibpath"
)

func TestExecFixtures(t *testing.T) {
	execMode := resolveTestExecMode(t)
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
			runExecFixture(t, dir, execMode)
		})
	}
}

func TestBuildExecSearchPathsRejectsDistinctStdlibRoots(t *testing.T) {
	root := t.TempDir()
	fixtureDir := filepath.Join(root, "fixture")
	cacheDir := filepath.Join(root, "cache")
	cacheRoot := filepath.Join(cacheDir, "pkg", "src", "able", "0.1.0")
	cacheSrc := filepath.Join(cacheRoot, "src")
	envRoot := filepath.Join(root, "env-stdlib")
	envSrc := filepath.Join(envRoot, "src")
	entryPath := filepath.Join(fixtureDir, "main.able")

	for _, dir := range []string{fixtureDir, cacheSrc, envSrc} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(cacheRoot, "package.yml"), []byte("name: able\nversion: 0.1.0\n"), 0o600); err != nil {
		t.Fatalf("write cache package.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(envRoot, "package.yml"), []byte("name: able\nversion: 9.9.9\n"), 0o600); err != nil {
		t.Fatalf("write env package.yml: %v", err)
	}

	t.Setenv("ABLE_HOME", cacheDir)
	t.Setenv("ABLE_MODULE_PATHS", envSrc)

	_, err := buildExecSearchPaths(entryPath, fixtureDir, fixtureManifest{})
	if err == nil {
		t.Fatalf("expected stdlib collision error")
	}
	if !strings.Contains(err.Error(), "stdlib collision") ||
		!strings.Contains(err.Error(), "env") ||
		!strings.Contains(err.Error(), "cache") {
		t.Fatalf("unexpected stdlib collision error: %v", err)
	}
}

func TestBuildExecSearchPathsSourceRootOnlyExcludesWorkingDirectoryPackage(t *testing.T) {
	root := t.TempDir()
	entryDir := filepath.Join(root, "entry")
	dataDir := filepath.Join(root, "data")
	entryPath := filepath.Join(entryDir, "main.able")
	for _, dir := range []string{entryDir, dataDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(entryPath, []byte("package sample\n\nfn main() {}\n"), 0o600); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "run.able"), []byte("package sample\n\nfn main() {}\n"), 0o600); err != nil {
		t.Fatalf("write data source: %v", err)
	}

	t.Chdir(dataDir)
	t.Setenv("ABLE_SOURCE_ROOT_ONLY", "1")
	paths, err := buildExecSearchPaths(entryPath, dataDir, fixtureManifest{})
	if err != nil {
		t.Fatalf("buildExecSearchPaths() error: %v", err)
	}
	for _, path := range paths {
		if path.Kind == driver.RootUser && path.Path == dataDir {
			t.Fatalf("source-root-only search paths included working directory %s", dataDir)
		}
	}

	loader, err := driver.NewLoader(paths)
	if err != nil {
		t.Fatalf("new loader: %v", err)
	}
	defer loader.Close()
	if _, err := loader.Load(entryPath); err != nil {
		t.Fatalf("load explicit entry with duplicate data package: %v", err)
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
	return selectExecFixtureBatch(t, dirs)
}

func selectExecFixtureBatch(t *testing.T, dirs []string) []string {
	t.Helper()
	rawIndex := os.Getenv("ABLE_EXEC_FIXTURE_BATCH_INDEX")
	rawCount := os.Getenv("ABLE_EXEC_FIXTURE_BATCH_COUNT")
	if rawIndex == "" && rawCount == "" {
		return dirs
	}
	if rawIndex == "" || rawCount == "" {
		t.Fatal("ABLE_EXEC_FIXTURE_BATCH_INDEX and ABLE_EXEC_FIXTURE_BATCH_COUNT must be set together")
	}
	index, indexErr := strconv.Atoi(rawIndex)
	count, countErr := strconv.Atoi(rawCount)
	if indexErr != nil || countErr != nil || count <= 0 || index < 0 || index >= count {
		t.Fatalf("invalid exec fixture batch %q of %q", rawIndex, rawCount)
	}
	selected := make([]string, 0, (len(dirs)+count-1)/count)
	for fixtureIndex, dir := range dirs {
		if fixtureIndex%count == index {
			selected = append(selected, dir)
		}
	}
	return selected
}

func TestSelectExecFixtureBatch(t *testing.T) {
	dirs := []string{"a", "b", "c", "d", "e"}
	t.Setenv("ABLE_EXEC_FIXTURE_BATCH_INDEX", "1")
	t.Setenv("ABLE_EXEC_FIXTURE_BATCH_COUNT", "2")
	if got, want := selectExecFixtureBatch(t, dirs), []string{"b", "d"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("selected batch = %v, want %v", got, want)
	}
}

func runExecFixture(t *testing.T, dir string, execMode testExecMode) {
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
	defer closeFixtureExecutor(executor)
	interp := newTestInterpreter(t, execMode, executor)
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
			actualErrs = expandFixtureLines([]string{DescribeRuntimeDiagnostic(interp.BuildRuntimeDiagnostic(runtimeErr))})
		}
		expectedErrs := expandFixtureLines(expected.Stderr)
		if !reflect.DeepEqual(actualErrs, expectedErrs) {
			t.Fatalf("stderr mismatch: expected %v, got %v", expectedErrs, actualErrs)
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

func expandFixtureLines(lines []string) []string {
	if len(lines) == 0 {
		return []string{}
	}
	var out []string
	for _, raw := range lines {
		trimmed := strings.TrimRight(raw, "\n")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		parts := strings.Split(trimmed, "\n")
		out = append(out, parts...)
	}
	return out
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

func closeFixtureExecutor(executor Executor) {
	if closer, ok := executor.(interface{ Close() }); ok {
		closer.Close()
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
	add := func(candidate string, kind driver.RootKind, source driver.StdlibSourceClass) {
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
		paths = append(paths, driver.SearchPath{Path: abs, Kind: kind, StdlibSource: source})
	}

	for _, extra := range []string{manifestRoot, entryDir} {
		add(extra, driver.RootUser, driver.StdlibSourceWorkspace)
	}
	if !sourceRootOnlyFixtureSearchPaths() {
		if cwd, err := os.Getwd(); err == nil {
			add(cwd, driver.RootUser, driver.StdlibSourceWorkspace)
		}
	}
	for _, entry := range resolveFixturePathList(ablePathEnv, fixtureDir) {
		add(entry, driver.RootUser, driver.StdlibSourceEnv)
	}
	for _, entry := range resolveFixturePathList(ableModulePathsEnv, fixtureDir) {
		add(entry, driver.RootUser, driver.StdlibSourceEnv)
	}
	for _, entry := range findKernelRoots(entryDir) {
		add(entry, driver.RootStdlib, driver.StdlibSourceUnknown)
	}
	stdlibRootsAdded := false
	addStdlibRoots := func(start string) {
		for _, entry := range findStdlibRoots(start) {
			pathCount := len(paths)
			add(entry, driver.RootStdlib, fixtureStdlibRootSource(entry))
			if len(paths) > pathCount {
				stdlibRootsAdded = true
			}
		}
	}
	addStdlibRoots(entryDir)
	if cwd, err := os.Getwd(); err == nil {
		for _, entry := range findKernelRoots(cwd) {
			add(entry, driver.RootStdlib, driver.StdlibSourceUnknown)
		}
		if !stdlibRootsAdded {
			addStdlibRoots(cwd)
		}
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, entry := range findKernelRoots(exeDir) {
			add(entry, driver.RootStdlib, driver.StdlibSourceUnknown)
		}
		if !stdlibRootsAdded {
			addStdlibRoots(exeDir)
		}
	}
	return driver.ResolveCanonicalStdlibSearchPaths(paths, false)
}

// sourceRootOnlyFixtureSearchPaths mirrors the CLI's source-root-only mode for
// the complete-program benchmark harness. The caller's CWD can supply input
// data without becoming a second source root containing a duplicate package.
func sourceRootOnlyFixtureSearchPaths() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ABLE_SOURCE_ROOT_ONLY"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
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
		if len(roots) > 0 {
			return roots
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
	if configured := strings.TrimSpace(os.Getenv("ABLE_STDLIB_ROOT")); configured != "" {
		add(configured)
		return roots
	}
	dir := start
	for {
		for _, candidate := range []string{
			filepath.Join(dir, "stdlib", "src"),
			filepath.Join(dir, "able-stdlib", "src"),
			filepath.Join(dir, "able_stdlib", "src"),
		} {
			add(candidate)
		}
		if len(roots) > 0 {
			return roots
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if installed := stdlibpath.ResolveInstalledSrc(); installed != "" {
		add(installed)
	}
	return roots
}

func fixtureStdlibRootSource(root string) driver.StdlibSourceClass {
	canonical, err := driver.CanonicalizeStdlibCandidateRoot(root)
	if err != nil {
		return driver.StdlibSourceUnknown
	}
	if configured := strings.TrimSpace(os.Getenv("ABLE_STDLIB_ROOT")); configured != "" {
		if configuredRoot, err := driver.CanonicalizeStdlibCandidateRoot(configured); err == nil && configuredRoot == canonical {
			return driver.StdlibSourceEnv
		}
	}
	if installed := stdlibpath.ResolveInstalledSrc(); installed != "" {
		if installedRoot, err := driver.CanonicalizeStdlibCandidateRoot(installed); err == nil && installedRoot == canonical {
			return driver.StdlibSourceCache
		}
	}
	if filepath.Base(filepath.Dir(canonical)) == "able-stdlib" {
		return driver.StdlibSourceOverride
	}
	return driver.StdlibSourceUnknown
}
