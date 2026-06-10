package main

import (
	"os"
	"path/filepath"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/compiler"
	"able/interpreter-go/pkg/driver"
)

func TestCompiledTestCacheKeyInvalidatesSemanticInputs(t *testing.T) {
	root := t.TempDir()
	runnerPath := filepath.Join(root, "runner.able")
	dependencyPath := filepath.Join(root, "dependency.able")
	if err := os.WriteFile(runnerPath, []byte("package compiled_tests\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dependencyPath, []byte("package dependency\nfn value() -> i32 { 1 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	program := &driver.Program{
		Entry: &driver.Module{Package: "compiled_tests", Files: []string{runnerPath}},
		Modules: []*driver.Module{
			{Package: "dependency", Files: []string{dependencyPath}},
			{Package: "compiled_tests", Files: []string{runnerPath}, Imports: []string{"dependency"}},
		},
	}
	base := compiledTestCacheKeyInput{
		Program:       program,
		EntryPath:     runnerPath,
		RunnerSource:  "package compiled_tests\n",
		HarnessSource: "package main\n",
		SearchPaths: []driver.SearchPath{{
			Path:         root,
			Kind:         driver.RootStdlib,
			StdlibSource: driver.StdlibSourceEnv,
		}},
		Packages: []string{"dependency"},
		CompilerOptions: compiler.Options{
			PackageName:               "main",
			RequireNoFallbacks:        true,
			ExperimentalMonoArrays:    true,
			ExperimentalMonoArraysSet: true,
		},
		ModuleRoot:    root,
		BuildIdentity: "able-and-go-v1",
	}
	baseKey := mustCompiledTestCacheKey(t, base)
	if repeated := mustCompiledTestCacheKey(t, base); repeated != baseKey {
		t.Fatalf("identical inputs produced keys %q and %q", baseKey, repeated)
	}

	assertCompiledTestCacheKeyChanges(t, baseKey, base, func(input *compiledTestCacheKeyInput) {
		input.HarnessSource = "package main\n// reporter configuration changed\n"
	})
	assertCompiledTestCacheKeyChanges(t, baseKey, base, func(input *compiledTestCacheKeyInput) {
		input.CompilerOptions.EmitTypedBoundaryTelemetry = true
	})
	assertCompiledTestCacheKeyChanges(t, baseKey, base, func(input *compiledTestCacheKeyInput) {
		input.BuildIdentity = "able-and-go-v2"
	})
	assertCompiledTestCacheKeyChanges(t, baseKey, base, func(input *compiledTestCacheKeyInput) {
		input.Salt = "manual-invalidation"
	})

	relocatedRoot := t.TempDir()
	relocatedRunner := filepath.Join(relocatedRoot, "runner.able")
	relocatedDependency := filepath.Join(relocatedRoot, "dependency.able")
	if err := os.WriteFile(relocatedRunner, []byte(base.RunnerSource), 0o600); err != nil {
		t.Fatal(err)
	}
	dependencySource, err := os.ReadFile(dependencyPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(relocatedDependency, dependencySource, 0o600); err != nil {
		t.Fatal(err)
	}
	relocated := base
	relocated.EntryPath = relocatedRunner
	relocated.SearchPaths = []driver.SearchPath{{
		Path:         relocatedRoot,
		Kind:         driver.RootStdlib,
		StdlibSource: driver.StdlibSourceEnv,
	}}
	relocated.Program = &driver.Program{
		Entry: &driver.Module{Package: "compiled_tests", Files: []string{relocatedRunner}},
		Modules: []*driver.Module{
			{Package: "dependency", Files: []string{relocatedDependency}},
			{Package: "compiled_tests", Files: []string{relocatedRunner}, Imports: []string{"dependency"}},
		},
	}
	if relocatedKey := mustCompiledTestCacheKey(t, relocated); relocatedKey == baseKey {
		t.Fatal("source relocation did not invalidate source-path-sensitive cache key")
	}

	if err := os.WriteFile(dependencyPath, []byte("package dependency\nfn value() -> i32 { 2 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if changed := mustCompiledTestCacheKey(t, base); changed == baseKey {
		t.Fatal("dependency source change did not invalidate the compiled-test cache key")
	}
}

func TestNormalizeCompiledTestRunnerOriginsLeavesUserSourcesIntact(t *testing.T) {
	root := t.TempDir()
	entryPath := filepath.Join(root, "runner.able")
	userPath := filepath.Join(root, "user.test.able")
	runnerNode := ast.NewIdentifier("runner")
	userNode := ast.NewIdentifier("user")
	program := &driver.Program{Modules: []*driver.Module{{
		Package: "compiled_test_runner.compiled_tests",
		NodeOrigins: map[ast.Node]string{
			runnerNode: entryPath,
			userNode:   userPath,
		},
	}}}

	normalizeCompiledTestRunnerOrigins(program, entryPath)

	origins := program.Modules[0].NodeOrigins
	if got := origins[runnerNode]; got != "compiled-test-runner/runner.able" {
		t.Fatalf("runner origin = %q, want stable synthetic origin", got)
	}
	if got := origins[userNode]; got != userPath {
		t.Fatalf("user origin = %q, want unchanged %q", got, userPath)
	}
}

func TestCompiledTestCachePublishesAndRejectsCorruption(t *testing.T) {
	root := t.TempDir()
	cache := &compiledTestCache{root: root}
	sourceBinary := filepath.Join(root, "source-binary")
	if err := os.WriteFile(sourceBinary, []byte("verified executable bytes"), 0o700); err != nil {
		t.Fatal(err)
	}

	cachedBinary, err := cache.publish("abc123", sourceBinary)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if hit, ok := cache.lookup("abc123"); !ok || hit != cachedBinary {
		t.Fatalf("lookup = %q, %v; want %q, true", hit, ok, cachedBinary)
	}
	second, err := cache.publish("abc123", sourceBinary)
	if err != nil {
		t.Fatalf("repeat publish: %v", err)
	}
	if second != cachedBinary {
		t.Fatalf("repeat publish = %q, want %q", second, cachedBinary)
	}

	if err := os.WriteFile(cachedBinary, []byte("corrupt"), 0o700); err != nil {
		t.Fatal(err)
	}
	if hit, ok := cache.lookup("abc123"); ok || hit != "" {
		t.Fatalf("corrupt lookup = %q, %v; want empty miss", hit, ok)
	}
	repaired, err := cache.publish("abc123", sourceBinary)
	if err != nil {
		t.Fatalf("repair corrupt entry: %v", err)
	}
	if hit, ok := cache.lookup("abc123"); !ok || hit != repaired {
		t.Fatalf("repaired lookup = %q, %v; want %q, true", hit, ok, repaired)
	}
}

func TestCompiledTestCacheIsExplicitlyOptIn(t *testing.T) {
	t.Setenv(compiledTestCacheDirEnv, "")
	if cache, err := openCompiledTestCache(); err != nil || cache != nil {
		t.Fatalf("disabled cache = %#v, %v; want nil, nil", cache, err)
	}

	root := filepath.Join(t.TempDir(), "compiled-cache")
	t.Setenv(compiledTestCacheDirEnv, root)
	cache, err := openCompiledTestCache()
	if err != nil {
		t.Fatalf("enable cache: %v", err)
	}
	if cache == nil || cache.root != cleanAbsolutePath(root) {
		t.Fatalf("enabled cache = %#v, want root %q", cache, cleanAbsolutePath(root))
	}
}

func mustCompiledTestCacheKey(t *testing.T, input compiledTestCacheKeyInput) string {
	t.Helper()
	key, err := compiledTestCacheKey(input)
	if err != nil {
		t.Fatalf("cache key: %v", err)
	}
	return key
}

func assertCompiledTestCacheKeyChanges(
	t *testing.T,
	baseKey string,
	input compiledTestCacheKeyInput,
	change func(*compiledTestCacheKeyInput),
) {
	t.Helper()
	change(&input)
	if changed := mustCompiledTestCacheKey(t, input); changed == baseKey {
		t.Fatalf("semantic input change did not invalidate cache key %q", baseKey)
	}
}
