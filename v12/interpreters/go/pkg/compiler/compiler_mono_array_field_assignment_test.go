package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCompilerMonoArrayMetadataAssignmentAvoidsLegacyWrapperFields(t *testing.T) {
	result := compileAndBuildCanonicalStdlibSource(t, "ablec-mono-array-field-assign-", strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{Array}",
		"",
		"fn touch(values: Array f64) -> i32 {",
		"  values.length = values.length",
		"  values.capacity = values.capacity",
		"  values.storage_handle = values.storage_handle",
		"  values.length + values.capacity + (values.storage_handle as i32)",
		"}",
		"",
		"fn main() -> i32 {",
		"  values: Array f64 := Array.with_capacity(4)",
		"  touch(values)",
		"}",
		"",
	}, "\n"))

	compiledSrc := string(result.Files["compiled.go"])
	for _, fragment := range []string{"self.Length", "self.Capacity", "self.Storage_handle"} {
		if strings.Contains(compiledSrc, fragment) {
			t.Fatalf("expected mono-array metadata assignment lowering to avoid legacy wrapper field %q:\n%s", fragment, compiledSrc)
		}
	}
}

func TestCompilerMonoArrayMetadataAssignmentExecutes(t *testing.T) {
	stdout := strings.TrimSpace(compileAndRunExecSourceWithOptions(t, "ablec-mono-array-field-assign-exec-", strings.Join([]string{
		"package demo",
		"",
		"import able.kernel.{Array}",
		"import able.collections.array",
		"",
		"fn main() {",
		"  values: Array f64 := Array.with_capacity(2)",
		"  values.push(1.0)",
		"  values.length = 3",
		"  values.capacity = 5",
		"  values.storage_handle = 99_i64",
		"  print(values.len())",
		"  print(\" \")",
		"  print(values.capacity())",
		"  print(\" \")",
		"  values.get(2) match {",
		"    case nil => print(-1_i32)",
		"    case value: f64 => print(value as i32)",
		"  }",
		"  print(\" \")",
		"  print(values.storage_handle as i32)",
		"}",
		"",
	}, "\n"), Options{
		PackageName:            "main",
		EmitMain:               true,
		ExperimentalMonoArrays: true,
	}))
	if strings.Join(strings.Fields(stdout), " ") != "3 5 0 0" {
		t.Fatalf("expected mono-array metadata assignment program to print values 3 5 0 0, got %q", stdout)
	}
}

func TestCompilerExperimentalMonoArraysMatrixMultiplyBuilds(t *testing.T) {
	sourcePath := filepath.Join(repositoryRoot(), "v12", "examples", "benchmarks", "matrixmultiply.able")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read matrixmultiply benchmark: %v", err)
	}

	result := compileSourceWithCanonicalStdlibPaths(t, string(source))

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-matrixmultiply-build-artifacts-")
	outputDir := filepath.Join(workDir, "out")
	if err := result.Write(outputDir); err != nil {
		t.Fatalf("write output: %v", err)
	}

	build := exec.Command("go", "test", "-run", "^$", ".")
	build.Dir = outputDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if runtime.GOOS == "windows" {
		build.Env = withEnv(build.Env, "GOFLAGS", "")
	}
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go test build failed for matrixmultiply benchmark: %v\n%s", err, string(output))
	}
}
