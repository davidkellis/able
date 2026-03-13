package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerCompiledHashSetUnionStdlib(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compiled stdlib hash set union test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, workDir := compilerTestWorkDirNoCleanup(t, "ablec-hashset-union")
	if os.Getenv("ABLE_TEST_KEEP_WORKDIR") == "" {
		t.Cleanup(func() { _ = os.RemoveAll(workDir) })
	} else {
		t.Logf("keeping workdir %s", workDir)
	}

	source := `import able.collections.hash_set.{HashSet}
extern go fn __able_os_exit(code: i32) -> void {}

fn main() {
  left := HashSet.new()
  left.add("a")
  left.add("b")
  right := HashSet.new()
  right.add("b")
  right.add("c")
  merged := left.union(right)
  if merged.size() == 3 && merged.contains("c") {
    __able_os_exit(0)
  } else {
    __able_os_exit(1)
  }
}
`
	entryPath := filepath.Join(workDir, "app.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	searchPaths := compilerStdlibSearchPaths(moduleRoot)
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}

	outputDir := filepath.Join(workDir, "out")
	comp := New(Options{
		PackageName: "main",
		EmitMain:    true,
		EntryPath:   entryPath,
	})
	result, err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	binPath := filepath.Join(workDir, "compiled-bin")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = outputDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(output))
	}

	run := exec.Command(binPath)
	if output, err := run.CombinedOutput(); err != nil {
		t.Fatalf("run failed: %v\n%s", err, string(output))
	}
}

func compilerStdlibSearchPaths(moduleRoot string) []driver.SearchPath {
	paths := make([]driver.SearchPath, 0)
	seen := map[string]struct{}{}
	add := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, exists := seen[clean]; exists {
			return
		}
		seen[clean] = struct{}{}
		paths = append(paths, driver.SearchPath{Path: clean, Kind: driver.RootStdlib})
	}
	for _, path := range findKernelRoots(moduleRoot) {
		add(path)
	}
	for _, path := range findStdlibRoots(moduleRoot) {
		add(path)
	}
	return paths
}
