package interpreter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	goRuntime "runtime"
	"runtime/pprof"
	"testing"
	"time"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/runtime"
)

const (
	bytecodeRuntimeBenchTargetEnv   = "ABLE_BENCH_RUNTIME_TARGET"
	bytecodeRuntimeBenchRunFromEnv  = "ABLE_BENCH_RUNTIME_RUN_FROM"
	bytecodeRuntimeBenchArgsJSONEnv = "ABLE_BENCH_RUNTIME_ARGS_JSON"
	bytecodeRuntimeBenchCPUProfEnv  = "ABLE_BENCH_RUNTIME_CPU_PROFILE"
	bytecodeRuntimeBenchLoadCPUEnv  = "ABLE_BENCH_RUNTIME_LOAD_CPU_PROFILE"
	bytecodeRuntimeBenchMemProfEnv  = "ABLE_BENCH_RUNTIME_MEM_PROFILE"
	bytecodeRuntimeBenchStatsOutEnv = "ABLE_BENCH_RUNTIME_STATS_OUT"
	bytecodeRuntimeBenchPhaseOutEnv = "ABLE_BENCH_RUNTIME_PHASE_STATS_OUT"
	bytecodeRuntimeBenchTraceOutEnv = "ABLE_BENCH_RUNTIME_TRACE_OUT"
	bytecodeRuntimeBenchTraceTopEnv = "ABLE_BENCH_RUNTIME_TRACE_TOP"
	// bytecodeRuntimeBenchArrayOwnershipOutEnv enables the release-disabled
	// frame-ownership observer for one bounded loader-driven program run.
	bytecodeRuntimeBenchArrayOwnershipOutEnv = "ABLE_BENCH_RUNTIME_ARRAY_OWNERSHIP_OUT"
	// bytecodeRuntimeBenchRetentionOutEnv enables the opt-in, fresh-process
	// retention probe below. It deliberately shares the runtime benchmark
	// target/run-from setup so every source benchmark can use the same bounded
	// execution path without adding diagnostic work to normal VM runs.
	bytecodeRuntimeBenchRetentionOutEnv         = "ABLE_BENCH_RUNTIME_RETENTION_OUT"
	bytecodeRuntimeBenchRetentionHeapProfileEnv = "ABLE_BENCH_RUNTIME_RETENTION_HEAP_PROFILE"
)

type bytecodeProgramRuntimeBenchConfig struct {
	TargetPath               string
	RunFrom                  string
	ProgramArgs              []string
	StatsOutputPath          string
	PhaseOutputPath          string
	LoadCPUProfilePath       string
	TraceOutputPath          string
	RetentionOutputPath      string
	RetentionHeapProfilePath string
	ArrayOwnershipOutputPath string
	TraceTop                 int
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
	if rawStatsOut := os.Getenv(bytecodeRuntimeBenchStatsOutEnv); rawStatsOut != "" {
		statsOutPath, err := resolveBytecodeProgramRuntimeBenchPath(rawStatsOut)
		if err != nil {
			return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchStatsOutEnv, err)
		}
		cfg.StatsOutputPath = statsOutPath
	}
	if rawPhaseOut := os.Getenv(bytecodeRuntimeBenchPhaseOutEnv); rawPhaseOut != "" {
		phaseOutPath, err := resolveBytecodeProgramRuntimeBenchPath(rawPhaseOut)
		if err != nil {
			return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchPhaseOutEnv, err)
		}
		cfg.PhaseOutputPath = phaseOutPath
	}
	if rawLoadCPU := os.Getenv(bytecodeRuntimeBenchLoadCPUEnv); rawLoadCPU != "" {
		loadCPUPath, err := resolveBytecodeProgramRuntimeBenchPath(rawLoadCPU)
		if err != nil {
			return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchLoadCPUEnv, err)
		}
		cfg.LoadCPUProfilePath = loadCPUPath
	}
	if rawTraceOut := os.Getenv(bytecodeRuntimeBenchTraceOutEnv); rawTraceOut != "" {
		traceOutPath, err := resolveBytecodeProgramRuntimeBenchPath(rawTraceOut)
		if err != nil {
			return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchTraceOutEnv, err)
		}
		cfg.TraceOutputPath = traceOutPath
	}
	if rawRetentionOut := os.Getenv(bytecodeRuntimeBenchRetentionOutEnv); rawRetentionOut != "" {
		retentionOutPath, err := resolveBytecodeProgramRuntimeBenchPath(rawRetentionOut)
		if err != nil {
			return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchRetentionOutEnv, err)
		}
		cfg.RetentionOutputPath = retentionOutPath
	}
	if rawRetentionHeapProfile := os.Getenv(bytecodeRuntimeBenchRetentionHeapProfileEnv); rawRetentionHeapProfile != "" {
		retentionHeapProfilePath, err := resolveBytecodeProgramRuntimeBenchPath(rawRetentionHeapProfile)
		if err != nil {
			return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchRetentionHeapProfileEnv, err)
		}
		cfg.RetentionHeapProfilePath = retentionHeapProfilePath
	}
	if rawOwnershipOut := os.Getenv(bytecodeRuntimeBenchArrayOwnershipOutEnv); rawOwnershipOut != "" {
		ownershipOutPath, err := resolveBytecodeProgramRuntimeBenchPath(rawOwnershipOut)
		if err != nil {
			return cfg, fmt.Errorf("resolve %s: %w", bytecodeRuntimeBenchArrayOwnershipOutEnv, err)
		}
		cfg.ArrayOwnershipOutputPath = ownershipOutPath
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
	return startBytecodeProgramRuntimeCPUProfileAt(path)
}

func startBytecodeProgramRuntimeCPUProfileAt(path string) (func(), error) {
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

type bytecodeProgramRuntimeStatsReport struct {
	TargetPath  string                `json:"target_path"`
	RunFrom     string                `json:"run_from"`
	ProgramArgs []string              `json:"program_args,omitempty"`
	Stats       BytecodeStatsSnapshot `json:"stats"`
}

type bytecodeProgramRuntimePhase struct {
	Name       string `json:"name"`
	DurationNS int64  `json:"duration_ns"`
}

type bytecodeProgramRuntimeLoadSubphase struct {
	Name        string `json:"name"`
	DurationNS  int64  `json:"duration_ns"`
	Samples     int    `json:"samples"`
	SourceBytes int    `json:"source_bytes,omitempty"`
}

type bytecodeProgramRuntimePhaseReport struct {
	TargetPath              string                               `json:"target_path"`
	RunFrom                 string                               `json:"run_from"`
	ProgramArgs             []string                             `json:"program_args,omitempty"`
	BenchmarkEntryToReadyNS int64                                `json:"benchmark_entry_to_ready_ns"`
	Phases                  []bytecodeProgramRuntimePhase        `json:"phases"`
	LoadSubphases           []bytecodeProgramRuntimeLoadSubphase `json:"load_subphases,omitempty"`
}

// bytecodeProgramRuntimeHeapStats captures only post-GC live heap state. The
// process-monotonic allocation counters in runtime.MemStats do not answer the
// retention question, so keep this report intentionally small and focused.
type bytecodeProgramRuntimeHeapStats struct {
	HeapAlloc   uint64 `json:"heap_alloc"`
	HeapInuse   uint64 `json:"heap_inuse"`
	HeapObjects uint64 `json:"heap_objects"`
	NumGC       uint32 `json:"num_gc"`
}

type bytecodeProgramRuntimeMainAllocationStats struct {
	AllocatedBytes uint64 `json:"allocated_bytes"`
	Allocations    uint64 `json:"allocations"`
	Frees          uint64 `json:"frees"`
	GCCount        uint32 `json:"gc_count"`
}

type bytecodeProgramRuntimeRetentionReport struct {
	TargetPath                  string                                     `json:"target_path"`
	RunFrom                     string                                     `json:"run_from"`
	ProgramArgs                 []string                                   `json:"program_args,omitempty"`
	MainDurationNS              int64                                      `json:"main_duration_ns"`
	MainAllocation              bytecodeProgramRuntimeMainAllocationStats  `json:"main_allocation"`
	StringStats                 BytecodeStringStatsSnapshot                `json:"string_stats"`
	BeforeFinalGC               runtime.ArrayStoreStats                    `json:"before_final_gc"`
	AfterFinalGC                runtime.ArrayStoreStats                    `json:"after_final_gc"`
	HeapAfterFinalGC            bytecodeProgramRuntimeHeapStats            `json:"heap_after_final_gc"`
	DynamicIntegerBoxCacheSize  map[string]int                             `json:"dynamic_integer_box_cache_size"`
	DynamicIntegerBoxCacheReuse map[string]bytecodeDynamicIntBoxCacheReuse `json:"dynamic_integer_box_cache_reuse,omitempty"`
}

type bytecodeProgramRuntimeArrayOwnershipSnapshot struct {
	Created        int            `json:"created"`
	Transferred    int            `json:"transferred"`
	PublicReturned int            `json:"public_returned"`
	Escaped        int            `json:"escaped"`
	FrameLocal     int            `json:"frame_local"`
	ErrorUnwound   int            `json:"error_unwound"`
	Escapes        map[string]int `json:"escapes,omitempty"`
}

type bytecodeProgramRuntimeArrayOwnershipMarker struct {
	Event     string                                       `json:"event"`
	Ownership bytecodeProgramRuntimeArrayOwnershipSnapshot `json:"ownership"`
}

type bytecodeProgramRuntimeArrayOwnershipReport struct {
	TargetPath  string                                       `json:"target_path"`
	RunFrom     string                                       `json:"run_from"`
	ProgramArgs []string                                     `json:"program_args,omitempty"`
	Markers     []bytecodeProgramRuntimeArrayOwnershipMarker `json:"markers"`
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

func writeBytecodeProgramRuntimeStats(cfg bytecodeProgramRuntimeBenchConfig, stats BytecodeStatsSnapshot) error {
	if cfg.StatsOutputPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cfg.StatsOutputPath), 0o755); err != nil {
		return fmt.Errorf("mkdir stats output: %w", err)
	}
	file, err := os.Create(cfg.StatsOutputPath)
	if err != nil {
		return fmt.Errorf("create stats output: %w", err)
	}
	defer file.Close()
	report := bytecodeProgramRuntimeStatsReport{
		TargetPath:  cfg.TargetPath,
		RunFrom:     cfg.RunFrom,
		ProgramArgs: append([]string(nil), cfg.ProgramArgs...),
		Stats:       stats,
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("write stats output: %w", err)
	}
	return nil
}

func writeBytecodeProgramRuntimePhases(cfg bytecodeProgramRuntimeBenchConfig, entryToReady time.Duration, phases []bytecodeProgramRuntimePhase, loadSubphases []bytecodeProgramRuntimeLoadSubphase) error {
	if cfg.PhaseOutputPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cfg.PhaseOutputPath), 0o755); err != nil {
		return fmt.Errorf("mkdir phase output: %w", err)
	}
	file, err := os.Create(cfg.PhaseOutputPath)
	if err != nil {
		return fmt.Errorf("create phase output: %w", err)
	}
	defer file.Close()
	report := bytecodeProgramRuntimePhaseReport{
		TargetPath:              cfg.TargetPath,
		RunFrom:                 cfg.RunFrom,
		ProgramArgs:             append([]string(nil), cfg.ProgramArgs...),
		BenchmarkEntryToReadyNS: entryToReady.Nanoseconds(),
		Phases:                  append([]bytecodeProgramRuntimePhase(nil), phases...),
		LoadSubphases:           append([]bytecodeProgramRuntimeLoadSubphase(nil), loadSubphases...),
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("write phase output: %w", err)
	}
	return nil
}

func writeBytecodeProgramRuntimeRetention(cfg bytecodeProgramRuntimeBenchConfig, mainDuration time.Duration, mainAllocation bytecodeProgramRuntimeMainAllocationStats, stringStats BytecodeStringStatsSnapshot, beforeFinalGC runtime.ArrayStoreStats, afterFinalGC runtime.ArrayStoreStats, heapStats bytecodeProgramRuntimeHeapStats, dynamicIntegerBoxCacheSize map[string]int, dynamicIntegerBoxCacheReuse map[string]bytecodeDynamicIntBoxCacheReuse) error {
	if cfg.RetentionOutputPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cfg.RetentionOutputPath), 0o755); err != nil {
		return fmt.Errorf("mkdir retention output: %w", err)
	}
	file, err := os.Create(cfg.RetentionOutputPath)
	if err != nil {
		return fmt.Errorf("create retention output: %w", err)
	}
	defer file.Close()
	report := bytecodeProgramRuntimeRetentionReport{
		TargetPath:                  cfg.TargetPath,
		RunFrom:                     cfg.RunFrom,
		ProgramArgs:                 append([]string(nil), cfg.ProgramArgs...),
		MainDurationNS:              mainDuration.Nanoseconds(),
		MainAllocation:              mainAllocation,
		StringStats:                 stringStats,
		BeforeFinalGC:               beforeFinalGC,
		AfterFinalGC:                afterFinalGC,
		HeapAfterFinalGC:            heapStats,
		DynamicIntegerBoxCacheSize:  dynamicIntegerBoxCacheSize,
		DynamicIntegerBoxCacheReuse: dynamicIntegerBoxCacheReuse,
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("write retention output: %w", err)
	}
	return nil
}

// writeBytecodeProgramRuntimeRetentionHeapProfile records only the heap after
// the retention probe has returned its interpreter scope and forced its final
// collections. It is test-harness-only and must stay separate from the normal
// runtime benchmark heap profile, which intentionally captures a live VM.
func writeBytecodeProgramRuntimeRetentionHeapProfile(cfg bytecodeProgramRuntimeBenchConfig) error {
	if cfg.RetentionHeapProfilePath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cfg.RetentionHeapProfilePath), 0o755); err != nil {
		return fmt.Errorf("mkdir retention heap profile output: %w", err)
	}
	file, err := os.Create(cfg.RetentionHeapProfilePath)
	if err != nil {
		return fmt.Errorf("create retention heap profile: %w", err)
	}
	defer file.Close()
	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("write retention heap profile: %w", err)
	}
	return nil
}

func bytecodeProgramRuntimeArrayOwnershipReportSnapshot(snapshot bytecodeArrayOwnershipSnapshot) bytecodeProgramRuntimeArrayOwnershipSnapshot {
	report := bytecodeProgramRuntimeArrayOwnershipSnapshot{
		Created:        snapshot.Created,
		Transferred:    snapshot.Transferred,
		PublicReturned: snapshot.PublicReturned,
		Escaped:        snapshot.Escaped,
		FrameLocal:     snapshot.FrameLocal,
		ErrorUnwound:   snapshot.ErrorUnwound,
	}
	if len(snapshot.Escapes) == 0 {
		return report
	}
	report.Escapes = make(map[string]int, len(snapshot.Escapes))
	for reason, count := range snapshot.Escapes {
		report.Escapes[reason.String()] += count
	}
	return report
}

func writeBytecodeProgramRuntimeArrayOwnership(cfg bytecodeProgramRuntimeBenchConfig, markers []bytecodeProgramRuntimeArrayOwnershipMarker) error {
	if cfg.ArrayOwnershipOutputPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cfg.ArrayOwnershipOutputPath), 0o755); err != nil {
		return fmt.Errorf("mkdir array ownership output: %w", err)
	}
	file, err := os.Create(cfg.ArrayOwnershipOutputPath)
	if err != nil {
		return fmt.Errorf("create array ownership output: %w", err)
	}
	defer file.Close()
	report := bytecodeProgramRuntimeArrayOwnershipReport{
		TargetPath:  cfg.TargetPath,
		RunFrom:     cfg.RunFrom,
		ProgramArgs: append([]string(nil), cfg.ProgramArgs...),
		Markers:     markers,
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("write array ownership output: %w", err)
	}
	return nil
}

// TestBytecodeProgramRuntimeRetention is an opt-in, one-program-per-process
// probe. The pre-GC snapshot is taken while the interpreter still owns the
// loaded program; after the helper returns, three forced collections allow
// token-only Array cleanup callbacks to run before measuring retained state.
// Keeping this in the benchmark harness makes it reproducible under the same
// OOM guardrails as wall-clock and allocation readings.
func TestBytecodeProgramRuntimeRetention(t *testing.T) {
	cfg, err := loadBytecodeProgramRuntimeBenchConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TargetPath == "" || cfg.RetentionOutputPath == "" {
		t.Skipf("set %s and %s to probe post-GC ArrayStore retention", bytecodeRuntimeBenchTargetEnv, bytecodeRuntimeBenchRetentionOutEnv)
	}

	var mainAllocation bytecodeProgramRuntimeMainAllocationStats
	var stringStats BytecodeStringStatsSnapshot
	beforeFinalGC, mainDuration, err := runBytecodeProgramRuntimeRetention(cfg, &mainAllocation, &stringStats)
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		goRuntime.GC()
	}
	afterFinalGC := runtime.ArrayStoreStatsSnapshot()
	var memStats goRuntime.MemStats
	goRuntime.ReadMemStats(&memStats)
	if err := writeBytecodeProgramRuntimeRetentionHeapProfile(cfg); err != nil {
		t.Fatal(err)
	}
	if err := writeBytecodeProgramRuntimeRetention(cfg, mainDuration, mainAllocation, stringStats, beforeFinalGC, afterFinalGC, bytecodeProgramRuntimeHeapStats{
		HeapAlloc:   memStats.HeapAlloc,
		HeapInuse:   memStats.HeapInuse,
		HeapObjects: memStats.HeapObjects,
		NumGC:       memStats.NumGC,
	}, bytecodeDynamicIntBoxCacheEntriesForTest(), bytecodeDynamicIntBoxCacheReuseForTest()); err != nil {
		t.Fatal(err)
	}
}

// TestBytecodeProgramRuntimeArrayOwnership is an opt-in, one-program-per-
// process observer run. It uses the ordinary loader and cached bytecode-module
// path, resets after program setup, and snapshots at every program print plus
// the main-call boundaries. The observer never releases an ArrayStore lease.
func TestBytecodeProgramRuntimeArrayOwnership(t *testing.T) {
	cfg, err := loadBytecodeProgramRuntimeBenchConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TargetPath == "" || cfg.ArrayOwnershipOutputPath == "" {
		t.Skipf("set %s and %s to probe bytecode Array ownership", bytecodeRuntimeBenchTargetEnv, bytecodeRuntimeBenchArrayOwnershipOutEnv)
	}

	markers, err := runBytecodeProgramRuntimeArrayOwnership(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeBytecodeProgramRuntimeArrayOwnership(cfg, markers); err != nil {
		t.Fatal(err)
	}
}

func runBytecodeProgramRuntimeRetention(cfg bytecodeProgramRuntimeBenchConfig, mainAllocation *bytecodeProgramRuntimeMainAllocationStats, stringStats *BytecodeStringStatsSnapshot) (runtime.ArrayStoreStats, time.Duration, error) {
	restoreWD, err := chdirBenchRuntime(cfg.RunFrom)
	if err != nil {
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("chdir run-from: %w", err)
	}
	defer restoreWD()

	searchPaths, err := buildExecSearchPaths(cfg.TargetPath, cfg.RunFrom, fixtureManifest{})
	if err != nil {
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("bench search paths: %w", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("loader init: %w", err)
	}
	defer loader.Close()
	program, err := loader.Load(cfg.TargetPath)
	if err != nil {
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("load program: %w", err)
	}

	executor, err := NewExecutorFromEnvironment()
	if err != nil {
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("executor from environment: %w", err)
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
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("evaluate program: %w", err)
	}
	env := entryEnv
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	mainValue, err := env.Get("main")
	if err != nil {
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("lookup main: %w", err)
	}
	// Exclude loading/lowering setup: the opt-in reuse snapshot represents only
	// the benchmark program's main execution.
	bytecodeResetDynamicIntBoxCacheReuseForTest()
	interp.ResetBytecodeStringStats()
	stopCPUProfile, err := startBytecodeProgramRuntimeCPUProfile()
	if err != nil {
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("start runtime cpu profile: %w", err)
	}
	var beforeMainMem goRuntime.MemStats
	goRuntime.ReadMemStats(&beforeMainMem)
	mainStart := time.Now()
	_, callErr := interp.CallFunction(mainValue, nil)
	mainDuration := time.Since(mainStart)
	stopCPUProfile()
	if callErr != nil {
		return runtime.ArrayStoreStats{}, 0, fmt.Errorf("call main: %w", callErr)
	}
	executor.Flush()
	if stringStats != nil {
		*stringStats = interp.BytecodeStringStats()
	}
	if mainAllocation != nil {
		var afterMainMem goRuntime.MemStats
		goRuntime.ReadMemStats(&afterMainMem)
		*mainAllocation = bytecodeProgramRuntimeMainAllocationStats{
			AllocatedBytes: afterMainMem.TotalAlloc - beforeMainMem.TotalAlloc,
			Allocations:    afterMainMem.Mallocs - beforeMainMem.Mallocs,
			Frees:          afterMainMem.Frees - beforeMainMem.Frees,
			GCCount:        afterMainMem.NumGC - beforeMainMem.NumGC,
		}
	}
	return runtime.ArrayStoreStatsSnapshot(), mainDuration, nil
}

func runBytecodeProgramRuntimeArrayOwnership(cfg bytecodeProgramRuntimeBenchConfig) ([]bytecodeProgramRuntimeArrayOwnershipMarker, error) {
	restoreWD, err := chdirBenchRuntime(cfg.RunFrom)
	if err != nil {
		return nil, fmt.Errorf("chdir run-from: %w", err)
	}
	defer restoreWD()

	searchPaths, err := buildExecSearchPaths(cfg.TargetPath, cfg.RunFrom, fixtureManifest{})
	if err != nil {
		return nil, fmt.Errorf("bench search paths: %w", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		return nil, fmt.Errorf("loader init: %w", err)
	}
	defer loader.Close()
	program, err := loader.Load(cfg.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("load program: %w", err)
	}

	executor, err := NewExecutorFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("executor from environment: %w", err)
	}
	if closer, ok := executor.(interface{ Close() }); ok {
		defer closer.Close()
	}
	interp := NewBytecodeWithExecutor(executor)
	interp.SetArgs(cfg.ProgramArgs)
	profile := interp.enableBytecodeArrayOwnershipProfile()
	defer interp.disableBytecodeArrayOwnershipProfile()

	markers := make([]bytecodeProgramRuntimeArrayOwnershipMarker, 0, 8)
	record := func(event string) {
		markers = append(markers, bytecodeProgramRuntimeArrayOwnershipMarker{
			Event:     event,
			Ownership: bytecodeProgramRuntimeArrayOwnershipReportSnapshot(profile.snapshot()),
		})
	}
	printCount := 0
	interp.GlobalEnvironment().Define("print", runtime.NativeFunctionValue{
		Name:  "print",
		Arity: 1,
		Impl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {
			printCount++
			record(fmt.Sprintf("print-%02d", printCount))
			return runtime.NilValue{}, nil
		},
	})

	skipTypecheck := benchSkipTypecheck()
	_, entryEnv, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    skipTypecheck,
		AllowDiagnostics: !skipTypecheck,
	})
	if err != nil {
		return nil, fmt.Errorf("evaluate program: %w", err)
	}
	profile.reset()
	markers = markers[:0]
	printCount = 0
	record("before-main")

	env := entryEnv
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	mainValue, err := env.Get("main")
	if err != nil {
		return nil, fmt.Errorf("lookup main: %w", err)
	}
	if err := runBytecodeProgramRuntimeOwnershipMain(interp, env, mainValue); err != nil {
		return nil, err
	}
	executor.Flush()
	record("after-main")
	return markers, nil
}

func runBytecodeProgramRuntimeOwnershipMain(interp *Interpreter, env *runtime.Environment, mainValue runtime.Value) error {
	if interp == nil {
		return fmt.Errorf("bytecode ownership main: interpreter is nil")
	}
	if mainValue == nil {
		return fmt.Errorf("bytecode ownership main: main is nil")
	}
	callProgram, err := interp.lowerExpressionToBytecode(ast.Call("main"))
	if err != nil {
		return fmt.Errorf("lower bytecode main call: %w", err)
	}
	vm := interp.acquireBytecodeVM(env)
	defer interp.releaseBytecodeVM(vm)
	if _, err := vm.run(callProgram); err != nil {
		return fmt.Errorf("bytecode call main: %w", err)
	}
	return nil
}

func BenchmarkBytecodeProgramRuntime(b *testing.B) {
	benchmarkEntry := time.Now()
	cfg, err := loadBytecodeProgramRuntimeBenchConfig()
	if err != nil {
		b.Fatal(err)
	}
	if cfg.TargetPath == "" {
		b.Skipf("set %s to benchmark a bytecode target", bytecodeRuntimeBenchTargetEnv)
	}
	phaseEnabled := cfg.PhaseOutputPath != ""
	phases := make([]bytecodeProgramRuntimePhase, 0, 9)
	recordPhase := func(name string, start time.Time) {
		if phaseEnabled {
			phases = append(phases, bytecodeProgramRuntimePhase{Name: name, DurationNS: time.Since(start).Nanoseconds()})
		}
	}

	resumeMemProfile := suspendMemProfileSampling()
	defer resumeMemProfile()

	phaseStart := time.Now()
	restoreWD, err := chdirBenchRuntime(cfg.RunFrom)
	if err != nil {
		b.Fatalf("chdir run-from: %v", err)
	}
	defer restoreWD()
	recordPhase("working_directory", phaseStart)

	phaseStart = time.Now()
	searchPaths, err := buildExecSearchPaths(cfg.TargetPath, cfg.RunFrom, fixtureManifest{})
	if err != nil {
		b.Fatalf("bench search paths: %v", err)
	}
	recordPhase("search_paths", phaseStart)
	phaseStart = time.Now()
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		b.Fatalf("loader init: %v", err)
	}
	defer loader.Close()
	loadSubphaseByName := make(map[string]*bytecodeProgramRuntimeLoadSubphase, 3)
	if phaseEnabled {
		loader.SetPhaseObserver(func(sample driver.LoaderPhaseSample) {
			name := string(sample.Phase)
			aggregate := loadSubphaseByName[name]
			if aggregate == nil {
				aggregate = &bytecodeProgramRuntimeLoadSubphase{Name: name}
				loadSubphaseByName[name] = aggregate
			}
			aggregate.DurationNS += sample.Duration.Nanoseconds()
			aggregate.Samples++
			aggregate.SourceBytes += sample.SourceBytes
		})
	}
	recordPhase("loader_init", phaseStart)

	phaseStart = time.Now()
	stopLoadCPUProfile, err := startBytecodeProgramRuntimeCPUProfileAt(cfg.LoadCPUProfilePath)
	if err != nil {
		b.Fatalf("start loader cpu profile: %v", err)
	}
	program, err := loader.Load(cfg.TargetPath)
	stopLoadCPUProfile()
	if err != nil {
		b.Fatalf("load program: %v", err)
	}
	recordPhase("program_load", phaseStart)

	phaseStart = time.Now()
	executor, err := NewExecutorFromEnvironment()
	if err != nil {
		b.Fatalf("executor from environment: %v", err)
	}
	if closer, ok := executor.(interface{ Close() }); ok {
		defer closer.Close()
	}
	recordPhase("executor_init", phaseStart)

	phaseStart = time.Now()
	interp := NewBytecodeWithExecutor(executor)
	interp.SetArgs(cfg.ProgramArgs)
	registerBenchPrint(interp)
	recordPhase("interpreter_init", phaseStart)

	skipTypecheck := benchSkipTypecheck()
	phaseStart = time.Now()
	_, entryEnv, _, err := interp.EvaluateProgram(program, ProgramEvaluationOptions{
		SkipTypecheck:    skipTypecheck,
		AllowDiagnostics: !skipTypecheck,
	})
	if err != nil {
		b.Fatalf("evaluate program: %v", err)
	}
	recordPhase("program_evaluation", phaseStart)

	phaseStart = time.Now()
	env := entryEnv
	if env == nil {
		env = interp.GlobalEnvironment()
	}
	mainValue, err := env.Get("main")
	if err != nil {
		b.Fatalf("lookup main: %v", err)
	}
	recordPhase("entry_lookup", phaseStart)

	phaseStart = time.Now()
	goRuntime.GC()
	recordPhase("prewarm_gc", phaseStart)
	phaseStart = time.Now()
	if _, err := interp.CallFunction(mainValue, nil); err != nil {
		b.Fatalf("warmup call main: %v", err)
	}
	executor.Flush()
	recordPhase("warm_main", phaseStart)
	phaseStart = time.Now()
	goRuntime.GC()
	recordPhase("postwarm_gc", phaseStart)
	loadSubphases := make([]bytecodeProgramRuntimeLoadSubphase, 0, len(loadSubphaseByName))
	for _, name := range []string{"native_parse", "ast_mapping", "origin_annotation"} {
		if aggregate := loadSubphaseByName[name]; aggregate != nil {
			loadSubphases = append(loadSubphases, *aggregate)
		}
	}
	if err := writeBytecodeProgramRuntimePhases(cfg, time.Since(benchmarkEntry), phases, loadSubphases); err != nil {
		b.Fatalf("write runtime phases: %v", err)
	}
	interp.ResetBytecodeStats()
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
	if err := writeBytecodeProgramRuntimeStats(cfg, interp.BytecodeStats()); err != nil {
		b.Fatalf("write runtime stats: %v", err)
	}
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
	statsOut := filepath.Join(t.TempDir(), "stats.json")
	phaseOut := filepath.Join(t.TempDir(), "phases.json")
	loadCPUOut := filepath.Join(t.TempDir(), "load.cpu.pprof")
	traceOut := filepath.Join(t.TempDir(), "trace.json")
	argsJSON := `["wordlist.txt","second"]`

	t.Setenv(bytecodeRuntimeBenchTargetEnv, targetPath)
	t.Setenv(bytecodeRuntimeBenchRunFromEnv, runFrom)
	t.Setenv(bytecodeRuntimeBenchArgsJSONEnv, argsJSON)
	t.Setenv(bytecodeRuntimeBenchStatsOutEnv, statsOut)
	t.Setenv(bytecodeRuntimeBenchPhaseOutEnv, phaseOut)
	t.Setenv(bytecodeRuntimeBenchLoadCPUEnv, loadCPUOut)
	t.Setenv(bytecodeRuntimeBenchTraceOutEnv, traceOut)
	t.Setenv(bytecodeRuntimeBenchRetentionOutEnv, filepath.Join(t.TempDir(), "retention.json"))
	retentionHeapProfile := filepath.Join(t.TempDir(), "retention.heap.pprof")
	t.Setenv(bytecodeRuntimeBenchRetentionHeapProfileEnv, retentionHeapProfile)
	arrayOwnershipOut := filepath.Join(t.TempDir(), "array-ownership.json")
	t.Setenv(bytecodeRuntimeBenchArrayOwnershipOutEnv, arrayOwnershipOut)
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
	if cfg.StatsOutputPath != statsOut {
		t.Fatalf("stats output mismatch: got %q want %q", cfg.StatsOutputPath, statsOut)
	}
	if cfg.PhaseOutputPath != phaseOut {
		t.Fatalf("phase output mismatch: got %q want %q", cfg.PhaseOutputPath, phaseOut)
	}
	if cfg.LoadCPUProfilePath != loadCPUOut {
		t.Fatalf("load cpu profile mismatch: got %q want %q", cfg.LoadCPUProfilePath, loadCPUOut)
	}
	if cfg.TraceOutputPath != traceOut {
		t.Fatalf("trace output mismatch: got %q want %q", cfg.TraceOutputPath, traceOut)
	}
	if cfg.RetentionOutputPath == "" {
		t.Fatalf("retention output should be resolved")
	}
	if cfg.RetentionHeapProfilePath != retentionHeapProfile {
		t.Fatalf("retention heap profile mismatch: got %q want %q", cfg.RetentionHeapProfilePath, retentionHeapProfile)
	}
	if cfg.ArrayOwnershipOutputPath != arrayOwnershipOut {
		t.Fatalf("array ownership output mismatch: got %q want %q", cfg.ArrayOwnershipOutputPath, arrayOwnershipOut)
	}
	if cfg.TraceTop != 7 {
		t.Fatalf("trace top mismatch: got %d want 7", cfg.TraceTop)
	}
}

func TestWriteBytecodeProgramRuntimeRetentionHeapProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "retention.heap.pprof")
	if err := writeBytecodeProgramRuntimeRetentionHeapProfile(bytecodeProgramRuntimeBenchConfig{
		RetentionHeapProfilePath: path,
	}); err != nil {
		t.Fatalf("write retention heap profile: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat retention heap profile: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("retention heap profile is empty")
	}
}
