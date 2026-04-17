package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func compileAndBuildStdlibSource(t *testing.T, tempPrefix string, source string) *Result {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping stdlib compiler build regression in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	result := compileSourceWithStdlibPaths(t, source)
	moduleRoot, workDir := compilerTestWorkDir(t, tempPrefix)
	outputDir := filepath.Join(workDir, "out")
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	build := exec.Command("go", "test", "-run", "^$", ".")
	build.Dir = outputDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go test build failed: %v\n%s", err, string(output))
	}
	return result
}

func TestCompilerReceiverRefinementKeepsDeclaredGenericBindingName(t *testing.T) {
	_ = compileAndBuildStdlibSource(t, "ablec-receiver-refinement-", strings.Join([]string{
		"package demo",
		"",
		"import able.collections.deque.{Deque}",
		"",
		"fn main() -> void {",
		"  deque := Deque.new()",
		"  i: i32 = 0",
		"  while i < 3 {",
		"    deque.push_back(i)",
		"    i = i + 1",
		"  }",
		"  _ = deque.len()",
		"}",
		"",
	}, "\n"))
}
