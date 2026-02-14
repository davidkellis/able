package compiler

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"testing"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
)

const (
	compilerBenchFixtureEnv     = "ABLE_COMPILER_BENCH_FIXTURE"
	defaultCompilerBenchFixture = "v12/fixtures/exec/07_09_bytecode_iterator_yield"
)

func BenchmarkCompilerExecFixtureBinary(b *testing.B) {
	if _, err := exec.LookPath("go"); err != nil {
		b.Skip("go toolchain not available")
	}

	dir := resolveCompilerBenchFixtureDir(b)
	manifest, err := interpreter.LoadFixtureManifest(dir)
	if err != nil {
		b.Fatalf("read manifest: %v", err)
	}
	if shouldSkipTarget(manifest.SkipTargets, "go") {
		b.Skip("fixture skipped for go target")
	}
	if manifest.Expect.TypecheckDiagnostics != nil && len(manifest.Expect.TypecheckDiagnostics) > 0 {
		b.Skip("diagnostic-only fixture is not suitable for runtime baseline")
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		b.Fatalf("module root: %v", err)
	}
	tmpRoot := filepath.Join(moduleRoot, "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		b.Fatalf("mkdir tmp: %v", err)
	}
	workDir, err := os.MkdirTemp(tmpRoot, "ablec-bench-")
	if err != nil {
		b.Fatalf("temp dir: %v", err)
	}
	b.Cleanup(func() { _ = os.RemoveAll(workDir) })

	binPath, env, err := buildCompilerBenchmarkBinary(workDir, dir, manifest)
	if err != nil {
		b.Fatalf("prepare compiled benchmark binary: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binPath)
		cmd.Env = env
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Run(); err != nil {
			failCmd := exec.Command(binPath)
			failCmd.Env = env
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			failCmd.Stdout = &stdout
			failCmd.Stderr = &stderr
			_ = failCmd.Run()
			b.Fatalf("compiled benchmark run failed: %v\nstdout=%q\nstderr=%q", err, stdout.String(), stderr.String())
		}
	}
}

func resolveCompilerBenchFixtureDir(b *testing.B) string {
	b.Helper()
	root := repositoryRoot()
	if root == "" {
		b.Fatalf("repository root not found")
	}
	if raw := os.Getenv(compilerBenchFixtureEnv); raw != "" {
		if filepath.IsAbs(raw) {
			return raw
		}
		return filepath.Join(root, filepath.FromSlash(raw))
	}
	return filepath.Join(root, filepath.FromSlash(defaultCompilerBenchFixture))
}

func buildCompilerBenchmarkBinary(workDir string, fixtureDir string, manifest interpreter.FixtureManifest) (string, []string, error) {
	entry := manifest.Entry
	if entry == "" {
		entry = "main.able"
	}
	entryPath := filepath.Join(fixtureDir, entry)

	searchPaths, err := buildExecSearchPaths(entryPath, fixtureDir, manifest)
	if err != nil {
		return "", nil, err
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		return "", nil, err
	}
	defer loader.Close()

	program, err := loader.Load(entryPath)
	if err != nil {
		return "", nil, err
	}

	comp := New(Options{PackageName: "main"})
	result, err := comp.Compile(program)
	if err != nil {
		return "", nil, err
	}
	if err := result.Write(workDir); err != nil {
		return "", nil, err
	}

	harness := compilerHarnessSource(entryPath, searchPaths, manifest.Executor)
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte(harness), 0o600); err != nil {
		return "", nil, err
	}

	moduleRoot, err := filepath.Abs(filepath.Join(".", "..", ".."))
	if err != nil {
		return "", nil, err
	}
	binPath := filepath.Join(workDir, "compiled-bench")
	if goruntime.GOOS == "windows" {
		binPath += ".exe"
	}
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = workDir
	build.Env = withEnv(os.Environ(), "GOCACHE", compilerExecGocache(moduleRoot))
	if output, err := build.CombinedOutput(); err != nil {
		return "", nil, &benchmarkBuildError{cause: err, output: string(output)}
	}

	return binPath, applyFixtureEnv(os.Environ(), manifest.Env), nil
}

type benchmarkBuildError struct {
	cause  error
	output string
}

func (e *benchmarkBuildError) Error() string {
	if e == nil {
		return "benchmark build failed"
	}
	if e.output == "" {
		return "benchmark build failed: " + e.cause.Error()
	}
	return "benchmark build failed: " + e.cause.Error() + "\n" + e.output
}
