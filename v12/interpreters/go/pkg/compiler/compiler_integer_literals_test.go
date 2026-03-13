package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"able/interpreter-go/pkg/driver"
)

func TestCompilerBuildsLargeI128AndU128Literals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large integer literal build test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-intlit")

	source := `fn main() -> i32 {
  min := -9223372036854775808_i128
  max := 9223372036854775807_i128
  unsigned_max := 18446744073709551615_i128
  huge_unsigned := 340282366920938463463374607431768211455_u128
  if min < max && unsigned_max > 0_i128 && huge_unsigned > 0_u128 { 0 } else { 1 }
}
`
	entryPath := filepath.Join(workDir, "app.able")
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	loader, err := driver.NewLoader(nil)
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
}
