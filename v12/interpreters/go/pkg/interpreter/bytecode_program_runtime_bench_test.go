package interpreter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	goRuntime "runtime"
	"runtime/pprof"
	"testing"

	"able/interpreter-go/pkg/driver"
)

const (
	bytecodeRuntimeBenchTargetEnv   = "ABLE_BENCH_RUNTIME_TARGET"
	bytecodeRuntimeBenchRunFromEnv  = "ABLE_BENCH_RUNTIME_RUN_FROM"
	bytecodeRuntimeBenchArgsJSONEnv = "ABLE_BENCH_RUNTIME_ARGS_JSON"
	bytecodeRuntimeBenchCPUProfEnv  = "ABLE_BENCH_RUNTIME_CPU_PROFILE"
	bytecodeRuntimeBenchMemProfEnv  = "ABLE_BENCH_RUNTIME_MEM_PROFILE"
	bytecodeRuntimeBenchTraceOutEnv = "ABLE_BENCH_RUNTIME_TRACE_OUT"
	bytecodeRuntimeBenchTraceTopEnv = "ABLE_BENCH_RUNTIME_TRACE_TOP"
)

type bytecodeProgramRuntimeBenchConfig struct {
	TargetPath      string
	RunFrom         string
	ProgramArgs     []string
	TraceOutputPath string
	TraceTop        int
}

func loadBytecodeProgramRuntimeBenchConfig() (bytecodeProgramRuntimeBenchConfig, error) {
	var cfg bytecodeProgramRuntimeBenchConfig

	rawTarget := os.Getenv(bytecodeRuntimeBenchTargetEnv)
	if rawTarget == "" {
		return cfg, nil
	}
	targetPath, err := resolveBytecodeProgramRuntimeBenchPath(rawTarget)
	if err != nil {
		return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchTargetEnv, err)
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return cfg, fmt.Errorf("stat %s: %w", targetPath, err)
	}
	if info.IsDir() {
		return cfg, fmt.Errorf("%s must reference a file, got directory %s", bytecodeRuntimeBenchTargetEnv, targetPath)
	}
	cfg.TargetPath = targetPath

	runFrom := os.Getenv(bytecodeRuntimeBenchRunFromEnv)
	if runFrom == "" {
		runFrom, err = os.Getwd()
		if err != nil {
			return cfg, fmt.Errorf("getwd: %w", err)
		}
	}
	runFromPath, err := resolveBytecodeProgramRuntimeBenchPath(runFrom)
	if err != nil {
		return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchRunFromEnv, err)
	}
	info, err = os.Stat(runFromPath)
	if err != nil {
		return cfg, fmt.Errorf("stat %s: %w", runFromPath, err)
	}
	if !info.IsDir() {
		return cfg, fmt.Errorf("%s must reference a directory, got file %s", bytecodeRuntimeBenchRunFromEnv, runFromPath)
	}
	cfg.RunFrom = runFromPath

	rawArgs := os.Getenv(bytecodeRuntimeBenchArgsJSONEnv)
	if rawArgs != "" {
		if err := json.Unmarshal([]byte(rawArgs), &cfg.ProgramArgs); err != nil {
			return cfg, fmt.Errorf("parse %s: %w", bytecodeRuntimeBenchArgsJSONEnv, err)
		}
	}
	if rawTraceOut := os.Getenv(bytecodeRuntimeBenchTraceOutEnv); rawTraceOut != "" {
		traceOutPath, err := resolveBytecodeProgramRuntimeBenchPath(rawTraceOut)
		if err != nil {
			return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchTraceOutEnv, err)
		}
		cfg.TraceOutputPath = traceOutPath
	}
	if rawTraceTop := os.Getenv(bytecodeRuntimeBenchTraceTopEnv); rawTraceTop != "" {
		var traceTop int
		if _, err := fmt.Sscanf(rawTraceTop, "%d", &traceTop); err != nil || traceTop < 0 {
			return cfg, fmt.Errorf("parse %s: expected non-negative integer, got %q", bytecodeRuntimeBenchTraceTopEnv, rawTraceTop)
		}
		cfg.TraceTop = traceTop
	}
	return cfg, nil
}

func resolveBytecodeProgramRuntimeBenchPath(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw), nil
	}
	if abs, err := filepath.Abs(raw); err == nil {
		if _, statErr := os.Stat(abs); statErr == nil {
			return abs, nil
		}
	}
	if root := repositoryRoot(); root != "" {
		candidate := filepath.Join(root, filepath.FromSlash(raw))
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Abs(candidate)
		}
	}
	return filepath.Abs(raw)
}

func chdirBenchRuntime(dir string) (func(), error) {
	prev, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if err := os.Chdir(dir); err != nil {
		return nil, err
	}
	return func() {
		_ = os.Chdir(prev)
	}, nil
}

func startBytecodeProgramRuntimeCPUProfile() (func(), error) {
	path := os.Getenv(bytecodeRuntimeBenchCPUProfEnv)
	if path == "" {
		return func() {}, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create cpu profile: %w", err)
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("start cpu profile: %w", err)
	}
	return func() {
		pprof.StopCPUProfile()
		_ = file.Close()
	}, nil
}

func writeBytecodeProgramRuntimeHeapProfile() error {
	path := os.Getenv(bytecodeRuntimeBenchMemProfEnv)
	if path == "" {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create mem profile: %w", err)
	}
	defer file.Close()
	goRuntime.GC()
	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("write mem profile: %w", err)
	}
	return nil
}

type bytecodeProgramRuntimeTraceReport struct {
	TargetPath  string                `json:"target_path"`
	RunFrom     string                `json:"run_from"`
	ProgramArgs []string              `json:"program_args,omitempty"`
	Trace       BytecodeTraceSnapshot `json:"trace"`
}

func writeBytecodeProgramRuntimeTrace(cfg bytecodeProgramRuntimeBenchConfig, trace BytecodeTraceSnapshot) error {
	if cfg.TraceOutputPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cfg.TraceOutputPath), 0o755); err != nil {
		return fmt.Errorf("mkdir trace output: %w", err)
	}
	file, err := os.Create(cfg.TraceOutputPath)
	if err != nil {
		return fmt.Errorf("create trace output: %w", err)
	}
	defer file.Close()
	report := bytecodeProgramRuntimeTraceReport{
		TargetPath:  cfg.TargetPath,
		RunFrom:     cfg.RunFrom,
		ProgramArgs: append([]string(nil), cfg.ProgramArgs...),
		Trace:       trace,
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("write trace output: %w", err)
	}
	return nil
}

func BenchmarkBytecodeProgramRuntime(b *testing.B) {
	cfg, err := loadBytecodeProgramRuntimeBenchConfig()
	if err != nil {
		b.Fatal(err)
	}
	if cfg.TargetPath == "" {
		b.Skipf("set %s to benchmark a bytecode target", bytecodeRuntimeBenchTargetEnv)
	}

	resumeMemProfile := suspendMemProfileSampling()
	defer resumeMemProfile()

	restoreWD, err := chdirBenchRuntime(cfg.RunFrom)
	if err != nil {
		b.Fatalf("chdir run-from: %v", err)
	}
	defer restoreWD()

	searchPaths, err := buildExecSearchPaths(cfg.TargetPath, cfg.RunFrom, fixtureManifest{})
	if err != nil {
		b.Fatalf("bench search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		b.Fatalf("loader init: %v", err)
	}
	defer loader.Close()

	program, err := loader.Load(cfg.TargetPath)
	if err != nil {
		b.Fatalf("load program: %v", err)
	}

	executor, err := NewExecutorFromEnvironment()
	if err != nil {
		b.Fatalf("executor from environment: %v", err)
	}
	if closer, ok := executor.(interface{ Close() }); ok {
		defer closer.Close()
	}

	interp := NewBytecodeWithExecutor(executor)
	interp.SetArgs(cfg.ProgramArgs)
	registerBenchPrint(interp)

	skipTypecheck := benchSkipTypecheck()
	_, entryEnv, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    skipTypecheck,
		AllowDiagnostics: !skipTypecheck,
	})
	if err != nil {
		b.Fatalf("evaluate program: %v", err)
	}

	env := entryEnv
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	mainValue, err := env.Get("main")
	if err != nil {
		b.Fatalf("lookup main: %v", err)
	}

	goRuntime.GC()
	if _, err := interp.CallFunction(mainValue, nil); err != nil {
		b.Fatalf("warmup call main: %v", err)
	}
	executor.Flush()
	goRuntime.GC()
	interp.ResetBytecodeTrace()
	resumeMemProfile()
	goRuntime.GC()

	stopCPUProfile, err := startBytecodeProgramRuntimeCPUProfile()
	if err != nil {
		b.Fatalf("start runtime cpu profile: %v", err)
	}
	defer stopCPUProfile()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := interp.CallFunction(mainValue, nil); err != nil {
			b.Fatalf("call main: %v", err)
		}
		executor.Flush()
	}
	b.StopTimer()
	if err := writeBytecodeProgramRuntimeTrace(cfg, interp.BytecodeTrace(cfg.TraceTop)); err != nil {
		b.Fatalf("write runtime trace: %v", err)
	}
	if err := writeBytecodeProgramRuntimeHeapProfile(); err != nil {
		b.Fatalf("write runtime heap profile: %v", err)
	}
}

func TestLoadBytecodeProgramRuntimeBenchConfig(t *testing.T) {
	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "main.able")
	if err := os.WriteFile(targetPath, []byte("fn main() {}\n"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}

	runFrom := t.TempDir()
	traceOut := filepath.Join(t.TempDir(), "trace.json")
	argsJSON := `["wordlist.txt","second"]`

	t.Setenv(bytecodeRuntimeBenchTargetEnv, targetPath)
	t.Setenv(bytecodeRuntimeBenchRunFromEnv, runFrom)
	t.Setenv(bytecodeRuntimeBenchArgsJSONEnv, argsJSON)
	t.Setenv(bytecodeRuntimeBenchTraceOutEnv, traceOut)
	t.Setenv(bytecodeRuntimeBenchTraceTopEnv, "7")

	cfg, err := loadBytecodeProgramRuntimeBenchConfig()
	if err != nil {
		t.Fatalf("loadBytecodeProgramRuntimeBenchConfig: %v", err)
	}
	if cfg.TargetPath != targetPath {
		t.Fatalf("target path mismatch: got %q want %q", cfg.TargetPath, targetPath)
	}
	if cfg.RunFrom != runFrom {
		t.Fatalf("run-from mismatch: got %q want %q", cfg.RunFrom, runFrom)
	}
	if len(cfg.ProgramArgs) != 2 || cfg.ProgramArgs[0] != "wordlist.txt" || cfg.ProgramArgs[1] != "second" {
		t.Fatalf("unexpected program args: %#v", cfg.ProgramArgs)
	}
	if cfg.TraceOutputPath != traceOut {
		t.Fatalf("trace output mismatch: got %q want %q", cfg.TraceOutputPath, traceOut)
	}
	if cfg.TraceTop != 7 {
		t.Fatalf("trace top mismatch: got %d want 7", cfg.TraceTop)
	}
}
